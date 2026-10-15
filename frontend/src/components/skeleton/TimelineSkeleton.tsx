import React from 'react'

import { Skeleton } from './Skeleton'

interface TimelineSkeletonProps {
  count?: number
}

export const TimelineSkeleton: React.FC<TimelineSkeletonProps> = ({ count = 4 }) => (
  <div className="relative pl-6 space-y-6 before:absolute before:left-2 before:top-2 before:bottom-2 before:w-0.5 before:bg-slate-200 dark:before:bg-slate-800">
    {[...Array(count).keys()].map((index) => (
      <div key={`timeline-skel-${index}`} className="relative space-y-2">
        {/* Dot */}
        <span className="absolute -left-6 top-1 w-2.5 h-2.5 rounded-full bg-slate-300 dark:bg-slate-700 ring-4 ring-white dark:ring-slate-900" />

        <div className="flex items-center gap-2">
          <Skeleton className="h-3 w-20 rounded" />
          <Skeleton className="h-3 w-28 rounded" />
        </div>

        <Skeleton
          className="h-3.5 rounded"
          style={{ width: `${Math.floor(70 + (index % 3) * 15)}%` }}
        />
      </div>
    ))}
  </div>
)
