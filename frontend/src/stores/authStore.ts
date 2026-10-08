import { create } from 'zustand'

import { authApi } from '@/api'
import { getCookie, removeCookie, setCookie } from '@/api/client'
import { appConfig } from '@/config'
import type { DevLoginPayload, OAuthLoginPayload, User } from '@/types'
import { logger } from '@/utils'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  loading: boolean
  error: string | null
  loginWithOAuth: (payload: OAuthLoginPayload) => Promise<void>
  loginDev: (payload: DevLoginPayload) => Promise<void>
  logout: () => void
  checkAuth: () => void
}

const STORAGE_KEY_TOKEN = 'auth_token'
const STORAGE_KEY_USER = 'auth_user'

export const useAuthStore = create<AuthState>((set) => ({
  user: (() => {
    try {
      const u = localStorage.getItem(STORAGE_KEY_USER)
      return u ? JSON.parse(u) : null
    } catch {
      return null
    }
  })(),
  token: getCookie('access_token') || localStorage.getItem(STORAGE_KEY_TOKEN),
  isAuthenticated: Boolean(getCookie('access_token') || localStorage.getItem(STORAGE_KEY_TOKEN)),
  loading: false,
  error: null,

  checkAuth: () => {
    const token = getCookie('access_token') || localStorage.getItem(STORAGE_KEY_TOKEN)
    const userStr = localStorage.getItem(STORAGE_KEY_USER)
    if (token && userStr) {
      try {
        set({ token, user: JSON.parse(userStr), isAuthenticated: true })
      } catch {
        set({ token: null, user: null, isAuthenticated: false })
      }
    } else {
      set({ token: null, user: null, isAuthenticated: false })
    }
  },

  loginWithOAuth: async (payload: OAuthLoginPayload) => {
    set({ loading: true, error: null })
    try {
      const res = await authApi.oauthLogin(payload)
      const { data } = res
      if (data.tokens?.access_token) {
        setCookie('access_token', data.tokens.access_token)
        localStorage.setItem(STORAGE_KEY_TOKEN, data.tokens.access_token)
      }
      // Note: refresh_token is handled automatically via HttpOnly cookie for web security.
      // Do NOT store refresh_token in localStorage.
      localStorage.setItem(STORAGE_KEY_USER, JSON.stringify(data.user))
      set({
        user: data.user,
        token: data.tokens.access_token,
        isAuthenticated: true,
        loading: false,
      })
    } catch (err: unknown) {
      if (!appConfig.enableMocks) {
        const errorMsg = err instanceof Error ? err.message : 'Login failed'
        set({ loading: false, error: errorMsg })
        throw err
      }
      // Local demo fallback if backend is offline and mocks are explicitly enabled
      const fallbackEmail = payload.email || 'user@example.com'
      const mockUser: User = {
        id: `user-demo-${Date.now()}`,
        name: payload.name || fallbackEmail.split('@')[0],
        email: fallbackEmail,
        avatar_url: payload.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${fallbackEmail}`,
        provider: payload.provider,
        created_at: new Date().toISOString(),
      }
      const demoToken = `demo_token_${Date.now()}`
      setCookie('access_token', demoToken)
      localStorage.setItem(STORAGE_KEY_TOKEN, demoToken)
      localStorage.setItem(STORAGE_KEY_USER, JSON.stringify(mockUser))
      set({
        user: mockUser,
        token: demoToken,
        isAuthenticated: true,
        loading: false,
      })
    }
  },

  loginDev: async (payload: DevLoginPayload) => {
    set({ loading: true, error: null })
    try {
      const res = await authApi.devLogin(payload)
      const { data } = res
      if (data.tokens?.access_token) {
        setCookie('access_token', data.tokens.access_token)
        localStorage.setItem(STORAGE_KEY_TOKEN, data.tokens.access_token)
      }
      // Note: refresh_token is handled automatically via HttpOnly cookie for web security.
      // Do NOT store refresh_token in localStorage.
      localStorage.setItem(STORAGE_KEY_USER, JSON.stringify(data.user))
      set({
        user: data.user,
        token: data.tokens.access_token,
        isAuthenticated: true,
        loading: false,
      })
    } catch (err: unknown) {
      if (!appConfig.enableMocks) {
        const errorMsg = err instanceof Error ? err.message : 'Dev login failed'
        set({ loading: false, error: errorMsg })
        throw err
      }
      // Fallback demo user if mocks are explicitly enabled
      const demoUser: User = {
        id: 'demo-user-local',
        name: payload.name || 'Demo Creator',
        email: payload.email,
        avatar_url: `https://api.dicebear.com/7.x/bottts/svg?seed=${payload.email}`,
        provider: 'google',
        created_at: new Date().toISOString(),
      }
      const demoToken = 'demo_token_local'
      setCookie('access_token', demoToken)
      localStorage.setItem(STORAGE_KEY_TOKEN, demoToken)
      localStorage.setItem(STORAGE_KEY_USER, JSON.stringify(demoUser))
      set({
        user: demoUser,
        token: demoToken,
        isAuthenticated: true,
        loading: false,
      })
    }
  },

  logout: () => {
    removeCookie('access_token')
    localStorage.removeItem(STORAGE_KEY_TOKEN)
    localStorage.removeItem(STORAGE_KEY_USER)
    // Clean up any legacy refresh token from previous versions if present
    localStorage.removeItem('auth_refresh_token')

    // Disable Google One Tap auto-selection on logout
    if (typeof window !== 'undefined' && window.google?.accounts?.id?.disableAutoSelect) {
      try {
        window.google.accounts.id.disableAutoSelect()
      } catch (err) {
        logger.warn('Google auto-select disable error:', err)
      }
    }
    // Record that the user intentionally logged out so One Tap doesn't auto-prompt
    try {
      sessionStorage.setItem('just_logged_out', 'true')
    } catch (err) {
      logger.warn('Session storage access error:', err)
    }

    authApi.logout().catch(() => { })
    set({ user: null, token: null, isAuthenticated: false })
  },
}))
