# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Purpose of this repo

A practice harness for the user's upcoming in-person Go interview, in
which they'll implement HTTP handlers using `chi` against a SQLite
database. The repo is intentionally a fill-in-the-blanks: most handlers
are stubs returning `501`, and the test suite is the spec.

**Do not implement the stub handlers unless the user explicitly asks.**
Solving them defeats the purpose of the practice. If the user gets stuck
and asks for help on a specific handler, prefer:

1. Pointing them at the relevant doc-comment above the stub.
2. Pointing at the matching reference handler (`listProducts`, `getProduct`)
   that already demonstrates the pattern.
3. Asking what they've tried before writing code.

If they explicitly ask for a worked solution, give it — but keep it
minimal and matched to the existing style.

## Architecture (single-package, flat)

Everything lives in `package main` at the repo root — no `internal/`
directories, no module sub-packages. This is deliberate: the user needs
to be able to scan the whole project quickly.

- `server.go` — `newRouter(db)` is the source of truth for which URLs
  exist. When adding a new endpoint, wire it here in addition to
  writing the handler.
- `db.go` — uses `//go:embed sql/schema.sql` and `//go:embed sql/seed.sql` so the
  binary is self-contained. `applySeed` is only called by tests, never
  by `main`.
- `helpers.go` — `writeJSON`, `writeError`, `decodeJSON`, `idParam`,
  `isUniqueConstraintErr`, `isCheckConstraintErr`. Use these in new
  handlers; don't reinvent.
- `models.go` — DTOs that double as DB row containers. Timestamps are
  `string` (ISO format) because the SQLite driver's `time.Time` mapping
  is awkward across drivers; this trade-off is intentional.
- `handlers_*.go` — one file per resource. The convention: each handler
  is a closure factory `func name(db *sql.DB) http.HandlerFunc`.
- `handlers_*_test.go` — paired with each handlers file. Each test gets
  a fresh DB via `newTestServer` (see `testhelpers_test.go`).

## Database driver

`modernc.org/sqlite` (pure-Go, no CGO). Driver name is **`"sqlite"`**, not
`"sqlite3"`. Don't switch to `mattn/go-sqlite3` without a reason — it
brings a CGO toolchain requirement that complicates the user's setup.

## Test pattern (important)

Every test calls `newTestServer(t)` which:

1. Creates a temp-dir SQLite file (auto-cleaned by `t.Cleanup`).
2. Applies `sql/schema.sql`.
3. Applies `sql/seed.sql`.
4. Returns the chi router + the underlying `*sql.DB`.

So tests get an isolated, fully-seeded DB on every run. Tests that need
to set up unusual state (e.g. a cancelled order to test reports) reach
into the returned `*sql.DB` directly with raw SQL — see
`TestRevenueReport/excludes_cancelled_orders` for the pattern.

The seed numbers (ids, prices, stock, totals) are baked into the test
assertions. **Don't change `sql/seed.sql` without updating every test that
references the affected rows.**

## Common commands

```bash
go test ./...                          # all tests
go test -v -run TestCreateOrder ./...  # one endpoint, verbose
go test -count=1 ./...                 # bypass test cache
go run .                               # boot the dev server on :8080
go build ./...                         # compile-only check
go vet ./...                           # quick static check
```

`go run .` applies `sql/schema.sql` to `app.db` but does NOT seed. To work
against seeded data manually: `sqlite3 app.db < sql/seed.sql`.

## Status-code conventions

The tests assume a specific mapping. Match these when adding new
endpoints so the codebase stays internally consistent:

- 200/201/204 — successful GET/PUT/PATCH, POST-create, DELETE
- 400 — malformed input (unparseable JSON, non-integer id, missing required field)
- 404 — addressed resource doesn't exist
- 409 — conflict (UNIQUE collision, FK in use, illegal state transition)
- 422 — body is fine but violates a business rule (e.g. insufficient stock)
- 501 — handler is a stub (the starting state for unimplemented endpoints)

## Documentation files

- [README.md](README.md) — quick start + workflow.
- [docs/ENDPOINTS.md](docs/ENDPOINTS.md) — the ordered list of endpoints to
  implement, by tier. Treat this as the user's curriculum; if they ask
  "what's next?", answer from this file.
- [docs/STUDY_GUIDE.md](docs/STUDY_GUIDE.md) — chi / `database/sql` / JSON
  cheat sheet. Cite this when explaining a pattern rather than re-deriving.

## When the user asks for help

1. They are interviewing soon under "Google allowed, AI not". Treat each
   question as if they need to internalize the answer, not just receive
   it. Brief explanations beat code dumps.
2. Prefer pointing at existing reference handlers (`listProducts`,
   `getProduct`) and the four templates at the bottom of
   [docs/STUDY_GUIDE.md](docs/STUDY_GUIDE.md) over writing new code.
3. If a test is failing in a way the user doesn't understand, read the
   exact assertion that failed and walk through the expectation; don't
   speculate from the test name.
