import { useExecutiveDetail, useTrainExecutive } from '@/api/executives.api'
import { audio } from '@/audio/AudioManager'
import { useCompany } from '@/api/company.api'
import {
  productionBonusAtLevel,
  salesBonusAtLevel,
  mgmtDiscountAtLevel,
  trainingCost,
  trainingTimeSeconds,
  formatMoney,
  formatDuration,
  RARITY_COLORS,
} from '@/game/executives'
import { formatTrainingRemaining, useTrainingNow } from './trainingTimer'

interface ExecutiveDetailProps {
  executiveId: string | null
  /** Called after a training action succeeds */
  onTrainingComplete?: () => void
}

export function ExecutiveDetail({ executiveId, onTrainingComplete }: ExecutiveDetailProps) {
  const { data: detail, isLoading, isError, error } = useExecutiveDetail(executiveId)
  const trainExec = useTrainExecutive()
  const { data: companyData } = useCompany()
  const now = useTrainingNow(detail?.status === 'training')

  const cash = companyData?.authCompany?.money ?? 0

  if (!executiveId) {
    return (
      <div className="rounded-xl border-2 border-amber-200/40 bg-white/50 p-6 text-center">
        <svg className="mx-auto h-10 w-10 text-amber-300" fill="none" stroke="currentColor" strokeWidth={1.5} viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
        </svg>
        <p className="mt-2 text-xs text-amber-500">Select an executive to view details</p>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="rounded-xl border-2 border-amber-200/40 bg-white/50 p-6 text-center">
        <div className="text-xs font-semibold text-amber-600 animate-pulse">Loading details...</div>
      </div>
    )
  }

  if (isError || !detail) {
    return (
      <div className="rounded-xl border-2 border-red-200 bg-red-50 p-4 text-center">
        <p className="text-xs font-semibold text-red-700">
          Error: {error instanceof Error ? error.message : 'Failed to load executive details'}
        </p>
      </div>
    )
  }

  const rarityColor = RARITY_COLORS[detail.rarity]
  const nextProd = productionBonusAtLevel(detail.level + 1)
  const nextSales = salesBonusAtLevel(detail.level + 1)
  const nextMgmt = mgmtDiscountAtLevel(detail.level + 1)
  const cost = trainingCost(detail.level)
  const time = trainingTimeSeconds(detail.level)
  const canTrain = cash >= cost && detail.status !== 'training'
  const trainingLabel = formatTrainingRemaining(detail.trainingEndTime, now)

  const handleTrain = async () => {
    try {
      audio.playSfx('executive_level_up')
      await trainExec.mutateAsync(detail.id)
      onTrainingComplete?.()
    } catch {
      // Error handled by mutation state
    }
  }

  return (
    <div className="rounded-xl border-2 border-amber-300/50 bg-gradient-to-b from-amber-50 to-amber-100/30 p-5 shadow-sm">
      {/* Decorative header */}
      <div className="mb-4 flex items-center gap-2 border-b border-amber-200/60 pb-3">
        <div className="flex h-5 w-5 items-center justify-center rounded bg-amber-700/10">
          <svg className="h-3 w-3 text-amber-800" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
          </svg>
        </div>
        <h3 className="text-sm font-black uppercase tracking-wider text-amber-900">Executive Details</h3>
      </div>

      {/* Avatar + basic info */}
      <div className="mb-4 flex items-center gap-3">
        <div className={`flex h-14 w-14 shrink-0 items-center justify-center rounded-full text-xl font-black text-white
          ${detail.rarity === 'Legendary' ? 'bg-orange-500' :
            detail.rarity === 'Epic' ? 'bg-purple-500' :
            detail.rarity === 'Rare' ? 'bg-blue-500' : 'bg-gray-400'}`}
        >
          {detail.name.charAt(0)}
        </div>
        <div className="min-w-0">
          <div className={`mb-1 inline-block rounded-md border px-2 py-0.5 text-[10px] font-black uppercase tracking-wider ${rarityColor}`}>
            {detail.rarity}
          </div>
          <div className="truncate text-lg font-bold text-amber-950">{detail.name}</div>
          <div className="text-xs font-semibold text-amber-700">{detail.title}</div>
          <div className="flex items-center gap-2 mt-0.5">
            <span className="rounded bg-amber-200/60 px-1.5 py-0.5 text-xs font-bold text-amber-900">
              Lv. {detail.level}
            </span>
            <span className="text-xs text-amber-700">{detail.stage}</span>
          </div>
        </div>
      </div>

      {/* Salary */}
      <div className="mb-4 flex items-center justify-between rounded-lg border border-amber-200/50 bg-white/60 px-3 py-2">
        <span className="text-xs font-bold text-amber-700">Salary</span>
        <span className="text-sm font-bold text-amber-950">${formatMoney(detail.salary)} /hr</span>
      </div>

      {/* Current Effects */}
      <div className="mb-4">
        <h4 className="mb-2 text-[10px] font-black uppercase tracking-wider text-amber-700">
          Current Effects
        </h4>
        <div className="space-y-1.5">
          <EffectRow
            icon={
              <svg className="h-3.5 w-3.5" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 1.5H8.25A2.25 2.25 0 006 3.75v16.5a2.25 2.25 0 002.25 2.25h7.5A2.25 2.25 0 0018 20.25V3.75a2.25 2.25 0 00-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 18.75h3" />
              </svg>
            }
            label="Production Bonus"
            value={`+${detail.productionBonus}%`}
          />
          <EffectRow
            icon={
              <svg className="h-3.5 w-3.5" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
            label="Sales Bonus"
            value={`+${detail.salesBonus}%`}
          />
          <EffectRow
            icon={
              <svg className="h-3.5 w-3.5" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M9.568 3H5.25A2.25 2.25 0 003 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 005.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 009.568 3z" />
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 6h.008v.008H6V6z" />
              </svg>
            }
            label="Management Discount"
            value={`${detail.mgmtDiscount}%`}
          />
        </div>
      </div>

      {/* Next Level Preview */}
      <div className="mb-4">
        <h4 className="mb-2 text-[10px] font-black uppercase tracking-wider text-green-700">
          Next Level Preview (Lv. {detail.level + 1})
        </h4>
        <div className="space-y-1">
          <PreviewRow label="Production Bonus" current={`+${detail.productionBonus}%`} next={`+${nextProd}%`} />
          <PreviewRow label="Sales Bonus" current={`+${detail.salesBonus}%`} next={`+${nextSales}%`} />
          <PreviewRow label="Mgmt Discount" current={`${detail.mgmtDiscount}%`} next={`${nextMgmt}%`} />
        </div>
      </div>

      {/* Training Cost */}
      <div className="mb-4 rounded-lg border border-amber-200/50 bg-white/60 p-3">
        <h4 className="mb-2 text-[10px] font-black uppercase tracking-wider text-amber-700">
          Training Cost
        </h4>
        <div className="flex items-center justify-between text-xs">
          <span className="text-amber-700">Cash</span>
          <span className="font-bold text-amber-950">${formatMoney(cost)}</span>
        </div>
        <div className="flex items-center justify-between text-xs mt-1">
          <span className="text-amber-700">Duration</span>
          <span className="font-bold text-amber-950">{formatDuration(time)}</span>
        </div>
      </div>

      {/* Train button */}
      <button
        onClick={handleTrain}
        disabled={!canTrain || trainExec.isPending}
        className={`w-full rounded-lg py-3 text-sm font-black uppercase tracking-wider transition-colors ${
          canTrain
            ? 'bg-green-700 text-white hover:bg-green-800 disabled:bg-green-300'
            : 'bg-gray-300 text-gray-500 cursor-not-allowed'
        }`}
      >
        {trainExec.isPending
          ? 'Training...'
          : detail.status === 'training'
            ? `Training ${trainingLabel}`
            : !canTrain
              ? `Need $${formatMoney(cost - cash)} more`
              : 'Train Executive'}
      </button>

      {/* Training feedback */}
      {trainExec.isSuccess && (
        <div className="mt-3 rounded-lg border border-green-200 bg-green-50 px-3 py-2 text-center text-[10px] font-semibold text-green-700">
          {detail.name} has been trained to Lv. {detail.level + 1}!
        </div>
      )}
      {trainExec.isError && (
        <div className="mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-center text-[10px] font-semibold text-red-700">
          Training failed: {trainExec.error instanceof Error ? trainExec.error.message : 'Unknown error'}
        </div>
      )}

      {/* Additive note */}
      <div className="mt-4 rounded-lg border border-amber-200/30 bg-amber-100/40 px-3 py-2">
        <p className="text-[10px] leading-relaxed text-amber-700">
          <span className="font-bold">Bonuses are additive:</span> Global reputation + executive bonus. Bonuses do not multiply.
        </p>
      </div>
    </div>
  )
}

function EffectRow({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode
  label: string
  value: string
}) {
  return (
    <div className="flex items-center justify-between rounded-lg border border-amber-200/40 bg-white/60 px-3 py-2 text-xs">
      <div className="flex items-center gap-2">
        <span className="text-amber-600">{icon}</span>
        <span className="text-amber-700">{label}</span>
      </div>
      <span className="font-bold text-green-700">{value}</span>
    </div>
  )
}

function PreviewRow({ label, current, next }: { label: string; current: string; next: string }) {
  return (
    <div className="flex items-center justify-between rounded border border-green-200/40 bg-green-50/60 px-3 py-1.5 text-xs">
      <span className="text-amber-700">{label}</span>
      <div className="flex items-center gap-2">
        <span className="text-amber-600">{current}</span>
        <svg className="h-3 w-3 text-green-600" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" d="M5 15l7-7 7 7" />
        </svg>
        <span className="font-bold text-green-700">{next}</span>
      </div>
    </div>
  )
}
