/**
 * Structured API Error models and localization helpers.
 */

import { logger } from '@/utils'

export interface ApiErrorResponseData {
  message?: string
  success?: boolean
  code?: string
  data?: unknown
  timestamp?: string
}

export class ApiError extends Error {
  code: string
  statusCode: number
  data?: unknown

  constructor(
    message: string,
    code: string = 'UNKNOWN_ERROR',
    statusCode: number = 500,
    data?: unknown
  ) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.statusCode = statusCode
    this.data = data
  }
}

function getFallbackCodeForStatus(status: number): string {
  switch (status) {
    case 400:
      return 'BAD_REQUEST'
    case 401:
      return 'UNAUTHORIZED'
    case 403:
      return 'FORBIDDEN'
    case 404:
      return 'NOT_FOUND'
    case 409:
      return 'CONFLICT'
    case 422:
      return 'VALIDATION_ERROR'
    case 429:
      return 'RATE_LIMIT_EXCEEDED'
    case 500:
      return 'INTERNAL_SERVER_ERROR'
    case 503:
      return 'SERVICE_ERROR'
    default:
      return 'UNKNOWN_ERROR'
  }
}

/**
 * Normalizes any caught error (AxiosError, Error, string) into a structured ApiError.
 */
export function parseApiError(err: unknown): ApiError {
  if (err instanceof ApiError) {
    return err
  }

  // Axios error with backend BaseResponse
  if (err && typeof err === 'object' && 'response' in err) {
    const axiosErr = err as {
      response?: {
        status?: number
        data?: ApiErrorResponseData | string
      }
      message?: string
    }

    const status = axiosErr.response?.status || 500
    const resData = axiosErr.response?.data

    if (resData && typeof resData === 'object') {
      const code = resData.code || getFallbackCodeForStatus(status)
      const message = resData.message || axiosErr.message || 'Request failed'
      return new ApiError(message, code, status, resData.data)
    }

    if (typeof resData === 'string' && resData.trim()) {
      return new ApiError(resData, getFallbackCodeForStatus(status), status)
    }

    return new ApiError(axiosErr.message || 'Request failed', getFallbackCodeForStatus(status), status)
  }

  // Network / connection failure
  if (
    err &&
    typeof err === 'object' &&
    'code' in err &&
    (err as { code?: string }).code === 'ERR_NETWORK'
  ) {
    return new ApiError('Network connection error', 'NETWORK_ERROR', 0)
  }

  if (err instanceof Error) {
    return new ApiError(err.message, 'UNKNOWN_ERROR', 500)
  }

  return new ApiError(String(err || 'Unknown error'), 'UNKNOWN_ERROR', 500)
}

/**
 * Returns a user-friendly, localized error message from the translation dictionary.
 * Raw technical backend messages are suppressed from the UI to protect UX and language consistency,
 * but are logged via logger.warn for developer diagnostics.
 */
export function getLocalizedErrorMessage(
  err: unknown,
  t: (key: string, params?: Record<string, string | number>) => string
): string {
  const apiErr = parseApiError(err)

  // Log raw technical diagnostics for developers in console
  logger.warn(`[ApiError ${apiErr.code}] (Status: ${apiErr.statusCode}): ${apiErr.message}`, apiErr.data)

  // 1. Exact match in translation dictionary
  const translationKey = `errors.${apiErr.code}`
  const localizedMsg = t(translationKey)
  if (localizedMsg && localizedMsg !== translationKey) {
    return localizedMsg
  }

  // 2. Intelligent pattern-based fallback
  if (apiErr.code.endsWith('_NOT_FOUND')) {
    return t('errors.NOT_FOUND')
  }
  if (apiErr.code.includes('DUPLICATE') || apiErr.code.includes('ALREADY_EXISTS')) {
    return t('errors.DUPLICATE_ENTRY')
  }
  if (apiErr.code.includes('VALIDATION')) {
    return t('errors.VALIDATION_ERROR')
  }
  if (apiErr.code.includes('AUTH') || apiErr.code.includes('TOKEN')) {
    return t('errors.UNAUTHORIZED')
  }

  // 3. Fallback to generic HTTP status code translation if available
  const fallbackCode = getFallbackCodeForStatus(apiErr.statusCode)
  const statusMsg = t(`errors.${fallbackCode}`)
  if (statusMsg && statusMsg !== `errors.${fallbackCode}`) {
    return statusMsg
  }

  return t('errors.operationFailed') || t('errors.UNKNOWN_ERROR')
}
