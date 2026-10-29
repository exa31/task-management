/**
 * Axios API Client with automated JWT silent-refresh interceptor and offline resilience.
 */
import type {
  AxiosError,
  AxiosInstance,
  AxiosRequestConfig,
  InternalAxiosRequestConfig
} from 'axios'
import axios from 'axios'

import { parseApiError } from './errors'

const STORAGE_KEY_TOKEN = 'auth_token'
const STORAGE_KEY_USER = 'auth_user'

interface CustomInternalRequestConfig extends InternalAxiosRequestConfig {
  _retry?: boolean
}

/**
 * Decode JWT token payload without external libraries.
 */
export const parseJwt = (token: string): { exp?: number; [key: string]: unknown } | null => {
  try {
    const parts = token.split('.')
    if (parts.length < 2) return null
    const base64Url = parts[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    if (typeof atob !== 'function') return null
    const decoded = atob(base64)
    const jsonPayload = decodeURIComponent(
      decoded
        .split('')
        .map((c: string) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload)
  } catch {
    return null
  }
}

/**
 * Checks whether a given JWT token string has expired (with an optional threshold buffer).
 */
export const isTokenExpired = (token: string, thresholdSeconds = 10): boolean => {
  const claims = parseJwt(token)
  if (!claims?.exp) return true
  return claims.exp * 1000 <= Date.now() + thresholdSeconds * 1000
}

export const getCookie = (name: string): string | null => {
  if (typeof document === 'undefined') {
    return null
  }
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`))
  return match ? decodeURIComponent(match[1]) : null
}

export const setCookie = (name: string, value: string, days = 7): void => {
  if (typeof document === 'undefined') {
    return
  }
  // Clear any existing cookie on current sub-path first to prevent path shadowing/duplicates
  if (typeof window !== 'undefined' && window.location.pathname && window.location.pathname !== '/') {
    document.cookie = `${name}=; path=${window.location.pathname}; expires=Thu, 01 Jan 1970 00:00:00 GMT; SameSite=Lax`
  }
  const expires = new Date(Date.now() + days * 864e5).toUTCString()
  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; expires=${expires}; SameSite=Lax`
}

export const removeCookie = (name: string): void => {
  if (typeof document === 'undefined') {
    return
  }
  const paths = ['/', '/api', '/api/v1', typeof window !== 'undefined' ? window.location.pathname : '']
  const uniquePaths = Array.from(new Set(paths.filter(Boolean)))
  for (const p of uniquePaths) {
    document.cookie = `${name}=; path=${p}; expires=Thu, 01 Jan 1970 00:00:00 GMT; SameSite=Lax`
  }
}

/**
 * Retrieve the active, unexpired access token, preferring authoritative localStorage.
 */
export const getValidAccessToken = (): string | null => {
  if (typeof window === 'undefined') {
    return null
  }

  // 1. Check localStorage first (our authoritative client store)
  const localToken = localStorage.getItem(STORAGE_KEY_TOKEN)
  if (localToken && !isTokenExpired(localToken)) {
    return localToken
  }

  // 2. Check cookie if localStorage is empty or expired
  const cookieToken = getCookie('access_token')
  if (cookieToken && !isTokenExpired(cookieToken)) {
    return cookieToken
  }

  // 3. Fallback: return token if exists rather than null
  return localToken || cookieToken || null
}

export const apiBaseUrl = (
  import.meta.env.VITE_API_URL ||
  import.meta.env.VITE_APP_API_URL ||
  ''
).replace(/\/+$/, '')

export const axiosInstance: AxiosInstance = axios.create({
  baseURL: apiBaseUrl,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Initialize common Authorization header if a valid token is present
const initialToken = getValidAccessToken()
if (initialToken) {
  axiosInstance.defaults.headers.common['Authorization'] = `Bearer ${initialToken}`
}

// === Request Interceptor ===
axiosInstance.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // If headers already has an Authorization Bearer token explicitly set (e.g. from retry),
    // and it is not expired, preserve it!
    const existingAuth = config.headers?.get
      ? config.headers.get('Authorization')
      : config.headers?.['Authorization']

    if (typeof existingAuth === 'string' && existingAuth.startsWith('Bearer ')) {
      const existingToken = existingAuth.slice(7).trim()
      if (existingToken && !isTokenExpired(existingToken)) {
        return config
      }
    }

    const token = getValidAccessToken()
    if (token) {
      if (config.headers?.set) {
        config.headers.set('Authorization', `Bearer ${token}`)
      } else if (config.headers) {
        config.headers.Authorization = `Bearer ${token}`
      }
    }
    return config
  },
  (error: AxiosError) => Promise.reject(error)
)

// === Response Interceptor with Token Refresh Flow ===
let refreshPromise: Promise<string> | null = null

const performSilentRefresh = async (): Promise<string> => {
  try {
    const refreshResponse = await axios.post(
      `${apiBaseUrl}/api/v1/auth/refresh`,
      {},
      { withCredentials: true }
    )

    const rawData = refreshResponse.data?.data || refreshResponse.data
    const newAccessToken = rawData?.access_token || rawData?.tokens?.access_token

    if (!newAccessToken) {
      throw new Error('No access token returned from refresh endpoint')
    }

    // Persist new token in localStorage and cookie
    localStorage.setItem(STORAGE_KEY_TOKEN, newAccessToken)
    setCookie('access_token', newAccessToken)

    // Update default instance header so subsequent requests pick it up immediately
    axiosInstance.defaults.headers.common['Authorization'] = `Bearer ${newAccessToken}`

    return newAccessToken
  } catch (refreshErr) {
    removeCookie('access_token')
    localStorage.removeItem(STORAGE_KEY_TOKEN)
    localStorage.removeItem('auth_refresh_token')
    localStorage.removeItem(STORAGE_KEY_USER)
    delete axiosInstance.defaults.headers.common['Authorization']

    if (typeof window !== 'undefined' && window.location.pathname !== '/auth/login') {
      window.location.href = '/auth/login'
    }
    throw refreshErr
  }
}

axiosInstance.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as CustomInternalRequestConfig

    if (!originalRequest) {
      return Promise.reject(error)
    }

    const requestUrl = originalRequest.url || ''
    const isAuthRoute =
      requestUrl.includes('/auth/refresh') ||
      requestUrl.includes('/auth/oauth') ||
      requestUrl.includes('/auth/dev-login')

    // If 401 and not already retried and not an auth route itself
    if (error.response?.status === 401 && !originalRequest._retry && !isAuthRoute) {
      originalRequest._retry = true

      const sentAuth = originalRequest.headers?.get
        ? (originalRequest.headers.get('Authorization') as string)
        : (originalRequest.headers?.['Authorization'] as string)
      const sentToken =
        typeof sentAuth === 'string' ? sentAuth.replace(/^Bearer\s+/, '').trim() : null

      const currentToken = getValidAccessToken()

      // If a newer token was already refreshed while this request was in flight,
      // retry immediately with currentToken without triggering another refresh!
      if (
        currentToken &&
        sentToken &&
        currentToken !== sentToken &&
        !isTokenExpired(currentToken)
      ) {
        if (originalRequest.headers?.set) {
          originalRequest.headers.set('Authorization', `Bearer ${currentToken}`)
        } else if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${currentToken}`
        }
        return axiosInstance(originalRequest)
      }

      // If refresh is already in progress, wait for it; otherwise initiate it
      if (!refreshPromise) {
        refreshPromise = performSilentRefresh().finally(() => {
          refreshPromise = null
        })
      }

      return refreshPromise
        .then((newAccessToken) => {
          if (originalRequest.headers?.set) {
            originalRequest.headers.set('Authorization', `Bearer ${newAccessToken}`)
          } else if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${newAccessToken}`
          }
          return axiosInstance(originalRequest)
        })
        .catch((refreshErr) => Promise.reject(refreshErr))
    }

    if (axios.isCancel(error)) {
      return Promise.reject(error)
    }

    return Promise.reject(parseApiError(error))
  }
)

