import { Button } from 'primereact/button'
import React, { useState } from 'react'
import { isRouteErrorResponse, useNavigate, useRouteError } from 'react-router-dom'

import { BrandLogo } from '@/components'
import { useTranslation } from '@/i18n'

interface ErrorPageProps {
  error?: Error | unknown
  resetErrorBoundary?: () => void
}

export const ErrorPage: React.FC<ErrorPageProps> = ({ error: propError, resetErrorBoundary }) => {
  const routeError = useRouteError()
  const navigate = useNavigate()
  const { t, locale } = useTranslation()
  const [showDetails, setShowDetails] = useState(false)
  const [copied, setCopied] = useState(false)

  // Prioritize error passed via prop (e.g. from ErrorBoundary), otherwise routeError
  const activeError = propError ?? routeError

  // Determine error characteristics
  let status = 500
  let statusText = 'Internal Application Error'
  let errorMessage = 'An unexpected runtime error occurred.'
  let errorStack: string | undefined

  if (isRouteErrorResponse(activeError)) {
    const { status: routeStatus, statusText: routeText, data: routeData } = activeError
    status = routeStatus
    statusText = routeText || (status === 404 ? 'Page Not Found' : 'Routing Error')
    errorMessage =
      typeof routeData === 'string'
        ? routeData
        : (routeData?.message ?? routeText ?? 'Route resolution failed.')
  } else if (activeError instanceof Error) {
    errorMessage = activeError.message
    errorStack = activeError.stack
    statusText = activeError.name || 'TypeError'
  } else if (typeof activeError === 'string') {
    errorMessage = activeError
  } else if (activeError && typeof activeError === 'object') {
    try {
      errorMessage = JSON.stringify(activeError, null, 2)
    } catch {
      errorMessage = String(activeError)
    }
  }

  const is404 = status === 404

  const handleReload = () => {
    if (resetErrorBoundary) {
      resetErrorBoundary()
    } else {
      window.location.reload()
    }
  }

  const handleGoHome = () => {
    if (resetErrorBoundary) {
      resetErrorBoundary()
    }
    navigate('/')
  }

  const handleClearCacheAndReload = () => {
    try {
      // Clear app caches that might cause stale state
      localStorage.removeItem('local_tasks')
      localStorage.removeItem('local_notes')
      Object.keys(localStorage)
        .filter((k) => k.startsWith('cache_'))
        .forEach((k) => localStorage.removeItem(k))
    } catch {
      // ignore
    }
    window.location.href = '/'
  }

  const handleCopyDetails = async () => {
    const diagnosticReport = [
      `TaskHub Error Diagnostics Report`,
      `=================================`,
      `Timestamp: ${new Date().toISOString()}`,
      `URL: ${window.location.href}`,
      `Locale: ${locale}`,
      `Status: ${status} (${statusText})`,
      `Message: ${errorMessage}`,
      errorStack ? `\nStack Trace:\n${errorStack}` : '',
    ]
      .filter(Boolean)
      .join('\n')

    try {
      await navigator.clipboard.writeText(diagnosticReport)
      setCopied(true)
      setTimeout(() => setCopied(false), 2500)
    } catch {
      // fallback
    }
  }

  return (
    <div className="min-h-screen flex flex-col justify-between bg-slate-950 text-slate-100 p-4 sm:p-6 font-sans relative overflow-hidden selection:bg-rose-500/30 selection:text-rose-200">
      {/* Background ambient gradient glow */}
      <div className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[350px] bg-gradient-to-tr from-rose-600/15 via-indigo-600/15 to-transparent blur-[120px] pointer-events-none rounded-full" />

      {/* Top Navbar Header */}
      <header className="relative z-10 w-full max-w-4xl mx-auto flex items-center justify-between py-2 border-b border-slate-800/80">
        <div className="flex items-center gap-2.5">
          <BrandLogo size={26} />
          <span className="font-bold text-sm tracking-tight text-white">
            Task<span className="text-indigo-400">Hub</span>
          </span>
          <span className="text-[10px] font-mono px-2 py-0.5 rounded-full bg-rose-500/10 text-rose-400 border border-rose-500/20">
            {is404 ? '404' : 'Crash Guard'}
          </span>
        </div>

        <button
          id="btn-error-header-home"
          onClick={handleGoHome}
          className="text-xs text-slate-400 hover:text-slate-200 transition-colors flex items-center gap-1.5 cursor-pointer"
        >
          <i className="pi pi-home text-[11px]" />
          <span>Dashboard</span>
        </button>
      </header>

      {/* Main Content Card */}
      <main className="relative z-10 w-full max-w-2xl mx-auto my-auto py-10">
        <div className="rounded-2xl border border-slate-800/90 bg-slate-900/70 backdrop-blur-xl p-6 sm:p-8 shadow-2xl space-y-6">
          {/* Badge & Icon Header */}
          <div className="flex items-start justify-between gap-4">
            <div className="space-y-2">
              <div className="inline-flex items-center gap-2 px-2.5 py-1 rounded-full text-xs font-medium bg-rose-500/15 text-rose-300 border border-rose-500/30">
                <i className={`pi ${is404 ? 'pi-compass' : 'pi-exclamation-triangle'} text-[11px]`} />
                <span>{is404 ? '404 Not Found' : (statusText || 'Unexpected Runtime Error')}</span>
              </div>
              <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-white">
                {is404 ? 'Page Not Found' : t('errors.pageTitle')}
              </h1>
              <p className="text-xs sm:text-sm text-slate-400 leading-relaxed">
                {is404
                  ? "We couldn't find the route you are trying to visit."
                  : t('errors.pageSubtitle')}
              </p>
            </div>

            <div className="shrink-0 h-12 w-12 rounded-xl bg-gradient-to-br from-rose-500/20 to-indigo-500/20 border border-rose-500/30 flex items-center justify-center text-rose-400">
              <i className={`pi ${is404 ? 'pi-question-circle' : 'pi-shield'} text-2xl`} />
            </div>
          </div>

          {/* Error Message Box */}
          <div className="rounded-xl border border-rose-950/60 bg-rose-950/20 p-3.5 flex items-start gap-3">
            <i className="pi pi-info-circle text-rose-400 text-sm mt-0.5 shrink-0" />
            <div className="min-w-0 flex-1">
              <p className="text-xs font-mono text-rose-200 break-words font-medium">
                {errorMessage}
              </p>
              <p className="text-[10px] text-slate-400 mt-1">
                Safe recovery mechanisms are active. Your server data is secure.
              </p>
            </div>
          </div>

          {/* Action Recovery Buttons */}
          <div className="flex flex-wrap items-center gap-2.5 pt-1">
            <Button
              id="btn-error-reload"
              label={t('errors.reloadPage')}
              icon="pi pi-refresh"
              onClick={handleReload}
              className="bg-indigo-600 hover:bg-indigo-500 text-white border-none text-xs px-4 py-2 rounded-xl font-medium shadow-md shadow-indigo-600/20 cursor-pointer"
            />
            <Button
              id="btn-error-home"
              label={t('errors.backHome')}
              icon="pi pi-home"
              onClick={handleGoHome}
              className="bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700/80 text-xs px-4 py-2 rounded-xl font-medium cursor-pointer"
            />
            <Button
              id="btn-error-clear-cache"
              label={t('errors.clearCacheReload')}
              icon="pi pi-trash"
              onClick={handleClearCacheAndReload}
              className="bg-rose-950/40 hover:bg-rose-900/50 text-rose-300 border border-rose-800/50 text-xs px-3.5 py-2 rounded-xl font-medium cursor-pointer"
            />
          </div>

          {/* Technical Details Accordion */}
          <div className="pt-2 border-t border-slate-800/80">
            <div className="flex items-center justify-between">
              <button
                id="btn-error-details-toggle"
                onClick={() => setShowDetails((prev) => !prev)}
                className="text-xs text-slate-400 hover:text-slate-200 flex items-center gap-1.5 font-medium transition-colors cursor-pointer"
              >
                <i className={`pi ${showDetails ? 'pi-chevron-down' : 'pi-chevron-right'} text-[10px]`} />
                <span>{t('errors.technicalDetails')}</span>
              </button>

              {showDetails && (
                <button
                  id="btn-error-copy"
                  onClick={handleCopyDetails}
                  className="text-xs text-indigo-400 hover:text-indigo-300 flex items-center gap-1 transition-colors cursor-pointer"
                >
                  <i className={`pi ${copied ? 'pi-check text-emerald-400' : 'pi-copy'} text-[11px]`} />
                  <span>{copied ? t('errors.errorCopied') : t('errors.copyError')}</span>
                </button>
              )}
            </div>

            {showDetails && (
              <div className="mt-3 rounded-xl border border-slate-800 bg-slate-950 p-3.5 overflow-x-auto text-[11px] font-mono text-slate-300 space-y-2 animate-fade-in max-h-60 overflow-y-auto">
                <div className="text-slate-400">
                  <span className="text-slate-400">URL:</span> {window.location.href}
                </div>
                <div className="text-slate-400">
                  <span className="text-slate-400">Error Name:</span> {statusText}
                </div>
                <div className="text-slate-400">
                  <span className="text-slate-400">Message:</span> {errorMessage}
                </div>
                {errorStack && (
                  <pre className="text-slate-400 text-[10px] whitespace-pre-wrap leading-relaxed border-t border-slate-900 pt-2 mt-2">
                    {errorStack}
                  </pre>
                )}
              </div>
            )}
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="relative z-10 w-full max-w-4xl mx-auto py-3 text-center text-[11px] text-slate-400 border-t border-slate-800/50">
        TaskHub Resilience Shield &bull; Error Boundary &bull; Protected Session
      </footer>
    </div>
  )
}

export default ErrorPage
