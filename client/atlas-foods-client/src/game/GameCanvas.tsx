import { useEffect, useMemo, useRef } from 'react'
import { IMAGE_LOAD_TIMEOUT_MS } from '@/constants'

import { Texture, Sprite, Container, type Application } from 'pixi.js'
import { useBuildings } from '@/api/buildings.api'
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
  const selectedBuildingId = useUIStore((s) => s.selectedBuildingId)
  const selectBuilding = useUIStore((s) => s.selectBuilding)
  const buildings = useMemo(
    () => Array.isArray(buildingsData)
      ? buildingsData.filter((building) => building.placed !== false)
      : [],
    [buildingsData],
  )

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

        // Warm earth-tone background as safety net for map edges
        app.renderer.background.color = 0xe8dcc8

        const sprite = new Sprite(Texture.from(img))
        sprite.anchor.set(0, 0)
        const screenW = app.screen.width
        const screenH = app.screen.height
        const imgW = sprite.width
        const imgH = sprite.height
        // Scale sprite so it fills viewport at 1x
        const baseScale = Math.max(screenW / imgW, screenH / imgH)
        sprite.scale.set(baseScale)
        app.stage.addChild(sprite)

        // Map world bounds
        const mapW = imgW * baseScale
        const mapH = imgH * baseScale

        // ── Zoom state ──
        let zoom = 1.0 // stage scale; 1x = map fills viewport
        const MIN_ZOOM = 1.0 // no zoom-out past initial fill
        const MAX_ZOOM = 4.0

        const clampPivot = () => {
          const maxPx = Math.max(0, mapW - screenW / zoom)
          const maxPy = Math.max(0, mapH - screenH / zoom)
          app!.stage.pivot.x = Math.max(0, Math.min(maxPx, app!.stage.pivot.x))
          app!.stage.pivot.y = Math.max(0, Math.min(maxPy, app!.stage.pivot.y))
        }


        const buildingLayer = new Container()
        buildingLayer.sortableChildren = true
        buildingLayerRef.current = buildingLayer
        app.stage.addChild(buildingLayer)
        const canvas = app.canvas as HTMLCanvasElement

        // ── Wheel zoom (geometric, clamped) ──
        canvas.addEventListener('wheel', (e: WheelEvent) => {
          e.preventDefault()
          const factor = e.deltaY > 0 ? 0.92 : 1.08
          const next = Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, zoom * factor))
          if (next !== zoom) {
            zoom = next
            app!.stage.scale.set(zoom)
            clampPivot()
          }
        }, { passive: false })

        // ── Middle-mouse pan (clamped) ──
        let panning = false
        const ps = { x: 0, y: 0 }
        const pp = { x: 0, y: 0 }
        canvas.addEventListener('mousedown', (e: MouseEvent) => {
          if (e.button === 1) { panning = true; ps.x = e.clientX; ps.y = e.clientY; pp.x = app!.stage.pivot.x; pp.y = app!.stage.pivot.y }
        })
        canvas.addEventListener('mousemove', (e: MouseEvent) => {
          if (panning) {
            app!.stage.pivot.x = pp.x - (e.clientX - ps.x) / app!.stage.scale.x
            app!.stage.pivot.y = pp.y - (e.clientY - ps.y) / app!.stage.scale.y
            clampPivot()
          }
        })
        canvas.addEventListener('mouseup', () => { panning = false })
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
  }, [])

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

  return <div ref={containerRef} style={{ width: '100%', height: '100%', background: '#e8dcc8' }} />
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
