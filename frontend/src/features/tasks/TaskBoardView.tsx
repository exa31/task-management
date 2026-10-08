import React, { useMemo } from 'react'

import { useTasks, useToggleFocus, useUpdateTask } from '@/hooks'
import { useTranslation } from '@/i18n'
import { useTaskStore } from '@/stores'
import type { TaskFilterParams, TaskStatus } from '@/types'

const COLUMNS: { id: TaskStatus; color: string }[] = [
  { id: 'backlog', color: 'border-slate-800' },
  { id: 'todo', color: 'border-blue-900/50' },
  { id: 'in_progress', color: 'border-amber-900/50' },
  { id: 'review', color: 'border-purple-900/50' },
  { id: 'done', color: 'border-emerald-900/50' },
]

export const TaskBoardView: React.FC = () => {
  const {
    openDrawer,
    filterPriority,
    filterTag,
    filterFocusOnly,
    searchQuery,
  } = useTaskStore()
  const { t, locale } = useTranslation()

  const filterParams: TaskFilterParams = useMemo(
    () => ({
      // In board view, always fetch all statuses so each column displays its tasks
      status: undefined,
      priority: filterPriority || undefined,
      tag: filterTag || undefined,
      is_focus: filterFocusOnly ? true : undefined,
      search: searchQuery || undefined,
    }),
    [filterPriority, filterTag, filterFocusOnly, searchQuery]
  )

  const { data: tasks = [], isLoading, isError, error, refetch } = useTasks(filterParams)
  const updateTaskMutation = useUpdateTask()
  const toggleFocusMutation = useToggleFocus()

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
  }

  const handleDrop = (e: React.DragEvent, status: TaskStatus) => {
    e.preventDefault()
    const taskIdStr = e.dataTransfer.getData('text/plain')
    if (taskIdStr) {
      const taskId = parseInt(taskIdStr, 10)
      if (!isNaN(taskId)) {
        updateTaskMutation.mutate({ id: taskId, payload: { status } })
      }
    }
  }

  const handleDragStart = (e: React.DragEvent, taskId: number) => {
    e.dataTransfer.setData('text/plain', String(taskId))
  }

  if (isLoading) {
    return (
      <div className="py-16 text-center text-slate-500 text-xs flex items-center justify-center gap-2">
        <i className="pi pi-spin pi-spinner text-indigo-400" /> {t('tasks.loadingTasks')}
      </div>
    )
  }

  if (isError) {
    return (
      <div className="py-12 text-center rounded-2xl border border-rose-900/40 bg-rose-950/20 p-8">
        <i className="pi pi-exclamation-circle text-3xl text-rose-400 mb-2" />
        <h3 className="text-sm font-semibold text-rose-300">{t('tasks.failedLoad')}</h3>
        <p className="text-xs text-rose-400/80 max-w-sm mx-auto mt-1 mb-4">
          {(error as Error)?.message || 'An error occurred while connecting to the API server.'}
        </p>
        <button
          onClick={() => refetch()}
          className="px-3.5 py-1.5 rounded-lg bg-rose-600 hover:bg-rose-500 text-xs font-medium text-white inline-flex items-center gap-1.5"
        >
          <i className="pi pi-refresh text-xs" /> {t('common.retry')}
        </button>
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-3.5 items-start overflow-x-auto pb-4">
      {COLUMNS.map((col) => {
        const columnTasks = tasks.filter((t) => t.status === col.id)
        const colTitle = t(`tasks.status.${col.id}`)

        return (
          <div
            key={col.id}
            onDragOver={handleDragOver}
            onDrop={(e) => handleDrop(e, col.id)}
            className="flex flex-col rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-100/70 dark:bg-slate-900/50 p-2.5 min-h-[500px]"
          >
            {/* Column Header */}
            <div className="flex items-center justify-between pb-2 mb-2 border-b border-slate-200 dark:border-slate-800/80 px-1">
              <span className="text-xs font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-2">
                <span
                  className={`h-2 w-2 rounded-full ${
                    col.id === 'done'
                      ? 'bg-emerald-500'
                      : col.id === 'in_progress'
                      ? 'bg-amber-500'
                      : 'bg-slate-400 dark:bg-slate-500'
                  }`}
                />
                {colTitle}
              </span>
              <span className="px-1.5 py-0.2 rounded-full text-[10px] font-mono bg-slate-200 dark:bg-slate-800 text-slate-700 dark:text-slate-400">
                {columnTasks.length}
              </span>
            </div>

            {/* Task Cards in Column */}
            <div className="space-y-2 flex-1 overflow-y-auto max-h-[calc(100vh-270px)] pr-1 scrollbar-thin">
              {columnTasks.map((task) => (
                <div
                  key={task.id}
                  draggable
                  onDragStart={(e) => handleDragStart(e, task.id)}
                  onClick={() => openDrawer(task)}
                  className="group rounded-lg border border-slate-200 dark:border-slate-800/80 bg-white dark:bg-slate-900 p-3 hover:border-indigo-400 dark:hover:border-indigo-500/50 hover:bg-slate-50 dark:hover:bg-slate-800 cursor-grab active:cursor-grabbing transition-all shadow-xs"
                >
                  <div className="flex items-start justify-between gap-1.5 mb-1.5">
                    <h4 className="text-xs font-medium text-slate-800 dark:text-slate-200 group-hover:text-indigo-600 dark:group-hover:text-indigo-200 line-clamp-2">
                      {task.title}
                    </h4>
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        toggleFocusMutation.mutate(task.id)
                      }}
                      className="p-0.5 text-slate-400 dark:text-slate-600 hover:text-amber-500 dark:hover:text-amber-400 shrink-0"
                    >
                      <i
                        className={`pi ${task.is_focus ? 'pi-star-fill text-amber-500 dark:text-amber-400' : 'pi-star'} text-[10px]`}
                      />
                    </button>
                  </div>

                  {/* Last progress note */}
                  {task.last_progress_note && (
                    <p className="text-[10px] text-indigo-600 dark:text-indigo-400 line-clamp-2 mb-2 italic bg-indigo-50/70 dark:bg-indigo-950/30 p-1.5 rounded border border-indigo-200/80 dark:border-indigo-900/30">
                      <i className="pi pi-history text-[8px] mr-1" />
                      {task.last_progress_note}
                    </p>
                  )}

                  {/* Subtasks Progress Bar */}
                  {task.subtasks_count > 0 && (
                    <div className="mb-2">
                      <div className="flex items-center justify-between text-[10px] text-slate-500 mb-1 font-mono">
                        <span>{t('tasks.checklist')}</span>
                        <span>
                          {task.subtasks_completed}/{task.subtasks_count} ({task.progress_percent}%)
                        </span>
                      </div>
                      <div className="w-full h-1 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                        <div
                          className="h-full bg-indigo-500 transition-all duration-300"
                          style={{ width: `${task.progress_percent}%` }}
                        />
                      </div>
                    </div>
                  )}

                  {/* Card Footer */}
                  <div className="flex items-center justify-between text-[10px] text-slate-500 pt-1 border-t border-slate-100 dark:border-slate-800/50">
                    <span className="uppercase font-mono tracking-wider font-semibold">
                      {t(`tasks.priority.${task.priority}`)}
                    </span>
                    {task.due_date && (
                      <span>
                        {new Date(task.due_date).toLocaleDateString(locale === 'id' ? 'id-ID' : 'en-US', {
                          day: 'numeric',
                          month: 'short',
                        })}
                      </span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )
      })}
    </div>
  )
}
