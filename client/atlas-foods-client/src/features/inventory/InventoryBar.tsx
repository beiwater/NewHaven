import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useWarehouse } from '@/api/inventory.api'
import { useResources } from '@/api/market.api'
import { useResearch } from '@/api/research.api'
import { qualitySalesBonusPct } from '@/game/quality'
import { resourceIcon, resourceName } from '@/game/resources'
import { SupplyChainPage } from '@/features/chain/SupplyChainPage'
import { useUIStore, type ActiveView } from '@/store/ui.store'
import { audio } from '@/audio/AudioManager'

export function InventoryBar() {
  const { t } = useTranslation()
  const { data } = useWarehouse()
  const { data: resourcesData } = useResources()
  const { data: research = [] } = useResearch()
  const activeView = useUIStore((s) => s.activeView)
  const setActiveView = useUIStore((s) => s.setActiveView)
  const inventory = useMemo(() => data?.inventory ?? [], [data?.inventory])
  const used = data?.used ?? 0
  const capacity = data?.capacity ?? 100
  const [selectedStack, setSelectedStack] = useState<string | null>(null)

  const resourceReferences = useMemo(() => new Map(
    (resourcesData?.resources ?? []).map((resource) => [resource.resourceId, resource.recommendedPrice ?? 0]),
  ), [resourcesData?.resources])
  const maxQualityByResource = useMemo(() => new Map(research.map((item) => [item.resourceId, item.maxQuality])), [research])
  const groups = useMemo(() => {
    const grouped = new Map<number, typeof inventory>()
    for (const item of inventory) {
      const stacks = grouped.get(item.resourceId) ?? []
      stacks.push(item)
      grouped.set(item.resourceId, stacks)
    }
    return [...grouped.entries()]
      .map(([resourceId, stacks]) => ({
        resourceId,
        stacks: stacks.sort((left, right) => (left.quality ?? 0) - (right.quality ?? 0)),
        total: stacks.reduce((sum, stack) => sum + stack.quantity, 0),
      }))
      .sort((left, right) => left.resourceId - right.resourceId)
  }, [inventory])
  const selected = inventory.find((item) => `${item.resourceId}:${item.quality ?? 0}` === selectedStack)

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
    <div className="h-full overflow-y-auto bg-gradient-to-br from-[#f8edd7] via-[#fffaf0] to-[#f2dcb5] p-4 sm:p-6">
      <InventorySubnav active="warehouse" onChange={setActiveView} />
      <header className="mb-4 flex flex-wrap items-end justify-between gap-3">
        <div>
          <p className="text-[10px] font-black uppercase tracking-[0.24em] text-amber-600">{t('inventory.stockLedger')}</p>
          <h2 className="text-2xl font-black text-amber-950">{t('inventory.warehouse')}</h2>
        </div>
        <div className="min-w-56 rounded-2xl border border-amber-200 bg-white/70 px-4 py-3 shadow-sm">
          <div className="mb-1 flex items-center justify-between text-xs text-amber-700">
            <span className="font-semibold uppercase tracking-wider">{t('inventory.capacityLabel')}</span>
            <span className="tabular-nums">{t('inventory.capacity', { used, capacity })}</span>
          </div>
          <div className="h-2.5 overflow-hidden rounded-full bg-amber-200/60">
            <div className={`h-full rounded-full transition-all duration-500 ${usedPct > 90 ? 'bg-red-500' : usedPct > 70 ? 'bg-yellow-500' : 'bg-green-500'}`} style={{ width: `${usedPct}%` }} />
          </div>
        </div>
      </header>

      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        {groups.map((group) => (
          <article key={group.resourceId} className="rounded-2xl border border-amber-200 bg-white/75 p-4 shadow-sm">
            <div className="flex items-center gap-3">
              <div className="grid h-16 w-16 shrink-0 place-items-center rounded-2xl bg-gradient-to-br from-amber-50 to-amber-200/70">
                <img src={resourceIcon(group.resourceId)} alt="" className="h-12 w-12 object-contain" />
              </div>
              <div className="min-w-0 flex-1">
                <h3 className="truncate text-base font-black text-amber-950">{resourceName(group.resourceId)}</h3>
                <div className="mt-1 text-xs font-bold text-amber-700">{group.total.toLocaleString()} {t('inventory.units')}</div>
                {(resourceReferences.get(group.resourceId) ?? 0) > 0 && <div className="mt-0.5 text-[10px] text-cyan-700">{t('inventory.marketReference')}: ${(resourceReferences.get(group.resourceId) ?? 0).toFixed(2)}</div>}
                <div className="mt-0.5 text-[9px] font-black text-violet-700">{t('research.unlockedThrough', { quality: maxQualityByResource.get(group.resourceId) ?? 0 })}</div>
              </div>
            </div>
            <div className="mt-3 border-t border-amber-100 pt-3">
              <div className="mb-2 text-[9px] font-black uppercase tracking-[0.18em] text-amber-600">{t('inventory.qualityStacks')}</div>
              <div className="flex flex-wrap gap-2">
                {group.stacks.map((stack) => {
                  const quality = stack.quality ?? 0
                  const key = `${stack.resourceId}:${quality}`
                  return (
                    <button key={key} type="button" onClick={() => { setSelectedStack(key); audio.playSfx('inventory_open') }} className={`rounded-xl border px-2.5 py-2 text-left transition ${selectedStack === key ? 'border-violet-400 bg-violet-100 ring-2 ring-violet-200' : 'border-amber-200 bg-amber-50 hover:border-violet-300'}`}>
                      <div className="text-[10px] font-black text-violet-700">Q{quality}</div>
                      <div className="text-xs font-black tabular-nums text-amber-950">{stack.quantity.toLocaleString()}</div>
                      <div className="text-[8px] font-bold text-green-700">+{qualitySalesBonusPct(quality)}% {t('inventory.saleSpeed')}</div>
                    </button>
                  )
                })}
              </div>
            </div>
          </article>
        ))}
      </div>
      {groups.length === 0 && <div className="rounded-2xl border border-dashed border-amber-300 bg-white/50 p-10 text-center text-sm font-semibold text-amber-600">{t('inventory.emptyWarehouse')}</div>}
      {selected && (
        <div className="sticky bottom-3 mt-4 flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-violet-300 bg-[#251f35]/95 px-4 py-3 text-violet-50 shadow-xl backdrop-blur">
          <div><span className="font-black">{resourceName(selected.resourceId)} · Q{selected.quality ?? 0}</span><span className="ml-2 text-xs text-violet-200">{selected.quantity.toLocaleString()} {t('inventory.units')}</span></div>
          <div className="text-xs font-bold text-green-300">+{qualitySalesBonusPct(selected.quality ?? 0)}% {t('inventory.retailDemandSpeed')}</div>
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
