/** Building on the map */
export interface Building {
  id: string
  kind: number
  name?: string
  x?: number
  y?: number
  level: number
  baseCost?: number
  placed?: boolean
  produces?: number[]
  starterProduces?: number[]
  starterRole?: string
  status?: 'idle' | 'running' | 'ready'
}

/** Production job */
export interface ProductionJob {
  id: string
  buildingId: string
  resourceId: number
  amount: number
  claimedAmount?: number
  claimableAmount?: number
  input: Record<string, number>
  output: Record<string, number>
  startedAt: string
  completesAt: string
  status: 'running' | 'ready' | 'claimed'
  meta?: Record<string, unknown>
}

/** Production queue (actual API shape) */
export interface ProductionQueue {
  byBuilding: Record<string, unknown>
  inUse: number
  maxSlots: number
}

/** Market order */
export interface MarketOrder {
  id: string
  resourceId: number
  kind: 0 | 1
  price: number
  quality: number
  quantity: number
  remaining: number
  companyId: number
  createdAt: string
  status?: 'open' | 'filled' | 'cancelled'
}

export interface MarketDepth {
  buys: Array<{ price: number; quantity?: number; qty?: number }>
  sells: Array<{ price: number; quantity?: number; qty?: number }>
}

/** Market ticker (actual API shape) */
export interface MarketTickerData {
  resource: number
  series: Array<{ price: number; time: string }>
}

/** Daily Order (actual API shape) */
export interface DailyOrder {
  id: string
  resourceId: number
  quality: number
  quantity: number
  rewardCash: number
  rewardXP: number
  status: string
  createdAt?: string
}

/** Government Contract */
export interface GovContract {
  id: string
  resourceId: number
  quality: number
  quantity: number
  maxPrice?: number
  depositRate?: number
  status: string
  bids?: Array<Record<string, unknown>>
  winnerCompanyId?: number
}

/** Warehouse inventory */
export interface WarehouseItem {
  resourceId: number
  quantity: number
  name?: string
  quality?: number
  estimatedValue?: number
}

export interface WarehouseData {
  capacity: number
  inventory: WarehouseItem[]
  used: number
}

export interface ResourceDefinition {
  resourceId: number
  name: string
  producedFrom?: Record<string, number>
  producedPerHourRaw?: number
  unitsSoldAnHour?: number
  hasEconomyModel?: boolean
  recipe?: Array<{ resourceId: number; resourceName?: string; quantity: number }>
}
