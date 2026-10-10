import React, { useEffect, useState } from 'react'

import { getLocalizedErrorMessage } from '@/api'
import { ProjectModal } from '@/features/projects'
import { useCreateTask, useProjects } from '@/hooks'
import { useTranslation } from '@/i18n'
import { useUIStore } from '@/stores'
import type { TaskPriority, TaskStatus } from '@/types'

export const TaskModal: React.FC = () => {
  const { isCreateTaskModalOpen, closeCreateTaskModal, createTaskInitialProjectId } = useUIStore()
  const createTaskMutation = useCreateTask()
  const { data: projects = [] } = useProjects()
  const { t } = useTranslation()

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [status, setStatus] = useState<TaskStatus>('todo')
  const [priority, setPriority] = useState<TaskPriority>('medium')
  const [projectId, setProjectId] = useState<number | null>(null)
  const [isCreateProjectOpen, setIsCreateProjectOpen] = useState(false)
  const [isFocus, setIsFocus] = useState(false)
  const [tagsInput, setTagsInput] = useState('')
  const [subtasksInput, setSubtasksInput] = useState('')
  const [touched, setTouched] = useState(false)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  useEffect(() => {
    if (isCreateTaskModalOpen) {
      setProjectId(createTaskInitialProjectId)
    }
  }, [isCreateTaskModalOpen, createTaskInitialProjectId])

  if (!isCreateTaskModalOpen) return null

  const isTitleInvalid = touched && !title.trim()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setTouched(true)
    setErrorMsg(null)
    if (!title.trim()) return

    const tags = tagsInput
      .split(',')
      .map((t) => t.trim().toLowerCase())
      .filter(Boolean)

    const subtasks = subtasksInput
      .split('\n')
      .map((s) => s.trim())
      .filter(Boolean)

    try {
      await createTaskMutation.mutateAsync({
        title: title.trim(),
        description: description.trim(),
        status,
        priority,
        project_id: projectId,
        is_focus: isFocus,
        tags,
        subtasks,
      })

      setTitle('')
      setDescription('')
      setStatus('todo')
      setPriority('medium')
      setProjectId(null)
      setIsFocus(false)
      setTagsInput('')
      setSubtasksInput('')
      setTouched(false)
      setErrorMsg(null)
      closeCreateTaskModal()
    } catch (err) {
      setErrorMsg(getLocalizedErrorMessage(err, t))
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/45 dark:bg-slate-950/75 backdrop-blur-sm animate-fade-in">
      <div className="relative w-full max-w-lg rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-2xl text-slate-900 dark:text-slate-100">
        <div className="flex items-center justify-between pb-3 border-b border-slate-200 dark:border-slate-800 mb-4">
          <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <i className="pi pi-plus-circle text-indigo-500 dark:text-indigo-400" />
            {t('taskModal.createTitle')}
          </h2>
          <button
            onClick={closeCreateTaskModal}
            title={t('common.close')}
            className="p-1 rounded text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
          >
            <i className="pi pi-times text-xs" />
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
          <div>
            <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
              {t('taskModal.titleLabel')} <span className="text-rose-500 font-bold ml-0.5">*</span>
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
              placeholder={t('taskModal.titlePlaceholder')}
              className={`w-full rounded-lg border px-3 py-2 text-xs text-slate-900 dark:text-slate-100 bg-slate-50 dark:bg-slate-950 placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none transition-colors ${
                isTitleInvalid
                  ? 'border-rose-500 focus:border-rose-500 ring-1 ring-rose-500/30'
                  : 'border-slate-200 dark:border-slate-800 focus:border-indigo-500'
              }`}
            />
            {isTitleInvalid && (
              <p className="text-[11px] text-rose-500 mt-1 flex items-center gap-1 font-medium">
                <i className="pi pi-exclamation-circle text-[10px]" /> {t('taskModal.titleRequired')}
              </p>
            )}
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
              {t('taskModal.descriptionLabel')}
            </label>
            <textarea
              rows={3}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={t('taskModal.descriptionPlaceholder')}
              className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3 py-2 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none resize-none transition-colors"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                {t('taskModal.statusLabel')} <span className="text-rose-500 font-bold ml-0.5">*</span>
              </label>
              <select
                value={status}
                onChange={(e) => setStatus(e.target.value as TaskStatus)}
                className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:outline-none transition-colors"
              >
                <option value="backlog" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">{t('tasks.status.backlog')}</option>
                <option value="todo" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">{t('tasks.status.todo')}</option>
                <option value="in_progress" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">{t('tasks.status.in_progress')}</option>
                <option value="review" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">{t('tasks.status.review')}</option>
              </select>
            </div>

            <div>
              <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                {t('taskModal.priorityLabel')} <span className="text-rose-500 font-bold ml-0.5">*</span>
              </label>
              <select
                value={priority}
                onChange={(e) => setPriority(e.target.value as TaskPriority)}
                className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:outline-none transition-colors"
              >
                <option value="low" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">{t('tasks.priority.low')}</option>
                <option value="medium" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">{t('tasks.priority.medium')}</option>
                <option value="high" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">{t('tasks.priority.high')}</option>
                <option value="urgent" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">{t('tasks.priority.urgent')}</option>
              </select>
            </div>
          </div>

          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="block text-xs font-medium text-slate-700 dark:text-slate-300">
                {t('taskModal.projectLabel')}
              </label>
              <button
                type="button"
                onClick={() => setIsCreateProjectOpen(true)}
                className="text-[11px] font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-500 hover:underline flex items-center gap-1 cursor-pointer"
              >
                <i className="pi pi-plus text-[9px]" />
                {t('taskModal.createNewProject')}
              </button>
            </div>
            <select
              value={projectId === null ? '' : String(projectId)}
              onChange={(e) => setProjectId(e.target.value ? parseInt(e.target.value, 10) : null)}
              className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:outline-none transition-colors"
            >
              <option value="" className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">
                {t('taskModal.noProjectOption')}
              </option>
              {projects.map((p) => (
                <option key={p.id} value={String(p.id)} className="bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">
                  📁 {p.name}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
              {t('tasks.checklist')}
            </label>
            <textarea
              rows={3}
              value={subtasksInput}
              onChange={(e) => setSubtasksInput(e.target.value)}
              placeholder={t('taskModal.checklistPlaceholder')}
              className="w-full rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3 py-1.5 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none resize-none transition-colors"
            />
          </div>

          <div className="flex items-center justify-between pt-1">
            <input
              type="text"
              value={tagsInput}
              onChange={(e) => setTagsInput(e.target.value)}
              placeholder={t('taskModal.tagsPlaceholder')}
              className="w-60 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-2.5 py-1.5 text-xs text-slate-900 dark:text-slate-200 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none transition-colors"
            />

            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={isFocus}
                onChange={(e) => setIsFocus(e.target.checked)}
                className="rounded border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-950 text-indigo-500 focus:ring-0"
              />
              <span className="text-xs text-amber-500 dark:text-amber-400 flex items-center gap-1 font-medium">
                <i className="pi pi-star-fill text-[10px]" /> {t('dashboard.todaysFocus')}
              </span>
            </label>
          </div>

          <div className="flex items-center justify-end gap-2 pt-3 border-t border-slate-200 dark:border-slate-800">
            <button
              type="button"
              onClick={closeCreateTaskModal}
              className="px-3 py-1.5 text-xs text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
              {t('common.cancel')}
            </button>
            <button
              type="submit"
              disabled={createTaskMutation.isPending || !title.trim()}
              className="px-4 py-1.5 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-lg disabled:opacity-50 flex items-center gap-1.5 shadow-xs transition-colors"
            >
              {createTaskMutation.isPending && (
                <i className="pi pi-spin pi-spinner text-xs" />
              )}
              {t('taskModal.submit')}
            </button>
          </div>
        </form>
      </div>

      {/* Inline Project Modal */}
      <ProjectModal
        isOpen={isCreateProjectOpen}
        onClose={() => setIsCreateProjectOpen(false)}
        onSuccess={(created) => setProjectId(created.id)}
      />
    </div>
  )
}
