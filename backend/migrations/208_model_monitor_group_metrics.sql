-- Group-scoped model monitoring with sparse minute/hour metric buckets.
ALTER TABLE model_monitor_groups
    ADD COLUMN IF NOT EXISTS enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS interval_seconds INTEGER NOT NULL DEFAULT 300,
    ADD COLUMN IF NOT EXISTS last_traffic_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS last_probe_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS probe_claimed_until TIMESTAMPTZ NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'model_monitor_groups_interval_check'
          AND conrelid = 'model_monitor_groups'::regclass
    ) THEN
        ALTER TABLE model_monitor_groups
            ADD CONSTRAINT model_monitor_groups_interval_check
            CHECK (interval_seconds BETWEEN 60 AND 3600) NOT VALID;
    END IF;
END $$;

UPDATE model_monitor_groups mg
SET enabled = m.enabled,
    interval_seconds = GREATEST(60, m.interval_seconds)
FROM model_monitors m
WHERE m.id = mg.monitor_id;

ALTER TABLE model_monitor_groups
    VALIDATE CONSTRAINT model_monitor_groups_interval_check;

CREATE INDEX IF NOT EXISTS idx_model_monitor_groups_probe_due
    ON model_monitor_groups (enabled, probe_claimed_until, last_traffic_at, last_probe_at);

ALTER TABLE model_monitor_histories
    ADD COLUMN IF NOT EXISTS first_token_ms INTEGER NULL,
    ADD COLUMN IF NOT EXISTS input_tokens INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS output_tokens INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS generation_ms BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS probe_cost DECIMAL(20, 10) NULL;

CREATE TABLE IF NOT EXISTS model_monitor_metric_buckets (
    monitor_id BIGINT NOT NULL REFERENCES model_monitors(id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    resolution VARCHAR(8) NOT NULL,
    source VARCHAR(8) NOT NULL,
    bucket_start TIMESTAMPTZ NOT NULL,
    request_count BIGINT NOT NULL DEFAULT 0,
    success_count BIGINT NOT NULL DEFAULT 0,
    latency_sum_ms BIGINT NOT NULL DEFAULT 0,
    latency_count BIGINT NOT NULL DEFAULT 0,
    ttft_sum_ms BIGINT NOT NULL DEFAULT 0,
    ttft_count BIGINT NOT NULL DEFAULT 0,
    output_tokens BIGINT NOT NULL DEFAULT 0,
    generation_ms BIGINT NOT NULL DEFAULT 0,
    probe_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    probe_cost_known BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (monitor_id, group_id, resolution, source, bucket_start),
    CONSTRAINT model_monitor_metric_resolution_check CHECK (resolution IN ('minute', 'hour')),
    CONSTRAINT model_monitor_metric_source_check CHECK (source IN ('traffic', 'probe'))
);

CREATE INDEX IF NOT EXISTS idx_model_monitor_metric_bucket_time
    ON model_monitor_metric_buckets (resolution, bucket_start);

CREATE INDEX IF NOT EXISTS idx_model_monitor_metric_lookup
    ON model_monitor_metric_buckets (monitor_id, group_id, resolution, bucket_start DESC);
