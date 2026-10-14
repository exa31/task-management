import { useQueryClient } from '@tanstack/react-query'
import React, { useMemo, useState } from 'react'

import { TaskDrawer } from '@/features/tasks/TaskDrawer'
import { TaskModal } from '@/features/tasks/TaskModal'
import {
  calendarKeys,
  reminderKeys,
  useCalendarGraphQL,
  useDebounce,
  useDismissReminder,
  useSnoozeReminder
} from '@/hooks'
import { useTranslation } from '@/i18n'
import { useTaskStore, useUIStore } from '@/stores'
import type { Task } from '@/types'

type CalendarViewMode = 'month' | 'week'
type FilterType = 'all' | 'tasks' | 'reminders' | 'activities'

export const CalendarPage: React.FC = () => {
  const { t, locale } = useTranslation()
  const { openDrawer } = useTaskStore()
  const { openCreateTaskModal } = useUIStore()
  const queryClient = useQueryClient()

  const [currentDate, setCurrentDate] = useState(() => new Date())
  const [selectedDate, setSelectedDate] = useState(() => new Date())
  const [viewMode, setViewMode] = useState<CalendarViewMode>('month')
  const [filterType, setFilterType] = useState<FilterType>('all')

  // Calculate range for query
  const queryRange = useMemo(() => {
    const year = currentDate.getFullYear()
    const month = currentDate.getMonth()
    // 1 week padding before & after
    const start = new Date(year, month - 1, 20)
    const end = new Date(year, month + 2, 10)
    return {
      from: start.toISOString(),
      to: end.toISOString()
    }
  }, [currentDate])

  // Debounce rapid month/week navigation (250ms), combined with AbortController for in-flight cancellation
  const debouncedRange = useDebounce(queryRange, 250)

  const { data, isFetching, refetch } = useCalendarGraphQL(debouncedRange.from, debouncedRange.to)

  const snoozeMutation = useSnoozeReminder()
  const dismissMutation = useDismissReminder()

  const tasks = data?.tasks || []
  const reminders = data?.reminders || []
  const activities = data?.activities || []

  // Date helper: YYYY-MM-DD
  const formatDateKey = (d: Date): string => {
    const y = d.getFullYear()
    const m = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${y}-${m}-${day}`
  }

  // Index items by date key
  const itemsByDate = useMemo(() => {
    const map = new Map<
      string,
      {
        tasks: typeof tasks
        reminders: typeof reminders
        activities: typeof activities
      }
    >()

    const getEntry = (key: string) => {
      let entry = map.get(key)
      if (!entry) {
        entry = { tasks: [], reminders: [], activities: [] }
        map.set(key, entry)
      }
      return entry
    }

    tasks.forEach((item) => {
      if (item.due_date) {
        const key = item.due_date.slice(0, 10)
        getEntry(key).tasks.push(item)
      }
    })

    reminders.forEach((item) => {
      if (item.remind_at) {
        const key = item.remind_at.slice(0, 10)
        getEntry(key).reminders.push(item)
      }
    })

    activities.forEach((item) => {
      if (item.created_at) {
        const key = item.created_at.slice(0, 10)
        getEntry(key).activities.push(item)
      }
    })

    return map
  }, [tasks, reminders, activities])

  // Navigation handlers
  const handlePrev = () => {
    if (viewMode === 'month') {
      setCurrentDate((prev) => new Date(prev.getFullYear(), prev.getMonth() - 1, 1))
    } else {
      setCurrentDate((prev) => new Date(prev.getFullYear(), prev.getMonth(), prev.getDate() - 7))
    }
  }

  const handleNext = () => {
    if (viewMode === 'month') {
      setCurrentDate((prev) => new Date(prev.getFullYear(), prev.getMonth() + 1, 1))
    } else {
      setCurrentDate((prev) => new Date(prev.getFullYear(), prev.getMonth(), prev.getDate() + 7))
    }
  }

  const handleToday = () => {
    const now = new Date()
    setCurrentDate(now)
    setSelectedDate(now)
  }

  // Generate Month Grid Days (Monday start)
  const monthDays = useMemo(() => {
    const year = currentDate.getFullYear()
    const month = currentDate.getMonth()

    const firstDayOfMonth = new Date(year, month, 1)
    const lastDayOfMonth = new Date(year, month + 1, 0)

    // 0 = Sunday, 1 = Monday ... 6 = Saturday
    let startDay = firstDayOfMonth.getDay() - 1
    if (startDay < 0) startDay = 6 // Make Monday 0

    const days: Array<{ date: Date; isCurrentMonth: boolean }> = []

    // Previous month filler
    for (let i = startDay - 1; i >= 0; i--) {
      days.push({
        date: new Date(year, month, -i),
        isCurrentMonth: false
      })
    }

    // Current month days
    for (let i = 1; i <= lastDayOfMonth.getDate(); i++) {
      days.push({
        date: new Date(year, month, i),
        isCurrentMonth: true
      })
    }

    // Next month filler to complete 35 or 42 cells
    const remaining = 42 - days.length
    if (remaining < 7) {
      for (let i = 1; i <= remaining; i++) {
        days.push({
          date: new Date(year, month + 1, i),
          isCurrentMonth: false
        })
      }
    } else {
      const target = days.length <= 35 ? 35 : 42
      const toAdd = target - days.length
      for (let i = 1; i <= toAdd; i++) {
        days.push({
          date: new Date(year, month + 1, i),
          isCurrentMonth: false
        })
      }
    }

    return days
  }, [currentDate])

  // Generate Week Grid Days
  const weekDays = useMemo(() => {
    const curr = new Date(currentDate)
    let day = curr.getDay() - 1
    if (day < 0) day = 6
    const startOfWeek = new Date(curr)
    startOfWeek.setDate(curr.getDate() - day)

    const days: Date[] = []
    for (let i = 0; i < 7; i++) {
      const d = new Date(startOfWeek)
      d.setDate(startOfWeek.getDate() + i)
      days.push(d)
    }
    return days
  }, [currentDate])

  // Weekday Names
  const weekDayNames = useMemo(() => {
    const base = new Date(2026, 9, 12) // Monday
    const names: string[] = []
    for (let i = 0; i < 7; i++) {
      const d = new Date(base)
      d.setDate(base.getDate() + i)
      names.push(d.toLocaleDateString(locale === 'id' ? 'id-ID' : 'en-US', { weekday: 'short' }))
    }
    return names
  }, [locale])

  // Selected date key & data
  const selectedDateKey = formatDateKey(selectedDate)
  const selectedDayItems = itemsByDate.get(selectedDateKey) || {
    tasks: [],
    reminders: [],
    activities: []
  }

  const todayKey = formatDateKey(new Date())

  // Formatting helpers
  const monthTitle = useMemo(
    () =>
      currentDate.toLocaleDateString(locale === 'id' ? 'id-ID' : 'en-US', {
        month: 'long',
        year: 'numeric'
      }),
    [currentDate, locale]
  )

  const selectedDateTitle = useMemo(
    () =>
      selectedDate.toLocaleDateString(locale === 'id' ? 'id-ID' : 'en-US', {
        weekday: 'long',
        day: 'numeric',
        month: 'long',
        year: 'numeric'
      }),
    [selectedDate, locale]
  )

  // Snooze reminder handler
  const handleSnooze = async (id: number, minutes: number) => {
    await snoozeMutation.mutateAsync({ id, payload: { minutes } })
    queryClient.invalidateQueries({ queryKey: calendarKeys.all })
    queryClient.invalidateQueries({ queryKey: reminderKeys.all })
  }

  // Dismiss reminder handler
  const handleDismiss = async (id: number) => {
    await dismissMutation.mutateAsync(id)
    queryClient.invalidateQueries({ queryKey: calendarKeys.all })
    queryClient.invalidateQueries({ queryKey: reminderKeys.all })
  }

  return (
    <div className="h-full flex flex-col space-y-4 animate-fade-in min-h-0">
      {/* Top Header & Toolbar */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 shrink-0">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2">
              <i className="pi pi-calendar text-indigo-500 dark:text-indigo-400" />
              {t('calendar.title')}
            </h1>
          </div>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            {t('calendar.subtitle')}
          </p>
        </div>

        {/* Action Controls */}
        <div className="flex flex-wrap items-center gap-2">
          {/* View Mode Toggle */}
          <div className="inline-flex rounded-lg bg-slate-100 dark:bg-slate-900 p-0.5 border border-slate-200 dark:border-slate-800 text-xs">
            <button
              onClick={() => setViewMode('month')}
              className={`px-3 py-1 rounded-md font-medium transition-colors ${
                viewMode === 'month'
                  ? 'bg-white dark:bg-slate-800 text-indigo-600 dark:text-indigo-300 shadow-sm'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
              }`}
            >
              {t('calendar.monthView')}
            </button>
            <button
              onClick={() => setViewMode('week')}
              className={`px-3 py-1 rounded-md font-medium transition-colors ${
                viewMode === 'week'
                  ? 'bg-white dark:bg-slate-800 text-indigo-600 dark:text-indigo-300 shadow-sm'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
              }`}
            >
              {t('calendar.weekView')}
            </button>
          </div>

          {/* Navigation Buttons */}
          <div className="flex items-center gap-1 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg p-0.5">
            <button
              onClick={handlePrev}
              className="p-1.5 hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 rounded"
              title={t('calendar.prev')}
            >
              <i className="pi pi-chevron-left text-xs" />
            </button>
            <button
              onClick={handleToday}
              className="px-2.5 py-1 text-xs font-medium hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-200 rounded"
            >
              {t('calendar.today')}
            </button>
            <button
              onClick={handleNext}
              className="p-1.5 hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 rounded"
              title={t('calendar.next')}
            >
              <i className="pi pi-chevron-right text-xs" />
            </button>
          </div>

          <span className="text-sm font-semibold text-slate-800 dark:text-slate-200 min-w-[140px] text-center capitalize">
            {monthTitle}
          </span>

          {/* Refresh Button */}
          <button
            onClick={() => refetch()}
            disabled={isFetching}
            className="p-2 border border-slate-200 dark:border-slate-800 rounded-lg bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 transition-colors"
            title={t('calendar.refresh')}
          >
            <i className={`pi pi-sync text-xs ${isFetching ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>

      {/* Filter Tabs */}
      <div className="flex items-center gap-2 border-b border-slate-200 dark:border-slate-800 pb-2 shrink-0">
        <button
          onClick={() => setFilterType('all')}
          className={`px-3 py-1 text-xs font-medium rounded-full transition-colors ${
            filterType === 'all'
              ? 'bg-indigo-600 text-white'
              : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-700'
          }`}
        >
          {t('calendar.filterAll')}
        </button>
        <button
          onClick={() => setFilterType('tasks')}
          className={`inline-flex items-center gap-1.5 px-3 py-1 text-xs font-medium rounded-full transition-colors ${
            filterType === 'tasks'
              ? 'bg-blue-600 text-white'
              : 'bg-blue-500/10 text-blue-600 dark:text-blue-400 hover:bg-blue-500/20'
          }`}
        >
          <i className="pi pi-check-square text-[10px]" />
          {t('calendar.filterTasks')} ({tasks.length})
        </button>
        <button
          onClick={() => setFilterType('reminders')}
          className={`inline-flex items-center gap-1.5 px-3 py-1 text-xs font-medium rounded-full transition-colors ${
            filterType === 'reminders'
              ? 'bg-amber-600 text-white'
              : 'bg-amber-500/10 text-amber-600 dark:text-amber-400 hover:bg-amber-500/20'
          }`}
        >
          <i className="pi pi-bell text-[10px]" />
          {t('calendar.filterReminders')} ({reminders.length})
        </button>
        <button
          onClick={() => setFilterType('activities')}
          className={`inline-flex items-center gap-1.5 px-3 py-1 text-xs font-medium rounded-full transition-colors ${
            filterType === 'activities'
              ? 'bg-emerald-600 text-white'
              : 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-500/20'
          }`}
        >
          <i className="pi pi-history text-[10px]" />
          {t('calendar.filterActivities')} ({activities.length})
        </button>
      </div>

      {/* Main Grid + Agenda Split Layout */}
      <div className="flex-1 grid grid-cols-1 lg:grid-cols-12 gap-4 min-h-0 overflow-hidden">
        {/* Calendar Grid Area */}
        <div className="lg:col-span-8 flex flex-col h-full bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden shadow-sm">
          {/* Weekday Header */}
          <div className="grid grid-cols-7 border-b border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950/60 text-center py-2 shrink-0">
            {weekDayNames.map((name, idx) => (
              <div
                key={idx}
                className="text-[11px] font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
              >
                {name}
              </div>
            ))}
          </div>

          {/* Month / Week Grid */}
          <div className="flex-1 overflow-y-auto">
            {viewMode === 'month' ? (
              <div className="grid grid-cols-7 auto-rows-fr h-full divide-x divide-y divide-slate-100 dark:divide-slate-800/80">
                {monthDays.map((cell, idx) => {
                  const key = formatDateKey(cell.date)
                  const dayItems = itemsByDate.get(key) || {
                    tasks: [],
                    reminders: [],
                    activities: []
                  }
                  const isSelected = key === selectedDateKey
                  const isToday = key === todayKey

                  const visibleTasks =
                    filterType === 'all' || filterType === 'tasks' ? dayItems.tasks : []
                  const visibleReminders =
                    filterType === 'all' || filterType === 'reminders' ? dayItems.reminders : []
                  const visibleActivities =
                    filterType === 'all' || filterType === 'activities' ? dayItems.activities : []

                  return (
                    <div
                      key={idx}
                      onClick={() => setSelectedDate(cell.date)}
                      className={`min-h-[90px] p-1.5 transition-colors cursor-pointer flex flex-col justify-between group relative ${
                        cell.isCurrentMonth
                          ? 'bg-white dark:bg-slate-900'
                          : 'bg-slate-50/50 dark:bg-slate-950/40 opacity-50'
                      } ${
                        isSelected
                          ? 'ring-2 ring-indigo-500 ring-inset bg-indigo-50/30 dark:bg-indigo-950/30'
                          : 'hover:bg-slate-50 dark:hover:bg-slate-800/50'
                      }`}
                    >
                      {/* Cell Header: Day Number */}
                      <div className="flex items-center justify-between">
                        <span
                          className={`text-xs font-semibold inline-flex items-center justify-center w-6 h-6 rounded-full ${
                            isToday
                              ? 'bg-indigo-600 text-white shadow-sm'
                              : isSelected
                                ? 'text-indigo-600 dark:text-indigo-400 font-bold'
                                : 'text-slate-700 dark:text-slate-300'
                          }`}
                        >
                          {cell.date.getDate()}
                        </span>

                        {/* Quick Add Task button on hover */}
                        <button
                          onClick={(e) => {
                            e.stopPropagation()
                            setSelectedDate(cell.date)
                            openCreateTaskModal()
                          }}
                          className="opacity-0 group-hover:opacity-100 p-0.5 text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-opacity"
                          title={t('calendar.createTaskForDate')}
                        >
                          <i className="pi pi-plus text-[10px]" />
                        </button>
                      </div>

                      {/* Item Badges / Chips */}
                      <div className="mt-1 space-y-1 flex-1 overflow-hidden">
                        {visibleTasks.slice(0, 2).map((taskItem) => (
                          <div
                            key={`task-${taskItem.id}`}
                            className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-blue-50 dark:bg-blue-950/50 text-blue-700 dark:text-blue-300 border border-blue-200/60 dark:border-blue-800/60 truncate flex items-center gap-1"
                            title={t('calendar.taskTitlePrefix', { title: taskItem.title })}
                          >
                            <span className="w-1.5 h-1.5 rounded-full bg-blue-500 shrink-0" />
                            <span className="truncate">{taskItem.title}</span>
                          </div>
                        ))}

                        {visibleReminders.slice(0, 2).map((r) => (
                          <div
                            key={`rem-${r.id}`}
                            className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-amber-50 dark:bg-amber-950/50 text-amber-700 dark:text-amber-300 border border-amber-200/60 dark:border-amber-800/60 truncate flex items-center gap-1"
                            title={t('calendar.reminderTitlePrefix', { title: r.title })}
                          >
                            <span className="w-1.5 h-1.5 rounded-full bg-amber-500 shrink-0" />
                            <span className="truncate">{r.title}</span>
                          </div>
                        ))}

                        {visibleActivities.slice(0, 1).map((a) => (
                          <div
                            key={`act-${a.id}`}
                            className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-emerald-50 dark:bg-emerald-950/50 text-emerald-700 dark:text-emerald-300 border border-emerald-200/60 dark:border-emerald-800/60 truncate flex items-center gap-1"
                            title={t('calendar.activityTitlePrefix', {
                              description: a.description
                            })}
                          >
                            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 shrink-0" />
                            <span className="truncate">{a.task_title || a.description}</span>
                          </div>
                        ))}

                        {/* More Items Indicator */}
                        {visibleTasks.length + visibleReminders.length + visibleActivities.length >
                          4 && (
                          <span className="text-[9px] text-slate-400 font-medium px-1 block">
                            {t('calendar.moreItems', {
                              count:
                                visibleTasks.length +
                                visibleReminders.length +
                                visibleActivities.length -
                                4
                            })}
                          </span>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            ) : (
              /* Week View */
              <div className="grid grid-cols-7 h-full divide-x divide-slate-100 dark:divide-slate-800">
                {weekDays.map((dayDate, idx) => {
                  const key = formatDateKey(dayDate)
                  const dayItems = itemsByDate.get(key) || {
                    tasks: [],
                    reminders: [],
                    activities: []
                  }
                  const isSelected = key === selectedDateKey
                  const isToday = key === todayKey

                  return (
                    <div
                      key={idx}
                      onClick={() => setSelectedDate(dayDate)}
                      className={`p-2 flex flex-col space-y-2 cursor-pointer transition-colors ${
                        isSelected
                          ? 'bg-indigo-50/20 dark:bg-indigo-950/20'
                          : 'hover:bg-slate-50 dark:hover:bg-slate-800/30'
                      }`}
                    >
                      <div className="text-center pb-2 border-b border-slate-100 dark:border-slate-800">
                        <span
                          className={`text-sm font-semibold inline-flex items-center justify-center w-7 h-7 rounded-full ${
                            isToday
                              ? 'bg-indigo-600 text-white'
                              : isSelected
                                ? 'text-indigo-600 font-bold'
                                : 'text-slate-700 dark:text-slate-300'
                          }`}
                        >
                          {dayDate.getDate()}
                        </span>
                      </div>

                      <div className="space-y-1.5 flex-1 overflow-y-auto">
                        {dayItems.tasks.map((taskItem) => (
                          <div
                            key={`week-t-${taskItem.id}`}
                            className="p-1.5 rounded-lg text-xs bg-blue-50 dark:bg-blue-950/40 border border-blue-200/60 dark:border-blue-800/60 text-blue-900 dark:text-blue-100"
                          >
                            <div className="font-medium truncate">{taskItem.title}</div>
                            <div className="text-[10px] text-blue-600 dark:text-blue-400 capitalize">
                              {taskItem.status}
                            </div>
                          </div>
                        ))}

                        {dayItems.reminders.map((r) => (
                          <div
                            key={`week-r-${r.id}`}
                            className="p-1.5 rounded-lg text-xs bg-amber-50 dark:bg-amber-950/40 border border-amber-200/60 dark:border-amber-800/60 text-amber-900 dark:text-amber-100"
                          >
                            <div className="font-medium truncate">{r.title}</div>
                            <div className="text-[10px] text-amber-600 dark:text-amber-400">
                              {r.remind_at?.slice(11, 16)}
                            </div>
                          </div>
                        ))}

                        {dayItems.activities.map((a) => (
                          <div
                            key={`week-a-${a.id}`}
                            className="p-1.5 rounded-lg text-xs bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200/60 dark:border-emerald-800/60 text-emerald-900 dark:text-emerald-100"
                          >
                            <div className="font-medium truncate">
                              {a.task_title || t('calendar.activityDefault')}
                            </div>
                            <div className="text-[10px] text-emerald-600 dark:text-emerald-400">
                              {a.description}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        </div>

        {/* Selected Date Agenda & Work Log Panel */}
        <div className="lg:col-span-4 flex flex-col h-full bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden shadow-sm">
          {/* Panel Header */}
          <div className="p-4 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/40 shrink-0">
            <div className="flex items-start justify-between gap-2">
              <div>
                <span className="text-[10px] uppercase font-mono tracking-wider text-slate-400 dark:text-slate-500">
                  {selectedDateKey === todayKey ? t('calendar.today') : t('calendar.agenda')}
                </span>
                <h2 className="text-sm font-bold text-slate-900 dark:text-slate-100 capitalize">
                  {selectedDateTitle}
                </h2>
              </div>
              <button
                onClick={() => openCreateTaskModal()}
                className="px-2.5 py-1 text-xs font-medium rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white transition-colors flex items-center gap-1 shadow-sm"
              >
                <i className="pi pi-plus text-[10px]" />
                {t('calendar.createTaskForDate')}
              </button>
            </div>
            <div className="mt-2 text-[11px] text-slate-500 dark:text-slate-400">
              {t('calendar.itemsSummary', {
                tasks: selectedDayItems.tasks.length,
                reminders: selectedDayItems.reminders.length,
                activities: selectedDayItems.activities.length
              })}
            </div>
          </div>

          {/* Agenda Scrollable Content */}
          <div className="flex-1 overflow-y-auto p-4 space-y-6">
            {/* 1. Tasks Section */}
            <div>
              <div className="flex items-center gap-1.5 text-xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider mb-2">
                <i className="pi pi-check-square text-blue-500" />
                <span>{t('calendar.tasksSection')}</span>
                <span className="text-[10px] text-slate-400 font-mono">
                  ({selectedDayItems.tasks.length})
                </span>
              </div>

              {selectedDayItems.tasks.length === 0 ? (
                <p className="text-xs text-slate-400 italic py-2">{t('calendar.noTasks')}</p>
              ) : (
                <div className="space-y-2">
                  {selectedDayItems.tasks.map((taskItem) => (
                    <div
                      key={taskItem.id}
                      onClick={() => openDrawer(taskItem as unknown as Task)}
                      className="p-2.5 rounded-lg border border-slate-200 dark:border-slate-800 hover:border-indigo-300 dark:hover:border-indigo-600 bg-slate-50/50 dark:bg-slate-800/40 cursor-pointer transition-all hover:shadow-xs group"
                    >
                      <div className="flex items-start justify-between gap-2">
                        <h4 className="text-xs font-semibold text-slate-800 dark:text-slate-200 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">
                          {taskItem.title}
                        </h4>
                        <span
                          className={`text-[9px] px-1.5 py-0.5 rounded font-mono uppercase font-bold shrink-0 ${
                            taskItem.priority === 'urgent'
                              ? 'bg-rose-500/10 text-rose-600 border border-rose-500/20'
                              : taskItem.priority === 'high'
                                ? 'bg-amber-500/10 text-amber-600 border border-amber-500/20'
                                : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300'
                          }`}
                        >
                          {taskItem.priority}
                        </span>
                      </div>
                      {taskItem.project && (
                        <div className="mt-1 flex items-center gap-1 text-[10px] text-slate-500">
                          <span
                            className="w-2 h-2 rounded-full"
                            style={{ backgroundColor: taskItem.project.color || '#6366f1' }}
                          />
                          <span>{taskItem.project.name}</span>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* 2. Reminders Section */}
            <div>
              <div className="flex items-center gap-1.5 text-xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider mb-2">
                <i className="pi pi-bell text-amber-500" />
                <span>{t('calendar.remindersSection')}</span>
                <span className="text-[10px] text-slate-400 font-mono">
                  ({selectedDayItems.reminders.length})
                </span>
              </div>

              {selectedDayItems.reminders.length === 0 ? (
                <p className="text-xs text-slate-400 italic py-2">{t('calendar.noReminders')}</p>
              ) : (
                <div className="space-y-2">
                  {selectedDayItems.reminders.map((rem) => (
                    <div
                      key={rem.id}
                      className="p-2.5 rounded-lg border border-amber-200/60 dark:border-amber-900/40 bg-amber-50/40 dark:bg-amber-950/20"
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div>
                          <h4 className="text-xs font-semibold text-slate-800 dark:text-slate-200">
                            {rem.title}
                          </h4>
                          {rem.task_title && (
                            <p className="text-[10px] text-amber-700 dark:text-amber-400 mt-0.5">
                              🔗 {rem.task_title}
                            </p>
                          )}
                          <p className="text-[10px] text-slate-400 mt-1">
                            ⏰ {rem.remind_at?.slice(11, 16)}
                          </p>
                        </div>
                        <div className="flex items-center gap-1">
                          <button
                            onClick={() => handleSnooze(rem.id, 60)}
                            className="px-2 py-0.5 text-[10px] rounded bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-100"
                            title={t('calendar.snoozeOneHour')}
                          >
                            {t('calendar.snooze1h')}
                          </button>
                          <button
                            onClick={() => handleDismiss(rem.id)}
                            className="px-2 py-0.5 text-[10px] rounded bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-rose-50 hover:text-rose-600"
                            title={t('calendar.dismiss')}
                          >
                            <i className="pi pi-check text-[9px]" />
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* 3. Work Log / Activity History Section */}
            <div>
              <div className="flex items-center gap-1.5 text-xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider mb-2">
                <i className="pi pi-history text-emerald-500" />
                <span>{t('calendar.activitiesSection')}</span>
                <span className="text-[10px] text-slate-400 font-mono">
                  ({selectedDayItems.activities.length})
                </span>
              </div>

              {selectedDayItems.activities.length === 0 ? (
                <p className="text-xs text-slate-400 italic py-2">{t('calendar.noActivities')}</p>
              ) : (
                <div className="relative pl-4 space-y-4 before:absolute before:left-1.5 before:top-2 before:bottom-2 before:w-0.5 before:bg-slate-200 dark:before:bg-slate-800">
                  {selectedDayItems.activities.map((act) => (
                    <div key={act.id} className="relative group">
                      {/* Timeline Dot */}
                      <span className="absolute -left-4 top-1 w-2.5 h-2.5 rounded-full bg-emerald-500 ring-4 ring-white dark:ring-slate-900" />

                      <div className="text-[11px]">
                        <div className="flex items-center gap-1.5 text-slate-400 text-[10px]">
                          <span>{act.created_at?.slice(11, 16)}</span>
                          {act.task_title && (
                            <span className="font-semibold text-slate-700 dark:text-slate-300">
                              • {act.task_title}
                            </span>
                          )}
                        </div>
                        <p className="text-slate-700 dark:text-slate-300 font-medium mt-0.5">
                          {act.description}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Global Task Drawer & Modal */}
      <TaskDrawer />
      <TaskModal />
    </div>
  )
}

export default CalendarPage
