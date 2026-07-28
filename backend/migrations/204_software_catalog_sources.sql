ALTER TABLE software_catalog_items
    ADD COLUMN IF NOT EXISTS source_type VARCHAR(20) NOT NULL DEFAULT 'github',
    ADD COLUMN IF NOT EXISTS source_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS manual_version VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS manual_release_name VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS manual_release_notes TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS manual_published_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS download_assets JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE software_catalog_items ALTER COLUMN repository DROP NOT NULL;
ALTER TABLE software_catalog_items ALTER COLUMN repository_url DROP NOT NULL;

UPDATE software_catalog_items
SET source_type = 'github', source_url = COALESCE(NULLIF(source_url, ''), repository_url, '')
WHERE source_type = 'github';

ALTER TABLE software_catalog_items DROP CONSTRAINT IF EXISTS software_catalog_repository_format;
ALTER TABLE software_catalog_items DROP CONSTRAINT IF EXISTS software_catalog_source_type;
ALTER TABLE software_catalog_items ADD CONSTRAINT software_catalog_source_type
    CHECK (source_type IN ('github', 'manual'));

CREATE UNIQUE INDEX IF NOT EXISTS idx_software_catalog_repository_unique
    ON software_catalog_items (LOWER(repository)) WHERE repository IS NOT NULL;
