import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './client'
import type { Building } from '@/game/types'

export function useBuildings() {
  return useQuery({
    queryKey: ['buildings'],
    queryFn: () => api.get<Building[]>('/api/v2/companies/me/buildings/'),
  })
}

export function useCompanyBuildings(companyId: number | null) {
  return useQuery({
    queryKey: ['company-buildings', companyId],
    queryFn: () => api.get<Building[]>(`/api/v2/companies/${companyId}/buildings/`),
    enabled: companyId !== null,
    staleTime: 30_000,
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
      qc.invalidateQueries({ queryKey: ['buildings'] })
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

export function useMoveBuilding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { buildingId: string; x: number; y: number }) =>
      api.post<{ building: Building }>('/api/v2/buildings/move/', params),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buildings'] })
    },
  })
}

export function useUpgradeBuilding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (buildingId: string) =>
      api.post<{
        buildingId: string
        oldLevel: number
        newLevel: number
        cost: number
        outputMultiplier: number
      }>(`/api/v1/buildings/${buildingId}/upgrade/`),
    onSuccess: (data) => {
      qc.setQueryData<Building[]>(['buildings'], (current = []) =>
        current.map((building) =>
          building.id === data.buildingId
            ? { ...building, level: data.newLevel }
            : building,
        ),
      )
      qc.invalidateQueries({ queryKey: ['buildings'] })
      qc.invalidateQueries({ queryKey: ['company'] })
      qc.invalidateQueries({ queryKey: ['productionOptions'] })
    },
  })
}

export function useDemolishBuilding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (buildingId: string) =>
      api.post<{ refund: number; status: string }>('/api/v2/buildings/demolish/', { buildingId }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buildings'] })
      qc.invalidateQueries({ queryKey: ['company'] })
    },
  })
}

