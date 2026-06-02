import { create } from 'zustand'

export type ActiveView = 'map' | 'build' | 'warehouse' | 'market' | 'contracts' | 'research'

interface UIState {
  activeView: ActiveView
  selectedBuildingId: string | null
  sidebarOpen: boolean
  marketPanelOpen: boolean
  chatOpen: boolean

  setActiveView: (view: ActiveView) => void
  selectBuilding: (id: string | null) => void
  setSidebarOpen: (open: boolean) => void
  setMarketPanelOpen: (open: boolean) => void
  setChatOpen: (open: boolean) => void
}

export const useUIStore = create<UIState>((set) => ({
  activeView: 'map',
  selectedBuildingId: null,
  sidebarOpen: true,
  marketPanelOpen: false,
  chatOpen: false,

  setActiveView: (view) => set({ activeView: view, selectedBuildingId: null }),
  selectBuilding: (id) => set({ selectedBuildingId: id }),
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
  setMarketPanelOpen: (open) => set({ marketPanelOpen: open }),
  setChatOpen: (open) => set({ chatOpen: open }),
}))
