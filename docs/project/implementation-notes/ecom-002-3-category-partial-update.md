# ECOM-002.3 — Category partial updates

Status: review. Approved for implementation on September 12, 2026.

## Objective and reason

Updating a category changes only fields present in the request. An explicit `is_active: false` deactivates it, and `parent_id: null` clears its parent. Currently the HTTP handler loses boolean presence when constructing a `Category`; domain `UpdateState` always copies the resulting false value. A null parent is indistinguishable from omission.

## Dependencies and assumptions

ECOM-002.2 is integrated. Keep the existing admin-only `PUT /api/v1/categories/{id}` route and repository. `null` is an explicit request to clear `parent_id`; omission preserves it.

## Scope

Introduce a typed, presence-preserving update contract across HTTP, service, and domain. Merge and validate the resulting category. Add domain, service, and HTTP regression tests. Do not address indirect hierarchy cycles (ECOM-002.5), complete cache invalidation (ECOM-004.2), or post-write read errors (ECOM-002.6).

## Contract and persistence

Document omitted fields, `is_active: false`, and the three `parent_id` states (omitted, ID, null). No migration is expected; `categories.parent_id` is nullable.

## Acceptance criteria

1. Omitted fields preserve persisted values.
2. Explicit `is_active: false` deactivates; omission preserves activation.
3. `parent_id: null` clears, a valid ID assigns, and omission preserves the parent.
4. Invalid input and missing parent references do not persist partial changes.
5. Domain, service, and HTTP tests cover the update matrix and route authorization.

## Test matrix and gates

First add regressions that fail for unintended deactivation and inability to clear the parent. Cover merge behavior and validation in domain tests; lookup, persistence, and failure propagation in service tests; and presence, malformed payloads, and authorization in HTTP tests. Run formatting, vet, lint, focused and full unit tests, race tests, applicable PostgreSQL integration tests, build, OpenAPI drift, and available security checks. Record unavailable gates as limitations.

## Risk, rollout, and rollback

The main risk is collapsing JSON null into omission. Keep the existing route and storage schema. Roll out as an application change; revert the unit commit to roll back without data migration.

## Validation evidence — September 12, 2026

The initial `TestUpdateCategoryPreservesOmittedActive` failed because an omitted boolean deactivated the category. The initial `TestCategoryUpdateStateClearsParent` failed because the old model could not clear a parent. Both pass after the correction. `go test ./internal/catalog/... -count=1`, `go test ./... -count=1`, `go test -race ./... -count=1`, `go vet ./...`, `go build ./...`, `gofmt -l` on changed Go files, and `git diff --check` passed. The repository tests requiring `CATEGORY_REPOSITORY_TEST_DATABASE_URL` skip without that variable; no repository code or schema changed in this unit. `make linter` failed because `golangci-lint` is not installed. Pinned Swaggo v1.16.6 generation failed at the existing `@Accept JSON` annotation in `category_handler.go`; generated OpenAPI remains stale pending ECOM-009.1. Dedicated secret and vulnerability scanners are not installed. No migration test applies because there is no migration.

### Pre-correction regression transcript

The following commands were run on the unit branch after adding each regression test and before editing production code. The original tool output is in the implementation session transcript; it was not saved as a separate file at execution time. Recorded output and exit codes:

```text
$ GOCACHE=/tmp/ecommerce-go-build-cache GOMODCACHE=/tmp/ecommerce-go-mod-cache go test ./internal/catalog/service -run TestUpdateCategoryPreservesOmittedActive -count=1
--- FAIL: TestUpdateCategoryPreservesOmittedActive (0.00s)
    category_service_test.go:43: omitted is_active deactivated category: {ID:1 Name:After Slug:before ParentID:<nil> IsActive:false Description:Description CreatedAt:0001-01-01 00:00:00 +0000 UTC UpdatedAt:0001-01-01 00:00:00 +0000 UTC DeletedAt:{Time:0001-01-01 00:00:00 +0000 UTC Valid:false} DeletedReason:}
FAIL
FAIL e-commerce-go/internal/catalog/service 0.003s
exit code: 1

$ GOCACHE=/tmp/ecommerce-go-build-cache GOMODCACHE=/tmp/ecommerce-go-mod-cache go test ./internal/catalog/domain -run TestCategoryUpdateStateClearsParent -count=1
--- FAIL: TestCategoryUpdateStateClearsParent (0.00s)
    category_test.go:10: explicit null parent did not clear relationship: 2
FAIL
FAIL e-commerce-go/internal/catalog/domain 0.003s
exit code: 1
```

The first test exercised the old service signature using a `Category{Name: "After"}`; the second exercised the old `UpdateState` signature using `Category{ParentID: nil}`. Both tests were adapted to `CategoryChanges` after the production signature changed. This transcript supports the failure reason but is not a standalone pre-correction worktree artifact.

After independent review, the service matrix was corrected to change parent ID 2 to a distinct existing parent ID 3; a domain test now checks the same transition. The catalog tests, full tests, catalog race tests, vet, and build passed again after that change.

### OpenAPI review decision

Runtime behavior and the source annotation describe `parent_id: null`, but `docs/swagger.yaml`, `docs/swagger.json`, and `docs/docs.go` still show the old update description and integer-only `parent_id`. The pinned generator stops on the pre-existing `@Accept JSON` annotation before emitting artifacts. Repairing the annotation set and regenerating the full contract is ECOM-009.1. Until that unit lands, clients must use the source annotation and this note for the null-clearing behavior; the published generated OpenAPI is incomplete for this field.

Independent review approved ECOM-002.3 with this non-blocking observation and accepted deferring the pre-existing generator repair to ECOM-009.1 for review purposes. Human approval of the temporary published-contract limitation was then given in the implementation conversation. The unit remains `review` pending the separately restricted stage, commit, and push steps.
