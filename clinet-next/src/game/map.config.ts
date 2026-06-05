// ── Map config: single source of truth for all map definitions ──
// To add a new map, add an entry to the MAPS record below.
// Each map defines: id, name, background, unlockLevel, slots, hotspots.
// Frontend automatically shows the map entry, lock state, slots, and
// hotspot navigation; no GameCanvas changes required for a new map.

export type MapId = string

export interface MapSlot {
  id: string
  label: string
  x: number
  y: number
  unlockLevel: number
  occupiedByArt?: boolean
  artKey?: string
  width?: number
  height?: number
  /** If true, this slot is sorted to the bottom in the building shop */
  premium?: boolean
}

export interface MapHotspot {
  id: string
  label: string
  x: number
  y: number
  targetMapId?: MapId
}

export interface MapDefinition {
  id: MapId
  name: string
  background: string
  unlockLevel: number
  slots: MapSlot[]
  hotspots: MapHotspot[]
}

export const INITIAL_OPEN_SLOTS = 3

// ── Map definitions ──────────────────────────────────────────────
// To add a new map: add a new MapDefinition entry below, create the
// corresponding backend slot entries in building_shop.go, and add the
// background image.

export const MAPS: Record<MapId, MapDefinition> = {
  farm: {
    id: 'farm',
    name: 'Farm',
    background: '/assets/maps/farm.png',
    unlockLevel: 1,
    slots: [
      { id: 'a1', label: 'A1', x: 120, y: 80, unlockLevel: 1 },
      { id: 'a2', label: 'A2', x: 320, y: 80, unlockLevel: 2 },
      { id: 'a3', label: 'A3', x: 520, y: 80, unlockLevel: 4 },
      { id: 'b1', label: 'B1', x: 120, y: 260, unlockLevel: 5 },
      { id: 'b2', label: 'B2', x: 320, y: 260, unlockLevel: 6 },
      { id: 'b3', label: 'B3', x: 520, y: 260, unlockLevel: 7 },
      { id: 'c1', label: 'C1', x: 220, y: 440, unlockLevel: 8 },
      { id: 'c2', label: 'C2', x: 420, y: 440, unlockLevel: 10 },
      { id: 'art_fountain', label: 'Fountain', x: 320, y: 260, unlockLevel: 1, occupiedByArt: true, artKey: 'farm_fountain', width: 120, height: 100 },
    ],
    hotspots: [
      { id: 'farm_to_mill', label: '→ Mill', x: 640, y: 200, targetMapId: 'mill' },
    ],
  },
  mill: {
    id: 'mill',
    name: 'Mill',
    background: '/assets/maps/mill.png',
    unlockLevel: 3,
    slots: [
      { id: 'a1', label: 'A1', x: 120, y: 80, unlockLevel: 3 },
      { id: 'b1', label: 'B1', x: 120, y: 260, unlockLevel: 5 },
      { id: 'c1', label: 'C1', x: 120, y: 440, unlockLevel: 8 },
      { id: 'a2', label: 'A2', x: 320, y: 80, unlockLevel: 6 },
    ],
    hotspots: [
      { id: 'mill_to_farm', label: '← Farm', x: -20, y: 200, targetMapId: 'farm' },
      { id: 'mill_to_kitchen', label: '→ Kitchen', x: 640, y: 200, targetMapId: 'kitchen' },
    ],
  },
  kitchen: {
    id: 'kitchen',
    name: 'Kitchen',
    background: '/assets/maps/kitchen.png',
    unlockLevel: 5,
    slots: [
      { id: 'a1', label: 'A1', x: 120, y: 80, unlockLevel: 5 },
      { id: 'b1', label: 'B1', x: 120, y: 260, unlockLevel: 7 },
      { id: 'c1', label: 'C1', x: 120, y: 440, unlockLevel: 9 },
    ],
    hotspots: [
      { id: 'kitchen_to_mill', label: '← Mill', x: -20, y: 200, targetMapId: 'mill' },
    ],
  },
}

// ── Helper functions ─────────────────────────────────────────────

/** Check whether a map is unlocked for the given player level */
export function isMapUnlocked(mapId: MapId, level: number): boolean {
  return level >= (MAPS[mapId]?.unlockLevel ?? Infinity)
}

/** Check whether a single slot is placeable at the given level */
export function isSlotUnlocked(slot: MapSlot, level: number): boolean {
  return !slot.occupiedByArt && level >= slot.unlockLevel
}

/** All placeable (unlocked, non-art) slots for a map */
export function placeableSlots(mapId: MapId, level: number): MapSlot[] {
  return MAPS[mapId]?.slots.filter((slot) => isSlotUnlocked(slot, level)) ?? []
}

/** All non-art slots (regardless of unlock state) */
export function allRealSlots(mapId: MapId): MapSlot[] {
  return MAPS[mapId]?.slots.filter((slot) => !slot.occupiedByArt) ?? []
}

/** Find a slot by id */
export function findSlot(mapId: MapId, slotId?: string | null): MapSlot | undefined {
  return MAPS[mapId]?.slots.find((slot) => slot.id === slotId)
}

/** Fallback to first placeable slot, then first real slot, then first slot */
export function fallbackSlot(mapId: MapId, level = 1): MapSlot {
  return (placeableSlots(mapId, level)[0] ?? allRealSlots(mapId)[0] ?? MAPS[mapId]?.slots[0])!
}

/** Find the nearest slot to (imgX, imgY) within all real slots */
export function nearestSlot(mapId: MapId, imgX: number, imgY: number, slots = allRealSlots(mapId)): MapSlot {
  let best = slots[0]
  let bestDist = Infinity
  for (const slot of slots) {
    const dx = slot.x - imgX
    const dy = slot.y - imgY
    const dist = dx * dx + dy * dy
    if (dist < bestDist) {
      bestDist = dist
      best = slot
    }
  }
  return best!
}

/** Build the set of occupied slot ids for a list of buildings */
export function occupiedSlotSet(buildings: { mapId?: MapId; slotId?: string; x?: number; y?: number }[]): Set<string> {
  const occupied = new Set<string>()
  for (const b of buildings) {
    if (b.slotId) occupied.add(b.slotId)
  }
  return occupied
}
