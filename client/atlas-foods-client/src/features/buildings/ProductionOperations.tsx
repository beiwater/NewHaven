import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useWarehouse } from '@/api/inventory.api'
import { useClaimProduction, useProductionJobs, useProductionOptions, useStartProduction } from '@/api/production.api'
import { audio } from '@/audio/AudioManager'
import { resourceIcon, resourceName } from '@/game/resources'
import type { Building, ProductionJob, ResourceDefinition } from '@/game/types'
import { useUIStore } from '@/store/ui.store'

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
  const totalSeconds = Math.ceil(remaining / 1000)
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60
  if (hours > 0) return `${hours}h ${minutes.toString().padStart(2, '0')}m`
  return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
}

function progressPercent(job: ProductionJob, now: number): number {
  const total = new Date(job.completesAt).getTime() - new Date(job.startedAt).getTime()
  if (total <= 0) return 100
  return Math.min(100, Math.max(0, ((now - new Date(job.startedAt).getTime()) / total) * 100))
}

function accruedAmount(job: ProductionJob, now: number): number {
  if (job.status === 'claimed') return job.amount
  const elapsed = (now - new Date(job.startedAt).getTime()) / 1000
  const duration = Math.max(1, (new Date(job.completesAt).getTime() - new Date(job.startedAt).getTime()) / 1000)
  return Math.min(job.amount, Math.floor((elapsed / duration) * job.amount))
}

function collectableAmount(job: ProductionJob, now: number): number {
  return Math.max(0, accruedAmount(job, now) - (job.claimedAmount ?? 0))
}

function ProductionOptionRow({
  building,
  option,
  currentStock,
  amount,
  setAmount,
  blocked,
  isStarting,
  guided,
  onStart,
}: {
  building: Building
  option: ResourceDefinition
  currentStock: number
  amount: string
  setAmount: (value: string) => void
  blocked: boolean
  isStarting: boolean
  guided: boolean
  onStart: (amount: number) => void
}) {
  const { t } = useTranslation()
  const rate = outputPerHour(option.producedPerHourRaw, building.level)
  const maxAmount = maxAmountFor48Hours(rate)
  const numericAmount = Math.max(1, Math.min(maxAmount, parseInt(amount, 10) || 1))
  const hours24 = Math.max(1, Math.min(maxAmount, Math.floor(rate * 24)))
  const ingredients = option.recipe ?? []

  return (
    <article className="rounded-2xl border border-amber-200/80 bg-white/70 p-4 shadow-sm transition-colors hover:border-amber-400/80">
      <div className="grid gap-4 lg:grid-cols-[minmax(190px,1.05fr)_minmax(180px,0.95fr)_minmax(230px,0.9fr)] lg:items-center">
        <div className="flex min-w-0 items-center gap-4">
          <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-amber-50 to-amber-200/70">
            <img src={resourceIcon(option.resourceId)} alt="" className="h-14 w-14 object-contain" />
          </div>
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="truncate text-base font-black text-amber-950">{resourceName(option.resourceId)}</h3>
              {building.starterProduces?.includes(option.resourceId) && (
                <span className="rounded-full bg-green-100 px-2 py-0.5 text-[9px] font-black uppercase tracking-wider text-green-700">
                  {t('building.starter')}
                </span>
              )}
            </div>
            <dl className="mt-2 space-y-1 text-[11px] text-amber-700">
              <div className="flex justify-between gap-3"><dt>{t('building.rate')}</dt><dd className="font-bold text-amber-950">{rate.toLocaleString()} {t('building.perHour')}</dd></div>
              <div className="flex justify-between gap-3"><dt>{t('building.currentStock')}</dt><dd className="font-bold text-amber-950">{currentStock.toLocaleString()}</dd></div>
              <div className="flex justify-between gap-3"><dt>{t('building.estimatedTime')}</dt><dd className="font-bold text-amber-950">{durationLabel(numericAmount, rate)}</dd></div>
            </dl>
          </div>
        </div>

        <div className="border-y border-amber-200/70 py-3 lg:border-x lg:border-y-0 lg:px-5 lg:py-2">
          <div className="mb-2 text-[10px] font-black uppercase tracking-[0.16em] text-amber-600">{t('building.requirements')}</div>
          {ingredients.length > 0 ? (
            <div className="flex flex-wrap gap-2">
              {ingredients.map((ingredient) => (
                <div key={ingredient.resourceId} className="flex items-center gap-1.5 rounded-lg bg-amber-50 px-2 py-1.5 text-xs font-bold text-amber-900">
                  <span>{ingredient.quantity}×</span>
                  <img src={resourceIcon(ingredient.resourceId)} alt="" className="h-6 w-6 object-contain" />
                  <span className="hidden xl:inline">{resourceName(ingredient.resourceId)}</span>
                </div>
              ))}
            </div>
          ) : (
            <div className="rounded-lg bg-green-50 px-3 py-2 text-xs font-semibold text-green-700">{t('building.noRequirements')}</div>
          )}
        </div>

        <div>
          <div className="mb-2 flex items-center justify-between text-[10px] font-black uppercase tracking-[0.16em] text-amber-600">
            <span>{t('building.quantity')}</span>
            <span>{t('building.max48h')}: {maxAmount.toLocaleString()}</span>
          </div>
          <div className="flex gap-2">
            <input
              aria-label={`${resourceName(option.resourceId)} ${t('building.quantity')}`}
              type="number"
              min="1"
              max={maxAmount}
              value={amount}
              onChange={(event) => setAmount(event.target.value)}
              className={`${guided ? 'tutorial-production-amount' : ''} min-w-0 flex-1 rounded-lg border border-amber-300 bg-white px-3 py-2 text-sm font-bold text-amber-950`}
            />
            <button type="button" onClick={() => setAmount(String(hours24))} className="rounded-lg border border-amber-300 bg-amber-50 px-2.5 text-[10px] font-black text-amber-800 hover:bg-amber-100">24H</button>
            <button type="button" onClick={() => setAmount(String(maxAmount))} className="rounded-lg border border-amber-300 bg-amber-50 px-2.5 text-[10px] font-black text-amber-800 hover:bg-amber-100">MAX</button>
          </div>
          <button
            type="button"
            onClick={() => onStart(numericAmount)}
            disabled={blocked || isStarting || rate <= 0}
            className={`${guided ? 'tutorial-start-production' : ''} mt-2 w-full rounded-lg bg-cyan-700 px-3 py-2.5 text-xs font-black text-white transition-colors hover:bg-cyan-800 disabled:cursor-not-allowed disabled:bg-slate-300`}
          >
            {isStarting ? t('building.starting') : blocked ? t('building.finishCurrentRun') : t('building.startThisRun')}
          </button>
        </div>
      </div>
    </article>
  )
}

