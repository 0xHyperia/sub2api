CREATE TABLE IF NOT EXISTS software_catalog_items (
    id BIGSERIAL PRIMARY KEY,
    repository VARCHAR(255) NOT NULL UNIQUE,
    repository_url TEXT NOT NULL,
    name VARCHAR(120) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    logo_url TEXT NOT NULL DEFAULT '',
    featured BOOLEAN NOT NULL DEFAULT FALSE,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 100,
    supported_platforms JSONB NOT NULL DEFAULT '[]'::jsonb,
    release_snapshot JSONB,
    release_fetched_at TIMESTAMPTZ,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT software_catalog_repository_format CHECK (repository ~ '^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$')
);

CREATE INDEX IF NOT EXISTS idx_software_catalog_public_order
    ON software_catalog_items (enabled, featured DESC, sort_order, id);

INSERT INTO software_catalog_items
    (repository, repository_url, name, description, logo_url, featured, enabled, sort_order, supported_platforms)
VALUES
    ('USA-Zero/ZeroAgent', 'https://github.com/USA-Zero/ZeroAgent', 'ZeroAgent', '可扩展的 AI Agent 桌面客户端，覆盖对话、工具调用、WebUI 与跨平台工作流。', '/software-center/zeroagent.png', TRUE, TRUE, 10, '["windows","macos","linux","android"]'),
    ('farion1231/cc-switch', 'https://github.com/farion1231/cc-switch', 'CC Switch', '跨平台 AI 编程助手管理工具，集中切换 Claude Code、Codex、Gemini 等供应商配置。', '/software-center/cc-switch.png', FALSE, TRUE, 20, '["windows","macos","linux"]'),
    ('BigPizzaV3/CodexPlusPlus', 'https://github.com/BigPizzaV3/CodexPlusPlus', 'Codex++', '面向 Codex 桌面应用的外部启动与管理工具，提供供应商切换、会话管理和界面增强。', '/software-center/codex-plus-plus.png', FALSE, TRUE, 30, '["windows","macos"]')
ON CONFLICT (repository) DO NOTHING;
