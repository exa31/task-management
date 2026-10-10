import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { noteApi } from '@/api'
import { useAppToast } from '@/components'
import { useTranslation } from '@/i18n'
import type {
  ConvertToTaskPayload,
  CreateNotePayload,
  Note,
  NoteFilterParams,
  UpdateNotePayload
} from '@/types'

import { taskKeys } from './useTasksQuery'

export const noteKeys = {
  all: ['notes'] as const,
  list: (params?: NoteFilterParams) => ['notes', 'list', params || {}] as const,
  detail: (id: number) => ['notes', 'detail', id] as const
}

export const useNotes = (params?: NoteFilterParams) =>
  useQuery({
    queryKey: noteKeys.list(params),
    queryFn: async (): Promise<Note[]> => {
      const res = await noteApi.getAll(params)
      const raw = res.data
      if (Array.isArray(raw)) return raw
      if (raw && typeof raw === 'object') {
        const obj = raw as { data?: Note[]; notes?: Note[]; items?: Note[] }
        if (Array.isArray(obj.data)) return obj.data
        if (Array.isArray(obj.notes)) return obj.notes
        if (Array.isArray(obj.items)) return obj.items
      }
      return []
    }
  })

export const useCreateNote = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async (payload: CreateNotePayload) => {
      const res = await noteApi.create(payload)
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: noteKeys.all })
      toast.success(t('notes.createdSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useUpdateNote = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async ({ id, payload }: { id: number; payload: UpdateNotePayload }) => {
      const res = await noteApi.update(id, payload)
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: noteKeys.all })
      toast.success(t('notes.updatedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useDeleteNote = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async (id: number) => {
      await noteApi.delete(id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: noteKeys.all })
      toast.success(t('notes.deletedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useConvertToTask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async ({ id, payload }: { id: number; payload: ConvertToTaskPayload }) => {
      const res = await noteApi.convertToTask(id, payload)
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: noteKeys.all })
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      toast.success(t('notes.convertedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}
