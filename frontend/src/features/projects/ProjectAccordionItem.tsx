import React, { useMemo, useState } from 'react'

import { confirmDialog } from '@/components'
import { useDeleteProject, useTasks, useToggleFocus, useUpdateTask } from '@/hooks'
import { useTranslation } from '@/i18n'
import { useTaskStore, useUIStore } from '@/stores'
import {
  compareTasks,
  type Project,
  STATUS_COLORS,
  type Task,
  type TaskPriority,
  type TaskStatus
} from '@/types'

const STATUS_KEYS: TaskStatus[] = ['in_progress', 'todo', 'review', 'backlog', 'done']

const PRIORITY_ICONS: Record<TaskPriority, { icon: string; class: string }> = {
  urgent: { icon: 'pi-exclamation-triangle', class: 'text-rose-500' },
  high: { icon: 'pi-angle-double-up', class: 'text-amber-600 dark:text-amber-400' },
  medium: { icon: 'pi-minus', class: 'text-blue-600 dark:text-blue-400' },
  low: { icon: 'pi-angle-down', class: 'text-slate-400' }
}

interface ProjectAccordionItemProps {
  project: Project
  isExpanded: boolean
  onToggleExpand: () => void
  onEdit: (project: Project) => void
}

export const ProjectAccordionItem: React.FC<ProjectAccordionItemProps> = ({
  project,
  isExpanded,
  onToggleExpand,
  onEdit
}) => {
  const { t, locale } = useTranslation()
  const { openDrawer } = useTaskStore()
  const { openCreateTaskModal } = useUIStore()

  const deleteProjectMutation = useDeleteProject()
  const updateTaskMutation = useUpdateTask()
  const toggleFocusMutation = useToggleFocus()

  // Fetch tasks belonging specifically to this project (enabled only when expanded or to calculate live stats)
  const { data: rawTasks, isLoading: isLoadingTasks } = useTasks({
    project_id: project.id
  })
  const tasks = useMemo(
    () => (Array.isArray(rawTasks) ? rawTasks.filter((t) => t && t.project_id === project.id) : []),
    [rawTasks, project.id]
  )

  // Filter tasks by status and sort
  const tasksByStatus = useMemo(() => {
    const map: Record<TaskStatus, Task[]> = {
      in_progress: [],
      todo: [],
      review: [],
      backlog: [],
      done: []
    }
    for (const task of tasks) {
      if (task.status in map) {
        map[task.status].push(task)
      }
    }
    for (const status of Object.keys(map) as TaskStatus[]) {
      map[status].sort(compareTasks)
    }
    return map
  }, [tasks])

  const totalTasks = tasks.length
  const completedTasks = tasksByStatus.done.length
  const percentDone = totalTasks > 0 ? Math.round((completedTasks / totalTasks) * 100) : 0

  const [activeStatusTab, setActiveStatusTab] = useState<TaskStatus | 'all'>('all')

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    confirmDialog({
      header: t('projects.deleteHeader'),
      message: t('projects.deleteMessage', { name: project.name }),
      icon: 'pi pi-trash text-rose-500',
      acceptClassName: 'p-button-danger',
      acceptLabel: t('projects.deleteAccept'),
      rejectLabel: t('projects.deleteReject'),
      accept: () => {
        deleteProjectMutation.mutate(project.id)
      }
    })
  }

  return (
    <div className="rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/90 shadow-xs overflow-hidden transition-all duration-200 hover:border-slate-300 dark:hover:border-slate-700">
      {/* Header Bar */}
      <div
        onClick={onToggleExpand}
        className="p-4 sm:p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4 cursor-pointer select-none bg-slate-50/50 dark:bg-slate-950/20 hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors"
      >
        <div className="flex items-start gap-3.5 min-w-0 flex-1">
          {/* Accent Color Dot / Tag */}
          <span
            className="w-4 h-4 rounded-full mt-1 shrink-0 ring-4 ring-white dark:ring-slate-900 shadow-sm"
            style={{ backgroundColor: project.color || '#6366f1' }}
          />

          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 flex-wrap">
              <h3 className="text-base font-bold text-slate-900 dark:text-white truncate">
                {project.name}
              </h3>
              <span className="px-2 py-0.5 rounded-full text-[10px] font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700/60">
                {t('projects.tasksConnected', { count: totalTasks })}
              </span>
              {totalTasks > 0 && (
                <span className="text-[11px] font-mono font-medium text-emerald-600 dark:text-emerald-400">
                  {percentDone}% {t('tasks.status.done')}
                </span>
              )}
            </div>

            {project.description && (
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-1 line-clamp-1">
                {project.description}
              </p>
            )}

            {/* Mini Progress Multi-bar */}
            {totalTasks > 0 && (
              <div className="w-full max-w-md h-1.5 rounded-full bg-slate-200 dark:bg-slate-800 mt-2.5 overflow-hidden flex">
                {STATUS_KEYS.map((s) => {
                  const count = tasksByStatus[s].length
                  if (count === 0) return null
                  const widthPct = (count / totalTasks) * 100
                  return (
                    <div
                      key={s}
                      className={STATUS_COLORS[s].dot}
                      style={{ width: `${widthPct}%` }}
                      title={`${t(`tasks.status.${s}`)}: ${count}`}
                    />
                  )
                })}
              </div>
            )}
          </div>
        </div>

        {/* Action Controls & Chevron */}
        <div className="flex items-center gap-2 shrink-0 self-end sm:self-center">
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation()
              openCreateTaskModal(project.id)
            }}
            className="px-2.5 py-1.5 text-xs font-medium rounded-lg bg-indigo-600/10 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-600 hover:text-white dark:hover:bg-indigo-600 transition-colors flex items-center gap-1.5"
            title={t('projects.addTaskToProject')}
          >
            <i className="pi pi-plus text-[10px]" />
            <span className="hidden md:inline">{t('dashboard.newTask')}</span>
          </button>

          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation()
              onEdit(project)
            }}
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-200/50 dark:hover:bg-slate-800 transition-colors"
            title={t('projects.editProject')}
          >
            <i className="pi pi-pencil text-xs" />
          </button>

          <button
            type="button"
            onClick={handleDelete}
            className="p-1.5 rounded-lg text-rose-400 hover:text-rose-600 dark:hover:text-rose-300 hover:bg-rose-50 dark:hover:bg-rose-950/40 transition-colors"
            title={t('projects.deleteProject')}
          >
            <i className="pi pi-trash text-xs" />
          </button>

          <div
            className={`w-7 h-7 rounded-lg flex items-center justify-center text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-transform ${
              isExpanded ? 'rotate-180' : ''
            }`}
          >
            <i className="pi pi-chevron-down text-xs" />
          </div>
        </div>
      </div>

      {/* Expanded Task Breakdown */}
      {isExpanded && (
        <div className="border-t border-slate-200 dark:border-slate-800 p-4 sm:p-5 space-y-4 bg-white dark:bg-slate-900 animate-fade-in">
          {/* Status Tabs and Quick Add */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-2 border-b border-slate-100 dark:border-slate-800/80">
            {/* Status Filter Tabs */}
            <div className="flex items-center gap-1 overflow-x-auto scrollbar-none py-0.5">
              <button
                type="button"
                onClick={() => setActiveStatusTab('all')}
                className={`px-2.5 py-1 rounded-lg text-xs font-medium transition-colors whitespace-nowrap ${
                  activeStatusTab === 'all'
                    ? 'bg-slate-900 text-white dark:bg-white dark:text-slate-900'
                    : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800'
                }`}
              >
                {t('tasks.allStatuses')} ({totalTasks})
              </button>
              {STATUS_KEYS.map((s) => (
                <button
                  key={s}
                  type="button"
                  onClick={() => setActiveStatusTab(s)}
                  className={`px-2.5 py-1 rounded-lg text-xs font-medium transition-colors flex items-center gap-1.5 whitespace-nowrap ${
                    activeStatusTab === s
                      ? 'bg-indigo-600 text-white'
                      : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800'
                  }`}
                >
                  <span className={`w-1.5 h-1.5 rounded-full ${STATUS_COLORS[s].dot}`} />
                  {t(`tasks.status.${s}`)} ({tasksByStatus[s].length})
                </button>
              ))}
            </div>

            <button
              type="button"
              onClick={() => openCreateTaskModal(project.id)}
              className="text-xs font-semibold text-indigo-600 dark:text-indigo-400 hover:text-indigo-500 inline-flex items-center gap-1 self-start sm:self-auto cursor-pointer"
            >
              <i className="pi pi-plus text-xs" />
              {t('projects.addTaskToProject')}
            </button>
          </div>

          {/* Loading Indicator */}
          {isLoadingTasks ? (
            <div className="py-8 text-center text-xs text-slate-400 flex items-center justify-center gap-2">
              <i className="pi pi-spin pi-spinner text-indigo-500" />
              <span>{t('tasks.loadingTasks')}</span>
            </div>
          ) : totalTasks === 0 ? (
            <div className="py-10 text-center border border-dashed border-slate-200 dark:border-slate-800 rounded-xl">
              <i className="pi pi-folder-open text-3xl text-slate-300 dark:text-slate-600 mb-2" />
              <p className="text-xs text-slate-500 dark:text-slate-400">
                {t('projects.emptyTasks')}
              </p>
              <button
                type="button"
                onClick={() => openCreateTaskModal(project.id)}
                className="mt-3 px-3 py-1.5 text-xs font-medium rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white transition-colors inline-flex items-center gap-1"
              >
                <i className="pi pi-plus text-xs" />
                {t('projects.addTaskToProject')}
              </button>
            </div>
          ) : (
            /* Render Grouped by Status or Filtered */
            <div className="space-y-4">
              {STATUS_KEYS.filter((s) => activeStatusTab === 'all' || activeStatusTab === s).map(
                (status) => {
                  const list = tasksByStatus[status]
                  if (activeStatusTab === 'all' && list.length === 0) return null

                  return (
                    <div key={status} className="space-y-2">
                      <div className="flex items-center gap-2 px-1">
                        <span className={`w-2 h-2 rounded-full ${STATUS_COLORS[status].dot}`} />
                        <h4 className="text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-300">
                          {t(`tasks.status.${status}`)}
                        </h4>
                        <span className="text-[10px] font-mono text-slate-400 font-medium">
                          ({list.length})
                        </span>
                      </div>

                      {list.length === 0 ? (
                        <p className="text-[11px] text-slate-400 italic px-2 py-1">
                          {t('projects.noTasksInStatus')}
                        </p>
                      ) : (
                        <div className="divide-y divide-slate-100 dark:divide-slate-800/60 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/40 dark:bg-slate-950/30 overflow-hidden">
                          {list.map((task) => {
                            const isDone = task.status === 'done'
                            const priorityMeta =
                              PRIORITY_ICONS[task.priority] || PRIORITY_ICONS.medium

                            return (
                              <div
                                key={task.id}
                                onClick={() => openDrawer(task)}
                                className="group flex items-center justify-between p-3 hover:bg-white dark:hover:bg-slate-800/70 cursor-pointer transition-colors text-xs gap-3"
                              >
                                {/* Left: Complete Checkbox, Star Focus, Title */}
                                <div className="flex items-center gap-3 min-w-0 flex-1">
                                  <button
                                    type="button"
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
                                        : 'border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 hover:border-indigo-500'
                                    }`}
                                    title={isDone ? t('tasks.reopenTask') : t('tasks.completeTask')}
                                  >
                                    {isDone && <i className="pi pi-check text-[9px]" />}
                                  </button>

                                  <button
                                    type="button"
                                    onClick={(e) => {
                                      e.stopPropagation()
                                      toggleFocusMutation.mutate(task.id)
                                    }}
                                    className={`p-0.5 rounded transition-colors ${
                                      task.is_focus
                                        ? 'text-amber-500 dark:text-amber-400'
                                        : 'text-slate-400 dark:text-slate-600 opacity-0 group-hover:opacity-100 hover:text-slate-600'
                                    }`}
                                    title={
                                      task.is_focus
                                        ? t('focus.removeFromFocus')
                                        : t('focus.pinToFocus')
                                    }
                                  >
                                    <i
                                      className={`pi ${task.is_focus ? 'pi-star-fill' : 'pi-star'} text-[11px]`}
                                    />
                                  </button>

                                  <div className="min-w-0 flex-1">
                                    <span
                                      className={`font-medium truncate ${
                                        isDone
                                          ? 'line-through text-slate-400 dark:text-slate-500'
                                          : 'text-slate-800 dark:text-slate-200 group-hover:text-indigo-600 dark:group-hover:text-indigo-300'
                                      }`}
                                    >
                                      {task.title}
                                    </span>

                                    {task.last_progress_note && (
                                      <p className="text-[11px] text-indigo-600 dark:text-indigo-400 truncate mt-0.5 flex items-center gap-1 italic">
                                        <i className="pi pi-history text-[9px]" />
                                        <span>{task.last_progress_note}</span>
                                      </p>
                                    )}
                                  </div>
                                </div>

                                {/* Right: Subtasks, Priority, Due Date */}
                                <div className="flex items-center gap-3 shrink-0">
                                  {task.subtasks_count > 0 && (
                                    <span className="hidden sm:inline-block text-[10px] font-mono text-slate-500">
                                      ✓ {task.subtasks_completed}/{task.subtasks_count}
                                    </span>
                                  )}

                                  <span
                                    className={`inline-flex items-center gap-1 text-[11px] ${priorityMeta.class}`}
                                    title={t(`tasks.priority.${task.priority}`)}
                                  >
                                    <i className={`pi ${priorityMeta.icon} text-[10px]`} />
                                  </span>

                                  {task.due_date && (
                                    <span className="hidden md:inline-block text-[10px] font-mono text-slate-400">
                                      {new Date(task.due_date).toLocaleDateString(
                                        locale === 'id' ? 'id-ID' : 'en-US',
                                        { day: 'numeric', month: 'short' }
                                      )}
                                    </span>
                                  )}
                                </div>
                              </div>
                            )
                          })}
                        </div>
                      )}
                    </div>
                  )
                }
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
export default ProjectAccordionItem
