import { Link, useLocation } from 'react-router-dom'

import { BrandLogo } from '@/components'
import { useNotes, useProjects, useTasks } from '@/hooks'
import { useTranslation } from '@/i18n'
import { useUIStore } from '@/stores'

interface SidebarProps {
  isSidebarOpen: boolean
}

export default function Sidebar({ isSidebarOpen }: SidebarProps) {
  const location = useLocation()
  const { t } = useTranslation()
  const { data: rawTasks } = useTasks()
  const { data: rawNotes } = useNotes()
  const { data: rawProjects } = useProjects()
  const tasks = Array.isArray(rawTasks) ? rawTasks : []
  const notes = Array.isArray(rawNotes) ? rawNotes : []
  const projects = Array.isArray(rawProjects) ? rawProjects : []
  const { openCommandPalette } = useUIStore()

  const isActive = (path: string) => location.pathname === path

  const inProgressCount = tasks.filter((t) => t?.status === 'in_progress').length

  return (
    <aside
      className={`bg-white dark:bg-slate-950 border-r border-slate-200 dark:border-slate-800 transition-all duration-300 flex flex-col fixed inset-y-0 left-0 z-30 md:static ${
        isSidebarOpen ? 'translate-x-0 w-64' : '-translate-x-full md:translate-x-0 md:w-20'
      }`}
    >
      {/* Brand Header */}
      <div className="h-14 flex items-center px-4 border-b border-slate-200 dark:border-slate-800 shrink-0">
        <Link to="/dashboard" className="flex items-center gap-2.5">
          <BrandLogo size={28} />
          {isSidebarOpen && (
            <span className="font-bold text-sm tracking-tight text-slate-900 dark:text-slate-100">
              Task<span className="text-indigo-600 dark:text-indigo-400">Hub</span>
            </span>
          )}
        </Link>
      </div>

      {/* Nav Items */}
      <div className="flex-1 overflow-y-auto p-3 space-y-1">
        {isSidebarOpen && (
          <div className="text-[10px] font-mono uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-2 px-2">
            {t('nav.workspace')}
          </div>
        )}

        <Link
          to="/dashboard"
          className={`flex items-center justify-between px-3 py-2 rounded-lg text-xs font-medium transition-colors ${
            isActive('/dashboard')
              ? 'bg-indigo-50 dark:bg-indigo-600/20 text-indigo-600 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-500/30'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-900'
          }`}
          title={t('nav.dashboard')}
        >
          <div className="flex items-center gap-2.5">
            <i className="pi pi-th-large text-sm" />
            {isSidebarOpen && <span>{t('nav.dashboard')}</span>}
          </div>
        </Link>

        <Link
          to="/tasks"
          className={`flex items-center justify-between px-3 py-2 rounded-lg text-xs font-medium transition-colors ${
            isActive('/tasks')
              ? 'bg-indigo-50 dark:bg-indigo-600/20 text-indigo-600 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-500/30'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-900'
          }`}
          title={t('nav.tasks')}
        >
          <div className="flex items-center gap-2.5">
            <i className="pi pi-check-square text-sm" />
            {isSidebarOpen && <span>{t('nav.tasks')}</span>}
          </div>
          {isSidebarOpen && inProgressCount > 0 && (
            <span className="px-1.5 py-0.2 rounded-full text-[10px] font-mono bg-amber-500/10 dark:bg-amber-500/20 text-amber-600 dark:text-amber-300 border border-amber-500/20 dark:border-amber-500/30">
              {inProgressCount}
            </span>
          )}
        </Link>

        <Link
          to="/projects"
          className={`flex items-center justify-between px-3 py-2 rounded-lg text-xs font-medium transition-colors ${
            isActive('/projects')
              ? 'bg-indigo-50 dark:bg-indigo-600/20 text-indigo-600 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-500/30'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-900'
          }`}
          title={t('nav.projects')}
        >
          <div className="flex items-center gap-2.5">
            <i className="pi pi-folder text-sm" />
            {isSidebarOpen && <span>{t('nav.projects')}</span>}
          </div>
          {isSidebarOpen && projects.length > 0 && (
            <span className="px-1.5 py-0.2 rounded-full text-[10px] font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
              {projects.length}
            </span>
          )}
        </Link>

        <Link
          to="/notes"
          className={`flex items-center justify-between px-3 py-2 rounded-lg text-xs font-medium transition-colors ${
            isActive('/notes')
              ? 'bg-indigo-50 dark:bg-indigo-600/20 text-indigo-600 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-500/30'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-900'
          }`}
          title={t('nav.notes')}
        >
          <div className="flex items-center gap-2.5">
            <i className="pi pi-lightbulb text-sm" />
            {isSidebarOpen && <span>{t('nav.notes')}</span>}
          </div>
          {isSidebarOpen && notes.length > 0 && (
            <span className="px-1.5 py-0.2 rounded-full text-[10px] font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
              {notes.length}
            </span>
          )}
        </Link>

        {isSidebarOpen && (
          <div className="text-[10px] font-mono uppercase tracking-wider text-slate-400 dark:text-slate-500 mt-6 mb-2 px-2">
            {t('nav.aiIntegration')}
          </div>
        )}

        <Link
          to="/mcp-settings"
          className={`flex items-center justify-between px-3 py-2 rounded-lg text-xs font-medium transition-colors ${
            isActive('/mcp-settings')
              ? 'bg-indigo-50 dark:bg-indigo-600/20 text-indigo-600 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-500/30'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-900'
          }`}
          title={t('nav.mcp')}
        >
          <div className="flex items-center gap-2.5">
            <i className="pi pi-code text-sm" />
            {isSidebarOpen && <span>{t('nav.mcp')}</span>}
          </div>
          {isSidebarOpen && (
            <span className="px-1.5 py-0.2 rounded text-[9px] font-mono bg-indigo-500/10 dark:bg-indigo-500/20 text-indigo-600 dark:text-indigo-300 border border-indigo-500/20 dark:border-indigo-500/30">
              MCP
            </span>
          )}
        </Link>

        <Link
          to="/settings"
          className={`flex items-center justify-between px-3 py-2 rounded-lg text-xs font-medium transition-colors ${
            isActive('/settings')
              ? 'bg-indigo-50 dark:bg-indigo-600/20 text-indigo-600 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-500/30'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-900'
          }`}
          title={t('nav.settings')}
        >
          <div className="flex items-center gap-2.5">
            <i className="pi pi-cog text-sm" />
            {isSidebarOpen && <span>{t('nav.settings')}</span>}
          </div>
        </Link>
      </div>

      {/* Footer shortcut helper */}
      {isSidebarOpen && (
        <div className="p-3 border-t border-slate-200 dark:border-slate-800">
          <button
            onClick={openCommandPalette}
            className="w-full flex items-center justify-between px-3 py-2 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 hover:border-slate-300 dark:hover:border-slate-700 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 text-xs transition-colors cursor-pointer"
          >
            <span className="flex items-center gap-2">
              <i className="pi pi-search text-xs" />
              <span>{t('nav.commandPalette')}</span>
            </span>
            <kbd className="px-1.5 py-0.5 rounded text-[10px] font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700">
              ⌘K
            </kbd>
          </button>
        </div>
      )}
    </aside>
  )
}
