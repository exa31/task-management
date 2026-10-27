import { confirmDialog } from 'primereact/confirmdialog'
import { Dialog } from 'primereact/dialog'
import { InputSwitch } from 'primereact/inputswitch'
import React, { useMemo, useState } from 'react'

import {
  useCreateRecurringTask,
  useDeleteRecurringTask,
  useProjects,
  useRecurringTasks,
  useToggleActiveRecurringTask,
  useUpdateRecurringTask
} from '@/hooks'
import { useTranslation } from '@/i18n'
import type {
  CreateRecurringTaskPayload,
  RecurrencePattern,
  RecurringTask,
  TaskPriority,
  UpdateRecurringTaskPayload
} from '@/types'

const PRIORITY_META: Record<TaskPriority, { icon: string; class: string; badge: string }> = {
  urgent: {
    icon: 'pi-exclamation-triangle',
    class: 'text-rose-500 dark:text-rose-400',
    badge: 'bg-rose-50 dark:bg-rose-950/40 text-rose-600 dark:text-rose-400 border-rose-200 dark:border-rose-800/40'
  },
  high: {
    icon: 'pi-angle-double-up',
    class: 'text-amber-600 dark:text-amber-400',
    badge: 'bg-amber-50 dark:bg-amber-950/40 text-amber-600 dark:text-amber-400 border-amber-200 dark:border-amber-800/40'
  },
  medium: {
    icon: 'pi-minus',
    class: 'text-blue-600 dark:text-blue-400',
    badge: 'bg-blue-50 dark:bg-blue-950/40 text-blue-600 dark:text-blue-400 border-blue-200 dark:border-blue-800/40'
  },
  low: {
    icon: 'pi-angle-down',
    class: 'text-slate-500',
    badge: 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-700'
  }
}

