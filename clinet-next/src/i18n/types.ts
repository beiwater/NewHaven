export type SupportedLocale = 'en-US' | 'zh-CN'

export const SUPPORTED_LOCALES: SupportedLocale[] = ['en-US', 'zh-CN']

export const LOCALE_LABELS: Record<SupportedLocale, string> = {
  'en-US': 'English',
  'zh-CN': '简体中文',
}

export function isSupportedLocale(value: string): value is SupportedLocale {
  return SUPPORTED_LOCALES.includes(value as SupportedLocale)
}
