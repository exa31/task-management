import type { InfiniteData, QueryClient } from '@tanstack/react-query'
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { taskApi } from '@/api'
import { useAppToast } from '@/components'
import { useTranslation } from '@/i18n'
import { useTaskStore } from '@/stores'
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

/**
 * Optimistically and directly update a task across all query caches and the Zustand store.
 */
export const updateTaskInQueryCache = (queryClient: QueryClient, updatedTask: Task) => {
  // 1. Update task detail query
  queryClient.setQueryData<Task>(taskKeys.detail(updatedTask.id), (old) => {
    if (!old) return updatedTask
    return { ...old, ...updatedTask }
  })

  // 2. Update all infinite queries (e.g. TaskListView)
  queryClient.setQueriesData<InfiniteData<TaskCursorResponse>>(
    { queryKey: ['tasks', 'infinite'] },
    (old) => {
      if (!old || !Array.isArray(old.pages)) return old
      return {
        ...old,
        pages: old.pages.map((page) => ({
          ...page,
          items: Array.isArray(page.items)
            ? page.items.map((item) =>
                item.id === updatedTask.id ? { ...item, ...updatedTask } : item
              )
            : []
        }))
      }
    }
  )

  // 3. Update all regular list queries (TaskBoardView, Dashboard, Today's Focus Horizon, etc.)
  queryClient.setQueriesData<Task[]>(
    { queryKey: ['tasks', 'list'] },
    (old) => {
      if (!Array.isArray(old)) return old
      return old.map((item) =>
        item.id === updatedTask.id ? { ...item, ...updatedTask } : item
      )
    }
  )

  // 4. Update Zustand taskStore selectedTask if active
  const store = useTaskStore.getState()
  if (store.selectedTask?.id === updatedTask.id) {
    store.setSelectedTask({ ...store.selectedTask, ...updatedTask })
  }
}

/**
 * Directly remove a task from all query caches and close drawer if matching.
 */
export const removeTaskFromQueryCache = (queryClient: QueryClient, taskId: number) => {
  // 1. Remove task detail query
  queryClient.removeQueries({ queryKey: taskKeys.detail(taskId) })

  // 2. Remove from all infinite queries
  queryClient.setQueriesData<InfiniteData<TaskCursorResponse>>(
    { queryKey: ['tasks', 'infinite'] },
    (old) => {
      if (!old || !Array.isArray(old.pages)) return old
      return {
        ...old,
        pages: old.pages.map((page) => ({
          ...page,
          items: Array.isArray(page.items)
            ? page.items.filter((item) => item.id !== taskId)
            : []
        }))
      }
    }
  )

  // 3. Remove from regular list queries
  queryClient.setQueriesData<Task[]>(
    { queryKey: ['tasks', 'list'] },
    (old) => {
      if (!Array.isArray(old)) return old
      return old.filter((item) => item.id !== taskId)
    }
  )

  // 4. Close drawer or clear selectedTask in store
  const store = useTaskStore.getState()
  if (store.selectedTask?.id === taskId) {
    store.closeDrawer()
  }
}

/**
 * Directly prepend a newly created task into infinite and list query caches.
 */
