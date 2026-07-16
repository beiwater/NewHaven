import { useState, useCallback } from 'react'
import { useAssignExecutivePosition, useMyExecutives, useTrainExecutive } from '@/api/executives.api'
import type { ExecutivePosition } from '@/game/executives'
import { ExecutiveMarket } from './ExecutiveMarket'
import { MyExecutives } from './MyExecutives'
import { ExecutiveDetail } from './ExecutiveDetail'
import { audio } from '@/audio/AudioManager'

export function ExecutivePage() {
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [assigningId, setAssigningId] = useState<string | null>(null)

  const { data: myExecs, isLoading: myLoading, refetch: refetchMy } = useMyExecutives()
  const trainExec = useTrainExecutive()
  const assignPosition = useAssignExecutivePosition()

  const ownedExecs = myExecs ?? []

  const handleTrain = useCallback(async (id: string) => {
    try {
      audio.playSfx('executive_level_up')
      await trainExec.mutateAsync(id)
      refetchMy()
    } catch {
      // handled by mutation state
    }
  }, [trainExec, refetchMy])

  const handleAssign = useCallback(async (id: string, position: ExecutivePosition) => {
    setAssigningId(id)
    try {
      await assignPosition.mutateAsync({ executiveId: id, position })
      refetchMy()
    } finally {
      setAssigningId(null)
    }
  }, [assignPosition, refetchMy])

  const handleTrainingComplete = useCallback(() => {
    refetchMy()
  }, [refetchMy])

  return (
    <div className="h-full overflow-y-auto">
      <div className="mx-auto max-w-7xl p-4 md:p-6">
        {/* Page header */}
        <div className="mb-5">
          <p className="text-[10px] font-bold uppercase tracking-[0.24em] text-amber-700/70">
            Human Resources
          </p>
          <h2 className="text-2xl font-black text-amber-950">Executives</h2>
        </div>

        {/* Main grid: left side (market + my execs + training) | right side (detail) */}
        <div className="grid gap-5 lg:grid-cols-[1fr_340px]">
          {/* Left column */}
          <div className="space-y-6">
            <ExecutiveMarket />

            <MyExecutives
              executives={ownedExecs}
              isLoading={myLoading}
              onSelect={setSelectedId}
              onTrain={handleTrain}
              onAssign={handleAssign}
              selectedId={selectedId}
              assigningId={assigningId}
            />
          </div>

          {/* Right column — detail panel */}
          <div className="lg:sticky lg:top-4 lg:self-start">
            <ExecutiveDetail
              executiveId={selectedId}
              onTrainingComplete={handleTrainingComplete}
            />
          </div>
        </div>

        {/* System hint (mobile) */}
        <div className="mt-6 rounded-lg border border-amber-200/30 bg-amber-100/30 px-4 py-2 text-center lg:hidden">
          <p className="text-[10px] text-amber-600">
            <span className="font-bold">Leadership is explicit:</span> only an assigned CMO or CTO changes the current retail or production loop.
          </p>
        </div>
      </div>
    </div>
  )
}
