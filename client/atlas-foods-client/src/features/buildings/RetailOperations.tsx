import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useStockShelf } from '@/api/buildings.api'
import { useWarehouse } from '@/api/inventory.api'
import { useResources } from '@/api/market.api'
import { QualitySelector } from '@/features/quality/QualitySelector'
import { qualitySalesBonusPct } from '@/game/quality'
import { resourceIcon, resourceName } from '@/game/resources'
import type { Building, ResourceDefinition, ShelfItem } from '@/game/types'

function priceSpeedMultiplier(price: number, recommendedPrice: number) {
  if (!Number.isFinite(price) || !Number.isFinite(recommendedPrice) || price <= 0 || recommendedPrice <= 0) return 0
  const ratio = price / recommendedPrice
  return ratio <= 1 ? Math.min(1.25, 1 + (1 - ratio) * 0.25) : 1 / (ratio * ratio)
}

function snapToMarketTick(price: number) {
  if (!Number.isFinite(price) || price <= 0) return 0
  const step = price >= 20_000 ? 500
    : price >= 10_000 ? 100
      : price >= 5_000 ? 25
        : price >= 1_000 ? 10
          : price >= 500 ? 5
            : price >= 200 ? 2
              : price >= 100 ? 1
                : price >= 50 ? 0.5
                  : price >= 20 ? 0.25
                    : price >= 5 ? 0.1
                      : price >= 2 ? 0.05
                        : price >= 1 ? 0.01
                          : 0.001
  return Math.round(price / step) * step
}

function retailRecommendation(resource: ResourceDefinition, hourlyWage: number, level: number) {
  const sourcePrice = resource.recommendedPrice ?? 0
  const demand = (resource.retailDemandPerHour ?? 30) * (resource.demandMultiplier ?? 1) * Math.max(1, level)
  if (sourcePrice <= 0 || demand <= 0) return 0
  return snapToMarketTick(sourcePrice + (hourlyWage + 300 * Math.max(1, level)) / demand)
}

function ActiveSaleRow({ shelf, workerCount, hourlyWage }: { shelf: ShelfItem; workerCount: number; hourlyWage: number }) {
  const { t } = useTranslation()
  const soldPercent = shelf.maxQty > 0 ? Math.max(0, Math.min(100, (shelf.quantity / shelf.maxQty) * 100)) : 0

  return (
    <article className="rounded-2xl border border-green-200 bg-white/75 p-4 shadow-sm">
      <div className="grid gap-4 lg:grid-cols-[minmax(190px,1.05fr)_minmax(180px,0.95fr)_minmax(230px,0.9fr)] lg:items-center">
        <div className="flex items-center gap-4">
          <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-green-50 to-amber-100">
            <img src={resourceIcon(shelf.resourceId)} alt="" className="h-14 w-14 object-contain" />
          </div>
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="truncate text-base font-black text-amber-950">{resourceName(shelf.resourceId)}</h3>
              <span className="rounded-full bg-violet-100 px-2 py-0.5 text-[9px] font-black text-violet-700">Q{shelf.quality} · +{qualitySalesBonusPct(shelf.quality)}%</span>
              <span className="rounded-full bg-green-100 px-2 py-0.5 text-[9px] font-black uppercase tracking-wider text-green-700">{t('building.saleActive')}</span>
            </div>
            <dl className="mt-2 space-y-1 text-[11px] text-amber-700">
              <div className="flex justify-between gap-3"><dt>{t('building.price')}</dt><dd className="font-bold text-amber-950">${shelf.price.toFixed(2)}</dd></div>
              <div className="flex justify-between gap-3"><dt>{t('building.quantity')}</dt><dd className="font-bold text-amber-950">{shelf.quantity.toLocaleString()} / {shelf.maxQty.toLocaleString()}</dd></div>
              <div className="flex justify-between gap-3"><dt>{t('building.workers')}</dt><dd className="font-bold text-amber-950">{workerCount.toLocaleString()}</dd></div>
              <div className="flex justify-between gap-3"><dt>{t('building.hourlyPayroll')}</dt><dd className="font-bold text-red-700">${hourlyWage.toFixed(2)}</dd></div>
              {shelf.revenue > 0 && <div className="flex justify-between gap-3"><dt>{t('building.revenue')}</dt><dd className="font-bold text-green-700">${shelf.revenue.toFixed(2)}</dd></div>}
            </dl>
          </div>
        </div>

        <div className="border-y border-green-200/70 py-3 lg:border-x lg:border-y-0 lg:px-5 lg:py-2">
          <div className="mb-2 text-[10px] font-black uppercase tracking-[0.16em] text-green-700">{t('building.saleProgress')}</div>
          <div className="h-2.5 overflow-hidden rounded-full bg-green-100">
            <div className="h-full rounded-full bg-gradient-to-r from-green-600 to-lime-500" style={{ width: `${soldPercent}%` }} />
          </div>
          <div className="mt-2 text-[10px] font-semibold text-green-700">{t('building.remainingToSell', { count: shelf.quantity.toLocaleString() })}</div>
        </div>

        <div className="rounded-xl border border-amber-200 bg-amber-50 px-3 py-3 text-xs font-semibold leading-5 text-amber-800">
          <div className="mb-1 font-black">{t('building.saleBatchLocked')}</div>
          {t('building.saleLockedUntilComplete')}
        </div>
      </div>
    </article>
  )
}

