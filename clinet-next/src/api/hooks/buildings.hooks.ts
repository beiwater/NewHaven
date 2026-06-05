import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { Building } from '@/game/types'
import type { MapId } from '@/game/map.config'

export function useBuildings() {
  return useQuery({
    queryKey: queryKeys.buildings.all(),
    queryFn: () => api.get<Building[]>('/api/v2/companies/me/buildings/'),
  })
}

export function useCompanyBuildings(companyId: number | null) {
  return useQuery({
    queryKey: queryKeys.buildings.byCompany(companyId),
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
      qc.setQueryData<Building[]>(queryKeys.buildings.all(), (current = []) => {
        const next = { ...data.building, placed: false }
        if (current.some((building) => building.id === next.id)) {
          return current
        }
        return [...current, next]
      })
      qc.invalidateQueries({ queryKey: queryKeys.company.all() })
      qc.invalidateQueries({ queryKey: queryKeys.buildings.all() })
    },
  })
}

export function usePlaceBuilding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { buildingId: string; x?: number; y?: number; mapId?: MapId; slotId?: string }) =>
      api.post<{ building: Building; money: number }>('/api/v2/buildings/place/', params),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.buildings.all() })
      qc.invalidateQueries({ queryKey: queryKeys.company.all() })
    },
  })
}

export function useMoveBuilding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { buildingId: string; x?: number; y?: number; mapId?: MapId; slotId?: string }) =>
      api.post<{ building: Building }>('/api/v2/buildings/move/', params),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.buildings.all() })
    },
  })
}

export function useUpgradeBuilding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (buildingId: string) =>
      api.post<{ buildingId: string; oldLevel: number; newLevel: number; cost: number; outputMultiplier: number }>(
        `/api/v1/buildings/${buildingId}/upgrade/`,
      ),
    onSuccess: (data) => {
      qc.setQueryData<Building[]>(queryKeys.buildings.all(), (current = []) =>
        current.map((building) =>
          building.id === data.buildingId
            ? { ...building, level: data.newLevel }
            : building,
        ),
      )
      qc.invalidateQueries({ queryKey: queryKeys.buildings.all() })
      qc.invalidateQueries({ queryKey: queryKeys.company.all() })
      qc.invalidateQueries({ queryKey: queryKeys.buildings.productionOptions(undefined) })
    },
  })
}

export function useDemolishBuilding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (buildingId: string) =>
      api.post<{ refund: number; status: string }>('/api/v2/buildings/demolish/', { buildingId }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.buildings.all() })
      qc.invalidateQueries({ queryKey: queryKeys.company.all() })
    },
  })
}
