export interface Project {
  id: number
  user_id: string
  name: string
  color: string
  description?: string
  task_count?: number
  created_at: string
  updated_at: string
}

export interface CreateProjectPayload {
  name: string
  color?: string
  description?: string
}

export interface UpdateProjectPayload {
  name?: string
  color?: string
  description?: string
}
