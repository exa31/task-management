import { Calendar } from 'primereact/calendar'
import React, { type FormEvent, useEffect, useRef, useState } from 'react'

import { getLocalizedErrorMessage } from '@/api'
import { Skeleton } from '@/components'
import { useCreateReminder, useTask, useTasks, useUpdateReminder } from '@/hooks'
import { useTranslation } from '@/i18n'
import type { Reminder, Task } from '@/types'

interface ReminderModalProps {
  isOpen: boolean
  onClose: () => void
  initialTaskId?: number | null
  initialReminder?: Reminder | null
}

export const ReminderModal: React.FC<ReminderModalProps> = ({
  isOpen,
  onClose,
  initialTaskId = null,
  initialReminder = null
}) => {
  const { t, locale } = useTranslation()

  const createMutation = useCreateReminder()
  const updateMutation = useUpdateReminder()

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [taskId, setTaskId] = useState<number | null>(initialTaskId)
  const [remindAtDate, setRemindAtDate] = useState<Date | null>(null)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  // Lightweight on-demand task combobox
  const [taskSearchQuery, setTaskSearchQuery] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [isTaskSelectorOpen, setIsTaskSelectorOpen] = useState(false)
  const taskSelectorRef = useRef<HTMLDivElement>(null)

  // Initialize or reset form state
  useEffect(() => {
    if (isOpen) {
      if (initialReminder) {
        setTitle(initialReminder.title)
        setDescription(initialReminder.description || '')
        setTaskId(initialReminder.task_id || null)
        setRemindAtDate(new Date(initialReminder.remind_at))
      } else {
        setTitle('')
        setDescription('')
        setTaskId(initialTaskId)
        // Default 1 hour in the future
        setRemindAtDate(new Date(Date.now() + 60 * 60 * 1000))
      }
      setErrorMsg(null)
      setIsTaskSelectorOpen(false)
      setTaskSearchQuery('')
    }
  }, [isOpen, initialReminder, initialTaskId])

  // Debounce task search (250ms)
  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedSearch(taskSearchQuery.trim())
    }, 250)
    return () => clearTimeout(handler)
  }, [taskSearchQuery])

  // Close dropdown on outside click
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (taskSelectorRef.current && !taskSelectorRef.current.contains(event.target as Node)) {
        setIsTaskSelectorOpen(false)
      }
    }
    if (isTaskSelectorOpen) {
      document.addEventListener('mousedown', handleClickOutside)
    }
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [isTaskSelectorOpen])

  // Fetch only the single selected task by ID
  const { data: selectedTask } = useTask(taskId || undefined)

  // Fetch search results on-demand with limit 15 only when selector is open
  const { data: rawSearchTasks, isLoading: isSearchingTasks } = useTasks(
    {
      search: debouncedSearch || undefined,
      limit: 15
    },
    { enabled: isOpen && isTaskSelectorOpen }
  )
  const searchResults = Array.isArray(rawSearchTasks) ? rawSearchTasks : []

  if (!isOpen) return null

  // Apply quick presets
  const applyPreset = (minutes: number) => {
    setRemindAtDate(new Date(Date.now() + minutes * 60 * 1000))
  }

  const applyTomorrowMorning = () => {
    const tomorrow = new Date()
    tomorrow.setDate(tomorrow.getDate() + 1)
    tomorrow.setHours(9, 0, 0, 0)
    setRemindAtDate(tomorrow)
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!title.trim()) {
      setErrorMsg(`${t('reminders.fieldTitle')} ${t('common.required')}`)
      return
    }

    if (!remindAtDate) {
      setErrorMsg(`${t('reminders.fieldRemindAt')} ${t('common.required')}`)
      return
    }

    const isoDate = remindAtDate.toISOString()

    try {
      if (initialReminder) {
        await updateMutation.mutateAsync({
          id: initialReminder.id,
          payload: {
            title: title.trim(),
            description: description.trim(),
            task_id: taskId,
            remind_at: isoDate
          }
        })
      } else {
        await createMutation.mutateAsync({
          title: title.trim(),
          description: description.trim(),
          task_id: taskId,
          remind_at: isoDate
        })
      }
      onClose()
    } catch (err) {
      setErrorMsg(getLocalizedErrorMessage(err, t))
    }
  }

  const isSubmitting = createMutation.isPending || updateMutation.isPending

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/45 dark:bg-slate-950/75 backdrop-blur-sm animate-fade-in">
      <div className="relative w-full max-w-lg rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-2xl text-slate-900 dark:text-slate-100">
        {/* Header */}
        <div className="flex items-center justify-between pb-3 border-b border-slate-200 dark:border-slate-800 mb-4">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-indigo-50 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400 flex items-center justify-center">
              <i className="pi pi-bell text-sm" />
            </div>
            <div>
              <h3 className="text-sm font-bold text-slate-900 dark:text-slate-100">
                {initialReminder ? t('reminders.editReminder') : t('reminders.newReminder')}
              </h3>
              <p className="text-[11px] text-slate-500 dark:text-slate-400">
                {t('reminders.createPrompt')}
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            title={t('common.close')}
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
              className="text-rose-500/70 hover:text-rose-500 cursor-pointer"
            >
              <i className="pi pi-times text-xs" />
            </button>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Title */}
          <div>
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
              {t('reminders.fieldTitle')} <span className="text-rose-500">*</span>
            </label>
            <input
              type="text"
              required
              autoFocus
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder={t('reminders.titlePlaceholder')}
              className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3 py-2 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none transition-colors"
            />
          </div>

          {/* Linked Task Selector */}
          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300">
                {t('reminders.fieldTask')}
              </label>
              {taskId !== null && (
                <button
                  type="button"
                  onClick={() => {
                    setTaskId(null)
                    setIsTaskSelectorOpen(false)
                    setTaskSearchQuery('')
                  }}
                  className="text-[11px] text-rose-500 hover:text-rose-600 dark:text-rose-400 font-medium cursor-pointer"
                >
                  {t('reminders.unlinkTask')}
                </button>
              )}
            </div>

            {taskId !== null && !isTaskSelectorOpen ? (
              <div className="flex items-center justify-between gap-3 p-2.5 rounded-xl border border-indigo-200/80 dark:border-indigo-900/60 bg-indigo-50/50 dark:bg-indigo-950/30">
                <div className="flex items-center gap-2.5 min-w-0">
                  <div className="w-7 h-7 rounded-lg bg-indigo-100 dark:bg-indigo-900/60 text-indigo-600 dark:text-indigo-400 flex items-center justify-center shrink-0">
                    <i className="pi pi-check-square text-xs" />
                  </div>
                  <div className="min-w-0">
                    <p className="text-xs font-semibold text-slate-900 dark:text-slate-100 truncate">
                      {selectedTask?.title || initialReminder?.task_title || `#${taskId}`}
                    </p>
                    <div className="flex items-center gap-2 mt-0.5 text-[10px] text-slate-500 dark:text-slate-400">
                      {selectedTask?.status && (
                        <span className="px-1.5 py-0.2 rounded font-mono uppercase bg-white/70 dark:bg-slate-900/70 border border-slate-200 dark:border-slate-800 text-slate-600 dark:text-slate-300">
                          {selectedTask.status}
                        </span>
                      )}
                      {selectedTask?.project_name && (
                        <span className="flex items-center gap-1">
                          <i className="pi pi-folder text-[9px]" />
                          {selectedTask.project_name}
                        </span>
                      )}
                      {selectedTask?.due_date && (
                        <span className="flex items-center gap-1 font-mono">
                          <i className="pi pi-calendar text-[9px]" />
                          {new Date(selectedTask.due_date).toLocaleDateString(
                            locale === 'id' ? 'id-ID' : 'en-US'
                          )}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
                <button
                  type="button"
                  onClick={() => setIsTaskSelectorOpen(true)}
                  className="px-2.5 py-1 rounded-lg text-[11px] font-medium border border-indigo-200 dark:border-indigo-800 bg-white dark:bg-slate-900 hover:bg-slate-50 dark:hover:bg-slate-800 text-indigo-600 dark:text-indigo-400 transition-colors shrink-0 cursor-pointer"
                >
                  {t('common.edit') || 'Ganti'}
                </button>
              </div>
            ) : (
              <div className="relative" ref={taskSelectorRef}>
                <div className="relative">
                  <i className="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-xs text-slate-400" />
                  <input
                    type="text"
                    value={taskSearchQuery}
                    onChange={(e) => {
                      setTaskSearchQuery(e.target.value)
                      if (!isTaskSelectorOpen) setIsTaskSelectorOpen(true)
                    }}
                    onFocus={() => setIsTaskSelectorOpen(true)}
                    placeholder={t('reminders.searchTaskPlaceholder')}
                    className="w-full pl-8 pr-8 py-2 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:border-indigo-500 transition-colors"
                  />
                  {taskSearchQuery && (
                    <button
                      type="button"
                      onClick={() => setTaskSearchQuery('')}
                      className="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 cursor-pointer"
                    >
                      <i className="pi pi-times-circle text-xs" />
                    </button>
                  )}
                </div>

                {isTaskSelectorOpen && (
                  <div className="absolute left-0 right-0 top-full mt-1 max-h-56 overflow-y-auto rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xl z-30 p-1 divide-y divide-slate-100 dark:divide-slate-800/60">
                    <button
                      type="button"
                      onClick={() => {
                        setTaskId(null)
                        setIsTaskSelectorOpen(false)
                        setTaskSearchQuery('')
                      }}
                      className="w-full px-3 py-2 text-left rounded-lg text-xs flex items-center justify-between hover:bg-slate-50 dark:hover:bg-slate-800/60 transition-colors cursor-pointer text-slate-500 dark:text-slate-400"
                    >
                      <span className="flex items-center gap-2">
                        <i className="pi pi-times text-[10px]" />
                        <span>{t('reminders.noTask')}</span>
                      </span>
                      {taskId === null && <i className="pi pi-check text-xs text-indigo-500" />}
                    </button>

                    {isSearchingTasks ? (
                      <div className="p-2 space-y-1.5">
                        <Skeleton className="h-8 w-full rounded-lg" />
                        <Skeleton className="h-8 w-full rounded-lg" />
                      </div>
                    ) : searchResults.length === 0 ? (
                      <div className="py-4 text-center text-xs text-slate-400">
                        {t('reminders.noTasksFound')}
                      </div>
                    ) : (
                      searchResults.map((tk: Task) => (
                        <button
                          key={tk.id}
                          type="button"
                          onClick={() => {
                            setTaskId(tk.id)
                            setIsTaskSelectorOpen(false)
                            setTaskSearchQuery('')
                          }}
                          className={`w-full px-3 py-2 text-left rounded-lg text-xs flex items-center justify-between hover:bg-slate-50 dark:hover:bg-slate-800/60 transition-colors cursor-pointer ${
                            taskId === tk.id
                              ? 'bg-indigo-50/70 dark:bg-indigo-950/40 text-indigo-600 dark:text-indigo-400 font-semibold'
                              : 'text-slate-800 dark:text-slate-200'
                          }`}
                        >
                          <div className="min-w-0 flex-1 pr-2">
                            <div className="flex items-center gap-2">
                              <span className="truncate">{tk.title}</span>
                              <span className="shrink-0 text-[9px] px-1.5 py-0.2 rounded font-mono font-medium bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400">
                                {tk.status}
                              </span>
                            </div>
                            {(tk.project_name || tk.due_date) && (
                              <div className="flex items-center gap-2 mt-0.5 text-[10px] text-slate-400">
                                {tk.project_name && <span>{tk.project_name}</span>}
                                {tk.due_date && (
                                  <span>
                                    {new Date(tk.due_date).toLocaleDateString(
                                      locale === 'id' ? 'id-ID' : 'en-US'
                                    )}
                                  </span>
                                )}
                              </div>
                            )}
                          </div>
                          {taskId === tk.id && (
                            <i className="pi pi-check text-xs text-indigo-500 shrink-0" />
                          )}
                        </button>
                      ))
                    )}
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Date & Time Picker + Quick Presets */}
          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300">
                {t('reminders.fieldRemindAt')} <span className="text-rose-500">*</span>
              </label>
              <span className="text-[10px] text-slate-400 font-medium">
                {t('reminders.quickPreset')}
              </span>
            </div>

            {/* Presets Button Row */}
            <div className="flex flex-wrap items-center gap-1.5 mb-2">
              <button
                type="button"
                onClick={() => applyPreset(15)}
                className="px-2.5 py-1 rounded-md text-[11px] font-medium border border-slate-200 dark:border-slate-800 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 transition-colors cursor-pointer"
              >
                {t('reminders.quickAdd15m')}
              </button>
              <button
                type="button"
                onClick={() => applyPreset(60)}
                className="px-2.5 py-1 rounded-md text-[11px] font-medium border border-slate-200 dark:border-slate-800 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 transition-colors cursor-pointer"
              >
                {t('reminders.quickAdd1h')}
              </button>
              <button
                type="button"
                onClick={() => applyPreset(180)}
                className="px-2.5 py-1 rounded-md text-[11px] font-medium border border-slate-200 dark:border-slate-800 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 transition-colors cursor-pointer"
              >
                {t('reminders.quickAdd3h')}
              </button>
              <button
                type="button"
                onClick={applyTomorrowMorning}
                className="px-2.5 py-1 rounded-md text-[11px] font-medium border border-slate-200 dark:border-slate-800 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 transition-colors cursor-pointer"
              >
                {t('reminders.snoozeTomorrow')}
              </button>
            </div>

            <Calendar
              value={remindAtDate}
              onChange={(e) => setRemindAtDate(e.value as Date)}
              showTime
              hourFormat="24"
              showIcon
              dateFormat="dd/mm/yy"
              locale={locale}
              placeholder={t('reminders.dateTimePlaceholder')}
              className="w-full text-xs"
              inputClassName="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:outline-none transition-colors"
              touchUI={false}
              showButtonBar
              required
            />
          </div>

          {/* Description */}
          <div>
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
              {t('reminders.fieldDescription')}
            </label>
            <textarea
              rows={2}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={t('reminders.descriptionPlaceholder')}
              className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3 py-2 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none resize-none transition-colors"
            />
          </div>

          {/* Footer Buttons */}
          <div className="flex items-center justify-end gap-2 pt-3 border-t border-slate-200 dark:border-slate-800">
            <button
              type="button"
              onClick={onClose}
              disabled={isSubmitting}
              className="px-3 py-1.5 text-xs text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
            >
              {t('common.cancel')}
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className="px-4 py-1.5 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-lg disabled:opacity-50 flex items-center gap-1.5 shadow-xs transition-colors cursor-pointer"
            >
              {isSubmitting && <i className="pi pi-spin pi-spinner text-xs" />}
              {initialReminder ? t('common.save') : t('reminders.newReminder')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
