import React, { useMemo, useState } from 'react'

import { ProjectsSkeleton } from '@/components'
import { ProjectAccordionItem, ProjectModal } from '@/features/projects'
import { TaskDrawer, TaskModal } from '@/features/tasks'
import { useProjects, useTasks } from '@/hooks'
import { useTranslation } from '@/i18n'
import type { Project } from '@/types'

export const ProjectsPage: React.FC = () => {
  const { t } = useTranslation()
  const { data: rawProjects = [], isLoading } = useProjects()
  const { data: rawTasks = [] } = useTasks()
  const projects = useMemo(() => (Array.isArray(rawProjects) ? rawProjects : []), [rawProjects])
  const allTasks = useMemo(() => (Array.isArray(rawTasks) ? rawTasks : []), [rawTasks])

  // Project Task Metrics
  const projectTasks = useMemo(
    () => allTasks.filter((t) => t && typeof t.project_id === 'number' && t.project_id > 0),
    [allTasks]
  )
  const inProgressProjectTasks = useMemo(
    () => projectTasks.filter((t) => t.status === 'in_progress').length,
    [projectTasks]
  )
  const completedProjectTasks = useMemo(
    () => projectTasks.filter((t) => t.status === 'done').length,
    [projectTasks]
  )
  const completionRate = useMemo(
    () =>
      projectTasks.length > 0 ? Math.round((completedProjectTasks / projectTasks.length) * 100) : 0,
    [projectTasks.length, completedProjectTasks]
  )

  const [search, setSearch] = useState('')
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())
  const [isProjectModalOpen, setIsProjectModalOpen] = useState(false)
  const [editingProject, setEditingProject] = useState<Project | null>(null)

  // Filter projects by name
  const filteredProjects = useMemo(() => {
    if (!search.trim()) return projects
    const q = search.toLowerCase()
    return projects.filter(
      (p) => p.name.toLowerCase().includes(q) || p.description?.toLowerCase().includes(q)
    )
  }, [projects, search])

  const toggleExpand = (id: number) => {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const handleExpandAll = () => {
    setExpandedIds(new Set(projects.map((p) => p.id)))
  }

  const handleCollapseAll = () => {
    setExpandedIds(new Set())
  }

  const handleOpenCreate = () => {
    setEditingProject(null)
    setIsProjectModalOpen(true)
  }

  const handleEdit = (proj: Project) => {
    setEditingProject(proj)
    setIsProjectModalOpen(true)
  }

  return (
    <div className="space-y-6 pb-12 animate-fade-in w-full">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-xl bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 flex items-center justify-center text-sm">
              <i className="pi pi-folder" />
            </div>
            {t('projects.title')}
          </h1>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
            {t('projects.subtitle')}
          </p>
        </div>

        <button
          type="button"
          onClick={handleOpenCreate}
          className="inline-flex items-center justify-center gap-2 px-3.5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-xs font-semibold text-white shadow-sm transition-all shrink-0 cursor-pointer"
        >
          <i className="pi pi-plus text-xs" />
          {t('projects.addProject')}
        </button>
      </div>

      {/* Project Metrics Summary Grid */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs">
          <div className="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs mb-2">
            <span>{t('projects.statsTotalProjects')}</span>
            <i className="pi pi-folder text-indigo-500 dark:text-indigo-400 text-xs" />
          </div>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono">
            {projects.length}
          </p>
          <span className="text-[10px] text-slate-500">{t('dashboard.activeProjects')}</span>
        </div>

        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs">
          <div className="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs mb-2">
            <span>{t('projects.statsActiveTasks')}</span>
            <i className="pi pi-check-square text-indigo-500 dark:text-indigo-400 text-xs" />
          </div>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono">
            {projectTasks.length}
          </p>
          <span className="text-[10px] text-slate-500">{t('dashboard.activeTasks')}</span>
        </div>

        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs">
          <div className="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs mb-2">
            <span>{t('projects.statsInProgress')}</span>
            <i className="pi pi-bolt text-amber-500 dark:text-amber-400 text-xs" />
          </div>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono">
            {inProgressProjectTasks}
          </p>
          <span className="text-[10px] text-slate-500">{t('dashboard.inProgress')}</span>
        </div>

        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs">
          <div className="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs mb-2">
            <span>{t('projects.statsCompleted')}</span>
            <i className="pi pi-check-circle text-emerald-500 dark:text-emerald-400 text-xs" />
          </div>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono">
            {completedProjectTasks}
          </p>
          <span className="text-[10px] text-slate-500">{t('dashboard.tasksFinished')}</span>
        </div>

        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs col-span-2 sm:col-span-1">
          <div className="flex items-center justify-between text-slate-500 dark:text-slate-400 text-xs mb-2">
            <span>{t('projects.statsCompletionRate')}</span>
            <i className="pi pi-chart-pie text-purple-500 dark:text-purple-400 text-xs" />
          </div>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono">
            {completionRate}%
          </p>
          <span className="text-[10px] text-slate-500">{t('projects.title')}</span>
        </div>
      </div>

      {/* Control Bar: Search & Expand/Collapse All */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-slate-200 dark:border-slate-800">
        <div className="relative flex-1 sm:max-w-md">
          <i className="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 text-xs" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('projects.searchPlaceholder')}
            className="w-full rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 pl-8 pr-3 py-1.5 text-xs text-slate-900 dark:text-slate-200 placeholder-slate-400 focus:outline-none focus:border-indigo-500 transition-colors shadow-2xs"
          />
        </div>

        {projects.length > 0 && (
          <div className="flex items-center gap-2 self-end sm:self-auto">
            <button
              type="button"
              onClick={handleExpandAll}
              className="px-2.5 py-1 text-xs font-medium rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
              {t('projects.expandAll')}
            </button>
            <span className="text-slate-300 dark:text-slate-700">|</span>
            <button
              type="button"
              onClick={handleCollapseAll}
              className="px-2.5 py-1 text-xs font-medium rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
              {t('projects.collapseAll')}
            </button>
          </div>
        )}
      </div>

      {/* Projects List Accordion */}
      {isLoading ? (
        <ProjectsSkeleton count={4} />
      ) : projects.length === 0 ? (
        <div className="py-20 text-center rounded-2xl border-2 border-dashed border-slate-200 dark:border-slate-800 bg-white/50 dark:bg-slate-900/40 p-8">
          <div className="w-12 h-12 rounded-2xl bg-indigo-500/10 text-indigo-500 flex items-center justify-center text-xl mx-auto mb-3">
            <i className="pi pi-folder-open" />
          </div>
          <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200">
            {t('projects.title')}
          </h3>
          <p className="text-xs text-slate-500 dark:text-slate-400 max-w-sm mx-auto mt-1 mb-4">
            {t('projects.empty')}
          </p>
          <button
            type="button"
            onClick={handleOpenCreate}
            className="px-4 py-2 text-xs font-semibold rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white transition-all shadow-sm inline-flex items-center gap-1.5"
          >
            <i className="pi pi-plus text-xs" />
            {t('projects.addProject')}
          </button>
        </div>
      ) : filteredProjects.length === 0 ? (
        <div className="py-12 text-center text-xs text-slate-400">
          <i className="pi pi-search text-xl mb-1 block" />
          <p>{t('tasks.noTasksFound')}</p>
        </div>
      ) : (
        <div className="space-y-3.5">
          {filteredProjects.map((proj) => (
            <ProjectAccordionItem
              key={proj.id}
              project={proj}
              isExpanded={expandedIds.has(proj.id)}
              onToggleExpand={() => toggleExpand(proj.id)}
              onEdit={handleEdit}
            />
          ))}
        </div>
      )}

      {/* Create / Edit Project Modal */}
      <ProjectModal
        isOpen={isProjectModalOpen}
        onClose={() => setIsProjectModalOpen(false)}
        initialProject={editingProject}
      />

      {/* Task Creation & Drawer Modals */}
      <TaskModal />
      <TaskDrawer />
    </div>
  )
}
export default ProjectsPage
