-- Independent platform-model monitoring. This intentionally does not reference
-- channel_monitors: channel checks validate configured endpoints, while these
-- rows validate the platform scheduler path for a supported model.
CREATE TABLE IF NOT EXISTS model_monitors (
    id BIGSERIAL PRIMARY KEY,
    platform VARCHAR(32) NOT NULL,
    model VARCHAR(200) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    interval_seconds INTEGER NOT NULL DEFAULT 300,
    last_checked_at TIMESTAMPTZ NULL,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT model_monitors_platform_check CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok')),
    CONSTRAINT model_monitors_interval_check CHECK (interval_seconds BETWEEN 15 AND 3600),
    CONSTRAINT model_monitors_platform_model_key UNIQUE (platform, model)
);

CREATE INDEX IF NOT EXISTS idx_model_monitors_enabled_last_checked
    ON model_monitors (enabled, last_checked_at);

CREATE TABLE IF NOT EXISTS model_monitor_histories (
    id BIGSERIAL PRIMARY KEY,
    monitor_id BIGINT NOT NULL REFERENCES model_monitors(id) ON DELETE CASCADE,
    status VARCHAR(16) NOT NULL,
    latency_ms INTEGER NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    message VARCHAR(500) NOT NULL DEFAULT '',
    checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT model_monitor_histories_status_check CHECK (status IN ('operational', 'degraded', 'failed', 'error'))
);

CREATE INDEX IF NOT EXISTS idx_model_monitor_histories_monitor_checked
    ON model_monitor_histories (monitor_id, checked_at DESC);
CREATE INDEX IF NOT EXISTS idx_model_monitor_histories_checked
    ON model_monitor_histories (checked_at);
