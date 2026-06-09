import { useEffect, useState, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { useLogin, useRegister } from '@/api/company.api'
import { SUPPORTED_LOCALES, LOCALE_LABELS, getStoredLocale, setStoredLocale } from '@/i18n'
import { AUTH_CHANGED_EVENT, isAuthenticated } from '@/api/client'
import { audio } from '@/audio/AudioManager'
import { LoadingScreen } from './LoadingScreen'
export function AuthGate({ children }: { children: React.ReactNode }) {
  const { t } = useTranslation()
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [name, setName] = useState('')
  const [gender, setGender] = useState('')
  const [customGender, setCustomGender] = useState('')
  const [showCustomGender, setShowCustomGender] = useState(false)
  const [email, setEmail] = useState('')
  const login = useLogin()
  const [locale, setLocaleState] = useState(getStoredLocale())
  const register = useRegister()
  // Initialize from stored token so refresh doesn't show login flash.
  const [authenticated, setAuthenticated] = useState(isAuthenticated)
  const [showLoading, setShowLoading] = useState(false)

  useEffect(() => {
    const syncAuth = () => setAuthenticated(isAuthenticated())
    window.addEventListener(AUTH_CHANGED_EVENT, syncAuth)
    window.addEventListener('storage', syncAuth)
    return () => {
      window.removeEventListener(AUTH_CHANGED_EVENT, syncAuth)
      window.removeEventListener('storage', syncAuth)
    }
  }, [])

  // Play BGM on login/register screen
  useEffect(() => {
    if (!authenticated) {
      audio.init()
      audio.playMusic('bgm_main_menu')
    }
  }, [authenticated])

  if (authenticated && !showLoading) {
    return <>{children}</>
  }

  if (authenticated && showLoading) {
    return <LoadingScreen onFinished={() => setShowLoading(false)} />
  }

  const isPending = login.isPending || register.isPending
  const error = login.error || register.error

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (!username.trim() || !password.trim()) return
    setShowLoading(true)
    audio.unlockAudio()
    audio.playSfx('ui_confirm')
    if (mode === 'login') {
      login.mutate({ username: username.trim(), password })
    } else {
      const finalGender = showCustomGender && customGender.trim() ? customGender.trim() : gender
      register.mutate({
        username: username.trim(),
        password,
        name: name.trim() || undefined,
        gender: finalGender || undefined,
        email: email.trim() || undefined,
      })
    }
  }

  return (
    <div className="w-screen h-screen bg-gradient-to-br from-amber-900 via-amber-800 to-amber-950 flex items-center justify-center">
      <div className="bg-amber-50 rounded-2xl shadow-2xl p-8 w-80 border-2 border-amber-700/30">
        {/* Logo */}
        <div className="flex justify-center mb-6">
          <div className="w-20 h-20 rounded-full bg-amber-100 border-2 border-amber-700/30 flex items-center justify-center">
            <img
              src="/assets/icons/icon_level_badge_v1.png"
              alt="Logo"
              className="w-14 h-14"
            />
          </div>
        </div>

        <h1 className="text-xl font-bold text-amber-900 text-center mb-1">
          {t('auth.title')}
        </h1>
        <p className="text-xs text-amber-600 text-center mb-6">
          {t('auth.subtitle')}
        </p>

        <form onSubmit={handleSubmit} className="space-y-3">
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder={t('auth.usernamePlaceholder')}
            minLength={3}
            maxLength={32}
            autoComplete="username"
            className="w-full px-3 py-2.5 bg-white border border-amber-300 rounded-lg text-sm text-amber-900 placeholder-amber-400 focus:outline-none focus:ring-2 focus:ring-amber-500"
            autoFocus
          />
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder={t('auth.passwordPlaceholder')}
            minLength={mode === 'register' ? 6 : undefined}
            maxLength={72}
            autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
            className="w-full px-3 py-2.5 bg-white border border-amber-300 rounded-lg text-sm text-amber-900 placeholder-amber-400 focus:outline-none focus:ring-2 focus:ring-amber-500"
          />

          {mode === 'register' && (
            <>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder={t('auth.displayNamePlaceholder')}
                maxLength={80}
                autoComplete="name"
                className="w-full px-3 py-2.5 bg-white border border-amber-300 rounded-lg text-sm text-amber-900 placeholder-amber-400 focus:outline-none focus:ring-2 focus:ring-amber-500"
              />
              <div className="flex gap-1">
                <select
                  value={gender}
                  onChange={(e) => { setGender(e.target.value); if (e.target.value) setShowCustomGender(false) }}
                  className="flex-1 px-3 py-2.5 bg-white border border-amber-300 rounded-lg text-sm text-amber-900 focus:outline-none focus:ring-2 focus:ring-amber-500 appearance-none"
                >
                  <option value="">{t('auth.genderPlaceholder')}</option>
                  <option value="Male">{t('auth.genderMale')}</option>
                  <option value="Female">{t('auth.genderFemale')}</option>
                </select>
                <button
                  type="button"
                  onClick={() => { setShowCustomGender(!showCustomGender); if (!showCustomGender) setGender('') }}
                  className={`px-2.5 py-2.5 border border-amber-300 rounded-lg text-sm text-amber-600 hover:bg-amber-100 transition-colors ${showCustomGender ? 'bg-amber-200' : 'bg-white'}`}
                  title={t('auth.customGender')}
                >
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="w-4 h-4">
                    <path fillRule="evenodd" d="M5.22 8.22a.75.75 0 0 1 1.06 0L10 11.94l3.72-3.72a.75.75 0 1 1 1.06 1.06l-4.25 4.25a.75.75 0 0 1-1.06 0L5.22 9.28a.75.75 0 0 1 0-1.06Z" clipRule="evenodd" />
                  </svg>
                </button>
              </div>
              {showCustomGender && (
                <input
                  type="text"
                  value={customGender}
                  onChange={(e) => setCustomGender(e.target.value)}
                  placeholder={t('auth.customGenderPlaceholder')}
                  maxLength={64}
                  className="w-full px-3 py-2.5 bg-white border border-amber-300 rounded-lg text-sm text-amber-900 placeholder-amber-400 focus:outline-none focus:ring-2 focus:ring-amber-500"
                  autoFocus
                />
              )}
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder={t('auth.emailPlaceholder')}
                maxLength={254}
                autoComplete="email"
                className="w-full px-3 py-2.5 bg-white border border-amber-300 rounded-lg text-sm text-amber-900 placeholder-amber-400 focus:outline-none focus:ring-2 focus:ring-amber-500"
              />
            </>
          )}

          {error && (
            <div className="text-xs text-red-500 bg-red-50 px-2 py-1.5 rounded">
              {error instanceof Error ? error.message : t('auth.failed')}
            </div>
          )}

          <button
            type="submit"
            disabled={isPending || !username.trim()}
            className="w-full py-2.5 bg-amber-700 hover:bg-amber-800 disabled:bg-amber-400 text-white text-sm font-bold rounded-lg transition-colors"
          >
            {isPending ? t('auth.connecting') : mode === 'login' ? t('auth.signIn') : t('auth.createAccount')}
          </button>
        </form>

        <div className="mt-4 text-center">
          <button
            onClick={() => { audio.playSfx('ui_tab_switch', { volume: 0.4 }); setMode(mode === 'login' ? 'register' : 'login') }}
            className="text-xs text-amber-600 hover:text-amber-800 underline"
          >
            {mode === 'login' ? t('auth.switchToRegister') : t('auth.switchToLogin')}
          </button>
        </div>

        {/* Language switcher */}
        <div className="mt-5 pt-4 border-t border-amber-200/50">
          <div className="flex items-center justify-center gap-1.5">
            <span className="text-[10px] text-amber-500 font-medium mr-1">{t('auth.switchLanguage')}:</span>
            {SUPPORTED_LOCALES.map((loc) => (
              <button
                key={loc}
                type="button"
                onClick={() => {
                  audio.playSfx('ui_tab_switch', { volume: 0.4 })
                  setLocaleState(loc)
                  setStoredLocale(loc)
                }}
                className={`px-2 py-1 text-[11px] rounded-md font-medium transition-colors ${
                  locale === loc
                    ? 'bg-amber-700 text-white shadow-sm'
                    : 'bg-amber-100 text-amber-600 hover:bg-amber-200'
                }`}
              >
                {LOCALE_LABELS[loc]}
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
