import type { Executive } from '@/game/executives'
import { RARITY_COLORS, RARITY_BG, formatMoney } from '@/game/executives'
import { formatTrainingRemaining, useTrainingNow } from './trainingTimer'

interface ExecutiveCardProps {
  executive: Executive
  /** Show recruit button instead of details/train */
  mode: 'market' | 'owned'
  /** Whether recruit is pending */
  isPending?: boolean
  /** Whether the player can afford this executive */
  canAfford?: boolean
  onRecruit?: (id: string) => void
  onDetails?: (id: string) => void
  onTrain?: (id: string) => void
}

export function ExecutiveCard({
  executive,
  mode,
  isPending,
  canAfford = true,
  onRecruit,
  onDetails,
  onTrain,
}: ExecutiveCardProps) {
  const rarityColor = RARITY_COLORS[executive.rarity]
  const cardBg = RARITY_BG[executive.rarity]
  const isTraining = executive.status === 'training'
  const now = useTrainingNow(isTraining)
  const trainingLabel = formatTrainingRemaining(executive.trainingEndTime, now)

  return (
    <div className={`rounded-xl border-2 p-4 shadow-sm transition-shadow hover:shadow-md ${cardBg}`}>
      {/* Rarity tag */}
      <div className={`mb-2 inline-block rounded-md border px-2 py-0.5 text-[10px] font-black uppercase tracking-wider ${rarityColor}`}>
        {executive.rarity}
      </div>

      {/* Avatar placeholder + name/title */}
      <div className="flex items-center gap-3 mb-3">
        <div className={`flex h-12 w-12 shrink-0 items-center justify-center rounded-full text-lg font-black text-white
          ${executive.rarity === 'Legendary' ? 'bg-orange-500' :
            executive.rarity === 'Epic' ? 'bg-purple-500' :
            executive.rarity === 'Rare' ? 'bg-blue-500' : 'bg-gray-400'}`}
        >
          {executive.name.charAt(0)}
        </div>
        <div className="min-w-0">
          <div className="truncate text-sm font-bold text-amber-950">{executive.name}</div>
          <div className="text-[10px] font-semibold text-amber-700">{executive.title}</div>
        </div>
      </div>

      {/* Level + stage */}
      <div className="mb-3 flex items-center justify-between">
        <span className="rounded bg-amber-200/60 px-2 py-0.5 text-xs font-bold text-amber-900">
          Lv. {executive.level}
        </span>
        <span className="text-[11px] font-medium text-amber-700">{executive.stage}</span>
      </div>

      {/* Stats */}
      <div className="mb-3 space-y-1 border-t border-amber-200/50 pt-2">
        <StatRow label="Salary" value={`$${formatMoney(executive.salary)}/hr`} />
        <StatRow label="Production" value={`+${executive.productionBonus}%`} tone="up" />
        <StatRow label="Sales" value={`+${executive.salesBonus}%`} tone="up" />
        <StatRow label="Mgmt Discount" value={`${executive.mgmtDiscount}%`} tone="up" />
      </div>

      {/* Actions */}
      {mode === 'market' && (
        <button
          onClick={() => onRecruit?.(executive.id)}
          disabled={isPending || !canAfford}
          className={`w-full rounded-lg py-2 text-xs font-black uppercase tracking-wider transition-colors ${
            canAfford
              ? 'bg-green-700 text-white hover:bg-green-800 disabled:bg-green-300'
              : 'bg-gray-300 text-gray-500 cursor-not-allowed'
          }`}
        >
          {isPending ? 'Recruiting...' : canAfford ? `Recruit $${formatMoney(executive.recruitCost)}` : 'Not Enough Cash'}
        </button>
      )}

      {mode === 'owned' && (
        <div className="flex gap-2">
          <button
            onClick={() => onDetails?.(executive.id)}
            className="flex-1 rounded-lg border border-amber-300 bg-amber-50 py-2 text-xs font-bold text-amber-800 hover:bg-amber-100 transition-colors"
          >
            Details
          </button>
          <button
            onClick={() => onTrain?.(executive.id)}
            disabled={isPending || isTraining}
            className={`flex-1 rounded-lg px-2 py-2 text-xs font-black transition-colors whitespace-nowrap tabular-nums ${
              isTraining
                ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                : 'bg-green-700 text-white hover:bg-green-800'
            }`}
          >
            {isPending ? 'Training...' : isTraining ? `Training ${trainingLabel}` : 'Train'}
          </button>
        </div>
      )}
    </div>
  )
}

function StatRow({ label, value, tone }: { label: string; value: string; tone?: 'up' | 'down' }) {
  return (
    <div className="flex items-center justify-between text-xs">
      <span className="text-amber-700">{label}</span>
      <span className={`font-semibold ${tone === 'up' ? 'text-green-700' : tone === 'down' ? 'text-red-600' : 'text-amber-900'}`}>
        {value}
      </span>
    </div>
  )
}
