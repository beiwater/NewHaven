/** Executive data model aligned with Executive Curve spec */

export type ExecutiveRarity = 'Legendary' | 'Epic' | 'Rare' | 'Common'
export type ExecutivePosition = 'coo' | 'cfo' | 'cmo' | 'cto' | ''

export interface ExecutiveSkills {
  management: number
  accounting: number
  communication: number
  science: number
}

/** A single executive — market candidate or owned */
export interface Executive {
  id: string
  name: string
  title: string
  specialty: Exclude<ExecutivePosition, ''>
  position: ExecutivePosition
  skills: ExecutiveSkills
  level: number
  rarity: ExecutiveRarity
  stage: string
  salary: number       // per hour
  productionBonus: number  // percent (e.g. 4 means +4%)
  salesBonus: number       // percent
  mgmtDiscount: number     // percent
  recruitCost: number
  trainingCost: number
  trainingTime: number     // seconds
  status: 'idle' | 'training' | 'recruiting'
  trainingEndTime?: string // ISO timestamp
}

/** Shape returned by POST /api/v2/executives/search/ */
export interface ExecutiveSearchResult {
  executives: Executive[]
  total: number
  refreshCooldown: string
}

/** Shape returned by POST /api/v2/executives/recruit/ */
export interface RecruitResult {
  ok: boolean
  executive?: Executive
  error?: string
}

/** Shape returned by POST /api/v2/executives/train/{id}/ */
export interface TrainResult {
  ok: boolean
  executive?: Executive
  error?: string
}

export interface AssignExecutivePositionResult {
  ok: boolean
  executive?: Executive
  error?: string
}

/** Shape returned by GET /api/v3/executives/{id}/ */
export interface ExecutiveDetail {
  id: string
  name: string
  title: string
  specialty: Exclude<ExecutivePosition, ''>
  position: ExecutivePosition
  skills: ExecutiveSkills
  level: number
  rarity: ExecutiveRarity
  stage: string
  salary: number
  productionBonus: number
  salesBonus: number
  mgmtDiscount: number
  trainingCost: number
  trainingTime: number
  status: 'idle' | 'training'
  trainingEndTime?: string
  morale: number
}

// ── Rarity ordering ────────────────────────────────────

export const RARITY_ORDER: Record<ExecutiveRarity, number> = {
  Legendary: 4,
  Epic: 3,
  Rare: 2,
  Common: 1,
}

export const RARITY_COLORS: Record<ExecutiveRarity, string> = {
  Legendary: 'text-orange-600 bg-orange-100 border-orange-400',
  Epic: 'text-purple-700 bg-purple-100 border-purple-400',
  Rare: 'text-blue-700 bg-blue-100 border-blue-400',
  Common: 'text-gray-600 bg-gray-100 border-gray-400',
}

export const RARITY_BG: Record<ExecutiveRarity, string> = {
  Legendary: 'bg-gradient-to-br from-orange-50 to-amber-100 border-orange-300',
  Epic: 'bg-gradient-to-br from-purple-50 to-violet-100 border-purple-300',
  Rare: 'bg-gradient-to-br from-blue-50 to-sky-100 border-blue-300',
  Common: 'bg-gradient-to-br from-gray-50 to-stone-100 border-gray-300',
}

// ── Executive curve helpers ────────────────────────────

/**
 * Production bonus per level.
 * Marginal convergence: each level adds less than the previous.
 * Lv1 base: 2%, increment decays by 0.12 per level.
 */
export function productionBonusAtLevel(level: number): number {
  if (level < 1) return 0
  const base = 2.0
  const increment = 0.9 - level * 0.06
  const raw = base + (level - 1) * Math.max(0.12, increment)
  return Math.round(raw * 10) / 10
}

/**
 * Sales bonus per level.
 * Sales scales faster than production. Lv1 base: 4%, decays slower.
 */
export function salesBonusAtLevel(level: number): number {
  if (level < 1) return 0
  const base = 4.0
  const increment = 1.6 - level * 0.08
  const raw = base + (level - 1) * Math.max(0.2, increment)
  return Math.round(raw * 10) / 10
}

/**
 * Management discount per level.
 * Lv1 base: 2%, decays moderately.
 */
export function mgmtDiscountAtLevel(level: number): number {
  if (level < 1) return 0
  const base = 2.0
  const increment = 0.7 - level * 0.035
  const raw = base + (level - 1) * Math.max(0.1, increment)
  return Math.round(raw * 10) / 10
}

/**
 * Training cost (cash) to go from `fromLevel` to `fromLevel + 1`.
 * Early levels cheap, then rises steeply mid-game (sweet spot ~7).
 * Formula: cost = 5000 * level^1.6
 */
export function trainingCost(level: number): number {
  return Math.round(5000 * Math.pow(level, 1.6))
}

/**
 * Training time (seconds) to go from `fromLevel` to `fromLevel + 1`.
 * Scales with level.
 * Formula: time = 3600 * level^0.7
 */
export function trainingTimeSeconds(level: number): number {
  return Math.round(3600 * Math.pow(level, 0.7))
}

/**
 * Salary per hour at a given level.
 * Salary = 600 + 80 * level^1.3
 */
export function salaryAtLevel(level: number): number {
  return Math.round(600 + 80 * Math.pow(level, 1.3))
}

/**
 * Recruit cost based on rarity and level.
 */
export function recruitCost(rarity: ExecutiveRarity, level: number): number {
  const rarityFactor: Record<ExecutiveRarity, number> = {
    Legendary: 2.5,
    Epic: 1.8,
    Rare: 1.2,
    Common: 0.8,
  }
  return Math.round(15000 * rarityFactor[rarity] * Math.pow(level, 0.8))
}

/**
 * Stage name based on level.
 */
export function stageAtLevel(level: number): string {
  if (level >= 10) return 'Executive VP'
  if (level >= 8) return 'Director'
  if (level >= 6) return 'Senior Manager'
  if (level >= 4) return 'Manager'
  if (level >= 2) return 'Junior Manager'
  return 'Trainee'
}

/**
 * Compute training cost and time for a given executive.
 */
export function computeTrainingCost(exec: {
  level: number
  rarity: ExecutiveRarity
}): { cost: number; timeSeconds: number } {
  return {
    cost: trainingCost(exec.level),
    timeSeconds: trainingTimeSeconds(exec.level),
  }
}

/**
 * Format seconds into a human-readable duration string.
 */
export function formatDuration(seconds: number): string {
  if (seconds <= 0) return '0s'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  const parts: string[] = []
  if (h > 0) parts.push(`${h}h`)
  if (m > 0) parts.push(`${m}m`)
  if (s > 0) parts.push(`${s}s`)
  return parts.join(' ') || '0s'
}

/**
 * Format a large number with commas.
 */
export function formatMoney(amount: number): string {
  return amount.toLocaleString('en-US')
}
