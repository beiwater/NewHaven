import { useQueryClient } from '@tanstack/react-query'
import { clearAuth } from '@/api/client'
import { useCompany, usePlayerLevel } from '@/api/company.api'
import { useUIStore } from '@/store/ui.store'
import { Icon } from '@/features/ui/Icon'

export function TopBar() {
  const queryClient = useQueryClient()
  const { data: companyData } = useCompany()
  const { data: levelData } = usePlayerLevel()
  const setActiveView = useUIStore((s) => s.setActiveView)
  const setPowerupOpen = useUIStore((s) => s.setPowerupOpen)

  const companyName = companyData?.authCompany?.company ?? 'Mellow Acres Co.'
  const cash = companyData?.authCompany?.money ?? 0
  const lvl = companyData?.levelInfo?.level ?? levelData?.level ?? 1
  const currentXp = levelData?.currentXp ?? companyData?.levelInfo?.xp ?? 0
  const xpToNext = levelData?.xpToNextLevel ?? 100
  const xpPct = xpToNext > 0 ? Math.min(100, (currentXp / xpToNext) * 100) : 0
  const handleLogout = () => {
    queryClient.clear()
    clearAuth()
  }

  return (
    <header className="topbar flex items-center bg-gradient-to-r from-amber-900 via-amber-800 to-amber-900 text-white border-b-2 border-amber-700 shadow-md z-50">
      {/* Company Logo & Name */}
      <div className="flex items-center gap-3 px-4 min-w-[260px] h-full border-r border-amber-700/50">
        <div className="w-12 h-12 rounded-full bg-amber-100 flex items-center justify-center overflow-hidden">
          <Icon name="icon_level_badge_v1" className="w-10 h-10" />
        </div>
        <div className="leading-tight">
          <div className="font-semibold text-sm tracking-tight">{companyName}</div>
          <div className="text-[10px] text-amber-300/80">Farm & Factory</div>
        </div>
      </div>

      {/* Cash */}
      <button
        onClick={() => setActiveView('finance')}
        className="flex items-center gap-2 px-5 h-full border-r border-amber-700/50 hover:bg-amber-700/30 transition-colors text-left"
        title="View financials"
      >
        <Icon name="icon_coin_v1" className="w-7 h-7" />
        <div>
          <div className="text-[10px] text-amber-300/70 uppercase tracking-wider">Cash</div>
          <div className="font-bold text-sm tabular-nums">${cash.toLocaleString()}</div>
        </div>
      </button>

      {/* Level & XP */}
      <div className="flex items-center gap-3 px-5 h-full flex-1">
        <Icon name="icon_level_badge_v1" className="w-7 h-7" />
        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between text-xs">
            <span className="font-semibold">Level {lvl}</span>
            <span className="text-amber-300/80">{currentXp.toLocaleString()} / {xpToNext.toLocaleString()} XP</span>
          </div>
          <div className="mt-1 h-2.5 bg-amber-950/50 rounded-full overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-green-500 to-green-400 rounded-full transition-all duration-500"
              style={{ width: `${xpPct}%` }}
            />
          </div>
        </div>
        <Icon name="icon_xp_v1" className="w-5 h-5" />
      </div>

      {/* Energy */}
      <div className="flex items-center gap-2 px-4 h-full border-l border-amber-700/50">
        <Icon name="icon_energy_v1" className="w-6 h-6" />
        <div className="text-xs">
          <div className="text-[10px] text-amber-300/70">Energy</div>
          <div className="font-semibold">120 / 120</div>
        </div>
      </div>

      {/* Workers */}
      <div className="flex items-center gap-2 px-4 h-full border-l border-amber-700/50">
        <svg className="w-4 h-4 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
          <path d="M9 6a3 3 0 11-6 0 3 3 0 016 0zM17 6a3 3 0 11-6 0 3 3 0 016 0zM12.93 17c.046-.327.07-.66.07-1a6.97 6.97 0 00-1.5-4.33A5 5 0 0119 16v1h-6.07zM6 11a5 5 0 015 5v1H1v-1a5 5 0 015-5z" />
        </svg>
        <div className="text-xs">
          <div className="text-[10px] text-amber-300/70">Workers</div>
          <div className="font-semibold">8 / 10</div>
        </div>
      </div>

      {/* Power-up */}
      <button
        onClick={() => setPowerupOpen(true)}
        className="flex items-center gap-2 px-4 h-full border-l border-amber-700/50 hover:bg-amber-700/30 transition-colors"
        title="Power-ups"
      >
        <svg className="w-5 h-5 text-yellow-300" fill="currentColor" viewBox="0 0 24 24">
          <path d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        <div className="text-xs leading-tight">
          <div className="text-[10px] text-amber-300/70 uppercase tracking-wider">Power-up</div>
          <div className="font-semibold text-yellow-200 text-[11px]">Active</div>
        </div>
      </button>
      {/* Top icons */}
      <div className="flex items-center gap-1 px-3 ml-auto">
        <button className="p-2 rounded-lg hover:bg-amber-700/50 transition-colors" title="Notifications">
          <svg className="w-5 h-5 text-amber-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
          </svg>
        </button>
        <button className="p-2 rounded-lg hover:bg-amber-700/50 transition-colors" title="Settings">
          <svg className="w-5 h-5 text-amber-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </button>
        <button
          type="button"
          onClick={handleLogout}
          className="ml-1 rounded-lg border border-amber-300/30 bg-amber-950/25 px-3 py-2 text-xs font-semibold text-amber-100 transition-colors hover:bg-red-900/45 hover:text-white"
          title="Log out"
        >
          Logout
        </button>
      </div>
    </header>
  )
}
