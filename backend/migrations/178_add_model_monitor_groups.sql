-- Model monitoring follows business groups, not the ungrouped scheduler pool.
-- A missing configuration remains valid and means "all compatible groups by rate".
CREATE TABLE IF NOT EXISTS model_monitor_groups (
    monitor_id BIGINT NOT NULL REFERENCES model_monitors(id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    priority INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (monitor_id, group_id),
    CONSTRAINT model_monitor_groups_priority_check CHECK (priority >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_model_monitor_groups_priority
    ON model_monitor_groups (monitor_id, priority);

ALTER TABLE model_monitor_histories
    ADD COLUMN IF NOT EXISTS group_id BIGINT NULL REFERENCES groups(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS group_name VARCHAR(100) NOT NULL DEFAULT '';

