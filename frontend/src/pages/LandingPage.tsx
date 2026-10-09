import React, { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { BrandLogo, useAppToast } from '@/components'
import { useTranslation } from '@/i18n'
import { useAuthStore, useUIStore } from '@/stores'

export const LandingPage: React.FC = () => {
  const navigate = useNavigate()
  const { user, isAuthenticated, loginDev } = useAuthStore()
  const { theme, toggleTheme } = useUIStore()
  const { t, locale, toggleLocale } = useTranslation()
  const toast = useAppToast()
  const [copiedMcp, setCopiedMcp] = useState(false)
  const [isDemoLoading, setIsDemoLoading] = useState(false)

  // If user is already authenticated, redirect straight to workspace dashboard
  React.useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    if (isAuthenticated && !params.get('preview')) {
      navigate('/dashboard', { replace: true })
    }
  }, [isAuthenticated, navigate])

  const handleDemoLogin = async () => {
    setIsDemoLoading(true)
    try {
      await loginDev({
        email: 'creator@taskhub.dev',
        name: 'Creator Demo',
      })
      toast.success('Welcome to your TaskHub workspace sandbox!', 'Demo Access Granted')
      navigate('/dashboard')
    } catch (err: unknown) {
      toast.error((err as Error).message || 'Could not start demo session', 'Demo Login Failed')
    } finally {
      setIsDemoLoading(false)
    }
  }

  const mcpConfigCode = JSON.stringify(
    {
      mcpServers: {
        taskhub: {
          command: 'npx',
          args: ['-y', '@taskhub/mcp-server', '--api-url', 'https://be-task-hub.eka-dev.cloud'],
        },
      },
    },
    null,
    2
  )

  const handleCopyMcp = async () => {
    try {
      await navigator.clipboard.writeText(mcpConfigCode)
      setCopiedMcp(true)
      toast.success('Ready to paste into your cursor_settings.json or agy config.', t('landing.mcpShowcase.copied'))
      setTimeout(() => setCopiedMcp(false), 2500)
    } catch {
      // ignore
    }
  }

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 transition-colors selection:bg-indigo-500/30 selection:text-indigo-300 relative overflow-x-hidden">
      {/* Ambient background glows */}
      <div className="pointer-events-none absolute top-0 left-1/2 -translate-x-1/2 w-full max-w-7xl h-[600px] overflow-hidden opacity-40 dark:opacity-20 z-0">
        <div className="absolute -top-32 left-1/4 w-96 h-96 bg-indigo-500 rounded-full blur-3xl animate-pulse" />
        <div className="absolute top-10 right-1/4 w-96 h-96 bg-purple-600 rounded-full blur-3xl" />
        <div className="absolute top-60 left-1/3 w-80 h-80 bg-sky-500 rounded-full blur-3xl" />
      </div>

      {/* =========================================================================
          1. NAVIGATION BAR
          ========================================================================= */}
      <header className="sticky top-0 z-50 backdrop-blur-xl bg-white/70 dark:bg-slate-950/70 border-b border-slate-200/80 dark:border-slate-800/80 transition-colors">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between">
          {/* Logo & Brand */}
          <Link to="/" className="flex items-center gap-2.5 group">
            <BrandLogo size={32} className="transition-transform group-hover:scale-105" />
            <span className="font-extrabold text-base tracking-tight text-slate-900 dark:text-slate-100">
              Task<span className="text-indigo-600 dark:text-indigo-400">Hub</span>
            </span>
          </Link>

          {/* Center Navigation Links (Hidden on mobile) */}
          <nav className="hidden md:flex items-center gap-6 text-xs font-medium text-slate-600 dark:text-slate-300">
            <a href="#features" className="hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors">
              {t('landing.features.sectionTitle')}
            </a>
            <a href="#kanban" className="hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors">
              {t('landing.features.kanbanTitle')}
            </a>
            <a href="#mcp" className="hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors">
              {t('landing.features.mcpTitle')}
            </a>
            <a href="#comparison" className="hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors">
              {t('landing.comparison.title')}
            </a>
          </nav>

          {/* Right Controls: Language, Theme, & Action Buttons */}
          <div className="flex items-center gap-2 sm:gap-3">
            {/* Language toggle */}
            <button
              onClick={toggleLocale}
              className="px-2.5 py-1.5 rounded-xl border border-slate-200 dark:border-slate-800 bg-white/80 dark:bg-slate-900/80 text-xs font-semibold text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors shadow-xs flex items-center gap-1.5 cursor-pointer"
              title={t('nav.switchLanguage')}
            >
              <i className="pi pi-globe text-xs text-indigo-500" />
              <span>{locale.toUpperCase()}</span>
            </button>

            {/* Theme toggle */}
            <button
              onClick={toggleTheme}
              className="p-2 rounded-xl border border-slate-200 dark:border-slate-800 bg-white/80 dark:bg-slate-900/80 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors shadow-xs cursor-pointer"
              title={theme === 'dark' ? t('nav.lightMode') : t('nav.darkMode')}
            >
              <i className={`pi ${theme === 'dark' ? 'pi-sun text-amber-400' : 'pi-moon text-indigo-500'} text-xs`} />
            </button>

            {/* Auth Buttons */}
            {isAuthenticated ? (
              <div className="flex items-center gap-2">
                <span className="hidden sm:inline text-xs text-slate-600 dark:text-slate-300">
                  {t('landing.welcomeBack', { name: user?.name?.split(' ')[0] || 'User' })}
                </span>
                <Link
                  to="/dashboard"
                  className="px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-md shadow-indigo-500/20 transition-all flex items-center gap-2 cursor-pointer"
                >
                  <span>{t('landing.ctaOpenWorkspace')}</span>
                  <i className="pi pi-arrow-right text-[10px]" />
                </Link>
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <Link
                  to="/auth/login"
                  className="px-3 py-1.5 rounded-xl text-xs font-semibold text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
                >
                  {t('nav.signIn')}
                </Link>
                <Link
                  to="/auth/login"
                  className="px-3.5 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-md shadow-indigo-500/20 transition-all flex items-center gap-1.5 cursor-pointer"
                >
                  <span>{t('landing.ctaGetStarted')}</span>
                  <i className="pi pi-sparkles text-[10px]" />
                </Link>
              </div>
            )}
          </div>
        </div>
      </header>

      {/* =========================================================================
          2. HERO SECTION
          ========================================================================= */}
      <section className="relative z-10 pt-16 sm:pt-24 pb-16 px-4 sm:px-6 max-w-7xl mx-auto text-center">
        {/* Eyebrow Pill */}
        <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-indigo-200 dark:border-indigo-500/30 bg-indigo-50/70 dark:bg-indigo-950/40 text-indigo-600 dark:text-indigo-400 text-xs font-medium mb-6 shadow-xs backdrop-blur-md">
          <span className="w-2 h-2 rounded-full bg-emerald-500 animate-ping" />
          <span>{t('landing.badge')}</span>
        </div>

        {/* Hero Title */}
        <h1 className="text-4xl sm:text-6xl lg:text-7xl font-extrabold tracking-tight max-w-4xl mx-auto leading-tight sm:leading-none mb-6">
          {t('landing.heroTitle1')}{' '}
          <span className="bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 bg-clip-text text-transparent">
            {t('landing.heroTitle2')}
          </span>
        </h1>

        {/* Hero Subtitle */}
        <p className="text-base sm:text-lg text-slate-600 dark:text-slate-300 max-w-2xl mx-auto mb-10 leading-relaxed">
          {t('landing.heroSubtitle')}
        </p>

        {/* Hero CTA Button Row */}
        <div className="flex flex-wrap items-center justify-center gap-3 sm:gap-4 mb-16">
          {isAuthenticated ? (
            <Link
              to="/dashboard"
              className="px-6 py-3.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-sm shadow-xl shadow-indigo-500/25 transition-all flex items-center gap-2.5 cursor-pointer group"
            >
              <span>{t('landing.ctaOpenWorkspace')}</span>
              <i className="pi pi-arrow-right text-xs group-hover:translate-x-0.5 transition-transform" />
            </Link>
          ) : (
            <Link
              to="/auth/login"
              className="px-6 py-3.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-sm shadow-xl shadow-indigo-500/25 transition-all flex items-center gap-2.5 cursor-pointer group"
            >
              <span>{t('landing.ctaGetStarted')}</span>
              <i className="pi pi-sparkles text-xs group-hover:rotate-12 transition-transform" />
            </Link>
          )}

          {/* Instant Sandbox Button */}
          <button
            onClick={handleDemoLogin}
            disabled={isDemoLoading}
            className="px-5 py-3.5 rounded-xl border border-slate-200 dark:border-slate-800 bg-white/80 dark:bg-slate-900/80 hover:bg-slate-100 dark:hover:bg-slate-800/90 text-slate-800 dark:text-slate-200 font-semibold text-sm shadow-sm transition-all flex items-center gap-2 cursor-pointer disabled:opacity-50"
          >
            {isDemoLoading ? (
              <i className="pi pi-spin pi-spinner text-xs text-indigo-500" />
            ) : (
              <i className="pi pi-play text-xs text-emerald-500" />
            )}
            <span>{t('landing.ctaInstantDemo')}</span>
          </button>

          {/* Scroll to features */}
          <a
            href="#features"
            className="px-5 py-3.5 rounded-xl text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 text-sm font-medium transition-colors flex items-center gap-1.5"
          >
            <span>{t('landing.ctaDocs')}</span>
            <i className="pi pi-chevron-down text-xs" />
          </a>
        </div>

        {/* =====================================================================
            HERO APP MOCKUP PREVIEW (State-of-the-Art Interactive Window)
            ===================================================================== */}
        <div className="relative max-w-5xl mx-auto rounded-2xl border border-slate-200/80 dark:border-slate-800/80 bg-white/90 dark:bg-slate-900/90 shadow-2xl backdrop-blur-xl overflow-hidden text-left animate-fade-in">
          {/* Mock Window Top Bar */}
          <div className="h-10 px-4 bg-slate-100/80 dark:bg-slate-950/80 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="w-3 h-3 rounded-full bg-rose-500/80" />
              <span className="w-3 h-3 rounded-full bg-amber-500/80" />
              <span className="w-3 h-3 rounded-full bg-emerald-500/80" />
            </div>
            <div className="text-[11px] font-mono text-slate-400 dark:text-slate-500 flex items-center gap-1.5">
              <i className="pi pi-lock text-[9px]" />
              <span>taskhub.eka-dev.cloud/dashboard</span>
            </div>
            <div className="flex items-center gap-2 text-slate-400 dark:text-slate-600 text-xs">
              <i className="pi pi-minus" />
              <i className="pi pi-clone" />
              <i className="pi pi-times" />
            </div>
          </div>

          {/* Mock Workspace Content Grid */}
          <div className="p-4 sm:p-6 grid grid-cols-1 md:grid-cols-12 gap-4 bg-slate-50/50 dark:bg-slate-950/50">
            {/* Left Mock Sidebar */}
            <div className="hidden md:block md:col-span-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/80 p-3 space-y-1.5">
              <div className="text-[10px] font-mono uppercase tracking-wider text-slate-400 dark:text-slate-500 px-2 mb-2">
                Workspace
              </div>
              <div className="flex items-center justify-between px-2.5 py-1.5 rounded-lg bg-indigo-50 dark:bg-indigo-600/20 text-indigo-600 dark:text-indigo-300 text-xs font-semibold border border-indigo-200 dark:border-indigo-500/30">
                <span className="flex items-center gap-2">
                  <i className="pi pi-th-large text-xs" />
                  <span>Dashboard</span>
                </span>
                <span className="text-[10px] bg-indigo-200/50 dark:bg-indigo-500/30 px-1.5 py-0.5 rounded-full font-mono">
                  12
                </span>
              </div>
              <div className="flex items-center justify-between px-2.5 py-1.5 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 text-xs font-medium transition-colors">
                <span className="flex items-center gap-2">
                  <i className="pi pi-check-square text-xs" />
                  <span>Tasks Board</span>
                </span>
                <span className="text-[10px] text-slate-400 font-mono">8</span>
              </div>
              <div className="flex items-center justify-between px-2.5 py-1.5 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 text-xs font-medium transition-colors">
                <span className="flex items-center gap-2">
                  <i className="pi pi-file-edit text-xs" />
                  <span>Scratchpad</span>
                </span>
                <span className="text-[10px] text-slate-400 font-mono">4</span>
              </div>
              <div className="flex items-center justify-between px-2.5 py-1.5 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 text-xs font-medium transition-colors">
                <span className="flex items-center gap-2">
                  <i className="pi pi-microchip text-xs text-purple-400" />
                  <span>MCP & AI Agent</span>
                </span>
                <span className="w-2 h-2 rounded-full bg-emerald-500" />
              </div>
            </div>

            {/* Center Mock Kanban Board */}
            <div className="col-span-1 md:col-span-6 space-y-3">
              <div className="flex items-center justify-between px-1">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-bold text-slate-800 dark:text-slate-200">Active Sprint Tasks</span>
                  <span className="text-[10px] bg-slate-200 dark:bg-slate-800 px-2 py-0.5 rounded-md text-slate-500 font-mono">
                    3 In Progress
                  </span>
                </div>
                <div className="flex items-center gap-1.5 text-xs text-indigo-500 font-medium">
                  <i className="pi pi-filter text-[10px]" />
                  <span className="text-[11px]">Horizon Focus</span>
                </div>
              </div>

              {/* Task 1 */}
              <div className="rounded-xl border border-indigo-200 dark:border-indigo-500/30 bg-white dark:bg-slate-900 p-3.5 shadow-sm space-y-2">
                <div className="flex items-center justify-between text-xs">
                  <span className="px-2 py-0.5 rounded-md bg-rose-50 dark:bg-rose-950/50 text-rose-600 dark:text-rose-400 text-[10px] font-bold uppercase tracking-wider">
                    Urgent
                  </span>
                  <span className="text-[10px] font-mono text-slate-400">#TSK-104</span>
                </div>
                <h4 className="text-xs font-semibold text-slate-900 dark:text-slate-100">
                  Connect Cursor MCP stdio bridge for realtime status sync
                </h4>
                <div className="flex items-center gap-2 text-[10px] text-slate-500">
                  <span className="bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded font-mono">#mcp</span>
                  <span className="bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded font-mono">#ai</span>
                  <span className="ml-auto text-emerald-500 font-medium flex items-center gap-1">
                    <i className="pi pi-check text-[9px]" /> 3/3 subtasks
                  </span>
                </div>
                {/* Progress bar */}
                <div className="w-full bg-slate-100 dark:bg-slate-800 h-1.5 rounded-full overflow-hidden">
                  <div className="bg-emerald-500 h-full w-full rounded-full" />
                </div>
              </div>

              {/* Task 2 */}
              <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-3.5 shadow-sm space-y-2">
                <div className="flex items-center justify-between text-xs">
                  <span className="px-2 py-0.5 rounded-md bg-amber-50 dark:bg-amber-950/50 text-amber-600 dark:text-amber-400 text-[10px] font-bold uppercase tracking-wider">
                    High
                  </span>
                  <span className="text-[10px] font-mono text-slate-400">#TSK-108</span>
                </div>
                <h4 className="text-xs font-semibold text-slate-900 dark:text-slate-100">
                  Implement Apple Sign-In JWKS cryptographic validation
                </h4>
                <div className="flex items-center gap-2 text-[10px] text-slate-500">
                  <span className="bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded font-mono">#auth</span>
                  <span className="bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded font-mono">#security</span>
                  <span className="ml-auto text-indigo-500 font-medium">80%</span>
                </div>
                {/* Progress bar */}
                <div className="w-full bg-slate-100 dark:bg-slate-800 h-1.5 rounded-full overflow-hidden">
                  <div className="bg-indigo-500 h-full w-4/5 rounded-full" />
                </div>
              </div>
            </div>

            {/* Right Mock Scratchpad Note */}
            <div className="hidden md:block md:col-span-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/80 p-3.5 space-y-2.5">
              <div className="flex items-center justify-between">
                <div className="text-[10px] font-mono uppercase tracking-wider text-slate-400 dark:text-slate-500">
                  Scratchpad Note
                </div>
                <span className="text-[10px] text-amber-500 font-mono flex items-center gap-1">
                  <i className="pi pi-bookmark-fill text-[9px]" /> Pinned
                </span>
              </div>
              <div className="text-xs font-bold text-slate-800 dark:text-slate-200">
                Architecture Roadmap
              </div>
              <div className="text-[11px] font-mono text-slate-500 dark:text-slate-400 space-y-1.5 bg-slate-50 dark:bg-slate-950 p-2.5 rounded-lg border border-slate-200/50 dark:border-slate-800/50">
                <p className="text-indigo-400 font-semibold">## AI Tools</p>
                <p>- [x] TaskHub MCP stdio</p>
                <p>- [x] Redis PubSub bridge</p>
                <p>- [ ] Automated changelog</p>
              </div>
              <button
                onClick={handleDemoLogin}
                className="w-full py-1.5 rounded-lg border border-indigo-200 dark:border-indigo-500/30 bg-indigo-50/50 dark:bg-indigo-950/30 text-indigo-600 dark:text-indigo-300 text-[10px] font-medium flex items-center justify-center gap-1.5 hover:bg-indigo-100 dark:hover:bg-indigo-950/60 transition-colors cursor-pointer"
              >
                <i className="pi pi-arrow-right-arrow-left text-[9px]" />
                <span>1-Click Convert to Task</span>
              </button>
            </div>
          </div>

          {/* Floating Live Badge Overlays */}
          <div className="hidden sm:flex absolute bottom-5 left-8 px-3 py-2 rounded-xl bg-slate-900/90 dark:bg-slate-800/90 border border-slate-700 text-white text-xs font-medium items-center gap-2 shadow-xl backdrop-blur-md">
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-ping" />
            <span>🤖 MCP Server: Connected to Cursor Agent</span>
          </div>

          <div className="hidden sm:flex absolute bottom-5 right-8 px-3 py-2 rounded-xl bg-slate-900/90 dark:bg-slate-800/90 border border-slate-700 text-white text-xs font-medium items-center gap-2 shadow-xl backdrop-blur-md">
            <i className="pi pi-bolt text-amber-400 text-xs" />
            <span>Sub-10ms Local-First Response</span>
          </div>
        </div>
      </section>

      {/* =========================================================================
          3. STATS STRIP
          ========================================================================= */}
      <section className="relative z-10 py-12 border-y border-slate-200/80 dark:border-slate-800/80 bg-slate-100/50 dark:bg-slate-900/30">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6 text-center">
            <div className="space-y-1">
              <div className="text-2xl sm:text-3xl font-extrabold text-indigo-600 dark:text-indigo-400 font-mono">
                {t('landing.stats.stat1Value')}
              </div>
              <div className="text-xs text-slate-500 dark:text-slate-400 font-medium">
                {t('landing.stats.stat1Label')}
              </div>
            </div>
            <div className="space-y-1">
              <div className="text-2xl sm:text-3xl font-extrabold text-purple-600 dark:text-purple-400 font-mono">
                {t('landing.stats.stat2Value')}
              </div>
              <div className="text-xs text-slate-500 dark:text-slate-400 font-medium">
                {t('landing.stats.stat2Label')}
              </div>
            </div>
            <div className="space-y-1">
              <div className="text-2xl sm:text-3xl font-extrabold text-sky-600 dark:text-sky-400 font-mono">
                {t('landing.stats.stat3Value')}
              </div>
              <div className="text-xs text-slate-500 dark:text-slate-400 font-medium">
                {t('landing.stats.stat3Label')}
              </div>
            </div>
            <div className="space-y-1">
              <div className="text-2xl sm:text-3xl font-extrabold text-emerald-600 dark:text-emerald-400 font-mono">
                {t('landing.stats.stat4Value')}
              </div>
              <div className="text-xs text-slate-500 dark:text-slate-400 font-medium">
                {t('landing.stats.stat4Label')}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* =========================================================================
          4. BENTO GRID FEATURES SECTION
          ========================================================================= */}
      <section id="features" className="py-20 px-4 sm:px-6 max-w-7xl mx-auto">
        <div className="text-center max-w-3xl mx-auto mb-16">
          <div className="text-xs font-mono uppercase tracking-wider text-indigo-500 font-bold mb-2">
            Features
          </div>
          <h2 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-slate-900 dark:text-slate-100 mb-4">
            {t('landing.features.sectionTitle')}
          </h2>
          <p className="text-sm sm:text-base text-slate-600 dark:text-slate-400">
            {t('landing.features.sectionSubtitle')}
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {/* Card 1: Kanban (Large Bento) */}
          <div
            id="kanban"
            className="md:col-span-2 rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-6 sm:p-8 shadow-sm hover:border-indigo-400/50 transition-all space-y-4"
          >
            <div className="w-10 h-10 rounded-xl bg-indigo-50 dark:bg-indigo-950/50 border border-indigo-200 dark:border-indigo-500/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400 text-lg">
              <i className="pi pi-th-large" />
            </div>
            <h3 className="text-xl font-bold text-slate-900 dark:text-slate-100">
              {t('landing.features.kanbanTitle')}
            </h3>
            <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
              {t('landing.features.kanbanDesc')}
            </p>
            <div className="pt-2 flex flex-wrap gap-2 text-xs font-mono">
              <span className="px-2.5 py-1 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300">
                Horizon Focus Mode
              </span>
              <span className="px-2.5 py-1 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300">
                Subtasks Progress %
              </span>
              <span className="px-2.5 py-1 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300">
                Progress Notes Log
              </span>
            </div>
          </div>

          {/* Card 2: MCP Protocol */}
          <div className="rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-6 sm:p-8 shadow-sm hover:border-purple-400/50 transition-all space-y-4">
            <div className="w-10 h-10 rounded-xl bg-purple-50 dark:bg-purple-950/50 border border-purple-200 dark:border-purple-500/30 flex items-center justify-center text-purple-600 dark:text-purple-400 text-lg">
              <i className="pi pi-microchip" />
            </div>
            <h3 className="text-xl font-bold text-slate-900 dark:text-slate-100">
              {t('landing.features.mcpTitle')}
            </h3>
            <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
              {t('landing.features.mcpDesc')}
            </p>
            <div className="pt-2 flex flex-wrap gap-2 text-xs font-mono">
              <span className="px-2.5 py-1 rounded-md bg-purple-100/60 dark:bg-purple-950/60 text-purple-700 dark:text-purple-300">
                Cursor & AGY Ready
              </span>
            </div>
          </div>

          {/* Card 3: Notes & Scratchpad */}
          <div className="rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-6 sm:p-8 shadow-sm hover:border-sky-400/50 transition-all space-y-4">
            <div className="w-10 h-10 rounded-xl bg-sky-50 dark:bg-sky-950/50 border border-sky-200 dark:border-sky-500/30 flex items-center justify-center text-sky-600 dark:text-sky-400 text-lg">
              <i className="pi pi-file-edit" />
            </div>
            <h3 className="text-xl font-bold text-slate-900 dark:text-slate-100">
              {t('landing.features.notesTitle')}
            </h3>
            <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
              {t('landing.features.notesDesc')}
            </p>
          </div>

          {/* Card 4: Command Palette */}
          <div className="rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-6 sm:p-8 shadow-sm hover:border-amber-400/50 transition-all space-y-4">
            <div className="w-10 h-10 rounded-xl bg-amber-50 dark:bg-amber-950/50 border border-amber-200 dark:border-amber-500/30 flex items-center justify-center text-amber-600 dark:text-amber-400 text-lg">
              <i className="pi pi-search" />
            </div>
            <h3 className="text-xl font-bold text-slate-900 dark:text-slate-100">
              {t('landing.features.commandTitle')}
            </h3>
            <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
              {t('landing.features.commandDesc')}
            </p>
          </div>

          {/* Card 5: Extreme Speed */}
          <div className="rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-6 sm:p-8 shadow-sm hover:border-emerald-400/50 transition-all space-y-4">
            <div className="w-10 h-10 rounded-xl bg-emerald-50 dark:bg-emerald-950/50 border border-emerald-200 dark:border-emerald-500/30 flex items-center justify-center text-emerald-600 dark:text-emerald-400 text-lg">
              <i className="pi pi-bolt" />
            </div>
            <h3 className="text-xl font-bold text-slate-900 dark:text-slate-100">
              {t('landing.features.speedTitle')}
            </h3>
            <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
              {t('landing.features.speedDesc')}
            </p>
          </div>
        </div>
      </section>

      {/* =========================================================================
          5. MCP SHOWCASE SECTION
          ========================================================================= */}
      <section id="mcp" className="py-16 px-4 sm:px-6 max-w-7xl mx-auto">
        <div className="rounded-3xl border border-indigo-200 dark:border-indigo-500/30 bg-gradient-to-b from-indigo-50/50 to-white dark:from-slate-900/90 dark:to-slate-950 p-6 sm:p-12 shadow-xl relative overflow-hidden">
          <div className="max-w-3xl mb-8 space-y-3">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-indigo-100 dark:bg-indigo-950/80 text-indigo-600 dark:text-indigo-400 text-xs font-mono font-bold">
              <i className="pi pi-code" />
              <span>Model Context Protocol</span>
            </div>
            <h2 className="text-2xl sm:text-3xl font-extrabold text-slate-900 dark:text-slate-100">
              {t('landing.mcpShowcase.title')}
            </h2>
            <p className="text-sm text-slate-600 dark:text-slate-300">
              {t('landing.mcpShowcase.subtitle')}
            </p>
          </div>

          {/* Interactive Code Window */}
          <div className="rounded-2xl border border-slate-300 dark:border-slate-800 bg-slate-950 text-slate-100 p-4 sm:p-6 shadow-2xl relative">
            <div className="flex items-center justify-between pb-3 mb-3 border-b border-slate-800 text-xs text-slate-400 font-mono">
              <div className="flex items-center gap-2">
                <span className="w-2.5 h-2.5 rounded-full bg-rose-500/80" />
                <span className="w-2.5 h-2.5 rounded-full bg-amber-500/80" />
                <span className="w-2.5 h-2.5 rounded-full bg-emerald-500/80" />
                <span className="ml-2">cursor_settings.json / agy_config.json</span>
              </div>
              <button
                onClick={handleCopyMcp}
                className="px-3 py-1 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-[11px] font-semibold flex items-center gap-1.5 transition-colors cursor-pointer"
              >
                <i className={`pi ${copiedMcp ? 'pi-check' : 'pi-copy'} text-[10px]`} />
                <span>{copiedMcp ? t('landing.mcpShowcase.copied') : t('landing.mcpShowcase.copyButton')}</span>
              </button>
            </div>
            <pre className="text-xs sm:text-sm font-mono overflow-x-auto text-emerald-400 leading-relaxed">
              <code>{mcpConfigCode}</code>
            </pre>
          </div>
        </div>
      </section>

      {/* =========================================================================
          6. COMPARISON SECTION
          ========================================================================= */}
      <section id="comparison" className="py-20 px-4 sm:px-6 max-w-7xl mx-auto">
        <div className="text-center max-w-3xl mx-auto mb-16">
          <h2 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-slate-900 dark:text-slate-100 mb-4">
            {t('landing.comparison.title')}
          </h2>
          <p className="text-sm sm:text-base text-slate-600 dark:text-slate-400">
            {t('landing.comparison.subtitle')}
          </p>
        </div>

        <div className="overflow-x-auto rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-md">
          <table className="w-full text-left text-xs sm:text-sm">
            <thead className="bg-slate-100 dark:bg-slate-950/80 border-b border-slate-200 dark:border-slate-800 text-slate-600 dark:text-slate-400 font-semibold uppercase font-mono tracking-wider text-[11px]">
              <tr>
                <th className="p-4 sm:p-5">{t('landing.comparison.featureCol')}</th>
                <th className="p-4 sm:p-5 text-slate-400">{t('landing.comparison.traditionalCol')}</th>
                <th className="p-4 sm:p-5 text-indigo-600 dark:text-indigo-400">{t('landing.comparison.taskhubCol')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800/80">
              <tr>
                <td className="p-4 sm:p-5 font-medium text-slate-900 dark:text-slate-100">
                  {t('landing.comparison.row1Feature')}
                </td>
                <td className="p-4 sm:p-5 text-slate-500">{t('landing.comparison.row1Trad')}</td>
                <td className="p-4 sm:p-5 font-semibold text-emerald-600 dark:text-emerald-400 flex items-center gap-2">
                  <i className="pi pi-check text-xs" />
                  <span>{t('landing.comparison.row1Hub')}</span>
                </td>
              </tr>
              <tr>
                <td className="p-4 sm:p-5 font-medium text-slate-900 dark:text-slate-100">
                  {t('landing.comparison.row2Feature')}
                </td>
                <td className="p-4 sm:p-5 text-slate-500">{t('landing.comparison.row2Trad')}</td>
                <td className="p-4 sm:p-5 font-semibold text-emerald-600 dark:text-emerald-400 flex items-center gap-2">
                  <i className="pi pi-check text-xs" />
                  <span>{t('landing.comparison.row2Hub')}</span>
                </td>
              </tr>
              <tr>
                <td className="p-4 sm:p-5 font-medium text-slate-900 dark:text-slate-100">
                  {t('landing.comparison.row3Feature')}
                </td>
                <td className="p-4 sm:p-5 text-slate-500">{t('landing.comparison.row3Trad')}</td>
                <td className="p-4 sm:p-5 font-semibold text-emerald-600 dark:text-emerald-400 flex items-center gap-2">
                  <i className="pi pi-check text-xs" />
                  <span>{t('landing.comparison.row3Hub')}</span>
                </td>
              </tr>
              <tr>
                <td className="p-4 sm:p-5 font-medium text-slate-900 dark:text-slate-100">
                  {t('landing.comparison.row4Feature')}
                </td>
                <td className="p-4 sm:p-5 text-slate-500">{t('landing.comparison.row4Trad')}</td>
                <td className="p-4 sm:p-5 font-semibold text-emerald-600 dark:text-emerald-400 flex items-center gap-2">
                  <i className="pi pi-check text-xs" />
                  <span>{t('landing.comparison.row4Hub')}</span>
                </td>
              </tr>
              <tr>
                <td className="p-4 sm:p-5 font-medium text-slate-900 dark:text-slate-100">
                  {t('landing.comparison.row5Feature')}
                </td>
                <td className="p-4 sm:p-5 text-slate-500">{t('landing.comparison.row5Trad')}</td>
                <td className="p-4 sm:p-5 font-semibold text-emerald-600 dark:text-emerald-400 flex items-center gap-2">
                  <i className="pi pi-check text-xs" />
                  <span>{t('landing.comparison.row5Hub')}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* =========================================================================
          7. BOTTOM CTA BANNER
          ========================================================================= */}
      <section className="py-16 px-4 sm:px-6 max-w-7xl mx-auto">
        <div className="rounded-3xl bg-gradient-to-r from-indigo-600 via-purple-600 to-indigo-700 p-8 sm:p-14 text-center text-white shadow-2xl relative overflow-hidden">
          <div className="relative z-10 max-w-2xl mx-auto space-y-4">
            <h2 className="text-3xl sm:text-4xl font-extrabold tracking-tight">
              {t('landing.ctaBanner.title')}
            </h2>
            <p className="text-sm sm:text-base text-indigo-100 leading-relaxed">
              {t('landing.ctaBanner.subtitle')}
            </p>
            <div className="pt-4 flex flex-wrap items-center justify-center gap-3">
              <Link
                to="/auth/login"
                className="px-6 py-3.5 rounded-xl bg-white hover:bg-slate-100 text-indigo-700 font-bold text-xs sm:text-sm shadow-lg transition-all flex items-center gap-2 cursor-pointer"
              >
                <span>{t('landing.ctaBanner.button')}</span>
                <i className="pi pi-arrow-right text-xs" />
              </Link>
              <button
                onClick={handleDemoLogin}
                disabled={isDemoLoading}
                className="px-6 py-3.5 rounded-xl border border-white/30 hover:bg-white/10 text-white font-semibold text-xs sm:text-sm transition-all flex items-center gap-2 cursor-pointer"
              >
                <i className="pi pi-play text-xs text-amber-300" />
                <span>{t('landing.ctaBanner.demoButton')}</span>
              </button>
            </div>
          </div>
        </div>
      </section>

      {/* =========================================================================
          8. FOOTER
          ========================================================================= */}
      <footer className="border-t border-slate-200 dark:border-slate-800 py-12 px-4 sm:px-6 max-w-7xl mx-auto">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-6">
          <div className="flex items-center gap-2.5">
            <BrandLogo size={24} />
            <span className="font-bold text-sm tracking-tight text-slate-900 dark:text-slate-100">
              Task<span className="text-indigo-600 dark:text-indigo-400">Hub</span>
            </span>
            <span className="text-xs text-slate-400 ml-2">
              — {t('landing.footer.tagline')}
            </span>
          </div>

          <div className="flex items-center gap-4 text-xs text-slate-500">
            <span className="flex items-center gap-1.5 text-emerald-500 font-medium">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
              <span>{t('landing.footer.status')}</span>
            </span>
            <span>•</span>
            <span>{t('landing.footer.copyright', { year: new Date().getFullYear().toString() })}</span>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default LandingPage
