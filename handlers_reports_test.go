package main

import "testing"

func TestRevenueReport(t *testing.T) {
	t.Run("totals all non-cancelled orders by default", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/reports/revenue", nil)
		expectStatus(t, rec, 200)

		var resp struct {
			TotalCents int64 `json:"total_cents"`
			OrderCount int   `json:"order_count"`
		}
		decodeBody(t, rec, &resp)
		// Seed totals: 6998 + 999 + 7999 = 15996, all non-cancelled.
		if resp.TotalCents != 15996 {
			t.Errorf("expected total 15996, got %d", resp.TotalCents)
		}
		if resp.OrderCount != 3 {
			t.Errorf("expected 3 orders, got %d", resp.OrderCount)
		}
	})

	t.Run("filters by date range", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// Only April orders: order 1 (6998) and order 3 (7999) -> 14997.
		rec := do(t, srv, "GET", "/reports/revenue?from=2026-04-01&to=2026-04-30", nil)
		expectStatus(t, rec, 200)
		var resp struct {
			TotalCents int64 `json:"total_cents"`
			OrderCount int   `json:"order_count"`
		}
		decodeBody(t, rec, &resp)
		if resp.TotalCents != 14997 {
			t.Errorf("expected total 14997, got %d", resp.TotalCents)
		}
		if resp.OrderCount != 2 {
			t.Errorf("expected 2 orders, got %d", resp.OrderCount)
		}
	})

	t.Run("excludes cancelled orders", func(t *testing.T) {
		srv, db := newTestServer(t)
		// Cancel order 1 directly in the DB (handler may not be implemented yet).
		if _, err := db.Exec(`UPDATE orders SET status='cancelled' WHERE id=1`); err != nil {
			t.Fatalf("setup: %v", err)
		}
		rec := do(t, srv, "GET", "/reports/revenue", nil)
		expectStatus(t, rec, 200)
		var resp struct {
			TotalCents int64 `json:"total_cents"`
			OrderCount int   `json:"order_count"`
		}
		decodeBody(t, rec, &resp)
		// Without order 1: 999 + 7999 = 8998.
		if resp.TotalCents != 8998 {
			t.Errorf("expected total 8998, got %d", resp.TotalCents)
		}
		if resp.OrderCount != 2 {
			t.Errorf("expected 2 orders, got %d", resp.OrderCount)
		}
	})

	t.Run("returns zero when no orders match", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/reports/revenue?from=2099-01-01&to=2099-12-31", nil)
		expectStatus(t, rec, 200)
		var resp struct {
			TotalCents int64 `json:"total_cents"`
			OrderCount int   `json:"order_count"`
		}
		decodeBody(t, rec, &resp)
		if resp.TotalCents != 0 || resp.OrderCount != 0 {
			t.Errorf("expected zeros, got total=%d count=%d", resp.TotalCents, resp.OrderCount)
		}
	})

	t.Run("rejects bad date format", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/reports/revenue?from=not-a-date", nil)
		expectStatus(t, rec, 400)
	})
}

func TestTopProducts(t *testing.T) {
	t.Run("returns products ranked by revenue", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/reports/top-products", nil)
		expectStatus(t, rec, 200)

		var resp struct {
			Items []struct {
				ProductID    int64  `json:"product_id"`
				Name         string `json:"name"`
				UnitsSold    int    `json:"units_sold"`
				RevenueCents int64  `json:"revenue_cents"`
			} `json:"items"`
		}
		decodeBody(t, rec, &resp)
		// Seed revenue per product: p1=1999, p2=4999, p3=999, p4=7999.
		// Top should be p4 (Pro Gadget) at 7999.
		if len(resp.Items) == 0 {
			t.Fatalf("expected items, got none")
		}
		if resp.Items[0].ProductID != 4 || resp.Items[0].RevenueCents != 7999 {
			t.Errorf("expected top to be product 4 with revenue 7999, got %+v", resp.Items[0])
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/reports/top-products?limit=2", nil)
		expectStatus(t, rec, 200)
		var resp struct {
			Items []struct {
				ProductID int64 `json:"product_id"`
			} `json:"items"`
		}
		decodeBody(t, rec, &resp)
		if len(resp.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(resp.Items))
		}
	})

	t.Run("excludes cancelled orders", func(t *testing.T) {
		srv, db := newTestServer(t)
		// Cancel order 3 (which contains product 4, the top earner).
		if _, err := db.Exec(`UPDATE orders SET status='cancelled' WHERE id=3`); err != nil {
			t.Fatalf("setup: %v", err)
		}
		rec := do(t, srv, "GET", "/reports/top-products", nil)
		expectStatus(t, rec, 200)
		var resp struct {
			Items []struct {
				ProductID    int64 `json:"product_id"`
				RevenueCents int64 `json:"revenue_cents"`
			} `json:"items"`
		}
		decodeBody(t, rec, &resp)
		// Now top should be product 2 (Premium Widget) at 4999.
		if resp.Items[0].ProductID != 2 || resp.Items[0].RevenueCents != 4999 {
			t.Errorf("expected top product 2 at 4999, got %+v", resp.Items[0])
		}
		// Product 4 should not appear at all.
		for _, it := range resp.Items {
			if it.ProductID == 4 {
				t.Errorf("product 4 should be excluded (cancelled)")
			}
		}
	})

	t.Run("rejects bad limit", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/reports/top-products?limit=abc", nil)
		expectStatus(t, rec, 400)
	})
}
