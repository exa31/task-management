import React from 'react'

import { NoteGrid } from '@/features/notes/NoteGrid'
import { TaskDrawer } from '@/features/tasks/TaskDrawer'
import { useNotes } from '@/hooks'
import { useTranslation } from '@/i18n'

export const NotesPage: React.FC = () => {
  const { data: notes = [] } = useNotes()
  const { t } = useTranslation()

  return (
    <div className="space-y-5 pb-12 animate-fade-in">
      {/* Page Header */}
      <div>
        <h1 className="text-xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2">
          <i className="pi pi-lightbulb text-amber-500 dark:text-amber-400" />
          {t('notes.title', { count: notes.length })}
        </h1>
        <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
          {t('notes.subtitle')}
        </p>
      </div>

      {/* Note Grid */}
      <NoteGrid />

      {/* Drawer */}
      <TaskDrawer />
    </div>
  )
}

export default NotesPage
