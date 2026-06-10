# Phase 3 — Horizon Contracts UI : Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Operator-facing contracts UI in Horizon: login, contracts list, contract detail (lifecycle timeline, schedule, audit trail with live chain verification, contextual transition actions), contract creation.

**Architecture:** Extend the existing React 16 + styled-components + axios app — no rewrite, follow existing page/component/lib patterns. Auth = session token in localStorage + axios Authorization header; a 401 on boot `/_info` means auth is enabled → redirect to `/login`. Role from `GET /auth/me` (readonly hides all action buttons).

**Tech Stack:** React 16, react-router-dom 5, styled-components, react-modal, axios (all already in package.json — zero new dependency).

**Repo:** `/home/kali/Desktop/horizon ` (trailing space in dir name), branch `feature/contracts-ui` from `main`. NEVER commit to main. Corren API on `http://localhost:3068`.

---

## File map

| File | Responsibility |
|---|---|
| `src/lib/auth.js` (new) | token store (localStorage `corren_token`), `login/logout/me`, axios default header, `onUnauthorized` redirect |
| `src/lib/ledger.js` (modify) | add `getContracts`, `getContract(id)`, `createContract(body)`, `transition(id, name, input)`, `getAudit(id)` |
| `src/pages/Login.jsx` (new) | username/password form → `POST /auth/login`, stores token, redirects `/contracts` |
| `src/pages/Contracts.jsx` (new) | list with state badges + "New contract" button |
| `src/pages/Contract.jsx` (new) | detail: header, timeline, schedule table, audit trail (+`chain_valid` badge), actions |
| `src/pages/ContractCreate.jsx` (new) | params form → POST, redirect to detail |
| `src/components/ContractsTable.jsx` (new) | table (id, state, cost, markup, progression x/N) |
| `src/components/ScheduleTable.jsx` (new) | seq/date/amount/principal/profit/status with colors |
| `src/components/AuditTrail.jsx` (new) | event list, denied in red with `standard_ref`, chain badge |
| `src/parts/StateBadge.jsx` (new) | colored pill per contract state / installment status |
| `src/parts/TransitionModal.jsx` (new) | confirm modal with per-transition fields (seq, rebate, penalty) |
| `src/parts/Navbar.jsx` (modify) | "Contracts" link + session widget (subject/role/logout) |
| `src/index.jsx` (modify) | routes `/login`, `/contracts`, `/contracts/new`, `/contracts/:id`; 401-on-boot → login |

## Behavior contracts

- API responses are `{ok, data}`; errors are `{error, message, standard_ref?, contract_id?}` (sharia codes) or `{error_message}` (legacy) — the UI shows `message + standard_ref` verbatim on rejections (the Sharia refusal IS a feature, display it prominently).
- Legal transitions per state (drive the action buttons): PROMISE→[acquire,cancel], ACQUIRED→[sell,cancel], SOLD→[pay_installment,early_settle,late_penalty], SETTLED/CANCELLED→[].
- `late_penalty` modal pre-fills destination `@charity:pool` (readonly field — the UI must not even offer another destination).
- Role `readonly` (from `/auth/me`): no action buttons anywhere. Auth disabled: full access, no session widget.
- State colors: PROMISE gray, ACQUIRED blue, SOLD orange, SETTLED green, CANCELLED red; installments: paid green, pending gray, overdue red, settled_early teal.

### Tasks

- [ ] **T1** Branch `feature/contracts-ui`; write `src/lib/auth.js` + extend `src/lib/ledger.js` (methods above, all passing the Authorization header via axios default). Commit.
- [ ] **T2** `Login.jsx` + Navbar session widget + boot-401 handling in `index.jsx`. Commit.
- [ ] **T3** `Contracts.jsx` + `ContractsTable` + `StateBadge` + route + Navbar link. Commit.
- [ ] **T4** `Contract.jsx` + `ScheduleTable` + `AuditTrail` + `TransitionModal` with per-transition fields and error display. Commit.
- [ ] **T5** `ContractCreate.jsx` (fields: id optional, asset_code, cost, markup, client, supplier, installments, first_due, period_days; amounts entered in minor units with a live major-unit hint). Commit.
- [ ] **T6** Browser verification (golden paths, auth off then on, readonly vs operator): corren server + `yarn dev`, walk create → acquire → sell → pay → audit chain_valid; screenshot evidence. Fix everything found. Commit.

Checklist for T6 (manual QA per design spec): boot without API → friendly error; auth on + no token → login; bad password → error shown; operator: create/transition flows; readonly: no buttons; sell-before-acquire shows `AAOIFI-SS-8` rejection in UI; audit badge green `chain_valid`.
