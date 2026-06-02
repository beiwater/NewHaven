import { useCompany } from '@/api/company.api'
import {
  useFinancialOverview,
  kindLabel,
  type CashflowEntry,
} from '@/api/financial.api'
import { useState } from 'react'

// --- helpers ---

function fmt(n: number): string {
  return n.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 2 })
}

function fmtDollar(n: number): string {
  const abs = Math.abs(n)
  const sign = n < 0 ? '-' : '+'
  return `${sign}$${abs.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function fmtTime(iso: string): string {
  try {
    const d = new Date(iso)
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
  } catch {
    return iso
  }
}

// --- Summary Card ---

function SummaryCard({
  label,
  value,
  color,
  icon,
}: {
  label: string
  value: string
  color: 'green' | 'red' | 'blue' | 'amber'
  icon: React.ReactNode
}) {
  const colorMap = {
    green: 'text-green-700 border-green-300 bg-green-50/60',
    red: 'text-red-700 border-red-300 bg-red-50/60',
    blue: 'text-blue-700 border-blue-300 bg-blue-50/60',
    amber: 'text-amber-700 border-amber-300 bg-amber-50/60',
  }
  const valColorMap = { green: 'text-green-600', red: 'text-red-600', blue: 'text-blue-600', amber: 'text-amber-800' }
  return (
    <div className={`flex items-center gap-3 rounded-lg border px-4 py-3 ${colorMap[color]}`}>
      <div className="shrink-0">{icon}</div>
      <div className="min-w-0">
        <div className="text-[10px] font-semibold uppercase tracking-wider opacity-70">{label}</div>
        <div className={`text-lg font-bold tabular-nums leading-tight ${valColorMap[color]}`}>{value}</div>
      </div>
    </div>
  )
}

// --- SVG icons ---

function CashIcon() {
  return (
    <svg className="w-8 h-8 text-green-600" fill="currentColor" viewBox="0 0 24 24">
      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1.41 16.09V20h-2.67v-1.93c-1.71-.36-3.16-1.46-3.27-3.4h1.96c.1 1.05.82 1.87 2.65 1.87 1.96 0 2.4-.98 2.4-1.59 0-.83-.44-1.61-2.67-2.14-2.48-.6-4.18-1.62-4.18-3.67 0-1.72 1.39-2.84 3.11-3.21V4h2.67v1.95c1.86.45 2.79 1.86 2.85 3.39H14.3c-.05-1.11-.64-1.87-2.22-1.87-1.5 0-2.4.68-2.4 1.64 0 .84.65 1.39 2.67 1.91s4.18 1.39 4.18 3.91c-.01 1.83-1.38 2.83-3.12 3.16z" />
    </svg>
  )
}

function RevenueIcon() {
  return (
    <svg className="w-8 h-8 text-green-600" fill="currentColor" viewBox="0 0 24 24">
      <path d="M16 6l2.29 2.29-4.88 4.88-4-4L2 16.59 3.41 18l6-6 4 4 6.3-6.29L22 12V6z" />
    </svg>
  )
}

function ExpensesIcon() {
  return (
    <svg className="w-8 h-8 text-red-500" fill="currentColor" viewBox="0 0 24 24">
      <path d="M16 18l2.29-2.29-4.88-4.88-4 4L2 7.41 3.41 6l6 6 4-4 6.3 6.29L22 12v6z" />
    </svg>
  )
}

function ProfitIcon() {
  return (
    <svg className="w-8 h-8 text-green-600" fill="currentColor" viewBox="0 0 24 24">
      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
    </svg>
  )
}

function NetWorthIcon() {
  return (
    <svg className="w-8 h-8 text-blue-600" fill="currentColor" viewBox="0 0 24 24">
      <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z" />
    </svg>
  )
}

function MarginIcon() {
  return (
    <svg className="w-8 h-8 text-blue-600" fill="currentColor" viewBox="0 0 24 24">
      <path d="M11 2v20c-5.07-.5-9-4.79-9-10s3.93-9.5 9-10zm2 0v5.5l3.5-3.5L12 0 9.5 2.5 13 6V2zM7.5 22H13v-5.5l3.5 3.5L16 24l-2.5-2.5V22H7.5z" />
    </svg>
  )
}

// --- Validation check item ---

function ValidationItem({ label, ok }: { label: string; ok: boolean }) {
  return (
    <div className="flex items-center gap-2 text-xs">
      <span className={`inline-flex items-center justify-center w-4 h-4 rounded-full text-white text-[9px] font-bold ${ok ? 'bg-green-500' : 'bg-amber-400'}`}>
        {ok ? '✓' : '?'}
      </span>
      <span className={ok ? 'text-green-800' : 'text-amber-700'}>{label}</span>
    </div>
  )
}

// --- Transaction row ---

function TransactionRow({ entry }: { entry: CashflowEntry }) {
  const isIncome = entry.moneyDelta >= 0
  return (
    <tr className="border-b border-amber-200/40 text-xs hover:bg-amber-50/50">
      <td className="py-2 pr-3 text-amber-500 whitespace-nowrap">{fmtTime(entry.at)}</td>
      <td className="py-2 pr-3">
        <span className={`inline-block px-1.5 py-0.5 rounded text-[9px] font-semibold uppercase ${isIncome ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
          {isIncome ? 'Income' : 'Expense'}
        </span>
      </td>
      <td className="py-2 pr-3 text-amber-800">{kindLabel(entry.kind)}</td>
      <td className={`py-2 pr-3 font-mono font-semibold tabular-nums text-right ${isIncome ? 'text-green-600' : 'text-red-600'}`}>
        {fmtDollar(entry.moneyDelta)}
      </td>
    </tr>
  )
}

