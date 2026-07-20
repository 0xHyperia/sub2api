\set ON_ERROR_STOP on

BEGIN;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM distribution_commission_sources s
        JOIN payment_orders o ON o.id = s.payment_order_id
        WHERE o.out_trade_no NOT LIKE 'DIST-DEMO-%'
    ) OR EXISTS (
        SELECT 1
        FROM distribution_customer_bindings b
        JOIN users u ON u.id = b.user_id
        WHERE u.email NOT LIKE 'demo.distribution.%@sub2api.local'
    ) THEN
        RAISE EXCEPTION 'Refusing to seed: non-demo distribution business data already exists';
    END IF;
END $$;

DELETE FROM distribution_withdrawals WHERE review_note = '[distribution-demo]';
DELETE FROM distribution_commission_entries
WHERE source_id IN (
    SELECT s.id
    FROM distribution_commission_sources s
    JOIN payment_orders o ON o.id = s.payment_order_id
    WHERE o.out_trade_no LIKE 'DIST-DEMO-%'
);
DELETE FROM distribution_commission_sources
WHERE payment_order_id IN (SELECT id FROM payment_orders WHERE out_trade_no LIKE 'DIST-DEMO-%');
DELETE FROM payment_orders WHERE out_trade_no LIKE 'DIST-DEMO-%';

CREATE TEMP TABLE distribution_demo_users (
    email TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    kind TEXT NOT NULL
) ON COMMIT DROP;

INSERT INTO distribution_demo_users (email, username, kind) VALUES
    ('demo.distribution.agent.active@sub2api.local', '演示下级代理-正常', 'agent'),
    ('demo.distribution.agent.suspended@sub2api.local', '演示下级代理-暂停', 'agent'),
    ('demo.distribution.agent.revoked@sub2api.local', '演示下级代理-撤销', 'agent'),
    ('demo.distribution.customer.01@sub2api.local', '演示客户-星河', 'customer'),
    ('demo.distribution.customer.02@sub2api.local', '演示客户-远山', 'customer'),
    ('demo.distribution.customer.03@sub2api.local', '演示客户-青禾', 'customer'),
    ('demo.distribution.customer.04@sub2api.local', '演示客户-长风', 'customer'),
    ('demo.distribution.customer.05@sub2api.local', '演示客户-云帆', 'customer'),
    ('demo.distribution.customer.06@sub2api.local', '演示客户-知夏', 'customer'),
    ('demo.distribution.customer.07@sub2api.local', '演示客户-林深', 'customer'),
    ('demo.distribution.customer.08@sub2api.local', '演示客户-南星', 'customer');

INSERT INTO users (email, password_hash, username, role, status, notes)
SELECT d.email, admin.password_hash, d.username, 'user', 'active', '[distribution-demo]'
FROM distribution_demo_users d
CROSS JOIN LATERAL (
    SELECT password_hash FROM users WHERE email = 'admin@sub2api.local' AND deleted_at IS NULL LIMIT 1
) admin
WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.email = d.email AND u.deleted_at IS NULL);

UPDATE users u
SET username = d.username, status = 'active', notes = '[distribution-demo]', updated_at = NOW()
FROM distribution_demo_users d
WHERE u.email = d.email AND u.deleted_at IS NULL;

UPDATE distribution_settings
SET enabled = TRUE,
    withdrawal_enabled = TRUE,
    withdrawal_fee_rate_bps = 100,
    updated_by = (SELECT id FROM users WHERE email = 'admin@sub2api.local' LIMIT 1),
    updated_at = NOW()
WHERE id = 1;

WITH root AS (
    SELECT a.id
    FROM distribution_agents a
    JOIN users u ON u.id = a.user_id
    JOIN distribution_agent_levels l ON l.id = a.level_id AND l.depth = 1
    WHERE u.email = 'admin@sub2api.local'
), level AS (
    SELECT id FROM distribution_agent_levels WHERE depth = 2
), demo_agents(email, promotion_code, status, rate_bps) AS (
    VALUES
        ('demo.distribution.agent.active@sub2api.local', 'DEMO_ACTIVE', 'active', 600),
        ('demo.distribution.agent.suspended@sub2api.local', 'DEMO_PAUSED', 'suspended', 550),
        ('demo.distribution.agent.revoked@sub2api.local', 'DEMO_REVOKED', 'revoked', 500)
)
INSERT INTO distribution_agents (user_id, level_id, parent_agent_id, promotion_code, rate_override_bps, status, granted_by)
SELECT u.id, level.id, root.id, d.promotion_code, d.rate_bps, d.status, admin.id
FROM demo_agents d
JOIN users u ON u.email = d.email AND u.deleted_at IS NULL
CROSS JOIN root
CROSS JOIN level
CROSS JOIN LATERAL (SELECT id FROM users WHERE email = 'admin@sub2api.local' LIMIT 1) admin
ON CONFLICT (user_id) DO UPDATE SET
    parent_agent_id = EXCLUDED.parent_agent_id,
    promotion_code = EXCLUDED.promotion_code,
    rate_override_bps = EXCLUDED.rate_override_bps,
    status = EXCLUDED.status,
    updated_at = NOW();

