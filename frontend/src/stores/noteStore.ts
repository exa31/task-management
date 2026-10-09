import { create } from 'zustand'

import { noteApi } from '@/api'
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
      const raw = res.data
      let fetched: Note[] = []
      if (Array.isArray(raw)) {
        fetched = raw
      } else if (raw && typeof raw === 'object') {
        const obj = raw as { data?: Note[]; notes?: Note[]; items?: Note[] }
        if (Array.isArray(obj.data)) fetched = obj.data
        else if (Array.isArray(obj.notes)) fetched = obj.notes
        else if (Array.isArray(obj.items)) fetched = obj.items
      }
      set({ notes: fetched, loading: false })
    } catch (err: unknown) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to load notes'
      set({ error: errorMsg, loading: false })
    }
  },

  createNote: async (payload: CreateNotePayload) => {
    const res = await noteApi.create(payload)
    const created = res.data
    set((state) => ({ notes: [created, ...state.notes] }))
    return created
  },

  updateNote: async (id: number, payload: UpdateNotePayload) => {
    const res = await noteApi.update(id, payload)
    const updated = res.data
    set((state) => ({
      notes: state.notes.map((n) => (n.id === id ? updated : n)),
    }))
    return updated
  },

  deleteNote: async (id: number) => {
    await noteApi.delete(id)
    set((state) => ({
      notes: state.notes.filter((n) => n.id !== id),
    }))
  },

  convertToTask: async (id: number, payload: ConvertToTaskPayload) => {
    const res = await noteApi.convertToTask(id, payload)
    const { data } = res
    set((state) => ({
      notes: state.notes.map((n) => (n.id === id ? data.note : n)),
    }))
    return data
  },
}))
