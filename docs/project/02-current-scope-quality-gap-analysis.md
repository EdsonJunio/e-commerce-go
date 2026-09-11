# Current-Scope Quality Gap Analysis

Status: active implementation roadmap

Baseline date: September 11, 2026

Baseline branch: `fix/ECOM-001-stabilize-existing-api`

## 1. Purpose

This document defines the work required to bring the currently executable API to a market-grade software quality baseline. It covers only behavior that already exists: authentication, authorization, categories, products, SKU listing, PostgreSQL, Redis, HTTP operations, observability, and delivery tooling.

Cart, order, external payment/fiscal integration, and fulfillment capabilities are not part of this stabilization roadmap. Their implementation is tracked separately in [the complete E-commerce roadmap](./03-complete-ecommerce-implementation-roadmap.md).

"Market-grade" means explicit contracts, safe defaults, automated verification, operational visibility, and recoverable changes. It does not claim conformance with any company's private engineering standards.

## 2. Verified baseline

The following foundation work is complete on the baseline branch:

- Configuration and secrets are externalized and validated.
- JWT validation enforces HS256, issuer, audience, expiration, and not-before claims.
- Identity roles and user status are persisted.
- Catalog mutations require the `admin` role.
- Login has a basic fixed-window rate limit.
- HTTP request bodies have a configurable size limit.
- PostgreSQL startup retries are bounded and context-aware.
- Graceful HTTP shutdown and resource cleanup are present.
- The seed command requires explicit configuration and destructive-operation confirmation.
- Database migrations have been exercised against PostgreSQL 13.
- `go test -race ./...`, `go vet ./...`, and `gofmt` pass.

The total statement coverage is currently 9.7%. Identity has meaningful unit coverage, while the catalog, repositories, database, cache, response, and transport packages remain effectively uncovered.

## 3. Current risks

| Priority | Gap | Current impact |
| --- | --- | --- |
| Critical | Handler and repository filter keys do not match | Category, product, and SKU filters are silently ignored |
| Critical | Partial updates lose boolean field presence | Updating an unrelated field can deactivate a category or product |
| Critical | SKU model does not match the database schema | SKU listing can query nonexistent columns and fail at runtime |
| High | Category cache invalidation is incomplete | Renamed or deleted categories can remain visible through stale slug keys |
| High | Catalog business paths have almost no tests | Regressions can pass every current quality command |
| High | No mandatory CI pipeline exists | Local hooks can be bypassed and merges are not independently verified |
| High | Database invariants are incomplete | Invalid stock and ambiguous soft-delete behavior remain possible |
| Medium | Redis is required at startup but absent from readiness | Orchestrators can receive an incomplete health signal |
| Medium | Login limiting is process-local | Limits reset and diverge when the API has multiple replicas |
| Medium | Generated OpenAPI and manual documentation are stale | Consumers receive contracts that do not match runtime behavior |
| Medium | Repository and tooling contain dead or redundant artifacts | Maintenance and onboarding remain unnecessarily confusing |

## 4. Stabilization subtasks

### ECOM-002 — Correct category and product behavior

Objective: make existing category and product operations deterministic and contract-safe.

Implementation:

- Replace `map[string]interface{}` filters with typed query/filter structures.
- Make handler, service, and repository filter contracts identical.
- Reject malformed query parameters with a stable `400` error instead of ignoring them.
- Model partial updates with explicit field presence so omitted fields remain unchanged.
- Support `is_active=false` in create and update operations.
- Define whether nullable relationships can be explicitly cleared and represent that operation safely.
- Normalize and validate names and slugs consistently.
- Prevent direct and indirect category hierarchy cycles.
- Return the updated entity without discarding repository or follow-up read errors.
- Correct inaccurate domain errors and HTTP mappings.

Required tests:

- Domain validation and update behavior.
- Service orchestration and repository error propagation.
- Handler success, malformed input, not found, conflict, authentication, and authorization.
- Every supported filter, including `false` boolean values.
- Parent-cycle scenarios and relationship clearing.

Completion criteria:

- Every documented category and product operation has behavioral tests.
- No filter is represented by raw SQL fragments outside the repository.
- An omitted update field never changes persisted state.
- Race tests, vet, lint, and integration checks pass.

