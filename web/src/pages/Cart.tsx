import { Link } from 'react-router-dom'
import { useCart } from '../hooks/useCart'

export default function Cart() {
  const { items, removeItem, updateQuantity, total, count } = useCart()

  if (count === 0) {
    return (
      <div className="cart-page">
        <h2>Cart</h2>
        <p className="empty-state">Your cart is empty.</p>
        <Link to="/" className="btn btn-primary">Browse Products</Link>
      </div>
    )
  }

  return (
    <div className="cart-page">
      <h2>Cart</h2>
      <div className="cart-items">
        {items.map(item => (
          <div key={item.productId} className="cart-item">
            <div className="cart-item-info">
              <span className="cart-item-name">{item.name}</span>
              <span className="cart-item-price">${item.price.toFixed(2)}</span>
            </div>
            <div className="cart-item-controls">
              <button onClick={() => updateQuantity(item.productId, item.quantity - 1)}>-</button>
              <span>{item.quantity}</span>
              <button onClick={() => updateQuantity(item.productId, item.quantity + 1)}>+</button>
              <button className="remove" onClick={() => removeItem(item.productId)}>Remove</button>
            </div>
          </div>
        ))}
      </div>
      <div className="cart-summary">
        <span className="cart-total">Total: ${total.toFixed(2)}</span>
        <Link to="/checkout" className="btn btn-primary">Checkout</Link>
      </div>
    </div>
  )
}
