import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { noteApi } from '@/api'
import { useAppToast } from '@/components'
import { appConfig } from '@/config'
import { useTranslation } from '@/i18n'
import type { ConvertToTaskPayload, CreateNotePayload, Note, NoteFilterParams, UpdateNotePayload } from '@/types'

import { taskKeys } from './useTasksQuery'

export const noteKeys = {
  all: ['notes'] as const,
  list: (params?: NoteFilterParams) => ['notes', 'list', params || {}] as const,
  detail: (id: number) => ['notes', 'detail', id] as const,
}

const DEFAULT_MOCK_NOTES: Note[] = [
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

export const useNotes = (params?: NoteFilterParams) => useQuery({
  queryKey: noteKeys.list(params),
  queryFn: async () => {
    try {
      const res = await noteApi.getAll(params)
      const notes = res.data || []
      if (appConfig.enableMocks) {
        localStorage.setItem('local_notes', JSON.stringify(notes))
      }
      return notes
    } catch (err) {
      if (!appConfig.enableMocks) {
        throw err
      }
      const local = localStorage.getItem('local_notes')
      if (local) {
        try {
          const parsed: Note[] = JSON.parse(local)
          let filtered = parsed
          if (params?.tag) {
            filtered = filtered.filter((n) => n.tags?.includes(params.tag!))
          }
          if (params?.search) {
            const q = params.search.toLowerCase()
            filtered = filtered.filter((n) => n.title.toLowerCase().includes(q) || n.content.toLowerCase().includes(q))
          }
          return filtered
        } catch {
          // ignore
        }
      }
      return DEFAULT_MOCK_NOTES
    }
  },
})

export const useCreateNote = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async (payload: CreateNotePayload) => {
      try {
        const res = await noteApi.create(payload)
        return res.data
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
        const cached = localStorage.getItem('local_notes')
        const list: Note[] = cached ? JSON.parse(cached) : DEFAULT_MOCK_NOTES
        const updated = [localNote, ...list]
        localStorage.setItem('local_notes', JSON.stringify(updated))
        return localNote
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: noteKeys.all })
      toast.success(t('notes.createdSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}

export const useUpdateNote = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async ({ id, payload }: { id: number; payload: UpdateNotePayload }) => {
      try {
        const res = await noteApi.update(id, payload)
        return res.data
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const cached = localStorage.getItem('local_notes')
        const list: Note[] = cached ? JSON.parse(cached) : DEFAULT_MOCK_NOTES
        let targetNote = list.find((n) => n.id === id)
        if (targetNote) {
          targetNote = {
            ...targetNote,
            title: payload.title ?? targetNote.title,
            content: payload.content ?? targetNote.content,
            tags: payload.tags ?? targetNote.tags,
            color: payload.color ?? targetNote.color,
            is_pinned: payload.is_pinned ?? targetNote.is_pinned,
            updated_at: new Date().toISOString(),
          }
          const updated = list.map((n) => (n.id === id ? targetNote! : n))
          localStorage.setItem('local_notes', JSON.stringify(updated))
          return targetNote
        }
        throw new Error('Note not found')
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: noteKeys.all })
      toast.success(t('notes.updatedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}

export const useDeleteNote = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async (id: number) => {
      try {
        await noteApi.delete(id)
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const cached = localStorage.getItem('local_notes')
        const list: Note[] = cached ? JSON.parse(cached) : DEFAULT_MOCK_NOTES
        const updated = list.filter((n) => n.id !== id)
        localStorage.setItem('local_notes', JSON.stringify(updated))
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: noteKeys.all })
      toast.success(t('notes.deletedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}

export const useConvertToTask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async ({ id, payload }: { id: number; payload: ConvertToTaskPayload }) => {
      try {
        const res = await noteApi.convertToTask(id, payload)
        return res.data
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const cached = localStorage.getItem('local_notes')
        const list: Note[] = cached ? JSON.parse(cached) : DEFAULT_MOCK_NOTES
        const updated = list.map((n) => (n.id === id ? { ...n, converted_task_id: Date.now() } : n))
        localStorage.setItem('local_notes', JSON.stringify(updated))
        return {
          note: updated.find((n) => n.id === id)!,
          task_id: Date.now(),
        }
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: noteKeys.all })
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      toast.success(t('notes.convertedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}
