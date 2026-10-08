/**
 * Product domain types
 */
export interface Product {
  id: string
  name: string
  price: number
  stock: number
  category: string
  status: 'Published' | 'Draft' | 'Archived'
}
