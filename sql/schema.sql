-- Schema for the order-management practice app.
--
-- Domain: a small inventory + orders system. Customers buy products; an
-- order has a status and one or more line items. Stock is decremented when
-- an order is created and restored when an order is cancelled.

CREATE TABLE IF NOT EXISTS customers (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    email      TEXT    NOT NULL UNIQUE,
    name       TEXT    NOT NULL,
    created_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S','now'))
);

CREATE TABLE IF NOT EXISTS products (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    sku            TEXT    NOT NULL UNIQUE,
    name           TEXT    NOT NULL,
    description    TEXT    NOT NULL DEFAULT '',
    price_cents    INTEGER NOT NULL CHECK (price_cents >= 0),
    stock_quantity INTEGER NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
    category       TEXT    NOT NULL DEFAULT '',
    created_at     TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S','now')),
    updated_at     TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S','now'))
);

CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);

CREATE TABLE IF NOT EXISTS orders (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    customer_id INTEGER NOT NULL REFERENCES customers(id),
    status      TEXT    NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending','paid','shipped','delivered','cancelled')),
    total_cents INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S','now')),
    updated_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S','now'))
);

CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_status   ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created  ON orders(created_at);

CREATE TABLE IF NOT EXISTS order_items (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id         INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id       INTEGER NOT NULL REFERENCES products(id),
    quantity         INTEGER NOT NULL CHECK (quantity > 0),
    unit_price_cents INTEGER NOT NULL CHECK (unit_price_cents >= 0)
);

CREATE INDEX IF NOT EXISTS idx_order_items_order   ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product ON order_items(product_id);

-- Warehouses hold a flat maximum number of product units (capacity_units) and
-- can store any mix of products. The status field gates whether a warehouse
-- can be used at all.
CREATE TABLE IF NOT EXISTS warehouses (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    code           TEXT    NOT NULL UNIQUE,
    name           TEXT    NOT NULL,
    address        TEXT    NOT NULL DEFAULT '',
    capacity_units INTEGER NOT NULL CHECK (capacity_units >= 0),
    status         TEXT    NOT NULL DEFAULT 'active'
                   CHECK (status IN ('active','closed')),
    created_at     TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S','now')),
    updated_at     TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S','now'))
);

CREATE INDEX IF NOT EXISTS idx_warehouses_status ON warehouses(status);

-- warehouse_inventory is the per-warehouse breakdown of which products are
-- stored where, and how many of each. SUM(quantity) per warehouse must stay
-- <= the warehouse's capacity_units (enforced in handler code, not by SQL).
-- This is separate from products.stock_quantity, which is the company-wide
-- total — the two are deliberately not auto-synced in this practice repo.
CREATE TABLE IF NOT EXISTS warehouse_inventory (
    warehouse_id INTEGER NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    product_id   INTEGER NOT NULL REFERENCES products(id),
    quantity     INTEGER NOT NULL CHECK (quantity >= 0),
    PRIMARY KEY (warehouse_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_warehouse_inventory_product ON warehouse_inventory(product_id);

-- Personnel are employees assigned to (at most) one warehouse. warehouse_id
-- is nullable to allow "unassigned" staff (newly hired, on leave, terminated).
-- terminated_at is NULL while the employee is active.
CREATE TABLE IF NOT EXISTS personnel (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    email         TEXT    NOT NULL UNIQUE,
    name          TEXT    NOT NULL,
    role          TEXT    NOT NULL CHECK (role IN ('manager','clerk','picker','driver')),
    warehouse_id  INTEGER REFERENCES warehouses(id),
    hired_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S','now')),
    terminated_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_personnel_warehouse ON personnel(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_personnel_role      ON personnel(role);
