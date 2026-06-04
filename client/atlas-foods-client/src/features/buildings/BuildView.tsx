import { useMemo, useState, useEffect, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import { useBuyBuilding, usePlaceBuilding, useBuildings } from '@/api/buildings.api'
import { Icon } from '@/features/ui/Icon'
import { useUIStore } from '@/store/ui.store'
import { useCompany, usePlayerLevel } from '@/api/company.api'
import { audio } from '@/audio/AudioManager'
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
  const { data: marketData } = useQuery({
    queryKey: ['buildingMarket'],
    queryFn: () => api.get<BuildingMarketItem[]>('/api/v2/buildings/market/'),
  })
  const { data: placedData } = useBuildings()
  const buyBuilding = useBuyBuilding()
  const placeBuilding = usePlaceBuilding()
  const startBuildingPlacement = useUIStore((s) => s.startBuildingPlacement)
  const currentMapIdRaw = useUIStore((s) => s.currentMapId)
  const { data: companyData } = useCompany()
  const { data: levelData } = usePlayerLevel()

  const marketItems = Array.isArray(marketData) ? marketData : []
  const allBuildings = Array.isArray(placedData) ? placedData : []
  const placedBuildings = allBuildings.filter((b) => b.placed !== false) as Building[]
  const unplacedBuildings = allBuildings.filter((b) => b.placed === false) as Building[]
  
  const playerLevel = levelData?.level ?? companyData?.levelInfo?.level ?? 1
  const currentMapId: string = isMapUnlocked(currentMapIdRaw, playerLevel) ? currentMapIdRaw : 'harbor'

  const allSlotsInfo = useMemo(() => availableMaps(placedBuildings, playerLevel), [placedBuildings, playerLevel])
  const currentMapInfo = allSlotsInfo.find((m) => m.mapId === currentMapId)

  // Best auto-placement candidate
  const bestPlacement = useMemo(
    () => findBestPlacement(placedBuildings, currentMapId, playerLevel),
    [placedBuildings, currentMapId, playerLevel],
  )

  // Auto-open MapPicker on new purchase when multiple maps unlocked
  const [pickerBuildingId, setPickerBuildingId] = useState<string | null>(null)
  const prevUnplacedCount = useRef(unplacedBuildings.length)
  useEffect(() => {
    if (unplacedBuildings.length > prevUnplacedCount.current && allSlotsInfo.filter(m => m.unlocked).length > 1) {
      const newest = unplacedBuildings[unplacedBuildings.length - 1]
      if (newest) setPickerBuildingId(newest.id)
    }
    prevUnplacedCount.current = unplacedBuildings.length
  }, [unplacedBuildings.length, allSlotsInfo])

  const buyMessage = buyBuilding.error instanceof Error
    ? buyBuilding.error.message
    : buyBuilding.data?.building
      ? `${buyBuilding.data.building.name ?? 'Building'} 已购买。`
      : ''
  const placeMessage = placeBuilding.error instanceof Error
    ? placeBuilding.error.message
    : placeBuilding.data?.building
      ? `${placeBuilding.data.building.name ?? 'Building'} 已放置。`
      : ''

  const handleBuy = (buildingId: string) => {
    audio.playSfx('build_confirm')
    buyBuilding.mutate(buildingId)
  }

  // Auto-place: best available map. If current map full, show picker.
  const handleAutoPlace = (buildingId: string) => {
    audio.playSfx('build_place')
    if (bestPlacement) {
      placeBuilding.mutate({ buildingId, mapId: bestPlacement.mapId, slotId: bestPlacement.slotId })
    } else {
      // No slots anywhere → show picker (so user can see lock info)
      setPickerBuildingId(buildingId)
    }
  }

  // Manual placement: enter map placement mode
  const handleManualPlace = (buildingId: string) => {
    audio.playSfx('build_confirm')
    const hasAnySlot = allSlotsInfo.some((m) => m.availableSlots > 0)
    if (!hasAnySlot) {
      setPickerBuildingId(buildingId) // show map info with lock status
      return
    }
    // Try switching to a map with available slots
    const target = allSlotsInfo.find((m) => m.availableSlots > 0)
    if (target && target.mapId !== currentMapId) {
      useUIStore.getState().setCurrentMapId(target.mapId)
    }
    startBuildingPlacement(buildingId)
  }

  const handlePickerPlace = (mapId: string, slotId: string) => {
    setPickerBuildingId(null)
    if (pickerBuildingId) {
      placeBuilding.mutate({ buildingId: pickerBuildingId, mapId, slotId })
    }
  }
  const marketList = marketItems.length === 0 ? (
    <div className="text-xs text-amber-400 italic text-center py-8">加载中...</div>
  ) : (
    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
      {(marketItems as BuildingMarketItem[]).map((item) => {
        const locked = playerLevel < (item.unlockLevel ?? 1)
        return (
          <div key={item.id} className={'rounded-xl p-4 border transition-all duration-200 ' + (locked ? 'bg-amber-50/40 border-amber-200/20 opacity-50' : 'bg-white border-amber-200/60 shadow-sm hover:shadow-md hover:border-amber-300')}>
            <div className="flex items-start justify-between mb-2">
              <span className="text-sm font-bold text-amber-900">{item.name}</span>
              <span className="text-xs font-semibold text-amber-700 bg-amber-100/70 px-2 py-0.5 rounded-full whitespace-nowrap ml-2">{item.cost.toLocaleString()}¥</span>
            </div>
            <p className="text-[11px] text-amber-600/80 leading-relaxed mb-3 min-h-[2em]">{item.description ?? ''}</p>
            <div className="flex items-center gap-2">
              <button
                onClick={() => handleBuy(item.id)}
                disabled={locked || buyBuilding.isPending}
                className="flex-1 py-2 bg-amber-700 hover:bg-amber-800 active:bg-amber-900 disabled:bg-amber-300 text-white text-xs font-bold rounded-lg transition-colors"
              >
                {locked ? `Lv.${item.unlockLevel} 解锁` : buyBuilding.isPending ? '购买中...' : `购买 ${item.cost.toLocaleString()}¥`}
              </button>
            </div>
            {item.produces && (
              <div className="flex items-center gap-1.5 mt-2 pt-2 border-t border-amber-100">
                <span className="text-[10px] text-amber-500 font-medium">生产:</span>
                {item.produces.map((r) => (
                  <span key={r} className="text-[10px] bg-amber-100 text-amber-700 px-2 py-0.5 rounded-full font-medium">{r}</span>
                ))}
              </div>
            )}
          </div>
        )
      })}
    </div>
  )

  return (
    <div className="h-full overflow-y-auto">
      <div className="mx-auto max-w-2xl p-6 space-y-6">
        {/* Header with slot pills */}
        <div className="flex items-center justify-between flex-wrap gap-3">
          <h2 className="text-lg font-bold text-amber-900 flex items-center gap-2">
            <Icon system name="market" className="w-6 h-6" />
            建造
          </h2>
          <div className="flex items-center gap-2 flex-wrap">
            {allSlotsInfo.map((m) => (
              <span key={m.mapId} className={'text-[10px] font-medium px-2.5 py-1 rounded-full ' + (m.unlocked ? 'bg-amber-100 text-amber-700' : 'bg-amber-50 text-amber-400 border border-amber-200')}>
                {m.config.name}: {m.usedSlots}/{m.totalSlots}
                {m.unlocked ? '' : ` · Lv.${m.config.unlockLevel}`}
              </span>
            ))}
          </div>
        </div>

        {/* Buy / place feedback */}
        {buyMessage && (
          <div className={'rounded-xl px-4 py-3 text-xs font-semibold border ' + (buyBuilding.error ? 'bg-red-50 text-red-700 border-red-200' : 'bg-green-50 text-green-700 border-green-200')}>
            {buyMessage}
          </div>
        )}
        {placeMessage && (
          <div className={'rounded-xl px-4 py-3 text-xs font-semibold border ' + (placeBuilding.error ? 'bg-red-50 text-red-700 border-red-200' : 'bg-green-50 text-green-700 border-green-200')}>
            {placeMessage}
          </div>
        )}

        {/* Available Buildings Market */}
        <div>
          <h3 className="text-xs font-bold text-amber-700 uppercase tracking-wider mb-3">可用建筑</h3>
          {marketList}
        </div>

        {/* Unplaced Buildings */}
        {unplacedBuildings.length > 0 && (
          <div>
            <h3 className="text-xs font-bold text-amber-700 uppercase tracking-wider mb-3">等待放置</h3>
            <div className="space-y-3">
              {unplacedBuildings.map((b) => {
                const canAuto = bestPlacement && allSlotsInfo.some(m => m.availableSlots > 0)
                return (
                  <div key={b.id} className="rounded-xl bg-white border border-amber-200/60 shadow-sm p-4">
                    <div className="flex items-center justify-between mb-3">
                      <div className="flex items-center gap-2">
                        <div className="w-8 h-8 rounded-lg bg-amber-100 flex items-center justify-center text-sm font-bold text-amber-800">{b.kind}</div>
                        <div>
                          <div className="text-sm font-bold text-amber-900">{b.name ?? ('建筑 ' + b.kind)}</div>
                          <div className="text-[10px] text-amber-500">等待放置</div>
                        </div>
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleManualPlace(b.id)}
                        className="flex-1 py-2 bg-amber-700 hover:bg-amber-800 text-white text-xs font-bold rounded-lg transition-colors"
                      >手动放置</button>
                      <button
                        onClick={() => handleAutoPlace(b.id)}
                        disabled={placeBuilding.isPending || !canAuto}
                        className="flex-1 py-2 bg-emerald-600 hover:bg-emerald-700 disabled:bg-amber-300 text-white text-xs font-bold rounded-lg transition-colors"
                      >{placeBuilding.isPending ? '放置中...' : canAuto ? '自动放置' : '无可用地块'}</button>
                      <button
                        onClick={() => setPickerBuildingId(b.id)}
                        className="py-2 px-3 bg-blue-600 hover:bg-blue-700 text-white text-xs font-bold rounded-lg transition-colors"
                      >选择地图</button>
                    </div>
                    {currentMapInfo && currentMapInfo.availableSlots === 0 && (
                      <p className="mt-2 text-[10px] text-red-500">
                        {currentMapInfo.unlocked ? '当前地图已满，试试其他地图或升级解锁' : `当前地图 Lv.${currentMapInfo.config.unlockLevel} 解锁`}
                      </p>
                    )}
                  </div>
                )
              })}
            </div>
          </div>
        )}

        {/* Hint when all map slots are full */}
        {unplacedBuildings.length > 0 && !bestPlacement && (
          <div className="rounded-xl bg-amber-50 border border-amber-200/60 px-4 py-3 text-xs text-amber-600 text-center">
            所有地图地块已满，升级或拆除旧建筑腾出空间。
          </div>
        )}
      </div>

      {/* MapPicker overlay */}
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