import type { Executive } from '@/game/executives'
import { productionBonusAtLevel, salesBonusAtLevel, mgmtDiscountAtLevel, formatMoney } from '@/game/executives'
import { formatTrainingRemaining, useTrainingNow } from './trainingTimer'

interface MyExecutivesProps {
  executives: Executive[]
  isLoading: boolean
  onSelect: (id: string) => void
  onTrain: (id: string) => void
  selectedId: string | null
}

export function MyExecutives({
  executives,
  isLoading,
  onSelect,
  onTrain,
  selectedId,
}: MyExecutivesProps) {
  const hasTraining = executives.some((exec) => exec.status === 'training')
  const now = useTrainingNow(hasTraining)

  if (isLoading) {
    return (
      <section>
        <h3 className="mb-3 text-base font-black uppercase tracking-wider text-amber-900">
          My Executives
        </h3>
        <div className="flex items-center justify-center rounded-xl border border-amber-200/60 bg-white/50 py-8">
          <div className="text-xs font-semibold text-amber-600 animate-pulse">Loading executives...</div>
        </div>
      </section>
    )
  }

  return (
    <section>
      <div className="mb-3 flex items-center gap-2">
        <svg className="h-5 w-5 text-amber-800" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
        <h3 className="text-base font-black uppercase tracking-wider text-amber-900">
          My Executives ({executives.length})
        </h3>
      </div>

      {executives.length === 0 && (
        <div className="rounded-xl border border-dashed border-amber-300/50 bg-white/40 py-8 text-center">
          <p className="text-xs text-amber-500">You haven't recruited any executives yet. Check the market above!</p>
        </div>
      )}

      <div className="space-y-2">
        {executives.map((exec) => {
          const nextProd = productionBonusAtLevel(exec.level + 1)
          const nextSales = salesBonusAtLevel(exec.level + 1)
          const nextMgmt = mgmtDiscountAtLevel(exec.level + 1)
          const isSelected = selectedId === exec.id
          const isTraining = exec.status === 'training'
          const trainingLabel = formatTrainingRemaining(exec.trainingEndTime, now)

          return (
            <div
              key={exec.id}
              onClick={() => onSelect(exec.id)}
              className={`cursor-pointer rounded-xl border-2 p-3 transition-all hover:shadow-sm ${
                isSelected
                  ? 'border-amber-500 bg-amber-50 shadow-inner'
                  : 'border-amber-200/60 bg-white/60 hover:border-amber-300'
              }`}
            >
              <div className="flex items-start justify-between gap-3">
                {/* Left: avatar + name */}
                <div className="flex items-center gap-3 min-w-0">
                  <div className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-base font-black text-white
                    ${exec.rarity === 'Legendary' ? 'bg-orange-500' :
                      exec.rarity === 'Epic' ? 'bg-purple-500' :
                      exec.rarity === 'Rare' ? 'bg-blue-500' : 'bg-gray-400'}`}
                  >
                    {exec.name.charAt(0)}
                  </div>
                  <div className="min-w-0">
                    <div className="truncate text-sm font-bold text-amber-950">{exec.name}</div>
                    <div className="text-[10px] font-medium text-amber-700">{exec.title}</div>
                    <div className="flex items-center gap-2 mt-0.5">
                      <span className="rounded bg-amber-200/60 px-1.5 py-0.5 text-[10px] font-bold text-amber-900">
                        Lv. {exec.level}
                      </span>
                      <span className="text-[10px] text-amber-600">{exec.stage}</span>
                      <span className="text-[10px] text-amber-500">${formatMoney(exec.salary)}/hr</span>
                    </div>
                  </div>
                </div>

                {/* Right: stat comparison */}
                <div className="shrink-0 text-right">
                  <div className="grid grid-cols-[1fr_auto_1fr] gap-x-2 gap-y-0.5 text-[11px]">
                    {/* Header */}
                    <span className="text-amber-500 text-[10px]">Current</span>
                    <span />
                    <span className="text-green-700 text-[10px]">Next Lv</span>

                    {/* Production Bonus */}
                    <span className="text-amber-800 font-semibold">+{exec.productionBonus}%</span>
                    <ArrowUp />
                    <span className="text-green-700 font-semibold">+{nextProd}%</span>

                    {/* Sales Bonus */}
                    <span className="text-amber-800 font-semibold">+{exec.salesBonus}%</span>
                    <ArrowUp />
                    <span className="text-green-700 font-semibold">+{nextSales}%</span>

                    {/* Mgmt Discount */}
                    <span className="text-amber-800 font-semibold">{exec.mgmtDiscount}%</span>
                    <ArrowUp />
                    <span className="text-green-700 font-semibold">{nextMgmt}%</span>
                  </div>

                  {/* Actions */}
                  <div className="mt-2 flex gap-1.5 justify-end">
                    <button
                      onClick={(e) => { e.stopPropagation(); onSelect(exec.id) }}
                      className="rounded-md border border-amber-300 bg-amber-50 px-2.5 py-1 text-[10px] font-bold text-amber-800 hover:bg-amber-100 transition-colors"
                    >
                      Details
                    </button>
                    <button
                      onClick={(e) => { e.stopPropagation(); onTrain(exec.id) }}
                      disabled={isTraining}
                      className={`rounded-md px-2.5 py-1 text-[10px] font-bold transition-colors whitespace-nowrap tabular-nums ${
                        isTraining
                          ? 'bg-gray-200 text-gray-500 cursor-not-allowed'
                          : 'bg-green-700 text-white hover:bg-green-800'
                      }`}
                    >
                      {isTraining ? `Training ${trainingLabel}` : 'Train'}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </section>
  )
}

function ArrowUp() {
  return (
    <svg className="h-3 w-3 text-green-600" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" d="M5 15l7-7 7 7" />
    </svg>
  )
}
