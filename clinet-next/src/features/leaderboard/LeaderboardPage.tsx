import { useState } from 'react'
import { useLeaderboard, rankColor, formatMainStat, isCurrentCompany, SORT_LABELS, SORT_DIMENSIONS, type SortDimension } from '@/api/hooks/leaderboard.hooks'

export function LeaderboardPage() {
  const [sort, setSort] = useState<SortDimension>('net_worth')
  const [page, setPage] = useState(1)
  const { data, isLoading } = useLeaderboard(sort, page)

  return (
    <div className="p-4">
      <h2 className="text-lg font-bold text-amber-900 mb-3">Leaderboard</h2>

      {/* Sort tabs */}
      <div className="flex gap-1 mb-3 p-1 bg-white/50 rounded-lg border border-amber-200/60">
        {SORT_DIMENSIONS.map((d) => (
          <button
            key={d}
            onClick={() => { setSort(d); setPage(1) }}
            className={`flex-1 py-1 text-xs font-semibold rounded-md ${sort === d ? 'bg-amber-200/70 text-amber-900' : 'text-amber-600 hover:text-amber-800'}`}
          >
            {SORT_LABELS[d]}
          </button>
        ))}
      </div>

      {isLoading ? (
        <div className="text-xs text-amber-400 italic">Loading...</div>
      ) : (
        <div className="space-y-1">
          {(data?.entries ?? []).map((entry) => (
            <div
              key={entry.companyId}
              className={`flex items-center gap-3 p-2.5 rounded-lg border text-xs ${isCurrentCompany(entry) ? 'bg-amber-100 border-amber-400' : 'bg-white/60 border-amber-200/40'}`}
            >
              <span className={`w-8 h-8 flex items-center justify-center rounded-full font-bold text-xs border ${rankColor(entry.rank)}`}>
                {entry.rank}
              </span>
              <div className="flex-1 min-w-0">
                <div className="font-semibold text-amber-900 truncate">{entry.companyName}</div>
                <div className="text-[10px] text-amber-600">Lv.{entry.level}</div>
              </div>
              <div className="text-right">
                <div className="font-bold text-amber-900">{formatMainStat(entry.mainStat, sort)}</div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Pagination */}
      {data && data.totalPages > 1 && (
        <div className="flex justify-center gap-2 mt-4">
          <button
            disabled={page <= 1}
            onClick={() => setPage(p => p - 1)}
            className="px-3 py-1 text-xs bg-white rounded border border-amber-200 disabled:opacity-30"
          >
            Prev
          </button>
          <span className="text-xs text-amber-700 self-center">{page} / {data.totalPages}</span>
          <button
            disabled={page >= data.totalPages}
            onClick={() => setPage(p => p + 1)}
            className="px-3 py-1 text-xs bg-white rounded border border-amber-200 disabled:opacity-30"
          >
            Next
          </button>
        </div>
      )}
    </div>
  )
}
