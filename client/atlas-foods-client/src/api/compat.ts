import type { Building, ProductionJob } from '@/game/types'

type JsonRecord = Record<string, unknown>

function record(value: unknown): JsonRecord {
  return typeof value === 'object' && value !== null ? value as JsonRecord : {}
}

function number(value: unknown, fallback = 0): number {
  return typeof value === 'number' ? value : fallback
}

function string(value: unknown, fallback = ''): string {
  return typeof value === 'string' ? value : fallback
}

function stringMap(value: unknown): Record<string, number> {
  const source = record(value)
  return Object.fromEntries(
    Object.entries(source).filter((entry): entry is [string, number] => typeof entry[1] === 'number'),
  )
}

export function normalizeBuilding(value: unknown): Building {
  const raw = record(value)
  const starterProduces = raw.starterProduces ?? raw.starter_produces
  const shelvesRaw = raw.shelves
  const shelves = Array.isArray(shelvesRaw) ? shelvesRaw.map(sh => ({
    resourceId: number(sh.resourceId ?? sh.resource_id),
    quantity: number(sh.quantity),
    maxQty: number(sh.maxQty ?? sh.max_qty),
    price: number(sh.price),
    priceLock: Boolean(sh.priceLock ?? sh.price_lock),
    revenue: number(sh.revenue, 0),
  })) : undefined
  return {
    id: string(raw.id),
    kind: number(raw.kind ?? raw.buildingId ?? raw.building_id),
    name: string(raw.name) || undefined,
    x: number(raw.x),
    y: number(raw.y),
    mapId: string(raw.mapId ?? raw.map_id) as Building['mapId'],
    slotId: string(raw.slotId ?? raw.slot_id) || undefined,
    level: number(raw.level, 1),
    baseCost: typeof raw.baseCost === 'number' ? raw.baseCost : undefined,
    placed: typeof raw.placed === 'boolean' ? raw.placed : Boolean(raw.mapId ?? raw.map_id),
    produces: Array.isArray(raw.produces) ? raw.produces.filter((v): v is number => typeof v === 'number') : undefined,
    starterProduces: Array.isArray(starterProduces)
      ? starterProduces.filter((v): v is number => typeof v === 'number')
      : undefined,
    starterRole: string(raw.starterRole ?? raw.starter_role) || undefined,
    isRetail: typeof raw.isRetail === 'boolean' ? raw.isRetail : undefined,
    shelves,
  }
}

export function normalizeBuildingList(value: unknown): Building[] {
  const raw = record(value)
  const list = Array.isArray(value) ? value : raw.buildings
  return Array.isArray(list) ? list.map(normalizeBuilding) : []
}

export function normalizeProductionJob(value: unknown): ProductionJob {
  const raw = record(value)
  const startedAt = string(raw.startedAt ?? raw.started_at)
  const durationSeconds = number(raw.durationSeconds ?? raw.duration_seconds)
  const completesAt = string(raw.completesAt ?? raw.completes_at)
    || (startedAt ? new Date(new Date(startedAt).getTime() + durationSeconds * 1000).toISOString() : '')
  return {
    id: string(raw.id ?? raw.jobId ?? raw.job_id),
    buildingId: string(raw.buildingId ?? raw.building_id),
    resourceId: number(raw.resourceId ?? raw.resource_id),
    amount: number(raw.amount ?? raw.targetQuantity ?? raw.target_quantity ?? raw.quantity),
    claimedAmount: number(raw.claimedAmount ?? raw.claimed_amount),
    claimableAmount: number(raw.claimableAmount ?? raw.claimable_amount),
    input: stringMap(raw.input),
    output: stringMap(raw.output),
    startedAt,
    completesAt,
    status: string(raw.status, 'running') as ProductionJob['status'],
  }
}

export function normalizeProductionJobList(value: unknown): ProductionJob[] {
  const raw = record(value)
  const list = Array.isArray(value) ? value : raw.jobs
  return Array.isArray(list) ? list.map(normalizeProductionJob) : []
}

export function camelBuildingResponse(value: unknown): JsonRecord & { building: Building } {
  const raw = record(value)
  return {
    ...raw,
    building: normalizeBuilding(raw.building),
  }
}

export function normalizeWarehouseData(value: unknown): { inventory: Array<{ resourceId: number; quantity: number; quality?: number }>; capacity: number; used: number } {
  const raw = record(value)
  const candidateItems = raw.items ?? raw.inventory
  const itemsRaw: unknown[] = Array.isArray(candidateItems) ? candidateItems : []
  const inventory = itemsRaw.map((item: unknown) => {
    const i = record(item)
    return {
      resourceId: number(i.resourceId ?? i.resource_id),
      quantity: number(i.quantity ?? i.amount ?? i.Amount),
      quality: i.quality !== undefined ? number(i.quality) : undefined,
    }
  })
  return {
    inventory,
    capacity: number(raw.capacity),
    used: number(raw.used ?? raw.used_capacity ?? raw.usedCapacity),
  }
}
