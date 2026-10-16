import { useEffect, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { taskApi } from '@/api'
import { useDismissReminder, useReminders, useSnoozeReminder } from '@/hooks'
import { useTranslation } from '@/i18n'
import { useTaskStore } from '@/stores'
import type { Reminder } from '@/types'

export const NotificationBell = () => {
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)
  const navigate = useNavigate()
  const { t } = useTranslation()
  const { openDrawer } = useTaskStore()

  // Fetch pending reminders
  const { data: reminderData } = useReminders({ status: 'pending' })
  const reminders = reminderData?.items || []

  const dismissMutation = useDismissReminder()
  const snoozeMutation = useSnoozeReminder()

  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 30000)
    return () => clearInterval(timer)
  }, [])

  // Close dropdown on outside click
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }
    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside)
    }
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [isOpen])

  const handleOpenTask = async (reminder: Reminder) => {
    setIsOpen(false)
    if (!reminder.task_id) return
    try {
      const res = await taskApi.getById(reminder.task_id)
      if (res?.data) {
        openDrawer(res.data)
      }
    } catch {
      // Fallback to tasks list
    }
    navigate('/tasks')
  }

  const formatRelativeTime = (dateStr: string) => {
    const target = new Date(dateStr).getTime()
    const diffMs = target - now
    const diffMins = Math.round(diffMs / 60000)

    if (diffMins < 0) {
      const overdueMins = Math.abs(diffMins)
      if (overdueMins < 60) return t('reminders.minutesAgo', { count: overdueMins })
      const hours = Math.round(overdueMins / 60)
      if (hours < 24) return t('reminders.hoursAgo', { count: hours })
      return t('reminders.daysAgo', { count: Math.round(hours / 24) })
    }

    if (diffMins === 0) return t('reminders.justNow')
    if (diffMins < 60) return t('reminders.inMinutes', { count: diffMins })
    const hours = Math.round(diffMins / 60)
    if (hours < 24) return t('reminders.inHours', { count: hours })
    return t('reminders.inDays', { count: Math.round(hours / 24) })
  }

  const count = reminders.length

  return (
    <div className="relative" ref={dropdownRef}>
      {/* Bell Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="relative p-2 rounded-lg text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-900 transition-colors cursor-pointer"
        title={t('reminders.notificationsTitle')}
        aria-label={t('reminders.notificationsTitle')}
      >
        <i className="pi pi-bell text-sm" />
        {count > 0 && (
          <span className="absolute top-1 right-1 flex h-4 min-w-[16px] px-1 items-center justify-center rounded-full bg-rose-500 text-[10px] font-bold text-white shadow-xs">
            {count > 99 ? '99+' : count}
          </span>
        )}
      </button>

      {/* Popover Dropdown */}
      {isOpen && (
        <div className="absolute right-0 mt-2 w-80 sm:w-96 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-xl z-50 overflow-hidden animate-in fade-in slide-in-from-top-2 duration-150">
          {/* Header */}
          <div className="flex items-center justify-between px-4 py-3 border-b border-slate-100 dark:border-slate-800/80 bg-slate-50/50 dark:bg-slate-950/50">
            <div className="flex items-center gap-2">
              <i className="pi pi-bell text-indigo-500 text-xs" />
              <span className="font-semibold text-xs text-slate-900 dark:text-slate-100">
                {t('reminders.notificationsTitle')}
              </span>
              {count > 0 && (
                <span className="px-1.5 py-0.2 rounded-full text-[10px] font-bold bg-indigo-100 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400">
                  {count}
                </span>
              )}
            </div>
            <Link
              to="/reminders"
              onClick={() => setIsOpen(false)}
              className="text-[11px] font-medium text-indigo-600 dark:text-indigo-400 hover:underline"
            >
              {t('reminders.viewAllReminders')}
            </Link>
          </div>

          {/* List */}
          <div className="max-h-[360px] overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800/50">
            {reminders.length === 0 ? (
              <div className="p-8 text-center">
                <div className="w-10 h-10 mx-auto rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center text-slate-400 mb-2">
                  <i className="pi pi-check text-base text-emerald-500" />
                </div>
                <p className="text-xs font-medium text-slate-700 dark:text-slate-300">
                  {t('reminders.noNotifications')}
                </p>
                <p className="text-[11px] text-slate-400 mt-0.5">{t('reminders.emptyDesc')}</p>
              </div>
            ) : (
              reminders.map((rem: Reminder) => {
                const isOverdue = new Date(rem.remind_at).getTime() < now
                return (
                  <div
                    key={rem.id}
                    className="p-3.5 hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors flex flex-col gap-2 group"
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-1.5 mb-1">
                          <span
                            className={`w-2 h-2 rounded-full shrink-0 ${
                              isOverdue ? 'bg-rose-500 animate-pulse' : 'bg-amber-500'
                            }`}
                          />
                          <h4 className="text-xs font-semibold text-slate-800 dark:text-slate-200 truncate">
                            {rem.title}
                          </h4>
                        </div>
                        {rem.description && (
                          <p className="text-[11px] text-slate-500 dark:text-slate-400 line-clamp-2">
                            {rem.description}
                          </p>
                        )}
                      </div>
                      <span
                        className={`text-[10px] font-mono shrink-0 px-1.5 py-0.5 rounded ${
                          isOverdue
                            ? 'bg-rose-50 dark:bg-rose-950/40 text-rose-600 dark:text-rose-400 font-semibold'
                            : 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400'
                        }`}
                      >
                        {formatRelativeTime(rem.remind_at)}
                      </span>
                    </div>

                    {/* Linked task tag */}
                    {rem.task_title && (
                      <button
                        onClick={() => handleOpenTask(rem)}
                        className="self-start text-[10px] px-2 py-0.5 rounded bg-indigo-50 dark:bg-indigo-950/40 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-100 dark:hover:bg-indigo-900/50 flex items-center gap-1 cursor-pointer transition-colors"
                      >
                        <i className="pi pi-check-square text-[9px]" />
                        <span className="truncate max-w-[200px]">{rem.task_title}</span>
                      </button>
                    )}

                    {/* Action buttons */}
                    <div className="flex items-center justify-between pt-1 border-t border-slate-100 dark:border-slate-800/40">
                      <div className="flex items-center gap-1">
                        <button
                          onClick={() =>
                            snoozeMutation.mutate({ id: rem.id, payload: { minutes: 15 } })
                          }
                          disabled={snoozeMutation.isPending}
                          className="px-2 py-1 rounded text-[10px] font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-slate-200 flex items-center gap-1 cursor-pointer transition-colors"
                          title={t('reminders.snooze15m')}
                        >
                          <i className="pi pi-clock text-[9px]" />
                          {t('reminders.quickAdd15m')}
                        </button>
                        <button
                          onClick={() =>
                            snoozeMutation.mutate({ id: rem.id, payload: { minutes: 60 } })
                          }
                          disabled={snoozeMutation.isPending}
                          className="px-2 py-1 rounded text-[10px] font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-slate-200 flex items-center gap-1 cursor-pointer transition-colors"
                          title={t('reminders.snooze1h')}
                        >
                          <i className="pi pi-clock text-[9px]" />
                          {t('reminders.quickAdd1h')}
                        </button>
                      </div>

                      <button
                        onClick={() => dismissMutation.mutate(rem.id)}
                        disabled={dismissMutation.isPending}
                        className="px-2.5 py-1 rounded text-[10px] font-semibold bg-emerald-50 dark:bg-emerald-950/40 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-100 dark:hover:bg-emerald-900/50 flex items-center gap-1 cursor-pointer transition-colors"
                      >
                        <i className="pi pi-check text-[9px]" />
                        {t('reminders.dismiss')}
                      </button>
                    </div>
                  </div>
                )
              })
            )}
          </div>
        </div>
      )}
    </div>
  )
}
