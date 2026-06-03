import { useState } from 'react'
import {
  useLeaderboard,
  rankColor,
  formatMainStat,
  isCurrentCompany,
  SORT_LABELS,
  SORT_DIMENSIONS,
  type SortDimension,
  type LeaderboardEntry,
} from '@/api/leaderboard.api'

export function LeaderboardPage() {
  const [sort, setSort] = useState<SortDimension>('net_worth')
  const [page, setPage] = useState(1)
  const limit = 10

  const { data, isLoading, isError, error } = useLeaderboard(sort, page, limit)

  const totalPages = data?.totalPages ?? 1
  const entries = data?.entries ?? []
  const podium = entries.slice(0, 3)

  return (
    <div className="h-full overflow-y-auto">
      <div className="mx-auto max-w-7xl p-4 md:p-6">
        {/* Header */}
        <div className="mb-5 flex items-center gap-3">
          <svg className="h-7 w-7 text-yellow-500" fill="currentColor" viewBox="0 0 24 24">
            <path d="M5.5 21v-6.5H4a1 1 0 01-1-1V9a1 1 0 011-1h1.5V3A1.5 1.5 0 017 1.5h10A1.5 1.5 0 0118.5 3v5H20a1 1 0 011 1v4.5a1 1 0 01-1 1h-1.5V21a1.5 1.5 0 01-1.5 1.5H7A1.5 1.5 0 015.5 21zM7 3v5h10V3H7z" />
          </svg>
          <div>
            <p className="text-[10px] font-bold uppercase tracking-[0.24em] text-amber-700/70">
              Rankings & Competition
            </p>
            <h2 className="text-2xl font-black text-amber-950">Leaderboard</h2>
          </div>
        </div>

        {/* Grid: left main area | right info panel */}
        <div className="grid gap-5 lg:grid-cols-[1fr_280px]">
          <div className="space-y-5">
            {/* Sort tabs */}
            <div className="flex flex-wrap gap-2">
              {SORT_DIMENSIONS.map((dim) => (
                <button
                  key={dim}
                  onClick={() => { setSort(dim); setPage(1) }}
                  className={`rounded-full px-4 py-2 text-xs font-bold transition-colors ${
                    sort === dim
                      ? 'bg-amber-800 text-white shadow'
                      : 'bg-white/60 text-amber-800 hover:bg-amber-100'
                  }`}
                >
                  {SORT_LABELS[dim]}
                </button>
              ))}
            </div>

            {/* Loading */}
            {isLoading && (
              <div className="flex items-center justify-center rounded-2xl border border-amber-200/60 bg-white/50 py-16">
                <div className="text-xs font-semibold text-amber-600 animate-pulse">Loading rankings...</div>
              </div>
            )}

            {/* Error */}
            {isError && !isLoading && (
              <div className="rounded-2xl border border-red-200 bg-red-50 p-6 text-center">
                <p className="text-xs font-semibold text-red-700">
                  Failed to load leaderboard: {error instanceof Error ? error.message : 'Unknown error'}
                </p>
              </div>
            )}

            {/* Podium — top 3 */}
            {!isLoading && !isError && podium.length > 0 && (
              <PodiumSection entries={podium} sort={sort} />
            )}

            {/* Ranking table */}
            {!isLoading && !isError && (
              <>
                <RankingTable entries={entries} sort={sort} />

                {/* Pagination */}
                <div className="flex items-center justify-between rounded-xl border border-amber-200/60 bg-white/60 px-4 py-3">
                  <button
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    disabled={page <= 1}
                    className="rounded-lg border border-amber-300 px-4 py-1.5 text-xs font-bold text-amber-800 hover:bg-amber-100 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                  >
                    Previous
                  </button>
                  <span className="text-xs font-semibold text-amber-700">
                    Page {page} / {totalPages}
                  </span>
                  <button
                    onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                    disabled={page >= totalPages}
                    className="rounded-lg bg-amber-800 px-4 py-1.5 text-xs font-bold text-white hover:bg-amber-900 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                  >
                    Next
                  </button>
                </div>
              </>
            )}
          </div>

          {/* Right panel — Public Data Only */}
          <PublicDataPanel sort={sort} />
        </div>
      </div>
    </div>
  )
}

// ── Podium ─────────────────────────────────────────

