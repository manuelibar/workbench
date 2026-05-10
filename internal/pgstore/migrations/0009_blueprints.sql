-- 0009_blueprints: versioned, immutable composition roots. (project, name,
-- version) is unique; updates write a new row with `version = MAX + 1`.

CREATE TABLE IF NOT EXISTS blueprints (
    id              UUID         PRIMARY KEY,
    project_id      UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT         NOT NULL CHECK (length(name) > 0),
    version         INTEGER      NOT NULL CHECK (version > 0),
    definition_jsonb JSONB       NOT NULL DEFAULT '{}'::jsonb,
    request_id      UUID,
    correlation_id  UUID,
    causation_id    UUID,
    idempotency_key TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (project_id, name, version)
);

CREATE INDEX IF NOT EXISTS blueprints_project_name_version_desc
    ON blueprints (project_id, name, version DESC);

CREATE UNIQUE INDEX IF NOT EXISTS blueprints_idempotency
    ON blueprints (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
