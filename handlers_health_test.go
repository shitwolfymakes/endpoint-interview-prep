package main

import "testing"

func TestHealth(t *testing.T) {
	srv, _ := newTestServer(t)

	rec := do(t, srv, "GET", "/health", nil)
	expectStatus(t, rec, 200)

	var got map[string]string
	decodeBody(t, rec, &got)
	if got["status"] != "ok" {
		t.Fatalf("expected status=ok, got %q", got["status"])
	}
}