INSERT INTO distribution_wallets (agent_id)
SELECT a.id
FROM distribution_agents a
JOIN users u ON u.id = a.user_id
WHERE u.email = 'admin@sub2api.local' OR u.email LIKE 'demo.distribution.agent.%@sub2api.local'
ON CONFLICT (agent_id) DO NOTHING;

WITH bindings(customer_email, agent_email, promotion_code, days_ago) AS (
    VALUES
        ('demo.distribution.customer.01@sub2api.local', 'admin@sub2api.local', '2H99MBBW9WE6', 32),
        ('demo.distribution.customer.02@sub2api.local', 'admin@sub2api.local', '2H99MBBW9WE6', 20),
        ('demo.distribution.customer.03@sub2api.local', 'admin@sub2api.local', '2H99MBBW9WE6', 8),
        ('demo.distribution.customer.04@sub2api.local', 'demo.distribution.agent.active@sub2api.local', 'DEMO_ACTIVE', 18),
        ('demo.distribution.customer.05@sub2api.local', 'demo.distribution.agent.active@sub2api.local', 'DEMO_ACTIVE', 12),
        ('demo.distribution.customer.06@sub2api.local', 'demo.distribution.agent.active@sub2api.local', 'DEMO_ACTIVE', 4),
        ('demo.distribution.customer.07@sub2api.local', 'demo.distribution.agent.suspended@sub2api.local', 'DEMO_PAUSED', 28),
        ('demo.distribution.customer.08@sub2api.local', 'demo.distribution.agent.revoked@sub2api.local', 'DEMO_REVOKED', 40)
)
INSERT INTO distribution_customer_bindings (user_id, agent_id, promotion_code, bound_at)
SELECT customer.id, agent.id, b.promotion_code, NOW() - make_interval(days => b.days_ago)
FROM bindings b
JOIN users customer ON customer.email = b.customer_email AND customer.deleted_at IS NULL
JOIN users agent_user ON agent_user.email = b.agent_email AND agent_user.deleted_at IS NULL
JOIN distribution_agents agent ON agent.user_id = agent_user.id
ON CONFLICT (user_id) DO UPDATE SET
    agent_id = EXCLUDED.agent_id,
    promotion_code = EXCLUDED.promotion_code,
    bound_at = EXCLUDED.bound_at,
    updated_at = NOW();

CREATE TEMP TABLE distribution_demo_orders (
    demo_no TEXT PRIMARY KEY,
    customer_email TEXT NOT NULL,
    currency TEXT NOT NULL,
    actual_paid NUMERIC(20,8) NOT NULL,
    rate_to_cny NUMERIC(20,8) NOT NULL,
    base_cny NUMERIC(20,8) NOT NULL,
    entry_status TEXT NOT NULL,
    days_ago INTEGER NOT NULL
) ON COMMIT DROP;

INSERT INTO distribution_demo_orders VALUES
    ('DIST-DEMO-001', 'demo.distribution.customer.01@sub2api.local', 'CNY', 500, 1, 500, 'available', 26),
    ('DIST-DEMO-002', 'demo.distribution.customer.01@sub2api.local', 'CNY', 300, 1, 300, 'frozen', 2),
    ('DIST-DEMO-003', 'demo.distribution.customer.02@sub2api.local', 'USD', 100, 7, 700, 'available', 16),
    ('DIST-DEMO-004', 'demo.distribution.customer.03@sub2api.local', 'CNY', 1200, 1, 1200, 'paid', 7),
    ('DIST-DEMO-010', 'demo.distribution.customer.03@sub2api.local', 'CNY', 1000, 1, 1000, 'available', 5),
    ('DIST-DEMO-005', 'demo.distribution.customer.04@sub2api.local', 'CNY', 800, 1, 800, 'available', 14),
    ('DIST-DEMO-006', 'demo.distribution.customer.05@sub2api.local', 'CNY', 600, 1, 600, 'frozen', 1),
    ('DIST-DEMO-007', 'demo.distribution.customer.06@sub2api.local', 'CNY', 1000, 1, 1000, 'converted', 3),
    ('DIST-DEMO-008', 'demo.distribution.customer.07@sub2api.local', 'CNY', 400, 1, 400, 'paid', 22),
    ('DIST-DEMO-009', 'demo.distribution.customer.08@sub2api.local', 'CNY', 500, 1, 500, 'paid', 35);

