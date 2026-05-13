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
    (6, 'TOOL-002',   'Screwdriver Set',   '6-piece screwdriver set',            3499,   0, 'tools',   '2026-01-01 00:00:00', '2026-01-01 00:00:00'),
    (7, 'TOOL-003',   'Tape Measure',      '25ft retractable tape measure',      1299,  40, 'tools',   '2026-01-01 00:00:00', '2026-01-01 00:00:00');

INSERT INTO orders (id, customer_id, status, total_cents, created_at, updated_at) VALUES
    (1, 1, 'paid',    6998, '2026-04-01 10:00:00', '2026-04-01 10:00:00'),
    (2, 1, 'pending',  999, '2026-05-01 11:00:00', '2026-05-01 11:00:00'),
    (3, 2, 'shipped', 7999, '2026-04-15 14:30:00', '2026-04-16 09:00:00');

INSERT INTO order_items (order_id, product_id, quantity, unit_price_cents) VALUES
    (1, 1, 1, 1999),
    (1, 2, 1, 4999),
    (2, 3, 1,  999),
    (3, 4, 1, 7999);

INSERT INTO warehouses (id, code, name, address, capacity_units, status, created_at, updated_at) VALUES
    (1, 'WH-NORTH', 'North Distribution Center', '100 North Rd',  1000, 'active', '2026-01-01 00:00:00', '2026-01-01 00:00:00'),
    (2, 'WH-SOUTH', 'South Distribution Center', '200 South Ave',  500, 'active', '2026-01-01 00:00:00', '2026-01-01 00:00:00'),
    (3, 'WH-EAST',  'East Storage Annex',        '300 East Blvd',  200, 'active', '2026-01-01 00:00:00', '2026-01-01 00:00:00'),
    (4, 'WH-WEST',  'West Storage Annex',        '400 West St',    300, 'closed', '2026-01-01 00:00:00', '2026-01-01 00:00:00');

-- WH-NORTH holds 100+50+100 = 250 of 1000 (25% full)
-- WH-SOUTH holds 100+75      = 175 of  500 (35% full)
-- WH-EAST  is empty
-- WH-WEST  is closed and empty
INSERT INTO warehouse_inventory (warehouse_id, product_id, quantity) VALUES
    (1, 1, 100),
    (1, 2,  50),
    (1, 3, 100),
    (2, 3, 100),
    (2, 5,  75);

INSERT INTO personnel (id, email, name, role, warehouse_id, hired_at, terminated_at) VALUES
    (1, 'dana@example.com',  'Dana Lin',   'manager', 1,    '2025-06-01 09:00:00', NULL),
    (2, 'evan@example.com',  'Evan Park',  'clerk',   1,    '2025-08-15 09:00:00', NULL),
    (3, 'faye@example.com',  'Faye Ortiz', 'picker',  1,    '2025-09-01 09:00:00', NULL),
    (4, 'greg@example.com',  'Greg Hwang', 'manager', 2,    '2025-07-10 09:00:00', NULL),
    (5, 'hana@example.com',  'Hana Iyer',  'driver',  2,    '2025-11-01 09:00:00', NULL),
    (6, 'ivan@example.com',  'Ivan Jones', 'clerk',   NULL, '2024-03-01 09:00:00', '2025-12-31 17:00:00');
