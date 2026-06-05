import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { MarketOrder, MarketDepth, MarketTickerData, ResourceDefinition } from '@/game/types'

export function useMarketTicker(resourceId: number) {
  return useQuery({
    queryKey: queryKeys.market.ticker(resourceId),
    queryFn: () => api.get<MarketTickerData>(`/api/v3/market-ticker/${resourceId}/`),
    refetchInterval: 10_000,
  })
}

export function useMarketDepth(resourceId: number, quality = 0) {
  return useQuery({
    queryKey: queryKeys.market.depth(resourceId, quality),
    queryFn: () => api.get<MarketDepth>(`/api/v3/market-depth/${resourceId}/${quality}/`),
    refetchInterval: 10_000,
  })
}

export function useMarketOrders(resourceId: number, quality = 0) {
  return useQuery({
    queryKey: queryKeys.market.orders(resourceId, quality),
    queryFn: () => api.get<MarketOrder[]>(`/api/v3/market/${resourceId}/${quality}/`),
    refetchInterval: 10_000,
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
      // TODO(audio): play sfx here when audio system is migrated
      qc.invalidateQueries({ queryKey: queryKeys.market.orders(vars.resourceId, vars.quality) })
      qc.invalidateQueries({ queryKey: queryKeys.market.depth(vars.resourceId, vars.quality) })
      qc.invalidateQueries({ queryKey: queryKeys.company.all() })
    },
  })
}

export function useCancelOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (orderId: string) =>
      api.delete<{ id: string; status: string }>(`/api/v2/market-order/cancel/${orderId}/`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.market.orders(0, 0).slice(0, 2) })
      // More precise: invalidate all market orders
      qc.invalidateQueries({ queryKey: ['market'] })
      qc.invalidateQueries({ queryKey: queryKeys.company.all() })
      qc.invalidateQueries({ queryKey: queryKeys.inventory.warehouse() })
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
      // TODO(audio): play sfx here
      qc.invalidateQueries({ queryKey: queryKeys.market.orders(vars.resourceId, vars.quality) })
      qc.invalidateQueries({ queryKey: queryKeys.market.depth(vars.resourceId, vars.quality) })
      qc.invalidateQueries({ queryKey: queryKeys.company.all() })
      qc.invalidateQueries({ queryKey: queryKeys.inventory.warehouse() })
    },
  })
}

export function useResources() {
  return useQuery({
    queryKey: queryKeys.market.resources(),
    queryFn: () => api.get<{ resources: ResourceDefinition[] }>('/api/v3/resources/'),
    staleTime: 10 * 60_000,
  })
}
