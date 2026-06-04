import { audio } from '@/audio/AudioManager'
import { useEffect, useMemo, useRef, useState } from 'react'
import { IMAGE_LOAD_TIMEOUT_MS } from '@/constants'

import { Assets, Texture, Sprite, Container, Text, Graphics, type Application } from 'pixi.js'
import { useBuildings, usePlaceBuilding, useMoveBuilding } from '@/api/buildings.api'
import { useProductionJobs, useClaimProduction } from '@/api/production.api'
import { useUIStore } from '@/store/ui.store'
import { createBuildingNode } from './pixi/buildingLayer'
import { createGameApp } from './pixi/createApp'
import { resourceIcon } from '@/game/resources'
import { buildingIcon } from '@/game/icons'
import type { Building, ProductionJob } from './types'

const MAP_BG = '/assets/backgrounds/map_background_v1.png'
const BUILDING_TEXTURE_URLS: Record<number, string> = {
  1: '/assets/buildings/grain_plot_lv1_idle_trimmed.png',
  2: buildingIcon(2),
  3: '/assets/buildings/mill_house_lv1_idle_trimmed.png',
  4: buildingIcon(4),
  5: '/assets/buildings/bakery_shop_lv1_idle_trimmed.png',
  6: buildingIcon(6),
  7: buildingIcon(7),
  8: buildingIcon(8),
  9: '/assets/buildings/meal_kiosk_lv1_idle_trimmed.png',
  10: buildingIcon(10),
  11: buildingIcon(11),
  12: buildingIcon(12),
}
const RESOURCE_TEXTURE_URLS: Record<number, string> = Object.fromEntries(
  Array.from({ length: 12 }, (_, i) => [i + 1, resourceIcon(i + 1)]),
)

function loadImage(url: string): Promise<HTMLImageElement> {
  const img = new Image()
  const { promise, resolve, reject } = Promise.withResolvers<HTMLImageElement>()
  const t = setTimeout(() => reject(new Error(`Timeout: ${url}`)), IMAGE_LOAD_TIMEOUT_MS)
  img.onload = () => { clearTimeout(t); resolve(img) }
  img.onerror = () => { clearTimeout(t); reject(new Error(`Failed: ${url}`)) }
  img.src = url
  return promise
}

async function preloadBuildingTextures(): Promise<Record<number, Texture>> {
  const urls = Object.values(BUILDING_TEXTURE_URLS)
  const loaded: Record<string, Texture> = await Assets.load(urls)
  const cache: Record<number, Texture> = {}
  for (const [kindStr, url] of Object.entries(BUILDING_TEXTURE_URLS)) {
    if (loaded[url]) cache[Number(kindStr)] = loaded[url]
  }
  return cache
}

async function preloadResourceTextures(): Promise<Record<number, Texture>> {
  const urls = Object.values(RESOURCE_TEXTURE_URLS)
  const loaded: Record<string, Texture> = await Assets.load(urls)
  const cache: Record<number, Texture> = {}
  for (const [idStr, url] of Object.entries(RESOURCE_TEXTURE_URLS)) {
    if (loaded[url]) cache[Number(idStr)] = loaded[url]
  }
  return cache
}

function jobProgress(job: ProductionJob, now: number): number {
  const start = new Date(job.startedAt).getTime()
  const end = new Date(job.completesAt).getTime()
  if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) return 1
  return Math.max(0, Math.min(1, (now - start) / (end - start)))
}

function accruedAmount(job: ProductionJob, now: number): number {
  if (job.status === 'claimed') return job.amount
  return Math.floor(jobProgress(job, now) * job.amount)
}

function collectableAmount(job: ProductionJob, now: number): number {
  if (job.status === 'claimed') return 0
  return Math.max(job.claimableAmount ?? 0, accruedAmount(job, now) - (job.claimedAmount ?? 0), 0)
}

