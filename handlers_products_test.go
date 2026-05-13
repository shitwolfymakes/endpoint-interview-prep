package main

import (
	"fmt"
	"testing"
)

// ----------------------------------------------------------------------------
// listProducts (already implemented — these tests should pass out of the box)
// ----------------------------------------------------------------------------

func TestListProducts(t *testing.T) {
	srv, _ := newTestServer(t)

	t.Run("returns all seeded products", func(t *testing.T) {
		rec := do(t, srv, "GET", "/products", nil)
		expectStatus(t, rec, 200)

		var resp struct {
			Items  []Product `json:"items"`
			Limit  int       `json:"limit"`
			Offset int       `json:"offset"`
		}
		decodeBody(t, rec, &resp)
		if len(resp.Items) != 7 {
			t.Fatalf("expected 7 items, got %d", len(resp.Items))
		}
		if resp.Limit != 50 || resp.Offset != 0 {
			t.Errorf("expected limit=50 offset=0, got limit=%d offset=%d", resp.Limit, resp.Offset)
		}
	})

	t.Run("filters by category", func(t *testing.T) {
		rec := do(t, srv, "GET", "/products?category=widgets", nil)
		expectStatus(t, rec, 200)
		var resp struct{ Items []Product `json:"items"` }
		decodeBody(t, rec, &resp)
		if len(resp.Items) != 2 {
			t.Fatalf("expected 2 widgets, got %d", len(resp.Items))
		}
		for _, p := range resp.Items {
			if p.Category != "widgets" {
				t.Errorf("got category %q, want widgets", p.Category)
			}
		}
	})

	t.Run("filters by price range", func(t *testing.T) {
		rec := do(t, srv, "GET", "/products?min_price=2000&max_price=5000", nil)
		expectStatus(t, rec, 200)
		var resp struct{ Items []Product `json:"items"` }
		decodeBody(t, rec, &resp)
		for _, p := range resp.Items {
			if p.PriceCents < 2000 || p.PriceCents > 5000 {
				t.Errorf("price %d out of range", p.PriceCents)
			}
		}
	})

	t.Run("filters by name substring (case-insensitive)", func(t *testing.T) {
		rec := do(t, srv, "GET", "/products?q=widget", nil)
		expectStatus(t, rec, 200)
		var resp struct{ Items []Product `json:"items"` }
		decodeBody(t, rec, &resp)
		if len(resp.Items) != 2 {
			t.Fatalf("expected 2 results, got %d", len(resp.Items))
		}
	})

	t.Run("paginates with limit and offset", func(t *testing.T) {
		rec := do(t, srv, "GET", "/products?limit=2&offset=2", nil)
		expectStatus(t, rec, 200)
		var resp struct {
			Items  []Product `json:"items"`
			Limit  int       `json:"limit"`
			Offset int       `json:"offset"`
		}
		decodeBody(t, rec, &resp)
		if len(resp.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(resp.Items))
		}
		if resp.Items[0].ID != 3 {
			t.Errorf("expected first id=3, got %d", resp.Items[0].ID)
		}
	})

	t.Run("rejects bad limit", func(t *testing.T) {
		rec := do(t, srv, "GET", "/products?limit=abc", nil)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// getProduct (already implemented)
// ----------------------------------------------------------------------------

func TestGetProduct(t *testing.T) {
	srv, _ := newTestServer(t)

	t.Run("returns product by id", func(t *testing.T) {
		rec := do(t, srv, "GET", "/products/1", nil)
		expectStatus(t, rec, 200)
		var p Product
		decodeBody(t, rec, &p)
		if p.ID != 1 || p.SKU != "WIDGET-001" {
			t.Errorf("got id=%d sku=%q, want 1, WIDGET-001", p.ID, p.SKU)
		}
	})

	t.Run("returns 404 for missing id", func(t *testing.T) {
		rec := do(t, srv, "GET", "/products/99999", nil)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 400 for non-integer id", func(t *testing.T) {
		rec := do(t, srv, "GET", "/products/notanumber", nil)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// createProduct (stub — implement to make these pass)
// ----------------------------------------------------------------------------

func TestCreateProduct(t *testing.T) {
	t.Run("creates a product with valid input", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{
			"sku": "NEW-001",
			"name": "New Widget",
			"description": "shiny",
			"price_cents": 1500,
			"stock_quantity": 10,
			"category": "widgets"
		}`
		rec := do(t, srv, "POST", "/products", body)
		expectStatus(t, rec, 201)

		var p Product
		decodeBody(t, rec, &p)
		if p.ID == 0 {
			t.Errorf("expected non-zero id")
		}
		if p.SKU != "NEW-001" || p.Name != "New Widget" || p.PriceCents != 1500 {
			t.Errorf("response fields wrong: %+v", p)
		}
	})

	t.Run("persists the product so it can be fetched", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"sku":"NEW-002","name":"X","price_cents":100}`
		rec := do(t, srv, "POST", "/products", body)
		expectStatus(t, rec, 201)
		var created Product
		decodeBody(t, rec, &created)

		rec2 := do(t, srv, "GET", fmt.Sprintf("/products/%d", created.ID), nil)
		expectStatus(t, rec2, 200)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/products", `{not json`)
		expectStatus(t, rec, 400)
	})

	t.Run("rejects missing required fields", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/products", `{"name":"no sku"}`)
		expectStatus(t, rec, 400)
	})

	t.Run("rejects negative price", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"sku":"NEG-001","name":"x","price_cents":-100}`
		rec := do(t, srv, "POST", "/products", body)
		expectStatus(t, rec, 400)
	})

	t.Run("rejects duplicate SKU with 409", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"sku":"WIDGET-001","name":"dup","price_cents":100}`
		rec := do(t, srv, "POST", "/products", body)
		expectStatus(t, rec, 409)
	})
}

// ----------------------------------------------------------------------------
// updateProduct (stub)
// ----------------------------------------------------------------------------

func TestUpdateProduct(t *testing.T) {
	t.Run("updates an existing product", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{
			"sku": "WIDGET-001",
			"name": "Updated Widget",
			"description": "now improved",
			"price_cents": 2999,
			"stock_quantity": 80,
			"category": "widgets"
		}`
		rec := do(t, srv, "PUT", "/products/1", body)
		expectStatus(t, rec, 200)
		var p Product
		decodeBody(t, rec, &p)
		if p.Name != "Updated Widget" || p.PriceCents != 2999 {
			t.Errorf("update did not stick: %+v", p)
		}

		rec2 := do(t, srv, "GET", "/products/1", nil)
		expectStatus(t, rec2, 200)
		var refetched Product
		decodeBody(t, rec2, &refetched)
		if refetched.Name != "Updated Widget" {
			t.Errorf("update not persisted")
		}
	})

	t.Run("returns 404 for missing id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"sku":"X","name":"x","price_cents":100,"stock_quantity":0}`
		rec := do(t, srv, "PUT", "/products/99999", body)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 409 on SKU collision with another product", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// Try to rename product 1 to product 2's SKU.
		body := `{"sku":"WIDGET-002","name":"x","price_cents":100,"stock_quantity":0}`
		rec := do(t, srv, "PUT", "/products/1", body)
		expectStatus(t, rec, 409)
	})

	t.Run("returns 400 on invalid id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"sku":"X","name":"x","price_cents":100,"stock_quantity":0}`
		rec := do(t, srv, "PUT", "/products/abc", body)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// deleteProduct (stub)
// ----------------------------------------------------------------------------

func TestDeleteProduct(t *testing.T) {
	t.Run("deletes a product not referenced by any order", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// Product 7 (TOOL-003) has no order_items or warehouse_inventory rows.
		rec := do(t, srv, "DELETE", "/products/7", nil)
		expectStatus(t, rec, 204)

		rec2 := do(t, srv, "GET", "/products/7", nil)
		expectStatus(t, rec2, 404)
	})

	t.Run("returns 404 for missing id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "DELETE", "/products/99999", nil)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 409 when product is referenced by an order", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// Product 1 is in order 1.
		rec := do(t, srv, "DELETE", "/products/1", nil)
		expectStatus(t, rec, 409)
	})

	t.Run("returns 400 on invalid id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "DELETE", "/products/abc", nil)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// adjustStock (stub)
// ----------------------------------------------------------------------------

func TestAdjustStock(t *testing.T) {
	t.Run("increments stock with positive delta", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "PATCH", "/products/1/stock", `{"delta": 25}`)
		expectStatus(t, rec, 200)
		var p Product
		decodeBody(t, rec, &p)
		if p.StockQuantity != 125 {
			t.Errorf("expected stock=125, got %d", p.StockQuantity)
		}
	})

	t.Run("decrements stock with negative delta", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "PATCH", "/products/1/stock", `{"delta": -10}`)
		expectStatus(t, rec, 200)
		var p Product
		decodeBody(t, rec, &p)
		if p.StockQuantity != 90 {
			t.Errorf("expected stock=90, got %d", p.StockQuantity)
		}
	})

	t.Run("rejects delta that would drive stock negative with 422", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// Product 6 (TOOL-002) has stock 0; delta -1 would underflow.
		rec := do(t, srv, "PATCH", "/products/6/stock", `{"delta": -1}`)
		expectStatus(t, rec, 422)

		// Confirm stock is unchanged.
		rec2 := do(t, srv, "GET", "/products/6", nil)
		var p Product
		decodeBody(t, rec2, &p)
		if p.StockQuantity != 0 {
			t.Errorf("expected stock unchanged at 0, got %d", p.StockQuantity)
		}
	})

	t.Run("returns 404 for missing product", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "PATCH", "/products/99999/stock", `{"delta": 1}`)
		expectStatus(t, rec, 404)
	})

	t.Run("rejects bad JSON", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "PATCH", "/products/1/stock", `{not json`)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// bulkCreateProducts (stub)
