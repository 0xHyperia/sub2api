-- Independent two-level agent distribution system.
-- All money in this module is CNY and all rates are integer basis points.

CREATE TABLE IF NOT EXISTS distribution_agent_levels (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    depth SMALLINT NOT NULL UNIQUE CHECK (depth IN (1, 2)),
    default_rate_bps INTEGER NOT NULL CHECK (default_rate_bps BETWEEN 0 AND 10000),
    max_child_rate_bps INTEGER NOT NULL DEFAULT 0 CHECK (max_child_rate_bps BETWEEN 0 AND 10000),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO distribution_agent_levels (name, depth, default_rate_bps, max_child_rate_bps)
VALUES ('一级代理', 1, 1000, 800), ('二级代理', 2, 600, 0)
ON CONFLICT (depth) DO NOTHING;

CREATE TABLE IF NOT EXISTS distribution_agents (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
    level_id BIGINT NOT NULL REFERENCES distribution_agent_levels(id) ON DELETE RESTRICT,
    parent_agent_id BIGINT REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    promotion_code VARCHAR(32) NOT NULL UNIQUE,
    rate_override_bps INTEGER CHECK (rate_override_bps BETWEEN 0 AND 10000),
    status VARCHAR(16) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspended', 'revoked')),
    can_recruit_subagents BOOLEAN NOT NULL DEFAULT FALSE,
    granted_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    status_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (user_id > 0),
    CHECK (parent_agent_id IS NULL OR parent_agent_id <> id),
    CHECK (promotion_code ~ '^[A-Z0-9_-]{4,32}$')
);

CREATE INDEX IF NOT EXISTS idx_distribution_agents_parent ON distribution_agents(parent_agent_id);
CREATE INDEX IF NOT EXISTS idx_distribution_agents_status ON distribution_agents(status);
CREATE INDEX IF NOT EXISTS idx_distribution_agents_admin_status_created
    ON distribution_agents(status, created_at DESC, id DESC);

COMMENT ON COLUMN distribution_agents.can_recruit_subagents IS
    'Administrator-granted permission for an active top-level agent to recruit subagents.';

CREATE TABLE IF NOT EXISTS distribution_agent_events (
    id BIGSERIAL PRIMARY KEY,
    agent_id BIGINT NOT NULL REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    event_type VARCHAR(32) NOT NULL CHECK (event_type IN (
        'created', 'status_changed', 'rate_changed', 'permission_changed'
    )),
    old_status VARCHAR(16),
    new_status VARCHAR(16),
    old_rate_override_bps INTEGER CHECK (old_rate_override_bps BETWEEN 0 AND 10000),
    new_rate_override_bps INTEGER CHECK (new_rate_override_bps BETWEEN 0 AND 10000),
    old_effective_rate_bps INTEGER CHECK (old_effective_rate_bps BETWEEN 0 AND 10000),
    new_effective_rate_bps INTEGER CHECK (new_effective_rate_bps BETWEEN 0 AND 10000),
    reason TEXT NOT NULL DEFAULT '',
    actor_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_distribution_agent_events_agent
    ON distribution_agent_events(agent_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS distribution_customer_bindings (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE RESTRICT,
    agent_id BIGINT NOT NULL REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    promotion_code VARCHAR(32) NOT NULL,
    bound_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    corrected_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    correction_reason TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_distribution_customer_bindings_agent
    ON distribution_customer_bindings(agent_id, bound_at DESC);
CREATE INDEX IF NOT EXISTS idx_distribution_customer_bindings_admin_bound
    ON distribution_customer_bindings(bound_at DESC, user_id);

CREATE TABLE IF NOT EXISTS distribution_binding_events (
    id BIGSERIAL PRIMARY KEY,
    customer_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    old_agent_id BIGINT REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    new_agent_id BIGINT NOT NULL REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    old_promotion_code VARCHAR(32),
    new_promotion_code VARCHAR(32) NOT NULL,
    reason TEXT NOT NULL,
    actor_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_distribution_binding_events_customer
    ON distribution_binding_events(customer_user_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS distribution_commission_sources (
    id BIGSERIAL PRIMARY KEY,
    payment_order_id BIGINT NOT NULL UNIQUE REFERENCES payment_orders(id) ON DELETE RESTRICT,
    customer_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    direct_agent_id BIGINT NOT NULL REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    payment_type VARCHAR(32) NOT NULL,
    payment_currency VARCHAR(8) NOT NULL,
    actual_paid_amount NUMERIC(20,8) NOT NULL CHECK (actual_paid_amount >= 0),
    fx_rate_to_cny NUMERIC(20,8) CHECK (fx_rate_to_cny > 0),
    commission_base_cny NUMERIC(20,8) CHECK (commission_base_cny >= 0),
    refunded_amount_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (refunded_amount_cny >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'pending_fx'
        CHECK (status IN ('pending_fx', 'settled', 'partially_refunded', 'fully_refunded', 'void')),
    provider_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    paid_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_distribution_commission_sources_customer
    ON distribution_commission_sources(customer_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_distribution_commission_sources_agent
    ON distribution_commission_sources(direct_agent_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_distribution_commission_sources_status
    ON distribution_commission_sources(status, created_at);
CREATE INDEX IF NOT EXISTS idx_distribution_commission_sources_paid_at
    ON distribution_commission_sources(paid_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS distribution_commission_entries (
    id BIGSERIAL PRIMARY KEY,
    source_id BIGINT NOT NULL REFERENCES distribution_commission_sources(id) ON DELETE RESTRICT,
    beneficiary_agent_id BIGINT NOT NULL REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    entry_type VARCHAR(16) NOT NULL CHECK (entry_type IN ('direct', 'team')),
    rate_bps INTEGER NOT NULL CHECK (rate_bps BETWEEN 0 AND 10000),
    original_amount_cny NUMERIC(20,8) NOT NULL CHECK (original_amount_cny >= 0),
    reversed_amount_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (reversed_amount_cny >= 0),
    status VARCHAR(16) NOT NULL DEFAULT 'frozen'
        CHECK (status IN ('frozen', 'available', 'reversed')),
    available_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source_id, beneficiary_agent_id, entry_type),
    CHECK (reversed_amount_cny <= original_amount_cny)
);

CREATE INDEX IF NOT EXISTS idx_distribution_commission_entries_beneficiary
    ON distribution_commission_entries(beneficiary_agent_id, status, available_at);
CREATE INDEX IF NOT EXISTS idx_distribution_commission_entries_admin_sort
    ON distribution_commission_entries(status, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS distribution_wallets (
    agent_id BIGINT PRIMARY KEY REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    frozen_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (frozen_cny >= 0),
    available_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (available_cny >= 0),
    reserved_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (reserved_cny >= 0),
    debt_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (debt_cny >= 0),
    total_earned_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (total_earned_cny >= 0),
    total_withdrawn_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (total_withdrawn_cny >= 0),
    total_converted_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (total_converted_cny >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS distribution_wallet_ledger (
    id BIGSERIAL PRIMARY KEY,
    agent_id BIGINT NOT NULL REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    commission_entry_id BIGINT REFERENCES distribution_commission_entries(id) ON DELETE RESTRICT,
    withdrawal_id BIGINT,
    entry_type VARCHAR(24) NOT NULL CHECK (entry_type IN (
        'commission_frozen', 'commission_released', 'refund_reversal', 'debt_offset',
        'withdrawal_reserved', 'withdrawal_paid', 'withdrawal_released',
        'balance_conversion', 'manual_adjustment'
    )),
    amount_cny NUMERIC(20,8) NOT NULL,
    frozen_after_cny NUMERIC(20,8) NOT NULL,
    available_after_cny NUMERIC(20,8) NOT NULL,
    reserved_after_cny NUMERIC(20,8) NOT NULL,
    debt_after_cny NUMERIC(20,8) NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    note TEXT,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_distribution_wallet_ledger_agent
    ON distribution_wallet_ledger(agent_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS distribution_payout_accounts (
    agent_id BIGINT PRIMARY KEY REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    alipay_name VARCHAR(128) NOT NULL,
    alipay_account VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE SEQUENCE IF NOT EXISTS distribution_withdrawal_no_seq;

CREATE TABLE IF NOT EXISTS distribution_withdrawals (
    id BIGSERIAL PRIMARY KEY,
    request_no VARCHAR(32) NOT NULL DEFAULT (
        'DW' || TO_CHAR(CURRENT_DATE, 'YYYYMMDD') ||
        LPAD(NEXTVAL('distribution_withdrawal_no_seq')::TEXT, 10, '0')
    ),
    agent_id BIGINT NOT NULL REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    amount_cny NUMERIC(20,8) NOT NULL CHECK (amount_cny > 0),
    fee_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (fee_cny >= 0),
    payout_cny NUMERIC(20,8) NOT NULL CHECK (payout_cny > 0),
    alipay_name VARCHAR(128) NOT NULL,
    alipay_account VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected', 'paying', 'paid', 'failed', 'cancelled')),
    review_note TEXT,
    approved_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    approved_at TIMESTAMPTZ,
    reviewed_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    payment_reference VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT distribution_withdrawals_paid_reference_check
        CHECK (status <> 'paid' OR (payment_reference IS NOT NULL AND BTRIM(payment_reference) <> ''))
);

ALTER TABLE distribution_wallet_ledger
    DROP CONSTRAINT IF EXISTS fk_distribution_wallet_ledger_withdrawal;
ALTER TABLE distribution_wallet_ledger
    ADD CONSTRAINT fk_distribution_wallet_ledger_withdrawal
    FOREIGN KEY (withdrawal_id) REFERENCES distribution_withdrawals(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_distribution_withdrawals_agent
    ON distribution_withdrawals(agent_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_distribution_withdrawals_status
    ON distribution_withdrawals(status, created_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_distribution_withdrawals_request_no
    ON distribution_withdrawals(request_no);
CREATE UNIQUE INDEX IF NOT EXISTS idx_distribution_withdrawals_payment_reference_unique
    ON distribution_withdrawals(payment_reference)
    WHERE payment_reference IS NOT NULL AND BTRIM(payment_reference) <> '';
CREATE INDEX IF NOT EXISTS idx_distribution_withdrawals_approved_by
    ON distribution_withdrawals(approved_by, approved_at DESC)
    WHERE approved_by IS NOT NULL;

CREATE TABLE IF NOT EXISTS distribution_withdrawal_events (
    id BIGSERIAL PRIMARY KEY,
    withdrawal_id BIGINT NOT NULL REFERENCES distribution_withdrawals(id) ON DELETE RESTRICT,
    from_status VARCHAR(20),
    to_status VARCHAR(20) NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    payment_reference VARCHAR(255),
    actor_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_distribution_withdrawal_events_withdrawal
    ON distribution_withdrawal_events(withdrawal_id, created_at, id);

CREATE TABLE IF NOT EXISTS distribution_withdrawal_attachments (
    id BIGSERIAL PRIMARY KEY,
    withdrawal_id BIGINT NOT NULL REFERENCES distribution_withdrawals(id) ON DELETE RESTRICT,
    object_key VARCHAR(1024) NOT NULL UNIQUE,
    original_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes > 0),
    sha256 VARCHAR(64) NOT NULL,
    evidence_type VARCHAR(24) NOT NULL DEFAULT 'payment_receipt'
        CHECK (evidence_type IN ('payment_receipt', 'bank_statement', 'other')),
    note TEXT NOT NULL DEFAULT '',
    uploaded_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    deleted_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    delete_reason TEXT
);

CREATE INDEX IF NOT EXISTS idx_distribution_withdrawal_attachments_withdrawal
    ON distribution_withdrawal_attachments(withdrawal_id, created_at, id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS distribution_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    freeze_hours INTEGER NOT NULL DEFAULT 168 CHECK (freeze_hours >= 0),
    withdrawal_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    withdrawal_dual_approval_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    minimum_withdrawal_cny NUMERIC(20,8) NOT NULL DEFAULT 100 CHECK (minimum_withdrawal_cny >= 0),
    maximum_withdrawal_cny NUMERIC(20,8) NOT NULL DEFAULT 50000 CHECK (maximum_withdrawal_cny >= minimum_withdrawal_cny),
    withdrawal_fee_rate_bps INTEGER NOT NULL DEFAULT 0 CHECK (withdrawal_fee_rate_bps BETWEEN 0 AND 10000),
    withdrawal_fee_fixed_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (withdrawal_fee_fixed_cny >= 0),
    daily_withdrawal_limit_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (daily_withdrawal_limit_cny >= 0),
    monthly_withdrawal_limit_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (monthly_withdrawal_limit_cny >= 0),
    cny_per_platform_usd NUMERIC(20,8) NOT NULL DEFAULT 1 CHECK (cny_per_platform_usd = 1),
    updated_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO distribution_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS distribution_fx_rates (
    currency VARCHAR(8) PRIMARY KEY,
    rate_to_cny NUMERIC(20,8) NOT NULL CHECK (rate_to_cny > 0),
    updated_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO distribution_fx_rates (currency, rate_to_cny)
VALUES ('CNY', 1)
ON CONFLICT (currency) DO UPDATE SET rate_to_cny = 1, updated_at = NOW();

COMMENT ON TABLE distribution_customer_bindings IS
    'Permanent registration-time customer-to-agent binding, independent from user_affiliates';
COMMENT ON TABLE distribution_commission_sources IS
    'One idempotent fee-excluded recharge or subscription commission source per payment order';
COMMENT ON TABLE distribution_wallets IS
    'Independent CNY distribution wallet; never shares the invitation rebate quota ledger';
COMMENT ON TABLE distribution_agent_events IS
    'Append-only history for agent lifecycle, commission-rate, and recruitment-permission changes';
COMMENT ON TABLE distribution_binding_events IS
    'Append-only customer ownership correction history';
COMMENT ON TABLE distribution_withdrawal_events IS
    'Append-only withdrawal workflow history';
COMMENT ON TABLE distribution_withdrawal_attachments IS
    'Private payment evidence metadata; object content lives in S3-compatible storage';
COMMENT ON COLUMN distribution_commission_entries.status IS
    'Commission arrival status: frozen, available, or reversed. Settlement actions are recorded separately.';
COMMENT ON COLUMN distribution_settings.withdrawal_dual_approval_enabled IS
    'When enabled, the administrator who approves a withdrawal cannot start its payment';
COMMENT ON COLUMN distribution_withdrawals.approved_by IS
    'Administrator responsible for the first-stage withdrawal approval';
