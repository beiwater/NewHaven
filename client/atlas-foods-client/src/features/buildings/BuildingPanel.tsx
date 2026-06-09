import { useTranslation } from 'react-i18next'
import { useUIStore } from '@/store/ui.store'
import { useBuildings } from '@/api/buildings.api'
import { useClaimableJobs, useClaimAll } from '@/api/production.api'
import { useCompany } from '@/api/company.api'
import { isMapUnlocked, MAPS, placeableSlots, type MapId } from '@/game/map.config'
import { buildingIcon } from '@/game/icons'
import { BuildingCard } from './BuildingCard'
export function BuildingPanel() {
  const { t } = useTranslation()
  const selectedBuildingId = useUIStore((s) => s.selectedBuildingId)
  const selectBuilding = useUIStore((s) => s.selectBuilding)
  const currentMapIdRaw = useUIStore((s) => s.currentMapId)
  const { data: buildingsData } = useBuildings()
  const { data: claimableData } = useClaimableJobs()
  const { data: companyData } = useCompany()
  const claimAll = useClaimAll()
  const level = companyData?.levelInfo?.level ?? 1
  const currentMapId: MapId = isMapUnlocked(currentMapIdRaw, level) ? currentMapIdRaw : 'harbor'

  // buildings: null when empty, [] when fetched with no buildings
  const allPlacedBuildings = Array.isArray(buildingsData)
    ? buildingsData.filter((b) => b.placed !== false)
    : []
  const buildings = allPlacedBuildings.filter((building) => {
    const id = building.mapId ?? ''
    return (MAPS[id] ? id : 'harbor') === currentMapId
  })
  const selected = buildings.find((b) => b.id === selectedBuildingId) ?? null

  // claimableData: ProductionJob[] directly from API
  const claimableCount = Array.isArray(claimableData) ? claimableData.length : 0
  const openMapPlots = placeableSlots(currentMapId, level).length

  return (
    <div className="right-panel flex flex-col bg-amber-50 border-l-2 border-amber-700/30 overflow-y-auto">
      {selected ? (
        <BuildingCard building={selected} hasFreeSlots={openMapPlots > buildings.length} onClose={() => selectBuilding(null)} />
      ) : (
        <>
          <div className="p-3 border-b border-amber-200/60 flex items-center justify-between">
            <div>
              <h2 className="text-xs font-bold text-amber-800 uppercase tracking-wider">{t('nav.buildings')}</h2>
              <p className="text-[10px] text-amber-600/70 mt-0.5">
                地块 {buildings.length}/{openMapPlots}
              </p>
            </div>
            <span className="text-[10px] font-medium text-amber-600 bg-amber-100/70 px-2 py-1 rounded-full">
              {MAPS[currentMapId].name}
            </span>
          </div>

          {claimableCount > 0 && (
            <button
              onClick={() => claimAll.mutate()}
              className="mx-3 mt-2 py-1.5 bg-green-600 hover:bg-green-700 text-white text-xs font-semibold rounded-md transition-colors"
            >
              {t('building.collect')} ({claimableCount})
            </button>
          )}

          <div className="flex-1 overflow-y-auto p-2 space-y-1.5">
            {buildings.length === 0 && (
              <div className="text-xs text-amber-400 italic text-center py-4">
                {t('building.noBuildingsYet')}
              </div>
            )}
            {buildings.map((b) => (
              <button
                key={b.id}
                onClick={() => selectBuilding(b.id)}
                className="w-full text-left p-2.5 rounded-lg bg-white/60 hover:bg-amber-100/60 border border-amber-200/40 transition-colors"
              >
                <div className="flex items-center gap-2">
                  <img src={buildingIcon(b.kind)} alt="" className="w-7 h-7 rounded bg-amber-100 object-contain p-1" />
                  <div className="flex-1 min-w-0">
                    <div className="text-xs font-semibold text-amber-900 truncate">
                      {t('building.name_' + b.kind, b.name ?? '')}
                    </div>
                    <div className="text-[10px] text-amber-600/70">
                      {t('building.level', { level: b.level })} · {b.status ?? t('building.idle')}
                    </div>
                  </div>
                </div>
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  )
}
