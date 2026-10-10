import { initializeApp } from 'firebase/app'
import { getMessaging, getToken, isSupported, onMessage } from 'firebase/messaging'

import { appToast } from '@/components'
import { reminderKeys } from '@/hooks/useRemindersQuery'
import { queryClient } from '@/queryClient'

export const firebaseConfig = {
  apiKey: import.meta.env.VITE_FIREBASE_API_KEY || '',
  authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN || '',
  projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID || '',
  storageBucket: import.meta.env.VITE_FIREBASE_STORAGE_BUCKET || '',
  messagingSenderId: import.meta.env.VITE_FIREBASE_MESSAGING_SENDER_ID || '',
  appId: import.meta.env.VITE_FIREBASE_APP_ID || '',
  vapidKey: import.meta.env.VITE_FIREBASE_VAPID_KEY || ''
}

export const isFirebaseConfigured = (): boolean => Boolean(
    firebaseConfig.apiKey &&
    firebaseConfig.projectId &&
    firebaseConfig.messagingSenderId &&
    firebaseConfig.vapidKey
  )

let messagingInstance: ReturnType<typeof getMessaging> | null = null
let foregroundListenerRegistered = false

export const requestFcmToken = async (): Promise<string | null> => {
  if (typeof window === 'undefined') return null

  if (!isFirebaseConfigured()) {
    return null
  }

  const supported = await isSupported().catch(() => false)
  if (!supported) {
    return null
  }

  if (!('serviceWorker' in navigator) || !('Notification' in window)) {
    return null
  }

  // Request browser notification permission if not yet decided
  if (Notification.permission === 'default') {
    const permission = await Notification.requestPermission()
    if (permission !== 'granted') {
      return null
    }
  } else if (Notification.permission !== 'granted') {
    return null
  }

  try {
    const app = initializeApp(firebaseConfig)
    if (!messagingInstance) {
      messagingInstance = getMessaging(app)
    }

    // Register service worker with query params so SW has config
    const swUrl = `/firebase-messaging-sw.js?apiKey=${encodeURIComponent(
      firebaseConfig.apiKey
    )}&projectId=${encodeURIComponent(
      firebaseConfig.projectId
    )}&messagingSenderId=${encodeURIComponent(
      firebaseConfig.messagingSenderId
    )}&appId=${encodeURIComponent(firebaseConfig.appId)}`

    const registration = await navigator.serviceWorker.register(swUrl)
    await navigator.serviceWorker.ready

    const token = await getToken(messagingInstance, {
      vapidKey: firebaseConfig.vapidKey,
      serviceWorkerRegistration: registration
    })

    // Listen for foreground push messages
    if (!foregroundListenerRegistered) {
      foregroundListenerRegistered = true
      onMessage(messagingInstance, (payload) => {
        const title = payload.notification?.title || payload.data?.title || 'Reminder Alert'
        const body =
          payload.notification?.body ||
          payload.data?.body ||
          'You have a scheduled reminder due now.'

        // Invalidate queries to refresh badges and list
        queryClient.invalidateQueries({ queryKey: reminderKeys.all })

        // Toast in app
        appToast.info(`⏰ ${title}: ${body}`)
      })
    }

    return token || null
  } catch (err) {
    console.warn('Failed to obtain FCM token:', err)
    return null
  }
}
