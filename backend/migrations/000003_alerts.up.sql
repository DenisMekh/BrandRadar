-- Конфигурации spike-алертов
CREATE TABLE IF NOT EXISTS alert_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    brand_id UUID NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    threshold INTEGER NOT NULL,
    window_minutes INTEGER NOT NULL,
    cooldown_minutes INTEGER NOT NULL,
    sentiment_filter VARCHAR(50) NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_threshold_positive CHECK (threshold > 0),
    CONSTRAINT chk_window_positive CHECK (window_minutes > 0),
    CONSTRAINT chk_cooldown_positive CHECK (cooldown_minutes > 0)
);

CREATE INDEX IF NOT EXISTS idx_alert_configs_brand_id ON alert_configs (brand_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_alert_configs_brand_unique ON alert_configs (brand_id);

-- Сработавшие алерты
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_id UUID NOT NULL REFERENCES alert_configs(id) ON DELETE CASCADE,
    brand_id UUID NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    mentions_count INTEGER NOT NULL,
    window_start TIMESTAMPTZ NOT NULL,
    window_end TIMESTAMPTZ NOT NULL,
    fired_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_alerts_brand_id ON alerts (brand_id);
CREATE INDEX IF NOT EXISTS idx_alerts_fired_at ON alerts (fired_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_config_id ON alerts (config_id);
