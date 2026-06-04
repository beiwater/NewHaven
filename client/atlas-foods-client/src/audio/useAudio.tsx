// React context + hook for the AudioManager singleton.
// Wrap your app with <AudioProvider> and call useAudio() anywhere.

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  useRef,
  type ReactNode,
} from 'react'
import { audio, type AudioSettings } from './AudioManager'
import type { SfxKey, MusicKey, AmbienceKey } from './audioManifest'

// ── Context shape ──

export interface AudioContextValue {
  /** Play a one-shot SFX. */
  playSfx: (key: SfxKey, opts?: { volume?: number; throttleMs?: number }) => void
  /** Start (or switch to) a BGM track. */
  playMusic: (key: MusicKey) => void
  /** Stop current BGM. */
  stopMusic: () => void
  /** Crossfade to a different BGM track. */
  fadeMusicTo: (key: MusicKey, durationMs: number) => void
  /** Start an ambience loop. */
  playAmbience: (key: AmbienceKey) => void
  /** Stop ambience. */
  stopAmbience: () => void
  /** Current settings snapshot. */
  settings: AudioSettings
  /** Update a volume slider (0-1). */
  setMasterVolume: (v: number) => void
  setSfxVolume: (v: number) => void
  setMusicVolume: (v: number) => void
  setAmbienceVolume: (v: number) => void
  /** Mute toggle. */
  muted: boolean
  toggleMute: () => void
  /** Unlock audio context (call on first user gesture). */
  unlockAudio: () => void
}

const AudioCtx = createContext<AudioContextValue | null>(null)

// ── Provider ──

export function AudioProvider({ children }: { children: ReactNode }) {
  // Use a counter to force re-render when settings change externally
  const [, refresh] = useState(0)
  const settingsRef = useRef(audio.getSettings())

  // Sync from AudioManager's settings on mount
  useEffect(() => {
    settingsRef.current = audio.getSettings()
    refresh((n) => n + 1)
  }, [])

  const updateAndRefresh = useCallback((fn: () => void) => {
    fn()
    settingsRef.current = audio.getSettings()
    refresh((n) => n + 1)
  }, [])

  const value: AudioContextValue = {
    playSfx: useCallback((key, opts) => audio.playSfx(key, opts), []),
    playMusic: useCallback((key) => audio.playMusic(key), []),
    stopMusic: useCallback(() => audio.stopMusic(), []),
    fadeMusicTo: useCallback((key, ms) => audio.fadeMusicTo(key, ms), []),
    playAmbience: useCallback((key) => audio.playAmbience(key), []),
    stopAmbience: useCallback(() => audio.stopAmbience(), []),
    settings: settingsRef.current,
    setMasterVolume: useCallback(
      (v) => updateAndRefresh(() => audio.setMasterVolume(v)),
      [updateAndRefresh]
    ),
    setSfxVolume: useCallback(
      (v) => updateAndRefresh(() => audio.setSfxVolume(v)),
      [updateAndRefresh]
    ),
    setMusicVolume: useCallback(
      (v) => updateAndRefresh(() => audio.setMusicVolume(v)),
      [updateAndRefresh]
    ),
    setAmbienceVolume: useCallback(
      (v) => updateAndRefresh(() => audio.setAmbienceVolume(v)),
      [updateAndRefresh]
    ),
    get muted() {
      return audio.muted
    },
    toggleMute: useCallback(() => {
      updateAndRefresh(() => audio.toggleMute())
      return audio.muted
    }, [updateAndRefresh]),
    unlockAudio: useCallback(() => audio.unlockAudio(), []),
  }

  return <AudioCtx.Provider value={value}>{children}</AudioCtx.Provider>
}

// ── Hook ──

export function useAudio(): AudioContextValue {
  const ctx = useContext(AudioCtx)
  if (!ctx) {
    throw new Error('useAudio must be used within <AudioProvider>')
  }
  return ctx
}
