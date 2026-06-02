import { useMemo, useState } from 'react'
import { getCompanyId } from '@/api/client'
import { useCompany } from '@/api/company.api'
import { useWarehouse } from '@/api/inventory.api'
import {
  useMarketTicker,
  useMarketDepth,
  useMarketOrders,
  useCreateOrder,
  useCancelOrder,
  useTakeOrder,
  useResources,
} from '@/api/market.api'
import {
  FALLBACK_MARKET_RESOURCES,
  MARKET_GROUPS,
  formatResourceName,
  resourceIcon,
} from '@/game/resources'

export function MarketPage() {
  const [selectedResource, setSelectedResource] = useState(121)
  const [selectedGroup, setSelectedGroup] = useState<(typeof MARKET_GROUPS)[number]['id']>('kitchen')
  const [orderKind, setOrderKind] = useState<'buy' | 'sell'>('buy')
  const [quantity, setQuantity] = useState('10')
  const [price, setPrice] = useState('10.00')

  const { data: resourcesData } = useResources()
  const resources = resourcesData?.resources?.length ? resourcesData.resources : FALLBACK_MARKET_RESOURCES
  const activeGroup = MARKET_GROUPS.find((g) => g.id === selectedGroup) ?? MARKET_GROUPS[0]
  const visibleResources = useMemo(
    () => activeGroup.ids.map((id) => resources.find((r) => r.resourceId === id)).filter(Boolean),
    [activeGroup, resources],
  )

  const { data: ticker } = useMarketTicker(selectedResource)
  const { data: depth } = useMarketDepth(selectedResource)
  const { data: orders } = useMarketOrders(selectedResource)
  const { data: companyData } = useCompany()
  const { data: warehouse } = useWarehouse()
  const createOrder = useCreateOrder()
  const cancelOrder = useCancelOrder()
  const takeOrder = useTakeOrder()

  const series = ticker?.series ?? []
  const lastPrice = series.length > 0 ? series[series.length - 1].price : 0
  const prevPrice = series.length > 1 ? series[series.length - 2].price : lastPrice
  const high = series.length > 0 ? Math.max(...series.map((s) => s.price)) : 0
  const low = series.length > 0 ? Math.min(...series.map((s) => s.price)) : 0
  const change = prevPrice > 0 ? ((lastPrice - prevPrice) / prevPrice) * 100 : 0

  const orderList = Array.isArray(orders) ? orders : []
  const currentCompanyId = Number(getCompanyId())
  const myOrders = orderList.filter((o) => o.companyId === currentCompanyId && o.status !== 'filled')
  const activeOrders = orderList.filter((o) => o.remaining > 0 && o.status !== 'filled' && o.status !== 'cancelled')
  const sellOrders = activeOrders
    .filter((o) => o.kind === 0)
    .sort((a, b) => a.price - b.price)
  const buyOrders = activeOrders
    .filter((o) => o.kind === 1)
    .sort((a, b) => b.price - a.price)
  const bestBid = depth?.buys?.[0]?.price ?? 0
  const bestAsk = depth?.sells?.[0]?.price ?? 0
  const selectedName = formatResourceName(selectedResource, resources)
  const selectedInventory = warehouse?.inventory
    ?.filter((item) => item.resourceId === selectedResource && (item.quality ?? 0) === 0)
    .reduce((sum, item) => sum + item.quantity, 0) ?? 0
  const cash = companyData?.authCompany?.money ?? 0

  const numericQuantity = Math.max(1, parseInt(quantity, 10) || 1)
  const numericPrice = Math.max(0.01, parseFloat(price) || 1)
  const orderValue = numericQuantity * numericPrice
  const localOrderError = orderKind === 'sell' && numericQuantity > selectedInventory
    ? `Not enough ${selectedName}. You have ${selectedInventory.toLocaleString()} available.`
    : orderKind === 'buy' && orderValue > cash
      ? `Not enough cash. Need $${orderValue.toLocaleString()}, have $${cash.toLocaleString()}.`
      : ''
  const mutationError = createOrder.error ?? takeOrder.error ?? cancelOrder.error
  const mutationMessage = mutationError instanceof Error ? mutationError.message : ''

  const handleCreateOrder = () => {
    if (localOrderError) return
    createOrder.mutate({
      resourceId: selectedResource,
      kind: orderKind === 'buy' ? 1 : 0,
      quality: 0,
      quantity: numericQuantity,
      price: numericPrice,
    })
  }

  const handleTakeBestAsk = () => {
    takeOrder.mutate({
      resourceId: selectedResource,
      quantity: numericQuantity,
      quality: 0,
      maxPrice: bestAsk || numericPrice,
    })
  }

  return (
    <div className="h-full overflow-y-auto p-4 md:p-6">
      <div className="mx-auto max-w-6xl space-y-4">
        <div className="flex flex-wrap items-end justify-between gap-3">
          <div>
            <p className="text-[10px] font-bold uppercase tracking-[0.24em] text-amber-700/70">
              Commodity Exchange
            </p>
            <h2 className="text-2xl font-black text-amber-950">Market</h2>
          </div>
          <div className="rounded-full border border-amber-300/70 bg-white/60 px-3 py-1 text-xs font-semibold text-amber-800">
            {resources.length} tradable resources
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          {MARKET_GROUPS.map((group) => (
            <button
              key={group.id}
              onClick={() => {
                setSelectedGroup(group.id)
                setSelectedResource(group.ids[0])
              }}
              className={`rounded-full px-4 py-2 text-xs font-bold transition-colors ${
                selectedGroup === group.id
                  ? 'bg-amber-800 text-white shadow'
                  : 'bg-white/60 text-amber-800 hover:bg-amber-100'
              }`}
            >
              {group.label}
            </button>
          ))}
        </div>

        <div className="grid gap-4 lg:grid-cols-[280px_1fr]">
          <aside className="rounded-2xl border border-amber-300/60 bg-white/55 p-3 shadow-sm">
            <div className="mb-2 text-[10px] font-bold uppercase tracking-wider text-amber-700">
              Select Product
            </div>
            <div className="grid gap-2">
              {visibleResources.map((resource) => resource && (
                <button
                  key={resource.resourceId}
                  onClick={() => setSelectedResource(resource.resourceId)}
                  className={`flex items-center gap-3 rounded-xl border p-2 text-left transition-colors ${
                    selectedResource === resource.resourceId
                      ? 'border-amber-700 bg-amber-100 text-amber-950'
                      : 'border-amber-200/60 bg-white/45 text-amber-800 hover:bg-white/80'
                  }`}
                >
                  <img src={resourceIcon(resource.resourceId)} alt="" className="h-8 w-8 object-contain" />
                  <div className="min-w-0">
                    <div className="truncate text-sm font-bold">{resource.name}</div>
                    <div className="text-[10px] text-amber-600">ID {resource.resourceId}</div>
                  </div>
                </button>
              ))}
            </div>
          </aside>

          <main className="space-y-4">
            <section className="rounded-2xl border border-amber-300/60 bg-white/60 p-4 shadow-sm">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="flex items-center gap-3">
                  <img src={resourceIcon(selectedResource)} alt="" className="h-12 w-12 object-contain" />
                  <div>
                    <h3 className="text-xl font-black text-amber-950">{selectedName}</h3>
                    <p className="text-xs text-amber-700">Quality 0 spot market</p>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                  <Metric label="Last" value={`$${lastPrice.toFixed(2)}`} />
                  <Metric label="Change" value={`${change >= 0 ? '+' : ''}${change.toFixed(1)}%`} tone={change >= 0 ? 'up' : 'down'} />
                  <Metric label="Bid" value={bestBid ? `$${bestBid.toFixed(2)}` : '-'} />
                  <Metric label="Ask" value={bestAsk ? `$${bestAsk.toFixed(2)}` : '-'} />
                </div>
                </div>
                <div className="mt-3 grid grid-cols-2 gap-2 text-center sm:grid-cols-4">
                  <Metric label="High" value={`$${high.toFixed(2)}`} subtle />
                  <Metric label="Low" value={`$${low.toFixed(2)}`} subtle />
                  <Metric label="Inventory" value={`${selectedInventory.toLocaleString()}`} subtle />
                  <Metric label="Cash" value={`$${Math.floor(cash).toLocaleString()}`} subtle />
                </div>
              <PriceCurve series={series} />
            </section>

            <section className="grid gap-4 xl:grid-cols-[1fr_360px]">
              <div className="grid gap-4 sm:grid-cols-2">
                <OrderBookSide title="Buy Orders" tone="buy" rows={depth?.buys ?? []} />
                <OrderBookSide title="Sell Orders" tone="sell" rows={depth?.sells ?? []} />
              </div>

              <div className="rounded-2xl border border-amber-300/60 bg-white/60 p-4 shadow-sm">
                <div className="mb-3 flex gap-2">
                  <button
                    onClick={() => setOrderKind('buy')}
                    className={`flex-1 rounded-lg px-3 py-2 text-xs font-black ${orderKind === 'buy' ? 'bg-green-700 text-white' : 'bg-amber-100 text-amber-800'}`}
                  >
                    Limit Buy
                  </button>
                  <button
                    onClick={() => setOrderKind('sell')}
                    className={`flex-1 rounded-lg px-3 py-2 text-xs font-black ${orderKind === 'sell' ? 'bg-red-600 text-white' : 'bg-amber-100 text-amber-800'}`}
                  >
                    Limit Sell
                  </button>
                </div>
                <label className="mb-2 block text-[10px] font-bold uppercase tracking-wider text-amber-700">
                  Quantity
                </label>
                <input
                  type="number"
                  min="1"
                  value={quantity}
                  onChange={(e) => setQuantity(e.target.value)}
                  className="mb-3 w-full rounded-lg border border-amber-300 bg-white px-3 py-2 text-sm text-amber-950"
                />
                <label className="mb-2 block text-[10px] font-bold uppercase tracking-wider text-amber-700">
                  Limit Price
                </label>
                <input
                  type="number"
                  min="0.01"
                  step="0.01"
                  value={price}
                  onChange={(e) => setPrice(e.target.value)}
                  className="mb-3 w-full rounded-lg border border-amber-300 bg-white px-3 py-2 text-sm text-amber-950"
                />
                <button
                  onClick={handleCreateOrder}
                  disabled={createOrder.isPending || !!localOrderError}
                  className="w-full rounded-lg bg-amber-800 px-4 py-2.5 text-sm font-black text-white hover:bg-amber-900 disabled:bg-amber-400"
                >
                  {createOrder.isPending ? 'Submitting...' : `Place ${orderKind === 'buy' ? 'Buy' : 'Sell'} Order`}
                </button>
                {(localOrderError || mutationMessage) && (
                  <div className="mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-xs font-semibold text-red-700">
                    {localOrderError || mutationMessage}
                  </div>
                )}
                <button
                  onClick={handleTakeBestAsk}
                  disabled={takeOrder.isPending || !bestAsk}
                  className="mt-2 w-full rounded-lg bg-blue-700 px-4 py-2.5 text-sm font-black text-white hover:bg-blue-800 disabled:bg-blue-300"
                >
                  {takeOrder.isPending ? 'Buying...' : bestAsk ? `Buy Best Ask up to $${bestAsk.toFixed(2)}` : 'No Sell Orders'}
                </button>
              </div>
            </section>

            <section className="grid gap-4 xl:grid-cols-2">
              <ParticipantList
                title="Who is selling"
                emptyText={`No sellers are offering ${selectedName} right now.`}
                orders={sellOrders}
                currentCompanyId={currentCompanyId}
              />
              <ParticipantList
                title="Who is buying"
                emptyText={`No buyers are bidding for ${selectedName} right now.`}
                orders={buyOrders}
                currentCompanyId={currentCompanyId}
              />
            </section>

            <section className="rounded-2xl border border-amber-300/60 bg-white/60 p-4 shadow-sm">
              <h3 className="mb-2 text-xs font-black uppercase tracking-wider text-amber-800">
                My Open Orders ({myOrders.length})
              </h3>
              <div className="space-y-2">
                {myOrders.length === 0 && (
                  <div className="rounded-lg bg-amber-50 px-3 py-4 text-center text-xs text-amber-500">
                    You have no open orders for {selectedName}.
                  </div>
                )}
                {myOrders.map((order, index) => (
                  <div key={`${order.id}-${index}`} className="flex items-center justify-between gap-3 rounded-lg border border-amber-200/70 bg-white/70 px-3 py-2 text-xs">
                    <span className={order.kind === 1 ? 'font-black text-green-700' : 'font-black text-red-600'}>
                      {order.kind === 1 ? 'BUY' : 'SELL'}
                    </span>
                    <span className="text-amber-900">${order.price.toFixed(2)} x {order.remaining}</span>
                    <button
                      onClick={() => cancelOrder.mutate(order.id)}
                      disabled={cancelOrder.isPending}
                      className="rounded bg-red-50 px-2 py-1 font-bold text-red-600 hover:bg-red-100 disabled:text-red-300"
                    >
                      Cancel
                    </button>
                  </div>
                ))}
              </div>
            </section>
          </main>
        </div>
      </div>
    </div>
  )
}

