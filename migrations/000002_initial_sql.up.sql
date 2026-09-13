CREATE TABLE
    IF NOT EXISTS users (
        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        email TEXT UNIQUE,
        password_hash TEXT,
        display_name TEXT,
        email_verified_at TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        deleted_at TIMESTAMPTZ
    );

CREATE TABLE
    IF NOT EXISTS user_sessions (
        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        refresh_token_hash TEXT NOT NULL,
        expires_at TIMESTAMPTZ NOT NULL,
        revoked_at TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        last_used_at TIMESTAMPTZ
    );

CREATE INDEX idx_user_sessions_user ON user_sessions (user_id);

CREATE INDEX idx_user_sessions_expires ON user_sessions (expires_at);

CREATE TABLE
    IF NOT EXISTS user_devices (
        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        device_name TEXT,
        platform TEXT NOT NULL,
        app_version TEXT,
        push_token TEXT,
        last_seen_at TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT now ()
    );

CREATE INDEX idx_user_devices_user ON user_devices (user_id);