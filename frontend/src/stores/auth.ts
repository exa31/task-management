/**
 * Example Zustand Store
 * Manage global application state, like authentication.
 */

import { create } from 'zustand'

import type { User } from '@/types'

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  login: (userData: User) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,

  login: (userData: User) => {
    set({ user: userData, isAuthenticated: true })
    // e.g. set token in localStorage
  },

  logout: () => {
    set({ user: null, isAuthenticated: false })
    // e.g. clear token
  }
}))
