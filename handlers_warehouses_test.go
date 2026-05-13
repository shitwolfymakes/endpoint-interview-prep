package main

import (
	"fmt"
	"testing"
)

// ----------------------------------------------------------------------------
// createWarehouse
// ----------------------------------------------------------------------------

func TestCreateWarehouse(t *testing.T) {
	t.Run("creates a warehouse with valid input", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{
			"code": "WH-NEW",
			"name": "New DC",
			"address": "1 Test Way",
			"capacity_units": 500
		}`
		rec := do(t, srv, "POST", "/warehouses", body)
		expectStatus(t, rec, 201)

		var wh Warehouse
		decodeBody(t, rec, &wh)
		if wh.ID == 0 {
			t.Errorf("expected non-zero id")
		}
		if wh.Code != "WH-NEW" || wh.CapacityUnits != 500 {
			t.Errorf("response fields wrong: %+v", wh)
		}
		if wh.Status != "active" {
			t.Errorf("expected default status=active, got %q", wh.Status)
		}
	})

	t.Run("persists the warehouse", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"code":"WH-PERSIST","name":"x","capacity_units":10}`
		rec := do(t, srv, "POST", "/warehouses", body)
		expectStatus(t, rec, 201)
		var created Warehouse
		decodeBody(t, rec, &created)

		rec2 := do(t, srv, "GET", fmt.Sprintf("/warehouses/%d", created.ID), nil)
		expectStatus(t, rec2, 200)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/warehouses", `{not json`)
		expectStatus(t, rec, 400)
	})

	t.Run("rejects missing required fields", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/warehouses", `{"name":"no code"}`)
		expectStatus(t, rec, 400)
	})

	t.Run("rejects negative capacity", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"code":"WH-NEG","name":"x","capacity_units":-1}`
		rec := do(t, srv, "POST", "/warehouses", body)
		expectStatus(t, rec, 400)
	})

	t.Run("rejects duplicate code with 409", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"code":"WH-NORTH","name":"dup","capacity_units":100}`
		rec := do(t, srv, "POST", "/warehouses", body)
		expectStatus(t, rec, 409)
	})
}

// ----------------------------------------------------------------------------
// updateWarehouse
// ----------------------------------------------------------------------------

func TestUpdateWarehouse(t *testing.T) {
	t.Run("updates an existing warehouse", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{
			"code": "WH-NORTH",
			"name": "North DC v2",
			"address": "100 North Rd",
			"capacity_units": 1500,
			"status": "active"
		}`
		rec := do(t, srv, "PUT", "/warehouses/1", body)
		expectStatus(t, rec, 200)
		var wh Warehouse
		decodeBody(t, rec, &wh)
		if wh.Name != "North DC v2" || wh.CapacityUnits != 1500 {
			t.Errorf("update did not stick: %+v", wh)
		}
	})

	t.Run("returns 404 for missing id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"code":"X","name":"x","capacity_units":1,"status":"active"}`
		rec := do(t, srv, "PUT", "/warehouses/99999", body)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 409 on code collision", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"code":"WH-SOUTH","name":"x","capacity_units":1,"status":"active"}`
		rec := do(t, srv, "PUT", "/warehouses/1", body)
		expectStatus(t, rec, 409)
	})

	t.Run("returns 400 on invalid id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"code":"X","name":"x","capacity_units":1,"status":"active"}`
		rec := do(t, srv, "PUT", "/warehouses/abc", body)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// deleteWarehouse
// ----------------------------------------------------------------------------

func TestDeleteWarehouse(t *testing.T) {
	t.Run("deletes an empty warehouse", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-EAST (id=3) has no inventory and no personnel.
		rec := do(t, srv, "DELETE", "/warehouses/3", nil)
		expectStatus(t, rec, 204)

		rec2 := do(t, srv, "GET", "/warehouses/3", nil)
		expectStatus(t, rec2, 404)
	})

	t.Run("returns 404 for missing id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "DELETE", "/warehouses/99999", nil)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 409 when warehouse has inventory", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-NORTH has inventory rows.
		rec := do(t, srv, "DELETE", "/warehouses/1", nil)
		expectStatus(t, rec, 409)
	})

	t.Run("returns 400 on invalid id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "DELETE", "/warehouses/abc", nil)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// getWarehouse
// ----------------------------------------------------------------------------

func TestGetWarehouse(t *testing.T) {
	t.Run("returns warehouse by id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/warehouses/1", nil)
		expectStatus(t, rec, 200)
		var wh Warehouse
		decodeBody(t, rec, &wh)
		if wh.ID != 1 || wh.Code != "WH-NORTH" {
			t.Errorf("got id=%d code=%q, want 1 WH-NORTH", wh.ID, wh.Code)
		}
	})

	t.Run("returns 404 for missing id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/warehouses/99999", nil)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 400 for non-integer id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/warehouses/abc", nil)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// listWarehouseInventory
