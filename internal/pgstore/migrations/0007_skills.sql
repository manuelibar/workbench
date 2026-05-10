-- 0007_skills: project-scoped markdown skill documents.

CREATE TABLE IF NOT EXISTS skills (
    id              UUID         PRIMARY KEY,
    project_id      UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT         NOT NULL,
    body_md         TEXT         NOT NULL DEFAULT '',
    request_id      UUID,
    correlation_id  UUID,
    causation_id    UUID,
    idempotency_key TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS skills_unique_name_per_project
    ON skills (project_id, name);

CREATE UNIQUE INDEX IF NOT EXISTS skills_idempotency
    ON skills (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
