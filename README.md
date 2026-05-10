# endpoint-interview-prep

A small Go web service (chi + SQLite) you can use to drill on building HTTP
handlers under interview-style conditions. The skeleton is in place; most
handlers are stubs returning `501`. The job is to implement them, one at a
time, until every test passes.

## Quick start

```bash
go test ./...        # see what's failing
go run .             # boot the server on :8080 with a real SQLite file
curl localhost:8080/health
```

By default the server uses `app.db` in the current directory and listens on
`:8080`. Override with environment variables:

```bash
DB_PATH=/tmp/foo.db ADDR=:9000 go run .
```

## Layout

```
docs/
  ENDPOINTS.md             -- endpoints to implement, ordered by tier
  STUDY_GUIDE.md           -- chi / database/sql / JSON cheat sheet

sql/
  schema.sql               -- table definitions, embedded into the binary
  seed.sql                 -- deterministic seed data, used by tests

main.go                    -- entrypoint: opens the DB, applies schema, serves
server.go                  -- newRouter: wires every URL pattern to its handler
db.go                      -- openDB / applySchema / applySeed
models.go                  -- Customer, Product, Order, OrderItem
helpers.go                 -- writeJSON, writeError, decodeJSON, idParam, ...

handlers_health.go         -- GET /health  (reference implementation)
handlers_products.go       -- products: list + get implemented; rest are stubs
handlers_customers.go      -- customers: all stubs
handlers_orders.go         -- orders: all stubs
handlers_reports.go        -- reports: all stubs

handlers_*_test.go         -- tests for every endpoint (passing & failing)
testhelpers_test.go        -- newTestServer, do, decodeBody, expectStatus
```

## Workflow during practice

1. Open [docs/ENDPOINTS.md](docs/ENDPOINTS.md), pick the next endpoint.
2. Run only the tests for that endpoint, e.g.
   `go test -v -run TestCreateProduct ./...`.
3. Watch the test output to see exactly what the handler must do.
4. Open the handler file (e.g. `handlers_products.go`), find the stub, read
   the comment block above it for hints, and implement.
5. Re-run the tests until they're green.
6. Move to the next endpoint.

## Reference handlers — copy these patterns

- [`health`](handlers_health.go) — minimal handler shape.
- [`listProducts`](handlers_products.go) — query-param filtering, pagination,
  iterating rows.
- [`getProduct`](handlers_products.go) — path param + single-row scan +
  `sql.ErrNoRows` -> 404.

Read [docs/STUDY_GUIDE.md](docs/STUDY_GUIDE.md) for chi, `database/sql`,
and JSON cheat-sheet snippets you may want at hand during the interview.

## Useful commands

```bash
go test ./...                          # run all tests
go test -v -run TestCreateOrder ./...  # run one endpoint's tests, verbose
go test -count=1 ./...                 # disable test cache (force fresh run)
go vet ./...                           # quick static check
go build ./...                         # compile without running

# Inspect the live database created by `go run .`:
sqlite3 app.db '.schema'
sqlite3 app.db 'SELECT * FROM products;'
```

## Re-seeding the dev database

`go run .` only applies the schema, not the seed. To work against seeded
data while developing, run the seed manually:

```bash
sqlite3 app.db < sql/seed.sql
```

(Tests don't need this — `newTestServer` always seeds a fresh DB.)
