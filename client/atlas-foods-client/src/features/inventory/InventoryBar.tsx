import { useWarehouse } from '@/api/inventory.api'

const RESOURCE_ICONS: Record<number, string> = {
  1: '⚡', 2: '💧', 3: '🍎', 4: '🌾', 5: '🌾', 6: '🧂',
  8: '🥩', 9: '🥩', 10: '🍞', 11: '🎂', 12: '🍕',
}

export function InventoryBar() {
  const { data } = useWarehouse()
  const inventory = data?.inventory ?? []
  const used = data?.used ?? 0
  const capacity = data?.capacity ?? 100

  const usedPct = capacity > 0 ? Math.min(100, (used / capacity) * 100) : 0

  return (
    <div className="p-4">
      {/* Capacity bar */}
      <div className="mb-3">
        <div className="flex items-center justify-between text-xs text-amber-700 mb-1">
          <span className="font-semibold uppercase tracking-wider">Warehouse</span>
          <span className="tabular-nums">{used} / {capacity}</span>
        </div>
        <div className="h-2.5 bg-amber-200/60 rounded-full overflow-hidden">
          <div
            className={`h-full rounded-full transition-all duration-500 ${
              usedPct > 90 ? 'bg-red-500' : usedPct > 70 ? 'bg-yellow-500' : 'bg-green-500'
            }`}
            style={{ width: `${usedPct}%` }}
          />
        </div>
      </div>

      {/* Items grid */}
      <div className="grid grid-cols-4 gap-2">
        {inventory.slice(0, 12).map((item) => (
          <div
            key={item.resourceId}
            className="flex flex-col items-center p-1.5 bg-white/50 rounded-lg border border-amber-200/30"
          >
            <span className="text-lg">{RESOURCE_ICONS[item.resourceId] ?? '📦'}</span>
            <span className="text-[10px] font-semibold text-amber-800 tabular-nums mt-0.5">
              {item.quantity}
            </span>
          </div>
        ))}
      </div>

      {inventory.length > 12 && (
        <div className="text-[10px] text-amber-500 text-center mt-2">
          +{inventory.length - 12} more items
        </div>
      )}
    </div>
  )
}
