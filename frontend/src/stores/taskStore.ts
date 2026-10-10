import { create } from 'zustand'

import type { Task } from '@/types'

interface TaskUIState {
  selectedTask: Task | null
  isDrawerOpen: boolean
  activeView: 'list' | 'board'
  filterStatus: string
  filterPriority: string
  filterTag: string
  filterProject: number | null
  filterFocusOnly: boolean
  searchQuery: string

  setActiveView: (view: 'list' | 'board') => void
  setFilterStatus: (status: string) => void
  setFilterPriority: (priority: string) => void
  setFilterTag: (tag: string) => void
  setFilterProject: (projectId: number | null) => void
  setFilterFocusOnly: (focus: boolean) => void
  setSearchQuery: (query: string) => void
  openDrawer: (task: Task) => void
  closeDrawer: () => void
  setSelectedTask: (task: Task | null) => void
}

export const useTaskStore = create<TaskUIState>((set) => ({
  selectedTask: null,
  isDrawerOpen: false,
  activeView: 'list',
  filterStatus: '',
  filterPriority: '',
  filterTag: '',
  filterProject: null,
  filterFocusOnly: false,
  searchQuery: '',

  setActiveView: (view) => set({ activeView: view }),
  setFilterStatus: (status) => set({ filterStatus: status }),
  setFilterPriority: (priority) => set({ filterPriority: priority }),
  setFilterTag: (tag) => set({ filterTag: tag }),
  setFilterProject: (projectId) => set({ filterProject: projectId }),
  setFilterFocusOnly: (focus) => set({ filterFocusOnly: focus }),
  setSearchQuery: (query) => set({ searchQuery: query }),
  openDrawer: (task) => set({ selectedTask: task, isDrawerOpen: true }),
  closeDrawer: () => set({ selectedTask: null, isDrawerOpen: false }),
  setSelectedTask: (task) => set({ selectedTask: task })
}))
