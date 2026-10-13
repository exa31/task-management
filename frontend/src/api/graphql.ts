import { axiosInstance } from './client'

export interface GraphQLTask {
  id: number
  title: string
  description?: string
  status: string
  priority: string
  due_date?: string
  is_focus?: boolean
  progress?: number
  tags?: string[]
  project?: {
    id: number
    name: string
    color?: string
  }
}

export interface GraphQLReminder {
  id: number
  title: string
  description?: string
  remind_at: string
  status: string
  task_id?: number
  task_title?: string
  snooze_count?: number
  created_at?: string
}

export interface GraphQLActivity {
  id: number
  task_id: number
  task_title?: string
  activity_type: string
  description: string
  created_at: string
}

export interface CalendarGraphQLResponse {
  tasks: GraphQLTask[]
  reminders: GraphQLReminder[]
  activities: GraphQLActivity[]
}

export const CALENDAR_DATA_QUERY = `
  query GetCalendarData($from: String, $to: String) {
    tasks {
      id
      title
      description
      status
      priority
      due_date
      is_focus
      progress
      tags
      project {
        id
        name
        color
      }
    }
    reminders {
      id
      title
      description
      remind_at
      status
      task_id
      task_title
      snooze_count
      created_at
    }
    activities(from: $from, to: $to) {
      id
      task_id
      task_title
      activity_type
      description
      created_at
    }
  }
`

export async function fetchGraphQL<T = unknown>(
  query: string,
  variables?: Record<string, unknown>,
  signal?: AbortSignal
): Promise<T> {
  const response = await axiosInstance.post<{ data?: T; errors?: Array<{ message: string }> }>(
    '/api/v1/graphql',
    { query, variables },
    { signal }
  )

  if (response.data.errors && response.data.errors.length > 0) {
    const errorMsg = response.data.errors.map((e) => e.message).join(', ')
    throw new Error(errorMsg || 'GraphQL query execution failed')
  }

  return response.data.data as T
}

export async function fetchCalendarData(
  from?: string,
  to?: string,
  signal?: AbortSignal
): Promise<CalendarGraphQLResponse> {
  return fetchGraphQL<CalendarGraphQLResponse>(CALENDAR_DATA_QUERY, { from, to }, signal)
}
