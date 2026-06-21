import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'

export interface CompanyData {
  authCompany: { id: number; company: string; money: number; level?: number; simBoosts?: number }
  authUser: { id: number; username: string }
  levelInfo: { level: number; xp: number; xpToNext: number; tutorialCompleted?: boolean }
  unlocks?: {
    features?: Record<string, boolean>
    featureLevels?: Record<string, number>
  }
  preferences: Record<string, unknown>
}

export function useCompany() {
  return useQuery({
    queryKey: queryKeys.company.all(),
    queryFn: () => api.get<CompanyData>('/api/v3/companies/'),
  })
}

export function useCompleteTutorial() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<{ ok: boolean }>('/api/v2/companies/me/tutorial/', {}),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.company.all() })
    },
  })
}