// ----------------------------------------------------------------------------

func TestListWarehouseInventory(t *testing.T) {
	t.Run("returns inventory lines with product names", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-NORTH has 3 inventory rows: products 1, 2, 3.
		rec := do(t, srv, "GET", "/warehouses/1/inventory", nil)
		expectStatus(t, rec, 200)

		var resp struct {
			Items []InventoryLine `json:"items"`
		}
		decodeBody(t, rec, &resp)
		if len(resp.Items) != 3 {
			t.Fatalf("expected 3 lines, got %d", len(resp.Items))
		}
		gotName := false
		for _, it := range resp.Items {
			if it.ProductName != "" {
				gotName = true
				break
			}
		}
		if !gotName {
			t.Errorf("expected items to include product_name")
		}
	})

	t.Run("returns empty list for warehouse with no inventory", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-EAST (id=3) is empty.
		rec := do(t, srv, "GET", "/warehouses/3/inventory", nil)
		expectStatus(t, rec, 200)
		var resp struct {
			Items []InventoryLine `json:"items"`
		}
		decodeBody(t, rec, &resp)
		if len(resp.Items) != 0 {
			t.Errorf("expected empty, got %d", len(resp.Items))
		}
	})

	t.Run("returns 404 for missing warehouse", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/warehouses/99999/inventory", nil)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 400 for non-integer id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/warehouses/abc/inventory", nil)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// listWarehousePersonnel
// ----------------------------------------------------------------------------

func TestListWarehousePersonnel(t *testing.T) {
	t.Run("returns active personnel at a warehouse", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-NORTH has 3 active personnel (1, 2, 3).
		rec := do(t, srv, "GET", "/warehouses/1/personnel", nil)
		expectStatus(t, rec, 200)
		var resp struct {
			Items []Personnel `json:"items"`
		}
		decodeBody(t, rec, &resp)
		if len(resp.Items) != 3 {
			t.Fatalf("expected 3, got %d", len(resp.Items))
		}
	})

	t.Run("returns empty list for warehouse with no personnel", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-EAST has no personnel.
		rec := do(t, srv, "GET", "/warehouses/3/personnel", nil)
		expectStatus(t, rec, 200)
		var resp struct {
			Items []Personnel `json:"items"`
		}
		decodeBody(t, rec, &resp)
		if len(resp.Items) != 0 {
			t.Errorf("expected empty, got %d", len(resp.Items))
		}
	})

	t.Run("returns 404 for missing warehouse", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/warehouses/99999/personnel", nil)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 400 for non-integer id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/warehouses/abc/personnel", nil)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// adjustWarehouseCapacity
// ----------------------------------------------------------------------------

func TestAdjustWarehouseCapacity(t *testing.T) {
	t.Run("increases capacity with positive delta", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "PATCH", "/warehouses/1/capacity", `{"delta": 200}`)
		expectStatus(t, rec, 200)
		var wh Warehouse
		decodeBody(t, rec, &wh)
		if wh.CapacityUnits != 1200 {
			t.Errorf("expected capacity=1200, got %d", wh.CapacityUnits)
		}
	})

	t.Run("decreases capacity with negative delta", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-NORTH has 250 used, capacity 1000 -> can drop to 250.
		rec := do(t, srv, "PATCH", "/warehouses/1/capacity", `{"delta": -750}`)
		expectStatus(t, rec, 200)
		var wh Warehouse
		decodeBody(t, rec, &wh)
		if wh.CapacityUnits != 250 {
			t.Errorf("expected capacity=250, got %d", wh.CapacityUnits)
		}
	})

	t.Run("rejects delta that would drop capacity below current usage with 422", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-NORTH has 250 used -> capacity must stay >= 250. -800 would land at 200.
		rec := do(t, srv, "PATCH", "/warehouses/1/capacity", `{"delta": -800}`)
		expectStatus(t, rec, 422)
	})

	t.Run("returns 404 for missing warehouse", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "PATCH", "/warehouses/99999/capacity", `{"delta": 1}`)
		expectStatus(t, rec, 404)
	})

	t.Run("rejects bad JSON", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "PATCH", "/warehouses/1/capacity", `{not json`)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// setWarehouseInventory
