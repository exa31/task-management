import { useEffect, useState } from 'react'

import { confirmDialog } from '@/components'
import {
  useCreateReminder,
  useDeleteReminder,
  useDismissReminder,
  useReminders,
  useSnoozeReminder
} from '@/hooks'
import { useTranslation } from '@/i18n'
import type { Reminder } from '@/types'

import { ReminderModal } from './ReminderModal'

interface TaskReminderSectionProps {
  taskId: number
  taskTitle: string
  dueDate?: string | null
}

export const TaskReminderSection = ({ taskId, taskTitle, dueDate }: TaskReminderSectionProps) => {
  const { t, locale } = useTranslation()
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingReminder, setEditingReminder] = useState<Reminder | null>(null)

  const { data: reminderData } = useReminders({ task_id: taskId, status: 'all' })
  const reminders = reminderData?.items || []

  const createMutation = useCreateReminder()
  const deleteMutation = useDeleteReminder()
  const snoozeMutation = useSnoozeReminder()
  const dismissMutation = useDismissReminder()

  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 30000)
    return () => clearInterval(timer)
  }, [])

  // Preset handlers relative to task deadline
  const handleCreatePreset = (minutesBefore: number, label: string) => {
    let remindDate: Date
    if (dueDate) {
      remindDate = new Date(new Date(dueDate).getTime() - minutesBefore * 60 * 1000)
    } else {
      remindDate = new Date(Date.now() + 60 * 60 * 1000) // 1h from now fallback
    }

    // Don't create past reminder
    if (remindDate.getTime() < Date.now()) {
      remindDate = new Date(Date.now() + 5 * 60 * 1000) // 5 mins ahead
    }

    createMutation.mutate({
      task_id: taskId,
      title: `${label}: ${taskTitle}`,
      remind_at: remindDate.toISOString()
    })
  }

  const formatTime = (dateStr: string) => {
    const d = new Date(dateStr)
    return d.toLocaleString(locale === 'id' ? 'id-ID' : 'en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: locale !== 'id'
    })
  }

  return (
    <div className="space-y-3">
      {/* Section Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <i className="pi pi-bell text-xs text-indigo-500" />
          <h4 className="text-xs font-semibold text-slate-800 dark:text-slate-200">
            {t('reminders.title')}
          </h4>
          {reminders.length > 0 && (
            <span className="px-1.5 py-0.2 rounded-full text-[10px] font-bold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
              {reminders.length}
            </span>
          )}
        </div>

        <button
          type="button"
          onClick={() => {
            setEditingReminder(null)
            setIsModalOpen(true)
          }}
          className="text-[11px] font-semibold text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 flex items-center gap-1 cursor-pointer"
        >
          <i className="pi pi-plus text-[10px]" />
          {t('reminders.newReminder')}
        </button>
      </div>

      {/* Relative Presets (If task has due date) */}
      {dueDate && (
        <div className="flex flex-wrap items-center gap-1.5 p-2 rounded-lg bg-slate-50 dark:bg-slate-900/60 border border-slate-100 dark:border-slate-800">
          <span className="text-[10px] text-slate-400 font-medium mr-1">
            {t('reminders.quickPreset')}:
          </span>
          <button
            type="button"
            onClick={() => handleCreatePreset(15, t('reminders.preset15mBefore'))}
            disabled={createMutation.isPending}
            className="px-2 py-0.5 rounded-md text-[10px] font-semibold border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300 transition-colors cursor-pointer"
          >
            {t('reminders.preset15m')}
          </button>
          <button
            type="button"
            onClick={() => handleCreatePreset(60, t('reminders.preset1hBefore'))}
            disabled={createMutation.isPending}
            className="px-2 py-0.5 rounded-md text-[10px] font-semibold border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300 transition-colors cursor-pointer"
          >
            {t('reminders.preset1h')}
          </button>
          <button
            type="button"
            onClick={() => handleCreatePreset(1440, t('reminders.preset1dBefore'))}
            disabled={createMutation.isPending}
            className="px-2 py-0.5 rounded-md text-[10px] font-semibold border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300 transition-colors cursor-pointer"
          >
            {t('reminders.preset1d')}
          </button>
        </div>
      )}

      {/* Reminders List */}
      <div className="space-y-2">
        {reminders.length === 0 ? (
          <p className="text-[11px] text-slate-400 italic py-1">{t('reminders.emptyDesc')}</p>
        ) : (
          reminders.map((rem: Reminder) => {
            const isPending = rem.status === 'pending'
            const isSent = rem.status === 'sent'
            const isOverdue = isPending && new Date(rem.remind_at).getTime() < now

            return (
              <div
                key={rem.id}
                className="flex items-center justify-between p-2.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 text-xs hover:border-slate-300 dark:hover:border-slate-700 transition-colors"
              >
                <div className="flex items-center gap-2 min-w-0">
                  <span
                    className={`w-2 h-2 rounded-full shrink-0 ${
                      isOverdue
                        ? 'bg-rose-500 animate-pulse'
                        : isPending
                          ? 'bg-amber-500'
                          : isSent
                            ? 'bg-indigo-500'
                            : 'bg-slate-400'
                    }`}
                  />
                  <div className="min-w-0">
                    <p className="font-medium text-slate-800 dark:text-slate-200 truncate leading-snug">
                      {rem.title}
                    </p>
                    <p className="text-[10px] text-slate-400 font-mono mt-0.5">
                      {formatTime(rem.remind_at)}
                      {rem.snooze_count > 0 &&
                        ` • ${t('reminders.snoozedTimes', { count: rem.snooze_count })}`}
                    </p>
                  </div>
                </div>

                {/* Quick Actions */}
                <div className="flex items-center gap-1 shrink-0 ml-2">
                  {isPending && (
                    <>
                      <button
                        type="button"
                        onClick={() =>
                          snoozeMutation.mutate({ id: rem.id, payload: { minutes: 15 } })
                        }
                        disabled={snoozeMutation.isPending}
                        className="px-1.5 py-0.5 rounded text-[10px] font-semibold border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300 transition-colors cursor-pointer"
                        title={t('reminders.snooze15m')}
                      >
                        {t('reminders.quickAdd15m')}
                      </button>
                      <button
                        type="button"
                        onClick={() => dismissMutation.mutate(rem.id)}
                        disabled={dismissMutation.isPending}
                        title={t('reminders.dismiss')}
                        className="w-6 h-6 rounded flex items-center justify-center text-emerald-600 dark:text-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-950/40 transition-colors cursor-pointer"
                      >
                        <i className="pi pi-check text-[11px]" />
                      </button>
                    </>
                  )}

                  <button
                    type="button"
                    onClick={() => {
                      setEditingReminder(rem)
                      setIsModalOpen(true)
                    }}
                    title={t('common.edit')}
                    className="w-6 h-6 rounded flex items-center justify-center text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
                  >
                    <i className="pi pi-pencil text-[11px]" />
                  </button>

                  <button
                    type="button"
                    onClick={() => {
                      confirmDialog({
                        message: t('reminders.deleteConfirm'),
                        header: t('common.confirm'),
                        icon: 'pi pi-exclamation-triangle',
                        acceptClassName: 'p-button-danger',
                        acceptLabel: t('common.yes'),
                        rejectLabel: t('common.no'),
                        accept: () => deleteMutation.mutate(rem.id)
                      })
                    }}
                    disabled={deleteMutation.isPending}
                    title={t('common.delete')}
                    className="w-6 h-6 rounded flex items-center justify-center text-rose-500 hover:bg-rose-50 dark:hover:bg-rose-950/40 transition-colors cursor-pointer"
                  >
                    <i className="pi pi-trash text-[11px]" />
                  </button>
                </div>
              </div>
            )
          })
        )}
      </div>

      {/* Modal */}
      <ReminderModal
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false)
          setEditingReminder(null)
        }}
        initialTaskId={taskId}
        initialReminder={editingReminder}
      />
    </div>
  )
}
