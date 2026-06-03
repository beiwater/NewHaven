import type { ResourceDefinition } from './types'
import { resourceIcon as iconLookup } from '@/game/icons'

export const MARKET_GROUPS = [
  { id: 'core', label: 'Core Chain', ids: [1, 2, 3, 4] },
] as const

export const FALLBACK_MARKET_RESOURCES: ResourceDefinition[] = [
  { resourceId: 1, name: 'Wheat' },
  { resourceId: 2, name: 'Flour' },
  { resourceId: 3, name: 'Bread' },
  { resourceId: 4, name: 'Meals' },
]

export function resourceIcon(resourceId: number): string {
  return iconLookup(resourceId)
}

export function formatResourceName(resourceId: number, resources: ResourceDefinition[]): string {
  return resources.find((r) => r.resourceId === resourceId)?.name ?? `Resource #${resourceId}`
}
