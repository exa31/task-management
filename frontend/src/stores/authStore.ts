import { create } from 'zustand'

import { authApi } from '@/api'
import { getCookie, removeCookie, setCookie } from '@/api/client'
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
  updateUser: (user: User) => void
  fetchProfile: () => Promise<void>
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
      const errorMsg = err instanceof Error ? err.message : 'Login failed'
      set({ loading: false, error: errorMsg })
      throw err
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
      const errorMsg = err instanceof Error ? err.message : 'Dev login failed'
      set({ loading: false, error: errorMsg })
      throw err
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

  updateUser: (user: User) => {
    localStorage.setItem(STORAGE_KEY_USER, JSON.stringify(user))
    set({ user })
  },

  fetchProfile: async () => {
    try {
      const res = await authApi.getProfile()
      if (res.data) {
        localStorage.setItem(STORAGE_KEY_USER, JSON.stringify(res.data))
        set({ user: res.data })
      }
    } catch (err) {
      logger.warn('Failed to fetch user profile:', err)
    }
  },
}))
