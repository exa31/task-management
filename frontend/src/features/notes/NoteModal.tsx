import React, { useEffect, useState } from 'react'

import { getLocalizedErrorMessage } from '@/api'
import { useTranslation } from '@/i18n'
import type { CreateNotePayload, Note, UpdateNotePayload } from '@/types'

interface NoteModalProps {
  isOpen: boolean
  initialNote?: Note | null
  isPending?: boolean
  onClose: () => void
  onSave: (payload: CreateNotePayload | UpdateNotePayload) => Promise<void>
}

const COLORS = [
  { id: 'default', bg: 'bg-slate-400 dark:bg-slate-700' },
  { id: 'indigo', bg: 'bg-indigo-600' },
  { id: 'purple', bg: 'bg-purple-600' },
  { id: 'blue', bg: 'bg-blue-600' },
  { id: 'green', bg: 'bg-emerald-600' },
  { id: 'amber', bg: 'bg-amber-600' },
  { id: 'rose', bg: 'bg-rose-600' }
]

export const NoteModal: React.FC<NoteModalProps> = ({
  isOpen,
  initialNote,
  isPending = false,
  onClose,
  onSave
}) => {
  const { t } = useTranslation()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [tagsInput, setTagsInput] = useState('')
  const [color, setColor] = useState('default')
  const [isPinned, setIsPinned] = useState(false)
  const [noteType, setNoteType] = useState<'actionable' | 'knowledge'>('actionable')
  const [localSubmitting, setLocalSubmitting] = useState(false)
  const [touched, setTouched] = useState(false)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  const isSubmitting = isPending || localSubmitting
  const isTitleInvalid = touched && !title.trim()

  useEffect(() => {
    setErrorMsg(null)
    if (initialNote) {
      setTitle(initialNote.title)
      setContent(initialNote.content || '')
      setTagsInput((initialNote.tags || []).join(', '))
      setColor(initialNote.color || 'default')
      setIsPinned(initialNote.is_pinned || false)
      setNoteType(initialNote.type || 'actionable')
      setTouched(false)
    } else {
      setTitle('')
      setContent('')
      setTagsInput('')
      setColor('default')
      setIsPinned(false)
      setNoteType('actionable')
      setTouched(false)
    }
  }, [initialNote, isOpen])

  if (!isOpen) return null

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setTouched(true)
    setErrorMsg(null)
    if (!title.trim()) return

    const tags = tagsInput
      .split(',')
      .map((t) => t.trim().toLowerCase())
      .filter(Boolean)

    setLocalSubmitting(true)
    try {
      await onSave({
        title: title.trim(),
        content: content.trim(),
        tags,
        color,
        is_pinned: isPinned,
        type: noteType
      })
      setErrorMsg(null)
      onClose()
    } catch (err) {
      setErrorMsg(getLocalizedErrorMessage(err, t))
    } finally {
      setLocalSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/45 dark:bg-slate-950/75 backdrop-blur-sm animate-fade-in">
      <div className="relative w-full max-w-lg rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-2xl text-slate-900 dark:text-slate-100">
        <div className="flex items-center justify-between pb-4 border-b border-slate-200 dark:border-slate-800 mb-4">
          <h2 className="text-base font-semibold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <i className="pi pi-bolt text-indigo-500 dark:text-indigo-400" />
            {initialNote ? t('noteModal.editTitle') : t('noteModal.createTitle')}
          </h2>
          <button
            onClick={onClose}
            title={t('common.close')}
            className="p-1 rounded text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            <i className="pi pi-times text-sm" />
          </button>
        </div>

        {errorMsg && (
          <div className="mb-4 flex items-start gap-2 rounded-lg border border-rose-500/20 bg-rose-500/10 p-3 text-xs text-rose-600 dark:text-rose-400">
            <i className="pi pi-exclamation-circle text-sm mt-0.5 shrink-0" />
            <span className="flex-1 font-medium">{errorMsg}</span>
            <button
              type="button"
              onClick={() => setErrorMsg(null)}
              className="text-rose-500/70 hover:text-rose-500"
            >
              <i className="pi pi-times text-xs" />
            </button>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Note Type Selector */}
          <div>
            <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">
              {t('noteModal.typeLabel')}
            </label>
            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => setNoteType('actionable')}
                className={`p-2.5 rounded-xl border text-left transition-all ${
                  noteType === 'actionable'
                    ? 'border-indigo-500 bg-indigo-50/70 dark:bg-indigo-950/40 text-indigo-950 dark:text-indigo-200 ring-1 ring-indigo-500'
                    : 'border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 text-slate-600 dark:text-slate-400 hover:border-slate-300'
                }`}
              >
                <div className="flex items-center gap-1.5 font-semibold text-xs text-indigo-600 dark:text-indigo-400">
                  <i className="pi pi-bolt text-xs" /> {t('noteModal.actionableTitle')}
                </div>
                <p className="text-[10px] text-slate-500 dark:text-slate-400 mt-0.5">
                  {t('noteModal.actionableDesc')}
                </p>
              </button>

              <button
                type="button"
                onClick={() => setNoteType('knowledge')}
                className={`p-2.5 rounded-xl border text-left transition-all ${
                  noteType === 'knowledge'
                    ? 'border-amber-500 bg-amber-50/70 dark:bg-amber-950/40 text-amber-950 dark:text-amber-200 ring-1 ring-amber-500'
                    : 'border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 text-slate-600 dark:text-slate-400 hover:border-slate-300'
                }`}
              >
                <div className="flex items-center gap-1.5 font-semibold text-xs text-amber-600 dark:text-amber-400">
                  <i className="pi pi-book text-xs" /> {t('noteModal.knowledgeTitle')}
                </div>
                <p className="text-[10px] text-slate-500 dark:text-slate-400 mt-0.5">
                  {t('noteModal.knowledgeDesc')}
                </p>
              </button>
            </div>
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
              {t('noteModal.titleLabel')} <span className="text-rose-500 font-bold ml-0.5">*</span>
            </label>
            <input
              type="text"
              required
              autoFocus
              value={title}
              onBlur={() => setTouched(true)}
              onChange={(e) => {
                setTitle(e.target.value)
                if (touched) setTouched(true)
              }}
              placeholder={t('noteModal.titlePlaceholder')}
              className={`w-full rounded-lg border px-3 py-2 text-sm text-slate-900 dark:text-slate-100 bg-slate-50 dark:bg-slate-950 placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none transition-colors ${
                isTitleInvalid
                  ? 'border-rose-500 focus:border-rose-500 ring-1 ring-rose-500/30'
                  : 'border-slate-200 dark:border-slate-800 focus:border-indigo-500'
              }`}
            />
            {isTitleInvalid && (
              <p className="text-[11px] text-rose-500 mt-1 flex items-center gap-1 font-medium">
                <i className="pi pi-exclamation-circle text-[10px]" />{' '}
                {t('noteModal.titleRequired')}
              </p>
            )}
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
              {t('noteModal.contentLabel')}
            </label>
            <textarea
              rows={5}
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder={t('noteModal.contentPlaceholder')}
              className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none resize-none transition-colors"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
              {t('noteModal.tagsLabel')}
            </label>
            <input
              type="text"
              value={tagsInput}
              onChange={(e) => setTagsInput(e.target.value)}
              placeholder={t('noteModal.tagsPlaceholder')}
              className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3 py-1.5 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none transition-colors"
            />
          </div>

          <div className="flex items-center justify-between pt-2">
            <div>
              <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1.5">
                {t('noteModal.colorTheme')}
              </label>
              <div className="flex items-center gap-1.5">
                {COLORS.map((c) => (
                  <button
                    key={c.id}
                    type="button"
                    onClick={() => setColor(c.id)}
                    className={`h-5 w-5 rounded-full ${c.bg} transition-transform ${
                      color === c.id
                        ? 'ring-2 ring-indigo-500 ring-offset-2 ring-offset-white dark:ring-offset-slate-900 scale-110'
                        : 'opacity-70 hover:opacity-100'
                    }`}
                    title={t(`noteModal.colors.${c.id}`)}
                  />
                ))}
              </div>
            </div>

            <label className="flex items-center gap-2 cursor-pointer mt-4">
              <input
                type="checkbox"
                checked={isPinned}
                onChange={(e) => setIsPinned(e.target.checked)}
                className="rounded border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-950 text-indigo-500 focus:ring-0"
              />
              <span className="text-xs text-slate-700 dark:text-slate-300">
                {t('noteModal.pinLabel')}
              </span>
            </label>
          </div>

          <div className="flex items-center justify-end gap-2 pt-4 border-t border-slate-200 dark:border-slate-800 mt-4">
            <button
              type="button"
              onClick={onClose}
              className="px-3 py-1.5 text-xs font-medium text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
              {t('common.cancel')}
            </button>
            <button
              type="submit"
              disabled={isSubmitting || !title.trim()}
              className="px-4 py-1.5 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-lg disabled:opacity-50 flex items-center gap-1.5 shadow-xs transition-colors"
            >
              {isSubmitting && <i className="pi pi-spin pi-spinner text-xs" />}
              {t('noteModal.save')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
