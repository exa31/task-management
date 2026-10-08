import { create } from 'zustand'

import { noteApi } from '@/api'
import { appConfig } from '@/config'
import type { ConvertToTaskPayload, CreateNotePayload, Note, NoteFilterParams, UpdateNotePayload } from '@/types'

interface NoteState {
  notes: Note[]
  loading: boolean
  error: string | null
  activeTag: string
  searchQuery: string
  setActiveTag: (tag: string) => void
  setSearchQuery: (query: string) => void
  fetchNotes: (params?: NoteFilterParams) => Promise<void>
  createNote: (payload: CreateNotePayload) => Promise<Note>
  updateNote: (id: number, payload: UpdateNotePayload) => Promise<Note>
  deleteNote: (id: number) => Promise<void>
  convertToTask: (id: number, payload: ConvertToTaskPayload) => Promise<{ note: Note; task_id: number }>
}

export const useNoteStore = create<NoteState>((set, get) => ({
  notes: [],
  loading: false,
  error: null,
  activeTag: '',
  searchQuery: '',

  setActiveTag: (tag) => set({ activeTag: tag }),
  setSearchQuery: (query) => set({ searchQuery: query }),

  fetchNotes: async (params) => {
    set({ loading: true, error: null })
    try {
      const p: NoteFilterParams = {
        ...params,
        tag: get().activeTag || undefined,
        search: get().searchQuery || undefined,
      }
      const res = await noteApi.getAll(p)
      const fetched = res.data || []
      if (appConfig.enableMocks) {
        localStorage.setItem('local_notes', JSON.stringify(fetched))
      }
      set({ notes: fetched, loading: false })
    } catch (err: unknown) {
      if (!appConfig.enableMocks) {
        const errorMsg = err instanceof Error ? err.message : 'Failed to load notes'
        set({ error: errorMsg, loading: false })
        return
      }
      // Local fallback state only if mocks enabled
      const local = localStorage.getItem('local_notes')
      if (local) {
        set({ notes: JSON.parse(local), loading: false })
      } else {
        const seedNotes: Note[] = [
          {
            id: 1,
            user_id: 'local',
            title: 'Refactor UI to Linear / Raycast dark mode theme',
            content: 'Use subtle 1px slate-800 borders, smooth micro-interactions, and snug card padding.',
            tags: ['feature', 'design'],
            color: 'indigo',
            is_pinned: true,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          },
          {
            id: 2,
            user_id: 'local',
            title: 'Explore MCP integration with Cursor & Antigravity',
            content: 'Configure create_task, list_tasks, and log_task_progress tools so AI agents can automatically sync task status.',
            tags: ['mcp', 'explore'],
            color: 'purple',
            is_pinned: false,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          },
        ]
        localStorage.setItem('local_notes', JSON.stringify(seedNotes))
        set({ notes: seedNotes, loading: false })
      }
    }
  },

  createNote: async (payload: CreateNotePayload) => {
    try {
      const res = await noteApi.create(payload)
      const created = res.data
      set((state) => {
        const updated = [created, ...state.notes]
        if (appConfig.enableMocks) {
          localStorage.setItem('local_notes', JSON.stringify(updated))
        }
        return { notes: updated }
      })
      return created
    } catch (err) {
      if (!appConfig.enableMocks) {
        throw err
      }
      const localNote: Note = {
        id: Date.now(),
        user_id: 'local',
        title: payload.title,
        content: payload.content || '',
        tags: payload.tags || [],
        color: payload.color || 'default',
        is_pinned: payload.is_pinned || false,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      }
      set((state) => {
        const updated = [localNote, ...state.notes]
        localStorage.setItem('local_notes', JSON.stringify(updated))
        return { notes: updated }
      })
      return localNote
    }
  },

  updateNote: async (id: number, payload: UpdateNotePayload) => {
    try {
      const res = await noteApi.update(id, payload)
      const updated = res.data
      set((state) => {
        const notes = state.notes.map((n) => (n.id === id ? updated : n))
        if (appConfig.enableMocks) {
          localStorage.setItem('local_notes', JSON.stringify(notes))
        }
        return { notes }
      })
      return updated
    } catch (err) {
      if (!appConfig.enableMocks) {
        throw err
      }
      let updatedNote: Note | null = null
      set((state) => {
        const notes = state.notes.map((n) => {
          if (n.id === id) {
            updatedNote = {
              ...n,
              ...payload,
              updated_at: new Date().toISOString(),
            }
            return updatedNote
          }
          return n
        })
        localStorage.setItem('local_notes', JSON.stringify(notes))
        return { notes }
      })
      if (!updatedNote) throw new Error('Note not found')
      return updatedNote
    }
  },

  deleteNote: async (id: number) => {
    try {
      await noteApi.delete(id)
      set((state) => {
        const notes = state.notes.filter((n) => n.id !== id)
        if (appConfig.enableMocks) {
          localStorage.setItem('local_notes', JSON.stringify(notes))
        }
        return { notes }
      })
    } catch (err) {
      if (!appConfig.enableMocks) {
        throw err
      }
      set((state) => {
        const notes = state.notes.filter((n) => n.id !== id)
        localStorage.setItem('local_notes', JSON.stringify(notes))
        return { notes }
      })
    }
  },

  convertToTask: async (id: number, payload: ConvertToTaskPayload) => {
    try {
      const res = await noteApi.convertToTask(id, payload)
      const { data } = res
      set((state) => {
        const notes = state.notes.map((n) => (n.id === id ? data.note : n))
        if (appConfig.enableMocks) {
          localStorage.setItem('local_notes', JSON.stringify(notes))
        }
        return { notes }
      })
      return data
    } catch (err) {
      if (!appConfig.enableMocks) {
        throw err
      }
      const taskId = Date.now()
      let convertedNote: Note | null = null
      set((state) => {
        const notes = state.notes.map((n) => {
          if (n.id === id) {
            convertedNote = { ...n, converted_task_id: taskId }
            return convertedNote
          }
          return n
        })
        localStorage.setItem('local_notes', JSON.stringify(notes))
        return { notes }
      })
      if (!convertedNote) throw new Error('Note not found')
      return { note: convertedNote, task_id: taskId }
    }
  },
}))
