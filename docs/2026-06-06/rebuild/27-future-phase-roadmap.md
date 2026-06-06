# Future Phase Roadmap

Date: 2026-06-06
Status: Abstract and changeable plan
Baseline: Phase 15 complete at `96210ba`
Planning review: `6bb17a1`

## 1. How To Use This Roadmap

This roadmap converts the post-Phase-15 capability batches into tentative
delivery phases.

It is intentionally changeable:

- a phase may be split when reconnaissance finds excessive risk
- adjacent phases may be merged when their contracts are inseparable
- implementation order may change when a dependency is discovered
- a phase may retire legacy behavior instead of migrating it
- no phase begins until its entry gate is satisfied
- every implementation phase ends with tests, review, documentation, and a Git
  checkpoint

Phase numbers describe the current expected order, not an irreversible promise.

## 2. Standard Phase Workflow

Every future phase follows this control loop:

```text
Codex defines scope and acceptance criteria
-> OMP performs bounded reconnaissance
-> Codex reviews and approves the plan
-> OMP implements the approved scope
-> Codex reviews the diff and corrects issues
-> tests, vet, build, contract checks, and relevant parity checks run
-> phase document and Git checkpoint are created
```

Every phase must explicitly list:

- goal
- included responsibilities
- excluded responsibilities
- dependencies
- risks
- verification
- rollback or failure behavior
- completion evidence

## 3. Planned Phases

### Phase 16: Compatibility and Completion Baseline

Goal:

- create a trusted inventory of supported legacy routes, frontend workflows,
  scheduler responsibilities, and data ownership

Expected outcomes:

- classify every legacy route as migrated, remaining, dev-only, stub, or retire
- map frontend workflows to required backend capabilities
- produce a reviewable responsibility matrix with valid dispositions and owners
- define compatibility levels for status, shape, semantics, and side effects
- establish the measurable completion dashboard

Non-goals:

- no gameplay migration
- no database implementation
- no frontend cutover

Completion gate:

- the responsibility matrix contains every production route, background job,
  persistent data group, and supported frontend workflow, with an approved
  owner, disposition, dependency group, and verification requirement

### Phase 17: Formula and Static Data Governance

Goal:

- establish one governed source of truth for economic formulas and static data

Expected outcomes:

- define the backend-next formula boundary
- identify legacy formulas that must retain exact behavior
- define formula and static-data versioning expectations
- add parity fixtures for critical economic calculations

Non-goals:

- no broad gameplay migration
- no intentional economy rebalance

Completion gate:

- critical formulas cannot silently diverge between legacy and backend-next

### Phase 18: Domain, Error, and Transaction Boundaries

Goal:

- stabilize the architectural rules needed by the remaining high-risk writes

Expected outcomes:

- define what belongs in domain rules versus app orchestration
- define typed application error categories
- define transaction, rollback, idempotency, and concurrency expectations
- consolidate runtime dependency composition expectations
- assign ownership and replacement policy for rate limiting, anti-cheat, AML,
  and other cross-cutting safety controls

Non-goals:

- no large-scale refactor merely for directory purity
- no new gameplay scope unless required to prove the boundaries

Completion gate:

- new write paths can follow one documented atomicity and error model

### Phase 19: Trusted Parity and Shadow Verification

Goal:

- turn the existing bridge utilities into trustworthy compatibility evidence

Expected outcomes:

- make shape comparison symmetric
- define real semantic and side-effect comparison
- establish reproducible state fixtures
- report approved differences separately from regressions

Non-goals:

- no production traffic cutover

Completion gate:

- migrated capabilities can be compared against legacy without false passes

### Phase 20: Low-Risk Remaining Read Groups

Goal:

- migrate or retire remaining read-only capabilities required by supported
  frontend workflows

Candidate responsibility groups:

- leaderboard and public company/player views
- remaining market and finance reads
- recipes, production options, and configuration views
- social, contract, research, and progression reads

Delivery rule:

