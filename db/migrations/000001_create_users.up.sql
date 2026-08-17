CREATE TABLE users (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
