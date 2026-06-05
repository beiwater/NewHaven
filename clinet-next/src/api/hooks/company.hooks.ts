import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'

export interface CompanyData {
  authCompany: { id: number; name: string; money: number }
  authUser: { id: number; username: string }
  levelInfo: { level: number; xp: number; xpToNext: number }
  preferences: Record<string, unknown>
}

export function useCompany() {
  return useQuery({
    queryKey: queryKeys.company.all(),
    queryFn: () => api.get<CompanyData>('/api/v2/companies/me/'),
  })
}
