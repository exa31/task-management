import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { confirmDialog } from '@/components'
import { ReminderModal } from '@/features/reminder'
import {
  useDeleteReminder,
  useDismissReminder,
  useReminders,
  useSnoozeReminder,
  useTasks
} from '@/hooks'
import { useTranslation } from '@/i18n'
import { useTaskStore } from '@/stores'
import type { Reminder, Task } from '@/types'

type FilterTab = 'all' | 'pending' | 'sent' | 'dismissed'

export default function RemindersPage() {
  const { t, locale } = useTranslation()
  const navigate = useNavigate()
  const { openDrawer } = useTaskStore()

  const [activeTab, setActiveTab] = useState<FilterTab>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [selectedReminder, setSelectedReminder] = useState<Reminder | null>(null)

  // Fetch reminders
  const { data: reminderData, isLoading } = useReminders({
    status: activeTab === 'all' ? 'all' : activeTab
  })
  const reminders = useMemo(() => reminderData?.items || [], [reminderData?.items])

  const { data: rawTasks } = useTasks()
  const tasks = useMemo(() => (Array.isArray(rawTasks) ? rawTasks : []), [rawTasks])

  const deleteMutation = useDeleteReminder()
  const snoozeMutation = useSnoozeReminder()
  const dismissMutation = useDismissReminder()

  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 30000)
    return () => clearInterval(timer)
  }, [])

  // Calculate statistics
  const stats = useMemo(() => {
    const total = reminders.length
    const pending = reminders.filter((r) => r.status === 'pending').length
    const sent = reminders.filter((r) => r.status === 'sent').length
    const dismissed = reminders.filter((r) => r.status === 'dismissed').length
    return { total, pending, sent, dismissed }
  }, [reminders])

  // Filter by search
  const filteredReminders = useMemo(() => {
    if (!searchQuery.trim()) return reminders
    const q = searchQuery.toLowerCase()
    return reminders.filter(
      (r) =>
        r.title.toLowerCase().includes(q) ||
        (r.description && r.description.toLowerCase().includes(q)) ||
        (r.task_title && r.task_title.toLowerCase().includes(q))
    )
  }, [reminders, searchQuery])

  const handleOpenTask = (reminder: Reminder) => {
    if (!reminder.task_id) return
    const target = tasks.find((tk: Task) => tk.id === reminder.task_id)
    if (target) {
      openDrawer(target)
    } else {
      navigate('/tasks')
    }
  }

  const formatDateTime = (dateStr: string) => {
    const d = new Date(dateStr)
    return d.toLocaleString(locale === 'id' ? 'id-ID' : 'en-US', {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: locale !== 'id'
    })
  }

  const formatOverdueDuration = (dateStr: string) => {
    const target = new Date(dateStr).getTime()
    const diffMs = now - target
    const diffMins = Math.max(1, Math.round(diffMs / 60000))

    if (diffMins < 60) return t('reminders.durationMinutes', { count: diffMins })
    const hours = Math.round(diffMins / 60)
    if (hours < 24) return t('reminders.durationHours', { count: hours })
    return t('reminders.durationDays', { count: Math.round(hours / 24) })
  }

  const formatRelativeTime = (dateStr: string) => {
    const target = new Date(dateStr).getTime()
    const diffMs = target - now
    const diffMins = Math.round(diffMs / 60000)

    if (diffMins <= 0) return t('reminders.justNow')
    if (diffMins < 60) return t('reminders.inMinutes', { count: diffMins })
    const hours = Math.round(diffMins / 60)
    if (hours < 24) return t('reminders.inHours', { count: hours })
    return t('reminders.inDays', { count: Math.round(hours / 24) })
  }

  return (
    <div className="w-full flex-1 flex flex-col gap-6">
      {/* Top Header Card */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 p-6 rounded-2xl bg-white/70 dark:bg-slate-900/70 border border-slate-200/80 dark:border-slate-800/80 backdrop-blur-md shadow-xs">
        <div>
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-indigo-600 to-violet-500 text-white flex items-center justify-center shadow-md shadow-indigo-500/20">
              <i className="pi pi-bell text-lg" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">
                {t('reminders.title')}
              </h1>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                {t('reminders.subtitle')}
              </p>
            </div>
          </div>
        </div>

        {/* Action Button */}
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => {
              setSelectedReminder(null)
              setIsModalOpen(true)
            }}
            className="inline-flex items-center justify-center gap-1.5 px-3.5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-xs font-semibold text-white shadow-sm shadow-indigo-600/20 transition-all shrink-0 cursor-pointer"
          >
            <i className="pi pi-plus text-xs" />
            <span>{t('reminders.newReminder')}</span>
          </button>
        </div>
      </div>

      {/* Filter Tabs & Search Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        {/* Tabs */}
        <div className="flex items-center gap-1.5 p-1 rounded-xl bg-slate-100 dark:bg-slate-900/80 border border-slate-200/80 dark:border-slate-800 w-fit">
          <button
            onClick={() => setActiveTab('all')}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer ${
              activeTab === 'all'
                ? 'bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 shadow-xs'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
            }`}
          >
            {t('reminders.tabAll')}
          </button>
          <button
            onClick={() => setActiveTab('pending')}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer flex items-center gap-1.5 ${
              activeTab === 'pending'
                ? 'bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 shadow-xs'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
            }`}
          >
            <span>{t('reminders.tabPending')}</span>
            {stats.pending > 0 && (
              <span className="px-1.5 py-0.2 rounded-full text-[10px] font-bold bg-amber-100 dark:bg-amber-950/60 text-amber-700 dark:text-amber-400">
                {stats.pending}
              </span>
            )}
          </button>
          <button
            onClick={() => setActiveTab('sent')}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer ${
              activeTab === 'sent'
                ? 'bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 shadow-xs'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
            }`}
          >
            {t('reminders.tabSent')}
          </button>
          <button
            onClick={() => setActiveTab('dismissed')}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer ${
              activeTab === 'dismissed'
                ? 'bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 shadow-xs'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
            }`}
          >
            {t('reminders.tabDismissed')}
          </button>
        </div>

        {/* Search */}
        <div className="relative w-full sm:w-64">
          <i className="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-xs text-slate-400 z-10" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder={t('reminders.searchPlaceholder')}
            className="w-full pl-8 pr-3 py-1.5 text-xs rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:border-indigo-500 transition-colors"
          />
        </div>
      </div>

      {/* Reminders List / Grid */}
      {isLoading ? (
        <div className="flex items-center justify-center py-20">
          <div className="flex items-center gap-2 text-xs text-slate-400">
            <i className="pi pi-spin pi-spinner text-base text-indigo-500" />
            <span>{t('reminders.loading')}</span>
          </div>
        </div>
      ) : filteredReminders.length === 0 ? (
        <div className="flex flex-col items-center justify-center p-12 rounded-2xl bg-white/50 dark:bg-slate-900/50 border border-slate-200/80 dark:border-slate-800/80 text-center">
          <div className="w-14 h-14 rounded-2xl bg-indigo-50 dark:bg-indigo-950/40 text-indigo-600 dark:text-indigo-400 flex items-center justify-center mb-3">
            <i className="pi pi-bell text-2xl" />
          </div>
          <h3 className="text-sm font-bold text-slate-900 dark:text-slate-100">
            {t('reminders.emptyTitle')}
          </h3>
          <p className="text-xs text-slate-500 dark:text-slate-400 max-w-sm mt-1 mb-4">
            {t('reminders.emptyDesc')}
          </p>
          <button
            onClick={() => {
              setSelectedReminder(null)
              setIsModalOpen(true)
            }}
            className="px-4 py-2 rounded-xl text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white shadow-xs transition-all flex items-center gap-1.5 cursor-pointer"
          >
            <i className="pi pi-plus text-xs" />
            <span>{t('reminders.newReminder')}</span>
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredReminders.map((rem: Reminder) => {
            const isPending = rem.status === 'pending'
            const isSent = rem.status === 'sent'
            const isDismissed = rem.status === 'dismissed'
            const isOverdue = isPending && new Date(rem.remind_at).getTime() < now

            return (
              <div
                key={rem.id}
                className="group relative flex flex-col justify-between p-5 rounded-2xl bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 hover:border-indigo-500/40 hover:shadow-lg dark:hover:shadow-indigo-950/20 transition-all duration-200"
              >
                <div>
                  {/* Status & Relative Time Badge Header */}
                  <div className="flex items-center justify-between gap-2 mb-3">
                    <span
                      className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-semibold ${
                        isOverdue
                          ? 'bg-rose-50 dark:bg-rose-950/40 text-rose-600 dark:text-rose-400 border border-rose-200 dark:border-rose-900/40'
                          : isPending
                            ? 'bg-amber-50 dark:bg-amber-950/40 text-amber-600 dark:text-amber-400 border border-amber-200 dark:border-amber-900/40'
                            : isSent
                              ? 'bg-indigo-50 dark:bg-indigo-950/40 text-indigo-600 dark:text-indigo-400 border border-indigo-200 dark:border-indigo-900/40'
                              : 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 border border-slate-200 dark:border-slate-700'
                      }`}
                    >
                      <span
                        className={`w-1.5 h-1.5 rounded-full ${
                          isOverdue
                            ? 'bg-rose-500 animate-pulse'
                            : isPending
                              ? 'bg-amber-500'
                              : isSent
                                ? 'bg-indigo-500'
                                : 'bg-slate-400'
                        }`}
                      />
                      <span>
                        {isOverdue
                          ? t('reminders.overdue', { time: formatOverdueDuration(rem.remind_at) })
                          : isPending
                            ? formatRelativeTime(rem.remind_at)
                            : isSent
                              ? t('reminders.statusSent')
                              : t('reminders.statusDismissed')}
                      </span>
                    </span>
                  </div>

                  {/* Title */}
                  <h3 className="text-sm font-bold text-slate-900 dark:text-slate-100 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">
                    {rem.title}
                  </h3>

                  {/* Description */}
                  {rem.description && (
                    <p className="text-xs text-slate-500 dark:text-slate-400 mt-1 line-clamp-2 leading-relaxed">
                      {rem.description}
                    </p>
                  )}

                  {/* Date & Time info */}
                  <div className="flex items-center gap-1.5 text-[11px] text-slate-400 font-mono mt-3">
                    <i className="pi pi-calendar text-[10px]" />
                    <span>{formatDateTime(rem.remind_at)}</span>
                    {rem.snooze_count > 0 && (
                      <span className="text-amber-500 ml-1">
                        ({t('reminders.snoozedTimes', { count: rem.snooze_count })})
                      </span>
                    )}
                  </div>

                  {/* Linked Task chip */}
                  {rem.task_title && (
                    <div className="mt-3">
                      <button
                        onClick={() => handleOpenTask(rem)}
                        className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-medium bg-indigo-50/70 dark:bg-indigo-950/40 text-indigo-700 dark:text-indigo-300 hover:bg-indigo-100 dark:hover:bg-indigo-900/50 transition-colors cursor-pointer border border-indigo-100 dark:border-indigo-900/30"
                      >
                        <i className="pi pi-check-square text-[10px]" />
                        <span className="truncate max-w-[220px]">{rem.task_title}</span>
                        <i className="pi pi-arrow-up-right text-[8px] opacity-60" />
                      </button>
                    </div>
                  )}
                </div>

                {/* Footer Controls */}
                <div className="flex items-center justify-between pt-4 mt-4 border-t border-slate-100 dark:border-slate-800">
                  {/* Snooze presets if pending */}
                  {isPending ? (
                    <div className="flex items-center gap-1.5">
                      <button
                        type="button"
                        onClick={() =>
                          snoozeMutation.mutate({ id: rem.id, payload: { minutes: 15 } })
                        }
                        disabled={snoozeMutation.isPending}
                        className="px-2 py-0.5 rounded-md text-[10px] font-semibold border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/80 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300 transition-colors cursor-pointer"
                        title={t('reminders.snooze15m')}
                      >
                        {t('reminders.quickAdd15m')}
                      </button>
                      <button
                        type="button"
                        onClick={() =>
                          snoozeMutation.mutate({ id: rem.id, payload: { minutes: 60 } })
                        }
                        disabled={snoozeMutation.isPending}
                        className="px-2 py-0.5 rounded-md text-[10px] font-semibold border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/80 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300 transition-colors cursor-pointer"
                        title={t('reminders.snooze1h')}
                      >
                        {t('reminders.quickAdd1h')}
                      </button>
                    </div>
                  ) : (
                    <div />
                  )}

                  <div className="flex items-center gap-1">
                    {/* Dismiss Button */}
                    {!isDismissed && (
                      <button
                        type="button"
                        onClick={() => dismissMutation.mutate(rem.id)}
                        disabled={dismissMutation.isPending}
                        title={t('reminders.dismiss')}
                        className="w-7 h-7 rounded-lg flex items-center justify-center text-emerald-600 dark:text-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-950/40 transition-colors cursor-pointer"
                      >
                        <i className="pi pi-check text-xs" />
                      </button>
                    )}

                    {/* Edit Button */}
                    <button
                      type="button"
                      onClick={() => {
                        setSelectedReminder(rem)
                        setIsModalOpen(true)
                      }}
                      title={t('common.edit')}
                      className="w-7 h-7 rounded-lg flex items-center justify-center text-slate-500 dark:text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
                    >
                      <i className="pi pi-pencil text-xs" />
                    </button>

                    {/* Delete Button */}
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
                      className="w-7 h-7 rounded-lg flex items-center justify-center text-rose-500 hover:bg-rose-50 dark:hover:bg-rose-950/40 transition-colors cursor-pointer"
                    >
                      <i className="pi pi-trash text-xs" />
                    </button>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Create / Edit Modal */}
      <ReminderModal
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false)
          setSelectedReminder(null)
        }}
        initialReminder={selectedReminder}
      />
    </div>
  )
}
