import { useActivePowerup, type ActivePowerup } from '@/api/powerup.api'
import { useResearchProgress, type ResearchProject } from '@/api/research.api'
import { useCompany, usePlayerLevel, type UnlockInfo } from '@/api/company.api'
import { useTranslation } from 'react-i18next'

export function InspectPage() {
  const { data: activeData } = useActivePowerup()
  const { data: progressData } = useResearchProgress()
  const { data: companyData } = useCompany()
  const { data: levelData } = usePlayerLevel()
  const { t } = useTranslation()

  const activePowerups: ActivePowerup[] = activeData?.active ?? []
  const completedResearch: ResearchProject[] = (progressData?.projects ?? []).filter(
    (p) => p.status === 'completed',
  )
  const unlocks: UnlockInfo | undefined = levelData?.unlocks ?? companyData?.unlocks
  const unlockedFeatures = unlocks
    ? Object.entries(unlocks.features).filter(([, v]) => v).map(([k]) => k)
    : []

  return (
    <div className="h-full overflow-y-auto">
      <div className="mx-auto max-w-lg p-4 space-y-5">
        {/* Header */}
        <div className="flex items-center gap-3">
          <svg className="h-6 w-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
          </svg>
          <div>
            <p className="text-[10px] font-bold uppercase tracking-[0.24em] text-amber-700/70">{t('inspect.subtitle')}</p>
            <h2 className="text-xl font-black text-amber-950">{t('inspect.title')}</h2>
          </div>
        </div>

        {/* Active Power-ups */}
        <section>
          <h3 className="flex items-center gap-2 text-xs font-bold text-amber-800 uppercase tracking-wider mb-2">
            <svg className="w-4 h-4 text-yellow-500" fill="currentColor" viewBox="0 0 24 24">
              <path d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            {t('inspect.activePowerups')}
          </h3>
          {activePowerups.length === 0 ? (
            <div className="text-[10px] text-amber-400 italic py-2">{t('inspect.noActivePowerups')}</div>
          ) : (
            <div className="space-y-1.5">
              {activePowerups.map((p) => (
                <div key={p.type} className="flex items-center gap-3 p-3 bg-green-50/80 rounded-xl border border-green-200/40">
                  <span className="text-lg">⚡</span>
                  <div className="flex-1">
                    <div className="text-xs font-semibold text-green-800">{p.type}</div>
                    <div className="text-[10px] text-green-600">
                      {t('inspect.ends')}: {new Date(p.endsAt).toLocaleTimeString()}
                    </div>
                  </div>
                  <span className="text-[10px] font-bold text-green-700 bg-green-200/50 px-2 py-0.5 rounded-full">{t('inspect.active')}</span>
                </div>
              ))}
            </div>
          )}
        </section>

        {/* Research Bonuses */}
        <section>
          <h3 className="flex items-center gap-2 text-xs font-bold text-amber-800 uppercase tracking-wider mb-2">
            <svg className="w-4 h-4 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
            </svg>
            {t('inspect.researchBonuses')}
          </h3>
          {completedResearch.length === 0 ? (
            <div className="text-[10px] text-amber-400 italic py-2">{t('inspect.noResearchBonuses')}</div>
          ) : (
            <div className="space-y-1.5">
              {completedResearch.map((r) => (
                <div key={r.id} className="flex items-center gap-3 p-3 bg-purple-50/80 rounded-xl border border-purple-200/40">
                  <span className="text-lg">🔬</span>
                  <div className="flex-1">
                    <div className="text-xs font-semibold text-purple-800">{r.name}</div>
                    {r.unlockPct != null && (
                      <div className="text-[10px] text-purple-600">+{r.unlockPct}% {t('inspect.quality')}</div>
                    )}
                  </div>
                  <span className="text-[10px] font-bold text-purple-700 bg-purple-200/50 px-2 py-0.5 rounded-full">{t('inspect.done')}</span>
                </div>
              ))}
            </div>
          )}
        </section>

        {/* Level Unlocks */}
        <section>
          <h3 className="flex items-center gap-2 text-xs font-bold text-amber-800 uppercase tracking-wider mb-2">
            <svg className="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
            {t('inspect.unlockedFeatures')}
          </h3>
          {unlockedFeatures.length === 0 ? (
            <div className="text-[10px] text-amber-400 italic py-2">{t('inspect.noUnlockedFeatures')}</div>
          ) : (
            <div className="flex flex-wrap gap-1.5">
              {unlockedFeatures.map((f) => (
                <span key={f} className="text-[10px] font-semibold text-blue-700 bg-blue-100/70 px-2.5 py-1 rounded-full border border-blue-200/40 capitalize">
                  {t('nav.' + f.toLowerCase(), f)}
                </span>
              ))}
            </div>
          )}
          {unlocks && (
            <div className="mt-2 space-y-0.5">
              {Object.entries(unlocks.features)
                .filter(([, v]) => !v)
                .map(([k]) => (
                  <div key={k} className="text-[10px] text-amber-400 flex items-center gap-1">
                    <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                    </svg>
                    <span className="capitalize">{t('nav.' + k.toLowerCase(), k)}</span>
                    <span>{t('inspect.unlocksAt', { level: unlocks.featureLevels?.[k] ?? '?' })}</span>
                  </div>
                ))}
            </div>
          )}
        </section>
      </div>
    </div>
  )
}