// === Universal Unwrapped API Client Adapter ===
export const apiClient = {
  instance: axiosInstance,

  get: async <T>(url: string, config?: AxiosRequestConfig): Promise<{ data: T }> => {
    try {
      const res = await axiosInstance.get(url, config)
      const trimmed = typeof res.data === 'string' ? res.data.trim().toLowerCase() : ''
      if (trimmed.startsWith('<!doctype') || trimmed.startsWith('<html')) {
        throw new Error(
          'Received HTML response instead of JSON. Check your backend URL or proxy configuration.'
        )
      }
      const unwrapped = res.data?.data !== undefined ? res.data.data : res.data
      return { data: unwrapped as T }
    } catch (err) {
      if (axios.isCancel(err)) {
        throw err
      }
      throw parseApiError(err)
    }
  },

  post: async <T>(
    url: string,
    payload?: unknown,
    config?: AxiosRequestConfig
  ): Promise<{ data: T }> => {
    try {
      const res = await axiosInstance.post(url, payload, config)
      const unwrapped = res.data?.data !== undefined ? res.data.data : res.data
      return { data: unwrapped as T }
    } catch (err) {
      if (axios.isCancel(err)) {
        throw err
      }
      throw parseApiError(err)
    }
  },

  put: async <T>(
    url: string,
    payload?: unknown,
    config?: AxiosRequestConfig
  ): Promise<{ data: T }> => {
    try {
      const res = await axiosInstance.put(url, payload, config)
      const unwrapped = res.data?.data !== undefined ? res.data.data : res.data
      return { data: unwrapped as T }
    } catch (err) {
      if (axios.isCancel(err)) {
        throw err
      }
      throw parseApiError(err)
    }
  },

  delete: async <T>(url: string, config?: AxiosRequestConfig): Promise<{ data: T }> => {
    try {
      const res = await axiosInstance.delete(url, config)
      const unwrapped = res.data?.data !== undefined ? res.data.data : res.data
      return { data: unwrapped as T }
    } catch (err) {
      if (axios.isCancel(err)) {
        throw err
      }
      throw parseApiError(err)
    }
  }
}
