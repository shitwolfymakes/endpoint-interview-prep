package main

import "net/http"

// health is a fully-implemented example handler. Use it as a reference for
// the basic shape: set headers, write status, write JSON body.
//
//	GET /health  ->  200 {"status":"ok"}
func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
