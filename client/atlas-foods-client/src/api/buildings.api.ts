import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './client'
import type { Building } from '@/game/types'

export function useBuildings() {
  return useQuery({
    queryKey: ['buildings'],
    queryFn: () => api.get<Building[]>('/api/v2/companies/me/buildings/'),
  })
}

export function useBuyBuilding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (buildingId: string) =>
      api.post<{ building: Building; cost?: number; money?: number }>('/api/v2/buildings/buy/', { buildingId }),
    onSuccess: (data) => {
      qc.setQueryData<Building[]>(['buildings'], (current = []) => {
        const next = { ...data.building, placed: false }
        if (current.some((building) => building.id === next.id)) {
          return current
        }
        return [...current, next]
      })
      qc.invalidateQueries({ queryKey: ['company'] })
    },
  })
}

export function usePlaceBuilding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { buildingId: string; x: number; y: number }) =>
      api.post<{ building: Building; money: number }>('/api/v2/buildings/place/', params),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buildings'] })
      qc.invalidateQueries({ queryKey: ['company'] })
    },
  })
}

export function useBuildingDetail(id: string | undefined) {
  return useQuery({
    queryKey: ['building', id],
    queryFn: () => api.get<{ building: Building }>(`/api/v1/buildings/${id}/`),
    enabled: !!id,
  })
}
