import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createOrder } from '../api/client'
import { useCart } from '../hooks/useCart'

export default function Checkout() {
  const { items, total, clearCart } = useCart()
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const navigate = useNavigate()

  if (items.length === 0) {
    navigate('/cart')
    return null
  }

  async function handleSubmit() {
    setSubmitting(true)
    setError(null)
    try {
      const order = await createOrder(
        items.map(i => ({ product_id: i.productId, quantity: i.quantity }))
      )
      clearCart()
      navigate(`/orders/${order.id}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Order failed')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="checkout-page">
      <h2>Checkout</h2>
      <div className="checkout-summary">
        <h3>Order Summary</h3>
        {items.map(item => (
          <div key={item.productId} className="checkout-item">
            <span>{item.name} x{item.quantity}</span>
            <span>${(item.price * item.quantity).toFixed(2)}</span>
          </div>
        ))}
        <div className="checkout-total">
          <strong>Total: ${total.toFixed(2)}</strong>
        </div>
      </div>
      {error && <div className="error">{error}</div>}
      <button className="btn btn-primary" onClick={handleSubmit} disabled={submitting}>
        {submitting ? 'Placing Order...' : 'Place Order'}
      </button>
    </div>
  )
}