INSERT INTO payment_orders (
    user_id, user_email, user_name, amount, pay_amount, payment_type, payment_trade_no,
    order_type, status, expires_at, paid_at, completed_at, out_trade_no,
    provider_key, provider_snapshot, created_at, updated_at
)
SELECT
    u.id, u.email, u.username, d.actual_paid, d.actual_paid, 'demo', d.demo_no || '-TRADE',
    'balance', 'COMPLETED', NOW() + INTERVAL '1 day', NOW() - make_interval(days => d.days_ago),
    NOW() - make_interval(days => d.days_ago), d.demo_no, 'demo',
    jsonb_build_object('currency', d.currency, 'demo', true),
    NOW() - make_interval(days => d.days_ago), NOW() - make_interval(days => d.days_ago)
FROM distribution_demo_orders d
JOIN users u ON u.email = d.customer_email AND u.deleted_at IS NULL;

INSERT INTO distribution_commission_sources (
    payment_order_id, customer_user_id, direct_agent_id, payment_type, payment_currency,
    actual_paid_amount, fx_rate_to_cny, commission_base_cny, status, provider_snapshot, paid_at, created_at, updated_at
)
SELECT
    o.id, customer.id, agent.id, 'cash', d.currency, d.actual_paid, d.rate_to_cny,
    d.base_cny, 'settled', jsonb_build_object('demo', true),
    NOW() - make_interval(days => d.days_ago), NOW() - make_interval(days => d.days_ago), NOW()
FROM distribution_demo_orders d
JOIN payment_orders o ON o.out_trade_no = d.demo_no
JOIN users customer ON customer.email = d.customer_email AND customer.deleted_at IS NULL
JOIN distribution_customer_bindings binding ON binding.user_id = customer.id
JOIN distribution_agents agent ON agent.id = binding.agent_id;

INSERT INTO distribution_commission_entries (
    source_id, beneficiary_agent_id, entry_type, rate_bps, original_amount_cny,
    status, available_at, created_at, updated_at
)
SELECT
    source.id, direct.id, 'direct', COALESCE(direct.rate_override_bps, level.default_rate_bps),
    ROUND(demo.base_cny * COALESCE(direct.rate_override_bps, level.default_rate_bps) / 10000, 8),
    demo.entry_status,
    CASE WHEN demo.entry_status = 'frozen' THEN NOW() + INTERVAL '5 days' ELSE NOW() - INTERVAL '1 day' END,
    source.paid_at, NOW()
FROM distribution_demo_orders demo
JOIN payment_orders orders ON orders.out_trade_no = demo.demo_no
JOIN distribution_commission_sources source ON source.payment_order_id = orders.id
JOIN distribution_agents direct ON direct.id = source.direct_agent_id
JOIN distribution_agent_levels level ON level.id = direct.level_id;

INSERT INTO distribution_commission_entries (
    source_id, beneficiary_agent_id, entry_type, rate_bps, original_amount_cny,
    status, available_at, created_at, updated_at
)
SELECT
    source.id, parent.id, 'team', parent_level.default_rate_bps - COALESCE(direct.rate_override_bps, direct_level.default_rate_bps),
    ROUND(demo.base_cny * (parent_level.default_rate_bps - COALESCE(direct.rate_override_bps, direct_level.default_rate_bps)) / 10000, 8),
    demo.entry_status,
    CASE WHEN demo.entry_status = 'frozen' THEN NOW() + INTERVAL '5 days' ELSE NOW() - INTERVAL '1 day' END,
    source.paid_at, NOW()
FROM distribution_demo_orders demo
JOIN payment_orders orders ON orders.out_trade_no = demo.demo_no
JOIN distribution_commission_sources source ON source.payment_order_id = orders.id
JOIN distribution_agents direct ON direct.id = source.direct_agent_id
JOIN distribution_agent_levels direct_level ON direct_level.id = direct.level_id AND direct_level.depth = 2
JOIN distribution_agents parent ON parent.id = direct.parent_agent_id
JOIN distribution_agent_levels parent_level ON parent_level.id = parent.level_id
WHERE parent_level.default_rate_bps > COALESCE(direct.rate_override_bps, direct_level.default_rate_bps);

