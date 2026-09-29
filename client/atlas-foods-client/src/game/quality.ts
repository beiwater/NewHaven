import type { ResourceDefinition } from './types'

export const MIN_PRODUCT_QUALITY = 0
export const MAX_PRODUCT_QUALITY = 12
export const QUALITY_INPUT_MULTIPLIER = 2

export function clampQuality(quality: number): number {
  return Math.max(MIN_PRODUCT_QUALITY, Math.min(MAX_PRODUCT_QUALITY, Math.trunc(quality)))
}

export function qualitySalesBonusPct(quality: number): number {
  return clampQuality(quality) * 2
}

function baseIngredients(option: ResourceDefinition): Array<{ resourceId: number; quantity: number }> {
  if (option.recipe?.length) return option.recipe
  return Object.entries(option.producedFrom ?? {})
    .map(([resourceId, quantity]) => ({ resourceId: Number(resourceId), quantity }))
    .filter((ingredient) => Number.isInteger(ingredient.resourceId) && ingredient.resourceId > 0 && ingredient.quantity > 0)
}

export function qualityRequirements(option: ResourceDefinition, quality: number): Array<{ resourceId: number; quality: number; quantity: number }> {
  const outputQuality = clampQuality(quality)
  const ingredients = baseIngredients(option)
  if (outputQuality === 0) {
    return ingredients.map((ingredient) => ({ ...ingredient, quality: 0 }))
  }
  if (ingredients.length === 0) {
    return [{ resourceId: option.resourceId, quality: outputQuality - 1, quantity: QUALITY_INPUT_MULTIPLIER }]
  }
  return ingredients.map((ingredient) => ({
    ...ingredient,
    quality: outputQuality - 1,
    quantity: ingredient.quantity * QUALITY_INPUT_MULTIPLIER,
  }))
}
