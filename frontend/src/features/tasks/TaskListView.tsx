import React, { useEffect, useMemo, useRef } from 'react'

import { TaskListSkeleton } from '@/components'
import { useInfiniteTasks, useToggleFocus, useUpdateTask } from '@/hooks'
import { useTranslation } from '@/i18n'
import { useTaskStore } from '@/stores'
import { compareTasks, type TaskFilterParams, type TaskPriority, type TaskStatus } from '@/types'

const STATUS_CLASSES: Record<TaskStatus, string> = {
  backlog:
    'text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800/80 border-slate-200 dark:border-slate-700/50',
  todo: 'text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-950/40 border-blue-200 dark:border-blue-800/50',
  in_progress:
    'text-amber-700 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/40 border-amber-200 dark:border-amber-800/50',
  review:
    'text-purple-600 dark:text-purple-400 bg-purple-50 dark:bg-purple-950/40 border-purple-200 dark:border-purple-800/50',
  done: 'text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/40 border-emerald-200 dark:border-emerald-800/50'
}

const PRIORITY_META: Record<TaskPriority, { icon: string; class: string }> = {
  urgent: { icon: 'pi-exclamation-triangle', class: 'text-rose-500 dark:text-rose-400' },
  high: { icon: 'pi-angle-double-up', class: 'text-amber-600 dark:text-amber-400' },
  medium: { icon: 'pi-minus', class: 'text-blue-600 dark:text-blue-400' },
  low: { icon: 'pi-angle-down', class: 'text-slate-500' }
}

