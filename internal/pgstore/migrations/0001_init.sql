-- 0001_init: bootstrap users and work_sessions tables.
-- Each table carries the four-part audit columns from the manifesto's
-- id-schema reference: request_id, correlation_id, causation_id,
-- idempotency_key, plus standard created_at / updated_at.

CREATE TABLE IF NOT EXISTS users (
    id              UUID         PRIMARY KEY,
    display_name    TEXT         NOT NULL,
    request_id      UUID,
    correlation_id  UUID,
    causation_id    UUID,
    idempotency_key TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS work_sessions (
    id              UUID         PRIMARY KEY,
    user_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT         NOT NULL,
    started_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    last_seen       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    closed_at       TIMESTAMPTZ,
    selection_jsonb JSONB        NOT NULL DEFAULT '{}'::jsonb,
    request_id      UUID,
    correlation_id  UUID,
    causation_id    UUID,
    idempotency_key TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Singleton invariant: at most one open WorkSession per user.
CREATE UNIQUE INDEX IF NOT EXISTS work_sessions_one_open_per_user
    ON work_sessions (user_id) WHERE closed_at IS NULL;
