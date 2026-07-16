CREATE TABLE IF NOT EXISTS ticket_categories (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name_zh VARCHAR(100) NOT NULL,
    name_en VARCHAR(100) NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tickets (
    id BIGSERIAL PRIMARY KEY,
    number VARCHAR(32) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    category_id BIGINT NOT NULL REFERENCES ticket_categories(id) ON DELETE RESTRICT,
    subject VARCHAR(200) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'answered', 'closed')),
    user_unread_count INT NOT NULL DEFAULT 0 CHECK (user_unread_count >= 0),
    admin_unread_count INT NOT NULL DEFAULT 0 CHECK (admin_unread_count >= 0),
    last_actor_type VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (last_actor_type IN ('user', 'admin', 'system')),
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    closed_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    closed_by_role VARCHAR(20),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ticket_messages (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL REFERENCES tickets(id) ON DELETE RESTRICT,
    sender_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    sender_type VARCHAR(20) NOT NULL CHECK (sender_type IN ('user', 'admin', 'system')),
    event_type VARCHAR(32) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ticket_attachments (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL REFERENCES ticket_messages(id) ON DELETE RESTRICT,
    object_key VARCHAR(1024) NOT NULL UNIQUE,
    original_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    sha256 VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ticket_categories_active_sort ON ticket_categories(active, sort_order, id);
CREATE INDEX IF NOT EXISTS idx_tickets_user_activity ON tickets(user_id, last_message_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_tickets_status_activity ON tickets(status, last_message_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_tickets_category_activity ON tickets(category_id, last_message_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_tickets_admin_unread ON tickets(last_message_at DESC, id DESC) WHERE admin_unread_count > 0;
CREATE INDEX IF NOT EXISTS idx_ticket_messages_ticket_created ON ticket_messages(ticket_id, created_at, id);
CREATE INDEX IF NOT EXISTS idx_ticket_attachments_message ON ticket_attachments(message_id);

INSERT INTO ticket_categories (code, name_zh, name_en, active, sort_order)
VALUES
    ('account', '账户问题', 'Account', TRUE, 10),
    ('billing', '充值/账单', 'Billing', TRUE, 20),
    ('incident', '错误/故障', 'Errors and incidents', TRUE, 30),
    ('feature', '功能建议', 'Feature request', TRUE, 40),
    ('performance', '延迟/性能', 'Latency and performance', TRUE, 50),
    ('merchant', '商家问题', 'Merchant', TRUE, 60),
    ('other', '其他', 'Other', TRUE, 70),
    ('quality', '质量问题', 'Quality', TRUE, 80)
ON CONFLICT (code) DO NOTHING;
