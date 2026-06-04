import { useUIStore, type ActiveView } from '@/store/ui.store'
import { systemIcon } from '@/game/icons'
import { useCompany, usePlayerLevel } from '@/api/company.api'
import { clearAuth } from '@/api/client'
import { useQueryClient } from '@tanstack/react-query'

const navItems: Array<{ id: ActiveView; label: string; sysIcon: string; feature: string }> = [
  { id: 'map', label: 'Map', sysIcon: 'inventory', feature: 'map' },
  { id: 'build', label: 'Build', sysIcon: 'market', feature: 'build' },
  { id: 'warehouse', label: 'Warehouse', sysIcon: 'inventory', feature: 'warehouse' },
  { id: 'market', label: 'Market', sysIcon: 'market', feature: 'market' },
  { id: 'contracts', label: 'Contracts', sysIcon: 'quest', feature: 'contracts' },
  { id: 'research', label: 'Research', sysIcon: 'research', feature: 'research' },
  { id: 'executives', label: 'Executives', sysIcon: 'executive', feature: 'executives' },
  { id: 'finance', label: 'Finance', sysIcon: 'financial', feature: 'finance' },
  { id: 'inspect', label: 'Inspect', sysIcon: 'achievement', feature: 'map' },
  { id: 'leaderboard', label: 'Leaderboard', sysIcon: 'leaderboard', feature: 'leaderboard' },
  { id: 'settings', label: 'Settings', sysIcon: 'settings', feature: 'map' },
];

export function MobileDrawer({ open, onClose }: { open: boolean; onClose: () => void }) {
  const activeView = useUIStore((s) => s.activeView)
  const setActiveView = useUIStore((s) => s.setActiveView)
  const { data: companyData } = useCompany()
  const { data: levelData } = usePlayerLevel()
  const queryClient = useQueryClient()

  const level = companyData?.levelInfo?.level ?? levelData?.level ?? 1
  const cash = companyData?.authCompany?.money ?? 0
  const companyName = companyData?.authCompany?.company ?? 'Atlas'

  const handleNav = (view: ActiveView) => {
    setActiveView(view)
    onClose()
  }

  const handleLogout = () => {
    queryClient.clear()
    clearAuth()
  }

  return (
    <>
      {/* Backdrop */}
      {open && (
        <div
          className="fixed inset-0 bg-black/50 z-50 transition-opacity"
          onClick={onClose}
        />
      )}

      {/* Drawer */}
      <div
        className={`fixed top-0 left-0 bottom-0 w-64 bg-[#f5e6c8] border-r-2 border-amber-700/30 z-50 flex flex-col transition-transform duration-250 ease-out ${
          open ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        {/* Header */}
        <div className="p-4 bg-amber-900 text-white">
          <div className="font-bold text-sm">{companyName}</div>
          <div className="text-[10px] text-amber-300/70 mt-0.5">
            Lv.{level} · ${cash.toLocaleString()}
          </div>
        </div>

        {/* Nav items */}
        <div className="flex-1 overflow-y-auto py-2">
          {navItems.map((item) => {
            const isUnlocked = levelData?.unlocks?.features?.[item.feature] ?? true
            const unlockLevel = levelData?.unlocks?.featureLevels?.[item.feature] ?? 1
            const isActive = activeView === item.id || (item.id === 'warehouse' && activeView === 'chain')

            return (
              <button
                key={item.id}
                onClick={() => { if (isUnlocked) handleNav(item.id) }}
                disabled={!isUnlocked}
                className={`
                  w-full flex items-center gap-3 px-4 py-3 text-left transition-colors
                  ${!isUnlocked ? 'opacity-40' : ''}
                  ${isActive
                    ? 'bg-amber-200/70 text-amber-900 font-semibold'
                    : 'text-amber-800/70 hover:bg-amber-100/50'
                  }
                `}
              >
                <img
                  src={systemIcon(item.sysIcon)}
                  alt={item.label}
                  className="w-6 h-6 object-contain"
                  loading="lazy"
                />
                <span className="text-xs uppercase tracking-wider">{item.label}</span>
                {!isUnlocked && (
                  <span className="ml-auto text-[9px] text-amber-500 font-bold">Lv.{unlockLevel}</span>
                )}
              </button>
            )
          })}
        </div>

        {/* Sign Out */}
        <div className="p-3 border-t border-amber-700/20">
          <button
            onClick={handleLogout}
            className="w-full flex items-center justify-center gap-2 py-2 rounded-lg border border-red-400/40 bg-red-900/10 text-red-600 text-xs font-semibold hover:bg-red-100/50 transition-colors"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
            Sign Out
          </button>
        </div>
      </div>
    </>
  )
}
