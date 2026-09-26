# ECOM-003.1 — SKU schema mapping

Status: done after independent review and user approval. Branch: `fix/ecom-003-1-sku-schema-mapping`. Base: `main` at `d0ffb5e`. The user authorized completion, commit, and push on September 26, 2026; merge into `main` remains a separate decision.

## Objective and observable outcome

Align the SKU persistence entity with the migrated `product_skus` table so that the repository can read representative PostgreSQL rows without querying an inferred table or a nonexistent column.

## Reason and current evidence

The current `Product_skus` model is not idiomatic, has no explicit table mapping, and declares `Stock` as a `product_skus.stock` column. Migration `000002_create_catalog_and_stock.up.sql` does not create that column; inventory is stored in the separate `stock` table. The repository queries the model directly, and no PostgreSQL integration test currently proves that SKU rows can be read through it.

## Dependencies and assumptions

ECOM-002.1 through ECOM-002.6 are implemented, reviewed, and merged. Migration `000002` is the executable source of truth for the current schema. `product_skus` remains the SKU table, while `stock` remains the separate inventory source. This unit does not define availability or inactive-SKU visibility policy and is not expected to add or alter a migration.

## In scope

- Rename `Product_skus` to the idiomatic `ProductSKU` and update its current consumers.
- Declare an explicit `product_skus` table mapping.
- Remove the nonexistent `stock` column from the SKU persistence entity and current JSON representation.
- Add a PostgreSQL repository integration test that reads representative migrated SKU rows.
- Verify the persisted scalar fields, JSONB attributes, active state, count, and basic pagination needed by the current listing path.
- Update reverse engineering, architecture notes where current behavior is recorded, roadmap status, and this implementation note with real evidence.

## Out of scope

Typed SKU filters, correction of the current `sku_id` filter, HTTP query validation, inactive-SKU visibility policy, stock joins, availability projections, OpenAPI annotations, SKU mutation operations, price or currency redesign, migrations, and unrelated cleanup are excluded. Those behaviors remain assigned to ECOM-003.2, ECOM-003.3, or later units.

## Contract, persistence, migrations, and data

The persistence entity will match the columns that currently exist in `product_skus`; no schema or stored-data change is planned. Removing `Stock` also removes the unsupported `stock` field from the current SKU JSON response. This compatibility change is explicitly approved because that value has no authoritative source in the SKU table. A correct availability projection remains reserved for ECOM-003.3.

## Acceptance criteria

1. The persisted type is named `ProductSKU` and explicitly maps to `product_skus`.
2. No field on the persisted SKU entity references the nonexistent `product_skus.stock` column.
3. The repository reads and counts representative SKU rows after the real migration is applied to PostgreSQL.
4. Repository reads preserve the migrated scalar fields, JSONB attributes, active state, and basic pagination behavior.
5. Filters, availability, and inactive-SKU visibility policy remain unchanged by this unit.

## Test matrix and expected pre-correction evidence

| Layer | Cases |
| --- | --- |
| PostgreSQL mapping | Apply the real catalog migration, insert a product and representative SKUs, then list them through `ProductSkuRepository`. |
| PostgreSQL fields | Verify ID, product ID, SKU code, barcode, price, JSONB attributes, active state, timestamps, and total count. |
| PostgreSQL pagination | Verify the existing descending-ID order and a bounded page without expanding filter behavior. |

Before the production correction, the focused PostgreSQL test is expected to fail while reading the model from the migrated table. The planning hypothesis was that GORM would request the nonexistent `stock` column or infer the wrong table; the executed RED cycle must record the actual behavior instead. Compilation, Docker, connection, migration, or fixture failures do not count as defect reproduction.

The executable production change that should make the regression test pass is the JSONB serializer on `Attributes`; removing it must reproduce the scan failure. The explicit table mapping, idiomatic type name, and removal of the unsupported inventory field are separately required structural acceptance criteria. GORM's current `SELECT *` behavior means this integration test does not use an accidental query failure as proof of those source-level decisions.

