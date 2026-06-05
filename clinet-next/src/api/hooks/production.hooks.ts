import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { ProductionJob, ProductionQueue, ResourceDefinition } from '@/game/types'

export function useProductionJobs() {
  return useQuery({
    queryKey: queryKeys.production.jobs(),
    queryFn: () => api.get<ProductionJob[]>('/api/v2/production/jobs/'),
    refetchInterval: 10_000,
  })
}

export function useProductionQueue() {
  return useQuery({
    queryKey: queryKeys.production.queue(),
    queryFn: () => api.get<ProductionQueue>('/api/v2/production/queue/'),
    refetchInterval: 10_000,
  })
}

export function useClaimableJobs() {
  return useQuery({
    queryKey: queryKeys.production.claimable(),
    queryFn: () => api.get<ProductionJob[]>('/api/v2/production/claimable/'),
    refetchInterval: 10_000,
  })
}

export function useProductionOptions(buildingId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.buildings.productionOptions(buildingId),
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
      estimatedSecondsToFinish?: number
    }) => api.post(`/api/v1/buildings/${params.buildingId}/busy/`, {
      kind: params.kind,
      amount: params.amount,
      estimatedSecondsToFinish: params.estimatedSecondsToFinish,
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.production.jobs() })
      qc.invalidateQueries({ queryKey: queryKeys.production.queue() })
      qc.invalidateQueries({ queryKey: queryKeys.company.all() })
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
      qc.invalidateQueries({ queryKey: queryKeys.production.jobs() })
      qc.invalidateQueries({ queryKey: queryKeys.production.claimable() })
      qc.invalidateQueries({ queryKey: queryKeys.company.all() })
      qc.invalidateQueries({ queryKey: queryKeys.inventory.warehouse() })
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
      qc.invalidateQueries({ queryKey: queryKeys.production.jobs() })
      qc.invalidateQueries({ queryKey: queryKeys.production.claimable() })
      qc.invalidateQueries({ queryKey: queryKeys.company.all() })
      qc.invalidateQueries({ queryKey: queryKeys.inventory.warehouse() })
    },
  })
}

export function useCancelJob() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (jobId: string) =>
      api.post<{ jobId: string; status: string }>('/api/v2/production/cancel/', { jobId }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.production.jobs() })
      qc.invalidateQueries({ queryKey: queryKeys.production.queue() })
    },
  })
}
