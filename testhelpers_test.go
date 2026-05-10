package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

// newTestServer spins up a fresh router backed by a fresh SQLite database
// (in a temp dir, auto-cleaned by t.Cleanup) with the schema applied and
// seed data inserted. Each test gets an isolated database — there is no
// state leak between tests.
func newTestServer(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := openDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := applySchema(db); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	if err := applySeed(db); err != nil {
		t.Fatalf("apply seed: %v", err)
	}
	return newRouter(db), db
}

// do performs an HTTP request against srv and returns the response recorder.
// Pass a string for body to send JSON; pass nil for no body.
func do(t *testing.T, srv http.Handler, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != nil {
		switch b := body.(type) {
		case string:
			r = strings.NewReader(b)
		case []byte:
			r = strings.NewReader(string(b))
		default:
			buf, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("marshal body: %v", err)
			}
			r = strings.NewReader(string(buf))
		}
	}
	req := httptest.NewRequest(method, target, r)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

// decodeBody decodes rec.Body into v. Fails the test on parse error.
func decodeBody(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("decode response body %q: %v", rec.Body.String(), err)
	}
}

// expectStatus fails the test if rec.Code != want.
func expectStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("expected status %d, got %d. body: %s", want, rec.Code, rec.Body.String())
	}
}
