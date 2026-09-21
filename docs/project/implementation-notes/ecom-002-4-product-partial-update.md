# ECOM-002.4 — Product partial updates

Status: done after independent review approval on September 21, 2026. Branch: `fix/ecom-002-4-product-partial-update`. Local commit pending push and integration.

## Objective and reason

Updating a product must change only fields present in the request. An omitted `is_active` currently becomes `false` when the HTTP request is converted to `Product`, and `Product.UpdateState` always copies it. The observable outcome is that omitted fields keep their stored values while explicit `is_active: false` deactivates the product.

## Dependencies and assumptions

ECOM-002.3 is reviewed and integrated into `main`. Keep the existing admin-only `PUT /api/v1/products/{id}` route and product repository. A product category remains required by domain validation. Although the database column is nullable, this unit does not define a category-clearing operation.

## Scope

Carry field presence from HTTP through service and domain using a typed update contract. Merge and validate before persisting. Cover domain, service, and HTTP behavior, including invalid input, persistence failures, and route authorization. Exclude inactive product creation, post-write read failure handling (ECOM-002.6), SKU behavior, and unrelated cleanup.

## Contract, persistence, migrations, and data

Omitted update fields preserve their stored values; explicit `is_active: false` deactivates. Explicit text and category changes still undergo existing validation. The route and response shape remain the same. No persistence or migration change is planned.

## Acceptance criteria

1. Omitted fields preserve persisted values.
2. Explicit `is_active: false` deactivates, while omission preserves activation.
3. Explicit text and category changes pass existing validation; invalid input does not persist.
4. Lookup, validation, and write errors propagate without partial writes.
5. Domain, service, and HTTP tests cover presence, omission, errors, and authorization.

## Test matrix and pre-correction evidence

First add a service regression that updates only a name on an active product and fails because the current code deactivates it. Then cover domain merge states, service lookup and write behavior, and HTTP omitted/false fields, malformed input, missing product, and unauthorized access. Record the failing command, output, and reason before changing production code.

### Pre-correction regression — September 21, 2026

The first attempt to run the test was blocked during dependency setup because the sandbox could not reach `proxy.golang.org`; it did not provide behavioral evidence. After dependencies were downloaded, the unchanged production code compiled and the regression failed for the expected reason:

```text
$ GOCACHE=/tmp/ecommerce-go-build-cache GOMODCACHE=/tmp/ecommerce-go-mod-cache go test ./internal/catalog/service -run '^TestUpdateProductPreservesOmittedActive$' -count=1
--- FAIL: TestUpdateProductPreservesOmittedActive (0.00s)
    product_service_test.go:47: omitted is_active deactivated product: {ID:1 Name:After Slug:before Description:Description SeoTitle:Title SeoDescription:SEO description CategoryID:<address> IsActive:false ...}
FAIL e-commerce-go/internal/catalog/service
exit code: 1
```

`Name` changed to `After`, so the service reached the write path. The active product became inactive because `UpdateState` copied the zero-value `false` from the partial `Product`. The pointer address and timestamps above are abbreviated from the original output; the exact output remains in the implementation session.

## Validation gates

Run focused catalog tests, all tests, race tests, `go vet ./...`, `go build ./...`, formatting, lint, applicable integration tests, OpenAPI drift, available secret and vulnerability scans, and `git diff --check`. Classify each gate by its actual result. Inspect `Makefile` and scripts before using commands.

## Compatibility, rollout, and rollback

Clients that relied on accidental deactivation when omitting `is_active` will observe the corrected behavior. Roll out as an application change. Revert the unit commit to roll back; there is no data migration. Generated OpenAPI currently has a documented pre-existing Swaggo blocker; the approved temporary contract limitation remains until ECOM-009.1.

## Validation evidence — September 21, 2026

| Gate | Classification | Real result |
| --- | --- | --- |
| Focused catalog tests | pass | `go test ./internal/catalog/... -count=1` passed. |
| Full tests | pass | `go test ./... -count=1` passed. `make test` passed and reported 21.4% total statement coverage. |
| Race tests | pass | `go test -race ./... -count=1` passed after dependencies were available. The first attempt was blocked by sandbox DNS during setup. |
| Vet and build | pass | `go vet ./...` and `go build ./...` passed after dependencies were available. Their first attempts were blocked by sandbox DNS during setup. |
| Formatting | pass | `gofmt` applied to changed Go files; `gofmt -l` returned no paths. `goimports` is unavailable. |
| PostgreSQL integration | blocked | The existing product repository test skips without `PRODUCT_REPOSITORY_TEST_DATABASE_URL`; no PostgreSQL container was running. This unit changes no repository SQL or schema. |
| Migrations | not applicable | No migration changed. |
| OpenAPI drift | blocked | Pinned Swaggo v1.16.6 failed on the pre-existing `@Accept JSON` in `category_handler.go` while generating to `/tmp`; no artifact comparison was possible. The product update source annotation was updated. |
| Lint | unavailable | `golangci-lint` is not installed. |
| Secret scan | unavailable | `gitleaks` is not installed. |
| Vulnerability scan | unavailable | `govulncheck` is not installed. |
| Diff whitespace | pass | `git diff --check` returned no errors. |

The successful full test command does not make the skipped PostgreSQL repository test a pass. No code, migration, or generated documentation was changed outside this unit.

## Independent review — September 21, 2026

Verdict: **approved with non-blocking observations**. The reviewer found no blocking, important, or suggestion findings. The review confirmed the acceptance criteria, the 11-file scope, and the pre-correction evidence as recorded by the implementer; the reviewer did not repeat the pre-correction run. The accepted limitations remain: PostgreSQL integration skipped without a test database, OpenAPI drift blocked by the pre-existing Swaggo annotation error, and lint, secret, and vulnerability tools unavailable. Review did not turn these gates into passes. The approved diff matched the pre-stage worktree byte for byte before this administrative update.
