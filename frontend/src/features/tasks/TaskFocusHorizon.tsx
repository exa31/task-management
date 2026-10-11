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
    <div className="rounded-2xl border border-slate-200/80 dark:border-slate-800/80 bg-white/80 dark:bg-slate-900/70 p-4 sm:p-5 shadow-xs backdrop-blur-xl space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 rounded-lg bg-amber-500/10 dark:bg-amber-400/15 border border-amber-500/20 dark:border-amber-400/20 flex items-center justify-center text-amber-500 dark:text-amber-400">
            <i className="pi pi-star-fill text-xs" />
          </div>
          <span className="text-xs font-bold uppercase tracking-wider text-slate-800 dark:text-slate-200">
            {t('focus.title')}
          </span>
          <span className="px-2 py-0.5 rounded-full text-[10px] font-mono font-semibold bg-amber-500/10 dark:bg-amber-400/10 text-amber-700 dark:text-amber-400 border border-amber-500/20 dark:border-amber-400/20">
            {t('focus.activeCount', { count: focusTasks.length })}
          </span>
        </div>
        <span className="text-[11px] text-slate-500 dark:text-slate-400 hidden sm:inline-flex items-center gap-1.5">
          {t('focus.subtitle')}
        </span>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2.5">
        {focusTasks.map((task) => (
          <div
            key={task.id}
            onClick={() => openDrawer(task)}
            className="group relative flex items-center justify-between gap-3 rounded-xl border border-slate-200/80 dark:border-slate-800/80 bg-slate-50/70 dark:bg-slate-950/40 p-3 hover:border-indigo-400/60 dark:hover:border-indigo-500/50 hover:bg-white dark:hover:bg-slate-900/90 cursor-pointer transition-all shadow-2xs"
          >
            <div className="flex items-center gap-2.5 min-w-0">
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  updateTaskMutation.mutate({
                    id: task.id,
                    payload: { status: 'done' }
                  })
                }}
                className="h-4 w-4 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 hover:border-indigo-500 hover:bg-indigo-50 dark:hover:bg-indigo-950/50 flex items-center justify-center shrink-0 transition-colors"
                title={t('focus.completeTask')}
                data-pr-tooltip={t('focus.completeTask')}
              >
                <i className="pi pi-check text-[9px] text-transparent hover:text-indigo-500 dark:hover:text-indigo-400" />
              </button>
              <div className="min-w-0">
                <p className="text-xs font-semibold text-slate-800 dark:text-slate-200 truncate group-hover:text-indigo-600 dark:group-hover:text-indigo-300 transition-colors">
                  {task.title}
                </p>
                {task.last_progress_note ? (
                  <p className="text-[10px] text-indigo-600 dark:text-indigo-400 truncate mt-0.5 flex items-center gap-1 font-medium">
                    <i className="pi pi-history text-[8px]" />
                    <span>{task.last_progress_note}</span>
                  </p>
                ) : (
                  <p className="text-[10px] text-slate-500 dark:text-slate-400 truncate mt-0.5">
                    {task.subtasks_count > 0
                      ? t('focus.checklistProgress', {
                          completed: task.subtasks_completed,
                          total: task.subtasks_count,
                          percent: task.progress_percent
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
              className="p-1.5 rounded-lg text-amber-500 dark:text-amber-400 hover:bg-amber-500/10 dark:hover:bg-amber-400/10 shrink-0 transition-colors"
              title={t('focus.removeFromFocus')}
              data-pr-tooltip={t('focus.removeFromFocus')}
            >
              <i className="pi pi-star-fill text-xs" />
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
