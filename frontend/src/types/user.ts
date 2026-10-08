/**
 * User domain types
 */
export interface User {
  id: string
  name: string
  email: string
  role: 'Admin' | 'User'
  status: 'Active' | 'Inactive'
  avatarUrl?: string
}
