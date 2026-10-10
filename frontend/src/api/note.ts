import type {
  ConvertToTaskPayload,
  CreateNotePayload,
  Note,
  NoteFilterParams,
  UpdateNotePayload
} from '@/types'

import { apiClient } from './client'

export const noteApi = {
  getAll: (params?: NoteFilterParams) => apiClient.get<Note[]>('/api/v1/notes', { params }),

  getById: (id: number) => apiClient.get<Note>(`/api/v1/notes/${id}`),

  create: (payload: CreateNotePayload) => apiClient.post<Note>('/api/v1/notes', payload),

  update: (id: number, payload: UpdateNotePayload) =>
    apiClient.put<Note>(`/api/v1/notes/${id}`, payload),

  delete: (id: number) => apiClient.delete(`/api/v1/notes/${id}`),

  convertToTask: (id: number, payload: ConvertToTaskPayload) =>
    apiClient.post<{ note: Note; task_id: number }>(`/api/v1/notes/${id}/convert`, payload)
}
