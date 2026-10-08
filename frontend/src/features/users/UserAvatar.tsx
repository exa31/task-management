/**
 * Example Feature Component
 * Place domain-specific UI components in features/<domain>/
 */
import { Avatar } from 'primereact/avatar'

import type { User } from '@/types'

interface UserAvatarProps {
  user: User
}

export default function UserAvatar({ user }: UserAvatarProps) {
  return (
    <div className="flex items-center gap-3">
      <Avatar
        image={user.avatarUrl}
        label={!user.avatarUrl ? user.name.charAt(0) : undefined}
        shape="circle"
        className="bg-indigo-100 text-indigo-700"
      />
      <div>
        <div className="font-medium text-sm text-slate-900 dark:text-slate-50">{user.name}</div>
        <div className="text-xs text-slate-500">{user.email}</div>
      </div>
    </div>
  )
}
