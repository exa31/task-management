import { useCallback, useEffect, useRef, useState } from 'react'

import { useAuthStore } from '@/stores'
import { logger } from '@/utils'

export interface AppleAuthorizationData {
  code: string
  id_token: string
  state?: string
}

export interface AppleUserData {
  name?: {
    firstName?: string
    lastName?: string
  }
  email?: string
}

export interface AppleSignInSuccessEventDetail {
  data: {
    authorization: AppleAuthorizationData
    user?: AppleUserData
  }
}

export interface AppleSignInFailureEventDetail {
  error: string
}

declare global {
  interface Window {
    // eslint-disable-next-line @typescript-eslint/naming-convention
    AppleID?: {
      auth: {
        init: (config: {
          clientId: string
          scope?: string
          redirectURI: string
          state?: string
          nonce?: string
          usePopup?: boolean
        }) => void
        signIn: () => Promise<{
          authorization: AppleAuthorizationData
          user?: AppleUserData
        }>
        renderButton?: () => void
      }
    }
  }
}

interface UseAppleAuthOptions {
  onSuccess?: () => void
  onError?: (error: string) => void
  onCancel?: () => void
  onMissingConfig?: () => void
}

/**
 * Decodes the payload portion of an Apple ID Token JWT.
 * Apple only sends the user profile (name and email) on the initial sign-in.
 * Subsequent sign-ins only contain the authorization token, from which email can be extracted.
 */
function parseAppleJwt(token: string): { email?: string; sub?: string } | null {
  try {
    const parts = token.split('.')
    if (parts.length < 2) return null
    const base64Url = parts[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      window
        .atob(base64)
        .split('')
        .map((c) => `%${`00${c.charCodeAt(0).toString(16)}`.slice(-2)}`)
        .join('')
    )
    return JSON.parse(jsonPayload)
  } catch {
    return null
  }
}

