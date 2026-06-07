import { useCompany, usePlayerLevel } from '@/api/company.api'
import { useActivePowerup } from '@/api/powerup.api'
import { useUIStore } from '@/store/ui.store'
import { Icon } from '@/features/ui/Icon'
import { useTranslation } from 'react-i18next'

export function MobileResourceBar() {
  const { data: companyData } = useCompany()
  const { data: levelData } = usePlayerLevel()
  const { data: activePowerupData } = useActivePowerup()
  const setActiveView = useUIStore((s) => s.setActiveView)
  const setPowerupOpen = useUIStore((s) => s.setPowerupOpen)
  const { t } = useTranslation()

  const cash = companyData?.authCompany?.money ?? 0
  const lvl = companyData?.levelInfo?.level ?? levelData?.level ?? 1
  const currentXp = levelData?.currentXp ?? companyData?.levelInfo?.xp ?? 0
  const xpToNext = levelData?.xpToNextLevel ?? 100
  const xpPct = xpToNext > 0 ? Math.min(100, (currentXp / xpToNext) * 100) : 0
  const financeUnlocked = levelData?.unlocks?.features?.finance ?? companyData?.unlocks?.features?.finance ?? true
  const financeUnlockLevel = levelData?.unlocks?.featureLevels?.finance ?? companyData?.unlocks?.featureLevels?.finance ?? 6
  const activePowerupCount = activePowerupData?.active?.length ?? 0
  const remainingPowerups = activePowerupData?.remaining ?? companyData?.authCompany?.simBoosts ?? 0

  return (
    <div className="flex items-center gap-1 bg-[#f5e6c8] border-b border-amber-700/30 px-1.5 py-1.5 shrink-0" style={{ minHeight: 52 }}>
      {/* Cash */}
      <button
        onClick={() => { if (financeUnlocked) setActiveView('finance') }}
        disabled={!financeUnlocked}
        className={`flex items-center gap-1.5 bg-white/70 rounded-lg border border-amber-300/50 px-2 py-1 min-w-[72px] shrink-0 text-left ${
          financeUnlocked ? 'active:bg-white/90' : 'opacity-60'
        }`}
        title={financeUnlocked ? t('mobile.viewFinancials') : t('mobile.financeUnlock', { level: financeUnlockLevel })}
      >
        <Icon name="icon_coin_v1" className="w-5 h-5" />
        <div className="leading-tight">
          <div className="text-[8px] text-amber-600/80 uppercase tracking-wider">{t('mobile.cash')}</div>
          <div className="text-xs font-bold text-amber-900 tabular-nums">${cash.toLocaleString()}</div>
          {!financeUnlocked && <div className="text-[8px] text-amber-200/80">Lv.{financeUnlockLevel}</div>}
        </div>
      </button>

      {/* Level & XP */}
      <div className="flex items-center gap-1.5 bg-white/70 rounded-lg border border-amber-300/50 px-2 py-1 flex-1 min-w-[110px] shrink-0">
        <Icon name="icon_level_badge_v1" className="w-5 h-5" />
        <div className="flex-1 min-w-0 leading-tight">
          <div className="flex items-center justify-between">
            <span className="text-[10px] font-bold text-amber-900">{t('building.level', { level: lvl })}</span>
            <span className="text-[7px] text-amber-500 tabular-nums">{currentXp.toLocaleString()} / {xpToNext.toLocaleString()} {t('mobile.xp')}</span>
          </div>
          <div className="mt-0.5 h-1.5 bg-amber-200/60 rounded-full overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-green-500 to-green-400 rounded-full transition-all duration-500"
              style={{ width: `${xpPct}%` }}
            />
          </div>
        </div>
        <Icon name="icon_xp_v1" className="w-4 h-4" />
      </div>

      {/* Power-up (Energy) */}
      <button
        onClick={() => setPowerupOpen(true)}
        className="flex items-center gap-1.5 bg-white/70 rounded-lg border border-amber-300/50 px-2 py-1 min-w-[64px] shrink-0 active:bg-white/90"
        title={t('mobile.powerup')}
      >
        <svg className="w-4 h-4 text-yellow-600" fill="currentColor" viewBox="0 0 24 24">
          <path d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        <div className="leading-tight">
          <div className="text-[8px] text-amber-600/80 uppercase tracking-wider">{t('mobile.powerup')}</div>
          <div className="text-[10px] font-semibold text-yellow-700">{activePowerupCount > 0 ? t('mobile.activeCount', { count: activePowerupCount }) : t('mobile.readyCount', { count: remainingPowerups })}</div>
        </div>
      </button>
    </div>
  )
}
