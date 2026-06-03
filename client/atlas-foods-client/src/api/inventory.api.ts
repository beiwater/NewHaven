import { useQuery } from '@tanstack/react-query'
import { api } from './client'

export interface WarehouseData {
  inventory: Array<{ resourceId: number; quantity: number; quality?: number }>
  capacity: number
  used: number
}

export function useWarehouse() {
  return useQuery({
    queryKey: ['warehouse'],
    queryFn: () => api.get<WarehouseData>('/api/v2/companies/me/warehouse/'),
  })
}
