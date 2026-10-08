export interface APIToken {
  id: number
  user_id: string
  name: string
  token_prefix: string
  last_used_at?: string | null
  created_at: string
  expires_at?: string | null
}

export interface GenerateTokenPayload {
  name: string
}

export interface GenerateTokenResponse {
  id: number
  name: string
  token_prefix: string
  raw_token: string
  created_at: string
  expires_at?: string | null
}
