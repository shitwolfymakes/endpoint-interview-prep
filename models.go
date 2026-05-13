package main

// Customer is a buyer who places orders.
type Customer struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// Product is something a customer can buy. price_cents avoids floating-point
// money. stock_quantity is decremented when an order is placed and restored
// when an order is cancelled.
type Product struct {
	ID            int64  `json:"id"`
	SKU           string `json:"sku"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	PriceCents    int64  `json:"price_cents"`
	StockQuantity int    `json:"stock_quantity"`
	Category      string `json:"category"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// Order is a customer purchase. Total is the sum of (qty * unit_price_cents)
// across its line items at the time of creation.
type Order struct {
	ID         int64       `json:"id"`
	CustomerID int64       `json:"customer_id"`
	Status     string      `json:"status"`
	TotalCents int64       `json:"total_cents"`
	CreatedAt  string      `json:"created_at"`
	UpdatedAt  string      `json:"updated_at"`
	Customer   *Customer   `json:"customer,omitempty"`
	Items      []OrderItem `json:"items,omitempty"`
}

// OrderItem is a single line in an order. unit_price_cents captures the
// product price at the time of the order — products can change price later
// without affecting historical orders.
type OrderItem struct {
	ID             int64  `json:"id"`
	OrderID        int64  `json:"order_id"`
	ProductID      int64  `json:"product_id"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	ProductName    string `json:"product_name,omitempty"`
}

// Warehouse is a physical storage location. capacity_units is the maximum
// total number of product units it can hold, summed across every product type.
// status is either 'active' or 'closed'.
type Warehouse struct {
	ID            int64  `json:"id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	Address       string `json:"address"`
	CapacityUnits int    `json:"capacity_units"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// InventoryLine is one (product, quantity) row at a warehouse. ProductName is
// populated by handlers that join against the products table.
type InventoryLine struct {
	WarehouseID int64  `json:"warehouse_id"`
	ProductID   int64  `json:"product_id"`
	Quantity    int    `json:"quantity"`
	ProductName string `json:"product_name,omitempty"`
}

// Personnel is an employee assigned to at most one warehouse. WarehouseID is
// a pointer so it can be null (unassigned / newly hired / terminated).
// TerminatedAt is "" for active employees.
type Personnel struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	Role         string     `json:"role"`
	WarehouseID  *int64     `json:"warehouse_id"`
	HiredAt      string     `json:"hired_at"`
	TerminatedAt string     `json:"terminated_at,omitempty"`
	Warehouse    *Warehouse `json:"warehouse,omitempty"`
}
