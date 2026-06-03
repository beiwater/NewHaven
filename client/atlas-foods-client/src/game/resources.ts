import type { ResourceDefinition } from './types'
import { resourceIcon as iconLookup } from '@/game/icons'

export const MARKET_GROUPS = [
  { id: 'raw', label: 'Raw Goods', ids: [1, 2, 6, 12] },
  { id: 'processed', label: 'Processed', ids: [3, 4, 5, 7] },
  { id: 'finished', label: 'Finished Food', ids: [8, 9, 10, 11] },
] as const

export const FALLBACK_MARKET_RESOURCES: ResourceDefinition[] = [
  { resourceId: 1, name: 'Grain' },
  { resourceId: 2, name: 'Dairy Milk' },
  { resourceId: 3, name: 'Flour' },
  { resourceId: 4, name: 'Dough' },
  { resourceId: 5, name: 'Butter' },
  { resourceId: 6, name: 'Sugar' },
  { resourceId: 7, name: 'Cheese' },
  { resourceId: 8, name: 'Steak' },
  { resourceId: 9, name: 'Pizza' },
  { resourceId: 10, name: 'Cake' },
  { resourceId: 11, name: 'Coffee' },
  { resourceId: 12, name: 'Vegetables' },
]

export function resourceIcon(resourceId: number): string {
  return iconLookup(resourceId)
}

export function formatResourceName(resourceId: number, resources: ResourceDefinition[]): string {
  return resources.find((r) => r.resourceId === resourceId)?.name ?? `Resource #${resourceId}`
}
