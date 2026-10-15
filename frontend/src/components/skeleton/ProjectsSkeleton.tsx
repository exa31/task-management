import React from 'react'

import { Skeleton } from './Skeleton'

interface ProjectsSkeletonProps {
  count?: number
}

export const ProjectsSkeleton: React.FC<ProjectsSkeletonProps> = ({ count = 3 }) => (
  <div className="space-y-3">
    {[...Array(count).keys()].map((index) => (
      <div
        key={`proj-skel-${index}`}
        className="rounded-2xl border border-slate-200/80 dark:border-slate-800/80 bg-white/80 dark:bg-slate-900/60 p-4 sm:p-5 shadow-xs space-y-3.5"
      >
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-3 min-w-0">
            <Skeleton className="w-3.5 h-3.5 rounded-full shrink-0" />
            <div className="space-y-1">
              <Skeleton className="h-4 w-40 sm:w-56 rounded" />
              <Skeleton className="h-3 w-28 rounded hidden sm:block" />
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Skeleton className="h-5 w-20 rounded-full" />
            <Skeleton className="w-7 h-7 rounded-lg" />
          </div>
        </div>

        {/* Progress bar */}
        <div className="space-y-1.5 pt-1">
          <div className="flex items-center justify-between text-xs">
            <Skeleton className="h-3 w-16 rounded" />
            <Skeleton className="h-3 w-8 rounded" />
          </div>
          <Skeleton className="h-2 w-full rounded-full" />
        </div>
      </div>
    ))}
  </div>
)