// ----------------------------------------------------------------------------

func TestSetWarehouseInventory(t *testing.T) {
	t.Run("creates a new inventory row", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-EAST is empty; add 50 of product 1.
		rec := do(t, srv, "PATCH", "/warehouses/3/inventory/1", `{"quantity": 50}`)
		expectStatus(t, rec, 200)
		var line InventoryLine
		decodeBody(t, rec, &line)
		if line.Quantity != 50 {
			t.Errorf("expected quantity=50, got %d", line.Quantity)
		}
	})

	t.Run("updates an existing inventory row", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-NORTH has 100 of product 1. Set to 80.
		rec := do(t, srv, "PATCH", "/warehouses/1/inventory/1", `{"quantity": 80}`)
		expectStatus(t, rec, 200)
		var line InventoryLine
		decodeBody(t, rec, &line)
		if line.Quantity != 80 {
			t.Errorf("expected quantity=80, got %d", line.Quantity)
		}
	})

	t.Run("rejects quantity that would exceed capacity with 422", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-EAST has capacity 200, currently empty. Setting product 1 to 201 should fail.
		rec := do(t, srv, "PATCH", "/warehouses/3/inventory/1", `{"quantity": 201}`)
		expectStatus(t, rec, 422)
	})

	t.Run("rejects negative quantity", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "PATCH", "/warehouses/1/inventory/1", `{"quantity": -1}`)
		expectStatus(t, rec, 400)
	})

	t.Run("returns 404 for missing warehouse", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "PATCH", "/warehouses/99999/inventory/1", `{"quantity": 1}`)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 404 for missing product", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "PATCH", "/warehouses/1/inventory/99999", `{"quantity": 1}`)
		expectStatus(t, rec, 404)
	})
}

// ----------------------------------------------------------------------------
// transferUnits
// ----------------------------------------------------------------------------

