import { useCallback, useEffect, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'

import { BrandLogo, OAuthSetupModal, useAppToast } from '@/components'
import { useAppleAuth, useGoogleAuth } from '@/hooks'
import { useTranslation } from '@/i18n'
import { useAuthStore, useUIStore } from '@/stores'

export default function Login() {
  const navigate = useNavigate()
  const location = useLocation()
  const { loginDev, loading, isAuthenticated } = useAuthStore()
  const { theme, toggleTheme } = useUIStore()
  const { t, locale, toggleLocale } = useTranslation()
  const toast = useAppToast()

  const [authError, setAuthError] = useState<string | null>(null)
  const [setupModalProvider, setSetupModalProvider] = useState<'google' | 'apple' | null>(null)

  const targetDestination =
    (location.state as { from?: { pathname?: string } })?.from?.pathname || '/dashboard'

  // If already authenticated, redirect directly to workspace
  useEffect(() => {
    if (isAuthenticated) {
      navigate(targetDestination, { replace: true })
    }
  }, [isAuthenticated, navigate, targetDestination])

  const {
    isAuthenticating: isGoogleAuthenticating,
    loginWithGooglePopup,
  } = useGoogleAuth({
    onSuccess: () => navigate(targetDestination, { replace: true }),
    onError: (err) => setAuthError(err),
    onMissingConfig: () => setSetupModalProvider('google'),
  })

  const {
    isAuthenticating: isAppleAuthenticating,
    loginWithApplePopup,
  } = useAppleAuth({
    onSuccess: () => navigate(targetDestination, { replace: true }),
    onError: (err) => setAuthError(err),
    onMissingConfig: () => setSetupModalProvider('apple'),
  })

  const isBusy = loading || isGoogleAuthenticating || isAppleAuthenticating

  const handleDevDemoLogin = useCallback(async () => {
    setAuthError(null)
    try {
      await loginDev({
        email: 'creator@taskhub.dev',
        name: 'Creator Demo',
      })
      toast.success('Signed in as Creator Demo', 'Access Granted')
      navigate(targetDestination, { replace: true })
    } catch (err: unknown) {
      setAuthError((err as Error).message || 'Demo sign-in failed')
    }
  }, [loginDev, navigate, targetDestination, toast])

  // Keyboard shortcut listeners (D = Demo, G = Google, A = Apple, Esc = Home)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (['INPUT', 'TEXTAREA'].includes((e.target as HTMLElement)?.tagName)) return
      if (setupModalProvider) return

      if (e.key === 'd' || e.key === 'D') {
        e.preventDefault()
        if (!isBusy) handleDevDemoLogin()
      } else if (e.key === 'g' || e.key === 'G') {
        e.preventDefault()
        if (!isBusy) loginWithGooglePopup()
      } else if (e.key === 'a' || e.key === 'A') {
        e.preventDefault()
        if (!isBusy) loginWithApplePopup()
      } else if (e.key === 'Escape') {
        navigate('/')
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isBusy, setupModalProvider, handleDevDemoLogin, loginWithGooglePopup, loginWithApplePopup, navigate])

  return (
    <div className="min-h-screen w-full flex flex-col lg:flex-row bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 transition-colors selection:bg-indigo-500/30 selection:text-indigo-200 relative overflow-x-hidden">
      {/* Background Ambient Glows */}
      <div className="pointer-events-none fixed inset-0 overflow-hidden opacity-40 dark:opacity-25 z-0">
        <div className="absolute top-1/4 -left-20 w-[500px] h-[500px] bg-indigo-500/25 rounded-full blur-[130px]" />
        <div className="absolute bottom-10 right-1/4 w-[450px] h-[450px] bg-purple-600/25 rounded-full blur-[140px]" />
        <div className="absolute top-10 right-10 w-[400px] h-[400px] bg-sky-500/20 rounded-full blur-[110px]" />
      </div>

      {/* =========================================================================
          LEFT HALF: NATURAL WORKSPACE SHOWCASE (Kanban + Scratchpad Duo)
          ========================================================================= */}
      <div className="hidden lg:flex lg:col-span-7 flex-1 flex-col justify-between p-10 xl:p-14 relative z-10 border-r border-slate-200/80 dark:border-slate-900 bg-white/40 dark:bg-slate-950/70 backdrop-blur-xl">
        {/* Top Header: Brand & Back to Home */}
        <div className="flex items-center justify-between">
          <Link to="/" className="flex items-center gap-3 group cursor-pointer">
            <BrandLogo size={36} className="transition-transform group-hover:scale-105 duration-200" />
            <span className="font-extrabold text-lg tracking-tight text-slate-900 dark:text-white">
              Task<span className="text-indigo-600 dark:text-indigo-400">Hub</span>
            </span>
          </Link>

          <Link
            to="/"
            className="flex items-center gap-2 text-xs font-medium text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white transition-colors px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white/80 dark:bg-slate-900/60 hover:bg-slate-100 dark:hover:bg-slate-800 shadow-xs"
          >
            <i className="pi pi-arrow-left text-[10px]" />
            <span>Back to Home</span>
            <kbd className="px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-[10px] text-slate-500 dark:text-slate-400 font-mono">
              Esc
            </kbd>
          </Link>
        </div>

        {/* Center Showcase Content: Realistic Kanban Board & Notes Duo */}
        <div className="max-w-2xl mx-auto w-full my-8 space-y-6">
          <div className="space-y-2 text-left">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-indigo-200 dark:border-indigo-500/30 bg-indigo-50/80 dark:bg-indigo-950/40 text-indigo-600 dark:text-indigo-400 text-xs font-medium shadow-xs">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-ping" />
              <span>High-Velocity Engineering Workspace</span>
            </div>
            <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-white leading-tight">
              Manage Tasks & AI Agents at the Speed of Thought
            </h2>
            <p className="text-sm text-slate-600 dark:text-slate-300 leading-relaxed">
              Designed for developers who value flow state. Native Kanban boards, instant markdown notes, and Model Context Protocol (MCP) server integration.
            </p>
          </div>

          {/* Product Duo: Kanban Cards Column & Scratchpad Card */}
          <div className="grid grid-cols-1 md:grid-cols-12 gap-4 text-left">
            {/* Left Column: Kanban Tasks Column */}
            <div className="md:col-span-7 rounded-2xl border border-slate-200 dark:border-slate-800/80 bg-white/90 dark:bg-slate-900/80 p-4 shadow-xl backdrop-blur-xl space-y-3">
              <div className="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-800">
                <div className="flex items-center gap-2">
                  <span className="w-2.5 h-2.5 rounded-full bg-sky-500" />
                  <span className="text-xs font-bold text-slate-800 dark:text-slate-200">In Progress</span>
                  <span className="text-[10px] bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded text-slate-500 font-mono">
                    3
                  </span>
                </div>
                <span className="text-[11px] text-indigo-600 dark:text-indigo-400 font-medium flex items-center gap-1">
                  <i className="pi pi-compass text-[10px]" /> Horizon Focus
                </span>
              </div>

              {/* Task 1 */}
              <div className="p-3.5 rounded-xl border border-indigo-200/80 dark:border-indigo-500/30 bg-indigo-50/30 dark:bg-indigo-950/20 shadow-xs space-y-2 hover:border-indigo-400 transition-colors">
                <div className="flex items-center justify-between text-xs">
                  <span className="px-2 py-0.5 rounded bg-rose-100 dark:bg-rose-950/60 text-rose-600 dark:text-rose-400 text-[10px] font-bold uppercase tracking-wider">
                    Urgent
                  </span>
                  <span className="text-[10px] font-mono text-slate-400">#TSK-104</span>
                </div>
                <h4 className="text-xs font-semibold text-slate-900 dark:text-slate-100 leading-snug">
                  Connect Cursor MCP stdio bridge for realtime status sync
                </h4>
                <div className="flex items-center gap-2 text-[10px] text-slate-500">
                  <span className="bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded font-mono">#mcp</span>
                  <span className="bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded font-mono">#ai</span>
                  <span className="ml-auto text-emerald-600 dark:text-emerald-400 font-medium flex items-center gap-1">
                    <i className="pi pi-check text-[9px]" /> 3/3 done
                  </span>
                </div>
                <div className="w-full bg-slate-200 dark:bg-slate-800 h-1.5 rounded-full overflow-hidden">
                  <div className="bg-emerald-500 h-full w-full rounded-full" />
                </div>
              </div>

              {/* Task 2 */}
              <div className="p-3.5 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xs space-y-2 hover:border-slate-300 dark:hover:border-slate-700 transition-colors">
                <div className="flex items-center justify-between text-xs">
                  <span className="px-2 py-0.5 rounded bg-amber-100 dark:bg-amber-950/60 text-amber-600 dark:text-amber-400 text-[10px] font-bold uppercase tracking-wider">
                    High
                  </span>
                  <span className="text-[10px] font-mono text-slate-400">#TSK-108</span>
                </div>
                <h4 className="text-xs font-semibold text-slate-900 dark:text-slate-100 leading-snug">
                  Implement Apple Sign-In JWKS cryptographic validation
                </h4>
                <div className="flex items-center gap-2 text-[10px] text-slate-500">
                  <span className="bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded font-mono">#security</span>
                  <span className="ml-auto text-indigo-600 dark:text-indigo-400 font-medium">80%</span>
                </div>
                <div className="w-full bg-slate-200 dark:bg-slate-800 h-1.5 rounded-full overflow-hidden">
                  <div className="bg-indigo-500 h-full w-4/5 rounded-full" />
                </div>
              </div>
            </div>

            {/* Right Column: Markdown Scratchpad */}
            <div className="md:col-span-5 rounded-2xl border border-slate-200 dark:border-slate-800/80 bg-white/90 dark:bg-slate-900/80 p-4 shadow-xl backdrop-blur-xl flex flex-col justify-between space-y-3">
              <div className="space-y-2">
                <div className="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-800">
                  <div className="flex items-center gap-1.5 text-xs font-bold text-slate-800 dark:text-slate-200">
                    <i className="pi pi-file-edit text-indigo-500 text-xs" />
                    <span>Scratchpad</span>
                  </div>
                  <span className="text-[10px] text-amber-500 flex items-center gap-1 font-mono font-medium">
                    <i className="pi pi-bookmark-fill text-[9px]" /> Pinned
                  </span>
                </div>
                <h4 className="text-xs font-semibold text-slate-900 dark:text-slate-100">
                  Sprint Roadmap Notes
                </h4>
                <div className="p-2.5 rounded-xl bg-slate-50 dark:bg-slate-950/70 border border-slate-200/60 dark:border-slate-800/60 font-mono text-[11px] text-slate-600 dark:text-slate-300 space-y-1">
                  <p className="text-indigo-600 dark:text-indigo-400 font-semibold">## Goals</p>
                  <p className="flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400">
                    <i className="pi pi-check text-[9px]" /> Sub-10ms query latency
                  </p>
                  <p className="flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400">
                    <i className="pi pi-check text-[9px]" /> Native MCP Server bridge
                  </p>
                  <p className="flex items-center gap-1.5 text-slate-400">
                    <i className="pi pi-circle text-[8px]" /> Automated changelogs
                  </p>
                </div>
              </div>

              <div className="pt-2">
                <div className="w-full py-2 rounded-xl bg-indigo-50 dark:bg-indigo-950/40 border border-indigo-200 dark:border-indigo-500/30 text-indigo-700 dark:text-indigo-300 text-[11px] font-semibold flex items-center justify-center gap-1.5">
                  <i className="pi pi-bolt text-[10px]" />
                  <span>1-Click Convert to Task</span>
                </div>
              </div>
            </div>
          </div>

          {/* Quick Metrics Strip */}
          <div className="grid grid-cols-3 gap-3 pt-2">
            <div className="p-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-white/60 dark:bg-slate-900/40 text-center space-y-0.5 shadow-xs">
              <div className="text-base font-extrabold text-emerald-600 dark:text-emerald-400 font-mono">&lt; 10ms</div>
              <div className="text-[11px] text-slate-500 dark:text-slate-400 font-medium">UI Latency</div>
            </div>
            <div className="p-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-white/60 dark:bg-slate-900/40 text-center space-y-0.5 shadow-xs">
              <div className="text-base font-extrabold text-indigo-600 dark:text-indigo-400 font-mono">MCP 2.0</div>
              <div className="text-[11px] text-slate-500 dark:text-slate-400 font-medium">AI Integration</div>
            </div>
            <div className="p-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-white/60 dark:bg-slate-900/40 text-center space-y-0.5 shadow-xs">
              <div className="text-base font-extrabold text-purple-600 dark:text-purple-400 font-mono">100%</div>
              <div className="text-[11px] text-slate-500 dark:text-slate-400 font-medium">Isolated Data</div>
            </div>
          </div>
        </div>

        {/* Bottom Footer Credits */}
        <div className="pt-4 border-t border-slate-200 dark:border-slate-900 text-xs text-slate-500 flex items-center justify-between font-mono">
          <span>TaskHub Developer Edition</span>
          <span>Go Fiber • React 19 • Tailwind v4</span>
        </div>
      </div>

      {/* =========================================================================
          RIGHT HALF: SPACIOUS, NATURAL AUTHENTICATION STUDIO
          ========================================================================= */}
      <div className="lg:col-span-5 flex-1 flex flex-col justify-center items-center p-6 sm:p-10 lg:p-14 relative z-10">
        {/* Top Controls: Language Switcher & Theme Toggle */}
        <div className="absolute top-6 right-6 flex items-center gap-2">
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
            className="p-2 rounded-xl border border-slate-200 dark:border-slate-800 bg-white/80 dark:bg-slate-900/80 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors shadow-xs cursor-pointer"
            title={theme === 'dark' ? t('nav.lightMode') : t('nav.darkMode')}
          >
            <i className={`pi ${theme === 'dark' ? 'pi-sun text-amber-400' : 'pi-moon text-indigo-500'} text-xs`} />
          </button>
        </div>

        {/* Spacious, Comfortable Studio Card (Not Small, Well-Balanced) */}
        <div className="w-full max-w-lg rounded-3xl border border-slate-200/90 dark:border-slate-800 bg-white/95 dark:bg-slate-900/90 p-8 sm:p-12 shadow-2xl backdrop-blur-2xl space-y-7 animate-fade-in text-left">
          {/* Header */}
          <div className="space-y-2">
            <div className="flex items-center gap-2.5 mb-1">
              <BrandLogo size={40} className="drop-shadow-sm" />
              <div className="flex items-center gap-1.5">
                <span className="font-extrabold text-lg tracking-tight text-slate-900 dark:text-white">TaskHub</span>
                <span className="px-2 py-0.5 rounded-full bg-indigo-50 dark:bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 text-[10px] font-mono font-semibold border border-indigo-200 dark:border-indigo-500/30">
                  Workspace
                </span>
              </div>
            </div>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-slate-900 dark:text-white tracking-tight">
              Sign In to Your Workspace
            </h1>
            <p className="text-xs sm:text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
              Welcome back. Choose your preferred authentication method to continue to your boards.
            </p>
          </div>

          {/* Auth Error Banner */}
          {authError && (
            <div className="p-3.5 rounded-2xl border border-rose-200 dark:border-rose-900/60 bg-rose-50 dark:bg-rose-950/40 text-rose-600 dark:text-rose-300 text-xs flex items-center gap-2.5">
              <i className="pi pi-exclamation-circle text-rose-500 shrink-0 text-sm" />
              <span>{authError}</span>
            </div>
          )}

          {/* FEATURED: Instant One-Click Demo Sandbox */}
          <div className="p-5 rounded-2xl border border-emerald-300/80 dark:border-emerald-500/30 bg-gradient-to-r from-emerald-50/80 to-indigo-50/60 dark:from-emerald-950/30 dark:to-indigo-950/30 space-y-3 relative overflow-hidden group shadow-xs">
            <div className="flex items-center justify-between">
              <span className="inline-flex items-center gap-1.5 text-xs font-mono font-bold uppercase tracking-wider text-emerald-700 dark:text-emerald-400">
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-ping" />
                Instant Demo Sandbox
              </span>
              <kbd className="px-2 py-0.5 rounded-lg bg-white/90 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-[10px] text-slate-600 dark:text-slate-300 font-mono shadow-2xs">
                Press [D]
              </kbd>
            </div>
            <p className="text-xs text-slate-600 dark:text-slate-300 leading-relaxed">
              Jump straight into a live interactive workspace preloaded with sprint tasks and notes. Zero setup required.
            </p>
            <button
              onClick={handleDevDemoLogin}
              disabled={isBusy}
              className="w-full h-11 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-bold text-xs sm:text-sm shadow-md shadow-emerald-600/20 transition-all flex items-center justify-center gap-2 cursor-pointer disabled:opacity-50"
            >
              {loading ? (
                <i className="pi pi-spin pi-spinner text-xs" />
              ) : (
                <i className="pi pi-play text-xs text-white" />
              )}
              <span>Launch Demo Workspace (1-Click)</span>
            </button>
          </div>

          {/* Divider */}
          <div className="relative my-4">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-slate-200 dark:border-slate-800" />
            </div>
            <div className="relative flex justify-center text-[10px] uppercase font-mono tracking-wider">
              <span className="bg-white dark:bg-slate-900 px-3 text-slate-500">Or continue with SSO</span>
            </div>
          </div>

          {/* Spacious OAuth Buttons (Google & Apple) */}
          <div className="space-y-3">
            {/* Continue with Google */}
            <button
              onClick={loginWithGooglePopup}
              disabled={isBusy}
              className="w-full h-12 flex items-center justify-between px-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-800/80 hover:bg-slate-50 dark:hover:bg-slate-800 text-xs sm:text-sm font-semibold text-slate-800 dark:text-slate-100 transition-all shadow-xs disabled:opacity-50 cursor-pointer group"
            >
              <div className="flex items-center gap-3">
                {isGoogleAuthenticating ? (
                  <i className="pi pi-spin pi-spinner text-sm text-indigo-500" />
                ) : (
                  <svg className="h-4 w-4 shrink-0" viewBox="0 0 24 24">
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
                <span>Continue with Google</span>
              </div>
              <kbd className="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-700/60 text-[10px] text-slate-500 dark:text-slate-400 font-mono group-hover:text-slate-800 dark:group-hover:text-slate-200">
                [G]
              </kbd>
            </button>

            {/* Continue with Apple */}
            <button
              onClick={loginWithApplePopup}
              disabled={isBusy}
              className="w-full h-12 flex items-center justify-between px-4 rounded-xl border border-slate-900 dark:border-slate-700 bg-slate-900 dark:bg-black hover:bg-slate-800 dark:hover:bg-slate-900 text-xs sm:text-sm font-semibold text-white transition-all shadow-xs disabled:opacity-50 cursor-pointer group"
            >
              <div className="flex items-center gap-3">
                {isAppleAuthenticating ? (
                  <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                ) : (
                  <i className="pi pi-apple text-base" />
                )}
                <span>Continue with Apple</span>
              </div>
              <kbd className="px-2 py-0.5 rounded bg-slate-800 text-[10px] text-slate-400 font-mono group-hover:text-white">
                [A]
              </kbd>
            </button>
          </div>

          {/* Keyboard Shortcuts Hint Bar */}
          <div className="pt-2 flex items-center justify-center gap-4 text-[11px] text-slate-400 font-mono">
            <span>[D] Demo</span>
            <span>•</span>
            <span>[G] Google</span>
            <span>•</span>
            <span>[A] Apple</span>
            <span>•</span>
            <span>[Esc] Home</span>
          </div>

          {/* Privacy & Tenant Isolation Note */}
          <div className="pt-3 border-t border-slate-100 dark:border-slate-800/80 text-center">
            <p className="text-[11px] text-slate-500 flex items-center justify-center gap-1.5">
              <i className="pi pi-shield text-[10px] text-emerald-500" />
              <span>Multi-tenant encrypted workspace with isolated data storage</span>
            </p>
          </div>
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
