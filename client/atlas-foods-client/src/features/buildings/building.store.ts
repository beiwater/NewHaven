import { create } from 'zustand'

interface BuildingStore {
  /** Production queue panel open on right */
  queuePanelOpen: boolean
  setQueuePanelOpen: (open: boolean) => void
}

export const useBuildingStore = create<BuildingStore>((set) => ({
  queuePanelOpen: false,
  setQueuePanelOpen: (open) => set({ queuePanelOpen: open }),
}))
