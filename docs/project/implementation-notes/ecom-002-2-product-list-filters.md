# ECOM-002.2 — Product list filters

Status: done

Branch: `fix/ecom-002-2-product-list-filters`

## Objective and reason

Make `GET /api/v1/products` apply `category_id` and `is_active` consistently to rows and total, including `is_active=false`. Before this unit, the handler passed SQL-fragment map keys while the repository read plain keys, so valid filters were ignored. Malformed values were silently omitted.

## Dependencies and assumptions

ECOM-002.1 is merged into `main`. Preserve the existing route, pagination, and Gin first-value behavior for repeated filter parameters, matching category listing.

## Scope and impact

In scope: typed filters across HTTP, service, and repository; filter validation; focused tests; OpenAPI annotations and behavior documentation. Out of scope: SKU, partial updates, new visibility policy, pagination changes, and schema changes. Invalid supported filters return `400 invalid_request`; persistence schema and data do not change.

## Acceptance criteria

1. Positive `category_id` filters rows and total.
2. Both active values work separately and combined with category.
3. Omitted filters preserve existing listing and pagination.
4. Empty and malformed supported filters return `400` before service invocation.
5. No SQL-fragment keys or raw maps cross the product list service boundary.

## Test matrix and validation

Add failing handler regression cases first for valid and malformed filters. Cover absent filters, repeated values, and pagination. Add PostgreSQL repository cases for category, active true/false, combined filters, total, ordering, and pagination. Run focused tests before and after the fix, then formatting, vet, lint, unit, race, integration, build, OpenAPI generation/drift, and applicable security checks. Record unavailable gates.

## Risks and rollback

Clients relying on silently ignored filters or malformed input acceptance will see changed results or `400`. Revert the unit commit, including documentation, if rollback is required; there is no migration.

## Validation record

Handler regression tests failed before the correction because valid keys did not match the repository and invalid filters returned `200`. Catalog package tests and focused PostgreSQL integration passed after correction. The full unit and race suites, vet, build, gofmt, and `git diff --check` passed. `make linter` could not run because `golangci-lint` is absent from PATH. `make docs` could not find `swag` in PATH, so the pinned module version was tried temporarily with `GOCACHE=/tmp/ecommerce-go-build-cache GOMODCACHE=/tmp/ecommerce-go-mod-cache go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/docs.go --output docs --parseDependency --parseInternal`. Generation stopped at the first fatal error: `ParseComment error in file internal/catalog/delivery/http/category_handler.go for comment: '// @Accept       JSON': JSON accept type can't be accepted`. Correcting the pre-existing annotation set and regenerating OpenAPI belongs to ECOM-009.1. The generated Swagger artifacts remain unchanged and stale. No vulnerability or dedicated secret scanning tool is installed; those gates remain unavailable. No schema changes require migration tests.

The unrelated review-workflow files were preserved outside the repository under `/tmp/ecommerce-review-workflow-backup/`, retaining their relative paths. Reverse-engineering text now labels the old product filter defect as historical and records the current SKU-only remainder.
