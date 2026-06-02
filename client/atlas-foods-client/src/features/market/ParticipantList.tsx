import { BOT_LABELS } from './constants'

function companyLabel(companyId: number, currentCompanyId: number): string {
  if (companyId === currentCompanyId) return 'You'
  const botLabel = BOT_LABELS[companyId]
  if (botLabel) return botLabel
  return `Company #${companyId}`
}

function shortOrderId(id: string): string {
  if (id.length <= 18) return id
  return `${id.slice(0, 10)}...${id.slice(-6)}`
}

export function ParticipantList({
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
