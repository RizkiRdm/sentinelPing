CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    smtp_host     TEXT,
    smtp_port     INTEGER,
    smtp_username TEXT,
    smtp_password TEXT,
    smtp_from     TEXT,
    created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS monitors (
    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id                 INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                    TEXT NOT NULL,
    ping_token              TEXT NOT NULL UNIQUE,
    expected_interval_seconds INTEGER NOT NULL,
    grace_period_seconds      INTEGER NOT NULL,
    max_runtime_seconds       INTEGER NOT NULL,
    current_state           TEXT NOT NULL DEFAULT 'PENDING' CHECK (current_state IN ('PENDING', 'HEALTHY', 'RUNNING', 'LATE', 'FAILED')),
    last_state_change_at    TEXT,
    created_at              TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    paused                  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS pings (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    monitor_id  INTEGER NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    ping_type   TEXT NOT NULL CHECK (ping_type IN ('start', 'success', 'fail')),
    received_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    source_ip   TEXT
);

CREATE TABLE IF NOT EXISTS state_transitions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    monitor_id      INTEGER NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    from_state      TEXT NOT NULL,
    to_state        TEXT NOT NULL,
    transitioned_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    reason          TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS notification_log (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    monitor_id   INTEGER NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    attempted_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    status       TEXT NOT NULL CHECK (status IN ('sent', 'failed')),
    error_detail TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_monitors_ping_token ON monitors(ping_token);
CREATE INDEX IF NOT EXISTS idx_pings_monitor_id ON pings(monitor_id);
CREATE INDEX IF NOT EXISTS idx_pings_monitor_received ON pings(monitor_id, received_at);
CREATE INDEX IF NOT EXISTS idx_state_transitions_monitor_id ON state_transitions(monitor_id);
