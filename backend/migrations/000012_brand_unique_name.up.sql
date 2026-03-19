-- Add unique constraint on brand name to prevent duplicates
CREATE UNIQUE INDEX IF NOT EXISTS idx_brands_name_unique ON brands (name);
