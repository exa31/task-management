import { apiClient } from './client'

export interface WorkspaceStats {
  tasks_in_progress: number
  tasks_todo: number
  tasks_total: number
  total_projects: number
  pending_reminders: number
  total_notes: number
}

export const workspaceApi = {
  getStats: () => apiClient.get<WorkspaceStats>('/api/v1/workspace/stats')
}
