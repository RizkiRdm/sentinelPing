# PRD.md

Project Name: SentinelPing (working name)
Category: Cron Job Monitoring / Dead Man's Switch
Version: 1.0 — MVP

---

# Project Summary

## Overview
SentinelPing is a self-hosted, open-source monitoring service for scheduled jobs
(cron jobs, backup scripts, scheduled syncs, batch reports). Jobs report their
lifecycle (start, success, fail) via HTTP ping. If a job does not report on
schedule, or reports failure, SentinelPing detects it and sends an email alert.

## Objective
Give solo developers and small teams visibility into silent failures of
background jobs, without requiring a paid SaaS subscription or a hosted
third-party dependency.

## Value Proposition
- Self-hosted: full data ownership, no per-monitor pricing.
- No monitor count limit (unlike free tiers of Healthchecks.io, Cronitor).
- Start+End ping model: distinguishes "job never ran" from "job ran but failed"
  — most free alternatives only support passive/end-only pings.
- Single binary deployment: one process, embedded frontend, SQLite storage.

---

# Target Users

## Persona: Solo Backend Developer
- Runs 1-20 scheduled jobs (backups, data sync, report generation) in
  production for personal projects or small business systems.
- Technical sophistication: HIGH. Comfortable editing crontab, adding a `curl`
  line to a script, reading HTTP status codes.
- Frustration: discovers job failures days/weeks late, usually via a
  downstream symptom (missing report, stale data, failed restore).

## Persona: Small Team Lead (2-10 engineers)
- Owns infrastructure for a small SaaS or internal tool with several
  scheduled maintenance tasks.
- Technical sophistication: HIGH.
- Frustration: existing SaaS monitoring tools (Healthchecks.io, Cronitor)
  become cost-prohibitive past free tier as monitor count grows; wants
  self-hosted alternative with no per-monitor billing.

---

# Problem Statement

Scheduled jobs (cron, systemd timers, CI scheduled pipelines) fail silently.
There is no built-in mechanism to detect:
1. A job that stopped running entirely (crontab removed, container down,
   host offline).
2. A job that ran but exited with a failure (script error, unhandled
   exception, non-zero exit code).

Existing free/cheap solutions either:
- Only support passive end-of-job pings (cannot distinguish "never ran" from
  "ran but failed" without extra client-side logic), or
- Cap free usage at a small number of monitors, pushing cost onto users who
  scale their infrastructure.

Impact is asymmetric: the cost of missing one failure (e.g., 2 weeks of
un-backed-up data) vastly exceeds the cost of running a lightweight monitor.

---

# Success Metrics

All metrics MUST be measurable at MVP completion.

- Ping ingestion endpoint MUST respond in <100ms p95 (excludes DB write
  flush latency under load testing).
- Checker service MUST detect a late/dead monitor within
  `grace_period + check_interval` (max drift: 60s, given check_interval
  default of 60s).
- Dashboard MUST load monitor list in <2s on a dataset of 100 monitors.
- Alert email MUST be dispatched within 60s of a monitor transitioning to
  `late` or `failed` state.
- A new user MUST be able to complete signup + create first monitor + receive
  the correct `curl` snippet in <60s (measured via manual walkthrough, no
  instrumentation required for MVP).
- System MUST support at least 500 concurrent monitors on a single instance
  without check cycle exceeding 5s total execution time (SQLite-bound
  assumption, single-node).

If a metric cannot be measured through code or manual walkthrough, it is
INVALID and MUST NOT be included in future iterations.

---

# Core Capabilities (MVP)

Capabilities are system-level. UI pages are NOT capabilities.

1. **Authentication** — user signup, login, session management (multi-user,
   no email verification, no password reset).
2. **Monitor Management** — create, read, update, delete monitor
   definitions (name, expected interval, grace period, max runtime).
3. **Ping Ingestion** — public HTTP endpoint accepting start/success/fail
   pings scoped to a monitor via unique token in URL path.
4. **State Machine Engine** — deterministic computation of monitor state
   (healthy, running, late, failed, dead) based on ping history and time.
