import { useEffect, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import {
  useDemolishBuilding,
  useStashBuilding,
  useUpgradeBuilding,
} from '@/api/buildings.api'
import { useProductionJobs } from '@/api/production.api'
import { audio } from '@/audio/AudioManager'
import { buildingIcon } from '@/game/icons'
import type { Building } from '@/game/types'
import { useUIStore } from '@/store/ui.store'
import { ProductionOperations } from './ProductionOperations'
import { RetailOperations } from './RetailOperations'

const BUILDING_PREVIEW: Record<number, string> = {
  1: '/assets/buildings/grain_plot_lv1_idle_trimmed.png',
  2: '/assets/buildings/mill_house_lv1_idle_trimmed.png',
  3: '/assets/buildings/bakery_shop_lv1_idle_trimmed.png',
  4: '/assets/buildings/meal_kiosk_lv1_idle_trimmed.png',
}

function getBuildingPreview(kind: number): string {
  return BUILDING_PREVIEW[kind] ?? buildingIcon(kind)
}

function timeLabel(seconds: number): string {
  const minutes = Math.max(1, Math.ceil(seconds / 60))
  const hours = Math.floor(minutes / 60)
  return hours > 0 ? `${hours}h ${minutes % 60}m` : `${minutes}m`
}

function countdown(completesAt: string | undefined, now: number): string {
  if (!completesAt) return '--'
  const seconds = Math.max(0, Math.ceil((new Date(completesAt).getTime() - now) / 1000))
  return timeLabel(seconds)
}

interface BuildingCardProps {
  building: Building
  hasFreeSlots: boolean
  onClose: () => void
}

function ActionButton({
  children,
  disabled,
  onClick,
  tone = 'neutral',
}: {
  children: ReactNode
  disabled?: boolean
  onClick: () => void
  tone?: 'neutral' | 'danger' | 'primary'
}) {
  const tones = {
    neutral: 'border-amber-300 bg-white text-amber-900 hover:bg-amber-50',
    danger: 'border-red-200 bg-red-50 text-red-700 hover:bg-red-100',
    primary: 'border-cyan-700 bg-cyan-700 text-white hover:bg-cyan-800',
  }
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className={`rounded-xl border px-3 py-2 text-xs font-black transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${tones[tone]}`}
    >
      {children}
    </button>
  )
}

