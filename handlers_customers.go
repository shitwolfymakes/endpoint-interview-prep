package main

import (
	"database/sql"
	"net/http"
)

// getCustomer: GET /customers/{id}
//
// Status codes:
//
//	200  on success — return the customer as JSON
//	400  on non-integer {id}
//	404  if no customer has that id
//
// Hint: see getProduct in handlers_products.go for the same pattern.
func getCustomer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = db
		writeError(w, http.StatusNotImplemented, "TODO: implement getCustomer")
	}
}

// listCustomerOrders: GET /customers/{id}/orders
//
// Returns every order belonging to the given customer, newest first
// (ORDER BY created_at DESC). Each order should include its customer_id,
// status, total_cents, and timestamps. Embedding line items is OPTIONAL —
// the test only checks the order-level fields.
//
// Status codes:
//
//	200  on success — return {"items": [...]}, possibly empty
//	400  on non-integer {id}
//	404  if no customer has that id (do a separate existence check —
//	     an empty result alone shouldn't 404, since a real customer can
//	     genuinely have zero orders)
func listCustomerOrders(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = db
		writeError(w, http.StatusNotImplemented, "TODO: implement listCustomerOrders")
	}
}
