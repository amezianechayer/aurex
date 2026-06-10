# Phase 2 — Postgres prouvé : Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove the entire sharia + auth stack against a real Postgres (the postgres code paths have never been executed), gated by `CORREN_TEST_PG_CONN` so `go test ./...` stays green without a database.

**Architecture:** Reuse the local `postgres:16` docker image via `docker run` for the actual verification; expose the compose postgres service on host port 5433 for future runs. Tests skip when the env var is absent.

**Tech Stack:** Go 1.16, docker, pgx v4. No new dependencies.

**Spec:** `docs/superpowers/specs/2026-06-10-pilot-hardening-design.md` (Phase 2). Branch: `feature/pilot-hardening`.

---

### Task 1: Expose postgres in docker-compose

**Files:** Modify `docker-compose.yml`

- [ ] Add to the existing `postgres` service:

```yaml
    ports:
      - "5433:5432"
```

- [ ] Commit: `chore: expose compose postgres for integration tests`

### Task 2: Gated postgres tests (store + auth + full contract cycle)

**Files:** Create `storage/postgres/sharia_test.go`, `auth/store_pg_test.go`, `sharia/engine_pg_test.go`

- [ ] `storage/postgres/sharia_test.go`: round-trip tests mirroring the sqlite ones (contract, schedule+MarkInstallment, FindDue, audit chain + LastAuditHash) on a unique ledger schema per run (`pgsh<unix-ts>`); skip when `CORREN_TEST_PG_CONN` unset.
- [ ] `auth/store_pg_test.go`: user/key/session round-trip against the `corren_auth` schema; same gating. Drop the schema at test start for idempotence.
- [ ] `sharia/engine_pg_test.go`: `TestPostgresContractLifecycle` — condensed scenarios on a fresh ledger schema: fund treasury, create (24 inst.), sell-before-acquire → `ERR_SHARIA_VIOLATION`, acquire, sell, pay 3 (balance invariant checked), replay pay 1 → `ERR_DUPLICATE`, penalty to income → SS-3 rejection, audit `chain_valid`.
- [ ] Run WITHOUT the env var: all three skip, `go test ./...` green.
- [ ] Commit: `test(postgres): env-gated integration tests for sharia, auth and engine`

### Task 3: Run against real Postgres and fix what breaks

- [ ] Start: `docker run -d --name corren-pg-test -e POSTGRES_USER=ledger -e POSTGRES_PASSWORD=ledger -e POSTGRES_DB=ledger -p 5433:5432 postgres:16`
- [ ] Run: `CORREN_TEST_PG_CONN="postgresql://ledger:ledger@localhost:5433/ledger" go test ./storage/postgres/ ./auth/ ./sharia/ -count=1 -run 'Postgres|RoundTrip'`
- [ ] Fix every failure found in `storage/postgres/sharia.go`, `storage/postgres/migration/v002.sql`, `auth/store.go` (likely suspects: pgx vs database/sql behavior, schema quoting, nullable scans). Each fix: failing run output → fix → green run.
- [ ] Full gated suite green, then `docker rm -f corren-pg-test`.
- [ ] Commit: `fix(postgres): make sharia + auth stores pass integration tests`

### Task 4: Wrap up

- [ ] `go test ./... -count=1` (without env var) — all green.
- [ ] Document the env var in `sharia/README.md` (one line in the test section).
- [ ] Commit.
