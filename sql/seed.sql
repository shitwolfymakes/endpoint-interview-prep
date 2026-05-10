-- Deterministic seed data. Tests rely on these exact rows — do not change
-- ids, SKUs, prices, or stock numbers without updating the test files.

INSERT INTO customers (id, email, name, created_at) VALUES
    (1, 'alice@example.com', 'Alice Smith', '2026-01-15 09:00:00'),
    (2, 'bob@example.com',   'Bob Jones',   '2026-02-10 11:30:00'),
    (3, 'carol@example.com', 'Carol White', '2026-03-05 14:00:00');

INSERT INTO products (id, sku, name, description, price_cents, stock_quantity, category, created_at, updated_at) VALUES
    (1, 'WIDGET-001', 'Standard Widget',   'A reliable everyday widget',         1999, 100, 'widgets', '2026-01-01 00:00:00', '2026-01-01 00:00:00'),
    (2, 'WIDGET-002', 'Premium Widget',    'A premium widget for power users',   4999,  50, 'widgets', '2026-01-01 00:00:00', '2026-01-01 00:00:00'),
    (3, 'GADGET-001', 'Basic Gadget',      'An entry-level gadget',               999, 200, 'gadgets', '2026-01-01 00:00:00', '2026-01-01 00:00:00'),
    (4, 'GADGET-002', 'Pro Gadget',        'A professional-grade gadget',        7999,  25, 'gadgets', '2026-01-01 00:00:00', '2026-01-01 00:00:00'),
    (5, 'TOOL-001',   'Claw Hammer',       'Standard 16oz claw hammer',          2499,  75, 'tools',   '2026-01-01 00:00:00', '2026-01-01 00:00:00'),
    (6, 'TOOL-002',   'Screwdriver Set',   '6-piece screwdriver set',            3499,   0, 'tools',   '2026-01-01 00:00:00', '2026-01-01 00:00:00');

INSERT INTO orders (id, customer_id, status, total_cents, created_at, updated_at) VALUES
    (1, 1, 'paid',    6998, '2026-04-01 10:00:00', '2026-04-01 10:00:00'),
    (2, 1, 'pending',  999, '2026-05-01 11:00:00', '2026-05-01 11:00:00'),
    (3, 2, 'shipped', 7999, '2026-04-15 14:30:00', '2026-04-16 09:00:00');

INSERT INTO order_items (order_id, product_id, quantity, unit_price_cents) VALUES
    (1, 1, 1, 1999),
    (1, 2, 1, 4999),
    (2, 3, 1,  999),
    (3, 4, 1, 7999);
