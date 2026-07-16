import type { Executive } from '@/game/executives'
import { RARITY_COLORS, RARITY_BG, formatMoney } from '@/game/executives'
import { useTranslation } from 'react-i18next'

interface ExecutiveCardProps {
  executive: Executive
  mode: 'market' | 'owned'
  isPending?: boolean
  canAfford?: boolean
  onRecruit?: (id: string) => void
  onDetails?: (id: string) => void
  onTrain?: (id: string) => void
}

const roleName: Record<Executive['specialty'], string> = {
  coo: 'COO · Operations',
  cfo: 'CFO · Finance',
  cmo: 'CMO · Marketing',
  cto: 'CTO · Technology',
}

export function ExecutiveCard({ executive, mode, isPending, canAfford = true, onRecruit, onDetails, onTrain }: ExecutiveCardProps) {
  const { t } = useTranslation()
  const rarityColor = RARITY_COLORS[executive.rarity]
  const cardBg = RARITY_BG[executive.rarity]

  return (
    <div className={`group relative overflow-hidden rounded-[22px] border p-4 shadow-sm transition duration-200 hover:-translate-y-1 hover:shadow-xl ${cardBg}`}>
      <div className="pointer-events-none absolute -right-8 -top-8 h-24 w-24 rounded-full bg-white/45 blur-2xl" />
      <div className="mb-2 flex items-center justify-between gap-2">
        <div className={`inline-block rounded-md border px-2 py-0.5 text-[10px] font-black uppercase tracking-wider ${rarityColor}`}>
          {t('executives.rarity_' + executive.rarity.toLowerCase())}
        </div>
        <span className="rounded bg-cyan-950/10 px-2 py-0.5 text-[10px] font-black uppercase tracking-wide text-cyan-900">
          {roleName[executive.specialty]}
        </span>
      </div>

      <div className="mb-3 flex items-center gap-3">
        <div className={`flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl border border-white/70 text-lg font-black text-white shadow-md ${
          executive.rarity === 'Legendary' ? 'bg-orange-500' : executive.rarity === 'Epic' ? 'bg-purple-500' : executive.rarity === 'Rare' ? 'bg-blue-500' : 'bg-gray-400'
        }`}>
          {executive.name.charAt(0)}
        </div>
        <div className="min-w-0">
          <div className="truncate text-sm font-bold text-amber-950">{executive.name}</div>
          <div className="text-[10px] font-semibold text-amber-700">{executive.title}</div>
          <span className="mt-1 inline-block rounded bg-amber-200/60 px-1.5 py-0.5 text-[10px] font-bold text-amber-900">Lv. {executive.level}</span>
        </div>
      </div>

      <SkillGrid executive={executive} />

      {mode === 'market' && (
        <button
          onClick={() => onRecruit?.(executive.id)}
          disabled={isPending || !canAfford}
          className={`mt-3 w-full rounded-xl py-2.5 text-[11px] font-black uppercase tracking-wider transition ${canAfford ? 'bg-[#2e6a51] text-white shadow hover:bg-[#245540] disabled:bg-green-300' : 'cursor-not-allowed bg-gray-300 text-gray-500'}`}
        >
          {isPending ? t('executives.recruiting') : canAfford ? `${t('executives.hire')} $${formatMoney(executive.recruitCost)}` : t('executives.notEnoughCash')}
        </button>
      )}

      {mode === 'owned' && (
        <div className="mt-3 flex gap-2">
          <button onClick={() => onDetails?.(executive.id)} className="flex-1 rounded-lg border border-amber-300 bg-amber-50 py-2 text-xs font-bold text-amber-800 hover:bg-amber-100">Details</button>
          <button onClick={() => onTrain?.(executive.id)} disabled={isPending} className="flex-1 rounded-lg bg-green-700 py-2 text-xs font-black text-white hover:bg-green-800 disabled:bg-green-300">
            {isPending ? 'Developing…' : 'Develop'}
          </button>
        </div>
      )}
    </div>
  )
}

export function SkillGrid({ executive }: { executive: Pick<Executive, 'skills'> }) {
  const skills = [
    ['Management', 'MGT', executive.skills.management],
    ['Accounting', 'ACC', executive.skills.accounting],
    ['Communication', 'COM', executive.skills.communication],
    ['Science', 'SCI', executive.skills.science],
  ] as const
  return (
    <div className="grid grid-cols-2 gap-2 border-t border-amber-900/10 pt-3 text-[10px]">
      {skills.map(([label, shortLabel, value]) => (
        <div key={label} title={label} className="rounded-lg bg-white/60 px-2 py-1.5 text-amber-700">
          <div className="flex justify-between"><span className="text-[8px] font-black tracking-wider">{shortLabel}</span><span className="font-black text-amber-950">{Math.round(value)}</span></div>
          <div className="mt-1 h-1 overflow-hidden rounded-full bg-amber-900/10"><div className="h-full rounded-full bg-amber-500" style={{ width: `${Math.min(100, Math.max(0, value))}%` }} /></div>
        </div>
      ))}
    </div>
  )
}
