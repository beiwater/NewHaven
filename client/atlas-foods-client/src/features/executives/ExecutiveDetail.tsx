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
      <div className="-mx-5 -mt-5 mb-5 rounded-t-[24px] border-b border-white/10 bg-[#20312b] p-5 text-white">
        <div className="flex items-center gap-3">
          <div className="grid h-14 w-14 place-items-center rounded-2xl border border-amber-100/20 bg-gradient-to-br from-amber-200 to-amber-500 text-xl font-black text-amber-950">{detail.name.slice(0, 1)}</div>
          <div className="min-w-0">
            <div className={`mb-1 inline-block rounded-full border px-2 py-0.5 text-[9px] font-black uppercase tracking-wider ${RARITY_COLORS[detail.rarity]}`}>{detail.rarity}</div>
            <h3 className="truncate text-lg font-black">{detail.name}</h3>
            <p className="text-[10px] font-semibold text-amber-100/55">{detail.title} · Level {detail.level}</p>
          </div>
        </div>
      </div>

      <div className="mb-4">
        <h4 className="mb-2 text-[9px] font-black uppercase tracking-[0.25em] text-amber-700">Skill profile</h4>
        <SkillGrid executive={detail} />
      </div>

      <div className="mb-4 rounded-2xl border border-emerald-900/15 bg-[#e8f1e6] p-3">
        <div className="flex items-center justify-between gap-2"><div className="text-xs font-black text-emerald-950">{effect.title}</div><span className="rounded-full bg-emerald-950 px-2 py-1 text-[8px] font-black uppercase tracking-wider text-emerald-100">{detail.position ? 'Active chair' : 'Bench'}</span></div>
        <p className="mt-2 text-[10px] leading-5 text-emerald-900/75">{effect.body}</p>
      </div>

      <div className="mb-3 rounded-2xl border border-amber-900/10 bg-white/70 p-3 text-xs">
        <div className="flex justify-between"><span className="text-amber-700">Development cost</span><span className="font-black text-amber-950">${formatMoney(detail.trainingCost)}</span></div>
        <p className="mt-1 text-[10px] leading-4 text-amber-600">Development is immediate and server-charged: +1 to every skill and +3 more to this executive's specialty. There is no fake timer.</p>
      </div>

      <button onClick={develop} disabled={!canDevelop || train.isPending} className={`w-full rounded-2xl py-3 text-xs font-black uppercase tracking-[0.18em] transition ${canDevelop ? 'bg-[#2e6a51] text-white shadow-lg shadow-emerald-950/10 hover:-translate-y-0.5 hover:bg-[#245540] disabled:bg-green-300' : 'bg-gray-300 text-gray-500'}`}>
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
  return <div className="overflow-hidden rounded-[24px] border border-amber-950/10 bg-[#fffaf0] p-5 shadow-[0_18px_50px_rgba(89,57,24,0.14)]">{children}</div>
}
