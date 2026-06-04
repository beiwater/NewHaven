// Dev-only sound test panel. Renders nothing in production builds.
// Use from Settings page or open programmatically via __openSoundTest.

import { useState } from 'react'
import { audio } from './AudioManager'
import {
  ALL_SFX_KEYS,
  ALL_MUSIC_KEYS,
  ALL_AMBIENCE_KEYS,
  SFX_CATEGORIES,
  MUSIC_LABELS,
  AMBIENCE_LABELS,
  type SfxKey,
} from './audioManifest'

interface SoundTestPanelProps {
  open: boolean
  onClose: () => void
}

const DEV = import.meta.env.DEV

/** Hidden dev panel — only included in dev builds. */
export function SoundTestPanel({ open, onClose }: SoundTestPanelProps) {
  const [category, setCategory] = useState<string | 'all'>('all')

  if (!DEV) return null

  const categories = ['all', ...Object.keys(SFX_CATEGORIES)]

  const visibleKeys: SfxKey[] =
    category === 'all'
      ? ALL_SFX_KEYS
      : SFX_CATEGORIES[category] ?? ALL_SFX_KEYS

  return open ? (
    <div className="fixed inset-0 z-[9998] flex items-start justify-center pt-8 bg-black/40">
      <div className="bg-white rounded-2xl shadow-2xl border border-amber-200 max-w-2xl w-full max-h-[85vh] overflow-y-auto p-5 space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-bold text-amber-900">Sound Test Panel</h2>
          <button
            onClick={onClose}
            className="text-amber-500 hover:text-amber-700 text-xl leading-none"
          >
            ✕
          </button>
        </div>

        {/* Volume sliders */}
        <div className="bg-amber-50 rounded-xl p-3 space-y-2">
          {(
            [
              ['Master', audio.getSettings().masterVolume, (v: number) => audio.setMasterVolume(v)],
              ['SFX', audio.getSettings().sfxVolume, (v: number) => audio.setSfxVolume(v)],
              ['Music', audio.getSettings().musicVolume, (v: number) => audio.setMusicVolume(v)],
              ['Ambience', audio.getSettings().ambienceVolume, (v: number) => audio.setAmbienceVolume(v)],
            ] as const
          ).map(([label, val, setter]) => (
            <div key={label} className="flex items-center gap-3">
              <span className="text-xs font-semibold text-amber-800 w-20">{label}</span>
              <input
                type="range"
                min={0}
                max={1}
                step={0.05}
                value={val}
                onChange={(e) => setter(Number(e.target.value))}
                className="flex-1 accent-amber-600"
              />
              <span className="text-[10px] text-amber-600 w-8 text-right">
                {Math.round(val * 100)}%
              </span>
            </div>
          ))}
          <button
            onClick={() => audio.toggleMute()}
            className={`px-3 py-1 text-xs font-semibold rounded-lg ${
              audio.muted
                ? 'bg-red-100 text-red-700'
                : 'bg-green-100 text-green-700'
            }`}
          >
            {audio.muted ? 'Muted' : 'Mute'}
          </button>
        </div>

        {/* BGM */}
        <section>
          <h3 className="text-xs font-bold uppercase tracking-wider text-amber-600 mb-2">
            BGM
          </h3>
          <div className="flex flex-wrap gap-1.5">
            {ALL_MUSIC_KEYS.map((key) => (
              <button
                key={key}
                onClick={() => audio.playMusic(key)}
                className="px-2.5 py-1 bg-indigo-100 hover:bg-indigo-200 text-indigo-800 text-[10px] font-semibold rounded-lg"
              >
                {MUSIC_LABELS[key]}
              </button>
            ))}
            <button
              onClick={() => audio.stopMusic()}
              className="px-2.5 py-1 bg-red-100 hover:bg-red-200 text-red-700 text-[10px] font-semibold rounded-lg"
            >
              Stop
            </button>
          </div>
        </section>

        {/* Ambience */}
        <section>
          <h3 className="text-xs font-bold uppercase tracking-wider text-amber-600 mb-2">
            Ambience
          </h3>
          <div className="flex flex-wrap gap-1.5">
            {ALL_AMBIENCE_KEYS.map((key) => (
              <button
                key={key}
                onClick={() => audio.playAmbience(key)}
                className="px-2.5 py-1 bg-teal-100 hover:bg-teal-200 text-teal-800 text-[10px] font-semibold rounded-lg"
              >
                {AMBIENCE_LABELS[key]}
              </button>
            ))}
            <button
              onClick={() => audio.stopAmbience()}
              className="px-2.5 py-1 bg-red-100 hover:bg-red-200 text-red-700 text-[10px] font-semibold rounded-lg"
            >
              Stop
            </button>
          </div>
        </section>

        {/* SFX Category filter */}
        <section>
          <h3 className="text-xs font-bold uppercase tracking-wider text-amber-600 mb-2">
            SFX
          </h3>
          <div className="flex flex-wrap gap-1 mb-3">
            {categories.map((cat) => (
              <button
                key={cat}
                onClick={() => setCategory(cat)}
                className={`px-2 py-0.5 text-[10px] font-semibold rounded ${
                  category === cat
                    ? 'bg-amber-700 text-white'
                    : 'bg-amber-100 text-amber-700 hover:bg-amber-200'
                }`}
              >
                {cat === 'all' ? 'All' : cat}
              </button>
            ))}
          </div>
          <div className="grid grid-cols-4 sm:grid-cols-5 gap-1.5">
            {visibleKeys.map((key) => (
              <button
                key={key}
                onClick={() => audio.playSfx(key, { throttleMs: 0 })}
                className="px-1.5 py-1 bg-amber-100 hover:bg-amber-200 text-amber-800 text-[9px] font-medium rounded truncate"
                title={key}
              >
                {key}
              </button>
            ))}
          </div>
        </section>
      </div>
    </div>
  ) : null
}
