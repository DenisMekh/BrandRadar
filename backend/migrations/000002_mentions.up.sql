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

    -- ML-предсказания (обязательно по ТЗ)
    ml_label VARCHAR(50) NOT NULL DEFAULT 'unknown',
    ml_score REAL NOT NULL DEFAULT 0.0,
    ml_is_relevant BOOLEAN NOT NULL DEFAULT false,
    ml_similar_ids UUID[] NOT NULL DEFAULT '{}',

    -- Статус обработки (state machine)
    status VARCHAR(20) NOT NULL DEFAULT 'new',
    deduplicated BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Дедупликация: уникальность в рамках бренда + источника + external_id
    CONSTRAINT uq_mention_dedup UNIQUE (brand_id, source_id, external_id),
    CONSTRAINT chk_mention_status CHECK (status IN ('new', 'processing', 'processed', 'archived', 'discarded')),
    CONSTRAINT chk_ml_score CHECK (ml_score >= 0.0 AND ml_score <= 1.0)
);

-- Индексы для фильтрации ленты
CREATE INDEX IF NOT EXISTS idx_mentions_brand_id ON mentions (brand_id);
CREATE INDEX IF NOT EXISTS idx_mentions_status ON mentions (status);
CREATE INDEX IF NOT EXISTS idx_mentions_ml_label ON mentions (ml_label);
CREATE INDEX IF NOT EXISTS idx_mentions_ml_is_relevant ON mentions (ml_is_relevant);
CREATE INDEX IF NOT EXISTS idx_mentions_published_at ON mentions (published_at DESC);
CREATE INDEX IF NOT EXISTS idx_mentions_created_at ON mentions (created_at DESC);

-- Композитные индексы для типичных запросов ленты
CREATE INDEX IF NOT EXISTS idx_mentions_brand_status_published ON mentions (brand_id, status, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_mentions_brand_label ON mentions (brand_id, ml_label, published_at DESC);
