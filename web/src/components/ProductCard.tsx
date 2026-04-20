import { Link } from 'react-router-dom'
import type { Product } from '../api/client'

interface Props {
  product: Product
  stock?: number
  onAddToCart: () => void
}

export default function ProductCard({ product, stock, onAddToCart }: Props) {
  const stockQty = stock ?? product.stock
  const inStock = stockQty > 0
  const lowStock = stockQty > 0 && stockQty <= 10

  return (
    <div className="product-card">
      <Link to={`/products/${product.id}`} className="product-link">
        <h3>{product.name}</h3>
        <p className="product-price">${product.price.toFixed(2)}</p>
      </Link>
      <div className="product-stock">
        {lowStock && <span className="stock-low">Low stock: {stockQty}</span>}
        {!inStock && <span className="stock-out">Out of stock</span>}
        {inStock && !lowStock && <span className="stock-ok">{stockQty} in stock</span>}
      </div>
      <button
        className="btn btn-primary"
        onClick={onAddToCart}
        disabled={!inStock}
      >
        Add to Cart
      </button>
    </div>
  )
}