function addHarvestFloat(layer: Container, pos: { x: number; y: number }, texture: Texture | undefined, amount: number) {
  const node = new Container()
  node.x = pos.x + 52
  node.y = pos.y + 4

  const bubble = new Graphics()
  bubble.roundRect(-12, -18, 68, 28, 10)
  bubble.fill({ color: 0xfff7dc, alpha: 0.96 })
  bubble.stroke({ width: 1, color: 0x7aa34b, alpha: 0.7 })
  node.addChild(bubble)

  if (texture) {
    const icon = new Sprite(texture)
    const scale = Math.min(18 / Math.max(1, texture.width), 18 / Math.max(1, texture.height))
    icon.scale.set(scale)
    icon.anchor.set(0.5)
    icon.x = 0
    icon.y = -4
    node.addChild(icon)
  }

  const text = new Text({
    text: `+${amount}`,
    style: { fontSize: 12, fontWeight: 'bold', fill: 0x3f6212, fontFamily: 'Inter, system-ui, sans-serif' },
  })
  text.anchor.set(0, 0.5)
  text.x = 12
  text.y = -4
  node.addChild(text)
  layer.addChild(node)

  const start = performance.now()
  const tick = () => {
    if (node.destroyed) return
    const elapsed = performance.now() - start
    const pct = Math.min(1, elapsed / 900)
    node.y = pos.y + 4 - pct * 34
    node.alpha = 1 - pct
    if (pct >= 1) {
      node.destroy({ children: true })
      return
    }
    requestAnimationFrame(tick)
  }
  requestAnimationFrame(tick)
}

