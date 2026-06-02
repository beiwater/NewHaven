// Central mapping from game entity IDs to asset paths.
// Assets live under the public /assets/ URL root.

const BASE_ITEMS = '/assets/items'
const BASE_BUILDINGS = '/assets/icons/buildings'
const BASE_SYSTEM = '/assets/icons/system'

export const RESOURCE_ICONS: Record<number, string> = {
  1: `${BASE_ITEMS}/01_grain.png`,
  2: `${BASE_ITEMS}/03_flour.png`,
  3: `${BASE_ITEMS}/04_dough.png`,
  4: `${BASE_ITEMS}/09_pizza.png`,
}

export function resourceIcon(id: number): string {
  return RESOURCE_ICONS[id] ?? RESOURCE_ICONS[1]
}

export const BUILDING_ICONS: Record<number, string> = {
  1: `${BASE_BUILDINGS}/01_farm.png`,
  2: `${BASE_BUILDINGS}/03_mill.png`,
  3: `${BASE_BUILDINGS}/05_bakery.png`,
  4: `${BASE_BUILDINGS}/09_restaurant.png`,
}

export function buildingIcon(id: number): string {
  return BUILDING_ICONS[id] ?? BUILDING_ICONS[1]
}

export const SYSTEM_ICONS: Record<string, string> = {
  research: `${BASE_SYSTEM}/01_research.png`,
  financial: `${BASE_SYSTEM}/02_financial_report.png`,
  executive: `${BASE_SYSTEM}/03_executive.png`,
  leaderboard: `${BASE_SYSTEM}/04_leaderboard.png`,
  powerup: `${BASE_SYSTEM}/05_power_up.png`,
  chat: `${BASE_SYSTEM}/06_chat.png`,
  market: `${BASE_SYSTEM}/07_market.png`,
  inventory: `${BASE_SYSTEM}/08_inventory.png`,
  settings: `${BASE_SYSTEM}/09_settings.png`,
  notification: `${BASE_SYSTEM}/10_notification.png`,
  quest: `${BASE_SYSTEM}/11_quest.png`,
  achievement: `${BASE_SYSTEM}/12_achievement.png`,
}

export function systemIcon(name: string): string {
  return SYSTEM_ICONS[name] ?? SYSTEM_ICONS.research
}
