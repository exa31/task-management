import type { CreateProjectPayload, Project, UpdateProjectPayload } from '@/types'

import { apiClient } from './client'

export const projectApi = {
  getProjects: () =>
    apiClient.get<Project[]>('/api/v1/projects'),

  getProject: (id: number) =>
    apiClient.get<Project>(`/api/v1/projects/${id}`),

  createProject: (payload: CreateProjectPayload) =>
    apiClient.post<Project>('/api/v1/projects', payload),

  updateProject: (id: number, payload: UpdateProjectPayload) =>
    apiClient.put<Project>(`/api/v1/projects/${id}`, payload),

  deleteProject: (id: number) =>
    apiClient.delete(`/api/v1/projects/${id}`),
}
