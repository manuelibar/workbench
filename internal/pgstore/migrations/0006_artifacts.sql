-- 0006_artifacts: typed, versioned, project-scoped outputs.
-- artifacts holds the head row (LatestVersion); artifact_versions is
-- append-only, keyed by (artifact_id, version).

CREATE TABLE IF NOT EXISTS artifacts (
    id              UUID         PRIMARY KEY,
    project_id      UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type            TEXT         NOT NULL CHECK (length(type) > 0),
    status          TEXT         NOT NULL DEFAULT 'draft'
                                 CHECK (status IN ('draft', 'reviewing', 'signed_off', 'archived')),
    parents         UUID[]       NOT NULL DEFAULT '{}',
    latest_version  INTEGER      NOT NULL DEFAULT 0,
    request_id      UUID,
    correlation_id  UUID,
    causation_id    UUID,
    idempotency_key TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS artifacts_project_type
    ON artifacts (project_id, type);

CREATE INDEX IF NOT EXISTS artifacts_project_status
    ON artifacts (project_id, status);

CREATE UNIQUE INDEX IF NOT EXISTS artifacts_idempotency
    ON artifacts (idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS artifact_versions (
    artifact_id     UUID         NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    version         INTEGER      NOT NULL CHECK (version > 0),
    content_jsonb   JSONB        NOT NULL DEFAULT '{}'::jsonb,
    content_text    TEXT         NOT NULL DEFAULT '',
    request_id      UUID,
    correlation_id  UUID,
    causation_id    UUID,
    idempotency_key TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    PRIMARY KEY (artifact_id, version)
);
