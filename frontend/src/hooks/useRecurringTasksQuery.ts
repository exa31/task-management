import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { recurringTaskApi } from '@/api/recurringTask'
import { useAppToast } from '@/components'
import { useTranslation } from '@/i18n'
import type {
  CreateRecurringTaskPayload,
  RecurringTask,
  UpdateRecurringTaskPayload
} from '@/types'

import { taskKeys } from './useTasksQuery'

export const recurringTaskKeys = {
  all: ['recurring-tasks'] as const,
  list: () => ['recurring-tasks', 'list'] as const,
  detail: (id: number) => ['recurring-tasks', 'detail', id] as const
}

export const useRecurringTasks = () =>
  useQuery({
    queryKey: recurringTaskKeys.list(),
    queryFn: async (): Promise<RecurringTask[]> => {
      const res = await recurringTaskApi.getAll()
      const raw = res.data
      if (Array.isArray(raw)) return raw
      if (raw && typeof raw === 'object') {
        const obj = raw as { data?: RecurringTask[]; items?: RecurringTask[] }
        if (Array.isArray(obj.data)) return obj.data
        if (Array.isArray(obj.items)) return obj.items
      }
      return []
    }
  })

export const useCreateRecurringTask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async (payload: CreateRecurringTaskPayload) => {
      const res = await recurringTaskApi.create(payload)
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: recurringTaskKeys.all })
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      toast.success(t('recurring.createdSuccess', { defaultValue: 'Template tugas berulang berhasil dibuat' }))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useUpdateRecurringTask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async ({ id, payload }: { id: number; payload: UpdateRecurringTaskPayload }) => {
      const res = await recurringTaskApi.update(id, payload)
      return res.data
    },
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: recurringTaskKeys.all })
      queryClient.invalidateQueries({ queryKey: recurringTaskKeys.detail(vars.id) })
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      toast.success(t('recurring.updatedSuccess', { defaultValue: 'Template tugas berulang berhasil diperbarui' }))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useDeleteRecurringTask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async (id: number) => {
      await recurringTaskApi.delete(id)
      return id
    },
    onSuccess: (deletedId) => {
      queryClient.invalidateQueries({ queryKey: recurringTaskKeys.all })
      queryClient.removeQueries({ queryKey: recurringTaskKeys.detail(deletedId) })
      toast.success(t('recurring.deletedSuccess', { defaultValue: 'Template tugas berulang berhasil dihapus' }))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useToggleActiveRecurringTask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async (id: number) => {
      const res = await recurringTaskApi.toggleActive(id)
      return res.data
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: recurringTaskKeys.all })
      if (data && data.is_active) {
        toast.success(t('recurring.resumedSuccess', { defaultValue: 'Jadwal tugas berulang diaktifkan' }))
      } else {
        toast.info(t('recurring.pausedSuccess', { defaultValue: 'Jadwal tugas berulang dijeda' }))
      }
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}
