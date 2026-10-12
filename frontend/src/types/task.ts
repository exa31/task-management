export type TaskStatus = 'backlog' | 'todo' | 'in_progress' | 'review' | 'done'
export type TaskPriority = 'low' | 'medium' | 'high' | 'urgent'

export const STATUS_COLORS: Record<TaskStatus, { badge: string; dot: string }> = {
  backlog: {
    badge: 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300',
    dot: 'bg-slate-400'
  },
  todo: {
    badge: 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300',
    dot: 'bg-blue-500'
  },
  in_progress: {
    badge: 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300',
    dot: 'bg-amber-500'
  },
  review: {
    badge: 'bg-purple-50 text-purple-700 dark:bg-purple-950/40 dark:text-purple-300',
    dot: 'bg-purple-500'
  },
  done: {
    badge: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300',
    dot: 'bg-emerald-500'
  }
}

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
  due_date_warning_sent_at?: string | null
  is_focus: boolean
  tags: string[]
  project_id?: number | null
  project_name?: string | null
  project_color?: string | null
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
  project_id?: number | null
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
  project_id?: number | null
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
  project_id?: number | null
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

export interface TaskActivity {
  id: number
  task_id: number
  user_id: string
  activity_type: string
  description: string
  changes?: Record<string, unknown>
  created_at: string
}

export interface CreateTaskActivityPayload {
  description: string
  activity_type?: string
}
