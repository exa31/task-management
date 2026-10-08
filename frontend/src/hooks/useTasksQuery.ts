import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { taskApi } from '@/api'
import { useAppToast } from '@/components'
import { appConfig } from '@/config'
import { useTranslation } from '@/i18n'
import type {
  CreateTaskPayload,
  LogProgressPayload,
  Subtask,
  Task,
  TaskCursorResponse,
  TaskFilterParams,
  UpdateTaskPayload,
} from '@/types'

export const taskKeys = {
  all: ['tasks'] as const,
  list: (params?: TaskFilterParams) => ['tasks', 'list', params || {}] as const,
  infinite: (params?: TaskFilterParams) => ['tasks', 'infinite', params || {}] as const,
  detail: (id: number) => ['tasks', 'detail', id] as const,
}

const cleanParams = (params?: TaskFilterParams): TaskFilterParams | undefined => {
  if (!params) return undefined
  const cleaned: TaskFilterParams = {}
  if (params.status) cleaned.status = params.status
  if (params.priority) cleaned.priority = params.priority
  if (params.tag) cleaned.tag = params.tag
  if (params.is_focus !== undefined) cleaned.is_focus = params.is_focus
  if (params.search) cleaned.search = params.search
  if (params.limit !== undefined) cleaned.limit = params.limit
  if (params.offset !== undefined) cleaned.offset = params.offset
  if (params.cursor !== undefined) cleaned.cursor = params.cursor
  if (params.paginate) cleaned.paginate = params.paginate
  return cleaned
}

