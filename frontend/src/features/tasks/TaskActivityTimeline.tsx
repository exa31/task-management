import React, { useState } from 'react'

import { useCreateTaskActivity, useTaskActivities } from '@/hooks'
import { useTranslation } from '@/i18n'
import type { TaskActivity } from '@/types'

interface TaskActivityTimelineProps {
  taskId: number
}

const getActivityBadgeInfo = (
  type: string
): { icon: string; bg: string; text: string; ring: string } => {
  switch (type) {
    case 'task_created':
      return {
        icon: 'pi pi-sparkles',
        bg: 'bg-emerald-100 dark:bg-emerald-950/60',
        text: 'text-emerald-700 dark:text-emerald-300',
        ring: 'ring-emerald-500/30'
      }
    case 'task_updated':
      return {
        icon: 'pi pi-pencil',
        bg: 'bg-blue-100 dark:bg-blue-950/60',
        text: 'text-blue-700 dark:text-blue-300',
        ring: 'ring-blue-500/30'
      }
    case 'progress_logged':
      return {
        icon: 'pi pi-history',
        bg: 'bg-indigo-100 dark:bg-indigo-950/60',
        text: 'text-indigo-700 dark:text-indigo-300',
        ring: 'ring-indigo-500/30'
      }
    case 'subtask_created':
      return {
        icon: 'pi pi-plus',
        bg: 'bg-cyan-100 dark:bg-cyan-950/60',
        text: 'text-cyan-700 dark:text-cyan-300',
        ring: 'ring-cyan-500/30'
      }
    case 'subtask_toggled':
      return {
        icon: 'pi pi-check-circle',
        bg: 'bg-teal-100 dark:bg-teal-950/60',
        text: 'text-teal-700 dark:text-teal-300',
        ring: 'ring-teal-500/30'
      }
    case 'subtask_deleted':
      return {
        icon: 'pi pi-trash',
        bg: 'bg-rose-100 dark:bg-rose-950/60',
        text: 'text-rose-700 dark:text-rose-300',
        ring: 'ring-rose-500/30'
      }
    case 'focus_toggled':
      return {
        icon: 'pi pi-star',
        bg: 'bg-amber-100 dark:bg-amber-950/60',
        text: 'text-amber-700 dark:text-amber-300',
        ring: 'ring-amber-500/30'
      }
    default:
      return {
        icon: 'pi pi-info-circle',
        bg: 'bg-slate-100 dark:bg-slate-800',
        text: 'text-slate-700 dark:text-slate-300',
        ring: 'ring-slate-500/20'
      }
  }
}

