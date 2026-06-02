import { useMemo, useState } from 'react'
import { useProductionJobs, useStartProduction, useClaimProduction, useProductionOptions } from '@/api/production.api'
import type { Building } from '@/game/types'
import { Icon } from '@/features/ui/Icon'
import { resourceIcon } from '@/game/resources'

/** Map building kind to the appropriate asset image */
const BUILDING_PREVIEW: Record<number, string> = {
  1: '/assets/buildings/grain_plot_lv1_idle_trimmed.png',
  2: '/assets/buildings/mill_house_lv1_idle_trimmed.png',
  3: '/assets/buildings/bakery_shop_lv1_idle_trimmed.png',
  4: '/assets/buildings/meal_kiosk_lv1_idle_trimmed.png',
}

function getBuildingPreview(kind: number): string {
  return BUILDING_PREVIEW[kind] ?? '/assets/buildings/bakery_shop_lv1_idle_trimmed.png'
}

interface BuildingCardProps {
  building: Building
  onClose: () => void
}

export function BuildingCard({ building, onClose }: BuildingCardProps) {
  const [selectedResourceId, setSelectedResourceId] = useState<number | null>(null)
  const [amount, setAmount] = useState('10')
  const { data: jobsData } = useProductionJobs()
  const { data: optionsData = [] } = useProductionOptions(building.id)
  const buildingJobs = (jobsData ?? []).filter((j) => j.buildingId === building.id)
  const startProd = useStartProduction()
  const claimProd = useClaimProduction()
  const selectedOption = useMemo(() => {
    if (optionsData.length === 0) return null
    return optionsData.find((option) => option.resourceId === selectedResourceId) ?? optionsData[0]
  }, [optionsData, selectedResourceId])
  const sortedOptions = useMemo(() => {
    const starter = new Set(building.starterProduces ?? [])
    return [...optionsData].sort((a, b) => {
      const ar = starter.has(a.resourceId) ? 0 : 1
      const br = starter.has(b.resourceId) ? 0 : 1
      return ar - br
    })
  }, [building.starterProduces, optionsData])

  const runningJob = buildingJobs.find((j) => j.status === 'running')
  const readyJob = buildingJobs.find((j) => j.status === 'ready')

  const handleStart = () => {
    if (!selectedOption) return
    startProd.mutate({
      buildingId: building.id,
      kind: selectedOption.resourceId,
      amount: Math.max(1, parseInt(amount, 10) || 1),
    })
  }

  const handleCollect = () => {
    if (readyJob) {
      claimProd.mutate(readyJob.id)
    }
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
        <button
          onClick={onClose}
          className="w-7 h-7 flex items-center justify-center rounded-md hover:bg-amber-200/60 text-amber-700 transition-colors"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      {/* Building Preview - actual sprite */}
      <div className="flex items-center justify-center py-4 bg-amber-100/30 border-b border-amber-200/40 min-h-[100px]">
        <img
          src={getBuildingPreview(building.kind)}
          alt={building.name ?? 'Building'}
          className="max-h-[90px] object-contain"
        />
      </div>

      {/* Produces */}
      <div className="px-3 py-2 border-b border-amber-200/40 flex items-center gap-2">
        <Icon name="icon_factory_v1" className="w-5 h-5" />
        <div>
          <span className="text-[10px] text-amber-600 uppercase tracking-wider">Produces </span>
          <span className="text-xs font-bold text-amber-900">
            {optionsData.length > 0
              ? sortedOptions.map((option) => option.name).slice(0, 2).join(', ')
              : 'No production options'}
          </span>
        </div>
      </div>

      {optionsData.length > 0 && (
        <div className="px-3 py-3 border-b border-amber-200/40">
          <div className="text-[10px] text-amber-600 uppercase tracking-wider mb-2">
            Production Recipe
          </div>
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
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            className="mt-2 w-full rounded-md border border-amber-300 bg-white px-2 py-1.5 text-xs text-amber-900"
          />
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
                {runningJob.completesAt ? new Date(runningJob.completesAt).toLocaleTimeString() : '...'}
              </span>
            </div>
            <div className="h-2.5 bg-amber-200/60 rounded-full overflow-hidden">
              <div className="h-full w-2/5 bg-gradient-to-r from-green-500 to-green-400 rounded-full" />
            </div>
          </>
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
        {readyJob ? (
          <button
            onClick={handleCollect}
            className="flex-1 py-2 bg-green-600 hover:bg-green-700 text-white text-xs font-bold rounded-md transition-colors btn-collect-pulse flex items-center justify-center gap-1"
          >
            <Icon name="icon_collect_v1" className="w-4 h-4" />
            Collect
          </button>
        ) : (
          <button
            onClick={handleStart}
            disabled={startProd.isPending || !selectedOption}
            className="flex-1 py-2 bg-cyan-700 hover:bg-cyan-800 disabled:bg-cyan-900/50 text-white text-xs font-bold rounded-md transition-colors"
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
        <button className="px-3 py-1.5 bg-amber-200/80 hover:bg-amber-300/80 text-amber-900 text-xs font-bold rounded-md transition-colors">
          ${(building.level + 1) * 10000}
        </button>
      </div>

      {/* Active Jobs */}
      <div className="flex-1 px-3 py-2 overflow-y-auto">
        <div className="text-[10px] font-semibold text-amber-700 uppercase tracking-wider mb-1">
          Active Jobs ({buildingJobs.length})
        </div>
        <div className="space-y-1.5">
          {buildingJobs.map((job) => (
            <div key={job.id} className="flex items-center gap-2 py-1 px-2 rounded bg-white/40 text-xs">
              <span className={`w-1.5 h-1.5 rounded-full ${job.status === 'ready' ? 'bg-green-500' : 'bg-blue-500'}`} />
              <span className="text-amber-800 truncate">#{job.resourceId} x{job.amount}</span>
              <span className="ml-auto text-[10px] text-amber-500">{job.status}</span>
            </div>
          ))}
          {buildingJobs.length === 0 && (
            <div className="text-[10px] text-amber-400 italic py-2">No active jobs</div>
          )}
        </div>
        <button className="w-full mt-2 py-1.5 bg-amber-200/50 hover:bg-amber-300/50 text-amber-800 text-xs font-semibold rounded-md transition-colors flex items-center justify-center gap-1">
          <Icon name="icon_timer_v1" className="w-4 h-4" />
          View All Jobs
        </button>
      </div>
    </div>
  )
}
