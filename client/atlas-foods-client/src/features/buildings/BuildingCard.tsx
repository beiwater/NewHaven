import { useMemo, useState, useEffect } from 'react'
import { useUIStore } from '@/store/ui.store'
import { useProductionJobs, useStartProduction, useClaimProduction, useProductionOptions } from '@/api/production.api'
import { useMoveBuilding, useDemolishBuilding, useStashBuilding, useUpgradeBuilding, useStockShelf, useUnstockShelf, useSetShelfPrice } from '@/api/buildings.api'
import type { Building, ProductionJob } from '@/game/types'
import { Icon } from '@/features/ui/Icon'
import { resourceIcon, resourceName } from '@/game/resources'
import { buildingIcon } from '@/game/icons'
import { audio } from '@/audio/AudioManager'
import { useTranslation } from 'react-i18next'

const BUILDING_PREVIEW: Record<number, string> = {
  1: '/assets/buildings/grain_plot_lv1_idle_trimmed.png',
  2: '/assets/buildings/mill_house_lv1_idle_trimmed.png',
  3: '/assets/buildings/bakery_shop_lv1_idle_trimmed.png',
  4: '/assets/buildings/meal_kiosk_lv1_idle_trimmed.png',
}

function getBuildingPreview(kind: number): string {
  return BUILDING_PREVIEW[kind] ?? buildingIcon(kind)
}

function outputPerHour(baseOutput: number | undefined, level: number): number {
  return Math.max(0, baseOutput ?? 0) * Math.max(1, level)
}

function maxAmountFor48Hours(rate: number): number {
  return Math.max(1, Math.floor(rate * 48))
}

function durationLabel(amount: number, rate: number): string {
  if (rate <= 0) return '--'
  const totalSeconds = Math.max(30, Math.ceil((amount / rate) * 3600))
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.ceil((totalSeconds % 3600) / 60)
  if (hours > 0 && minutes > 0) return `${hours}h ${minutes}m`
  if (hours > 0) return `${hours}h`
  return `${Math.max(1, minutes)}m`
}

function countdownDisplay(completesAt: string | undefined, now: number): string {
  if (!completesAt) return ''
  const remaining = new Date(completesAt).getTime() - now
  if (remaining <= 0) return '00:00'
  const totalSec = Math.ceil(remaining / 1000)
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  if (h > 0) return `${h}h ${m.toString().padStart(2, '0')}m`
  return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
}

function progressPct(startedAt: string | undefined, completesAt: string | undefined, now: number): number {
  if (!startedAt || !completesAt) return 0
  const total = new Date(completesAt).getTime() - new Date(startedAt).getTime()
  if (total <= 0) return 100
  const elapsed = now - new Date(startedAt).getTime()
  return Math.min(100, Math.max(0, (elapsed / total) * 100))
}

function accruedAmount(job: ProductionJob, now: number): number {
  if (job.status === 'claimed') return job.amount
  if (!job.startedAt) return 0
  const elapsed = (now - new Date(job.startedAt).getTime()) / 1000
  const duration = Math.max(1, (new Date(job.completesAt).getTime() - new Date(job.startedAt).getTime()) / 1000)
  return Math.min(job.amount, Math.floor((elapsed / duration) * job.amount))
}

function collectableAmount(job: ProductionJob, now: number): number {
  return Math.max(0, accruedAmount(job, now) - (job.claimedAmount ?? 0))
}

interface BuildingCardProps {
  building: Building
  hasFreeSlots: boolean
  onClose: () => void
}

