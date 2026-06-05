import type { ResourceDefinition } from './types'
import { resourceIcon as iconLookup } from './icons'

export const MARKET_GROUPS = [
  { id: 'raw', labelKey: 'marketGroup.raw', ids: [1, 2, 6, 12] },
  { id: 'processed', labelKey: 'marketGroup.processed', ids: [3, 4, 5, 7] },
  { id: 'finished', labelKey: 'marketGroup.finished', ids: [8, 9, 10, 11] },
] as const

export const ALL_RESOURCE_IDS = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12] as const
export type MarketGroupId = (typeof MARKET_GROUPS)[number]['id'] | 'all'

export const FALLBACK_MARKET_RESOURCES: ResourceDefinition[] = [
  { resourceId: 1, name: '' }, { resourceId: 2, name: '' }, { resourceId: 3, name: '' },
  { resourceId: 4, name: '' }, { resourceId: 5, name: '' }, { resourceId: 6, name: '' },
  { resourceId: 7, name: '' }, { resourceId: 8, name: '' }, { resourceId: 9, name: '' },
  { resourceId: 10, name: '' }, { resourceId: 11, name: '' }, { resourceId: 12, name: '' },
]

export function resourceIcon(resourceId: number): string {
  return iconLookup(resourceId)
}

export function formatResourceName(resourceId: number): string {
  return `Resource #${resourceId}`
}

export function resourceName(resourceId: number): string {
  return `Resource #${resourceId}`
}
