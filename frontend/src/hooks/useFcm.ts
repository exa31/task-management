import { useEffect, useRef } from 'react'

import { reminderApi } from '@/api/reminders'
import { isFirebaseConfigured, requestFcmToken } from '@/config/firebase'
import { useAuthStore } from '@/stores'

export const useFcm = () => {
  const token = useAuthStore((s) => s.token)
  const userId = useAuthStore((s) => s.user?.id)
  const isRegisteredRef = useRef(false)

  useEffect(() => {
    if (!token || !userId) {
      isRegisteredRef.current = false
      return
    }

    if (!isFirebaseConfigured()) {
      return
    }

    const storageKey = `fcm_reg_${userId}`
    const alreadyRegistered = sessionStorage.getItem(storageKey)

    if (alreadyRegistered || isRegisteredRef.current) {
      return
    }

    let isMounted = true

    const registerFcm = async () => {
      try {
        const fcmToken = await requestFcmToken()
        if (fcmToken && isMounted) {
          await reminderApi.registerDeviceToken({
            token: fcmToken,
            platform: 'web'
          })
          sessionStorage.setItem(storageKey, fcmToken)
          isRegisteredRef.current = true
        }
      } catch (err) {
        console.warn('Could not register FCM device token with backend:', err)
      }
    }

    registerFcm()

    return () => {
      isMounted = false
    }
  }, [token, userId])
}
