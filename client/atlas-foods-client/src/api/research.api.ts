import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from './client'

export interface QualityResearch {
  resourceId: number
  name: string
  tier: number
  maxQuality: number
  nextQuality?: number
  nextCost?: number
  salesSpeedBonus: number
  nextSalesSpeedPct?: number
}

interface ResearchListResponse {
  research: QualityResearch[]
}

interface UnlockQualityResponse {
  research: {
    resourceId: number
    maxQuality: number
    cost: number
    charged: boolean
    salesSpeedBonus: number
    nextQuality?: number
    nextCost?: number
  }
}

const researchKey = ['research', 'quality'] as const

export function useResearch() {
  return useQuery({
    queryKey: researchKey,
    queryFn: async () => {
      const data = await api.get<ResearchListResponse>('/api/v2/research/')
      return data.research ?? []
    },
  })
}

export function useUnlockQuality() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ resourceId, targetQuality }: { resourceId: number; targetQuality: number }) =>
      api.post<UnlockQualityResponse>('/api/v2/research/quality/', { resourceId, targetQuality }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: researchKey })
      queryClient.invalidateQueries({ queryKey: ['company'] })
    },
  })
}
