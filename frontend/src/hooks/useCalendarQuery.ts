import { useQuery } from '@tanstack/react-query'

import { fetchCalendarData, type CalendarGraphQLResponse } from '@/api/graphql'

export const calendarKeys = {
  all: ['calendar-graphql'] as const,
  range: (from?: string, to?: string) => ['calendar-graphql', from || '', to || ''] as const
}

export const useCalendarGraphQL = (from?: string, to?: string) =>
  useQuery<CalendarGraphQLResponse>({
    queryKey: calendarKeys.range(from, to),
    queryFn: () => fetchCalendarData(from, to),
    staleTime: 30 * 1000
  })
