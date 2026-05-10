package main

import (
	"database/sql"
	"net/http"
)

// getOrder: GET /orders/{id}
//
// Returns one order with its customer embedded (the Customer field on the
// Order struct) and its line items embedded (the Items field). For each
// item, also populate ProductName by joining on products.
//
// Status codes:
//
//	200  on success
//	400  on non-integer {id}
//	404  if no order has that id
//
// Hints:
//   - One query for the order + customer (JOIN customers)
//   - One query for the line items (JOIN products to get the name)
//   - Or do it all in a single query and dedupe in Go — either is fine
func getOrder(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = db
		writeError(w, http.StatusNotImplemented, "TODO: implement getOrder")
	}
}

// createOrder: POST /orders
//
// Creates a new order in a single transaction. For each item, validate that
// the product exists and has enough stock; then decrement the stock,
// snapshot the unit_price_cents from the product row, and insert the line
// item. Compute total_cents as the sum of (quantity * unit_price_cents).
//
// Request body:
//
//	{
//	  "customer_id": int,
//	  "items": [
//	    {"product_id": int, "quantity": int},
//	    ...
//	  ]
//	}
//
// Status codes:
//
//	201  on success — return the created order with items embedded
//	400  on invalid JSON, missing fields, empty items, quantity <= 0,
//	     or unknown customer/product
//	422  if any item has insufficient stock — error message should name
//	     the offending product so the client can fix it
//
// Hints:
//   - Begin a tx; defer tx.Rollback(); call tx.Commit() at the end
//   - Lock-free approach for SQLite: use a guarded UPDATE on each product:
//
//	    UPDATE products SET stock_quantity = stock_quantity - ?,
//	           updated_at = ...
//	    WHERE id = ? AND stock_quantity >= ?
//
//     If RowsAffected() == 0, either the product is missing or stock is
//     insufficient — distinguish with a follow-up SELECT to give a good
//     error message.
//   - Insert the order row first to get an order_id, then insert each item.
func createOrder(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = db
		writeError(w, http.StatusNotImplemented, "TODO: implement createOrder")
	}
}

// cancelOrder: POST /orders/{id}/cancel
//
// Cancels an order and restores the stock for every item. Only orders in
// status 'pending' or 'paid' may be cancelled. 'shipped', 'delivered', and
// 'cancelled' are terminal/post-shipment and cannot be cancelled here.
//
// Status codes:
//
//	200  on success — return the updated order
//	400  on non-integer {id}
//	404  if no order has that id
//	409  if the order is in a status that cannot be cancelled — error
//	     message should include the current status
//
// Hints:
//   - Use a transaction
//   - SELECT the current status with `... FOR UPDATE` style guard:
//	    UPDATE orders SET status='cancelled', updated_at=...
//	    WHERE id = ? AND status IN ('pending','paid')
//     RowsAffected == 0 -> either missing or wrong status; do a follow-up
//     SELECT to tell the client which.
//   - Then for each line item: UPDATE products SET stock_quantity = stock_quantity + ?
func cancelOrder(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = db
		writeError(w, http.StatusNotImplemented, "TODO: implement cancelOrder")
	}
}
