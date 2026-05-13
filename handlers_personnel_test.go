package main

import (
	"fmt"
	"testing"
)

// ----------------------------------------------------------------------------
// createPersonnel
// ----------------------------------------------------------------------------

func TestCreatePersonnel(t *testing.T) {
	t.Run("creates a person with valid input", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{
			"email": "newhire@example.com",
			"name": "New Hire",
			"role": "clerk",
			"warehouse_id": 1
		}`
		rec := do(t, srv, "POST", "/personnel", body)
		expectStatus(t, rec, 201)
		var p Personnel
		decodeBody(t, rec, &p)
		if p.ID == 0 {
			t.Errorf("expected non-zero id")
		}
		if p.Email != "newhire@example.com" || p.Role != "clerk" {
			t.Errorf("response fields wrong: %+v", p)
		}
		if p.WarehouseID == nil || *p.WarehouseID != 1 {
			t.Errorf("expected warehouse_id=1, got %+v", p.WarehouseID)
		}
	})

	t.Run("creates a person without warehouse", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"email":"unassigned@example.com","name":"X","role":"clerk"}`
		rec := do(t, srv, "POST", "/personnel", body)
		expectStatus(t, rec, 201)
		var p Personnel
		decodeBody(t, rec, &p)
		if p.WarehouseID != nil {
			t.Errorf("expected nil warehouse_id, got %+v", p.WarehouseID)
		}
	})

	t.Run("persists the person", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"email":"persist@example.com","name":"X","role":"picker"}`
		rec := do(t, srv, "POST", "/personnel", body)
		expectStatus(t, rec, 201)
		var created Personnel
		decodeBody(t, rec, &created)

		rec2 := do(t, srv, "GET", fmt.Sprintf("/personnel/%d", created.ID), nil)
		expectStatus(t, rec2, 200)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/personnel", `{not json`)
		expectStatus(t, rec, 400)
	})

	t.Run("rejects missing email", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/personnel", `{"name":"x","role":"clerk"}`)
		expectStatus(t, rec, 400)
	})

	t.Run("rejects invalid role", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"email":"role@example.com","name":"x","role":"wizard"}`
		rec := do(t, srv, "POST", "/personnel", body)
		expectStatus(t, rec, 400)
	})

	t.Run("rejects duplicate email with 409", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"email":"dana@example.com","name":"dup","role":"clerk"}`
		rec := do(t, srv, "POST", "/personnel", body)
		expectStatus(t, rec, 409)
	})

	t.Run("rejects assignment to closed warehouse with 409", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-WEST (id=4) is closed.
		body := `{"email":"closed@example.com","name":"x","role":"clerk","warehouse_id":4}`
		rec := do(t, srv, "POST", "/personnel", body)
		expectStatus(t, rec, 409)
	})
}

// ----------------------------------------------------------------------------
// updatePersonnel
// ----------------------------------------------------------------------------

func TestUpdatePersonnel(t *testing.T) {
	t.Run("updates an existing person", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{
			"email": "dana@example.com",
			"name": "Dana Renamed",
			"role": "manager",
			"warehouse_id": 1
		}`
		rec := do(t, srv, "PUT", "/personnel/1", body)
		expectStatus(t, rec, 200)
		var p Personnel
		decodeBody(t, rec, &p)
		if p.Name != "Dana Renamed" {
			t.Errorf("update did not stick: %+v", p)
		}
	})

	t.Run("returns 404 for missing id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"email":"x@example.com","name":"x","role":"clerk"}`
		rec := do(t, srv, "PUT", "/personnel/99999", body)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 409 on email collision", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// Try to give person 1 person 2's email.
		body := `{"email":"evan@example.com","name":"x","role":"clerk"}`
		rec := do(t, srv, "PUT", "/personnel/1", body)
		expectStatus(t, rec, 409)
	})

	t.Run("returns 400 on invalid id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"email":"x@example.com","name":"x","role":"clerk"}`
		rec := do(t, srv, "PUT", "/personnel/abc", body)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// deletePersonnel
// ----------------------------------------------------------------------------

func TestDeletePersonnel(t *testing.T) {
	t.Run("deletes a person", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "DELETE", "/personnel/1", nil)
		expectStatus(t, rec, 204)

		rec2 := do(t, srv, "GET", "/personnel/1", nil)
		expectStatus(t, rec2, 404)
	})

	t.Run("returns 404 for missing id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "DELETE", "/personnel/99999", nil)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 400 on invalid id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "DELETE", "/personnel/abc", nil)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// getPersonnel
// ----------------------------------------------------------------------------

func TestGetPersonnel(t *testing.T) {
	t.Run("returns person with warehouse embedded", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/personnel/1", nil)
		expectStatus(t, rec, 200)
		var p Personnel
		decodeBody(t, rec, &p)
		if p.ID != 1 || p.Email != "dana@example.com" {
			t.Errorf("got id=%d email=%q, want 1 dana@example.com", p.ID, p.Email)
		}
		if p.Warehouse == nil || p.Warehouse.ID != 1 {
			t.Errorf("expected warehouse embedded, got %+v", p.Warehouse)
		}
	})

	t.Run("returns unassigned person with nil warehouse", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// Person 6 (Ivan) is unassigned and terminated.
		rec := do(t, srv, "GET", "/personnel/6", nil)
		expectStatus(t, rec, 200)
		var p Personnel
		decodeBody(t, rec, &p)
		if p.Warehouse != nil {
			t.Errorf("expected nil warehouse for unassigned person, got %+v", p.Warehouse)
		}
	})

	t.Run("returns 404 for missing id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/personnel/99999", nil)
		expectStatus(t, rec, 404)
	})

	t.Run("returns 400 for non-integer id", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "GET", "/personnel/abc", nil)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// reassignPersonnel
