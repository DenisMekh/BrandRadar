-- Remove threshold and aggression columns from alert_configs
-- These fields are no longer used in the simplified alert system

ALTER TABLE alert_configs DROP COLUMN IF EXISTS threshold;
ALTER TABLE alert_configs DROP COLUMN IF EXISTS aggression;

-- Remove the aggression check constraint if it exists
ALTER TABLE alert_configs DROP CONSTRAINT IF EXISTS alert_configs_aggression_check;
ALTER TABLE alert_configs DROP CONSTRAINT IF EXISTS chk_threshold_positive;
