import React, { useState } from 'react'

import { confirmDialog } from '@/components'
import { useConvertToTask, useCreateNote, useDeleteNote, useNotes, useUpdateNote } from '@/hooks'
import { useTranslation } from '@/i18n'
import type {
  ConvertToTaskPayload,
  CreateNotePayload,
  Note,
  NoteFilterParams,
  UpdateNotePayload
} from '@/types'

import { ConvertToTaskModal } from './ConvertToTaskModal'
import { NoteCard } from './NoteCard'
import { NoteModal } from './NoteModal'

export const NoteGrid: React.FC = () => {
  const [activeTag, setActiveTag] = useState('')
  const [activeType, setActiveType] = useState<'' | 'actionable' | 'knowledge'>('')
  const { data: rawAllNotes } = useNotes()

  const filterParams: NoteFilterParams = {}
  if (activeTag) filterParams.tag = activeTag
  if (activeType) filterParams.type = activeType

  const {
    data: rawNotes,
    isLoading,
    isError,
    error,
    refetch
  } = useNotes(Object.keys(filterParams).length > 0 ? filterParams : undefined)
  const allNotes = Array.isArray(rawAllNotes) ? rawAllNotes : []
  const notes = Array.isArray(rawNotes) ? rawNotes : []
  const { t } = useTranslation()

  const createNoteMutation = useCreateNote()
  const updateNoteMutation = useUpdateNote()
  const deleteNoteMutation = useDeleteNote()
  const convertToTaskMutation = useConvertToTask()

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
      }
    })
  }

  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingNote, setEditingNote] = useState<Note | null>(null)
  const [convertingNote, setConvertingNote] = useState<Note | null>(null)

  // Collect all unique tags from all notes
  const allTags = Array.from(new Set(allNotes.flatMap((n) => n.tags || [])))

  const handleOpenCreate = () => {
    setEditingNote(null)
    setIsModalOpen(true)
  }

  const handleEdit = (note: Note) => {
    setEditingNote(note)
    setIsModalOpen(true)
  }

  const handleTogglePin = async (note: Note) => {
    await updateNoteMutation.mutateAsync({ id: note.id, payload: { is_pinned: !note.is_pinned } })
  }

  const handleSave = async (payload: CreateNotePayload | UpdateNotePayload) => {
    if (editingNote) {
      await updateNoteMutation.mutateAsync({ id: editingNote.id, payload })
    } else {
      await createNoteMutation.mutateAsync(payload as CreateNotePayload)
    }
  }

  const handleConfirmConvert = async (noteId: number, payload: ConvertToTaskPayload) => {
    await convertToTaskMutation.mutateAsync({ id: noteId, payload })
  }

  return (
    <div className="space-y-4">
      {/* Action Bar & Types / Tags */}
      <div className="flex flex-col gap-3 pb-3 border-b border-slate-200 dark:border-slate-800">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          {/* Note Type Filter Tabs */}
          <div className="flex items-center gap-1.5 p-1 rounded-xl bg-slate-100 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 self-start">
            <button
              onClick={() => setActiveType('')}
              className={`px-3 py-1 rounded-lg text-xs font-semibold transition-all ${
                activeType === ''
                  ? 'bg-white dark:bg-slate-800 text-slate-900 dark:text-white shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'
              }`}
            >
              {t('notes.allIdeas', { count: notes.length })}
            </button>
            <button
              onClick={() => setActiveType('actionable')}
              className={`px-3 py-1 rounded-lg text-xs font-semibold transition-all flex items-center gap-1.5 ${
                activeType === 'actionable'
                  ? 'bg-indigo-600 text-white shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400'
              }`}
            >
              <i className="pi pi-bolt text-xs" />
              {t('notes.typeActionable')}
            </button>
            <button
              onClick={() => setActiveType('knowledge')}
              className={`px-3 py-1 rounded-lg text-xs font-semibold transition-all flex items-center gap-1.5 ${
                activeType === 'knowledge'
                  ? 'bg-amber-600 text-white shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-amber-600 dark:hover:text-amber-400'
              }`}
            >
              <i className="pi pi-book text-xs" />
              {t('notes.typeKnowledge')}
            </button>
          </div>

          <button
            onClick={handleOpenCreate}
            className="inline-flex items-center justify-center gap-2 px-3.5 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-xs font-medium text-white shadow-xs transition-colors shrink-0"
          >
            <i className="pi pi-plus text-xs" />
            {t('notes.writeNewIdea')}
          </button>
        </div>

        {/* Tag Filters (if any tags exist) */}
        {allTags.length > 0 && (
          <div className="flex items-center gap-1.5 overflow-x-auto py-0.5 scrollbar-none text-xs">
            <span className="text-[11px] text-slate-400 font-medium mr-1">
              {t('notes.tagsLabel')}:
            </span>
            <button
              onClick={() => setActiveTag('')}
              className={`px-2 py-0.5 rounded-md text-[11px] font-medium transition-colors ${
                activeTag === ''
                  ? 'bg-slate-800 text-white dark:bg-slate-200 dark:text-slate-900'
                  : 'bg-slate-100 dark:bg-slate-900 text-slate-600 dark:text-slate-400 hover:bg-slate-200'
              }`}
            >
              {t('notes.allTags')}
            </button>
            {allTags.map((tag) => (
              <button
                key={tag}
                onClick={() => setActiveTag(tag === activeTag ? '' : tag)}
                className={`px-2 py-0.5 rounded-md text-[11px] font-medium transition-colors ${
                  activeTag === tag
                    ? 'bg-slate-800 text-white dark:bg-slate-200 dark:text-slate-900'
                    : 'bg-slate-100 dark:bg-slate-900 text-slate-600 dark:text-slate-400 hover:bg-slate-200'
                }`}
              >
                #{tag}
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Grid of Notes */}
      {isLoading ? (
        <div className="py-16 text-center text-slate-500 text-xs flex items-center justify-center gap-2">
          <i className="pi pi-spin pi-spinner text-indigo-400" /> {t('notes.loadingIdeas')}
        </div>
      ) : isError ? (
        <div className="py-12 text-center rounded-2xl border border-rose-900/40 bg-rose-950/20 p-8">
          <i className="pi pi-exclamation-circle text-3xl text-rose-400 mb-2" />
          <h3 className="text-sm font-semibold text-rose-300">{t('notes.failedLoad')}</h3>
          <p className="text-xs text-rose-400/80 max-w-sm mx-auto mt-1 mb-4">
            {(error as Error)?.message || 'An error occurred while connecting to the API server.'}
          </p>
          <button
            onClick={() => refetch()}
            className="px-3.5 py-1.5 rounded-lg bg-rose-600 hover:bg-rose-500 text-xs font-medium text-white inline-flex items-center gap-1.5"
          >
            <i className="pi pi-refresh text-xs" /> {t('common.retry')}
          </button>
        </div>
      ) : notes.length === 0 ? (
        <div className="py-20 text-center rounded-2xl border border-dashed border-slate-300 dark:border-slate-800 bg-white/50 dark:bg-slate-950/40 p-8">
          <i className="pi pi-lightbulb text-4xl text-slate-400 dark:text-slate-600 mb-3" />
          <h3 className="text-sm font-medium text-slate-700 dark:text-slate-300">
            {t('notes.noIdeasTitle')}
          </h3>
          <p className="text-xs text-slate-500 max-w-sm mx-auto mt-1 mb-4">
            {t('notes.noIdeasSubtitle')}
          </p>
          <button
            onClick={handleOpenCreate}
            className="px-3.5 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-xs font-medium text-white inline-flex items-center gap-1.5 shadow-xs"
          >
            <i className="pi pi-plus text-xs" /> {t('notes.captureFirst')}
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-3.5">
          {notes.map((note) => (
            <NoteCard
              key={note.id}
              note={note}
              onEdit={handleEdit}
              onDelete={handleDeleteNote}
              onTogglePin={handleTogglePin}
              onConvertToTask={(n) => setConvertingNote(n)}
            />
          ))}
        </div>
      )}

      {/* Modals */}
      <NoteModal
        isOpen={isModalOpen}
        initialNote={editingNote}
        isPending={createNoteMutation.isPending || updateNoteMutation.isPending}
        onClose={() => setIsModalOpen(false)}
        onSave={handleSave}
      />

      <ConvertToTaskModal
        isOpen={Boolean(convertingNote)}
        note={convertingNote}
        isPending={convertToTaskMutation.isPending}
        onClose={() => setConvertingNote(null)}
        onConvert={handleConfirmConvert}
      />
    </div>
  )
}
