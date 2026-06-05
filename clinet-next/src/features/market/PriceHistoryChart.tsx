import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts'

export function PriceHistoryChart({ series }: { series: Array<{ price: number; time: string }> }) {
  if (!series || series.length === 0) {
    return (
      <div className="flex items-center justify-center h-40 text-xs text-amber-400">
        No price data available
      </div>
    )
  }

  const data = series.map((p) => ({
    price: p.price,
    time: new Date(p.time).toLocaleTimeString(),
  }))

  return (
    <div className="mt-4 rounded-2xl border border-amber-200/70 bg-gradient-to-br from-amber-50 to-white/70 p-3">
      <div className="mb-2 flex items-center justify-between">
        <h3 className="text-xs font-black uppercase tracking-wider text-amber-800">Price History</h3>
        <span className="text-[10px] font-semibold text-amber-700">
          {series.length} data points
        </span>
      </div>
      <ResponsiveContainer width="100%" height={160}>
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="#d4a76a" opacity={0.3} />
          <XAxis dataKey="time" tick={{ fontSize: 10, fill: '#8b6914' }} interval="preserveStartEnd" />
          <YAxis tick={{ fontSize: 10, fill: '#8b6914' }} domain={['auto', 'auto']} />
          <Tooltip
            contentStyle={{
              background: '#fef3c7',
              border: '1px solid #d4a76a',
              borderRadius: '8px',
              fontSize: '12px',
            }}
          />
          <Line type="monotone" dataKey="price" stroke="#b45309" strokeWidth={2} dot={false} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}