export const TaskListView: React.FC = () => {
  const {
    openDrawer,
    filterStatus,
    filterPriority,
    filterTag,
    filterProject,
    filterFocusOnly,
    searchQuery
  } = useTaskStore()
  const { t, locale } = useTranslation()

  const filterParams: TaskFilterParams = useMemo(
    () => ({
      status: filterStatus || undefined,
      priority: filterPriority || undefined,
      tag: filterTag || undefined,
      project_id: filterProject || undefined,
      is_focus: filterFocusOnly ? true : undefined,
      search: searchQuery || undefined
    }),
    [filterStatus, filterPriority, filterTag, filterProject, filterFocusOnly, searchQuery]
  )

  const {
    data,
    isLoading,
    isError,
    error,
    refetch,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage
  } = useInfiniteTasks(filterParams)

  const tasks = useMemo(() => {
    const raw = data?.pages.flatMap((page) => (Array.isArray(page?.items) ? page.items : [])) || []
    const uniqueMap = new Map<number, (typeof raw)[number]>()
    for (const item of raw) {
      if (item && item.id != null && !uniqueMap.has(item.id)) {
        uniqueMap.set(item.id, item)
      }
    }
    return Array.from(uniqueMap.values()).sort(compareTasks)
  }, [data])

  const sentinelRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    const el = sentinelRef.current
    if (!el) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
          fetchNextPage()
        }
      },
      { threshold: 0.1 }
    )

    observer.observe(el)
    return () => observer.disconnect()
  }, [hasNextPage, isFetchingNextPage, fetchNextPage])

  const updateTaskMutation = useUpdateTask()
  const toggleFocusMutation = useToggleFocus()

  if (isLoading) {
    return <TaskListSkeleton count={6} />
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

  if (tasks.length === 0) {
    return (
      <div className="py-20 text-center rounded-2xl border border-dashed border-slate-300 dark:border-slate-800 bg-white/50 dark:bg-slate-950/40 p-8">
        <i className="pi pi-check-circle text-4xl text-slate-400 dark:text-slate-600 mb-3" />
        <h3 className="text-sm font-medium text-slate-700 dark:text-slate-300">
          {t('tasks.noTasksFound')}
        </h3>
        <p className="text-xs text-slate-500 max-w-sm mx-auto mt-1 mb-4">
          {t('tasks.noTasksSubtitle')}
        </p>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col min-h-0 overflow-y-auto rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 divide-y divide-slate-100 dark:divide-slate-800/80 shadow-xs scrollbar-thin">
      {tasks.map((task) => {
        const statusClass = STATUS_CLASSES[task.status] || STATUS_CLASSES.todo
        const statusLabel = t(`tasks.status.${task.status}`)
        const priorityMeta = PRIORITY_META[task.priority] || PRIORITY_META.medium
        const priorityLabel = t(`tasks.priority.${task.priority}`)
        const isDone = task.status === 'done'

        return (
          <div
            key={task.id}
            onClick={() => openDrawer(task)}
            className="group flex items-center justify-between px-3.5 py-2.5 hover:bg-slate-50 dark:hover:bg-slate-800/80 cursor-pointer transition-colors text-xs gap-3"
          >
            {/* Left: Complete toggle, Focus pin, Title & Last Progress Note */}
            <div className="flex items-center gap-3 min-w-0 flex-1">
              {/* Done Checkbox */}
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  updateTaskMutation.mutate({
                    id: task.id,
                    payload: { status: isDone ? 'todo' : 'done' }
                  })
                }}
                className={`h-4 w-4 rounded border flex items-center justify-center shrink-0 transition-colors ${
                  isDone
                    ? 'border-emerald-500 bg-emerald-500 text-white'
                    : 'border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-950 hover:border-indigo-500'
                }`}
                title={isDone ? t('tasks.reopenTask') : t('tasks.completeTask')}
              >
                {isDone && <i className="pi pi-check text-[9px]" />}
              </button>

              {/* Pin Focus Toggle */}
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  toggleFocusMutation.mutate(task.id)
                }}
                className={`p-0.5 rounded transition-colors ${
                  task.is_focus
                    ? 'text-amber-500 dark:text-amber-400'
                    : 'text-slate-400 dark:text-slate-600 opacity-0 group-hover:opacity-100 hover:text-slate-600 dark:hover:text-slate-400'
                }`}
                title={task.is_focus ? t('focus.removeFromFocus') : t('focus.pinToFocus')}
              >
                <i className={`pi ${task.is_focus ? 'pi-star-fill' : 'pi-star'} text-[11px]`} />
              </button>

              {/* Title & Last Progress Note */}
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <span
                    className={`font-medium truncate ${
                      isDone
                        ? 'line-through text-slate-400 dark:text-slate-500'
                        : 'text-slate-800 dark:text-slate-200 group-hover:text-indigo-600 dark:group-hover:text-indigo-200'
                    }`}
                  >
                    {task.title}
                  </span>
                  {task.project_name && (
                    <span
                      className="inline-flex items-center gap-1.5 px-1.5 py-0.5 rounded text-[10px] font-semibold border shadow-2xs shrink-0"
                      style={{
                        backgroundColor: `${task.project_color || '#6366f1'}15`,
                        borderColor: `${task.project_color || '#6366f1'}35`,
                        color: task.project_color || '#6366f1'
                      }}
                    >
                      <span
                        className="w-1.5 h-1.5 rounded-full shrink-0"
                        style={{ backgroundColor: task.project_color || '#6366f1' }}
                      />
                      {task.project_name}
                    </span>
                  )}
                  {(task.is_recurring || task.recurring_task_id) && (
                    <span
                      title={t('tasks.recurringTooltip', {
                        pattern: task.recurrence_pattern
                          ? t(`taskModal.patterns.${task.recurrence_pattern}`)
                          : ''
                      })}
                      className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-purple-50 dark:bg-purple-950/40 text-purple-600 dark:text-purple-300 border border-purple-200 dark:border-purple-800/60 shrink-0"
                    >
                      <i className="pi pi-sync text-[9px]" />
                      <span>{t('tasks.recurringBadge')}</span>
                    </span>
                  )}
                  {task.tags && task.tags.length > 0 && (
                    <span className="hidden md:inline-flex items-center gap-1 text-[10px] font-mono text-slate-500">
                      {task.tags.map((t) => `#${t}`).join(' ')}
                    </span>
                  )}
                </div>

                {/* Progress Context Note (remember where you left off) */}
                {task.last_progress_note ? (
                  <p className="text-[11px] text-indigo-600 dark:text-indigo-400 truncate mt-0.5 flex items-center gap-1">
                    <i className="pi pi-history text-[9px]" />
                    <span className="italic">{task.last_progress_note}</span>
                  </p>
                ) : null}
              </div>
            </div>

            {/* Right: Subtasks progress bar, Priority, Status, Due Date */}
            <div className="flex items-center gap-3 shrink-0">
              {/* Subtasks Progress */}
              {task.subtasks_count > 0 && (
                <div className="hidden sm:flex items-center gap-1.5 min-w-[90px]">
                  <div className="w-14 h-1.5 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                    <div
                      className="h-full bg-indigo-500 transition-all duration-300"
                      style={{ width: `${task.progress_percent}%` }}
                    />
                  </div>
                  <span className="text-[10px] font-mono text-slate-500 dark:text-slate-400">
                    {task.subtasks_completed}/{task.subtasks_count}
                  </span>
                </div>
              )}

              {/* Priority */}
              <span
                className={`inline-flex items-center gap-1 text-[11px] font-medium ${priorityMeta.class}`}
                title={`Priority: ${priorityLabel}`}
              >
                <i className={`pi ${priorityMeta.icon} text-[10px]`} />
                <span className="hidden md:inline">{priorityLabel}</span>
              </span>

              {/* Status Badge */}
              <span
                className={`px-2 py-0.5 rounded-full text-[10px] font-medium border ${statusClass}`}
              >
                {statusLabel}
              </span>

              {/* Due Date */}
              {!task.is_recurring && task.due_date && (
                <span className="hidden lg:inline-flex items-center gap-1 text-[10px] text-slate-500 font-mono">
                  <i className="pi pi-calendar text-[9px]" />
                  {new Date(task.due_date).toLocaleDateString(locale === 'id' ? 'id-ID' : 'en-US', {
                    day: 'numeric',
                    month: 'short'
                  })}
                </span>
              )}
            </div>
          </div>
        )
      })}

      {/* Infinite Scroll Sentinel & Loader */}
      <div ref={sentinelRef} className="py-2.5 flex items-center justify-center bg-transparent">
        {isFetchingNextPage ? (
          <div className="w-full py-2">
            <TaskListSkeleton count={2} />
          </div>
        ) : hasNextPage ? (
          <button
            onClick={() => fetchNextPage()}
            className="text-xs text-slate-400 hover:text-indigo-500 py-1 transition-colors font-mono cursor-pointer"
          >
            {t('tasks.loadingMore')}
          </button>
        ) : tasks.length > 10 ? (
          <span className="text-[11px] text-slate-400 dark:text-slate-600 py-1 font-mono">
            {t('tasks.allTasksLoaded')}
          </span>
        ) : null}
      </div>
    </div>
  )
}
