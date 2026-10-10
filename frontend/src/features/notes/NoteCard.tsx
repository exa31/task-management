import React from 'react'

import { useTranslation } from '@/i18n'
import type { Note } from '@/types'

interface NoteCardProps {
  note: Note
  onEdit: (note: Note) => void
  onDelete: (id: number) => void
  onTogglePin: (note: Note) => void
  onConvertToTask: (note: Note) => void
}

const COLOR_CLASSES: Record<string, string> = {
  default:
    'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 hover:border-slate-300 dark:hover:border-slate-700',
  indigo:
    'border-indigo-200 dark:border-indigo-500/30 bg-indigo-50/70 dark:bg-indigo-950/20 hover:border-indigo-300 dark:hover:border-indigo-500/50',
  purple:
    'border-purple-200 dark:border-purple-500/30 bg-purple-50/70 dark:bg-purple-950/20 hover:border-purple-300 dark:hover:border-purple-500/50',
  blue: 'border-blue-200 dark:border-blue-500/30 bg-blue-50/70 dark:bg-blue-950/20 hover:border-blue-300 dark:hover:border-blue-500/50',
  green:
    'border-emerald-200 dark:border-emerald-500/30 bg-emerald-50/70 dark:bg-emerald-950/20 hover:border-emerald-300 dark:hover:border-emerald-500/50',
  amber:
    'border-amber-200 dark:border-amber-500/30 bg-amber-50/70 dark:bg-amber-950/20 hover:border-amber-300 dark:hover:border-amber-500/50',
  rose: 'border-rose-200 dark:border-rose-500/30 bg-rose-50/70 dark:bg-rose-950/20 hover:border-rose-300 dark:hover:border-rose-500/50'
}

export const NoteCard: React.FC<NoteCardProps> = ({
  note,
  onEdit,
  onDelete,
  onTogglePin,
  onConvertToTask
}) => {
  const { t } = useTranslation()
  const cardColor = COLOR_CLASSES[note.color] || COLOR_CLASSES.default

  return (
    <div
      className={`group relative rounded-xl border p-4 transition-all duration-200 backdrop-blur-md shadow-xs flex flex-col justify-between ${cardColor}`}
    >
      <div>
        {/* Header with Title and Pin */}
        <div className="flex items-start justify-between gap-2 mb-1.5">
          <div className="flex-1">
            <div className="flex items-center gap-1.5 mb-1">
              {note.type === 'knowledge' ? (
                <span className="inline-flex items-center gap-1 px-1.5 py-0.2 rounded text-[10px] font-semibold bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-500/20">
                  <i className="pi pi-book text-[9px]" /> {t('notes.typeKnowledge')}
                </span>
              ) : (
                <span className="inline-flex items-center gap-1 px-1.5 py-0.2 rounded text-[10px] font-semibold bg-indigo-500/10 text-indigo-700 dark:text-indigo-400 border border-indigo-500/20">
                  <i className="pi pi-bolt text-[9px]" /> {t('notes.typeActionable')}
                </span>
              )}
            </div>
            <h3
              onClick={() => onEdit(note)}
              className="font-medium text-slate-800 dark:text-slate-100 text-sm leading-snug cursor-pointer group-hover:text-indigo-600 dark:group-hover:text-indigo-300 transition-colors line-clamp-2"
            >
              {note.title}
            </h3>
          </div>
          <button
            onClick={() => onTogglePin(note)}
            title={note.is_pinned ? t('notes.unpin') : t('notes.pin')}
            className={`p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800/80 transition-colors ${
              note.is_pinned
                ? 'text-amber-500 dark:text-amber-400'
                : 'text-slate-400 dark:text-slate-500 opacity-0 group-hover:opacity-100'
            }`}
          >
            <i className={`pi ${note.is_pinned ? 'pi-bookmark-fill' : 'pi-bookmark'} text-xs`} />
          </button>
        </div>

        {/* Content Snippet */}
        {note.content && (
          <p
            onClick={() => onEdit(note)}
            className="text-xs text-slate-600 dark:text-slate-400 leading-relaxed mb-3 whitespace-pre-wrap line-clamp-5 cursor-pointer"
          >
            {note.content}
          </p>
        )}

        {/* Tags */}
        {note.tags && note.tags.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-3">
            {note.tags.map((tag) => (
              <span
                key={tag}
                className="px-1.5 py-0.5 rounded text-[10px] font-mono bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700/50"
              >
                #{tag}
              </span>
            ))}
          </div>
        )}
      </div>

      {/* Footer Actions */}
      <div className="pt-2 border-t border-slate-100 dark:border-slate-800/60 flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
        <div>
          {note.converted_task_id ? (
            <span className="inline-flex items-center gap-1 text-[11px] text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/40 px-2 py-0.5 rounded-full border border-emerald-200 dark:border-emerald-800/50">
              <i className="pi pi-check text-[9px]" />{' '}
              {t('notes.converted', { id: note.converted_task_id })}
            </span>
          ) : note.type === 'knowledge' ? (
            <span className="inline-flex items-center gap-1 text-[11px] text-slate-400 dark:text-slate-500 italic">
              <i className="pi pi-lock text-[9px]" /> {t('notes.howToStorage')}
            </span>
          ) : (
            <button
              onClick={() => onConvertToTask(note)}
              className="inline-flex items-center gap-1.5 text-[11px] font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 hover:bg-indigo-50 dark:hover:bg-indigo-950/50 px-2 py-1 rounded transition-colors"
              title={t('notes.convertToTask')}
            >
              <i className="pi pi-arrow-up-right text-[10px]" /> {t('notes.convertToTask')}
            </button>
          )}
        </div>

        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            onClick={() => onEdit(note)}
            className="p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
            title={t('common.edit')}
          >
            <i className="pi pi-pencil text-xs" />
          </button>
          <button
            onClick={() => onDelete(note.id)}
            className="p-1 rounded hover:bg-rose-50 dark:hover:bg-rose-950/50 text-slate-400 hover:text-rose-500 dark:hover:text-rose-400"
            title={t('common.delete')}
          >
            <i className="pi pi-trash text-xs" />
          </button>
        </div>
      </div>
    </div>
  )
}