// ----------------------------------------------------------------------------

func TestReassignPersonnel(t *testing.T) {
	t.Run("moves a person to a different warehouse", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// Person 1 is at WH-NORTH; move to WH-SOUTH (id=2).
		rec := do(t, srv, "PATCH", "/personnel/1/reassign", `{"warehouse_id": 2}`)
		expectStatus(t, rec, 200)
		var p Personnel
		decodeBody(t, rec, &p)
		if p.WarehouseID == nil || *p.WarehouseID != 2 {
			t.Errorf("expected warehouse_id=2, got %+v", p.WarehouseID)
		}
	})

	t.Run("unassigns when warehouse_id is null", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "PATCH", "/personnel/1/reassign", `{"warehouse_id": null}`)
		expectStatus(t, rec, 200)
		var p Personnel
		decodeBody(t, rec, &p)
		if p.WarehouseID != nil {
			t.Errorf("expected nil warehouse_id, got %+v", p.WarehouseID)
		}
	})

	t.Run("rejects target that is closed with 409", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// WH-WEST (id=4) is closed.
		rec := do(t, srv, "PATCH", "/personnel/1/reassign", `{"warehouse_id": 4}`)
		expectStatus(t, rec, 409)
	})

	t.Run("rejects target that doesn't exist with 409", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "PATCH", "/personnel/1/reassign", `{"warehouse_id": 99999}`)
		expectStatus(t, rec, 409)
	})

	t.Run("refuses to reassign a terminated person with 409", func(t *testing.T) {
		srv, _ := newTestServer(t)
		// Person 6 (Ivan) is terminated.
		rec := do(t, srv, "PATCH", "/personnel/6/reassign", `{"warehouse_id": 1}`)
		expectStatus(t, rec, 409)
	})

	t.Run("returns 404 for missing person", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "PATCH", "/personnel/99999/reassign", `{"warehouse_id": 1}`)
		expectStatus(t, rec, 404)
	})

	t.Run("rejects bad JSON", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "PATCH", "/personnel/1/reassign", `{not json`)
		expectStatus(t, rec, 400)
	})
}

// ----------------------------------------------------------------------------
// bulkCreatePersonnel
// ----------------------------------------------------------------------------

func TestBulkCreatePersonnel(t *testing.T) {
	t.Run("creates multiple in one call", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"personnel": [
			{"email":"a@example.com","name":"A","role":"clerk"},
			{"email":"b@example.com","name":"B","role":"picker"},
			{"email":"c@example.com","name":"C","role":"driver"}
		]}`
		rec := do(t, srv, "POST", "/personnel/bulk", body)
		expectStatus(t, rec, 201)
		var resp struct {
			Items []Personnel `json:"items"`
		}
		decodeBody(t, rec, &resp)
		if len(resp.Items) != 3 {
			t.Fatalf("expected 3 created, got %d", len(resp.Items))
		}
	})

	t.Run("rolls back on duplicate email", func(t *testing.T) {
		srv, _ := newTestServer(t)
		body := `{"personnel": [
			{"email":"newperson@example.com","name":"X","role":"clerk"},
			{"email":"dana@example.com","name":"dup","role":"clerk"}
		]}`
		rec := do(t, srv, "POST", "/personnel/bulk", body)
		expectStatus(t, rec, 409)
	})

	t.Run("rejects empty list", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/personnel/bulk", `{"personnel":[]}`)
		expectStatus(t, rec, 400)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		srv, _ := newTestServer(t)
		rec := do(t, srv, "POST", "/personnel/bulk", `{not json`)
		expectStatus(t, rec, 400)
	})
}
