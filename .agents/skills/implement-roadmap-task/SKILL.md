---
name: implement-roadmap-task
description: Execute one numbered E-Commerce Go roadmap unit with task selection, implementation, tests, documentation, validation, commit, and push. Use for ECOM units or requests to continue the project roadmap.
---

Follow all instructions from the repository `AGENTS.md`.

## Select one unit

If the user provides an `ECOM-*.*` identifier, use that unit.

Otherwise:

1. Read `docs/project/02-current-scope-quality-gap-analysis.md`.
2. Continue the unit marked `in_progress`, if one exists.
3. Otherwise, select the first incomplete and unblocked unit marked `ready`.
4. Verify in code, tests, documentation, and Git history that it remains
   incomplete.
5. Use `docs/project/03-complete-ecommerce-implementation-roadmap.md` after the
   stabilization gate is complete or when explicitly requested.

Execute one numbered child unit, such as `ECOM-002.1`. Never execute an entire
epic, such as `ECOM-002`, in one task.

## Prepare

Before modifying code:

1. Read the unit's outcome, dependencies, required evidence, and Definition of
   Done.
2. Read the relevant reverse-engineering and architecture sections.
3. Inspect the affected code and tests.
4. Define scope, acceptance criteria, test matrix, validation, and rollback.
5. Create or continue the unit branch.

If the user requests identification or planning only, report the requested
analysis and stop. Do not modify files, create branches, run tests, commit, or
push.

## Execute

Implement only the selected unit.

For a defect, reproduce it with a regression test before applying the
correction.

Run the applicable repository checks and report their actual results.

Update the reverse-engineering document, roadmap status, contracts, and other
affected documentation.

Mark the unit as `review` after implementation and validation. Mark it as
`done` only after review and approval.

Commit with a coherent conventional message referencing the ECOM unit. Push
only to the unit branch. Do not merge without explicit user authorization.

Report the branch, commit, changed files, tests, checks, documentation,
remaining limitations, and next eligible unit.
