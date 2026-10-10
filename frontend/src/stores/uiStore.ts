import { create } from 'zustand'

const getInitialTheme = (): 'dark' | 'light' => {
  if (typeof window === 'undefined') return 'dark'
  try {
    const stored = localStorage.getItem('theme')
    if (stored === 'light' || stored === 'dark') {
      return stored
    }
  } catch {
    // ignore
  }
  return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark'
}

const applyThemeToDOM = (theme: 'dark' | 'light') => {
  if (typeof document === 'undefined') return
  if (theme === 'dark') {
    document.documentElement.classList.add('dark')
    document.documentElement.setAttribute('data-theme', 'dark')
    document.documentElement.style.colorScheme = 'dark'
  } else {
    document.documentElement.classList.remove('dark')
    document.documentElement.setAttribute('data-theme', 'light')
    document.documentElement.style.colorScheme = 'light'
  }
  try {
    localStorage.setItem('theme', theme)
  } catch {
    // ignore
  }
}

const getInitialLocale = (): 'en' | 'id' => {
  if (typeof window === 'undefined') return 'en'
  try {
    const stored = localStorage.getItem('locale')
    if (stored === 'en' || stored === 'id') {
      return stored
    }
    if (navigator.language.startsWith('id')) {
      return 'id'
    }
  } catch {
    // ignore
  }
  return 'en'
}

const saveLocale = (locale: 'en' | 'id') => {
  try {
    localStorage.setItem('locale', locale)
  } catch {
    // ignore
  }
}

const initialTheme = getInitialTheme()
applyThemeToDOM(initialTheme)
const initialLocale = getInitialLocale()

interface UIState {
  isCommandPaletteOpen: boolean
  isCreateTaskModalOpen: boolean
  createTaskInitialProjectId: number | null
  isCreateNoteModalOpen: boolean
  theme: 'dark' | 'light'
  locale: 'en' | 'id'
  openCommandPalette: () => void
  closeCommandPalette: () => void
  toggleCommandPalette: () => void
  openCreateTaskModal: (initialProjectId?: number | null | unknown) => void
  closeCreateTaskModal: () => void
  openCreateNoteModal: () => void
  closeCreateNoteModal: () => void
  toggleTheme: () => void
  setTheme: (theme: 'dark' | 'light') => void
  setLocale: (locale: 'en' | 'id') => void
  toggleLocale: () => void
}

export const useUIStore = create<UIState>((set) => ({
  isCommandPaletteOpen: false,
  isCreateTaskModalOpen: false,
  createTaskInitialProjectId: null,
  isCreateNoteModalOpen: false,
  theme: initialTheme,
  locale: initialLocale,

  openCommandPalette: () => set({ isCommandPaletteOpen: true }),
  closeCommandPalette: () => set({ isCommandPaletteOpen: false }),
  toggleCommandPalette: () =>
    set((state) => ({ isCommandPaletteOpen: !state.isCommandPaletteOpen })),

  openCreateTaskModal: (initialProjectId?: number | null | unknown) =>
    set({
      isCreateTaskModalOpen: true,
      createTaskInitialProjectId: typeof initialProjectId === 'number' ? initialProjectId : null
    }),
  closeCreateTaskModal: () =>
    set({ isCreateTaskModalOpen: false, createTaskInitialProjectId: null }),

  openCreateNoteModal: () => set({ isCreateNoteModalOpen: true }),
  closeCreateNoteModal: () => set({ isCreateNoteModalOpen: false }),

  toggleTheme: () =>
    set((state) => {
      const nextTheme = state.theme === 'dark' ? 'light' : 'dark'
      applyThemeToDOM(nextTheme)
      return { theme: nextTheme }
    }),

  setTheme: (theme: 'dark' | 'light') => {
    applyThemeToDOM(theme)
    set({ theme })
  },

  setLocale: (locale: 'en' | 'id') => {
    saveLocale(locale)
    set({ locale })
  },

  toggleLocale: () =>
    set((state) => {
      const nextLocale = state.locale === 'en' ? 'id' : 'en'
      saveLocale(nextLocale)
      return { locale: nextLocale }
    })
}))
