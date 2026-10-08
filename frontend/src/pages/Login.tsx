import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { BrandLogo, OAuthSetupModal } from '@/components'
import { useAppleAuth, useGoogleAuth } from '@/hooks'
import { useTranslation } from '@/i18n'
import { useAuthStore, useUIStore } from '@/stores'

export default function Login() {
  const navigate = useNavigate()
  const { loginDev, loading } = useAuthStore()
  const { theme, toggleTheme } = useUIStore()
  const { t, locale, toggleLocale } = useTranslation()
  const [authError, setAuthError] = useState<string | null>(null)
  const [setupModalProvider, setSetupModalProvider] = useState<'google' | 'apple' | null>(null)

  const {
    isAuthenticating: isGoogleAuthenticating,
    loginWithGooglePopup,
  } = useGoogleAuth({
    onSuccess: () => navigate('/'),
    onError: (err) => setAuthError(err),
    onMissingConfig: () => setSetupModalProvider('google'),
  })

  const {
    isAuthenticating: isAppleAuthenticating,
    loginWithApplePopup,
  } = useAppleAuth({
    onSuccess: () => navigate('/'),
    onError: (err) => setAuthError(err),
    onMissingConfig: () => setSetupModalProvider('apple'),
  })

  const isBusy = loading || isGoogleAuthenticating || isAppleAuthenticating

  const handleDevDemoLogin = async () => {
    setAuthError(null)
    try {
      await loginDev({
        email: 'creator@taskhub.dev',
        name: 'Creator Demo',
      })
      navigate('/')
    } catch (err: unknown) {
      setAuthError((err as Error).message || 'Demo sign-in failed')
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 relative transition-colors">
      {/* Top Controls: Language Switcher & Theme Toggle */}
      <div className="absolute top-4 right-4 flex items-center gap-2">
        <button
          onClick={toggleLocale}
          className="px-2.5 py-1.5 rounded-xl border border-slate-200 dark:border-slate-800 bg-white/80 dark:bg-slate-900/80 text-xs font-semibold text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors shadow-xs flex items-center gap-1.5 cursor-pointer"
          title={t('nav.switchLanguage')}
        >
          <i className="pi pi-globe text-xs text-indigo-500" />
          <span>{locale.toUpperCase()}</span>
        </button>

        <button
          onClick={toggleTheme}
          className="p-2 rounded-xl border border-slate-200 dark:border-slate-800 bg-white/80 dark:bg-slate-900/80 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors shadow-xs cursor-pointer"
          title={theme === 'dark' ? t('nav.lightMode') : t('nav.darkMode')}
        >
          <i className={`pi ${theme === 'dark' ? 'pi-sun text-amber-400' : 'pi-moon text-indigo-500'} text-xs`} />
        </button>
      </div>

      <div className="w-full max-w-sm rounded-2xl border border-slate-200 dark:border-slate-800 bg-white/90 dark:bg-slate-900/90 p-8 shadow-xl dark:shadow-2xl backdrop-blur-xl animate-fade-in">
        {/* Brand Header */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center mb-3">
            <BrandLogo size={48} className="drop-shadow-md" />
          </div>
          <h1 className="text-xl font-bold tracking-tight text-slate-900 dark:text-slate-100">
            {t('login.brandTitle')}
          </h1>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
            {t('login.brandSubtitle')}
          </p>
        </div>

        {/* Error message */}
        {authError && (
          <div className="mb-5 p-3 rounded-xl border border-rose-200 dark:border-rose-900/50 bg-rose-50 dark:bg-rose-950/40 text-rose-600 dark:text-rose-300 text-xs text-center flex items-center justify-center gap-2">
            <i className="pi pi-exclamation-circle text-xs shrink-0" />
            <span>{authError}</span>
          </div>
        )}

        {/* OAuth Buttons (Google & Apple) */}
        <div className="space-y-3">
          {/* Continue with Google (Popup initCodeClient flow) */}
          <button
            onClick={loginWithGooglePopup}
            disabled={isBusy}
            className="w-full h-11 flex items-center justify-center gap-3 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800/80 hover:bg-slate-50 dark:hover:bg-slate-700 text-xs font-medium text-slate-800 dark:text-slate-100 transition-all shadow-xs disabled:opacity-50 cursor-pointer"
          >
            {isBusy ? (
              <i className="pi pi-spin pi-spinner text-sm text-indigo-500" />
            ) : (
              <svg className="h-4 w-4" viewBox="0 0 24 24">
                <path
                  fill="#4285F4"
                  d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                />
                <path
                  fill="#34A853"
                  d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                />
                <path
                  fill="#FBBC05"
                  d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z"
                />
                <path
                  fill="#EA4335"
                  d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"
                />
              </svg>
            )}
            <span>{t('login.googleLogin')}</span>
          </button>

          {/* Continue with Apple */}
          <button
            onClick={loginWithApplePopup}
            disabled={isBusy}
            className="w-full h-11 flex items-center justify-center gap-3 rounded-xl border border-slate-900 dark:border-slate-700 bg-slate-900 dark:bg-black hover:bg-slate-800 dark:hover:bg-slate-900 text-xs font-medium text-white transition-all shadow-xs disabled:opacity-50 cursor-pointer"
          >
            {isAppleAuthenticating ? (
              <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            ) : (
              <i className="pi pi-apple text-base" />
            )}
            <span>{t('login.appleLogin')}</span>
          </button>
        </div>

        {/* Divider */}
        <div className="relative my-6">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-slate-200 dark:border-slate-800" />
          </div>
          <div className="relative flex justify-center text-[10px] uppercase font-mono tracking-wider">
            <span className="bg-white dark:bg-slate-900 px-3 text-slate-500">
              {t('login.devTesting')}
            </span>
          </div>
        </div>

        {/* Quick Dev Demo Access */}
        <button
          onClick={handleDevDemoLogin}
          disabled={isBusy}
          className="w-full py-2.5 rounded-xl border border-indigo-200 dark:border-indigo-500/30 bg-indigo-50/70 dark:bg-indigo-950/30 hover:bg-indigo-100 dark:hover:bg-indigo-950/60 text-xs font-medium text-indigo-700 dark:text-indigo-300 transition-colors flex items-center justify-center gap-2 cursor-pointer"
        >
          <i className="pi pi-user text-xs" />
          <span>{t('login.devDemoLogin')}</span>
        </button>

        {/* Privacy & Isolation note */}
        <div className="mt-6 pt-4 border-t border-slate-100 dark:border-slate-800/80 text-center">
          <p className="text-[11px] text-slate-500 flex items-center justify-center gap-1.5">
            <i className="pi pi-lock text-[10px] text-emerald-500 dark:text-emerald-400" />
            {t('login.isolatedData')}
          </p>
        </div>
      </div>

      {/* OAuth Developer Setup Modal */}
      <OAuthSetupModal
        isOpen={Boolean(setupModalProvider)}
        provider={setupModalProvider || 'google'}
        onClose={() => setSetupModalProvider(null)}
        onUseDemoLogin={handleDevDemoLogin}
      />
    </div>
  )
}
