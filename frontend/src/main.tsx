import 'primeicons/primeicons.css'
import './index.css'

import { QueryClientProvider } from '@tanstack/react-query'
import { PrimeReactProvider } from 'primereact/api'
import React from 'react'
import ReactDOM from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'

import App from './App.tsx'
import ProtectedRoute from './components/auth/ProtectedRoute.tsx'
import AuthLayout from './layouts/AuthLayout.tsx'
import DashboardLayout from './layouts/DashboardLayout.tsx'
import Dashboard from './pages/Dashboard.tsx'
import ErrorPage from './pages/ErrorPage.tsx'
import LandingPage from './pages/LandingPage.tsx'
import Login from './pages/Login.tsx'
import McpSettingsPage from './pages/McpSettingsPage.tsx'
import NotesPage from './pages/NotesPage.tsx'
import NotFound from './pages/NotFound.tsx'
import ProjectsPage from './pages/ProjectsPage.tsx'
import SettingsPage from './pages/SettingsPage.tsx'
import TasksPage from './pages/TasksPage.tsx'
import { queryClient } from './queryClient.ts'

const router = createBrowserRouter([
  {
    path: '/',
    element: <App />,
    errorElement: <ErrorPage />,
    children: [
      // 1. Public Landing Page
      {
        index: true,
        element: <LandingPage />,
      },
      // 2. Authentication Pages (Public)
      {
        path: 'auth',
        element: <AuthLayout />,
        errorElement: <ErrorPage />,
        children: [
          { path: 'login', element: <Login /> },
        ],
      },
      {
        path: 'login',
        element: <Login />,
      },
      // 3. Protected Workspace Routes (Authenticate-Only)
      {
        element: <ProtectedRoute />,
        errorElement: <ErrorPage />,
        children: [
          {
            element: <DashboardLayout />,
            children: [
              { path: 'dashboard', element: <Dashboard /> },
              { path: 'tasks', element: <TasksPage /> },
              { path: 'projects', element: <ProjectsPage /> },
              { path: 'notes', element: <NotesPage /> },
              { path: 'settings', element: <SettingsPage /> },
              { path: 'mcp-settings', element: <McpSettingsPage /> },
            ],
          },
        ],
      },
      {
        path: '*',
        element: <NotFound />,
      },
    ],
  },
])

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <PrimeReactProvider value={{ ripple: true }}>
        <RouterProvider router={router} />
      </PrimeReactProvider>
    </QueryClientProvider>
  </React.StrictMode>
)
