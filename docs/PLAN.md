# PLAN.md

Project: SentinelPing
Scope: MVP ONLY — from repository init through production deployment.
No non-MVP phase is included in this document. Anything not listed here is
OUT OF SCOPE for this plan (see PRD.md Non-MVP Features and
ARCHITECTURE.md Non-goals / Out-of-Scope).

Budget constraint: ~1 hour/day solo development. Sequencing below is
dependency-aware — each phase assumes the prior phase is functionally
complete, not perfect. Do not polish beyond what a checklist item requires
before moving to the next item.

---

## Phase 0 — Repository & Environment Foundation

- [ ] Init Go module (`go mod init`), commit `.gitignore` (Go + Node +
      SQLite file + `.env`).
- [ ] Set up directory skeleton per ARCHITECTURE.md Section 5 Module Rules:
      `internal/auth`, `internal/monitor`, `internal/ping`,
      `internal/notify`, `internal/scheduler`, `internal/db`.
- [ ] Add `internal/db/migrations/0001_init.sql` with all 5 tables from
      ARCHITECTURE.md Section 3 (users, monitors, pings, state_transitions,
      notification_log).
- [ ] Write minimal hand-rolled migration runner (embedded SQL files,
      version tracking table) — PROHIBITED to add external migration
      framework per Context Lock.
- [ ] Verify: `go run .` boots, connects to local SQLite file, runs
      migration once, does not re-run on second boot.
- [ ] Init frontend scaffold: Vite + React 19 + TailwindCSS + shadcn/ui
      init. Confirm `npm run build` outputs to `web/dist`.
- [ ] Add root `embed.go` with `//go:embed web/dist` and a placeholder
      static handler serving `index.html`.
- [ ] Verify: single `go build` produces one binary that serves the Vite
      placeholder page.

**Exit criteria**: one binary boots, connects to DB, serves a static page.
No business logic yet.

---

## Phase 1 — Auth (Foundation for everything else)

- [ ] Implement `internal/auth`: signup (email + password), bcrypt hash
      (cost 10 per ARCHITECTURE.md), login, session cookie issuance
      (HTTP-only, Secure, SameSite=Lax).
- [ ] Implement session validation middleware (STRICT: applied to all
      `/api/*` except `/api/auth/signup` and `/api/auth/login`, per
      Security Contract).
- [ ] Write unit tests for signup/login happy path + duplicate email
      rejection + wrong password rejection.
- [ ] Build minimal frontend: signup form, login form, logout button. No
      styling polish beyond shadcn/ui defaults.
- [ ] Verify end-to-end: signup via UI → session cookie set → protected
      route accessible → logout clears session.

**Exit criteria**: a user can sign up, log in, log out. No monitor
functionality yet.

---

## Phase 2 — Monitor CRUD (No State Machine Yet)

- [ ] Implement `internal/monitor` repository: create, list (scoped to
      `user_id`), get by ID (scoped to `user_id`), update, soft state via
      `paused` flag, delete.
- [ ] Implement service layer validation per ARCHITECTURE.md Execution
      Constraints (positive integers, MAX_LIMIT 2592000s).
- [ ] Implement `ping_token` generation via `crypto/rand`, minimum 32
      chars, unique constraint enforced at DB level.
- [ ] Implement handlers: `POST /api/monitors`, `GET /api/monitors`,
      `GET /api/monitors/:id`, `PUT /api/monitors/:id`,
      `DELETE /api/monitors/:id`.
- [ ] STRICT: every query MUST filter by authenticated `user_id` — write a
      test that confirms User A cannot fetch/edit/delete User B's monitor
      (403 or 404, per Error Contract).
- [ ] Build frontend: monitor list page (empty state per DESIGN.md), create
      monitor dialog form, monitor detail page showing generated
      `ping_token` and copy-paste `curl` snippets for start/success/fail.
- [ ] Verify end-to-end: create monitor via UI → see it in list → copy
      curl snippet → delete monitor → confirm removed from list.

**Exit criteria**: full monitor CRUD works through the UI. State is always
`PENDING` (state machine not implemented yet — this phase is intentionally
inert).

---

## Phase 3 — Ping Ingestion

- [ ] Implement `internal/ping` handler:
      `GET /ping/:token/{start|success|fail}` — no auth required.
- [ ] Implement token format validation BEFORE DB query (reject malformed
      tokens with 400, per Security Rules).
- [ ] Implement `pings` table insert (fast path, no transaction with state
      evaluation, per Execution Constraints Transaction Policy).
- [ ] Write tests: valid token + valid type → 200; invalid token format →
      400; well-formed but nonexistent token → 404; duplicate rapid pings
      → both persisted (per Verification Rules Failure Scenario 3).
- [ ] Manual verification: run `curl` against a live monitor's ping URLs
      from a terminal, confirm rows appear in `pings` table.

**Exit criteria**: pings are ingested and persisted. State machine still
not wired — this phase is pure ingestion plumbing.

---

## Phase 4 — State Machine (Core Value Delivery)

⚠️ This is the highest-complexity phase. Do not rush past the acceptance
scenarios — this is the differentiating logic of the entire product.

- [ ] Implement state machine logic in `internal/monitor` service layer per
      ARCHITECTURE.md Section 6 (5 states: PENDING, HEALTHY, RUNNING, LATE,
      FAILED).
- [ ] Write unit tests for ALL 8 Acceptance Scenarios in
      ARCHITECTURE.md Section 6 BEFORE wiring the scheduler (test the pure
      function first, in isolation, against fixed/mocked timestamps).
