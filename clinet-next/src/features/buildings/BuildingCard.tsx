import { useState, useEffect } from 'react'
import { useProductionJobs, useStartProduction, useClaimProduction } from '@/api/hooks/production.hooks'
import { useDemolishBuilding, useUpgradeBuilding } from '@/api/hooks/buildings.hooks'
import { useUIStore } from '@/store/ui.store'
import type { Building, ProductionJob } from '@/game/types'
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
  if (!completesAt) return '--'
  const diff = new Date(completesAt).getTime() - now
  if (diff <= 0) return 'Ready'
  const mins = Math.ceil(diff / 60000)
  if (mins >= 60) return `${Math.floor(mins / 60)}h ${mins % 60}m`
  return `${mins}m`
}

function progressPct(startedAt: string | undefined, completesAt: string | undefined, now: number): number {
  if (!startedAt || !completesAt) return 0
  const total = new Date(completesAt).getTime() - new Date(startedAt).getTime()
  if (total <= 0) return 100
  return Math.min(100, Math.round(((now - new Date(startedAt).getTime()) / total) * 100))
}

function accruedAmount(job: ProductionJob, now: number): number {
  if (!job.startedAt || !job.completesAt) return 0
  const total = new Date(job.completesAt).getTime() - new Date(job.startedAt).getTime()
  if (total <= 0) return job.amount
  const elapsed = now - new Date(job.startedAt).getTime()
  return Math.min(job.amount, Math.max(0, Math.floor((elapsed / total) * job.amount)))
}

interface BuildingCardProps {
  building: Building
  onClose: () => void
}

export function BuildingCard({ building, onClose }: BuildingCardProps) {
  const [now, setNow] = useState(() => Date.now())
  const [amount, setAmount] = useState(1)
  const { data: allJobs } = useProductionJobs()
  const startProduction = useStartProduction()
  const claimProduction = useClaimProduction()
  const demolishBuilding = useDemolishBuilding()
  const upgradeBuilding = useUpgradeBuilding()
  const startBuildingMove = useUIStore((s) => s.startBuildingMove)

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(id)
  }, [])

  const jobs = (allJobs ?? []).filter((j) => j.buildingId === building.id)
  const activeJob = jobs.find((j) => j.status !== 'claimed' && j.claimableAmount === 0)
  const collectableJobs = jobs.filter((j) => (j.claimableAmount ?? 0) > 0)
  const collectableTotal = collectableJobs.reduce((sum, j) => sum + (j.claimableAmount ?? 0), 0)

  const rate = outputPerHour(building.produces?.[0] ? 60 : undefined, building.level)
  const maxAmount = maxAmountFor48Hours(rate)
  const handleStartProduction = () => {
    startProduction.mutate({
      buildingId: building.id,
      kind: (building.produces?.[0] ?? building.kind),
      amount: amount,
      estimatedSecondsToFinish: Math.max(30, Math.ceil((amount / Math.max(1, rate)) * 3600)),
    })
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between p-3 border-b border-amber-200/60">
        <div className="flex items-center gap-2">
          <img src={getBuildingPreview(building.kind)} alt="" className="w-10 h-10 rounded-lg object-cover" />
          <div>
            <h2 className="text-sm font-bold text-amber-900">{building.name ?? `Building ${building.kind}`}</h2>
            <p className="text-[10px] text-amber-600">Lv.{building.level}</p>
          </div>
        </div>
        <button onClick={onClose} className="p-1 text-amber-400 hover:text-amber-600">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        {/* Active job progress */}
        {activeJob && (
          <div className="bg-white/60 rounded-lg p-3 border border-amber-200/40">
            <div className="text-xs font-semibold text-amber-900 mb-2">
              Producing #{activeJob.resourceId}
            </div>
            <div className="h-2 bg-amber-200/60 rounded-full overflow-hidden mb-1">
              <div className="h-full bg-blue-500 rounded-full" style={{ width: `${progressPct(activeJob.startedAt, activeJob.completesAt, now)}%` }} />
            </div>
            <div className="flex justify-between text-[10px] text-amber-600">
              <span>{accruedAmount(activeJob, now)}/{activeJob.amount}</span>
              <span>{countdownDisplay(activeJob.completesAt, now)}</span>
            </div>
          </div>
        )}

        {/* Collectable jobs */}
        {collectableJobs.length > 0 && (
          <div className="space-y-1">
            <h3 className="text-xs font-semibold text-green-700">Ready to Collect ({collectableTotal})</h3>
            {collectableJobs.map((job) => (
              <button
                key={job.id}
                onClick={() => claimProduction.mutate(job.id)}
                className="w-full flex items-center gap-2 p-2 bg-green-50 rounded-lg border border-green-200 text-xs hover:bg-green-100"
              >
                <span className="w-2 h-2 rounded-full bg-green-500" />
                <span className="text-amber-900">#{job.resourceId}</span>
                <span className="text-amber-600">x{job.claimableAmount}</span>
                <span className="ml-auto text-green-600">Collect</span>
              </button>
            ))}
          </div>
        )}

        {/* Start production */}
        {!activeJob && (
          <div className="bg-white/60 rounded-lg p-3 border border-amber-200/40">
            <h3 className="text-xs font-semibold text-amber-900 mb-2">Start Production</h3>
            <div className="flex gap-2 mb-2">
              <input
                type="number"
                min={1}
                max={maxAmount}
                value={amount}
                onChange={(e) => setAmount(Math.min(maxAmount, Math.max(1, Number(e.target.value) || 1)))}
                className="w-20 px-2 py-1 text-xs border border-amber-200 rounded"
              />
              <span className="text-[10px] text-amber-600 self-center">
                {durationLabel(amount, Math.max(1, rate))}
              </span>
            </div>
            <button
              onClick={handleStartProduction}
              disabled={startProduction.isPending}
              className="w-full py-1.5 bg-amber-600 hover:bg-amber-700 text-white text-xs font-semibold rounded-md transition-colors disabled:opacity-50"
            >
              {startProduction.isPending ? 'Starting...' : 'Start'}
            </button>
          </div>
        )}

        {/* Actions */}
        <div className="flex gap-2">
          <button
            onClick={() => upgradeBuilding.mutate(building.id)}
            disabled={upgradeBuilding.isPending}
            className="flex-1 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-md transition-colors disabled:opacity-50"
          >
            Upgrade
          </button>
          <button
            onClick={() => startBuildingMove(building.id)}
            className="flex-1 py-2 bg-amber-100 hover:bg-amber-200 text-amber-900 text-xs font-semibold rounded-md transition-colors"
          >
            Move
          </button>
          <button
            onClick={() => { if (confirm('Demolish this building?')) demolishBuilding.mutate(building.id) }}
            disabled={demolishBuilding.isPending}
            className="flex-1 py-2 bg-red-100 hover:bg-red-200 text-red-700 text-xs font-semibold rounded-md transition-colors disabled:opacity-50"
          >
            Demolish
          </button>
        </div>
      </div>
    </div>
  )
}
