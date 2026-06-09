import { useState, useEffect, useCallback, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { useCompany } from '@/api/company.api'
import { useSavePreferences } from '@/api/company.api'
import { clearAuth } from '@/api/client'
import { useQueryClient } from '@tanstack/react-query'
import { useAudio } from '@/audio/useAudio'
import { SoundTestPanel } from '@/audio/SoundTestPanel'
import {
  type SupportedLocale,
  SUPPORTED_LOCALES,
  LOCALE_LABELS,
  getStoredLocale,
  setStoredLocale,
} from '@/i18n'

interface GameSettings {
  soundEnabled: boolean
  notificationsEnabled: boolean
  showTutorialTips: boolean
  autoCollect: boolean
}

const DEFAULTS: GameSettings = {
  soundEnabled: true,
  notificationsEnabled: true,
  showTutorialTips: true,
  autoCollect: false,
}

const SETTINGS_KEY = 'atlas_foods_settings'

const PRESET_COLORS = [
  '#4a7c59',
  '#3b82f6',
  '#8b5cf6',
  '#dc2626',
  '#d4a843',
  '#5c3d2e',
  '#ec4899',
  '#06b6d4',
]

function loadSettings(): GameSettings {
  try {
    const raw = localStorage.getItem(SETTINGS_KEY)
    if (raw) return { ...DEFAULTS, ...JSON.parse(raw) }
  } catch { /* ignore */ }
  return DEFAULTS
}

function saveSettings(s: GameSettings) {
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(s))
}

function ToggleRow({
  label,
  description,
  checked,
  onChange,
}: {
  label: string
  description?: string
  checked: boolean
  onChange: (v: boolean) => void
}) {
  return (
    <label className="flex items-center justify-between px-4 py-3 bg-white/60 rounded-xl border border-amber-200/40 cursor-pointer hover:bg-white/80 transition-colors">
      <div>
        <div className="text-xs font-semibold text-amber-900">{label}</div>
        {description && <div className="text-[10px] text-amber-500 mt-0.5">{description}</div>}
      </div>
      <button
        role="switch"
        aria-checked={checked}
        onClick={(e) => { e.preventDefault(); onChange(!checked) }}
        className={`relative w-10 h-5.5 rounded-full transition-colors ${
          checked ? 'bg-green-500' : 'bg-amber-300'
        }`}
      >
        <span
          className={`absolute top-0.5 left-0.5 w-4.5 h-4.5 rounded-full bg-white shadow transition-transform ${
            checked ? 'translate-x-4.5' : ''
          }`}
        />
      </button>
    </label>
  )
}

