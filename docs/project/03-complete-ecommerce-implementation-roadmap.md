# Complete E-Commerce Implementation Roadmap

Status: proposed incremental delivery plan

Baseline date: September 11, 2026

Target: complete the E-commerce service maintained in this repository

## 1. Purpose and scope

This roadmap decomposes the complete construction of the E-commerce application into bounded, testable subtasks. It begins after the currently executable API reaches the stabilization gate in [the current-scope quality analysis](./02-current-scope-quality-gap-analysis.md).

The first complete version is a single-seller physical-goods store using BRL and one warehouse. It includes:

- Customer identity, sessions, profiles, and addresses.
- Categories, products, SKUs, prices, and catalog availability.
- Inventory, stock movements, and reservations.
- Cart and wishlist.
- Checkout, immutable snapshots, and orders.
- Payment and fiscal status projections received through external contracts.
- Cancellation, expiration, refund requests, fulfillment, shipment, and returns.
- Administrative operations, audit, observability, security, and recovery.

The following are outside this repository and outside this roadmap:

- Construction of Payment or Billing services.
- PSP, banking, boleto-settlement, refund-processing, or reconciliation internals.
- Fiscal-provider implementation and tax calculation internals.
- Marketplace sellers, commissions, split payments, custody of funds, multiple currencies, and multiple warehouses.

Payment and Billing appear only as external ports and versioned contracts required by E-commerce. Local fakes will provide deterministic integration tests until those external services exist.

"Market-grade" means explicit contracts, safe defaults, automated verification, operational visibility, and recoverable changes. It does not claim conformance with any company's private engineering standards.

## 2. E-commerce ownership

| Capability | E-commerce responsibility |
| --- | --- |
| Identity | Customer account, credentials, session policy, profile, address, role, and ownership authorization |
| Catalog | Categories, products, SKUs, attributes, activation, and public catalog queries |
| Pricing | Authoritative catalog price, currency, history, and price snapshot at checkout |
| Inventory | On-hand balance, reservations, availability, movements, and invariant enforcement |
| Shopping | Cart, cart items, wishlist, and customer ownership |
| Checkout | Server-side validation, totals, snapshots, idempotency, order creation, and reservation |
| Order | Order lifecycle and customer/admin queries |
| Financial projection | Requested/pending/paid/refunded/review status received from an external authority |
| Fiscal projection | Requested/authorized/rejected/cancelled status received from an external authority |
| Fulfillment | Shipment preparation, carrier reference, tracking, delivery, cancellation, and return |
| Integration | E-commerce outbox/inbox, versioned contracts, retries, deduplication, and reconciliation of its projections |

E-commerce never reads an external service database and never marks a payment as captured from frontend input. Redis is not authoritative for inventory, orders, or financial state.

## 3. Non-negotiable invariants

1. Money uses integer minor units and an explicit currency. Totals never use floating point.
2. `total = subtotal - discount + shipping + tax`, with overflow and negative-result protection.
3. Orders preserve customer, address, item, SKU, price, discount, shipping, and tax snapshots.
4. Inventory always satisfies `0 <= reserved <= on_hand`.
5. A cart does not reserve inventory.
6. Checkout persists order, items, snapshots, reservations, idempotency response, and outbox record atomically.
7. Repeating the same command or event cannot duplicate its effect.
8. No SQL transaction remains open during a broker, HTTP, Payment, Billing, carrier, or other external call.
9. Catalog changes never rewrite historical order data.
10. Order, financial, fiscal, and fulfillment states remain separate.
11. Cancellation and deletion never erase commercial history.
12. Administrative actions record actor, timestamp, reason, and affected resource.
13. Personal data follows least privilege, retention, and audited-access policies.

## 4. Decisions required before affected implementation

| Decision | Planning default | Must be resolved before |
| --- | --- | --- |
| Checkout pricing and discount policy | Server recalculates authoritative catalog prices | Checkout implementation |
| Reservation duration | Configurable short reservation | Expiration worker |
| Payment interaction | Versioned asynchronous contract with a deterministic fake | Payment-status integration |
| Fiscal interaction | Versioned asynchronous contract with a deterministic fake | Fiscal-status integration |
| Shipping calculation | Provider-neutral port with a deterministic fake | Checkout shipping quote |
| Shipment provider | Provider-neutral adapter | Real fulfillment integration |
| Anonymous cart | Not included initially | Public cart support |
| Search engine | PostgreSQL first | Measured catalog search limitation |
| Broker | RabbitMQ initially | Durable integration implementation |
| Production traffic and SLOs | Measured targets, not assumptions | Production-readiness approval |

