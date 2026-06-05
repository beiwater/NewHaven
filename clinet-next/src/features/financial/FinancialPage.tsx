import { useIncomeStatement, useBalanceSheet, useRecentCashflow, usePastFinances, kindLabel } from '@/api/hooks/financial.hooks'
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function FinancialPage() {
  const { data: income } = useIncomeStatement()
  const { data: balance } = useBalanceSheet()
  const { data: recent } = useRecentCashflow()
  const { data: past } = usePastFinances()

  // Cast from unknown — the API returns these shapes but the query type is unparameterized
  const balanceData = balance as unknown as { cash?: number } | undefined
  const incomeData = income as unknown as { revenue?: number; expenses?: number } | undefined
  const cash = balanceData?.cash ?? 0
  const revenue = incomeData?.revenue ?? 0
  const expenses = incomeData?.expenses ?? 0

  return (
    <div className="p-4 space-y-4">
      <h2 className="text-lg font-bold text-amber-900">Financial</h2>

      {/* Summary cards */}
      <div className="grid grid-cols-3 gap-3">
        <Card><CardContent className="p-3">
          <div className="text-[10px] text-amber-600 uppercase">Cash</div>
          <div className="text-lg font-bold text-amber-900">${cash.toLocaleString()}</div>
        </CardContent></Card>
        <Card><CardContent className="p-3">
          <div className="text-[10px] text-amber-600 uppercase">Revenue</div>
          <div className="text-lg font-bold text-green-700">${revenue.toLocaleString()}</div>
        </CardContent></Card>
        <Card><CardContent className="p-3">
          <div className="text-[10px] text-amber-600 uppercase">Expenses</div>
          <div className="text-lg font-bold text-red-600">${expenses.toLocaleString()}</div>
        </CardContent></Card>
      </div>

      {/* Chart */}
      {past?.series && past.series.length > 0 && (
        <Card>
          <CardHeader><CardTitle>Net Worth History</CardTitle></CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={200}>
              <LineChart data={past.series.map((p) => ({ ...p, time: new Date(p.at).toLocaleDateString() }))}>
                <CartesianGrid strokeDasharray="3 3" stroke="#d4a76a" opacity={0.3} />
                <XAxis dataKey="time" tick={{ fontSize: 10 }} />
                <YAxis tick={{ fontSize: 10 }} />
                <Tooltip />
                <Line type="monotone" dataKey="netWorth" stroke="#b45309" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      )}

      {/* Recent transactions */}
      {recent?.entries && recent.entries.length > 0 && (
        <Card>
          <CardHeader><CardTitle>Recent Transactions</CardTitle></CardHeader>
          <CardContent>
            <div className="space-y-1">
              {recent.entries.slice(0, 10).map((e, i) => (
                <div key={i} className="flex justify-between text-xs py-1 border-b border-amber-100">
                  <span className="text-amber-700">{kindLabel(e.kind)}</span>
                  <span className={e.amount >= 0 ? 'text-green-600' : 'text-red-600'}>
                    {e.amount >= 0 ? '+' : ''}${Math.abs(e.amount).toLocaleString()}
                  </span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
