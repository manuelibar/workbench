-- Workbench durable state foundation.
-- This schema intentionally stays relational: namespace_edges provides tree/DAG
-- flexibility without requiring a graph database.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS namespaces (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug text NOT NULL UNIQUE,
    display_name text NOT NULL,
    description text NOT NULL DEFAULT '',
    kind text NOT NULL DEFAULT 'namespace',
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS namespace_edges (
    parent_id uuid NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    child_id uuid NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    edge_type text NOT NULL DEFAULT 'contains',
    sort_order integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (parent_id, child_id, edge_type),
    CONSTRAINT namespace_edges_no_self_edge CHECK (parent_id <> child_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS namespace_edges_one_contains_parent_per_child
    ON namespace_edges (child_id)
    WHERE edge_type = 'contains';

CREATE TABLE IF NOT EXISTS projects (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    namespace_id uuid REFERENCES namespaces(id) ON DELETE SET NULL,
    slug text NOT NULL,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    system_prompt text NOT NULL DEFAULT '',
    root_path text NOT NULL DEFAULT '',
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (namespace_id, slug)
);

CREATE UNIQUE INDEX IF NOT EXISTS projects_global_slug_unique
    ON projects (slug)
    WHERE namespace_id IS NULL;

CREATE TABLE IF NOT EXISTS roles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    namespace_id uuid REFERENCES namespaces(id) ON DELETE SET NULL,
    slug text NOT NULL,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    system_prompt text NOT NULL DEFAULT '',
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (namespace_id, slug)
);

CREATE UNIQUE INDEX IF NOT EXISTS roles_global_slug_unique
    ON roles (slug)
    WHERE namespace_id IS NULL;

CREATE TABLE IF NOT EXISTS prompt_templates (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    namespace_id uuid REFERENCES namespaces(id) ON DELETE SET NULL,
    project_id uuid REFERENCES projects(id) ON DELETE CASCADE,
    slug text NOT NULL,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    template text NOT NULL,
    scope text NOT NULL DEFAULT 'global',
    status text NOT NULL DEFAULT 'active',
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (namespace_id, project_id, slug)
);

CREATE UNIQUE INDEX IF NOT EXISTS prompt_templates_global_slug_unique
    ON prompt_templates (slug)
    WHERE namespace_id IS NULL AND project_id IS NULL;

CREATE TABLE IF NOT EXISTS tasks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    namespace_id uuid REFERENCES namespaces(id) ON DELETE SET NULL,
    project_id uuid REFERENCES projects(id) ON DELETE CASCADE,
    title text NOT NULL,
    description text NOT NULL DEFAULT '',
    state text NOT NULL,
    kind text NOT NULL DEFAULT 'workbench_task',
    evidence jsonb NOT NULL DEFAULT '[]'::jsonb,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS knowledge_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    namespace_id uuid REFERENCES namespaces(id) ON DELETE SET NULL,
    project_id uuid REFERENCES projects(id) ON DELETE CASCADE,
    task_id uuid REFERENCES tasks(id) ON DELETE SET NULL,
    kind text NOT NULL,
    uri text NOT NULL DEFAULT '',
    summary text NOT NULL,
    details text NOT NULL DEFAULT '',
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    client_name text NOT NULL DEFAULT '',
    namespace_id uuid REFERENCES namespaces(id) ON DELETE SET NULL,
    project_id uuid REFERENCES projects(id) ON DELETE SET NULL,
    role_id uuid REFERENCES roles(id) ON DELETE SET NULL,
    board_id uuid,
    task_id uuid REFERENCES tasks(id) ON DELETE SET NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS projects_namespace_idx ON projects(namespace_id);
CREATE INDEX IF NOT EXISTS roles_namespace_idx ON roles(namespace_id);
CREATE INDEX IF NOT EXISTS prompt_templates_scope_idx ON prompt_templates(namespace_id, project_id, status);
CREATE INDEX IF NOT EXISTS tasks_scope_idx ON tasks(namespace_id, project_id, state);
CREATE INDEX IF NOT EXISTS knowledge_items_scope_idx ON knowledge_items(namespace_id, project_id, task_id);
CREATE INDEX IF NOT EXISTS sessions_scope_idx ON sessions(namespace_id, project_id, role_id, task_id);

INSERT INTO namespaces (slug, display_name, description, kind)
VALUES ('default', 'Default', 'Default Workbench namespace', 'namespace')
ON CONFLICT (slug) DO NOTHING;
