ALTER TABLE crawler_items DROP COLUMN IF EXISTS source_id;

CREATE TABLE IF NOT EXISTS mentions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    brand_id UUID NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    source_id UUID NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    external_id VARCHAR(512) NOT NULL,
    title VARCHAR(1024) NOT NULL DEFAULT '',
    text TEXT NOT NULL,
    url VARCHAR(2048) NOT NULL DEFAULT '',
    author VARCHAR(255) NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ NOT NULL,
    ml_label VARCHAR(50) NOT NULL DEFAULT 'unknown',
    ml_score REAL NOT NULL DEFAULT 0.0,
    ml_is_relevant BOOLEAN NOT NULL DEFAULT false,
    ml_similar_ids UUID[] NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'new',
    deduplicated BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_mention_dedup UNIQUE (brand_id, source_id, external_id),
    CONSTRAINT chk_mention_status CHECK (status IN ('new', 'processing', 'processed', 'archived', 'discarded')),
    CONSTRAINT chk_ml_score CHECK (ml_score >= 0.0 AND ml_score <= 1.0)
);
