import { useQuery } from '@tanstack/react-query'

import { type CalendarGraphQLResponse, fetchCalendarData } from '@/api/graphql'

export const calendarKeys = {
  all: ['calendar-graphql'] as const,
  range: (from?: string, to?: string) => ['calendar-graphql', from || '', to || ''] as const
}

export const useCalendarGraphQL = (from?: string, to?: string) =>
  useQuery<CalendarGraphQLResponse>({
    queryKey: calendarKeys.range(from, to),
    queryFn: ({ signal }) => fetchCalendarData(from, to, signal),
    staleTime: 30 * 1000
  })
