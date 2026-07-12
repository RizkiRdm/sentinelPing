# AGENTS.md

Project: SentinelPing

---

# Quick Rules

- MUST read ARCHITECTURE.md before writing any code — layer boundaries and
  the state machine definition there are non-negotiable.
- MUST NOT introduce an ORM, external job queue, or migration framework —
  PROHIBITED per ARCHITECTURE.md Context Lock.
- MUST NOT skip layers (`handler -> repository` direct calls PROHIBITED).
- MUST write state machine unit tests BEFORE wiring the scheduler (see
  PLAN.md Phase 4).
- MUST NOT implement any feature listed in PRD.md Non-MVP Features or
  ARCHITECTURE.md Out-of-Scope, even if asked casually in a follow-up
  prompt — STOP and flag it instead of implementing.
- MUST NOT add authentication mechanisms beyond session cookie (no OAuth,
  no API keys, no SSO) in MVP scope.

---

# Project Context

SentinelPing is a self-hosted, open-source dead-man's-switch / cron job
monitoring tool. Solo developers and small teams add a `curl` ping to their
scheduled jobs (start/success/fail); SentinelPing detects silent failures
(job never ran, or ran but errored) and sends an email alert. Single binary
deployment, SQLite storage, no external dependencies beyond user-supplied
SMTP credentials. Full context: PRD.md.

---

# Tech Stack

- Backend: Go 1.25+, `chi` router, `database/sql` (no ORM),
  `modernc.org/sqlite` (no CGO), `golang.org/x/crypto/bcrypt`.
- Frontend: React 19 (typescript compiler) + Vite + TailwindCSS v4 + shadcn/ui + Recharts, built to
  static assets, embedded via Go `embed.FS`.
- Database: SQLite, single file, hand-rolled migration runner (see
  `internal/db/migrations/`).
- No Redis, no external queue, no ORM. See ARCHITECTURE.md Context Lock
  for the full allowed/forbidden library list.

---

# Architecture Overview

Layers: `handler -> service -> repository -> database`. Scheduler is a
separate background goroutine that calls into `service`, never into
`repository` directly. Full diagrams and call-flow rules: ARCHITECTURE.md
Sections 1-2 and Diagram A.

Modules: `internal/auth`, `internal/monitor` (CRUD + state machine),
`internal/ping` (public ingestion endpoint), `internal/notify` (email),
`internal/scheduler` (checker loop). See ARCHITECTURE.md Section 5.

---

# Project Structure

```
/
├── main.go
├── embed.go                  # //go:embed web/dist
├── internal/
│   ├── auth/
│   ├── monitor/               # includes state machine logic
│   ├── ping/
│   ├── notify/
│   ├── scheduler/
│   └── db/
│       └── migrations/
├── web/                        # React source
│   └── dist/                   # build output, embedded (gitignored source, built at CI/build time)
├── PRD.md
├── ARCHITECTURE.md
├── DESIGN.md
├── PLAN.md
├── AGENTS.md
└── Dockerfile
```

File ownership: each `internal/<module>` owns its own handler, service, and
repository files. MUST NOT reach into another module's internal types
directly — communicate through exported service interfaces only.

---

# Key Commands

- Build backend: `go build -o sentinelping .`
- Run backend (dev): `go run .`
- Build frontend: `cd web && bun install && bun run build`
- Run tests: `go test ./...`
- Run single package tests: `go test ./internal/monitor/...`
- Lint: `go vet ./...` (MVP scope does not require additional linters
  beyond stdlib `go vet` — introducing golangci-lint is acceptable but not
  REQUIRED).
- Format: `gofmt -w .`
- Docker build: `docker build -t sentinelping .`
- Docker run: `docker run -p 8080:8080 -v ./data:/data sentinelping`

---

# Coding Conventions

- MUST follow standard Go project layout and naming (no underscores in Go
  identifiers, exported names PascalCase, unexported camelCase).
- MUST return typed errors from `service` layer, NOT raw `errors.New`
  strings passed directly to HTTP responses.
- MUST centralize HTTP status code mapping in one function/middleware —
  PROHIBITED to scatter `w.WriteHeader(...)` decisions ad-hoc per handler.
- MUST use `log/slog` structured logging exclusively — PROHIBITED to use
  `fmt.Println` or `log.Println` for anything beyond local debug scratch
  work that gets removed before commit.
- MUST use prepared statements for all SQL — PROHIBITED to use string
  concatenation for query building.
- Design pattern: repository pattern for data access, service layer for
  business logic. No repository/service generics abstraction layer unless
  a second database backend is actually introduced (it is not, in MVP).

---

# API Conventions

- Request/response format: JSON only.
- Error contract: `{"error": "code", "message": "human readable"}` — see
  ARCHITECTURE.md Section E Error Contract. MUST be used for every non-2xx
  response.
- Status codes: 400 (validation), 401 (unauthenticated), 403
  (cross-tenant/forbidden), 404 (not found), 500 (internal). No other
  status codes in MVP scope.
- Versioning: none (`/api/*`, no `/v1/` prefix) — see ARCHITECTURE.md
  Versioning Contract.
- Ping ingestion routes (`/ping/:token/*`) are unauthenticated by design —
  MUST NOT apply session middleware to this route group.

---

# Database Conventions

- Schema: see ARCHITECTURE.md Section 3 for the 5 canonical tables. MUST
  NOT deviate from these table definitions without updating
  ARCHITECTURE.md first (spec is source of truth, not the code).
- Migrations: sequentially numbered `.sql` files in
  `internal/db/migrations/`, applied by the hand-rolled runner. MUST NOT
  edit a migration file that has already been applied in any existing
  environment — MUST create a new migration instead.
