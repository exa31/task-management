import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { mcpApi } from '@/api'
import { useAppToast } from '@/components'
import { useTranslation } from '@/i18n'
import type { APIToken, GenerateTokenResponse } from '@/types'

export const mcpKeys = {
  all: ['mcp-tokens'] as const,
}

export const useMcpTokens = () =>
  useQuery({
    queryKey: mcpKeys.all,
    queryFn: async (): Promise<APIToken[]> => {
      const res = await mcpApi.listTokens()
      return res.data || []
    },
  })

export const useCreateMcpToken = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async ({ name }: { name: string }): Promise<GenerateTokenResponse> => {
      const res = await mcpApi.generateToken({ name })
      return res.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mcpKeys.all })
      toast.success(t('mcp.tokenCreatedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}

export const useDeleteMcpToken = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async (id: number) => {
      await mcpApi.deleteToken(id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mcpKeys.all })
      toast.success(t('mcp.tokenRevokedSuccess'))
    },
    onError: (err) => {
      toast.error(err)
    },
  })
}
