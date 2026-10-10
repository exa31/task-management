import React from 'react'

import { NoteGrid } from '@/features/notes/NoteGrid'
import { TaskDrawer } from '@/features/tasks/TaskDrawer'
import { useNotes } from '@/hooks'
import { useTranslation } from '@/i18n'

export const NotesPage: React.FC = () => {
  const { data: rawNotes } = useNotes()
  const notes = Array.isArray(rawNotes) ? rawNotes : []
  const { t } = useTranslation()

  const actionableCount = notes.filter((n) => n.type === 'actionable').length
  const knowledgeCount = notes.filter((n) => n.type === 'knowledge').length
  const pinnedCount = notes.filter((n) => n.is_pinned).length

  return (
    <div className="space-y-6 pb-12 animate-fade-in w-full">
      {/* Page Header */}
      <div>
        <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-xl bg-amber-500/10 text-amber-500 dark:text-amber-400 flex items-center justify-center text-sm">
            <i className="pi pi-lightbulb" />
          </div>
          {t('notes.title', { count: notes.length })}
        </h1>
        <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
          {t('notes.subtitle')}
        </p>
      </div>

      {/* KPI Stats Grid */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs">
          <div className="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs mb-2">
            <span>{t('notes.statsTotal')}</span>
            <i className="pi pi-file text-indigo-500 dark:text-indigo-400 text-xs" />
          </div>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono">{notes.length}</p>
          <span className="text-[10px] text-slate-500">{t('dashboard.savedInNotes')}</span>
        </div>

        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs">
          <div className="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs mb-2">
            <span>{t('notes.statsActionable')}</span>
            <i className="pi pi-bolt text-indigo-500 dark:text-indigo-400 text-xs" />
          </div>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono">{actionableCount}</p>
          <span className="text-[10px] text-slate-500">{t('noteModal.actionableTitle')}</span>
        </div>

        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs">
          <div className="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs mb-2">
            <span>{t('notes.statsKnowledge')}</span>
            <i className="pi pi-book text-emerald-500 dark:text-emerald-400 text-xs" />
          </div>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono">{knowledgeCount}</p>
          <span className="text-[10px] text-slate-500">{t('noteModal.knowledgeTitle')}</span>
        </div>

        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs">
          <div className="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs mb-2">
            <span>{t('notes.statsPinned')}</span>
            <i className="pi pi-bookmark-fill text-amber-500 dark:text-amber-400 text-xs" />
          </div>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono">{pinnedCount}</p>
          <span className="text-[10px] text-slate-500">{t('notes.pin')}</span>
        </div>
      </div>

      {/* Note Grid */}
      <NoteGrid />

      {/* Drawer */}
      <TaskDrawer />
    </div>
  )
}

export default NotesPage
