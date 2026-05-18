package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// writeJSON writes v as JSON with the given status code. Pass nil for v to
// write an empty body (handy for 204 No Content).
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// writeError writes a JSON error response of the form {"error": msg}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// decodeJSON decodes the request body into v. Returns an error suitable for
// reporting back to the client with 400 Bad Request.
func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	// Reject trailing junk after the JSON value.
	if dec.More() {
		return errors.New("body must contain a single JSON value")
	}
	return nil
}

// idParam reads the {id} path segment from the request and parses it as an
// int64. Returns 0 and an error if the segment is missing or non-numeric.
func idParam(r *http.Request, name string) (int64, error) {
	raw := chi.URLParam(r, name)
	if raw == "" {
		return 0, errors.New(name + " is required")
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, errors.New(name + " must be an integer")
	}
	return id, nil
}

// isUniqueConstraintErr is a pragmatic SQLite UNIQUE-constraint detector.
// In production you would type-assert to the driver's error type, but for
// this practice app the message check is good enough and works across both
// modernc.org/sqlite and mattn/go-sqlite3.
func isUniqueConstraintErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// isCheckConstraintErr detects CHECK-constraint violations (e.g. negative
// price, negative stock).
func isCheckConstraintErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "CHECK constraint failed")
}

// isFKeyConstraintErr detects FOREIGN KEY-constraint violations (e.g.
// deleting a product still referenced by an order).
func isFKeyConstraintErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "FOREIGN KEY constraint failed")
}
