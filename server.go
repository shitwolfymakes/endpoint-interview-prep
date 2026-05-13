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

	r.Route("/warehouses", func(r chi.Router) {
		r.Post("/", createWarehouse(db))
		r.Post("/bulk", bulkCreateWarehouses(db))
		r.Get("/{id}", getWarehouse(db))
		r.Put("/{id}", updateWarehouse(db))
		r.Delete("/{id}", deleteWarehouse(db))
		r.Get("/{id}/inventory", listWarehouseInventory(db))
		r.Get("/{id}/personnel", listWarehousePersonnel(db))
		r.Patch("/{id}/capacity", adjustWarehouseCapacity(db))
		r.Patch("/{id}/inventory/{product_id}", setWarehouseInventory(db))
		r.Post("/{id}/transfer", transferUnits(db))
		r.Post("/{id}/close", closeWarehouse(db))
	})

	r.Route("/personnel", func(r chi.Router) {
		r.Post("/", createPersonnel(db))
		r.Post("/bulk", bulkCreatePersonnel(db))
		r.Get("/{id}", getPersonnel(db))
		r.Put("/{id}", updatePersonnel(db))
		r.Delete("/{id}", deletePersonnel(db))
		r.Patch("/{id}/reassign", reassignPersonnel(db))
	})

	r.Route("/reports", func(r chi.Router) {
		r.Get("/revenue", revenueReport(db))
		r.Get("/top-products", topProducts(db))
		r.Get("/warehouse-utilization", warehouseUtilization(db))
		r.Get("/headcount", headcount(db))
	})

	return r
}