export function BuildingCard({ building, hasFreeSlots, onClose }: BuildingCardProps) {
  const { t } = useTranslation()
  const [showDemolishConfirm, setShowDemolishConfirm] = useState(false)
  const [showUpgradeConfirm, setShowUpgradeConfirm] = useState(false)
  const [now, setNow] = useState(() => Date.now())
  const { data: jobs = [] } = useProductionJobs()
  const upgradeBuilding = useUpgradeBuilding()
  const demolishBuilding = useDemolishBuilding()
  const stashBuilding = useStashBuilding()
  const startBuildingMove = useUIStore((state) => state.startBuildingMove)

  const isUpgrading = Boolean(building.upgradeTargetLevel && building.upgradeTargetLevel > building.level)
  const pendingJobs = jobs.filter((job) => job.buildingId === building.id && job.status !== 'claimed').length
  const activeSales = building.shelves?.length ?? 0
  const upgradeRequiresIdle = pendingJobs > 0 || activeSales > 0
  const upgradeCost = building.nextUpgradeCost ?? building.level * (building.baseCost ?? 10000)
  const upgradeSeconds = building.nextUpgradeDurationSeconds ?? 60
  const demolitionRefund = Math.floor((building.baseCost ?? 10000) * 0.5)
  const buildingName = building.name ?? t(`building.name_${building.kind}`, t('building.building'))
  const completionTime = new Date(now + upgradeSeconds * 1000)

  useEffect(() => {
    if (!isUpgrading && !showUpgradeConfirm) return
    const interval = window.setInterval(() => setNow(Date.now()), 1_000)
    return () => window.clearInterval(interval)
  }, [isUpgrading, showUpgradeConfirm])

  const handleMove = () => {
    startBuildingMove(building.id)
    onClose()
  }
  const handleStartUpgrade = () => {
    audio.playSfx('build_construction_start')
    upgradeBuilding.mutate(building.id, { onSuccess: () => setShowUpgradeConfirm(false) })
  }
  const handleStash = () => stashBuilding.mutate(building.id, { onSuccess: onClose })
  const handleDemolish = () => {
    audio.playSfx('build_demolish')
    demolishBuilding.mutate(building.id, {
      onSuccess: () => {
        setShowDemolishConfirm(false)
        onClose()
      },
    })
  }
  const actionError = upgradeBuilding.error ?? stashBuilding.error ?? demolishBuilding.error

  return (
    <div className="relative flex h-full min-h-0 flex-col bg-[#fff8ea] text-amber-950">
      <header className="flex shrink-0 items-center gap-3 border-b border-amber-300 bg-[#fffaf0] px-4 py-3 sm:px-6">
        <img src={buildingIcon(building.kind)} alt="" className="h-10 w-10 rounded-xl border border-amber-200 bg-amber-50 object-contain p-1.5" />
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="truncate text-base font-black sm:text-lg">{buildingName}</h1>
            <span className={`rounded-full px-2 py-0.5 text-[9px] font-black uppercase tracking-wider ${building.isRetail ? 'bg-green-100 text-green-700' : 'bg-cyan-100 text-cyan-800'}`}>
              {building.isRetail ? t('building.retail') : t('building.productionBuilding')}
            </span>
          </div>
          <p className="text-[11px] font-semibold text-amber-700">{t('building.level', { level: building.level })} · {t('building.operationCenter')}</p>
        </div>
        <button type="button" autoFocus aria-label={t('common.close')} onClick={onClose} className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-amber-300 bg-white text-xl text-amber-900 transition-colors hover:bg-amber-100 focus:outline-none focus:ring-2 focus:ring-amber-400">×</button>
      </header>

      <div className="shrink-0 border-b border-amber-200 bg-white/75 px-4 py-3 sm:px-6">
        <div className="mx-auto flex max-w-5xl flex-wrap items-center gap-3">
          <div className="flex h-20 w-24 shrink-0 items-center justify-center rounded-2xl border border-amber-200 bg-gradient-to-br from-amber-50 to-amber-200/80 p-2">
            <img src={getBuildingPreview(building.kind)} alt={buildingName} className="max-h-full max-w-full object-contain" />
          </div>
          <dl className="grid min-w-[240px] flex-1 grid-cols-2 gap-x-5 gap-y-1.5 text-[11px] sm:grid-cols-3">
            <div><dt className="font-semibold text-amber-600">{t('building.status')}</dt><dd className="font-black text-amber-950">{isUpgrading ? t('building.underConstruction') : building.status ?? t('building.idle')}</dd></div>
            <div><dt className="font-semibold text-amber-600">{building.isRetail ? t('building.activeSales') : t('building.activeLines')}</dt><dd className="font-black text-amber-950">{building.isRetail ? activeSales : pendingJobs}</dd></div>
            <div><dt className="font-semibold text-amber-600">{t('building.nextUpgrade')}</dt><dd className="font-black text-amber-950">${upgradeCost.toLocaleString()}</dd></div>
          </dl>
          {!isUpgrading && (
            <div className="grid grid-cols-2 gap-2 sm:flex sm:flex-wrap">
              <ActionButton tone="primary" disabled={upgradeBuilding.isPending || upgradeRequiresIdle} onClick={() => setShowUpgradeConfirm(true)}>{t('building.upgrade')}</ActionButton>
              <ActionButton disabled={!hasFreeSlots} onClick={handleMove}>{t('building.move')}</ActionButton>
              {building.placed !== false && <ActionButton disabled={stashBuilding.isPending} onClick={handleStash}>{stashBuilding.isPending ? t('building.stashing') : t('building.stash')}</ActionButton>}
              <ActionButton tone="danger" onClick={() => setShowDemolishConfirm(true)}>{t('building.demolish')}</ActionButton>
            </div>
          )}
        </div>
        {!isUpgrading && upgradeRequiresIdle && <p className="mx-auto mt-2 max-w-5xl text-[10px] font-semibold text-amber-700">{t('building.upgradeNeedsIdle')}</p>}
      </div>

      {isUpgrading ? (
        <div className="flex min-h-0 flex-1 items-center justify-center bg-gradient-to-br from-[#fff4dc] via-[#fffaf0] to-[#f3dfb8] p-6">
          <section className="w-full max-w-xl rounded-3xl border border-amber-300 bg-white/85 p-6 text-center shadow-sm">
            <img src={getBuildingPreview(building.kind)} alt="" className="mx-auto h-28 w-36 object-contain opacity-75" />
            <p className="mt-2 text-[10px] font-black uppercase tracking-[0.2em] text-cyan-700">{t('building.underConstruction')}</p>
            <h2 className="mt-1 text-2xl font-black text-amber-950">{t('building.upgradingTo', { level: building.upgradeTargetLevel })}</h2>
            <p className="mt-3 text-sm font-semibold text-amber-800">{t('building.constructionLocksBuilding')}</p>
            <div className="mt-5 rounded-2xl bg-amber-50 p-4">
              <div className="text-3xl font-black tabular-nums text-cyan-800">{countdown(building.upgradeCompletesAt, now)}</div>
              <div className="mt-1 text-xs font-semibold text-amber-700">{t('building.untilUpgradeComplete')}</div>
            </div>
          </section>
        </div>
      ) : (
        <div className="min-h-0 flex-1">{building.isRetail ? <RetailOperations building={building} /> : <ProductionOperations building={building} />}</div>
      )}

      {showUpgradeConfirm && (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-amber-950/25 p-4 backdrop-blur-sm">
          <section aria-label={t('building.upgrade')} className="w-full max-w-md rounded-3xl border border-amber-300 bg-[#fffaf0] p-5 shadow-2xl">
            <div className="flex items-start gap-3">
              <img src={getBuildingPreview(building.kind)} alt="" className="h-16 w-20 rounded-2xl bg-amber-100 object-contain p-1" />
              <div><p className="text-[10px] font-black uppercase tracking-[0.16em] text-cyan-700">{t('building.constructionPlan')}</p><h2 className="text-xl font-black text-amber-950">{t('building.upgradeTo', { level: building.level + 1 })}</h2></div>
            </div>
            <dl className="mt-5 space-y-2 rounded-2xl border border-amber-200 bg-white/75 p-4 text-sm">
              <div className="flex justify-between gap-4"><dt className="text-amber-700">{t('building.upgradeCost')}</dt><dd className="font-black text-amber-950">${upgradeCost.toLocaleString()}</dd></div>
              <div className="flex justify-between gap-4"><dt className="text-amber-700">{t('building.constructionTime')}</dt><dd className="font-black text-amber-950">{timeLabel(upgradeSeconds)}</dd></div>
              <div className="flex justify-between gap-4"><dt className="text-amber-700">{t('building.completesAt')}</dt><dd className="font-black text-amber-950">{completionTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</dd></div>
            </dl>
            <p className="mt-3 text-xs leading-5 text-amber-700">{t('building.constructionNotice')}</p>
            <div className="mt-5 flex justify-end gap-2"><ActionButton onClick={() => setShowUpgradeConfirm(false)}>{t('common.cancel')}</ActionButton><ActionButton tone="primary" disabled={upgradeBuilding.isPending} onClick={handleStartUpgrade}>{upgradeBuilding.isPending ? '…' : t('building.startUpgrade')}</ActionButton></div>
          </section>
        </div>
      )}

      {showDemolishConfirm && (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-amber-950/25 p-4 backdrop-blur-sm">
          <section className="w-full max-w-md rounded-3xl border border-red-200 bg-[#fffaf0] p-5 shadow-2xl">
            <h2 className="text-lg font-black text-red-800">{t('building.demolish')}</h2>
            <p className="mt-2 text-sm font-semibold leading-5 text-red-700">{t('building.demolishWarning', { name: buildingName, amount: demolitionRefund.toLocaleString() })}</p>
            <div className="mt-5 flex justify-end gap-2"><ActionButton onClick={() => setShowDemolishConfirm(false)}>{t('common.cancel')}</ActionButton><ActionButton tone="danger" disabled={demolishBuilding.isPending} onClick={handleDemolish}>{demolishBuilding.isPending ? t('building.demolishing') : t('building.confirmDemolish')}</ActionButton></div>
          </section>
        </div>
      )}

      {actionError instanceof Error && <div className="absolute bottom-4 left-4 right-4 z-20 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-xs font-semibold text-red-700 sm:left-auto sm:max-w-md">{actionError.message}</div>}
    </div>
  )
}
