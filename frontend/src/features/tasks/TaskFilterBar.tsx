import React from 'react'

import { useTasks } from '@/hooks'
import { useTranslation } from '@/i18n'
import { useTaskStore, useUIStore } from '@/stores'

export const TaskFilterBar: React.FC = () => {
  const {
    activeView,
    setActiveView,
    filterStatus,
    setFilterStatus,
    filterFocusOnly,
    setFilterFocusOnly,
    searchQuery,
    setSearchQuery,
  } = useTaskStore()

  const { data: rawAllTasks } = useTasks()
  const allTasks = Array.isArray(rawAllTasks) ? rawAllTasks : []
  const { openCreateTaskModal } = useUIStore()
  const { t } = useTranslation()

  const statusFilters = [
    { id: '', label: t('tasks.allStatuses') },
    { id: 'in_progress', label: t('tasks.status.in_progress') },
    { id: 'todo', label: t('tasks.status.todo') },
    { id: 'backlog', label: t('tasks.status.backlog') },
    { id: 'review', label: t('tasks.status.review') },
    { id: 'done', label: t('tasks.status.done') },
  ]

  const focusCount = allTasks.filter((t) => t?.is_focus).length

  return (
    <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 pb-3 border-b border-slate-200 dark:border-slate-800">
      {/* Left: View Switcher & Search */}
      <div className="flex items-center gap-2 flex-wrap">
        {/* View Switcher Toggle */}
        <div className="inline-flex rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-100 dark:bg-slate-950 p-0.5">
          <button
            onClick={() => setActiveView('list')}
            className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-colors ${
              activeView === 'list'
                ? 'bg-white dark:bg-slate-800 text-indigo-600 dark:text-indigo-400 shadow-xs'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
            }`}
            title={t('tasks.listView')}
          >
            <i className="pi pi-list text-xs" />
            <span>{t('tasks.listView')}</span>
          </button>
          <button
            onClick={() => setActiveView('board')}
            className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-colors ${
              activeView === 'board'
                ? 'bg-white dark:bg-slate-800 text-indigo-600 dark:text-indigo-400 shadow-xs'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
            }`}
            title={t('tasks.boardView')}
          >
            <i className="pi pi-table text-xs" />
            <span>{t('tasks.boardView')}</span>
          </button>
        </div>

        {/* Status Filter Pills (Only shown in list view; hidden in board view since columns represent statuses) */}
        <div className="flex items-center gap-1 overflow-x-auto scrollbar-none py-1">
          {activeView === 'list' &&
            statusFilters.map((s) => (
              <button
                key={s.id}
                onClick={() => setFilterStatus(s.id)}
                className={`px-2.5 py-1 rounded-full text-xs font-medium transition-colors whitespace-nowrap ${
                  filterStatus === s.id
                    ? 'bg-indigo-600 text-white'
                    : 'bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 border border-slate-200 dark:border-slate-800'
                }`}
              >
                {s.label}
              </button>
            ))}
          <button
            onClick={() => setFilterFocusOnly(!filterFocusOnly)}
            className={`px-2.5 py-1 rounded-full text-xs font-medium transition-colors whitespace-nowrap inline-flex items-center gap-1 ${
              filterFocusOnly
                ? 'bg-amber-500 text-slate-950 font-semibold'
                : 'bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 border border-slate-200 dark:border-slate-800'
            }`}
          >
            <i className="pi pi-star-fill text-[10px]" />
            {t('tasks.focusFilter', { count: focusCount })}
          </button>
        </div>
      </div>

      {/* Right: Search & Create Button */}
      <div className="flex items-center gap-2">
        <div className="relative flex-1 sm:w-48">
          <i className="pi pi-search absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 text-xs" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder={t('tasks.searchPlaceholder')}
            className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 pl-7 pr-3 py-1.5 text-xs text-slate-900 dark:text-slate-200 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none transition-colors"
          />
        </div>

        <button
          onClick={openCreateTaskModal}
          className="inline-flex items-center justify-center gap-1.5 px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-xs font-medium text-white shadow-xs transition-colors shrink-0"
        >
          <i className="pi pi-plus text-xs" />
          {t('dashboard.newTask')}
        </button>
      </div>
    </div>
  )
}
