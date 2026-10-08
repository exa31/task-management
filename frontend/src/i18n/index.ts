import { addLocale, locale as setPrimeReactLocale } from 'primereact/api'

import { useUIStore } from '@/stores/uiStore'

import { en, type TranslationSchema } from './locales/en'
import { id } from './locales/id'

export type Locale = 'en' | 'id'

export const translations: Record<Locale, TranslationSchema> = {
  en,
  id,
}

// Register Indonesian locale in PrimeReact
addLocale('id', {
  firstDayOfWeek: 1,
  dayNames: ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu'],
  dayNamesShort: ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab'],
  dayNamesMin: ['Mg', 'Sn', 'Sl', 'Rb', 'Km', 'Jm', 'Sb'],
  monthNames: [
    'Januari',
    'Februari',
    'Maret',
    'April',
    'Mei',
    'Juni',
    'Juli',
    'Agustus',
    'September',
    'Oktober',
    'November',
    'Desember',
  ],
  monthNamesShort: ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des'],
  today: 'Hari ini',
  clear: 'Hapus',
  accept: 'Ya',
  reject: 'Batal',
  choose: 'Pilih',
  upload: 'Unggah',
  cancel: 'Batal',
})

/**
 * Resolves dot-notated string key path into a translated template string,
 * then interpolates {params}.
 */
export function translate(
  currentLocale: Locale,
  path: string,
  params?: Record<string, string | number>
): string {
  const dictionary = translations[currentLocale] || translations.en
  const parts = path.split('.')
  let current: unknown = dictionary

  for (const part of parts) {
    if (current && typeof current === 'object' && part in current) {
      current = (current as Record<string, unknown>)[part]
    } else {
      current = undefined
      break
    }
  }

  // Fallback to English if missing in target locale
  if (typeof current !== 'string' && currentLocale !== 'en') {
    let fallback: unknown = translations.en
    for (const part of parts) {
      if (fallback && typeof fallback === 'object' && part in fallback) {
        fallback = (fallback as Record<string, unknown>)[part]
      } else {
        fallback = undefined
        break
      }
    }
    if (typeof fallback === 'string') {
      current = fallback
    }
  }

  const template = typeof current === 'string' ? current : path

  if (!params) return template

  return template.replace(/\{(\w+)\}/g, (match, key) => key in params ? String(params[key]) : match)
}

/**
 * Custom React Hook providing translation function and locale switching helpers.
 */
export function useTranslation() {
  const { locale, setLocale, toggleLocale } = useUIStore()

  const t = (path: string, params?: Record<string, string | number>) => translate(locale, path, params)

  return {
    t,
    locale,
    setLocale: (next: Locale) => {
      setPrimeReactLocale(next)
      setLocale(next)
    },
    toggleLocale: () => {
      const next: Locale = locale === 'en' ? 'id' : 'en'
      setPrimeReactLocale(next)
      toggleLocale()
    },
  }
}

export * from './locales/en'
export * from './locales/id'