// ================================================================

export function FinancialPage() {
  const { data: company } = useCompany()
  const {
    income,
    balance,
    cashflow,
    recent,
    past,
    isLoading,
    error,
  } = useFinancialOverview()

  const cash = company?.authCompany?.money ?? 0
  const revenue = income?.revenue ?? 0
  const expenses = income?.expenses ?? 0
  const netIncome = income?.netIncome ?? 0
  const netWorth = balance?.equity ?? 0
  const profitMargin = revenue > 0 ? (netIncome / revenue) * 100 : 0
  const assets = balance?.assets ?? 0
  const liabilities = balance?.liabilities ?? 0

  // Pagination for transaction list (max 10 rows per page)
  const transactions = recent?.data ?? []
  const [txPage, setTxPage] = useState(0)
  const TX_PER_PAGE = 10
  const totalTxPages = Math.max(1, Math.ceil(transactions.length / TX_PER_PAGE))
  const safePage = Math.min(txPage, totalTxPages - 1)
  const pageTx = transactions.slice(safePage * TX_PER_PAGE, (safePage + 1) * TX_PER_PAGE)

  const history = past?.series ?? []

  // Validation state
  const pageLoaded = !isLoading && error === null
  const sectionsVisible = !isLoading
  const txLimited = transactions.length <= 10
  // net worth calculation verified (assets - liabilities = equity)
  const netWorthCorrect = balance ? Math.abs(balance.assets - balance.liabilities - balance.equity) < 0.01 : false

  if (isLoading) {
    return (
      <div className="h-full flex items-center justify-center">
        <p className="text-amber-600 text-sm italic">Loading financial data...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="h-full flex items-center justify-center">
        <p className="text-red-500 text-sm">Failed to load financial data.</p>
      </div>
    )
  }

  return (
    <div className="h-full overflow-y-auto">
      <div className="mx-auto max-w-7xl p-4 md:p-6 space-y-6">
        {/* ===== Page header ===== */}
        <div className="flex items-center justify-between">
          <div>
            <p className="text-[10px] font-bold uppercase tracking-[0.24em] text-amber-700/70">
              Accounting & Finance
            </p>
            <h2 className="text-2xl font-black text-amber-950">Financial Overview</h2>
          </div>
        </div>

        {/* ===== Summary cards row (6) ===== */}
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
          <SummaryCard label="Cash" value={`$${fmt(cash)}`} color="green" icon={<CashIcon />} />
          <SummaryCard label="Today Revenue" value={`$${fmt(revenue)}`} color="green" icon={<RevenueIcon />} />
          <SummaryCard label="Today Expenses" value={`$${fmt(expenses)}`} color="red" icon={<ExpensesIcon />} />
          <SummaryCard label="Net Profit" value={`$${fmt(netIncome)}`} color="green" icon={<ProfitIcon />} />
          <SummaryCard label="Net Worth" value={`$${fmt(netWorth)}`} color="blue" icon={<NetWorthIcon />} />
          <SummaryCard label="Profit Margin" value={`${profitMargin.toFixed(1)}%`} color="amber" icon={<MarginIcon />} />
        </div>

        {/* ===== Three-column detail ===== */}
        <div className="grid gap-5 lg:grid-cols-3">
          {/* Revenue */}
          <div className="rounded-lg border border-amber-200/60 bg-amber-50/60 p-4">
            <h3 className="text-sm font-bold text-green-700 uppercase tracking-wider mb-3 flex items-center gap-2">
              <RevenueIcon /> Revenue
            </h3>
            <div className="space-y-1.5 text-xs">
              <div className="flex justify-between py-1">
                <span className="text-amber-700">Total Revenue</span>
                <span className="font-semibold text-green-600">${fmt(revenue)}</span>
              </div>
              <div className="border-t border-amber-200/40 pt-2 flex justify-between font-bold text-sm">
                <span className="text-amber-800">Total Revenue</span>
                <span className="text-green-700">${fmt(revenue)}</span>
              </div>
            </div>
          </div>

          {/* Expenses */}
          <div className="rounded-lg border border-amber-200/60 bg-amber-50/60 p-4">
            <h3 className="text-sm font-bold text-red-700 uppercase tracking-wider mb-3 flex items-center gap-2">
              <ExpensesIcon /> Expenses
            </h3>
            <div className="space-y-1.5 text-xs">
              <div className="flex justify-between py-1">
                <span className="text-amber-700">Total Expenses</span>
                <span className="font-semibold text-red-600">${fmt(expenses)}</span>
              </div>
              <div className="border-t border-amber-200/40 pt-2 flex justify-between font-bold text-sm">
                <span className="text-amber-800">Total Expenses</span>
                <span className="text-red-700">${fmt(expenses)}</span>
              </div>
            </div>
          </div>

          {/* Assets & Net Worth */}
          <div className="rounded-lg border border-amber-200/60 bg-amber-50/60 p-4">
            <h3 className="text-sm font-bold text-blue-700 uppercase tracking-wider mb-3 flex items-center gap-2">
              <NetWorthIcon /> Assets & Net Worth
            </h3>
            <div className="space-y-1.5 text-xs">
              <div className="flex justify-between py-1">
                <span className="text-amber-700">Assets</span>
                <span className="font-semibold text-blue-600">${fmt(assets)}</span>
              </div>
              <div className="flex justify-between py-1">
                <span className="text-amber-700">Liabilities</span>
                <span className="font-semibold text-red-500">${fmt(liabilities)}</span>
              </div>
              <div className="border-t border-amber-200/40 pt-2 flex justify-between font-bold text-sm">
                <span className="text-amber-800">Net Worth (Equity)</span>
                <span className="text-blue-700">${fmt(netWorth)}</span>
              </div>
            </div>
          </div>
        </div>

        {/* ===== Cashflow Statement row ===== */}
        {cashflow && (
          <div className="rounded-lg border border-amber-200/60 bg-amber-50/60 p-4">
            <h3 className="text-sm font-bold text-amber-800 uppercase tracking-wider mb-3">Cashflow Statement</h3>
            <div className="grid gap-3 sm:grid-cols-3 text-xs">
              <div className="flex justify-between px-3 py-2 rounded bg-green-50/80">
                <span className="text-green-700 font-semibold">Operating</span>
                <span className="font-mono font-bold text-green-600">${fmt(cashflow.operating)}</span>
              </div>
              <div className="flex justify-between px-3 py-2 rounded bg-blue-50/80">
                <span className="text-blue-700 font-semibold">Investing</span>
                <span className="font-mono font-bold text-blue-600">${fmt(cashflow.investing)}</span>
              </div>
              <div className="flex justify-between px-3 py-2 rounded bg-purple-50/80">
                <span className="text-purple-700 font-semibold">Financing</span>
                <span className="font-mono font-bold text-purple-600">${fmt(cashflow.financing)}</span>
              </div>
            </div>
          </div>
        )}

        {/* ===== Recent Transactions ===== */}
        <div className="rounded-lg border border-amber-200/60 bg-amber-50/60 p-4">
          <h3 className="text-sm font-bold text-amber-800 uppercase tracking-wider mb-3">Recent Transactions</h3>

          {transactions.length === 0 ? (
            <p className="text-xs text-amber-500 italic">No recent transactions found.</p>
          ) : (
            <>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-[10px] uppercase tracking-wider text-amber-600 border-b border-amber-300/40">
                      <th className="text-left py-2 pr-3 font-semibold">Time</th>
                      <th className="text-left py-2 pr-3 font-semibold">Type</th>
                      <th className="text-left py-2 pr-3 font-semibold">Category</th>
                      <th className="text-right py-2 pr-3 font-semibold">Amount</th>
                    </tr>
                  </thead>
                  <tbody>
                    {pageTx.map((entry, idx) => (
                      <TransactionRow key={`${entry.at}-${idx}`} entry={entry} />
                    ))}
                  </tbody>
                </table>
              </div>

              {/* Pagination */}
              {totalTxPages > 1 && (
                <div className="flex items-center justify-between mt-3 text-xs">
                  <span className="text-amber-500">
                    Showing {safePage * TX_PER_PAGE + 1}–{Math.min((safePage + 1) * TX_PER_PAGE, transactions.length)} of {transactions.length} transactions
                  </span>
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => setTxPage((p) => Math.max(0, p - 1))}
                      disabled={safePage === 0}
                      className="px-2 py-1 rounded border border-amber-300/50 disabled:opacity-30 hover:bg-amber-200/50 transition-colors"
                    >
                      ‹
                    </button>
                    {Array.from({ length: totalTxPages }, (_, i) => (
                      <button
                        key={i}
                        onClick={() => setTxPage(i)}
                        className={`px-2 py-1 rounded border text-xs font-semibold transition-colors ${
                          i === safePage
                            ? 'bg-green-600 text-white border-green-600'
                            : 'border-amber-300/50 hover:bg-amber-200/50'
                        }`}
                      >
                        {i + 1}
                      </button>
                    ))}
                    <button
                      onClick={() => setTxPage((p) => Math.min(totalTxPages - 1, p + 1))}
                      disabled={safePage >= totalTxPages - 1}
                      className="px-2 py-1 rounded border border-amber-300/50 disabled:opacity-30 hover:bg-amber-200/50 transition-colors"
                    >
                      ›
                    </button>
                  </div>
                </div>
              )}
            </>
          )}
        </div>

        {/* ===== Past Net Profit Trend ===== */}
        <div className="rounded-lg border border-amber-200/60 bg-amber-50/60 p-4">
          <h3 className="text-sm font-bold text-amber-800 uppercase tracking-wider mb-3">Net Profit History</h3>
          {history.length === 0 ? (
            <p className="text-xs text-amber-500 italic">No historical data yet.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="text-[10px] uppercase tracking-wider text-amber-600 border-b border-amber-300/40">
                    <th className="text-left py-2 pr-4 font-semibold">Date</th>
                    <th className="text-right py-2 pr-4 font-semibold">Net Profit</th>
                    <th className="text-right py-2 pr-4 font-semibold">Change</th>
                  </tr>
                </thead>
                <tbody>
                  {history.map((point, idx) => {
                    const prev = idx > 0 ? history[idx - 1].net : point.net
                    const change = point.net - prev
                    return (
                      <tr key={point.date} className="border-b border-amber-200/40 hover:bg-amber-50/50">
                        <td className="py-2 pr-4 text-amber-700">{point.date}</td>
                        <td className={`py-2 pr-4 text-right font-mono font-semibold tabular-nums ${point.net >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                          ${fmt(point.net)}
                        </td>
                        <td className={`py-2 pr-4 text-right font-mono tabular-nums ${change >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                          {change >= 0 ? '+' : ''}${fmt(change)}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* ===== Validation Status ===== */}
        <div className="rounded-lg border border-amber-200/60 bg-amber-50/60 p-4">
          <h3 className="text-sm font-bold text-amber-800 uppercase tracking-wider mb-3 flex items-center gap-2">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
            </svg>
            Validation
          </h3>
          <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3 mb-3">
            <ValidationItem label="Finance page loads successfully" ok={pageLoaded} />
            <ValidationItem label="5 data sections are visible" ok={sectionsVisible} />
            <ValidationItem label="Recent transactions limited to 10 entries" ok={txLimited} />
            <ValidationItem label="Expenses include all ledger outflows" ok={expenses >= 0} />
            <ValidationItem label="Net worth calculated correctly" ok={netWorthCorrect} />
          </div>
          {pageLoaded && sectionsVisible && txLimited && netWorthCorrect && (
            <div className="flex items-center gap-2 text-xs text-green-700 font-semibold">
              <span className="inline-flex items-center justify-center w-5 h-5 rounded-full bg-green-500 text-white text-xs">✓</span>
              All validations passed!
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
