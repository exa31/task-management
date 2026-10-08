export type TaskStatus = 'backlog' | 'todo' | 'in_progress' | 'review' | 'done'
export type TaskPriority = 'low' | 'medium' | 'high' | 'urgent'

export interface Subtask {
  id: number
  task_id: number
  title: string
  is_completed: boolean
  order_index: number
  created_at: string
  updated_at: string
}

export interface Task {
  id: number
  user_id: string
  title: string
  description: string
  status: TaskStatus
  priority: TaskPriority
  due_date?: string | null
  is_focus: boolean
  tags: string[]
  last_progress_note: string
  last_progress_at?: string | null
  source_note_id?: number | null
  subtasks: Subtask[]
  subtasks_count: number
  subtasks_completed: number
  progress_percent: number
  created_at: string
  updated_at: string
}

export interface CreateTaskPayload {
  title: string
  description?: string
  status?: TaskStatus
  priority?: TaskPriority
  due_date?: string
  is_focus?: boolean
  tags?: string[]
  source_note_id?: number
  subtasks?: string[]
}

export interface UpdateTaskPayload {
  title?: string
  description?: string
  status?: TaskStatus
  priority?: TaskPriority
  due_date?: string | null
  is_focus?: boolean
  tags?: string[]
  last_progress_note?: string
}

export interface LogProgressPayload {
  progress_note: string
  status?: TaskStatus
  completed_subtask_ids?: number[]
}

export interface TaskFilterParams {
  status?: string
  priority?: string
  is_focus?: boolean
  tag?: string
  search?: string
  limit?: number
  offset?: number
  cursor?: number
  paginate?: 'cursor'
}

export interface TaskCursorResponse {
  items: Task[]
  next_cursor?: number | null
  has_more: boolean
}

