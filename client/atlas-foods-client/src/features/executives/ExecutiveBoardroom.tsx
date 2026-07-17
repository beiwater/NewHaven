import type { Executive, ExecutivePosition } from '@/game/executives'

interface ExecutiveBoardroomProps {
  executives: Executive[]
  selectedId: string | null
  assigningId: string | null
  onSelect: (id: string) => void
}

const chairs: Array<{
  position: Exclude<ExecutivePosition, ''>
  shortName: string
  title: string
  skill: keyof Executive['skills']
  effect: string
  accent: string
  glow: string
}> = [
  { position: 'coo', shortName: 'COO', title: 'Operations', skill: 'management', effect: 'Administration system reserved', accent: 'text-emerald-200', glow: 'border-emerald-400/40 bg-emerald-400/10' },
  { position: 'cfo', shortName: 'CFO', title: 'Finance', skill: 'accounting', effect: 'Finance system reserved', accent: 'text-sky-200', glow: 'border-sky-400/40 bg-sky-400/10' },
  { position: 'cmo', shortName: 'CMO', title: 'Marketing', skill: 'communication', effect: 'Retail demand +0.5% / effective point', accent: 'text-rose-200', glow: 'border-rose-400/40 bg-rose-400/10' },
  { position: 'cto', shortName: 'CTO', title: 'Technology', skill: 'science', effect: 'Production speed +2% / effective point', accent: 'text-violet-200', glow: 'border-violet-400/40 bg-violet-400/10' },
]

function effectiveSkill(raw: number): number {
  if (raw <= 60) return Math.max(0, raw)
  if (raw <= 80) return 60 + (raw - 60) / 2
  return 70 + (raw - 80) / 2
}

export function ExecutiveBoardroom({ executives, selectedId, assigningId, onSelect }: ExecutiveBoardroomProps) {
  const cto = executives.find((executive) => executive.position === 'cto')
  const cmo = executives.find((executive) => executive.position === 'cmo')
  const productionBonus = cto ? Math.min(200, effectiveSkill(cto.skills.science) * 2) : 0
  const retailBonus = cmo ? Math.min(50, effectiveSkill(cmo.skills.communication) / 2) : 0
  const activeSeats = executives.filter((executive) => executive.position).length
  return (
    <section className="overflow-hidden rounded-[28px] border border-amber-200/20 bg-[#16221f] shadow-[0_22px_70px_rgba(63,39,16,0.22)]">
      <div className="relative border-b border-white/10 px-5 py-5 md:px-7">
        <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_right,rgba(251,191,36,0.15),transparent_42%)]" />
        <div className="relative flex flex-wrap items-end justify-between gap-3">
          <div>
            <p className="text-[10px] font-black uppercase tracking-[0.32em] text-amber-300/70">Company command</p>
            <h3 className="mt-1 text-xl font-black text-amber-50">The Boardroom</h3>
          </div>
          <p className="max-w-md text-right text-[11px] leading-5 text-amber-100/55">
            One active holder per chair. Moving an executive here atomically releases the previous holder.
          </p>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-px border-b border-white/10 bg-white/10 sm:grid-cols-4">
        {[
          ['Active cabinet', `${activeSeats} / 4`],
          ['Production speed', `+${productionBonus.toFixed(1)}%`],
          ['Retail demand', `+${retailBonus.toFixed(1)}%`],
          ['Admin & finance', 'Reserved'],
        ].map(([label, value]) => (
          <div key={label} className="bg-[#1b2a26] px-4 py-3">
            <div className="text-[8px] font-black uppercase tracking-[0.18em] text-amber-200/40">{label}</div>
            <div className="mt-1 text-sm font-black text-amber-50">{value}</div>
          </div>
        ))}
      </div>

      <div className="grid gap-3 p-4 sm:grid-cols-2 md:p-6">
        {chairs.map((chair) => {
          const holder = executives.find((executive) => executive.position === chair.position)
          const skillValue = holder ? Math.round(holder.skills[chair.skill]) : 0
          const selected = holder?.id === selectedId
          return (
            <button
              key={chair.position}
              type="button"
              onClick={() => holder && onSelect(holder.id)}
              className={`group min-h-40 rounded-2xl border p-4 text-left transition duration-200 ${chair.glow} ${selected ? 'ring-2 ring-amber-300 ring-offset-2 ring-offset-[#16221f]' : 'hover:-translate-y-0.5 hover:border-amber-200/55'}`}
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <div className={`text-[11px] font-black uppercase tracking-[0.25em] ${chair.accent}`}>{chair.shortName}</div>
                  <div className="mt-0.5 text-xs font-semibold text-white/55">{chair.title}</div>
                </div>
                <span className="rounded-full border border-white/10 bg-black/15 px-2 py-1 text-[9px] font-bold uppercase tracking-wider text-white/45">
                  {chair.skill}
                </span>
              </div>

              {holder ? (
                <div className="mt-5 flex items-center gap-3">
                  <div className="grid h-12 w-12 shrink-0 place-items-center rounded-2xl border border-amber-100/20 bg-gradient-to-br from-amber-200 to-amber-500 text-lg font-black text-amber-950 shadow-lg">
                    {holder.name.slice(0, 1)}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-sm font-black text-white">{holder.name}</div>
                    <div className="mt-0.5 text-[10px] font-semibold text-white/45">Level {holder.level} · {holder.rarity}</div>
                    <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-black/25">
                      <div className="h-full rounded-full bg-amber-300" style={{ width: `${Math.min(100, skillValue)}%` }} />
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-xl font-black text-amber-100">{skillValue}</div>
                    <div className="text-[8px] font-black uppercase tracking-wider text-white/35">skill</div>
                  </div>
                </div>
              ) : (
                <div className="mt-5 rounded-xl border border-dashed border-white/15 bg-black/10 px-3 py-4 text-center">
                  <div className="text-xs font-black text-white/55">Chair vacant</div>
                  <div className="mt-1 text-[10px] text-white/30">Assign someone from the talent bench</div>
                </div>
              )}

              <div className="mt-4 border-t border-white/10 pt-3 text-[9px] font-semibold leading-4 text-white/40">{chair.effect}</div>
            </button>
          )
        })}
      </div>

      {executives.some((executive) => !executive.position) && (
        <div className="border-t border-white/10 bg-black/10 px-5 py-3 text-[10px] text-amber-100/45 md:px-7">
          Vacant chairs can be filled from the role selector below.
          {assigningId && <span className="ml-2 font-bold text-amber-300">Updating appointment…</span>}
        </div>
      )}
    </section>
  )
}
