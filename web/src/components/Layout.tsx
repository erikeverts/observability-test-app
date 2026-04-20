import { Link, Outlet } from 'react-router-dom'
import { useCart } from '../hooks/useCart'

export default function Layout() {
  const { count } = useCart()

  return (
    <div className="app">
      <header className="header">
        <Link to="/" className="logo">ObsDemo Store</Link>
        <nav className="nav">
          <Link to="/">Products</Link>
          <Link to="/orders">Orders</Link>
          <Link to="/load-generator">Load Generator</Link>
          <Link to="/cart" className="cart-link">
            Cart {count > 0 && <span className="badge">{count}</span>}
          </Link>
        </nav>
      </header>
      <main className="main">
        <Outlet />
      </main>
    </div>
  )
}
