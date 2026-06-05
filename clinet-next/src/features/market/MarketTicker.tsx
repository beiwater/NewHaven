import { useMarketTicker } from '@/api/hooks/market.hooks'
import { resourceIcon } from '@/game/icons'

const TRACKED_RESOURCE_IDS = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12]

function PriceCard({ resourceId }: { resourceId: number }) {
  const { data } = useMarketTicker(resourceId)
  const series = data?.series ?? []
  const last = series.length > 0 ? series[series.length - 1].price : 0
  const prev = series.length > 1 ? series[series.length - 2].price : last
  const isUp = last >= prev

  return (
    <div className="flex min-w-[176px] items-center gap-3 border-r border-amber-200/60 px-4 py-2">
      <img src={resourceIcon(resourceId)} alt="" className="h-10 w-10 object-contain" loading="lazy" />
      <div>
        <div className="text-[10px] uppercase tracking-wider text-amber-700/60">Resource #{resourceId}</div>
        <div className="text-sm font-bold tabular-nums text-amber-900">${last.toFixed(2)}</div>
      </div>
      <div className={`ml-auto text-xs font-semibold tabular-nums ${isUp ? 'text-green-600' : 'text-red-500'}`}>
        {isUp ? '+' : ''}{prev > 0 ? (((last - prev) / prev) * 100).toFixed(1) : '0.0'}%
      </div>
    </div>
  )
}

export function MarketTicker() {
  return (
    <div className="market-ticker flex items-center overflow-hidden border-t-2 border-amber-700/30 bg-[#f5e6c8]">
      <div className="flex min-w-[150px] items-center gap-2 border-r border-amber-200/60 bg-amber-100/50 px-4 py-2">
        <svg className="h-5 w-5 text-amber-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
        </svg>
        <div>
          <div className="text-[10px] uppercase tracking-wider text-amber-700/60">Food Market</div>
          <div className="text-xs font-bold uppercase text-amber-800">Live Prices</div>
        </div>
      </div>
      <div className="ticker-window flex-1 overflow-hidden">
        <div className="ticker-track flex w-max">
          {[...TRACKED_RESOURCE_IDS, ...TRACKED_RESOURCE_IDS].map((id, index) => (
            <PriceCard key={`${id}-${index}`} resourceId={id} />
          ))}
        </div>
      </div>
    </div>
  )
}
