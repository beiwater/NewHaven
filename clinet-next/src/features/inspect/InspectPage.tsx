import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import { useTranslation } from 'react-i18next'

export function InspectPage() {
  const { t } = useTranslation()
  const { data: companyData } = useQuery({
    queryKey: ['company'],
    queryFn: () => api.get('/api/v2/companies/me/'),
  })

  return (
    <div className="p-4 overflow-y-auto h-full">
      <h2 className="text-lg font-bold text-amber-900 mb-4">{t('nav.inspect', 'Inspect')}</h2>
      <div className="space-y-2 text-xs text-amber-700">
        <pre className="bg-white/60 p-4 rounded-lg border border-amber-200/40 overflow-auto max-h-[70vh]">
          {JSON.stringify(companyData, null, 2)}
        </pre>
      </div>
    </div>
  )
}
