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

interface QueuedPromise {
  resolve: (token: string) => void
  reject: (err: unknown) => void
}

let isRefreshing = false
let failedQueue: QueuedPromise[] = []

const processQueue = (error: unknown, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error)
    } else if (token) {
      prom.resolve(token)
    }
  })
  failedQueue = []
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
  const expires = new Date(Date.now() + days * 864e5).toUTCString()
  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; expires=${expires}; SameSite=Lax`
}

export const removeCookie = (name: string): void => {
  if (typeof document === 'undefined') {
    return
  }
  document.cookie = `${name}=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; SameSite=Lax`
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

// === Request Interceptor ===
axiosInstance.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getCookie('access_token') || localStorage.getItem(STORAGE_KEY_TOKEN)
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error: AxiosError) => Promise.reject(error)
)

// === Response Interceptor with Token Refresh Flow ===
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
      // If refresh is already in progress by another request, queue this one
      if (isRefreshing) {
        return new Promise<string>((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        })
          .then((newToken) => {
            if (originalRequest.headers) {
              originalRequest.headers.Authorization = `Bearer ${newToken}`
            }
            return axiosInstance(originalRequest)
          })
          .catch((err) => Promise.reject(err))
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        const storedRefreshToken = getCookie('refresh_token') || ''
        const refreshResponse = await axios.post(
          `${apiBaseUrl}/api/v1/auth/refresh`,
          storedRefreshToken ? { refresh_token: storedRefreshToken } : {},
          { withCredentials: true }
        )

        const rawData = refreshResponse.data?.data || refreshResponse.data
        const newAccessToken = rawData?.access_token || rawData?.tokens?.access_token

        if (newAccessToken) {
          setCookie('access_token', newAccessToken)
          localStorage.setItem(STORAGE_KEY_TOKEN, newAccessToken)

          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${newAccessToken}`
          }

          processQueue(null, newAccessToken)
          return axiosInstance(originalRequest)
        }
        throw new Error('No access token returned from refresh endpoint')
      } catch (refreshErr) {
        processQueue(refreshErr, null)
        removeCookie('access_token')
        localStorage.removeItem(STORAGE_KEY_TOKEN)
        localStorage.removeItem('auth_refresh_token')
        localStorage.removeItem(STORAGE_KEY_USER)
        if (window.location.pathname !== '/auth/login') {
          window.location.href = '/auth/login'
        }
        return Promise.reject(refreshErr)
      } finally {
        isRefreshing = false
      }
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
