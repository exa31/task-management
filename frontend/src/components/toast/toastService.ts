import type { Toast } from 'primereact/toast'

import { getLocalizedErrorMessage } from '@/api/errors'
import { useTranslation } from '@/i18n'

interface ToastOptions {
  severity?: 'success' | 'info' | 'warn' | 'error'
  summary?: string
  detail?: string
  life?: number
}

let toastInstance: Toast | null = null
const pendingQueue: ToastOptions[] = []

export const appToast = {
  setInstance: (ref: Toast | null) => {
    toastInstance = ref
    if (ref && pendingQueue.length > 0) {
      while (pendingQueue.length > 0) {
        const item = pendingQueue.shift()
        if (item) {
          ref.show(item)
        }
      }
    }
  },
  show: (options: ToastOptions) => {
    if (toastInstance) {
      toastInstance.show(options)
    } else {
      pendingQueue.push(options)
    }
  },
  success: (detail: string, summary: string = 'Success') => {
    appToast.show({ severity: 'success', summary, detail, life: 3000 })
  },
  error: (detail: string, summary: string = 'Error') => {
    appToast.show({ severity: 'error', summary, detail, life: 4000 })
  },
  info: (detail: string, summary: string = 'Info') => {
    appToast.show({ severity: 'info', summary, detail, life: 3000 })
  },
  warn: (detail: string, summary: string = 'Warning') => {
    appToast.show({ severity: 'warn', summary, detail, life: 3500 })
  }
}

if (typeof window !== 'undefined') {
  ;(window as unknown as { taskHubToast: typeof appToast }).taskHubToast = appToast
}

export function useAppToast() {
  const { t } = useTranslation()

  return {
    success: (detail: string, summary?: string) => {
      appToast.success(detail, summary || t('common.success'))
    },
    error: (errOrDetail: unknown, summary?: string) => {
      const msg =
        typeof errOrDetail === 'string' ? errOrDetail : getLocalizedErrorMessage(errOrDetail, t)
      appToast.error(msg, summary || t('errors.defaultTitle'))
    },
    info: (detail: string, summary: string = 'Info') => {
      appToast.info(detail, summary)
    },
    warn: (detail: string, summary?: string) => {
      appToast.warn(detail, summary || 'Warning')
    }
  }
}
