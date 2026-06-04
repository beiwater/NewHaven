// ── Map config: single source of truth for all map definitions ──
// To add a new map, add an entry to the MAPS record below.
// Each map defines: id, name, background, unlockLevel, slots, hotspots.
// Frontend automatically shows the map entry, lock state, slots, and
// hotspot navigation; no GameCanvas changes required for a new map.

export type MapId = string

export interface MapSlot {
  id: string
  mapId: MapId
  px: number
  py: number
  unlockOrder: number
  occupiedByArt?: boolean
}

export interface MapHotspot {
  id: string
  targetMapId: MapId
  px: number
  py: number
  label: string
  unlockLevel?: number
}

export interface MapDefinition {
  id: MapId
  name: string
  background: string
  /** Level required to unlock this map (1 = immediately available) */
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
  harbor: {
    id: 'harbor',
    name: '新港码头',
    background: '/assets/backgrounds/map_harbor_v1.png',
    unlockLevel: 1,
    slots: [
      { id: 'harbor-plot-01', mapId: 'harbor', px: 116, py: 122, unlockOrder: 1 },
      { id: 'harbor-plot-02', mapId: 'harbor', px: 274, py: 142, unlockOrder: 2 },
      { id: 'harbor-plot-03', mapId: 'harbor', px: 194, py: 330, unlockOrder: 3 },
      { id: 'harbor-plot-04', mapId: 'harbor', px: 356, py: 378, unlockOrder: 4 },
      { id: 'harbor-plot-05', mapId: 'harbor', px: 183, py: 566, unlockOrder: 5 },
      { id: 'harbor-plot-06', mapId: 'harbor', px: 380, py: 610, unlockOrder: 6 },
      { id: 'harbor-plot-07', mapId: 'harbor', px: 317, py: 755, unlockOrder: 7 },
      { id: 'harbor-plot-08', mapId: 'harbor', px: 525, py: 780, unlockOrder: 8 },
      { id: 'harbor-art-dock', mapId: 'harbor', px: 820, py: 405, unlockOrder: 99, occupiedByArt: true },
      { id: 'harbor-art-warehouse', mapId: 'harbor', px: 448, py: 590, unlockOrder: 99, occupiedByArt: true },
    ],
    hotspots: [
      {
        id: 'harbor-to-inland',
        targetMapId: 'inland',
        px: 80,
        py: 470,
        label: '内陆庄园',
        unlockLevel: 5,
      },
      {
        id: 'harbor-to-desert',
        targetMapId: 'desert',
        px: 1540,
        py: 220,
        label: '沙漠绿洲',
        unlockLevel: 10,
      },
    ],
  },
  inland: {
    id: 'inland',
    name: '内陆庄园',
    background: '/assets/backgrounds/map_background_v1.png',
    unlockLevel: 5,
    slots: [
      { id: 'inland-plot-01', mapId: 'inland', px: 226, py: 228, unlockOrder: 1 },
      { id: 'inland-plot-02', mapId: 'inland', px: 474, py: 230, unlockOrder: 2 },
      { id: 'inland-plot-03', mapId: 'inland', px: 765, py: 170, unlockOrder: 3 },
      { id: 'inland-plot-04', mapId: 'inland', px: 1124, py: 252, unlockOrder: 4 },
      { id: 'inland-plot-05', mapId: 'inland', px: 325, py: 502, unlockOrder: 5 },
      { id: 'inland-plot-06', mapId: 'inland', px: 702, py: 470, unlockOrder: 6 },
      { id: 'inland-plot-07', mapId: 'inland', px: 1056, py: 520, unlockOrder: 7 },
      { id: 'inland-plot-08', mapId: 'inland', px: 488, py: 770, unlockOrder: 8 },
      { id: 'inland-art-town', mapId: 'inland', px: 1450, py: 382, unlockOrder: 99, occupiedByArt: true },
      { id: 'inland-art-field', mapId: 'inland', px: 245, py: 322, unlockOrder: 99, occupiedByArt: true },
    ],
    hotspots: [
      {
        id: 'inland-to-harbor',
        targetMapId: 'harbor',
        px: 1540,
        py: 455,
        label: '新港码头',
      },
    ],
  },
  desert: {
    id: 'desert',
    name: '沙漠绿洲',
    background: '/assets/backgrounds/map_harbor_v1.png',
    unlockLevel: 10,
    slots: [
      { id: 'desert-plot-01', mapId: 'desert', px: 150, py: 150, unlockOrder: 1 },
      { id: 'desert-plot-02', mapId: 'desert', px: 350, py: 180, unlockOrder: 2 },
      { id: 'desert-plot-03', mapId: 'desert', px: 220, py: 360, unlockOrder: 3 },
    ],
    hotspots: [
      {
        id: 'desert-to-harbor',
        targetMapId: 'harbor',
        px: 70,
        py: 455,
        label: '新港码头',
      },
    ],
  },
}

// ── Helper functions ─────────────────────────────────────────────

/** Check whether a map is unlocked for the given player level */
export function isMapUnlocked(mapId: MapId, level: number): boolean {
  const map = MAPS[mapId]
  if (!map) return false
  return level >= map.unlockLevel
}

/** Check whether a single slot is placeable at the given level */
export function isSlotUnlocked(slot: MapSlot, level: number): boolean {
  if (slot.occupiedByArt) return false
  if (!isMapUnlocked(slot.mapId, level)) return false
  // First 3 slots: available at map's unlock level
  if (slot.unlockOrder <= INITIAL_OPEN_SLOTS) return true
  // Additional slots: +2 levels per order past 3
  const map = MAPS[slot.mapId]
  const baseLevel = map ? map.unlockLevel : 1
  return level >= baseLevel + (slot.unlockOrder - INITIAL_OPEN_SLOTS) * 2
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
  return slots.reduce((best, slot) => {
    const bestDist = (best.px - imgX) ** 2 + (best.py - imgY) ** 2
    const dist = (slot.px - imgX) ** 2 + (slot.py - imgY) ** 2
    return dist < bestDist ? slot : best
  })
}

/** Build the set of occupied slot ids for a list of buildings */
export function occupiedSlotSet(buildings: { mapId?: MapId; slotId?: string; x?: number; y?: number }[]): Set<string> {
  return new Set(
    buildings.map((b) => b.slotId || '').filter(Boolean),
  )
}
