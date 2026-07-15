-- Optional marketplace presentation metadata managed alongside model monitors.
ALTER TABLE model_monitors
    ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS label VARCHAR(24) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_model_monitors_display_order
    ON model_monitors (display_order DESC, platform, model);