Every accepted decision must become an ADR before code depends on it.

## 5. Delivery phases and subtasks

### Phase A — Stabilized foundation

#### ECOM-010 — Complete the current-scope stabilization roadmap

Deliverables:

- Complete ECOM-002 through ECOM-009 from the current-scope gap analysis.
- Correct category, product, SKU, persistence, cache, security, tests, CI, and documentation.
- Establish the first mandatory quality baseline.

Completion criteria:

- Every currently documented endpoint is executable, tested, observable, and represented accurately in OpenAPI.
- No expansion feature is built on known catalog or persistence defects.

#### ECOM-011 — Approve E-commerce requirements and ADR baseline

Deliverables:

- Versioned functional and non-functional requirements.
- E-commerce aggregate and module ownership map.
- ADRs for identifiers, money, transaction boundaries, inventory authority, cache, events, idempotency, and soft deletion.
- Initial threat model and personal-data inventory.
- Traceability from requirements to APIs, events, migrations, tests, and metrics.

Required evidence:

- Reviewed success, failure, retry, compensation, and authorization scenarios.
- No unresolved decision blocks the next vertical slice.

### Phase B — Customer identity and ownership

#### ECOM-012 — Implement customer registration and account lifecycle

Deliverables:

- Customer registration with normalized email uniqueness.
- Password policy and secure hashing.
- Email verification through an external notification port and local fake.
- Password reset with expiring, one-use tokens.
- Account disablement and audited administrative actions.
- Enumeration-resistant public responses.

Required tests:

- Duplicate normalized email, weak password, expired/replayed token, disabled account, enumeration behavior, and concurrent registration.

#### ECOM-013 — Implement sessions, profile, and authorization policies

Deliverables:

- Access and refresh/session strategy with rotation and revocation.
- Logout and revoke-all-sessions use cases.
- Role- and resource-ownership authorization policies.
- Customer profile reads and updates with explicit field-presence semantics.
- Immediate or explicitly bounded handling of role and account-status changes.

Required tests:

- Token rotation replay, revoked session, expired token, role change, disabled user, and cross-customer access.

#### ECOM-014 — Implement customer addresses

Deliverables:

- Address create, read, update, deactivate, and list operations.
- Shipping, billing, and other address types.
- At most one active default address per customer and type.
- Postal/address validation behind a replaceable port when external validation is needed.
- Ownership enforcement in every query and mutation.

Required tests:

- Concurrent default selection, cross-user access, invalid address, deactivation of a default, and snapshot independence after checkout.

Completion criteria for Phase B:

- A customer can safely create and manage an account, sessions, profile, and addresses without accessing another customer's resources.

### Phase C — Complete catalog and pricing

#### ECOM-015 — Complete category, product, and SKU administration

Deliverables:

- Complete category, product, and SKU lifecycle.
- Category hierarchy without cycles.
- Validated SKU attributes and stable public identifiers.
- Activation and deactivation rules that preserve historical references.
- Administrative audit records.
- Deterministic filtering, ordering, and bounded pagination.

Required tests:

- Slug and SKU conflicts, category cycles, inactive relations, partial updates, authorization, concurrent edits, and contract compatibility.

#### ECOM-016 — Implement authoritative pricing and history

Deliverables:

- Price value in integer minor units with currency.
- Atomic current-price update and append-only price history.
- Actor, reason, and effective timestamp for administrative changes.
- Price query used by checkout through an application port.
- Explicit initial discount policy; no hidden client-provided discount authority.

Required tests:

- Invalid/overflowing value, unsupported currency, concurrent update, history atomicity, inactive SKU, and historical-order independence.

#### ECOM-017 — Complete public catalog discovery

Deliverables:

- Public category, product, SKU, price, and availability projections.
- Search and filters implemented with PostgreSQL first.
- Cache keys, TTL, invalidation, and database fallback.
- Stable public pagination and response contracts.
- Query metrics for latency, result count, and cache effectiveness without personal data.

Required tests:

- Filter combinations, stable ordering, empty pages, cache failure, stale-write prevention, inactive-item visibility, and representative query plans.

Completion criteria for Phase C:

