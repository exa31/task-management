import React, { useState } from 'react'

import { useTranslation } from '@/i18n'

interface OAuthSetupModalProps {
  isOpen: boolean
  onClose: () => void
  provider: 'google' | 'apple'
  onUseDemoLogin?: () => void
}

export const OAuthSetupModal: React.FC<OAuthSetupModalProps> = ({
  isOpen,
  onClose,
  provider,
  onUseDemoLogin
}) => {
  const { t } = useTranslation()
  const [copiedSnippet, setCopiedSnippet] = useState(false)

  if (!isOpen) return null

  const isGoogle = provider === 'google'

  const envSnippet = isGoogle
    ? `# frontend/.env
VITE_GOOGLE_CLIENT_ID=YOUR_GOOGLE_CLIENT_ID.apps.googleusercontent.com

# backend/.env
GOOGLE_CLIENT_ID=YOUR_GOOGLE_CLIENT_ID.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=YOUR_GOOGLE_CLIENT_SECRET
GOOGLE_REDIRECT_URI=postmessage`
    : `# frontend/.env
VITE_APPLE_CLIENT_ID=com.yourbrand.app.service

# backend/.env
APPLE_CLIENT_ID=com.yourbrand.app.service`

  const handleCopy = () => {
    navigator.clipboard.writeText(envSnippet)
    setCopiedSnippet(true)
    setTimeout(() => setCopiedSnippet(false), 2000)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/50 dark:bg-slate-950/80 backdrop-blur-sm animate-fade-in">
      <div className="relative w-full max-w-lg rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-6 shadow-2xl text-slate-900 dark:text-slate-100">
        {/* Modal Header */}
        <div className="flex items-center justify-between pb-4 border-b border-slate-200 dark:border-slate-800">
          <div className="flex items-center gap-2.5">
            <div className="h-8 w-8 rounded-xl bg-indigo-50 dark:bg-indigo-950/60 border border-indigo-200 dark:border-indigo-800 flex items-center justify-center text-indigo-600 dark:text-indigo-400">
              <i className={`pi ${isGoogle ? 'pi-google' : 'pi-apple'} text-sm`} />
            </div>
            <div>
              <h2 className="text-sm font-bold text-slate-900 dark:text-slate-100">
                {isGoogle ? 'Setup Google OAuth Console' : 'Setup Apple Developer Console'}
              </h2>
              <p className="text-[11px] text-slate-500 dark:text-slate-400">
                Client ID belum terdeteksi di berkas{' '}
                <code className="font-mono text-indigo-600 dark:text-indigo-400">.env</code>
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
            title={t('common.close')}
          >
            <i className="pi pi-times text-xs" />
          </button>
        </div>

        {/* Modal Body */}
        <div className="my-4 space-y-3.5 text-xs">
          {isGoogle ? (
            <>
              <div className="space-y-2 text-slate-600 dark:text-slate-300">
                <p className="font-semibold text-slate-800 dark:text-slate-200">
                  Langkah konfigurasi Google Cloud Console:
                </p>
                <ol className="list-decimal list-inside space-y-1 text-[11px] text-slate-600 dark:text-slate-400">
                  <li>
                    Buka{' '}
                    <a
                      href="https://console.cloud.google.com/apis/credentials"
                      target="_blank"
                      rel="noreferrer"
                      className="text-indigo-600 dark:text-indigo-400 underline font-medium"
                    >
                      Google Cloud Console &gt; Credentials
                    </a>
                  </li>
                  <li>
                    Buat kredensial <strong>OAuth 2.0 Client ID</strong> tipe{' '}
                    <strong>Web application</strong>.
                  </li>
                  <li>
                    Di bagian <strong>Authorized JavaScript origins</strong>, tambahkan:{' '}
                    <code className="px-1 py-0.5 rounded bg-slate-100 dark:bg-slate-800 font-mono text-indigo-600 dark:text-indigo-400">
                      {typeof window !== 'undefined' ? window.location.origin : 'http://localhost:5173'}
                    </code>
                  </li>
                  <li>
                    Salin Client ID dan Client Secret ke berkas{' '}
                    <code className="font-mono">.env</code> berikut:
                  </li>
                </ol>
              </div>

              {/* Code snippet */}
              <div className="relative rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 p-3 font-mono text-[11px] text-slate-700 dark:text-slate-300 overflow-x-auto">
                <button
                  onClick={handleCopy}
                  className="absolute top-2 right-2 px-2 py-1 rounded bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-[10px] text-slate-600 dark:text-slate-300 hover:text-indigo-600 transition-colors flex items-center gap-1 cursor-pointer"
                >
                  <i className={`pi ${copiedSnippet ? 'pi-check text-emerald-500' : 'pi-copy'}`} />
                  {copiedSnippet ? t('common.copied') : 'Copy'}
                </button>
                <pre>{envSnippet}</pre>
              </div>
            </>
          ) : (
            <>
              <div className="space-y-2 text-slate-600 dark:text-slate-300">
                <p className="font-semibold text-slate-800 dark:text-slate-200">
                  Langkah konfigurasi Apple Developer Account:
                </p>
                <ol className="list-decimal list-inside space-y-1 text-[11px] text-slate-600 dark:text-slate-400">
                  <li>
                    Buka{' '}
                    <a
                      href="https://developer.apple.com/account/resources/identifiers/list/serviceId"
                      target="_blank"
                      rel="noreferrer"
                      className="text-indigo-600 dark:text-indigo-400 underline font-medium"
                    >
                      Apple Developer &gt; Services IDs
                    </a>
                  </li>
                  <li>
                    Buat <strong>Services ID</strong> baru dan aktifkan{' '}
                    <strong>Sign in with Apple</strong>.
                  </li>
                  <li>Daftarkan domain dan redirect URL web Anda.</li>
                  <li>
                    Salin Services ID ke berkas <code className="font-mono">.env</code> berikut:
                  </li>
                </ol>
              </div>

              <div className="relative rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 p-3 font-mono text-[11px] text-slate-700 dark:text-slate-300 overflow-x-auto">
                <button
                  onClick={handleCopy}
                  className="absolute top-2 right-2 px-2 py-1 rounded bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-[10px] text-slate-600 dark:text-slate-300 hover:text-indigo-600 transition-colors flex items-center gap-1 cursor-pointer"
                >
                  <i className={`pi ${copiedSnippet ? 'pi-check text-emerald-500' : 'pi-copy'}`} />
                  {copiedSnippet ? t('common.copied') : 'Copy'}
                </button>
                <pre>{envSnippet}</pre>
              </div>
            </>
          )}
        </div>

        {/* Modal Footer */}
        <div className="pt-3 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between">
          {onUseDemoLogin ? (
            <button
              onClick={() => {
                onClose()
                onUseDemoLogin()
              }}
              className="text-xs text-indigo-600 dark:text-indigo-400 hover:underline cursor-pointer flex items-center gap-1 font-medium"
            >
              <i className="pi pi-bolt text-xs" />
              Gunakan Login Akun Demo Sementara
            </button>
          ) : (
            <span />
          )}

          <button
            onClick={onClose}
            className="px-3.5 py-1.5 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs font-semibold text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors cursor-pointer"
          >
            {t('common.close')}
          </button>
        </div>
      </div>
    </div>
  )
}

export default OAuthSetupModal
