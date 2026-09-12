# ECOM-002.1 — Typed Category List Filters

Status: done

Branch: `fix/ecom-002-1-category-list-filters`

Implementation date: September 11, 2026

## Objective and observable outcome

Make `GET /api/v1/categories` use one typed filter contract across HTTP, service, and repository layers. Valid `parent_id` and `is_active` values filter both returned rows and the reported total; malformed supported filters return `400 invalid_request` without invoking the service.

## Defect evidence

The handler previously sent raw SQL-like keys (`parent_id = ?` and `is_active = ?`) while the repository read different keys (`parent_id` and `is_active`). Both layers accepted `map[string]interface{}`, and parse errors were silently discarded. No category handler or repository tests covered this behavior.

## Dependencies and assumptions

- ECOM-002.1 has no predecessor in the stabilization roadmap.
- The supported filters are the documented `parent_id` and `is_active` parameters.
- A present `parent_id` must be a positive integer.
- Boolean parsing retains the behavior supported by `strconv.ParseBool`, including an explicit `false` value.
- Pagination parsing and bounds are unchanged.
- Gin's current first-value behavior for repeated query parameters is preserved. A policy for repeated parameters is a future contract decision.

## Acceptance criteria

1. `CategoryListFilters` replaces `map[string]interface{}` in the HTTP, service, and repository layers.
2. `parent_id` and `is_active` preserve presence and filter both records and total.
3. Filters work separately and together, including `is_active=false`.
4. Malformed supported filters return `400 invalid_request` without invoking the service.
5. Omitting filters preserves the current paginated listing behavior.

## In scope

- Typed category list filters and consistent layer contracts.
- Validation of supported filter values.
- Handler regression tests and focused PostgreSQL repository tests for filters, total, and pagination.
- The HTTP `@Failure 400` annotation and current-behavior documentation updates.

## Out of scope

- Product and SKU filters.
- Category mutations, partial-update presence, hierarchy cycles, cache behavior, and error-catalog consolidation.
- Pagination validation or behavior changes.
- A public category name filter.
- A new policy for repeated query parameters.
- Reusable database test infrastructure from ECOM-005.2.
- Comprehensive correction of pre-existing Swagger annotation syntax and regeneration of Swagger artifacts.

## Contract and persistence impact

`GET /api/v1/categories` now returns `400` for malformed `parent_id` and `is_active` values, and its HTTP annotation includes `@Failure 400`. Successful response and pagination envelopes are unchanged. Internal category list interfaces accept `CategoryListFilters`. Repository predicates remain parameterized. There is no schema, migration, cache, or data-write impact.

The generated Swagger artifacts were not regenerated. `make docs` is blocked by pre-existing unsupported annotation syntax, beginning with `@Accept JSON`, before generation reaches this contract change. Correcting all affected annotations and regenerating the artifacts is a broader reconciliation that remains outside ECOM-002.1.

## Expected files

- `internal/catalog/domain/category.go`
- `internal/catalog/service/category_service.go`
- `internal/catalog/repository/category_repository.go`
- `internal/catalog/delivery/http/category_handler.go`
- Category handler and repository test files
- Generated Swagger artifacts, only after the pre-existing generation blocker is resolved outside this unit
- Reverse-engineering and stabilization roadmap documents

## Test matrix

| Layer | Cases | Evidence |
| --- | --- | --- |
| HTTP | no filters; parent; active true; active false; combined | typed values and unchanged pagination reach the service |
| HTTP | invalid, zero, and negative parent; invalid boolean | `400 invalid_request`; service is not called |
| HTTP | repeated supported filters | existing first-value behavior is preserved |
| Repository | no filters with limit/offset | page contents and unfiltered total |
| Repository | parent; active false; combined | filtered rows and filtered total |

The repository test requires a temporary, isolated PostgreSQL instance with the existing migrations applied. It does not introduce the reusable integration-test harness assigned to ECOM-005.2.

## Validation commands

```bash
gofmt -w <changed-go-files>
go test ./internal/catalog/delivery/http -run TestCategoryHandlerListCategories -count=1
CATEGORY_REPOSITORY_TEST_DATABASE_URL='<redacted>' go test ./internal/catalog/repository -run TestCategoryRepositoryListFiltersTotalAndPagination -count=1
go test ./internal/catalog/...
make fmt
go vet ./...
make linter
go test ./...
go test -race ./...
go build ./...
make docs
git diff --check
```

`make docs` was executed but stopped on pre-existing `@Accept JSON` syntax, so no Swagger artifact diff was produced. The comprehensive annotation correction and artifact regeneration remain outside this unit. Record unavailable repository gates rather than claiming they passed.

## Rollback

Revert the single ECOM-002.1 application, test, HTTP annotation, and documentation change. No generated Swagger artifact changed. No database or cache rollback is required because the unit introduces no persistence change.
