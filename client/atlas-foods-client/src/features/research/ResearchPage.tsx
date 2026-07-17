import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useCompany } from '@/api/company.api'
import { useResearch, useUnlockQuality } from '@/api/research.api'
import type { QualityResearch } from '@/api/research.api'
import { audio } from '@/audio/AudioManager'
import { MAX_PRODUCT_QUALITY } from '@/game/quality'
import { FlaskIcon } from './icons'
import { QualityResearchCard } from './QualityResearchCard'

const FILTERS = ['all', '1', '2', '3', '4'] as const
type TierFilter = (typeof FILTERS)[number]

export function ResearchPage() {
  const { t } = useTranslation()
  const [tier, setTier] = useState<TierFilter>('all')
  const { data: company } = useCompany()
  const { data: research = [], isLoading, error } = useResearch()
  const unlock = useUnlockQuality()
  const cash = company?.authCompany.money ?? 0

  const filtered = useMemo(() => tier === 'all'
    ? research
    : research.filter((item) => item.tier === Number(tier)), [research, tier])
  const totalUnlocks = research.reduce((sum, item) => sum + item.maxQuality, 0)
  const highestQuality = research.reduce((highest, item) => Math.max(highest, item.maxQuality), 0)
  const mastered = research.filter((item) => item.maxQuality >= MAX_PRODUCT_QUALITY).length
  const pendingResource = unlock.variables?.resourceId

  const handleUnlock = async (item: QualityResearch) => {
    if (!item.nextQuality) return
    audio.playSfx('research_start')
    try {
      const result = await unlock.mutateAsync({ resourceId: item.resourceId, targetQuality: item.nextQuality })
      if (result.research.charged) {
        audio.playSfx('tech_unlock')
      }
    } catch { /* mutation state renders the server error */ }
  }

  return (
    <div className="h-full overflow-y-auto bg-gradient-to-br from-[#f4e5c5] via-[#fffaf0] to-[#dceff0] p-4 sm:p-6">
      <div className="mx-auto max-w-7xl">
        <header className="overflow-hidden rounded-3xl border border-amber-200 bg-white/80 shadow-sm">
          <div className="grid gap-5 p-5 md:grid-cols-[1.35fr_1fr] md:p-7">
            <div>
              <div className="flex items-center gap-3">
                <div className="grid h-12 w-12 place-items-center rounded-2xl bg-cyan-700 text-white"><FlaskIcon className="h-7 w-7" /></div>
                <div>
                  <p className="text-[10px] font-black uppercase tracking-[0.24em] text-cyan-700">{t('research.qualityProgramme')}</p>
                  <h2 className="text-2xl font-black text-amber-950 sm:text-3xl">{t('research.qualityLab')}</h2>
                </div>
              </div>
              <p className="mt-3 max-w-2xl text-xs font-semibold leading-5 text-amber-700">{t('research.qualityLabHelp')}</p>
            </div>
            <div className="grid grid-cols-2 gap-2 sm:grid-cols-4 md:grid-cols-2">
              <ResearchMetric label={t('research.cash')} value={`$${Math.round(cash).toLocaleString()}`} tone="green" />
              <ResearchMetric label={t('research.highestQuality')} value={`Q${highestQuality}`} tone="violet" />
              <ResearchMetric label={t('research.totalUnlocks')} value={String(totalUnlocks)} tone="cyan" />
              <ResearchMetric label={t('research.masteredProducts')} value={`${mastered}/${research.length}`} tone="amber" />
            </div>
          </div>
          <div className="grid gap-px border-t border-amber-100 bg-amber-100 sm:grid-cols-3">
            <RuleStep number="1" text={t('research.ruleResearch')} />
            <RuleStep number="2" text={t('research.ruleSource')} />
            <RuleStep number="3" text={t('research.ruleSell')} />
          </div>
        </header>

        <div className="my-5 flex gap-2 overflow-x-auto pb-1">
          {FILTERS.map((filter) => (
            <button
              key={filter}
              type="button"
              onClick={() => setTier(filter)}
              className={`whitespace-nowrap rounded-full px-4 py-2 text-xs font-black transition ${tier === filter ? 'bg-amber-900 text-white shadow-sm' : 'border border-amber-200 bg-white/70 text-amber-800 hover:bg-white'}`}
            >
              {filter === 'all' ? t('research.allProducts') : t('research.productTier', { tier: filter })}
            </button>
          ))}
        </div>

        {unlock.error && <div className="mb-4 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-xs font-bold text-red-700">{unlock.error.message}</div>}
        {isLoading && <div className="rounded-3xl border border-dashed border-amber-300 bg-white/60 p-12 text-center text-sm font-semibold text-amber-600">{t('research.loading')}</div>}
        {error && <div className="rounded-3xl border border-red-200 bg-red-50 p-12 text-center text-sm font-semibold text-red-700">{t('research.failedToLoad')}</div>}
        {!isLoading && !error && filtered.length === 0 && <div className="rounded-3xl border border-dashed border-amber-300 bg-white/60 p-12 text-center text-sm font-semibold text-amber-600">{t('research.noProjects')}</div>}

        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {filtered.map((item) => (
            <QualityResearchCard
              key={item.resourceId}
              item={item}
              cash={cash}
              pending={unlock.isPending && pendingResource === item.resourceId}
              onUnlock={() => handleUnlock(item)}
            />
          ))}
        </div>
      </div>
    </div>
  )
}

function ResearchMetric({ label, value, tone }: { label: string; value: string; tone: 'green' | 'violet' | 'cyan' | 'amber' }) {
  const tones = { green: 'text-green-700', violet: 'text-violet-700', cyan: 'text-cyan-800', amber: 'text-amber-800' }
  return <div className="rounded-2xl border border-amber-100 bg-amber-50/70 px-3 py-2"><div className="text-[8px] font-black uppercase tracking-wider text-amber-500">{label}</div><div className={`mt-0.5 text-lg font-black tabular-nums ${tones[tone]}`}>{value}</div></div>
}

function RuleStep({ number, text }: { number: string; text: string }) {
  return <div className="flex items-center gap-3 bg-white/70 px-5 py-3"><span className="grid h-7 w-7 shrink-0 place-items-center rounded-full bg-cyan-700 text-xs font-black text-white">{number}</span><span className="text-[10px] font-bold leading-4 text-amber-800">{text}</span></div>
}
