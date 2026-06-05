import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { Executive, ExecutiveSearchResult, RecruitResult, TrainResult, ExecutiveDetail } from '@/game/executives'

export function useExecutiveSearch() {
  return useQuery({
    queryKey: queryKeys.executives.all(),
    queryFn: () => api.post<ExecutiveSearchResult>('/api/v2/executives/search/'),
    staleTime: 60_000,
  })
}

export function useRecruitExecutive() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (executiveId: string) =>
      api.post<RecruitResult>('/api/v2/executives/recruit/', { executiveId }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.executives.all() })
      qc.invalidateQueries({ queryKey: queryKeys.executives.my() })
      qc.invalidateQueries({ queryKey: queryKeys.company.all() })
    },
  })
}

export function useMyExecutives() {
  return useQuery({
    queryKey: queryKeys.executives.my(),
    queryFn: () => api.get<Executive[]>('/api/v2/executives/'),
  })
}

export function useTrainExecutive() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (executiveId: string) =>
      api.post<TrainResult>(`/api/v2/executives/train/${executiveId}/`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.executives.my() })
      qc.invalidateQueries({ queryKey: queryKeys.executives.detail(null) })
    },
  })
}

export function useExecutiveDetail(id: string | null) {
  return useQuery({
    queryKey: queryKeys.executives.detail(id),
    queryFn: () => api.get<ExecutiveDetail>(`/api/v3/executives/${id}/`),
    enabled: !!id,
  })
}
