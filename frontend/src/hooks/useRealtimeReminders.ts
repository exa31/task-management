import { useEffect } from 'react'

import { apiBaseUrl } from '@/api/client'
import { appToast } from '@/components'
import { reminderKeys } from '@/hooks/useRemindersQuery'
import { queryClient } from '@/queryClient'
import { useAuthStore } from '@/stores'
import type { Reminder } from '@/types'

let activeEventSource: EventSource | null = null
let activeToken: string | null = null
let activeUserId: string | null = null

export const useRealtimeReminders = () => {
  const token = useAuthStore((state) => state.token)
  const userId = useAuthStore((state) => state.user?.id)

  useEffect(() => {
    // If user is not authenticated, teardown any active connection
    if (!token || !userId) {
      if (activeEventSource) {
        activeEventSource.close()
        activeEventSource = null
        activeToken = null
        activeUserId = null
      }
      return
    }

    // If already connected with the same session, reuse existing connection
    if (
      activeEventSource &&
      activeToken === token &&
      activeUserId === userId &&
      activeEventSource.readyState !== EventSource.CLOSED
    ) {
      return
    }

    // Teardown previous if token or user changed
    if (activeEventSource) {
      activeEventSource.close()
      activeEventSource = null
    }

    activeToken = token
    activeUserId = userId

    // Request browser notification permission if available
    if (typeof window !== 'undefined' && 'Notification' in window) {
      if (Notification.permission === 'default') {
        Notification.requestPermission().catch(() => {})
      }
    }

    const sseUrl = `${apiBaseUrl}/api/v1/realtime/sse/events?token=${encodeURIComponent(token)}&user_id=${encodeURIComponent(userId)}`
    const es = new EventSource(sseUrl)
    activeEventSource = es

    es.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data)
        if (parsed.type === 'reminder:triggered') {
          const reminder = parsed.data as Reminder

          // Invalidate reminder queries to refresh lists and badges
          queryClient.invalidateQueries({ queryKey: reminderKeys.all })

          // Toast alert
          appToast.info(`⏰ Reminder: ${reminder.title}`)

          // Native browser notification
          if (
            typeof window !== 'undefined' &&
            'Notification' in window &&
            Notification.permission === 'granted'
          ) {
            new Notification(reminder.title, {
              body: reminder.description || 'You have a scheduled reminder due now.',
              icon: '/favicon.ico'
            })
          }
        } else if (parsed.type === 'task:due_warning') {
          const taskData = (parsed.data || {}) as { task_id?: number; title?: string }

          // Invalidate task queries so UI and warning badges refresh immediately
          queryClient.invalidateQueries({ queryKey: ['tasks'] })

          const taskTitle = taskData.title || 'Task'
          const toastMsg = `⚠️ Deadline Mendekati: ${taskTitle}`
          const bodyMsg = 'Tugas ini mendekati batas tenggat waktu (due date).'

          appToast.warn(toastMsg)

          if (
            typeof window !== 'undefined' &&
            'Notification' in window &&
            Notification.permission === 'granted'
          ) {
            new Notification(toastMsg, {
              body: bodyMsg,
              icon: '/favicon.ico'
            })
          }
        }
      } catch {
        // Non-JSON SSE ping/keep-alive message
      }
    }

    es.onerror = () => {
      // Reconnection is automatically handled by browser's EventSource
    }
  }, [token, userId])
}
