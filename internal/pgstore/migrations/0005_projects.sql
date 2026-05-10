-- 0005_projects: the leaf of the namespace hierarchy — the unit that maps
-- to a code base. namespace_id may be NULL for standalone projects.
-- (namespace_id, name) is unique; NULLS NOT DISTINCT so two standalone
-- projects can't share a name either.

CREATE TABLE IF NOT EXISTS projects (
    id              UUID         PRIMARY KEY,
    namespace_id    UUID         REFERENCES namespaces(id) ON DELETE CASCADE,
    name            TEXT         NOT NULL,
    description     TEXT         NOT NULL DEFAULT '',
    settings_jsonb  JSONB        NOT NULL DEFAULT '{}'::jsonb,
    request_id      UUID,
    correlation_id  UUID,
    causation_id    UUID,
    idempotency_key TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS projects_unique_name_per_namespace
    ON projects (namespace_id, name) NULLS NOT DISTINCT;

CREATE UNIQUE INDEX IF NOT EXISTS projects_idempotency
    ON projects (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
