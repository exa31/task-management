import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { useNotes, useTasks } from '@/hooks'
import { useTranslation } from '@/i18n'
import { useTaskStore, useUIStore } from '@/stores'

export const CommandPalette: React.FC = () => {
  const {
    isCommandPaletteOpen,
    closeCommandPalette,
    toggleCommandPalette,
    openCreateTaskModal,
    openCreateNoteModal,
    theme,
    toggleTheme
  } = useUIStore()
  const { openDrawer, setActiveView } = useTaskStore()
  const { data: rawTasks } = useTasks()
  const { data: rawNotes } = useNotes()
  const tasks = Array.isArray(rawTasks) ? rawTasks : []
  const notes = Array.isArray(rawNotes) ? rawNotes : []
  const { t, locale, toggleLocale } = useTranslation()
  const navigate = useNavigate()

  const [query, setQuery] = useState('')

  // Reset search query whenever palette is opened/closed
  useEffect(() => {
    if (!isCommandPaletteOpen) {
      setQuery('')
    }
  }, [isCommandPaletteOpen])

  // Listen to keyboard shortcut Cmd+K / Ctrl+K
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        toggleCommandPalette()
      }
      if (e.key === 'Escape' && isCommandPaletteOpen) {
        closeCommandPalette()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isCommandPaletteOpen, toggleCommandPalette, closeCommandPalette])

  if (!isCommandPaletteOpen) return null

  const quickActions = [
    {
      id: 'create-task',
      title: t('commandPalette.createNewTask'),
      icon: 'pi pi-plus text-indigo-500 dark:text-indigo-400',
      badge: t('commandPalette.badgeTask'),
      run: () => {
        closeCommandPalette()
        openCreateTaskModal()
      }
    },
    {
      id: 'create-note',
      title: t('commandPalette.writeNewIdea'),
      icon: 'pi pi-lightbulb text-amber-500 dark:text-amber-400',
      badge: t('commandPalette.badgeNote'),
      run: () => {
        closeCommandPalette()
        openCreateNoteModal()
      }
    },
    {
      id: 'switch-board',
      title: t('commandPalette.switchBoardView'),
      icon: 'pi pi-table text-indigo-500 dark:text-indigo-400',
      badge: t('commandPalette.badgeView'),
      run: () => {
        closeCommandPalette()
        setActiveView('board')
        navigate('/tasks')
      }
    },
    {
      id: 'open-projects',
      title: t('projects.title'),
      icon: 'pi pi-folder text-indigo-500 dark:text-indigo-400',
      badge: t('commandPalette.navigation'),
      run: () => {
        closeCommandPalette()
        navigate('/projects')
      }
    },
    {
      id: 'mcp-settings',
      title: t('commandPalette.openMcpSettings'),
      icon: 'pi pi-code text-indigo-500 dark:text-indigo-400',
      badge: t('commandPalette.badgeSettings'),
      run: () => {
        closeCommandPalette()
        navigate('/mcp-settings')
      }
    },
    {
      id: 'toggle-theme',
      title: t('commandPalette.toggleTheme'),
      icon: theme === 'dark' ? 'pi pi-sun text-amber-400' : 'pi pi-moon text-indigo-500',
      badge: theme === 'dark' ? t('nav.lightMode') : t('nav.darkMode'),
      run: () => {
        toggleTheme()
        closeCommandPalette()
      }
    },
    {
      id: 'switch-language',
      title: t('commandPalette.switchLanguage'),
      icon: 'pi pi-globe text-indigo-400',
      badge: locale.toUpperCase(),
      run: () => {
        toggleLocale()
        closeCommandPalette()
      }
    }
  ]

  const trimmedQuery = query.trim().toLowerCase()

  const filteredActions = trimmedQuery
    ? quickActions.filter(
        (action) =>
          action.title.toLowerCase().includes(trimmedQuery) ||
          action.badge.toLowerCase().includes(trimmedQuery)
      )
    : quickActions

  const filteredTasks = trimmedQuery
    ? tasks.filter((taskItem) => taskItem.title.toLowerCase().includes(trimmedQuery))
    : tasks.slice(0, 4)

  const filteredNotes = trimmedQuery
    ? notes.filter(
        (noteItem) =>
          noteItem.title.toLowerCase().includes(trimmedQuery) ||
          (noteItem.content && noteItem.content.toLowerCase().includes(trimmedQuery))
      )
    : notes.slice(0, 3)

  const hasAnyResults =
    filteredActions.length > 0 || filteredTasks.length > 0 || filteredNotes.length > 0

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-24 p-4 bg-slate-950/40 dark:bg-slate-950/70 backdrop-blur-md animate-fade-in">
      <div className="relative w-full max-w-xl rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-2xl overflow-hidden divide-y divide-slate-200 dark:divide-slate-800">
        {/* Search Input Bar */}
        <div className="flex items-center px-4 py-3 gap-3">
          <i className="pi pi-search text-slate-400 text-sm" />
          <input
            type="text"
            autoFocus
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t('commandPalette.placeholder')}
            className="flex-1 bg-transparent text-sm text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none"
          />
          <kbd className="px-1.5 py-0.5 rounded text-[10px] font-mono bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 border border-slate-200 dark:border-slate-700">
            ESC
          </kbd>
        </div>

        {/* Results / Commands List */}
        <div className="max-h-96 overflow-y-auto p-2 space-y-3">
          {/* Quick Actions */}
          {filteredActions.length > 0 && (
            <div>
              <p className="px-2.5 py-1 text-[10px] font-semibold text-slate-500 uppercase tracking-wider font-mono">
                {t('commandPalette.actions')}
              </p>
              <div className="space-y-0.5">
                {filteredActions.map((action) => (
                  <button
                    key={action.id}
                    onClick={action.run}
                    className="w-full flex items-center justify-between px-3 py-2 rounded-lg text-xs text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800/80 text-left transition-colors cursor-pointer"
                  >
                    <span className="flex items-center gap-2">
                      <i className={action.icon} />
                      {action.title}
                    </span>
                    <span className="text-[10px] text-slate-500 font-mono">{action.badge}</span>
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Tasks Found */}
          {filteredTasks.length > 0 && (
            <div>
              <p className="px-2.5 py-1 text-[10px] font-semibold text-slate-500 uppercase tracking-wider font-mono">
                {t('commandPalette.tasksSection')}
              </p>
              <div className="space-y-0.5">
                {filteredTasks.map((taskItem) => (
                  <button
                    key={taskItem.id}
                    onClick={() => {
                      closeCommandPalette()
                      navigate('/tasks')
                      openDrawer(taskItem)
                    }}
                    className="w-full flex items-center justify-between px-3 py-2 rounded-lg text-xs text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800/80 text-left transition-colors cursor-pointer"
                  >
                    <span className="truncate flex items-center gap-2">
                      <i className="pi pi-check-circle text-slate-500" />
                      {taskItem.title}
                    </span>
                    <span className="text-[10px] font-mono px-1.5 py-0.2 rounded bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-400">
                      {t(`tasks.status.${taskItem.status}`)}
                    </span>
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Notes Found */}
          {filteredNotes.length > 0 && (
            <div>
              <p className="px-2.5 py-1 text-[10px] font-semibold text-slate-500 uppercase tracking-wider font-mono">
                {t('commandPalette.notesSection')}
              </p>
              <div className="space-y-0.5">
                {filteredNotes.map((noteItem) => (
                  <button
                    key={noteItem.id}
                    onClick={() => {
                      closeCommandPalette()
                      navigate('/notes')
                    }}
                    className="w-full flex items-center justify-between px-3 py-2 rounded-lg text-xs text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800/80 text-left transition-colors cursor-pointer"
                  >
                    <span className="truncate flex items-center gap-2">
                      <i className="pi pi-file text-amber-400" />
                      {noteItem.title}
                    </span>
                    <span className="text-[10px] text-slate-500 font-mono">
                      {t('commandPalette.badgeNote')}
                    </span>
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Empty State when no matches */}
          {!hasAnyResults && (
            <div className="py-8 text-center text-xs text-slate-400 dark:text-slate-500">
              <i className="pi pi-search mb-2 text-base text-slate-300 dark:text-slate-600 block" />
              {t('commandPalette.noResults')}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
