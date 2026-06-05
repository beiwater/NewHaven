import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'

export interface PowerupType { id: string; name: string; desc: string; duration: number }
export interface ActivePowerup { type: string; endsAt: string; remaining: string }

export function usePowerupTypes() {
  return useQuery({
    queryKey: queryKeys.powerups.types(),
    queryFn: () => api.get<{ boosts: PowerupType[] }>('/api/v2/players/simboosts/'),
  })
}

export function useActivePowerup() {
  return useQuery({
    queryKey: queryKeys.powerups.active(),
    queryFn: () => api.get<{ remaining: number; active: ActivePowerup[] }>('/api/v2/players/simboosts-use/'),
    refetchInterval: 30_000,
  })
}

export function useActivatePowerup() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (boostId: string) =>
      api.post<{ boostId: string; endsAt: string; multiplier: number }>('/api/v2/players/simboosts-use/', { boostId }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.powerups.active() })
      qc.invalidateQueries({ queryKey: ['production'] })
    },
  })
}
