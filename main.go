package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	dbPath := envOr("DB_PATH", "app.db")
	addr := envOr("ADDR", ":8080")

	db, err := openDB(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := applySchema(db); err != nil {
		log.Fatalf("apply schema: %v", err)
	}

	log.Printf("listening on %s (db=%s)", addr, dbPath)
	if err := http.ListenAndServe(addr, newRouter(db)); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
