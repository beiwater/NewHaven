// Central mapping from game entity IDs to asset paths.
// Assets live under the public /assets/ URL root.

const BASE_ITEMS = '/assets/items'
const BASE_BUILDINGS = '/assets/icons/buildings'
const BASE_SYSTEM = '/assets/icons/system'

export const RESOURCE_ICONS: Record<number, string> = {
  1: `${BASE_ITEMS}/01_grain.png`,
  2: `${BASE_ITEMS}/02_dairy_milk.png`,
  3: `${BASE_ITEMS}/03_flour.png`,
  4: `${BASE_ITEMS}/04_dough.png`,
  5: `${BASE_ITEMS}/05_butter.png`,
  6: `${BASE_ITEMS}/06_sugar.png`,
  7: `${BASE_ITEMS}/07_cheese.png`,
  8: `${BASE_ITEMS}/08_steak.png`,
  9: `${BASE_ITEMS}/09_pizza.png`,
  10: `${BASE_ITEMS}/10_cake.png`,
  11: `${BASE_ITEMS}/11_coffee.png`,
  12: `${BASE_ITEMS}/12_vegetables.png`,
}

export function resourceIcon(id: number): string {
  return RESOURCE_ICONS[id] ?? RESOURCE_ICONS[1]
}

export const BUILDING_ICONS: Record<number, string> = {
  1: `${BASE_BUILDINGS}/01_farm.png`,
  2: `${BASE_BUILDINGS}/02_barn.png`,
  3: `${BASE_BUILDINGS}/03_mill.png`,
  4: `${BASE_BUILDINGS}/04_kitchen.png`,
  5: `${BASE_BUILDINGS}/05_bakery.png`,
  6: `${BASE_BUILDINGS}/06_market_stall.png`,
  7: `${BASE_BUILDINGS}/07_cafe.png`,
  8: `${BASE_BUILDINGS}/08_food_truck.png`,
  9: `${BASE_BUILDINGS}/09_restaurant.png`,
  10: `${BASE_BUILDINGS}/10_trading_hub.png`,
  11: `${BASE_BUILDINGS}/11_warehouse.png`,
  12: `${BASE_BUILDINGS}/12_shop.png`,
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
