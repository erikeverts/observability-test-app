CREATE TABLE IF NOT EXISTS products (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price       DOUBLE PRECISION NOT NULL DEFAULT 0,
    stock       INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS orders (
    id         TEXT PRIMARY KEY,
    status     TEXT NOT NULL DEFAULT 'pending',
    total      DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    id         SERIAL PRIMARY KEY,
    order_id   TEXT NOT NULL REFERENCES orders(id),
    product_id TEXT NOT NULL,
    quantity   INT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);

CREATE TABLE IF NOT EXISTS inventory (
    product_id TEXT PRIMARY KEY,
    quantity   INT NOT NULL DEFAULT 0
);

-- Seed products
INSERT INTO products (id, name, description, price, stock) VALUES
    ('prod-1', 'Mechanical Keyboard', 'Cherry MX Blue switches', 149.99, 50),
    ('prod-2', 'Wireless Mouse', 'Ergonomic design, 2.4GHz', 49.99, 120),
    ('prod-3', 'USB-C Hub', '7-in-1 multiport adapter', 39.99, 200),
    ('prod-4', 'Monitor Stand', 'Adjustable aluminum stand', 79.99, 35),
    ('prod-5', 'Webcam HD', '1080p with autofocus', 69.99, 80)
ON CONFLICT (id) DO NOTHING;

-- Seed inventory
INSERT INTO inventory (product_id, quantity) VALUES
    ('prod-1', 50),
    ('prod-2', 120),
    ('prod-3', 200),
    ('prod-4', 35),
    ('prod-5', 80)
ON CONFLICT (product_id) DO NOTHING;
