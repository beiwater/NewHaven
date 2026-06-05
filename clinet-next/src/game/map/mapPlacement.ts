// ── Map-aware placement logic ──
// Provides helpers for finding open slots across maps, so the build
// UI and auto-place can work with any number of maps.

import { allRealSlots, isMapUnlocked, MAPS, placeableSlots, type MapId } from '@/game/map.config'
import type { Building } from '@/game/types'

export interface MapSlotStatus {
  mapId: MapId
  config: (typeof MAPS)[MapId]
  unlocked: boolean
  /** Number of placeable (unlocked, non-art) slots */
  totalSlots: number
  /** Number of occupied slots on this map */
  usedSlots: number
  availableSlots: number
  /** First available slot id, if any */
  firstOpenSlotId: string | null
}

/** Build a set of slot IDs occupied across ALL maps */
function globalOccupiedSet(buildings: Building[]): Set<string> {
  const set = new Set<string>()
  for (const b of buildings) {
    if (b.slotId) set.add(b.slotId)
  }
  return set
}

/** Get placement status for every known map */
export function availableMaps(buildings: Building[], level: number): MapSlotStatus[] {
  const occupied = globalOccupiedSet(buildings)
  return Object.values(MAPS).map((cfg) => {
    const unlocked = isMapUnlocked(cfg.id, level)
    const allPlots = allRealSlots(cfg.id)
    const usable = unlocked ? placeableSlots(cfg.id, level) : allPlots
    const used = allPlots.filter((slot) => occupied.has(slot.id)).length
    const avail = unlocked ? usable.filter((slot) => !occupied.has(slot.id)).length : 0
    const firstAvail = unlocked ? (usable.find((slot) => !occupied.has(slot.id))?.id ?? null) : null
    return {
      mapId: cfg.id,
      config: cfg,
      unlocked,
      totalSlots: allPlots.length,
      usedSlots: used,
      availableSlots: avail,
      firstOpenSlotId: firstAvail,
    }
  })
}

/** Find the best map and slot for auto-placement */
export function findBestPlacement(
  buildings: Building[],
  preferredMapId: MapId,
  level: number,
): { mapId: MapId; slotId: string } | null {
  // 1. Try preferred (current) map first
  if (isMapUnlocked(preferredMapId, level)) {
    const slots = placeableSlots(preferredMapId, level)
    const occupied = globalOccupiedSet(buildings)
    for (const slot of slots) {
      if (!occupied.has(slot.id)) {
        return { mapId: preferredMapId, slotId: slot.id }
      }
    }
  }
  // 2. Try other unlocked maps
  for (const cfg of Object.values(MAPS)) {
    if (cfg.id === preferredMapId || !isMapUnlocked(cfg.id, level)) continue
    const slots = placeableSlots(cfg.id, level)
    const occupied = globalOccupiedSet(buildings)
    for (const slot of slots) {
      if (!occupied.has(slot.id)) {
        return { mapId: cfg.id, slotId: slot.id }
      }
    }
  }
  return null
}

/** Find next available slot on a specific map */
export function findNextAvailableSlot(
  buildings: Building[],
  mapId: MapId,
  level: number,
): { mapId: MapId; slotId: string } | null {
  if (!isMapUnlocked(mapId, level)) return null
  const slots = placeableSlots(mapId, level)
  const occupied = globalOccupiedSet(buildings)
  for (const slot of slots) {
    if (!occupied.has(slot.id)) {
      return { mapId, slotId: slot.id }
    }
  }
  return null
}

/** Count available (unlocked + unoccupied) slots on a map */
export function countAvailableSlots(buildings: Building[], mapId: MapId, level: number): number {
  if (!isMapUnlocked(mapId, level)) return 0
  const slots = placeableSlots(mapId, level)
  const occupied = globalOccupiedSet(buildings)
  let count = 0
  for (const slot of slots) {
    if (!occupied.has(slot.id)) count++
  }
  return count
}
