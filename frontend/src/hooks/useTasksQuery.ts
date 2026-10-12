import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { taskApi } from '@/api'
import { useAppToast } from '@/components'
import { useTranslation } from '@/i18n'
import type {
  CreateTaskPayload,
  LogProgressPayload,
  Task,
  TaskCursorResponse,
  TaskFilterParams,
  UpdateTaskPayload
} from '@/types'

export const taskKeys = {
  all: ['tasks'] as const,
  list: (params?: TaskFilterParams) => ['tasks', 'list', params || {}] as const,
  infinite: (params?: TaskFilterParams) => ['tasks', 'infinite', params || {}] as const,
  detail: (id: number) => ['tasks', 'detail', id] as const,
  activities: (id: number) => ['tasks', 'activities', id] as const
}

const cleanParams = (params?: TaskFilterParams): TaskFilterParams | undefined => {
  if (!params) return undefined
  const cleaned: TaskFilterParams = {}
  if (params.status) cleaned.status = params.status
  if (params.priority) cleaned.priority = params.priority
  if (params.tag) cleaned.tag = params.tag
  if (params.project_id !== undefined && params.project_id !== null)
    cleaned.project_id = params.project_id
  if (params.is_focus !== undefined) cleaned.is_focus = params.is_focus
  if (params.search) cleaned.search = params.search
  if (params.limit !== undefined) cleaned.limit = params.limit
  if (params.offset !== undefined) cleaned.offset = params.offset
  if (params.cursor !== undefined) cleaned.cursor = params.cursor
  if (params.paginate) cleaned.paginate = params.paginate
  return cleaned
}

export const useTasks = (params?: TaskFilterParams) => {
  const cleaned = cleanParams(params)
  return useQuery({
    queryKey: taskKeys.list(cleaned),
    queryFn: async (): Promise<Task[]> => {
      const res = await taskApi.getAll(cleaned)
      const raw = res.data
      if (Array.isArray(raw)) return raw
      if (raw && typeof raw === 'object') {
        const obj = raw as { items?: Task[]; data?: Task[]; tasks?: Task[] }
        if (Array.isArray(obj.items)) return obj.items
        if (Array.isArray(obj.data)) return obj.data
        if (Array.isArray(obj.tasks)) return obj.tasks
      }
      return []
    }
  })
}

export const useInfiniteTasks = (params?: TaskFilterParams) => {
  const cleaned = cleanParams(params)
  return useInfiniteQuery({
    queryKey: taskKeys.infinite(cleaned),
    initialPageParam: undefined as number | undefined,
    queryFn: async ({ pageParam }): Promise<TaskCursorResponse> => {
      const queryParams: TaskFilterParams = {
        ...cleaned,
        cursor: pageParam,
        limit: 15,
        paginate: 'cursor'
      }
      const res = await taskApi.getAll(queryParams)
      const raw = res.data
      if (raw && typeof raw === 'object' && 'items' in raw) {
        const cursorResp = raw as TaskCursorResponse
        return {
          items: Array.isArray(cursorResp.items) ? cursorResp.items : [],
          next_cursor: cursorResp.next_cursor ?? null,
          has_more: Boolean(cursorResp.has_more)
        }
      }
      const items = Array.isArray(raw) ? raw : []
      const nextId = items.length >= 15 ? items[items.length - 1].id : null
      return {
        items,
        next_cursor: nextId,
        has_more: items.length >= 15
      }
    },
    getNextPageParam: (lastPage) =>
      lastPage.has_more ? (lastPage.next_cursor ?? undefined) : undefined
  })
}

export const useTask = (id?: number) =>
  useQuery({
    queryKey: id ? taskKeys.detail(id) : ['tasks', 'detail', 'null'],
    enabled: !!id,
    queryFn: async () => {
      if (!id) return null
      const res = await taskApi.getById(id)
      return res.data
    }
  })

export const useCreateTask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async (payload: CreateTaskPayload) => {
      const res = await taskApi.create(payload)
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      toast.success(t('tasks.createdSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useUpdateTask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async ({ id, payload }: { id: number; payload: UpdateTaskPayload }) => {
      const res = await taskApi.update(id, payload)
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      toast.success(t('tasks.updatedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useDeleteTask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async (id: number) => {
      await taskApi.delete(id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      toast.success(t('tasks.deletedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useLogProgress = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async ({ id, payload }: { id: number; payload: LogProgressPayload }) => {
      const res = await taskApi.logProgress(id, payload)
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      toast.success(t('tasks.progressLoggedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useToggleFocus = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async (id: number) => {
      const res = await taskApi.toggleFocus(id)
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      toast.success(t('tasks.focusToggledSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useAddSubtask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()

  return useMutation({
    mutationFn: async ({ taskId, title }: { taskId: number; title: string }) => {
      const res = await taskApi.createSubtask(taskId, title)
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useToggleSubtask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()

  return useMutation({
    mutationFn: async ({
      taskId,
      subtaskId,
      isCompleted
    }: {
      taskId: number
      subtaskId: number
      isCompleted: boolean
    }) => {
      await taskApi.updateSubtask(taskId, subtaskId, { is_completed: isCompleted })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useDeleteSubtask = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()

  return useMutation({
    mutationFn: async ({ taskId, subtaskId }: { taskId: number; subtaskId: number }) => {
      await taskApi.deleteSubtask(taskId, subtaskId)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}

export const useTaskActivities = (taskId?: number) =>
  useQuery({
    queryKey: taskId ? taskKeys.activities(taskId) : ['tasks', 'activities', 'null'],
    enabled: !!taskId,
    queryFn: async () => {
      if (!taskId) return []
      const res = await taskApi.getActivities(taskId)
      const { data } = res
      return Array.isArray(data) ? data : []
    }
  })

export const useCreateTaskActivity = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()

  return useMutation({
    mutationFn: async ({ taskId, description }: { taskId: number; description: string }) => {
      const res = await taskApi.createActivity(taskId, { description })
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
    },
    onError: (err) => {
      toast.error(err)
    }
  })
}
