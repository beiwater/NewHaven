import { useQuery, useMutation } from '@tanstack/react-query'
import { api, setAuth, getCompanyId } from './client'

interface LoginResponse {
  player?: { id: number; username: string; token: string; companyId: number }
  company?: {
    id: number
    name: string
    money: number
    level: number
    inventory: Record<string, number>
  }
  token?: string
  companyId?: number
}

interface RegisterResponse extends LoginResponse {
  companyID: number
}

export interface CompanyData {
  authCompany: {
    company: string
    companyId: number
    money: number
    level: number
    simBoosts: number
  }
  authUser: {
    isModerator: boolean
    playerId: string
  }
  levelInfo: {
    level: number
    xp: number
  }
  preferences: Record<string, unknown>
}

export interface PlayerLevel {
  level: number
  currentXp: number
  xpToNextLevel: number
  buildingSlots: number
  buildingsUsed: number
}

export function useCompany() {
  const companyId = getCompanyId()
  return useQuery({
    queryKey: ['company', companyId],
    queryFn: () => api.get<CompanyData>(`/api/v3/companies/${companyId}/`),
    refetchInterval: 30_000,
    enabled: !!companyId,
  })
}

export function usePlayerLevel() {
  return useQuery({
    queryKey: ['playerLevel'],
    queryFn: () => api.get<PlayerLevel>('/api/v2/players/me/level/'),
    refetchInterval: 30_000,
  })
}

export function useLogin() {
  return useMutation({
    mutationFn: ({ username, password }: { username: string; password: string }) =>
      api.post<LoginResponse>('/api/login', { username, password }),
    onSuccess: (data) => {
      const token = data.player?.token ?? data.token
      const companyId = data.player?.companyId ?? data.companyId ?? data.company?.id
      if (!token || !companyId) {
        throw new Error('Login response missing token or company id')
      }
      setAuth(token, String(companyId))
    },
  })
}

export function useRegister() {
  return useMutation({
    mutationFn: ({ username, password }: { username: string; password: string }) =>
      api.post<RegisterResponse>('/api/register', { username, password }),
    onSuccess: (data) => {
      const token = data.player?.token ?? data.token
      const companyId = data.companyID ?? data.player?.companyId ?? data.companyId ?? data.company?.id
      if (!token || !companyId) {
        throw new Error('Register response missing token or company id')
      }
      setAuth(token, String(companyId))
    },
  })
}