- Administrators can manage a consistent catalog and customers can discover only sellable items through tested contracts.

### Phase D — Inventory and reservations

#### ECOM-018 — Implement inventory balances and movements

Deliverables:

- Inventory balance per SKU for the initial warehouse.
- Append-only stock movements with reason, actor, reference, and timestamp.
- Atomic increase, decrease, correction, and reconciliation use cases.
- Database constraints enforcing non-negative values.
- Administrative inventory query and adjustment endpoints.

Required tests:

- Concurrent adjustments, insufficient balance, duplicate reference, rollback, audit trail, and invariant enforcement in PostgreSQL.

#### ECOM-019 — Implement reservation lifecycle

Deliverables:

- Reservation entity with `active`, `consumed`, `released`, and `expired` states.
- Atomic multi-SKU reserve, consume, release, and expire operations.
- Stable lock ordering by SKU to reduce deadlocks.
- Command idempotency and unique business references.
- Expiration worker with retry, metrics, and safe shutdown.

Required tests:

- Two customers competing for the last unit.
- Complete rollback when the last SKU in a request is unavailable.
- Duplicate reserve/consume/release commands.
- Expiration racing with consumption.
- Deadlock retry and worker restart.

Completion criteria for Phase D:

- Stock cannot be oversold under tested concurrent checkout behavior, and every balance change is traceable.

### Phase E — Shopping experience

#### ECOM-020 — Implement persistent cart

Deliverables:

- One active cart per authenticated customer under the initial policy.
- Add, update, remove, clear, and list item operations.
- Upsert behavior for the same SKU.
- Server-side validation of SKU activation and quantity.
- Cart expiration without inventory reservation.
- Price and availability warnings without silently rewriting customer intent.

Required tests:

- Concurrent updates, duplicate SKU addition, invalid quantity, inactive SKU, expired cart, cross-user access, and catalog price change.

#### ECOM-021 — Implement wishlist

Deliverables:

- Add, remove, and list wishlist items.
- One active entry per customer and SKU.
- Ownership enforcement and inactive-item presentation policy.

Required tests:

- Duplicate addition, cross-user access, inactive/deleted SKU, pagination, and concurrent add/remove.

Completion criteria for Phase E:

- Customers can maintain shopping intent safely without mutating inventory.

### Phase F — Checkout and orders

#### ECOM-022 — Implement checkout quotation

Deliverables:

- Server-side cart validation and current-price recalculation.
- Shipping quote port with deterministic local fake.
- Explicit subtotal, discount, shipping, tax, total, and currency.
- Customer and address validation.
- Short-lived quote identifier if the chosen policy requires it.
- No external call inside a database transaction.

Required tests:

- Price change, inactive SKU, insufficient stock, invalid address, shipping timeout, overflow, and quote expiration.

#### ECOM-023 — Implement idempotent order creation

Deliverables:

- Mandatory checkout idempotency key.
- Immutable customer, address, item, SKU, attribute, price, discount, shipping, and tax snapshots.
- Atomic order, items, inventory reservations, idempotency response, and outbox record.
- Separate order, financial, fiscal, and fulfillment states.
- Customer and administrative order queries.

Required tests:

- Identical retry returns the original order.
- Same key with a different payload is rejected.
- Multi-SKU rollback, stale cart, concurrent checkout, address mutation after checkout, and database retry.

#### ECOM-024 — Implement order lifecycle and expiration

Deliverables:

- Explicit valid transitions for awaiting payment, confirmed, cancelled, expired, completed, and review.
- Expiration worker coordinated with reservation locks.
- Immutable transition history with actor or event source.
- Customer-visible state and reason mapping without internal leakage.

Required tests:

- Duplicate transition, invalid transition, expiration racing with confirmation, worker restart, and clock-boundary behavior.

Completion criteria for Phase F:

- A customer can transform a cart into a durable, idempotent order with valid totals, snapshots, and reserved inventory.

### Phase G — External contracts owned by E-commerce

#### ECOM-025 — Implement transactional outbox and inbox

Deliverables:

- Transactional outbox for order and fulfillment events.
- Publisher worker with retry, backoff, metrics, and safe shutdown.
- Inbox and deduplication for external events consumed by E-commerce.
- Versioned envelope with event ID, type, schema version, correlation, causation, producer, and occurrence time.
- Contract compatibility checks in CI.
- Dead-letter handling and audited replay command.