function GameCanvas() {
  const containerRef = useRef<HTMLDivElement>(null)
  const appRef = useRef<Application | null>(null)
  const textureCacheRef = useRef<Record<number, Texture>>({})
  const resourceTextureCacheRef = useRef<Record<number, Texture>>({})
  const buildingLayerRef = useRef<Container | null>(null)
  const startedRef = useRef(false)
  const [layerReady, setLayerReady] = useState(false)
  const [clock, setClock] = useState(Date.now())
  const moveOverlayRef = useRef<Container | null>(null)

  const { data: buildingsData } = useBuildings()
  const { data: jobsData } = useProductionJobs()
  const claimProduction = useClaimProduction()
  const placeBuilding = usePlaceBuilding()
  const moveBuilding = useMoveBuilding()
  const placeBuildingRef = useRef(placeBuilding.mutate)
  const moveBuildingRef = useRef(moveBuilding.mutate)

  const selectedBuildingId = useUIStore((s) => s.selectedBuildingId)
  const selectBuilding = useUIStore((s) => s.selectBuilding)
  const placementBuildingId = useUIStore((s) => s.placementBuildingId)
  const clearBuildingPlacement = useUIStore((s) => s.clearBuildingPlacement)
  const movingBuildingId = useUIStore((s) => s.movingBuildingId)
  const clearBuildingMove = useUIStore((s) => s.clearBuildingMove)

  const buildings = useMemo(
    () => Array.isArray(buildingsData) ? buildingsData.filter((b) => b.placed !== false) : [],
    [buildingsData],
  )

  useEffect(() => {
    const timer = setInterval(() => setClock(Date.now()), 1000)
    return () => clearInterval(timer)
  }, [])

  useEffect(() => {
    placeBuildingRef.current = placeBuilding.mutate
    moveBuildingRef.current = moveBuilding.mutate
  }, [placeBuilding.mutate, moveBuilding.mutate])

  useEffect(() => {
    if (startedRef.current) return
    startedRef.current = true
    const container = containerRef.current
    if (!container) return
    let destroyed = false
    let app: Application | null = null

    const init = async () => {
      try {
        const img = await loadImage(MAP_BG)
        if (destroyed) return
        app = await createGameApp(container)
        appRef.current = app
        app.renderer.background.color = 0xe8dcc8
        textureCacheRef.current = await preloadBuildingTextures()
        resourceTextureCacheRef.current = await preloadResourceTextures()
        const sprite = new Sprite(Texture.from(img))
        sprite.anchor.set(0, 0)

        const screenW = app.screen.width
        const screenH = app.screen.height
        const baseScale = Math.max(screenW / sprite.width, screenH / sprite.height)
        const mapW = sprite.width   // image-pixel width
        const mapH = sprite.height  // image-pixel height

        // Content container at baseScale — everything inside uses image-pixel coords
        const mapContent = new Container()
        mapContent.scale.set(baseScale)
        app.stage.addChild(mapContent)
        mapContent.addChild(sprite)

        let zoom = 1.0
        const minZoom = 1.0; const maxZoom = 4.0
        const clampPivot = () => {
          const bleed = 180 / zoom
          // pivot operates in stage coords; mapW * baseScale = stage width
          const maxPx = Math.max(0, mapW * baseScale - screenW / zoom)
          const maxPy = Math.max(0, mapH * baseScale - screenH / zoom)
          app!.stage.pivot.x = Math.max(-bleed, Math.min(maxPx + bleed, app!.stage.pivot.x))
          app!.stage.pivot.y = Math.max(-bleed, Math.min(maxPy + bleed, app!.stage.pivot.y))
        }

        const buildingLayer = new Container()
        buildingLayer.sortableChildren = true
        buildingLayerRef.current = buildingLayer
        mapContent.addChild(buildingLayer)

        const moveOverlayLayer = new Container()
        moveOverlayLayer.sortableChildren = true
        moveOverlayLayer.name = 'moveOverlay'
        moveOverlayRef.current = moveOverlayLayer
        mapContent.addChild(moveOverlayLayer)
        const drawOverlay = (grid: { x: number; y: number } | null, showGhost: boolean) => {
          moveOverlayLayer.removeChildren()
          if (!grid) return
          const bw = 100
          const bh = 80
          // Grid cells — centered on grid position to match isometric tiles
          for (let gy = 1; gy <= 10; gy++) {
            for (let gx = 1; gx <= 12; gx++) {
              const cx = 110 + gx * 112 + gy * 36
              const cy = 95 + gy * 74
              const cell = new Graphics()
              const isHovered = gx === grid.x && gy === grid.y
              cell.rect(cx - 50, cy - 40, bw, bh)
              if (isHovered) {
                cell.fill({ color: 0x3b82f6, alpha: 0.20 })
                cell.stroke({ width: 2, color: 0x3b82f6, alpha: 0.60 })
              } else {
                cell.stroke({ width: 1, color: 0xffffff, alpha: 0.30 })
              }
              moveOverlayLayer.addChild(cell)
            }
          }
          // Ghost preview
          if (showGhost && grid) {
            const cx = 110 + grid.x * 112 + grid.y * 36
            const cy = 95 + grid.y * 74
            const ghost = new Graphics()
            ghost.rect(cx - 40, cy - 40, 80, 80)
            ghost.stroke({ width: 2, color: 0x3b82f6, alpha: 0.75 })
            moveOverlayLayer.addChild(ghost)
          }
        }
        const clearOverlay = () => moveOverlayLayer.removeChildren()
        setLayerReady(true)
        const canvas = app.canvas as HTMLCanvasElement

        // screenToMap: screen CSS coords → stage world coords
        const screenToMap = (clientX: number, clientY: number) => {
          const rect = canvas.getBoundingClientRect()
          return {
            x: app!.stage.pivot.x + (clientX - rect.left) / app!.stage.scale.x,
            y: app!.stage.pivot.y + (clientY - rect.top) / app!.stage.scale.y,
          }
        }
        // toGrid: image-pixel coords → grid coords (1..12, 1..10)
        const toGrid = (imgX: number, imgY: number) => {
          const y = Math.round((imgY - 95) / 74)
          const x = Math.round((imgX - 110 - y * 36) / 112)
          return { x: Math.max(1, Math.min(12, x)), y: Math.max(1, Math.min(10, y)) }
        }

        let panning = false; let dragged = false
        const panStart = { x: 0, y: 0 }; const pivotStart = { x: 0, y: 0 }

         // ── Touch events (mobile) ──
        let touches: Map<number, { clientX: number; clientY: number }> = new Map()
        let pinchStartDist = 0
        let pinchStartZoom = 1

        canvas.addEventListener('touchstart', (e: TouchEvent) => {
          e.preventDefault()
          for (let i = 0; i < e.changedTouches.length; i++) {
            const t = e.changedTouches[i]
            touches.set(t.identifier, { clientX: t.clientX, clientY: t.clientY })
          }
          if (touches.size === 1) {
            panning = true; dragged = false
            const first = touches.values().next().value!
            panStart.x = first.clientX; panStart.y = first.clientY
            pivotStart.x = app!.stage.pivot.x; pivotStart.y = app!.stage.pivot.y
          }
          if (touches.size === 2) {
            panning = false
            const pts = [...touches.values()]
            pinchStartDist = Math.hypot(pts[0].clientX - pts[1].clientX, pts[0].clientY - pts[1].clientY)
            pinchStartZoom = zoom
          }
        }, { passive: false })

        canvas.addEventListener('touchmove', (e: TouchEvent) => {
          e.preventDefault()
          for (let i = 0; i < e.changedTouches.length; i++) {
            const t = e.changedTouches[i]
            touches.set(t.identifier, { clientX: t.clientX, clientY: t.clientY })
          }
          if (touches.size === 1 && panning) {
            const first = touches.values().next().value!
            if (Math.abs(first.clientX - panStart.x) + Math.abs(first.clientY - panStart.y) > 6) dragged = true
            app!.stage.pivot.x = pivotStart.x - (first.clientX - panStart.x) / app!.stage.scale.x
            app!.stage.pivot.y = pivotStart.y - (first.clientY - panStart.y) / app!.stage.scale.y
            clampPivot()
          }
          if (touches.size === 2) {
            const pts = [...touches.values()]
            const dist = Math.hypot(pts[0].clientX - pts[1].clientX, pts[0].clientY - pts[1].clientY)
            if (pinchStartDist > 0) {
              const scale = dist / pinchStartDist
              const next = Math.max(minZoom, Math.min(maxZoom, pinchStartZoom * scale))
              if (next !== zoom) { zoom = next; app!.stage.scale.set(zoom); clampPivot() }
            }
          }
        }, { passive: false })

        canvas.addEventListener('touchend', (e: TouchEvent) => {
          for (let i = 0; i < e.changedTouches.length; i++) {
            touches.delete(e.changedTouches[i].identifier)
          }
          if (touches.size === 0 && !dragged && panning) {
            const state = useUIStore.getState()
            const world = screenToMap(panStart.x, panStart.y)
            // Account for building visual offset (sprite foot at node.x+50, node.y+80)
            const grid = toGrid(world.x / baseScale - 50, world.y / baseScale - 80)
            if (state.movingBuildingId) {
              moveBuildingRef.current({ buildingId: state.movingBuildingId, x: grid.x, y: grid.y }, {
                onSuccess: () => { clearBuildingMove(); clearOverlay() },
              })
            } else if (state.placementBuildingId) {
              placeBuildingRef.current({ buildingId: state.placementBuildingId, x: grid.x, y: grid.y }, {
                onSuccess: () => clearBuildingPlacement(),
              })
            }
          }
          if (touches.size === 1) {
            // Transition from pinch to pan
            panning = true; dragged = false
            const first = touches.values().next().value!
            panStart.x = first.clientX; panStart.y = first.clientY
            pivotStart.x = app!.stage.pivot.x; pivotStart.y = app!.stage.pivot.y
          }
          if (touches.size === 0) {
            panning = false
            pinchStartDist = 0
            clearOverlay()
          }
        })
        canvas.addEventListener('wheel', (e: WheelEvent) => {
          e.preventDefault()
          const factor = e.deltaY > 0 ? 0.92 : 1.08
          const next = Math.max(minZoom, Math.min(maxZoom, zoom * factor))
          if (next !== zoom) { zoom = next; app!.stage.scale.set(zoom); clampPivot() }
        }, { passive: false })
        // ── Mouse events (desktop) ──
        canvas.addEventListener('mousedown', (e: MouseEvent) => {
          if (e.button !== 0 && e.button !== 1) return
          panning = true; dragged = false
          panStart.x = e.clientX; panStart.y = e.clientY
          pivotStart.x = app!.stage.pivot.x; pivotStart.y = app!.stage.pivot.y
        })
        canvas.addEventListener('mousemove', (e: MouseEvent) => {
          if (panning) {
            if (Math.abs(e.clientX - panStart.x) + Math.abs(e.clientY - panStart.y) > 6) dragged = true
            app!.stage.pivot.x = pivotStart.x - (e.clientX - panStart.x) / app!.stage.scale.x
            app!.stage.pivot.y = pivotStart.y - (e.clientY - panStart.y) / app!.stage.scale.y
            clampPivot()
          }
          // Ghost preview tracking during move/placement
          const state = useUIStore.getState()
          if (state.movingBuildingId || state.placementBuildingId) {
            const world = screenToMap(e.clientX, e.clientY)
            const grid = toGrid(world.x / baseScale - 50, world.y / baseScale - 80)
            drawOverlay(grid, !!state.movingBuildingId)
          } else {
            clearOverlay()
          }
        })
        canvas.addEventListener('mouseup', (e: MouseEvent) => {
          if (!dragged && e.button === 0) {
            const state = useUIStore.getState()
            const world = screenToMap(e.clientX, e.clientY)
            // Account for building visual offset (sprite foot at node.x+50, node.y+80)
            const grid = toGrid(world.x / baseScale - 50, world.y / baseScale - 80)
            if (state.movingBuildingId) {
              moveBuildingRef.current({ buildingId: state.movingBuildingId, x: grid.x, y: grid.y }, {
                onSuccess: () => { clearBuildingMove(); clearOverlay() },
              })
            } else if (state.placementBuildingId) {
              placeBuildingRef.current({ buildingId: state.placementBuildingId, x: grid.x, y: grid.y }, {
                onSuccess: () => clearBuildingPlacement(),
              })
            }
          }
          panning = false
          clearOverlay()
        })
        canvas.addEventListener('mouseleave', () => {
          panning = false
        })
      } catch (e) {
        console.error('[GameCanvas]', e)
        if (container) container.innerHTML = '<div style="padding:20px;color:#8b7355">Map unavailable</div>'
      }
    }
    init()
    return () => {
      destroyed = true; startedRef.current = false
      if (app) { app.destroy(true); appRef.current = null }
      buildingLayerRef.current = null
      setLayerReady(false)
    }
  }, [clearBuildingPlacement, clearBuildingMove])

  useEffect(() => {
    const layer = buildingLayerRef.current
    if (!layer) return
    layer.removeChildren().forEach((c) => c.destroy({ children: true }))
    const cache = textureCacheRef.current
    const resourceCache = resourceTextureCacheRef.current
    const jobs = jobsData ?? []
    const now = clock
    const pulse = (Math.sin(now / 650) + 1) / 2
    for (const b of buildings) {
      const tex = cache[b.kind] ?? Texture.from(BUILDING_TEXTURE_URLS[b.kind] ?? BUILDING_TEXTURE_URLS[2])
      const buildingJobs = jobs.filter((job) => job.buildingId === b.id && job.status !== 'claimed')
      const harvestJob = buildingJobs.find((job) => collectableAmount(job, now) > 0)
      const activeJob = harvestJob ?? buildingJobs[0]
      const collectable = harvestJob ? collectableAmount(harvestJob, now) : 0
      const progress = activeJob ? jobProgress(activeJob, now) : undefined
      const node = createBuildingNode(tex, {
        name: b.name ?? `Building ${b.kind}`,
        level: b.level ?? 1,
        status: activeJob ? activeJob.status : (b.status ?? 'idle'),
        progress,
        collectableAmount: collectable,
        pulse,
        resourceTexture: activeJob ? resourceCache[activeJob.resourceId] : undefined,
        isSelected: b.id === selectedBuildingId,
      }, () => {
        const state = useUIStore.getState()
        if (state.movingBuildingId || state.placementBuildingId) return
        selectBuilding(b.id)
      }, () => {
        if (!harvestJob) return
        const amount = collectableAmount(harvestJob, Date.now())
        claimProduction.mutate(harvestJob.id, {
          onSuccess: () => {
            audio.playSfx('resource_pickup')
            const currentLayer = buildingLayerRef.current
            if (currentLayer) {
              addHarvestFloat(currentLayer, buildingToMapPosition(b), resourceCache[harvestJob.resourceId], amount)
            }
          },
        })
      })
      const pos = buildingToMapPosition(b)
      node.x = pos.x; node.y = pos.y; node.zIndex = pos.y * 100 + pos.x
      layer.addChild(node)
    }
  }, [buildings, jobsData, selectedBuildingId, selectBuilding, layerReady, claimProduction, clock])

  return (
    <div ref={containerRef} className="relative h-full w-full bg-[#e8dcc8]">
      {placementBuildingId && (
        <div className="pointer-events-none absolute left-4 top-4 z-10 rounded-md border border-amber-700/30 bg-amber-50/95 px-3 py-2 text-xs font-semibold text-amber-900 shadow">
          Click a map plot to place this building. Drag to pan.
        </div>
      )}
      {movingBuildingId && (
        <div className="pointer-events-none absolute left-4 top-4 z-10 rounded-md border border-blue-500/30 bg-blue-50/95 px-3 py-2 text-xs font-semibold text-blue-800 shadow">
          Click a map plot to move this building. Drag to pan.
        </div>
      )}

    </div>
  )
}

function buildingToMapPosition(building: Building): { x: number; y: number } {
  const x = building.x ?? 0; const y = building.y ?? 0
  return { x: 110 + x * 112 + y * 36, y: 95 + y * 74 }
}

export default GameCanvas
