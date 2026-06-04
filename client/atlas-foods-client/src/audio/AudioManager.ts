// AudioManager — singleton that wraps the Web Audio API for game-wide sound.
//
// Usage:
//   import { audio } from './AudioManager'
//   audio.init()
//   audio.playSfx('ui_button_click')
//   audio.playMusic('bgm_harbor_town')
//
// Asset files are loaded lazily on first play. Missing files → warning, no crash.
// Short SFX are throttled (per-key) to prevent ear-blast on rapid clicks.
// Settings are persisted to localStorage.

import {
  type SfxKey,
  type MusicKey,
  type AmbienceKey,
  SFX_PATHS,
  MUSIC_PATHS,
  AMBIENCE_PATHS,
} from './audioManifest'

// ── Settings ──

export interface AudioSettings {
  masterVolume: number
  sfxVolume: number
  musicVolume: number
  ambienceVolume: number
  muted: boolean
}

const SETTINGS_KEY = 'newhaven_audio_settings'

const DEFAULT_SETTINGS: AudioSettings = {
  masterVolume: 0.8,
  sfxVolume: 0.75,
  musicVolume: 0.35,
  ambienceVolume: 0.25,
  muted: false,
}

function loadSettings(): AudioSettings {
  try {
    const raw = localStorage.getItem(SETTINGS_KEY)
    if (raw) return { ...DEFAULT_SETTINGS, ...JSON.parse(raw) }
  } catch { /* ignore corrupt data */ }
  return { ...DEFAULT_SETTINGS }
}

function persistSettings(s: AudioSettings): void {
  try {
    localStorage.setItem(SETTINGS_KEY, JSON.stringify(s))
  } catch { /* quota exceeded — ignore */ }
}

// ── Helpers ──

/** Throttle: ignore repeated plays within N ms. */
const SFX_THROTTLE_MS = 60

/** Minor pitch variation range (± semitones) for variety on repeated SFX. */
const PITCH_VARIATION = 0.06
const VOLUME_VARIATION = 0.08

// ── Manager ──

class AudioManager {
  // Audio graph
  private ctx: AudioContext | null = null
  private masterGain: GainNode | null = null
  private sfxGain: GainNode | null = null
  private musicGain: GainNode | null = null
  private ambienceGain: GainNode | null = null

  // Buffers (decoded once, reused)
  private buffers = new Map<string, AudioBuffer>()
  private loading = new Map<string, Promise<void>>()

  // Currently playing sources — one active BGM + one active ambience
  private musicSource: AudioBufferSourceNode | null = null
  private musicKey: MusicKey | null = null
  private ambienceSource: AudioBufferSourceNode | null = null
  private ambienceKey: AmbienceKey | null = null

  // Throttle timestamps
  private lastPlayed = new Map<string, number>()

  // Settings
  private settings: AudioSettings = loadSettings()

  // Ready flag; true after init()
  private _ready = false

  // ── Lifecycle ──

  init(): void {
    if (this._ready) return
    // AudioContext created on first user gesture (see unlockAudio)
    this._ready = true
  }

  /** Call after the first user interaction (click/touch/keydown) to comply with autoplay policy. */
  unlockAudio(): void {
    if (this.ctx?.state === 'running') return
    if (!this.ctx) {
      this.ctx = new AudioContext()
      this.buildGraph()
    }
    if (this.ctx.state === 'suspended') {
      this.ctx.resume().catch(() => {})
    }
  }

  private ensureContext(): boolean {
    if (!this.ctx) return false
    return this.ctx.state !== 'closed'
  }

  private buildGraph(): void {
    if (!this.ctx) return
    this.masterGain = this.ctx.createGain()
    this.sfxGain = this.ctx.createGain()
    this.musicGain = this.ctx.createGain()
    this.ambienceGain = this.ctx.createGain()

    this.sfxGain.connect(this.masterGain)
    this.musicGain.connect(this.masterGain)
    this.ambienceGain.connect(this.masterGain)
    this.masterGain.connect(this.ctx.destination)

    this.applyVolumes()
  }

  // ── Volume ──

  getSettings(): AudioSettings {
    return { ...this.settings }
  }

  setMasterVolume(v: number): void {
    this.settings.masterVolume = clamp01(v)
    this.applyVolumes()
    persistSettings(this.settings)
  }

