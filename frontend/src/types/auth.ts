export interface User {
  id: string
  name: string
  email: string
  avatar_url?: string
  avatarUrl?: string
  role?: string
  status?: string
  provider?: string
  google_bound?: boolean
  apple_bound?: boolean
  has_password?: boolean
  timezone?: string
  due_date_warning_threshold?: number
  email_notifications_enabled?: boolean
  created_at?: string
}

export interface AuthTokens {
  access_token: string
  refresh_token: string
  expires_at: string
  token_type: string
}

export interface AuthResponse {
  user: User
  tokens: AuthTokens
}

export interface OAuthLoginPayload {
  provider: 'google' | 'apple'
  id_token?: string
  code?: string
  email?: string
  name?: string
  avatar_url?: string
}

export interface BindOAuthPayload {
  provider: 'google' | 'apple'
  id_token?: string
  code?: string
  email?: string
  name?: string
}

export interface UnbindOAuthPayload {
  provider: 'google' | 'apple'
}

export interface DeleteAccountPayload {
  confirmation: string
}

export interface DevLoginPayload {
  email: string
  name?: string
}
