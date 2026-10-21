import { Calendar } from 'primereact/calendar'
import React, { useCallback, useEffect, useRef, useState } from 'react'

import { getLocalizedErrorMessage } from '@/api'
import { confirmDialog } from '@/components'
import { TaskReminderSection } from '@/features/reminder'
import { TaskActivityTimeline } from '@/features/tasks/TaskActivityTimeline'
import {
  useAddSubtask,
  useDeleteSubtask,
  useDeleteTask,
  useLogProgress,
  useProjects,
  useTask,
  useToggleFocus,
  useToggleSubtask,
  useUpdateSubtask,
  useUpdateTask
} from '@/hooks'
import { useTranslation } from '@/i18n'
import { useTaskStore } from '@/stores'
import type { RecurrencePattern, Subtask, TaskPriority, TaskStatus } from '@/types'

interface SubtaskItemRowProps {
  subtask: Subtask
  taskId: number
  isRecurring?: boolean
  onToggle: (completed: boolean) => void
  onDelete: (id: number, title: string) => void
  onUpdate: (params: {
    taskId: number
    subtaskId: number
    title?: string
    description?: string
    dueDate?: string | null
  }) => Promise<unknown>
}

const SubtaskItemRow: React.FC<SubtaskItemRowProps> = ({
  subtask,
  taskId,
  isRecurring = false,
  onToggle,
  onDelete,
  onUpdate
}) => {
  const { t, locale } = useTranslation()
  const [isExpanded, setIsExpanded] = useState(false)
  const [editTitle, setEditTitle] = useState(subtask.title)
  const [editDescription, setEditDescription] = useState(subtask.description || '')
  const [editDueDate, setEditDueDate] = useState<Date | null>(
    subtask.due_date ? new Date(subtask.due_date) : null
  )
  const [isSaving, setIsSaving] = useState(false)

  useEffect(() => {
    setEditTitle(subtask.title)
    setEditDescription(subtask.description || '')
    setEditDueDate(subtask.due_date ? new Date(subtask.due_date) : null)
  }, [subtask.title, subtask.description, subtask.due_date])

  const handleSave = async () => {
    if (!editTitle.trim()) return
    setIsSaving(true)
    try {
      await onUpdate({
        taskId,
        subtaskId: subtask.id,
        title: editTitle.trim(),
        description: editDescription.trim(),
        dueDate: isRecurring ? null : (editDueDate ? editDueDate.toISOString() : null)
      })
      setIsExpanded(false)
    } finally {
      setIsSaving(false)
    }
  }

  const handleCancel = () => {
    setEditTitle(subtask.title)
    setEditDescription(subtask.description || '')
    setEditDueDate(subtask.due_date ? new Date(subtask.due_date) : null)
    setIsExpanded(false)
  }

  const formattedDueDate = subtask.due_date
    ? new Date(subtask.due_date).toLocaleDateString(locale === 'id' ? 'id-ID' : 'en-US', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      })
    : null

  return (
    <div
      className={`rounded-lg border transition-all duration-200 ${
        isExpanded
          ? 'bg-slate-50 dark:bg-slate-900/90 border-indigo-300 dark:border-indigo-800 shadow-xs'
          : 'bg-slate-50/70 dark:bg-slate-950/60 hover:bg-slate-100/80 dark:hover:bg-slate-950 border-slate-200 dark:border-slate-800/80'
      }`}
    >
      {/* Header / Summary Row */}
      <div className="flex items-start justify-between gap-2 p-2.5 text-xs">
        <label className="flex items-start gap-2.5 flex-1 min-w-0 cursor-pointer">
          <input
            type="checkbox"
            checked={subtask.is_completed}
            onChange={(e) => onToggle(e.target.checked)}
            className="mt-0.5 rounded border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-indigo-600 focus:ring-0 shrink-0 cursor-pointer"
          />
          <div className="flex-1 min-w-0 space-y-1">
            <span
              className={`block break-words leading-snug ${
                subtask.is_completed
                  ? 'line-through text-slate-400 dark:text-slate-500'
                  : 'text-slate-800 dark:text-slate-200 font-medium'
              }`}
            >
              {subtask.title}
            </span>

            {/* Badges / indicators when collapsed */}
            {!isExpanded && (subtask.description || (!isRecurring && subtask.due_date)) && (
              <div className="flex flex-wrap items-center gap-1.5 pt-0.5">
                {subtask.description && (
                  <span
                    className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] text-slate-600 dark:text-slate-300 bg-slate-200/60 dark:bg-slate-800/80 max-w-full"
                    title={subtask.description}
                  >
                    <i className="pi pi-align-left text-[9px] shrink-0 text-slate-500 dark:text-slate-400" />
                    <span className="truncate max-w-[200px]">{subtask.description}</span>
                  </span>
                )}
                {!isRecurring && formattedDueDate && (
                  <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950/60 border border-indigo-200/50 dark:border-indigo-800/40">
                    <i className="pi pi-calendar text-[9px]" />
                    {formattedDueDate}
                  </span>
                )}
              </div>
            )}
          </div>
        </label>

        <div className="flex items-center gap-1 shrink-0 pt-0.5">
          <button
            type="button"
            onClick={() => setIsExpanded(!isExpanded)}
            data-pr-tooltip={
              isExpanded
                ? t('taskDrawer.subtaskCollapseTooltip')
                : t('taskDrawer.subtaskExpandTooltip')
            }
            className={`p-1 rounded text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors cursor-pointer ${
              isExpanded
                ? 'text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950/50'
                : ''
            }`}
          >
            <i className={`pi ${isExpanded ? 'pi-chevron-up' : 'pi-chevron-down'} text-[10px]`} />
          </button>

          <button
            type="button"
            onClick={() => onDelete(subtask.id, subtask.title)}
            data-pr-tooltip={t('taskDrawer.deleteSubtaskTooltip')}
            className="p-1 rounded text-slate-400 dark:text-slate-500 hover:text-rose-500 dark:hover:text-rose-400 transition-colors cursor-pointer"
          >
            <i className="pi pi-times text-[10px]" />
          </button>
        </div>
      </div>

      {/* Expanded Details Form */}
      {isExpanded && (
        <div className="px-2.5 pb-2.5 pt-1 border-t border-slate-200/70 dark:border-slate-800/70 space-y-2.5 animate-fade-in text-xs">
          <div>
            <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
              {t('taskDrawer.subtaskTitlePlaceholder')}
            </label>
            <input
              type="text"
              value={editTitle}
              onChange={(e) => setEditTitle(e.target.value)}
              className="w-full rounded-md border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:border-indigo-500 focus:outline-none"
            />
          </div>

          <div>
            <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
              {t('taskDrawer.subtaskDetails')}
            </label>
            <textarea
              rows={2}
              value={editDescription}
              onChange={(e) => setEditDescription(e.target.value)}
              placeholder={t('taskDrawer.subtaskDescriptionPlaceholder')}
              className="w-full rounded-md border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none resize-none"
            />
          </div>

          {!isRecurring && (
            <div>
              <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
                {t('taskDrawer.subtaskDueDate')}
              </label>
              <Calendar
                value={editDueDate}
                onChange={(e) => setEditDueDate(e.value as Date | null)}
                showTime
                hourFormat="24"
                placeholder={t('taskDrawer.subtaskDueDatePlaceholder')}
                className="w-full text-xs"
                inputClassName="w-full rounded-md border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:border-indigo-500 focus:outline-none"
                showIcon
                showButtonBar
              />
            </div>
          )}

          <div className="flex items-center justify-end gap-1.5 pt-1">
            <button
              type="button"
              onClick={handleCancel}
              className="px-2.5 py-1 rounded text-xs text-slate-600 dark:text-slate-400 hover:bg-slate-200/50 dark:hover:bg-slate-800 transition-colors cursor-pointer"
            >
              {t('common.cancel')}
            </button>
            <button
              type="button"
              disabled={isSaving || !editTitle.trim()}
              onClick={handleSave}
              className="px-3 py-1 rounded bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium disabled:opacity-50 flex items-center gap-1 shadow-xs transition-colors cursor-pointer"
            >
              {isSaving && <i className="pi pi-spin pi-spinner text-[10px]" />}
              {t('common.save')}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

export const TaskDrawer: React.FC = () => {
  const { selectedTask, isDrawerOpen, closeDrawer } = useTaskStore()
  const { t, locale } = useTranslation()
  const [activeTab, setActiveTab] = useState<'details' | 'activities'>('details')

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
  const updateSubtaskMutation = useUpdateSubtask()
  const deleteSubtaskMutation = useDeleteSubtask()

  const [title, setTitle] = useState('')
  const [titleError, setTitleError] = useState<string | null>(null)
  const [description, setDescription] = useState('')
  const [status, setStatus] = useState<TaskStatus>('todo')
  const [priority, setPriority] = useState<TaskPriority>('medium')
  const [dueDate, setDueDate] = useState<Date | null>(null)
  const [projectId, setProjectId] = useState<number | null>(null)
  const [newSubtaskTitle, setNewSubtaskTitle] = useState('')
  const [newSubtaskDescription, setNewSubtaskDescription] = useState('')
  const [newSubtaskDueDate, setNewSubtaskDueDate] = useState<Date | null>(null)
  const [isAddingWithDetails, setIsAddingWithDetails] = useState(false)
  const [newProgressNote, setNewProgressNote] = useState('')
  const [isRecurring, setIsRecurring] = useState(false)
  const [recurrencePattern, setRecurrencePattern] = useState<RecurrencePattern>('daily')
  const [recurrenceInterval, setRecurrenceInterval] = useState<number>(1)
  const [recurrenceEndDate, setRecurrenceEndDate] = useState<Date | null>(null)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  const prevTaskIdRef = useRef<number | null>(null)

  useEffect(() => {
    if (task && task.id !== prevTaskIdRef.current) {
      prevTaskIdRef.current = task.id
      setTitle(task.title)
      setTitleError(null)
      setErrorMsg(null)
      setDescription(task.description || '')
      setStatus(task.status)
      setPriority(task.priority)
      setDueDate(task.due_date ? new Date(task.due_date) : null)
      setProjectId(task.project_id ?? null)
      setIsRecurring(task.is_recurring ?? false)
      setRecurrencePattern(task.recurrence_pattern || 'daily')
      setRecurrenceInterval(task.recurrence_interval || 1)
      setRecurrenceEndDate(task.recurrence_end_date ? new Date(task.recurrence_end_date) : null)
      setNewProgressNote('')
    }
  }, [task])

  const isDirty =
    Boolean(task) &&
    (title.trim() !== task!.title ||
      description !== (task!.description || '') ||
      status !== task!.status ||
      priority !== task!.priority ||
      (dueDate ? dueDate.toISOString() : null) !==
        (task!.due_date ? new Date(task!.due_date).toISOString() : null) ||
      projectId !== (task!.project_id ?? null) ||
      isRecurring !== (task!.is_recurring ?? false) ||
      (isRecurring && recurrencePattern !== (task!.recurrence_pattern || 'daily')) ||
      (isRecurring && recurrenceInterval !== (task!.recurrence_interval || 1)) ||
      (isRecurring &&
        (recurrenceEndDate ? recurrenceEndDate.toISOString() : null) !==
          (task!.recurrence_end_date ? new Date(task!.recurrence_end_date).toISOString() : null)))

  const handleResetMain = () => {
    if (!task) return
    setTitle(task.title)
    setTitleError(null)
    setErrorMsg(null)
    setDescription(task.description || '')
    setStatus(task.status)
    setPriority(task.priority)
    setDueDate(task.due_date ? new Date(task.due_date) : null)
    setProjectId(task.project_id ?? null)
    setIsRecurring(task.is_recurring ?? false)
    setRecurrencePattern(task.recurrence_pattern || 'daily')
    setRecurrenceInterval(task.recurrence_interval || 1)
    setRecurrenceEndDate(task.recurrence_end_date ? new Date(task.recurrence_end_date) : null)
  }

  const handleStatusChange = (newStatus: TaskStatus) => {
    setStatus(newStatus)
    setErrorMsg(null)
  }

  const handlePriorityChange = (newPriority: TaskPriority) => {
    setPriority(newPriority)
    setErrorMsg(null)
  }

  const handleProjectChange = (newProjectId: number | null) => {
    setProjectId(newProjectId)
    setErrorMsg(null)
  }

  const handleDueDateChange = (newDate: Date | null) => {
    setDueDate(newDate)
    setErrorMsg(null)
  }

  const handleTitleBlur = () => {
    const trimmed = title.trim()
    if (!trimmed) {
      setTitleError(t('taskDrawer.titleRequired'))
    } else {
      setTitleError(null)
    }
  }

  const handleClose = useCallback(() => {
    closeDrawer()
  }, [closeDrawer])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isDrawerOpen) {
        handleClose()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isDrawerOpen, handleClose])

  const handleSubmitMain = async (e?: React.FormEvent) => {
    if (e) e.preventDefault()
    if (!task) return
    if (!title.trim()) {
      setTitleError(t('taskDrawer.titleRequired'))
      return
    }
    setTitleError(null)
    setErrorMsg(null)
    try {
      const updated = await updateTaskMutation.mutateAsync({
        id: task.id,
        payload: {
          title: title.trim(),
          description,
          status,
          priority,
          due_date: isRecurring ? null : (dueDate ? dueDate.toISOString() : null),
          project_id: projectId,
          is_recurring: isRecurring,
          recurrence_pattern: isRecurring ? recurrencePattern : null,
          recurrence_interval: isRecurring ? recurrenceInterval : null,
          recurrence_end_date:
            isRecurring && recurrenceEndDate ? recurrenceEndDate.toISOString() : null
        }
      })
      if (updated) {
        setTitle(updated.title)
        setDescription(updated.description || '')
        setStatus(updated.status)
        setPriority(updated.priority)
        setDueDate(updated.due_date ? new Date(updated.due_date) : null)
        setProjectId(updated.project_id ?? null)
        setIsRecurring(updated.is_recurring ?? false)
        setRecurrencePattern(updated.recurrence_pattern || 'daily')
        setRecurrenceInterval(updated.recurrence_interval || 1)
        setRecurrenceEndDate(
          updated.recurrence_end_date ? new Date(updated.recurrence_end_date) : null
        )
      }
    } catch (err) {
      setErrorMsg(getLocalizedErrorMessage(err, t))
    }
  }

  const handleAddSubtaskSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!task || !newSubtaskTitle.trim()) return
    await addSubtaskMutation.mutateAsync({
      taskId: task.id,
      title: newSubtaskTitle.trim(),
      description: newSubtaskDescription.trim() || undefined,
      dueDate: isRecurring ? undefined : (newSubtaskDueDate ? newSubtaskDueDate.toISOString() : undefined)
    })
    setNewSubtaskTitle('')
    setNewSubtaskDescription('')
    setNewSubtaskDueDate(null)
    setIsAddingWithDetails(false)
  }

  const handleLogProgressSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!task || !newProgressNote.trim()) return
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
    if (!task) return
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
    if (!task) return
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

  if (!isDrawerOpen || !task) return null

  return (
    <div className="fixed inset-0 z-50 overflow-hidden">
      {/* Backdrop */}
      <div
        onClick={handleClose}
        className="fixed inset-0 bg-slate-950/40 dark:bg-slate-950/70 backdrop-blur-sm transition-opacity cursor-pointer"
      />

      {/* Drawer Panel */}
      <div className="fixed inset-y-0 right-0 max-w-full flex pl-10 pointer-events-none">
        <div className="w-screen max-w-lg bg-white dark:bg-slate-900 border-l border-slate-200 dark:border-slate-800 shadow-2xl flex flex-col text-slate-900 dark:text-slate-100 animate-slide-left pointer-events-auto">
          {/* Header */}
          <div className="flex items-center justify-between p-4 border-b border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950/50">
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={() => toggleFocusMutation.mutate(task.id)}
                className={`p-1.5 rounded-lg border transition-colors cursor-pointer ${
                  task.is_focus
                    ? 'border-amber-500/40 bg-amber-500/10 text-amber-500'
                    : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-400 hover:text-amber-500'
                }`}
                data-pr-tooltip={
                  task.is_focus
                    ? t('taskDrawer.unpinFocusTooltip')
                    : t('taskDrawer.pinFocusTooltip')
                }
              >
                <i className={`pi ${task.is_focus ? 'pi-star-fill' : 'pi-star'} text-xs`} />
              </button>
              <span className="text-xs font-mono text-slate-500">
                {t('taskDrawer.taskIdBadge', { id: task.id })}
              </span>
            </div>

            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={handleDelete}
                disabled={deleteTaskMutation.isPending}
                className="p-1.5 rounded-lg hover:bg-rose-50 dark:hover:bg-rose-950/50 text-slate-400 hover:text-rose-500 dark:hover:text-rose-400 cursor-pointer transition-colors"
                data-pr-tooltip={t('taskDrawer.deleteTaskTooltip')}
              >
                <i className="pi pi-trash text-xs" />
              </button>
              <button
                type="button"
                onClick={handleClose}
                className="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 cursor-pointer transition-colors"
                data-pr-tooltip={t('taskDrawer.closeDrawerTooltip')}
              >
                <i className="pi pi-times text-xs" />
              </button>
            </div>
          </div>

          {/* Navigation Tabs */}
          <div className="flex border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50 px-5 pt-1.5 shrink-0">
            <button
              type="button"
              onClick={() => setActiveTab('details')}
              className={`pb-2.5 px-3 text-xs font-semibold border-b-2 transition-all flex items-center gap-1.5 cursor-pointer ${
                activeTab === 'details'
                  ? 'border-indigo-600 text-indigo-600 dark:border-indigo-400 dark:text-indigo-400'
                  : 'border-transparent text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
              }`}
            >
              <i className="pi pi-align-left text-xs" />
              <span>{t('taskDrawer.tabDetails')}</span>
            </button>
            <button
              type="button"
              onClick={() => setActiveTab('activities')}
              className={`pb-2.5 px-3 text-xs font-semibold border-b-2 transition-all flex items-center gap-1.5 cursor-pointer ${
                activeTab === 'activities'
                  ? 'border-indigo-600 text-indigo-600 dark:border-indigo-400 dark:text-indigo-400'
                  : 'border-transparent text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
              }`}
            >
              <i className="pi pi-history text-xs" />
              <span>{t('taskDrawer.tabActivities')}</span>
            </button>
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
                  data-pr-tooltip={t('common.cancel') || 'Dismiss'}
                  className="text-rose-500/70 hover:text-rose-500 cursor-pointer"
                >
                  <i className="pi pi-times text-xs" />
                </button>
              </div>
            )}

            {activeTab === 'activities' ? (
              <TaskActivityTimeline taskId={task.id} />
            ) : (
              <>
                {/* Controls: Status, Priority, Due Date & Project */}
                <div className="grid grid-cols-2 gap-3 p-3 rounded-xl bg-slate-50 dark:bg-slate-950/60 border border-slate-200 dark:border-slate-800">
                  <div>
                    <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
                      {t('taskDrawer.status')}
                    </label>
                    <select
                      value={status}
                      onChange={(e) => handleStatusChange(e.target.value as TaskStatus)}
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
                      onChange={(e) => handlePriorityChange(e.target.value as TaskPriority)}
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

                  {!isRecurring && (
                    <div>
                      <div className="flex items-center justify-between mb-1">
                        <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400">
                          {t('taskDrawer.dueDate')}
                        </label>
                        {dueDate && (
                          <button
                            type="button"
                            onClick={() => handleDueDateChange(null)}
                            className="text-[10px] text-slate-400 hover:text-rose-500 flex items-center gap-1 transition-colors cursor-pointer"
                            title={t('common.clear') || 'Clear'}
                          >
                            <i className="pi pi-times text-[9px]" />
                            <span>{t('common.clear') || 'Hapus'}</span>
                          </button>
                        )}
                      </div>
                      <Calendar
                        value={dueDate}
                        onChange={(e) => handleDueDateChange((e.value as Date) || null)}
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
                  )}

                  <div className={isRecurring ? 'col-span-2' : ''}>
                    <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
                      {t('taskDrawer.projectLabel')}
                    </label>
                    <select
                      value={projectId !== null && projectId !== undefined ? String(projectId) : ''}
                      onChange={(e) => {
                        const next = e.target.value ? parseInt(e.target.value, 10) : null
                        handleProjectChange(next)
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

                {/* Repeat Task Section */}
                <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-950/40 p-3 transition-all">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div
                        className={`w-6 h-6 rounded-lg flex items-center justify-center transition-colors ${
                          isRecurring
                            ? 'bg-indigo-600 text-white shadow-xs'
                            : 'bg-slate-200 dark:bg-slate-800 text-slate-500 dark:text-slate-400'
                        }`}
                      >
                        <i className="pi pi-sync text-[11px]" />
                      </div>
                      <span className="text-xs font-semibold text-slate-800 dark:text-slate-200">
                        {t('taskDrawer.repeatTaskLabel')}
                      </span>
                    </div>

                    <button
                      type="button"
                      role="switch"
                      aria-checked={isRecurring}
                      onClick={() => setIsRecurring(!isRecurring)}
                      className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                        isRecurring ? 'bg-indigo-600' : 'bg-slate-300 dark:bg-slate-700'
                      }`}
                    >
                      <span
                        className={`pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-md ring-0 transition duration-200 ease-in-out ${
                          isRecurring ? 'translate-x-4' : 'translate-x-0'
                        }`}
                      />
                    </button>
                  </div>

                  {isRecurring && (
                    <div className="mt-3 pt-3 border-t border-slate-200/80 dark:border-slate-800/80 space-y-3 animate-fade-in">
                      {/* Pattern Selector */}
                      <div>
                        <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
                          {t('taskDrawer.repeatPatternLabel')}
                        </label>
                        <div className="grid grid-cols-4 gap-1.5">
                          {(['daily', 'weekdays', 'weekly', 'monthly'] as const).map((pat) => (
                            <button
                              key={pat}
                              type="button"
                              onClick={() => setRecurrencePattern(pat)}
                              className={`px-2 py-1 rounded-lg text-xs font-medium transition-all text-center cursor-pointer ${
                                recurrencePattern === pat
                                  ? 'bg-indigo-600 text-white shadow-xs'
                                  : 'bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-800 hover:bg-slate-100 dark:hover:bg-slate-800'
                              }`}
                            >
                              {t(`taskDrawer.patterns.${pat}`)}
                            </button>
                          ))}
                        </div>
                      </div>

                      {/* Interval & End Date */}
                      <div className="grid grid-cols-2 gap-2.5">
                        <div>
                          <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
                            {t('taskDrawer.repeatIntervalLabel')}
                          </label>
                          <input
                            type="number"
                            min={1}
                            max={365}
                            value={recurrenceInterval}
                            onChange={(e) =>
                              setRecurrenceInterval(Math.max(1, parseInt(e.target.value, 10) || 1))
                            }
                            className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-2.5 py-1.5 text-xs text-slate-800 dark:text-slate-200 focus:border-indigo-500 focus:outline-none"
                          />
                        </div>

                        <div>
                          <div className="flex items-center justify-between mb-1">
                            <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400">
                              {t('taskDrawer.repeatEndDateLabel')}
                            </label>
                            {recurrenceEndDate && (
                              <button
                                type="button"
                                onClick={() => setRecurrenceEndDate(null)}
                                className="text-[10px] text-slate-400 hover:text-rose-500 flex items-center gap-1 transition-colors cursor-pointer"
                                title={t('common.clear') || 'Clear'}
                              >
                                <i className="pi pi-times text-[9px]" />
                                <span>{t('common.clear') || 'Hapus'}</span>
                              </button>
                            )}
                          </div>
                          <Calendar
                            value={recurrenceEndDate}
                            onChange={(e) => setRecurrenceEndDate((e.value as Date) || null)}
                            showIcon
                            showButtonBar
                            dateFormat="dd/mm/yy"
                            locale={locale}
                            placeholder="DD/MM/YYYY"
                            className="w-full text-xs"
                            inputClassName="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-2.5 py-1.5 text-xs text-slate-800 dark:text-slate-200 focus:border-indigo-500 focus:outline-none"
                          />
                        </div>
                      </div>
                    </div>
                  )}
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
                    onBlur={handleTitleBlur}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        e.currentTarget.blur()
                      }
                    }}
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
                      data-pr-tooltip={t('taskDrawer.saveProgressTooltip')}
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
                  <div className="space-y-1.5 mb-3 max-h-72 overflow-y-auto pr-1">
                    {task.subtasks?.map((st) => (
                      <SubtaskItemRow
                        key={st.id}
                        subtask={st}
                        taskId={task.id}
                        isRecurring={isRecurring}
                        onToggle={(checked) =>
                          toggleSubtaskMutation.mutate({
                            taskId: task.id,
                            subtaskId: st.id,
                            isCompleted: checked
                          })
                        }
                        onDelete={handleDeleteSubtask}
                        onUpdate={(params) => updateSubtaskMutation.mutateAsync(params)}
                      />
                    ))}
                  </div>

                  {/* Add Subtask Input */}
                  <form onSubmit={handleAddSubtaskSubmit} className="space-y-2">
                    <div className="flex gap-2">
                      <input
                        type="text"
                        value={newSubtaskTitle}
                        onChange={(e) => setNewSubtaskTitle(e.target.value)}
                        placeholder={t('taskDrawer.addSubtaskPlaceholder')}
                        className="flex-1 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 px-3 py-1.5 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none"
                      />
                      <button
                        type="button"
                        onClick={() => setIsAddingWithDetails(!isAddingWithDetails)}
                        data-pr-tooltip={t('taskDrawer.subtaskDetails')}
                        className={`px-2.5 py-1.5 rounded-lg border text-xs transition-colors cursor-pointer ${
                          isAddingWithDetails
                            ? 'bg-indigo-50 dark:bg-indigo-950/70 border-indigo-400 dark:border-indigo-700 text-indigo-600 dark:text-indigo-400'
                            : 'bg-white dark:bg-slate-900 border-slate-200 dark:border-slate-800 text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'
                        }`}
                      >
                        <i className="pi pi-sliders-h text-xs" />
                      </button>
                      <button
                        type="submit"
                        disabled={addSubtaskMutation.isPending || !newSubtaskTitle.trim()}
                        data-pr-tooltip={t('taskDrawer.addSubtaskTooltip')}
                        className="px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white disabled:opacity-50 text-xs font-medium flex items-center gap-1 shadow-xs transition-colors cursor-pointer"
                      >
                        {addSubtaskMutation.isPending && (
                          <i className="pi pi-spin pi-spinner text-[10px]" />
                        )}
                        {t('taskDrawer.add')}
                      </button>
                    </div>

                    {/* Optional additional details when creating */}
                    {isAddingWithDetails && (
                      <div className="p-2.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/80 dark:bg-slate-950/40 space-y-2 animate-fade-in text-xs">
                        <div>
                          <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
                            {t('taskDrawer.subtaskDetails')}
                          </label>
                          <textarea
                            rows={2}
                            value={newSubtaskDescription}
                            onChange={(e) => setNewSubtaskDescription(e.target.value)}
                            placeholder={t('taskDrawer.subtaskDescriptionPlaceholder')}
                            className="w-full rounded-md border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none resize-none"
                          />
                        </div>
                        {!isRecurring && (
                          <div>
                            <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
                              {t('taskDrawer.subtaskDueDate')}
                            </label>
                            <Calendar
                              value={newSubtaskDueDate}
                              onChange={(e) => setNewSubtaskDueDate(e.value as Date | null)}
                              showTime
                              hourFormat="24"
                              placeholder={t('taskDrawer.subtaskDueDatePlaceholder')}
                              className="w-full text-xs"
                              inputClassName="w-full rounded-md border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:border-indigo-500 focus:outline-none"
                              showIcon
                              showButtonBar
                            />
                          </div>
                        )}
                      </div>
                    )}
                  </form>
                </div>

                {/* Task Reminders Section */}
                <div className="pt-3 border-t border-slate-200 dark:border-slate-800/80">
                  <TaskReminderSection
                    taskId={task.id}
                    taskTitle={task.title}
                    dueDate={dueDate ? dueDate.toISOString() : task.due_date}
                  />
                </div>
              </>
            )}
          </div>

          {/* Sticky Footer: Submit Changes */}
          <div className="p-4 border-t border-slate-200 dark:border-slate-800 bg-slate-50/95 dark:bg-slate-950/90 backdrop-blur-sm flex items-center justify-between gap-3 shrink-0">
            <div className="flex items-center gap-2 min-w-0">
              {updateTaskMutation.isPending ? (
                <span className="flex items-center gap-1.5 text-xs font-medium text-indigo-600 dark:text-indigo-400">
                  <i className="pi pi-spin pi-spinner text-xs" />
                  <span className="truncate">{t('taskDrawer.savingChanges')}</span>
                </span>
              ) : isDirty ? (
                <span className="flex items-center gap-1.5 text-xs font-medium text-amber-600 dark:text-amber-400">
                  <span className="w-2 h-2 rounded-full bg-amber-500 animate-pulse shrink-0" />
                  <span className="truncate">{t('taskDrawer.unsavedChanges')}</span>
                </span>
              ) : (
                <span className="text-xs text-emerald-600 dark:text-emerald-400 flex items-center gap-1.5 font-medium">
                  <i className="pi pi-check-circle text-xs text-emerald-500" />
                  <span className="truncate">
                    {t('taskDrawer.saved')} • {t('taskDrawer.taskIdBadge', { id: task.id })}
                  </span>
                </span>
              )}
            </div>

            <div className="flex items-center gap-2 shrink-0">
              {isDirty && (
                <button
                  type="button"
                  onClick={handleResetMain}
                  disabled={updateTaskMutation.isPending}
                  data-pr-tooltip={t('taskDrawer.resetTaskTooltip')}
                  className="px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800 text-xs font-medium text-slate-700 dark:text-slate-300 transition-colors cursor-pointer"
                >
                  {t('taskDrawer.resetChanges')}
                </button>
              )}
              <button
                type="button"
                onClick={() => handleSubmitMain()}
                disabled={updateTaskMutation.isPending || !isDirty}
                data-pr-tooltip={t('taskDrawer.saveTaskTooltip')}
                className="min-w-[145px] justify-center px-4 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed text-xs font-semibold text-white shadow-sm flex items-center gap-1.5 transition-colors cursor-pointer"
              >
                {updateTaskMutation.isPending && (
                  <i className="pi pi-spin pi-spinner text-[10px]" />
                )}
                {updateTaskMutation.isPending
                  ? t('taskDrawer.savingChanges')
                  : t('taskDrawer.saveChanges')}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
