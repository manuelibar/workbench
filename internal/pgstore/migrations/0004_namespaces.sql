-- 0004_namespaces: tree-shaped organisational containers. Roots have NULL
-- parent_id. (parent_id, name) is unique among siblings; NULLS NOT DISTINCT
-- (Postgres 15+) ensures roots can't reuse a name either.

CREATE TABLE IF NOT EXISTS namespaces (
    id              UUID         PRIMARY KEY,
    parent_id       UUID         REFERENCES namespaces(id) ON DELETE CASCADE,
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

CREATE UNIQUE INDEX IF NOT EXISTS namespaces_unique_name_per_parent
    ON namespaces (parent_id, name) NULLS NOT DISTINCT;

CREATE UNIQUE INDEX IF NOT EXISTS namespaces_idempotency
    ON namespaces (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
