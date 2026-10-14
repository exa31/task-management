import { useQuery } from '@tanstack/react-query'

import { workspaceApi, type WorkspaceStats } from '@/api'

export const workspaceKeys = {
  all: ['workspace-stats'] as const
}

export const useWorkspaceStats = () =>
  useQuery<WorkspaceStats>({
    queryKey: workspaceKeys.all,
    queryFn: async () => {
      const res = await workspaceApi.getStats()
      return res.data
    },
    staleTime: 30 * 1000,
    refetchOnWindowFocus: true
  })
