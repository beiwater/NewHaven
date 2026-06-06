import { audio } from '@/audio/AudioManager'
import { useEffect, useMemo, useRef, useState } from 'react'
import { Container, Graphics, Sprite, Text, Texture, type Application } from 'pixi.js'
import { useBuildings, useMoveBuilding, usePlaceBuilding } from '@/api/buildings.api'
import { useCompany } from '@/api/company.api'
import { useClaimProduction, useProductionJobs } from '@/api/production.api'
import { useUIStore } from '@/store/ui.store'
import { createBuildingNode } from './pixi/buildingLayer'
import { createGameApp } from './pixi/createApp'
import {
  allRealSlots,
  isSlotUnlocked,
  MAPS,
  nearestSlot,
  placeableSlots,
  type MapId,
  type MapSlot,
} from './map.config'
import { BUILDING_TEXTURE_URLS, loadImage, preloadBuildingTextures, preloadResourceTextures } from './map/textures'
import { addHarvestFloat, buildingMapId, buildingSlot, collectableAmount, jobProgress, normalizedMapId, slotPosition } from './map/mapGeometry'

const SLOT_W = 104
const SLOT_H = 70
const SLOT_PICK_RADIUS = 96

function GameCanvas() {
  const containerRef = useRef<HTMLDivElement>(null)
  const appRef = useRef<Application | null>(null)
  const stageScaleRef = useRef(1)
  const bgSizeRef = useRef({ w: 1, h: 1 })
  const zoomRef = useRef(1)
  const levelRef = useRef(1)
  const occupiedSlotIdsRef = useRef(new Set<string>())
  const currentMapIdRef = useRef<MapId>('harbor')
  const textureCacheRef = useRef<Record<number, Texture>>({})
  const resourceTextureCacheRef = useRef<Record<number, Texture>>({})
  const buildingLayerRef = useRef<Container | null>(null)
  const hotspotLayerRef = useRef<Container | null>(null)
  const overlayLayerRef = useRef<Container | null>(null)
  const mapContentRef = useRef<Container | null>(null)
  const bgSpriteRef = useRef<Sprite | null>(null)
  const startedRef = useRef(false)
  const [layerReady, setLayerReady] = useState(false)
  const [clock, setClock] = useState(Date.now())
  const [resizeWarning, setResizeWarning] = useState(false)
  const [toast, setToast] = useState<string | null>(null)

  const { data: buildingsData } = useBuildings()
  const { data: companyData } = useCompany()
  const { data: jobsData } = useProductionJobs()
  const claimProduction = useClaimProduction()
  const placeBuilding = usePlaceBuilding()
  const moveBuilding = useMoveBuilding()
  const placeBuildingRef = useRef(placeBuilding.mutate)
  const moveBuildingRef = useRef(moveBuilding.mutate)

  const level = companyData?.levelInfo?.level ?? 1
  const currentMapIdRaw = useUIStore((s) => s.currentMapId)
  const setCurrentMapId = useUIStore((s) => s.setCurrentMapId)
  const currentMapId = normalizedMapId(currentMapIdRaw, level)
  const currentMap = MAPS[currentMapId]
  const selectedBuildingId = useUIStore((s) => s.selectedBuildingId)
  const selectBuilding = useUIStore((s) => s.selectBuilding)
  const placementBuildingId = useUIStore((s) => s.placementBuildingId)
  const clearBuildingPlacement = useUIStore((s) => s.clearBuildingPlacement)
  const movingBuildingId = useUIStore((s) => s.movingBuildingId)
  const clearBuildingMove = useUIStore((s) => s.clearBuildingMove)



  const buildings = useMemo(
    () => Array.isArray(buildingsData) ? buildingsData.filter((b) => b.placed !== false && buildingMapId(b) === currentMapId) : [],
    [buildingsData, currentMapId],
  )

  const occupiedSlotIds = useMemo(() => new Set(buildings.map((building, index) => buildingSlot(building, level, index).id)), [buildings, level])
  const unlockedSlots = useMemo(() => placeableSlots(currentMapId, level), [currentMapId, level])
  const lockedSlots = useMemo(
    () => allRealSlots(currentMapId).filter((slot) => !isSlotUnlocked(slot, level)),
    [currentMapId, level],
  )
  const actionActive = Boolean(placementBuildingId || movingBuildingId)

  useEffect(() => {
    levelRef.current = level
    currentMapIdRef.current = currentMapId
    occupiedSlotIdsRef.current = occupiedSlotIds
  }, [currentMapId, level, occupiedSlotIds])

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
    let resizeObserver: ResizeObserver | null = null
    let app: Application | null = null
    const minZoom = 1
    const maxZoom = 4

    const clampPivot = () => {
      if (!app) return
      const screenW = app.screen.width
      const screenH = app.screen.height
      const baseScale = stageScaleRef.current
      const zoom = zoomRef.current
      const bgSize = bgSizeRef.current
      const bleed = 80 / zoom
      const maxPx = Math.max(0, bgSize.w * baseScale - screenW / zoom)
      const maxPy = Math.max(0, bgSize.h * baseScale - screenH / zoom)
      app.stage.pivot.x = Math.max(-bleed, Math.min(maxPx + bleed, app.stage.pivot.x))
      app.stage.pivot.y = Math.max(-bleed, Math.min(maxPy + bleed, app.stage.pivot.y))
    }

    const resizeMap = () => {
      if (!app || !mapContentRef.current || !bgSpriteRef.current) return
      const screenW = app.screen.width
      const screenH = app.screen.height
      if (screenW <= 0 || screenH <= 0) {
        setResizeWarning(true)
        return
      }
      const bgSize = bgSizeRef.current
      const baseScale = Math.max(screenW / bgSize.w, screenH / bgSize.h)
      stageScaleRef.current = baseScale
      mapContentRef.current.scale.set(baseScale)
      clampPivot()
      setResizeWarning(false)
    }

    const init = async () => {
      try {
        const firstBg = await loadImage(MAPS.harbor.background)
        if (destroyed) return
        app = await createGameApp(container)
        appRef.current = app
        app.renderer.background.color = 0xe8dcc8
        textureCacheRef.current = await preloadBuildingTextures()
        resourceTextureCacheRef.current = await preloadResourceTextures()

        const mapContent = new Container()
        mapContentRef.current = mapContent
        app.stage.addChild(mapContent)

        const bg = new Sprite(Texture.from(firstBg))
        bg.anchor.set(0, 0)
        bgSpriteRef.current = bg
        bgSizeRef.current = { w: bg.width, h: bg.height }
        mapContent.addChild(bg)

        const buildingLayer = new Container()
        buildingLayer.sortableChildren = true
        buildingLayerRef.current = buildingLayer
        mapContent.addChild(buildingLayer)

        const overlayLayer = new Container()
        overlayLayer.sortableChildren = true
        overlayLayerRef.current = overlayLayer
        const hotspotLayer = new Container()
        hotspotLayer.sortableChildren = true
        hotspotLayerRef.current = hotspotLayer
        mapContent.addChild(hotspotLayer)
        mapContent.addChild(overlayLayer)

        resizeMap()
        setLayerReady(true)

        const canvas = app.canvas as HTMLCanvasElement
        const screenToImage = (clientX: number, clientY: number) => {
          const rect = canvas.getBoundingClientRect()
          return {
            x: (app!.stage.pivot.x + (clientX - rect.left) / app!.stage.scale.x) / stageScaleRef.current,
            y: (app!.stage.pivot.y + (clientY - rect.top) / app!.stage.scale.y) / stageScaleRef.current,
          }
        }

        let panning = false
        let dragged = false
        const panStart = { x: 0, y: 0 }
        const pivotStart = { x: 0, y: 0 }

        let touches: Map<number, { clientX: number; clientY: number }> = new Map()
        let pinchStartDist = 0
        let pinchStartZoom = 1

        const tryUseSlot = (slot: MapSlot) => {
          const state = useUIStore.getState()
          if (occupiedSlotIdsRef.current.has(slot.id) || !isSlotUnlocked(slot, levelRef.current)) return
          if (state.movingBuildingId) {
            moveBuildingRef.current({ buildingId: state.movingBuildingId, mapId: slot.mapId, slotId: slot.id }, {
              onSuccess: () => { clearBuildingMove(); clearOverlay() },
            })
          } else if (state.placementBuildingId) {
            placeBuildingRef.current({ buildingId: state.placementBuildingId, mapId: slot.mapId, slotId: slot.id }, {
              onSuccess: () => clearBuildingPlacement(),
            })
          }
        }

        const tryHotspot = (imgX: number, imgY: number) => {
          const state = useUIStore.getState()
          if (state.movingBuildingId || state.placementBuildingId) return false
          const mapId = normalizedMapId(state.currentMapId, levelRef.current)
          const map = MAPS[mapId]
          const hotspot = map.hotspots.find((candidate) => Math.hypot(candidate.px - imgX, candidate.py - imgY) <= 70)
          if (!hotspot) return false
          if (hotspot.unlockLevel && levelRef.current < hotspot.unlockLevel) {
            setToast(`Unlocks at Level ${hotspot.unlockLevel}`)
            setTimeout(() => setToast(null), 2500)
            return true
          }
          // Reset zoom/pivot when switching maps
          zoomRef.current = 1
          app!.stage.scale.set(1)
          app!.stage.pivot.set(0, 0)
          setCurrentMapId(hotspot.targetMapId)
          return true
        }

        const slotAtPoint = (imgX: number, imgY: number) => {
          const mapId = currentMapIdRef.current
          const slots = allRealSlots(mapId)
          const slot = nearestSlot(mapId, imgX, imgY, slots)
          return Math.hypot(slot.px - imgX, slot.py - imgY) <= SLOT_PICK_RADIUS ? slot : null
        }

        const clearOverlay = () => overlayLayerRef.current?.removeChildren()

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
            pinchStartZoom = zoomRef.current
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
              zoomRef.current = Math.max(minZoom, Math.min(maxZoom, pinchStartZoom * (dist / pinchStartDist)))
              app!.stage.scale.set(zoomRef.current)
              clampPivot()
            }
          }
        }, { passive: false })

        canvas.addEventListener('touchend', (e: TouchEvent) => {
          for (let i = 0; i < e.changedTouches.length; i++) touches.delete(e.changedTouches[i].identifier)
          if (touches.size === 0 && !dragged && panning) {
            const img = screenToImage(panStart.x, panStart.y)
            const slot = slotAtPoint(img.x, img.y)
            if (slot) tryUseSlot(slot)
            else tryHotspot(img.x, img.y)
          }
          panning = touches.size === 1
          if (touches.size === 0) {
            pinchStartDist = 0
            clearOverlay()
          }
        })

        canvas.addEventListener('wheel', (e: WheelEvent) => {
          e.preventDefault()
          zoomRef.current = Math.max(minZoom, Math.min(maxZoom, zoomRef.current * (e.deltaY > 0 ? 0.92 : 1.08)))
          app!.stage.scale.set(zoomRef.current)
          clampPivot()
        }, { passive: false })

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
        })

        canvas.addEventListener('mouseup', (e: MouseEvent) => {
          if (!dragged && e.button === 0) {
            const img = screenToImage(e.clientX, e.clientY)
            const slot = slotAtPoint(img.x, img.y)
            if (slot) tryUseSlot(slot)
            else tryHotspot(img.x, img.y)
          }
          panning = false
          clearOverlay()
        })

        canvas.addEventListener('mouseleave', () => { panning = false })

        resizeObserver = new ResizeObserver(() => {
          app?.renderer.resize(container.clientWidth, container.clientHeight)
          resizeMap()
        })
        resizeObserver.observe(container)
      } catch (e) {
        console.error('[GameCanvas]', e)
        if (container) container.innerHTML = '<div style="padding:20px;color:#8b7355">Map unavailable</div>'
      }
    }
    init()
    return () => {
      destroyed = true
      startedRef.current = false
      resizeObserver?.disconnect()
      if (app) { app.destroy(true); appRef.current = null }
      buildingLayerRef.current = null
      overlayLayerRef.current = null
      hotspotLayerRef.current = null
      mapContentRef.current = null
      bgSpriteRef.current = null
      setLayerReady(false)
    }
  }, [clearBuildingMove, clearBuildingPlacement, setCurrentMapId])

  useEffect(() => {
    const sprite = bgSpriteRef.current
    const mapContent = mapContentRef.current
    const app = appRef.current
    if (!sprite || !mapContent || !app) return
    let cancelled = false
    loadImage(currentMap.background).then((img) => {
      if (cancelled) return
      sprite.texture = Texture.from(img)
      sprite.width = img.width
      sprite.height = img.height
      bgSizeRef.current = { w: img.width, h: img.height }
      const baseScale = Math.max(app.screen.width / img.width, app.screen.height / img.height)
      stageScaleRef.current = baseScale
      mapContent.scale.set(baseScale)
      app.stage.pivot.set(0, 0)
      zoomRef.current = 1
      app.stage.scale.set(1)
    }).catch((err) => console.warn('[GameCanvas] map background failed', err))
    return () => { cancelled = true }
  }, [currentMap.background])
  useEffect(() => {
    // Draw hotspot markers (always visible)
    const hLayer = hotspotLayerRef.current
    if (hLayer) {
      hLayer.removeChildren().forEach((child) => child.destroy({ children: true }))
      for (const hotspot of currentMap.hotspots) {
        const locked = Boolean(hotspot.unlockLevel && level < hotspot.unlockLevel)
        const node = new Container()
        node.x = hotspot.px
        node.y = hotspot.py
        const marker = new Graphics()
        marker.circle(0, 0, 28)
        marker.fill({ color: locked ? 0x5c3d2e : 0x2563eb, alpha: 0.62 })
        marker.stroke({ width: 2, color: 0xfff4c7, alpha: 0.78 })
        node.addChild(marker)
        const label = new Text({
          text: locked ? `Lv.${hotspot.unlockLevel}` : hotspot.label,
          style: { fontSize: 12, fontWeight: 'bold', fill: 0xfff4c7, fontFamily: 'Inter, system-ui, sans-serif' },
        })
        label.anchor.set(0.5)
        node.addChild(label)
        hLayer.addChild(node)
      }
    }
    // Draw slot overlays (only when placing/moving)
    const layer = overlayLayerRef.current
    if (!layer) return
    layer.removeChildren().forEach((child) => child.destroy({ children: true }))
    if (actionActive) {
      for (const slot of unlockedSlots) {
        if (occupiedSlotIds.has(slot.id)) continue
        drawSlot(layer, slot, 0x3b82f6, 0.16, 0.62)
      }
      for (const slot of lockedSlots) {
        drawSlot(layer, slot, 0xd4a843, 0.10, 0.46, true)
      }
    }
  }, [actionActive, currentMap.hotspots, level, lockedSlots, occupiedSlotIds, unlockedSlots, layerReady])

  useEffect(() => {
    const layer = buildingLayerRef.current
    if (!layer) return
    layer.removeChildren().forEach((child) => child.destroy({ children: true }))
    const cache = textureCacheRef.current
    const resourceCache = resourceTextureCacheRef.current
    const jobs = jobsData ?? []
    const now = clock
    const pulse = (Math.sin(now / 650) + 1) / 2
    for (const [index, b] of buildings.entries()) {
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
            if (currentLayer) addHarvestFloat(currentLayer, slotPosition(buildingSlot(b, level, index)), resourceCache[harvestJob.resourceId], amount)
          },
        })
      })
      const pos = slotPosition(buildingSlot(b, level, index))
      node.x = pos.x
      node.y = pos.y
      node.zIndex = pos.y * 100 + pos.x
      layer.addChild(node)
    }
  }, [buildings, claimProduction, clock, jobsData, layerReady, level, selectBuilding, selectedBuildingId])

  return (
    <div ref={containerRef} className="relative h-full w-full bg-[#e8dcc8]">
      <div className="pointer-events-none absolute left-4 top-4 z-10 rounded-md border border-amber-700/30 bg-amber-50/95 px-3 py-2 text-xs font-semibold text-amber-900 shadow">
        {currentMap.name}
        {currentMap.unlockLevel > 1 ? ` · Lv.${currentMap.unlockLevel}` : ' · 出发港口'}
      </div>
      {placementBuildingId && (
        <div className="pointer-events-none absolute left-4 top-14 z-10 rounded-md border border-blue-500/30 bg-blue-50/95 px-3 py-2 text-xs font-semibold text-blue-800 shadow">
          Blue plots are available. Gold plots unlock as you level up.
        </div>
      )}
      {movingBuildingId && (
        <div className="pointer-events-none absolute left-4 top-14 z-10 rounded-md border border-blue-500/30 bg-blue-50/95 px-3 py-2 text-xs font-semibold text-blue-800 shadow">
          Move to a blue plot on this map.
        </div>
      )}
      {resizeWarning && (
        <div className="pointer-events-none absolute right-4 top-4 z-10 rounded-md border border-red-300 bg-red-50/95 px-3 py-2 text-xs font-semibold text-red-700 shadow">
          Map view changed. Refresh if placement feels off.
        </div>
      )}
      {toast && (
        <div className="pointer-events-none absolute left-1/2 top-1/4 z-20 -translate-x-1/2 rounded-lg border border-amber-600/40 bg-amber-900/90 px-5 py-3 text-sm font-bold text-amber-100 shadow-lg">
          {toast}
        </div>
      )}
    </div>
  )
}

function drawSlot(layer: Container, slot: MapSlot, color: number, fillAlpha: number, strokeAlpha: number, locked = false) {
  const cell = new Graphics()
  cell.roundRect(slot.px - SLOT_W / 2, slot.py - SLOT_H / 2, SLOT_W, SLOT_H, 12)
  cell.fill({ color, alpha: fillAlpha })
  cell.stroke({ width: 2, color, alpha: strokeAlpha })
  layer.addChild(cell)
  if (locked) {
    const text = new Text({
      text: 'LOCK',
      style: { fontSize: 10, fontWeight: 'bold', fill: 0x5c3d2e, fontFamily: 'Inter, system-ui, sans-serif' },
    })
    text.anchor.set(0.5)
    text.x = slot.px
    text.y = slot.py
    layer.addChild(text)
  }
}

export default GameCanvas
