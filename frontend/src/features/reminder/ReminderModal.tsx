import { Calendar } from 'primereact/calendar'
import React, { type FormEvent, useEffect, useState } from 'react'

import { getLocalizedErrorMessage } from '@/api'
import { useCreateReminder, useTasks, useUpdateReminder } from '@/hooks'
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
  const { data: rawTasks } = useTasks()
  const tasks = Array.isArray(rawTasks) ? rawTasks : []

  const createMutation = useCreateReminder()
  const updateMutation = useUpdateReminder()

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [taskId, setTaskId] = useState<number | null>(initialTaskId)
  const [remindAtDate, setRemindAtDate] = useState<Date | null>(null)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  // Initialize or reset form state
  useEffect(() => {
    if (initialReminder) {
      if (isOpen) {
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
    }
  }, [isOpen, initialReminder, initialTaskId])

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
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
              {t('reminders.fieldTask')}
            </label>
            <select
              value={taskId === null ? '' : String(taskId)}
              onChange={(e) => setTaskId(e.target.value ? parseInt(e.target.value, 10) : null)}
              className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:outline-none transition-colors"
            >
              <option
                value=""
                className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
              >
                {t('reminders.noTask')}
              </option>
              {tasks.map((tk: Task) => (
                <option
                  key={tk.id}
                  value={String(tk.id)}
                  className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                >
                  {tk.title}{' '}
                  {tk.due_date
                    ? ` (${t('reminders.taskDueBadge', {
                        date: new Date(tk.due_date).toLocaleDateString(
                          locale === 'id' ? 'id-ID' : 'en-US'
                        )
                      })})`
                    : ''}
                </option>
              ))}
            </select>
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