- treat each independent read group as a separately reviewable sub-checkpoint;
  split this phase if one group develops non-trivial domain behavior

Non-goals:

- no mutations or scheduler behavior

Completion gate:

- supported read-only frontend workflows have backend-next ownership and parity
  evidence

### Phase 21: Research and Social Domains

Goal:

- complete bounded domains whose storage abstractions already exist

Candidate responsibility groups:

- research lifecycle
- messages, chat, contacts, and notifications

Non-goals:

- no scheduler-wide orchestration
- no unrelated player progression writes

Completion gate:

- selected domains are independently owned, tested, and no longer require
  legacy runtime behavior

### Phase 22: Building and Warehouse Lifecycle

Goal:

- migrate the complete player-triggered building lifecycle

Candidate responsibility groups:

- market and purchase
- placement and coordinate rules
- movement and demolition
- building and warehouse upgrades
- warehouse capacity, inventory grouping, and storage invariants

Non-goals:

- no scheduler work

Completion gate:

- money, placement, warehouse capacity, inventory storage, ownership, and
  rollback invariants are proven

### Phase 23: Production Lifecycle Completion

Goal:

- complete remaining production operations around the already migrated core

Candidate responsibility groups:

- queue views
- cancellation
- claim-all
- slot management
- remaining production entry compatibility

Non-goals:

- no background scheduler migration

Dependencies:

- Phase 22 warehouse capacity and inventory invariants

Completion gate:

- all supported player-triggered production workflows are owned by backend-next

### Phase 24: Player Progression and Offline Behavior

Goal:

- migrate supported progression and account-adjacent gameplay behavior

Candidate responsibility groups:

- level and XP
- achievements and rewards
- boosts
- tutorial state
- offline income

Non-goals:

- no broad scheduler migration unless offline behavior requires a bounded job

Completion gate:

- progression mutations and rewards are deterministic, persistent-ready, and
  protected against duplicate claims

### Phase 25: Market and Finance Lifecycle Completion

Goal:

- complete remaining player-triggered market and financial capabilities

Candidate responsibility groups:

- remaining market order lifecycle and trade views
- remaining finance views
- explicit ledger invariants
- fee and balance behavior

Non-goals:

- no bot market cycle
- no periodic settlement

Completion gate:

- player-triggered money, inventory, order, trade, and ledger behavior has
  atomicity and parity evidence

### Phase 26: Bond Lifecycle Completion

Goal:

- complete player-triggered bond ownership and liability behavior

Candidate responsibility groups:

- bond purchase
- bond call or redemption
- holdings and issuer state transitions
- default-related state required before scheduler settlement

Non-goals:

- no periodic interest scheduler yet

Completion gate:

- bond ownership, money movement, liability, and rollback invariants are proven

### Phase 27: Contracts and Daily Orders

Goal:

- migrate or formally retire the remaining bounded gameplay systems

Candidate responsibility groups:

- government contracts
- daily orders

Non-goals:

- no scheduler execution until domain ownership is complete
- no auctions or executives

Completion gate:

- player-triggered contract and daily-order state transitions are owned,
  deterministic, and ready for scheduler integration

### Phase 28: Auctions, Executives, and Legacy Scope Retirement

Goal:

- migrate the remaining supported bounded gameplay systems and formally retire
  unsupported legacy scope

Candidate responsibility groups:

- auctions
- executives
- aerospace and other legacy stubs

Non-goals:

- no scheduler execution until domain ownership is complete

Completion gate:

- each remaining gameplay system is migrated, retired, or explicitly deferred
  outside the rewrite definition

### Phase 29: Scheduler Framework and Job Contracts

Goal:

- establish the backend-next background runtime without migrating every job at
  once

Expected outcomes:

- typed job contracts
- preserved or deliberately revised ordering rules
- observability, cancellation, retries, idempotency, and replay expectations
- deterministic clock-driven tests

Non-goals:

- no unreviewed copy of the complete legacy scheduler

