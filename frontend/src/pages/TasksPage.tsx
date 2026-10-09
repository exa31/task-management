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
    <div className="space-y-5 pb-12 animate-fade-in">
      {/* Page Title */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <i className="pi pi-check-square text-indigo-500 dark:text-indigo-400" />
            {t('tasks.title', { count: tasks.length })}
          </h1>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            {t('tasks.subtitle')}
          </p>
        </div>
      </div>

      {/* Today's Focus Horizon */}
      <TaskFocusHorizon />

      {/* Filter and View Bar */}
      <TaskFilterBar />

      {/* Dual View Renderer */}
      {activeView === 'list' ? <TaskListView /> : <TaskBoardView />}

      {/* Quick Slide-Over Drawer & Create Modal */}
      <TaskDrawer />
      <TaskModal />
    </div>
  )
}

export default TasksPage
