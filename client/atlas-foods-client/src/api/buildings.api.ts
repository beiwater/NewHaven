import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './client'
import type { Building } from '@/game/types'
import type { MapId } from '@/game/map.config'
import { camelBuildingResponse, normalizeBuildingList } from './compat'

export function useBuildings() {
  return useQuery({
    queryKey: ['buildings'],
    queryFn: async () => normalizeBuildingList(await api.get<unknown>('/api/v2/companies/me/buildings/')),
  })
}

export function useCompanyBuildings(companyId: number | null) {
  return useQuery({
    queryKey: ['company-buildings', companyId],
    queryFn: async () => normalizeBuildingList(await api.get<unknown>(`/api/v2/companies/${companyId}/buildings/`)),
    enabled: companyId !== null,
    staleTime: 30_000,
  })
}

export function useBuyBuilding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (buildingId: string) =>
      api.post<Record<string, unknown>>('/api/v2/buildings/buy/', { buildingId }).then(camelBuildingResponse),
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
    mutationFn: (params: { buildingId: string; x?: number; y?: number; mapId?: MapId; slotId?: string }) =>
      api.post<Record<string, unknown>>('/api/v2/buildings/place/', params).then(camelBuildingResponse),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buildings'] })
      qc.invalidateQueries({ queryKey: ['company'] })
    },
  })
}

export function useMoveBuilding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { buildingId: string; x?: number; y?: number; mapId?: MapId; slotId?: string }) =>
      api.post<Record<string, unknown>>('/api/v2/buildings/move/', params).then(camelBuildingResponse),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buildings'] })
    },
  })
}

export function useUpgradeBuilding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (buildingId: string) =>
      api.post<Record<string, unknown>>(`/api/v1/buildings/${buildingId}/upgrade/`).then((data) => ({
        buildingId: String(data.buildingId ?? data.building_id ?? ''),
        oldLevel: Number(data.oldLevel ?? data.old_level ?? 0),
        newLevel: Number(data.newLevel ?? data.new_level ?? 0),
        cost: Number(data.cost ?? 0),
        outputMultiplier: Number(data.outputMultiplier ?? data.output_multiplier ?? 0),
      })),
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

export function useStashBuilding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (buildingId: string) =>
      api.post<Record<string, unknown>>('/api/v2/buildings/stash/', { buildingId }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buildings'] })
    },
  })
}
