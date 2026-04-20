-- 002_add_sessions.sql

CREATE TABLE IF NOT EXISTS sessions (
    id               TEXT PRIMARY KEY,
    user_id          INTEGER,
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_activity_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    current_vector   BLOB, -- Serialized float32 vector
    origin_vector    BLOB, -- Serialized float32 vector
    exploration_rate REAL DEFAULT 0.15
);

CREATE TABLE IF NOT EXISTS session_events (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    track_id   INTEGER NOT NULL,
    event_type TEXT NOT NULL, -- 'play', 'skip', 'complete'
    completion REAL,
    played_for INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(session_id) REFERENCES sessions(id) ON DELETE CASCADE
);
