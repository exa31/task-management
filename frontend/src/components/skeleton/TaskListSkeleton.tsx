import React from 'react'

import { Skeleton } from './Skeleton'

interface TaskListSkeletonProps {
  count?: number
}

export const TaskListSkeleton: React.FC<TaskListSkeletonProps> = ({ count = 6 }) => (
  <div className="space-y-2.5">
    {[...Array(count).keys()].map((index) => (
      <div
        key={`task-skel-${index}`}
        className="flex items-center gap-3.5 p-3.5 rounded-xl border border-slate-200/70 dark:border-slate-800/70 bg-white/70 dark:bg-slate-900/50 shadow-xs"
      >
        {/* Checkbox placeholder */}
        <Skeleton className="w-5 h-5 rounded-lg shrink-0" />

        {/* Title & Details placeholder */}
        <div className="flex-1 min-w-0 space-y-2">
          <div className="flex items-center gap-2.5">
            <Skeleton
              className="h-4 rounded"
              style={{ width: `${Math.floor(40 + (index % 3) * 20)}%` }}
            />
            {index % 2 === 0 && <Skeleton className="h-3.5 w-12 rounded-full hidden sm:block" />}
          </div>
          <div className="flex items-center gap-3">
            <Skeleton className="h-3 w-24 rounded" />
            <Skeleton className="h-3 w-16 rounded hidden sm:block" />
            <Skeleton className="h-3 w-20 rounded hidden md:block" />
          </div>
        </div>

        {/* Project pill */}
        <Skeleton className="h-5 w-20 rounded-full shrink-0 hidden sm:block" />

        {/* Priority pill */}
        <Skeleton className="h-5 w-16 rounded-full shrink-0 hidden md:block" />

        {/* Due date */}
        <Skeleton className="h-4 w-16 rounded shrink-0 hidden lg:block" />

        {/* Action buttons */}
        <div className="flex items-center gap-1 shrink-0">
          <Skeleton className="w-6 h-6 rounded-md" />
          <Skeleton className="w-6 h-6 rounded-md" />
        </div>
      </div>
    ))}
  </div>
)
