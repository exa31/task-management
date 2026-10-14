import { Calendar } from 'primereact/calendar'
import React, { useEffect, useState } from 'react'

import { getLocalizedErrorMessage } from '@/api'
import { ProjectModal } from '@/features/projects'
import { useCreateTask, useProjects } from '@/hooks'
import { useTranslation } from '@/i18n'
import { useUIStore } from '@/stores'
import type { RecurrencePattern, TaskPriority, TaskStatus } from '@/types'

interface LocalSubtask {
  id: string
  title: string
}

export const TaskModal: React.FC = () => {
  const { isCreateTaskModalOpen, closeCreateTaskModal, createTaskInitialProjectId } = useUIStore()
  const createTaskMutation = useCreateTask()
  const { data: projects = [] } = useProjects()
  const { t, locale } = useTranslation()

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [status, setStatus] = useState<TaskStatus>('todo')
  const [priority, setPriority] = useState<TaskPriority>('medium')
  const [dueDate, setDueDate] = useState<Date | null>(null)
  const [projectId, setProjectId] = useState<number | null>(null)
  const [isCreateProjectOpen, setIsCreateProjectOpen] = useState(false)
  const [isFocus, setIsFocus] = useState(false)
  const [tagsInput, setTagsInput] = useState('')
  const [subtasks, setSubtasks] = useState<LocalSubtask[]>([])
  const [newSubtaskTitle, setNewSubtaskTitle] = useState('')
  const [isRecurring, setIsRecurring] = useState(false)
  const [recurrencePattern, setRecurrencePattern] = useState<RecurrencePattern>('daily')
  const [recurrenceInterval, setRecurrenceInterval] = useState<number>(1)
  const [recurrenceEndDate, setRecurrenceEndDate] = useState<Date | null>(null)
  const [touched, setTouched] = useState(false)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  useEffect(() => {
    if (isCreateTaskModalOpen) {
      setProjectId(createTaskInitialProjectId)
    }
  }, [isCreateTaskModalOpen, createTaskInitialProjectId])

  if (!isCreateTaskModalOpen) return null

  const isTitleInvalid = touched && !title.trim()

  const handleClose = () => {
    setTitle('')
    setDescription('')
    setStatus('todo')
    setPriority('medium')
    setDueDate(null)
    setProjectId(null)
    setIsFocus(false)
    setTagsInput('')
    setSubtasks([])
    setNewSubtaskTitle('')
    setIsRecurring(false)
    setRecurrencePattern('daily')
    setRecurrenceInterval(1)
    setRecurrenceEndDate(null)
    setTouched(false)
    setErrorMsg(null)
    closeCreateTaskModal()
  }

  const handleAddSubtask = () => {
    if (!newSubtaskTitle.trim()) return
    setSubtasks((prev) => [
      ...prev,
      { id: `st_${Date.now()}_${prev.length}`, title: newSubtaskTitle.trim() }
    ])
    setNewSubtaskTitle('')
  }

  const handleRemoveSubtask = (idToRemove: string) => {
    setSubtasks((prev) => prev.filter((item) => item.id !== idToRemove))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setTouched(true)
    setErrorMsg(null)
    if (!title.trim()) return

    const tags = tagsInput
      .split(',')
      .map((t) => t.trim().toLowerCase())
      .filter(Boolean)

    const finalSubtasks = subtasks.map((item) => item.title)
    if (newSubtaskTitle.trim()) {
      finalSubtasks.push(newSubtaskTitle.trim())
    }

    try {
      await createTaskMutation.mutateAsync({
        title: title.trim(),
        description: description.trim(),
        status,
        priority,
        project_id: projectId,
        due_date: dueDate ? dueDate.toISOString() : undefined,
        is_focus: isFocus,
        tags,
        subtasks: finalSubtasks,
        is_recurring: isRecurring,
        recurrence_pattern: isRecurring ? recurrencePattern : undefined,
        recurrence_interval: isRecurring ? recurrenceInterval : undefined,
        recurrence_end_date: isRecurring && recurrenceEndDate ? recurrenceEndDate.toISOString() : undefined
      })

      handleClose()
    } catch (err) {
      setErrorMsg(getLocalizedErrorMessage(err, t))
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/45 dark:bg-slate-950/75 backdrop-blur-sm animate-fade-in">
      <div className="relative w-full max-w-lg rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-2xl text-slate-900 dark:text-slate-100">
        <div className="flex items-center justify-between pb-3 border-b border-slate-200 dark:border-slate-800 mb-4">
          <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <i className="pi pi-plus-circle text-indigo-500 dark:text-indigo-400" />
            {t('taskModal.createTitle')}
          </h2>
          <button
            type="button"
            onClick={handleClose}
            data-pr-tooltip={t('common.close')}
            className="p-1 rounded text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
          >
            <i className="pi pi-times text-xs" />
          </button>
        </div>

        {errorMsg && (
          <div className="mb-4 flex items-start gap-2 rounded-lg border border-rose-500/20 bg-rose-500/10 p-3 text-xs text-rose-600 dark:text-rose-400">
            <i className="pi pi-exclamation-circle text-sm mt-0.5 shrink-0" />
            <span className="flex-1 font-medium">{errorMsg}</span>
            <button
              type="button"
              onClick={() => setErrorMsg(null)}
              data-pr-tooltip={t('common.close')}
              className="text-rose-500/70 hover:text-rose-500 cursor-pointer"
            >
              <i className="pi pi-times text-xs" />
            </button>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Title */}
          <div>
            <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
              {t('taskModal.titleLabel')} <span className="text-rose-500 font-bold ml-0.5">*</span>
            </label>
            <input
              type="text"
              required
              autoFocus
              value={title}
              onBlur={() => setTouched(true)}
              onChange={(e) => {
                setTitle(e.target.value)
                if (touched) setTouched(true)
              }}
              placeholder={t('taskModal.titlePlaceholder')}
              className={`w-full rounded-lg border px-3 py-2 text-xs text-slate-900 dark:text-slate-100 bg-slate-50 dark:bg-slate-950 placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none transition-colors ${
                isTitleInvalid
                  ? 'border-rose-500 focus:border-rose-500 ring-1 ring-rose-500/30'
                  : 'border-slate-200 dark:border-slate-800 focus:border-indigo-500'
              }`}
            />
            {isTitleInvalid && (
              <p className="text-[11px] text-rose-500 mt-1 flex items-center gap-1 font-medium">
                <i className="pi pi-exclamation-circle text-[10px]" />{' '}
                {t('taskModal.titleRequired')}
              </p>
            )}
          </div>

          {/* Description */}
          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
              {t('taskModal.descriptionLabel')}
            </label>
            <textarea
              rows={3}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={t('taskModal.descriptionPlaceholder')}
              className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3 py-2 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none resize-none transition-colors"
            />
          </div>

          {/* Status & Priority */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                {t('taskModal.statusLabel')}{' '}
                <span className="text-rose-500 font-bold ml-0.5">*</span>
              </label>
              <select
                value={status}
                onChange={(e) => setStatus(e.target.value as TaskStatus)}
                className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:outline-none transition-colors"
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
              </select>
            </div>

            <div>
              <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                {t('taskModal.priorityLabel')}{' '}
                <span className="text-rose-500 font-bold ml-0.5">*</span>
              </label>
              <select
                value={priority}
                onChange={(e) => setPriority(e.target.value as TaskPriority)}
                className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:outline-none transition-colors"
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
          </div>

          {/* Due Date & Project */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                {t('taskModal.dueDateLabel')}
              </label>
              <Calendar
                value={dueDate}
                onChange={(e) => setDueDate(e.value as Date | null)}
                showTime
                hourFormat="24"
                dateFormat="dd/mm/yy"
                locale={locale}
                placeholder={t('taskModal.dueDateLabel')}
                className="w-full text-xs"
                inputClassName="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:outline-none transition-colors"
                showIcon
                touchUI={false}
                showButtonBar
              />
            </div>

            <div>
              <div className="flex items-center justify-between mb-1">
                <label className="block text-xs font-medium text-slate-700 dark:text-slate-300">
                  {t('taskModal.projectLabel')}
                </label>
                <button
                  type="button"
                  onClick={() => setIsCreateProjectOpen(true)}
                  data-pr-tooltip={t('taskModal.createNewProject')}
                  className="text-[11px] font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-500 hover:underline flex items-center gap-1 cursor-pointer"
                >
                  <i className="pi pi-plus text-[9px]" />
                  {t('taskModal.createNewProject')}
                </button>
              </div>
              <select
                value={projectId === null ? '' : String(projectId)}
                onChange={(e) => setProjectId(e.target.value ? parseInt(e.target.value, 10) : null)}
                className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:outline-none transition-colors"
              >
                <option
                  value=""
                  className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                >
                  {t('taskModal.noProjectOption')}
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

          {/* Repeat / Recurrence Task Section */}
          <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-950/40 p-3 transition-all duration-200">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2.5">
                <div
                  className={`w-7 h-7 rounded-lg flex items-center justify-center transition-colors ${
                    isRecurring
                      ? 'bg-indigo-600 text-white shadow-xs'
                      : 'bg-slate-200 dark:bg-slate-800 text-slate-500 dark:text-slate-400'
                  }`}
                >
                  <i className="pi pi-sync text-xs" />
                </div>
                <div>
                  <span className="text-xs font-semibold text-slate-900 dark:text-slate-100 block">
                    {t('taskModal.repeatTaskLabel')}
                  </span>
                  <span className="text-[10px] text-slate-500 dark:text-slate-400 block">
                    {t('taskModal.repeatTaskHelp')}
                  </span>
                </div>
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
                {/* Pattern selector */}
                <div>
                  <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1.5">
                    {t('taskModal.repeatPatternLabel')}
                  </label>
                  <div className="grid grid-cols-4 gap-1.5">
                    {(['daily', 'weekdays', 'weekly', 'monthly'] as const).map((pat) => (
                      <button
                        key={pat}
                        type="button"
                        onClick={() => setRecurrencePattern(pat)}
                        className={`px-2 py-1.5 rounded-lg text-xs font-medium transition-all text-center cursor-pointer ${
                          recurrencePattern === pat
                            ? 'bg-indigo-600 text-white shadow-xs'
                            : 'bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-800 hover:bg-slate-100 dark:hover:bg-slate-800'
                        }`}
                      >
                        {t(`taskModal.patterns.${pat}`)}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Interval & End Date */}
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
                      {t('taskModal.repeatIntervalLabel')}
                    </label>
                    <input
                      type="number"
                      min={1}
                      max={365}
                      value={recurrenceInterval}
                      onChange={(e) =>
                        setRecurrenceInterval(Math.max(1, parseInt(e.target.value, 10) || 1))
                      }
                      className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-[11px] font-medium text-slate-600 dark:text-slate-400 mb-1">
                      {t('taskModal.repeatEndDateLabel')}
                    </label>
                    <Calendar
                      value={recurrenceEndDate}
                      onChange={(e) => setRecurrenceEndDate((e.value as Date) || null)}
                      showIcon
                      showButtonBar
                      dateFormat="dd/mm/yy"
                      locale={locale}
                      placeholder="DD/MM/YYYY"
                      className="w-full text-xs"
                      inputClassName="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:outline-none"
                    />
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Checklist Subtasks */}
          <div>
            <div className="flex items-center justify-between mb-1.5">
              <label className="block text-xs font-medium text-slate-600 dark:text-slate-400">
                {t('tasks.checklist')}
                {subtasks.length > 0 && (
                  <span className="ml-1.5 px-1.5 py-0.5 rounded-full text-[10px] font-semibold bg-indigo-50 dark:bg-indigo-950/50 text-indigo-600 dark:text-indigo-400 border border-indigo-200 dark:border-indigo-800/60">
                    {subtasks.length}
                  </span>
                )}
              </label>
            </div>

            {/* Subtask Items */}
            {subtasks.length > 0 && (
              <div className="space-y-1.5 mb-2 max-h-36 overflow-y-auto pr-0.5">
                {subtasks.map((st) => (
                  <div
                    key={st.id}
                    className="group flex items-center justify-between p-2 rounded-lg bg-slate-50 dark:bg-slate-950/60 hover:bg-slate-100/80 dark:hover:bg-slate-950 border border-slate-200 dark:border-slate-800 text-xs transition-colors"
                  >
                    <div className="flex items-center gap-2 flex-1 min-w-0">
                      <i className="pi pi-check-square text-xs text-indigo-500 dark:text-indigo-400 shrink-0" />
                      <span className="truncate text-slate-800 dark:text-slate-200">
                        {st.title}
                      </span>
                    </div>

                    <button
                      type="button"
                      onClick={() => handleRemoveSubtask(st.id)}
                      data-pr-tooltip={t('taskDrawer.deleteSubtaskTooltip')}
                      className="p-1 text-slate-400 hover:text-rose-500 dark:hover:text-rose-400 opacity-70 group-hover:opacity-100 transition-opacity cursor-pointer"
                    >
                      <i className="pi pi-times text-[10px]" />
                    </button>
                  </div>
                ))}
              </div>
            )}

            {/* Add Subtask Input */}
            <div className="flex gap-2">
              <input
                type="text"
                value={newSubtaskTitle}
                onChange={(e) => setNewSubtaskTitle(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault()
                    handleAddSubtask()
                  }
                }}
                placeholder={t('taskDrawer.addSubtaskPlaceholder')}
                className="flex-1 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3 py-1.5 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none transition-colors"
              />
              <button
                type="button"
                onClick={handleAddSubtask}
                disabled={!newSubtaskTitle.trim()}
                data-pr-tooltip={t('taskDrawer.addSubtaskTooltip')}
                className="px-3 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 border border-slate-200 dark:border-slate-700 disabled:opacity-50 text-xs font-medium text-slate-700 dark:text-slate-200 flex items-center gap-1 transition-colors cursor-pointer"
              >
                <i className="pi pi-plus text-[10px]" />
                {t('taskDrawer.add')}
              </button>
            </div>
          </div>

          {/* Tags & Today's Focus */}
          <div className="flex items-center justify-between pt-1">
            <input
              type="text"
              value={tagsInput}
              onChange={(e) => setTagsInput(e.target.value)}
              placeholder={t('taskModal.tagsPlaceholder')}
              className="w-60 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-200 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none transition-colors"
            />

            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={isFocus}
                onChange={(e) => setIsFocus(e.target.checked)}
                className="rounded border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-950 text-indigo-500 focus:ring-0 cursor-pointer"
              />
              <span className="text-xs text-amber-500 dark:text-amber-400 flex items-center gap-1 font-medium">
                <i className="pi pi-star-fill text-[10px]" /> {t('dashboard.todaysFocus')}
              </span>
            </label>
          </div>

          {/* Footer Buttons */}
          <div className="flex items-center justify-end gap-2 pt-3 border-t border-slate-200 dark:border-slate-800">
            <button
              type="button"
              onClick={handleClose}
              data-pr-tooltip={t('common.cancel')}
              className="px-3 py-1.5 text-xs text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
            >
              {t('common.cancel')}
            </button>
            <button
              type="submit"
              disabled={createTaskMutation.isPending || !title.trim()}
              data-pr-tooltip={t('taskModal.submit')}
              className="px-4 py-1.5 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-lg disabled:opacity-50 flex items-center gap-1.5 shadow-xs transition-colors cursor-pointer"
            >
              {createTaskMutation.isPending && <i className="pi pi-spin pi-spinner text-xs" />}
              {t('taskModal.submit')}
            </button>
          </div>
        </form>
      </div>

      {/* Inline Project Modal */}
      <ProjectModal
        isOpen={isCreateProjectOpen}
        onClose={() => setIsCreateProjectOpen(false)}
        onSuccess={(created) => setProjectId(created.id)}
      />
    </div>
  )
}
