import { useEffect, useState } from 'react'
import { fetchProducts, fetchInventory, type Product } from '../api/client'
import { useCart } from '../hooks/useCart'
import { usePolling } from '../hooks/usePolling'
import ProductCard from '../components/ProductCard'

export default function Products() {
  const [products, setProducts] = useState<Product[]>([])
  const [error, setError] = useState<string | null>(null)
  const { addItem } = useCart()
  const { data: inventory } = usePolling(fetchInventory, 5000)

  useEffect(() => {
    fetchProducts()
      .then(setProducts)
      .catch(e => setError(e.message))
  }, [])

  const stockMap = new Map<string, number>()
  if (inventory) {
    for (const entry of inventory) {
      stockMap.set(entry.product_id, entry.quantity)
    }
  }

  if (error) return <div className="error">Error: {error}</div>

  return (
    <div className="products-page">
      <h2>Products</h2>
      <div className="product-grid">
        {products.map(p => (
          <ProductCard
            key={p.id}
            product={p}
            stock={stockMap.get(p.id)}
            onAddToCart={() => addItem(p.id, p.name, p.price)}
          />
        ))}
      </div>
    </div>
  )
}
