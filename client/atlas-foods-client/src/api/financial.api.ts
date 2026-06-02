import { useQuery } from '@tanstack/react-query'
import { api } from './client'

// --- Response types matching backend ---

export interface IncomeStatement {
  revenue: number
  expenses: number
  netIncome: number
}

export interface BalanceSheet {
  assets: number
  liabilities: number
  equity: number
}

export interface CashflowStatement {
  operating: number
  investing: number
  financing: number
}

export interface CashflowEntry {
  kind: string
  moneyDelta: number
  at: string
}

export interface RecentCashflow {
  data: CashflowEntry[]
  oldestPulled: string
  money: number
}

export interface PastFinancePoint {
  date: string
  net: number
}

export interface PastFinances {
  series: PastFinancePoint[]
}

// --- Human-readable labels for ledger kinds ---

export const KIND_LABELS: Record<string, string> = {
  market_trade: 'Market Sales',
  market_fee: 'Market Fees',
  market_take_buy: 'Market Purchases',
  gov_contract_reward: 'Contract Rewards',
  gov_bid_deposit: 'Contract Deposits',
  bond_interest_income: 'Bond Interest',
  bond_interest_expense: 'Bond Interest',
  bond_buy: 'Bond Investment',
  bond_call: 'Bond Call',
  bond_issue: 'Bond Issuance',
  production_input: 'Production Input',
  production_output: 'Production Output',
}

export function kindLabel(kind: string): string {
  return KIND_LABELS[kind] ?? kind.replace(/_/g, ' ')
}

// --- Hooks ---

export function useIncomeStatement() {
  return useQuery({
    queryKey: ['financial', 'income-statement'],
    queryFn: () => api.get<IncomeStatement>('/api/v2/companies/me/income-statement/'),
    refetchInterval: 30_000,
  })
}

export function useBalanceSheet() {
  return useQuery({
    queryKey: ['financial', 'balance-sheet'],
    queryFn: () => api.get<BalanceSheet>('/api/v2/companies/me/balance-sheet/'),
    refetchInterval: 30_000,
  })
}

export function useCashflowStatement() {
  return useQuery({
    queryKey: ['financial', 'cashflow-statement'],
    queryFn: () => api.get<CashflowStatement>('/api/v2/companies/me/cashflow-statement/'),
    refetchInterval: 30_000,
  })
}

export function useRecentCashflow() {
  return useQuery({
    queryKey: ['financial', 'recent-cashflow'],
    queryFn: () => api.get<RecentCashflow>('/api/v2/companies/me/cashflow/recent/'),
    refetchInterval: 30_000,
  })
}

export function usePastFinances() {
  return useQuery({
    queryKey: ['financial', 'past-finances'],
    queryFn: () => api.get<PastFinances>('/api/v3/companies/me/past-finances/'),
    staleTime: 60_000,
  })
}

/** Aggregated financial data for the dashboard overview */
export function useFinancialOverview() {
  const income = useIncomeStatement()
  const balance = useBalanceSheet()
  const cashflow = useCashflowStatement()
  const recent = useRecentCashflow()
  const past = usePastFinances()

  const isLoading = income.isLoading || balance.isLoading || cashflow.isLoading || recent.isLoading
  const error = income.error ?? balance.error ?? cashflow.error ?? recent.error

  return {
    income: income.data,
    balance: balance.data,
    cashflow: cashflow.data,
    recent: recent.data,
    past: past.data,
    isLoading,
    error,
  }
}
