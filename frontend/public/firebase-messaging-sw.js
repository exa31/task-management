// Service Worker for Firebase Cloud Messaging (Web Push)
// Handles background notifications when the web app is closed or minimized.

importScripts('https://www.gstatic.com/firebasejs/10.12.0/firebase-app-compat.js')
importScripts('https://www.gstatic.com/firebasejs/10.12.0/firebase-messaging-compat.js')

// Extract credentials passed via registration query string
const urlParams = new URLSearchParams(location.search)
const apiKey = urlParams.get('apiKey')
const projectId = urlParams.get('projectId')
const messagingSenderId = urlParams.get('messagingSenderId')
const appId = urlParams.get('appId')

if (apiKey && projectId && messagingSenderId) {
  try {
    firebase.initializeApp({
      apiKey,
      projectId,
      messagingSenderId,
      appId: appId || undefined
    })

    const messaging = firebase.messaging()

    messaging.onBackgroundMessage((payload) => {
      const notificationTitle =
        payload.notification?.title || payload.data?.title || 'Reminder Alert'
      const notificationOptions = {
        body:
          payload.notification?.body ||
          payload.data?.body ||
          'You have a scheduled reminder due now.',
        icon: '/favicon.ico',
        badge: '/favicon.ico',
        data: payload.data || {}
      }

      self.registration.showNotification(notificationTitle, notificationOptions)
    })
  } catch (err) {
    console.warn('[firebase-messaging-sw.js] Failed to initialize Firebase compat:', err)
  }
}

// Fallback listener for standard Web Push events
self.addEventListener('push', (event) => {
  if (!event.data) return

  try {
    const data = event.data.json()
    const notification = data.notification || {}
    const customData = data.data || {}
    const title = notification.title || customData.title || 'Reminder Alert'
    const body =
      notification.body ||
      customData.body ||
      customData.description ||
      'You have a scheduled reminder due now.'

    event.waitUntil(
      self.registration.showNotification(title, {
        body,
        icon: '/favicon.ico',
        badge: '/favicon.ico',
        data: customData
      })
    )
  } catch {
    const text = event.data.text()
    if (text) {
      event.waitUntil(
        self.registration.showNotification('Reminder Alert', {
          body: text,
          icon: '/favicon.ico'
        })
      )
    }
  }
})

// Focus or open web application when user clicks notification
self.addEventListener('notificationclick', (event) => {
  event.notification.close()
  const targetUrl = '/reminders'

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
      for (const client of clientList) {
        if (client.url && 'focus' in client) {
          return client.focus()
        }
      }
      if (clients.openWindow) {
        return clients.openWindow(targetUrl)
      }
    })
  )
})
