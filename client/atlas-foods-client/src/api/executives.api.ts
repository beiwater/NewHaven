import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './client'
import type {
  Executive,
  ExecutiveSearchResult,
  RecruitResult,
  TrainResult,
  ExecutiveDetail,
} from '@/game/executives'

const EXECUTIVES_KEY = ['executives'] as const
const MY_EXECUTIVES_KEY = ['myExecutives'] as const

// ── Transform: backend response → frontend model ────

/**
 * Backend stubs may return legacy { skills: {...} } format.
 * This mapping normalizes to the frontend model.
 * Once the backend returns the new schema, this transform
 * can pass through directly.
 */
function transformExecutive(raw: Record<string, unknown>): Executive {
  const level = (raw.level as number) ?? 1
  const rarity = (raw.rarity as string) ?? inferRarity(level)
  const status = (raw.status as string) ?? 'idle'
  const skills = (raw.skills as Record<string, number>) ?? {}

  return {
    id: (raw.id as string) ?? '',
    name: (raw.name as string) ?? 'Unknown',
    title: (raw.title as string) ?? 'Manager',
    level,
    rarity: rarity as Executive['rarity'],
    stage: (raw.stage as string) ?? stageAtLevel(level),
    salary: (raw.salary as number) ?? salaryAtLevel(level),
    productionBonus: (raw.productionBonus as number) ?? Math.round((skills.management ?? 50) / 10),
    salesBonus: (raw.salesBonus as number) ?? Math.round((skills.finance ?? 50) / 5),
    mgmtDiscount: (raw.mgmtDiscount as number) ?? Math.round((skills.science ?? 50) / 15),
    recruitCost: (raw.recruitCost as number) ?? recruitCost(rarity, level),
    trainingCost: (raw.trainingCost as number) ?? trainingCost(level),
    trainingTime: (raw.trainingTime as number) ?? trainingTimeSeconds(level),
    status: status as Executive['status'],
    trainingEndTime: raw.trainingEndTime as string | undefined,
  }
}

function inferRarity(level: number): string {
  if (level >= 9) return 'Legendary'
  if (level >= 7) return 'Epic'
  if (level >= 5) return 'Rare'
  return 'Common'
}

function stageAtLevel(level: number): string {
  if (level >= 10) return 'Executive VP'
  if (level >= 8) return 'Director'
  if (level >= 6) return 'Senior Manager'
  if (level >= 4) return 'Manager'
  if (level >= 2) return 'Junior Manager'
  return 'Trainee'
}

function trainingCost(level: number): number {
  return Math.round(5000 * Math.pow(level, 1.6))
}

function trainingTimeSeconds(level: number): number {
  return Math.round(3600 * Math.pow(level, 0.7))
}

function salaryAtLevel(level: number): number {
  return Math.round(600 + 80 * Math.pow(level, 1.3))
}

function recruitCost(rarity: string, level: number): number {
  const factor: Record<string, number> = {
    Legendary: 2.5, Epic: 1.8, Rare: 1.2, Common: 0.8,
  }
  return Math.round(15000 * (factor[rarity] ?? 1.0) * Math.pow(level, 0.8))
}

// ── API hooks ──────────────────────────────────────

/** Search/refresh the executive market */
export function useExecutiveSearch() {
  return useQuery({
    queryKey: EXECUTIVES_KEY,
    queryFn: async () => {
      const data = await api.post<{
        executives: Record<string, unknown>[]
        total: number
        refreshCooldown?: string
      }>('/api/v2/executives/search/')
      return {
        executives: (data.executives ?? []).map(transformExecutive),
        total: data.total ?? 0,
        refreshCooldown: data.refreshCooldown ?? '09:00:00',
      } satisfies ExecutiveSearchResult
    },
    refetchInterval: 60_000,
    staleTime: 30_000,
  })
}

/** Recruit an executive from the market */
export function useRecruitExecutive() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (executiveId: string) =>
      api.post<RecruitResult>('/api/v2/executives/recruit/', { executiveId }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: EXECUTIVES_KEY })
      qc.invalidateQueries({ queryKey: MY_EXECUTIVES_KEY })
      qc.invalidateQueries({ queryKey: ['company'] })
    },
  })
}

/** Fetch executives owned by the player */
export function useMyExecutives() {
  return useQuery({
    queryKey: MY_EXECUTIVES_KEY,
    queryFn: async () => {
      const data = await api.post<{
        executives: Record<string, unknown>[]
      }>('/api/v2/executives/search/', { scope: 'mine' })
      return (data.executives ?? []).map(transformExecutive) as Executive[]
    },
    refetchInterval: 15_000,
    staleTime: 10_000,
  })
}

/** Train (level up) an executive */
export function useTrainExecutive() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (executiveId: string) =>
      api.post<TrainResult>(`/api/v2/executives/train/${executiveId}/`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: MY_EXECUTIVES_KEY })
      qc.invalidateQueries({ queryKey: ['company'] })
    },
  })
}

/** Fetch single executive detail */
export function useExecutiveDetail(id: string | null) {
  return useQuery({
    queryKey: ['executiveDetail', id],
    queryFn: async () => {
      const data = await api.get<Record<string, unknown>>(`/api/v3/executives/${id}/`)
      return {
        ...transformExecutive(data),
        morale: (data.morale as number) ?? 100,
      } as ExecutiveDetail
    },
    enabled: !!id,
  })
}
