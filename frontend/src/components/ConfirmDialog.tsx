import React, { useEffect } from 'react'

export { ConfirmDialog as PrimeConfirmDialog } from 'primereact/confirmdialog'

export interface CustomConfirmDialogProps {
  isOpen: boolean
  title?: string
  message: string | React.ReactNode
  confirmLabel?: string
  cancelLabel?: string
  severity?: 'danger' | 'warning' | 'info'
  isLoading?: boolean
  onConfirm: () => void | Promise<void>
  onCancel: () => void
}

const SEVERITY_CONFIG = {
  danger: {
    icon: 'pi-trash',
    iconBg: 'bg-rose-500/10 text-rose-400 border-rose-500/20',
    confirmBtn: 'bg-rose-600 hover:bg-rose-500 text-white shadow-rose-950/50',
  },
  warning: {
    icon: 'pi-exclamation-triangle',
    iconBg: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
    confirmBtn: 'bg-amber-600 hover:bg-amber-500 text-white shadow-amber-950/50',
  },
  info: {
    icon: 'pi-info-circle',
    iconBg: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20',
    confirmBtn: 'bg-indigo-600 hover:bg-indigo-500 text-white shadow-indigo-950/50',
  },
}

export const ConfirmDialog: React.FC<CustomConfirmDialogProps> = ({
  isOpen,
  title = 'Confirmation Required',
  message,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  severity = 'danger',
  isLoading = false,
  onConfirm,
  onCancel,
}) => {
  useEffect(() => {
    if (!isOpen) return
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !isLoading) {
        onCancel()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, isLoading, onCancel])

  if (!isOpen) return null

  const config = SEVERITY_CONFIG[severity] || SEVERITY_CONFIG.danger

  return (
    <div
      role="dialog"
      aria-modal="true"
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/45 dark:bg-slate-950/75 backdrop-blur-sm animate-fade-in"
    >
      <div
        className="fixed inset-0"
        onClick={() => {
          if (!isLoading) onCancel()
        }}
      />
      <div className="relative w-full max-w-md rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-2xl z-10 animate-scale-in text-slate-900 dark:text-slate-100">
        <div className="flex items-start gap-4">
          <div
            className={`w-10 h-10 rounded-xl border flex items-center justify-center shrink-0 ${config.iconBg}`}
          >
            <i className={`pi ${config.icon} text-base`} />
          </div>

          <div className="flex-1 min-w-0">
            <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-1 leading-snug">
              {title}
            </h3>
            <div className="text-xs text-slate-600 dark:text-slate-400 leading-relaxed">
              {typeof message === 'string' ? <p>{message}</p> : message}
            </div>
          </div>
        </div>

        <div className="mt-6 flex items-center justify-end gap-2.5">
          <button
            type="button"
            disabled={isLoading}
            onClick={onCancel}
            className="px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-100 dark:bg-slate-800/80 hover:bg-slate-200 dark:hover:bg-slate-700 text-xs font-medium text-slate-700 dark:text-slate-300 disabled:opacity-50 transition-colors cursor-pointer"
          >
            {cancelLabel}
          </button>
          <button
            type="button"
            disabled={isLoading}
            onClick={onConfirm}
            className={`px-3.5 py-1.5 rounded-lg text-xs font-medium shadow-sm flex items-center gap-1.5 disabled:opacity-50 transition-colors cursor-pointer ${config.confirmBtn}`}
          >
            {isLoading && <i className="pi pi-spin pi-spinner text-xs" />}
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}
