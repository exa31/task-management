import React from 'react'

import { Skeleton } from './Skeleton'

interface NotesSkeletonProps {
  count?: number
}

export const NotesSkeleton: React.FC<NotesSkeletonProps> = ({ count = 8 }) => (
  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
    {[...Array(count).keys()].map((index) => (
      <div
        key={`note-skel-${index}`}
        className="p-4 rounded-2xl border border-slate-200/80 dark:border-slate-800/80 bg-white/80 dark:bg-slate-900/60 shadow-xs space-y-3"
      >
        <div className="flex items-start justify-between gap-2">
          <Skeleton
            className="h-4 rounded"
            style={{ width: `${Math.floor(55 + (index % 3) * 15)}%` }}
          />
          <Skeleton className="w-4 h-4 rounded shrink-0" />
        </div>

        <div className="space-y-1.5">
          <Skeleton className="h-3 w-full rounded" />
          <Skeleton className="h-3 w-4/5 rounded" />
          {index % 2 === 0 && <Skeleton className="h-3 w-3/5 rounded" />}
        </div>

        <div className="flex items-center gap-1.5 pt-1">
          <Skeleton className="h-4 w-12 rounded-full" />
          <Skeleton className="h-4 w-16 rounded-full" />
        </div>

        <div className="pt-2 border-t border-slate-100 dark:border-slate-800/60 flex items-center justify-between">
          <Skeleton className="h-3 w-16 rounded" />
          <div className="flex items-center gap-1.5">
            <Skeleton className="w-5 h-5 rounded" />
            <Skeleton className="w-5 h-5 rounded" />
          </div>
        </div>
      </div>
    ))}
  </div>
)
