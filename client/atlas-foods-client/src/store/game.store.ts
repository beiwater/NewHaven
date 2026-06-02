import { create } from 'zustand'

interface GameState {
  /** Map zoom level (0.5 - 3.0) */
  zoom: number
  /** Map pan offset */
  panX: number
  panY: number
  /** Selected building on the map */
  selectedMapBuildingId: string | null
  /** Production timer refresh tick */
  tick: number

  setZoom: (zoom: number) => void
  setPan: (x: number, y: number) => void
  selectMapBuilding: (id: string | null) => void
  tickClock: () => void
}

export const useGameStore = create<GameState>((set) => ({
  zoom: 1,
  panX: 0,
  panY: 0,
  selectedMapBuildingId: null,
  tick: 0,

  setZoom: (zoom) => set({ zoom: Math.max(0.5, Math.min(3, zoom)) }),
  setPan: (x, y) => set({ panX: x, panY: y }),
  selectMapBuilding: (id) => set({ selectedMapBuildingId: id }),
  tickClock: () => set((s) => ({ tick: s.tick + 1 })),
}))
