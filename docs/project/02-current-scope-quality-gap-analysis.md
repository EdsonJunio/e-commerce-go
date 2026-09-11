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

## 4. Execution-unit sizing policy

The `ECOM-002` through `ECOM-009` identifiers are epics. Implementation must select one numbered child unit from the catalog below, never an entire epic at once.

An executable unit should normally:

- Deliver one observable behavior or one cohesive quality gate.
- Cross only the layers required for that behavior.
- Include its unit, integration, or contract tests in the same change.
- Introduce at most one forward migration.
- Be independently reviewable and reversible.
- Have no more than five primary acceptance statements.
- Touch roughly 8 to 12 relevant files at most; generated files and mechanical mock updates do not count.
- End with validation, documentation adjustment, one coherent commit, and push.

The file-count guideline is a warning threshold, not a target. A unit must be split when it contains independent behaviors, multiple unrelated migrations, or acceptance criteria that can pass separately. A unit must not be split into incomplete technical fragments such as "create an interface" without an executable consumer.

Before coding, the implementation note for a unit must state:

1. Objective and observable outcome.
2. Dependencies and assumptions.
3. In-scope and explicitly out-of-scope behavior.
4. Contract and persistence impact.
5. Required test matrix.
6. Validation commands and rollback approach.

Tests and documentation belong to each unit. The testing and documentation epics later in this document are consolidation gates for cross-package evidence and final consistency, not containers for postponed work.

## 5. Executable stabilization units

| Unit | Bounded outcome | Required evidence |
| --- | --- | --- |
| `ECOM-002.1` | Make category list filters typed and consistent across HTTP, service, and repository | Handler and repository tests for every filter and malformed value |
| `ECOM-002.2` | Make product list filters typed and consistent across HTTP, service, and repository | Handler and repository tests for category and active filters |
| `ECOM-002.3` | Preserve field presence in category partial updates, including `false` and nullable parent behavior | Domain, service, and HTTP update matrix |
| `ECOM-002.4` | Preserve field presence in product partial updates, including `false` | Domain, service, and HTTP update matrix |
| `ECOM-002.5` | Prevent direct and indirect category hierarchy cycles | Ancestor-chain unit and PostgreSQL integration cases |
| `ECOM-002.6` | Correct catalog domain errors and stop discarding post-write read failures | Error-mapping and failure-propagation tests |
| `ECOM-003.1` | Align the SKU persistence entity and table mapping with the migrated schema | PostgreSQL mapping test that reads representative SKU rows |
| `ECOM-003.2` | Implement typed SKU listing filters, pagination, validation, and visibility rules | HTTP and repository listing matrix |
| `ECOM-003.3` | Add an explicit SKU availability projection backed by the stock table | Join, missing-stock, active-state, and OpenAPI contract tests |
| `ECOM-004.1` | Introduce a focused category cache port and use the configured TTL | Unit tests with a cache fake and TTL assertions |
| `ECOM-004.2` | Make category rename and delete invalidation cover ID, old slug, and new slug keys | Cache integration tests for update, rename, and deletion |
| `ECOM-004.3` | Define Redis degradation and align startup/readiness behavior with that policy | Redis outage, recovery, and readiness integration tests |
| `ECOM-004.4` | Add stock invariants and indexes required by existing access paths in one forward migration | Migration up/down and constraint/query-plan evidence |
| `ECOM-004.5` | Reconcile GORM soft deletion with database triggers under one documented policy | Repository integration tests for delete, repeated delete, and visibility |
| `ECOM-004.6` | Reconcile normalized email uniqueness and legacy administrator migration behavior | Upgrade, fresh-install, duplicate-email, and rollback tests |
| `ECOM-005.1` | Close remaining catalog domain and service test debt not already covered by prior units | Package-level behavior matrix and coverage report |
| `ECOM-005.2` | Build a reusable isolated PostgreSQL and Redis repository test harness | Repeatable local and CI execution without shared state |
| `ECOM-005.3` | Add the login-to-admin-catalog integration journey | Valid, missing, expired, customer, disabled, and admin token scenarios |
| `ECOM-005.4` | Add package coverage reporting and a monotonic ratchet | Deliberate threshold regression blocks the test command |
| `ECOM-006.1` | Introduce the standard public error envelope with request correlation | Response and middleware contract tests |
| `ECOM-006.2` | Translate validation failures and document the complete HTTP status matrix | Tests for `400`, `401`, `403`, `404`, `409`, `413`, `429`, and `500` |
| `ECOM-006.3` | Add security headers and restrict documentation, metrics, and profiling in production | Environment-specific route and header tests |
| `ECOM-006.4` | Replace the process-local login limiter with a shared implementation | Multi-instance semantics, expiry, failure-policy, and retry-header tests |
| `ECOM-006.5` | Define and implement token revocation, role-change, and disabled-user behavior | Session-state and authorization integration tests |
| `ECOM-007.1` | Close partially initialized resources and make health signals dependency-accurate | Startup failure, shutdown, liveness, and readiness tests |
| `ECOM-007.2` | Use structured logs consistently and redact sensitive request/dependency data | Logger assertions and representative redaction tests |
| `ECOM-007.3` | Add authentication, cache, repository, rate-limit, and dependency metrics | Metric registration and outcome-label tests |
| `ECOM-007.4` | Add API healthcheck, initial dashboard, alerts, SLOs, and dependency runbook | Compose validation and documented failure drill |
| `ECOM-008.1` | Remove dead composition, backup module, unused Node tooling, duplicate targets, and misleading ignore rules | Clean build plus repository inventory review |
| `ECOM-008.2` | Pin development tools and expose one reproducible local quality command | Fresh-environment command verification |
| `ECOM-008.3` | Add fast CI gates for modules, formatting, vet, lint, unit tests, and race tests | Deliberate format/test failures block a pull request |
| `ECOM-008.4` | Add isolated PostgreSQL, Redis, migration, and integration CI gates | Deliberate schema and integration failures block a pull request |
| `ECOM-008.5` | Add OpenAPI drift, secret, vulnerability, binary, container, and image gates | Deliberate drift or security fixture is detected safely |
| `ECOM-009.1` | Reconcile annotations, generated OpenAPI, envelopes, routes, and status codes | Generation produces no unexplained diff and contract tests pass |
| `ECOM-009.2` | Rewrite the reverse engineering and root README in English against verified behavior | Commands, routes, risks, and implementation status are evidence-backed |
| `ECOM-009.3` | Remove obsolete documentation and validate the complete internal link graph | No duplicate source of truth or unresolved local documentation link |

## 6. Stabilization epics

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

## 7. Execution order

```text
ECOM-002.1 -> ECOM-002.6
    -> ECOM-003.1 -> ECOM-003.3
    -> ECOM-004.1 -> ECOM-004.6
    -> ECOM-005.1 -> ECOM-005.4
    -> ECOM-006.1 -> ECOM-006.5
    -> ECOM-007.1 -> ECOM-007.4
    -> ECOM-008.1 -> ECOM-008.5
    -> ECOM-009.1 -> ECOM-009.3
```

Tests and documentation are part of every subtask. ECOM-005 and ECOM-009 are consolidation gates, not permission to postpone all tests or all documentation until the end.

## 8. Definition of Done

Every subtask must provide:

- A bounded branch or commit with a clear conventional message.
- Explicit behavior and error contracts.
- Unit or integration tests proportional to risk.
- Safe forward migrations when persistence changes.
- Updated OpenAPI and human documentation when contracts change.
- Structured logs without sensitive values.
- Passing format, vet, lint, race, test, build, migration, and applicable security gates.
- Push to the working remote branch only after validation.
