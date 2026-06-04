import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { useResources } from '@/api/market.api'
import { useWarehouse } from '@/api/inventory.api'
import { useProductionJobs } from '@/api/production.api'
import { useBuildings } from '@/api/buildings.api'
import { FALLBACK_MARKET_RESOURCES, MARKET_GROUPS, formatResourceName, resourceIcon, resourceName } from '@/game/resources'
import { buildingIcon } from '@/game/icons'

const GROUP_COPY: Record<string, string> = {
  raw: 'Soil, barns, and simple harvests.',
  processed: 'Work added by mills and kitchens.',
  finished: 'Food with a longer story inside.',
}

export function SupplyChainPage({ embedded = false }: { embedded?: boolean }) {
  const { t } = useTranslation()
  const { data: resourcesData } = useResources()
  const { data: warehouse } = useWarehouse()
  const { data: jobsData } = useProductionJobs()
  const { data: buildingsData } = useBuildings()

  const resources = resourcesData?.resources?.length ? resourcesData.resources : FALLBACK_MARKET_RESOURCES
  const inventoryByResource = useMemo(() => {
    const map = new Map<number, number>()
    for (const item of warehouse?.inventory ?? []) {
      map.set(item.resourceId, (map.get(item.resourceId) ?? 0) + item.quantity)
    }
    return map
  }, [warehouse?.inventory])
  const activeByResource = useMemo(() => {
    const map = new Map<number, number>()
    for (const job of jobsData ?? []) {
      if (job.status === 'claimed') continue
      map.set(job.resourceId, (map.get(job.resourceId) ?? 0) + 1)
    }
    return map
  }, [jobsData])
  const ownedBuildings = Array.isArray(buildingsData) ? buildingsData.filter((b) => b.placed !== false) : []

  return (
    <div className={embedded ? '' : 'h-full overflow-y-auto bg-amber-50/60 p-4 md:p-6'}>
      <div className="mx-auto max-w-6xl space-y-4">
        <header className="flex flex-wrap items-end justify-between gap-3">
          <div>
            <p className="text-[10px] font-black uppercase tracking-[0.24em] text-amber-700/70">Farm Handbook</p>
            <h2 className="text-2xl font-black text-amber-950">Supply Chain</h2>
          </div>
          <div className="flex gap-2">
            {ownedBuildings.slice(0, 5).map((building) => (
              <img
                key={building.id}
                src={buildingIcon(building.kind)}
                alt=""
                className="h-9 w-9 rounded-md border border-amber-300/60 bg-white/70 object-contain p-1"
              />
            ))}
          </div>
        </header>

        <div className="grid gap-4 xl:grid-cols-3">
          {MARKET_GROUPS.map((group) => (
            <section key={group.id} className="rounded-lg border border-amber-300/60 bg-white/65 p-3 shadow-sm">
              <div className="mb-3">
                <h3 className="text-xs font-black uppercase tracking-wider text-amber-900">{t(group.labelKey)}</h3>
                <p className="mt-0.5 text-[10px] text-amber-600">{GROUP_COPY[group.id]}</p>
              </div>
              <div className="space-y-2">
                {group.ids.map((id) => {
                  const resource = resources.find((r) => r.resourceId === id) ?? FALLBACK_MARKET_RESOURCES.find((r) => r.resourceId === id)
                  if (!resource) return null
                  const inputs = Object.entries(resource.producedFrom ?? {})
                  const inventory = inventoryByResource.get(id) ?? 0
                  const active = activeByResource.get(id) ?? 0
                  const producer = ownedBuildings.find((b) => b.produces?.includes(id))
                  return (
                    <article key={id} className="rounded-md border border-amber-200/70 bg-amber-50/55 p-2">
                      <div className="flex items-center gap-2">
                        <img src={resourceIcon(id)} alt="" className="h-9 w-9 object-contain" />
                        <div className="min-w-0 flex-1">
                          <div className="truncate text-sm font-black text-amber-950">{resourceName(resource.resourceId)}</div>
                          <div className="flex flex-wrap gap-1 pt-1">
                            {inventory > 0 && <StatusPill tone="stock" label={inventory.toLocaleString()} />}
                            {active > 0 && <StatusPill tone="active" label="growing" />}
                            {producer && <StatusPill tone="owned" label={producer.name ?? `B${producer.kind}`} />}
                          </div>
                        </div>
                      </div>
                      {inputs.length > 0 && (
                        <div className="mt-2 flex flex-wrap items-center gap-1.5 pl-1">
                          {inputs.map(([inputId, qty]) => (
                            <div key={inputId} className="flex items-center gap-1 rounded bg-white/75 px-1.5 py-1 text-[10px] font-semibold text-amber-800">
                              <img src={resourceIcon(Number(inputId))} alt="" className="h-4 w-4 object-contain" />
                              <span>{qty} {formatResourceName(Number(inputId), resources)}</span>
                            </div>
                          ))}
                          <span className="text-[10px] font-black text-amber-500">to</span>
                          <img src={resourceIcon(id)} alt="" className="h-5 w-5 object-contain" />
                        </div>
                      )}
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

function StatusPill({ tone, label }: { tone: 'stock' | 'active' | 'owned'; label: string }) {
  const className = tone === 'stock'
    ? 'bg-amber-100 text-amber-800'
    : tone === 'active'
      ? 'bg-green-100 text-green-700'
      : 'bg-blue-100 text-blue-700'
  return <span className={`rounded px-1.5 py-0.5 text-[9px] font-bold ${className}`}>{label}</span>
}