function PriceCurve({ series }: { series: Array<{ price: number; time: string }> }) {
  const width = 640
  const height = 160
  const padding = 14
  const prices = series.map((point) => point.price)
  const min = prices.length ? Math.min(...prices) : 0
  const max = prices.length ? Math.max(...prices) : 0
  const spread = Math.max(1, max - min)
  const points = series.map((point, index) => {
    const x = padding + (index / Math.max(1, series.length - 1)) * (width - padding * 2)
    const y = height - padding - ((point.price - min) / spread) * (height - padding * 2)
    return `${x.toFixed(2)},${y.toFixed(2)}`
  }).join(' ')
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
        <line x1={padding} y1={height - padding} x2={width - padding} y2={height - padding} stroke="#d6b27b" strokeDasharray="4 4" />
        <line x1={padding} y1={padding} x2={padding} y2={height - padding} stroke="#d6b27b" strokeDasharray="4 4" />
        {points ? (
          <>
            <polyline
              points={`${padding},${height - padding} ${points} ${width - padding},${height - padding}`}
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

function ParticipantList({
  title,
  emptyText,
  orders,
  currentCompanyId,
}: {
  title: string
  emptyText: string
  orders: Array<{ id: string; companyId: number; price: number; remaining: number; quantity: number }>
  currentCompanyId: number
}) {
  return (
    <div className="rounded-2xl border border-amber-300/60 bg-white/60 p-4 shadow-sm">
      <h3 className="mb-2 text-xs font-black uppercase tracking-wider text-amber-800">
        {title} ({orders.length})
      </h3>
      <div className="space-y-2">
        {orders.length === 0 && (
          <div className="rounded-lg bg-amber-50 px-3 py-4 text-center text-xs text-amber-500">
            {emptyText}
          </div>
        )}
        {orders.slice(0, 10).map((order, index) => (
          <div key={`${order.id}-${order.companyId}-${index}`} className="grid grid-cols-[1fr_auto] gap-2 rounded-lg border border-amber-200/70 bg-white/70 px-3 py-2 text-xs">
            <div className="min-w-0">
              <div className="truncate font-black text-amber-950">
                {companyLabel(order.companyId, currentCompanyId)}
              </div>
              <div className="text-[10px] text-amber-600">
                Order {shortOrderId(order.id)}
              </div>
            </div>
            <div className="text-right">
              <div className="font-black tabular-nums text-amber-950">${order.price.toFixed(2)}</div>
              <div className="text-[10px] font-semibold text-amber-700">
                {order.remaining.toLocaleString()} / {order.quantity.toLocaleString()}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function companyLabel(companyId: number, currentCompanyId: number): string {
  if (companyId === currentCompanyId) return 'You'
  if (companyId === 900001) return 'Atlas Trading Bot'
  if (companyId === 900002) return 'Nova Market Bot'
  if (companyId === 999999) return 'National Team'
  return `Company #${companyId}`
}

function shortOrderId(id: string): string {
  if (id.length <= 18) return id
  return `${id.slice(0, 10)}...${id.slice(-6)}`
}

function Metric({
  label,
  value,
  tone,
  subtle,
}: {
  label: string
  value: string
  tone?: 'up' | 'down'
  subtle?: boolean
}) {
  const toneClass = tone === 'up' ? 'text-green-700' : tone === 'down' ? 'text-red-600' : 'text-amber-950'
  return (
    <div className={`rounded-xl border border-amber-200/60 bg-white/65 px-3 py-2 ${subtle ? 'text-xs' : ''}`}>
      <div className="text-[10px] font-bold uppercase tracking-wider text-amber-600">{label}</div>
      <div className={`font-black tabular-nums ${toneClass}`}>{value}</div>
    </div>
  )
}

function OrderBookSide({
  title,
  tone,
  rows,
}: {
  title: string
  tone: 'buy' | 'sell'
  rows: Array<{ price: number; quantity?: number; qty?: number }>
}) {
  return (
    <div className="rounded-2xl border border-amber-300/60 bg-white/60 p-4 shadow-sm">
      <h3 className={`mb-2 text-xs font-black uppercase tracking-wider ${tone === 'buy' ? 'text-green-700' : 'text-red-600'}`}>
        {title}
      </h3>
      <div className="space-y-1">
        {rows.length === 0 && (
          <div className="rounded-lg bg-amber-50 px-3 py-4 text-center text-xs text-amber-500">No orders yet</div>
        )}
        {rows.slice(0, 8).map((row, index) => (
          <div key={`${row.price}-${index}`} className="grid grid-cols-2 rounded-lg bg-white/65 px-3 py-2 text-xs">
            <span className="font-black text-amber-950">${row.price.toFixed(2)}</span>
            <span className="text-right font-semibold text-amber-700">
              {(row.quantity ?? row.qty ?? 0).toLocaleString()}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
