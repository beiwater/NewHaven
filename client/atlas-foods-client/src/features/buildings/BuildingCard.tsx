import { useMemo, useState, useEffect } from 'react'
import { useProductionJobs, useStartProduction, useClaimProduction, useProductionOptions } from '@/api/production.api'
import { useMoveBuilding, useDemolishBuilding, useUpgradeBuilding } from '@/api/buildings.api'
import { useUIStore } from '@/store/ui.store'
import type { Building, ProductionJob } from '@/game/types'
import { Icon } from '@/features/ui/Icon'
import { resourceIcon } from '@/game/resources'
import { buildingIcon } from '@/game/icons'

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
  if (!completesAt) return '...'
  const remaining = Math.max(0, new Date(completesAt).getTime() - now)
  if (remaining <= 0) return 'Ready!'
  const h = Math.floor(remaining / 3_600_000)
  const m = Math.floor((remaining % 3_600_000) / 60_000)
  const s = Math.floor((remaining % 60_000) / 1000)
  const parts = h > 0 ? [`${h}h`, `${String(m).padStart(2, '0')}m`, `${String(s).padStart(2, '0')}s`] : [`${m}m`, `${String(s).padStart(2, '0')}s`]
  return parts.join(' ')
}

function progressPct(startedAt: string | undefined, completesAt: string | undefined, now: number): number {
  if (!startedAt || !completesAt) return 0
  const total = new Date(completesAt).getTime() - new Date(startedAt).getTime()
  if (total <= 0) return 0
  const elapsed = now - new Date(startedAt).getTime()
  return Math.min(100, Math.round((elapsed / total) * 100))
}

function accruedAmount(job: ProductionJob, now: number): number {
  const start = new Date(job.startedAt).getTime()
  const end = new Date(job.completesAt).getTime()
  if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) return job.amount
  if (now >= end) return job.amount
  if (now <= start) return 0
  return Math.floor(((now - start) / (end - start)) * job.amount)
}

function collectableAmount(job: ProductionJob, now: number): number {
  if (job.status === 'claimed') return 0
  const claimed = job.claimedAmount ?? 0
  return Math.max(job.claimableAmount ?? 0, accruedAmount(job, now) - claimed, 0)
}

interface BuildingCardProps {
  building: Building
  onClose: () => void
}