const DEFAULT_MOCK_TASKS: Task[] = [
  {
    id: 101,
    user_id: 'local',
    title: 'Integrate REST API with React TanStack Query',
    description: 'Replace legacy manual fetching with React Query hooks for standardized auto-caching and clean invalidations.',
    status: 'in_progress',
    priority: 'urgent',
    is_focus: true,
    tags: ['react', 'tanstack', 'frontend'],
    last_progress_note: 'All query keys standardized and migrated.',
    last_progress_at: new Date().toISOString(),
    subtasks_count: 3,
    subtasks_completed: 2,
    progress_percent: 67,
    subtasks: [
      { id: 1, task_id: 101, title: 'Install @tanstack/react-query', is_completed: true, order_index: 0, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
      { id: 2, task_id: 101, title: 'Create query keys factory', is_completed: true, order_index: 1, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
      { id: 3, task_id: 101, title: 'Migrate useTasks and useNotes hooks', is_completed: false, order_index: 2, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
    ],
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 102,
    user_id: 'local',
    title: 'Design MCP Endpoint Server & Token Management',
    description: 'JSON-RPC 2.0 endpoint for tools / list and call.',
    status: 'todo',
    priority: 'high',
    is_focus: true,
    tags: ['backend', 'mcp', 'go'],
    last_progress_note: 'Database schema mcp_tokens and RPC handlers ready.',
    last_progress_at: new Date().toISOString(),
    subtasks_count: 2,
    subtasks_completed: 0,
    progress_percent: 0,
    subtasks: [
      { id: 4, task_id: 102, title: 'Migrate API tokens database table', is_completed: false, order_index: 0, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
      { id: 5, task_id: 102, title: 'Verify SHA256 hashed secret tokens', is_completed: false, order_index: 1, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
    ],
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
]

export const useTasks = (params?: TaskFilterParams) => {
  const cleaned = cleanParams(params)
  return useQuery({
    queryKey: taskKeys.list(cleaned),
    queryFn: async () => {
      try {
        const res = await taskApi.getAll(cleaned)
        const raw = res.data
        let tasks: Task[] = []
        if (Array.isArray(raw)) {
          tasks = raw
        } else if (raw && 'items' in raw) {
          tasks = (raw as TaskCursorResponse).items || []
        }
        if (appConfig.enableMocks) {
          localStorage.setItem('local_tasks', JSON.stringify(tasks))
        }
        return tasks
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const local = localStorage.getItem('local_tasks')
        if (local) {
          try {
            let tasks: Task[] = JSON.parse(local)
            if (params?.status) {
              tasks = tasks.filter((t) => t.status === params.status)
            }
            if (params?.priority) {
              tasks = tasks.filter((t) => t.priority === params.priority)
            }
            if (params?.is_focus) {
              tasks = tasks.filter((t) => t.is_focus)
            }
            if (params?.tag) {
              tasks = tasks.filter((t) => t.tags?.includes(params.tag!))
            }
            if (params?.search) {
              const q = params.search.toLowerCase()
              tasks = tasks.filter((t) => t.title.toLowerCase().includes(q) || (t.description || '').toLowerCase().includes(q))
            }
            return tasks
          } catch {
            // ignore
          }
        }
        return DEFAULT_MOCK_TASKS
      }
    },
  })
}

export const useInfiniteTasks = (params?: TaskFilterParams) => {
  const cleaned = cleanParams(params)
  return useInfiniteQuery({
    queryKey: taskKeys.infinite(cleaned),
    initialPageParam: undefined as number | undefined,
    queryFn: async ({ pageParam }) => {
      try {
        const queryParams: TaskFilterParams = {
          ...cleaned,
          cursor: pageParam,
          limit: 15,
          paginate: 'cursor',
        }
        const res = await taskApi.getAll(queryParams)
        const raw = res.data
        if (raw && 'items' in raw) {
          return raw as TaskCursorResponse
        }
        const items = Array.isArray(raw) ? raw : []
        const nextId = items.length >= 15 ? items[items.length - 1].id : null
        return {
          items,
          next_cursor: nextId,
          has_more: items.length >= 15,
        }
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const local = localStorage.getItem('local_tasks')
        let all: Task[] = local ? JSON.parse(local) : DEFAULT_MOCK_TASKS
        if (pageParam) {
          all = all.filter((t) => t.id < pageParam)
        }
        const slice = all.slice(0, 15)
        const nextId = slice.length >= 15 ? slice[slice.length - 1].id : null
        return {
          items: slice,
          next_cursor: nextId,
          has_more: slice.length >= 15,
        }
      }
    },
    getNextPageParam: (lastPage) => (lastPage.has_more ? (lastPage.next_cursor ?? undefined) : undefined),
  })
}

export const useTask = (id?: number) => useQuery({
    queryKey: id ? taskKeys.detail(id) : ['tasks', 'detail', 'null'],
    enabled: !!id,
    queryFn: async () => {
      if (!id) return null
      try {
        const res = await taskApi.getById(id)
        return res.data
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const local = localStorage.getItem('local_tasks')
        const list: Task[] = local ? JSON.parse(local) : DEFAULT_MOCK_TASKS
        return list.find((t) => t.id === id) || null
      }
    },
  })

export const useCreateTask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async (payload: CreateTaskPayload) => {
      try {
        const res = await taskApi.create(payload)
        return res.data
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const subtasks: Subtask[] = (payload.subtasks || []).map((title, i) => ({
          id: Date.now() + i,
          task_id: Date.now(),
          title,
          is_completed: false,
          order_index: i,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        }))
        const localTask: Task = {
          id: Date.now(),
          user_id: 'local',
          title: payload.title,
          description: payload.description || '',
          status: payload.status || 'todo',
          priority: payload.priority || 'medium',
          due_date: payload.due_date,
          is_focus: payload.is_focus || false,
          tags: payload.tags || [],
          last_progress_note: '',
          subtasks_count: subtasks.length,
          subtasks_completed: 0,
          progress_percent: 0,
          subtasks,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        }
        const cached = localStorage.getItem('local_tasks')
        const list: Task[] = cached ? JSON.parse(cached) : DEFAULT_MOCK_TASKS
        const updated = [localTask, ...list]
        localStorage.setItem('local_tasks', JSON.stringify(updated))
        return localTask
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      toast.success(t('tasks.createdSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}

export const useUpdateTask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async ({ id, payload }: { id: number; payload: UpdateTaskPayload }) => {
      try {
        const res = await taskApi.update(id, payload)
        return res.data
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const cached = localStorage.getItem('local_tasks')
        const list: Task[] = cached ? JSON.parse(cached) : DEFAULT_MOCK_TASKS
        let targetTask = list.find((t) => t.id === id)
        if (targetTask) {
          targetTask = {
            ...targetTask,
            ...payload,
            updated_at: new Date().toISOString(),
          }
          const updated = list.map((t) => (t.id === id ? targetTask! : t))
          localStorage.setItem('local_tasks', JSON.stringify(updated))
          return targetTask
        }
        throw new Error('Task not found')
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      toast.success(t('tasks.updatedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}

export const useDeleteTask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async (id: number) => {
      try {
        await taskApi.delete(id)
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const cached = localStorage.getItem('local_tasks')
        const list: Task[] = cached ? JSON.parse(cached) : DEFAULT_MOCK_TASKS
        const updated = list.filter((t) => t.id !== id)
        localStorage.setItem('local_tasks', JSON.stringify(updated))
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      toast.success(t('tasks.deletedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}

export const useLogProgress = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async ({ id, payload }: { id: number; payload: LogProgressPayload }) => {
      try {
        const res = await taskApi.logProgress(id, payload)
        return res.data
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const cached = localStorage.getItem('local_tasks')
        const list: Task[] = cached ? JSON.parse(cached) : DEFAULT_MOCK_TASKS
        const targetTask = list.find((t) => t.id === id)
        if (targetTask) {
          const updatedSubtasks = (targetTask.subtasks || []).map((st) =>
            payload.completed_subtask_ids?.includes(st.id) ? { ...st, is_completed: true } : st
          )
          const completedCount = updatedSubtasks.filter((s) => s.is_completed).length
          const totalCount = updatedSubtasks.length
          const updatedTask: Task = {
            ...targetTask,
            status: payload.status || targetTask.status,
            last_progress_note: payload.progress_note,
            last_progress_at: new Date().toISOString(),
            subtasks: updatedSubtasks,
            subtasks_completed: completedCount,
            progress_percent: totalCount > 0 ? Math.round((completedCount * 100) / totalCount) : 0,
            updated_at: new Date().toISOString(),
          }
          const updated = list.map((t) => (t.id === id ? updatedTask : t))
          localStorage.setItem('local_tasks', JSON.stringify(updated))
          return updatedTask
        }
        throw new Error('Task not found')
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      toast.success(t('tasks.progressLoggedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}

export const useToggleFocus = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async (id: number) => {
      try {
        const res = await taskApi.toggleFocus(id)
        return res.data
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const cached = localStorage.getItem('local_tasks')
        const list: Task[] = cached ? JSON.parse(cached) : DEFAULT_MOCK_TASKS
        const target = list.find((t) => t.id === id)
        if (target) {
          const updatedTask = { ...target, is_focus: !target.is_focus }
          const updated = list.map((t) => (t.id === id ? updatedTask : t))
          localStorage.setItem('local_tasks', JSON.stringify(updated))
          return updatedTask
        }
        throw new Error('Task not found')
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      toast.success(t('tasks.focusToggledSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}

export const useAddSubtask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()

  return useMutation({
    mutationFn: async ({ taskId, title }: { taskId: number; title: string }) => {
      try {
        const res = await taskApi.createSubtask(taskId, title)
        return res.data
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const st: Subtask = {
          id: Date.now(),
          task_id: taskId,
          title,
          is_completed: false,
          order_index: 99,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        }
        const cached = localStorage.getItem('local_tasks')
        const list: Task[] = cached ? JSON.parse(cached) : DEFAULT_MOCK_TASKS
        const target = list.find((t) => t.id === taskId)
        if (target) {
          const subtasks = [...(target.subtasks || []), st]
          const updatedTask = {
            ...target,
            subtasks,
            subtasks_count: subtasks.length,
            progress_percent: Math.round(
              (subtasks.filter((s) => s.is_completed).length * 100) / subtasks.length
            ),
          }
          const updated = list.map((t) => (t.id === taskId ? updatedTask : t))
          localStorage.setItem('local_tasks', JSON.stringify(updated))
        }
        return st
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}

export const useToggleSubtask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()

  return useMutation({
    mutationFn: async ({
      taskId,
      subtaskId,
      isCompleted,
    }: {
      taskId: number
      subtaskId: number
      isCompleted: boolean
    }) => {
      try {
        await taskApi.updateSubtask(taskId, subtaskId, { is_completed: isCompleted })
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const cached = localStorage.getItem('local_tasks')
        const list: Task[] = cached ? JSON.parse(cached) : DEFAULT_MOCK_TASKS
        const target = list.find((t) => t.id === taskId)
        if (target) {
          const subtasks = (target.subtasks || []).map((s) =>
            s.id === subtaskId ? { ...s, is_completed: isCompleted } : s
          )
          const updatedTask = {
            ...target,
            subtasks,
            subtasks_completed: subtasks.filter((s) => s.is_completed).length,
            progress_percent:
              subtasks.length > 0
                ? Math.round((subtasks.filter((s) => s.is_completed).length * 100) / subtasks.length)
                : 0,
          }
          const updated = list.map((t) => (t.id === taskId ? updatedTask : t))
          localStorage.setItem('local_tasks', JSON.stringify(updated))
        }
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}

export const useDeleteSubtask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()

  return useMutation({
    mutationFn: async ({ taskId, subtaskId }: { taskId: number; subtaskId: number }) => {
      try {
        await taskApi.deleteSubtask(taskId, subtaskId)
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const cached = localStorage.getItem('local_tasks')
        const list: Task[] = cached ? JSON.parse(cached) : DEFAULT_MOCK_TASKS
        const target = list.find((t) => t.id === taskId)
        if (target) {
          const subtasks = (target.subtasks || []).filter((s) => s.id !== subtaskId)
          const updatedTask = {
            ...target,
            subtasks,
            subtasks_count: subtasks.length,
            subtasks_completed: subtasks.filter((s) => s.is_completed).length,
            progress_percent:
              subtasks.length > 0
                ? Math.round((subtasks.filter((s) => s.is_completed).length * 100) / subtasks.length)
                : 0,
          }
          const updated = list.map((t) => (t.id === taskId ? updatedTask : t))
          localStorage.setItem('local_tasks', JSON.stringify(updated))
        }
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}
