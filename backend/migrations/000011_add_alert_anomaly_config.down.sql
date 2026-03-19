-- Rollback: Remove anomaly detection configuration columns from alert_configs
ALTER TABLE alert_configs DROP CONSTRAINT IF EXISTS chk_percentile_range;
ALTER TABLE alert_configs DROP CONSTRAINT IF EXISTS chk_anomaly_window_size_positive;

ALTER TABLE alert_configs DROP COLUMN IF EXISTS percentile;
ALTER TABLE alert_configs DROP COLUMN IF EXISTS anomaly_window_size;
