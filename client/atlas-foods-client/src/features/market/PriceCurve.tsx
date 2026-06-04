/**
 * PriceCurve — sparkline chart with y-axis price labels and X-axis time labels.
 * The left side shows price levels (low, mid, high) so you can read approximate values.
 */
export function PriceCurve({ series }: { series: Array<{ price: number; time: string }> }) {
  const width = 640
  const height = 160
  const padding = { top: 14, right: 14, bottom: 24, left: 52 }
  const plotW = width - padding.left - padding.right
  const plotH = height - padding.top - padding.bottom

  const prices = series.map((p) => p.price)
  const min = prices.length ? Math.min(...prices) : 0
  const max = prices.length ? Math.max(...prices) : 0
  const spread = Math.max(1, max - min)

  // Y-axis ticks: 5 evenly-spaced labels
  const yTicks = 5
  const yLabels = Array.from({ length: yTicks }, (_, i) => {
    const val = min + (spread * i) / (yTicks - 1)
    return { value: val, y: padding.top + plotH - ((val - min) / spread) * plotH }
  })

  // Polyline points
  const points = series
    .map((point, index) => {
      const x = padding.left + (index / Math.max(1, series.length - 1)) * plotW
      const y = padding.top + plotH - ((point.price - min) / spread) * plotH
      return `${x.toFixed(2)},${y.toFixed(2)}`
    })
    .join(' ')

  const lastPoint = series.at(-1)

  return (
    <div className="mt-4 rounded-2xl border border-amber-200/70 bg-gradient-to-br from-amber-50 to-white/70 p-3">
      <div className="mb-2 flex items-center justify-between gap-2">
        <div>
          <div className="text-[10px] font-black uppercase tracking-wider text-amber-700">Price Curve</div>
          <div className="text-xs text-amber-600">48-hour synthetic + trade history ticker</div>
        </div>
        {lastPoint && (
          <div className="text-right text-[10px] font-semibold text-amber-700">
            Last update {new Date(lastPoint.time).toLocaleTimeString()}
          </div>
        )}
      </div>
      <svg viewBox={`0 0 ${width} ${height}`} className="h-40 w-full overflow-visible">
        <defs>
          <linearGradient id="marketCurveFill" x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="#d97706" stopOpacity="0.26" />
            <stop offset="100%" stopColor="#d97706" stopOpacity="0.02" />
          </linearGradient>
        </defs>

        {/* Y-axis labels */}
        {yLabels.map(({ value, y }) => (
          <g key={y}>
            <text
              x={padding.left - 6}
              y={y + 3}
              textAnchor="end"
              className="fill-amber-600 text-[9px] font-semibold"
            >
              ${value.toFixed(2)}
            </text>
            <line
              x1={padding.left}
              y1={y}
              x2={width - padding.right}
              y2={y}
              stroke="#d6b27b"
              strokeDasharray="4 4"
              strokeOpacity={0.4}
            />
          </g>
        ))}

        {/* X-axis baseline */}
        <line
          x1={padding.left}
          y1={padding.top + plotH}
          x2={width - padding.right}
          y2={padding.top + plotH}
          stroke="#d6b27b"
          strokeDasharray="4 4"
        />
        <line
          x1={padding.left}
          y1={padding.top}
          x2={padding.left}
          y2={padding.top + plotH}
          stroke="#d6b27b"
          strokeDasharray="4 4"
        />

        {points ? (
          <>
            <polyline
              points={`${padding.left},${padding.top + plotH} ${points} ${width - padding.right},${padding.top + plotH}`}
              fill="url(#marketCurveFill)"
              stroke="none"
            />
            <polyline points={points} fill="none" stroke="#b45309" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round" />
          </>
        ) : (
          <text x={width / 2} y={height / 2} textAnchor="middle" className="fill-amber-500 text-xs">
            No price history yet
          </text>
        )}
      </svg>
      <div className="mt-1 flex justify-between text-[10px] font-semibold text-amber-600">
        <span>Low ${min.toFixed(2)}</span>
        <span>High ${max.toFixed(2)}</span>
      </div>
    </div>
  )
}
