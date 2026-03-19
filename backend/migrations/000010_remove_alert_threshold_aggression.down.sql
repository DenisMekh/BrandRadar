-- Rollback: Add back threshold and aggression columns to alert_configs

ALTER TABLE alert_configs ADD COLUMN threshold INTEGER NOT NULL DEFAULT 10;
ALTER TABLE alert_configs ADD COLUMN aggression DOUBLE PRECISION NOT NULL DEFAULT 0.5;

-- Recreate constraints
ALTER TABLE alert_configs ADD CONSTRAINT chk_threshold_positive CHECK (threshold > 0);
ALTER TABLE alert_configs ADD CONSTRAINT alert_configs_aggression_check CHECK (aggression >= 0 AND aggression <= 1);
