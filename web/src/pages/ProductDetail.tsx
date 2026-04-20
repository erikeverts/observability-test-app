import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { fetchProduct, restockProduct, type Product, fetchInventory } from '../api/client'
import { useCart } from '../hooks/useCart'
import { usePolling } from '../hooks/usePolling'

export default function ProductDetail() {
  const { id } = useParams<{ id: string }>()
  const [product, setProduct] = useState<Product | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [restockQty, setRestockQty] = useState(10)
  const [restocking, setRestocking] = useState(false)
  const { addItem } = useCart()
  const { data: inventory } = usePolling(fetchInventory, 5000)

  useEffect(() => {
    if (!id) return
    fetchProduct(id)
      .then(setProduct)
      .catch(e => setError(e.message))
  }, [id])

  const stock = inventory?.find(e => e.product_id === id)?.quantity ?? product?.stock

  async function handleRestock() {
    if (!id) return
    setRestocking(true)
    try {
      await restockProduct(id, restockQty)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Restock failed')
    } finally {
      setRestocking(false)
    }
  }

  if (error) return <div className="error">Error: {error}</div>
  if (!product) return <div className="loading">Loading...</div>

  return (
    <div className="product-detail">
      <Link to="/" className="back-link">&larr; Back to products</Link>
      <h2>{product.name}</h2>
      <p className="product-description">{product.description}</p>
      <p className="product-price">${product.price.toFixed(2)}</p>
      <p className="product-stock-detail">
        Stock: <strong>{stock ?? '...'}</strong>
      </p>

      <div className="detail-actions">
        <button
          className="btn btn-primary"
          onClick={() => addItem(product.id, product.name, product.price)}
          disabled={stock !== undefined && stock <= 0}
        >
          Add to Cart
        </button>
      </div>

      <div className="restock-section">
        <h3>Restock</h3>
        <div className="restock-form">
          <input
            type="number"
            min={1}
            value={restockQty}
            onChange={e => setRestockQty(Number(e.target.value))}
          />
          <button className="btn btn-secondary" onClick={handleRestock} disabled={restocking}>
            {restocking ? 'Restocking...' : 'Restock'}
          </button>
        </div>
      </div>
    </div>
  )
}
