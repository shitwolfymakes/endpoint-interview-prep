package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// ============================================================================
// REFERENCE IMPLEMENTATIONS — study these before implementing the stubs
// ============================================================================

// listProducts is a fully-implemented example showing query-parameter
// parsing, optional filtering, pagination, and scanning multiple rows.
//
//	GET /products?category=&min_price=&max_price=&q=&limit=&offset=
//
// Response: {"items": [...], "limit": 50, "offset": 0}
func listProducts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		sqlStr := `SELECT id, sku, name, description, price_cents,
		                  stock_quantity, category, created_at, updated_at
		           FROM products WHERE 1=1`
		var args []any

		if v := q.Get("category"); v != "" {
			sqlStr += " AND category = ?"
			args = append(args, v)
		}
		if v := q.Get("min_price"); v != "" {
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "min_price must be an integer")
				return
			}
			sqlStr += " AND price_cents >= ?"
			args = append(args, n)
		}
		if v := q.Get("max_price"); v != "" {
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "max_price must be an integer")
				return
			}
			sqlStr += " AND price_cents <= ?"
			args = append(args, n)
		}
		if v := q.Get("q"); v != "" {
			sqlStr += " AND LOWER(name) LIKE ?"
			args = append(args, "%"+strings.ToLower(v)+"%")
		}

		limit := 50
		if v := q.Get("limit"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 {
				writeError(w, http.StatusBadRequest, "limit must be a positive integer")
				return
			}
			if n > 200 {
				n = 200
			}
			limit = n
		}
		offset := 0
		if v := q.Get("offset"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				writeError(w, http.StatusBadRequest, "offset must be non-negative")
				return
			}
			offset = n
		}
		sqlStr += " ORDER BY id LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

		rows, err := db.QueryContext(r.Context(), sqlStr, args...)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		out := []Product{}
		for rows.Next() {
			var p Product
			if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Description,
				&p.PriceCents, &p.StockQuantity, &p.Category,
				&p.CreatedAt, &p.UpdatedAt); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			out = append(out, p)
		}
		if err := rows.Err(); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"items":  out,
			"limit":  limit,
			"offset": offset,
		})
	}
}

// getProduct is a fully-implemented example showing path-parameter parsing,
// single-row scanning, and the sql.ErrNoRows -> 404 pattern.
//
//	GET /products/{id}  ->  200 with product, 404 if not found, 400 on bad id
func getProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := idParam(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		var p Product
		err = db.QueryRowContext(r.Context(), `
			SELECT id, sku, name, description, price_cents,
			       stock_quantity, category, created_at, updated_at
			FROM products WHERE id = ?`, id).
			Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.PriceCents,
				&p.StockQuantity, &p.Category, &p.CreatedAt, &p.UpdatedAt)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

// ============================================================================
// STUBS — implement these. Run `go test ./...` to see which tests fail.
// ============================================================================

