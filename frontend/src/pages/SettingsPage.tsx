import React, { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { authApi } from '@/api'
import { useAppToast } from '@/components'
import { AvatarCropModal } from '@/features/settings/AvatarCropModal'
import { useTranslation } from '@/i18n'
import { useAuthStore } from '@/stores'

export const SettingsPage: React.FC = () => {
  const { user, updateUser, logout } = useAuthStore()
  const { t } = useTranslation()
  const toast = useAppToast()

  // Profile Form State
  const [name, setName] = useState(user?.name || '')
  const [isSavingProfile, setIsSavingProfile] = useState(false)
  const [isAvatarModalOpen, setIsAvatarModalOpen] = useState(false)

  // Notification & Due Date Alert State
  const [timezone, setTimezone] = useState(user?.timezone || 'Asia/Jakarta')
  const [threshold, setThreshold] = useState<number>(user?.due_date_warning_threshold ?? 1440)
  const [emailEnabled, setEmailEnabled] = useState<boolean>(
    user?.email_notifications_enabled ?? true
  )
  const [isSavingNotifications, setIsSavingNotifications] = useState(false)

  // Danger Zone Delete Modal State
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)
  const [deleteConfirmation, setDeleteConfirmation] = useState('')
  const [isDeletingAccount, setIsDeletingAccount] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const [isBindingGoogle, setIsBindingGoogle] = useState(false)
  const [isBindingApple, setIsBindingApple] = useState(false)

  useEffect(() => {
    if (user) {
      if (user.name) setName(user.name)
      if (user.timezone) setTimezone(user.timezone)
      if (user.due_date_warning_threshold !== undefined)
        setThreshold(user.due_date_warning_threshold)
      if (user.email_notifications_enabled !== undefined)
        setEmailEnabled(user.email_notifications_enabled)
    }
  }, [user])

  // Google Bind Handler
  const handleBindGoogle = () => {
    const googleClientId = import.meta.env.VITE_GOOGLE_CLIENT_ID || ''
    if (typeof window === 'undefined' || !window.google?.accounts?.oauth2) {
      toast.error(t('settings.oauth.googleSdkNotReady'))
      return
    }
    if (!googleClientId) {
      toast.error(t('settings.oauth.googleClientNotConfigured'))
      return
    }

    setIsBindingGoogle(true)
    try {
      const client = window.google.accounts.oauth2.initCodeClient({
        client_id: googleClientId,
        scope: 'openid email profile',
        ux_mode: 'popup',
        callback: async (resp: { code?: string; error?: string }) => {
          setIsBindingGoogle(false)
          if (resp.error) {
            toast.error(t('settings.oauth.googleAuthCancelled'))
            return
          }
          if (resp.code) {
            try {
              const res = await authApi.bindOAuth({
                provider: 'google',
                code: resp.code
              })
              if (res.data) {
                updateUser(res.data)
                toast.success(t('settings.oauth.googleLinkedSuccess'))
              }
            } catch (err: unknown) {
              toast.error(err instanceof Error ? err.message : t('settings.oauth.googleLinkError'))
            }
          }
        },
        // eslint-disable-next-line @typescript-eslint/naming-convention
        error_callback: () => {
          setIsBindingGoogle(false)
        }
      })
      client.requestCode()
    } catch (err: unknown) {
      setIsBindingGoogle(false)
      toast.error(err instanceof Error ? err.message : t('settings.oauth.googleInitError'))
    }
  }

  // Apple Bind Handler
  const handleBindApple = async () => {
    const appleClientId = import.meta.env.VITE_APPLE_CLIENT_ID || ''
    if (typeof window === 'undefined' || !window.AppleID?.auth) {
      toast.error(t('settings.oauth.appleSdkNotReady'))
      return
    }

    setIsBindingApple(true)
    try {
      if (appleClientId) {
        window.AppleID.auth.init({
          clientId: appleClientId,
          scope: 'name email',
          redirectURI: window.location.origin,
          state: 'taskhub-apple-bind',
          usePopup: true
        })
      }
      const res = await window.AppleID.auth.signIn()
      if (res?.authorization) {
        const bindRes = await authApi.bindOAuth({
          provider: 'apple',
          id_token: res.authorization.id_token,
          code: res.authorization.code
        })
        if (bindRes.data) {
          updateUser(bindRes.data)
          toast.success(t('settings.oauth.appleLinkedSuccess'))
        }
      }
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : t('settings.oauth.appleLinkError'))
    } finally {
      setIsBindingApple(false)
    }
  }

  // Handle Unbind OAuth
  const handleUnbind = async (provider: 'google' | 'apple') => {
    try {
      const res = await authApi.unbindOAuth({ provider })
      if (res.data) {
        updateUser(res.data)
        toast.success(
          t('settings.oauth.unbindSuccess', {
            provider: provider === 'google' ? 'Google' : 'Apple'
          })
        )
      }
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : t('settings.oauth.unbindError'))
    }
  }

  // Handle Save Profile
  const handleSaveProfile = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    setIsSavingProfile(true)
    try {
      const res = await authApi.updateProfile({ name: name.trim() })
      if (res.data) {
        updateUser(res.data)
        toast.success(t('settings.profile.updateSuccess'))
      }
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : t('settings.profile.updateError'))
    } finally {
      setIsSavingProfile(false)
    }
  }

  // Handle Save Notification Preferences
  const handleSaveNotificationPreferences = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSavingNotifications(true)
    try {
      const res = await authApi.updateProfile({
        timezone,
        due_date_warning_threshold: threshold,
        email_notifications_enabled: emailEnabled
      })
      if (res.data) {
        updateUser(res.data)
        toast.success(t('settings.notifications.savedSuccess'))
      }
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : t('settings.notifications.savedError'))
    } finally {
      setIsSavingNotifications(false)
    }
  }

  // Handle Account Deletion
  const handleDeleteAccount = async () => {
    if (deleteConfirmation.trim().toUpperCase() !== 'DELETE') {
      setDeleteError(t('settings.dangerZone.errorConfirmation'))
      return
    }

    setIsDeletingAccount(true)
    setDeleteError(null)

    try {
      await authApi.deleteAccount({ confirmation: 'DELETE' })
      toast.success(t('settings.dangerZone.deletedToast'))
      setIsDeleteModalOpen(false)
      logout()
    } catch (err: unknown) {
      setDeleteError(err instanceof Error ? err.message : t('settings.dangerZone.deleteErrorToast'))
      setIsDeletingAccount(false)
    }
  }

  return (
    <div className="space-y-6 pb-12 animate-fade-in w-full">
      {/* Header */}
      <div>
        <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-slate-900 dark:text-white flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 flex items-center justify-center text-lg">
            <i className="pi pi-cog" />
          </div>
          {t('settings.title')}
        </h1>
        <p className="text-xs sm:text-sm text-slate-500 dark:text-slate-400 mt-1">
          {t('settings.subtitle')}
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 items-start">
        {/* Left Column: Profile Information & Linked Accounts */}
        <div className="space-y-6">
          {/* 1. Profile Information */}
          <section className="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm">
            <h2 className="text-base font-semibold text-slate-900 dark:text-white mb-6 flex items-center gap-2">
              <i className="pi pi-user text-indigo-500" />
              {t('settings.profile.title')}
            </h2>

            <div className="flex flex-col sm:flex-row items-center sm:items-start gap-6 pb-6 border-b border-slate-200 dark:border-slate-800">
              {/* Avatar with auto-crop launcher */}
              <div
                className="relative group cursor-pointer"
                onClick={() => setIsAvatarModalOpen(true)}
              >
                <div className="w-24 h-24 rounded-full overflow-hidden border-2 border-indigo-500/20 bg-slate-100 dark:bg-slate-800 flex items-center justify-center shadow-md">
                  {user?.avatar_url ? (
                    <img
                      src={user.avatar_url}
                      alt={user.name}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <span className="text-2xl font-bold text-indigo-600 dark:text-indigo-400">
                      {user?.name?.charAt(0).toUpperCase() || 'U'}
                    </span>
                  )}
                </div>

                {/* Hover Camera Overlay */}
                <div className="absolute inset-0 rounded-full bg-slate-950/60 opacity-0 group-hover:opacity-100 transition-opacity flex flex-col items-center justify-center text-white text-xs gap-1">
                  <i className="pi pi-camera text-base" />
                  <span>{t('settings.profile.changePhoto')}</span>
                </div>

                <div className="absolute bottom-0 right-0 w-7 h-7 rounded-full bg-indigo-600 text-white flex items-center justify-center text-xs shadow-md">
                  <i className="pi pi-pencil text-[10px]" />
                </div>
              </div>

              <div className="flex-1 text-center sm:text-left space-y-1">
                <h3 className="text-base font-bold text-slate-900 dark:text-white">{user?.name}</h3>
                <p className="text-xs text-slate-500 dark:text-slate-400">{user?.email}</p>
                <div className="pt-2 flex flex-wrap gap-2 justify-center sm:justify-start">
                  <button
                    type="button"
                    onClick={() => setIsAvatarModalOpen(true)}
                    data-pr-tooltip={t('settings.profile.uploadCropWebp')}
                    className="px-3 py-1.5 text-xs font-medium rounded-lg border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors flex items-center gap-1.5 cursor-pointer"
                  >
                    <i className="pi pi-upload text-xs" /> {t('settings.profile.uploadCropWebp')}
                  </button>
                </div>
              </div>
            </div>

            {/* Edit Name Form */}
            <form onSubmit={handleSaveProfile} className="mt-6 space-y-4">
              <div>
                <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                  {t('settings.profile.fullName')}
                </label>
                <input
                  type="text"
                  required
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3.5 py-2 text-sm text-slate-900 dark:text-slate-100 focus:outline-none focus:border-indigo-500 transition-colors"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                  {t('settings.profile.email')}
                </label>
                <input
                  type="email"
                  disabled
                  value={user?.email || ''}
                  className="w-full rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-100 dark:bg-slate-950/40 px-3.5 py-2 text-sm text-slate-500 dark:text-slate-400 cursor-not-allowed"
                />
              </div>

              <button
                type="submit"
                disabled={isSavingProfile || name.trim() === user?.name}
                data-pr-tooltip={t('settings.profile.saveChanges')}
                className="px-4 py-2 text-xs font-medium rounded-xl bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white transition-all shadow-sm shadow-indigo-500/20 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 cursor-pointer"
              >
                {isSavingProfile ? (
                  <>
                    <i className="pi pi-spin pi-spinner text-xs" /> {t('settings.profile.saving')}
                  </>
                ) : (
                  t('settings.profile.saveChanges')
                )}
              </button>
            </form>
          </section>

          {/* 2. Linked Accounts (Google & Apple) */}
          <section className="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm">
            <h2 className="text-base font-semibold text-slate-900 dark:text-white mb-2 flex items-center gap-2">
              <i className="pi pi-link text-indigo-500" />
              {t('settings.oauth.title')}
            </h2>
            <p className="text-xs text-slate-500 dark:text-slate-400 mb-6">
              {t('settings.oauth.subtitle')}
            </p>

            <div className="space-y-4">
              {/* Google Card */}
              <div className="flex items-center justify-between p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/30">
                <div className="flex items-center gap-3.5">
                  <div className="w-10 h-10 rounded-xl bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 flex items-center justify-center shadow-xs">
                    <svg className="w-5 h-5" viewBox="0 0 24 24">
                      <path
                        fill="#4285F4"
                        d="M23.745 12.27c0-.7-.06-1.4-.19-2.07H12v4.51h6.6c-.29 1.52-1.14 2.82-2.4 3.68v3.05h3.88c2.27-2.09 3.665-5.17 3.665-9.17z"
                      />
                      <path
                        fill="#34A853"
                        d="M12 24c3.24 0 5.95-1.08 7.93-2.91l-3.88-3.05c-1.08.72-2.45 1.16-4.05 1.16-3.12 0-5.77-2.1-6.72-4.93H1.25v3.15C3.26 21.36 7.33 24 12 24z"
                      />
                      <path
                        fill="#FBBC05"
                        d="M5.28 14.27c-.25-.72-.38-1.49-.38-2.27s.13-1.55.38-2.27V6.58H1.25C.45 8.18 0 9.99 0 12s.45 3.82 1.25 5.42l4.03-3.15z"
                      />
                      <path
                        fill="#EA4335"
                        d="M12 4.75c1.77 0 3.35.61 4.6 1.8l3.42-3.42C17.95 1.19 15.24 0 12 0 7.33 0 3.26 2.64 1.25 6.58l4.03 3.15c.95-2.83 3.6-4.98 6.72-4.98z"
                      />
                    </svg>
                  </div>
                  <div>
                    <h4 className="text-sm font-semibold text-slate-900 dark:text-white">
                      {t('settings.oauth.google')}
                    </h4>
                    <div className="flex items-center gap-1.5 mt-0.5">
                      <span
                        className={`inline-block w-2 h-2 rounded-full ${
                          user?.google_bound ? 'bg-emerald-500' : 'bg-slate-400'
                        }`}
                      />
                      <span className="text-xs text-slate-500 dark:text-slate-400">
                        {user?.google_bound
                          ? t('settings.oauth.connected')
                          : t('settings.oauth.notLinked')}
                      </span>
                    </div>
                  </div>
                </div>

                <div>
                  {user?.google_bound ? (
                    <button
                      type="button"
                      onClick={() => handleUnbind('google')}
                      data-pr-tooltip={t('settings.oauth.disconnect')}
                      className="px-3 py-1.5 text-xs font-medium rounded-lg border border-rose-200 dark:border-rose-900/50 text-rose-600 dark:text-rose-400 hover:bg-rose-50 dark:hover:bg-rose-950/30 transition-colors cursor-pointer"
                    >
                      {t('settings.oauth.disconnect')}
                    </button>
                  ) : (
                    <button
                      type="button"
                      onClick={handleBindGoogle}
                      disabled={isBindingGoogle}
                      data-pr-tooltip={t('settings.oauth.linkGoogle')}
                      className="px-3.5 py-1.5 text-xs font-medium rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white transition-colors flex items-center gap-1.5 cursor-pointer"
                    >
                      {isBindingGoogle ? (
                        <i className="pi pi-spin pi-spinner text-xs" />
                      ) : (
                        <i className="pi pi-plus text-xs" />
                      )}
                      {t('settings.oauth.linkGoogle')}
                    </button>
                  )}
                </div>
              </div>

              {/* Apple Card */}
              <div className="flex items-center justify-between p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/30">
                <div className="flex items-center gap-3.5">
                  <div className="w-10 h-10 rounded-xl bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 flex items-center justify-center shadow-xs">
                    <svg
                      className="w-5 h-5 fill-current text-black dark:text-white"
                      viewBox="0 0 170 170"
                    >
                      <path d="M150.37 130.25c-2.45 5.66-5.35 10.87-8.71 15.66-4.58 6.53-8.33 11.05-11.22 13.56-4.48 4.12-9.28 6.23-14.42 6.35-3.69 0-8.14-1.05-13.32-3.18-5.19-2.12-9.97-3.17-14.34-3.17-4.58 0-9.49 1.05-14.75 3.17-5.26 2.13-9.5 3.24-12.74 3.35-4.35.13-9.16-1.9-14.42-6.08-3.69-3.04-7.67-7.85-11.96-14.41-6.19-9.5-10.97-20.5-14.33-33.01-3.37-12.51-5.06-24.16-5.06-34.94 0-16.75 4.1-30.43 12.3-41.04 8.2-10.62 18.52-16.03 30.98-16.25 4.8 0 10.21 1.25 16.24 3.76 6.03 2.51 10.13 3.82 12.3 3.93 1.74 0 5.86-1.37 12.37-4.11 6.51-2.73 11.9-4 16.16-3.8 11.75.55 21.09 4.79 28.02 12.72-10.45 6.32-15.56 15.24-15.34 26.77.22 9.04 3.74 16.63 10.57 22.77 6.83 6.13 15.02 9.69 24.58 10.67-2.18 6.42-4.9 12.64-8.16 18.66zM119.22 31.86c0-7.39 2.67-14.46 8.01-21.2 5.34-6.75 11.98-10.66 19.92-11.74.33 1.52.49 3.04.49 4.56 0 7.39-2.77 14.53-8.32 21.41-5.55 6.88-12.24 10.74-20.1 11.58z" />
                    </svg>
                  </div>
                  <div>
                    <h4 className="text-sm font-semibold text-slate-900 dark:text-white">
                      {t('settings.oauth.apple')}
                    </h4>
                    <div className="flex items-center gap-1.5 mt-0.5">
                      <span
                        className={`inline-block w-2 h-2 rounded-full ${
                          user?.apple_bound ? 'bg-emerald-500' : 'bg-slate-400'
                        }`}
                      />
                      <span className="text-xs text-slate-500 dark:text-slate-400">
                        {user?.apple_bound
                          ? t('settings.oauth.connected')
                          : t('settings.oauth.notLinked')}
                      </span>
                    </div>
                  </div>
                </div>

                <div>
                  {user?.apple_bound ? (
                    <button
                      type="button"
                      onClick={() => handleUnbind('apple')}
                      data-pr-tooltip={t('settings.oauth.disconnect')}
                      className="px-3 py-1.5 text-xs font-medium rounded-lg border border-rose-200 dark:border-rose-900/50 text-rose-600 dark:text-rose-400 hover:bg-rose-50 dark:hover:bg-rose-950/30 transition-colors cursor-pointer"
                    >
                      {t('settings.oauth.disconnect')}
                    </button>
                  ) : (
                    <button
                      type="button"
                      onClick={handleBindApple}
                      disabled={isBindingApple}
                      data-pr-tooltip={t('settings.oauth.linkApple')}
                      className="px-3.5 py-1.5 text-xs font-medium rounded-lg bg-slate-900 dark:bg-white text-white dark:text-slate-900 hover:bg-slate-800 dark:hover:bg-slate-100 transition-colors flex items-center gap-1.5 cursor-pointer"
                    >
                      {isBindingApple ? (
                        <i className="pi pi-spin pi-spinner text-xs" />
                      ) : (
                        <i className="pi pi-plus text-xs" />
                      )}
                      {t('settings.oauth.linkApple')}
                    </button>
                  )}
                </div>
              </div>
            </div>
          </section>
        </div>

        {/* Right Column: Project Modules Link, MCP & AI Integration, Danger Zone */}
        <div className="space-y-6">
          {/* 3. Project Task Modules Link Card */}
          <section className="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div>
                <h2 className="text-base font-semibold text-slate-900 dark:text-white flex items-center gap-2">
                  <i className="pi pi-folder text-indigo-500" />
                  {t('projects.manageInProjects')}
                </h2>
                <p className="text-xs text-slate-500 dark:text-slate-400 mt-1 max-w-xl">
                  {t('projects.settingsLinkDesc')}
                </p>
              </div>
              <Link
                to="/projects"
                data-pr-tooltip={t('projects.goToProjects')}
                className="px-4 py-2 text-xs font-semibold rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white transition-all shadow-sm shadow-indigo-500/20 inline-flex items-center gap-2 shrink-0 self-start sm:self-auto cursor-pointer"
              >
                <i className="pi pi-arrow-up-right text-xs" />
                {t('projects.goToProjects')}
              </Link>
            </div>
          </section>

          {/* 4. MCP & AI Integration Link Card */}
          <section className="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div>
                <h2 className="text-base font-semibold text-slate-900 dark:text-white flex items-center gap-2">
                  <i className="pi pi-bolt text-indigo-500" />
                  {t('settings.mcp.title')}
                </h2>
                <p className="text-xs text-slate-500 dark:text-slate-400 mt-1 max-w-xl">
                  {t('settings.mcp.subtitle')}
                </p>
              </div>
              <Link
                to="/mcp-settings"
                data-pr-tooltip={t('settings.mcp.button')}
                className="px-4 py-2 text-xs font-semibold rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white transition-all shadow-sm shadow-indigo-500/20 inline-flex items-center gap-2 shrink-0 self-start sm:self-auto cursor-pointer"
              >
                <i className="pi pi-arrow-up-right text-xs" />
                {t('settings.mcp.button')}
              </Link>
            </div>
          </section>

          {/* 5. Notification & Due Date Warning Preferences */}
          <section className="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm">
            <h2 className="text-base font-semibold text-slate-900 dark:text-white mb-2 flex items-center gap-2">
              <i className="pi pi-bell text-indigo-500" />
              {t('settings.notifications.title')}
            </h2>
            <p className="text-xs text-slate-500 dark:text-slate-400 mb-6">
              {t('settings.notifications.subtitle')}
            </p>

            <form onSubmit={handleSaveNotificationPreferences} className="space-y-4">
              {/* Display & Reminder Timezone */}
              <div>
                <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                  {t('settings.notifications.timezoneLabel')}
                </label>
                <select
                  value={timezone}
                  onChange={(e) => setTimezone(e.target.value)}
                  className="w-full rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3.5 py-2 text-sm text-slate-900 dark:text-slate-100 focus:outline-none focus:border-indigo-500 transition-colors"
                >
                  <option value="Asia/Jakarta">Asia/Jakarta (WIB, UTC+7)</option>
                  <option value="Asia/Makassar">Asia/Makassar (WITA, UTC+8)</option>
                  <option value="Asia/Jayapura">Asia/Jayapura (WIT, UTC+9)</option>
                  <option value="Asia/Singapore">Asia/Singapore (SGT, UTC+8)</option>
                  <option value="Asia/Tokyo">Asia/Tokyo (JST, UTC+9)</option>
                  <option value="UTC">UTC (UTC+0)</option>
                  <option value="Europe/London">Europe/London (BST/GMT)</option>
                  <option value="America/New_York">America/New_York (EDT/EST, UTC-4/5)</option>
                  <option value="America/Los_Angeles">
                    America/Los_Angeles (PDT/PST, UTC-7/8)
                  </option>
                </select>
                <p className="text-[11px] text-slate-500 dark:text-slate-400 mt-1">
                  {t('settings.notifications.timezoneHint')}
                </p>
              </div>

              {/* Due Date Alert Threshold */}
              <div>
                <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                  {t('settings.notifications.thresholdLabel')}
                </label>
                <select
                  value={threshold}
                  onChange={(e) => setThreshold(Number(e.target.value))}
                  className="w-full rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 px-3.5 py-2 text-sm text-slate-900 dark:text-slate-100 focus:outline-none focus:border-indigo-500 transition-colors"
                >
                  <option value={15}>{t('settings.notifications.threshold15m')}</option>
                  <option value={60}>{t('settings.notifications.threshold1h')}</option>
                  <option value={180}>{t('settings.notifications.threshold3h')}</option>
                  <option value={720}>{t('settings.notifications.threshold12h')}</option>
                  <option value={1440}>{t('settings.notifications.threshold24h')}</option>
                  <option value={2880}>{t('settings.notifications.threshold48h')}</option>
                </select>
                <p className="text-[11px] text-slate-500 dark:text-slate-400 mt-1">
                  {t('settings.notifications.thresholdHint')}
                </p>
              </div>

              {/* Email Notifications Toggle */}
              <div className="pt-2 border-t border-slate-200 dark:border-slate-800 flex items-start gap-3">
                <input
                  type="checkbox"
                  id="email_notifications_toggle"
                  checked={emailEnabled}
                  onChange={(e) => setEmailEnabled(e.target.checked)}
                  className="mt-0.5 rounded border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-indigo-600 focus:ring-0 cursor-pointer"
                />
                <label
                  htmlFor="email_notifications_toggle"
                  className="cursor-pointer select-none space-y-0.5"
                >
                  <div className="text-xs font-semibold text-slate-900 dark:text-white">
                    {t('settings.notifications.emailNotificationsLabel')}
                  </div>
                  <div className="text-[11px] text-slate-500 dark:text-slate-400">
                    {t('settings.notifications.emailNotificationsDesc')}
                  </div>
                </label>
              </div>

              <div className="pt-2">
                <button
                  type="submit"
                  disabled={
                    isSavingNotifications ||
                    (timezone === (user?.timezone || 'Asia/Jakarta') &&
                      threshold === (user?.due_date_warning_threshold ?? 1440) &&
                      emailEnabled === (user?.email_notifications_enabled ?? true))
                  }
                  data-pr-tooltip={t('settings.notifications.savePreferences')}
                  className="px-4 py-2 text-xs font-medium rounded-xl bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white transition-all shadow-sm shadow-indigo-500/20 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 cursor-pointer"
                >
                  {isSavingNotifications ? (
                    <>
                      <i className="pi pi-spin pi-spinner text-xs" />{' '}
                      {t('settings.notifications.saving')}
                    </>
                  ) : (
                    t('settings.notifications.savePreferences')
                  )}
                </button>
              </div>
            </form>
          </section>

          {/* 6. Danger Zone (Permanent Account Deletion) */}
          <section className="bg-rose-50/40 dark:bg-rose-950/10 rounded-2xl border border-rose-200 dark:border-rose-900/40 p-6">
            <h2 className="text-base font-semibold text-rose-600 dark:text-rose-400 mb-2 flex items-center gap-2">
              <i className="pi pi-exclamation-triangle" />
              {t('settings.dangerZone.title')}
            </h2>
            <p className="text-xs text-slate-600 dark:text-slate-400 mb-6">
              {t('settings.dangerZone.subtitle')}
            </p>

            <button
              type="button"
              onClick={() => {
                setDeleteConfirmation('')
                setDeleteError(null)
                setIsDeleteModalOpen(true)
              }}
              data-pr-tooltip={t('settings.dangerZone.deleteAccountBtn')}
              className="px-4 py-2 text-xs font-semibold rounded-xl bg-rose-600 hover:bg-rose-500 active:bg-rose-700 text-white transition-all shadow-sm shadow-rose-500/20 flex items-center gap-1.5 cursor-pointer"
            >
              <i className="pi pi-trash text-xs" />
              {t('settings.dangerZone.deleteAccountBtn')}
            </button>
          </section>
        </div>
      </div>

      {/* Avatar Crop Modal */}
      <AvatarCropModal
        isOpen={isAvatarModalOpen}
        onClose={() => setIsAvatarModalOpen(false)}
        onSuccess={() => {
          toast.success(t('settings.avatarModal.uploadSuccess'))
        }}
      />

      {/* Danger Zone Confirmation Modal */}
      {isDeleteModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-md animate-fade-in">
          <div className="relative w-full max-w-md rounded-2xl border border-rose-500/30 bg-white dark:bg-slate-900 p-6 shadow-2xl text-slate-900 dark:text-slate-100">
            <div className="w-12 h-12 rounded-2xl bg-rose-500/10 text-rose-600 dark:text-rose-400 flex items-center justify-center text-xl mb-4">
              <i className="pi pi-exclamation-triangle" />
            </div>

            <h3 className="text-base font-bold text-slate-900 dark:text-white">
              {t('settings.dangerZone.modalTitle')}
            </h3>
            <p className="text-xs text-slate-600 dark:text-slate-400 mt-2 leading-relaxed">
              {t('settings.dangerZone.modalDesc')}
            </p>

            <div className="mt-4 p-3 rounded-xl bg-slate-100 dark:bg-slate-950 border border-slate-200 dark:border-slate-800">
              <p className="text-xs text-slate-700 dark:text-slate-300 font-medium mb-1.5">
                {t('settings.dangerZone.typeDeletePrompt')}
              </p>
              <input
                type="text"
                autoFocus
                value={deleteConfirmation}
                onChange={(e) => {
                  setDeleteConfirmation(e.target.value)
                  setDeleteError(null)
                }}
                placeholder="DELETE"
                className="w-full rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 px-3 py-1.5 text-xs font-mono font-bold tracking-widest text-slate-900 dark:text-slate-100 focus:outline-none focus:border-rose-500"
              />
            </div>

            {deleteError && (
              <p className="text-xs text-rose-500 mt-2 font-medium flex items-center gap-1">
                <i className="pi pi-times-circle" /> {deleteError}
              </p>
            )}

            <div className="mt-6 flex justify-end gap-2.5">
              <button
                type="button"
                onClick={() => setIsDeleteModalOpen(false)}
                disabled={isDeletingAccount}
                data-pr-tooltip={t('settings.dangerZone.cancelBtn')}
                className="px-4 py-2 text-xs font-medium rounded-xl border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 cursor-pointer"
              >
                {t('settings.dangerZone.cancelBtn')}
              </button>
              <button
                type="button"
                onClick={handleDeleteAccount}
                disabled={deleteConfirmation.trim().toUpperCase() !== 'DELETE' || isDeletingAccount}
                data-pr-tooltip={t('settings.dangerZone.confirmDeleteBtn')}
                className="px-4 py-2 text-xs font-bold rounded-xl bg-rose-600 hover:bg-rose-500 text-white disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 cursor-pointer"
              >
                {isDeletingAccount ? (
                  <>
                    <i className="pi pi-spin pi-spinner text-xs" />{' '}
                    {t('settings.dangerZone.deleting')}
                  </>
                ) : (
                  t('settings.dangerZone.confirmDeleteBtn')
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
export default SettingsPage