Required tests:

- Broker unavailable after database commit.
- Publish succeeds but acknowledgement is lost.
- Duplicate, delayed, and out-of-order delivery.
- Consumer restart between persistence and acknowledgement.
- Poison message and controlled replay.

#### ECOM-026 — Integrate external payment outcomes

Scope: E-commerce contract and projection only; Payment implementation remains external.

Deliverables:

- Publish an idempotent payment request or order event according to the accepted contract.
- Consume pending, captured, failed, cancelled, expired, partially refunded, and refunded outcomes.
- Reject mismatched order, amount, currency, producer, or schema version.
- Consume active reservations exactly once after confirmed capture.
- Implement late-payment handling: re-reserve when allowed or enter review and request compensation.
- Store a financial projection separate from order state.
- Provide deterministic Payment fake for integration and E2E tests.

Required tests:

- Duplicate capture, failed event before captured event, amount mismatch, payment after expiration, capture racing with cancellation, and fake-service outage.

#### ECOM-027 — Integrate external fiscal outcomes

Scope: E-commerce request and projection only; Billing/fiscal-provider implementation remains external.

Deliverables:

- Produce a versioned fiscal request containing the approved immutable snapshot.
- Consume requested, processing, authorized, rejected, cancellation-pending, and cancelled outcomes.
- Gate fulfillment on fiscal authorization only when required by accepted policy.
- Present operational rejection reasons without exposing sensitive provider payloads.
- Provide deterministic Billing/fiscal fake for integration and E2E tests.

Required tests:

- Duplicate outcome, rejection and correction, out-of-order authorization, cancellation race, missing required snapshot, and fake-service outage.

Completion criteria for Phase G:

- E-commerce remains consistent across duplicate, delayed, unavailable, and reordered external interactions without accessing external databases.

### Phase H — Fulfillment, cancellation, and returns

#### ECOM-028 — Implement fulfillment and shipment tracking

Deliverables:

- Fulfillment and shipment aggregates independent from financial state.
- Valid transitions for pending, ready, in transit, delivered, returned, and cancelled.
- Carrier port with deterministic local fake.
- Tracking reference, shipment cost, dimensions, timestamps, and audit history.
- Customer and administrative shipment views.

Required tests:

- Invalid transition, duplicate carrier event, fiscal gate, cancellation after dispatch, delivered-order behavior, cross-user access, and malformed dimensions.

#### ECOM-029 — Implement cancellation and compensation orchestration

Deliverables:

- Cancellation policy based on order, reservation, financial, fiscal, and fulfillment projections.
- Idempotent reservation release, payment-refund request, fiscal-cancellation request, and shipment-cancellation request.
- Visible pending-compensation and manual-review states.
- Audited administrative override policy without direct database state editing.

Required tests:

- Cancellation before payment, during uncertain payment, after capture, after fiscal authorization, after shipment, and during partial refund.

#### ECOM-030 — Implement return lifecycle

Deliverables:

- Return request, approval, receipt, rejection, restock decision, and completion states.
- Item-level return quantities bounded by fulfilled quantities.
- Inventory movement and refund-request contracts triggered by approved policy.
- Customer and administrative views with audit history.

Required tests:

- Partial return, duplicate request, excess quantity, non-returnable state, concurrent cancellation, restock/no-restock, and refund-outcome delay.

Completion criteria for Phase H:

- Orders can be delivered, cancelled, compensated, or returned with complete history and without erasing inventory or commercial evidence.

### Phase I — Administration and operations

#### ECOM-031 — Implement administrative operations and audit

Deliverables:

- Least-privilege administrative roles or permissions.
- Search and inspection for customers, catalog, inventory, orders, integrations, and pending compensation.
- Audited inventory adjustment, account disablement, order review, and replay requests.
- Reason required for sensitive actions.
- No endpoint that directly edits financial or immutable historical state.

Required tests:

- Permission matrix, self-escalation prevention, audit completeness, sensitive-data redaction, and concurrent operator actions.

#### ECOM-032 — Complete observability and incident response

Deliverables:

- Correlated logs, metrics, and traces across HTTP, database, Redis, workers, broker, and external ports.
- HTTP RED metrics and backlog/age metrics for reservations, outbox, inbox, retries, and compensation.
- Business-invariant and projection-discrepancy alerts.
- Initial SLOs and error-budget policy based on measured traffic.
- Runbooks for inventory discrepancy, stale reservation, message replay, late payment, external outage, and failed compensation.