Completion gate:

- isolated jobs are observable, replayable, cancellation-aware, failure-aware,
  and verifiable with deterministic tests

### Phase 30: Economic Background Jobs

Goal:

- migrate scheduler-owned economic behavior

Candidate responsibility groups:

- bond settlement and defaults
- government awards and defaults
- bot market cycle
- market lock behavior
- cleanup and refresh jobs

Non-goals:

- no production cutover before deterministic replay evidence exists

Completion gate:

- approved state fixtures produce approved results across complete scheduler
  cycles

### Phase 31: PostgreSQL Storage and Schema

Goal:

- implement the persistent backend-next storage model

Expected outcomes:

- versioned schema migrations
- PostgreSQL implementations of required storage contracts
- explicit transaction boundaries
- shared contract tests for memory and PostgreSQL

Non-goals:

- no legacy data cutover yet

Completion gate:

- PostgreSQL passes storage, transaction, restart, and failure tests

### Phase 32: Legacy Data Migration and Recovery

Goal:

- safely transform legacy persisted state into the backend-next storage model

Expected outcomes:

- migration tooling and validation reports
- backups and rollback procedure
- reconciliation for money, inventory, ownership, orders, trades, ledger, bonds,
  progression, and social data

Non-goals:

- no irreversible production migration without rollback evidence

Completion gate:

- representative and full-size migration rehearsals preserve approved state

### Phase 33: Frontend Workflow Cutover

Goal:

- move supported frontend workflows to backend-next in controlled groups

Expected outcomes:

- workflow-level compatibility checks
- feature flags or routing controls
- monitoring and rollback for each cutover group
- removal of frontend dependencies on retired routes

Non-goals:

- no all-at-once cutover without evidence

Completion gate:

- each approved cutover group passes its workflow suite and rollback check
- the phase completes when every supported workflow is assigned to a successful
  cutover group and no supported workflow requires legacy traffic

### Phase 34: Production Readiness and Security Review

Goal:

- verify the complete replacement as an operable production service

Expected outcomes:

- race tests in a supported environment
- load and failure testing
- health and readiness dependency checks
- security, auth, rate-limit, anti-cheat, and AML disposition
- scheduler and database operational runbooks

Non-goals:

- no legacy deletion yet

Completion gate:

- production readiness checklist and rollback rehearsal are approved

### Phase 35: Legacy Retirement

Goal:

- disable and remove the legacy runtime after replacement evidence is complete

Expected outcomes:

- legacy traffic disabled
- final parity and data reconciliation
- dead route and dependency removal
- archived migration records and operational documentation

Completion gate:

- legacy receives zero required production traffic for an agreed observation
  period
- backend-next-only frontend and operational suites pass
- data reconciliation and rollback evidence are approved
- deleting or archiving legacy does not remove any responsibility listed as
  supported in the Phase 16 matrix

## 4. Expected Change Points

The roadmap should be reconsidered at these checkpoints:

- after Phase 16, because route and workflow classification may retire scope
- after Phase 19, because trusted parity evidence may expose hidden behavior
- after Phase 28, because all gameplay ownership must be known before scheduler
- before Phase 31, because stable domain contracts should shape the database
- before Phase 33, because cutover groups depend on actual frontend usage
- before Phase 35, because legacy retirement requires operational evidence

## 5. Rules for Changing the Plan

A roadmap change is healthy when it:

- reduces cross-domain blast radius
- makes completion evidence stronger
- exposes hidden dependencies earlier
- retires unsupported or unused behavior deliberately
- improves rollback and data safety

A roadmap change is unhealthy when it:

- skips verification to move faster
- combines unrelated domains into one large phase
- introduces database schema before domain ownership is understood
- moves frontend traffic before scheduler and persistence responsibilities are
  clear
- treats passing unit tests as proof of full behavioral parity

Any material roadmap change should update this document or its successor and
create a Git checkpoint before implementation continues.