5. **Checker Service** — background process that periodically evaluates all
   monitors against the state machine and triggers alerts on state
   transition into a failing state.
6. **Notification Dispatch** — email delivery of alerts via user-configured
   SMTP credentials.
7. **Ping History & Status Dashboard** — retrieval and visualization of
   monitor status, uptime percentage, and ping history timeline.

---

# Non-MVP Features

Explicitly excluded from MVP. Reference only — full detail in ARCHITECTURE.md
Non-Goals section.

- Email verification / password reset flow.
- Role-based access control (RBAC), teams, or shared workspaces.
- Non-email notification channels (Telegram, Discord, generic webhook,
  Slack).
- Public status pages.
- Monitor grouping / tagging / folders.
- API rate limiting beyond basic abuse protection.
- Metrics export (Prometheus, etc).
- Mobile app.
- Billing / subscription system.
- Multi-instance / distributed checker (horizontal scaling).

---

# User Flows

## Primary Flow: Add a Monitor and Integrate
1. User signs up / logs in.
2. User creates a monitor: sets name, expected interval, grace period, max
   runtime.
3. System generates a unique ping token and displays ready-to-copy `curl`
   snippets for start/success/fail.
4. User adds snippet to their cron job / script.
5. Job runs; pings arrive; dashboard reflects `healthy` state after first
   successful cycle.

## Primary Flow: Detect and Alert on Failure
1. Monitor is in `healthy` or `running` state.
2. Expected ping does not arrive within `grace_period`, OR a `fail` ping is
   received, OR `running` state exceeds `max_runtime`.
3. Checker service transitions monitor state on next check cycle.
4. Notification dispatch sends email alert to user's registered address.
5. Dashboard reflects new state and timestamp of transition.

## Edge Case: Duplicate/Out-of-Order Pings
- A `success` ping arriving without a preceding `start` ping in the current
  cycle MUST still be accepted and MUST transition state to `healthy` (job
  may not always call the start endpoint; success/fail pings are
  self-sufficient signals).
- A `start` ping arriving while monitor is already `running` MUST reset the
  `running` cycle timer (treated as a new run beginning).

## Failure Scenario: Checker Service Downtime
- If the checker process itself is down, monitors MUST NOT falsely show
  `healthy` — the dashboard MUST display a distinct system-level warning
  ("last check performed at: X") if the most recent check cycle timestamp is
  older than 3x the global check interval.

## Failure Scenario: SMTP Delivery Failure
- Failed email delivery MUST be logged (structured log) and MUST NOT block
  or crash the checker cycle. Retry policy: MAX_LIMIT 3 attempts per alert
  event, no further retry after exhaustion for that event.

---

# High-Level Tech Stack

- **Frontend**: React 19 + Vite + TailwindCSS + shadcn/ui + Recharts
  (history graph). Built as static assets, embedded into Go binary.
- **Backend**: Go (standard library + minimal router, e.g. chi).
- **Database**: SQLite via `modernc.org/sqlite` (no CGO).
- **Background Scheduler**: in-process Go goroutine + ticker (no external
  job queue, no Redis).
- **Email Delivery**: standard `net/smtp` or minimal SMTP client library
  against user-supplied SMTP credentials.
- **Deployment**: single compiled binary; optional Dockerfile wrapping the
  same binary.

---

# Technical Assumptions

⚠️ Assumptions — validated only by architecture design, not by production
load testing at MVP stage.

- Single-node deployment is sufficient for target scale (solo dev / small
  team, up to ~500 monitors per instance).
- SQLite is sufficient for expected write volume (ping ingestion is
  low-frequency per monitor — typically 1 ping per job run, jobs run at
  minute-to-daily intervals, not sub-second).
- Users possess their own SMTP credentials (Gmail app password, SendGrid,
  Mailgun, etc) or are willing to configure one; SentinelPing does not
  provide a managed email sending service.
- No horizontal scaling requirement at MVP. If monitor count or check
  frequency demands exceed single-node capacity, this is an explicit
  future-phase architectural revision, not a Phase 1 concern.
