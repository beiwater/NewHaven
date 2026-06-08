import { Assets, Texture } from 'pixi.js'
import { IMAGE_LOAD_TIMEOUT_MS } from '@/constants'
import { buildingIcon } from '@/game/icons'
import { resourceIcon } from '@/game/resources'

export const BUILDING_TEXTURE_URLS: Record<number, string> = {
  1: '/assets/buildings/grain_plot_lv1_idle_trimmed.png',
  2: buildingIcon(2),
  3: '/assets/buildings/mill_house_lv1_idle_trimmed.png',
  4: buildingIcon(4),
  5: '/assets/buildings/bakery_shop_lv1_idle_trimmed.png',
  6: buildingIcon(6),
  7: buildingIcon(7),
  8: buildingIcon(8),
  9: '/assets/buildings/meal_kiosk_lv1_idle_trimmed.png',
  10: buildingIcon(10),
  11: buildingIcon(11),
  12: buildingIcon(12),
}

export const RESOURCE_TEXTURE_URLS: Record<number, string> = Object.fromEntries(
  Array.from({ length: 12 }, (_, i) => [i + 1, resourceIcon(i + 1)]),
)

export function loadImage(url: string): Promise<HTMLImageElement> {
  const img = new Image()
  const { promise, resolve, reject } = Promise.withResolvers<HTMLImageElement>()
  const t = setTimeout(() => reject(new Error(`Timeout: ${url}`)), IMAGE_LOAD_TIMEOUT_MS)
  img.onload = () => { clearTimeout(t); resolve(img) }
  img.onerror = () => { clearTimeout(t); reject(new Error(`Failed: ${url}`)) }
  img.src = url
  return promise
}

export async function preloadBuildingTextures(): Promise<Record<number, Texture>> {
  const urls = Object.values(BUILDING_TEXTURE_URLS)
  const loaded: Record<string, Texture> = await Assets.load(urls)
  const cache: Record<number, Texture> = {}
  for (const [kindStr, url] of Object.entries(BUILDING_TEXTURE_URLS)) {
    if (loaded[url]) cache[Number(kindStr)] = loaded[url]
  }
  return cache
}

export async function preloadResourceTextures(): Promise<Record<number, Texture>> {
  const urls = Object.values(RESOURCE_TEXTURE_URLS)
  const loaded: Record<string, Texture> = await Assets.load(urls)
  const cache: Record<number, Texture> = {}
  for (const [idStr, url] of Object.entries(RESOURCE_TEXTURE_URLS)) {
    if (loaded[url]) cache[Number(idStr)] = loaded[url]
  }
  return cache
}

/**
 * Preload all game textures (buildings + resources + default map background)
 * before the player enters the game. Call early during login loading screen.
 * Errors are swallowed — individual GameCanvas fallbacks still apply.
 */
export async function preloadGameAssets(): Promise<void> {
  await Promise.allSettled([
    preloadBuildingTextures(),
    preloadResourceTextures(),
    loadImage('/assets/backgrounds/map_harbor_v1.png'),
  ])
}
