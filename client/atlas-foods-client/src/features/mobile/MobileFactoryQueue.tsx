import { useProductionJobs, useProductionQueue, useCancelJob } from '@/api/production.api'
import { resourceIcon } from '@/game/resources'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { resourceName } from '@/game/resources'


export function MobileFactoryQueue() {
  const { data: jobsData } = useProductionJobs()
  const { data: queueData } = useProductionQueue()
  const cancelJob = useCancelJob()
  const { t } = useTranslation()
  function countdownDisplay(completesAt: string | undefined): string {
    if (!completesAt) return '...'
    const remaining = Math.max(0, new Date(completesAt).getTime() - Date.now())
    if (remaining <= 0) return t('mobile.ready')
    const h = Math.floor(remaining / 3_600_000)
    const m = Math.floor((remaining % 3_600_000) / 60_000)
    const s = Math.floor((remaining % 60_000) / 1000)
    if (h > 0) return `${h}h ${String(m).padStart(2, '0')}m`
    return `${m}m ${String(s).padStart(2, '0')}s`
  }

  const runningJobs = useMemo(
    () => (Array.isArray(jobsData) ? jobsData.filter((j) => j.status === 'running') : []),
    [jobsData],
  )
  const inUse = queueData?.inUse ?? 0
  const maxSlots = queueData?.maxSlots ?? 0

  return (
    <div className="bg-white/60 rounded-xl border border-amber-300/50 p-3 min-w-[220px] shrink-0">
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-xs font-bold text-amber-800 uppercase tracking-wider">{t('mobile.production')}</h3>
        <span className="text-[10px] font-semibold text-amber-700 tabular-nums">
          {t('mobile.slots', { used: inUse, max: maxSlots })}
        </span>
      </div>

      {runningJobs.length === 0 && (
        <div className="text-[10px] text-amber-400 italic text-center py-2">
          {t('mobile.noProduction')}
        </div>
      )}

      <div className="space-y-1">
        {runningJobs.slice(0, 3).map((job) => (
          <div key={job.id} className="flex items-center gap-2 p-1.5 bg-amber-50/70 rounded-lg border border-amber-200/30">
            <img
              src={resourceIcon(job.resourceId)}
              alt=""
              className="w-6 h-6 object-contain shrink-0 rounded"
              loading="lazy"
            />
            <div className="flex-1 min-w-0">
              <div className="text-[9px] font-semibold text-amber-900 truncate">
                {resourceName(job.resourceId)} ×{job.amount}
              </div>
              <div className="text-[9px] text-amber-500 tabular-nums">
                {countdownDisplay(job.completesAt)}
              </div>
            </div>
            <button
              onClick={() => cancelJob.mutate(job.id)}
              disabled={cancelJob.isPending}
              className="w-5 h-5 bg-red-100 hover:bg-red-200 rounded-full flex items-center justify-center shrink-0 transition-colors disabled:opacity-50"
              title={t('mobile.cancel')}
            >
              <svg className="w-3 h-3 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
