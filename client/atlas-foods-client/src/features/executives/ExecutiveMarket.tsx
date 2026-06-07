import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useExecutiveSearch, useRecruitExecutive } from '@/api/executives.api'
import { useCompany } from '@/api/company.api'
import { audio } from '@/audio/AudioManager'
import { ExecutiveCard } from './ExecutiveCard'

export function ExecutiveMarket() {
  const { t } = useTranslation()
  const { data, isLoading, isError, error, refetch } = useExecutiveSearch()
  const recruitExec = useRecruitExecutive()
  const { data: companyData } = useCompany()
  const [recruitingId, setRecruitingId] = useState<string | null>(null)

  const cash = companyData?.authCompany?.money ?? 0
  const executives = data?.executives ?? []
  const refreshCooldown = data?.refreshCooldown ?? '09:00:00'

  const handleRecruit = async (id: string) => {
    setRecruitingId(id)
    audio.playSfx('executive_hire')
    try {
      await recruitExec.mutateAsync(id)
    } finally {
      setRecruitingId(null)
    }
  }

  return (
    <section>
      {/* Header */}
      <div className="mb-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <svg className="h-5 w-5 text-amber-800" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z" />
          </svg>
          <h3 className="text-base font-black uppercase tracking-wider text-amber-900">
            {t('executives.market')}
          </h3>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-[10px] font-semibold text-amber-600">
            {t('executives.refreshesIn')}: {refreshCooldown}
          </span>
          <button
            onClick={() => refetch()}
            disabled={isLoading}
            className="rounded-lg bg-amber-800 px-3 py-1.5 text-[11px] font-black text-white hover:bg-amber-900 disabled:bg-amber-400 transition-colors"
          >
            {isLoading ? t('executives.refreshing') : t('executives.refresh')}
          </button>
        </div>
      </div>

      {/* Loading */}
      {isLoading && (
        <div className="flex items-center justify-center rounded-xl border border-amber-200/60 bg-white/50 py-12">
          <div className="text-xs font-semibold text-amber-600 animate-pulse">{t('executives.searching')}</div>
        </div>
      )}

      {/* Error */}
      {isError && !isLoading && (
        <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-center">
          <p className="text-xs font-semibold text-red-700">
            {t('executives.loadFailed')}: {error instanceof Error ? error.message : t('common.error')}
          </p>
          <button
            onClick={() => refetch()}
            className="mt-2 rounded-lg bg-red-100 px-4 py-1.5 text-xs font-bold text-red-800 hover:bg-red-200 transition-colors"
          >
            {t('common.retry')}
          </button>
        </div>
      )}

      {/* Recruit success/error feedback */}
      {recruitExec.isSuccess && (
        <div className="mb-3 rounded-lg border border-green-200 bg-green-50 px-3 py-2 text-xs font-semibold text-green-700">
          {t('executives.recruitSuccess')}
        </div>
      )}
      {recruitExec.isError && (
        <div className="mb-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-xs font-semibold text-red-700">
          {t('executives.recruitFailed')}: {recruitExec.error instanceof Error ? recruitExec.error.message : t('common.error')}
        </div>
      )}

      {/* Market cards grid */}
      {!isLoading && !isError && (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {executives.length === 0 && (
            <div className="col-span-full rounded-xl border border-dashed border-amber-300/50 bg-white/40 py-8 text-center">
              <p className="text-xs text-amber-500">{t('executives.noExecutivesAvailable')}</p>
            </div>
          )}
          {executives.map((exec) => (
            <ExecutiveCard
              key={exec.id}
              executive={exec}
              mode="market"
              isPending={recruitingId === exec.id}
              canAfford={cash >= exec.recruitCost}
              onRecruit={handleRecruit}
            />
          ))}
        </div>
      )}
    </section>
  )
}
