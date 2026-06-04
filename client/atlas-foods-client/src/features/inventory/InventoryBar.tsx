import { useTranslation } from 'react-i18next'
import { useWarehouse } from '@/api/inventory.api'
import { resourceIcon } from '@/game/icons'
import { SupplyChainPage } from '@/features/chain/SupplyChainPage'
import { useUIStore, type ActiveView } from '@/store/ui.store'
import { audio } from '@/audio/AudioManager'

export function InventoryBar() {
  const { t } = useTranslation()
  const { data } = useWarehouse()
  const activeView = useUIStore((s) => s.activeView)
  const setActiveView = useUIStore((s) => s.setActiveView)
  const inventory = data?.inventory ?? []
  const used = data?.used ?? 0
  const capacity = data?.capacity ?? 100

  const usedPct = capacity > 0 ? Math.min(100, (used / capacity) * 100) : 0
  const activeTab = activeView === 'chain' ? 'chain' : 'warehouse'

  if (activeTab === 'chain') {
    return (
      <div className="h-full overflow-y-auto p-4">
        <InventorySubnav active="chain" onChange={setActiveView} />
        <SupplyChainPage embedded />
      </div>
    )
  }

  return (
    <div className="p-4">
      <InventorySubnav active="warehouse" onChange={setActiveView} />
      {/* Capacity bar */}
      <div className="mb-3">
        <div className="flex items-center justify-between text-xs text-amber-700 mb-1">
          <span className="font-semibold uppercase tracking-wider">{t('inventory.warehouse')}</span>
          <span className="tabular-nums">{t('inventory.capacity', { used, capacity })}</span>
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
            className="flex flex-col items-center gap-1 rounded-lg bg-white/60 p-2 border border-amber-200/40 cursor-pointer"
            onClick={() => audio.playSfx('inventory_open')}
            title={`#${item.resourceId}: ${item.quantity}`}
          >
            <img src={resourceIcon(item.resourceId)} alt="" className="h-7 w-7 object-contain" />
            <span className="text-[9px] font-semibold text-amber-800 text-center leading-tight truncate w-full">
              #{item.resourceId}
            </span>
            <span className="text-[10px] font-bold text-amber-900 tabular-nums">{item.quantity}</span>
          </div>
        ))}
      </div>
      {inventory.length > 12 && (
        <div className="text-[10px] text-amber-500 text-center mt-2">
          {t('inventory.moreItems', { count: inventory.length - 12 })}
        </div>
      )}
    </div>
  )
}

function InventorySubnav({
  active,
  onChange,
}: {
  active: 'warehouse' | 'chain'
  onChange: (view: ActiveView) => void
}) {
  const { t } = useTranslation()
  return (
    <div className="mb-3 flex gap-2 rounded-lg border border-amber-300/50 bg-white/50 p-1">
      <button
        onClick={() => { onChange('warehouse'); audio.playSfx('inventory_sort') }}
        className={`flex-1 rounded-md px-3 py-1.5 text-xs font-bold transition-colors ${
          active === 'warehouse'
            ? 'bg-amber-800 text-white'
            : 'text-amber-800 hover:bg-amber-100'
        }`}
      >
        {t('inventory.warehouse')}
      </button>
      <button
        onClick={() => { onChange('chain'); audio.playSfx('inventory_sort') }}
        className={`flex-1 rounded-md px-3 py-1.5 text-xs font-bold transition-colors ${
          active === 'chain'
            ? 'bg-amber-800 text-white'
            : 'text-amber-800 hover:bg-amber-100'
        }`}
      >
        {t('inventory.supplyChain')}
      </button>
    </div>
  )
}
