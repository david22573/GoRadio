-- 001_initial_schema.sql

CREATE TABLE IF NOT EXISTS stations (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    url  TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS shows (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL UNIQUE,
    duration   INTEGER NOT NULL,
    day        INTEGER NOT NULL,
    hour       INTEGER NOT NULL,
    min        INTEGER NOT NULL,
    scheduled  INTEGER NOT NULL DEFAULT 0,
    station_id INTEGER NOT NULL,
    FOREIGN KEY(station_id) REFERENCES stations(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tracks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    station_id  INTEGER NOT NULL,
    title       TEXT NOT NULL,
    artist      TEXT NOT NULL,
    url         TEXT NOT NULL,
    duration    INTEGER NOT NULL,
    analyzed_at DATETIME,
    FOREIGN KEY(station_id) REFERENCES stations(id) ON DELETE CASCADE
);

-- Vector Support (sqlite-vec extension must be loaded)
CREATE VIRTUAL TABLE IF NOT EXISTS track_vectors USING vec0(
    track_id INTEGER PRIMARY KEY,
    embedding FLOAT[128] distance_metric=l2
);
