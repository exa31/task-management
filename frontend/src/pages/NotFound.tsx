import { Link } from 'react-router-dom'

import { useTranslation } from '@/i18n'

export default function NotFound() {
  const { t } = useTranslation()

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-4 bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-50 text-center">
      <div className="max-w-md w-full">
        <div className="text-indigo-500 font-bold text-9xl tracking-tighter mb-4 opacity-50 dark:opacity-20">
          404
        </div>
        <h1 className="text-3xl font-bold tracking-tight mb-3">{t('errors.notFoundTitle')}</h1>
        <p className="text-slate-500 dark:text-slate-400 mb-8">{t('errors.notFoundMessage')}</p>
        <Link
          to="/"
          className="inline-flex items-center gap-2 bg-indigo-600 hover:bg-indigo-500 text-white px-6 py-2.5 rounded-xl text-xs font-semibold shadow-md shadow-indigo-600/20 transition-all cursor-pointer"
        >
          <i className="pi pi-home text-xs" />
          <span>{t('errors.backHome')}</span>
        </Link>
      </div>
    </div>
  )
}