  setSfxVolume(v: number): void {
    this.settings.sfxVolume = clamp01(v)
    this.applyVolumes()
    persistSettings(this.settings)
  }

  setMusicVolume(v: number): void {
    this.settings.musicVolume = clamp01(v)
    this.applyVolumes()
    persistSettings(this.settings)
  }

  setAmbienceVolume(v: number): void {
    this.settings.ambienceVolume = clamp01(v)
    this.applyVolumes()
    persistSettings(this.settings)
  }

  mute(): void {
    this.settings.muted = true
    this.applyVolumes()
    persistSettings(this.settings)
  }

  unmute(): void {
    this.settings.muted = false
    this.applyVolumes()
    persistSettings(this.settings)
  }

  toggleMute(): boolean {
    if (this.settings.muted) this.unmute()
    else this.mute()
    return this.settings.muted
  }

  get muted(): boolean {
    return this.settings.muted
  }

  private applyVolumes(): void {
    const s = this.settings
    const master = s.muted ? 0 : s.masterVolume
    if (this.masterGain) this.masterGain.gain.value = master
    if (this.sfxGain) this.sfxGain.gain.value = s.sfxVolume
    if (this.musicGain) this.musicGain.gain.value = s.musicVolume
    if (this.ambienceGain) this.ambienceGain.gain.value = s.ambienceVolume
  }

  // ── Loading ──

  private async loadBuffer(key: string, url: string): Promise<AudioBuffer | null> {
    // Deduplicate concurrent loads
    const pending = this.loading.get(key)
    if (pending) {
      await pending
      return this.buffers.get(key) ?? null
    }

    const { promise, resolve } = Promise.withResolvers<void>()
    this.loading.set(key, promise)

    try {
      const resp = await fetch(url)
      if (!resp.ok) {
        if (import.meta.env.DEV) {
          console.warn(`[Audio] File not found: ${url} (${resp.status})`)
        }
        resolve()
        return null
      }
      const arrayBuf = await resp.arrayBuffer()
      if (!this.ctx) {
        resolve()
        return null
      }
      const audioBuf = await this.ctx.decodeAudioData(arrayBuf)
      this.buffers.set(key, audioBuf)
      resolve()
      return audioBuf
    } catch (err) {
      if (import.meta.env.DEV) {
        console.warn(`[Audio] Failed to load ${url}:`, err)
      }
      resolve()
      return null
    }
  }

  /** Preload a list of SFX keys (call early for sounds you know you'll need). */
  async preloadSfx(keys: SfxKey[]): Promise<void> {
    await Promise.allSettled(
      keys.map((k) => this.loadBuffer(k, SFX_PATHS[k]))
    )
  }

  // ── Play SFX ──

  playSfx(
    key: SfxKey,
    opts?: { volume?: number; throttleMs?: number }
  ): void {
    if (!this.ensureContext()) return
    if (this.settings.muted) return

    // Throttle
    const throttle = opts?.throttleMs ?? SFX_THROTTLE_MS
    const now = performance.now()
    const last = this.lastPlayed.get(key) ?? 0
    if (now - last < throttle) return
    this.lastPlayed.set(key, now)

    this.playBuffer(SFX_PATHS[key], key, {
      volume: opts?.volume,
      loop: false,
      destination: this.sfxGain,
      vary: true,
    })
  }

  // ── Play Music ──

  playMusic(key: MusicKey): void {
    if (!this.ensureContext()) return
    if (key === this.musicKey && this.musicSource) return // already playing
    this.fadeOutAndStart(MUSIC_PATHS[key], key, {
      destination: this.musicGain,
      loop: true,
      crossfadeMs: 800,
      onStart: (k) => { this.musicKey = k as MusicKey },
      onStop: () => { this.musicKey = null },
    })
    // Update the active reference
    const oldSource = this.musicSource
    if (oldSource) {
      try { oldSource.stop() } catch { /* already stopped */ }
    }
  }

  stopMusic(): void {
    if (this.musicSource) {
      try { this.musicSource.stop() } catch { /* nop */ }
      this.musicSource = null
    }
    this.musicKey = null
  }

