/**
 * Order domain types
 */
export interface Order {
  id: string
  customer: string
  total: number
  status: 'Pending' | 'Processing' | 'Shipped' | 'Delivered' | 'Cancelled'
  date: string
}
