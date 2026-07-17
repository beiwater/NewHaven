import { useState, useCallback } from 'react'
import { useAssignExecutivePosition, useMyExecutives, useTrainExecutive } from '@/api/executives.api'
import type { ExecutivePosition } from '@/game/executives'
import { ExecutiveMarket } from './ExecutiveMarket'
import { MyExecutives } from './MyExecutives'
import { ExecutiveDetail } from './ExecutiveDetail'
import { ExecutiveBoardroom } from './ExecutiveBoardroom'
import { audio } from '@/audio/AudioManager'

export function ExecutivePage() {
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [assigningId, setAssigningId] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'boardroom' | 'market'>('boardroom')

  const { data: myExecs, isLoading: myLoading, refetch: refetchMy } = useMyExecutives()
  const trainExec = useTrainExecutive()
  const assignPosition = useAssignExecutivePosition()

  const ownedExecs = myExecs ?? []
  const resolvedSelectedId = selectedId ?? ownedExecs.find((executive) => executive.position)?.id ?? ownedExecs[0]?.id ?? null

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
    <div className="h-full overflow-y-auto bg-[radial-gradient(circle_at_top,#fff8df_0%,#f6e7bd_42%,#ead09a_100%)]">
      <div className="mx-auto max-w-[1440px] p-3 md:p-6">
        <header className="mb-5 overflow-hidden rounded-[26px] border border-amber-950/10 bg-amber-50/80 shadow-sm backdrop-blur">
          <div className="flex flex-col gap-4 px-5 py-5 md:flex-row md:items-end md:justify-between md:px-7">
            <div>
              <p className="text-[10px] font-black uppercase tracking-[0.32em] text-amber-700/65">New Haven Commerce Guild</p>
              <h2 className="mt-1 text-3xl font-black tracking-tight text-amber-950">Leadership HQ</h2>
              <p className="mt-1 max-w-xl text-xs leading-5 text-amber-800/65">Build a cabinet around the way your company earns: production, retail, operations and finance.</p>
            </div>
            <div className="flex gap-2 rounded-2xl border border-amber-900/10 bg-white/55 p-1.5">
              <button type="button" onClick={() => setActiveTab('boardroom')} className={`rounded-xl px-4 py-2 text-[11px] font-black uppercase tracking-wider transition ${activeTab === 'boardroom' ? 'bg-[#16221f] text-amber-100 shadow' : 'text-amber-800 hover:bg-amber-100'}`}>Boardroom</button>
              <button type="button" onClick={() => setActiveTab('market')} className={`rounded-xl px-4 py-2 text-[11px] font-black uppercase tracking-wider transition ${activeTab === 'market' ? 'bg-[#16221f] text-amber-100 shadow' : 'text-amber-800 hover:bg-amber-100'}`}>Talent market</button>
            </div>
          </div>
        </header>

        {activeTab === 'boardroom' ? (
          <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_360px]">
            <div className="space-y-5">
              <ExecutiveBoardroom executives={ownedExecs} selectedId={resolvedSelectedId} assigningId={assigningId} onSelect={setSelectedId} />
              <MyExecutives executives={ownedExecs} isLoading={myLoading} onSelect={setSelectedId} onTrain={handleTrain} onAssign={handleAssign} selectedId={resolvedSelectedId} assigningId={assigningId} />
            </div>
            <div className="xl:sticky xl:top-4 xl:self-start">
              <ExecutiveDetail executiveId={resolvedSelectedId} onTrainingComplete={handleTrainingComplete} />
            </div>
          </div>
        ) : (
          <ExecutiveMarket />
        )}
      </div>
    </div>
  )
}
