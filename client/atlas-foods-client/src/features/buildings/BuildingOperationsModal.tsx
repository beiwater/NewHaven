import { useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import { useBuildings } from '@/api/buildings.api'
import { useCompany } from '@/api/company.api'
import { MAPS, isMapUnlocked, placeableSlots, type MapId } from '@/game/map.config'
import { useUIStore } from '@/store/ui.store'
import { BuildingCard } from './BuildingCard'

export function BuildingOperationsModal() {
  const dialogRef = useRef<HTMLElement>(null)
  const selectedBuildingId = useUIStore((state) => state.selectedBuildingId)
  const selectBuilding = useUIStore((state) => state.selectBuilding)
  const currentMapIdRaw = useUIStore((state) => state.currentMapId)
  const { data: buildingsData } = useBuildings()
  const { data: companyData } = useCompany()

  const level = companyData?.levelInfo?.level ?? 1
  const currentMapId: MapId = isMapUnlocked(currentMapIdRaw, level) ? currentMapIdRaw : 'harbor'
  const buildings = Array.isArray(buildingsData) ? buildingsData : []
  const selected = buildings.find((building) => building.id === selectedBuildingId) ?? null
  const placedOnMap = buildings.filter((building) => {
    if (building.placed === false) return false
    const mapId = building.mapId && MAPS[building.mapId] ? building.mapId : 'harbor'
    return mapId === currentMapId
  })
  const hasFreeSlots = placeableSlots(currentMapId, level).length > placedOnMap.length

  useEffect(() => {
    if (!selected) return
    const previouslyFocused = document.activeElement instanceof HTMLElement ? document.activeElement : null
    const previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        selectBuilding(null)
        return
      }
      if (event.key !== 'Tab') return
      const focusable = [...(dialogRef.current?.querySelectorAll<HTMLElement>(
        'button:not([disabled]), input:not([disabled]), select:not([disabled]), a[href]',
      ) ?? [])].filter((element) => element.offsetParent !== null)
      if (focusable.length === 0) return
      const first = focusable[0]
      const last = focusable[focusable.length - 1]
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault()
        last.focus()
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault()
        first.focus()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => {
      document.body.style.overflow = previousOverflow
      window.removeEventListener('keydown', handleKeyDown)
      previouslyFocused?.focus()
    }
  }, [selectBuilding, selected])

  if (!selected) return null

  return createPortal(
    <div
      className="building-modal-backdrop fixed inset-0 z-[90] flex items-center justify-center bg-amber-950/25 p-0 backdrop-blur-[2px] sm:p-4 lg:p-8"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) selectBuilding(null)
      }}
    >
      <section
        ref={dialogRef}
        aria-label={`${selected.name ?? 'Building'} operations`}
        aria-modal="true"
        role="dialog"
        className="building-modal-shell h-full w-full overflow-hidden border border-amber-300/70 bg-[#f8edd7] shadow-2xl sm:h-[min(860px,calc(100dvh-2rem))] sm:max-w-6xl sm:rounded-[28px]"
      >
        <BuildingCard
          building={selected}
          hasFreeSlots={hasFreeSlots}
          onClose={() => selectBuilding(null)}
        />
      </section>
    </div>,
    document.body,
  )
}
