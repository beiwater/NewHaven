import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import { useBuyBuilding, usePlaceBuilding, useBuildings } from '@/api/buildings.api'
import { Icon } from '@/features/ui/Icon'
import { FALLBACK_MARKET_RESOURCES, formatResourceName } from '@/game/resources'

interface BuildingMarketItem {
  id: string
  name: string
  kind: number
  cost: number
  description?: string
  produces?: number[]
  starterProduces?: number[]
  starterRole?: string
  unlockLevel?: number
}

export function BuildView() {
  const [placeX, setPlaceX] = useState('5')
  const [placeY, setPlaceY] = useState('5')

  const { data: marketData } = useQuery({
    queryKey: ['buildingMarket'],
    queryFn: () => api.get<BuildingMarketItem[]>('/api/v2/buildings/market/'),
  })
  const { data: placedData } = useBuildings()
  const buyBuilding = useBuyBuilding()
  const placeBuilding = usePlaceBuilding()

  const marketItems = Array.isArray(marketData) ? marketData : []
  const allBuildings = Array.isArray(placedData) ? placedData : []
  const placedBuildings = allBuildings.filter((b) => b.placed !== false)
  const unplacedBuildings = allBuildings.filter((b) => b.placed === false)
  const buyMessage = buyBuilding.error instanceof Error
    ? buyBuilding.error.message
    : buyBuilding.data?.building
      ? `${buyBuilding.data.building.name ?? 'Building'} purchased and placed on the map.`
      : ''
  const placeMessage = placeBuilding.error instanceof Error
    ? placeBuilding.error.message
    : placeBuilding.data?.building
      ? `${placeBuilding.data.building.name ?? 'Building'} placed.`
      : ''

  const handleBuy = (buildingId: string) => {
    buyBuilding.mutate(buildingId, {
      onSuccess: (data) => {
        const position = findNextBuildingSpot(placedBuildings)
        placeBuilding.mutate({
          buildingId: data.building.id,
          x: position.x,
          y: position.y,
        })
      },
    })
  }

  const handlePlace = (buildingId: string) => {
    placeBuilding.mutate({
      buildingId,
      x: parseInt(placeX) || 5,
      y: parseInt(placeY) || 5,
    })
  }

  return (
    <div className="h-full overflow-y-auto p-4 space-y-4">
      <h2 className="text-lg font-bold text-amber-900 flex items-center gap-2">
        <Icon name="icon_builder_v1" className="w-6 h-6" />
        Build
      </h2>

      {/* Available Buildings Market */}
      <div>
        <h3 className="text-xs font-bold text-amber-700 uppercase tracking-wider mb-2">
          Available Buildings{marketItems.length > 0 ? ` (${marketItems.length})` : ''}
        </h3>
        {buyMessage && (
          <div className={`mb-2 rounded-lg px-3 py-2 text-xs font-semibold ${
            buyBuilding.error ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200'
          }`}>
            {buyMessage}
          </div>
        )}
        {marketItems.length === 0 && (
          <div className="text-xs text-amber-400 italic py-2">Loading available buildings...</div>
        )}
        <div className="space-y-2">
          {marketItems.map((item) => (
            <div
              key={item.id}
              className="bg-white/60 rounded-lg border border-amber-200/40 p-3"
            >
              <div className="flex items-start justify-between mb-2">
                <div>
                  <div className="text-sm font-bold text-amber-900">{item.name}</div>
                  <div className="text-[10px] text-amber-600">Kind {item.kind}</div>
                  {item.description && (
                    <div className="text-[10px] text-amber-600 mt-0.5">{item.description}</div>
                  )}
                  {item.starterRole && (
                    <div className="mt-1 rounded bg-green-50 px-2 py-1 text-[10px] font-semibold text-green-700">
                      Starter chain: {item.starterRole}
                    </div>
                  )}
                  {item.starterProduces && item.starterProduces.length > 0 && (
                    <div className="mt-1 text-[10px] text-green-800">
                      Recommended: {item.starterProduces
                        .map((id) => formatResourceName(id, FALLBACK_MARKET_RESOURCES))
                        .join(' -> ')}
                    </div>
                  )}
                  {item.produces && item.produces.length > 0 && (
                    <div className="mt-1 text-[10px] text-amber-700">
                      Produces: {item.produces
                        .slice(0, 4)
                        .map((id) => formatResourceName(id, FALLBACK_MARKET_RESOURCES))
                        .join(', ')}
                      {item.produces.length > 4 ? ` +${item.produces.length - 4} more` : ''}
                    </div>
                  )}
                  {item.produces && item.produces.length === 0 && (
                    <div className="mt-1 text-[10px] text-amber-500">Support building, no direct production.</div>
                  )}
                </div>
                <button
                  onClick={() => handleBuy(item.id)}
                  disabled={buyBuilding.isPending}
                  className="px-3 py-1.5 bg-amber-700 hover:bg-amber-800 disabled:bg-amber-400 text-white text-xs font-bold rounded transition-colors"
                >
                  {buyBuilding.isPending ? '...' : `Buy $${item.cost?.toLocaleString() ?? '?'}`}
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Placed Buildings */}
      <div>
        <h3 className="text-xs font-bold text-amber-700 uppercase tracking-wider mb-2">
          Your Buildings ({placedBuildings.length})
        </h3>
        {placedBuildings.length === 0 && (
          <div className="text-xs text-amber-400 italic py-2">No buildings placed yet</div>
        )}
        <div className="space-y-1.5">
          {placedBuildings.map((b) => (
            <div key={b.id} className="flex items-center gap-2 p-2 bg-white/40 rounded border border-amber-200/30 text-xs">
              <Icon name="icon_level_badge_v1" className="w-5 h-5" />
              <span className="text-amber-900 font-semibold">{b.name ?? `Building ${b.kind}`}</span>
              <span className="text-amber-500">Lv.{b.level}</span>
              <span className="ml-auto text-amber-600">({b.x ?? '-'}, {b.y ?? '-'})</span>
            </div>
          ))}
        </div>
      </div>

      {/* Unplaced Buildings (from API) */}
      {unplacedBuildings.length > 0 && (
        <div>
          <h3 className="text-xs font-bold text-amber-700 uppercase tracking-wider mb-2">
            Unplaced Buildings
          </h3>
          {placeMessage && (
            <div className={`mb-2 rounded-lg px-3 py-2 text-xs font-semibold ${
              placeBuilding.error ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200'
            }`}>
              {placeMessage}
            </div>
          )}
          <div className="space-y-2">
            {unplacedBuildings.map((b) => (
              <div key={b.id} className="bg-amber-100/60 rounded-lg p-3 border border-amber-300/40">
                <div className="text-sm font-bold text-amber-900 mb-2">{b.name ?? `Building ${b.kind}`}</div>
                <div className="flex gap-2">
                  <input
                    type="number"
                    value={placeX}
                    onChange={(e) => setPlaceX(e.target.value)}
                    placeholder="X"
                    className="w-16 px-2 py-1 text-xs bg-white border border-amber-300 rounded text-amber-900"
                  />
                  <input
                    type="number"
                    value={placeY}
                    onChange={(e) => setPlaceY(e.target.value)}
                    placeholder="Y"
                    className="w-16 px-2 py-1 text-xs bg-white border border-amber-300 rounded text-amber-900"
                  />
                  <button
                    onClick={() => handlePlace(b.id)}
                    disabled={placeBuilding.isPending}
                    className="px-3 py-1 bg-green-600 hover:bg-green-700 disabled:bg-green-400 text-white text-xs font-bold rounded transition-colors"
                  >
                    Place
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

function findNextBuildingSpot(buildings: Array<{ x?: number; y?: number }>): { x: number; y: number } {
  const occupied = new Set(buildings.map((building) => `${building.x ?? 0}:${building.y ?? 0}`))
  for (let y = 1; y <= 8; y += 1) {
    for (let x = 1; x <= 8; x += 1) {
      if (!occupied.has(`${x}:${y}`)) {
        return { x, y }
      }
    }
  }
  return { x: buildings.length + 1, y: 1 }
}
