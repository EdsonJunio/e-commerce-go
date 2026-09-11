# E-Commerce Go — Agent Instructions

## Mission

Evolve this repository through small, reviewable roadmap units while preserving
correctness, security, traceability, and operational quality.

"Market-grade quality" means explicit contracts, safe defaults, automated
verification, observable behavior, recoverable changes, and documented
decisions.

Do not claim compliance with Mercado Livre, Uber, or another company's private
engineering standards.

## Project status

This repository is a partial prototype under stabilization.

Never treat behavior as implemented only because it appears in a roadmap,
migration, seed, README, Swagger file, or architecture document.

Confirm current behavior in source code and executable tests.

## Required project documents

Before selecting or implementing work, read:

1. `docs/project/README.md`
2. `docs/project/01-engenharia-reversa-codigo-atual.md`
3. `docs/project/02-current-scope-quality-gap-analysis.md`
4. `docs/project/arquitetura-e-commerce-go.md`
5. `docs/project/03-complete-ecommerce-implementation-roadmap.md`

Read the relevant sections instead of loading unrelated content when documents
are large.

## Document responsibilities

- `01-engenharia-reversa-codigo-atual.md` describes verified current behavior.
- `02-current-scope-quality-gap-analysis.md` defines stabilization work.
- `arquitetura-e-commerce-go.md` describes the target architecture.
- `03-complete-ecommerce-implementation-roadmap.md` defines future capabilities.
- Accepted ADRs record architectural decisions.
- Generated OpenAPI describes the public HTTP contract.

## Source-of-truth precedence

When information conflicts, use this order:

1. Executable source code and tests for current behavior.
2. Accepted ADRs for architectural decisions.
3. Current-scope gap analysis for stabilization priority.
4. Complete roadmap for future implementation order.
5. Target architecture for intended system boundaries.
6. Reverse-engineering documentation for the last recorded code state.
7. README and generated documentation as summaries.

Do not silently resolve a meaningful conflict. Record it and determine whether
the selected unit can proceed safely.

## Roadmap selection policy

Work on exactly one numbered executable unit at a time.

Examples:

- Valid: `ECOM-002.1`
- Invalid: entire `ECOM-002` epic
- Invalid: "complete the catalog"
- Invalid: multiple unrelated units

Use this selection order:

1. Continue the explicitly requested in-progress unit.
2. Otherwise, select the first incomplete and unblocked stabilization unit from
   `02-current-scope-quality-gap-analysis.md`.
3. Do not start future functionality from
   `03-complete-ecommerce-implementation-roadmap.md` until the stabilization
   completion gate is satisfied or the user explicitly changes the priority.
4. If the next unit is blocked, record the blocker and select only a prerequisite
   that is already represented in the roadmap.
5. Do not invent hidden prerequisite projects.

Before coding, verify that the unit is still incomplete by inspecting the
current branch, source code, tests, documentation, and relevant commit history.

## Implementation note

Before changing code, create or update an implementation note containing:

1. Roadmap unit identifier.
2. Objective and observable outcome.
3. Reason the unit is needed.
4. Dependencies and assumptions.
5. In-scope behavior.
6. Out-of-scope behavior.
7. Contract and persistence impact.
8. Test matrix.
9. Validation commands.
10. Rollback approach.

If the unit cannot be described with a small, cohesive acceptance contract,
split it before implementation.

## Branch policy

Never implement directly on `main`.

Create one branch per executable roadmap unit.

Use:

- `fix/ecom-002-1-category-list-filters`
- `fix/ecom-002-3-category-partial-update`
- `feat/ecom-010-1-customer-registration`
- `refactor/ecom-008-1-remove-dead-tooling`
- `docs/ecom-009-2-english-documentation`

Before creating the branch:

1. Inspect the current branch and working tree.
2. Preserve uncommitted user work.
3. Confirm the base branch.
4. Check that a branch for the unit does not already exist.
5. Continue an existing unit branch instead of creating a duplicate.

Do not merge into `main` without explicit user authorization.

## Implementation quality

