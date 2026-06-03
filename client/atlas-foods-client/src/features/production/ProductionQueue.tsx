import { useProductionQueue, useProductionJobs, useCancelJob } from '@/api/production.api'

export function ProductionQueue() {
  const { data: queue } = useProductionQueue()
  const { data: jobs } = useProductionJobs()
  const cancelJob = useCancelJob()

  const allJobs = jobs ?? []
  const collectableJobs = allJobs.filter((j) => (j.claimableAmount ?? 0) > 0)
  const activeJobs = allJobs.filter((j) => j.status !== 'claimed')

  return (
    <div className="p-4">
      <h2 className="text-lg font-bold text-amber-900 mb-3">Production Queue</h2>

      {queue && (
        <div className="flex gap-3 mb-4">
          <div className="flex-1 bg-white/60 rounded-lg p-3 border border-amber-200/40 text-center">
            <div className="text-2xl font-bold text-amber-900">{queue?.inUse ?? 0}</div>
            <div className="text-[10px] text-amber-600 uppercase">Used Slots</div>
          </div>
          <div className="flex-1 bg-white/60 rounded-lg p-3 border border-amber-200/40 text-center">
            <div className="text-2xl font-bold text-amber-900">{queue.maxSlots}</div>
            <div className="text-[10px] text-amber-600 uppercase">Max Slots</div>
          </div>
        </div>
      )}

      {collectableJobs.length > 0 && (
        <div className="mb-4">
          <h3 className="text-xs font-semibold text-green-700 uppercase tracking-wider mb-2">
            Available to Collect ({collectableJobs.length})
          </h3>
          <div className="space-y-1.5">
            {collectableJobs.map((job) => (
              <div
                key={job.id}
                className="flex items-center gap-2 p-2 bg-green-50 rounded-lg border border-green-200 text-xs"
              >
                <span className="w-2 h-2 rounded-full bg-green-500" />
                <span className="text-amber-900 font-medium">#{job.resourceId}</span>
                <span className="text-amber-600">x{job.claimableAmount}</span>
                <span className="ml-auto text-green-600 font-semibold">Growing</span>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="space-y-1.5">
        <h3 className="text-xs font-semibold text-amber-700 uppercase tracking-wider mb-2">
          In Progress ({activeJobs.length})
        </h3>
        {activeJobs.map((job) => {
          const pct = Math.min(100, Math.round(((job.claimedAmount ?? 0) / Math.max(1, job.amount)) * 100))
          return (
            <div
              key={job.id}
              className="flex items-center gap-2 p-2 bg-white/60 rounded-lg border border-amber-200/40 text-xs"
            >
              <div className="w-1.5 h-1.5 rounded-full bg-blue-500" />
              <span className="text-amber-900 font-medium">Job #{job.id.slice(0, 6)}</span>
              <span className="text-amber-600">{job.claimedAmount ?? 0} / {job.amount}</span>
              <div className="flex-1 h-1.5 bg-amber-200/60 rounded-full mx-2">
                <div className="h-full bg-blue-500 rounded-full" style={{ width: `${pct}%` }} />
              </div>
              <span className="text-amber-500 tabular-nums text-[10px]">
                {job.completesAt ? new Date(job.completesAt).toLocaleTimeString() : '...'}
              </span>
              <button
                onClick={() => cancelJob.mutate(job.id)}
                className="p-1 text-red-400 hover:text-red-600 transition-colors"
              >
                <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          )
        })}
        {activeJobs.length === 0 && collectableJobs.length === 0 && (
          <div className="text-xs text-amber-400 italic py-4 text-center">No production jobs</div>
        )}
      </div>
    </div>
  )
}
