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
    return <section><h3 className="mb-3 text-base font-black uppercase tracking-wider text-amber-900">Leadership team</h3><div className="rounded-xl border border-amber-200/60 bg-white/50 py-8 text-center text-xs font-semibold text-amber-600">Loading executives…</div></section>
  }

  return (
    <section>
      <div className="mb-3 flex items-end justify-between gap-3">
        <div>
          <h3 className="text-base font-black uppercase tracking-wider text-amber-900">Leadership team ({executives.length})</h3>
          <p className="mt-1 text-[11px] text-amber-700">Each chair has one active holder. Reassigning a chair moves its previous holder to unassigned.</p>
        </div>
      </div>

      {executives.length === 0 && (
        <div className="rounded-xl border border-dashed border-amber-300/50 bg-white/40 py-8 text-center"><p className="text-xs text-amber-500">Recruit a candidate, then place them in the chair where their specialty matters.</p></div>
      )}

      <div className="space-y-2">
        {executives.map((executive) => {
          const selected = selectedId === executive.id
          return (
            <div key={executive.id} onClick={() => onSelect(executive.id)} className={`cursor-pointer rounded-xl border-2 p-3 transition ${selected ? 'border-amber-500 bg-amber-50' : 'border-amber-200/60 bg-white/60 hover:border-amber-300'}`}>
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2"><span className="font-bold text-amber-950">{executive.name}</span><span className="rounded bg-amber-200/60 px-1.5 py-0.5 text-[10px] font-bold text-amber-900">Lv. {executive.level}</span></div>
                  <div className="mt-0.5 text-[11px] font-semibold text-amber-700">{executive.title} · {executive.specialty.toUpperCase()} specialist</div>
                  <div className="mt-1 text-[10px] text-amber-600">M {Math.round(executive.skills.management)} · A {Math.round(executive.skills.accounting)} · C {Math.round(executive.skills.communication)} · S {Math.round(executive.skills.science)}</div>
                </div>
                <div className="flex shrink-0 items-center gap-2" onClick={(event) => event.stopPropagation()}>
                  <select value={executive.position} onChange={(event) => onAssign(executive.id, event.target.value as ExecutivePosition)} disabled={assigningId === executive.id} className="rounded-lg border border-cyan-800/25 bg-cyan-50 px-2 py-1.5 text-[11px] font-bold text-cyan-950 disabled:opacity-50">
                    {positions.map((position) => <option key={position.value || 'unassigned'} value={position.value}>{position.label}</option>)}
                  </select>
                  <button onClick={() => onTrain(executive.id)} className="rounded-lg bg-green-700 px-3 py-1.5 text-[11px] font-black text-white hover:bg-green-800">Develop · ${formatMoney(executive.trainingCost)}</button>
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </section>
  )
}