// ----------------------------------------------------------------------------

func TestBulkCreateProducts(t *testing.T) {
	t.Run("creates multiple products in one call", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"products": [
			{"sku":"BULK-001","name":"Bulk A","price_cents":100},
			{"sku":"BULK-002","name":"Bulk B","price_cents":200},
			{"sku":"BULK-003","name":"Bulk C","price_cents":300}
		]}`
		rec := do(t, srv, "POST", "/products/bulk", body)
		expectStatus(t, rec, 201)
		var resp struct{ Items []Product `json:"items"` }
		decodeBody(t, rec, &resp)
		if len(resp.Items) != 3 {
			t.Fatalf("expected 3 created, got %d", len(resp.Items))
		}
	})

	t.Run("rolls back on duplicate SKU within batch", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"products": [
			{"sku":"BULK-X","name":"X","price_cents":100},
			{"sku":"WIDGET-001","name":"dup","price_cents":200}
		]}`
		rec := do(t, srv, "POST", "/products/bulk", body)
		expectStatus(t, rec, 409)

		// BULK-X must NOT exist — the whole batch should have rolled back.
		rec2 := do(t, srv, "GET", "/products?q=BULK-X", nil)
		var resp struct{ Items []Product `json:"items"` }
		decodeBody(t, rec2, &resp)
		if len(resp.Items) != 0 {
			t.Errorf("expected rollback, but BULK-X exists: %+v", resp.Items)
		}
	})

	t.Run("rejects empty list", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/products/bulk", `{"products":[]}`)
		expectStatus(t, rec, 400)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/products/bulk", `{not json`)
		expectStatus(t, rec, 400)
	})
}
