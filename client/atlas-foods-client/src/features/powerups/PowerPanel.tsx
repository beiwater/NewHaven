import { useUIStore } from '@/store/ui.store'
import { audio } from '@/audio/AudioManager'
import { usePowerupTypes, useActivePowerup, useActivatePowerup, type PowerupType, type ActivePowerup } from '@/api/powerup.api'

/** Format seconds into a human-readable duration string */
function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  const min = Math.floor(seconds / 60)
  const sec = seconds % 60
  if (min < 60) return sec > 0 ? `${min}m ${sec}s` : `${min}m`
  const h = Math.floor(min / 60)
  const m = min % 60
  return m > 0 ? `${h}h ${m}m` : `${h}h`
}

/** Parse a Go-style duration string like "12m30s" into a display string */
function formatRemaining(duration: string): string {
  if (!duration) return ''
  // Match Go duration format: XhYmZs
  const h = duration.match(/(\d+)h/)
  const m = duration.match(/(\d+)m(?!s)/)
  const s = duration.match(/(\d+)s/)
  const parts: string[] = []
  if (h) parts.push(`${h[1]}h`)
  if (m) parts.push(`${m[1]}m`)
  if (s) parts.push(`${s[1]}s`)
  return parts.join(' ') || 'expiring'
}

/** Format an ISO timestamp into a localized time string */
function formatEndsAt(iso: string): string {
  try {
    const d = new Date(iso)
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } catch {
    return ''
  }
}

export function PowerPanel() {
  const powerupOpen = useUIStore((s) => s.powerupOpen)
  const setPowerupOpen = useUIStore((s) => s.setPowerupOpen)

  const { data: typesData, isLoading: typesLoading } = usePowerupTypes()
  const { data: activeData, isLoading: activeLoading } = useActivePowerup()
  const activate = useActivatePowerup()

  if (!powerupOpen) return null

  const types: PowerupType[] = typesData?.boosts ?? []
  const active: ActivePowerup[] = activeData?.active ?? []
  const remaining: number = activeData?.remaining ?? 0

  const hasActive = active.length > 0

  return (
    <div className="fixed bottom-[102px] right-[322px] w-72 max-h-96 bg-amber-50 border-2 border-amber-700/40 rounded-t-lg shadow-xl flex flex-col z-40">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 bg-amber-800 text-white rounded-t-[5px]">
        <span className="text-xs font-semibold flex items-center gap-1.5">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          Power-ups
        </span>
        <button onClick={() => setPowerupOpen(false)} className="text-amber-200 hover:text-white">
          <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-2 space-y-2">
        {/* Active power-up banner */}
        {hasActive && (
          <div className="bg-green-100 border border-green-400 rounded px-2.5 py-2">
            <div className="text-[11px] font-semibold text-green-800 flex items-center gap-1">
              <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              Active
            </div>
            {active.map((a) => (
              <div key={a.type} className="mt-1 text-xs text-green-700">
                <span className="font-medium">{a.type.replace('boost-', '').replace(/-/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())}</span>
                <span className="ml-2 text-green-600">ends {formatEndsAt(a.endsAt)}</span>
                {a.remaining && <span className="ml-1 text-green-500">({formatRemaining(a.remaining)})</span>}
              </div>
            ))}
          </div>
        )}

        {/* Loading state */}
        {typesLoading || activeLoading ? (
          <div className="text-xs text-amber-500 text-center py-4 animate-pulse">Loading...</div>
        ) : types.length === 0 ? (
          <div className="text-xs text-amber-500 text-center py-4">No power-ups available</div>
        ) : (
          types.map((boost) => {
            const isActive = active.some((a) => a.type === boost.id)
            const disabled = isActive || activate.isPending

            return (
              <div
                key={boost.id}
                className={`rounded border px-2.5 py-2 ${
                  isActive
                    ? 'border-green-300 bg-green-50'
                    : 'border-amber-200 bg-white hover:border-amber-400'
                } ${disabled ? 'opacity-75' : 'cursor-pointer transition-colors'}`}
                onClick={() => {
                  if (disabled) return
                  audio.playSfx('buff_activate')
                  activate.mutate(boost.id)
                }}
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0 flex-1">
                    <div className="text-xs font-semibold text-amber-900">{boost.name}</div>
                    <div className="text-[11px] text-amber-600 mt-0.5">{boost.desc}</div>
                    <div className="text-[10px] text-amber-400 mt-0.5">{formatDuration(boost.duration)}</div>
                  </div>
                  <div className="flex flex-col items-end shrink-0">
                    {isActive ? (
                      <span className="text-[10px] font-semibold text-green-600 bg-green-100 px-1.5 py-0.5 rounded">ACTIVE</span>
                    ) : (
                      <span className="text-[10px] font-semibold text-amber-600 bg-amber-100 px-1.5 py-0.5 rounded">USE</span>
                    )}
                  </div>
                </div>
              </div>
            )
          })
        )}

        {/* Remaining uses */}
        <div className="text-[10px] text-amber-400 text-center pt-1 border-t border-amber-200/60 mt-1">
          {remaining > 0
            ? `${remaining} power-up${remaining !== 1 ? 's' : ''} remaining`
            : 'No power-ups remaining'}
        </div>

        {/* Mutation error */}
        {activate.isError && (
          <div className="text-[11px] text-red-600 text-center bg-red-50 rounded px-2 py-1">
            {activate.error instanceof Error ? activate.error.message : 'Failed to activate power-up'}
          </div>
        )}
      </div>
    </div>
  )
}
