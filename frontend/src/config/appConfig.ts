/**
 * Application Configuration
 *
 * Controls whether mock data fallbacks are enabled when API requests fail.
 * Default is FALSE. When false, API errors are thrown directly and surfaced
 * in the UI so real API states and network failures are not masked.
 *
 * Precedence:
 * 1. Runtime override via localStorage ('app_enable_mocks': 'true' | 'false')
 * 2. Vite environment variable (VITE_ENABLE_MOCKS === 'true')
 * 3. Default: false
 */
export const appConfig = {
  get enableMocks(): boolean {
    if (typeof window !== 'undefined') {
      const override = localStorage.getItem('app_enable_mocks')
      if (override !== null) {
        return override === 'true'
      }
    }
    return import.meta.env.VITE_ENABLE_MOCKS === 'true'
  },

  setEnableMocks(enabled: boolean | null): void {
    if (typeof window !== 'undefined') {
      if (enabled === null) {
        localStorage.removeItem('app_enable_mocks')
      } else {
        localStorage.setItem('app_enable_mocks', String(enabled))
      }
    }
  },
}

// Expose on window in dev mode for quick toggling in browser DevTools
if (typeof window !== 'undefined') {
  ;(window as unknown as { appConfig?: typeof appConfig }).appConfig = appConfig
}
