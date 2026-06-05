import { useMemo, useState } from 'react'
import { getCompanyId } from '@/api/client'
import { useMarketTicker, useMarketDepth, useMarketOrders, useCancelOrder, useResources } from '@/api/hooks/market.hooks'
import { useWarehouse } from '@/api/hooks/warehouse.hooks'
import { FALLBACK_MARKET_RESOURCES, MARKET_GROUPS, formatResourceName, resourceIcon, type MarketGroupId } from '@/game/resources'
import type { MarketOrder } from '@/game/types'
import { PriceHistoryChart } from './PriceHistoryChart'
import { ParticipantList } from './ParticipantList'
import { CreateOrderForm } from './CreateOrderForm'

const companyId = getCompanyId

export function MarketPage() {
  const [group, setGroup] = useState<MarketGroupId>('all')
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [showForm, setShowForm] = useState(false)

  const { data: resourcesData } = useResources()
  const { data: ticker } = useMarketTicker(selectedId ?? 1)
  const { data: depth } = useMarketDepth(selectedId ?? 1)
  const { data: orders } = useMarketOrders(selectedId ?? 1)
  const { data: warehouse } = useWarehouse()
  const cancelOrder = useCancelOrder()

  const allResources = useMemo(
    () => resourcesData?.resources ?? FALLBACK_MARKET_RESOURCES,
    [resourcesData],
  )

  const filtered = useMemo(() => {
    if (group === 'all') return allResources
    const groupDef = MARKET_GROUPS.find((g) => g.id === group)
    return groupDef
      ? allResources.filter((r) => (groupDef.ids as readonly number[]).includes(r.resourceId))
      : allResources
  }, [allResources, group])

  const selectedResource = selectedId
    ? allResources.find((r) => r.resourceId === selectedId)
    : filtered[0]

  const buys = depth?.buys ?? []
  const sells = depth?.sells ?? []
  const currentCompanyId = companyId()
  const myOrders = useMemo(
    () => (orders ?? []).filter((o: MarketOrder) => o.companyId === Number(currentCompanyId)),
    [orders, currentCompanyId],
  )

  const inventory = warehouse?.inventory ?? []
  const myInv = selectedId
    ? inventory.find((i) => i.resourceId === selectedId)
    : null

  return (
    <div className="h-full overflow-y-auto bg-amber-50 p-4">
      {/* Group filter tabs */}
      <div className="mb-3 flex gap-1 rounded-lg border border-amber-300/50 bg-white/50 p-1">
        <button onClick={() => setGroup('all')} className={`flex-1 py-1.5 text-xs font-semibold rounded-md ${group === 'all' ? 'bg-amber-200/70 text-amber-900' : 'text-amber-600 hover:text-amber-800'}`}>
          All
        </button>
        {MARKET_GROUPS.map((g) => (
          <button key={g.id} onClick={() => setGroup(g.id)} className={`flex-1 py-1.5 text-xs font-semibold rounded-md ${group === g.id ? 'bg-amber-200/70 text-amber-900' : 'text-amber-600 hover:text-amber-800'}`}>
            {g.id}
          </button>
        ))}
      </div>

      {/* Resource grid */}
      <div className="grid grid-cols-6 gap-1.5 mb-3">
        {filtered.map((r) => (
          <button
            key={r.resourceId}
            onClick={() => { setSelectedId(r.resourceId); setShowForm(false) }}
            className={`flex flex-col items-center p-1.5 rounded-lg border transition-colors ${selectedId === r.resourceId ? 'bg-amber-200/70 border-amber-400' : 'bg-white/60 border-amber-200/40 hover:bg-amber-100/60'}`}
          >
            <img src={resourceIcon(r.resourceId)} alt="" className="w-5 h-5 object-contain" />
            <span className="text-[9px] text-amber-900 mt-0.5">{formatResourceName(r.resourceId)}</span>
          </button>
        ))}
      </div>

      {selectedResource && (
        <>
          {/* Price chart */}
          <PriceHistoryChart series={ticker?.series ?? []} />

          {/* Quick stats */}
          <div className="mt-3 flex gap-3">
            <div className="flex-1 bg-white/60 rounded-lg p-3 border border-amber-200/40 text-center">
              <div className="text-[9px] text-amber-600 uppercase">Your Stock</div>
              <div className="text-lg font-bold text-amber-900">{myInv?.quantity ?? 0}</div>
            </div>
            <button onClick={() => setShowForm(!showForm)} className="flex-1 bg-amber-600 hover:bg-amber-700 text-white rounded-lg p-3 text-xs font-semibold transition-colors">
              {showForm ? 'Cancel' : 'New Order'}
            </button>
          </div>

          {showForm && (
            <CreateOrderForm resourceId={selectedResource.resourceId} onSuccess={() => setShowForm(false)} />
          )}

          {/* Order book */}
          <div className="mt-3 grid grid-cols-2 gap-3">
            <ParticipantList title="Buy Orders" emptyText="No buy orders" orders={buys.map((b, i) => ({ id: `depth-buy-${i}`, companyId: 0, price: b.price, remaining: b.quantity, quantity: b.quantity }))} currentCompanyId={Number(currentCompanyId)} />
            <ParticipantList title="Sell Orders" emptyText="No sell orders" orders={sells.map((s, i) => ({ id: `depth-sell-${i}`, companyId: 0, price: s.price, remaining: s.quantity, quantity: s.quantity }))} currentCompanyId={Number(currentCompanyId)} />
          </div>

          {/* My orders */}
          {myOrders.length > 0 && (
            <div className="mt-3">
              <h3 className="text-xs font-black uppercase tracking-wider text-amber-800 mb-2">My Orders ({myOrders.length})</h3>
              <div className="space-y-1">
                {myOrders.map((o: MarketOrder) => (
                  <div key={o.id} className="flex items-center gap-2 p-2 bg-white/70 rounded-lg border border-amber-200/70 text-xs">
                    <span className="font-semibold text-amber-900">${o.price}</span>
                    <span className="text-amber-600">x{o.quantity}</span>
                    <button onClick={() => cancelOrder.mutate(o.id)} className="ml-auto text-red-400 hover:text-red-600 text-[10px]">Cancel</button>
                  </div>
                ))}
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