Completion criteria:

- An operator can identify a failed customer journey from request ID or order ID and execute a documented recovery procedure.

### Phase J — Production certification

#### ECOM-033 — Complete security and data-protection review

Deliverables:

- Updated threat model for all E-commerce trust boundaries.
- Least-privilege database, Redis, broker, object, and operator permissions.
- TLS and managed secret/key rotation procedures.
- Personal-data minimization, retention, export, anonymization, and audited-access procedures.
- Secret, dependency, source, container, and dynamic security scans.
- Independent security review before real customer or financial traffic.

#### ECOM-034 — Validate performance and recovery

Deliverables:

- Representative load, stress, and soak tests.
- Measured limits for HTTP, workers, PostgreSQL, Redis, and broker.
- Query plans and indexes validated against representative data volume.
- Encrypted backups and point-in-time recovery where available.
- Documented RPO/RTO and completed restoration exercises.
- Reconciliation after restoration.

#### ECOM-035 — Validate deployment and migration safety

Deliverables:

- Immutable signed image and SBOM.
- Progressive deployment and rollback procedure.
- Forward-compatible expand/backfill/validate/contract migrations.
- Feature flags for risky transitions.
- Resource limits, autoscaling inputs, probes, and graceful worker shutdown.
- Production configuration and access review.

#### ECOM-036 — Execute complete E2E and failure certification

Required journeys:

1. Registration through delivered order.
2. Checkout retry with the same idempotency key.
3. Two customers competing for the final unit.
4. Delayed, duplicated, and reordered payment outcomes.
5. Payment after reservation expiration.
6. Fiscal rejection and correction.
7. Cancellation before and after each external milestone.
8. Partial/full refund projection and partial/full return.
9. PostgreSQL, Redis, broker, worker, carrier, Payment fake, and Billing fake failures.
10. Backup restoration followed by state reconciliation.

Completion criteria:

- Every critical invariant has automated evidence.
- All high-severity security findings are resolved or formally risk-accepted.
- SLOs, alerts, ownership, runbooks, and recovery evidence exist.
- OpenAPI, event schemas, migrations, dashboards, documentation, and deployed versions are traceable.

## 6. Dependency map

```text
ECOM-010 Current stabilization
    -> ECOM-011 Requirements and ADRs
    -> ECOM-012..014 Identity, sessions, and addresses
    -> ECOM-015..017 Catalog, pricing, and discovery
    -> ECOM-018..019 Inventory and reservations
    -> ECOM-020..021 Cart and wishlist
    -> ECOM-022..024 Checkout and orders
    -> ECOM-025 Durable integration
       -> ECOM-026 Payment projection
       -> ECOM-027 Fiscal projection
    -> ECOM-028..030 Fulfillment, cancellation, and returns
    -> ECOM-031..032 Administration and operations
    -> ECOM-033..036 Production certification
```

Some work may proceed in parallel only after ownership, invariants, and contracts are accepted. Parallel implementation must retain compatible migrations and consumer-driven contract tests.

## 7. Delivery policy for every subtask

Each subtask is delivered as a small vertical slice:

1. Confirm requirements, invariants, ownership, and failure behavior.
2. Record or update the applicable ADR.
3. Define HTTP and event contracts before adapter integration.
4. Add a forward-safe migration when persistence changes.
5. Implement domain and application behavior behind focused ports.
6. Implement adapters without leaking infrastructure details into domain rules.
7. Add unit, integration, concurrency, contract, and failure tests proportional to risk.
8. Add logs, metrics, traces, alerts, and operational controls.
9. Update OpenAPI, event schemas, runbooks, reverse engineering, and maintained documentation in English.
10. Run all quality and security gates, commit, and push before beginning the next subtask.

## 8. E-commerce Definition of Done

The first complete E-commerce version is done only when:

- The customer journey from account creation through delivery is executable.
- Cancellation, expiration, late payment, fiscal rejection, return, and external-service ambiguity are recoverable.
- Inventory and order invariants hold under concurrency, retries, duplicates, and restarts.
- E-commerce owns and migrates its data without cross-service database access.
- Every outbound request and inbound event is idempotent or explicitly reconciled.
- Security, privacy, observability, deployment, backup, and incident procedures have executable evidence.
- Contracts and documentation describe only behavior that exists and is tested.
