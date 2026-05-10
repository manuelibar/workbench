-- 0010_modes: mutable mode containers nested inside a specific blueprint
-- version. Updates are only allowed when the blueprint is the latest
-- version for (project, blueprint_name); the workbench enforces this in Go,
-- not in SQL.

CREATE TABLE IF NOT EXISTS modes (
    id                 UUID         PRIMARY KEY,
    blueprint_id       UUID         NOT NULL REFERENCES blueprints(id) ON DELETE CASCADE,
    name               TEXT         NOT NULL,
    system_prompt      TEXT         NOT NULL DEFAULT '',
    capabilities_jsonb JSONB        NOT NULL DEFAULT '{}'::jsonb,
    definition_jsonb   JSONB        NOT NULL DEFAULT '{}'::jsonb,
    request_id         UUID,
    correlation_id     UUID,
    causation_id       UUID,
    idempotency_key    TEXT,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (blueprint_id, name)
);

CREATE UNIQUE INDEX IF NOT EXISTS modes_idempotency
    ON modes (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