export function SettingsPage() {
  const { t } = useTranslation()
  const { settings: audioSettings, setMasterVolume, setSfxVolume, setMusicVolume, muted, toggleMute } = useAudio()
  const [locale, setLocaleLocal] = useState<SupportedLocale>(getStoredLocale)
  const [showSoundTest, setShowSoundTest] = useState(false)
  const [settings, setSettings] = useState<GameSettings>(loadSettings)
  const [saved] = useState(false)
  const [avatarUrl, setAvatarUrl] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)
  const { data: companyData } = useCompany()
  const queryClient = useQueryClient()
  const handleLogout = () => {
    queryClient.clear()
    clearAuth()
  }
  const savePrefs = useSavePreferences()
  const companyName = companyData?.authCompany?.company ?? 'MC'
  const initial = companyName.charAt(0).toUpperCase()

  // Avatar: stored server-side in preferences.avatar and preferences.avatarBg
  const serverAvatar = (companyData?.preferences?.avatar as string) ?? null
  const serverAvatarBg = (companyData?.preferences?.avatarBg as string) ?? PRESET_COLORS[0]

  const update = useCallback((patch: Partial<GameSettings>) => {
    setSettings((prev) => {
      const next = { ...prev, ...patch }
      saveSettings(next)
      return next
    })
  }, [])

  useEffect(() => {
    setSettings(loadSettings())
  }, [])
  useEffect(() => {
    if (companyData?.preferences?.avatar_url) {
      setAvatarUrl(companyData.preferences.avatar_url as string)
    }
  }, [companyData])

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => {
      const dataUrl = reader.result as string
      savePrefs.mutate({ avatar: dataUrl, avatarBg: serverAvatarBg })
    }
    reader.readAsDataURL(file)
  }

  const handlePickColor = (color: string) => {
    savePrefs.mutate({ avatarBg: color, avatar: serverAvatar ?? undefined })
  }

  const handleRemoveAvatar = () => {
    savePrefs.mutate({ avatarBg: serverAvatarBg })
  }
  // Get current display values
  const displayAvatar = (companyData?.preferences?.avatar_url as string) || serverAvatar
  const displayBg = serverAvatarBg

  return (
    <div className="h-full overflow-y-auto">
      <div className="mx-auto max-w-lg p-4 space-y-5">
        {/* Header */}
        <div className="flex items-center gap-3">
          <svg className="h-6 w-6 text-amber-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <div>
            <p className="text-[10px] font-bold uppercase tracking-[0.24em] text-amber-700/70">{t('settings.preferences')}</p>
            <h2 className="text-xl font-black text-amber-950">{t('settings.title')}</h2>
          </div>
        </div>

        {/* Avatar */}
        <section className="space-y-2">
          <h3 className="text-[11px] font-bold uppercase tracking-wider text-amber-700 px-1">{t('settings.avatar')}</h3>
          <div className="bg-white/60 rounded-xl border border-amber-200/40 p-4">
            <div className="flex items-center gap-4">
              {/* Preview */}
              <div
                className="w-16 h-16 rounded-full flex items-center justify-center text-white text-xl font-black shrink-0 overflow-hidden border-2 border-amber-300/50"
                style={{ background: displayAvatar ? undefined : displayBg }}
              >
                {displayAvatar ? (
                  <img src={displayAvatar} alt={t('settings.avatar')} className="w-full h-full object-cover" />
                ) : (
                  initial
                )}
              </div>

              <div className="flex-1 space-y-2">
                <div className="flex gap-2">
                  <button
                    onClick={() => fileInputRef.current?.click()}
                    disabled={savePrefs.isPending}
                    className="px-3 py-1.5 bg-amber-100 hover:bg-amber-200 disabled:opacity-50 text-amber-800 text-[10px] font-semibold rounded-lg transition-colors"
                  >
                    {savePrefs.isPending ? t('settings.saving') : t('settings.uploadPhoto')}
                  </button>
                  {displayAvatar && (
                    <button
                      onClick={handleRemoveAvatar}
                      disabled={savePrefs.isPending}
                      className="px-3 py-1.5 bg-red-50 hover:bg-red-100 disabled:opacity-50 text-red-600 text-[10px] font-semibold rounded-lg transition-colors"
                    >
                      {t('settings.remove')}
                    </button>
                  )}
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/*"
                    onChange={handleFileChange}
                    className="hidden"
                  />
                </div>

                {/* Color picker */}
                <div>
                  <div className="text-[9px] text-amber-500 mb-1">{t('settings.bgColor')}</div>
                  <div className="flex gap-1.5">
                    {PRESET_COLORS.map((color) => (
                      <button
                        key={color}
                        onClick={() => handlePickColor(color)}
                        disabled={savePrefs.isPending}
                        className={`w-5 h-5 rounded-full border-2 transition-all disabled:opacity-50 ${
                          displayBg === color && !displayAvatar
                            ? 'border-amber-800 scale-110'
                            : 'border-transparent hover:scale-105'
                        }`}
                        style={{ backgroundColor: color }}
                      />
                    ))}
                  </div>
                </div>
                {savePrefs.isError && (
                  <div className="text-[9px] text-red-500">
                    {savePrefs.error instanceof Error ? savePrefs.error.message : t('settings.failedToSave')}
                  </div>
                )}
              </div>
            </div>
          </div>
        </section>
        {/* Avatar URL */}
        <section className="space-y-2">
          <div className="bg-white/60 rounded-xl border border-amber-200/40 p-4">
            <div className="flex gap-2">
              <input
                value={avatarUrl}
                onChange={e => setAvatarUrl(e.target.value)}
                placeholder="https://img..."
                className="flex-1 px-3 py-2 rounded-lg border border-amber-200/60 bg-white/80 text-xs text-amber-900 placeholder-amber-300"
              />
              <button
                onClick={() => savePrefs.mutate({ ...(companyData?.preferences as object || {}), avatar_url: avatarUrl })}
                disabled={savePrefs.isPending}
                className="px-3 py-2 rounded-lg bg-amber-800 text-white text-[10px] font-bold hover:bg-amber-900 disabled:opacity-50"
              >
                {savePrefs.isPending ? '保存中...' : '保存'}
              </button>
            </div>
            {savePrefs.isError && (
              <div className="text-[9px] text-red-500 mt-1">
                {savePrefs.error instanceof Error ? savePrefs.error.message : t('settings.failedToSave')}
              </div>
            )}
          </div>
        </section>

        {/* Audio */}
        <section className="space-y-2">
          <h3 className="text-[11px] font-bold uppercase tracking-wider text-amber-700 px-1">{t('settings.audio')}</h3>
          <div className="bg-white/60 rounded-xl border border-amber-200/40 p-4 space-y-3">
            {/* Master Volume */}
            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-[10px] font-semibold text-amber-800">{t('settings.masterVolume')}</span>
                <span className="text-[9px] text-amber-500">{Math.round(audioSettings.masterVolume * 100)}%</span>
              </div>
              <input
                type="range"
                min={0}
                max={1}
                step={0.05}
                value={audioSettings.masterVolume}
                onChange={(e) => setMasterVolume(Number(e.target.value))}
                className="w-full accent-amber-600"
              />
            </div>

            {/* SFX Volume */}
            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-[10px] font-semibold text-amber-800">{t('settings.sfxVolume')}</span>
                <span className="text-[9px] text-amber-500">{Math.round(audioSettings.sfxVolume * 100)}%</span>
              </div>
              <input
                type="range"
                min={0}
                max={1}
                step={0.05}
                value={audioSettings.sfxVolume}
                onChange={(e) => setSfxVolume(Number(e.target.value))}
                className="w-full accent-amber-600"
              />
            </div>

            {/* Music Volume */}
            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-[10px] font-semibold text-amber-800">{t('settings.musicVolume')}</span>
                <span className="text-[9px] text-amber-500">{Math.round(audioSettings.musicVolume * 100)}%</span>
              </div>
              <input
                type="range"
                min={0}
                max={1}
                step={0.05}
                value={audioSettings.musicVolume}
                onChange={(e) => setMusicVolume(Number(e.target.value))}
                className="w-full accent-amber-600"
              />
            </div>

            {/* Mute + Test */}
            <div className="flex items-center gap-2 pt-1">
              <button
                onClick={() => toggleMute()}
                className={`px-3 py-1.5 text-[10px] font-semibold rounded-lg transition-colors ${
                  muted
                    ? 'bg-red-100 text-red-700 hover:bg-red-200'
                    : 'bg-green-100 text-green-700 hover:bg-green-200'
                }`}
              >
                {muted ? t('settings.unmute') : t('settings.mute')}
              </button>
              <button
                onClick={() => setShowSoundTest(true)}
                className="px-3 py-1.5 bg-indigo-100 hover:bg-indigo-200 text-indigo-800 text-[10px] font-semibold rounded-lg transition-colors"
              >
                ♪ Sound Test
              </button>
            </div>
          </div>
        </section>

        {/* Notifications */}
        <section className="space-y-2">
          <h3 className="text-[11px] font-bold uppercase tracking-wider text-amber-700 px-1">{t('settings.notifications')}</h3>
          <ToggleRow
            label={t('settings.pushNotifications')}
            description={t('settings.pushNotificationsDesc')}
            checked={settings.notificationsEnabled}
            onChange={(v) => update({ notificationsEnabled: v })}
          />
        </section>

        {/* Help */}
        <section className="space-y-2">
          <h3 className="text-[11px] font-bold uppercase tracking-wider text-amber-700 px-1">{t('settings.help')}</h3>
          <ToggleRow
            label={t('settings.tutorialTips')}
            description={t('settings.tutorialTipsDesc')}
            checked={settings.showTutorialTips}
            onChange={(v) => update({ showTutorialTips: v })}
          />
        </section>

        {/* Language */}
        <section className="space-y-2">
          <h3 className="text-[11px] font-bold uppercase tracking-wider text-amber-700 px-1">{t('settings.language')}</h3>
          <div className="bg-white/60 rounded-xl border border-amber-200/40 p-3">
            <select
              value={locale}
              onChange={(e) => {
                const next = e.target.value as SupportedLocale
                setLocaleLocal(next)
                setStoredLocale(next)
              }}
              className="w-full px-3 py-2.5 bg-white border border-amber-300 rounded-lg text-sm text-amber-900 focus:outline-none focus:ring-2 focus:ring-amber-500 appearance-none cursor-pointer"
            >
              {SUPPORTED_LOCALES.map((loc) => (
                <option key={loc} value={loc}>
                  {LOCALE_LABELS[loc]}
                </option>
              ))}
            </select>
          </div>
        </section>

        {/* About */}
        <section className="rounded-xl border border-amber-200/40 bg-white/60 p-4 space-y-1">
          <div className="text-[10px] text-amber-500">{t('settings.about')}</div>
          <div className="text-[9px] text-amber-400">{t('settings.version')}</div>
        </section>
      </div>

      {/* Bottom actions */}
      <div className="border-t border-amber-200/40 pt-4 space-y-3">
        {/* GitHub link */}
        <a
          href="https://github.com"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center justify-center gap-2 px-4 py-2.5 bg-amber-50 hover:bg-amber-100 border border-amber-300/50 rounded-xl text-[11px] text-amber-700 font-semibold transition-colors"
        >
          <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
          </svg>
          GitHub
        </a>

        {/* Logout */}
        <button
          onClick={handleLogout}
          className="w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-red-50 hover:bg-red-100 border border-red-300/50 rounded-xl text-[11px] text-red-700 font-semibold transition-colors"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
          {t('settings.logout')}
        </button>
      </div>
      <SoundTestPanel open={showSoundTest} onClose={() => setShowSoundTest(false)} />
    </div>
  )
}