export function BuildingCard({ building, onClose }: BuildingCardProps) {
  const [selectedResourceId, setSelectedResourceId] = useState<number | null>(null)
  const [amount, setAmount] = useState('10')
  const [showDemolishConfirm, setShowDemolishConfirm] = useState(false)
  const [now, setNow] = useState(Date.now())

  const { data: jobsData } = useProductionJobs()
  const { data: optionsData = [] } = useProductionOptions(building.id)
  const buildingJobs = (jobsData ?? []).filter((j) => j.buildingId === building.id)
  const startProd = useStartProduction()
  const claimProd = useClaimProduction()
  const moveBuilding = useMoveBuilding()
  const upgradeBuilding = useUpgradeBuilding()
  const demolishBuilding = useDemolishBuilding()
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
    startProd.mutate({
      buildingId: building.id,
      kind: selectedOption.resourceId,
      amount: safeAmount,
    })
  }

  const handleCollect = () => {
    if (collectableJob) claimProd.mutate(collectableJob.id)
  }

  const handleMove = () => {
    startBuildingMove(building.id)
  }

  const handleUpgrade = () => {
    upgradeBuilding.mutate(building.id)
  }

  const handleViewJobs = () => {
    setActiveView('production')
  }

  const handleDemolish = () => {
    demolishBuilding.mutate(building.id)
    setShowDemolishConfirm(false)
    onClose()
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between p-3 border-b border-amber-200/60">
        <div className="flex items-center gap-2">
          <Icon name="icon_level_badge_v1" className="w-6 h-6" />
          <div>
            <h2 className="text-sm font-bold text-amber-900">{building.name ?? `Building #${building.kind}`}</h2>
            <span className="text-[10px] text-amber-600 font-semibold">Level {building.level}</span>
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
        <img src={getBuildingPreview(building.kind)} alt={building.name ?? 'Building'} className="max-h-[90px] object-contain" />
      </div>

      {/* Produces */}
      <div className="px-3 py-2 border-b border-amber-200/40 flex items-center gap-2">
        <Icon name="icon_factory_v1" className="w-5 h-5" />
        <div>
          <span className="text-[10px] text-amber-600 uppercase tracking-wider">Produces </span>
          <span className="text-xs font-bold text-amber-900">
            {optionsData.length > 0 ? sortedOptions.map((o) => o.name).slice(0, 2).join(', ') : 'No production options'}
          </span>
        </div>
      </div>

      {/* Production Recipe */}
      {optionsData.length > 0 && (
        <div className="px-3 py-3 border-b border-amber-200/40">
          <div className="text-[10px] text-amber-600 uppercase tracking-wider mb-2">Production Recipe</div>
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
                <span className="truncate">{option.name}</span>
                {building.starterProduces?.includes(option.resourceId) && (
                  <span className="ml-auto rounded bg-green-100 px-1 text-[9px] text-green-700">start</span>
                )}
              </button>
            ))}
          </div>
          {selectedOption?.recipe && selectedOption.recipe.length > 0 && (
            <div className="mt-2 rounded-md bg-white/60 px-2 py-1.5 text-[10px] text-amber-700">
              Needs {selectedOption.recipe.map((r) => `${r.quantity} ${r.resourceName ?? `#${r.resourceId}`}`).join(', ')}
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
                <div className="uppercase tracking-wider text-amber-500">Rate</div>
                <div className="font-bold text-amber-900">{selectedRate.toLocaleString()} / h</div>
              </div>
              <div className="rounded bg-white/60 px-2 py-1">
                <div className="uppercase tracking-wider text-amber-500">Time</div>
                <div className="font-bold text-amber-900">{estimatedDuration}</div>
              </div>
              <div className="rounded bg-white/60 px-2 py-1">
                <div className="uppercase tracking-wider text-amber-500">48h Max</div>
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
          Production Progress
        </div>
        {runningJob ? (
          <>
            <div className="flex items-center justify-between text-xs text-amber-800 mb-1">
              <span>In progress</span>
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
            Ready to collect {collectableAmount(collectableJob, now).toLocaleString()} units
          </div>
        ) : (
          <div className="text-xs text-amber-500 italic">Not producing</div>
        )}
      </div>

      {/* Output */}
      <div className="px-3 py-2 border-b border-amber-200/40 flex items-center gap-2">
        <Icon name="icon_warehouse_v1" className="w-5 h-5" />
        <div>
          <div className="text-[10px] text-amber-600 uppercase tracking-wider">Output</div>
          <div className="text-xs font-semibold text-amber-900">
            {buildingJobs.reduce((s, j) => s + j.amount, 0)} units
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="flex gap-2 px-3 py-3 border-b border-amber-200/40">
        {collectableJob ? (
          <button onClick={handleCollect} className="tutorial-collect-production flex-1 py-2 bg-green-600 hover:bg-green-700 text-white text-xs font-bold rounded-md transition-colors btn-collect-pulse flex items-center justify-center gap-1">
            <Icon name="icon_collect_v1" className="w-4 h-4" />
            Collect {collectableAmount(collectableJob, now).toLocaleString()}
          </button>
        ) : (
          <button
            onClick={handleStart}
            disabled={startProd.isPending || !selectedOption}
            className={`${shouldGuideProductionStart ? 'tutorial-start-production' : ''} flex-1 py-2 bg-cyan-700 hover:bg-cyan-800 disabled:bg-cyan-900/50 text-white text-xs font-bold rounded-md transition-colors`}
          >
            {startProd.isPending ? 'Starting...' : 'Start'}
          </button>
        )}
      </div>
      {startProd.error instanceof Error && (
        <div className="mx-3 mb-3 rounded-md border border-red-200 bg-red-50 px-2 py-1.5 text-[10px] font-semibold text-red-700">
          {startProd.error.message}
        </div>
      )}

      {/* Upgrade */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-amber-200/40">
        <div className="flex items-center gap-2">
          <Icon name="icon_upgrade_v1" className="w-5 h-5" />
          <div>
            <div className="text-[10px] text-amber-600 uppercase tracking-wider">Upgrade</div>
            <div className="text-[10px] text-amber-500">Increase output capacity</div>
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

      {/* Move Building */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-amber-200/40">
        <div className="flex items-center gap-2">
          <svg className="w-5 h-5 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
          </svg>
          <div>
            <div className="text-[10px] text-amber-600 uppercase tracking-wider">Move</div>
            <div className="text-[10px] text-amber-500">Click map to reposition (free)</div>
          </div>
        </div>
        <button
          onClick={handleMove}
          disabled={moveBuilding.isPending}
          className="px-3 py-1.5 bg-blue-100 hover:bg-blue-200 disabled:bg-blue-50 text-blue-800 text-xs font-bold rounded-md transition-colors"
        >
          {moveBuilding.isPending ? 'Moving...' : 'Move'}
        </button>
      </div>

      {/* Demolish Building */}
      {!showDemolishConfirm ? (
        <div className="flex items-center justify-between px-3 py-2 border-b border-amber-200/40">
          <div className="flex items-center gap-2">
            <svg className="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            <div>
              <div className="text-[10px] text-red-600 uppercase tracking-wider">Demolish</div>
              <div className="text-[10px] text-red-400">50% refund (${Math.floor((building.baseCost ?? 10000) * 0.5).toLocaleString()})</div>
            </div>
          </div>
          <button onClick={() => setShowDemolishConfirm(true)} className="px-3 py-1.5 bg-red-100 hover:bg-red-200 text-red-700 text-xs font-bold rounded-md transition-colors">
            Demolish
          </button>
        </div>
      ) : (
        <div className="px-3 py-2 border-b border-red-300 bg-red-50 space-y-2">
          <p className="text-[10px] font-semibold text-red-700">
            This will permanently destroy {building.name ?? 'this building'}.
            You'll receive 50% refund (${Math.floor((building.baseCost ?? 10000) * 0.5).toLocaleString()}).
            Active production jobs will be lost.
          </p>
          <div className="flex gap-2">
            <button onClick={handleDemolish} disabled={demolishBuilding.isPending} className="flex-1 py-1.5 bg-red-600 hover:bg-red-700 disabled:bg-red-400 text-white text-xs font-bold rounded-md">
              {demolishBuilding.isPending ? 'Demolishing...' : 'Confirm Demolish'}
            </button>
            <button onClick={() => setShowDemolishConfirm(false)} className="flex-1 py-1.5 bg-amber-200 hover:bg-amber-300 text-amber-900 text-xs font-bold rounded-md">
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Active Jobs */}
      <div className="flex-1 px-3 py-2 overflow-y-auto">
        <div className="text-[10px] font-semibold text-amber-700 uppercase tracking-wider mb-1">
          Active Jobs ({buildingJobs.length})
        </div>
        <div className="space-y-1.5">
          {buildingJobs.map((job) => (
            <div key={job.id} className="flex items-center gap-2 py-1 px-2 rounded bg-white/40 text-xs">
              <span className={`w-1.5 h-1.5 rounded-full ${collectableAmount(job, now) > 0 ? 'bg-green-500' : 'bg-blue-500'}`} />
              <span className="text-amber-800 truncate">#{job.resourceId} x{job.amount}</span>
              <span className="ml-auto text-[10px] text-amber-500">
                {(job.claimedAmount ?? 0).toLocaleString()} / {job.amount.toLocaleString()}
              </span>
            </div>
          ))}
          {buildingJobs.length === 0 && (
            <div className="text-[10px] text-amber-400 italic py-2">No active jobs</div>
          )}
        </div>
        <button
          onClick={handleViewJobs}
          className="w-full mt-2 py-1.5 bg-amber-200/50 hover:bg-amber-300/50 text-amber-800 text-xs font-semibold rounded-md transition-colors flex items-center justify-center gap-1"
        >
          <Icon name="icon_timer_v1" className="w-4 h-4" />
          View All Jobs
        </button>
      </div>
    </div>
  )
}
