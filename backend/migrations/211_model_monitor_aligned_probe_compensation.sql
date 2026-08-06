-- Align model monitor probes to natural time slots and optionally retry failed
-- probes at each minute boundary until health is confirmed again.
ALTER TABLE model_monitor_groups
    ADD COLUMN IF NOT EXISTS failure_compensation_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS failure_compensation_pending BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS last_scheduled_slot_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS next_compensation_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS consecutive_probe_failures INTEGER NOT NULL DEFAULT 0;

UPDATE model_monitor_groups
SET last_scheduled_slot_at = to_timestamp(
        floor(extract(epoch FROM NOW()) / interval_seconds) * interval_seconds
    )
WHERE last_scheduled_slot_at IS NULL;

DROP INDEX IF EXISTS idx_model_monitor_groups_probe_due;
CREATE INDEX IF NOT EXISTS idx_model_monitor_groups_probe_due
    ON model_monitor_groups (
        enabled,
        failure_compensation_pending,
        next_compensation_at,
        probe_claimed_until,
        last_scheduled_slot_at
    );

CREATE INDEX IF NOT EXISTS idx_model_monitor_groups_compensation_due
    ON model_monitor_groups (next_compensation_at)
    WHERE enabled = TRUE
      AND failure_compensation_enabled = TRUE
      AND failure_compensation_pending = TRUE;
