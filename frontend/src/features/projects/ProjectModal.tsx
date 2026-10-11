import React, { useEffect, useState } from 'react'

import { useCreateProject, useUpdateProject } from '@/hooks'
import { useTranslation } from '@/i18n'
import type { Project } from '@/types'

const PRESET_COLORS = [
  '#6366f1', // Indigo
  '#8b5cf6', // Violet
  '#ec4899', // Pink
  '#f43f5e', // Rose
  '#f59e0b', // Amber
  '#10b981', // Emerald
  '#06b6d4', // Cyan
  '#3b82f6' // Blue
]

interface ProjectModalProps {
  isOpen: boolean
  onClose: () => void
  initialProject?: Project | null
  onSuccess?: (createdOrUpdated: Project) => void
}

export const ProjectModal: React.FC<ProjectModalProps> = ({
  isOpen,
  onClose,
  initialProject,
  onSuccess
}) => {
  const { t } = useTranslation()
  const createMutation = useCreateProject()
  const updateMutation = useUpdateProject()

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [color, setColor] = useState(PRESET_COLORS[0])
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (initialProject) {
      setName(initialProject.name)
      setDescription(initialProject.description || '')
      setColor(initialProject.color || PRESET_COLORS[0])
    } else {
      setName('')
      setDescription('')
      setColor(PRESET_COLORS[0])
    }
  }, [initialProject, isOpen])

  if (!isOpen) return null

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    setIsSubmitting(true)
    try {
      if (initialProject) {
        const res = await updateMutation.mutateAsync({
          id: initialProject.id,
          payload: {
            name: name.trim(),
            description: description.trim(),
            color
          }
        })
        if (onSuccess && res) onSuccess(res)
      } else {
        const res = await createMutation.mutateAsync({
          name: name.trim(),
          description: description.trim(),
          color
        })
        if (onSuccess && res) onSuccess(res)
      }
      onClose()
    } catch {
      // Error is handled by mutation onError callback
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/60 backdrop-blur-sm animate-fade-in">
      <div className="relative w-full max-w-sm rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-2xl text-slate-900 dark:text-slate-100">
        <div className="flex items-center justify-between pb-3 border-b border-slate-200 dark:border-slate-800 mb-4">
          <h3 className="text-sm font-semibold flex items-center gap-2">
            <span
              className="w-3 h-3 rounded-full ring-2 ring-white dark:ring-slate-800"
              style={{ backgroundColor: color }}
            />
            {initialProject ? t('projects.editModalTitle') : t('projects.createModalTitle')}
          </h3>
          <button
            type="button"
            onClick={onClose}
            data-pr-tooltip={t('common.close')}
            className="p-1 rounded text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors cursor-pointer"
          >
            <i className="pi pi-times text-xs" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
              {t('projects.nameLabel')} <span className="text-rose-500 font-bold">*</span>
            </label>
            <input
              type="text"
              required
              autoFocus
              placeholder={t('projects.namePlaceholder')}
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3 py-2 text-xs text-slate-900 dark:text-slate-100 focus:outline-none focus:border-indigo-500 transition-colors"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
              {t('projects.descLabel')}
            </label>
            <textarea
              rows={2}
              placeholder={t('projects.descPlaceholder')}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3 py-2 text-xs text-slate-900 dark:text-slate-100 focus:outline-none focus:border-indigo-500 resize-none transition-colors"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-2">
              {t('projects.colorLabel')}
            </label>
            <div className="flex items-center gap-2">
              {PRESET_COLORS.map((c) => (
                <button
                  key={c}
                  type="button"
                  onClick={() => setColor(c)}
                  className={`w-6 h-6 rounded-full transition-transform ${
                    color === c
                      ? 'scale-125 ring-2 ring-indigo-500 ring-offset-2 dark:ring-offset-slate-900'
                      : 'hover:scale-110'
                  }`}
                  style={{ backgroundColor: c }}
                />
              ))}
            </div>
          </div>

          <div className="pt-2 flex justify-end gap-2 border-t border-slate-200 dark:border-slate-800">
            <button
              type="button"
              onClick={onClose}
              className="px-3 py-1.5 text-xs font-medium rounded-lg border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
              {t('projects.cancel')}
            </button>
            <button
              type="submit"
              disabled={isSubmitting || !name.trim()}
              className="px-4 py-1.5 text-xs font-medium rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white disabled:opacity-50 transition-colors flex items-center gap-1.5"
            >
              {isSubmitting && <i className="pi pi-spin pi-spinner text-xs" />}
              {initialProject ? t('projects.save') : t('projects.create')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
export default ProjectModal