## Validation gates and commands

- Focused red/green PostgreSQL test: `go test ./internal/catalog/repository -run 'TestProductSkuRepositoryListUsesMigratedSchema' -count=1`
- Catalog packages: `go test ./internal/catalog/... -count=1`
- Full suite and official coverage command: `go test ./... -count=1` and `make test`
- Race: `go test -race ./... -count=1`
- Vet and build: `go vet ./...` and `go build ./...`
- Formatting: `gofmt -l` and `goimports -l` for changed Go files; repository `make fmt` when the pinned/available tool can be used safely
- Lint: `make linter`
- Migration evidence: apply the existing migrations to an isolated PostgreSQL instance and run the focused repository test without skips
- OpenAPI generation/drift: pinned Swaggo command or `make docs`
- Available secret and vulnerability scanners
- Documentation links affected by this unit
- Final integrity: `git diff --check`, staged/unstaged/untracked inspection, and complete diff review

Use task-specific Go caches under `/tmp`. Classify every gate as pass, fail, blocked, unavailable, or not applicable, and never present a skipped or unexecuted check as passing.

## Compatibility, rollout, and rollback

This is an application and test change with no migration or data rewrite. Roll out with the normal API deployment. Roll back by reverting the unit commit; no database rollback is required. Clients that observed the unsupported `stock` field must stop relying on it until ECOM-003.3 supplies an explicit availability projection.

## Human decisions

The bounded scope, branch, removal of the unsupported public `stock` field, and deferral of availability to ECOM-003.3 are approved. No further product decision is required unless implementation reveals a conflict with the migrated schema or an existing tested contract.

## Evidence log

The isolated PostgreSQL 13 database accepted migrations `000001` through `000004` with `ON_ERROR_STOP=1`. The first focused test attempt inside the default sandbox was invalid as RED evidence because local socket access was blocked before the repository ran. It was repeated with approved localhost access.

### Pre-correction evidence — September 25, 2026

The unchanged production code compiled, inserted the fixtures, counted the rows, and reached the real repository `Find`. The focused test then failed for the relevant schema-mapping behavior:

```text
$ SKU_REPOSITORY_TEST_DATABASE_URL='<redacted local test DSN>' GOCACHE=/tmp/ecommerce-go-build-cache GOMODCACHE=/tmp/ecommerce-go-mod-cache go test ./internal/catalog/repository -run 'TestProductSkuRepositoryListUsesMigratedSchema' -count=1 -v
sql: Scan error on column index 5, name "attributes": unsupported Scan, storing driver.Value type []uint8 into type *map[string]interface {}
SELECT * FROM "product_skus" WHERE "product_skus"."deleted_at" IS NULL ORDER BY id DESC LIMIT 2 OFFSET 1
FAIL
```

This disproved two planning assumptions: GORM already inferred `product_skus`, and `SELECT *` did not explicitly request the model's nonexistent `stock` field. The executable defect was the JSONB-to-map scan. The approved test matrix requires JSONB fields to be read, so the minimal mapping correction includes GORM's native JSON serializer in addition to the approved explicit table mapping, type rename, and removal of `Stock`.

### Focused green evidence — September 25, 2026

After the minimal production correction, the identical focused command passed against the same migrated database and fixture path:

```text
=== RUN   TestProductSkuRepositoryListUsesMigratedSchema
--- PASS: TestProductSkuRepositoryListUsesMigratedSchema (0.01s)
PASS
ok  e-commerce-go/internal/catalog/repository  0.013s
```

### Implementation and validation evidence — September 25, 2026

The entity is now named `ProductSKU`, explicitly maps to `product_skus`, excludes the unsupported stock field, and uses GORM's JSON serializer for the migrated JSONB attributes. Repository and service signatures use the renamed type. Filters, handler parsing, availability, migrations, and generated contracts were not changed.

