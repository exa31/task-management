export interface Reminder {
  id: number
  user_id: string
  task_id?: number | null
  task_title?: string | null
  title: string
  description: string
  remind_at: string
  status: 'pending' | 'sent' | 'dismissed'
  snooze_count: number
  sent_at?: string | null
  created_at: string
  updated_at: string
}

export interface CreateReminderPayload {
  task_id?: number | null
  title: string
  description?: string
  remind_at: string
}

export interface UpdateReminderPayload {
  task_id?: number | null
  title?: string
  description?: string
  remind_at?: string
  status?: 'pending' | 'sent' | 'dismissed'
}

export interface SnoozeReminderPayload {
  minutes: number
}

export interface RegisterDeviceTokenPayload {
  token: string
  platform?: 'web' | 'android' | 'ios'
}

export interface ReminderFilterParams {
  task_id?: number
  status?: 'pending' | 'sent' | 'dismissed' | 'all'
  limit?: number
  offset?: number
}

export interface ReminderPaginatedResponse {
  items: Reminder[]
  total: number
  page: number
  perPage: number
  totalPages: number
}