export const addTaskToQueryCache = (queryClient: QueryClient, newTask: Task) => {
  // 1. Set task detail query
  queryClient.setQueryData<Task>(taskKeys.detail(newTask.id), newTask)

  // 2. Prepend to first page of infinite queries
  queryClient.setQueriesData<InfiniteData<TaskCursorResponse>>(
    { queryKey: ['tasks', 'infinite'] },
    (old) => {
      if (!old || !Array.isArray(old.pages) || old.pages.length === 0) return old
      const firstPage = old.pages[0]
      const existingItems = Array.isArray(firstPage.items) ? firstPage.items : []
      if (existingItems.some((item) => item.id === newTask.id)) return old
      return {
        ...old,
        pages: [
          {
            ...firstPage,
            items: [newTask, ...existingItems]
          },
          ...old.pages.slice(1)
        ]
      }
    }
  )

  // 3. Prepend to regular list queries
  queryClient.setQueriesData<Task[]>(
    { queryKey: ['tasks', 'list'] },
    (old) => {
      if (!Array.isArray(old)) return old
      if (old.some((item) => item.id === newTask.id)) return old
      return [newTask, ...old]
    }
  )
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

export const useTasks = (params?: TaskFilterParams, options?: { enabled?: boolean }) => {
  const cleaned = cleanParams(params)
  return useQuery({
    queryKey: taskKeys.list(cleaned),
    enabled: options?.enabled,
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
    onSuccess: (newTask) => {
      if (newTask) {
        addTaskToQueryCache(queryClient, newTask)
      }
      queryClient.invalidateQueries({ queryKey: taskKeys.all, refetchType: 'all' })
      queryClient.invalidateQueries({ queryKey: ['projects'], refetchType: 'all' })
      queryClient.invalidateQueries({ queryKey: ['workspace-stats'] })
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
    onSuccess: (updatedTask) => {
      if (updatedTask) {
        updateTaskInQueryCache(queryClient, updatedTask)
      }
      queryClient.invalidateQueries({ queryKey: taskKeys.all, refetchType: 'all' })
      queryClient.invalidateQueries({ queryKey: ['projects'], refetchType: 'all' })
      queryClient.invalidateQueries({ queryKey: ['workspace-stats'] })
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
      return id
    },
    onSuccess: (id) => {
      removeTaskFromQueryCache(queryClient, id)
      queryClient.invalidateQueries({ queryKey: taskKeys.all, refetchType: 'all' })
      queryClient.invalidateQueries({ queryKey: ['projects'], refetchType: 'all' })
      queryClient.invalidateQueries({ queryKey: ['workspace-stats'] })
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
    onSuccess: (updatedTask) => {
      if (updatedTask) {
        updateTaskInQueryCache(queryClient, updatedTask)
      }
      queryClient.invalidateQueries({ queryKey: taskKeys.all, refetchType: 'all' })
      queryClient.invalidateQueries({ queryKey: ['workspace-stats'] })
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
    onSuccess: (updatedTask) => {
      if (updatedTask) {
        updateTaskInQueryCache(queryClient, updatedTask)
      }
      queryClient.invalidateQueries({ queryKey: taskKeys.all, refetchType: 'all' })
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
      return { taskId, subtask: res.data }
    },
    onSuccess: ({ taskId, subtask }) => {
      // 1. Update detail cache
      queryClient.setQueryData<Task>(taskKeys.detail(taskId), (old) => {
        if (!old) return old
        const subtasks = [...(old.subtasks || []), subtask]
        const total = subtasks.length
        const completed = subtasks.filter((s) => s.is_completed).length
        const percent = total > 0 ? Math.round((completed / total) * 100) : 0
        return {
          ...old,
          subtasks,
          subtasks_count: total,
          subtasks_completed: completed,
          progress_percent: percent
        }
      })
      // 2. Update infinite query counts
      queryClient.setQueriesData<InfiniteData<TaskCursorResponse>>(
        { queryKey: ['tasks', 'infinite'] },
        (old) => {
          if (!old || !Array.isArray(old.pages)) return old
          return {
            ...old,
            pages: old.pages.map((page) => ({
              ...page,
              items: (page.items || []).map((item) => {
                if (item.id !== taskId) return item
                const total = item.subtasks_count + 1
                const percent = Math.round((item.subtasks_completed / total) * 100)
                return {
                  ...item,
                  subtasks_count: total,
                  progress_percent: percent
                }
              })
            }))
          }
        }
      )
      // 3. Update list queries
      queryClient.setQueriesData<Task[]>(
        { queryKey: ['tasks', 'list'] },
        (old) => {
          if (!Array.isArray(old)) return old
          return old.map((item) => {
            if (item.id !== taskId) return item
            const total = item.subtasks_count + 1
            const percent = Math.round((item.subtasks_completed / total) * 100)
            return {
              ...item,
              subtasks_count: total,
              progress_percent: percent
            }
          })
        }
      )
      queryClient.invalidateQueries({ queryKey: taskKeys.detail(taskId), refetchType: 'all' })
      queryClient.invalidateQueries({ queryKey: taskKeys.all, refetchType: 'all' })
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
      return { taskId, subtaskId, isCompleted }
    },
    onMutate: async ({ taskId, subtaskId, isCompleted }) => {
      // Optimistically update subtask in task detail
      queryClient.setQueryData<Task>(taskKeys.detail(taskId), (old) => {
        if (!old) return old
        const updatedSubtasks = (old.subtasks || []).map((st) =>
          st.id === subtaskId ? { ...st, is_completed: isCompleted } : st
        )
        const completedCount = updatedSubtasks.filter((st) => st.is_completed).length
        const totalCount = updatedSubtasks.length
        const percent = totalCount > 0 ? Math.round((completedCount / totalCount) * 100) : 0
        return {
          ...old,
          subtasks: updatedSubtasks,
          subtasks_completed: completedCount,
          subtasks_count: totalCount,
          progress_percent: percent
        }
      })
      // Optimistically update counts in infinite queries
      queryClient.setQueriesData<InfiniteData<TaskCursorResponse>>(
        { queryKey: ['tasks', 'infinite'] },
        (old) => {
          if (!old || !Array.isArray(old.pages)) return old
          return {
            ...old,
            pages: old.pages.map((page) => ({
              ...page,
              items: (page.items || []).map((item) => {
                if (item.id !== taskId) return item
                const completedDelta = isCompleted ? 1 : -1
                const newCompleted = Math.max(
                  0,
                  Math.min(item.subtasks_count, item.subtasks_completed + completedDelta)
                )
                const newPercent =
                  item.subtasks_count > 0
                    ? Math.round((newCompleted / item.subtasks_count) * 100)
                    : 0
                return {
                  ...item,
                  subtasks_completed: newCompleted,
                  progress_percent: newPercent
                }
              })
            }))
          }
        }
      )
      // Optimistically update counts in regular lists
      queryClient.setQueriesData<Task[]>(
        { queryKey: ['tasks', 'list'] },
        (old) => {
          if (!Array.isArray(old)) return old
          return old.map((item) => {
            if (item.id !== taskId) return item
            const completedDelta = isCompleted ? 1 : -1
            const newCompleted = Math.max(
              0,
              Math.min(item.subtasks_count, item.subtasks_completed + completedDelta)
            )
            const newPercent =
              item.subtasks_count > 0
                ? Math.round((newCompleted / item.subtasks_count) * 100)
                : 0
            return {
              ...item,
              subtasks_completed: newCompleted,
              progress_percent: newPercent
            }
          })
        }
      )
    },
    onSuccess: ({ taskId }) => {
      queryClient.invalidateQueries({ queryKey: taskKeys.detail(taskId), refetchType: 'all' })
      queryClient.invalidateQueries({ queryKey: taskKeys.all, refetchType: 'all' })
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
      return { taskId, subtaskId }
    },
    onSuccess: ({ taskId, subtaskId }) => {
      // 1. Update detail query cache
      queryClient.setQueryData<Task>(taskKeys.detail(taskId), (old) => {
        if (!old) return old
        const remaining = (old.subtasks || []).filter((s) => s.id !== subtaskId)
        const total = remaining.length
        const completed = remaining.filter((s) => s.is_completed).length
        const percent = total > 0 ? Math.round((completed / total) * 100) : 0
        return {
          ...old,
          subtasks: remaining,
          subtasks_count: total,
          subtasks_completed: completed,
          progress_percent: percent
        }
      })
      queryClient.invalidateQueries({ queryKey: taskKeys.detail(taskId), refetchType: 'all' })
      queryClient.invalidateQueries({ queryKey: taskKeys.all, refetchType: 'all' })
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