  /** Crossfade from current BGM to a new one over `durationMs`. */
  fadeMusicTo(key: MusicKey, durationMs: number): void {
    // Simple: fade out current, then play new
    const current = this.musicSource
    if (current && this.ctx) {
      const now = this.ctx.currentTime
      const gainNode = this.musicGain
      if (gainNode) {
        gainNode.gain.cancelScheduledValues(now)
        gainNode.gain.setValueAtTime(gainNode.gain.value, now)
        gainNode.gain.linearRampToValueAtTime(0, now + durationMs / 1000)
      }
      setTimeout(() => {
        this.stopMusic()
        // Restore music gain to its set volume
        if (gainNode) {
          gainNode.gain.cancelScheduledValues(this.ctx!.currentTime)
          gainNode.gain.setValueAtTime(0, this.ctx!.currentTime)
          gainNode.gain.linearRampToValueAtTime(
            this.settings.musicVolume,
            this.ctx!.currentTime + durationMs / 2000
          )
        }
        this.playMusic(key)
      }, durationMs)
    } else {
      this.playMusic(key)
    }
  }

  // ── Play Ambience ──

  playAmbience(key: AmbienceKey): void {
    if (!this.ensureContext()) return
    if (key === this.ambienceKey && this.ambienceSource) return

    const oldSource = this.ambienceSource
    if (oldSource) {
      try { oldSource.stop() } catch { /* nop */ }
    }

    this.playBuffer(AMBIENCE_PATHS[key], `amb:${key}`, {
      destination: this.ambienceGain,
      loop: true,
    })
    this.ambienceKey = key
  }

  stopAmbience(): void {
    if (this.ambienceSource) {
      try { this.ambienceSource.stop() } catch { /* nop */ }
      this.ambienceSource = null
    }
    this.ambienceKey = null
  }

  // ── Internal ──

  private async playBuffer(
    url: string,
    cacheKey: string,
    opts: {
      volume?: number
      loop?: boolean
      destination?: AudioNode | null
      vary?: boolean
      onStart?: (key: string) => void
      onStop?: () => void
    }
  ): Promise<void> {
    if (!this.ctx) return

    let buf = this.buffers.get(cacheKey)
    if (!buf) {
      const loaded = await this.loadBuffer(cacheKey, url)
      if (!loaded) return
      buf = loaded
    }

    const source = this.ctx.createBufferSource()
    source.buffer = buf
    source.loop = opts.loop ?? false

    // Pitch variation
    if (opts.vary) {
      const semitones = (Math.random() - 0.5) * 2 * PITCH_VARIATION
      source.detune.value = semitones * 100 // cents
    }

    // Volume variation
    let volume = opts.volume ?? 1
    if (opts.vary) {
      volume *= 1 + (Math.random() - 0.5) * 2 * VOLUME_VARIATION
    }

    // Connect
    const gainNode = this.ctx.createGain()
    gainNode.gain.value = clamp01(volume)
    source.connect(gainNode)
    gainNode.connect(opts.destination ?? this.masterGain ?? this.ctx.destination)

    source.onended = () => {
      gainNode.disconnect()
      opts.onStop?.()
    }

    source.start()

    // For BGM, store the source so we can stop it later
    const urlLower = url.toLowerCase()
    if (urlLower.includes('/bgm/')) {
      this.musicSource = source
      opts.onStart?.(cacheKey)
    } else if (urlLower.includes('/ambience/')) {
      this.ambienceSource = source
    }
  }

  /** Helper that fades out the current BGM source before starting a new one. */
  private fadeOutAndStart(
    url: string,
    cacheKey: string,
    opts: {
      destination?: AudioNode | null
      loop?: boolean
      crossfadeMs: number
      onStart?: (key: string) => void
      onStop?: () => void
    }
  ): void {
    const oldSrc = this.musicSource
    if (oldSrc && this.ctx) {
      // Create a temporary gain to fade
      try {
        oldSrc.stop()
      } catch { /* nop */ }
    }

    this.playBuffer(url, cacheKey, {
      destination: opts.destination,
      loop: opts.loop,
      onStart: opts.onStart,
      onStop: opts.onStop,
    })
  }

  /** Cleanup (call on unmount). */
  dispose(): void {
    this.stopMusic()
    this.stopAmbience()
    if (this.ctx) {
      this.ctx.close().catch(() => {})
      this.ctx = null
    }
    this.buffers.clear()
    this.loading.clear()
    this.lastPlayed.clear()
    this._ready = false
  }
}

function clamp01(v: number): number {
  return v < 0 ? 0 : v > 1 ? 1 : v
}

// Singleton
export const audio = new AudioManager()
export type { AudioManager as AudioManagerType }
