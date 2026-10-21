import type {
  CreateRecurringTaskPayload,
  RecurringTask,
  UpdateRecurringTaskPayload
} from '@/types'

import { apiClient } from './client'

export const recurringTaskApi = {
  getAll: () => apiClient.get<RecurringTask[]>('/api/v1/recurring-tasks'),

  getById: (id: number) => apiClient.get<RecurringTask>(`/api/v1/recurring-tasks/${id}`),

  create: (payload: CreateRecurringTaskPayload) =>
    apiClient.post<RecurringTask>('/api/v1/recurring-tasks', payload),

  update: (id: number, payload: UpdateRecurringTaskPayload) =>
    apiClient.put<RecurringTask>(`/api/v1/recurring-tasks/${id}`, payload),

  delete: (id: number) => apiClient.delete(`/api/v1/recurring-tasks/${id}`),

  toggleActive: (id: number) =>
    apiClient.post<RecurringTask>(`/api/v1/recurring-tasks/${id}/toggle-active`)
}