INSERT INTO distribution_payout_accounts (agent_id, alipay_name, alipay_account)
SELECT a.id, '演示收款人', 'demo.alipay@sub2api.local'
FROM distribution_agents a
JOIN users u ON u.id = a.user_id
WHERE u.email = 'admin@sub2api.local'
ON CONFLICT (agent_id) DO NOTHING;

WITH root AS (
    SELECT a.id
    FROM distribution_agents a
    JOIN users u ON u.id = a.user_id
    WHERE u.email = 'admin@sub2api.local'
)
INSERT INTO distribution_withdrawals (
    agent_id, amount_cny, fee_cny, payout_cny, alipay_name, alipay_account,
    status, review_note, reviewed_by, reviewed_at, paid_at, payment_reference, created_at, updated_at
)
SELECT root.id, v.amount, v.fee, v.payout, '演示收款人', 'demo.alipay@sub2api.local',
       v.status, '[distribution-demo]', admin.id, v.reviewed_at, v.paid_at,
       v.reference, v.created_at, NOW()
FROM root
CROSS JOIN LATERAL (SELECT id FROM users WHERE email = 'admin@sub2api.local' LIMIT 1) admin
CROSS JOIN (VALUES
    (100::numeric, 0::numeric, 100::numeric, 'pending', NULL::timestamptz, NULL::timestamptz, NULL::text, NOW() - INTERVAL '3 hours'),
    (161::numeric, 0::numeric, 161::numeric, 'paid', NOW() - INTERVAL '6 days', NOW() - INTERVAL '5 days', 'DIST-DEMO-WD-PAID', NOW() - INTERVAL '7 days'),
    (100::numeric, 0::numeric, 100::numeric, 'rejected', NOW() - INTERVAL '10 days', NULL::timestamptz, 'DIST-DEMO-WD-REJECTED', NOW() - INTERVAL '11 days'),
    (150::numeric, 0::numeric, 150::numeric, 'failed', NOW() - INTERVAL '14 days', NULL::timestamptz, 'DIST-DEMO-WD-FAILED', NOW() - INTERVAL '15 days')
) AS v(amount, fee, payout, status, reviewed_at, paid_at, reference, created_at);

WITH stats AS (
    SELECT
        a.id AS agent_id,
        COALESCE(SUM(e.original_amount_cny - e.reversed_amount_cny) FILTER (WHERE e.status = 'frozen'), 0) AS frozen,
        COALESCE(SUM(e.original_amount_cny - e.reversed_amount_cny) FILTER (WHERE e.status = 'available'), 0) AS available_before_reserve,
        COALESCE(SUM(e.original_amount_cny - e.reversed_amount_cny), 0) AS earned,
        COALESCE(SUM(e.original_amount_cny - e.reversed_amount_cny) FILTER (WHERE e.status = 'paid'), 0) AS withdrawn,
        COALESCE(SUM(e.original_amount_cny - e.reversed_amount_cny) FILTER (WHERE e.status = 'converted'), 0) AS converted
    FROM distribution_agents a
    LEFT JOIN distribution_commission_entries e ON e.beneficiary_agent_id = a.id
    JOIN users u ON u.id = a.user_id
    WHERE u.email = 'admin@sub2api.local' OR u.email LIKE 'demo.distribution.agent.%@sub2api.local'
    GROUP BY a.id
), reserved AS (
    SELECT agent_id, COALESCE(SUM(amount_cny), 0) AS amount
    FROM distribution_withdrawals
    WHERE status IN ('pending', 'approved', 'paying')
    GROUP BY agent_id
)
UPDATE distribution_wallets wallet
SET frozen_cny = stats.frozen,
    available_cny = GREATEST(stats.available_before_reserve - COALESCE(reserved.amount, 0), 0),
    reserved_cny = COALESCE(reserved.amount, 0),
    debt_cny = 0,
    total_earned_cny = stats.earned,
    total_withdrawn_cny = stats.withdrawn,
    total_converted_cny = stats.converted,
    updated_at = NOW()
FROM stats
LEFT JOIN reserved ON reserved.agent_id = stats.agent_id
WHERE wallet.agent_id = stats.agent_id;

COMMIT;

SELECT
    u.email,
    l.depth,
    a.status,
    a.promotion_code,
    wallet.available_cny,
    wallet.frozen_cny,
    wallet.reserved_cny,
    wallet.total_earned_cny
FROM distribution_agents a
JOIN users u ON u.id = a.user_id
JOIN distribution_agent_levels l ON l.id = a.level_id
JOIN distribution_wallets wallet ON wallet.agent_id = a.id
WHERE u.email = 'admin@sub2api.local' OR u.email LIKE 'demo.distribution.agent.%@sub2api.local'
ORDER BY l.depth, u.email;
