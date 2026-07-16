import { useExecutiveDetail, useTrainExecutive } from '@/api/executives.api'
import { useCompany } from '@/api/company.api'
import { formatMoney, RARITY_COLORS } from '@/game/executives'
import { SkillGrid } from './ExecutiveCard'

interface ExecutiveDetailProps {
  executiveId: string | null
  onTrainingComplete?: () => void
}

const liveEffect: Record<string, { title: string; body: string }> = {
  coo: { title: 'COO · Management', body: 'The management skill is now recorded and shown for the administration system. Administration overhead is not yet charged in the minimal economy, so this role has no hidden payroll discount.' },
  cfo: { title: 'CFO · Accounting', body: 'The accounting skill is now recorded for future finance tools. Exchange fees remain the published 4%; this role does not secretly alter a completed trade.' },
  cmo: { title: 'CMO · Communication', body: 'When assigned as CMO, effective Communication adds up to +50% retail demand speed. It improves sell-through, not the price a customer accepts.' },
  cto: { title: 'CTO · Science', body: 'When assigned as CTO, effective Science gives +2% production speed per point, capped at a 3× production multiplier.' },
}

export function ExecutiveDetail({ executiveId, onTrainingComplete }: ExecutiveDetailProps) {
  const { data: detail, isLoading, isError, error } = useExecutiveDetail(executiveId)
  const train = useTrainExecutive()
  const { data: companyData } = useCompany()

  if (!executiveId) return <EmptyDetail />
  if (isLoading) return <Panel><p className="text-xs font-semibold text-amber-600 animate-pulse">Loading executive…</p></Panel>
  if (isError || !detail) return <Panel><p className="text-xs font-semibold text-red-700">{error instanceof Error ? error.message : 'Failed to load executive details'}</p></Panel>

  const cash = companyData?.authCompany?.money ?? 0
  const canDevelop = cash >= detail.trainingCost
  const effect = liveEffect[detail.specialty]

  const develop = async () => {
    try {
      await train.mutateAsync(detail.id)
      onTrainingComplete?.()
    } catch {
      // The mutation result is rendered below.
    }
  }

  return (
    <Panel>
      <div className="mb-4 border-b border-amber-200/60 pb-3">
        <div className={`mb-2 inline-block rounded-md border px-2 py-0.5 text-[10px] font-black uppercase tracking-wider ${RARITY_COLORS[detail.rarity]}`}>{detail.rarity}</div>
        <h3 className="text-lg font-black text-amber-950">{detail.name}</h3>
        <p className="text-xs font-semibold text-amber-700">{detail.title} · Level {detail.level}</p>
      </div>

      <div className="mb-4">
        <h4 className="mb-2 text-[10px] font-black uppercase tracking-wider text-amber-700">Skill profile</h4>
        <SkillGrid executive={detail} />
      </div>

      <div className="mb-4 rounded-xl border border-cyan-800/20 bg-cyan-50 p-3">
        <div className="text-xs font-black text-cyan-950">{effect.title}</div>
        <p className="mt-1 text-[11px] leading-5 text-cyan-900">{effect.body}</p>
        <p className="mt-2 text-[10px] font-bold text-cyan-800">Current chair: {detail.position ? detail.position.toUpperCase() : 'unassigned'}</p>
      </div>

      <div className="mb-3 rounded-lg border border-amber-200/50 bg-white/60 p-3 text-xs">
        <div className="flex justify-between"><span className="text-amber-700">Development cost</span><span className="font-black text-amber-950">${formatMoney(detail.trainingCost)}</span></div>
        <p className="mt-1 text-[10px] leading-4 text-amber-600">Development is immediate and server-charged: +1 to every skill and +3 more to this executive's specialty. There is no fake timer.</p>
      </div>

      <button onClick={develop} disabled={!canDevelop || train.isPending} className={`w-full rounded-lg py-3 text-sm font-black uppercase tracking-wider ${canDevelop ? 'bg-green-700 text-white hover:bg-green-800 disabled:bg-green-300' : 'bg-gray-300 text-gray-500'}`}>
        {train.isPending ? 'Developing…' : canDevelop ? 'Develop executive' : `Need $${formatMoney(detail.trainingCost - cash)} more`}
      </button>
      {train.isError && <p className="mt-2 text-center text-[10px] font-semibold text-red-700">{train.error instanceof Error ? train.error.message : 'Development failed'}</p>}
    </Panel>
  )
}

function EmptyDetail() {
  return <Panel><p className="text-xs text-amber-500">Select an executive to inspect their four skills and current leadership effect.</p></Panel>
}

function Panel({ children }: { children: React.ReactNode }) {
  return <div className="rounded-xl border-2 border-amber-300/50 bg-gradient-to-b from-amber-50 to-amber-100/30 p-5 shadow-sm">{children}</div>
}
