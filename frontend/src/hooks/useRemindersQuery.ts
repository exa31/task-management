import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { reminderApi } from '@/api'
import { useAppToast } from '@/components'
import { useTranslation } from '@/i18n'
import type {
  CreateReminderPayload,
  Reminder,
  ReminderFilterParams,
  ReminderPaginatedResponse,
  SnoozeReminderPayload,
  UpdateReminderPayload
} from '@/types'

export const reminderKeys = {
  all: ['reminders'] as const,
  list: (params?: ReminderFilterParams) => ['reminders', 'list', params || {}] as const,
  detail: (id: number) => ['reminders', 'detail', id] as const
}

export const useReminders = (params?: ReminderFilterParams) =>
  useQuery({
    queryKey: reminderKeys.list(params),
    queryFn: async (): Promise<ReminderPaginatedResponse> => {
      const res = await reminderApi.getAll(params)
      const { data } = res
      if (data && 'items' in data) {
        return data as ReminderPaginatedResponse
      }
      return {
        items: Array.isArray(data) ? (data as unknown as Reminder[]) : [],
        total: Array.isArray(data) ? (data as unknown as Reminder[]).length : 0,
        page: 1,
        perPage: 50,
        totalPages: 1
      }
    }
  })

export const useReminder = (id: number) =>
  useQuery({
    queryKey: reminderKeys.detail(id),
    queryFn: async (): Promise<Reminder> => {
      const res = await reminderApi.getById(id)
      return res.data
    },
    enabled: !!id
  })

export const useCreateReminder = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: (payload: CreateReminderPayload) => reminderApi.create(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: reminderKeys.all })
      toast.success(t('reminders.createdSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useUpdateReminder = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: UpdateReminderPayload }) =>
      reminderApi.update(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: reminderKeys.all })
      toast.success(t('reminders.updatedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useDeleteReminder = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: (id: number) => reminderApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: reminderKeys.all })
      toast.success(t('reminders.deletedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useSnoozeReminder = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: SnoozeReminderPayload }) =>
      reminderApi.snooze(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: reminderKeys.all })
      toast.success(t('reminders.snoozedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useDismissReminder = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: (id: number) => reminderApi.dismiss(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: reminderKeys.all })
      toast.success(t('reminders.dismissedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}
