import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import enUS from './locales/en-US.json'
import zhCN from './locales/zh-CN.json'
import { SUPPORTED_LOCALES, isSupportedLocale, type SupportedLocale } from './types'

const STORAGE_KEY = 'atlas_foods_locale'

export { type SupportedLocale, SUPPORTED_LOCALES, isSupportedLocale, LOCALE_LABELS } from './types'

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      'en-US': { translation: enUS },
      'zh-CN': { translation: zhCN },
    },
    fallbackLng: 'en-US',
    supportedLngs: SUPPORTED_LOCALES,
    interpolation: {
      escapeValue: false, // React already escapes
    },
    detection: {
      order: ['localStorage', 'navigator'],
      lookupLocalStorage: STORAGE_KEY,
      caches: ['localStorage'],
    },
    returnNull: false,
    returnEmptyString: false,
  })

export function getStoredLocale(): SupportedLocale {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw && isSupportedLocale(raw)) return raw
  } catch { /* ignore */ }
  // Fall back to browser language
  if (typeof navigator !== 'undefined') {
    const browserLang = navigator.language
    // Try exact match first
    if (isSupportedLocale(browserLang)) return browserLang
    // Try prefix match (e.g. "zh" → "zh-CN")
    const prefix = browserLang.split('-')[0]
    const match = SUPPORTED_LOCALES.find((l) => l.startsWith(prefix))
    if (match) return match
  }
  return 'en-US'
}

export function setStoredLocale(locale: SupportedLocale): void {
  try {
    localStorage.setItem(STORAGE_KEY, locale)
  } catch { /* ignore */ }
  i18n.changeLanguage(locale)
}

export default i18n