function PodiumSection({
  entries,
  sort,
}: {
  entries: LeaderboardEntry[]
  sort: SortDimension
}) {
  // Order: 2nd (left), 1st (center), 3rd (right)
  const second = entries.find((e) => e.rank === 2)
  const first = entries.find((e) => e.rank === 1)
  const third = entries.find((e) => e.rank === 3)

  return (
    <div className="flex items-end justify-center gap-3 md:gap-5">
      {/* 2nd */}
      <PodiumCard
        entry={second}
        pos={2}
        sort={sort}
        className="w-1/3 md:w-1/4"
      />
      {/* 1st */}
      <PodiumCard
        entry={first}
        pos={1}
        sort={sort}
        className="w-1/3 md:w-1/3 scale-105 z-10"
        champion
      />
      {/* 3rd */}
      <PodiumCard
        entry={third}
        pos={3}
        sort={sort}
        className="w-1/3 md:w-1/4"
      />
    </div>
  )
}

function PodiumCard({
  entry,
  pos,
  sort,
  className = '',
  champion = false,
}: {
  entry?: LeaderboardEntry
  pos: number
  sort: SortDimension
  className?: string
  champion?: boolean
}) {
  if (!entry) {
    return <div className={`rounded-2xl border border-dashed border-amber-200/50 bg-white/30 p-4 text-center ${className}`}>
      <div className="text-[10px] text-amber-400">—</div>
    </div>
  }

  const badgeColors = ['bg-yellow-400', 'bg-slate-300', 'bg-orange-300']
  const badgeTexts = ['text-yellow-900', 'text-slate-800', 'text-orange-800']

  return (
    <div className={`rounded-2xl border-2 bg-white/60 p-4 text-center shadow-sm ${champion ? 'border-yellow-400 shadow-md' : 'border-amber-200/60'} ${className}`}>
      {/* Medal */}
      <div className={`mx-auto mb-2 flex h-8 w-8 items-center justify-center rounded-full ${badgeColors[pos - 1]} ${badgeTexts[pos - 1]} text-sm font-black`}>
        {pos}
      </div>
      {/* Avatar */}
      <div className={`mx-auto mb-2 flex h-10 w-10 items-center justify-center rounded-full text-white text-sm font-black
        ${champion ? 'bg-yellow-500' : 'bg-amber-400'}`}
      >
        {entry.companyName.charAt(0)}
      </div>
      <div className="truncate text-sm font-bold text-amber-950">{entry.companyName}</div>
      <div className="text-[10px] font-semibold text-amber-700">Level {entry.level}</div>
      <div className="mt-1 text-sm font-black text-green-700">
        {formatMainStat(entry.mainStat, sort)}
      </div>
    </div>
  )
}

// ── Ranking Table ──────────────────────────────────

