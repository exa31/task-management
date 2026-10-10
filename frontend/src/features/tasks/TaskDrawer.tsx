import { Calendar } from 'primereact/calendar'
import React, { useEffect, useState } from 'react'

import { getLocalizedErrorMessage } from '@/api'
import { confirmDialog } from '@/components'
import { TaskReminderSection } from '@/features/reminder'
import {
  useAddSubtask,
  useDeleteSubtask,
  useDeleteTask,
  useLogProgress,
  useProjects,
  useTask,
  useToggleFocus,
  useToggleSubtask,
  useUpdateTask
} from '@/hooks'
import { useTranslation } from '@/i18n'
import { useTaskStore } from '@/stores'
import type { TaskPriority, TaskStatus } from '@/types'

export const TaskDrawer: React.FC = () => {
  const { selectedTask, isDrawerOpen, closeDrawer } = useTaskStore()
  const { t, locale } = useTranslation()

  // Fetch live fresh task data for the selected task
  const { data: queriedTask } = useTask(selectedTask?.id)
  const { data: projects = [] } = useProjects()
  const task = queriedTask || selectedTask

  const updateTaskMutation = useUpdateTask()
  const deleteTaskMutation = useDeleteTask()
  const toggleFocusMutation = useToggleFocus()
  const logProgressMutation = useLogProgress()
  const addSubtaskMutation = useAddSubtask()
  const toggleSubtaskMutation = useToggleSubtask()
  const deleteSubtaskMutation = useDeleteSubtask()

  const [title, setTitle] = useState('')
  const [titleError, setTitleError] = useState<string | null>(null)
  const [description, setDescription] = useState('')
  const [status, setStatus] = useState<TaskStatus>('todo')
  const [priority, setPriority] = useState<TaskPriority>('medium')
  const [newSubtaskTitle, setNewSubtaskTitle] = useState('')
  const [newProgressNote, setNewProgressNote] = useState('')
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  useEffect(() => {
    if (task) {
      setTitle(task.title)
      setTitleError(null)
      setErrorMsg(null)
      setDescription(task.description || '')
      setStatus(task.status)
      setPriority(task.priority)
      setNewProgressNote('')
    }
  }, [task])

  if (!isDrawerOpen || !task) return null

  const handleSaveMain = async () => {
    if (!title.trim()) {
      setTitleError(t('taskDrawer.titleRequired'))
      return
    }
    setTitleError(null)
    setErrorMsg(null)
    try {
      await updateTaskMutation.mutateAsync({
        id: task.id,
        payload: {
          title: title.trim(),
          description,
          status,
          priority
        }
      })
    } catch (err) {
      setErrorMsg(getLocalizedErrorMessage(err, t))
    }
  }

  const handleAddSubtaskSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newSubtaskTitle.trim()) return
    await addSubtaskMutation.mutateAsync({
      taskId: task.id,
      title: newSubtaskTitle.trim()
    })
    setNewSubtaskTitle('')
  }

  const handleLogProgressSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newProgressNote.trim()) return
    await logProgressMutation.mutateAsync({
      id: task.id,
      payload: {
        progress_note: newProgressNote.trim(),
        status
      }
    })
    setNewProgressNote('')
  }

  const handleDelete = () => {
    confirmDialog({
      header: t('taskDrawer.deleteTaskConfirmHeader'),
      message: t('taskDrawer.deleteTaskConfirmMessage'),
      icon: 'pi pi-trash text-rose-500',
      defaultFocus: 'reject',
      acceptClassName: 'p-button-danger',
      acceptLabel: t('common.delete'),
      rejectLabel: t('common.cancel'),
      accept: async () => {
        await deleteTaskMutation.mutateAsync(task.id)
        closeDrawer()
      }
    })
  }

  const handleDeleteSubtask = (subtaskId: number, subtaskTitle: string) => {
    confirmDialog({
      header: t('taskDrawer.deleteSubtaskConfirmHeader'),
      message: t('taskDrawer.deleteSubtaskConfirmMessage', { title: subtaskTitle }),
      icon: 'pi pi-trash text-rose-500',
      defaultFocus: 'reject',
      acceptClassName: 'p-button-danger',
      acceptLabel: t('common.delete'),
      rejectLabel: t('common.cancel'),
      accept: () => {
        deleteSubtaskMutation.mutate({
          taskId: task.id,
          subtaskId
        })
      }
    })
  }

  return (
    <div className="fixed inset-0 z-50 overflow-hidden">
      {/* Backdrop */}
      <div
        onClick={closeDrawer}
        className="fixed inset-0 bg-slate-950/40 dark:bg-slate-950/70 backdrop-blur-sm transition-opacity cursor-pointer"
      />

      {/* Drawer Panel */}
      <div className="fixed inset-y-0 right-0 max-w-full flex pl-10 pointer-events-none">
        <div className="w-screen max-w-lg bg-white dark:bg-slate-900 border-l border-slate-200 dark:border-slate-800 shadow-2xl flex flex-col text-slate-900 dark:text-slate-100 animate-slide-left pointer-events-auto">
          {/* Header */}
          <div className="flex items-center justify-between p-4 border-b border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950/50">
            <div className="flex items-center gap-2">
              <button
                onClick={() => toggleFocusMutation.mutate(task.id)}
                className={`p-1.5 rounded-lg border transition-colors cursor-pointer ${
                  task.is_focus
                    ? 'border-amber-500/40 bg-amber-500/10 text-amber-500'
                    : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-400 hover:text-amber-500'
                }`}
                title={task.is_focus ? t('focus.removeFromFocus') : t('focus.pinToFocus')}
              >
                <i className={`pi ${task.is_focus ? 'pi-star-fill' : 'pi-star'} text-xs`} />
              </button>
              <span className="text-xs font-mono text-slate-500">
                {t('taskDrawer.taskIdBadge', { id: task.id })}
              </span>
            </div>

            <div className="flex items-center gap-2">
              <button
                onClick={handleDelete}
                disabled={deleteTaskMutation.isPending}
                className="p-1.5 rounded-lg hover:bg-rose-50 dark:hover:bg-rose-950/50 text-slate-400 hover:text-rose-500 dark:hover:text-rose-400 cursor-pointer transition-colors"
                title={t('taskDrawer.deleteTask')}
              >
                <i className="pi pi-trash text-xs" />
              </button>
              <button
                onClick={closeDrawer}
                className="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 cursor-pointer transition-colors"
                title={t('taskDrawer.closeEsc')}
              >
                <i className="pi pi-times text-xs" />
              </button>
            </div>
          </div>

          {/* Body Content */}
          <div className="flex-1 overflow-y-auto p-5 space-y-5">
            {errorMsg && (
              <div className="flex items-start gap-2 rounded-lg border border-rose-500/20 bg-rose-500/10 p-3 text-xs text-rose-600 dark:text-rose-400">
                <i className="pi pi-exclamation-circle text-sm mt-0.5 shrink-0" />
                <span className="flex-1 font-medium">{errorMsg}</span>
                <button
                  type="button"
                  onClick={() => setErrorMsg(null)}
                  className="text-rose-500/70 hover:text-rose-500 cursor-pointer"
                >
                  <i className="pi pi-times text-xs" />
                </button>
              </div>
            )}

            {/* Controls: Status, Priority, Due Date & Project */}
            <div className="grid grid-cols-2 gap-3 p-3 rounded-xl bg-slate-50 dark:bg-slate-950/60 border border-slate-200 dark:border-slate-800">
              <div>
                <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
                  {t('taskDrawer.status')}
                </label>
                <select
                  value={status}
                  onChange={(e) => {
                    const next = e.target.value as TaskStatus
                    setStatus(next)
                    updateTaskMutation.mutate({ id: task.id, payload: { status: next } })
                  }}
                  className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-2.5 py-1.5 text-xs text-slate-800 dark:text-slate-200 focus:border-indigo-500 focus:outline-none transition-colors"
                >
                  <option
                    value="backlog"
                    className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                  >
                    {t('tasks.status.backlog')}
                  </option>
                  <option
                    value="todo"
                    className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                  >
                    {t('tasks.status.todo')}
                  </option>
                  <option
                    value="in_progress"
                    className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                  >
                    {t('tasks.status.in_progress')}
                  </option>
                  <option
                    value="review"
                    className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                  >
                    {t('tasks.status.review')}
                  </option>
                  <option
                    value="done"
                    className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                  >
                    {t('tasks.status.done')}
                  </option>
                </select>
              </div>

              <div>
                <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
                  {t('taskDrawer.priority')}
                </label>
                <select
                  value={priority}
                  onChange={(e) => {
                    const next = e.target.value as TaskPriority
                    setPriority(next)
                    updateTaskMutation.mutate({ id: task.id, payload: { priority: next } })
                  }}
                  className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-2.5 py-1.5 text-xs text-slate-800 dark:text-slate-200 focus:border-indigo-500 focus:outline-none transition-colors"
                >
                  <option
                    value="low"
                    className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                  >
                    {t('tasks.priority.low')}
                  </option>
                  <option
                    value="medium"
                    className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                  >
                    {t('tasks.priority.medium')}
                  </option>
                  <option
                    value="high"
                    className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                  >
                    {t('tasks.priority.high')}
                  </option>
                  <option
                    value="urgent"
                    className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                  >
                    {t('tasks.priority.urgent')}
                  </option>
                </select>
              </div>

              <div>
                <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
                  {t('taskDrawer.dueDate')}
                </label>
                <Calendar
                  value={task.due_date ? new Date(task.due_date) : null}
                  onChange={(e) => {
                    const nextDate = e.value as Date | null
                    updateTaskMutation.mutate({
                      id: task.id,
                      payload: { due_date: nextDate ? nextDate.toISOString() : null }
                    })
                  }}
                  showTime
                  hourFormat="24"
                  showIcon
                  dateFormat="dd/mm/yy"
                  locale={locale}
                  placeholder={t('taskDrawer.dueDate')}
                  className="w-full text-xs"
                  inputClassName="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-2.5 py-1.5 text-xs text-slate-800 dark:text-slate-200 focus:border-indigo-500 focus:outline-none transition-colors"
                  touchUI={false}
                  showButtonBar
                />
              </div>

              <div>
                <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
                  {t('taskDrawer.projectLabel')}
                </label>
                <select
                  value={task.project_id || ''}
                  onChange={(e) => {
                    const next = e.target.value ? parseInt(e.target.value, 10) : null
                    updateTaskMutation.mutate({ id: task.id, payload: { project_id: next } })
                  }}
                  className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-2.5 py-1.5 text-xs text-slate-800 dark:text-slate-200 focus:border-indigo-500 focus:outline-none transition-colors"
                >
                  <option
                    value=""
                    className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                  >
                    {t('taskDrawer.noProjectOption')}
                  </option>
                  {projects.map((p) => (
                    <option
                      key={p.id}
                      value={String(p.id)}
                      className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                    >
                      📁 {p.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {/* Title & Description */}
            <div>
              <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
                {t('taskDrawer.titleLabel')}{' '}
                <span className="text-rose-500 font-bold ml-0.5">*</span>
              </label>
              <input
                type="text"
                value={title}
                onChange={(e) => {
                  setTitle(e.target.value)
                  if (e.target.value.trim()) setTitleError(null)
                }}
                onBlur={handleSaveMain}
                placeholder={`${t('taskDrawer.titleLabel')}...`}
                className={`w-full text-base font-semibold text-slate-900 dark:text-slate-100 bg-transparent border-b ${
                  titleError
                    ? 'border-rose-500'
                    : 'border-slate-200 dark:border-slate-800 hover:border-slate-400 dark:hover:border-slate-700 focus:border-indigo-500'
                } pb-1 focus:outline-none transition-colors`}
              />
              {titleError && (
                <p className="text-[11px] text-rose-500 mt-1 flex items-center gap-1">
                  <i className="pi pi-exclamation-circle text-[10px]" /> {titleError}
                </p>
              )}
              <textarea
                rows={3}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                onBlur={handleSaveMain}
                placeholder={t('taskDrawer.descriptionPlaceholder')}
                className="w-full mt-2 text-xs text-slate-700 dark:text-slate-300 bg-transparent border-none focus:ring-0 p-0 placeholder-slate-400 dark:placeholder-slate-600 resize-none focus:outline-none"
              />
            </div>

            {/* Progress Log */}
            <div className="rounded-xl border border-indigo-200 dark:border-indigo-500/30 bg-indigo-50/70 dark:bg-indigo-950/20 p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs font-semibold text-indigo-900 dark:text-indigo-300 flex items-center gap-1.5">
                  <i className="pi pi-history text-xs text-indigo-500 dark:text-indigo-400" />
                  {t('taskDrawer.lastProgressNote')}
                </span>
                {task.last_progress_at && (
                  <span className="text-[10px] text-slate-500 dark:text-slate-400">
                    {new Date(task.last_progress_at).toLocaleTimeString(
                      locale === 'id' ? 'id-ID' : 'en-US',
                      {
                        hour: '2-digit',
                        minute: '2-digit',
                        day: 'numeric',
                        month: 'short'
                      }
                    )}
                  </span>
                )}
              </div>

              {task.last_progress_note ? (
                <div className="p-2.5 rounded-lg bg-white dark:bg-slate-900/90 border border-slate-200 dark:border-slate-800 text-xs text-slate-800 dark:text-slate-200 mb-3 shadow-xs">
                  <p className="font-medium text-indigo-600 dark:text-indigo-300 text-[11px] mb-0.5">
                    {t('taskDrawer.lastWorkedOn')}
                  </p>
                  <p className="leading-relaxed">{task.last_progress_note}</p>
                </div>
              ) : (
                <p className="text-[11px] text-slate-500 mb-3">{t('taskDrawer.noLogYet')}</p>
              )}

              {/* Quick Update Log Form */}
              <form onSubmit={handleLogProgressSubmit} className="space-y-2">
                <input
                  type="text"
                  value={newProgressNote}
                  onChange={(e) => setNewProgressNote(e.target.value)}
                  placeholder={t('taskDrawer.progressPlaceholder')}
                  className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-3 py-1.5 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none"
                />
                <button
                  type="submit"
                  disabled={logProgressMutation.isPending || !newProgressNote.trim()}
                  className="w-full py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-xs font-medium text-white flex items-center justify-center gap-1.5 transition-colors shadow-xs cursor-pointer"
                >
                  {logProgressMutation.isPending && (
                    <i className="pi pi-spin pi-spinner text-[10px]" />
                  )}
                  {t('taskDrawer.saveProgressNote')}
                </button>
              </form>
            </div>

            {/* Checklist Subtasks */}
            <div>
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                  <i className="pi pi-check-square text-xs text-indigo-500 dark:text-indigo-400" />
                  {t('taskDrawer.subtasksChecklist', {
                    completed: task.subtasks_completed,
                    total: task.subtasks_count
                  })}
                </span>
                <span className="text-xs font-mono text-indigo-600 dark:text-indigo-400 font-semibold">
                  {task.progress_percent}%
                </span>
              </div>

              {/* Progress Bar */}
              <div className="w-full h-1.5 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden mb-3">
                <div
                  className="h-full bg-indigo-500 transition-all duration-300"
                  style={{ width: `${task.progress_percent}%` }}
                />
              </div>

              {/* Subtask Items */}
              <div className="space-y-1.5 mb-3">
                {task.subtasks?.map((st) => (
                  <div
                    key={st.id}
                    className="group flex items-center justify-between p-2 rounded-lg bg-slate-50 dark:bg-slate-950/60 hover:bg-slate-100/80 dark:hover:bg-slate-950 border border-slate-200 dark:border-slate-800/80 text-xs"
                  >
                    <label className="flex items-center gap-2.5 flex-1 cursor-pointer min-w-0">
                      <input
                        type="checkbox"
                        checked={st.is_completed}
                        onChange={(e) =>
                          toggleSubtaskMutation.mutate({
                            taskId: task.id,
                            subtaskId: st.id,
                            isCompleted: e.target.checked
                          })
                        }
                        className="rounded border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-indigo-600 focus:ring-0 shrink-0 cursor-pointer"
                      />
                      <span
                        className={`truncate ${
                          st.is_completed
                            ? 'line-through text-slate-400 dark:text-slate-500'
                            : 'text-slate-800 dark:text-slate-200'
                        }`}
                      >
                        {st.title}
                      </span>
                    </label>

                    <button
                      type="button"
                      onClick={() => handleDeleteSubtask(st.id, st.title)}
                      className="p-1 text-slate-400 dark:text-slate-600 hover:text-rose-500 dark:hover:text-rose-400 opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
                      title={t('taskDrawer.deleteSubtaskConfirmHeader')}
                    >
                      <i className="pi pi-times text-[10px]" />
                    </button>
                  </div>
                ))}
              </div>

              {/* Add Subtask Input */}
              <form onSubmit={handleAddSubtaskSubmit} className="flex gap-2">
                <input
                  type="text"
                  value={newSubtaskTitle}
                  onChange={(e) => setNewSubtaskTitle(e.target.value)}
                  placeholder={t('taskDrawer.addSubtaskPlaceholder')}
                  className="flex-1 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 px-3 py-1.5 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none"
                />
                <button
                  type="submit"
                  disabled={addSubtaskMutation.isPending || !newSubtaskTitle.trim()}
                  className="px-3 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 border border-slate-200 dark:border-slate-700 disabled:opacity-50 text-xs font-medium text-slate-700 dark:text-slate-200 flex items-center gap-1 transition-colors cursor-pointer"
                >
                  {addSubtaskMutation.isPending && (
                    <i className="pi pi-spin pi-spinner text-[10px]" />
                  )}
                  {t('taskDrawer.add')}
                </button>
              </form>
            </div>

            {/* Task Reminders Section */}
            <div className="pt-3 border-t border-slate-200 dark:border-slate-800/80">
              <TaskReminderSection
                taskId={task.id}
                taskTitle={task.title}
                dueDate={task.due_date}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
