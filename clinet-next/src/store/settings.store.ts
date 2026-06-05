/**
 * User preferences persisted to localStorage.
 */
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface SettingsState {
  /** Master volume (0-1) */
  volume: number
  /** UI language */
  language: string
  /** Dark mode toggle */
  darkMode: boolean

  setVolume: (volume: number) => void
  setLanguage: (lang: string) => void
  setDarkMode: (dark: boolean) => void
}

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      volume: 0.7,
      language: 'en',
      darkMode: false,

      setVolume: (volume) => set({ volume: Math.max(0, Math.min(1, volume)) }),
      setLanguage: (language) => set({ language }),
      setDarkMode: (darkMode) => set({ darkMode }),
    }),
    { name: 'newhaven-settings' },
  ),
)
