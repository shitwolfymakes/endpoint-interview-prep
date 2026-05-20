package main

import (
	"database/sql"
	"errors"
	"net/http"
	"slices"
	"time"
)

// ============================================================================
// STUBS — implement these. Run `go test ./...` to see which tests fail.
// ============================================================================

// createPersonnel: POST /personnel
//
// Request body:
//
//	{
//	  "email":        string (required, unique)
//	  "name":         string (required)
//	  "role":         string (required; one of: manager, clerk, picker, driver)
//	  "warehouse_id": int    (optional; if present, must reference an active warehouse)
//	}
//
// Status codes:
//
//	201  on success — return the created person
//	400  on invalid JSON, missing required fields, or invalid role
//	409  on duplicate email (UNIQUE), or warehouse_id referring to a
//	     non-existent or closed warehouse
//
// Mirrors createProduct.
func createPersonnel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// validate JSON request
		var req Personnel
		err := decodeJSON(r, &req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// validate request values
		if req.Email == "" {
			writeError(w, http.StatusBadRequest, "email cannot be empty")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "name cannot be empty")
			return
		}
		validRoles := []string{"manager", "clerk", "picker", "driver"}
		if !slices.Contains(validRoles, req.Role) {
			writeError(w, http.StatusBadRequest, "invalid role")
			return
		}

		// create tx
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer tx.Rollback()

		// check that warehouse exists and is active
		if req.WarehouseID != nil {
			var status string
			err = tx.QueryRowContext(r.Context(),
				`SELECT status FROM warehouses WHERE id = ?`,
				*req.WarehouseID).
				Scan(&status)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					writeError(w, http.StatusNotFound, err.Error())
				}
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if status == "closed" {
				writeError(w, http.StatusConflict,
					"warehouse cannot be closed",
				)
				return
			}
		}

		// run the sql
		result, err := tx.ExecContext(r.Context(),
			`INSERT INTO personnel
			(email, name, role, warehouse_id, hired_at)
			VALUES
			(?, ?, ?, ?, ?)`,
			req.Email, req.Name, req.Role, req.WarehouseID, time.Now(),
		)
		if err != nil {
			if isUniqueConstraintErr(err) {
				writeError(w, http.StatusConflict, err.Error())
				return
			}
			if isCheckConstraintErr(err) {
				writeError(w, http.StatusConflict, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		id, err := result.LastInsertId()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// query row for return
		var resp Personnel
		err = tx.QueryRowContext(r.Context(),
			`SELECT id, email, name, role, warehouse_id,
			hired_at, 
			COALESCE(terminated_at, '')
			FROM personnel WHERE id = ?`, id).
			Scan(&resp.ID, &resp.Email, &resp.Name, &resp.Role,
				&resp.WarehouseID, &resp.HiredAt, &resp.TerminatedAt,
			)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// commit tx
		if err := tx.Commit(); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, resp)
	}
}

// updatePersonnel: PUT /personnel/{id}
//
// Replaces every editable field. Same body shape as createPersonnel.
//
// Status codes:
//
//	200  on success — return the updated person
//	400  on invalid JSON, invalid {id}, or invalid role
//	404  if no person has that id
//	409  on duplicate email, or warehouse_id referring to a non-existent
//	     or closed warehouse
func updatePersonnel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = db
		writeError(w, http.StatusNotImplemented, "TODO: implement updatePersonnel")
	}
}

// deletePersonnel: DELETE /personnel/{id}
//
// Status codes:
//
//	204  on success (empty body)
//	400  on invalid {id}
//	404  if no person has that id
//
// Note: this is a hard delete. To mark someone as no-longer-working, set
// their terminated_at via the regular PUT instead.
func deletePersonnel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = db
		writeError(w, http.StatusNotImplemented, "TODO: implement deletePersonnel")
	}
}

// getPersonnel: GET /personnel/{id}
//
// Returns one person with their current warehouse embedded (via JOIN). When
// the person is unassigned (warehouse_id IS NULL), the embedded warehouse
// should be null/omitted.
//
// Status codes:
//
//	200  on success
//	400  on non-integer {id}
//	404  if no person has that id
//
// Mirrors getOrder, which embeds Customer the same way.
func getPersonnel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse path param
		id, err := idParam(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// query for return
		var resp Personnel
		err = db.QueryRowContext(r.Context(),
			`SELECT id, email, name, role, warehouse_id,
			hired_at, 
			COALESCE(terminated_at, '')
			FROM personnel WHERE id = ?`, id).
			Scan(&resp.ID, &resp.Email, &resp.Name, &resp.Role,
				&resp.WarehouseID, &resp.HiredAt, &resp.TerminatedAt,
			)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if resp.WarehouseID != nil {
			var wh Warehouse
			err = db.QueryRowContext(r.Context(),
				`SELECT id, code, name, address, capacity_units, status,
				created_at, updated_at
				FROM warehouses
				WHERE id = ?`, id).
				Scan(&wh.ID, &wh.Code, &wh.Name,
					&wh.Address, &wh.CapacityUnits,
					&wh.Status, &wh.CreatedAt,
					&wh.UpdatedAt)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					writeError(w, http.StatusNotFound, err.Error())
					return
				}
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			resp.Warehouse = &wh
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// reassignPersonnel: PATCH /personnel/{id}/reassign
//
// Moves a person to a different warehouse atomically. Passing
// "warehouse_id": null is allowed and unassigns the person.
//
// Request body:
//
//	{ "warehouse_id": int | null }
//
// Status codes:
//
//	200  on success — return the updated person
//	400  on invalid JSON or invalid {id}
//	404  if no person has that id
//	409  if the target warehouse doesn't exist, is closed, or the person
//	     is already terminated (terminated_at IS NOT NULL)
//
// Hint: a single guarded UPDATE works. Example shape:
//
//	UPDATE personnel
//	SET    warehouse_id = ?
//	WHERE  id = ? AND terminated_at IS NULL
//	  AND  (? IS NULL OR EXISTS (SELECT 1 FROM warehouses
//	                             WHERE id = ? AND status = 'active'))
//
// RowsAffected == 0 -> distinguish missing vs terminated vs bad target with
// a follow-up SELECT.
// Mirrors adjustStock.
func reassignPersonnel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = db
		writeError(w, http.StatusNotImplemented, "TODO: implement reassignPersonnel")
	}
}

// bulkCreatePersonnel: POST /personnel/bulk
//
// Inserts every row in a single transaction; roll the whole batch back on
// any failure. Same conventions as bulkCreateProducts.
//
// Request body:
//
//	{ "personnel": [ {email, name, role, warehouse_id?}, ... ] }
//
// Status codes:
//
//	201  on success — return {"items": [...]}
//	400  on invalid JSON, empty list, or any item failing validation
//	409  if any email collides with an existing person, or any warehouse_id
//	     refers to a missing or closed warehouse
func bulkCreatePersonnel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = db
		writeError(w, http.StatusNotImplemented, "TODO: implement bulkCreatePersonnel")
	}
}
