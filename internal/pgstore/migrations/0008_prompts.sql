-- 0008_prompts: project-scoped prompt templates with declared arguments.
-- args_jsonb stores the MCP-shape PromptArg array.

CREATE TABLE IF NOT EXISTS prompts (
    id              UUID         PRIMARY KEY,
    project_id      UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT         NOT NULL,
    description     TEXT         NOT NULL DEFAULT '',
    body            TEXT         NOT NULL DEFAULT '',
    args_jsonb      JSONB        NOT NULL DEFAULT '[]'::jsonb,
    request_id      UUID,
    correlation_id  UUID,
    causation_id    UUID,
    idempotency_key TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS prompts_unique_name_per_project
    ON prompts (project_id, name);

CREATE UNIQUE INDEX IF NOT EXISTS prompts_idempotency
    ON prompts (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
