import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { mcpApi } from '@/api'
import { useAppToast } from '@/components'
import { appConfig } from '@/config'
import { useTranslation } from '@/i18n'
import type { APIToken, GenerateTokenResponse } from '@/types'

export const mcpKeys = {
  all: ['mcp-tokens'] as const,
}

const DEFAULT_MOCK_TOKENS: APIToken[] = [
  {
    id: 1,
    user_id: 'local',
    name: 'Antigravity IDE Agent',
    token_prefix: 'mcp_pat_a8f9c2d1',
    last_used_at: new Date().toISOString(),
    created_at: new Date().toISOString(),
  },
]

export const useMcpTokens = () => useQuery({
    queryKey: mcpKeys.all,
    queryFn: async () => {
      try {
        const res = await mcpApi.listTokens()
        const tokens = res.data || []
        if (appConfig.enableMocks) {
          localStorage.setItem('local_mcp_tokens', JSON.stringify(tokens))
        }
        return tokens
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const local = localStorage.getItem('local_mcp_tokens')
        if (local) {
          try {
            return JSON.parse(local) as APIToken[]
          } catch {
            // ignore
          }
        }
        return DEFAULT_MOCK_TOKENS
      }
    },
  })

export const useCreateMcpToken = () => {
  const queryClient = useQueryClient()
  const toast = useAppToast()
  const { t } = useTranslation()

  return useMutation({
    mutationFn: async ({ name }: { name: string }): Promise<GenerateTokenResponse> => {
      try {
        const res = await mcpApi.generateToken({ name })
        return res.data
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const mockPAT: GenerateTokenResponse = {
          id: Date.now(),
          name,
          token_prefix: `mcp_pat_${Math.random().toString(16).slice(2, 10)}`,
          raw_token: `mcp_pat_${Array.from({ length: 48 }, () => Math.floor(Math.random() * 16).toString(16)).join('')}`,
          created_at: new Date().toISOString(),
        }
        const cached = localStorage.getItem('local_mcp_tokens')
        const tokens: APIToken[] = cached ? JSON.parse(cached) : DEFAULT_MOCK_TOKENS
        const updated: APIToken[] = [
          {
            id: mockPAT.id,
            user_id: 'local',
            name: mockPAT.name,
            token_prefix: mockPAT.token_prefix,
            created_at: mockPAT.created_at,
          },
          ...tokens,
        ]
        localStorage.setItem('local_mcp_tokens', JSON.stringify(updated))
        return mockPAT
      }
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
      try {
        await mcpApi.deleteToken(id)
      } catch (err) {
        if (!appConfig.enableMocks) {
          throw err
        }
        const cached = localStorage.getItem('local_mcp_tokens')
        const tokens: APIToken[] = cached ? JSON.parse(cached) : DEFAULT_MOCK_TOKENS
        const updated = tokens.filter((t) => t.id !== id)
        localStorage.setItem('local_mcp_tokens', JSON.stringify(updated))
      }
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
