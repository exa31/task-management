import React from 'react'

import { Skeleton } from './Skeleton'

interface RemindersSkeletonProps {
  count?: number
}

export const RemindersSkeleton: React.FC<RemindersSkeletonProps> = ({ count = 6 }) => (
  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3.5">
    {[...Array(count).keys()].map((index) => (
      <div
        key={`rem-skel-${index}`}
        className="p-4 rounded-2xl border border-slate-200/80 dark:border-slate-800/80 bg-white/80 dark:bg-slate-900/60 shadow-xs space-y-3"
      >
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2">
            <Skeleton className="w-2.5 h-2.5 rounded-full shrink-0" />
            <Skeleton className="h-4 w-36 sm:w-48 rounded" />
          </div>
          <Skeleton className="h-4 w-12 rounded-full" />
        </div>

        <Skeleton className="h-3.5 w-28 rounded" />

        <div className="pt-2 border-t border-slate-100 dark:border-slate-800/60 flex items-center justify-between">
          <Skeleton className="h-3.5 w-24 rounded" />
          <div className="flex items-center gap-1.5">
            <Skeleton className="h-6 w-12 rounded-md" />
            <Skeleton className="h-6 w-12 rounded-md" />
          </div>
        </div>
      </div>
    ))}
  </div>
)
