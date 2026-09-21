# ECOM-002.5 — Category hierarchy cycles

Status: done after independent review and user approval; local commit authorized. Branch: `fix/ecom-002-5-category-hierarchy-cycles`. Base: `main` at `ada2ab3`. No push or merge.

## Objective and reason

Reject a category parent change that places the category in its own ancestor chain. The current service rejects only direct self-parenting; assigning a descendant as parent can persist an indirect cycle.

## Dependencies and assumptions

ECOM-002.3 supplies field-preserving category updates, and ECOM-002.4 is reviewed and merged. Preserve the existing admin-only update route and explicit parent clearing. The application owns this update path; database writes outside the API are outside the approved contract.

## Scope

Walk the proposed parent's ancestors when assigning a parent, reject direct and indirect cycles, propagate lookup failures, and cover service behavior and PostgreSQL persistence. Preserve valid assignment and clearing. Exclude unrelated catalog rules, cache repair, and general tree reorganization.

## Contract, persistence, migrations, and data

Cycle attempts return the existing invalid category reference classification (`400`, `invalid_request`) and leave the stored relationship unchanged. The route and response envelope remain unchanged. The repository uses a transaction-scoped PostgreSQL advisory lock and recursive ancestor query to make the API update path safe for concurrent parent assignments. No migration or stored-data change is needed.

## Acceptance criteria

1. Direct self-parenting is rejected.
2. Indirect cycles are rejected.
3. Valid assignment and parent removal still work.
4. Ancestor lookup failure causes no write.
5. Ancestor-chain unit tests and PostgreSQL integration cases cover the behavior.

## Test matrix and pre-correction evidence

First add a regression for `A -> B -> C`, then attempt `A.parent_id = C`. Run it against the unchanged production code and record the failure. Add direct self-parenting, longer chain, valid assignment, clearing, missing ancestor, lookup failure, and PostgreSQL cases. Check concurrent update safety separately.

## Validation gates

Run focused catalog tests and PostgreSQL integration; `make test`; `go test -race ./...`; `go vet ./...`; `go build ./...`; formatting and lint; applicable migration and OpenAPI checks; available secret and vulnerability scans; and `git diff --check`. Record actual results and classify unavailable or blocked gates explicitly.

## Compatibility, rollout, and rollback

Clients currently creating cycles will receive a validation error. Roll out with the application change. Without a migration, reverting this unit's commit rolls back the behavior. Any migration requires an explicit up/down plan before implementation.

## Human decisions

No decision is required for the approved API behavior. Reconfirm scope if the guarantee must cover direct SQL writes outside the API.

## Pre-correction evidence — September 21, 2026

The initial test attempt was blocked because the default Go build cache was read-only in the sandbox. With task-specific caches, the unchanged production code failed for the expected behavior:

```text
$ GOCACHE=/tmp/ecommerce-go-build-cache GOMODCACHE=/tmp/ecommerce-go-mod-cache go test ./internal/catalog/service -run '^TestUpdateCategoryRejectsIndirectCycle$' -count=1
--- FAIL: TestUpdateCategoryRejectsIndirectCycle (0.00s)
    category_service_test.go:53: cycle error = <nil>, want invalid category reference
FAIL e-commerce-go/internal/catalog/service
```

The service accepted `A.parent_id = C` for the chain `C -> B -> A` and reached the write path. This is behavioral evidence, not a fixture or compilation failure.

## Implementation and integration evidence

The service rejects target, repeated, missing, and unreadable ancestors before writing. Its parent lookup bypasses Redis so stale cached categories cannot decide the hierarchy. The repository atomically rechecks the current PostgreSQL chain under a transaction-scoped advisory lock. The PostgreSQL tests cover direct and indirect cycles, valid assignment, parent clearing, unchanged persisted data after rejection, a stale cache entry, and two concurrent assignments that would otherwise form a cycle. The HTTP test verifies the existing `400`/`invalid_request` mapping.

The isolated PostgreSQL 13 container received the existing first two migrations. The first integration run exposed test fixture leakage into an older list test; the concurrent test now hides its rows during cleanup. The next run passed all category repository integration cases. No migration was added or changed.

## Validation evidence — September 21, 2026

