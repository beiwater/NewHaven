import { useState, useEffect } from 'react'

interface OHLC {
  open: number
  high: number
  low: number
  close: number
  time: Date
}

const RANGES = ['1h', '6h', '12h', '24h', '48h'] as const
type TimeRange = (typeof RANGES)[number]

const RANGE_MS: Record<TimeRange, number> = {
  '1h': 3_600_000,
  '6h': 21_600_000,
  '12h': 43_200_000,
  '24h': 86_400_000,
  '48h': 172_800_000,
}

const BUCKET_MS: Record<TimeRange, number> = {
  '1h': 60_000,     // 1 minute
  '6h': 300_000,    // 5 minutes
  '12h': 600_000,   // 10 minutes
  '24h': 1_200_000, // 20 minutes
  '48h': 1_800_000, // 30 minutes
}

function buildCandles(
  series: Array<{ price: number; time: string }>,
  range: TimeRange,
): OHLC[] {
  const now = Date.now()
  const cutoff = now - RANGE_MS[range]
  const filtered = series.filter((p) => new Date(p.time).getTime() >= cutoff)
  if (filtered.length === 0) return []

  const bucketSize = BUCKET_MS[range]
  const buckets = new Map<number, { open: number; high: number; low: number; close: number; firstTime: number }>()

  for (const pt of filtered) {
    const t = new Date(pt.time).getTime()
    const bkt = Math.floor(t / bucketSize) * bucketSize
    const existing = buckets.get(bkt)
    if (!existing) {
      buckets.set(bkt, { open: pt.price, high: pt.price, low: pt.price, close: pt.price, firstTime: t })
    } else {
      existing.high = Math.max(existing.high, pt.price)
      existing.low = Math.min(existing.low, pt.price)
      existing.close = pt.price
      if (t < existing.firstTime) {
        existing.open = pt.price
        existing.firstTime = t
      }
    }
  }

  // Fill empty buckets with the last known close price so the chart is continuous.
  const startBkt = Math.floor(cutoff / bucketSize) * bucketSize
  const endBkt = Math.floor(now / bucketSize) * bucketSize
  const numBuckets = Math.floor((endBkt - startBkt) / bucketSize) + 1
  let lastClose = filtered[0].price
  const filled: OHLC[] = []
  for (let i = 0; i < numBuckets; i++) {
    const bkt = startBkt + i * bucketSize
    const existing = buckets.get(bkt)
    if (existing) {
      lastClose = existing.close
      filled.push({
        open: existing.open,
        high: existing.high,
        low: existing.low,
        close: existing.close,
        time: new Date(existing.firstTime),
      })
    } else {
      // Carry last close forward so the line never breaks.
      filled.push({
        open: lastClose,
        high: lastClose,
        low: lastClose,
        close: lastClose,
        time: new Date(bkt),
      })
    }
  }
  return filled
}

