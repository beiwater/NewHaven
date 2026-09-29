import { useTranslation } from 'react-i18next'
import type { QualityResearch } from '@/api/research.api'
import { MAX_PRODUCT_QUALITY } from '@/game/quality'
import { resourceIcon, resourceName } from '@/game/resources'

function fmtCash(value: number): string {
  return `$${Math.round(value).toLocaleString()}`
}

export function QualityResearchCard({
  item,
  cash,
  pending,
  onUnlock,
}: {
  item: QualityResearch
  cash: number
  pending: boolean
  onUnlock: () => void
}) {
  const { t } = useTranslation()
  const maxed = item.maxQuality >= MAX_PRODUCT_QUALITY
  const nextQuality = item.nextQuality ?? item.maxQuality + 1
  const nextCost = item.nextCost ?? 0
  const canAfford = cash >= nextCost

  return (
    <article className="flex min-h-72 flex-col rounded-3xl border border-amber-200/90 bg-white/80 p-4 shadow-[0_12px_35px_rgba(109,76,25,0.08)]">
      <div className="flex items-start gap-3">
        <div className="grid h-16 w-16 shrink-0 place-items-center rounded-2xl bg-gradient-to-br from-amber-50 to-amber-200/80">
          <img src={resourceIcon(item.resourceId)} alt="" className="h-12 w-12 object-contain" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="truncate text-base font-black text-amber-950">{resourceName(item.resourceId)}</h3>
            <span className="rounded-full bg-cyan-100 px-2 py-0.5 text-[9px] font-black uppercase tracking-wider text-cyan-800">
              {t('research.productTier', { tier: item.tier })}
            </span>
          </div>
          <div className="mt-1 flex items-baseline gap-2">
            <span className="text-2xl font-black text-violet-700">Q{item.maxQuality}</span>
            <span className="text-[10px] font-bold text-green-700">+{item.salesSpeedBonus}% {t('quality.retailSpeed')}</span>
          </div>
        </div>
      </div>

      <div className="mt-4">
        <div className="mb-2 flex items-center justify-between text-[9px] font-black uppercase tracking-[0.16em] text-amber-600">
          <span>{t('research.qualityLicence')}</span>
          <span>Q0–Q{item.maxQuality}</span>
        </div>
        <div className="grid grid-cols-[repeat(13,minmax(0,1fr))] gap-1" aria-label={t('research.unlockedThrough', { quality: item.maxQuality })}>
          {Array.from({ length: MAX_PRODUCT_QUALITY + 1 }, (_, quality) => {
            const unlocked = quality <= item.maxQuality
            const next = quality === nextQuality && !maxed
            return (
              <div
                key={quality}
                title={`Q${quality}`}
                className={`grid h-7 place-items-center rounded-md text-[8px] font-black ${unlocked ? 'bg-violet-600 text-white' : next ? 'border border-dashed border-cyan-500 bg-cyan-50 text-cyan-800' : 'bg-stone-100 text-stone-400'}`}
              >
                {quality}
              </div>
            )
          })}
        </div>
      </div>

      <div className="mt-4 flex-1 rounded-2xl border border-amber-100 bg-amber-50/75 p-3">
        {maxed ? (
          <div className="flex h-full flex-col justify-center text-center">
            <div className="text-sm font-black text-green-700">{t('research.mastered')}</div>
            <p className="mt-1 text-[10px] font-semibold leading-4 text-green-700/80">{t('research.masteredHelp')}</p>
          </div>
        ) : (
          <>
            <div className="flex items-center justify-between gap-3">
              <span className="text-xs font-black text-amber-950">{t('research.unlockQuality', { quality: nextQuality })}</span>
              <span className="text-sm font-black tabular-nums text-green-700">{fmtCash(nextCost)}</span>
            </div>
            <p className="mt-2 text-[10px] font-semibold leading-4 text-amber-700">
              {t('research.unlockEffect', { quality: nextQuality, bonus: item.nextSalesSpeedPct ?? nextQuality * 2 })}
            </p>
          </>
        )}
      </div>

      {!maxed && (
        <button
          type="button"
          onClick={onUnlock}
          disabled={pending || !canAfford}
          className="mt-3 rounded-xl bg-cyan-700 px-4 py-3 text-xs font-black text-white transition hover:bg-cyan-800 disabled:cursor-not-allowed disabled:bg-stone-300"
        >
          {pending ? t('research.researching') : canAfford ? t('research.researchQuality', { quality: nextQuality }) : t('research.needMoreCash', { amount: fmtCash(nextCost - cash) })}
        </button>
      )}
    </article>
  )
}