- Transactions: state transitions (monitor state update +
  state_transitions insert) MUST be atomic — see ARCHITECTURE.md
  Transaction Policy.
- Indexing: see ARCHITECTURE.md Section 3 Index Policy — MUST include the
  composite `(monitor_id, received_at)` index on `pings`, this is
  performance-critical for the checker's per-tick query pattern.
- Naming: snake_case for table and column names, matching the SQL
  definitions in ARCHITECTURE.md exactly.

---

# Security Rules

- Passwords: bcrypt, cost factor MIN 10. PROHIBITED to store or log
  plaintext passwords anywhere, including debug logs.
- SMTP credentials: `smtp_password` MUST be encrypted at rest with
  AES-256-GCM. Encryption key MUST come from an environment variable.
  PROHIBITED to commit any key or secret to the repository.
- Session cookies: HTTP-only, Secure, SameSite=Lax. PROHIBITED to store
  session tokens in `localStorage` or `sessionStorage`.
- Ping tokens: MUST be generated via `crypto/rand`, minimum 32 characters.
  PROHIBITED to use `math/rand` anywhere in token generation.
- Tenant isolation: every monitor-scoped query MUST filter by
  authenticated `user_id`. PROHIBITED to trust a client-supplied user ID
  from request body/params for authorization decisions — MUST derive it
  from the validated session only.
- CORS: disabled, same-origin only. PROHIBITED to add permissive CORS
  headers (`Access-Control-Allow-Origin: *`) anywhere in MVP scope.

---

# Testing Rules

- State machine (`internal/monitor` service layer): MUST have unit test
  coverage for all 8 Acceptance Scenarios in ARCHITECTURE.md Section 6.
  PROHIBITED to merge a state transition code path without a corresponding
  test.
- Repository layer: MUST be tested against in-memory SQLite (`:memory:`),
  NOT mocked interfaces — this validates actual SQL correctness, not just
  Go logic.
- Tenant isolation: MUST have an explicit test proving cross-user data
  access is rejected (403/404), per ARCHITECTURE.md Data Contracts
  Ownership Boundaries.
- Coverage requirement: no fixed percentage target in MVP — REQUIRED
  coverage is scenario-based (all 8 acceptance scenarios, all documented
  failure scenarios in ARCHITECTURE.md Section 6), not an arbitrary
  line-coverage number.

---

# Git Conventions

- Branch naming: `phase-N-<short-description>` matching PLAN.md phase
  numbers (e.g. `phase-4-state-machine`).
- Commit format: conventional-commit style (`feat:`, `fix:`, `test:`,
  `docs:`, `chore:`) with a short imperative summary line.
- Pull requests: not required for solo-dev workflow at MVP stage — direct
  commits to main are acceptable, but MUST NOT commit a phase as "done"
  without its Exit Criteria (PLAN.md) being verifiably met.
- Release: tag `v0.1.0` at Phase 8 completion, per PLAN.md.

---

# AI Agent Rules

- Allowed actions: implement code within the module/layer boundaries
  defined in ARCHITECTURE.md; write tests; write migration files; update
  README.md; suggest Dockerfile/deployment configuration.
- Forbidden actions: MUST NOT add new external dependencies beyond those
  listed in ARCHITECTURE.md Context Lock without explicit user approval.
  MUST NOT implement any Non-MVP Feature or Out-of-Scope item from PRD.md
  / ARCHITECTURE.md even if a follow-up prompt requests it in isolation —
  flag the conflict with the locked spec instead of silently implementing.
  MUST NOT modify ARCHITECTURE.md's state machine definition (Section 6)
  without explicit user confirmation, since it is the core differentiating
  logic of the product.
- Escalation conditions: if a requested change conflicts with a MUST/
  PROHIBITED rule in ARCHITECTURE.md or this file, STOP and surface the
  conflict rather than resolving it silently in either direction.
- Decision boundaries: implementation details not covered by the spec
  (e.g. exact button label wording, minor CSS spacing) MAY be decided
  autonomously. Anything affecting data model, API contract, security
  rule, or state machine behavior MUST NOT be decided autonomously.

---

# Context Hints

- The state machine (ARCHITECTURE.md Section 6) is the single most
  important piece of logic in this codebase — it is the product's core
  differentiator (start+end ping model vs passive-only competitors).
  Treat changes to it with proportional caution.
- `monitors.current_state` is intentionally denormalized for read
  performance — it MUST stay in sync with `state_transitions` via the same
  transaction, never written independently.
- The scheduler tick MUST remain sequential (not parallelized per monitor)
  per Async Policy — this is a deliberate simplicity choice for MVP scale
  (up to ~500 monitors), not an oversight to "optimize later" without
  discussion.
- SMTP is user-bring-your-own — there is no managed email sending service
  in this product. Do not add one.

---

# Known Issues

- None yet (pre-implementation stage). This section MUST be updated by
  whoever implements each phase if workarounds or deferred bugs are
  introduced.

---

# Out of Scope

- Email verification, password reset.
- RBAC, teams, shared workspaces.
- Non-email notification channels (Telegram, Discord, webhook, Slack).
- Public status pages.
- Monitor grouping/tagging/folders.
- Metrics export (Prometheus, etc).
- Mobile app.
- Billing/subscription system.
- Multi-instance/distributed checker coordination.
- API versioning.

See PRD.md Non-MVP Features and ARCHITECTURE.md Non-goals/Out-of-Scope for
full authoritative list — this section is a pointer, not a duplicate
source of truth.
