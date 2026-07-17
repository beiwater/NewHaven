import type { Executive, ExecutivePosition } from '@/game/executives'
import { formatMoney } from '@/game/executives'

interface MyExecutivesProps {
  executives: Executive[]
  isLoading: boolean
  onSelect: (id: string) => void
  onTrain: (id: string) => void
  onAssign: (id: string, position: ExecutivePosition) => void
  selectedId: string | null
  assigningId?: string | null
}

const positions: Array<{ value: ExecutivePosition; label: string }> = [
  { value: '', label: 'Unassigned' },
  { value: 'coo', label: 'COO · Operations' },
  { value: 'cfo', label: 'CFO · Finance' },
  { value: 'cmo', label: 'CMO · Marketing' },
  { value: 'cto', label: 'CTO · Technology' },
]

export function MyExecutives({ executives, isLoading, onSelect, onTrain, onAssign, selectedId, assigningId }: MyExecutivesProps) {
  if (isLoading) {
    return <section className="rounded-[24px] border border-amber-900/10 bg-amber-50/75 p-5"><h3 className="mb-3 text-base font-black text-amber-950">Talent bench</h3><div className="py-8 text-center text-xs font-semibold text-amber-600">Loading executives…</div></section>
  }

  return (
    <section className="rounded-[24px] border border-amber-900/10 bg-amber-50/75 p-4 shadow-sm md:p-5">
      <div className="mb-4 flex items-end justify-between gap-3">
        <div>
          <p className="text-[9px] font-black uppercase tracking-[0.28em] text-amber-700/60">Owned executives</p>
          <h3 className="mt-0.5 text-lg font-black text-amber-950">Talent bench <span className="text-amber-600">{executives.length}</span></h3>
        </div>
      </div>

      {executives.length === 0 && (
        <div className="rounded-2xl border border-dashed border-amber-400/50 bg-white/40 py-8 text-center"><p className="text-xs text-amber-600">Recruit a candidate, then appoint them to a chair.</p></div>
      )}

      <div className="grid gap-3 md:grid-cols-2">
        {executives.map((executive) => {
          const selected = selectedId === executive.id
          return (
            <div key={executive.id} onClick={() => onSelect(executive.id)} className={`cursor-pointer rounded-2xl border p-3 transition ${selected ? 'border-amber-500 bg-white shadow-md ring-2 ring-amber-300/40' : 'border-amber-900/10 bg-white/55 hover:-translate-y-0.5 hover:border-amber-400/60 hover:bg-white'}`}>
              <div className="flex items-start gap-3">
                <div className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-[#24352f] text-sm font-black text-amber-100">{executive.name.slice(0, 1)}</div>
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2"><span className="truncate font-black text-amber-950">{executive.name}</span><span className="rounded-full bg-amber-200/70 px-2 py-0.5 text-[9px] font-black text-amber-900">LV {executive.level}</span></div>
                  <div className="mt-0.5 text-[10px] font-bold uppercase tracking-wider text-amber-700/65">{executive.specialty} specialist · {executive.rarity}</div>
                  <div className="mt-2 grid grid-cols-4 gap-1 text-center text-[9px] text-amber-700">
                    {([['M', executive.skills.management], ['A', executive.skills.accounting], ['C', executive.skills.communication], ['S', executive.skills.science]] as const).map(([label, value]) => <span key={label} className="rounded-md bg-amber-100/70 py-1"><b className="text-amber-950">{label}</b> {Math.round(value)}</span>)}
                  </div>
                </div>
              </div>
              <div className="mt-3 flex items-center gap-2 border-t border-amber-900/10 pt-3" onClick={(event) => event.stopPropagation()}>
                <select value={executive.position} onChange={(event) => onAssign(executive.id, event.target.value as ExecutivePosition)} disabled={assigningId === executive.id} aria-label={`Assign ${executive.name} to leadership chair`} className="min-w-0 flex-1 rounded-xl border border-amber-900/15 bg-amber-50 px-2 py-2 text-[10px] font-black text-amber-950 outline-none focus:border-amber-500 disabled:opacity-50">
                  {positions.map((position) => <option key={position.value || 'unassigned'} value={position.value}>{position.label}</option>)}
                </select>
                <button onClick={() => onTrain(executive.id)} className="rounded-xl bg-[#2e6a51] px-3 py-2 text-[10px] font-black text-white hover:bg-[#245540]">Develop · ${formatMoney(executive.trainingCost)}</button>
              </div>
            </div>
          )
        })}
      </div>
    </section>
  )
}