export const TaskActivityTimeline: React.FC<TaskActivityTimelineProps> = ({ taskId }) => {
  const { t, locale } = useTranslation()
  const { data: activities = [], isLoading } = useTaskActivities(taskId)
  const createActivityMutation = useCreateTaskActivity()
  const [newNote, setNewNote] = useState('')

  const handlePostNote = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newNote.trim()) return

    await createActivityMutation.mutateAsync({
      taskId,
      description: newNote.trim()
    })
    setNewNote('')
  }

  const formatTimestamp = (dateStr: string) => {
    const d = new Date(dateStr)
    const now = new Date()
    const diffSec = Math.floor((now.getTime() - d.getTime()) / 1000)

    if (diffSec < 60) return t('taskDrawer.justNow')
    if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m`
    if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}h`

    return d.toLocaleDateString(locale === 'id' ? 'id-ID' : 'en-US', {
      day: 'numeric',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  const getActivityTypeLabel = (type: string) => {
    switch (type) {
      case 'task_created':
        return t('taskDrawer.activityType.task_created')
      case 'task_updated':
        return t('taskDrawer.activityType.task_updated')
      case 'progress_logged':
        return t('taskDrawer.activityType.progress_logged')
      case 'subtask_created':
        return t('taskDrawer.activityType.subtask_created')
      case 'subtask_toggled':
        return t('taskDrawer.activityType.subtask_toggled')
      case 'subtask_deleted':
        return t('taskDrawer.activityType.subtask_deleted')
      case 'focus_toggled':
        return t('taskDrawer.activityType.focus_toggled')
      default:
        return type
    }
  }

  return (
    <div className="space-y-4">
      {/* Post Worklog / Progress Note Form */}
      <div className="rounded-xl border border-indigo-200 dark:border-indigo-500/30 bg-indigo-50/60 dark:bg-indigo-950/20 p-3.5 shadow-xs">
        <form onSubmit={handlePostNote} className="space-y-2.5">
          <textarea
            rows={2}
            value={newNote}
            onChange={(e) => setNewNote(e.target.value)}
            placeholder={t('taskDrawer.postActivityPlaceholder')}
            className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-2.5 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none resize-none transition-colors"
          />
          <div className="flex justify-end">
            <button
              type="submit"
              disabled={createActivityMutation.isPending || !newNote.trim()}
              className="px-3.5 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed text-xs font-semibold text-white shadow-xs flex items-center gap-1.5 transition-colors cursor-pointer"
            >
              {createActivityMutation.isPending && (
                <i className="pi pi-spin pi-spinner text-[10px]" />
              )}
              {createActivityMutation.isPending
                ? t('taskDrawer.postingActivity')
                : t('taskDrawer.postActivityBtn')}
            </button>
          </div>
        </form>
      </div>

      {/* Timeline Section */}
      <div>
        <div className="flex items-center justify-between mb-3 px-1">
          <div>
            <h4 className="text-xs font-semibold text-slate-800 dark:text-slate-200">
              {t('taskDrawer.activityTimelineTitle')}
            </h4>
            <p className="text-[11px] text-slate-500 dark:text-slate-400">
              {t('taskDrawer.activityTimelineSubtitle')}
            </p>
          </div>
          <span className="text-[11px] font-mono px-2 py-0.5 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700">
            {activities.length}
          </span>
        </div>

        {isLoading ? (
          <div className="p-8 text-center text-slate-400">
            <i className="pi pi-spin pi-spinner text-lg mb-2" />
            <p className="text-xs">{t('tasks.loadingTasks')}</p>
          </div>
        ) : activities.length === 0 ? (
          <div className="p-8 text-center rounded-xl border border-dashed border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/30">
            <i className="pi pi-history text-2xl text-slate-300 dark:text-slate-600 mb-2" />
            <p className="text-xs text-slate-500 dark:text-slate-400">
              {t('taskDrawer.noActivitiesYet')}
            </p>
          </div>
        ) : (
          <div className="relative pl-6 space-y-4 before:absolute before:left-2.5 before:top-3 before:bottom-3 before:w-0.5 before:bg-slate-200 dark:before:bg-slate-800">
            {activities.map((act: TaskActivity) => {
              const badge = getActivityBadgeInfo(act.activity_type)
              const isProgressNote = act.activity_type === 'progress_logged'

              return (
                <div key={act.id} className="relative group">
                  {/* Timeline Node Icon */}
                  <span
                    className={`absolute -left-6 top-1 flex h-5 w-5 items-center justify-center rounded-full ring-2 ${badge.ring} ${badge.bg} ${badge.text} shrink-0 shadow-xs`}
                  >
                    <i className={`${badge.icon} text-[10px]`} />
                  </span>

                  {/* Activity Entry Card */}
                  <div className="rounded-xl border border-slate-200/80 dark:border-slate-800/80 bg-white dark:bg-slate-900/90 p-3 shadow-xs hover:border-slate-300 dark:hover:border-slate-700 transition-colors">
                    <div className="flex items-center justify-between gap-2 mb-1.5">
                      <span
                        className={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-semibold tracking-wide uppercase ${badge.bg} ${badge.text}`}
                      >
                        {getActivityTypeLabel(act.activity_type)}
                      </span>
                      <time
                        dateTime={act.created_at}
                        className="text-[10px] text-slate-400 dark:text-slate-500 font-mono shrink-0"
                      >
                        {formatTimestamp(act.created_at)}
                      </time>
                    </div>

                    {isProgressNote ? (
                      <div className="p-2 rounded-lg bg-indigo-50/70 dark:bg-indigo-950/30 border border-indigo-100 dark:border-indigo-900/40 text-xs text-indigo-950 dark:text-indigo-200 leading-relaxed font-normal">
                        <p className="whitespace-pre-wrap">{act.description}</p>
                      </div>
                    ) : (
                      <p className="text-xs text-slate-700 dark:text-slate-300 leading-relaxed">
                        {act.description}
                      </p>
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
