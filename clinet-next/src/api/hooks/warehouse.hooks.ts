import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { WarehouseData } from '@/game/types'

export function useWarehouse() {
  return useQuery({
    queryKey: queryKeys.inventory.warehouse(),
    queryFn: () => api.get<WarehouseData>('/api/v2/companies/me/warehouse/'),
  })
}
