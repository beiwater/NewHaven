import { useWarehouse } from '@/api/hooks/warehouse.hooks'
import { resourceIcon } from '@/game/icons'
import { useUIStore } from '@/store/ui.store'

export function InventoryBar() {
  const { data } = useWarehouse()
  const setActiveView = useUIStore((s) => s.setActiveView)
  const inventory = data?.inventory ?? []
  const used = data?.used ?? 0
  const capacity = data?.capacity ?? 100

  const usedPct = capacity > 0 ? Math.min(100, (used / capacity) * 100) : 0

  return (
    <div className="p-4">
      <div className="mb-3 flex gap-2 rounded-lg border border-amber-300/50 bg-white/50 p-1">
        <button
          onClick={() => setActiveView('warehouse')}
          className="flex-1 py-1.5 text-xs font-semibold rounded-md bg-amber-200/70 text-amber-900"
        >
          Warehouse
        </button>
      </div>

      {/* Capacity bar */}
      <div className="mb-3">
        <div className="flex justify-between text-[10px] text-amber-700 mb-1">
          <span>Storage</span>
          <span>{used} / {capacity}</span>
        </div>
        <div className="w-full h-2 bg-amber-200/60 rounded-full overflow-hidden">
          <div
            className="h-full bg-amber-600 rounded-full transition-all duration-300"
            style={{ width: `${usedPct}%` }}
          />
        </div>
      </div>

      {/* Items grid */}
      <div className="grid grid-cols-4 gap-2">
        {inventory.slice(0, 12).map((item) => (
          <div
            key={item.resourceId}
            className="flex flex-col items-center gap-1 p-1.5 rounded-lg bg-white/60 border border-amber-200/40"
          >
            <img src={resourceIcon(item.resourceId)} alt="" className="w-6 h-6 object-contain" />
            <span className="text-[10px] font-medium text-amber-900">#{item.resourceId}</span>
            <span className="text-[9px] text-amber-600">x{item.quantity}</span>
          </div>
        ))}
      </div>
      {inventory.length > 12 && (
        <div className="text-[10px] text-amber-500 text-center mt-2">
          ...and {inventory.length - 12} more items
        </div>
      )}
    </div>
  )
}
