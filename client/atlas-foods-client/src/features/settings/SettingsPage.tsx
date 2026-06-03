import { useState, useEffect, useCallback, useRef } from 'react'
import { useCompany } from '@/api/company.api'
import { useSavePreferences } from '@/api/company.api'

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
  const [settings, setSettings] = useState<GameSettings>(loadSettings)
  const [saved, setSaved] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const { data: companyData } = useCompany()
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
  const displayAvatar = serverAvatar
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
            <p className="text-[10px] font-bold uppercase tracking-[0.24em] text-amber-700/70">Preferences</p>
            <h2 className="text-xl font-black text-amber-950">Settings</h2>
          </div>
        </div>

        {/* Avatar */}
        <section className="space-y-2">
          <h3 className="text-[11px] font-bold uppercase tracking-wider text-amber-700 px-1">Avatar</h3>
          <div className="bg-white/60 rounded-xl border border-amber-200/40 p-4">
            <div className="flex items-center gap-4">
              {/* Preview */}
              <div
                className="w-16 h-16 rounded-full flex items-center justify-center text-white text-xl font-black shrink-0 overflow-hidden border-2 border-amber-300/50"
                style={{ background: displayAvatar ? undefined : displayBg }}
              >
                {displayAvatar ? (
                  <img src={displayAvatar} alt="Avatar" className="w-full h-full object-cover" />
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
                    {savePrefs.isPending ? 'Saving...' : 'Upload Photo'}
                  </button>
                  {displayAvatar && (
                    <button
                      onClick={handleRemoveAvatar}
                      disabled={savePrefs.isPending}
                      className="px-3 py-1.5 bg-red-50 hover:bg-red-100 disabled:opacity-50 text-red-600 text-[10px] font-semibold rounded-lg transition-colors"
                    >
                      Remove
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
                  <div className="text-[9px] text-amber-500 mb-1">Background color</div>
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
                    {savePrefs.error instanceof Error ? savePrefs.error.message : 'Failed to save'}
                  </div>
                )}
              </div>
            </div>
          </div>
        </section>

        {/* Gameplay */}
        <section className="space-y-2">
          <h3 className="text-[11px] font-bold uppercase tracking-wider text-amber-700 px-1">Gameplay</h3>
          <ToggleRow
            label="Sound Effects"
            description="Play sounds for building actions and notifications"
            checked={settings.soundEnabled}
            onChange={(v) => update({ soundEnabled: v })}
          />
          <ToggleRow
            label="Auto-Collect"
            description="Automatically collect completed production when ready"
            checked={settings.autoCollect}
            onChange={(v) => update({ autoCollect: v })}
          />
        </section>

        {/* Notifications */}
        <section className="space-y-2">
          <h3 className="text-[11px] font-bold uppercase tracking-wider text-amber-700 px-1">Notifications</h3>
          <ToggleRow
            label="Push Notifications"
            description="Get notified when production completes or contracts are ready"
            checked={settings.notificationsEnabled}
            onChange={(v) => update({ notificationsEnabled: v })}
          />
        </section>

        {/* Help */}
        <section className="space-y-2">
          <h3 className="text-[11px] font-bold uppercase tracking-wider text-amber-700 px-1">Help</h3>
          <ToggleRow
            label="Tutorial Tips"
            description="Show in-game tips and guidance for new features"
            checked={settings.showTutorialTips}
            onChange={(v) => update({ showTutorialTips: v })}
          />
        </section>

        {/* About */}
        <section className="rounded-xl border border-amber-200/40 bg-white/60 p-4 space-y-1">
          <div className="text-[10px] text-amber-500">Atlas Foods — Farm & Factory Tycoon</div>
          <div className="text-[9px] text-amber-400">Version 1.0.0</div>
        </section>

        {/* Save indicator */}
        <div className="text-center">
          {saved ? (
            <span className="text-[10px] text-green-600 font-semibold">Settings saved ✓</span>
          ) : (
            <span className="text-[9px] text-amber-400">Changes are saved automatically</span>
          )}
        </div>
      </div>
    </div>
  )
}
