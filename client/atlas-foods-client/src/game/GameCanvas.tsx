import { useEffect, useMemo, useRef } from 'react'
import { IMAGE_LOAD_TIMEOUT_MS } from '@/constants'

import { Texture, Sprite, Container, type Application } from 'pixi.js'
import { useBuildings, usePlaceBuilding } from '@/api/buildings.api'
import { useUIStore } from '@/store/ui.store'
import { createBuildingNode } from './pixi/buildingLayer'
import { createGameApp } from './pixi/createApp'
import type { Building } from './types'

const MAP_BG = '/assets/backgrounds/map_background_v1.png'
const BUILDING_TEXTURES: Record<number, string> = {
  1: '/assets/buildings/grain_plot_lv1_idle_trimmed.png',
  2: '/assets/buildings/mill_house_lv1_idle_trimmed.png',
  3: '/assets/buildings/bakery_shop_lv1_idle_trimmed.png',
  4: '/assets/buildings/meal_kiosk_lv1_idle_trimmed.png',
}

function loadImage(url: string): Promise<HTMLImageElement> {
  const img = new Image()
  const { promise, resolve, reject } = Promise.withResolvers<HTMLImageElement>()
  const t = setTimeout(() => reject(new Error(`Timeout: ${url}`)), IMAGE_LOAD_TIMEOUT_MS)
  img.onload = () => { clearTimeout(t); resolve(img) }
  img.onerror = () => { clearTimeout(t); reject(new Error(`Failed: ${url}`)) }
  img.src = url
  return promise
}

