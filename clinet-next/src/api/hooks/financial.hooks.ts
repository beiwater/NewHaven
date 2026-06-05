import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'

export interface PastFinancePoint {
  at: string; cash: number; netWorth: number; revenue: number; expenses: number
}

export const KIND_LABELS: Record<string, string> = {
  sale: 'Sale', purchase: 'Purchase', salary: 'Salary',
  research_cost: 'Research', building_cost: 'Building',
  upgrade_cost: 'Upgrade', contract_reward: 'Contract',
  market_fee: 'Market Fee', investment: 'Investment',
  taxes: 'Taxes', other: 'Other',
}

export function kindLabel(kind: string): string {
  return KIND_LABELS[kind] ?? kind.replace(/_/g, ' ')
}

export function useIncomeStatement() {
  return useQuery({
    queryKey: queryKeys.financial.income(),
    queryFn: () => api.get('/api/v2/financial/income/'),
    staleTime: 60_000,
  })
}

export function useBalanceSheet() {
  return useQuery({
    queryKey: queryKeys.financial.balance(),
    queryFn: () => api.get('/api/v2/financial/balance/'),
    staleTime: 60_000,
  })
}

export function useCashflowStatement() {
  return useQuery({
    queryKey: queryKeys.financial.cashflow(),
    queryFn: () => api.get('/api/v2/financial/cashflow/'),
    staleTime: 60_000,
  })
}

export function useRecentCashflow() {
  return useQuery({
    queryKey: queryKeys.financial.recentCashflow(),
    queryFn: () => api.get<{ entries: Array<{ kind: string; amount: number; desc: string; at: string }> }>('/api/v2/financial/recent/'),
    staleTime: 30_000,
  })
}

export function usePastFinances() {
  return useQuery({
    queryKey: queryKeys.financial.pastFinances(),
    queryFn: () => api.get<{ series: PastFinancePoint[] }>('/api/v2/financial/history/'),
  })
}

export function useFinancialOverview() {
  return useQuery({
    queryKey: queryKeys.financial.overview(),
    queryFn: () => api.get('/api/v2/financial/overview/'),
    staleTime: 30_000,
  })
}
