import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './client'
import type { ProductionQueue, ResourceDefinition } from '@/game/types'
import { normalizeProductionJobList } from './compat'

function newProductionRequestId() {
  return globalThis.crypto?.randomUUID?.()
    ?? `production-${Date.now()}-${Math.random().toString(36).slice(2)}`
}

// Actual API: /api/v2/production/jobs/ returns ProductionJob[] directly
export function useProductionJobs() {
  return useQuery({
    queryKey: ['productionJobs'],
    queryFn: async () => normalizeProductionJobList(await api.get<unknown>('/api/v2/production/jobs/')),
    refetchInterval: 10_000,
  })
}

// Actual API: /api/v2/production/queue/ returns { byBuilding, inUse, maxSlots }
export function useProductionQueue() {
  return useQuery({
    queryKey: ['productionQueue'],
    queryFn: () => api.get<ProductionQueue>('/api/v2/production/queue/'),
    refetchInterval: 10_000,
  })
}

// Actual API: /api/v2/production/claimable/ returns ProductionJob[] directly
export function useClaimableJobs() {
  return useQuery({
    queryKey: ['claimableJobs'],
    queryFn: async () => normalizeProductionJobList(await api.get<unknown>('/api/v2/production/claimable/')),
    refetchInterval: 10_000,
  })
}

export function useProductionOptions(buildingId: string | undefined) {
  return useQuery({
    queryKey: ['productionOptions', buildingId],
    queryFn: () => api.get<ResourceDefinition[]>(`/api/v2/buildings/${buildingId}/production-options/`),
    enabled: !!buildingId,
  })
}

export function useStartProduction() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: {
      buildingId: string
      kind: number
      amount: number
      quality: number
      estimatedSecondsToFinish?: number
      requestId?: string
    }) => api.post(`/api/v1/buildings/${params.buildingId}/busy/`, {
      kind: params.kind,
      amount: params.amount,
      quality: params.quality,
      estimatedSecondsToFinish: params.estimatedSecondsToFinish,
      requestId: params.requestId ?? newProductionRequestId(),
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['productionJobs'] })
      qc.invalidateQueries({ queryKey: ['productionQueue'] })
      qc.invalidateQueries({ queryKey: ['company'] })
    },
  })
}

export function useClaimProduction() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (jobId: string) =>
      api.post<{ jobId: string; status: string; output: Record<string, number>; quality: number }>(
        `/api/v2/production/claim/${jobId}/`,
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['productionJobs'] })
      qc.invalidateQueries({ queryKey: ['claimableJobs'] })
      qc.invalidateQueries({ queryKey: ['company'] })
      qc.invalidateQueries({ queryKey: ['playerLevel'] })
      qc.invalidateQueries({ queryKey: ['warehouse'] })
    },
  })
}

export function useClaimAll() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<{ claimed: unknown[]; errors: string[]; total: number }>(
      '/api/v2/production/claim-all/',
    ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['productionJobs'] })
      qc.invalidateQueries({ queryKey: ['claimableJobs'] })
      qc.invalidateQueries({ queryKey: ['company'] })
      qc.invalidateQueries({ queryKey: ['playerLevel'] })
      qc.invalidateQueries({ queryKey: ['warehouse'] })
    },
  })
}

export function useCancelJob() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (jobId: string) =>
      api.post<{ jobId: string; status: string }>('/api/v2/production/cancel/', { jobId }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['productionJobs'] })
      qc.invalidateQueries({ queryKey: ['productionQueue'] })
    },
  })
}