export function useAppleAuth(options: UseAppleAuthOptions = {}) {
  const { loginWithOAuth } = useAuthStore()
  const [isScriptReady, setIsScriptReady] = useState(false)
  const [isAuthenticating, setIsAuthenticating] = useState(false)
  const isInitializedRef = useRef(false)
  const isProcessingRef = useRef(false)

  const appleClientId = import.meta.env.VITE_APPLE_CLIENT_ID || ''
  const isConfigured = Boolean(appleClientId.trim())

  // Check when Apple JS SDK script is available on window
  useEffect(() => {
    if (typeof window === 'undefined') return

    if (window.AppleID?.auth) {
      setIsScriptReady(true)
      return
    }

    const interval = setInterval(() => {
      if (window.AppleID?.auth) {
        setIsScriptReady(true)
        clearInterval(interval)
      }
    }, 100)

    const timeout = setTimeout(() => clearInterval(interval), 6000)

    return () => {
      clearInterval(interval)
      clearTimeout(timeout)
    }
  }, [])

  // Initialize Apple ID configuration
  useEffect(() => {
    if (!isConfigured || !isScriptReady || typeof window === 'undefined') return

    if (window.AppleID?.auth && !isInitializedRef.current) {
      try {
        window.AppleID.auth.init({
          clientId: appleClientId,
          scope: 'name email',
          redirectURI: window.location.origin,
          state: 'taskhub-apple-auth',
          usePopup: true
        })
        isInitializedRef.current = true
      } catch (err) {
        logger.warn('AppleID.auth.init skipped or failed:', err)
      }
    }
  }, [isConfigured, isScriptReady, appleClientId])

  // Process authorization result from Apple
  const processAppleAuthData = useCallback(
    async (authorization: AppleAuthorizationData, user?: AppleUserData) => {
      if (isProcessingRef.current) return
      isProcessingRef.current = true
      setIsAuthenticating(true)

      try {
        let email = user?.email || ''
        let name = ''

        if (user?.name) {
          const first = user.name.firstName || ''
          const last = user.name.lastName || ''
          name = `${first} ${last}`.trim()
        }

        // If email was omitted (on repeat logins), decode from Apple ID token JWT
        if (!email && authorization.id_token) {
          const decoded = parseAppleJwt(authorization.id_token)
          if (decoded?.email) {
            const { email: jwtEmail } = decoded
            email = jwtEmail
          }
        }

        // Default name if still empty
        if (!name) {
          name = email ? email.split('@')[0] : 'Apple User'
        }

        await loginWithOAuth({
          provider: 'apple',
          email: email || 'user@privaterelay.appleid.com',
          name,
          id_token: authorization.id_token,
          code: authorization.code,
          avatar_url: `https://api.dicebear.com/7.x/bottts/svg?seed=${email || 'apple-user'}`
        })

        localStorage.setItem('has_logged_in_before', 'true')
        sessionStorage.removeItem('just_logged_out')
        options.onSuccess?.()
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Apple sign-in failed'
        options.onError?.(message)
      } finally {
        setIsAuthenticating(false)
        isProcessingRef.current = false
      }
    },
    [loginWithOAuth, options]
  )

  // Listen to Apple CustomEvents (AppleIDSignInOnSuccess and AppleIDSignInOnFailure)
  useEffect(() => {
    if (typeof document === 'undefined') return

    const handleSuccessEvent = (event: Event) => {
      const customEvent = event as CustomEvent<AppleSignInSuccessEventDetail['data']>
      if (customEvent.detail?.authorization) {
        processAppleAuthData(customEvent.detail.authorization, customEvent.detail.user)
      }
    }

    const handleFailureEvent = (event: Event) => {
      const customEvent = event as CustomEvent<AppleSignInFailureEventDetail>
      const error = customEvent.detail?.error
      if (error === 'popup_closed_by_user' || error === 'user_cancelled') {
        options.onCancel?.()
      } else if (error) {
        options.onError?.(error)
      }
      setIsAuthenticating(false)
      isProcessingRef.current = false
    }

    document.addEventListener('AppleIDSignInOnSuccess', handleSuccessEvent)
    document.addEventListener('AppleIDSignInOnFailure', handleFailureEvent)

    return () => {
      document.removeEventListener('AppleIDSignInOnSuccess', handleSuccessEvent)
      document.removeEventListener('AppleIDSignInOnFailure', handleFailureEvent)
    }
  }, [processAppleAuthData, options])

  // Triggered when user clicks "Continue with Apple"
  const loginWithApplePopup = useCallback(async () => {
    if (!isConfigured) {
      options.onMissingConfig?.()
      return
    }

    if (!window.AppleID?.auth) {
      options.onError?.(
        'Apple Sign In SDK is still loading. Please check your internet connection or try again.'
      )
      return
    }

    // Ensure initialized before popup call
    if (!isInitializedRef.current) {
      try {
        window.AppleID.auth.init({
          clientId: appleClientId,
          scope: 'name email',
          redirectURI: window.location.origin,
          state: 'taskhub-apple-auth',
          usePopup: true
        })
        isInitializedRef.current = true
      } catch (err) {
        options.onError?.(err instanceof Error ? err.message : 'Failed to initialize Apple SDK')
        return
      }
    }

    setIsAuthenticating(true)
    try {
      const res = await window.AppleID.auth.signIn()
      if (res && res.authorization) {
        await processAppleAuthData(res.authorization, res.user)
      }
    } catch (err: unknown) {
      const errObj = err as { error?: string }
      const errorMsg = errObj?.error || (err instanceof Error ? err.message : '')
      if (errorMsg === 'popup_closed_by_user' || errorMsg === 'user_cancelled') {
        options.onCancel?.()
      } else if (errorMsg) {
        options.onError?.(errorMsg || 'Apple sign-in was cancelled or failed')
      }
      setIsAuthenticating(false)
      isProcessingRef.current = false
    }
  }, [isConfigured, appleClientId, processAppleAuthData, options])

  return {
    isConfigured,
    isScriptReady,
    isAuthenticating,
    loginWithApplePopup
  }
}
