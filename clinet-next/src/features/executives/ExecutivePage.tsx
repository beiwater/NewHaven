import { useState, useCallback } from 'react'
import { useMyExecutives, useTrainExecutive } from '@/api/hooks/executives.hooks'
import { ExecutiveMarket } from './ExecutiveMarket'
import { MyExecutives } from './MyExecutives'
import { TrainingQueue } from './TrainingQueue'
import { ExecutiveDetail } from './ExecutiveDetail'

export function ExecutivePage() {
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [completeNowId, setCompleteNowId] = useState<string | null>(null)

  const { data: myExecs, isLoading: myLoading, refetch: refetchMy } = useMyExecutives()
  const trainExec = useTrainExecutive()

  const ownedExecs = myExecs ?? []
  const trainees = ownedExecs.filter((e) => e.status === 'training')

  const handleTrain = useCallback(async (id: string) => {
    try {
      // TODO(audio): playSfx('executive_level_up')
      await trainExec.mutateAsync(id)
      refetchMy()
    } catch {
      // handled by mutation state
    }
  }, [trainExec, refetchMy])

  const handleCompleteNow = useCallback(async (id: string) => {
    setCompleteNowId(id)
    try {
      // TODO(audio): playSfx('executive_level_up')
      await trainExec.mutateAsync(id)
      refetchMy()
    } finally {
      setCompleteNowId(null)
    }
  }, [trainExec, refetchMy])

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
              selectedId={selectedId}
            />

            <TrainingQueue
              trainees={trainees}
              onCompleteNow={handleCompleteNow}
              completeNowPending={completeNowId}
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

        {/* Additive bonus hint (mobile) */}
        <div className="mt-6 rounded-lg border border-amber-200/30 bg-amber-100/30 px-4 py-2 text-center lg:hidden">
          <p className="text-[10px] text-amber-600">
            <span className="font-bold">Bonuses are additive:</span> Global reputation + executive bonus.
            Bonuses do not multiply.
          </p>
        </div>
      </div>
    </div>
  )
}
