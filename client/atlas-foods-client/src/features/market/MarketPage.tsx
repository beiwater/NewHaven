import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { audio } from '@/audio/AudioManager'
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
  useMyOrders,
} from '@/api/market.api'
import {
  ALL_RESOURCE_IDS,
  FALLBACK_MARKET_RESOURCES,
  MARKET_GROUPS,
  formatResourceName,
  resourceName,
  resourceIcon,
  type MarketGroupId,
} from '@/game/resources'
import { PriceCurve } from './PriceCurve'
import { ParticipantList } from './ParticipantList'

export function MarketPage() {
  const { t } = useTranslation()
  const [selectedResource, setSelectedResource] = useState(1)
  const [selectedGroup, setSelectedGroup] = useState<MarketGroupId>('all')
  const [orderKind, setOrderKind] = useState<'buy' | 'sell'>('buy')
  const [quantity, setQuantity] = useState('10')
  const [price, setPrice] = useState('10.00')
  const [activeTab, setActiveTab] = useState<'market' | 'myorders'>('market')

  const { data: resourcesData } = useResources()
  const resources = resourcesData?.resources?.length ? resourcesData.resources : FALLBACK_MARKET_RESOURCES
  const activeIds = selectedGroup === 'all' ? ALL_RESOURCE_IDS : (MARKET_GROUPS.find((g) => g.id === selectedGroup)?.ids ?? ALL_RESOURCE_IDS)
  const visibleResources = useMemo(
    () => activeIds.map((id) => resources.find((r) => r.resourceId === id)).filter(Boolean),
    [activeIds, resources],
  )

  const { data: ticker } = useMarketTicker(selectedResource)
  const { data: depth } = useMarketDepth(selectedResource)
  const { data: orders } = useMarketOrders(selectedResource)
  const { data: companyData } = useCompany()
  const { data: warehouse } = useWarehouse()
  const createOrder = useCreateOrder()
  const cancelOrder = useCancelOrder()
  const { data: myAllOrders } = useMyOrders()
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
    ? t('market.notEnoughItem', { name: selectedName, available: selectedInventory.toLocaleString() })
    : orderKind === 'buy' && orderValue > cash
      ? t('market.notEnoughCash', { need: orderValue.toLocaleString(), have: cash.toLocaleString() })
      : ''
  const mutationError = createOrder.error ?? takeOrder.error ?? cancelOrder.error
  const mutationMessage = mutationError instanceof Error ? mutationError.message : ''

  const handleCreateOrder = () => {
    if (localOrderError) return
    audio.playSfx(orderKind === 'buy' ? 'market_buy' : 'market_sell')
    audio.playSfx('market_order_created')
    createOrder.mutate({
      resourceId: selectedResource,
      kind: orderKind === 'buy' ? 1 : 0,
      quality: 0,
      quantity: numericQuantity,
      price: numericPrice,
    })
  }

  const handleTakeBestAsk = () => {
    audio.playSfx('market_buy')
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
            <p className="text-[10px] font-bold uppercase tracking-[0.24em] text-amber-700/70">{t('market.subtitle')}</p>
            <h2 className="text-2xl font-black text-amber-950">{t('market.title')}</h2>
          </div>
          <div className="rounded-full border border-amber-300/70 bg-white/60 px-3 py-1 text-xs font-semibold text-amber-800">
            {t('market.tradableResources', { count: resources.length })}
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          {/* All resources */}
          <button
            onClick={() => {
              setSelectedGroup('all')
              setSelectedResource(1)
            }}
            className={`rounded-full px-4 py-2 text-xs font-bold transition-colors ${
              selectedGroup === 'all'
                ? 'bg-amber-800 text-white shadow'
                : 'bg-white/60 text-amber-800 hover:bg-amber-100'
            }`}
          >
            {t('market.all')}
          </button>
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
              {t(`market.groupLabels.${group.id}`)}
            </button>
          ))}
        </div>

        <div className="grid gap-4 lg:grid-cols-[280px_1fr]">
          <aside className="rounded-2xl border border-amber-300/60 bg-white/55 p-3 shadow-sm">
            <div className="mb-2 text-[10px] font-bold uppercase tracking-wider text-amber-700">{t('market.selectProduct')}</div>
            <div className="grid gap-2">
              {visibleResources.map((resource, index) => resource && (
                <button
                  key={resource.resourceId}
                  onClick={() => setSelectedResource(resource.resourceId)}
                  className={`${index === 0 ? 'tutorial-market-product' : ''} flex items-center gap-3 rounded-xl border p-2 text-left transition-colors ${
                    selectedResource === resource.resourceId
                      ? 'border-amber-700 bg-amber-100 text-amber-950'
                      : 'border-amber-200/60 bg-white/45 text-amber-800 hover:bg-white/80'
                  }`}
                >
                  <img src={resourceIcon(resource.resourceId)} alt="" className="h-8 w-8 object-contain" />
                  <div className="min-w-0">
                    <div className="truncate text-sm font-bold">{resourceName(resource.resourceId)}</div>
                    <div className="text-[10px] text-amber-600">ID {resource.resourceId}</div>
                  </div>
                </button>
              ))}
            </div>
          </aside>

          <main className="space-y-4">
            {/* Tab buttons */}
            <div className="flex gap-2 border-b border-amber-200/60 pb-2">
              <button
                onClick={() => setActiveTab('market')}
                className={`px-4 py-2 text-xs font-bold rounded-t-lg transition-colors ${
                  activeTab === 'market'
                    ? 'bg-amber-800 text-white shadow'
                    : 'bg-amber-100/60 text-amber-700 hover:bg-amber-200/60'
                }`}
              >
                {t('market.title')}
              </button>
              <button
                onClick={() => setActiveTab('myorders')}
                className={`px-4 py-2 text-xs font-bold rounded-t-lg transition-colors ${
                  activeTab === 'myorders'
                    ? 'bg-amber-800 text-white shadow'
                    : 'bg-amber-100/60 text-amber-700 hover:bg-amber-200/60'
                }`}
              >
                {t('market.myOrders')}
              </button>
            </div>
            {activeTab === 'market' && (<>
            <section className="rounded-2xl border border-amber-300/60 bg-white/60 p-4 shadow-sm">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="flex items-center gap-3">
                  <img src={resourceIcon(selectedResource)} alt="" className="h-12 w-12 object-contain" />
                  <div>
                    <h3 className="text-xl font-black text-amber-950">{selectedName}</h3>
                    <p className="text-xs text-amber-700">{t('market.qualitySpot')}</p>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                  <Metric label={t('market.last')} value={`$${lastPrice.toFixed(2)}`} />
                  <Metric label={t('market.change')} value={`${change >= 0 ? '+' : ''}${change.toFixed(1)}%`} tone={change >= 0 ? 'up' : 'down'} />
                  <Metric label={t('market.bid')} value={bestBid ? `$${bestBid.toFixed(2)}` : '-'} />
                  <Metric label={t('market.ask')} value={bestAsk ? `$${bestAsk.toFixed(2)}` : '-'} />
                </div>
                </div>
                <div className="mt-3 grid grid-cols-2 gap-2 text-center sm:grid-cols-4">
                  <Metric label={t('market.high')} value={`$${high.toFixed(2)}`} subtle />
                  <Metric label={t('market.low')} value={`$${low.toFixed(2)}`} subtle />
                  <Metric label={t('common.inventory')} value={`${selectedInventory.toLocaleString()}`} subtle />
                  <Metric label={t('market.cash')} value={`$${Math.floor(cash).toLocaleString()}`} subtle />
                </div>
              <PriceCurve series={series} selectedResource={selectedResource} />
            </section>

            <section className="grid gap-4 xl:grid-cols-[1fr_360px]">
              <div className="grid gap-4 sm:grid-cols-2">
                <OrderBookSide title={t('market.buyOrders')} tone="buy" rows={depth?.buys ?? []} />
                <OrderBookSide title={t('market.sellOrders')} tone="sell" rows={depth?.sells ?? []} />
              </div>

              <div className="rounded-2xl border border-amber-300/60 bg-white/60 p-4 shadow-sm">
                <div className="mb-3 flex gap-2">
                  <button
                    onClick={() => setOrderKind('buy')}
                    className={`flex-1 rounded-lg px-3 py-2 text-xs font-black ${orderKind === 'buy' ? 'bg-green-700 text-white' : 'bg-amber-100 text-amber-800'}`}
                  >
                    {t('market.limitBuy')}
                  </button>
                  <button
                    onClick={() => setOrderKind('sell')}
                    className={`flex-1 rounded-lg px-3 py-2 text-xs font-black ${orderKind === 'sell' ? 'bg-red-600 text-white' : 'bg-amber-100 text-amber-800'}`}
                  >
                    {t('market.limitSell')}
                  </button>
                </div>
                <label className="mb-2 block text-[10px] font-bold uppercase tracking-wider text-amber-700">
                  {t('market.quantity')}
                </label>
                <input
                  type="number"
                  min="1"
                  value={quantity}
                  onChange={(e) => setQuantity(e.target.value)}
                  className="tutorial-market-quantity mb-3 w-full rounded-lg border border-amber-300 bg-white px-3 py-2 text-sm text-amber-950"
                />
                <label className="mb-2 block text-[10px] font-bold uppercase tracking-wider text-amber-700">
                  {t('market.limitPrice')}
                </label>
                <input
                  type="number"
                  min="0.01"
                  step="0.01"
                  value={price}
                  onChange={(e) => setPrice(e.target.value)}
                  className="tutorial-market-price mb-3 w-full rounded-lg border border-amber-300 bg-white px-3 py-2 text-sm text-amber-950"
                />
                <button
                  onClick={handleCreateOrder}
                  disabled={createOrder.isPending || !!localOrderError}
                  className="tutorial-market-order w-full rounded-lg bg-amber-800 px-4 py-2.5 text-sm font-black text-white hover:bg-amber-900 disabled:bg-amber-400"
                >
                  {createOrder.isPending ? t('market.submitting') : t('market.placeOrder', { kind: orderKind === 'buy' ? t('market.buyKind') : t('market.sellKind') })}
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
                  {takeOrder.isPending ? t('market.buying') : bestAsk ? t('market.buyBestAsk', { price: bestAsk.toFixed(2) }) : t('market.noSellOrders')}
                </button>
              </div>
            </section>

            <section className="grid gap-4 xl:grid-cols-2">
              <ParticipantList
                title={t('market.whoIsSelling')}
                emptyText={t('market.noSellers', { name: selectedName })}
                orders={sellOrders}
                currentCompanyId={currentCompanyId}
              />
              <ParticipantList
                title={t('market.whoIsBuying')}
                emptyText={t('market.noBuyers', { name: selectedName })}
                orders={buyOrders}
                currentCompanyId={currentCompanyId}
              />
            </section>
            </>)}

            {activeTab === 'myorders' && (
            <section className="rounded-2xl border border-amber-300/60 bg-white/60 p-4 shadow-sm">
              <h3 className="mb-3 text-xs font-black uppercase tracking-wider text-amber-800">
                {t('market.myOpenOrders', { count: (myAllOrders ?? []).length })}
              </h3>
              {(!myAllOrders || myAllOrders.length === 0) ? (
                <div className="rounded-lg bg-amber-50 px-3 py-4 text-center text-xs text-amber-500">
                  {t('market.noOrders')}
                </div>
              ) : (
                <div className="space-y-3">
                  {Object.entries(
                    (myAllOrders as Array<{ id: string; resourceId: number; kind: number; price: number; remaining: number }>).reduce<Record<string, { resourceId: number; kind: number; totalRemaining: number; orders: typeof myAllOrders }>>((acc, order) => {
                      const key = `${order.resourceId}-${order.kind}`
                      if (!acc[key]) {
                        acc[key] = { resourceId: order.resourceId, kind: order.kind, totalRemaining: 0, orders: [] }
                      }
                      acc[key].totalRemaining += order.remaining
                      acc[key].orders.push(order)
                      return acc
                    }, {})
                  ).map(([key, group]) => (
                    <div key={key} className="rounded-lg border border-amber-200/70 bg-white/70 p-3">
                      <div className="flex items-center justify-between gap-2">
                        <div className="flex items-center gap-2 min-w-0">
                          <img src={resourceIcon(group.resourceId)} alt="" className="w-6 h-6 object-contain shrink-0" />
                          <span className="text-sm font-bold text-amber-900 truncate">{resourceName(group.resourceId)}</span>
                          <span className={`text-[10px] font-black px-1.5 py-0.5 rounded shrink-0 ${group.kind === 1 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
                            {group.kind === 1 ? t('market.buyKind').toUpperCase() : t('market.sellKind').toUpperCase()}
                          </span>
                        </div>
                        <div className="flex items-center gap-2 shrink-0">
                          <span className="text-xs text-amber-700">{t('market.totalRemaining', { count: group.totalRemaining })}</span>
                          <span className="text-[10px] text-amber-400">({group.orders.length})</span>
                        </div>
                      </div>
                      {/* Individual orders */}
                      <div className="mt-2 space-y-1">
                        {(group.orders as Array<{ id: string; price: number; remaining: number }>).map((order) => (
                          <div key={order.id} className="flex items-center justify-between gap-2 pl-8 pr-1 py-1 rounded bg-amber-50/50 text-xs">
                            <span className="text-amber-900 font-medium">${order.price.toFixed(2)}</span>
                            <div className="flex items-center gap-2">
                              <span className="text-amber-500">x{order.remaining}</span>
                              <button
                                onClick={() => cancelOrder.mutate(order.id)}
                                disabled={cancelOrder.isPending}
                                className="text-[10px] font-bold text-red-600 hover:text-red-800 disabled:text-red-300"
                              >
                                {t('market.cancel')}
                              </button>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </section>
            )}
          </main>
        </div>
      </div>
    </div>
  )
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
  const { t } = useTranslation()
  return (
    <div className="rounded-2xl border border-amber-300/60 bg-white/60 p-4 shadow-sm">
      <h3 className={`mb-2 text-xs font-black uppercase tracking-wider ${tone === 'buy' ? 'text-green-700' : 'text-red-600'}`}>
        {title}
      </h3>
      <div className="space-y-1">
        {rows.length === 0 && (
          <div className="rounded-lg bg-amber-50 px-3 py-4 text-center text-xs text-amber-500">{t('market.noOrders')}</div>
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
