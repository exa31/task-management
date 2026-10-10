import { useLocation, useNavigate } from 'react-router-dom'

import { useTranslation } from '@/i18n'
import { useAuthStore, useUIStore } from '@/stores'

interface TopNavProps {
  isSidebarOpen: boolean
  toggleSidebar: () => void
}

export default function TopNav({ isSidebarOpen, toggleSidebar }: TopNavProps) {
  const location = useLocation()
  const navigate = useNavigate()
  const { user, logout } = useAuthStore()
  const { openCommandPalette, theme, toggleTheme } = useUIStore()
  const { t, locale, toggleLocale } = useTranslation()

  const getBreadcrumb = () => {
    const path = location.pathname
    if (path === '/') return t('nav.dashboard')
    if (path.startsWith('/tasks')) return t('nav.tasks')
    if (path.startsWith('/notes')) return t('nav.notes')
    if (path.startsWith('/mcp-settings')) return t('nav.mcp')
    if (path.startsWith('/users')) return t('nav.users')
    return path.substring(1).charAt(0).toUpperCase() + path.substring(2)
  }

  const handleLogout = () => {
    logout()
    navigate('/auth/login')
  }

  return (
    <header className="h-14 flex items-center justify-between px-4 bg-white/80 dark:bg-slate-950/80 backdrop-blur-md border-b border-slate-200 dark:border-slate-800 z-10 sticky top-0 transition-colors">
      <div className="flex items-center gap-3">
        <button
          onClick={toggleSidebar}
          aria-expanded={isSidebarOpen}
          aria-label={isSidebarOpen ? t('nav.collapseSidebar') : t('nav.expandSidebar')}
          className="p-1.5 rounded-lg text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-900 transition-colors cursor-pointer"
        >
          <i className="pi pi-bars text-sm" />
        </button>

        <div className="hidden sm:flex items-center text-xs text-slate-500 dark:text-slate-400">
          <span className="text-slate-400 dark:text-slate-500 font-mono">{t('nav.workspace').toLowerCase()}</span>
          <i className="pi pi-chevron-right text-[9px] mx-2 text-slate-400 dark:text-slate-600" />
          <span className="text-slate-800 dark:text-slate-200 font-medium">{getBreadcrumb()}</span>
        </div>
      </div>

      <div className="flex items-center gap-2">
        {/* Command Bar Trigger Button */}
        <button
          onClick={openCommandPalette}
          className="flex items-center gap-3 px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 hover:border-slate-300 dark:hover:border-slate-700 text-xs text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 transition-colors"
        >
          <div className="flex items-center gap-1.5">
            <i className="pi pi-search text-xs" />
            <span className="hidden sm:inline">{t('nav.searchPlaceholder')}</span>
          </div>
          <kbd className="hidden sm:inline-block px-1.5 py-0.2 rounded text-[10px] font-mono bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 border border-slate-200 dark:border-slate-700">
            ⌘K
          </kbd>
        </button>

        {/* Language Switcher */}
        <button
          onClick={toggleLocale}
          className="px-2 py-1.5 rounded-lg text-xs font-semibold border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-900 text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors flex items-center gap-1.5 shadow-xs cursor-pointer"
          title={`${t('nav.switchLanguage')}: ${locale === 'en' ? t('nav.switchToIndonesian') : t('nav.switchToEnglish')}`}
        >
          <i className="pi pi-globe text-xs text-indigo-500 dark:text-indigo-400" />
          <span className="text-[11px] font-mono font-bold text-indigo-600 dark:text-indigo-400">
            {locale.toUpperCase()}
          </span>
        </button>

        {/* Theme Toggle */}
        <button
          onClick={toggleTheme}
          className="p-1.5 rounded-lg text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-900 transition-colors cursor-pointer"
          title={theme === 'dark' ? t('nav.lightMode') : t('nav.darkMode')}
        >
          <i className={`pi ${theme === 'dark' ? 'pi-sun text-amber-400' : 'pi-moon text-indigo-500'} text-xs`} />
        </button>

        {/* Settings button */}
        <button
          onClick={() => navigate('/settings')}
          className="p-1.5 rounded-lg text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-900 transition-colors cursor-pointer"
          title="Pengaturan"
        >
          <i className="pi pi-cog text-xs" />
        </button>

        <div className="h-4 w-px bg-slate-200 dark:bg-slate-800 mx-1" />

        {/* User Profile */}
        <div className="flex items-center gap-2">
          {user ? (
            <div
              onClick={() => navigate('/settings')}
              className="flex items-center gap-2 cursor-pointer hover:opacity-85 transition-opacity"
              title={t('nav.profileSettings')}
            >
              <img
                src={user.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${user.email}`}
                alt={user.name}
                className="w-7 h-7 rounded-full bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 object-cover"
              />
              <div className="hidden md:block text-left">
                <p className="text-xs font-medium text-slate-800 dark:text-slate-200 leading-none">{user.name}</p>
                <p className="text-[10px] font-mono text-slate-400 dark:text-slate-500 leading-none mt-0.5">
                  {t('nav.viaProvider', { provider: user.provider || 'OAuth' })}
                </p>
              </div>
            </div>
          ) : (
            <button
              onClick={() => navigate('/auth/login')}
              className="text-xs text-indigo-600 dark:text-indigo-400 hover:underline cursor-pointer"
            >
              {t('nav.signIn')}
            </button>
          )}

          {/* Logout button */}
          {user && (
            <button
              onClick={handleLogout}
              className="p-1.5 rounded-lg text-slate-400 dark:text-slate-500 hover:text-rose-500 hover:bg-rose-50 dark:hover:bg-rose-950/30 transition-colors cursor-pointer"
              title={t('nav.signOut')}
            >
              <i className="pi pi-sign-out text-xs" />
            </button>
          )}
        </div>
      </div>
    </header>
  )
}
