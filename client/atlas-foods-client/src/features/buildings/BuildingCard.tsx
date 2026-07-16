import { useState, type ReactNode } from 'react'
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
    neutral: 'border-amber-300 bg-white/70 text-amber-900 hover:bg-amber-50',
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
  const { data: jobs = [] } = useProductionJobs()
  const upgradeBuilding = useUpgradeBuilding()
  const demolishBuilding = useDemolishBuilding()
  const stashBuilding = useStashBuilding()
  const startBuildingMove = useUIStore((state) => state.startBuildingMove)

  const pendingJobs = jobs.filter((job) => job.buildingId === building.id && job.status !== 'claimed').length
  const activeSales = building.shelves?.length ?? 0
  const upgradeCost = (building.level + 1) * (building.baseCost ?? 10000)
  const demolitionRefund = Math.floor((building.baseCost ?? 10000) * 0.5)
  const buildingName = building.name ?? t(`building.name_${building.kind}`, t('building.building'))

  const handleMove = () => {
    startBuildingMove(building.id)
    onClose()
  }

  const handleUpgrade = () => {
    audio.playSfx('build_upgrade')
    upgradeBuilding.mutate(building.id)
  }

  const handleStash = () => {
    stashBuilding.mutate(building.id, { onSuccess: onClose })
  }

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
    <div className="flex h-full min-h-0 flex-col">
      <header className="flex shrink-0 items-center gap-3 border-b border-amber-300/70 bg-[#2f302f] px-4 py-3 text-white sm:px-6">
        <img src={buildingIcon(building.kind)} alt="" className="h-10 w-10 rounded-xl bg-amber-50/95 object-contain p-1.5" />
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="truncate text-base font-black sm:text-lg">{buildingName}</h1>
            <span className={`rounded-full px-2 py-0.5 text-[9px] font-black uppercase tracking-wider ${building.isRetail ? 'bg-green-500/20 text-green-200' : 'bg-cyan-500/20 text-cyan-100'}`}>
              {building.isRetail ? t('building.retail') : t('building.productionBuilding')}
            </span>
          </div>
          <p className="text-[11px] font-semibold text-amber-100/75">{t('building.level', { level: building.level })} · {t('building.operationCenter')}</p>
        </div>
        <button
          type="button"
          autoFocus
          aria-label={t('common.close')}
          onClick={onClose}
          className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-white/15 bg-white/5 text-xl text-white transition-colors hover:bg-white/15 focus:outline-none focus:ring-2 focus:ring-amber-300"
        >
          ×
        </button>
      </header>

      <div className="grid min-h-0 flex-1 grid-rows-[auto_minmax(0,1fr)] md:grid-cols-[250px_minmax(0,1fr)] md:grid-rows-1">
        <aside className="max-h-[236px] overflow-y-auto border-b border-amber-300/70 bg-[#373835] p-4 text-white md:max-h-none md:border-b-0 md:border-r">
          <div className="grid grid-cols-[96px_minmax(0,1fr)] gap-4 md:block">
            <div className="flex h-24 items-center justify-center rounded-2xl bg-gradient-to-br from-amber-50 to-amber-200/90 p-2 md:h-36">
              <img src={getBuildingPreview(building.kind)} alt={buildingName} className="max-h-full max-w-full object-contain" />
            </div>

            <div className="min-w-0 md:mt-4">
              <dl className="space-y-1.5 text-[11px]">
                <div className="flex justify-between gap-2 border-b border-white/10 pb-1.5">
                  <dt className="text-white/55">{t('building.status')}</dt>
                  <dd className="font-black text-amber-100">{building.status ?? t('building.idle')}</dd>
                </div>
                <div className="flex justify-between gap-2 border-b border-white/10 pb-1.5">
                  <dt className="text-white/55">{building.isRetail ? t('building.activeSales') : t('building.activeLines')}</dt>
                  <dd className="font-black text-amber-100">{building.isRetail ? activeSales : pendingJobs}</dd>
                </div>
                <div className="flex justify-between gap-2">
                  <dt className="text-white/55">{t('building.nextUpgrade')}</dt>
                  <dd className="font-black text-amber-100">${upgradeCost.toLocaleString()}</dd>
                </div>
              </dl>
            </div>
          </div>

          <div className="mt-4 grid grid-cols-2 gap-2 md:grid-cols-1">
            <ActionButton tone="primary" disabled={upgradeBuilding.isPending} onClick={handleUpgrade}>
              {upgradeBuilding.isPending ? '…' : `${t('building.upgrade')} · $${upgradeCost.toLocaleString()}`}
            </ActionButton>
            <ActionButton disabled={!hasFreeSlots} onClick={handleMove}>{t('building.move')}</ActionButton>
            {building.placed !== false && (
              <ActionButton disabled={stashBuilding.isPending} onClick={handleStash}>
                {stashBuilding.isPending ? t('building.stashing') : t('building.stash')}
              </ActionButton>
            )}
            <ActionButton tone="danger" onClick={() => setShowDemolishConfirm(true)}>{t('building.demolish')}</ActionButton>
          </div>

          {showDemolishConfirm && (
            <div className="mt-3 rounded-xl border border-red-300/40 bg-red-950/45 p-3">
              <p className="text-[10px] font-semibold leading-4 text-red-100">
                {t('building.demolishWarning', { name: buildingName, amount: demolitionRefund.toLocaleString() })}
              </p>
              <div className="mt-2 grid grid-cols-2 gap-2">
                <button type="button" disabled={demolishBuilding.isPending} onClick={handleDemolish} className="rounded-lg bg-red-600 px-2 py-1.5 text-[10px] font-black text-white disabled:opacity-50">
                  {demolishBuilding.isPending ? t('building.demolishing') : t('building.confirmDemolish')}
                </button>
                <button type="button" onClick={() => setShowDemolishConfirm(false)} className="rounded-lg bg-white/10 px-2 py-1.5 text-[10px] font-black text-white hover:bg-white/20">
                  {t('common.cancel')}
                </button>
              </div>
            </div>
          )}

          {actionError instanceof Error && (
            <div className="mt-3 rounded-lg border border-red-300/40 bg-red-950/45 px-3 py-2 text-[10px] font-semibold text-red-100">{actionError.message}</div>
          )}
        </aside>

        {building.isRetail ? <RetailOperations building={building} /> : <ProductionOperations building={building} />}
      </div>
    </div>
  )
}
