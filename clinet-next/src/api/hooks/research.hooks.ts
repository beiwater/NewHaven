import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'

export interface ResearchProject {
  id: string; name: string; building?: string
  resourceCost?: Record<number, number>; cashCost?: number; durationHours?: number
  unlockRecipeId?: number; qualityResourceId?: number; unlockPct?: number
  status: string; progress: number; startedAt?: string; completesAt?: string
  producedPerHour?: number; sourcingCost?: number
}

export interface ResearchListResponse { projects: ResearchProject[] }

export function useResearch() {
  return useQuery({
    queryKey: queryKeys.research.all(),
    queryFn: async () => {
      const data = await api.get<ResearchListResponse>('/api/v2/research/')
      return data.projects ?? []
    },
  })
}

export function useResearchProgress() {
  return useQuery({
    queryKey: queryKeys.research.progress(),
    queryFn: () => api.get<{ projects: ResearchProject[] }>('/api/v2/research/progress/'),
    refetchInterval: 15_000,
  })
}

export function useStartResearch() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (projectId: string) =>
      api.post<{ project: ResearchProject; status: string }>('/api/v2/research/start/', { projectId }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.research.all() }),
  })
}

export function useCompleteResearch() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (projectId: string) =>
      api.post<{ ok: boolean; projectId: string; patentsGained: number; qualityImproved: number }>(
        `/api/v2/research/complete/${projectId}/`,
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.research.all() }),
  })
}
