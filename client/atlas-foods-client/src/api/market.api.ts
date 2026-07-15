import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './client'
import type { MarketOrder, MarketDepth, MarketTickerData, ResourceDefinition } from '@/game/types'
import { audio } from '@/audio/AudioManager'

function newMarketRequestId() {
  return globalThis.crypto?.randomUUID?.() ?? `market-${Date.now()}-${Math.random().toString(36).slice(2)}`
}

// /api/v3/market-ticker/{id}/ returns { resource, series: [{price, time}] }
export function useMarketTicker(resourceId: number, enabled = true) {
  return useQuery({
    queryKey: ['marketTicker', resourceId],
    queryFn: () => api.get<MarketTickerData>(`/api/v3/market-ticker/${resourceId}/`),
    refetchInterval: 10_000,
    enabled,
  })
}

interface AllMarketTickersResponse {
  tickers: MarketTickerData[]
}

export function useAllMarketTickers() {
  return useQuery({
    queryKey: ['allMarketTickers'],
    queryFn: () => api.get<AllMarketTickersResponse>('/api/v3/market-ticker/'),
    staleTime: 15_000,
    refetchInterval: 30_000,
  })
}

// /api/v3/market-depth/{id}/{quality}/ returns { buys, sells }

// /api/v3/companies/me/orders/ returns all active orders for the current company.
export function useMyOrders() {
  return useQuery({
    queryKey: ['myOrders'],
    queryFn: async () => {
      const resp = await api.get<{ orders: MarketOrder[] }>('/api/v3/companies/me/orders/');
      return resp.orders ?? [];
    },
    refetchInterval: 10_000,
  })
}
export function useMarketDepth(resourceId: number, quality = 0, enabled = true) {
  return useQuery({
    queryKey: ['marketDepth', resourceId, quality],
    queryFn: () => api.get<MarketDepth>(`/api/v3/market-depth/${resourceId}/${quality}/`),
    refetchInterval: 10_000,
    enabled,
  })
}

// /api/v3/market/{id}/{quality}/ returns { orders: MarketOrder[] }
export function useMarketOrders(resourceId: number, quality = 0, enabled = true) {
  return useQuery({
    queryKey: ['marketOrders', resourceId, quality],
    queryFn: async () => {
      const resp = await api.get<{ orders: MarketOrder[] }>(`/api/v3/market/${resourceId}/${quality}/`);
      return resp.orders ?? [];
    },
    refetchInterval: 10_000,
    enabled,
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
      requestId?: string
    }) => api.post<{ order: MarketOrder }>('/api/v2/market-order/', {
      ...order,
      requestId: order.requestId ?? newMarketRequestId(),
    }),
    onSuccess: (_, vars) => {
      audio.playSfx(vars.kind === 0 ? 'money_coin_spend' : 'money_coin_gain')
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
        requestId: newMarketRequestId(),
      },
    ),
    onSuccess: (_, vars) => {
      audio.playSfx('money_coin_spend')
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
