import type {
  CreateTaskActivityPayload,
  CreateTaskPayload,
  LogProgressPayload,
  Subtask,
  Task,
  TaskActivity,
  TaskCursorResponse,
  TaskFilterParams,
  UpdateTaskPayload
} from '@/types'

import { apiClient } from './client'

export const taskApi = {
  getAll: (params?: TaskFilterParams) =>
    apiClient.get<Task[] | TaskCursorResponse>('/api/v1/tasks', { params }),

  getById: (id: number) => apiClient.get<Task>(`/api/v1/tasks/${id}`),

  create: (payload: CreateTaskPayload) => apiClient.post<Task>('/api/v1/tasks', payload),

  update: (id: number, payload: UpdateTaskPayload) =>
    apiClient.put<Task>(`/api/v1/tasks/${id}`, payload),

  delete: (id: number) => apiClient.delete(`/api/v1/tasks/${id}`),

  logProgress: (id: number, payload: LogProgressPayload) =>
    apiClient.post<Task>(`/api/v1/tasks/${id}/progress`, payload),

  toggleFocus: (id: number) => apiClient.post<Task>(`/api/v1/tasks/${id}/focus`),

  createSubtask: (
    taskId: number,
    payload: { title: string; description?: string; due_date?: string | null } | string
  ) => {
    const body = typeof payload === 'string' ? { title: payload } : payload
    return apiClient.post<Subtask>(`/api/v1/tasks/${taskId}/subtasks`, body)
  },

  updateSubtask: (
    taskId: number,
    subtaskId: number,
    payload: { title?: string; description?: string; is_completed?: boolean; due_date?: string | null }
  ) => apiClient.put<Subtask>(`/api/v1/tasks/${taskId}/subtasks/${subtaskId}`, payload),

  deleteSubtask: (taskId: number, subtaskId: number) =>
    apiClient.delete(`/api/v1/tasks/${taskId}/subtasks/${subtaskId}`),

  getActivities: (taskId: number) =>
    apiClient.get<TaskActivity[]>(`/api/v1/tasks/${taskId}/activities`),

  createActivity: (taskId: number, payload: CreateTaskActivityPayload) =>
    apiClient.post<TaskActivity>(`/api/v1/tasks/${taskId}/activities`, payload)
}
