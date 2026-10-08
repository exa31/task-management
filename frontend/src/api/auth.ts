import type { AuthResponse, DevLoginPayload, OAuthLoginPayload, User } from '@/types'

import { apiClient } from './client'

export const authApi = {
  oauthLogin: (payload: OAuthLoginPayload) =>
    apiClient.post<AuthResponse>('/api/v1/auth/oauth', payload),

  devLogin: (payload: DevLoginPayload) =>
    apiClient.post<AuthResponse>('/api/v1/auth/dev-login', payload),

  getProfile: () =>
    apiClient.get<User>('/api/v1/auth/profile'),

  logout: () =>
    apiClient.post('/api/v1/auth/logout'),
}
