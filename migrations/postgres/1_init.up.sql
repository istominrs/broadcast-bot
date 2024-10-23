CREATE DATABASE IF NOT EXISTS freekeys;

\c freekeys;

CREATE TABLE IF NOT EXISTS servers (
    uuid UUID NOT NULL PRIMARY KEY,
    ip_address TEXT NOT NULL UNIQUE,
    port INTEGER NOT NULL,
    key TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS access_keys (
    uuid UUID NOT NULL PRIMARY KEY,
    key TEXT NOT NULL,
    api_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expired_at TIMESTAMP NOT NULL
);
