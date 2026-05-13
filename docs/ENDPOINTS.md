# Endpoints to implement

Listed roughly in order of difficulty. Implement them in the order shown —
later tiers reuse patterns from earlier ones.

Each tier contains multiple endpoints across three resources (`products`,
`warehouses`, `personnel`) so you can drill the same pattern several times.
If the first one in a tier feels hard, do it slowly; the others should feel
mechanical after that.

For each endpoint:
1. Read the doc-comment in the handler file (it lists the contract,
   status codes, and hints).
2. Run the matching test:
   `go test -v -run <TestName> ./...`
3. Implement until green.

## Reference implementations — read these first

| # | Tier | Method   | Path                        | Handler / Test |
|--:|:-----|:---------|:----------------------------|:---------------|
| 0 | ref  | GET      | `/health`                   | `health` / `TestHealth` |
| 1 | ref  | GET      | `/products`                 | `listProducts` / `TestListProducts` |
| 2 | ref  | GET      | `/products/{id}`            | `getProduct` / `TestGetProduct` |

## Tier 1 — basic CRUD

You write JSON bodies, parse path params, distinguish 200/201/204 from
400/404/409, and detect `UNIQUE constraint failed` from the SQLite error.

| Method   | Path                        | Handler / Test |
|:---------|:----------------------------|:---------------|
| POST     | `/products`                 | `createProduct` / `TestCreateProduct` |
| PUT      | `/products/{id}`            | `updateProduct` / `TestUpdateProduct` |
| DELETE   | `/products/{id}`            | `deleteProduct` / `TestDeleteProduct` |
| POST     | `/warehouses`               | `createWarehouse` / `TestCreateWarehouse` |
| PUT      | `/warehouses/{id}`          | `updateWarehouse` / `TestUpdateWarehouse` |
| DELETE   | `/warehouses/{id}`          | `deleteWarehouse` / `TestDeleteWarehouse` |
| POST     | `/personnel`                | `createPersonnel` / `TestCreatePersonnel` |
| PUT      | `/personnel/{id}`           | `updatePersonnel` / `TestUpdatePersonnel` |
| DELETE   | `/personnel/{id}`           | `deletePersonnel` / `TestDeletePersonnel` |

## Tier 2 — joins & nested resources

You write SQL with `JOIN`, scan into a struct that embeds another struct,
and decide whether an empty result is `200 []` or `404`.

| Method   | Path                              | Handler / Test |
|:---------|:----------------------------------|:---------------|
| GET      | `/customers/{id}`                 | `getCustomer` / `TestGetCustomer` |
| GET      | `/customers/{id}/orders`          | `listCustomerOrders` / `TestListCustomerOrders` |
| GET      | `/orders/{id}`                    | `getOrder` / `TestGetOrder` |
| GET      | `/warehouses/{id}`                | `getWarehouse` / `TestGetWarehouse` |
| GET      | `/warehouses/{id}/inventory`      | `listWarehouseInventory` / `TestListWarehouseInventory` |
| GET      | `/warehouses/{id}/personnel`      | `listWarehousePersonnel` / `TestListWarehousePersonnel` |
| GET      | `/personnel/{id}`                 | `getPersonnel` / `TestGetPersonnel` |

## Tier 3 — atomic updates

A single SQL statement does the validation and the mutation together,
using a guarded `WHERE`.

| Method   | Path                                          | Handler / Test |
|:---------|:----------------------------------------------|:---------------|
| PATCH    | `/products/{id}/stock`                        | `adjustStock` / `TestAdjustStock` |
| PATCH    | `/warehouses/{id}/capacity`                   | `adjustWarehouseCapacity` / `TestAdjustWarehouseCapacity` |
| PATCH    | `/warehouses/{id}/inventory/{product_id}`     | `setWarehouseInventory` / `TestSetWarehouseInventory` |
| PATCH    | `/personnel/{id}/reassign`                    | `reassignPersonnel` / `TestReassignPersonnel` |

## Tier 4 — multi-step transactions

Validate, mutate multiple rows, commit; on any failure, roll everything
back. This is the heart of "real" business logic in an interview.

| Method   | Path                              | Handler / Test |
|:---------|:----------------------------------|:---------------|
| POST     | `/orders`                         | `createOrder` / `TestCreateOrder` |
| POST     | `/orders/{id}/cancel`             | `cancelOrder` / `TestCancelOrder` |
| POST     | `/warehouses/{id}/transfer`       | `transferUnits` / `TestTransferUnits` |
| POST     | `/warehouses/{id}/close`          | `closeWarehouse` / `TestCloseWarehouse` |

## Tier 5 — aggregations / reporting

`GROUP BY`, `SUM`, NULL handling, and translating query params into
optional `WHERE` clauses.

| Method   | Path                                  | Handler / Test |
|:---------|:--------------------------------------|:---------------|
| GET      | `/reports/revenue`                    | `revenueReport` / `TestRevenueReport` |
| GET      | `/reports/top-products`               | `topProducts` / `TestTopProducts` |
| GET      | `/reports/warehouse-utilization`      | `warehouseUtilization` / `TestWarehouseUtilization` |
| GET      | `/reports/headcount`                  | `headcount` / `TestHeadcount` |

## Tier 6 — bulk operations

Same patterns as before, but in a transaction across many rows.

| Method   | Path                        | Handler / Test |
|:---------|:----------------------------|:---------------|
| POST     | `/products/bulk`            | `bulkCreateProducts` / `TestBulkCreateProducts` |
| POST     | `/warehouses/bulk`          | `bulkCreateWarehouses` / `TestBulkCreateWarehouses` |
| POST     | `/personnel/bulk`           | `bulkCreatePersonnel` / `TestBulkCreatePersonnel` |

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
