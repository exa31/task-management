import React, { useEffect } from 'react'
import { Navigate, Outlet, useLocation } from 'react-router-dom'

import { useAuthStore } from '@/stores'

export const ProtectedRoute: React.FC = () => {
  const { isAuthenticated, checkAuth, token } = useAuthStore()
  const location = useLocation()

  useEffect(() => {
    checkAuth()
  }, [checkAuth])

  if (!isAuthenticated && !token) {
    // Preserve intended destination so user can be redirected back after successful sign-in
    return <Navigate to="/auth/login" state={{ from: location }} replace />
  }

  return <Outlet />
}

export default ProtectedRoute
