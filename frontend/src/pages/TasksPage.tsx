import React from 'react'
import { useSearchParams } from 'react-router-dom'

import { RecurringTasksManagementView } from '@/features/tasks/RecurringTasksManagementView'
import { TaskBoardView } from '@/features/tasks/TaskBoardView'
import { TaskDrawer } from '@/features/tasks/TaskDrawer'
import { TaskFilterBar } from '@/features/tasks/TaskFilterBar'
import { TaskFocusHorizon } from '@/features/tasks/TaskFocusHorizon'
import { TaskListView } from '@/features/tasks/TaskListView'
import { TaskModal } from '@/features/tasks/TaskModal'
import { useTasks } from '@/hooks'
import { useTranslation } from '@/i18n'
import { useTaskStore } from '@/stores'

export const TasksPage: React.FC = () => {
  const { activeView } = useTaskStore()
  const { data: rawTasks } = useTasks()
  const tasks = Array.isArray(rawTasks) ? rawTasks : []
  const { t } = useTranslation()
  const [searchParams, setSearchParams] = useSearchParams()

  const currentTab = searchParams.get('tab') === 'recurring' ? 'recurring' : 'tasks'

  const handleTabChange = (tab: 'tasks' | 'recurring') => {
    if (tab === 'recurring') {
      searchParams.set('tab', 'recurring')
    } else {
      searchParams.delete('tab')
    }
    setSearchParams(searchParams, { replace: true })
  }

  return (
    <div className="h-full flex flex-col space-y-3.5 animate-fade-in min-h-0">
      {/* Page Title & Tab Switcher Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 shrink-0">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <i
              className={`pi ${
                currentTab === 'tasks' ? 'pi-check-square' : 'pi-sync'
              } text-indigo-500 dark:text-indigo-400`}
            />
            {currentTab === 'tasks'
              ? t('tasks.title', { count: tasks.length })
              : t('recurring.title')}
          </h1>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            {currentTab === 'tasks' ? t('tasks.subtitle') : t('recurring.subtitle')}
          </p>
        </div>

        {/* Tab Switcher */}
        <div className="inline-flex p-1 rounded-xl bg-slate-100 dark:bg-slate-800/80 border border-slate-200/80 dark:border-slate-700/80 shrink-0">
          <button
            type="button"
            onClick={() => handleTabChange('tasks')}
            className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer ${
              currentTab === 'tasks'
                ? 'bg-white dark:bg-slate-900 text-indigo-600 dark:text-indigo-400 shadow-xs'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
            }`}
          >
            <i className="pi pi-check-square text-xs" />
            <span>{t('recurring.tabAllTasks')}</span>
          </button>
          <button
            type="button"
            onClick={() => handleTabChange('recurring')}
            className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer ${
              currentTab === 'recurring'
                ? 'bg-white dark:bg-slate-900 text-indigo-600 dark:text-indigo-400 shadow-xs'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
            }`}
          >
            <i className="pi pi-sync text-xs" />
            <span>{t('recurring.tabRecurringTemplates')}</span>
          </button>
        </div>
      </div>

      {currentTab === 'tasks' ? (
        <>
          {/* Today's Focus Horizon */}
          <div className="shrink-0">
            <TaskFocusHorizon />
          </div>

          {/* Filter and View Bar */}
          <div className="shrink-0">
            <TaskFilterBar />
          </div>

          {/* Dual View Renderer */}
          <div className="flex-1 min-h-0">
            {activeView === 'list' ? <TaskListView /> : <TaskBoardView />}
          </div>
        </>
      ) : (
        /* Recurring Templates Management View */
        <div className="flex-1 min-h-0">
          <RecurringTasksManagementView />
        </div>
      )}

      {/* Quick Slide-Over Drawer & Create Modal */}
      <TaskDrawer />
      <TaskModal />
    </div>
  )
}

export default TasksPage
