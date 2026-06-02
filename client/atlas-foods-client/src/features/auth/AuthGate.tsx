import { useEffect, useState, type FormEvent } from 'react'
import { useLogin, useRegister } from '@/api/company.api'
import { AUTH_CHANGED_EVENT, isAuthenticated } from '@/api/client'

export function AuthGate({ children }: { children: React.ReactNode }) {
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const login = useLogin()
  const register = useRegister()
  const [authenticated, setAuthenticated] = useState(isAuthenticated())

  useEffect(() => {
    const syncAuth = () => setAuthenticated(isAuthenticated())
    window.addEventListener(AUTH_CHANGED_EVENT, syncAuth)
    window.addEventListener('storage', syncAuth)
    return () => {
      window.removeEventListener(AUTH_CHANGED_EVENT, syncAuth)
      window.removeEventListener('storage', syncAuth)
    }
  }, [])

  if (authenticated) {
    return <>{children}</>
  }

  const isPending = login.isPending || register.isPending
  const error = login.error || register.error

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (!username.trim() || !password.trim()) return
    if (mode === 'login') {
      login.mutate({ username: username.trim(), password })
    } else {
      register.mutate({ username: username.trim(), password })
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
          Mellow Acres Co.
        </h1>
        <p className="text-xs text-amber-600 text-center mb-6">
          Farm & Factory Tycoon
        </p>

        <form onSubmit={handleSubmit} className="space-y-3">
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="Enter your username"
            className="w-full px-3 py-2.5 bg-white border border-amber-300 rounded-lg text-sm text-amber-900 placeholder-amber-400 focus:outline-none focus:ring-2 focus:ring-amber-500"
            autoFocus
          />
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Enter your password"
            className="w-full px-3 py-2.5 bg-white border border-amber-300 rounded-lg text-sm text-amber-900 placeholder-amber-400 focus:outline-none focus:ring-2 focus:ring-amber-500"
          />

          {error && (
            <div className="text-xs text-red-500 bg-red-50 px-2 py-1.5 rounded">
              {error instanceof Error ? error.message : 'Authentication failed'}
            </div>
          )}

          <button
            type="submit"
            disabled={isPending || !username.trim()}
            className="w-full py-2.5 bg-amber-700 hover:bg-amber-800 disabled:bg-amber-400 text-white text-sm font-bold rounded-lg transition-colors"
          >
            {isPending ? 'Connecting...' : mode === 'login' ? 'Sign In' : 'Create Account'}
          </button>
        </form>

        <div className="mt-4 text-center">
          <button
            onClick={() => setMode(mode === 'login' ? 'register' : 'login')}
            className="text-xs text-amber-600 hover:text-amber-800 underline"
          >
            {mode === 'login' ? "Don't have an account? Register" : 'Already have an account? Sign In'}
          </button>
        </div>
      </div>
    </div>
  )
}
