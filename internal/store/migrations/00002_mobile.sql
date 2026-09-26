-- +goose Up
CREATE TABLE device_pairings (
 id TEXT PRIMARY KEY,
 user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 token_hash TEXT NOT NULL UNIQUE,
 created_at DATETIME NOT NULL,
 expires_at DATETIME NOT NULL,
 consumed_at DATETIME
);
CREATE INDEX device_pairings_expires_at_idx ON device_pairings(expires_at);
CREATE TABLE mobile_devices (
 id TEXT PRIMARY KEY,
 user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 name TEXT NOT NULL,
 platform TEXT NOT NULL CHECK(platform IN ('IOS', 'ANDROID')),
 created_at DATETIME NOT NULL,
 last_seen_at DATETIME
);
CREATE TABLE mobile_sessions (
 id TEXT PRIMARY KEY,
 device_id TEXT NOT NULL REFERENCES mobile_devices(id) ON DELETE CASCADE,
 token_hash TEXT NOT NULL UNIQUE,
 created_at DATETIME NOT NULL,
 revoked_at DATETIME
);
CREATE INDEX mobile_sessions_device_id_idx ON mobile_sessions(device_id);
-- +goose Down
DROP TABLE mobile_sessions;
DROP TABLE mobile_devices;
DROP TABLE device_pairings;
