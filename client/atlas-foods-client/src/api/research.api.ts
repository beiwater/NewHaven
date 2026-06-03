import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './client'

// ------ Types ------

export interface ResearchProject {
  id: string
  name: string
  building?: string
  resourceCost?: Record<number, number>
  cashCost?: number
  durationHours?: number
  unlockRecipeId?: number
  qualityResourceId?: number
  unlockPct?: number
  status: string // 'available' | 'in_progress' | 'completed' | 'locked'
  progress: number
  startedAt?: string
  completesAt?: string
  // Extra fields from the simplified research endpoint
  producedPerHour?: number
  sourcingCost?: number
}

export interface ResearchListResponse {
  projects: ResearchProject[]
}

export interface StartResearchPayload {
  projectId: string
}

export interface StartResearchResponse {
  project: ResearchProject
  status: string
}

export interface ResearchProgressResponse {
  projects: ResearchProject[]
}

export interface CompleteResearchResponse {
  ok: boolean
  projectId: string
  patentsGained: number
  qualityImproved: number
  completedAt: string
}

// ------ Hooks ------

const researchKey = ['research', 'projects'] as const

/** Fetch the research project list (GET /api/v2/research/) */
export function useResearch() {
  return useQuery({
    queryKey: researchKey,
    queryFn: async () => {
      const data = await api.get<ResearchListResponse>('/api/v2/research/')
      return data.projects ?? []
    },
  })
}

/** Fetch live research progress with auto-polling (GET /api/v2/research/progress/) */
export function useResearchProgress() {
  return useQuery({
    queryKey: [...researchKey, 'progress'],
    queryFn: () => api.get<ResearchProgressResponse>('/api/v2/research/progress/'),
    refetchInterval: 15_000,
  })
}

/** Start a research project (POST /api/v2/research/start/) */
export function useStartResearch() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: StartResearchPayload) =>
      api.post<StartResearchResponse>('/api/v2/research/start/', payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: researchKey })
    },
  })
}

/** Complete a research project (POST /api/v2/research/complete/{id}/) */
export function useCompleteResearch() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (projectId: string) =>
      api.post<CompleteResearchResponse>(`/api/v2/research/complete/${projectId}/`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: researchKey })
    },
  })
}
