import { create } from 'zustand'
import { MAPS, type MapId } from '@/game/map.config'

export type ActiveView = 'map' | 'build' | 'warehouse' | 'chain' | 'production' | 'market' | 'contracts' | 'research' | 'executives' | 'finance' | 'leaderboard' | 'settings' | 'inspect' | 'chat'

interface UIState {
  activeView: ActiveView
  selectedBuildingId: string | null
  placementBuildingId: string | null
  movingBuildingId: string | null
  sidebarOpen: boolean
  marketPanelOpen: boolean
  chatOpen: boolean
  powerupOpen: boolean
  currentMapId: MapId
  syncingRetail: boolean
  chatTab: 'messages' | 'public'

  setActiveView: (view: ActiveView) => void
  selectBuilding: (id: string | null) => void
  startBuildingPlacement: (id: string) => void
  clearBuildingPlacement: () => void
  startBuildingMove: (id: string) => void
  clearBuildingMove: () => void
  setSidebarOpen: (open: boolean) => void
  setMarketPanelOpen: (open: boolean) => void
  setChatOpen: (open: boolean) => void
  setPowerupOpen: (open: boolean) => void
  setChatTab: (tab: 'messages' | 'public') => void
  setCurrentMapId: (mapId: MapId) => void
  setSyncingRetail: (syncing: boolean) => void
}

function initialMapId(): MapId {
  const stored = localStorage.getItem('newhaven_current_map')
  if (stored && MAPS[stored]) return stored
  return 'harbor'
}

export const useUIStore = create<UIState>((set) => ({
  activeView: 'map',
  selectedBuildingId: null,
  placementBuildingId: null,
  movingBuildingId: null,
  sidebarOpen: true,
  marketPanelOpen: false,
  chatOpen: false,
  chatTab: 'messages',
  powerupOpen: false,
  syncingRetail: false,
  currentMapId: initialMapId(),

  setActiveView: (view) => set({ activeView: view, selectedBuildingId: null }),
  selectBuilding: (id) => set({ selectedBuildingId: id, placementBuildingId: null, movingBuildingId: null }),
  startBuildingPlacement: (id) => set({ activeView: 'map', selectedBuildingId: null, placementBuildingId: id, movingBuildingId: null }),
  clearBuildingPlacement: () => set({ placementBuildingId: null }),
  startBuildingMove: (id) => set({ activeView: 'map', selectedBuildingId: null, movingBuildingId: id, placementBuildingId: null }),
  clearBuildingMove: () => set({ movingBuildingId: null }),
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
  setMarketPanelOpen: (open) => set({ marketPanelOpen: open }),
  setChatOpen: (open) => set({ chatOpen: open }),
  setChatTab: (tab) => set({ chatTab: tab }),
  setPowerupOpen: (open) => set({ powerupOpen: open }),
  setCurrentMapId: (mapId) => {
    localStorage.setItem('newhaven_current_map', mapId)
    set({ currentMapId: mapId, selectedBuildingId: null })
  },
  setSyncingRetail: (syncing) => set({ syncingRetail: syncing }),
}))