- Follow idiomatic Go.
- Preserve the selected architecture and current stack unless an accepted ADR
  changes them.
- Keep domain rules independent from Gin, GORM, Redis, PostgreSQL drivers, and
  external provider SDKs.
- Keep HTTP parsing and response mapping in delivery packages.
- Keep business decisions in domain or application services.
- Keep persistence behavior in repositories.
- Use typed contracts instead of raw maps for business inputs.
- Preserve field presence in partial updates.
- Use explicit transactions for multi-write invariants.
- Propagate `context.Context` across I/O boundaries.
- Wrap errors with useful context and preserve error classification.
- Do not add an interface, dependency, abstraction, or distributed pattern
  without a concrete use in the selected unit.
- Do not access another service's database.
- Do not expand the selected roadmap unit with unrelated cleanup.

## Testing policy

Tests belong to every roadmap unit.

For a bug:

1. Reproduce the defect.
2. Add a regression test that fails for the expected reason.
3. Apply the smallest complete correction.
4. Confirm the regression test passes.
5. Run affected package and integration tests.

For new behavior, cover:

- Successful behavior.
- Relevant validation errors.
- Authentication and authorization when applicable.
- State transitions.
- Persistence constraints.
- Failure propagation.
- Idempotency and concurrency when relevant.

Do not delete, skip, weaken, or rewrite a valid test merely to make checks pass.

Use the exact evidence required by the selected roadmap unit.

## Required validation

Inspect the repository's Makefile and scripts before selecting commands.

Run the smallest relevant checks during implementation.

Before concluding a unit, run every available and applicable gate required by
its Definition of Done, including:

- Formatting.
- Vet.
- Lint.
- Unit tests.
- Race tests.
- Integration tests.
- Migration tests.
- OpenAPI drift checks.
- Secret scanning.
- Vulnerability scanning.
- Build checks.

Report commands and real results. Never claim that a check passed when it was
not executed.

If a required gate does not exist yet, report that limitation. Implement the
gate only when it belongs to the selected roadmap unit.

## Security policy

Never include credentials, tokens, personal data, production secrets, or
sensitive payloads in source code, tests, fixtures, documentation, commands,
logs, commits, or reports.

Redact detected credentials from output.

Use least privilege, safe defaults, explicit authorization, bounded input,
parameterized queries, and controlled error disclosure.

## Language policy

Use English for:

- Source code identifiers.
- Source code comments.
- Go documentation.
- Branch names.
- Commit messages.
- New maintained technical documentation.
- API fields and public error codes.

Translate existing maintained documentation as part of the roadmap unit
responsible for English documentation.

Do not delete content merely because it is written in Portuguese.

Do not hide a large translation inside an unrelated implementation unit.

## Documentation updates

Whenever current behavior changes:

1. Update the relevant contract or OpenAPI annotations.
2. Update the affected architecture or domain documentation.
3. Update `01-engenharia-reversa-codigo-atual.md`.
4. Update the selected roadmap unit with status and evidence.
5. Regenerate derived documentation when applicable.
6. Verify local documentation links affected by the change.

Only mark a behavior as implemented when source code and tests provide evidence.

Do not rewrite historical baseline statements as if they had always described
the new state. Add a dated update or clearly revise the baseline scope.

## Commit and push policy

Create one coherent conventional commit for the completed unit unless the unit
requires separately reviewable migration and application commits.

Before committing:

1. Inspect the complete diff.
2. Confirm the diff contains only the selected unit.
3. Run required validation.
4. Run the available secret scan.
5. Update documentation.
6. Confirm no generated or local files were accidentally included.

Push only to the roadmap unit's working branch.

Do not force-push, rewrite shared history, delete remote branches, or merge
without explicit user authorization.

## Completion report

At the end, report:

1. Roadmap unit.
2. Branch.
3. Objective delivered.
4. Files changed.
5. Tests added.
6. Commands executed and results.
7. Documentation updated.
8. Commit and push status.
9. Remaining risks or blockers.
10. Next eligible roadmap unit.

Present the final diff or pull request for human review.
