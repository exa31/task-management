import React, { useState } from 'react'
import { Link } from 'react-router-dom'

import { confirmDialog } from '@/components'
import { ConvertToTaskModal } from '@/features/notes/ConvertToTaskModal'
import { NoteCard } from '@/features/notes/NoteCard'
import { NoteModal } from '@/features/notes/NoteModal'
import { TaskDrawer } from '@/features/tasks/TaskDrawer'
import { TaskFocusHorizon } from '@/features/tasks/TaskFocusHorizon'
import { TaskModal } from '@/features/tasks/TaskModal'
import {
  useConvertToTask,
  useCreateNote,
  useDeleteNote,
  useNotes,
  useTasks,
  useUpdateNote,
  useUpdateTask,
} from '@/hooks'
import { useTranslation } from '@/i18n'
import { useAuthStore, useTaskStore, useUIStore } from '@/stores'
import type { ConvertToTaskPayload, CreateNotePayload, Note, Task, UpdateNotePayload } from '@/types'

export default function Dashboard() {
  const { user } = useAuthStore()
  const { openDrawer } = useTaskStore()
  const { openCreateTaskModal } = useUIStore()
  const { t, locale } = useTranslation()

  // TanStack Query server data
  const { data: rawTasks } = useTasks()
  const { data: rawNotes } = useNotes()
  const tasks: Task[] = Array.isArray(rawTasks) ? rawTasks : []
  const notes: Note[] = Array.isArray(rawNotes) ? rawNotes : []

  // Mutations
  const createNoteMutation = useCreateNote()
  const updateNoteMutation = useUpdateNote()
  const deleteNoteMutation = useDeleteNote()
  const convertToTaskMutation = useConvertToTask()
  const updateTaskMutation = useUpdateTask()

  const handleDeleteNote = (id: number) => {
    confirmDialog({
      header: t('notes.deleteConfirmHeader'),
      message: t('notes.deleteConfirmMessage'),
      icon: 'pi pi-trash text-rose-500',
      defaultFocus: 'reject',
      acceptClassName: 'p-button-danger',
      acceptLabel: t('common.delete'),
      rejectLabel: t('common.cancel'),
      accept: async () => {
        await deleteNoteMutation.mutateAsync(id)
      },
    })
  }

  const [quickIdea, setQuickIdea] = useState('')
  const [editingNote, setEditingNote] = useState<Note | null>(null)
  const [convertingNote, setConvertingNote] = useState<Note | null>(null)

  const handleQuickIdeaSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!quickIdea.trim()) return
    await createNoteMutation.mutateAsync({
      title: quickIdea.trim(),
      color: 'indigo',
    })
    setQuickIdea('')
  }

  // Stats
  const inProgressCount = tasks.filter((t) => t?.status === 'in_progress').length
  const focusCount = tasks.filter((t) => t?.is_focus && t?.status !== 'done').length
  const completedCount = tasks.filter((t) => t?.status === 'done').length

  const todayStr = new Date().toLocaleDateString(locale === 'id' ? 'id-ID' : 'en-US', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })

  return (
    <div className="space-y-6 pb-12 animate-fade-in">
      {/* Top Welcome Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <div className="flex items-center gap-2 text-xs font-mono text-slate-500 mb-1">
            <i className="pi pi-calendar text-[10px]" />
            <span>{todayStr}</span>
          </div>
          <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100">
            {t('dashboard.welcome', { name: user?.name || (locale === 'id' ? 'Teman' : 'Friend') })}
          </h1>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            {t('dashboard.subtitle')}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={openCreateTaskModal}
            className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-xs font-medium text-white shadow-xs transition-colors"
          >
            <i className="pi pi-plus text-xs" />
            {t('dashboard.newTask')}
          </button>
        </div>
      </div>

      {/* KPI Stats Grid */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs">
          <div className="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs mb-2">
            <span>{t('dashboard.todaysFocus')}</span>
            <i className="pi pi-star-fill text-amber-500 dark:text-amber-400 text-xs" />
          </div>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono">{focusCount}</p>
          <span className="text-[10px] text-slate-500">{t('dashboard.priorityTasksToday')}</span>
        </div>

        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs">
          <div className="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs mb-2">
            <span>{t('dashboard.inProgress')}</span>
            <i className="pi pi-bolt text-indigo-500 dark:text-indigo-400 text-xs" />
          </div>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono">{inProgressCount}</p>
          <span className="text-[10px] text-slate-500">{t('dashboard.activeTasks')}</span>
        </div>

        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs">
          <div className="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs mb-2">
            <span>{t('dashboard.scratchpadIdeas')}</span>
            <i className="pi pi-lightbulb text-purple-500 dark:text-purple-400 text-xs" />
          </div>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono">{notes.length}</p>
          <span className="text-[10px] text-slate-500">{t('dashboard.savedInNotes')}</span>
        </div>

        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs">
          <div className="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs mb-2">
            <span>{t('dashboard.completedTasks')}</span>
            <i className="pi pi-check-circle text-emerald-500 dark:text-emerald-400 text-xs" />
          </div>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono">{completedCount}</p>
          <span className="text-[10px] text-slate-500">{t('dashboard.tasksFinished')}</span>
        </div>
      </div>

      {/* Today's Focus Horizon Banner */}
      <TaskFocusHorizon />

      {/* Quick Idea Capture Scratchpad Input */}
      <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-gradient-to-r dark:from-slate-900 dark:to-indigo-950/20 p-4 shadow-xs">
        <form onSubmit={handleQuickIdeaSubmit} className="flex items-center gap-2.5">
          <i className="pi pi-bolt text-indigo-500 dark:text-indigo-400 text-sm pl-1" />
          <input
            type="text"
            value={quickIdea}
            onChange={(e) => setQuickIdea(e.target.value)}
            placeholder={t('dashboard.quickIdeaPlaceholder')}
            className="flex-1 bg-transparent text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none"
          />
          <button
            type="submit"
            disabled={createNoteMutation.isPending || !quickIdea.trim()}
            className="px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 text-xs font-medium text-white transition-colors flex items-center gap-1.5 shrink-0"
          >
            {createNoteMutation.isPending ? (
              <i className="pi pi-spin pi-spinner text-xs" />
            ) : (
              <i className="pi pi-plus text-xs" />
            )}
            {t('dashboard.captureIdea')}
          </button>
        </form>
      </div>

      {/* Main Two-Column Hub: Active Tasks & Recent Ideas */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        {/* Active Tasks Panel */}
        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/50 p-4 shadow-xs">
          <div className="flex items-center justify-between mb-3 pb-2 border-b border-slate-100 dark:border-slate-800/80">
            <h2 className="text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
              <i className="pi pi-list text-indigo-500 dark:text-indigo-400" />
              {t('dashboard.activeTasksTitle')}
            </h2>
            <Link
              to="/tasks"
              className="text-xs text-indigo-600 dark:text-indigo-400 hover:text-indigo-500 dark:hover:text-indigo-300 font-medium flex items-center gap-1"
            >
              {t('dashboard.viewAll', { count: tasks.length })} <i className="pi pi-arrow-right text-[10px]" />
            </Link>
          </div>

          <div className="divide-y divide-slate-100 dark:divide-slate-800/60">
            {tasks
              .filter((t) => t && t.status !== 'done')
              .slice(0, 5)
              .map((task) => (
                <div
                  key={task.id}
                  onClick={() => openDrawer(task)}
                  className="group py-2.5 flex items-center justify-between gap-3 hover:bg-slate-50 dark:hover:bg-slate-800/50 px-2 rounded-lg cursor-pointer transition-colors"
                >
                  <div className="flex items-center gap-2.5 min-w-0">
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        updateTaskMutation.mutate({
                          id: task.id,
                          payload: { status: 'done' },
                        })
                      }}
                      className="h-3.5 w-3.5 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-950 flex items-center justify-center shrink-0 hover:border-indigo-500"
                      title={t('dashboard.markCompleted')}
                    />
                    <div className="min-w-0">
                      <p className="text-xs font-medium text-slate-800 dark:text-slate-200 truncate group-hover:text-indigo-600 dark:group-hover:text-indigo-200">
                        {task.title}
                      </p>
                      {task.last_progress_note ? (
                        <p className="text-[10px] text-indigo-600 dark:text-indigo-400 truncate mt-0.5">
                          <i className="pi pi-history text-[8px] mr-1" />
                          {task.last_progress_note}
                        </p>
                      ) : null}
                    </div>
                  </div>

                  <div className="flex items-center gap-2 shrink-0">
                    <span className="text-[10px] font-mono px-2 py-0.5 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700/50">
                      {task.status}
                    </span>
                  </div>
                </div>
              ))}
            {tasks.filter((t) => t && t.status !== 'done').length === 0 && (
              <p className="py-6 text-center text-xs text-slate-500">
                {t('dashboard.noActiveTasks')}
              </p>
            )}
          </div>
        </div>

        {/* Recent Ideas Scratchpad Panel */}
        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/50 p-4 shadow-xs">
          <div className="flex items-center justify-between mb-3 pb-2 border-b border-slate-100 dark:border-slate-800/80">
            <h2 className="text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
              <i className="pi pi-lightbulb text-amber-500 dark:text-amber-400" />
              {t('dashboard.recentIdeasTitle')}
            </h2>
            <Link
              to="/notes"
              className="text-xs text-indigo-600 dark:text-indigo-400 hover:text-indigo-500 dark:hover:text-indigo-300 font-medium flex items-center gap-1"
            >
              {t('dashboard.viewScratchpad', { count: notes.length })} <i className="pi pi-arrow-right text-[10px]" />
            </Link>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
            {notes.slice(0, 4).map((note) => (
              <NoteCard
                key={note.id}
                note={note}
                onEdit={(n) => setEditingNote(n)}
                onDelete={handleDeleteNote}
                onTogglePin={(n) =>
                  updateNoteMutation.mutate({
                    id: n.id,
                    payload: { is_pinned: !n.is_pinned },
                  })
                }
                onConvertToTask={(n) => setConvertingNote(n)}
              />
            ))}
            {notes.length === 0 && (
              <div className="col-span-2 py-6 text-center text-xs text-slate-500">
                {t('dashboard.noIdeasSaved')}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Global Modals & Drawers */}
      <TaskDrawer />
      <TaskModal />

      <NoteModal
        isOpen={Boolean(editingNote)}
        initialNote={editingNote}
        isPending={updateNoteMutation.isPending}
        onClose={() => setEditingNote(null)}
        onSave={async (payload: CreateNotePayload | UpdateNotePayload) => {
          if (editingNote) {
            await updateNoteMutation.mutateAsync({
              id: editingNote.id,
              payload,
            })
          }
        }}
      />

      <ConvertToTaskModal
        isOpen={Boolean(convertingNote)}
        note={convertingNote}
        isPending={convertToTaskMutation.isPending}
        onClose={() => setConvertingNote(null)}
        onConvert={async (noteId: number, payload: ConvertToTaskPayload) => {
          await convertToTaskMutation.mutateAsync({ id: noteId, payload })
        }}
      />
    </div>
  )
}
