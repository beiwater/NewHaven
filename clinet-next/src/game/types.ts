/** Building on the map */
import type { MapId } from './map.config'

export interface Building {
  id: string
  kind: number
  name?: string
  x?: number
  y?: number
  mapId?: MapId
  slotId?: string
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
  quality: number
  input: Record<number, number>
  output: Record<number, number>
  claimedAmount: number
  claimableAmount?: number
  xpAwarded?: number
  startedAt: string
  completesAt: string
  status: string
  meta?: Record<string, unknown>
}

/** Production queue (actual API shape) */
export interface ProductionQueue {
  byBuilding: Record<string, ProductionJob[]>
  maxSlots: number
  inUse: number
}

/** Market order */
export interface MarketOrder {
  id: string
  resourceId: number
  kind: string
  price: number
  quantity: number
  filled: number
  status: string
  companyId?: number
  playerName?: string
}

export interface MarketDepth {
  buys: Array<{ price: number; quantity: number }>
  sells: Array<{ price: number; quantity: number }>
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
  amount: number
  reward: number
  completed: boolean
  expiresAt: number
}

/** Government Contract */
export interface GovContract {
  id: string
  resourceId: number
  totalAmount: number
  fulfilledAmount: number
  reward: number
  expiresAt: number
  playerName?: string
}

/** Warehouse inventory item from API */
export interface WarehouseItem {
  resourceId: number
  name: string
  quantity: number
  quality: number
  estimatedValue: number
}

export interface WarehouseData {
  inventory: WarehouseItem[]
  capacity: number
  used: number
}

export interface ResourceDefinition {
  resourceId: number
  name: string
}