export function BuildingCard({ building, hasFreeSlots, onClose }: BuildingCardProps) {
  const [selectedResourceId, setSelectedResourceId] = useState<number | null>(null)
  const [amount, setAmount] = useState('10')
  const [showDemolishConfirm, setShowDemolishConfirm] = useState(false)
  const [now, setNow] = useState(Date.now())
  const { t } = useTranslation()

  // Production hooks (only used for non-retail buildings)
  const { data: jobsData } = useProductionJobs()
  const { data: optionsData = [] } = useProductionOptions(building.id)
  const buildingJobs = (jobsData ?? []).filter((j) => j.buildingId === building.id)
  const startProd = useStartProduction()
  const claimProd = useClaimProduction()

  // Retail hooks (only used for retail buildings)
  const stockShelf = useStockShelf()
  const unstockShelf = useUnstockShelf()
  const setShelfPriceMut = useSetShelfPrice()

  const moveBuilding = useMoveBuilding()
  const upgradeBuilding = useUpgradeBuilding()
  const demolishBuilding = useDemolishBuilding()
  const stashBuilding = useStashBuilding()
  const startBuildingMove = useUIStore((s) => s.startBuildingMove)
  const setActiveView = useUIStore((s) => s.setActiveView)

  const runningJob = buildingJobs.find((j) => j.status === 'running')
  const collectableJob = buildingJobs.find((j) => collectableAmount(j, now) > 0)
  const shouldGuideProductionStart = !collectableJob && buildingJobs.length === 0

  // Countdown ticker
  useEffect(() => {
    if (!runningJob) return
    const interval = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(interval)
  }, [runningJob?.id])

  const selectedOption = useMemo(() => {
    if (optionsData.length === 0) return null
    return optionsData.find((o) => o.resourceId === selectedResourceId) ?? optionsData[0]
  }, [optionsData, selectedResourceId])

  const numericAmount = Math.max(1, parseInt(amount, 10) || 1)
  const selectedRate = outputPerHour(selectedOption?.producedPerHourRaw, building.level)
  const maxAmount = maxAmountFor48Hours(selectedRate)
  const safeAmount = Math.min(numericAmount, maxAmount)
  const estimatedDuration = durationLabel(safeAmount, selectedRate)
  const upgradeCost = (building.level + 1) * (building.baseCost ?? 10000)

  const sortedOptions = useMemo(() => {
    const starter = new Set(building.starterProduces ?? [])
    return [...optionsData].sort((a, b) => {
      const ar = starter.has(a.resourceId) ? 0 : 1
      const br = starter.has(b.resourceId) ? 0 : 1
      return ar - br
    })
  }, [building.starterProduces, optionsData])

  const handleStart = () => {
    if (!selectedOption) return
    audio.playSfx('build_confirm')
    startProd.mutate({
      buildingId: building.id,
      kind: selectedOption.resourceId,
      amount: safeAmount,
    })
  }

  const handleCollect = () => {
    if (collectableJob) {
      audio.playSfx('resource_pickup')
      claimProd.mutate(collectableJob.id)
    }
  }

  const handleMove = () => {
    startBuildingMove(building.id)
  }

  const handleUpgrade = () => {
    audio.playSfx('build_upgrade')
    upgradeBuilding.mutate(building.id)
  }

  const handleViewJobs = () => {
    setActiveView('production')
  }

  const handleDemolish = () => {
    audio.playSfx('build_demolish')
    demolishBuilding.mutate(building.id)
    setShowDemolishConfirm(false)
    onClose()
  }

  // ─── Stock state (retail only) ───
  const [stockQuantity, setStockQuantity] = useState('10')
  const [stockResourceId, setStockResourceId] = useState<number | null>(building.produces?.[0] ?? null)
  const [shelfPrice, setShelfPrice] = useState('')

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between p-3 border-b border-amber-200/60">
        <div className="flex items-center gap-2">
          <Icon name="icon_level_badge_v1" className="w-6 h-6" />
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-sm font-bold text-amber-900">{building.name ?? `${t('building.building')} #${building.kind}`}</h2>
              {building.isRetail && <span className="text-[9px] bg-blue-100 text-blue-700 px-1.5 py-0.5 rounded-full font-bold">{t('building.retail')}</span>}
            </div>
            <span className="text-[10px] text-amber-600 font-semibold">{t('building.level', { level: building.level })}</span>
          </div>
        </div>
        <button onClick={onClose} className="w-7 h-7 flex items-center justify-center rounded-md hover:bg-amber-200/60 text-amber-700 transition-colors">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      {/* Building Preview */}
      <div className="flex items-center justify-center py-4 bg-amber-100/30 border-b border-amber-200/40 min-h-[100px]">
        <img src={getBuildingPreview(building.kind)} alt={building.name ?? t('building.building')} className="max-h-[90px] object-contain" />
      </div>

      {/* ────── RETAIL BUILDING: Sell Tab ────── */}
      {building.isRetail ? <RetailSellSection
          building={building}
          stockShelf={stockShelf}
          unstockShelf={unstockShelf}
          setShelfPriceApi={setShelfPriceMut}
          stockQuantity={stockQuantity}
          setStockQuantity={setStockQuantity}
          stockResourceId={stockResourceId}
          setStockResourceId={setStockResourceId}
          shelfPrice={shelfPrice}
          setShelfPrice={setShelfPrice}
          t={t}
        />
      : <>
      {/* ────── PRODUCTION BUILDING: Production UI ────── */}

      {/* Produces */}
      <div className="px-3 py-2 border-b border-amber-200/40 flex items-center gap-2">
        <Icon name="icon_factory_v1" className="w-5 h-5" />
        <div>
          <span className="text-[10px] text-amber-600 uppercase tracking-wider">{t('building.produces')}</span>
          <span className="text-xs font-bold text-amber-900">
            {optionsData.length > 0 ? sortedOptions.map((o) => resourceName(o.resourceId)).slice(0, 2).join(', ') : t('building.noProductionOptions')}
          </span>
        </div>
      </div>

      {/* Production Recipe */}
      {optionsData.length > 0 && (
        <div className="px-3 py-3 border-b border-amber-200/40">
          <div className="text-[10px] text-amber-600 uppercase tracking-wider mb-2">{t('building.productionRecipe')}</div>
          <div className="grid grid-cols-2 gap-1.5">
            {sortedOptions.map((option) => (
              <button
                key={option.resourceId}
                onClick={() => setSelectedResourceId(option.resourceId)}
                className={`flex items-center gap-1.5 rounded-md border px-2 py-1.5 text-left text-[10px] font-semibold transition-colors ${
                  selectedOption?.resourceId === option.resourceId
                    ? 'border-amber-700 bg-amber-100 text-amber-950'
                    : 'border-amber-200 bg-white/60 text-amber-800 hover:bg-amber-50'
                }`}
              >
                <img src={resourceIcon(option.resourceId)} alt="" className="h-5 w-5 object-contain" />
                <span className="truncate">{resourceName(option.resourceId)}</span>
                {building.starterProduces?.includes(option.resourceId) && (
                  <span className="ml-auto rounded bg-green-100 px-1 text-[9px] text-green-700">{t('building.starter')}</span>
                )}
              </button>
            ))}
          </div>
          {selectedOption?.recipe && selectedOption.recipe.length > 0 && (
            <div className="mt-2 rounded-md bg-white/60 px-2 py-1.5 text-[10px] text-amber-700">
              {t('building.needs')} {selectedOption.recipe.map((r) => `${r.quantity} ${resourceName(r.resourceId)}`).join(', ')}
            </div>
          )}
          <input
            type="number"
            min="1"
            max={maxAmount}
            value={amount}
            onChange={(e) => {
              const next = Math.max(1, Math.min(maxAmount, parseInt(e.target.value, 10) || 1))
              setAmount(String(next))
            }}
            className={`${shouldGuideProductionStart ? 'tutorial-production-amount' : ''} mt-2 w-full rounded-md border border-amber-300 bg-white px-2 py-1.5 text-xs text-amber-900`}
          />
          {selectedOption && (
            <div className="mt-2 grid grid-cols-3 gap-1.5 text-[10px] text-amber-700">
              <div className="rounded bg-white/60 px-2 py-1">
                <div className="uppercase tracking-wider text-amber-500">{t('building.rate')}</div>
                <div className="font-bold text-amber-900">{selectedRate.toLocaleString()} {t('building.perHour')}</div>
              </div>
              <div className="rounded bg-white/60 px-2 py-1">
                <div className="uppercase tracking-wider text-amber-500">{t('building.time')}</div>
                <div className="font-bold text-amber-900">{estimatedDuration}</div>
              </div>
              <div className="rounded bg-white/60 px-2 py-1">
                <div className="uppercase tracking-wider text-amber-500">{t('building.max48h')}</div>
                <div className="font-bold text-amber-900">{maxAmount.toLocaleString()}</div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Progress */}
      <div className="px-3 py-3 border-b border-amber-200/40">
        <div className="text-[10px] text-amber-600 uppercase tracking-wider mb-1 flex items-center gap-1">
          <Icon name="icon_timer_v1" className="w-4 h-4" />
          {t('building.productionProgress')}
        </div>
        {runningJob ? (
          <>
            <div className="flex items-center justify-between text-xs text-amber-800 mb-1">
              <span>{t('building.inProgress')}</span>
              <span className="text-amber-500 tabular-nums">
                {countdownDisplay(runningJob.completesAt, now)}
              </span>
            </div>
            <div className="h-2.5 bg-amber-200/60 rounded-full overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-green-500 to-green-400 rounded-full transition-all duration-1000"
                style={{ width: `${progressPct(runningJob.startedAt, runningJob.completesAt, now)}%` }}
              />
            </div>
          </>
        ) : collectableJob ? (
          <div className="rounded-md bg-green-50 px-2 py-2 text-xs font-semibold text-green-700">
            {t('building.readyToCollect', { count: collectableAmount(collectableJob, now) })}
          </div>
        ) : (
          <div className="text-xs text-amber-500 italic">{t('building.notProducing')}</div>
        )}
      </div>

      {/* Output */}
      <div className="px-3 py-2 border-b border-amber-200/40 flex items-center gap-2">
        <Icon name="icon_warehouse_v1" className="w-5 h-5" />
        <div>
          <div className="text-[10px] text-amber-600 uppercase tracking-wider">{t('building.output')}</div>
          <div className="text-xs font-semibold text-amber-900">
            {buildingJobs.reduce((s, j) => s + j.amount, 0)} {t('building.units')}
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="flex gap-2 px-3 py-3 border-b border-amber-200/40">
        {collectableJob ? (
          <button onClick={handleCollect} className="tutorial-collect-production flex-1 py-2 bg-green-600 hover:bg-green-700 text-white text-xs font-bold rounded-md transition-colors btn-collect-pulse flex items-center justify-center gap-1">
            <Icon name="icon_collect_v1" className="w-4 h-4" />
            {t('building.collect')} {collectableAmount(collectableJob, now).toLocaleString()}
          </button>
        ) : (
          <button
            onClick={handleStart}
            disabled={startProd.isPending || !selectedOption}
            className={`${shouldGuideProductionStart ? 'tutorial-start-production' : ''} flex-1 py-2 bg-cyan-700 hover:bg-cyan-800 disabled:bg-cyan-900/50 text-white text-xs font-bold rounded-md transition-colors`}
          >
            {startProd.isPending ? t('building.starting') : t('building.start')}
          </button>
        )}
      </div>
      {startProd.error instanceof Error && (
        <div className="mx-3 mb-3 rounded-md border border-red-200 bg-red-50 px-2 py-1.5 text-[10px] font-semibold text-red-700">
          {startProd.error.message}
        </div>
      )}

      {/* Active Jobs */}
      <div className="flex-1 px-3 py-2 overflow-y-auto">
        <div className="text-[10px] font-semibold text-amber-700 uppercase tracking-wider mb-1">
          {t('building.activeJobs', { count: buildingJobs.length })}
        </div>
        <div className="space-y-1.5">
          {buildingJobs.map((job) => (
            <div key={job.id} className="flex items-center gap-2 py-1 px-2 rounded bg-white/40 text-xs">
              <span className={`w-1.5 h-1.5 rounded-full ${collectableAmount(job, now) > 0 ? 'bg-green-500' : 'bg-blue-500'}`} />
              <span className="text-amber-800 truncate">{resourceName(job.resourceId)} x{job.amount}</span>
              <span className="ml-auto text-[10px] text-amber-500">
                {(job.claimedAmount ?? 0).toLocaleString()} / {job.amount.toLocaleString()}
              </span>
            </div>
          ))}
          {buildingJobs.length === 0 && (
            <div className="text-[10px] text-amber-400 italic py-2">{t('building.noActiveJobs')}</div>
          )}
        </div>
        <button
          onClick={handleViewJobs}
          className="w-full mt-2 py-1.5 bg-amber-200/50 hover:bg-amber-300/50 text-amber-800 text-xs font-semibold rounded-md transition-colors flex items-center justify-center gap-1"
        >
          <Icon name="icon_timer_v1" className="w-4 h-4" />
          {t('building.viewAllJobs')}
        </button>
      </div>
      </>}

      {/* Upgrade */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-amber-200/40">
        <div className="flex items-center gap-2">
          <Icon name="icon_upgrade_v1" className="w-5 h-5" />
          <div>
            <div className="text-[10px] text-amber-600 uppercase tracking-wider">{t('building.upgrade')}</div>
            <div className="text-[10px] text-amber-500">{t('building.increaseOutput')}</div>
          </div>
        </div>
        <button
          onClick={handleUpgrade}
          disabled={upgradeBuilding.isPending}
          className="px-3 py-1.5 bg-amber-200/80 hover:bg-amber-300/80 disabled:bg-amber-100 disabled:text-amber-500 text-amber-900 text-xs font-bold rounded-md transition-colors"
        >
          {upgradeBuilding.isPending ? '...' : '$' + upgradeCost.toLocaleString()}
        </button>
      </div>
      {upgradeBuilding.error instanceof Error && (
        <div className="mx-3 mb-3 rounded-md border border-red-200 bg-red-50 px-2 py-1.5 text-[10px] font-semibold text-red-700">
          {upgradeBuilding.error.message}
        </div>
      )}

      {/* Move */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-amber-200/40">
        <div className="flex items-center gap-2">
          <svg className="w-5 h-5 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
          </svg>
          <div>
            <div className="text-[10px] text-amber-600 uppercase tracking-wider">{t('building.move')}</div>
            <div className="text-[10px] text-amber-500">{t('building.clickMapToReposition')}</div>
          </div>
        </div>
        <button
          onClick={handleMove}
          disabled={moveBuilding.isPending || !hasFreeSlots}
          className="px-3 py-1.5 bg-blue-100 hover:bg-blue-200 disabled:bg-blue-50 text-blue-800 text-xs font-bold rounded-md transition-colors disabled:opacity-50"
        >
          {moveBuilding.isPending ? t('building.moving') : t('building.move')}
        </button>
      </div>

      {/* Demolish */}
      {!showDemolishConfirm ? (
        <div className="flex items-center justify-between px-3 py-2 border-b border-amber-200/40">
          <div className="flex items-center gap-2">
            <svg className="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            <div>
              <div className="text-[10px] text-red-600 uppercase tracking-wider">{t('building.demolish')}</div>
              <div className="text-[10px] text-red-400">{t('building.refund50', { amount: Math.floor((building.baseCost ?? 10000) * 0.5).toLocaleString() })}</div>
            </div>
          </div>
          <button onClick={() => setShowDemolishConfirm(true)} className="px-3 py-1.5 bg-red-100 hover:bg-red-200 text-red-700 text-xs font-bold rounded-md transition-colors">
            {t('building.demolish')}
          </button>
        </div>
      ) : (
        <div className="px-3 py-2 border-b border-red-300 bg-red-50 space-y-2">
          <p className="text-[10px] font-semibold text-red-700">
            {t('building.demolishWarning', { name: building.name, amount: Math.floor((building.baseCost ?? 10000) * 0.5).toLocaleString() })}
          </p>
          <div className="flex gap-2">
            <button onClick={handleDemolish} disabled={demolishBuilding.isPending} className="flex-1 py-1.5 bg-red-600 hover:bg-red-700 disabled:bg-red-400 text-white text-xs font-bold rounded-md">
              {demolishBuilding.isPending ? t('building.demolishing') : t('building.confirmDemolish')}
            </button>
            <button onClick={() => setShowDemolishConfirm(false)} className="flex-1 py-1.5 bg-amber-200 hover:bg-amber-300 text-amber-900 text-xs font-bold rounded-md">
              {t('common.cancel')}
            </button>
          </div>
        </div>
      )}

      {/* Stash */}
      {building.placed !== false && (
        <div className="flex items-center justify-between px-3 py-2 border-b border-blue-200/40">
          <div className="flex items-center gap-2">
            <svg className="w-5 h-5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
            </svg>
            <div>
              <div className="text-[10px] text-slate-600 uppercase tracking-wider">{t('building.stash')}</div>
              <div className="text-[10px] text-slate-400">{t('building.returnToInventory')}</div>
            </div>
          </div>
          <button
            onClick={() => stashBuilding.mutate(building.id)}
            disabled={stashBuilding.isPending}
            className="px-3 py-1.5 bg-slate-100 hover:bg-slate-200 active:bg-slate-300 text-slate-700 text-xs font-bold rounded-md transition-colors disabled:opacity-50"
          >
            {stashBuilding.isPending ? t('building.stashing') : t('building.stash')}
          </button>
        </div>
      )}
    </div>
  )
}

/* ─── Retail Sell Section ─── */
function RetailSellSection({
  building, stockShelf, unstockShelf, setShelfPriceApi,
  stockQuantity, setStockQuantity, stockResourceId, setStockResourceId,
  shelfPrice, setShelfPrice, t,
}: {
  building: Building
  stockShelf: ReturnType<typeof useStockShelf>
  unstockShelf: ReturnType<typeof useUnstockShelf>
  setShelfPriceApi: ReturnType<typeof useSetShelfPrice>
  stockQuantity: string
  setStockQuantity: (v: string) => void
  stockResourceId: number | null
  setStockResourceId: (v: number | null) => void
  shelfPrice: string
  setShelfPrice: (v: string) => void
  t: (key: string) => string
}) {
  const sellableResources = building.produces ?? []
  const shelves = building.shelves ?? []
  const totalRevenue = shelves.reduce((s, sh) => s + (sh.revenue ?? 0), 0)

  return (
    <div className="flex-1 overflow-y-auto">
      {/* Sells label */}
      <div className="px-3 py-2 border-b border-amber-200/40 flex items-center gap-2">
        <Icon name="icon_tag_v1" className="w-5 h-5" />
        <div>
          <span className="text-[10px] text-blue-600 uppercase tracking-wider">🏪 {t('building.sells')}</span>
          <span className="text-xs font-bold text-amber-900">
            {sellableResources.map((r) => resourceName(r)).join(', ')}
          </span>
        </div>
      </div>

      {/* Shelves */}
      <div className="px-3 py-3 border-b border-amber-200/40">
        <div className="text-[10px] text-amber-600 uppercase tracking-wider mb-2">{t('building.shelves')}</div>
        {shelves.length === 0 ? (
          <div className="text-[10px] text-amber-400 italic">{t('building.noShelves')}</div>
        ) : (
          <div className="space-y-2">
            {shelves.map((sh) => (
              <div key={sh.resourceId} className="rounded-md border border-amber-200 bg-white/60 p-2">
                <div className="flex items-center justify-between mb-1">
                  <div className="flex items-center gap-1.5">
                    <img src={resourceIcon(sh.resourceId)} alt="" className="h-4 w-4 object-contain" />
                    <span className="text-xs font-bold text-amber-900">{resourceName(sh.resourceId)}</span>
                  </div>
                  <span className="text-[10px] text-amber-500">{sh.quantity} / {sh.maxQty}</span>
                </div>
                <div className="flex items-center gap-2 text-[10px]">
                  <span className="text-amber-700">{t('building.price')}: ${sh.price.toFixed(2)}</span>
                  {sh.priceLock && <span className="text-amber-400">🔒</span>}
                  {sh.revenue > 0 && (
                    <span className="ml-auto text-green-600">{t('building.revenue')}: ${sh.revenue.toFixed(2)}</span>
                  )}
                </div>
                <div className="flex gap-1.5 mt-2">
                  <button
                    onClick={() => {
                      unstockShelf.mutate({ buildingId: building.id, resourceId: sh.resourceId, quantity: sh.quantity })
                    }}
                    className="flex-1 py-1 bg-amber-200/70 hover:bg-amber-300/70 text-amber-900 text-[10px] font-bold rounded transition-colors"
                  >
                    {t('building.unstock')}
                  </button>
                  <button
                    onClick={() => {
                      const price = prompt(t('building.setPricePrompt'), String(sh.price))
                      if (price) {
                        const numPrice = parseFloat(price)
                        if (!isNaN(numPrice) && numPrice > 0) {
                          setShelfPriceApi.mutate({ buildingId: building.id, resourceId: sh.resourceId, price: numPrice, lock: true })
                        }
                      }
                    }}
                    className="flex-1 py-1 bg-blue-100 hover:bg-blue-200 text-blue-800 text-[10px] font-bold rounded transition-colors"
                  >
                    {t('building.setPrice')}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Stock form */}
      <div className="px-3 py-3 border-b border-amber-200/40">
        <div className="text-[10px] text-amber-600 uppercase tracking-wider mb-2">{t('building.stockItems')}</div>
        <div className="flex gap-2 mb-2">
          <select
            value={stockResourceId ?? ''}
            onChange={(e) => {
              const rid = parseInt(e.target.value, 10)
              if (!isNaN(rid)) setStockResourceId(rid)
            }}
            className="flex-1 rounded border border-amber-300 bg-white px-2 py-1 text-[10px] text-amber-900"
          >
            {sellableResources.map((rid) => (
              <option key={rid} value={rid}>{resourceName(rid)}</option>
            ))}
          </select>
          <input
            type="number"
            min="1"
            value={stockQuantity}
            onChange={(e) => setStockQuantity(e.target.value)}
            className="w-20 rounded border border-amber-300 bg-white px-2 py-1 text-[10px] text-amber-900"
          />
        </div>
        <div className="flex gap-2">
          <input
            type="number"
            step="0.01"
            min="0"
            placeholder={t('building.priceOptional')}
            value={shelfPrice}
            onChange={(e) => setShelfPrice(e.target.value)}
            className="flex-1 rounded border border-amber-300 bg-white px-2 py-1 text-[10px] text-amber-900"
          />
          <button
            onClick={() => {
              if (stockResourceId === null) return
              const price = shelfPrice ? parseFloat(shelfPrice) : undefined
              stockShelf.mutate({
                buildingId: building.id,
                resourceId: stockResourceId,
                quantity: parseInt(stockQuantity, 10) || 10,
                price,
              })
            }}
            disabled={stockShelf.isPending || stockResourceId === null}
            className="px-3 py-1 bg-cyan-700 hover:bg-cyan-800 disabled:bg-cyan-900/50 text-white text-[10px] font-bold rounded transition-colors"
          >
            {stockShelf.isPending ? t('building.stocking') : t('building.stock')}
          </button>
        </div>
      </div>

      {/* Revenue */}
      {totalRevenue > 0 && (
        <div className="px-3 py-2 border-b border-amber-200/40 flex items-center justify-between">
          <span className="text-[10px] text-green-700 uppercase tracking-wider font-bold">{t('building.revenue')}</span>
          <span className="text-xs font-bold text-green-800">+${totalRevenue.toFixed(2)}</span>
        </div>
      )}
    </div>
  )
}
