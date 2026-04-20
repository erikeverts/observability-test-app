const BASE = ''

export interface Product {
  id: string
  name: string
  description: string
  price: number
  stock: number
}

export interface StockEntry {
  product_id: string
  quantity: number
}

export interface OrderItem {
  product_id: string
  quantity: number
}

export interface Order {
  id: string
  items: OrderItem[]
  status: string
  total: number
  created_at: string
}

export async function fetchProducts(): Promise<Product[]> {
  const res = await fetch(`${BASE}/products`)
  if (!res.ok) throw new Error(`Failed to fetch products: ${res.status}`)
  return res.json()
}

export async function fetchProduct(id: string): Promise<Product> {
  const res = await fetch(`${BASE}/products/${id}`)
  if (!res.ok) throw new Error(`Failed to fetch product: ${res.status}`)
  return res.json()
}

export async function fetchInventory(): Promise<StockEntry[]> {
  const res = await fetch(`${BASE}/inventory`)
  if (!res.ok) throw new Error(`Failed to fetch inventory: ${res.status}`)
  return res.json()
}

export async function fetchOrders(): Promise<Order[]> {
  const res = await fetch(`${BASE}/orders`)
  if (!res.ok) throw new Error(`Failed to fetch orders: ${res.status}`)
  return res.json()
}

export async function fetchOrder(id: string): Promise<Order> {
  const res = await fetch(`${BASE}/orders/${id}`)
  if (!res.ok) throw new Error(`Failed to fetch order: ${res.status}`)
  return res.json()
}

export async function createOrder(items: OrderItem[]): Promise<Order> {
  const res = await fetch(`${BASE}/orders`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ items }),
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Failed to create order: ${res.status}`)
  }
  return res.json()
}

export async function restockProduct(productId: string, quantity: number): Promise<{ product_id: string; added: number; remaining: number }> {
  const res = await fetch(`${BASE}/inventory/restock`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ product_id: productId, quantity }),
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Failed to restock: ${res.status}`)
  }
  return res.json()
}
