import React, { useCallback, useEffect, useRef, useState } from 'react'

import { authApi } from '@/api'
import { useTranslation } from '@/i18n'
import { useAuthStore } from '@/stores'

interface AvatarCropModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess?: (avatarUrl: string) => void
}

export const AvatarCropModal: React.FC<AvatarCropModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const { t } = useTranslation()
  const { updateUser } = useAuthStore()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const imageRef = useRef<HTMLImageElement | null>(null)
  const previewCanvasRef = useRef<HTMLCanvasElement | null>(null)

  const [imageSrc, setImageSrc] = useState<string | null>(null)
  const [zoom, setZoom] = useState(1)
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const [isDragging, setIsDragging] = useState(false)
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 })
  const [uploading, setUploading] = useState(false)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  // Reset state when opening/closing
  useEffect(() => {
    if (!isOpen) {
      setImageSrc(null)
      setZoom(1)
      setPosition({ x: 0, y: 0 })
      setErrorMsg(null)
    }
  }, [isOpen])

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    if (!file.type.startsWith('image/')) {
      setErrorMsg(t('settings.avatarModal.invalidFileType'))
      return
    }

    if (file.size > 10 * 1024 * 1024) {
      setErrorMsg(t('settings.avatarModal.fileTooLarge'))
      return
    }

    setErrorMsg(null)
    const reader = new FileReader()
    reader.onload = () => {
      setImageSrc(reader.result as string)
      setZoom(1)
      setPosition({ x: 0, y: 0 })
    }
    reader.readAsDataURL(file)
  }

  // Draw crop preview onto offscreen canvas
  const drawPreview = useCallback(() => {
    const canvas = previewCanvasRef.current
    const img = imageRef.current
    if (!canvas || !img || !img.complete) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const size = 512
    canvas.width = size
    canvas.height = size

    ctx.clearRect(0, 0, size, size)

    // Compute dimensions
    const imgAspect = img.naturalWidth / img.naturalHeight
    let renderW = size
    let renderH = size

    if (imgAspect > 1) {
      renderH = size
      renderW = size * imgAspect
    } else {
      renderW = size
      renderH = size / imgAspect
    }

    renderW *= zoom
    renderH *= zoom

    const centerX = size / 2 + position.x
    const centerY = size / 2 + position.y

    ctx.save()
    // Smooth image rendering
    ctx.imageSmoothingEnabled = true
    ctx.imageSmoothingQuality = 'high'

    ctx.drawImage(
      img,
      centerX - renderW / 2,
      centerY - renderH / 2,
      renderW,
      renderH
    )
    ctx.restore()
  }, [zoom, position])

  useEffect(() => {
    if (imageSrc) {
      const img = new Image()
      img.src = imageSrc
      img.onload = () => {
        imageRef.current = img
        drawPreview()
      }
    }
  }, [imageSrc, drawPreview])

  useEffect(() => {
    if (imageRef.current) {
      drawPreview()
    }
  }, [zoom, position, drawPreview])

  const handleMouseDown = (e: React.MouseEvent) => {
    if (!imageSrc) return
    setIsDragging(true)
    setDragStart({ x: e.clientX - position.x, y: e.clientY - position.y })
  }

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!isDragging) return
    setPosition({
      x: e.clientX - dragStart.x,
      y: e.clientY - dragStart.y,
    })
  }

  const handleMouseUp = () => {
    setIsDragging(false)
  }

  const handleUpload = async () => {
    const canvas = previewCanvasRef.current
    if (!canvas) return

    setUploading(true)
    setErrorMsg(null)

    try {
      // Export canvas to WebP blob
      const blob = await new Promise<Blob | null>((resolve) => {
        canvas.toBlob((b) => resolve(b), 'image/webp', 0.92)
      })

      if (!blob) {
        throw new Error(t('settings.avatarModal.convertError'))
      }

      const file = new File([blob], `avatar-${Date.now()}.webp`, {
        type: 'image/webp',
      })

      const res = await authApi.uploadAvatar(file)
      if (res.data) {
        updateUser(res.data)
        if (onSuccess && res.data.avatar_url) {
          onSuccess(res.data.avatar_url)
        }
      }
      onClose()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t('settings.avatarModal.uploadError')
      setErrorMsg(msg)
    } finally {
      setUploading(false)
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/60 backdrop-blur-md animate-fade-in">
      <div className="relative w-full max-w-md rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-2xl text-slate-900 dark:text-slate-100">
        {/* Header */}
        <div className="flex items-center justify-between pb-4 border-b border-slate-200 dark:border-slate-800">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 flex items-center justify-center font-semibold text-sm">
              <i className="pi pi-camera" />
            </div>
            <div>
              <h2 className="text-base font-semibold leading-none">{t('settings.avatarModal.title')}</h2>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                {t('settings.avatarModal.subtitle')}
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            <i className="pi pi-times text-sm" />
          </button>
        </div>

        {errorMsg && (
          <div className="mt-4 flex items-start gap-2 rounded-lg border border-rose-500/20 bg-rose-500/10 p-3 text-xs text-rose-600 dark:text-rose-400">
            <i className="pi pi-exclamation-circle text-sm mt-0.5 shrink-0" />
            <span className="flex-1 font-medium">{errorMsg}</span>
          </div>
        )}

        {/* Content */}
        <div className="mt-5 flex flex-col items-center">
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            onChange={handleFileChange}
            className="hidden"
          />

          {!imageSrc ? (
            <div
              onClick={() => fileInputRef.current?.click()}
              className="w-full h-64 border-2 border-dashed border-slate-300 dark:border-slate-700 hover:border-indigo-500 dark:hover:border-indigo-400 rounded-2xl flex flex-col items-center justify-center cursor-pointer bg-slate-50 dark:bg-slate-950/50 hover:bg-indigo-50/30 dark:hover:bg-indigo-950/20 transition-all group"
            >
              <div className="w-14 h-14 rounded-full bg-slate-200 dark:bg-slate-800 flex items-center justify-center text-slate-500 dark:text-slate-400 group-hover:scale-110 group-hover:bg-indigo-500 group-hover:text-white transition-all">
                <i className="pi pi-cloud-upload text-2xl" />
              </div>
              <p className="mt-3 text-sm font-medium text-slate-700 dark:text-slate-300">
                {t('settings.avatarModal.uploadPrompt')}
              </p>
              <p className="text-xs text-slate-400 mt-1">{t('settings.avatarModal.supportedFormats')}</p>
            </div>
          ) : (
            <div className="w-full flex flex-col items-center">
              {/* Interactive Crop Viewport */}
              <div
                className="relative w-64 h-64 rounded-2xl overflow-hidden bg-slate-950 border border-slate-700 shadow-inner cursor-grab active:cursor-grabbing select-none"
                onMouseDown={handleMouseDown}
                onMouseMove={handleMouseMove}
                onMouseUp={handleMouseUp}
                onMouseLeave={handleMouseUp}
              >
                {/* Offscreen Canvas positioned to show live preview */}
                <canvas
                  ref={previewCanvasRef}
                  className="w-full h-full object-cover"
                />

                {/* Circular overlay mask guide */}
                <div className="absolute inset-0 pointer-events-none rounded-full border-2 border-white/80 shadow-[0_0_0_9999px_rgba(0,0,0,0.6)]" />
              </div>

              {/* Controls */}
              <div className="w-full mt-4 space-y-3">
                <div className="flex items-center gap-3">
                  <i className="pi pi-search-minus text-xs text-slate-400" />
                  <input
                    type="range"
                    min="1"
                    max="3"
                    step="0.05"
                    value={zoom}
                    onChange={(e) => setZoom(parseFloat(e.target.value))}
                    className="flex-1 accent-indigo-500 cursor-pointer h-1.5 bg-slate-200 dark:bg-slate-700 rounded-lg appearance-none"
                  />
                  <i className="pi pi-search-plus text-xs text-slate-400" />
                  <span className="text-xs font-mono text-slate-500 dark:text-slate-400 w-10 text-right">
                    {Math.round(zoom * 100)}%
                  </span>
                </div>

                <div className="flex justify-between items-center pt-1 text-xs">
                  <button
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                    className="text-indigo-600 dark:text-indigo-400 hover:underline flex items-center gap-1 font-medium"
                  >
                    <i className="pi pi-sync text-[10px]" /> {t('settings.avatarModal.changePhoto')}
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setZoom(1)
                      setPosition({ x: 0, y: 0 })
                    }}
                    className="text-slate-500 hover:text-slate-700 dark:hover:text-slate-300"
                  >
                    {t('settings.avatarModal.resetPosition')}
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Footer Actions */}
        <div className="mt-6 flex justify-end gap-2.5 pt-4 border-t border-slate-200 dark:border-slate-800">
          <button
            type="button"
            onClick={onClose}
            disabled={uploading}
            className="px-4 py-2 text-xs font-medium rounded-xl border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            {t('settings.avatarModal.cancel')}
          </button>
          <button
            type="button"
            onClick={handleUpload}
            disabled={!imageSrc || uploading}
            className="px-5 py-2 text-xs font-medium rounded-xl bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white transition-all shadow-md shadow-indigo-500/20 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {uploading ? (
              <>
                <i className="pi pi-spin pi-spinner text-xs" />
                {t('settings.avatarModal.saving')}
              </>
            ) : (
              <>
                <i className="pi pi-check text-xs" />
                {t('settings.avatarModal.save')}
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  )
}
