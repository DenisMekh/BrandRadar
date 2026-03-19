ALTER TABLE alert_configs
ADD COLUMN aggression DOUBLE PRECISION NOT NULL DEFAULT 0.5;

ALTER TABLE alert_configs
ADD CONSTRAINT alert_configs_aggression_check
CHECK (aggression >= 0 AND aggression <= 1);
