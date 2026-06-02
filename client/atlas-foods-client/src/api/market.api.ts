import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './client'
import type { MarketOrder, MarketDepth, MarketTickerData, ResourceDefinition } from '@/game/types'

// /api/v3/market-ticker/{id}/ returns { resource, series: [{price, time}] }
export function useMarketTicker(resourceId: number) {
  return useQuery({
    queryKey: ['marketTicker', resourceId],
    queryFn: () => api.get<MarketTickerData>(`/api/v3/market-ticker/${resourceId}/`),
    refetchInterval: 30_000,
  })
}

// /api/v3/market-depth/{id}/{quality}/ returns { buys, sells }
export function useMarketDepth(resourceId: number, quality = 0) {
  return useQuery({
    queryKey: ['marketDepth', resourceId, quality],
    queryFn: () => api.get<MarketDepth>(`/api/v3/market-depth/${resourceId}/${quality}/`),
    refetchInterval: 15_000,
  })
}

// /api/v3/market/{id}/{quality}/ returns MarketOrder[] directly
export function useMarketOrders(resourceId: number, quality = 0) {
  return useQuery({
    queryKey: ['marketOrders', resourceId, quality],
    queryFn: () =>
      api.get<MarketOrder[]>(`/api/v3/market/${resourceId}/${quality}/`),
    refetchInterval: 15_000,
  })
}

export function useCreateOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (order: {
      resourceId: number
      kind: 0 | 1
      quality: number
      quantity: number
      price: number
    }) => api.post<{ order: MarketOrder }>('/api/v2/market-order/', order),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: ['marketOrders', vars.resourceId] })
      qc.invalidateQueries({ queryKey: ['marketDepth', vars.resourceId] })
      qc.invalidateQueries({ queryKey: ['company'] })
    },
  })
}

export function useCancelOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (orderId: string) =>
      api.delete<{ id: string; status: string }>(`/api/v2/market-order/cancel/${orderId}/`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['marketOrders'] })
      qc.invalidateQueries({ queryKey: ['marketDepth'] })
      qc.invalidateQueries({ queryKey: ['company'] })
      qc.invalidateQueries({ queryKey: ['warehouse'] })
    },
  })
}

export function useTakeOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: {
      resourceId: number
      quantity: number
      quality: number
      maxPrice: number
    }) => api.post<{ amountBought: number; trades: unknown[]; moneyDelta: number }>(
      '/api/v2/market-order/take/',
      {
        resource: params.resourceId,
        quantity: params.quantity,
        quality: params.quality,
        maxPrice: params.maxPrice,
      },
    ),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: ['marketOrders', vars.resourceId] })
      qc.invalidateQueries({ queryKey: ['marketDepth', vars.resourceId] })
      qc.invalidateQueries({ queryKey: ['company'] })
      qc.invalidateQueries({ queryKey: ['warehouse'] })
    },
  })
}

export function useResources() {
  return useQuery({
    queryKey: ['resources'],
    queryFn: () => api.get<{ resources: ResourceDefinition[] }>('/api/v3/resources/'),
    staleTime: 10 * 60_000,
  })
}
