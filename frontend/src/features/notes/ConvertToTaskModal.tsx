import React, { useState } from 'react'

import { useTranslation } from '@/i18n'
import type { ConvertToTaskPayload, Note, TaskPriority } from '@/types'

interface ConvertToTaskModalProps {
  note: Note | null
  isOpen: boolean
  isPending?: boolean
  onClose: () => void
  onConvert: (noteId: number, payload: ConvertToTaskPayload) => Promise<void>
}

export const ConvertToTaskModal: React.FC<ConvertToTaskModalProps> = ({
  note,
  isOpen,
  isPending = false,
  onClose,
  onConvert,
}) => {
  const { t } = useTranslation()
  const [priority, setPriority] = useState<TaskPriority>('medium')
  const [subtasksInput, setSubtasksInput] = useState('')
  const [localSubmitting, setLocalSubmitting] = useState(false)

  if (!isOpen || !note) return null

  const isSubmitting = isPending || localSubmitting

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLocalSubmitting(true)

    const subtasks = subtasksInput
      .split('\n')
      .map((s) => s.trim())
      .filter(Boolean)

    try {
      await onConvert(note.id, {
        priority,
        subtasks,
      })
      onClose()
    } finally {
      setLocalSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/45 dark:bg-slate-950/75 backdrop-blur-sm animate-fade-in">
      <div className="relative w-full max-w-md rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-2xl text-slate-900 dark:text-slate-100">
        <div className="flex items-center justify-between pb-3 border-b border-slate-200 dark:border-slate-800 mb-4">
          <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <i className="pi pi-arrow-up-right text-indigo-500 dark:text-indigo-400" />
            {t('convertModal.title')}
          </h2>
          <button
            onClick={onClose}
            title={t('common.close')}
            className="p-1 rounded text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
          >
            <i className="pi pi-times text-xs" />
          </button>
        </div>

        <div className="mb-4 p-3 rounded-lg bg-slate-50 dark:bg-slate-950/80 border border-slate-200 dark:border-slate-800/80">
          <p className="text-xs font-medium text-slate-800 dark:text-slate-200 line-clamp-2">{note.title}</p>
          {note.content && (
            <p className="text-[11px] text-slate-500 dark:text-slate-400 line-clamp-3 mt-1 leading-relaxed">{note.content}</p>
          )}
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
              {t('convertModal.priorityLabel')} <span className="text-rose-500 font-bold ml-0.5">*</span>
            </label>
            <select
              value={priority}
              onChange={(e) => setPriority(e.target.value as TaskPriority)}
              className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3 py-2 text-xs text-slate-900 dark:text-slate-100 focus:border-indigo-500 focus:outline-none transition-colors"
            >
              <option value="low" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">{t('tasks.priority.low')}</option>
              <option value="medium" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">{t('tasks.priority.medium')}</option>
              <option value="high" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">{t('tasks.priority.high')}</option>
              <option value="urgent" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">{t('tasks.priority.urgent')}</option>
            </select>
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
              {t('tasks.checklist')}
            </label>
            <textarea
              rows={4}
              value={subtasksInput}
              onChange={(e) => setSubtasksInput(e.target.value)}
              placeholder="Example:&#10;Step 1: Research documentation&#10;Step 2: Build prototype&#10;Step 3: Review and ship"
              className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3 py-2 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none resize-none transition-colors"
            />
          </div>

          <div className="flex items-center justify-end gap-2 pt-3 border-t border-slate-200 dark:border-slate-800">
            <button
              type="button"
              onClick={onClose}
              className="px-3 py-1.5 text-xs text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
              {t('common.cancel')}
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className="px-4 py-1.5 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-lg disabled:opacity-50 flex items-center gap-1.5 shadow-xs transition-colors"
            >
              {isSubmitting && <i className="pi pi-spin pi-spinner text-xs" />}
              {t('convertModal.submit')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
