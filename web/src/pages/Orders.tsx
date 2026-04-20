import { Link } from 'react-router-dom'
import { fetchOrders } from '../api/client'
import { usePolling } from '../hooks/usePolling'

export default function Orders() {
  const { data: orders, error, loading } = usePolling(fetchOrders, 5000)

  if (loading) return <div className="loading">Loading orders...</div>
  if (error) return <div className="error">Error: {error}</div>
  if (!orders || orders.length === 0) {
    return (
      <div className="orders-page">
        <h2>Orders</h2>
        <p className="empty-state">No orders yet.</p>
        <Link to="/" className="btn btn-primary">Browse Products</Link>
      </div>
    )
  }

  return (
    <div className="orders-page">
      <h2>Orders</h2>
      <div className="order-list">
        {orders.map(order => (
          <Link to={`/orders/${order.id}`} key={order.id} className="order-card">
            <div className="order-header">
              <span className="order-id">{order.id}</span>
              <span className={`order-status status-${order.status}`}>{order.status}</span>
            </div>
            <div className="order-meta">
              <span>{order.items.length} item(s)</span>
              <span>${order.total.toFixed(2)}</span>
              <span>{new Date(order.created_at).toLocaleString()}</span>
            </div>
          </Link>
        ))}
      </div>
    </div>
  )
}
