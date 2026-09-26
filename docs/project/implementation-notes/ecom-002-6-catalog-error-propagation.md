# ECOM-002.6 — Catalog error propagation

Status: done after independent review and user approval. Branch: `fix/ecom-002-6-catalog-error-propagation`. Base: `main` at `329ab01`. Commit and push evidence are reported in the completion handoff; no pull request or merge is authorized.

## Objective and observable outcome

Correct inaccurate catalog domain errors and HTTP mappings, and stop category and product update handlers from returning success when the post-write read fails. Clients receive the existing public error classification instead of an incorrect validation error, an internal error for known validation failures, or a successful response with null data.

## Reason and current evidence

`GetCategoryBySlug` returns the category-description error for an empty slug. Product description and SEO validation errors are not mapped and therefore become `500 internal_error`. `ErrInvalidProductPrice` has an unrelated description message and is unused. Both update handlers discard the result error from their follow-up read. The reverse-engineering document records these gaps, and current tests do not cover them.

## Dependencies and assumptions

ECOM-002.1 through ECOM-002.5 are implemented and merged. Preserve the existing response envelope, public error codes, update route authorization, service interfaces, and repository behavior. A successful write followed by a failed read is reported using the existing error mapper; this unit does not attempt to roll back a completed write.

## In scope

- Return the category-slug-required error for an empty category slug.
- Map known product description and SEO validation errors to `400 invalid_request`.
- Correct or remove inaccurate unused catalog error declarations without inventing price behavior.
- Propagate post-write read failures from category and product update handlers through the existing HTTP mapper.
- Add focused service, mapper, and HTTP regression tests.
- Reconcile the stale ECOM-002.5 merge evidence while updating the execution table.

## Out of scope

Inactive product creation, SKU behavior, cache invalidation, repository changes, transactions spanning handler reads, new response envelopes, broad error-catalog redesign, OpenAPI-wide reconciliation, and unrelated cleanup are excluded.

## Contract, persistence, migrations, and data

Known validation errors in this unit return `400 invalid_request`. Post-write reads return their mapped `400`, `404`, `409`, or `500` response and redact internal error details consistently with other handler failures. Successful update response shapes remain unchanged. There is no persistence, migration, or stored-data impact.

## Acceptance criteria

1. Empty category slugs produce `ErrCategorySlugRequired`.
2. Product description and SEO validation errors map to `400 invalid_request`.
3. Category update returns a mapped error when its post-write read fails.
4. Product update returns a mapped error when its post-write read fails.
5. Successful update behavior and existing public response envelopes remain unchanged.

## Test matrix and expected pre-correction evidence

| Layer | Cases |
| --- | --- |
| Service | Empty category slug returns the slug-required sentinel. |
| Transport | Product description, SEO title, and SEO description sentinels map to `400 invalid_request`; unknown errors remain `500 internal_error`. |
| Category HTTP | Successful write plus failed follow-up read returns the mapped error and no success payload; successful read remains `200`. |
| Product HTTP | Successful write plus failed follow-up read returns the mapped error and no success payload; successful read remains `200`. |

Before production corrections, the focused tests are expected to fail because the service returns the description sentinel, the mapper returns 500 for known validation errors, and both handlers return 200 after discarding read failures. Failures must be behavioral rather than compilation, environment, or fixture failures.

## Validation gates and commands

- Focused red/green tests: `go test ./internal/catalog/service ./internal/catalog/delivery/http ./internal/shared/transport -run '<focused expression>' -count=1`
- Catalog and transport packages: `go test ./internal/catalog/... ./internal/shared/transport -count=1`
- Full suite: `make test` and `go test ./... -count=1`
- Race: `go test -race ./... -count=1`
- Vet and build: `go vet ./...` and `go build ./...`
- Formatting: `gofmt -l` for changed Go files and repository `make fmt` when `goimports` is available
- Lint: `make linter`
- OpenAPI generation/drift: pinned Swaggo command or `make docs`
- Available secret and vulnerability scanners
- Final integrity: `git diff --check`, staged/unstaged/untracked inspection, and complete diff review

Use task-specific Go caches under `/tmp`. Classify every gate as pass, fail, blocked, unavailable, or not applicable.

## Compatibility, rollout, and rollback

Clients that previously received 500 for known product validation errors will receive 400. A write may already be committed when a post-write read fails; the response accurately reports that the requested representation could not be produced and clients must not assume the write was rolled back. Roll out as an application-only change. Roll back by reverting this unit; no data rollback is needed.

## Human decisions

The bounded behavior and use of one ECOM-002.6 branch are approved. The orphaned `chore/agent-roadmap-workflow` branch remains untouched. No further product decision is required unless atomic write-and-read semantics are desired, which would exceed this unit.

## Evidence log

Baseline on the unchanged branch passed with `GOCACHE=/tmp/ecommerce-go-build-cache GOMODCACHE=/tmp/ecommerce-go-mod-cache go test ./... -count=1`. The initial dependency setup attempt was blocked by the read-only global module cache, and the first isolated-cache attempt was blocked by sandbox DNS; the approved network run populated the isolated cache and completed successfully.

### Pre-correction evidence — September 25, 2026

