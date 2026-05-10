package main

import "testing"

func TestGetCustomer(t *testing.T) {
	t.Run("returns customer by id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/customers/1", nil)
		expectStatus(t, rec, 200)
		var c Customer
		decodeBody(t, rec, &c)
		if c.ID != 1 || c.Email != "alice@example.com" || c.Name != "Alice Smith" {
			t.Errorf("unexpected fields: %+v", c)
		}
	})

	t.Run("returns 404 for missing id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/customers/99999", nil)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 400 for non-integer id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/customers/abc", nil)
		expectStatus(t, rec, 400)
	})
}

func TestListCustomerOrders(t *testing.T) {
	t.Run("returns orders for a customer, newest first", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/customers/1/orders", nil)
		expectStatus(t, rec, 200)

		var resp struct{ Items []Order `json:"items"` }
		decodeBody(t, rec, &resp)
		if len(resp.Items) != 2 {
			t.Fatalf("expected 2 orders for customer 1, got %d", len(resp.Items))
		}
		// Order 2 was created later than order 1 — newest first.
		if resp.Items[0].ID != 2 {
			t.Errorf("expected newest first (id=2), got id=%d", resp.Items[0].ID)
		}
	})

	t.Run("returns empty list for customer with no orders", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/customers/3/orders", nil)
		expectStatus(t, rec, 200)
		var resp struct{ Items []Order `json:"items"` }
		decodeBody(t, rec, &resp)
		if len(resp.Items) != 0 {
			t.Errorf("expected empty, got %d items", len(resp.Items))
		}
	})

	t.Run("returns 404 for missing customer", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/customers/99999/orders", nil)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 400 for non-integer id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/customers/abc/orders", nil)
		expectStatus(t, rec, 400)
	})
}