export function PriceCurve({
  series,
  selectedResource,
}: {
  series: Array<{ price: number; time: string }>
  selectedResource?: number
}) {
  const [range, setRange] = useState<TimeRange>('48h')

  // Reset zoom when resource changes
  useEffect(() => { setRange('48h') }, [selectedResource])

  const now = Date.now()
  const cutoff = now - RANGE_MS[range]
  const visibleSeries = series.filter((p) => new Date(p.time).getTime() >= cutoff)
  const candles = buildCandles(visibleSeries, range)
  const width = 640
  const height = 200
  const padding = { top: 14, right: 14, bottom: 24, left: 52 }
  const plotW = width - padding.left - padding.right
  const plotH = height - padding.top - padding.bottom

  // Y-axis from candle close prices (continuous even with sparse trades).
  const allPrices = candles.map((c) => c.close)
  const dataMin = allPrices.length ? Math.min(...allPrices) : 0
  const dataMax = allPrices.length ? Math.max(...allPrices) : 0
  const min = dataMin
  const max = dataMax
  const spread = Math.max(1, max - min)

  const yTicks = 5
  const yLabels = Array.from({ length: yTicks }, (_, i) => {
    const val = min + (spread * i) / (yTicks - 1)
    return { value: val, y: padding.top + plotH - ((val - min) / spread) * plotH }
  })

  // Helper: map candle's price to Y
  const toY = (price: number) => padding.top + plotH - ((price - min) / spread) * plotH

  // Background line — uses candle close prices so it's continuous.
  const points = candles
    .map((c, i) => {
      const x = padding.left + (i / Math.max(1, candles.length - 1)) * plotW
      const y = toY(c.close)
      return `${x.toFixed(2)},${y.toFixed(2)}`
    })
    .join(' ')
  const candleW = candles.length > 1 ? (plotW / candles.length) * 0.6 : 4
  const candleSpacing = candles.length > 1 ? plotW / (candles.length - 1) : plotW

  const lastCandle = candles.at(-1)
  return (
    <div className="mt-4 rounded-2xl border border-amber-200/70 bg-gradient-to-br from-amber-50 to-white/70 p-3">
      {/* Header */}
      <div className="mb-2 flex items-center justify-between gap-2">
        <div className="flex items-center gap-3">
          <div>
            <div className="text-[10px] font-black uppercase tracking-wider text-amber-700">K线图</div>
            <div className="text-[9px] text-amber-500">Candlestick</div>
          </div>
          {/* Zoom buttons */}
          <div className="flex gap-1 ml-2">
            {RANGES.map((r) => (
              <button
                key={r}
                onClick={() => setRange(r)}
                className={`px-2 py-1 text-[10px] font-bold rounded transition-colors ${
                  range === r
                    ? 'bg-amber-800 text-white'
                    : 'bg-amber-100 text-amber-700 hover:bg-amber-200'
                }`}
              >
                {r}
              </button>
            ))}
          </div>
        </div>
        {lastCandle && (
          <div className="text-right text-[10px] font-semibold text-amber-700">
            {new Date(lastCandle.time).toLocaleString()}
          </div>
        )}
      </div>

      <svg viewBox={`0 0 ${width} ${height}`} className="h-48 w-full overflow-visible">
        <defs>
          <linearGradient id="marketCurveFill" x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="#d97706" stopOpacity="0.26" />
            <stop offset="100%" stopColor="#d97706" stopOpacity="0.02" />
          </linearGradient>
        </defs>

        {/* Y-axis labels */}
        {yLabels.map(({ value, y }) => (
          <g key={y}>
            <text x={padding.left - 6} y={y + 3} textAnchor="end" className="fill-amber-600 text-[9px] font-semibold">
              ${value.toFixed(2)}
            </text>
            <line
              x1={padding.left} y1={y} x2={width - padding.right} y2={y}
              stroke="#d6b27b" strokeDasharray="4 4" strokeOpacity={0.4}
            />
          </g>
        ))}

        {/* X-axis baseline */}
        <line x1={padding.left} y1={padding.top + plotH} x2={width - padding.right} y2={padding.top + plotH} stroke="#d6b27b" strokeDasharray="4 4" />
        <line x1={padding.left} y1={padding.top} x2={padding.left} y2={padding.top + plotH} stroke="#d6b27b" strokeDasharray="4 4" />

        {candles.length > 0 ? (
          <>
            {/* Background line */}
            {points && (
              <>
                <polyline
                  points={`${padding.left},${padding.top + plotH} ${points} ${width - padding.right},${padding.top + plotH}`}
                  fill="url(#marketCurveFill)" stroke="none"
                />
                <polyline points={points} fill="none" stroke="#b45309" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" opacity={0.5} />
              </>
            )}

            {/* Candlesticks */}
            {candles.map((c, i) => {
              const cx = padding.left + (i / Math.max(1, candles.length - 1)) * plotW
              const bodyTop = toY(Math.max(c.open, c.close))
              const bodyBot = toY(Math.min(c.open, c.close))
              const bodyH = Math.max(1, bodyBot - bodyTop)
              const wickTop = toY(c.high)
              const wickBot = toY(c.low)
              const isBullish = c.close >= c.open

              return (
                <g key={i}>
                  {/* Wick */}
                  <line x1={cx} y1={wickTop} x2={cx} y2={wickBot} stroke={isBullish ? '#16a34a' : '#dc2626'} strokeWidth="1" />
                  {/* Body */}
                  <rect x={cx - candleW / 2} y={bodyTop} width={candleW} height={bodyH} fill={isBullish ? '#16a34a' : '#dc2626'} rx="0.5" />
                </g>
              )
            })}
          </>
        ) : (
          <text x={width / 2} y={height / 2} textAnchor="middle" className="fill-amber-500 text-xs">
            No price history yet
          </text>
        )}
      </svg>

      {/* Bottom stats */}
      <div className="mt-1 flex items-center justify-between text-[10px] font-semibold text-amber-600">
        <span>Low ${min.toFixed(2)}</span>
        {candles.length > 0 && (
          <span className="text-amber-500">
            {candles.length} candles · {range}
          </span>
        )}
        <span>High ${max.toFixed(2)}</span>
      </div>
    </div>
  )
}
