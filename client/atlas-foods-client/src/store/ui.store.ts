import { create } from 'zustand'

export type ActiveView = 'map' | 'build' | 'warehouse' | 'market' | 'contracts' | 'research' | 'executives' | 'finance' | 'leaderboard'

interface UIState {
  activeView: ActiveView
  selectedBuildingId: string | null
  sidebarOpen: boolean
  marketPanelOpen: boolean
  chatOpen: boolean
  powerupOpen: boolean

  setActiveView: (view: ActiveView) => void
  selectBuilding: (id: string | null) => void
  setSidebarOpen: (open: boolean) => void
  setMarketPanelOpen: (open: boolean) => void
  setChatOpen: (open: boolean) => void
  setPowerupOpen: (open: boolean) => void
}

export const useUIStore = create<UIState>((set) => ({
  activeView: 'map',
  selectedBuildingId: null,
  sidebarOpen: true,
  marketPanelOpen: false,
  chatOpen: false,
  powerupOpen: false,

  setActiveView: (view) => set({ activeView: view, selectedBuildingId: null }),
  selectBuilding: (id) => set({ selectedBuildingId: id }),
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
  setMarketPanelOpen: (open) => set({ marketPanelOpen: open }),
  setChatOpen: (open) => set({ chatOpen: open }),
  setPowerupOpen: (open) => set({ powerupOpen: open }),
}))
