import { useQuery } from '@tanstack/react-query'
import { api } from './client'
import { normalizeWarehouseData } from './compat'

export interface WarehouseData {
  inventory: Array<{ resourceId: number; quantity: number; quality?: number }>
  capacity: number
  used: number
}

export function useWarehouse() {
  return useQuery({
    queryKey: ['warehouse'],
    queryFn: async () => normalizeWarehouseData(await api.get<unknown>('/api/v2/companies/me/warehouse/')),
  })
}
