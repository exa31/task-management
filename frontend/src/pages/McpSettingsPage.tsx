import React, { useState } from 'react'

import { confirmDialog } from '@/components'
import { useCreateMcpToken, useDeleteMcpToken, useMcpTokens } from '@/hooks'
import { useTranslation } from '@/i18n'
import type { GenerateTokenResponse } from '@/types'

export const McpSettingsPage: React.FC = () => {
  const { data: tokens = [], isLoading, isError, error, refetch } = useMcpTokens()
  const createTokenMutation = useCreateMcpToken()
  const deleteTokenMutation = useDeleteMcpToken()
  const { t, locale } = useTranslation()

  const [newTokenName, setNewTokenName] = useState('')
  const [generatedToken, setGeneratedToken] = useState<GenerateTokenResponse | null>(null)
  const [copiedKey, setCopiedKey] = useState<string | null>(null)

  const handleGenerateToken = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newTokenName.trim()) return

    const res = await createTokenMutation.mutateAsync({ name: newTokenName.trim() })
    setGeneratedToken(res)
    setNewTokenName('')
  }

  const handleDeleteToken = (id: number, name?: string) => {
    confirmDialog({
      header: t('mcp.revokeConfirmHeader'),
      message: t('mcp.revokeConfirmMessage', { name: name || 'this token' }),
      icon: 'pi pi-exclamation-triangle text-rose-500',
      defaultFocus: 'reject',
      acceptClassName: 'p-button-danger',
      acceptLabel: t('mcp.revokeAccess'),
      rejectLabel: t('common.cancel'),
      accept: async () => {
        await deleteTokenMutation.mutateAsync(id)
      }
    })
  }

  const copyToClipboard = (text: string, key: string) => {
    navigator.clipboard.writeText(text)
    setCopiedKey(key)
    setTimeout(() => setCopiedKey(null), 2000)
  }

  const activeTokenSecret = generatedToken
    ? generatedToken.raw_token
    : '<YOUR_PERSONAL_ACCESS_TOKEN>'

  const antigravitySnippet = `{
  "mcpServers": {
    "task-management": {
      "url": "http://localhost:3000/api/v1/mcp",
      "headers": {
        "Authorization": "Bearer ${activeTokenSecret}"
      }
    }
  }
}`

  const cursorSnippet = `{
  "mcpServers": {
    "task-hub": {
      "url": "http://localhost:3000/api/v1/mcp",
      "headers": {
        "Authorization": "Bearer ${activeTokenSecret}"
      }
    }
  }
}`

  const claudeDesktopSnippet = `{
  "mcpServers": {
    "task-management": {
      "command": "curl",
      "args": [
        "-X", "POST",
        "http://localhost:3000/api/v1/mcp",
        "-H", "Authorization: Bearer ${activeTokenSecret}",
        "-H", "Content-Type: application/json"
      ]
    }
  }
}`

  return (
    <div className="space-y-6 pb-16 animate-fade-in w-full">
      {/* Header */}
      <div>
        <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-xl bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 flex items-center justify-center text-sm">
            <i className="pi pi-code" />
          </div>
          {t('mcp.title')}
        </h1>
        <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">{t('mcp.subtitle')}</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 items-start">
        {/* Left Column: Capabilities & Token Management */}
        <div className="space-y-6">
          {/* Feature capabilities banner */}
          <div className="rounded-xl border border-indigo-200 dark:border-indigo-500/30 bg-indigo-50/70 dark:bg-indigo-950/20 p-4 shadow-xs">
            <h3 className="text-xs font-semibold text-indigo-900 dark:text-indigo-300 mb-2 flex items-center gap-1.5">
              <i className="pi pi-sparkles text-xs" />
              {t('mcp.capabilitiesTitle')}
            </h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs text-slate-700 dark:text-slate-300">
              <div className="flex items-center gap-2">
                <i className="pi pi-check text-[10px] text-emerald-500 dark:text-emerald-400" />
                <span>
                  <strong className="text-slate-900 dark:text-slate-100">create_idea_note</strong>:{' '}
                  {t('mcp.createIdeaNote')}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <i className="pi pi-check text-[10px] text-emerald-500 dark:text-emerald-400" />
                <span>
                  <strong className="text-slate-900 dark:text-slate-100">
                    convert_note_to_task
                  </strong>
                  : {t('mcp.convertNoteToTask')}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <i className="pi pi-check text-[10px] text-emerald-500 dark:text-emerald-400" />
                <span>
                  <strong className="text-slate-900 dark:text-slate-100">
                    create_task & list_tasks
                  </strong>
                  : {t('mcp.createTaskAndList')}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <i className="pi pi-check text-[10px] text-emerald-500 dark:text-emerald-400" />
                <span>
                  <strong className="text-slate-900 dark:text-slate-100">log_task_progress</strong>:{' '}
                  {t('mcp.logProgress')}
                </span>
              </div>
            </div>
          </div>

          {/* Generated Token Warning Banner */}
          {generatedToken && (
            <div className="rounded-xl border border-emerald-500/30 dark:border-emerald-500/50 bg-emerald-50 dark:bg-emerald-950/40 p-4 shadow-sm dark:shadow-lg animate-fade-in">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-xs font-semibold text-emerald-800 dark:text-emerald-300 flex items-center gap-1.5">
                    <i className="pi pi-check-circle text-xs" />
                    {t('mcp.newTokenCreated', { name: generatedToken.name })}
                  </p>
                  <p className="text-[11px] text-emerald-700/80 dark:text-slate-300 mt-1">
                    {t('mcp.copyWarning')}
                  </p>
                  <div className="mt-2.5 flex items-center gap-2 bg-white dark:bg-slate-950 p-2 rounded-lg border border-emerald-200 dark:border-slate-800 font-mono text-xs text-indigo-600 dark:text-indigo-300">
                    <span className="truncate select-all">{generatedToken.raw_token}</span>
                    <button
                      onClick={() => copyToClipboard(generatedToken.raw_token, 'new_token')}
                      className="px-2 py-1 rounded bg-indigo-600 hover:bg-indigo-500 text-white text-[10px] shrink-0 font-sans cursor-pointer"
                    >
                      {copiedKey === 'new_token' ? t('common.copied') : t('mcp.copyToken')}
                    </button>
                  </div>
                </div>
                <button
                  onClick={() => setGeneratedToken(null)}
                  className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 cursor-pointer"
                >
                  <i className="pi pi-times text-xs" />
                </button>
              </div>
            </div>
          )}

          {/* Generate Token Section */}
          <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs">
            <h2 className="text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-300 mb-3">
              {t('mcp.generatePatTitle')}
            </h2>
            <form onSubmit={handleGenerateToken} className="space-y-2">
              <label className="block text-xs font-medium text-slate-700 dark:text-slate-300">
                {t('mcp.tokenNameLabel')} <span className="text-rose-500 font-bold ml-0.5">*</span>
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  required
                  value={newTokenName}
                  onChange={(e) => setNewTokenName(e.target.value)}
                  placeholder={t('mcp.tokenNamePlaceholder')}
                  className="flex-1 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 px-3 py-2 text-xs text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:border-indigo-500 focus:outline-none transition-colors"
                />
                <button
                  type="submit"
                  disabled={createTokenMutation.isPending || !newTokenName.trim()}
                  className="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-xs font-medium text-white transition-colors flex items-center gap-1.5 shrink-0 shadow-xs cursor-pointer"
                >
                  {createTokenMutation.isPending ? (
                    <i className="pi pi-spin pi-spinner text-xs" />
                  ) : (
                    <i className="pi pi-key text-xs" />
                  )}
                  {t('mcp.generateTokenBtn')}
                </button>
              </div>
            </form>

            {/* List of active tokens */}
            <div className="mt-5 border-t border-slate-100 dark:border-slate-800/80 pt-4">
              <div className="flex items-center justify-between mb-2">
                <h3 className="text-xs font-medium text-slate-500 dark:text-slate-400">
                  {t('mcp.activeTokens', { count: tokens.length })}
                </h3>
                {isError && (
                  <button
                    onClick={() => refetch()}
                    className="text-[11px] text-rose-500 dark:text-rose-400 hover:text-rose-600 dark:hover:text-rose-300 flex items-center gap-1 cursor-pointer"
                  >
                    <i className="pi pi-refresh text-[10px]" /> {t('common.retry')}
                  </button>
                )}
              </div>

              {isLoading ? (
                <div className="py-6 text-center text-slate-500 text-xs flex items-center justify-center gap-2">
                  <i className="pi pi-spin pi-spinner text-indigo-400" /> {t('common.loading')}
                </div>
              ) : isError ? (
                <div className="py-6 text-center rounded-lg border border-rose-200 dark:border-rose-900/40 bg-rose-50 dark:bg-rose-950/20 p-3">
                  <p className="text-xs text-rose-600 dark:text-rose-300 font-medium">
                    {t('mcp.failedLoadTokens')}
                  </p>
                  <p className="text-[10px] text-rose-500 dark:text-rose-400/80 mt-0.5">
                    {(error as Error)?.message ||
                      'An error occurred while connecting to the API server.'}
                  </p>
                </div>
              ) : (
                <div className="divide-y divide-slate-100 dark:divide-slate-800/60">
                  {tokens.map((tok) => (
                    <div
                      key={tok.id}
                      className="py-2.5 flex items-center justify-between text-xs gap-3"
                    >
                      <div>
                        <p className="font-medium text-slate-800 dark:text-slate-200">{tok.name}</p>
                        <p className="text-[10px] font-mono text-slate-500 mt-0.5">
                          {t('mcp.prefix')}{' '}
                          <span className="text-indigo-600 dark:text-indigo-400">
                            {tok.token_prefix}...
                          </span>{' '}
                          • {t('mcp.createdAt')}{' '}
                          {new Date(tok.created_at).toLocaleDateString(
                            locale === 'id' ? 'id-ID' : 'en-US'
                          )}
                        </p>
                      </div>
                      <button
                        onClick={() => handleDeleteToken(tok.id, tok.name)}
                        className="px-2 py-1 rounded text-rose-600 dark:text-rose-400 hover:bg-rose-50 dark:hover:bg-rose-950/40 text-[11px] border border-rose-200 dark:border-rose-900/40 transition-colors cursor-pointer"
                      >
                        {t('mcp.revokeAccess')}
                      </button>
                    </div>
                  ))}
                  {tokens.length === 0 && (
                    <p className="text-xs text-slate-500 py-3">{t('mcp.noTokensYet')}</p>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Right Column: Ready-to-copy Configuration Snippets */}
        <div className="space-y-6">
          <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-4 shadow-xs space-y-4">
            <h2 className="text-xs font-semibold uppercase tracking-wider text-slate-700 dark:text-slate-300">
              {t('mcp.snippetsTitle')}
            </h2>

            {/* Antigravity Config */}
            <div>
              <div className="flex items-center justify-between mb-1.5">
                <span className="text-xs font-medium text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                  <i className="pi pi-bolt text-indigo-500 dark:text-indigo-400" />
                  {t('mcp.antigravityConfig')}
                </span>
                <button
                  onClick={() => copyToClipboard(antigravitySnippet, 'antigravity')}
                  className="text-[11px] text-indigo-600 dark:text-indigo-400 hover:text-indigo-500 dark:hover:text-indigo-300 flex items-center gap-1 cursor-pointer"
                >
                  <i className="pi pi-copy text-[10px]" />
                  {copiedKey === 'antigravity' ? t('common.copied') : t('mcp.copyJson')}
                </button>
              </div>
              <pre className="p-3 rounded-lg bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 text-[11px] font-mono text-slate-800 dark:text-slate-300 overflow-x-auto">
                {antigravitySnippet}
              </pre>
            </div>

            {/* Cursor Config */}
            <div>
              <div className="flex items-center justify-between mb-1.5">
                <span className="text-xs font-medium text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                  <i className="pi pi-desktop text-purple-500 dark:text-purple-400" />
                  {t('mcp.cursorConfig')}
                </span>
                <button
                  onClick={() => copyToClipboard(cursorSnippet, 'cursor')}
                  className="text-[11px] text-indigo-600 dark:text-indigo-400 hover:text-indigo-500 dark:hover:text-indigo-300 flex items-center gap-1 cursor-pointer"
                >
                  <i className="pi pi-copy text-[10px]" />
                  {copiedKey === 'cursor' ? t('common.copied') : t('mcp.copyJson')}
                </button>
              </div>
              <pre className="p-3 rounded-lg bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 text-[11px] font-mono text-slate-800 dark:text-slate-300 overflow-x-auto">
                {cursorSnippet}
              </pre>
            </div>

            {/* Claude Desktop Config */}
            <div>
              <div className="flex items-center justify-between mb-1.5">
                <span className="text-xs font-medium text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                  <i className="pi pi-comment text-amber-500 dark:text-amber-400" />
                  {t('mcp.claudeConfig')}
                </span>
                <button
                  onClick={() => copyToClipboard(claudeDesktopSnippet, 'claude')}
                  className="text-[11px] text-indigo-600 dark:text-indigo-400 hover:text-indigo-500 dark:hover:text-indigo-300 flex items-center gap-1 cursor-pointer"
                >
                  <i className="pi pi-copy text-[10px]" />
                  {copiedKey === 'claude' ? t('common.copied') : t('mcp.copyJson')}
                </button>
              </div>
              <pre className="p-3 rounded-lg bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 text-[11px] font-mono text-slate-800 dark:text-slate-300 overflow-x-auto">
                {claudeDesktopSnippet}
              </pre>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default McpSettingsPage
