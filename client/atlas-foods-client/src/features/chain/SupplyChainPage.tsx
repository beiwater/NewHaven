import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { useResources } from '@/api/market.api'
import { useWarehouse } from '@/api/inventory.api'
import { useProductionJobs } from '@/api/production.api'
import { useBuildings } from '@/api/buildings.api'
import { FALLBACK_MARKET_RESOURCES, MARKET_GROUPS, formatResourceName, resourceIcon, resourceName } from '@/game/resources'
import { buildingIcon } from '@/game/icons'

export function SupplyChainPage({ embedded = false }: { embedded?: boolean }) {
  const { t } = useTranslation()
  const { data: resourcesData } = useResources()
  const { data: warehouse } = useWarehouse()
  const { data: jobsData } = useProductionJobs()
  const { data: buildingsData } = useBuildings()

  const resources = resourcesData?.resources?.length ? resourcesData.resources : FALLBACK_MARKET_RESOURCES
  const inventoryByResource = useMemo(() => {
    const map = new Map<number, number>()
    for (const item of warehouse?.inventory ?? []) map.set(item.resourceId, (map.get(item.resourceId) ?? 0) + item.quantity)
    return map
  }, [warehouse?.inventory])
  const activeByResource = useMemo(() => {
    const map = new Map<number, number>()
    for (const job of jobsData ?? []) {
      if (job.status !== 'claimed') map.set(job.resourceId, (map.get(job.resourceId) ?? 0) + 1)
    }
    return map
  }, [jobsData])
  const ownedBuildings = Array.isArray(buildingsData) ? buildingsData.filter((building) => building.placed !== false) : []

  return (
    <div className={embedded ? '' : 'h-full overflow-y-auto bg-gradient-to-br from-[#fff7e7] via-[#f8e4bd] to-[#f2d4a2] p-4 md:p-7'}>
      <div className="mx-auto max-w-7xl">
        <header className="mb-5 flex flex-wrap items-end justify-between gap-3">
          <div>
            <p className="text-[10px] font-black uppercase tracking-[0.28em] text-orange-600">{t('chain.handbook')}</p>
            <h2 className="text-3xl font-black tracking-tight text-amber-950">{t('inventory.supplyChain')}</h2>
          </div>
          <div className="flex gap-2 rounded-xl border border-amber-300/70 bg-white/50 p-1.5">
            {ownedBuildings.slice(0, 5).map((building) => <img key={building.id} src={buildingIcon(building.kind)} alt="" className="h-9 w-9 rounded-md bg-amber-50 object-contain p-1" />)}
          </div>
        </header>

        <div className="grid gap-4 xl:grid-cols-3">
          {MARKET_GROUPS.map((group) => (
            <section key={group.id} className="min-h-[430px] rounded-xl border border-amber-300 bg-[#fffaf1]/90 p-3 shadow-[0_2px_0_rgba(180,112,16,0.16)]">
              <div className="mb-3 px-1 pt-0.5">
                <h3 className="text-sm font-black text-amber-950">{t(group.labelKey)}</h3>
                <p className="mt-0.5 text-[11px] font-medium text-orange-600">{t(`chain.${group.id}`)}</p>
              </div>
              <div className="space-y-2.5">
                {group.ids.map((id) => {
                  const resource = resources.find((item) => item.resourceId === id) ?? FALLBACK_MARKET_RESOURCES.find((item) => item.resourceId === id)
                  if (!resource) return null
                  const inputs = Object.entries(resource.producedFrom ?? {})
                  const inventory = inventoryByResource.get(id) ?? 0
                  const active = activeByResource.get(id) ?? 0
                  const producer = ownedBuildings.find((building) => building.produces?.includes(id))
                  return (
                    <article key={id} className="rounded-lg border border-amber-200 bg-[#fffdf8] px-3 py-2.5 transition-transform hover:-translate-y-0.5 hover:border-amber-400">
                      <div className="flex items-center gap-3">
                        <img src={resourceIcon(id)} alt="" className="h-11 w-11 shrink-0 object-contain" />
                        <div className="min-w-0 flex-1">
                          <h4 className="truncate text-sm font-black text-amber-950">{resourceName(id)}</h4>
                          <div className="mt-1 flex flex-wrap gap-1">
                            <Pill tone="stock" label={`${t('chain.stock')} ${inventory.toLocaleString()}`} />
                            {producer && <Pill tone="building" label={producer.name ?? t(`building.name_${producer.kind}`)} />}
                            {active > 0 && <Pill tone="active" label={t('chain.running')} />}
                          </div>
                        </div>
                      </div>
                      {inputs.length > 0 ? (
                        <div className="mt-2.5 flex flex-wrap items-center gap-1.5 border-t border-amber-100 pt-2">
                          {inputs.map(([inputID, quantity]) => <RecipeChip key={inputID} resourceID={Number(inputID)} quantity={quantity} resources={resources} />)}
                          <span className="px-0.5 text-sm font-black text-orange-500">→</span>
                          <img src={resourceIcon(id)} alt="" className="h-5 w-5 object-contain" />
                        </div>
                      ) : <div className="mt-2 text-[10px] font-bold text-green-700">{t('chain.rawMaterial')}</div>}
                    </article>
                  )
                })}
              </div>
            </section>
          ))}
        </div>
      </div>
    </div>
  )
}

function RecipeChip({ resourceID, quantity, resources }: { resourceID: number; quantity: unknown; resources: typeof FALLBACK_MARKET_RESOURCES }) {
  return <span className="flex items-center gap-1 rounded-md bg-amber-50 px-1.5 py-1 text-[10px] font-bold text-amber-900"><img src={resourceIcon(resourceID)} alt="" className="h-4 w-4 object-contain" />{String(quantity)} {formatResourceName(resourceID, resources)}</span>
}

function Pill({ tone, label }: { tone: 'stock' | 'building' | 'active'; label: string }) {
  const classes = tone === 'stock' ? 'bg-amber-100 text-amber-800' : tone === 'building' ? 'bg-blue-100 text-blue-700' : 'bg-green-100 text-green-700'
  return <span className={`rounded px-1.5 py-0.5 text-[9px] font-bold ${classes}`}>{label}</span>
}
