import type { APIToken, GenerateTokenPayload, GenerateTokenResponse } from '@/types'

import { apiClient } from './client'

export const mcpApi = {
  listTokens: () => apiClient.get<APIToken[]>('/api/v1/tokens'),

  generateToken: (payload: GenerateTokenPayload) =>
    apiClient.post<GenerateTokenResponse>('/api/v1/tokens', payload),

  deleteToken: (id: number) => apiClient.delete(`/api/v1/tokens/${id}`),

  getDiscovery: () => apiClient.get<Record<string, unknown>>('/api/v1/mcp')
}