The unchanged production code compiled and the focused regressions failed for the expected behavior:

```text
$ GOCACHE=/tmp/ecommerce-go-build-cache GOMODCACHE=/tmp/ecommerce-go-mod-cache go test ./internal/catalog/service ./internal/catalog/delivery/http ./internal/shared/transport -run 'Test(GetCategoryBySlugRejectsEmptySlugWithSlugError|CategoryHandlerUpdatePropagatesPostWriteReadFailure|ProductHandlerUpdatePropagatesPostWriteReadFailure|HTTPErrorMapperClassifiesCatalogValidationErrors|HTTPErrorMapperKeepsUnknownErrorsInternal)$' -count=1
TestGetCategoryBySlugRejectsEmptySlugWithSlugError: got category description is required, want category slug is required
TestCategoryHandlerUpdatePropagatesPostWriteReadFailure: got 200 with {"data":null}, want 404 not_found
TestProductHandlerUpdatePropagatesPostWriteReadFailure: got 200 with {"data":null}, want 500 internal_error
TestHTTPErrorMapperClassifiesCatalogValidationErrors: product description and both SEO errors mapped to 500 internal_error instead of 400 invalid_request
FAIL
```

The unknown-error mapper control case passed. No failure was caused by compilation, environment, or fixture setup.

### Focused green evidence — September 25, 2026

After the minimal correction, the same command passed in all three packages. Full validation evidence is recorded below after all applicable gates run.

### Implementation and validation evidence — September 25, 2026

The service now uses the slug-required sentinel. The mapper classifies product description and SEO validation failures as invalid requests while preserving the unknown-error fallback. Both update handlers map and redact post-write read failures before returning. The two inaccurate and unused internal sentinels were removed. No service interface, repository, migration, generated contract, or persistence behavior changed.

| Gate | Classification | Real result |
| --- | --- | --- |
| Baseline | pass | `go test ./... -count=1` passed on unchanged `main` after dependencies were placed in task-specific caches. |
| Focused red | pass | All four defective behaviors failed for the expected reason; the unknown-error control passed. |
| Focused green | pass | The identical focused command passed in service, HTTP delivery, and transport. |
| Catalog and transport | pass | `go test ./internal/catalog/... ./internal/shared/transport -count=1` passed. Repository integration tests compiled and ran their environment-independent paths; this result is not claimed as PostgreSQL integration evidence. |
| Full Go suite | pass | `go test ./... -count=1` passed. |
| Official test/coverage command | pass | `make test` reported 12 passed packages, 0 failures, 0 skips detected by the script, and 23.4% total statement coverage. The script cannot prove that nested Go tests did not call `t.Skip`. |
| Race | pass | `go test -race ./... -count=1` passed. |
| Vet | pass | `go vet ./...` exited 0. |
| Build | pass | `go build ./...` exited 0. |
| Formatting | pass | Pinned `goimports v0.39.0` corrected one test import group; subsequent `goimports -l` and `gofmt -l` on every changed Go file returned no paths. |
| Lint | blocked | `golangci-lint v1.64.8` reached type checking but could not decode current Go export data and reported cascading undefined imports in unchanged packages. Go test, vet, and build all resolve those imports successfully, so no lint finding is claimed. |
| PostgreSQL/Redis integration | not applicable | No repository, cache, schema, or persistence behavior changed. |
| Migration up/down | not applicable | No migration was added or changed. |
| OpenAPI generation/drift | blocked | Pinned Swaggo v1.16.6 reached parsing and stopped on the pre-existing `// @Accept JSON` category annotation. No generated repository file changed, and the HTTP status set for the affected update routes did not change. |
| Secret scan | unavailable | `gitleaks` is not installed. |
| Vulnerability scan | unavailable | `govulncheck` is not installed. |
| Documentation links | not applicable | No local documentation link was added or changed. |
| Diff whitespace and scope | pass | `git diff --check` returned no errors. Index inspection showed no staged changes. Scope inspection found ten modified and two new files, all belonging to ECOM-002.6 and the approved ECOM-002.5 evidence reconciliation. |

The environment printed `Failed to create stream fd: Operation not permitted` around several successful commands; their child processes completed with exit 0 and emitted complete Go results. The first lint run was blocked earlier by its default read-only cache; the isolated-cache rerun produced the tool incompatibility described above. The first OpenAPI attempt was blocked by sandbox DNS; the approved network rerun reached the known annotation failure.

## Independent review — September 25, 2026

Verdict: **approved with non-blocking observations**. The independent reviewer found no blocking, important, or suggestion findings. The reviewer reproduced the pre-correction failures against base `329ab01`, confirmed all five acceptance criteria, reran the focused, full, race, vet, build, formatting, and diff checks, and verified that the worktree and empty index remained unchanged.

The reviewer retained the existing classifications: lint and OpenAPI drift are blocked by the documented tool and annotation limitations; PostgreSQL, Redis, and migrations are not applicable; `gitleaks` and `govulncheck` are unavailable. The reviewer also confirmed that internal read errors are redacted, no secrets or personal data were introduced, and the documented post-write/read ambiguity remains an accepted contract risk. The user approved proceeding with completion, commit, and push of the unit branch; pull request and merge remain unauthorized.
