import type { Executive } from '@/game/executives'
import { formatDuration } from '@/game/executives'

interface TrainingQueueProps {
  trainees: Executive[]
  onCompleteNow?: (id: string) => void
  completeNowPending?: string | null
}

export function TrainingQueue({
  trainees,
  onCompleteNow,
  completeNowPending,
}: TrainingQueueProps) {
  if (trainees.length === 0) {
    return null
  }

  return (
    <section>
      <div className="mb-3 flex items-center gap-2">
        <svg className="h-5 w-5 text-amber-800" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" d="M19 14l-7 7m0 0l-7-7m7 7V3" />
        </svg>
        <h3 className="text-base font-black uppercase tracking-wider text-amber-900">
          Training Queue
        </h3>
      </div>

      <div className="space-y-2">
        {trainees.map((exec) => {
          const endTime = exec.trainingEndTime ? new Date(exec.trainingEndTime).getTime() : 0
          const now = Date.now()
          const remainingMs = Math.max(0, endTime - now)
          const remainingSec = Math.floor(remainingMs / 1000)
          const progress = endTime > now
            ? 1 - remainingMs / (10 * 60 * 60 * 1000) // placeholder: assume 10h total
            : 1

          return (
            <div
              key={exec.id}
              className="flex items-center gap-3 rounded-xl border border-amber-200/60 bg-white/60 p-3"
            >
              {/* Avatar */}
              <div className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-sm font-black text-white
                ${exec.rarity === 'Legendary' ? 'bg-orange-500' :
                  exec.rarity === 'Epic' ? 'bg-purple-500' :
                  exec.rarity === 'Rare' ? 'bg-blue-500' : 'bg-gray-400'}`}
              >
                {exec.name.charAt(0)}
              </div>

              {/* Info */}
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <span className="truncate text-sm font-bold text-amber-950">{exec.name}</span>
                  <span className="text-[10px] font-medium text-amber-600">Lv. {exec.level} → Lv. {exec.level + 1}</span>
                </div>

                {/* Progress bar */}
                <div className="mt-1 h-2 w-full rounded-full bg-amber-200/50 overflow-hidden">
                  <div
                    className="h-full rounded-full bg-blue-500 transition-all duration-500"
                    style={{ width: `${Math.min(100, Math.round(progress * 100))}%` }}
                  />
                </div>
              </div>

              {/* Time remaining */}
              <div className="shrink-0 text-right">
                <div className="text-[11px] font-bold text-amber-900">
                  {remainingSec > 0 ? formatDuration(remainingSec) : 'Complete!'}
                </div>
                {remainingSec > 0 && (
                  <button
                    onClick={() => onCompleteNow?.(exec.id)}
                    disabled={completeNowPending === exec.id}
                    className="mt-1 rounded-md bg-orange-500 px-2 py-0.5 text-[10px] font-bold text-white hover:bg-orange-600 disabled:bg-orange-300 transition-colors"
                  >
                    {completeNowPending === exec.id ? '...' : 'Complete Now'}
                  </button>
                )}
                {remainingSec <= 0 && (
                  <div className="mt-1 rounded-md bg-green-100 px-2 py-0.5 text-[10px] font-semibold text-green-700">
                    Ready
                  </div>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </section>
  )
}
