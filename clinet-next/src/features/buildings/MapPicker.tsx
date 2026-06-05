import { useMemo } from 'react'
import { useBuildings } from '@/api/hooks/buildings.hooks'
import { useCompany } from '@/api/hooks/company.hooks'
import { availableMaps } from '@/game/map/mapPlacement'
import type { Building } from '@/game/types'

interface MapPickerProps {
  buildingId: string
  onPlace: (mapId: string, slotId: string) => void
  onCancel: () => void
}

export function MapPicker({ onPlace, onCancel }: MapPickerProps) {
  const { data: buildingsData } = useBuildings()
  const { data: companyData } = useCompany()

  const allBuildings = Array.isArray(buildingsData) ? buildingsData : []
  const placedBuildings = allBuildings.filter((b) => b.placed !== false).map((b) => b as Building)
  const level = companyData?.levelInfo?.level ?? 1

  const maps = useMemo(() => availableMaps(placedBuildings, level), [placedBuildings, level])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="mx-4 w-full max-w-sm rounded-xl border border-amber-600/40 bg-amber-50 p-5 shadow-2xl">
        <h2 className="mb-3 text-base font-bold text-amber-900">Choose Plot</h2>
        <p className="mb-4 text-xs text-amber-600">Select a map to place your building:</p>
        <div className="space-y-2">
          {maps.map((m) => {
            const locked = !m.unlocked
            const full = m.availableSlots <= 0
            return (
              <button
                key={m.mapId}
                disabled={locked || full}
                onClick={() => {
                  if (!m.firstOpenSlotId) return
                  onPlace(m.mapId, m.firstOpenSlotId)
                }}
                className={`w-full rounded-lg border px-4 py-3 text-left transition-colors ${
                  locked
                    ? 'border-amber-300/40 bg-amber-100/50 text-amber-400'
                    : full
                      ? 'border-amber-300/40 bg-amber-100/50 text-amber-500'
                      : 'border-amber-400/60 bg-amber-50 hover:bg-amber-100 text-amber-900'
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="text-sm font-bold">{m.config.name}</span>
                  {locked ? (
                    <span className="rounded bg-amber-200/60 px-2 py-0.5 text-[10px] font-semibold text-amber-600">
                      Lv.{m.config.unlockLevel} unlock
                    </span>
                  ) : full ? (
                    <span className="rounded bg-red-100 px-2 py-0.5 text-[10px] font-semibold text-red-500">
                      Full
                    </span>
                  ) : (
                    <span className="rounded bg-blue-100 px-2 py-0.5 text-[10px] font-semibold text-blue-600">
                      {m.availableSlots} slots
                    </span>
                  )}
                </div>
                <div className="mt-1 text-[10px] text-amber-500">
                  {m.usedSlots}/{m.totalSlots} used
                </div>
              </button>
            )
          })}
        </div>
        <button
          onClick={onCancel}
          className="mt-4 w-full rounded-lg border border-amber-300 bg-white py-2 text-xs font-semibold text-amber-700 hover:bg-amber-100 transition-colors"
        >
          Cancel
        </button>
      </div>
    </div>
  )
}
