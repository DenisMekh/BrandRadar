-- Add anomaly detection configuration columns to alert_configs
ALTER TABLE alert_configs 
ADD COLUMN percentile DOUBLE PRECISION NOT NULL DEFAULT 95.0;

ALTER TABLE alert_configs 
ADD COLUMN anomaly_window_size INTEGER NOT NULL DEFAULT 10;

-- Add check constraints
ALTER TABLE alert_configs 
ADD CONSTRAINT chk_percentile_range CHECK (percentile >= 0 AND percentile <= 100);

ALTER TABLE alert_configs 
ADD CONSTRAINT chk_anomaly_window_size_positive CHECK (anomaly_window_size >= 3);
