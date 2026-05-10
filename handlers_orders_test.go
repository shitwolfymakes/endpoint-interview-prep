package main

import (
	"testing"
)

func TestGetOrder(t *testing.T) {
	t.Run("returns order with customer and items embedded", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/orders/1", nil)
		expectStatus(t, rec, 200)

		var o Order
		decodeBody(t, rec, &o)
		if o.ID != 1 || o.CustomerID != 1 || o.Status != "paid" || o.TotalCents != 6998 {
			t.Errorf("order fields wrong: %+v", o)
		}
		if o.Customer == nil || o.Customer.ID != 1 {
			t.Errorf("expected customer embedded, got %+v", o.Customer)
		}
		if len(o.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(o.Items))
		}
		// Items should carry the product names too.
		gotName := false
		for _, it := range o.Items {
			if it.ProductName != "" {
				gotName = true
				break
			}
		}
		if !gotName {
			t.Errorf("expected items to include product_name")
		}
	})

	t.Run("returns 404 for missing order", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/orders/99999", nil)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 400 for non-integer id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/orders/abc", nil)
		expectStatus(t, rec, 400)
	})
}

func TestCreateOrder(t *testing.T) {
	t.Run("creates an order, decrements stock, computes total", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{
			"customer_id": 2,
			"items": [
				{"product_id": 1, "quantity": 3},
				{"product_id": 3, "quantity": 2}
			]
		}`
		rec := do(t, srv, "POST", "/orders", body)
		expectStatus(t, rec, 201)

		var o Order
		decodeBody(t, rec, &o)
		// total = 3*1999 + 2*999 = 5997 + 1998 = 7995
		if o.TotalCents != 7995 {
			t.Errorf("expected total 7995, got %d", o.TotalCents)
		}
		if o.CustomerID != 2 {
			t.Errorf("expected customer_id=2, got %d", o.CustomerID)
		}
		if o.Status != "pending" {
			t.Errorf("expected status=pending, got %q", o.Status)
		}
		if len(o.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(o.Items))
		}

		// Stock for product 1 should be 100 - 3 = 97.
		rec2 := do(t, srv, "GET", "/products/1", nil)
		var p Product
		decodeBody(t, rec2, &p)
		if p.StockQuantity != 97 {
			t.Errorf("expected product 1 stock=97, got %d", p.StockQuantity)
		}
	})

	t.Run("returns 422 on insufficient stock", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// Product 6 (TOOL-002) has stock 0.
		body := `{"customer_id":1,"items":[{"product_id":6,"quantity":1}]}`
		rec := do(t, srv, "POST", "/orders", body)
		expectStatus(t, rec, 422)
	})

	t.Run("rolls back stock decrement when a later item fails", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// First item OK; second item exceeds stock. The whole transaction
		// must roll back, so product 1's stock should remain at 100.
		body := `{"customer_id":1,"items":[
			{"product_id":1,"quantity":1},
			{"product_id":6,"quantity":1}
		]}`
		rec := do(t, srv, "POST", "/orders", body)
		expectStatus(t, rec, 422)

		rec2 := do(t, srv, "GET", "/products/1", nil)
		var p Product
		decodeBody(t, rec2, &p)
		if p.StockQuantity != 100 {
			t.Errorf("expected rollback to leave stock at 100, got %d", p.StockQuantity)
		}
	})

	t.Run("returns 400 on unknown customer", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"customer_id":99999,"items":[{"product_id":1,"quantity":1}]}`
		rec := do(t, srv, "POST", "/orders", body)
		expectStatus(t, rec, 400)
	})

	t.Run("returns 400 on unknown product", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"customer_id":1,"items":[{"product_id":99999,"quantity":1}]}`
		rec := do(t, srv, "POST", "/orders", body)
		expectStatus(t, rec, 400)
	})

	t.Run("returns 400 on empty items", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"customer_id":1,"items":[]}`
		rec := do(t, srv, "POST", "/orders", body)
		expectStatus(t, rec, 400)
	})

	t.Run("returns 400 on non-positive quantity", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"customer_id":1,"items":[{"product_id":1,"quantity":0}]}`
		rec := do(t, srv, "POST", "/orders", body)
		expectStatus(t, rec, 400)
	})
}

func TestCancelOrder(t *testing.T) {
	t.Run("cancels a pending order and restores stock", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// Order 2 is pending and contains 1x product 3 (gadget basic).
		// Product 3 starts at stock 200.
		rec := do(t, srv, "POST", "/orders/2/cancel", nil)
		expectStatus(t, rec, 200)

		var o Order
		decodeBody(t, rec, &o)
		if o.Status != "cancelled" {
			t.Errorf("expected status=cancelled, got %q", o.Status)
		}

		// Stock restored to 201.
		rec2 := do(t, srv, "GET", "/products/3", nil)
		var p Product
		decodeBody(t, rec2, &p)
		if p.StockQuantity != 201 {
			t.Errorf("expected stock restored to 201, got %d", p.StockQuantity)
		}
	})

	t.Run("cancels a paid order", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/orders/1/cancel", nil)
		expectStatus(t, rec, 200)
	})

	t.Run("refuses to cancel a shipped order with 409", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/orders/3/cancel", nil)
		expectStatus(t, rec, 409)
	})

	t.Run("refuses to cancel an already-cancelled order with 409", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// Cancel once, then try again.
		rec1 := do(t, srv, "POST", "/orders/2/cancel", nil)
		expectStatus(t, rec1, 200)

		rec2 := do(t, srv, "POST", "/orders/2/cancel", nil)
		expectStatus(t, rec2, 409)
	})

	t.Run("returns 404 for missing order", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/orders/99999/cancel", nil)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 400 for non-integer id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/orders/abc/cancel", nil)
		expectStatus(t, rec, 400)
	})
}