### ECOM-003 — Align SKU behavior with the schema

Objective: make the existing SKU endpoint executable and unambiguous.

Implementation:

- Rename `Product_skus` to idiomatic `ProductSKU`.
- Add an explicit table mapping where necessary.
- Remove the nonexistent `stock` column from the SKU persistence model.
- Represent inventory through the `stock` table and use an explicit join only when availability is requested.
- Replace the invalid `sku_id` filter with supported filters such as ID, product ID, SKU code, active status, and availability.
- Introduce typed filters, pagination, validation, and stable errors.
- Decide whether public responses expose inactive SKUs.
- Add OpenAPI annotations for the endpoint.

Required tests:

- SKU mapping and filtering against real PostgreSQL.
- Pagination and empty results.
- Active/inactive visibility policy.
- Stock join behavior and missing stock rows.
- Invalid query input and repository failures.

Completion criteria:

- The endpoint runs against the migrated schema without implicit-column failures.
- SQL behavior is verified by integration tests.
- The response contract is documented and stable.

### ECOM-004 — Make persistence and cache behavior safe

Objective: make PostgreSQL the clear source of truth and Redis a predictable optimization.

Implementation:

- Place Redis operations behind a small cache interface.
- Use the configured category TTL instead of a hard-coded duration.
- Invalidate ID, old slug, and new slug keys during updates.
- Invalidate all relevant keys during deletion.
- Prevent cached deleted or inactive records from authorizing new relationships.
- Define and implement the degraded mode when Redis is unavailable.
- Add `reserved_quantity <= quantity` and other missing database constraints.
- Add indexes based on actual foreign-key and filtering access paths.
- Reconcile GORM soft delete with database delete triggers and adopt one explicit policy.
- Review case-insensitive uniqueness and reuse rules for soft-deleted records.
- Remove implicit administrator promotion by numeric user ID, or document and test a safe compatibility migration.
- Preserve applied migrations and use forward migrations for shared environments.

Required tests:

- Cache hit, miss, corrupt payload, timeout, update, rename, and deletion.
- Repository behavior with Redis unavailable.
- Migration up/down in an isolated PostgreSQL instance.
- Constraints, unique indexes, soft deletion, and rollback behavior.

Completion criteria:

- Redis failure cannot corrupt authoritative data.
- No stale key can expose a renamed or deleted category beyond the documented policy.
- All schema invariants are enforced and integration-tested.

### ECOM-005 — Establish robust automated tests

Objective: make regressions observable before merge.

Implementation:

- Replace the empty product handler test with behavioral test cases.
- Add unit tests for catalog domains and services.
- Add HTTP tests for every route and error family.
- Add PostgreSQL repository tests with isolated data.
- Add Redis integration or protocol-compatible tests where behavior depends on Redis semantics.
- Add application wiring and startup-failure tests.
- Add a login-to-authorized-catalog integration flow.
- Produce package-level coverage reports and enforce a ratcheting threshold.

Coverage policy:

- Domain and service packages target at least 80% statement coverage.
- Security-critical branches require explicit scenario coverage regardless of percentage.
- Generated code and trivial entry points may be excluded only through a documented policy.
- The global threshold must increase monotonically until the agreed target is reached.

Completion criteria:

- The test suite fails when a known catalog defect is deliberately reintroduced.
- Integration tests are isolated, repeatable, and safe to run locally and in CI.
- `go test -race ./...` remains mandatory.

### ECOM-006 — Standardize HTTP contracts and security controls

Objective: expose a consistent, non-leaky API contract.

Implementation:

- Standardize errors with `code`, `message`, `details`, and `request_id` where useful.
- Translate validation failures into stable public messages instead of returning raw validator errors.
- Document and test `400`, `401`, `403`, `404`, `409`, `413`, `429`, and `500` behavior.
- Add appropriate HTTP security headers.
- Move login rate-limit state to Redis or another shared backend before horizontal scaling.
- Define token revocation, role-change, and disabled-user behavior during token lifetime.
- Restrict Swagger, metrics, and profiling endpoints in production.
- Add secret, dependency, and vulnerability scanning.

