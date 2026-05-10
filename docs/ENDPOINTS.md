# Endpoints to implement

Listed roughly in order of difficulty. Implement them in the order shown —
later tiers reuse patterns from earlier ones.

For each endpoint:
1. Read the doc-comment in the handler file (it lists the contract,
   status codes, and hints).
2. Run the matching test:
   `go test -v -run <TestName> ./...`
3. Implement until green.

| # | Tier | Method   | Path                        | Handler / Test |
|--:|:-----|:---------|:----------------------------|:---------------|
| 0 | ref  | GET      | `/health`                   | `health` / `TestHealth` (already done) |
| 1 | ref  | GET      | `/products`                 | `listProducts` / `TestListProducts` (already done) |
| 2 | ref  | GET      | `/products/{id}`            | `getProduct` / `TestGetProduct` (already done) |

## Tier 1 — basic CRUD

You write JSON bodies, parse path params, distinguish 200/201/204 from
400/404/409, and detect `UNIQUE constraint failed` from the SQLite error.

| # | Method   | Path                        | Handler / Test |
|--:|:---------|:----------------------------|:---------------|
| 3 | POST     | `/products`                 | `createProduct` / `TestCreateProduct` |
| 4 | PUT      | `/products/{id}`            | `updateProduct` / `TestUpdateProduct` |
| 5 | DELETE   | `/products/{id}`            | `deleteProduct` / `TestDeleteProduct` |

## Tier 2 — joins & nested resources

You write SQL with `JOIN`, scan into a struct that embeds another struct,
and decide whether an empty result is `200 []` or `404`.

| # | Method   | Path                        | Handler / Test |
|--:|:---------|:----------------------------|:---------------|
| 6 | GET      | `/customers/{id}`           | `getCustomer` / `TestGetCustomer` |
| 7 | GET      | `/customers/{id}/orders`    | `listCustomerOrders` / `TestListCustomerOrders` |
| 8 | GET      | `/orders/{id}`              | `getOrder` / `TestGetOrder` |

## Tier 3 — atomic updates

A single SQL statement does the validation and the mutation together,
using a guarded `WHERE`.

| # | Method   | Path                        | Handler / Test |
|--:|:---------|:----------------------------|:---------------|
| 9 | PATCH    | `/products/{id}/stock`      | `adjustStock` / `TestAdjustStock` |

## Tier 4 — multi-step transactions

Validate, mutate multiple rows, commit; on any failure, roll everything
back. This is the heart of "real" business logic in an interview.

| # | Method   | Path                        | Handler / Test |
|--:|:---------|:----------------------------|:---------------|
| 10 | POST    | `/orders`                   | `createOrder` / `TestCreateOrder` |
| 11 | POST    | `/orders/{id}/cancel`       | `cancelOrder` / `TestCancelOrder` |

## Tier 5 — aggregations / reporting

`GROUP BY`, `SUM`, NULL handling, and translating query params into
optional `WHERE` clauses.

| # | Method   | Path                        | Handler / Test |
|--:|:---------|:----------------------------|:---------------|
| 12 | GET     | `/reports/revenue`          | `revenueReport` / `TestRevenueReport` |
| 13 | GET     | `/reports/top-products`     | `topProducts` / `TestTopProducts` |

## Tier 6 — bulk operations

Same patterns as before, but in a transaction across many rows.

| # | Method   | Path                        | Handler / Test |
|--:|:---------|:----------------------------|:---------------|
| 14 | POST    | `/products/bulk`            | `bulkCreateProducts` / `TestBulkCreateProducts` |

---

## Status code conventions used in this codebase

| Code | Meaning                                                         |
|-----:|:----------------------------------------------------------------|
| 200  | Successful GET / PUT / PATCH                                    |
| 201  | Successful POST that creates a resource — return the new entity |
| 204  | Successful DELETE — empty body                                  |
| 400  | Malformed body or path param (bad JSON, non-integer id, etc.)   |
| 404  | Resource does not exist                                         |
| 409  | Conflict — duplicate UNIQUE, FK in use, illegal state change    |
| 422  | Semantically valid input that violates a business rule (e.g. insufficient stock) |
| 500  | Unexpected DB / server error                                    |
| 501  | Stub — handler not implemented yet (the starting state)         |

If a test expects a code you wouldn't have picked, default to what the
test expects. The interview almost certainly cares more about consistency
and well-handled error paths than nailing 409 vs 422 in every case.

## Stretch goals (no tests, but good practice)

If you finish early, try these. They have no tests; just decide on the
contract and implement.

- `GET /orders?status=&customer_id=&from=&to=&limit=&offset=` — full order
  search across all customers.
- `POST /orders/{id}/items` — append a line item to a `pending` order;
  re-decrement stock; bump `total_cents`. Reject if status != `pending`.
- `POST /orders/{id}/ship` — transition `paid` → `shipped`. 409 from any
  other state.
- Add a `Content-Type` check at the request level so wrong content types
  get rejected with `415`.
- Add a `request_id` in every error body, derived from the
  `X-Request-Id` header that chi's `RequestID` middleware sets.