function NewSaleRow({
  resource,
  warehouseStock,
  quality,
  setQuality,
  quantity,
  price,
  setQuantity,
  setPrice,
  pending,
  workerCount,
  hourlyWage,
  level,
  onStart,
}: {
  resource: ResourceDefinition
  warehouseStock: number
  quality: number
  setQuality: (quality: number) => void
  quantity: string
  price: string
  setQuantity: (value: string) => void
  setPrice: (value: string) => void
  pending: boolean
  workerCount: number
  hourlyWage: number
  level: number
  onStart: (quantity: number, price: number, quality: number) => void
}) {
  const { t } = useTranslation()
  const recommendedPrice = retailRecommendation(resource, hourlyWage, level)
  const qualityMultiplier = 1 + qualitySalesBonusPct(quality) / 100
  const liveDemand = (resource.retailDemandPerHour ?? 30) * (resource.demandMultiplier ?? 1) * Math.max(1, level) * qualityMultiplier
  const numericQuantity = parseInt(quantity, 10)
  const numericPrice = parseFloat(price)
  const validQuantity = Number.isInteger(numericQuantity) && numericQuantity > 0 && numericQuantity <= warehouseStock
  const validPrice = Number.isFinite(numericPrice) && numericPrice > 0
  const demandSpeed = priceSpeedMultiplier(numericPrice, recommendedPrice) * qualityMultiplier
  const demandSpeedLabel = demandSpeed * 100 < 0.0001
    ? '<0.0001%'
    : `${(demandSpeed * 100).toFixed(demandSpeed < 0.01 ? 4 : 1)}%`

  return (
    <article className="rounded-2xl border border-amber-200/80 bg-white/70 p-4 shadow-sm transition-colors hover:border-amber-400/80">
      <div className="grid gap-4 lg:grid-cols-[minmax(190px,1.05fr)_minmax(180px,0.95fr)_minmax(230px,0.9fr)] lg:items-center">
        <div className="flex items-center gap-4">
          <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-amber-50 to-cyan-100/80">
            <img src={resourceIcon(resource.resourceId)} alt="" className="h-14 w-14 object-contain" />
          </div>
          <div className="min-w-0">
            <h3 className="truncate text-base font-black text-amber-950">{resourceName(resource.resourceId)}</h3>
            <dl className="mt-2 space-y-1 text-[11px] text-amber-700">
              <div className="flex justify-between gap-3"><dt>{t('building.currentStock')}</dt><dd className="font-bold text-amber-950">{warehouseStock.toLocaleString()}</dd></div>
              <div className="flex justify-between gap-3"><dt>{t('building.recommendedSalePrice')}</dt><dd className="font-bold text-cyan-800">${recommendedPrice.toFixed(2)}</dd></div>
              <div className="flex justify-between gap-3"><dt>{t('building.liveDemand')}</dt><dd className="font-bold text-cyan-800">{liveDemand.toFixed(1)} /h</dd></div>
              <div className="flex justify-between gap-3"><dt>{t('building.workers')}</dt><dd className="font-bold text-amber-950">{workerCount.toLocaleString()}</dd></div>
              <div className="flex justify-between gap-3"><dt>{t('building.hourlyPayroll')}</dt><dd className="font-bold text-red-700">${hourlyWage.toFixed(2)}</dd></div>
            </dl>
          </div>
        </div>

        <div className="border-y border-amber-200/70 py-3 lg:border-x lg:border-y-0 lg:px-5 lg:py-2">
          <div className="mb-2 text-[10px] font-black uppercase tracking-[0.16em] text-amber-600">{t('building.saleTerms')}</div>
          <div className="rounded-xl bg-amber-50 px-3 py-2 text-[11px] font-semibold leading-5 text-amber-800">
            {t('building.saleTermsHelp')}
          </div>
        </div>

        <div>
          <QualitySelector value={quality} onChange={setQuality} disabled={pending} />
          <div className="mt-3 grid grid-cols-2 gap-2">
            <label className="text-[10px] font-black uppercase tracking-wider text-amber-600">
              {t('building.quantity')}
              <input
                type="number"
                min="1"
                max={warehouseStock}
                value={quantity}
                onChange={(event) => setQuantity(event.target.value)}
                className="mt-1 w-full rounded-lg border border-amber-300 bg-white px-3 py-2 text-sm font-bold normal-case text-amber-950"
              />
            </label>
            <label className="text-[10px] font-black uppercase tracking-wider text-amber-600">
              {t('building.price')}
              <input
                type="number"
                min="0.01"
                step="0.01"
                value={price}
                onChange={(event) => setPrice(event.target.value)}
                className="mt-1 w-full rounded-lg border border-amber-300 bg-white px-3 py-2 text-sm font-bold normal-case text-amber-950"
              />
            </label>
          </div>
          {recommendedPrice > 0 && (
            <button type="button" onClick={() => setPrice(recommendedPrice.toFixed(2))} className="mt-2 w-full rounded-lg border border-cyan-200 bg-cyan-50 px-3 py-1.5 text-[10px] font-black text-cyan-800 hover:bg-cyan-100">
              {t('market.useRecommended')} · ${recommendedPrice.toFixed(2)}
            </button>
          )}
          {validPrice && recommendedPrice > 0 && (
            <div className="mt-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-[10px] font-semibold leading-4 text-amber-800">
              <div className="flex items-center justify-between gap-3"><span>{t('building.demandSpeed')}</span><strong>{demandSpeedLabel}</strong></div>
              <p className="mt-1">{demandSpeed < 1 ? t('building.pricePayrollWarning') : t('building.pricePayrollStable')}</p>
            </div>
          )}
          <button
            type="button"
            onClick={() => onStart(numericQuantity, numericPrice, quality)}
            disabled={pending || !validQuantity || !validPrice}
            className="mt-2 w-full rounded-lg bg-cyan-700 px-3 py-2.5 text-xs font-black text-white transition-colors hover:bg-cyan-800 disabled:cursor-not-allowed disabled:bg-slate-300"
          >
            {pending ? t('building.stocking') : t('building.stock')}
          </button>
          {warehouseStock === 0 && <div className="mt-2 text-center text-[10px] font-semibold text-red-600">{t('building.noStockToSell')}</div>}
        </div>
      </div>
    </article>
  )
}