| Gate | Classification | Real result |
| --- | --- | --- |
| Focused catalog | pass | `go test ./internal/catalog/... -count=1` passed with the database variable and task-specific Go caches. An earlier run without the database variable compiled the package but skipped PostgreSQL cases. |
| PostgreSQL integration | pass | `CATEGORY_REPOSITORY_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5438/ecommerce_test?sslmode=disable GOCACHE=/tmp/ecommerce-go-build-cache GOMODCACHE=/tmp/ecommerce-go-mod-cache go test ./internal/catalog/repository -run '^TestCategoryRepository' -count=1 -v` passed after fixture cleanup. |
| Full suite and coverage | pass | The same database and cache variables with `make test` passed, reporting 11 tested packages and 24.6% total statement coverage. The Makefile's skip summary does not detect Go subtest skips; this result does not establish unrelated integration coverage. |
| Race | pass | The same variables with `go test -race ./... -count=1` passed. |
| Vet and build | pass | `go vet ./...` and `go build ./...` passed with task-specific Go caches. |
| Formatting | pass | `gofmt -l` on all changed Go files returned no paths. `goimports` is unavailable. |
| Lint | unavailable | `golangci-lint` is not installed. |
| Migrations | not applicable | No migration changed. Existing migrations 000001 and 000002 were applied successfully to the isolated test database. |
| OpenAPI drift | blocked | Pinned Swaggo v1.16.6 generation to `/tmp` failed on the pre-existing `@Accept JSON` annotation in the category handler. Generation produced no comparable artifact; the source annotation was updated. |
| Secret scan | unavailable | `gitleaks` is not installed. |
| Vulnerability scan | unavailable | `govulncheck` is not installed. |
| Diff whitespace | pass | `git diff --check` returned no errors. |

The first test attempt was blocked by the default Go build cache. The first PostgreSQL integration attempt was blocked by sandbox socket permissions; the approved local database run reached the tests. The first reached integration run found test data leakage, which was corrected before the passing run.

## Independent review follow-up — PostgreSQL evidence

The independent review could not verify persistence and concurrency because its local run omitted `CATEGORY_REPOSITORY_TEST_DATABASE_URL` and skipped the PostgreSQL cases. On September 21, 2026, the isolated `ecom-002-5-category-test` container was running. A read-only database check returned PostgreSQL `13.23` and the `categories` table. The focused tests were then rerun with the connection variable, `-count=1`, and `-v`; the complete output was:

```text
$ CATEGORY_REPOSITORY_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5438/ecommerce_test?sslmode=disable GOCACHE=/tmp/ecommerce-go-build-cache GOMODCACHE=/tmp/ecommerce-go-mod-cache go test ./internal/catalog/repository -run '^TestCategoryRepository(UpdateRejectsCycles|ConcurrentParentUpdatesDoNotCycle)$' -count=1 -v
=== RUN   TestCategoryRepositoryUpdateRejectsCycles
--- PASS: TestCategoryRepositoryUpdateRejectsCycles (0.01s)
=== RUN   TestCategoryRepositoryConcurrentParentUpdatesDoNotCycle
--- PASS: TestCategoryRepositoryConcurrentParentUpdatesDoNotCycle (0.01s)
PASS
ok  e-commerce-go/internal/catalog/repository 0.028s
```

Exit code: `0`. Both PostgreSQL cases ran and passed; neither was skipped. No production code or test was changed for this review follow-up.

## Independent review approval and accepted limitations

The independent reviewer approved ECOM-002.5 with no blocking or important findings after the PostgreSQL evidence follow-up. The user explicitly accepted only the following non-blocking limitations for this unit:

| Accepted limitation | Existing future unit |
| --- | --- |
| OpenAPI drift could not be checked because pinned Swaggo fails on the pre-existing `@Accept JSON` annotation. | `ECOM-009.1` reconciles annotations and generated OpenAPI; `ECOM-008.5` adds the drift gate. |
| `golangci-lint` was unavailable locally. | `ECOM-008.2` pins development tools; `ECOM-008.3` adds the lint CI gate. |
| `gitleaks` secret scanning was unavailable locally. | `ECOM-008.5` adds secret scanning. |
| `govulncheck` vulnerability scanning was unavailable locally. | `ECOM-008.5` adds vulnerability scanning. |

This acceptance applies only to ECOM-002.5. The independent review was approved, and the user subsequently authorized a local commit. Push, pull request, and merge remain unauthorized.
