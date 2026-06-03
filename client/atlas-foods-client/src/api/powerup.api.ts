import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './client'

/** Available power-up type from backend */
export interface PowerupType {
  id: string
  name: string
  desc: string
  duration: number
}

/** Active power-up returned by backend */
export interface ActivePowerup {
  type: string
  endsAt: string
  remaining: string
}

/** Response from GET /api/v2/players/simboosts/ */
interface BoostTypesResponse {
  boosts: PowerupType[]
}

/** Response from GET /api/v2/players/simboosts-use/ */
interface BoostStatusResponse {
  remaining: number
  active: ActivePowerup[]
}

/** Fetch available power-up types */
export function usePowerupTypes() {
  return useQuery({
    queryKey: ['powerupTypes'],
    queryFn: () =>
      api.get<BoostTypesResponse>('/api/v2/players/simboosts/'),
  })
}

/** Fetch active power-up and remaining uses */
export function useActivePowerup() {
  return useQuery({
    queryKey: ['activePowerup'],
    queryFn: () =>
      api.get<BoostStatusResponse>('/api/v2/players/simboosts-use/'),
    refetchInterval: 30_000,
  })
}

/** Activate a power-up */
export function useActivatePowerup() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (boostId: string) =>
      api.post<{ boostId: string; endsAt: string; multiplier: number }>(
        '/api/v2/players/simboosts-use/',
        { boostId },
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['activePowerup'] })
      qc.invalidateQueries({ queryKey: ['production'] })
    },
  })
}
