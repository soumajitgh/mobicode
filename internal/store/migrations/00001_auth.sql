-- +goose Up
CREATE TABLE users (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 email TEXT NOT NULL UNIQUE,
 password_hash TEXT NOT NULL,
 session_version INTEGER NOT NULL DEFAULT 1,
 created_at DATETIME NOT NULL,
 updated_at DATETIME NOT NULL
);
CREATE TABLE sessions (
 token TEXT PRIMARY KEY,
 data BLOB NOT NULL,
 expiry DATETIME NOT NULL
);
CREATE INDEX sessions_expiry_idx ON sessions(expiry);
-- +goose Down
DROP TABLE sessions;
DROP TABLE users;
