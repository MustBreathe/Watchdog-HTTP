CREATE TABLE monitors (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    method TEXT NOT NULL,
    interval_seconds INTEGER NOT NULL,
    timeout_seconds INTEGER NOT NULL,
    expected_status INTEGER NOT NULL,
    expected_body_substring TEXT,
    headers_json TEXT,
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE checks (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL,
    started_at TEXT NOT NULL,
    finished_at TEXT NOT NULL,
    duration_ms INTEGER NOT NULL,
    status_code INTEGER,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    response_body_sample TEXT,
    trigger TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
);

CREATE INDEX idx_checks_monitor_id_created_at ON checks (monitor_id, created_at DESC);

CREATE INDEX idx_checks_success ON checks (success);

CREATE TABLE incidents (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL,
    started_at TEXT NOT NULL,
    resolved_at TEXT,
    status TEXT NOT NULL,
    failure_count INTEGER NOT NULL DEFAULT 1,
    last_error_message TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
);