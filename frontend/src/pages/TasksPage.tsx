import React from 'react'

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

  return (
    <div className="h-full flex flex-col space-y-3.5 animate-fade-in min-h-0">
      {/* Page Title */}
      <div className="flex items-center justify-between shrink-0">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <i className="pi pi-check-square text-indigo-500 dark:text-indigo-400" />
            {t('tasks.title', { count: tasks.length })}
          </h1>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{t('tasks.subtitle')}</p>
        </div>
      </div>

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

      {/* Quick Slide-Over Drawer & Create Modal */}
      <TaskDrawer />
      <TaskModal />
    </div>
  )
}

export default TasksPage
