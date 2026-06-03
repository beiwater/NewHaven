import { useDailyOrders, useCompleteDailyOrder, useClaimDailyOrder } from '@/api/contracts.api'
import { resourceIcon } from '@/game/resources'
import { useUIStore } from '@/store/ui.store'

export function MobileOrderBoard() {
  const { data: daily } = useDailyOrders()
  const completeDaily = useCompleteDailyOrder()
  const claimDaily = useClaimDailyOrder()
  const setActiveView = useUIStore((s) => s.setActiveView)

  const orders = daily?.orders ?? []

  return (
    <div className="bg-white/60 rounded-xl border border-amber-300/50 p-3 min-w-[260px] shrink-0">
      <div className="flex items-center gap-2 mb-2">
        <svg className="w-4 h-4 text-amber-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
        <h3 className="text-xs font-bold text-amber-800 uppercase tracking-wider">Order Board</h3>
      </div>

      {orders.length === 0 && (
        <div className="text-[10px] text-amber-400 italic text-center py-2">No orders available</div>
      )}

      <div className="space-y-1.5">
        {orders.slice(0, 2).map((order) => {
          const isCompleted = order.status === 'completed'
          const rewardCash = order.rewardCash ?? 0
          const rewardXP = order.rewardXP ?? 0

          return (
            <div key={order.id} className="flex items-center gap-2 p-2 bg-amber-50/70 rounded-lg border border-amber-200/30">
              <img
                src={resourceIcon(order.resourceId)}
                alt=""
                className="w-7 h-7 object-contain shrink-0 rounded"
                loading="lazy"
              />
              <div className="flex-1 min-w-0">
                <div className="text-[10px] font-semibold text-amber-900 truncate">
                  Deliver {order.quantity} Resource #{order.resourceId}
                </div>
                <div className="text-[9px] text-amber-600">
                  Reward ${rewardCash.toLocaleString()}
                  {rewardXP > 0 && <span className="ml-1">{rewardXP} XP</span>}
                </div>
              </div>
              {isCompleted ? (
                <button
                  onClick={() => claimDaily.mutate(order.id)}
                  disabled={claimDaily.isPending}
                  className="px-2 py-1 bg-green-500 hover:bg-green-600 disabled:bg-green-300 text-white text-[9px] font-bold rounded-md shrink-0 transition-colors"
                >
                  Claim
                </button>
              ) : (
                <button
                  onClick={() => completeDaily.mutate(order.id)}
                  disabled={completeDaily.isPending}
                  className="px-2 py-1 bg-orange-500 hover:bg-orange-600 disabled:bg-orange-300 text-white text-[9px] font-bold rounded-md shrink-0 transition-colors"
                >
                  Go
                </button>
              )}
            </div>
          )
        })}
      </div>

      <button
        onClick={() => setActiveView('contracts')}
        className="w-full mt-2 py-1.5 bg-green-600 hover:bg-green-700 text-white text-[10px] font-semibold rounded-lg transition-colors"
      >
        View All Contracts
      </button>
    </div>
  )
}
