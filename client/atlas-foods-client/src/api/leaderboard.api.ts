import { useQuery } from '@tanstack/react-query'
import { api } from './client'
import { getCompanyId } from './client'

export type SortDimension = 'net_worth' | 'level' | 'production' | 'sales' | 'contracts'

export interface LeaderboardEntry {
  rank: number
  companyId: number
  companyName: string
  level: number
  mainStat: number
}

export interface LeaderboardResult {
  entries: LeaderboardEntry[]
  total: number
  page: number
  limit: number
  totalPages: number
  sort: SortDimension
}

export function useLeaderboard(sort: SortDimension = 'net_worth', page = 1, limit = 10) {
  return useQuery({
    queryKey: ['leaderboard', sort, page, limit],
    queryFn: () =>
      api.get<LeaderboardResult>(
        `/api/v2/leaderboard/?sort=${sort}&page=${page}&limit=${limit}`,
      ),
    refetchInterval: 60_000,
    staleTime: 30_000,
  })
}

/** Returns the rank label for a given numeric rank. */
export function rankLabel(rank: number): string {
  if (rank === 1) return '1st'
  if (rank === 2) return '2nd'
  if (rank === 3) return '3rd'
  return `${rank}th`
}

/** Returns a color class for the podium badge. */
export function rankColor(rank: number): string {
  switch (rank) {
    case 1: return 'bg-yellow-400 text-yellow-900 border-yellow-500'
    case 2: return 'bg-slate-300 text-slate-800 border-slate-400'
    case 3: return 'bg-orange-300 text-orange-800 border-orange-400'
    default: return 'bg-amber-100 text-amber-700 border-amber-300'
  }
}

/** Format a main stat value as currency for net_worth, number otherwise. */
export function formatMainStat(value: number, sort: SortDimension): string {
  if (sort === 'net_worth') {
    return `$${(value).toLocaleString('en-US', { maximumFractionDigits: 0 })}`
  }
  if (sort === 'level' || sort === 'production' || sort === 'sales') {
    return Math.round(value).toLocaleString()
  }
  return value.toLocaleString('en-US', { maximumFractionDigits: 0 })
}

/** Check if a leaderboard entry is the current player's company. */
export function isCurrentCompany(entry: LeaderboardEntry): boolean {
  return entry.companyId === Number(getCompanyId())
}

export const SORT_LABELS: Record<SortDimension, string> = {
  net_worth: 'Net Worth',
  level: 'Level',
  production: 'Production',
  sales: 'Sales',
  contracts: 'Contracts',
}

export const SORT_DIMENSIONS: SortDimension[] = [
  'net_worth',
  'level',
  'production',
  'sales',
  'contracts',
]
