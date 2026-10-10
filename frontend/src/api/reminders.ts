import type {
  CreateReminderPayload,
  RegisterDeviceTokenPayload,
  Reminder,
  ReminderFilterParams,
  ReminderPaginatedResponse,
  SnoozeReminderPayload,
  UpdateReminderPayload
} from '@/types'

import { apiClient } from './client'

export const reminderApi = {
  getAll: (params?: ReminderFilterParams) =>
    apiClient.get<ReminderPaginatedResponse>('/api/v1/reminders', { params }),

  getById: (id: number) => apiClient.get<Reminder>(`/api/v1/reminders/${id}`),

  create: (payload: CreateReminderPayload) =>
    apiClient.post<Reminder>('/api/v1/reminders', payload),

  update: (id: number, payload: UpdateReminderPayload) =>
    apiClient.put<Reminder>(`/api/v1/reminders/${id}`, payload),

  delete: (id: number) => apiClient.delete(`/api/v1/reminders/${id}`),

  snooze: (id: number, payload: SnoozeReminderPayload) =>
    apiClient.post<Reminder>(`/api/v1/reminders/${id}/snooze`, payload),

  dismiss: (id: number) => apiClient.post<Reminder>(`/api/v1/reminders/${id}/dismiss`),

  registerDeviceToken: (payload: RegisterDeviceTokenPayload) =>
    apiClient.post('/api/v1/reminders/device-token', payload)
}