export const RecurringTasksManagementView: React.FC = () => {
  const { t } = useTranslation()
  const { data: rawTemplates, isLoading, isError, refetch } = useRecurringTasks()
  const templates = Array.isArray(rawTemplates) ? rawTemplates : []
  const { data: projects = [] } = useProjects()

  const toggleActiveMutation = useToggleActiveRecurringTask()
  const deleteMutation = useDeleteRecurringTask()
  const createMutation = useCreateRecurringTask()
  const updateMutation = useUpdateRecurringTask()

  const [searchQuery, setSearchQuery] = useState('')
  const [filterActiveOnly, setFilterActiveOnly] = useState<'all' | 'active' | 'paused'>('all')

  // Modal State (Create / Edit)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<RecurringTask | null>(null)

  // Form State
  const [formTitle, setFormTitle] = useState('')
  const [formDesc, setFormDesc] = useState('')
  const [formPriority, setFormPriority] = useState<TaskPriority>('medium')
  const [formProjectId, setFormProjectId] = useState<number | null>(null)
  const [formPattern, setFormPattern] = useState<RecurrencePattern>('daily')
  const [formInterval, setFormInterval] = useState<number>(1)
  const [formDueTime, setFormDueTime] = useState<string>('')
  const [formIsActive, setFormIsActive] = useState<boolean>(true)
  const [formSubtasks, setFormSubtasks] = useState<string[]>([])
  const [newSubtaskInput, setNewSubtaskInput] = useState('')
  const [formTouched, setFormTouched] = useState(false)

  const openCreateModal = () => {
    setEditingTemplate(null)
    setFormTitle('')
    setFormDesc('')
    setFormPriority('medium')
    setFormProjectId(null)
    setFormPattern('daily')
    setFormInterval(1)
    setFormDueTime('')
    setFormIsActive(true)
    setFormSubtasks([])
    setNewSubtaskInput('')
    setFormTouched(false)
    setIsModalOpen(true)
  }

  const openEditModal = (template: RecurringTask) => {
    setEditingTemplate(template)
    setFormTitle(template.title || '')
    setFormDesc(template.description || '')
    setFormPriority(template.priority || 'medium')
    setFormProjectId(template.project_id || null)
    setFormPattern(template.recurrence_pattern || 'daily')
    setFormInterval(template.recurrence_interval || 1)
    setFormDueTime(template.due_time ? template.due_time.substring(0, 5) : '')
    setFormIsActive(template.is_active)
    setFormSubtasks(template.subtasks?.map((s) => s.title) || [])
    setNewSubtaskInput('')
    setFormTouched(false)
    setIsModalOpen(true)
  }

  const handleAddSubtask = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newSubtaskInput.trim()) return
    setFormSubtasks([...formSubtasks, newSubtaskInput.trim()])
    setNewSubtaskInput('')
  }

  const handleRemoveSubtask = (index: number) => {
    setFormSubtasks(formSubtasks.filter((_, i) => i !== index))
  }

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormTouched(true)
    if (!formTitle.trim()) return

    const subtaskPayload = formSubtasks.map((st, idx) => ({
      title: st,
      order_index: idx
    }))

    if (editingTemplate) {
      const payload: UpdateRecurringTaskPayload = {
        title: formTitle.trim(),
        description: formDesc.trim(),
        priority: formPriority,
        project_id: formProjectId,
        recurrence_pattern: formPattern,
        recurrence_interval: formInterval,
        due_time: formDueTime ? (formDueTime.length === 5 ? `${formDueTime}:00` : formDueTime) : null,
        is_active: formIsActive,
        subtasks: subtaskPayload
      }
      await updateMutation.mutateAsync({ id: editingTemplate.id, payload })
    } else {
      const payload: CreateRecurringTaskPayload = {
        title: formTitle.trim(),
        description: formDesc.trim(),
        priority: formPriority,
        project_id: formProjectId,
        recurrence_pattern: formPattern,
        recurrence_interval: formInterval,
        due_time: formDueTime ? (formDueTime.length === 5 ? `${formDueTime}:00` : formDueTime) : undefined,
        is_active: formIsActive,
        subtasks: subtaskPayload
      }
      await createMutation.mutateAsync(payload)
    }

    setIsModalOpen(false)
  }

  const handleDeleteTemplate = (template: RecurringTask) => {
    confirmDialog({
      header: t('recurring.deleteConfirmHeader'),
      message: t('recurring.deleteConfirmMessage', { title: template.title }),
      icon: 'pi pi-exclamation-triangle text-rose-500',
      defaultFocus: 'reject',
      acceptClassName: 'p-button-danger',
      acceptLabel: t('recurring.deleteBtn'),
      rejectLabel: t('recurring.cancel'),
      accept: async () => {
        await deleteMutation.mutateAsync(template.id)
      }
    })
  }

  const filteredTemplates = useMemo(() => {
    return templates.filter((tpl) => {
      if (filterActiveOnly === 'active' && !tpl.is_active) return false
      if (filterActiveOnly === 'paused' && tpl.is_active) return false
      if (searchQuery.trim()) {
        const query = searchQuery.toLowerCase()
        const matchesTitle = tpl.title.toLowerCase().includes(query)
        const matchesDesc = (tpl.description || '').toLowerCase().includes(query)
        const matchesProject = (tpl.project_name || '').toLowerCase().includes(query)
        if (!matchesTitle && !matchesDesc && !matchesProject) return false
      }
      return true
    })
  }, [templates, filterActiveOnly, searchQuery])

  const activeCount = useMemo(() => templates.filter((t) => t.is_active).length, [templates])
  const pausedCount = useMemo(() => templates.filter((t) => !t.is_active).length, [templates])

  const getPatternLabel = (pattern: RecurrencePattern, interval: number) => {
    const unitMap: Record<RecurrencePattern, string> = {
      daily: interval > 1 ? `${interval} Hari` : t('recurring.pattern.daily'),
      weekdays: t('recurring.pattern.weekdays'),
      weekly: interval > 1 ? `${interval} Minggu` : t('recurring.pattern.weekly'),
      monthly: interval > 1 ? `${interval} Bulan` : t('recurring.pattern.monthly')
    }
    return unitMap[pattern] || pattern
  }

  return (
    <div className="flex flex-col h-full space-y-4 animate-fade-in min-h-0">
      {/* Header and Controls Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 shrink-0 bg-white/70 dark:bg-slate-900/70 p-3.5 rounded-xl border border-slate-200/80 dark:border-slate-800 backdrop-blur-md shadow-xs">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-lg bg-indigo-50 dark:bg-indigo-950/60 border border-indigo-200/60 dark:border-indigo-800/60 flex items-center justify-center text-indigo-600 dark:text-indigo-400 shrink-0">
            <i className="pi pi-sync text-lg animate-spin-reverse" />
          </div>
          <div>
            <h2 className="text-base font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
              {t('recurring.title')}
              <span className="text-xs px-2 py-0.5 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 font-semibold border border-slate-200 dark:border-slate-700">
                {templates.length}
              </span>
            </h2>
            <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              {t('recurring.subtitle')}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <button
            type="button"
            onClick={openCreateModal}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white text-xs font-semibold shadow-xs transition-colors cursor-pointer"
          >
            <i className="pi pi-plus text-xs" />
            <span>{t('recurring.createTemplateBtn')}</span>
          </button>
        </div>
      </div>

      {/* Filter and Stats Bar */}
      <div className="flex flex-wrap items-center justify-between gap-2.5 shrink-0">
        <div className="flex items-center gap-2">
          {/* Status filter buttons */}
          <div className="inline-flex rounded-lg border border-slate-200 dark:border-slate-800 bg-white/80 dark:bg-slate-900/80 p-0.5 text-xs">
            <button
              type="button"
              onClick={() => setFilterActiveOnly('all')}
              className={`px-2.5 py-1 rounded-md font-medium transition-colors cursor-pointer ${
                filterActiveOnly === 'all'
                  ? 'bg-indigo-600 text-white shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
              }`}
            >
              Semua ({templates.length})
            </button>
            <button
              type="button"
              onClick={() => setFilterActiveOnly('active')}
              className={`px-2.5 py-1 rounded-md font-medium transition-colors cursor-pointer flex items-center gap-1.5 ${
                filterActiveOnly === 'active'
                  ? 'bg-emerald-600 text-white shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
              }`}
            >
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
              {t('recurring.active')} ({activeCount})
            </button>
            <button
              type="button"
              onClick={() => setFilterActiveOnly('paused')}
              className={`px-2.5 py-1 rounded-md font-medium transition-colors cursor-pointer flex items-center gap-1.5 ${
                filterActiveOnly === 'paused'
                  ? 'bg-amber-600 text-white shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
              }`}
            >
              <span className="w-1.5 h-1.5 rounded-full bg-amber-500" />
              {t('recurring.paused')} ({pausedCount})
            </button>
          </div>
        </div>

        {/* Search input */}
        <div className="relative min-w-[200px] max-w-xs">
          <i className="pi pi-search absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400 text-xs pointer-events-none" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Cari template berulang..."
            className="w-full pl-8 pr-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white/90 dark:bg-slate-900/90 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:border-indigo-500 transition-colors"
          />
          {searchQuery && (
            <button
              type="button"
              onClick={() => setSearchQuery('')}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 cursor-pointer"
            >
              <i className="pi pi-times text-[10px]" />
            </button>
          )}
        </div>
      </div>

      {/* Main Content Area */}
      <div className="flex-1 overflow-y-auto min-h-0 pr-1">
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3.5">
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <div
                key={i}
                className="h-44 rounded-xl border border-slate-200/80 dark:border-slate-800 bg-white/60 dark:bg-slate-900/60 animate-pulse p-4"
              />
            ))}
          </div>
        ) : isError ? (
          <div className="p-8 text-center bg-white dark:bg-slate-900 rounded-xl border border-rose-200 dark:border-rose-900/40">
            <i className="pi pi-exclamation-circle text-rose-500 text-2xl mb-2" />
            <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100">
              Gagal memuat template tugas berulang
            </h3>
            <button
              type="button"
              onClick={() => refetch()}
              className="mt-3 px-3 py-1.5 rounded-lg bg-indigo-600 text-white text-xs font-medium cursor-pointer"
            >
              Coba Lagi
            </button>
          </div>
        ) : filteredTemplates.length === 0 ? (
          <div className="flex flex-col items-center justify-center p-12 text-center bg-white/60 dark:bg-slate-900/60 rounded-2xl border border-slate-200/80 dark:border-slate-800">
            <div className="w-12 h-12 rounded-full bg-indigo-50 dark:bg-indigo-950/60 border border-indigo-200 dark:border-indigo-800 flex items-center justify-center text-indigo-500 dark:text-indigo-400 mb-3">
              <i className="pi pi-calendar-plus text-xl" />
            </div>
            <h3 className="text-sm font-bold text-slate-900 dark:text-slate-100">
              {t('recurring.noTemplatesFound')}
            </h3>
            <p className="text-xs text-slate-500 dark:text-slate-400 mt-1 max-w-sm">
              {t('recurring.noTemplatesSubtitle')}
            </p>
            <button
              type="button"
              onClick={openCreateModal}
              className="mt-4 inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-xs transition-colors cursor-pointer"
            >
              <i className="pi pi-plus text-xs" />
              <span>{t('recurring.createTemplateBtn')}</span>
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3.5">
            {filteredTemplates.map((template) => {
              const priorityInfo = PRIORITY_META[template.priority] || PRIORITY_META.medium
              const patternText = getPatternLabel(template.recurrence_pattern, template.recurrence_interval)

              return (
                <div
                  key={template.id}
                  className={`flex flex-col justify-between p-4 rounded-xl border transition-all duration-200 ${
                    template.is_active
                      ? 'bg-white dark:bg-slate-900 border-slate-200 dark:border-slate-800 shadow-xs hover:border-indigo-400/60 dark:hover:border-indigo-600/60'
                      : 'bg-slate-50/70 dark:bg-slate-950/50 border-slate-200/60 dark:border-slate-800/60 opacity-80'
                  }`}
                >
                  <div className="space-y-2.5">
                    {/* Top Row: Project badge & Active Toggle */}
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex items-center gap-1.5 min-w-0">
                        {template.project_name ? (
                          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200/60 dark:border-slate-700/60 truncate">
                            <span
                              className="w-1.5 h-1.5 rounded-full shrink-0"
                              style={{ backgroundColor: template.project_color || '#6366f1' }}
                            />
                            <span className="truncate">{template.project_name}</span>
                          </span>
                        ) : (
                          <span className="text-[10px] text-slate-400">Tanpa Project</span>
                        )}
                        <span
                          className={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-semibold border ${priorityInfo.badge}`}
                        >
                          <i className={`pi ${priorityInfo.icon} text-[9px]`} />
                          <span className="capitalize">{template.priority}</span>
                        </span>
                      </div>

                      {/* Active Toggle Switch */}
                      <div className="flex items-center gap-1.5 shrink-0">
                        <span
                          className={`text-[10px] font-semibold ${
                            template.is_active
                              ? 'text-emerald-600 dark:text-emerald-400'
                              : 'text-amber-600 dark:text-amber-400'
                          }`}
                        >
                          {template.is_active ? t('recurring.active') : t('recurring.paused')}
                        </span>
                        <InputSwitch
                          checked={template.is_active}
                          disabled={toggleActiveMutation.isPending}
                          onChange={() => toggleActiveMutation.mutate(template.id)}
                          className="scale-75"
                        />
                      </div>
                    </div>

                    {/* Title & Description */}
                    <div>
                      <h3 className="text-sm font-bold text-slate-900 dark:text-slate-100 leading-snug line-clamp-1">
                        {template.title}
                      </h3>
                      {template.description && (
                        <p className="text-xs text-slate-500 dark:text-slate-400 mt-1 line-clamp-2 leading-relaxed">
                          {template.description}
                        </p>
                      )}
                    </div>

                    {/* Recurrence & Due Meta Chips */}
                    <div className="flex flex-wrap items-center gap-1.5 pt-1">
                      <span className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-[11px] font-medium bg-indigo-50 dark:bg-indigo-950/50 text-indigo-700 dark:text-indigo-300 border border-indigo-200/60 dark:border-indigo-800/40">
                        <i className="pi pi-sync text-[10px]" />
                        <span>{patternText}</span>
                      </span>

                      {template.due_time && (
                        <span className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-[11px] font-medium bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200/60 dark:border-slate-700/60">
                          <i className="pi pi-clock text-[10px]" />
                          <span>{template.due_time.substring(0, 5)}</span>
                        </span>
                      )}

                      {template.subtasks && template.subtasks.length > 0 && (
                        <span className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-[11px] font-medium bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200/60 dark:border-slate-700/60">
                          <i className="pi pi-check-square text-[10px]" />
                          <span>{template.subtasks.length} Subtasks</span>
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Bottom Row: Actions */}
                  <div className="flex items-center justify-between pt-3 mt-3 border-t border-slate-100 dark:border-slate-800/80">
                    <span className="text-[10px] text-slate-400">
                      ID #{template.id}
                    </span>

                    <div className="flex items-center gap-1">
                      <button
                        type="button"
                        onClick={() => openEditModal(template)}
                        title="Edit template"
                        className="p-1.5 rounded-lg text-slate-500 hover:text-indigo-600 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
                      >
                        <i className="pi pi-pencil text-xs" />
                      </button>
                      <button
                        type="button"
                        onClick={() => handleDeleteTemplate(template)}
                        title="Hapus template master"
                        className="p-1.5 rounded-lg text-slate-500 hover:text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/40 transition-colors cursor-pointer"
                      >
                        <i className="pi pi-trash text-xs" />
                      </button>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>

      {/* Modal: Create / Edit Template */}
      <Dialog
        visible={isModalOpen}
        onHide={() => setIsModalOpen(false)}
        header={
          <div className="flex items-center gap-2 text-slate-900 dark:text-slate-100">
            <i className="pi pi-sync text-indigo-500" />
            <span className="text-sm font-bold">
              {editingTemplate ? t('recurring.editTemplateTitle') : t('recurring.createTemplateTitle')}
            </span>
          </div>
        }
        className="w-full max-w-lg mx-4"
        contentClassName="bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 p-5 rounded-b-xl"
        headerClassName="bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 p-4 border-b border-slate-200 dark:border-slate-800 rounded-t-xl"
      >
        <form onSubmit={handleSave} className="space-y-4 pt-1">
          {/* Title */}
          <div>
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
              Judul Template *
            </label>
            <input
              type="text"
              value={formTitle}
              onChange={(e) => setFormTitle(e.target.value)}
              placeholder="mis. Standup Harian Tim Backend"
              className={`w-full px-3 py-2 rounded-lg border text-xs bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none transition-colors ${
                formTouched && !formTitle.trim()
                  ? 'border-rose-500 focus:border-rose-500'
                  : 'border-slate-200 dark:border-slate-800 focus:border-indigo-500'
              }`}
            />
            {formTouched && !formTitle.trim() && (
              <span className="text-[11px] text-rose-500 mt-1 block">Judul template wajib diisi</span>
            )}
          </div>

          {/* Description */}
          <div>
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
              Deskripsi
            </label>
            <textarea
              rows={2}
              value={formDesc}
              onChange={(e) => setFormDesc(e.target.value)}
              placeholder="Deskripsi atau catatan khusus untuk tugas berulang ini..."
              className="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-800 text-xs bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-indigo-500 transition-colors"
            />
          </div>

          {/* Project & Priority */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                Project / Modul
              </label>
              <select
                value={formProjectId ?? ''}
                onChange={(e) => setFormProjectId(e.target.value ? Number(e.target.value) : null)}
                className="w-full px-2.5 py-2 rounded-lg border border-slate-200 dark:border-slate-800 text-xs bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-indigo-500"
              >
                <option value="">Tanpa Project</option>
                {projects.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                Prioritas
              </label>
              <select
                value={formPriority}
                onChange={(e) => setFormPriority(e.target.value as TaskPriority)}
                className="w-full px-2.5 py-2 rounded-lg border border-slate-200 dark:border-slate-800 text-xs bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-indigo-500"
              >
                <option value="low">Rendah (Low)</option>
                <option value="medium">Sedang (Medium)</option>
                <option value="high">Tinggi (High)</option>
                <option value="urgent">Mendesak (Urgent)</option>
              </select>
            </div>
          </div>

          {/* Recurrence Pattern & Interval */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                Pola Perulangan
              </label>
              <select
                value={formPattern}
                onChange={(e) => setFormPattern(e.target.value as RecurrencePattern)}
                className="w-full px-2.5 py-2 rounded-lg border border-slate-200 dark:border-slate-800 text-xs bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-indigo-500"
              >
                <option value="daily">Setiap Hari (Daily)</option>
                <option value="weekdays">Hari Kerja (Senin–Jumat)</option>
                <option value="weekly">Setiap Minggu (Weekly)</option>
                <option value="monthly">Setiap Bulan (Monthly)</option>
              </select>
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                Interval (Frekuensi)
              </label>
              <input
                type="number"
                min={1}
                max={30}
                value={formInterval}
                onChange={(e) => setFormInterval(Math.max(1, Number(e.target.value)))}
                className="w-full px-2.5 py-2 rounded-lg border border-slate-200 dark:border-slate-800 text-xs bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>

          {/* Due Time & Active Status */}
          <div className="grid grid-cols-2 gap-3 items-center">
            <div>
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                Jam Deadline (Opsional)
              </label>
              <input
                type="time"
                value={formDueTime}
                onChange={(e) => setFormDueTime(e.target.value)}
                className="w-full px-2.5 py-2 rounded-lg border border-slate-200 dark:border-slate-800 text-xs bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-indigo-500"
              />
            </div>

            <div className="pt-4">
              <label className="flex items-center gap-2 cursor-pointer text-xs font-medium text-slate-700 dark:text-slate-300">
                <InputSwitch
                  checked={formIsActive}
                  onChange={(e) => setFormIsActive(Boolean(e.value))}
                />
                <span>Status Template Aktif</span>
              </label>
            </div>
          </div>

          {/* Checklist Subtasks Template */}
          <div className="space-y-2 pt-2 border-t border-slate-200 dark:border-slate-800">
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300">
              Template Subtask (Checklist Bawaan)
            </label>

            <div className="flex gap-2">
              <input
                type="text"
                value={newSubtaskInput}
                onChange={(e) => setNewSubtaskInput(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault()
                    handleAddSubtask(e)
                  }
                }}
                placeholder="Tambahkan butir langkah..."
                className="flex-1 px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 text-xs bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-indigo-500"
              />
              <button
                type="button"
                onClick={handleAddSubtask}
                className="px-3 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-xs font-medium text-slate-700 dark:text-slate-300 cursor-pointer"
              >
                Tambah
              </button>
            </div>

            {formSubtasks.length > 0 && (
              <div className="space-y-1.5 max-h-36 overflow-y-auto pr-1 pt-1">
                {formSubtasks.map((st, idx) => (
                  <div
                    key={idx}
                    className="flex items-center justify-between px-2.5 py-1.5 rounded-md bg-slate-50 dark:bg-slate-950 border border-slate-200/60 dark:border-slate-800/60 text-xs text-slate-800 dark:text-slate-200"
                  >
                    <span className="truncate">{st}</span>
                    <button
                      type="button"
                      onClick={() => handleRemoveSubtask(idx)}
                      className="text-slate-400 hover:text-rose-500 cursor-pointer ml-2"
                    >
                      <i className="pi pi-times text-[10px]" />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Modal Footer */}
          <div className="flex items-center justify-end gap-2 pt-3 border-t border-slate-200 dark:border-slate-800">
            <button
              type="button"
              onClick={() => setIsModalOpen(false)}
              className="px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 text-xs font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 cursor-pointer transition-colors"
            >
              {t('recurring.cancel')}
            </button>
            <button
              type="submit"
              disabled={createMutation.isPending || updateMutation.isPending}
              className="inline-flex items-center gap-1.5 px-4 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-xs disabled:opacity-50 cursor-pointer transition-colors"
            >
              {(createMutation.isPending || updateMutation.isPending) && (
                <i className="pi pi-spin pi-spinner text-xs" />
              )}
              <span>{t('recurring.save')}</span>
            </button>
          </div>
        </form>
      </Dialog>
    </div>
  )
}
