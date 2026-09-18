-- Keep the independent USA0 model monitor aligned with concrete gateway platforms.
ALTER TABLE model_monitors
    DROP CONSTRAINT IF EXISTS model_monitors_platform_check;

ALTER TABLE model_monitors
    ADD CONSTRAINT model_monitors_platform_check
    CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok',
                        'kimi', 'zhipu', 'deepseek', 'minimax', 'opencode_go'));
