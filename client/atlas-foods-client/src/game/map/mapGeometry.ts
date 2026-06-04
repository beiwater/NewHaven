import { Container, Graphics, Sprite, Text, Texture } from 'pixi.js'
import { MAPS, isMapUnlocked, findSlot, fallbackSlot, allRealSlots, type MapSlot } from '@/game/map.config'
import type { Building, ProductionJob } from '@/game/types'

/** Progress of a production job 0..1 */
export function jobProgress(job: ProductionJob, now: number): number {
  const start = new Date(job.startedAt).getTime()
  const end = new Date(job.completesAt).getTime()
  if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) return 1
  return Math.max(0, Math.min(1, (now - start) / (end - start)))
}

/** Amount produced so far (floor) */
export function accruedAmount(job: ProductionJob, now: number): number {
  if (job.status === 'claimed') return job.amount
  return Math.floor(jobProgress(job, now) * job.amount)
}

/** Amount collectable right now */
export function collectableAmount(job: ProductionJob, now: number): number {
  if (job.status === 'claimed') return 0
  return Math.max(job.claimableAmount ?? 0, accruedAmount(job, now) - (job.claimedAmount ?? 0), 0)
}

/** Normalise a raw mapId string to a valid MapId based on unlock state */
export function normalizedMapId(mapId: string | undefined, level: number): string {
  if (mapId && isMapUnlocked(mapId, level)) {
    return mapId
  }
  return MAPS['harbor']?.id ?? 'harbor'
}

/** Normalise a building's mapId */
export function buildingMapId(building: { mapId?: string }): string {
  return building.mapId && MAPS[building.mapId] ? building.mapId : 'harbor'
}

/** Resolve the MapSlot for a building, with fallback */
export function buildingSlot(building: Building, level: number, fallbackIndex = 0): MapSlot {
  const slot = findSlot(buildingMapId(building), building.slotId)
  if (slot) return slot
  const fallbackSlots = allRealSlots(buildingMapId(building))
  const fallback = fallbackSlots[fallbackIndex] ?? fallbackSlots[0]
  if (fallback) return fallback
  return fallbackSlot(buildingMapId(building), level)
}

/** Convert a MapSlot to an image-pixel position (sprite offset applied) */
export function slotPosition(slot: MapSlot): { x: number; y: number } {
  return { x: slot.px - 50, y: slot.py - 80 }
}

/** Animated "+N" float that drifts upward and fades out */
export function addHarvestFloat(layer: Container, pos: { x: number; y: number }, texture: Texture | undefined, amount: number) {
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