function RankingTable({
  entries,
  sort,
}: {
  entries: LeaderboardEntry[]
  sort: SortDimension
}) {
  return (
    <div className="overflow-hidden rounded-2xl border border-amber-200/60 bg-white/60 shadow-sm">
      <table className="w-full text-left text-xs">
        <thead>
          <tr className="border-b border-amber-200/60 bg-amber-100/50">
            <th className="px-4 py-3 font-bold uppercase tracking-wider text-amber-700 w-14">Rank</th>
            <th className="px-4 py-3 font-bold uppercase tracking-wider text-amber-700">Company</th>
            <th className="px-4 py-3 font-bold uppercase tracking-wider text-amber-700 text-center w-20">Level</th>
            <th className="px-4 py-3 font-bold uppercase tracking-wider text-amber-700 text-right">
              Main Stat ({SORT_LABELS[sort]})
              <span className="ml-1 inline-flex h-3.5 w-3.5 items-center justify-center rounded-full bg-amber-200 text-[8px] font-bold text-amber-600 cursor-help" title={`Sorted by ${SORT_LABELS[sort]}`}>i</span>
            </th>
          </tr>
        </thead>
        <tbody>
          {entries.length === 0 && (
            <tr>
              <td colSpan={4} className="px-4 py-8 text-center text-amber-500">
                No companies found on this page.
              </td>
            </tr>
          )}
          {entries.map((entry) => {
            const isMe = isCurrentCompany(entry)
            return (
              <tr
                key={entry.companyId}
                className={`border-b border-amber-100/60 transition-colors ${
                  isMe
                    ? 'bg-green-50/80 hover:bg-green-100/80'
                    : 'hover:bg-amber-50/50'
                }`}
              >
                {/* Rank */}
                <td className="px-4 py-3">
                  <span className={`inline-flex h-6 w-6 items-center justify-center rounded-full text-[10px] font-black ${rankColor(entry.rank)}`}>
                    {entry.rank <= 3 ? entry.rank : entry.rank}
                  </span>
                </td>
                {/* Company */}
                <td className="px-4 py-3">
                  <div className="flex items-center gap-2">
                    <div className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-white text-xs font-bold
                      ${isMe ? 'bg-green-600' : 'bg-amber-500'}`}
                    >
                      {entry.companyName.charAt(0)}
                    </div>
                    <div className="min-w-0">
                      <span className={`truncate font-bold ${isMe ? 'text-green-800' : 'text-amber-950'}`}>
                        {entry.companyName}
                      </span>
                      {isMe && (
                        <span className="ml-1.5 rounded bg-green-200 px-1.5 py-0.5 text-[9px] font-bold text-green-800">You</span>
                      )}
                    </div>
                  </div>
                </td>
                {/* Level */}
                <td className="px-4 py-3 text-center font-bold text-amber-900">
                  {entry.level}
                </td>
                {/* Main Stat */}
                <td className="px-4 py-3 text-right font-bold text-green-700">
                  {formatMainStat(entry.mainStat, sort)}
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

// ── Public Data Panel ──────────────────────────────

function PublicDataPanel({ sort }: { sort: SortDimension }) {
  return (
    <div className="lg:sticky lg:top-4 lg:self-start rounded-2xl border-2 border-amber-300/40 bg-gradient-to-b from-amber-50 to-amber-100/20 p-5 shadow-sm">
      <div className="mb-4 flex items-center gap-2 border-b border-amber-200/50 pb-3">
        <svg className="h-4 w-4 text-blue-600" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <h3 className="text-sm font-black uppercase tracking-wider text-amber-900">
          Public Data Only
        </h3>
      </div>

      <p className="mb-4 text-[10px] leading-relaxed text-amber-700">
        The following information is visible on the leaderboard:
      </p>

      <div className="space-y-3">
        <InfoRow
          icon={
            <svg className="h-4 w-4 text-amber-600" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
            </svg>
          }
          label="Company Name"
        />
        <InfoRow
          icon={
            <svg className="h-4 w-4 text-amber-600" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M5.5 21v-6.5H4a1 1 0 01-1-1V9a1 1 0 011-1h1.5V3A1.5 1.5 0 017 1.5h10A1.5 1.5 0 0118.5 3v5H20a1 1 0 011 1v4.5a1 1 0 01-1 1h-1.5V21a1.5 1.5 0 01-1.5 1.5H7A1.5 1.5 0 015.5 21z" />
            </svg>
          }
          label="Company Logo"
        />
        <InfoRow
          icon={
            <svg className="h-4 w-4 text-amber-600" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 1.944A11.954 11.954 0 012.166 5C2.056 5.647 2 6.319 2 7c0 5.225 3.34 9.67 8 11.317C14.66 16.67 18 12.225 18 7c0-.682-.057-1.353-.166-2A11.954 11.954 0 0112 1.944z" />
            </svg>
          }
          label="Level"
        />
        <InfoRow
          icon={
            <svg className="h-4 w-4 text-green-600" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          }
          label={`${SORT_LABELS[sort]} / Main Stat`}
        />
      </div>

      <div className="mt-4 rounded-lg border border-amber-200/30 bg-amber-100/40 px-3 py-2">
        <div className="flex items-start gap-2">
          <svg className="mt-0.5 h-3.5 w-3.5 shrink-0 text-amber-600" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
          <p className="text-[10px] leading-relaxed text-amber-700">
            Private information such as resources, buildings, contracts, and player activity is hidden to ensure fair competition.
          </p>
        </div>
      </div>

      {/* Podium decoration */}
      <div className="mt-5 flex items-end justify-center gap-3 opacity-30">
        <div className="h-8 w-8 rounded-t bg-amber-300" />
        <div className="h-12 w-8 rounded-t bg-yellow-400" />
        <div className="h-6 w-8 rounded-t bg-orange-300" />
        <div className="absolute ml-14 flex gap-1.5 mt-1">
          <span className="text-[9px] font-bold text-amber-500">1</span>
          <span className="text-[9px] font-bold text-amber-500">2</span>
          <span className="text-[9px] font-bold text-amber-500">3</span>
        </div>
      </div>
    </div>
  )
}

function InfoRow({ icon, label }: { icon: React.ReactNode; label: string }) {
  return (
    <div className="flex items-center gap-2 rounded-lg border border-amber-200/40 bg-white/60 px-3 py-2">
      <span className="text-amber-600">{icon}</span>
      <span className="text-[11px] font-semibold text-amber-800">{label}</span>
    </div>
  )
}
