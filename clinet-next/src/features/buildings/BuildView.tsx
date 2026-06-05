import { useMemo, useState, useEffect, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import { useBuyBuilding, usePlaceBuilding, useBuildings } from '@/api/hooks/buildings.hooks'
import { Icon } from '@/features/ui/Icon'
import { useUIStore } from '@/store/ui.store'
import { useCompany } from '@/api/hooks/company.hooks'
import { isMapUnlocked } from '@/game/map.config'
import { availableMaps, findBestPlacement } from '@/game/map/mapPlacement'
import { MapPicker } from './MapPicker'
import type { Building } from '@/game/types'

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
  const { data: marketData } = useQuery<BuildingMarketItem[]>({
    queryKey: ['buildingMarket'],
    queryFn: () => api.get<BuildingMarketItem[]>('/api/buildings/'),
  })
  const { data: placedData } = useBuildings()
  const buyBuilding = useBuyBuilding()
  const placeBuilding = usePlaceBuilding()
  const currentMapIdRaw = useUIStore((s) => s.currentMapId)
  const { data: companyData } = useCompany()

  const marketItems = Array.isArray(marketData) ? marketData : []
  const allBuildings = Array.isArray(placedData) ? placedData : []
  const placedBuildings = allBuildings.filter((b) => b.placed !== false) as Building[]
  const unplacedBuildings = allBuildings.filter((b) => b.placed === false) as Building[]

  const playerLevel = companyData?.levelInfo?.level ?? 1
  const currentMapId: string = isMapUnlocked(currentMapIdRaw, playerLevel) ? currentMapIdRaw : 'harbor'

  const allSlotsInfo = useMemo(() => availableMaps(placedBuildings, playerLevel), [placedBuildings, playerLevel])
  const currentMapInfo = allSlotsInfo.find((m) => m.mapId === currentMapId)

  const bestPlacement = useMemo(
    () => findBestPlacement(placedBuildings, currentMapId, playerLevel),
    [placedBuildings, currentMapId, playerLevel],
  )

  const [pickerBuildingId, setPickerBuildingId] = useState<string | null>(null)
  const prevUnplacedCount = useRef(unplacedBuildings.length)
  useEffect(() => {
    if (unplacedBuildings.length > prevUnplacedCount.current && allSlotsInfo.length > 1) {
      setPickerBuildingId(unplacedBuildings[0].id)
    }
    prevUnplacedCount.current = unplacedBuildings.length
  }, [unplacedBuildings, allSlotsInfo])

  const handleBuy = (buildingId: string) => {
    buyBuilding.mutate(buildingId)
  }

  const handleAutoPlace = (buildingId: string) => {
    if (bestPlacement) {
      placeBuilding.mutate({
        buildingId,
        mapId: bestPlacement.mapId,
        slotId: bestPlacement.slotId,
      })
    } else {
      setPickerBuildingId(buildingId)
    }
  }

  const handlePickerPlace = (mapId: string, slotId: string) => {
    if (pickerBuildingId) {
      placeBuilding.mutate({ buildingId: pickerBuildingId, mapId, slotId })
    }
    setPickerBuildingId(null)
  }

  const marketList = marketItems.length === 0 ? (
    <div className="text-xs text-amber-400 italic text-center py-8">Loading...</div>
  ) : (
    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
      {marketItems.map((item) => {
        const owned = allBuildings.filter((b) => b.kind === item.kind && !b.placed)
        return (
          <div key={item.id} className="rounded-xl border border-amber-300/60 bg-white/70 p-3 shadow-sm">
            <div className="flex items-center gap-3 mb-2">
              <div className="w-10 h-10 rounded-lg bg-amber-100 flex items-center justify-center">
                <Icon building name={String(item.kind)} className="w-6 h-6" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="text-sm font-bold text-amber-900 truncate">{item.name}</div>
                <div className="text-[10px] text-amber-600">${item.cost?.toLocaleString?.() ?? item.cost}</div>
              </div>
            </div>
            {item.description && (
              <p className="text-[10px] text-amber-700/70 mb-2 line-clamp-2">{item.description}</p>
            )}
            <div className="flex gap-2">
              <button
                onClick={() => handleBuy(item.id)}
                disabled={buyBuilding.isPending}
                className="flex-1 py-1.5 bg-amber-600 hover:bg-amber-700 text-white text-xs font-semibold rounded-md transition-colors disabled:opacity-50"
              >
                Buy
              </button>
              {owned.length > 0 && (
                <button
                  onClick={() => handleAutoPlace(owned[0].id)}
                  className="flex-1 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-md transition-colors"
                >
                  Place
                </button>
              )}
            </div>
          </div>
        )
      })}
    </div>
  )

  return (
    <div className="h-full overflow-y-auto">
      <div className="p-4">
        <h2 className="text-lg font-bold text-amber-900 mb-3">Build</h2>
        <div className="mb-3 flex gap-2 text-[10px] text-amber-600">
          <span>Plots: {currentMapInfo?.usedSlots ?? 0}/{currentMapInfo?.totalSlots ?? 0}</span>
          <span>Available: {currentMapInfo?.availableSlots ?? 0}</span>
        </div>
        {marketList}
      </div>

      {pickerBuildingId && (
        <MapPicker
          buildingId={pickerBuildingId}
          onPlace={handlePickerPlace}
          onCancel={() => setPickerBuildingId(null)}
        />
      )}
    </div>
  )
}
