-- Добавляем source_id в crawler_items для сборки Mention
ALTER TABLE crawler_items
    ADD COLUMN IF NOT EXISTS source_id UUID REFERENCES sources(id) ON DELETE SET NULL;

-- Удаляем старую таблицу mentions (данные теперь собираются из crawler_items + sentiment_results)
DROP TABLE IF EXISTS mentions CASCADE;
