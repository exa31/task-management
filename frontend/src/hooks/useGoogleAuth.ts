import { useCallback, useEffect, useRef, useState } from 'react'

import { useAuthStore } from '@/stores'
import { logger } from '@/utils'

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (config: {
            client_id: string
            callback: (response: { credential: string }) => void
            auto_select?: boolean
            cancel_on_tap_outside?: boolean
          }) => void
          prompt: (
            notification?: (notification: {
              isNotDisplayed: () => boolean
              isSkippedMoment: () => boolean
              isDismissedMoment: () => boolean
              getNotDisplayedReason: () => string
              getSkippedReason: () => string
              getDismissedReason: () => string
            }) => void
          ) => void
          disableAutoSelect: () => void
        }
        oauth2: {
          initCodeClient: (config: {
            client_id: string
            scope: string
            ux_mode?: 'popup' | 'redirect'
            redirect_uri?: string
            callback: (response: {
              code?: string
              error?: string
              error_description?: string
            }) => void
          }) => {
            requestCode: () => void
          }
        }
      }
    }
  }
}

interface UseGoogleAuthOptions {
  onSuccess?: () => void
  onError?: (error: string) => void
  onMissingConfig?: () => void
}

export function useGoogleAuth(options: UseGoogleAuthOptions = {}) {
  const { loginWithOAuth } = useAuthStore()
  const [isScriptReady, setIsScriptReady] = useState(false)
  const [isAuthenticating, setIsAuthenticating] = useState(false)
  const codeClientRef = useRef<{ requestCode: () => void } | null>(null)

  const googleClientId = import.meta.env.VITE_GOOGLE_CLIENT_ID || ''
  const isConfigured = Boolean(googleClientId.trim())

  // Check when Google Identity Services (GIS) SDK script is available on window
  useEffect(() => {
    if (typeof window === 'undefined') return

    if (window.google?.accounts) {
      setIsScriptReady(true)
      return
    }

    const interval = setInterval(() => {
      if (window.google?.accounts) {
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

  // Initialize One Tap prompt only if user has logged in before and didn't just log out
  useEffect(() => {
    if (!isConfigured || !isScriptReady || typeof window === 'undefined') return

    const hasLoggedInBefore = localStorage.getItem('has_logged_in_before') === 'true'
    const justLoggedOut = sessionStorage.getItem('just_logged_out') === 'true'

    // If user has previous login history on this device and hasn't freshly logged out in this session:
    if (hasLoggedInBefore && !justLoggedOut && window.google?.accounts?.id) {
      try {
        window.google.accounts.id.initialize({
          client_id: googleClientId,
          cancel_on_tap_outside: true,
          callback: async (response) => {
            if (response.credential) {
              setIsAuthenticating(true)
              try {
                await loginWithOAuth({
                  provider: 'google',
                  id_token: response.credential,
                })
                localStorage.setItem('has_logged_in_before', 'true')
                sessionStorage.removeItem('just_logged_out')
                options.onSuccess?.()
              } catch (err) {
                const message = err instanceof Error ? err.message : 'Google One Tap sign-in failed'
                options.onError?.(message)
              } finally {
                setIsAuthenticating(false)
              }
            }
          },
        })

        // Trigger Google One Tap floating dialog
        window.google.accounts.id.prompt((notification) => {
          if (notification.isNotDisplayed()) {
            // Silently suppressed by Google browser cool-down or no Google session
          }
        })
      } catch (err) {
        logger.warn('Google One Tap init skipped:', err)
      }
    }
  }, [isConfigured, isScriptReady, googleClientId, loginWithOAuth, options])

  // Setup Popup Authorization Code Client (google.accounts.oauth2.initCodeClient)
  useEffect(() => {
    if (!isConfigured || !isScriptReady || typeof window === 'undefined') return

    if (window.google?.accounts?.oauth2?.initCodeClient) {
      try {
        codeClientRef.current = window.google.accounts.oauth2.initCodeClient({
          client_id: googleClientId,
          scope: 'openid email profile',
          ux_mode: 'popup',
          callback: async (response) => {
            if (response.error) {
              if (response.error !== 'popup_closed_by_user') {
                options.onError?.(response.error_description || response.error)
              }
              setIsAuthenticating(false)
              return
            }

            if (response.code) {
              setIsAuthenticating(true)
              try {
                await loginWithOAuth({
                  provider: 'google',
                  code: response.code,
                })
                localStorage.setItem('has_logged_in_before', 'true')
                sessionStorage.removeItem('just_logged_out')
                options.onSuccess?.()
              } catch (err) {
                const message = err instanceof Error ? err.message : 'Google sign-in failed'
                options.onError?.(message)
              } finally {
                setIsAuthenticating(false)
              }
            }
          },
        })
      } catch (err) {
        logger.warn('Google initCodeClient error:', err)
      }
    }
  }, [isConfigured, isScriptReady, googleClientId, loginWithOAuth, options])

  // Triggered when user clicks the "Continue with Google" button
  const loginWithGooglePopup = useCallback(() => {
    if (!isConfigured) {
      options.onMissingConfig?.()
      return
    }

    if (!codeClientRef.current) {
      // If code client isn't ready yet, try re-initializing if SDK is present
      if (window.google?.accounts?.oauth2?.initCodeClient) {
        codeClientRef.current = window.google.accounts.oauth2.initCodeClient({
          client_id: googleClientId,
          scope: 'openid email profile',
          ux_mode: 'popup',
          callback: async (response) => {
            if (response.code) {
              setIsAuthenticating(true)
              try {
                await loginWithOAuth({
                  provider: 'google',
                  code: response.code,
                })
                localStorage.setItem('has_logged_in_before', 'true')
                sessionStorage.removeItem('just_logged_out')
                options.onSuccess?.()
              } catch (err) {
                const message = err instanceof Error ? err.message : 'Google sign-in failed'
                options.onError?.(message)
              } finally {
                setIsAuthenticating(false)
              }
            }
          },
        })
      } else {
        options.onError?.('Google SDK is still loading. Please check your internet connection or try again.')
        return
      }
    }

    setIsAuthenticating(true)
    try {
      codeClientRef.current.requestCode()
    } catch (err) {
      setIsAuthenticating(false)
      options.onError?.(err instanceof Error ? err.message : 'Failed to launch Google sign-in dialog')
    }
  }, [isConfigured, googleClientId, loginWithOAuth, options])

  return {
    isConfigured,
    isScriptReady,
    isAuthenticating,
    loginWithGooglePopup,
  }
}
