package main

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// newRouter wires every endpoint to its handler. Add routes here when you
// pick up new endpoints from ENDPOINTS.md.
func newRouter(db *sql.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Get("/health", health)

	r.Route("/products", func(r chi.Router) {
		r.Get("/", listProducts(db))
		r.Post("/", createProduct(db))
		r.Post("/bulk", bulkCreateProducts(db))
		r.Get("/{id}", getProduct(db))
		r.Put("/{id}", updateProduct(db))
		r.Delete("/{id}", deleteProduct(db))
		r.Patch("/{id}/stock", adjustStock(db))
	})

	r.Route("/customers", func(r chi.Router) {
		r.Get("/{id}", getCustomer(db))
		r.Get("/{id}/orders", listCustomerOrders(db))
	})

	r.Route("/orders", func(r chi.Router) {
		r.Post("/", createOrder(db))
		r.Get("/{id}", getOrder(db))
		r.Post("/{id}/cancel", cancelOrder(db))
	})

	r.Route("/reports", func(r chi.Router) {
		r.Get("/revenue", revenueReport(db))
		r.Get("/top-products", topProducts(db))
	})

	return r
}
