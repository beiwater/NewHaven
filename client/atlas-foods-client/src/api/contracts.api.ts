import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './client'
import type { DailyOrder, GovContract } from '@/game/types'

export function useDailyOrders() {
  return useQuery({
    queryKey: ['dailyOrders'],
    queryFn: () =>
      api.get<{ orders: DailyOrder[]; date: string }>('/api/v2/orders/daily/'),
  })
}

export function useCompleteDailyOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) =>
      api.post<{ id: string; status: string }>('/api/v2/orders/daily/complete/', {
        orderId: id,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['dailyOrders'] })
      qc.invalidateQueries({ queryKey: ['warehouse'] })
    },
  })
}

export function useClaimDailyOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) =>
      api.post<{ id: string; cash: number; xp: number }>('/api/v2/orders/daily/claim/', {
        orderId: id,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['dailyOrders'] })
      qc.invalidateQueries({ queryKey: ['company'] })
      qc.invalidateQueries({ queryKey: ['playerLevel'] })
    },
  })
}

export function useGovContracts() {
  return useQuery({
    queryKey: ['govContracts'],
    queryFn: () =>
      api.get<GovContract[]>('/api/v3/government-orders/'),
  })
}

export function useBidContract() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { contractId: string; unitPrice: number }) =>
      api.post('/api/v3/government-orders/bid/', params),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['govContracts'] })
    },
  })
}
