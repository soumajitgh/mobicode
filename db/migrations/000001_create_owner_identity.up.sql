CREATE TABLE owner_identity (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    public_key TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE auth_replays (
    event_id TEXT PRIMARY KEY NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_auth_replays_expires_at ON auth_replays(expires_at);

CREATE TABLE setup_sessions (
    id TEXT PRIMARY KEY NOT NULL,
    browser_token_hash TEXT NOT NULL UNIQUE,
    pairing_token_hash TEXT NOT NULL UNIQUE,
    csrf_token_hash TEXT NOT NULL,
    candidate_public_key TEXT,
    state TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_setup_sessions_expires_at ON setup_sessions(expires_at);
