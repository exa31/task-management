import React from 'react'

import { Skeleton } from './Skeleton'

export const TaskBoardSkeleton: React.FC = () => (
  <div className="h-full flex gap-3.5 items-stretch overflow-x-auto pb-2">
    {[...Array(5).keys()].map((colIdx) => (
      <div
        key={`board-col-skel-${colIdx}`}
        className="flex flex-col rounded-xl border border-slate-200/80 dark:border-slate-800/80 bg-slate-100/70 dark:bg-slate-900/50 p-2.5 flex-1 min-w-[280px] space-y-3"
      >
        {/* Column Header */}
        <div className="flex items-center justify-between pb-2 border-b border-slate-200/60 dark:border-slate-800/60">
          <div className="flex items-center gap-2">
            <Skeleton className="w-2.5 h-2.5 rounded-full" />
            <Skeleton className="h-4 w-20 rounded" />
          </div>
          <Skeleton className="h-4 w-6 rounded-full" />
        </div>

        {/* Column Task Cards */}
        <div className="space-y-2.5 flex-1">
          {[...Array(colIdx === 0 ? 3 : colIdx === 1 ? 2 : 1).keys()].map((cardIdx) => (
            <div
              key={`col-${colIdx}-card-${cardIdx}`}
              className="p-3 rounded-xl border border-slate-200/80 dark:border-slate-800/80 bg-white/90 dark:bg-slate-900/80 shadow-xs space-y-2.5"
            >
              <div className="flex items-start justify-between gap-2">
                <Skeleton
                  className="h-3.5 rounded"
                  style={{ width: `${Math.floor(65 + ((cardIdx + colIdx) % 3) * 15)}%` }}
                />
                <Skeleton className="w-4 h-4 rounded shrink-0" />
              </div>
              <div className="flex items-center gap-2">
                <Skeleton className="h-4 w-16 rounded-full" />
                <Skeleton className="h-4 w-12 rounded-full" />
              </div>
              <div className="pt-2 border-t border-slate-100 dark:border-slate-800/60 flex items-center justify-between">
                <Skeleton className="h-3 w-16 rounded" />
                <Skeleton className="h-3 w-12 rounded" />
              </div>
            </div>
          ))}
        </div>
      </div>
    ))}
  </div>
)