- [ ] Implement the transactional write (state update + state_transitions
      insert in one DB transaction, per Transaction Policy).
- [ ] Implement notify-on-transition-edge-only logic (Acceptance Scenario 8) — write a specific test proving repeated checker ticks in the
      same failing state do NOT re-trigger notification dispatch.

**Exit criteria**: state machine function is fully unit-tested and passes
all 8 acceptance scenarios in isolation, without the scheduler or real
clock involved yet.

---

## Phase 5 — Scheduler (Wiring State Machine to Real Time)

- [ ] Implement `internal/scheduler`: single goroutine, `time.Ticker`,
      configurable interval (env var, DEFAULT 30s, MIN_LIMIT 10s per
      Non-goals).
- [ ] Implement sequential (not parallel) processing of all monitors per
      tick, per Async Policy.
- [ ] Implement `defer recover()` at tick level (one monitor's panic MUST
      NOT crash the loop or skip remaining monitors, per Error Handling
      Policy).
- [ ] Implement structured logging for tick start/end + monitor count +
      duration, per Logging Policy.
- [ ] Verify end-to-end: create a monitor with short intervals (e.g. 60s
      expected interval, 30s grace), do not ping it, observe state
      transition to LATE in the dashboard within expected drift window.
- [ ] Verify: send a `fail` ping manually via curl, observe immediate
      transition to FAILED on next tick regardless of prior state
      (Acceptance Scenario 7).

**Exit criteria**: state transitions happen automatically over real time
without manual DB manipulation. This is the point where the product
"works" end-to-end minus notifications.

---

## Phase 6 — Notification Dispatch

- [ ] Implement `internal/notify`: SMTP client using user's stored
      credentials (decrypt `smtp_password` via AES-256-GCM per Security
      Rules).
- [ ] Implement dispatch trigger ONLY on transition into LATE or FAILED
      (Event Contract).
- [ ] Implement retry policy: MAX_LIMIT 3 attempts, backoff 2s/4s/8s, 10s
      timeout per attempt (Retry/Timeout Policy).
- [ ] Implement `notification_log` insert (status sent/failed) regardless
      of outcome — state persistence MUST NOT depend on notification
      success (Failure Scenario 1).
- [ ] Build minimal frontend: Settings page for SMTP credential input
      (host, port, username, password, from address).
- [ ] Verify end-to-end: configure real SMTP (e.g. Gmail app password),
      trigger a monitor into FAILED state, confirm email arrives.
- [ ] Verify failure path: configure intentionally wrong SMTP credentials,
      confirm checker cycle does NOT crash and `notification_log` records
      `status=failed`.

**Exit criteria**: user receives a real email when a monitor fails. Full
core value loop (create monitor → integrate → silent failure → alert) is
functional.

---

## Phase 7 — Dashboard Polish (History & Status Visualization)

This phase is explicitly bounded — it is the ONLY phase where visual
polish beyond bare functionality is permitted, and ONLY for the two items
below. Anything beyond this list is deferred, no exceptions.

- [ ] Implement ping history table (paginated, per DESIGN.md Component
      System — no infinite scroll).
- [ ] Implement uptime percentage calculation (successful cycles / total
      expected cycles over a rolling window, e.g. last 30 days) and
      display on monitor detail page.
- [ ] Implement history graph using Recharts (lazy-loaded on monitor detail
      route only, per DESIGN.md Implementation Constraints) — timeline of
      state over time.
- [ ] Implement system-level staleness warning banner if checker's last
      tick timestamp exceeds 3x configured interval (per PRD.md Failure
      Scenario: Checker Service Downtime).
- [ ] Apply DESIGN.md skeleton loading states to dashboard list and
      monitor detail page (replace any generic spinners used during
      earlier phases).

**Exit criteria**: dashboard shows history graph, uptime %, and staleness
warning. STOP HERE. Do not add features not listed in this phase.

---

## Phase 8 — Deployment

- [ ] Write `Dockerfile`: multi-stage build (Node build stage for frontend
      → Go build stage embedding `web/dist` → minimal final image, e.g.
      `scratch` or `distroless` given no CGO dependency).
- [ ] Write `docker-compose.yml` example with volume mount for SQLite file
      persistence.
- [ ] Document required environment variables (DB path, session secret,
      AES encryption key, checker interval, port) in `README.md`.
- [ ] Verify: fresh `docker build` + `docker run` on a clean machine/VM
      completes signup → create monitor → receive ping → see state change,
      with zero manual intervention beyond `docker run`.
- [ ] Tag `v0.1.0` release. Write minimal `README.md` covering: what it is,
      quickstart (docker run), environment variables, and a note that this
      is self-hosted-only (no hosted version) for MVP.
- [ ] Publish repository as public/open-source (per PRD.md value
      proposition).

**Exit criteria**: a stranger can clone the repo, read the README, run one
`docker run` command, and have a working instance. This is the definition
of "done" for MVP — not "feature complete," but "deployable and usable by
someone who is not you."

---

## Explicit Non-Sequencing Notes

- Do NOT start Phase 4 (state machine) before Phase 2 and 3 are verified
  working — the state machine has nothing to operate on without monitors
  and pings existing first.
- Do NOT attempt Phase 7 (polish) before Phase 6 (notifications) is fully
  working — a pretty dashboard with no alerting is not the core value
  proposition of this product and is a scope-discipline trap.
- If time pressure forces a cut, the correct cut is Phase 7 polish items
  (history graph, uptime %) — NOT Phase 6 (notifications) and NOT Phase 4
  (state machine correctness). Those two phases ARE the product.
