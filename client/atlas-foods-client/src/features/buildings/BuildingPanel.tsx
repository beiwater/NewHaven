import { useUIStore } from '@/store/ui.store'
import { useBuildings } from '@/api/buildings.api'
import { useProductionQueue, useClaimableJobs, useClaimAll } from '@/api/production.api'
import { BuildingCard } from './BuildingCard'

export function BuildingPanel() {
  const selectedBuildingId = useUIStore((s) => s.selectedBuildingId)
  const selectBuilding = useUIStore((s) => s.selectBuilding)
  const { data: buildingsData } = useBuildings()
  const { data: queueData } = useProductionQueue()
  const { data: claimableData } = useClaimableJobs()
  const claimAll = useClaimAll()

  // buildings: null when empty, [] when fetched with no buildings
  const buildings = Array.isArray(buildingsData)
    ? buildingsData.filter((b) => b.placed !== false)
    : []
  const selected = buildings.find((b) => b.id === selectedBuildingId) ?? null

  // claimableData: ProductionJob[] directly from API
  const claimableCount = Array.isArray(claimableData) ? claimableData.length : 0

  // queueData: { byBuilding, inUse, maxSlots }
  const inUse = queueData?.inUse ?? 0
  const maxSlots = queueData?.maxSlots ?? 0

  return (
    <div className="right-panel flex flex-col bg-amber-50 border-l-2 border-amber-700/30 overflow-y-auto">
      {selected ? (
        <BuildingCard building={selected} onClose={() => selectBuilding(null)} />
      ) : (
        <>
          <div className="p-3 border-b border-amber-200/60">
            <h2 className="text-xs font-bold text-amber-800 uppercase tracking-wider">Buildings</h2>
            <p className="text-[10px] text-amber-600/70 mt-0.5">
              {inUse} / {maxSlots} slots used
            </p>
          </div>

          {claimableCount > 0 && (
            <button
              onClick={() => claimAll.mutate()}
              className="mx-3 mt-2 py-1.5 bg-green-600 hover:bg-green-700 text-white text-xs font-semibold rounded-md transition-colors"
            >
              Collect All ({claimableCount})
            </button>
          )}

          <div className="flex-1 overflow-y-auto p-2 space-y-1.5">
            {buildings.length === 0 && (
              <div className="text-xs text-amber-400 italic text-center py-4">
                No buildings yet
              </div>
            )}
            {buildings.map((b) => (
              <button
                key={b.id}
                onClick={() => selectBuilding(b.id)}
                className="w-full text-left p-2.5 rounded-lg bg-white/60 hover:bg-amber-100/60 border border-amber-200/40 transition-colors"
              >
                <div className="flex items-center gap-2">
                  <div className="w-7 h-7 rounded bg-amber-200 flex items-center justify-center text-xs font-bold text-amber-800">
                    {b.kind}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="text-xs font-semibold text-amber-900 truncate">
                      {b.name ?? `Building ${b.kind}`}
                    </div>
                    <div className="text-[10px] text-amber-600/70">
                      Lv.{b.level} · {b.status ?? 'idle'}
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
