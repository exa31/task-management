import type {
  AuthResponse,
  BindOAuthPayload,
  DeleteAccountPayload,
  DevLoginPayload,
  OAuthLoginPayload,
  UnbindOAuthPayload,
  User
} from '@/types'

import { apiClient } from './client'

export const authApi = {
  oauthLogin: (payload: OAuthLoginPayload) =>
    apiClient.post<AuthResponse>('/api/v1/auth/oauth', payload),

  devLogin: (payload: DevLoginPayload) =>
    apiClient.post<AuthResponse>('/api/v1/auth/dev-login', payload),

  getProfile: () => apiClient.get<User>('/api/v1/auth/profile'),

  updateProfile: (payload: {
    name?: string
    timezone?: string
    due_date_warning_threshold?: number
    email_notifications_enabled?: boolean
  }) => apiClient.put<User>('/api/v1/auth/profile', payload),

  uploadAvatar: (file: Blob | File) => {
    const formData = new FormData()
    formData.append('avatar', file, 'avatar.webp')
    return apiClient.post<User>('/api/v1/auth/avatar', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },

  bindOAuth: (payload: BindOAuthPayload) =>
    apiClient.post<User>('/api/v1/auth/bind-oauth', payload),

  unbindOAuth: (payload: UnbindOAuthPayload) =>
    apiClient.post<User>('/api/v1/auth/unbind-oauth', payload),

  deleteAccount: (payload: DeleteAccountPayload) =>
    apiClient.delete('/api/v1/auth/account', { data: payload }),

  logout: () => apiClient.post('/api/v1/auth/logout')
}
