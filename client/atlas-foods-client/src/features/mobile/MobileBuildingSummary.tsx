import { useUIStore } from '@/store/ui.store'
import { useTranslation } from 'react-i18next'
import { resourceName } from '@/game/resources'
import { useBuildings } from '@/api/buildings.api'
import { useClaimableJobs, useClaimAll } from '@/api/production.api'
import { useCompany } from '@/api/company.api'
import { MAPS, isMapUnlocked, placeableSlots, type MapId } from '@/game/map.config'

export function MobileBuildingSummary() {
  const { t } = useTranslation()
  const selectBuilding = useUIStore((s) => s.selectBuilding)
  const currentMapIdRaw = useUIStore((s) => s.currentMapId)
  const { data: buildingsData } = useBuildings()
  const { data: claimableData } = useClaimableJobs()
  const { data: companyData } = useCompany()
  const claimAll = useClaimAll()
  const level = companyData?.levelInfo?.level ?? 1
  const currentMapId: MapId = isMapUnlocked(currentMapIdRaw, level) ? currentMapIdRaw : 'harbor'

  const buildings = Array.isArray(buildingsData)
    ? buildingsData.filter((b) => {
      if (b.placed === false) return false
      const id = b.mapId ?? ''
      return (MAPS[id] ? id : 'harbor') === currentMapId
    })
    : []
  const claimableCount = Array.isArray(claimableData) ? claimableData.length : 0
  const openMapPlots = placeableSlots(currentMapId, level).length

  return (
    <div className="bg-white/60 rounded-xl border border-amber-300/50 p-3 min-w-[200px] shrink-0">
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-xs font-bold text-amber-800 uppercase tracking-wider">{t('nav.buildings')}</h3>
        <span className="text-[10px] text-amber-600 tabular-nums">
          {t('mobile.plots', { filled: buildings.length, total: openMapPlots })}
        </span>
      </div>

      {claimableCount > 0 && (
        <button
          onClick={() => claimAll.mutate()}
          className="w-full mb-2 py-1 bg-green-600 hover:bg-green-700 text-white text-[10px] font-semibold rounded-md transition-colors"
        >
          {t('mobile.collectAll', { count: claimableCount })}
        </button>
      )}

      {buildings.length === 0 && (
        <div className="text-[10px] text-amber-400 italic text-center py-2">
          {t('mobile.noBuildings')}
        </div>
      )}

      <div className="space-y-1 max-h-[120px] overflow-y-auto">
        {buildings.slice(0, 6).map((b) => (
          <button
            key={b.id}
            onClick={() => selectBuilding(b.id)}
            className="w-full text-left p-1.5 rounded-md bg-amber-50/70 hover:bg-amber-100/60 border border-amber-200/30 transition-colors flex items-center gap-2"
          >
            <div className="w-6 h-6 rounded bg-amber-200 flex items-center justify-center text-[10px] font-bold text-amber-800 shrink-0">
              {b.kind}
            </div>
            <div className="flex-1 min-w-0">
              <div className="text-[9px] font-semibold text-amber-900 truncate">
                {b.name ?? t('mobile.buildingFallback', { kind: b.kind })}
              </div>
              <div className="text-[8px] text-amber-600">
                {t('building.level', { level: b.level })} · {b.status ?? t('mobile.idle')}
              </div>
            </div>
          </button>
        ))}
      </div>
      {buildings.length > 6 && (
        <div className="text-[9px] text-amber-400 text-center mt-1">
          {t('mobile.more', { count: buildings.length - 6 })}
        </div>
      )}
    </div>
  )
}
