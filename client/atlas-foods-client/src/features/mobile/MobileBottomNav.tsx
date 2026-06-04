import { useUIStore, type ActiveView } from '@/store/ui.store'
import { systemIcon } from '@/game/icons'
import { usePlayerLevel } from '@/api/company.api'
import { useTranslation } from 'react-i18next'

const navItems: Array<{ id: ActiveView; label: string; sysIcon: string; feature: string }> = [
  { id: 'map', label: 'company', sysIcon: 'inventory', feature: 'map' },
  { id: 'build', label: 'buildings', sysIcon: 'market', feature: 'build' },
  { id: 'market', label: 'market', sysIcon: 'market', feature: 'market' },
  { id: 'research', label: 'research', sysIcon: 'research', feature: 'research' },
  { id: 'finance', label: 'finance', sysIcon: 'financial', feature: 'finance' },
  { id: 'inspect', label: 'more', sysIcon: 'leaderboard', feature: '__always__' },
]

export function MobileBottomNav() {
  const activeView = useUIStore((s) => s.activeView)
  const setActiveView = useUIStore((s) => s.setActiveView)
  const { data: levelData } = usePlayerLevel()
  const { t } = useTranslation()

  return (
    <nav className="flex items-stretch bg-[#3d2b1f] border-t-2 border-amber-700/40 shrink-0 safe-bottom" style={{ height: 56 }}>
      {navItems.map((item) => {
        const isUnlocked = levelData?.unlocks?.features?.[item.feature] ?? true
        const isActive = activeView === item.id || (item.id === 'map' && activeView === 'map')

        return (
          <button
            key={item.id}
            onClick={() => { if (isUnlocked) setActiveView(item.id) }}
            disabled={!isUnlocked}
            className={`
              flex-1 flex flex-col items-center justify-center gap-0.5 transition-colors min-w-0
              ${!isUnlocked ? 'opacity-40' : ''}
              ${isActive
                ? 'bg-amber-700/40 text-amber-200'
                : 'text-amber-600/70 active:bg-amber-800/20'
              }
            `}
          >
            <img
              src={systemIcon(item.sysIcon)}
              alt={t(`nav.${item.label}`)}
              className="w-5 h-5 object-contain"
              loading="lazy"
            />
            <span className="text-[9px] font-semibold uppercase tracking-wider truncate max-w-full px-0.5">
              {t(`nav.${item.label}`)}
            </span>
          </button>
        )
      })}
    </nav>
  )
}
