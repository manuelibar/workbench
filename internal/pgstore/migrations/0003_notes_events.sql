-- 0003_notes_events: the Zettelkasten note primitive plus the append-only
-- events log. Notes are user-scoped (namespace_id/project_id are recorded
-- but not enforced at this stage — the namespaces and projects tables
-- arrive in later phases). Events are scoped to a WorkSession and feed the
-- `recent_events` field in the refresh tool's response.

CREATE TABLE IF NOT EXISTS notes (
    id              UUID         PRIMARY KEY,
    user_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body_md         TEXT         NOT NULL,
    tags            TEXT[]       NOT NULL DEFAULT '{}',
    namespace_id    UUID,
    project_id      UUID,
    promoted_to     UUID,
    request_id      UUID,
    correlation_id  UUID,
    causation_id    UUID,
    idempotency_key TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS notes_user_created_at_desc
    ON notes (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS notes_tags_gin
    ON notes USING gin (tags);

-- Idempotency: at most one row per (user_id, idempotency_key) when key is set.
CREATE UNIQUE INDEX IF NOT EXISTS notes_idempotency
    ON notes (user_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS events (
    id              UUID         PRIMARY KEY,
    work_session_id UUID         NOT NULL REFERENCES work_sessions(id) ON DELETE CASCADE,
    mcp_session_id  TEXT,
    occurred_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    type            TEXT         NOT NULL,
    subject_kind    TEXT,
    subject_id      TEXT,
    payload_jsonb   JSONB        NOT NULL DEFAULT '{}'::jsonb,
    request_id      UUID,
    correlation_id  UUID,
    causation_id    UUID,
    idempotency_key TEXT
);

CREATE INDEX IF NOT EXISTS events_ws_occurred_at_desc
    ON events (work_session_id, occurred_at DESC);