| Gate | Classification | Real result |
| --- | --- | --- |
| Focused red | pass | After migrations and fixtures succeeded, the unchanged repository failed while scanning `attributes` JSONB into the map field. The earlier sandbox socket failure was rejected as invalid evidence. |
| Focused green | pass | The identical focused test passed against PostgreSQL 13 after the minimal mapping correction. |
| Catalog tests and PostgreSQL integration | pass | `go test ./internal/catalog/... -count=1 -v` passed with category, product, and SKU database variables set; the SKU mapping test and all existing repository integrations executed without skips. |
| Full Go suite | pass | `go test ./... -count=1` passed with all repository database variables set. |
| Official test/coverage command | pass | `make test` reported 12 passed packages, 0 failures, 0 skips detected by the script, and 27.6% total statement coverage. The Makefile emitted its pre-existing duplicate `air` target warning. |
| Race | pass | `go test -race ./... -count=1` passed with the repository integrations enabled. |
| Vet | pass | `go vet ./...` exited 0. |
| Build | pass | `go build ./...` exited 0. |
| Formatting | pass | `gofmt -l` and pinned `goimports -l` returned no changed Go paths. |
| Lint | blocked | Installed `golangci-lint` could not decode current Go export data and produced cascading undefined-import typecheck errors in unchanged packages; test, vet, and build resolve those imports successfully. No lint finding is claimed. |
| Migration up | pass | Migrations `000001` through `000004` applied in order to isolated PostgreSQL 13 with `psql -v ON_ERROR_STOP=1`. |
| Migration down | not applicable | This unit adds or changes no migration. |
| OpenAPI generation/drift | blocked | Pinned Swaggo v1.16.6 stopped on the pre-existing `// @Accept JSON` category annotation. No generated file changed. |
| Secret scan | unavailable | `gitleaks` is not installed. No credential or sensitive fixture was added to repository files. |
| Vulnerability scan | unavailable | `govulncheck` is not installed. No dependency was added or changed. |
| Documentation links | not applicable | The documentation updates add no links. |
| Final diff and scope integrity | pass | `git diff --check` and explicit trailing-whitespace checks returned no findings. The index is empty. Six tracked files are modified and two approved files are untracked; the complete tracked diff and full contents of both new files were inspected and contain only ECOM-003.1. |

The isolated database used for implementation evidence was test-only and removed after verification. At this stage the unit remained uncommitted and unpushed for independent review.

## Independent review — September 26, 2026

The independent review approved all five acceptance criteria and found no blocking or important issues. It reproduced the PostgreSQL mapping evidence without skips and confirmed the full, race, vet, build, formatting, and diff gates. Lint and OpenAPI remained blocked by the documented pre-existing limitations; secret and vulnerability scanners remained unavailable.

The initial review made one non-blocking documentation observation: the reverse-engineering update incorrectly said that the integration test applied migrations itself. The text was corrected to state that the test runs against a database where the real migrations were applied. A follow-up review confirmed the correction and returned an unqualified `approved` verdict. The user then authorized completion through commit and push; no merge authorization is inferred.

## Completion validation — September 26, 2026

After the follow-up approval and documentation correction, migrations `000001` through `000004` were applied again to a fresh PostgreSQL 13 container with `ON_ERROR_STOP=1`. The full and race suites passed with category, product, and SKU repository database variables set, so all existing PostgreSQL tests executed without skips. `make test` reported 12 passed packages, 0 failures, 0 skips detected by the script, and 27.6% total statement coverage. Vet, build, `gofmt -l`, `goimports -l`, and `git diff --check` passed.

The final lint rerun reproduced the previously documented export-data incompatibility and cascading typecheck errors in unchanged packages. Pinned Swaggo v1.16.6 again stopped on the pre-existing `// @Accept JSON` category annotation without changing generated files. `gitleaks` and `govulncheck` remained unavailable, and no dependency changed. The final scope inspection found only the eight approved ECOM-003.1 files; the index was empty before staging.
