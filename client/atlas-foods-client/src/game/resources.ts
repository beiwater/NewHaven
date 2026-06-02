import type { ResourceDefinition } from './types'

export const MARKET_GROUPS = [
  { id: 'staples', label: 'Staples', ids: [1, 2, 6, 66, 72, 120] },
  { id: 'farm', label: 'Farm Goods', ids: [3, 4, 5, 9, 115, 116, 117] },
  { id: 'kitchen', label: 'Kitchen Chain', ids: [7, 8, 121, 122, 127, 133, 134, 135, 137, 139, 141] },
] as const

export const FALLBACK_MARKET_RESOURCES: ResourceDefinition[] = [
  { resourceId: 1, name: 'Power' },
  { resourceId: 2, name: 'Water' },
  { resourceId: 3, name: 'Apples' },
  { resourceId: 4, name: 'Oranges' },
  { resourceId: 5, name: 'Grapes' },
  { resourceId: 6, name: 'Grain' },
  { resourceId: 7, name: 'Steak' },
  { resourceId: 8, name: 'Sausages' },
  { resourceId: 9, name: 'Eggs' },
  { resourceId: 66, name: 'Seeds' },
  { resourceId: 72, name: 'Sugarcane' },
  { resourceId: 115, name: 'Cow' },
  { resourceId: 116, name: 'Pig' },
  { resourceId: 117, name: 'Milk' },
  { resourceId: 120, name: 'Vegetables' },
  { resourceId: 121, name: 'Bread' },
  { resourceId: 122, name: 'Cheese' },
  { resourceId: 127, name: 'Pizza' },
  { resourceId: 133, name: 'Flour' },
  { resourceId: 134, name: 'Butter' },
  { resourceId: 135, name: 'Sugar' },
  { resourceId: 137, name: 'Dough' },
  { resourceId: 139, name: 'Fodder' },
  { resourceId: 141, name: 'Vegetable Oil' },
]

export function resourceIcon(resourceId: number): string {
  switch (resourceId) {
    case 6:
      return '/assets/items/item_wheat_v1.png'
    case 121:
      return '/assets/items/item_bread_v1.png'
    case 127:
      return '/assets/items/item_meal_v1.png'
    case 133:
      return '/assets/items/item_flour_v1.png'
    default:
      return '/assets/icons/icon_market_v1.png'
  }
}

export function formatResourceName(resourceId: number, resources: ResourceDefinition[]): string {
  return resources.find((r) => r.resourceId === resourceId)?.name ?? `Resource #${resourceId}`
}
