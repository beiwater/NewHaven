import { useTranslation } from 'react-i18next'
import { MAX_PRODUCT_QUALITY, qualitySalesBonusPct } from '@/game/quality'

export function QualitySelector({ value, onChange, disabled = false, maxQuality = MAX_PRODUCT_QUALITY }: { value: number; onChange: (quality: number) => void; disabled?: boolean; maxQuality?: number }) {
  const { t } = useTranslation()
  const qualityCeiling = Math.max(0, Math.min(MAX_PRODUCT_QUALITY, Math.trunc(maxQuality)))

  return (
    <label className="flex items-center justify-between gap-3 rounded-xl border border-violet-200 bg-violet-50/80 px-3 py-2">
      <span>
        <span className="block text-[9px] font-black uppercase tracking-[0.18em] text-violet-600">{t('quality.tier')}</span>
        <span className="block text-[10px] font-semibold text-violet-800">+{qualitySalesBonusPct(value)}% {t('quality.retailSpeed')}</span>
      </span>
      <select
        aria-label={t('quality.tier')}
        value={value}
        disabled={disabled}
        onChange={(event) => onChange(Number(event.target.value))}
        className="rounded-lg border border-violet-300 bg-white px-3 py-2 text-xs font-black text-violet-900 disabled:bg-slate-100"
      >
        {Array.from({ length: qualityCeiling + 1 }, (_, quality) => (
          <option key={quality} value={quality}>Q{quality}</option>
        ))}
      </select>
    </label>
  )
}