// createProduct: POST /products
//
// Request body:
//
//	{
//	  "sku":            string  (required, unique)
//	  "name":           string  (required)
//	  "description":    string  (optional, defaults to "")
//	  "price_cents":    int     (required, >= 0)
//	  "stock_quantity": int     (optional, defaults to 0, >= 0)
//	  "category":       string  (optional, defaults to "")
//	}
//
// Status codes:
//
//	201  on success — return the created product (with its new id) as JSON
//	400  on invalid JSON, missing required fields, or negative numeric values
//	409  on duplicate SKU (UNIQUE constraint violation — see isUniqueConstraintErr)
//
// Hints:
//   - decodeJSON(r, &input) handles parsing
//   - Use db.ExecContext + result.LastInsertId() to get the new id
//   - SELECT the row back so timestamps are populated, OR INSERT ... RETURNING
func createProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// validate JSON request
		var req Product
		err := decodeJSON(r, &req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// validate request values
		if req.SKU == "" {
			writeError(w, http.StatusBadRequest, "SKU cannot be empty")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "Name cannot be empty")
			return
		}
		if req.PriceCents < 0 {
			writeError(w, http.StatusBadRequest, "Price cannot be negative")
			return
		}
		if req.StockQuantity < 0 {
			writeError(w, http.StatusBadRequest, "Quantity cannot be negative")
			return
		}

		// Insert the row
		result, err := db.ExecContext(r.Context(),
			`INSERT INTO products 
			(sku, name, description, price_cents, stock_quantity, category)
			VALUES
			(?, ?, ?, ?, ?, ?)`,
			req.SKU, req.Name, req.Description, req.PriceCents,
			req.StockQuantity, req.Category,
		)
		if err != nil {
			if isUniqueConstraintErr(err) {
				writeError(w, http.StatusConflict, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// query and return the new row
		id, err := result.LastInsertId()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		var response Product
		err = db.QueryRowContext(r.Context(), `
			SELECT id, sku, name, description, price_cents,
			       stock_quantity, category, created_at, updated_at
			FROM products WHERE id = ?`, id).
			Scan(&response.ID, &response.SKU, &response.Name,
				&response.Description, &response.PriceCents,
				&response.StockQuantity, &response.Category,
				&response.CreatedAt, &response.UpdatedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, response)
	}
}

// updateProduct: PUT /products/{id}
//
// Replaces every editable field on the product. Body has the same shape as
// createProduct (all fields required).
//
// Status codes:
//
//	200  on success — return the updated product
//	400  on invalid JSON or invalid {id}
//	404  if no product has that id
//	409  if the new SKU collides with another product
//
// Hint: bump updated_at to CURRENT_TIMESTAMP in the UPDATE.
func updateProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// validate JSON request
		var req Product
		err := decodeJSON(r, &req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// parse path params
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// validate request values
		if req.SKU == "" {
			writeError(w, http.StatusBadRequest, "SKU cannot be empty")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "Name cannot be empty")
			return
		}
		if req.PriceCents < 0 {
			writeError(w, http.StatusBadRequest, "Price cannot be negative")
			return
		}
		if req.StockQuantity < 0 {
			writeError(w, http.StatusBadRequest, "Quantity cannot be negative")
			return
		}

		// execute the db call
		result, err := db.ExecContext(r.Context(),
			`UPDATE products SET
			sku = ?,
			name = ?,
			description = ?,
			price_cents = ?,
			stock_quantity = ?,
			category = ?,
			updated_at = ?
			WHERE id = ?`,
			req.SKU, req.Name, req.Description, req.PriceCents,
			req.StockQuantity, req.Category,
			time.Now(),
			id,
		)
		if err != nil {
			// check that SKU isn't changed to one that already exists
			if isUniqueConstraintErr(err) {
				writeError(w, http.StatusConflict, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		// check that the ID was found
		found, _, err := rowsAffected(result)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !found {
			writeError(w, http.StatusNotFound, fmt.Sprintf("id %d not found", id))
			return
		}

		// query and return the new row
		var response Product
		err = db.QueryRowContext(r.Context(), `
			SELECT id, sku, name, description, price_cents,
			       stock_quantity, category, created_at, updated_at
			FROM products WHERE id = ?`, id).
			Scan(&response.ID, &response.SKU, &response.Name,
				&response.Description, &response.PriceCents,
				&response.StockQuantity, &response.Category,
				&response.CreatedAt, &response.UpdatedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

// deleteProduct: DELETE /products/{id}
//
// Status codes:
//
//	204  on success (empty body)
//	400  on invalid {id}
//	404  if no product has that id (check result.RowsAffected())
//	409  if the product is referenced by an order (FK violation) — return a
//	     helpful error message
func deleteProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse path params
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		result, err := db.ExecContext(r.Context(),
			`DELETE FROM products
			WHERE id = ?`,
			id,
		)
		if err != nil {
			// check for foreign key errors
			if isFKeyConstraintErr(err) {
				writeError(w, http.StatusConflict, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		// check that the ID was found
		found, _, err := rowsAffected(result)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !found {
			writeError(w, http.StatusNotFound, fmt.Sprintf("id %d not found", id))
			return
		}

		writeJSON(w, http.StatusNoContent, nil)
	}
}

// adjustStock: PATCH /products/{id}/stock
//
// Adjusts a product's stock by a signed delta. Used both to receive new
// inventory (positive delta) and to write off damaged units (negative delta).
//
// Request body:
//
//	{ "delta": int }      // can be negative
//
// Status codes:
//
//	200  on success — return the updated product
//	400  on invalid JSON, missing delta, or non-integer {id}
//	404  if no product has that id
//	422  if applying the delta would make stock_quantity negative
//
// Hint: do this atomically with a single UPDATE that includes a guard, e.g.
//
//	UPDATE products SET stock_quantity = stock_quantity + ?, updated_at = ...
//	WHERE id = ? AND stock_quantity + ? >= 0
//
// then check RowsAffected: 0 means the guard rejected it (or the row is missing).
func adjustStock(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// validate JSON request
		var req struct {
			Delta *int `json:"delta"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// parse path params
		id, err := idParam(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// execute db call
		result, err := db.ExecContext(r.Context(),
			`UPDATE products SET
			stock_quantity = stock_quantity + ?,
			updated_at = ?
			WHERE id = ?
			`,
			req.Delta,
			time.Now(),
			id,
			req.Delta,
		)
		if err != nil {
			// check constraint for negative
			if isCheckConstraintErr(err) {
				writeError(w, http.StatusUnprocessableEntity, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		// check that id was found
		found, _, err := rowsAffected(result)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !found {
			writeError(w, http.StatusNotFound, fmt.Sprintf("id %d not found", id))
			return
		}

		// query and return the row
		var response Product
		err = db.QueryRowContext(r.Context(), `
			SELECT id, sku, name, description, price_cents,
			       stock_quantity, category, created_at, updated_at
			FROM products WHERE id = ?`, id).
			Scan(&response.ID, &response.SKU, &response.Name,
				&response.Description, &response.PriceCents,
				&response.StockQuantity, &response.Category,
				&response.CreatedAt, &response.UpdatedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

// bulkCreateProducts: POST /products/bulk
//
// Inserts a batch of products in a single database transaction. If any one
// fails (validation or UNIQUE collision), the entire batch is rolled back.
//
// Request body:
//
//	{ "products": [ {sku, name, price_cents, ...}, ... ] }
//
// Status codes:
//
//	201  on success — return {"items": [...]} with all created products
//	400  on invalid JSON, empty list, or any product failing validation
//	409  if any SKU collides with an existing product or a duplicate inside the batch
//
// Hints:
//   - db.BeginTx -> tx.ExecContext per row -> tx.Commit / tx.Rollback
//   - Defer tx.Rollback() right after BeginTx; Commit makes the rollback a no-op
func bulkCreateProducts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// validate JSON request
		var req struct {
			Products []Product `json:"products"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if len(req.Products) == 0 {
			writeError(w, http.StatusBadRequest, "list cannot be empty")
		}

		// validate request values
		for _, prod := range req.Products {
			if prod.SKU == "" {
				writeJSON(w, http.StatusBadRequest, prod)
				return
			}
			if prod.Name == "" {
				writeJSON(w, http.StatusBadRequest, prod)
				return
			}
			if prod.PriceCents < 0 {
				writeJSON(w, http.StatusBadRequest, prod)
				return
			}
			if prod.StockQuantity < 0 {
				writeJSON(w, http.StatusBadRequest, prod)
				return
			}
		}

		_ = db
		writeError(w, http.StatusNotImplemented, "TODO: implement bulkCreateProducts")
	}
}