function GameCanvas() {
  const containerRef = useRef<HTMLDivElement>(null)
  const appRef = useRef<Application | null>(null)
  const buildingLayerRef = useRef<Container | null>(null)
  const startedRef = useRef(false)
  const { data: buildingsData } = useBuildings()
  const placeBuilding = usePlaceBuilding()
  const placeBuildingRef = useRef(placeBuilding.mutate)
  const selectedBuildingId = useUIStore((s) => s.selectedBuildingId)
  const selectBuilding = useUIStore((s) => s.selectBuilding)
  const placementBuildingId = useUIStore((s) => s.placementBuildingId)
  const clearBuildingPlacement = useUIStore((s) => s.clearBuildingPlacement)
  const buildings = useMemo(
    () => Array.isArray(buildingsData)
      ? buildingsData.filter((building) => building.placed !== false)
      : [],
    [buildingsData],
  )

  useEffect(() => {
    placeBuildingRef.current = placeBuilding.mutate
  }, [placeBuilding.mutate])

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

        const sprite = new Sprite(Texture.from(img))
        sprite.anchor.set(0, 0)
        const screenW = app.screen.width
        const screenH = app.screen.height
        const imgW = sprite.width
        const imgH = sprite.height
        const baseScale = Math.max(screenW / imgW, screenH / imgH)
        sprite.scale.set(baseScale)
        app.stage.addChild(sprite)

        const mapW = imgW * baseScale
        const mapH = imgH * baseScale
        let zoom = 1.0
        const minZoom = 1.0
        const maxZoom = 4.0

        const clampPivot = () => {
          const bleed = 180 / zoom
          const maxPx = Math.max(0, mapW - screenW / zoom)
          const maxPy = Math.max(0, mapH - screenH / zoom)
          app!.stage.pivot.x = Math.max(-bleed, Math.min(maxPx + bleed, app!.stage.pivot.x))
          app!.stage.pivot.y = Math.max(-bleed, Math.min(maxPy + bleed, app!.stage.pivot.y))
        }

        const buildingLayer = new Container()
        buildingLayer.sortableChildren = true
        buildingLayerRef.current = buildingLayer
        app.stage.addChild(buildingLayer)
        const canvas = app.canvas as HTMLCanvasElement

        const screenToMap = (clientX: number, clientY: number) => {
          const rect = canvas.getBoundingClientRect()
          return {
            x: app!.stage.pivot.x + (clientX - rect.left) / app!.stage.scale.x,
            y: app!.stage.pivot.y + (clientY - rect.top) / app!.stage.scale.y,
          }
        }

        const toGrid = (worldX: number, worldY: number) => {
          const y = Math.round((worldY - 95) / 74)
          const x = Math.round((worldX - 110 - y * 36) / 112)
          return {
            x: Math.max(1, Math.min(12, x)),
            y: Math.max(1, Math.min(10, y)),
          }
        }

        canvas.addEventListener('wheel', (e: WheelEvent) => {
          e.preventDefault()
          const factor = e.deltaY > 0 ? 0.92 : 1.08
          const next = Math.max(minZoom, Math.min(maxZoom, zoom * factor))
          if (next !== zoom) {
            zoom = next
            app!.stage.scale.set(zoom)
            clampPivot()
          }
        }, { passive: false })

        let panning = false
        let dragged = false
        const panStart = { x: 0, y: 0 }
        const pivotStart = { x: 0, y: 0 }

        canvas.addEventListener('mousedown', (e: MouseEvent) => {
          if (e.button !== 0 && e.button !== 1) return
          panning = true
          dragged = false
          panStart.x = e.clientX
          panStart.y = e.clientY
          pivotStart.x = app!.stage.pivot.x
          pivotStart.y = app!.stage.pivot.y
        })

        canvas.addEventListener('mousemove', (e: MouseEvent) => {
          if (!panning) return
          if (Math.abs(e.clientX - panStart.x) + Math.abs(e.clientY - panStart.y) > 6) {
            dragged = true
          }
          app!.stage.pivot.x = pivotStart.x - (e.clientX - panStart.x) / app!.stage.scale.x
          app!.stage.pivot.y = pivotStart.y - (e.clientY - panStart.y) / app!.stage.scale.y
          clampPivot()
        })

        canvas.addEventListener('mouseup', (e: MouseEvent) => {
          if (!dragged && e.button === 0) {
            const buildingId = useUIStore.getState().placementBuildingId
            if (buildingId) {
              const world = screenToMap(e.clientX, e.clientY)
              const grid = toGrid(world.x, world.y)
              placeBuildingRef.current({ buildingId, x: grid.x, y: grid.y }, {
                onSuccess: () => clearBuildingPlacement(),
              })
            }
          }
          panning = false
        })
      } catch (e) {
        console.error('[GameCanvas]', e)
        if (container) container.innerHTML = '<div style="padding:20px;color:#8b7355">Map unavailable</div>'
      }
    }
    init()

    return () => {
      destroyed = true
      startedRef.current = false
      if (app) { app.destroy(true); appRef.current = null }
      buildingLayerRef.current = null
    }
  }, [clearBuildingPlacement])

  useEffect(() => {
    const layer = buildingLayerRef.current
    if (!layer) return

    layer.removeChildren().forEach((child) => child.destroy({ children: true }))

    for (const building of buildings) {
      const texture = Texture.from(BUILDING_TEXTURES[building.kind] ?? BUILDING_TEXTURES[2])
      const node = createBuildingNode(
        texture,
        {
          name: building.name ?? `Building ${building.kind}`,
          level: building.level ?? 1,
          status: building.status ?? 'idle',
          isSelected: building.id === selectedBuildingId,
        },
        () => selectBuilding(building.id),
      )
      const pos = buildingToMapPosition(building)
      node.x = pos.x
      node.y = pos.y
      node.zIndex = pos.y
      layer.addChild(node)
    }
  }, [buildings, selectedBuildingId, selectBuilding])

  return (
    <div ref={containerRef} className="relative h-full w-full bg-[#e8dcc8]">
      {placementBuildingId && (
        <div className="pointer-events-none absolute left-4 top-4 z-10 rounded-md border border-amber-700/30 bg-amber-50/95 px-3 py-2 text-xs font-semibold text-amber-900 shadow">
          Click a map plot to place this building. Drag to pan.
        </div>
      )}
    </div>
  )
}

function buildingToMapPosition(building: Building): { x: number; y: number } {
  const x = building.x ?? 0
  const y = building.y ?? 0
  return {
    x: 110 + x * 112 + y * 36,
    y: 95 + y * 74,
  }
}

export default GameCanvas
