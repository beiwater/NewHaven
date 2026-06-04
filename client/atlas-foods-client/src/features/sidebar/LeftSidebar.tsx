import { useUIStore, type ActiveView } from '@/store/ui.store'
import { useTranslation } from 'react-i18next'
import { systemIcon } from '@/game/icons'
import { useCompany, usePlayerLevel } from '@/api/company.api'
import { audio } from '@/audio/AudioManager'

const navItems: Array<{ id: ActiveView; label: string; sysIcon: string; feature: string }> = [
  { id: 'map', label: 'map', sysIcon: 'inventory', feature: 'map' },
  { id: 'build', label: 'build', sysIcon: 'market', feature: 'build' },
  { id: 'warehouse', label: 'warehouse', sysIcon: 'inventory', feature: 'warehouse' },
  { id: 'market', label: 'market', sysIcon: 'market', feature: 'market' },
  { id: 'contracts', label: 'contracts', sysIcon: 'quest', feature: 'contracts' },
  { id: 'research', label: 'research', sysIcon: 'research', feature: 'research' },
  { id: 'executives', label: 'executives', sysIcon: 'executive', feature: 'executives' },
  { id: 'finance', label: 'finance', sysIcon: 'financial', feature: 'finance' },
  { id: 'leaderboard', label: 'leaderboard', sysIcon: 'leaderboard', feature: 'leaderboard' },
  { id: 'inspect', label: 'inspect', sysIcon: 'achievement', feature: 'map' },
  { id: 'settings', label: 'settings', sysIcon: 'settings', feature: 'map' },
]
export function LeftSidebar() {
  const { t } = useTranslation()
  const activeView = useUIStore((s) => s.activeView)
  const setActiveView = useUIStore((s) => s.setActiveView)
  const { data: companyData } = useCompany()
  const { data: levelData } = usePlayerLevel()
  const level = companyData?.levelInfo?.level ?? companyData?.authCompany?.level ?? 1
  const cash = companyData?.authCompany?.money ?? 0
  const unlocks = levelData?.unlocks ?? companyData?.unlocks

  return (
    <nav className="sidebar flex flex-col bg-[#f5e6c8] border-r-2 border-amber-700/30 overflow-y-auto">
      {navItems.map((item) => (
        (() => {
          const isUnlocked = unlocks?.features?.[item.feature] ?? true
          const unlockLevel = unlocks?.featureLevels?.[item.feature] ?? 1
          const isActive = activeView === item.id || (item.id === 'warehouse' && activeView === 'chain')
          return (
            <button
              key={item.id}
              onClick={() => {
              if (isUnlocked) { setActiveView(item.id); audio.playSfx('ui_button_click', { volume: 0.4 }) }
              }}
              disabled={!isUnlocked}
              title={isUnlocked ? t(`nav.${item.label}`) : t('nav.locked', { level: unlockLevel })}
              className={`
            flex flex-col items-center justify-center gap-1 py-4 px-2
            border-b border-amber-700/10 transition-all duration-150
            ${!isUnlocked
              ? 'cursor-not-allowed opacity-45 grayscale text-amber-800/50'
              : isActive
                ? 'bg-amber-200/70 text-amber-900 font-semibold shadow-inner'
                : 'text-amber-800/70 hover:bg-amber-100/50 hover:text-amber-900'
            }
          `}
              style={{ height: isActive ? '68px' : '80px' }}
            >
              <img
                src={systemIcon(item.sysIcon)}
                alt={t(`nav.${item.label}`)}
                className="w-9 h-9 object-contain"
                loading="lazy"
              />
              <span className="text-[11px] uppercase tracking-wider">{t(`nav.${item.label}`)}</span>
              {!isUnlocked && <span className="text-[9px] font-bold">Lv.{unlockLevel}</span>}
            </button>
          )
        })()
      ))}

      {/* Company Status */}
      <div className="mt-auto p-3 mx-2 mb-3 bg-amber-900/10 rounded-lg border border-amber-700/20">
        <div className="text-[9px] uppercase tracking-widest text-amber-700/60 text-center mb-1">
          {t('nav.companyStatus')}
        </div>
        <div className="flex items-center justify-center gap-1">
          <svg className="w-3 h-3 text-yellow-600" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 1.944A11.954 11.954 0 012.166 5C2.056 5.647 2 6.319 2 7c0 5.225 3.34 9.67 8 11.317C14.66 16.67 18 12.225 18 7c0-.682-.057-1.353-.166-2A11.954 11.954 0 0110 1.944zM11 14a1 1 0 11-2 0 1 1 0 012 0zm0-7a1 1 0 10-2 0v3a1 1 0 102 0V7z" clipRule="evenodd" />
          </svg>
          <span className="text-xs font-bold text-amber-800">Lv.{level}</span>
        </div>
        <div className="text-[9px] text-amber-700/50 text-center mt-0.5">${cash.toLocaleString()}</div>
      </div>
    </nav>
  )
}