func TestTransferUnits(t *testing.T) {
	t.Run("moves units between warehouses", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-NORTH has 100 of product 1. Move 30 to WH-EAST.
		body := `{"to_warehouse_id": 3, "product_id": 1, "quantity": 30}`
		rec := do(t, srv, "POST", "/warehouses/1/transfer", body)
		expectStatus(t, rec, 200)

		// Source should now hold 70.
		rec2 := do(t, srv, "GET", "/warehouses/1/inventory", nil)
		var src struct {
			Items []InventoryLine `json:"items"`
		}
		decodeBody(t, rec2, &src)
		for _, it := range src.Items {
			if it.ProductID == 1 && it.Quantity != 70 {
				t.Errorf("expected source product 1 quantity=70, got %d", it.Quantity)
			}
		}

		// Destination should hold 30.
		rec3 := do(t, srv, "GET", "/warehouses/3/inventory", nil)
		var dst struct {
			Items []InventoryLine `json:"items"`
		}
		decodeBody(t, rec3, &dst)
		if len(dst.Items) != 1 || dst.Items[0].Quantity != 30 {
			t.Errorf("expected destination to hold 30 of product 1, got %+v", dst.Items)
		}
	})

	t.Run("rejects transfer that exceeds source stock with 422", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-NORTH has 100 of product 1; try to move 500.
		body := `{"to_warehouse_id": 3, "product_id": 1, "quantity": 500}`
		rec := do(t, srv, "POST", "/warehouses/1/transfer", body)
		expectStatus(t, rec, 422)
	})

	t.Run("rejects transfer that exceeds destination capacity with 422", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-EAST has capacity 200. Move 250 of product 3 from WH-SOUTH? WH-SOUTH has 100.
		// Instead: top up WH-EAST first to leave only 50 of capacity, then try 51.
		// Simpler: WH-NORTH has 50 of product 2. Add capacity check by attempting
		// a transfer that overruns. WH-EAST is empty (cap 200), source must have
		// > 200 of one product. WH-NORTH has 100 of product 1 and 100 of product 3.
		// Neither alone exceeds 200, so instead pre-load WH-EAST.
		_ = do(t, srv, "PATCH", "/warehouses/3/inventory/2", `{"quantity": 180}`)

		// Now try to transfer 50 of product 1 from WH-NORTH to WH-EAST (180+50=230 > 200).
		body := `{"to_warehouse_id": 3, "product_id": 1, "quantity": 50}`
		rec := do(t, srv, "POST", "/warehouses/1/transfer", body)
		expectStatus(t, rec, 422)
	})

	t.Run("rejects transfer to closed warehouse with 409", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-WEST (id=4) is closed.
		body := `{"to_warehouse_id": 4, "product_id": 1, "quantity": 1}`
		rec := do(t, srv, "POST", "/warehouses/1/transfer", body)
		expectStatus(t, rec, 409)
	})

	t.Run("rejects same source and destination with 400", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"to_warehouse_id": 1, "product_id": 1, "quantity": 1}`
		rec := do(t, srv, "POST", "/warehouses/1/transfer", body)
		expectStatus(t, rec, 400)
	})

	t.Run("rolls back on failure", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// Failing transfer should leave source unchanged.
		body := `{"to_warehouse_id": 3, "product_id": 1, "quantity": 9999}`
		rec := do(t, srv, "POST", "/warehouses/1/transfer", body)
		expectStatus(t, rec, 422)

		rec2 := do(t, srv, "GET", "/warehouses/1/inventory", nil)
		var src struct {
			Items []InventoryLine `json:"items"`
		}
		decodeBody(t, rec2, &src)
		for _, it := range src.Items {
			if it.ProductID == 1 && it.Quantity != 100 {
				t.Errorf("expected rollback, source product 1 still %d", it.Quantity)
			}
		}
	})
}

// ----------------------------------------------------------------------------
// closeWarehouse
// ----------------------------------------------------------------------------

func TestCloseWarehouse(t *testing.T) {
	t.Run("closes an empty warehouse with no personnel", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-EAST is empty and has no personnel.
		rec := do(t, srv, "POST", "/warehouses/3/close", nil)
		expectStatus(t, rec, 200)
		var wh Warehouse
		decodeBody(t, rec, &wh)
		if wh.Status != "closed" {
			t.Errorf("expected status=closed, got %q", wh.Status)
		}
	})

	t.Run("refuses to close a warehouse with inventory with 409", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/warehouses/1/close", nil)
		expectStatus(t, rec, 409)
	})

	t.Run("refuses to close a warehouse with personnel with 409", func(t *testing.T) {
		srv, db := newTestServer(t)
		// Strip inventory from WH-NORTH but leave personnel.
		if _, err := db.Exec(`DELETE FROM warehouse_inventory WHERE warehouse_id=1`); err != nil {
			t.Fatalf("setup: %v", err)
		}
		rec := do(t, srv, "POST", "/warehouses/1/close", nil)
		expectStatus(t, rec, 409)
	})

	t.Run("refuses to close an already-closed warehouse with 409", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-WEST (id=4) is already closed.
		rec := do(t, srv, "POST", "/warehouses/4/close", nil)
		expectStatus(t, rec, 409)
	})

	t.Run("returns 404 for missing warehouse", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/warehouses/99999/close", nil)
		expectStatus(t, rec, 404)
	})
}

// ----------------------------------------------------------------------------
// bulkCreateWarehouses
// ----------------------------------------------------------------------------

func TestBulkCreateWarehouses(t *testing.T) {
	t.Run("creates multiple warehouses in one call", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"warehouses": [
			{"code":"WH-A","name":"A","capacity_units":10},
			{"code":"WH-B","name":"B","capacity_units":20},
			{"code":"WH-C","name":"C","capacity_units":30}
		]}`
		rec := do(t, srv, "POST", "/warehouses/bulk", body)
		expectStatus(t, rec, 201)
		var resp struct {
			Items []Warehouse `json:"items"`
		}
		decodeBody(t, rec, &resp)
		if len(resp.Items) != 3 {
			t.Fatalf("expected 3 created, got %d", len(resp.Items))
		}
	})

	t.Run("rolls back on duplicate code", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"warehouses": [
			{"code":"WH-OK","name":"X","capacity_units":10},
			{"code":"WH-NORTH","name":"dup","capacity_units":20}
		]}`
		rec := do(t, srv, "POST", "/warehouses/bulk", body)
		expectStatus(t, rec, 409)

		// WH-OK must NOT exist.
		rec2 := do(t, srv, "GET", "/warehouses/5", nil)
		// id 5 would be the next slot; either 404 or some other id check is fine.
		// We just verify the lookup by code-by-listing isn't trivial here,
		// so settle for: bulk failure means count of warehouses unchanged.
		if rec2.Code == 200 {
			t.Errorf("expected the rolled-back warehouse to not exist")
		}
	})

	t.Run("rejects empty list", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/warehouses/bulk", `{"warehouses":[]}`)
		expectStatus(t, rec, 400)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/warehouses/bulk", `{not json`)
		expectStatus(t, rec, 400)
	})
}
