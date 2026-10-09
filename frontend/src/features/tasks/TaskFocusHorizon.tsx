import React from 'react'

import { useTasks, useToggleFocus, useUpdateTask } from '@/hooks'
import { useTranslation } from '@/i18n'
import { useTaskStore } from '@/stores'

export const TaskFocusHorizon: React.FC = () => {
  const { openDrawer } = useTaskStore()
  const { data: rawTasks } = useTasks()
  const tasks = Array.isArray(rawTasks) ? rawTasks : []
  const { t } = useTranslation()
  const updateTaskMutation = useUpdateTask()
  const toggleFocusMutation = useToggleFocus()

  // Pinned to focus and not done
  const focusTasks = tasks.filter((t) => t && t.is_focus && t.status !== 'done')

  if (focusTasks.length === 0) return null

  return (
    <div className="rounded-xl border border-indigo-200 dark:border-indigo-500/30 bg-indigo-50/70 dark:bg-gradient-to-r dark:from-indigo-950/40 dark:via-slate-900 dark:to-indigo-950/20 p-3.5 shadow-sm dark:shadow-lg backdrop-blur-md mb-5">
      <div className="flex items-center justify-between mb-2.5">
        <div className="flex items-center gap-2">
          <span className="flex h-2 w-2 rounded-full bg-indigo-500 dark:bg-indigo-400 animate-pulse" />
          <h2 className="text-xs font-semibold uppercase tracking-wider text-indigo-900 dark:text-indigo-300">
            {t('focus.title')}
          </h2>
          <span className="px-1.5 py-0.2 rounded text-[10px] font-mono bg-indigo-100 dark:bg-indigo-500/20 text-indigo-800 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-500/30">
            {t('focus.activeCount', { count: focusTasks.length })}
          </span>
        </div>
        <span className="text-[11px] text-slate-500 dark:text-slate-400">{t('focus.subtitle')}</span>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2.5">
        {focusTasks.map((task) => (
          <div
            key={task.id}
            onClick={() => openDrawer(task)}
            className="group relative flex items-center justify-between gap-3 rounded-lg border border-slate-200 dark:border-slate-800/80 bg-white dark:bg-slate-900/90 p-2.5 hover:border-indigo-400 dark:hover:border-indigo-500/40 hover:bg-slate-50 dark:hover:bg-slate-800 cursor-pointer transition-all shadow-xs"
          >
            <div className="flex items-center gap-2.5 min-w-0">
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  updateTaskMutation.mutate({
                    id: task.id,
                    payload: { status: 'done' },
                  })
                }}
                className="h-4 w-4 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-950 hover:border-indigo-500 flex items-center justify-center shrink-0 transition-colors"
                title={t('focus.completeTask')}
              >
                <i className="pi pi-check text-[9px] text-transparent hover:text-indigo-500 dark:hover:text-indigo-400" />
              </button>
              <div className="min-w-0">
                <p className="text-xs font-medium text-slate-800 dark:text-slate-200 truncate group-hover:text-indigo-600 dark:group-hover:text-indigo-200">
                  {task.title}
                </p>
                {task.last_progress_note ? (
                  <p className="text-[10px] text-indigo-600 dark:text-indigo-400 truncate mt-0.5">
                    <i className="pi pi-history text-[8px] mr-1" />
                    {task.last_progress_note}
                  </p>
                ) : (
                  <p className="text-[10px] text-slate-500 truncate mt-0.5">
                    {task.subtasks_count > 0
                      ? t('focus.checklistProgress', {
                          completed: task.subtasks_completed,
                          total: task.subtasks_count,
                          percent: task.progress_percent,
                        })
                      : t('focus.noProgressLog')}
                  </p>
                )}
              </div>
            </div>

            <button
              onClick={(e) => {
                e.stopPropagation()
                toggleFocusMutation.mutate(task.id)
              }}
              className="p-1 rounded text-slate-400 dark:text-slate-500 hover:text-amber-500 dark:hover:text-amber-400 shrink-0"
              title={t('focus.removeFromFocus')}
            >
              <i className="pi pi-star-fill text-amber-500 dark:text-amber-400 text-xs" />
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
