import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { fetchOrder, type Order } from '../api/client'

export default function OrderDetail() {
  const { id } = useParams<{ id: string }>()
  const [order, setOrder] = useState<Order | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    fetchOrder(id)
      .then(setOrder)
      .catch(e => setError(e.message))
  }, [id])

  if (error) return <div className="error">Error: {error}</div>
  if (!order) return <div className="loading">Loading order...</div>

  return (
    <div className="order-detail">
      <Link to="/orders" className="back-link">&larr; Back to orders</Link>
      <h2>Order {order.id}</h2>
      <div className={`order-status-badge status-${order.status}`}>{order.status}</div>
      <p className="order-date">Placed: {new Date(order.created_at).toLocaleString()}</p>

      <h3>Items</h3>
      <div className="order-items">
        {order.items.map((item, i) => (
          <div key={i} className="order-item-row">
            <span>{item.product_id}</span>
            <span>Qty: {item.quantity}</span>
          </div>
        ))}
      </div>

      <div className="order-total">
        <strong>Total: ${order.total.toFixed(2)}</strong>
      </div>
    </div>
  )
}
