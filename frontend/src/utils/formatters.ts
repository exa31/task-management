/**
 * Example Utility Functions
 * Keep generic helper functions here (e.g. date formatting, currency, etc).
 */

export const formatCurrency = (value: number): string => new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(value)

export const formatDate = (date: string | Date): string => new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  }).format(new Date(date))

export const cn = (...classes: (string | boolean | undefined | null)[]): string => classes.filter(Boolean).join(' ')
