# Study guide — Go HTTP handlers with chi + database/sql

A pragmatic cheat-sheet for the patterns you'll need. Everything here is in
the standard library or in `github.com/go-chi/chi/v5`. Copy snippets
freely while practicing.

## Table of contents

1. [The handler shape](#the-handler-shape)
2. [chi routing & path params](#chi-routing--path-params)
3. [Reading the request body](#reading-the-request-body)
4. [Reading query parameters](#reading-query-parameters)
5. [Writing JSON responses](#writing-json-responses)
6. [database/sql basics](#databasesql-basics)
7. [Single-row reads & 404 handling](#single-row-reads--404-handling)
8. [Multi-row reads](#multi-row-reads)
9. [INSERT and getting the new id](#insert-and-getting-the-new-id)
10. [UPDATE / DELETE with RowsAffected](#update--delete-with-rowsaffected)
11. [Transactions](#transactions)
12. [Detecting SQLite constraint errors](#detecting-sqlite-constraint-errors)
13. [HTTP status codes — when to use which](#http-status-codes--when-to-use-which)
14. [Common pitfalls](#common-pitfalls)
15. [Things to memorize before walking in](#things-to-memorize-before-walking-in)

---

## The handler shape

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    // 1. parse input  (path params, query params, body)
    // 2. talk to the database
    // 3. write the response
}
```

To inject a `*sql.DB` (or any other dependency), use a closure:

```go
func myHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // ...use db...
    }
}
```

Every handler in this codebase uses the closure form. You wire it up in
`server.go` like:

```go
r.Get("/products/{id}", getProduct(db))
```

## chi routing & path params

```go
r := chi.NewRouter()
r.Use(middleware.Recoverer)            // turn panics into 500s
r.Get("/health", health)               // method + path -> handler

r.Route("/products", func(r chi.Router) {
    r.Get("/", listProducts(db))       // GET  /products
    r.Post("/", createProduct(db))     // POST /products
    r.Get("/{id}", getProduct(db))     // GET  /products/123
})
```

Inside a handler, read path params with `chi.URLParam`:

```go
raw := chi.URLParam(r, "id")           // returns "" if missing
id, err := strconv.ParseInt(raw, 10, 64)
if err != nil { /* 400 */ }
```

This codebase wraps that into [`idParam`](../helpers.go) — use it.

## Reading the request body

```go
type input struct {
    SKU        string `json:"sku"`
    Name       string `json:"name"`
    PriceCents int64  `json:"price_cents"`
}

var in input
if err := decodeJSON(r, &in); err != nil {
    writeError(w, http.StatusBadRequest, err.Error())
    return
}
```

The `decodeJSON` helper in this repo turns on `DisallowUnknownFields` so
typos in the body are surfaced instead of silently ignored.

> Pointers vs zero values: `json:"price_cents"` on an `int64` can't tell
> "missing" from `0`. If you need to distinguish them, declare the field
> as `*int64` and check for `nil`. We don't bother in this practice repo.

## Reading query parameters

```go
q := r.URL.Query()
category := q.Get("category")          // "" if missing

if v := q.Get("limit"); v != "" {
    n, err := strconv.Atoi(v)
    if err != nil || n < 1 { /* 400 */ }
    limit = n
}
```

`q.Get` returns "". `q.Has("foo")` distinguishes missing from `?foo=`.

## Writing JSON responses

```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
_ = json.NewEncoder(w).Encode(payload)
```

Use `writeJSON(w, status, value)` from helpers.go to keep handlers tidy.

> Always set the status BEFORE writing the body. After the first write,
> `WriteHeader` is a no-op (and prints a warning to stderr).

## database/sql basics

The driver registers under `"sqlite"` (modernc) or `"sqlite3"` (mattn).
This repo uses modernc, so `sql.Open("sqlite", path)`.

Always use the placeholder form (`?`) — never string-concatenate user
input into SQL.

```go
rows, err := db.QueryContext(r.Context(), `
    SELECT id, name FROM products WHERE category = ?`, "widgets")
```

Always pass `r.Context()` as the first arg — it lets the request cancel
the query if the client hangs up.

## Single-row reads & 404 handling

```go
var p Product
err := db.QueryRowContext(r.Context(),
    `SELECT id, name FROM products WHERE id = ?`, id,
).Scan(&p.ID, &p.Name)

if errors.Is(err, sql.ErrNoRows) {
    writeError(w, http.StatusNotFound, "product not found")
    return
}
if err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
}
writeJSON(w, http.StatusOK, p)
```

`sql.ErrNoRows` is THE pattern — you'll write it for every "get by id"
endpoint. Memorize this block.

## Multi-row reads

```go
rows, err := db.QueryContext(r.Context(), `SELECT id, name FROM products`)
if err != nil { /* 500 */ }
defer rows.Close()                          // ALWAYS

out := []Product{}                          // not nil — JSON encodes [] not null
for rows.Next() {
    var p Product
    if err := rows.Scan(&p.ID, &p.Name); err != nil { /* 500 */ }
    out = append(out, p)
}
if err := rows.Err(); err != nil { /* 500 */ }
```

The two easy bugs:

- forgetting `defer rows.Close()` — leaks DB connections;
- forgetting `rows.Err()` after the loop — silently swallows errors.

## INSERT and getting the new id

```go
res, err := db.ExecContext(r.Context(),
    `INSERT INTO products (sku, name, price_cents) VALUES (?, ?, ?)`,
    in.SKU, in.Name, in.PriceCents)
if isUniqueConstraintErr(err) {
    writeError(w, http.StatusConflict, "sku already exists")
    return
}
if err != nil { /* 500 */ }

id, err := res.LastInsertId()
if err != nil { /* 500 */ }
```

To return the full row (including DEFAULT timestamps), do a follow-up
`SELECT ... WHERE id = ?` — or use SQLite's `INSERT ... RETURNING *`
(supported on modern SQLite) and `QueryRowContext` instead of `Exec`.

## UPDATE / DELETE with RowsAffected

```go
res, err := db.ExecContext(r.Context(),
    `UPDATE products SET name = ? WHERE id = ?`, in.Name, id)
if err != nil { /* 500 */ }

n, err := res.RowsAffected()
if err != nil { /* 500 */ }
if n == 0 {
    writeError(w, http.StatusNotFound, "product not found")
    return
}
```

`RowsAffected` is how you tell "no row matched" from "update worked".
Don't `SELECT` first to check existence — it's a race condition and an
extra round-trip.

## Transactions

```go
tx, err := db.BeginTx(r.Context(), nil)
if err != nil { /* 500 */ }
defer tx.Rollback()                          // safe: Commit makes this a no-op

if _, err := tx.ExecContext(r.Context(), `UPDATE ...`); err != nil {
    writeError(w, http.StatusUnprocessableEntity, err.Error())
    return
}
if _, err := tx.ExecContext(r.Context(), `INSERT ...`); err != nil { /* 500 */ }

if err := tx.Commit(); err != nil { /* 500 */ }
```

Two rules:

1. After `BeginTx`, every DB call must go through `tx`, not `db`.
   Mixing them defeats the transaction.
2. Defer `tx.Rollback()` immediately. After `Commit()` it's a no-op.

### Atomic compare-and-set in a single statement

For "decrement stock if there's enough", you don't need a SELECT:

```sql
UPDATE products
SET    stock_quantity = stock_quantity - ?
WHERE  id = ? AND stock_quantity >= ?
```

If `RowsAffected() == 0`, either the row is gone or stock was insufficient
— fall back to a SELECT to tell which.

## Detecting SQLite constraint errors

The pragmatic check (works across drivers):

```go
if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
    writeError(w, http.StatusConflict, "duplicate key")
    return
}
```

This repo provides `isUniqueConstraintErr` and `isCheckConstraintErr`.

In production you'd type-assert to the driver's error struct, but for
interview prep the message check is fine — and avoids tying handlers to
a specific driver.

## HTTP status codes — when to use which

| Code | Name                  | Use it for                                                       |
|-----:|:----------------------|:-----------------------------------------------------------------|
| 200  | OK                    | Successful GET / PUT / PATCH                                     |
| 201  | Created               | Successful POST that created a resource (return the new entity)  |
| 204  | No Content            | Successful DELETE; empty body                                    |
| 400  | Bad Request           | Bad request: malformed body, unparseable id, missing field       |
| 401  | Unauthorized          | Missing/invalid auth (we don't use this in this repo)            |
| 403  | Forbidden             | Authenticated but not allowed                                    |
| 404  | Not Found             | The thing you asked for doesn't exist                            |
| 409  | Conflict              | Conflict: duplicate UNIQUE, FK in use, illegal state transition  |
| 422  | Unprocessable Entity  | Body parses but violates a business rule (e.g. not enough stock) |
| 500  | Internal Server Error | Unexpected server / DB error                                     |

When in doubt between 400 and 422: 400 = "I can't even parse this", 422 =
"I parsed it fine, but it's not okay". When in doubt between 409 and 422:
409 = "conflict with existing data", 422 = "this object is unprocessable
on its own". Pick one and be consistent.

## Common pitfalls

- **Returning `nil` instead of `[]`.** `json.Marshal(nil)` produces `null`,
  not `[]`. Always initialize with `out := []T{}`.
- **Setting headers after writing the body.** `w.WriteHeader` and
  `w.Header().Set(...)` must come before any `w.Write` call.
- **Forgetting `defer rows.Close()` and `rows.Err()`.** Leaks and silent
  errors.
- **`%v` in error messages.** Go's `%v` formatting strips the type. Prefer
  `err.Error()` directly when shaping a user-facing message.
- **JSON tags on private fields.** `json:"foo"` on a lowercase field is
  ignored. Field names in Go must start with a capital letter to be
  exported and serialized.
- **Time.Time + SQLite.** SQLite stores times as TEXT. This repo uses
  `string` fields for timestamps to sidestep the conversion. If the
  interviewer wants `time.Time`, scan into a `string` first and parse
  with `time.Parse("2006-01-02 15:04:05", s)`.
- **Forgetting to bump `updated_at`.** Easy to miss on UPDATEs. Bake it
  into every UPDATE: `... SET ..., updated_at = strftime(...)`.
- **Mutating params from a path that wasn't validated.** Always validate
  `id` parses, body decodes, required fields present BEFORE touching
  the DB.

## Things to memorize before walking in

These are the templates you should be able to type from muscle memory.
Practice them until you can.

```go
// 1. Path-param + 404 pattern
id, err := idParam(r, "id")
if err != nil {
    writeError(w, http.StatusBadRequest, err.Error())
    return
}

var p Product
err = db.QueryRowContext(r.Context(),
    `SELECT id, name FROM products WHERE id = ?`, id,
).Scan(&p.ID, &p.Name)
if errors.Is(err, sql.ErrNoRows) {
    writeError(w, http.StatusNotFound, "not found")
    return
}
if err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
}
writeJSON(w, http.StatusOK, p)
```

```go
// 2. Decode + insert + return-with-id pattern
var in struct {
    SKU  string `json:"sku"`
    Name string `json:"name"`
}
if err := decodeJSON(r, &in); err != nil {
    writeError(w, http.StatusBadRequest, err.Error())
    return
}
res, err := db.ExecContext(r.Context(),
    `INSERT INTO things (sku, name) VALUES (?, ?)`, in.SKU, in.Name)
if isUniqueConstraintErr(err) {
    writeError(w, http.StatusConflict, "sku already exists")
    return
}
if err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
}
id, _ := res.LastInsertId()
writeJSON(w, http.StatusCreated, map[string]any{"id": id, "sku": in.SKU})
```

```go
// 3. Transaction pattern
tx, err := db.BeginTx(r.Context(), nil)
if err != nil { /* 500 */ }
defer tx.Rollback()

if _, err := tx.ExecContext(r.Context(), `...`); err != nil { /* handle */ }
if _, err := tx.ExecContext(r.Context(), `...`); err != nil { /* handle */ }

if err := tx.Commit(); err != nil { /* 500 */ }
```

```go
// 4. Multi-row scan pattern
rows, err := db.QueryContext(r.Context(), sqlStr, args...)
if err != nil { /* 500 */ }
defer rows.Close()
out := []Thing{}
for rows.Next() {
    var t Thing
    if err := rows.Scan(&t.A, &t.B, &t.C); err != nil { /* 500 */ }
    out = append(out, t)
}
if err := rows.Err(); err != nil { /* 500 */ }
writeJSON(w, http.StatusOK, map[string]any{"items": out})
```

If those four templates are at your fingertips, you can build everything
in this repo without thinking.
