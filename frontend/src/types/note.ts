export interface Note {
  id: number
  user_id: string
  title: string
  content: string
  tags: string[]
  color: string
  is_pinned: boolean
  converted_task_id?: number | null
  created_at: string
  updated_at: string
}

export interface CreateNotePayload {
  title: string
  content?: string
  tags?: string[]
  color?: string
  is_pinned?: boolean
}

export interface UpdateNotePayload {
  title?: string
  content?: string
  tags?: string[]
  color?: string
  is_pinned?: boolean
  converted_task_id?: number | null
}

export interface ConvertToTaskPayload {
  priority?: 'low' | 'medium' | 'high' | 'urgent'
  due_date?: string
  subtasks?: string[]
}

export interface NoteFilterParams {
  search?: string
  tag?: string
  limit?: number
  offset?: number
}
