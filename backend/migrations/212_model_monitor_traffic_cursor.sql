-- Keep the passive traffic scan position across restarts. The scanner still
-- overlaps its cursor by a small safety window so asynchronously written logs
-- are re-read without losing data after a process outage.
CREATE TABLE IF NOT EXISTS model_monitor_runtime_state (
    id BOOLEAN PRIMARY KEY DEFAULT TRUE,
    traffic_cursor_at TIMESTAMPTZ NULL,
    traffic_claimed_until TIMESTAMPTZ NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT model_monitor_runtime_state_singleton CHECK (id)
);

INSERT INTO model_monitor_runtime_state (id)
VALUES (TRUE)
ON CONFLICT (id) DO NOTHING;

ALTER TABLE model_monitor_runtime_state
    ADD COLUMN IF NOT EXISTS traffic_claimed_until TIMESTAMPTZ NULL;

-- Scheduled probes use the natural slot as an idempotency key. This prevents
-- duplicate history/cost rows when a transaction commits but its result cannot
-- be observed by the worker due to a connection interruption.
ALTER TABLE model_monitor_histories
    ADD COLUMN IF NOT EXISTS scheduled_slot_at TIMESTAMPTZ NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_model_monitor_histories_scheduled_slot
    ON model_monitor_histories (monitor_id, group_id, scheduled_slot_at)
    WHERE scheduled_slot_at IS NOT NULL AND group_id IS NOT NULL;
