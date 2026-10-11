import { ConfirmDialog } from 'primereact/confirmdialog'
import { Tooltip } from 'primereact/tooltip'
import { useEffect } from 'react'
import { Outlet, useLocation } from 'react-router-dom'

import { AppToastContainer, ErrorBoundary } from '@/components'
import { useTranslation } from '@/i18n'

function App() {
  const location = useLocation()
  const { t, locale } = useTranslation()

  // Keep <html lang="..."> and document.title synchronized with route & active locale
  useEffect(() => {
    document.documentElement.lang = locale

    let pageTitle = ''
    const path = location.pathname

    if (path === '/') {
      pageTitle = ''
    } else if (path === '/dashboard') {
      pageTitle = t('nav.dashboard')
    } else if (path.startsWith('/tasks')) {
      pageTitle = t('nav.tasks')
    } else if (path.startsWith('/projects')) {
      pageTitle = t('nav.projects')
    } else if (path.startsWith('/reminders')) {
      pageTitle = t('nav.reminders')
    } else if (path.startsWith('/notes')) {
      pageTitle = t('nav.notes')
    } else if (path.startsWith('/mcp-settings')) {
      pageTitle = t('nav.mcp')
    } else if (path.startsWith('/auth/login') || path === '/login') {
      pageTitle = t('nav.signIn')
    }

    if (pageTitle) {
      document.title = `${pageTitle} — TaskHub`
    } else {
      document.title = 'TaskHub — Developer Task & MCP AI Workspace'
    }
  }, [location.pathname, locale, t])

  return (
    <>
      <AppToastContainer />
      <ConfirmDialog />
      <Tooltip target="[data-pr-tooltip]" position="top" showDelay={200} />
      <ErrorBoundary>
        <Outlet />
      </ErrorBoundary>
    </>
  )
}

export default App
