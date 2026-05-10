-- 0002_pgvector: enable pgvector. The extension is unused in v0 — the
-- embeddings table is created in a later phase — but we want it available
-- from day one because the docker-compose image already ships it and the
-- semantic-memory work in Phase 1 of the roadmap depends on it.

CREATE EXTENSION IF NOT EXISTS vector;