export function RetailOperations({ building }: { building: Building }) {
  const { t } = useTranslation()
  const { data: resourcesData } = useResources()
  const { data: warehouse } = useWarehouse()
  const stockShelf = useStockShelf()
  const [quantities, setQuantities] = useState<Record<number, string>>({})
  const [prices, setPrices] = useState<Record<number, string>>({})
  const [qualities, setQualities] = useState<Record<number, number>>({})
  const resources = resourcesData?.resources ?? []
  const sellableResources = (building.produces ?? [])
    .map((resourceId) => resources.find((resource) => resource.resourceId === resourceId) ?? { resourceId, name: resourceName(resourceId) })
  const shelves = building.shelves ?? []
  const totalRevenue = shelves.reduce((sum, shelf) => sum + (shelf.revenue ?? 0), 0)
  const workerCount = building.workerCount ?? 0
  const hourlyWage = building.hourlyWage ?? 0
  const warehouseStock = (resourceId: number, quality: number) => warehouse?.inventory
    .filter((item) => item.resourceId === resourceId && (item.quality ?? 0) === quality)
    .reduce((sum, item) => sum + item.quantity, 0) ?? 0

  return (
    <div className="min-h-0 flex-1 overflow-y-auto bg-gradient-to-br from-[#f8edd7] via-[#fffaf0] to-[#f2dcb5] p-4 sm:p-6">
      <div className="mx-auto max-w-4xl">
        <div className="mb-4 flex flex-wrap items-end justify-between gap-3">
          <div>
            <p className="text-[10px] font-black uppercase tracking-[0.22em] text-cyan-700">{t('building.operationCenter')}</p>
            <h2 className="text-xl font-black text-amber-950 sm:text-2xl">{t('building.salesCatalog')}</h2>
          </div>
          <div className="rounded-full border border-green-200 bg-green-50 px-3 py-1.5 text-[11px] font-bold text-green-700">
            {t('building.revenue')}: ${totalRevenue.toFixed(2)}
          </div>
        </div>

        <div className="space-y-3">
          {sellableResources.map((resource) => {
            const activeShelf = shelves.find((shelf) => shelf.resourceId === resource.resourceId)
            if (activeShelf) return <ActiveSaleRow key={resource.resourceId} shelf={activeShelf} workerCount={workerCount} hourlyWage={hourlyWage} />

            const availableQualities = (warehouse?.inventory ?? [])
              .filter((item) => item.resourceId === resource.resourceId && item.quantity > 0)
              .map((item) => item.quality ?? 0)
              .sort((left, right) => left - right)
            const quality = qualities[resource.resourceId] ?? availableQualities[0] ?? 0
            const stock = warehouseStock(resource.resourceId, quality)
            const recommendation = retailRecommendation(resource, hourlyWage, building.level)
            const defaultQuantity = String(Math.min(10, Math.max(1, stock)))
            const price = prices[resource.resourceId] ?? (recommendation > 0 ? recommendation.toFixed(2) : '')
            return (
              <NewSaleRow
                key={resource.resourceId}
                resource={resource}
                warehouseStock={stock}
                quality={quality}
                setQuality={(nextQuality) => setQualities((current) => ({ ...current, [resource.resourceId]: nextQuality }))}
                quantity={quantities[resource.resourceId] ?? defaultQuantity}
                price={price}
                setQuantity={(value) => setQuantities((current) => ({ ...current, [resource.resourceId]: value }))}
                setPrice={(value) => setPrices((current) => ({ ...current, [resource.resourceId]: value }))}
                pending={stockShelf.isPending}
                workerCount={workerCount}
                hourlyWage={hourlyWage}
                level={building.level}
                onStart={(quantity, salePrice, saleQuality) => stockShelf.mutate({
                  buildingId: building.id,
                  resourceId: resource.resourceId,
                  quality: saleQuality,
                  quantity,
                  price: salePrice,
                })}
              />
            )
          })}
          {sellableResources.length === 0 && (
            <div className="rounded-2xl border border-dashed border-amber-300 bg-white/50 p-8 text-center text-sm font-semibold text-amber-500">{t('building.noShelves')}</div>
          )}
        </div>

        {stockShelf.error instanceof Error && (
          <div className="mt-4 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-xs font-semibold text-red-700">{stockShelf.error.message}</div>
        )}
      </div>
    </div>
  )
}
