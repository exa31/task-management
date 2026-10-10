import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { projectApi } from '@/api'
import { useAppToast } from '@/components'
import type { CreateProjectPayload, Project, UpdateProjectPayload } from '@/types'

import { taskKeys } from './useTasksQuery'

export const projectKeys = {
  all: ['projects'] as const,
  list: () => ['projects', 'list'] as const,
  detail: (id: number) => ['projects', 'detail', id] as const,
}

export const useProjects = () =>
  useQuery({
    queryKey: projectKeys.list(),
    queryFn: async (): Promise<Project[]> => {
      const res = await projectApi.getProjects()
      const raw = res.data
      if (Array.isArray(raw)) return raw
      if (raw && typeof raw === 'object') {
        const obj = raw as { data?: Project[]; projects?: Project[]; items?: Project[] }
        if (Array.isArray(obj.data)) return obj.data
        if (Array.isArray(obj.projects)) return obj.projects
        if (Array.isArray(obj.items)) return obj.items
      }
      return []
    },
  })

export const useCreateProject = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()

  return useMutation({
    mutationFn: async (payload: CreateProjectPayload) => {
      const res = await projectApi.createProject(payload)
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: projectKeys.all })
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      toast.success('Project berhasil dibuat')
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}

export const useUpdateProject = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()

  return useMutation({
    mutationFn: async ({ id, payload }: { id: number; payload: UpdateProjectPayload }) => {
      const res = await projectApi.updateProject(id, payload)
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: projectKeys.all })
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      toast.success('Project berhasil diperbarui')
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}

export const useDeleteProject = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()

  return useMutation({
    mutationFn: async (id: number) => {
      await projectApi.deleteProject(id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: projectKeys.all })
      queryClient.invalidateQueries({ queryKey: taskKeys.all })
      toast.success('Project berhasil dihapus')
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}