Completion criteria:

- Runtime responses match OpenAPI examples and schemas.
- Internal errors, credentials, hashes, and implementation details never reach clients or logs.
- Security controls behave consistently across multiple API replicas.

### ECOM-007 — Complete runtime resilience and observability

Objective: make the service diagnosable and safe to operate.

Implementation:

- Include Redis in readiness when it is a hard dependency, or make Redis optional and document degraded readiness.
- Close already-opened resources when a later startup step fails.
- Use structured logging throughout database and cache initialization.
- Redact or omit sensitive query values and headers.
- Add metrics for authentication results, cache hits/misses, repository failures, rate limiting, and dependency latency.
- Define initial SLOs and actionable alerts.
- Add an API container health check.
- Pin container image versions and define an image update policy.
- Keep operational endpoints off the public application surface where deployment architecture permits.

Completion criteria:

- Dependency failure modes are tested.
- Readiness accurately represents the service's ability to serve requests.
- A dashboard and runbook can identify HTTP, database, and cache degradation.

### ECOM-008 — Add CI/CD quality gates and clean the toolchain

Objective: make the quality baseline mandatory and reproducible.

Pipeline gates:

1. Verify `go mod tidy` produces no diff.
2. Verify `gofmt` and `goimports`.
3. Run `go vet` and pinned `golangci-lint` checks.
4. Run unit tests and `go test -race`.
5. Run PostgreSQL, Redis, and migration integration tests.
6. Enforce the coverage policy.
7. Regenerate OpenAPI and reject unexpected drift.
8. Run `govulncheck` and secret scanning.
9. Build the Go binaries and container image.
10. Scan the final image and retain test evidence.

Repository cleanup:

- Remove `go.mod.save`.
- Remove Google Wire or make it the single supported composition mechanism.
- Remove Node/Husky files if they have no maintained role.
- Remove the duplicate `air` Makefile target.
- Pin tool installation versions.
- Correct `.gitignore` entries that ignore tracked project source and documentation.
- Version reusable hooks outside `.git/hooks`, while keeping CI authoritative.

Completion criteria:

- A deliberately failing test, formatting error, leaked test secret, migration failure, or OpenAPI drift blocks a pull request.
- Local and CI commands use the same documented entry points.

### ECOM-009 — Reconcile OpenAPI and project documentation

Objective: leave one accurate English documentation set after stabilization.

Implementation:

- Regenerate OpenAPI from corrected handlers.
- Remove claims for checkout, real-time inventory, and complete SKU management until executable paths exist.
- Document response envelopes and every implemented status code.
- Update the reverse-engineering document to reflect the security and catalog corrections.
- Update the root README with verified setup, migration, seed, test, and operation instructions.
- Translate maintained manual documentation and public code documentation to English.
- Remove obsolete documentation and duplicate sources of truth.
- Keep code comments concise and focused on intent or non-obvious constraints.

Completion criteria:

- Every documented feature has an executable path and tests.
- Every executable public endpoint is represented in OpenAPI.
- All maintained documentation and public code documentation are in English.
- All internal documentation links resolve.

## 5. Execution order

```text
ECOM-002 Category and product correctness
    -> ECOM-003 SKU/schema alignment
    -> ECOM-004 Persistence and cache safety
    -> ECOM-005 Test and integration baseline
    -> ECOM-006 HTTP contracts and security
    -> ECOM-007 Runtime and observability
    -> ECOM-008 CI/CD and repository hygiene
    -> ECOM-009 Final documentation reconciliation
```

Tests and documentation are part of every subtask. ECOM-005 and ECOM-009 are consolidation gates, not permission to postpone all tests or all documentation until the end.

## 6. Definition of Done

Every subtask must provide:

- A bounded branch or commit with a clear conventional message.
- Explicit behavior and error contracts.
- Unit or integration tests proportional to risk.
- Safe forward migrations when persistence changes.
- Updated OpenAPI and human documentation when contracts change.
- Structured logs without sensitive values.
- Passing format, vet, lint, race, test, build, migration, and applicable security gates.
- Push to the working remote branch only after validation.
