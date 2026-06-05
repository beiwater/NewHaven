import { useTranslation } from 'react-i18next'

export default function MapSlot() {
  const { t } = useTranslation()
  return (
    <div className="flex items-center justify-center h-full bg-amber-100">
      <div className="text-center p-8">
        <div className="text-lg font-bold text-amber-800 mb-2">{t('map.title', 'Game Map')}</div>
        <p className="text-sm text-amber-600">Map view — PixiJS canvas placeholder</p>
      </div>
    </div>
  )
}
