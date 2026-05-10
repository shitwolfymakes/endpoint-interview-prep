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