export function ProductionOperations({ building }: { building: Building }) {
  const { t } = useTranslation()
  const { data: options = [] } = useProductionOptions(building.id)
  const { data: jobs = [] } = useProductionJobs()
  const { data: warehouse } = useWarehouse()
  const startProduction = useStartProduction()
  const claimProduction = useClaimProduction()
  const setActiveView = useUIStore((state) => state.setActiveView)
  const [amounts, setAmounts] = useState<Record<number, string>>({})
  const [now, setNow] = useState(() => Date.now())

  const buildingJobs = jobs.filter((job) => job.buildingId === building.id && job.status !== 'claimed')
  const runningJob = buildingJobs.find((job) => job.status === 'running')
  const runningJobId = runningJob?.id
  const collectableJob = buildingJobs.find((job) => collectableAmount(job, now) > 0)
  const blocked = buildingJobs.length > 0
  const sortedOptions = useMemo(() => {
    const starters = new Set(building.starterProduces ?? [])
    return [...options].sort((left, right) => Number(!starters.has(left.resourceId)) - Number(!starters.has(right.resourceId)))
  }, [building.starterProduces, options])

  useEffect(() => {
    if (!runningJobId) return
    const interval = window.setInterval(() => setNow(Date.now()), 1000)
    return () => window.clearInterval(interval)
  }, [runningJobId])

  const currentStock = (resourceId: number) => warehouse?.inventory
    .filter((item) => item.resourceId === resourceId && (item.quality ?? 0) === 0)
    .reduce((sum, item) => sum + item.quantity, 0) ?? 0

  return (
    <div className="min-h-0 flex-1 overflow-y-auto bg-gradient-to-br from-[#f8edd7] via-[#fffaf0] to-[#f2dcb5] p-4 sm:p-6">
      <div className="mx-auto max-w-4xl">
        <div className="mb-4 flex flex-wrap items-end justify-between gap-3">
          <div>
            <p className="text-[10px] font-black uppercase tracking-[0.22em] text-cyan-700">{t('building.operationCenter')}</p>
            <h2 className="text-xl font-black text-amber-950 sm:text-2xl">{t('building.productionCatalog')}</h2>
          </div>
          <button type="button" onClick={() => setActiveView('production')} className="rounded-full border border-amber-300 bg-white/70 px-3 py-1.5 text-[11px] font-bold text-amber-800 hover:bg-white">
            {t('building.viewAllJobs')} · {buildingJobs.length}
          </button>
        </div>

        {runningJob && (
          <div className="mb-4 rounded-2xl border border-cyan-200 bg-cyan-50/90 p-4">
            <div className="flex flex-wrap items-center gap-3">
              <img src={resourceIcon(runningJob.resourceId)} alt="" className="h-10 w-10 object-contain" />
              <div className="min-w-0 flex-1">
                <div className="flex justify-between gap-3 text-xs font-bold text-cyan-950">
                  <span>{resourceName(runningJob.resourceId)} × {runningJob.amount.toLocaleString()}</span>
                  <span className="tabular-nums">{countdownDisplay(runningJob.completesAt, now)}</span>
                </div>
                <div className="mt-2 h-2 overflow-hidden rounded-full bg-cyan-200/70">
                  <div className="h-full rounded-full bg-gradient-to-r from-cyan-600 to-green-500 transition-[width] duration-1000" style={{ width: `${progressPercent(runningJob, now)}%` }} />
                </div>
              </div>
              {collectableJob && (
                <button
                  type="button"
                  onClick={() => {
                    audio.playSfx('resource_pickup')
                    claimProduction.mutate(collectableJob.id)
                  }}
                  className="tutorial-collect-production btn-collect-pulse rounded-lg bg-green-600 px-4 py-2 text-xs font-black text-white hover:bg-green-700"
                >
                  {t('building.collect')} {collectableAmount(collectableJob, now).toLocaleString()}
                </button>
              )}
            </div>
          </div>
        )}

        {!runningJob && collectableJob && (
          <div className="mb-4 flex items-center justify-between gap-3 rounded-2xl border border-green-200 bg-green-50 p-4 text-sm font-bold text-green-800">
            <span>{t('building.readyToCollect', { count: collectableAmount(collectableJob, now) })}</span>
            <button type="button" onClick={() => claimProduction.mutate(collectableJob.id)} className="rounded-lg bg-green-600 px-4 py-2 text-xs font-black text-white hover:bg-green-700">{t('building.collect')}</button>
          </div>
        )}

        {blocked && (
          <div className="mb-4 rounded-xl border border-amber-300 bg-amber-50 px-3 py-2 text-xs font-semibold text-amber-800">
            {t('building.oneRunAtATime')}
          </div>
        )}

        <div className="space-y-3">
          {sortedOptions.map((option, index) => {
            const rate = outputPerHour(option.producedPerHourRaw, building.level)
            const defaultAmount = String(Math.min(10, maxAmountFor48Hours(rate)))
            return (
              <ProductionOptionRow
                key={option.resourceId}
                building={building}
                option={option}
                currentStock={currentStock(option.resourceId)}
                amount={amounts[option.resourceId] ?? defaultAmount}
                setAmount={(value) => setAmounts((current) => ({ ...current, [option.resourceId]: value }))}
                blocked={blocked}
                isStarting={startProduction.isPending}
                guided={index === 0 && !blocked}
                onStart={(amount) => {
                  audio.playSfx('build_confirm')
                  startProduction.mutate({ buildingId: building.id, kind: option.resourceId, amount })
                }}
              />
            )
          })}
          {sortedOptions.length === 0 && (
            <div className="rounded-2xl border border-dashed border-amber-300 bg-white/50 p-8 text-center text-sm font-semibold text-amber-500">
              {t('building.noProductionOptions')}
            </div>
          )}
        </div>

        {startProduction.error instanceof Error && (
          <div className="mt-4 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-xs font-semibold text-red-700">{startProduction.error.message}</div>
        )}
      </div>
    </div>
  )
}
