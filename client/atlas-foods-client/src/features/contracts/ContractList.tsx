import { useDailyOrders, useCompleteDailyOrder, useClaimDailyOrder, useGovContracts, useBidContract } from '@/api/contracts.api'
import { useState } from 'react'
import { audio } from '@/audio/AudioManager'

export function ContractList() {
  const { data: daily } = useDailyOrders()
  const { data: govContracts } = useGovContracts()
  const completeDaily = useCompleteDailyOrder()
  const claimDaily = useClaimDailyOrder()
  const bidContract = useBidContract()
  const [bidPrices, setBidPrices] = useState<Record<string, string>>({})

  const handleBid = (contractId: string) => {
    const price = parseFloat(bidPrices[contractId] ?? '0')
    if (price <= 0) return
    audio.playSfx('contract_signed')
    bidContract.mutate({ contractId, unitPrice: price })
  }

  return (
    <div className="h-full overflow-y-auto p-4 space-y-4">
      <h2 className="text-lg font-bold text-amber-900">Contracts</h2>

      {/* Daily Orders */}
      <div>
        <h3 className="text-xs font-bold text-amber-700 uppercase tracking-wider mb-2 flex items-center gap-2">
          <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          Daily Orders
          {daily?.date && <span className="text-[10px] text-amber-500 font-normal">({daily.date})</span>}
        </h3>
        <div className="space-y-2">
          {(daily?.orders ?? []).length === 0 && (
            <div className="text-xs text-amber-400 italic">No daily orders available</div>
          )}
          {(daily?.orders ?? []).map((order) => (
            <div
              key={order.id}
              className="flex items-center gap-3 p-2.5 bg-white/60 rounded-lg border border-amber-200/40"
            >
              <div className="w-8 h-8 rounded bg-amber-100 flex items-center justify-center text-sm">
                📦
              </div>
              <div className="flex-1">
                <div className="text-xs font-semibold text-amber-900">
                  Resource #{order.resourceId} × {order.quantity}
                </div>
                <div className="text-[10px] text-amber-600">
                  Reward: ${order.rewardCash?.toLocaleString() ?? 'N/A'}
                </div>
              </div>
              <span className={`text-[10px] px-2 py-0.5 rounded-full font-semibold ${
                order.status === 'completed' ? 'bg-green-100 text-green-700' : 'bg-amber-100 text-amber-700'
              }`}>
                {order.status}
              </span>
              {order.status === 'completed' ? (
                <button
                  onClick={() => claimDaily.mutate(order.id)}
                  disabled={claimDaily.isPending}
                  className="px-3 py-1 bg-green-600 hover:bg-green-700 disabled:bg-green-300 text-white text-xs font-semibold rounded transition-colors"
                >
                  Claim
                </button>
              ) : (
                <button
                  onClick={() => { audio.playSfx('delivery_complete'); completeDaily.mutate(order.id) }}
                  disabled={completeDaily.isPending}
                  className="px-3 py-1 bg-amber-700 hover:bg-amber-800 disabled:bg-amber-300 text-white text-xs font-semibold rounded transition-colors"
                >
                  Deliver
                </button>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Government Contracts */}
      <div>
        <h3 className="text-xs font-bold text-amber-700 uppercase tracking-wider mb-2 flex items-center gap-2">
          <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
          </svg>
          Government Contracts
        </h3>
        <div className="space-y-2">
          {(govContracts ?? []).length === 0 && (
            <div className="text-xs text-amber-400 italic">No government contracts available</div>
          )}
          {(govContracts ?? []).map((contract) => (
            <div
              key={contract.id}
              className="p-2.5 bg-white/60 rounded-lg border border-amber-200/40"
            >
              <div className="flex items-center justify-between mb-1.5">
                <div className="text-xs font-semibold text-amber-900">
                  Resource #{contract.resourceId} × {contract.quantity}
                </div>
                <span className={`text-[10px] px-2 py-0.5 rounded-full font-semibold ${
                  contract.status === 'awarded' ? 'bg-green-100 text-green-700'
                  : contract.status === 'delivered' ? 'bg-blue-100 text-blue-700'
                  : 'bg-amber-100 text-amber-700'
                }`}>
                  {contract.status}
                </span>
              </div>

              {contract.status === 'open' && (
                <div className="flex gap-2">
                  <input
                    type="number"
                    step="0.01"
                    value={bidPrices[contract.id] ?? ''}
                    onChange={(e) => setBidPrices((p) => ({ ...p, [contract.id]: e.target.value }))}
                    placeholder="Unit price"
                    className="flex-1 px-2 py-1 text-xs bg-white border border-amber-300 rounded text-amber-900 placeholder-amber-400"
                  />
                  <button
                    onClick={() => handleBid(contract.id)}
                    className="px-3 py-1 bg-amber-700 hover:bg-amber-800 text-white text-xs font-semibold rounded transition-colors"
                  >
                    Bid
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
