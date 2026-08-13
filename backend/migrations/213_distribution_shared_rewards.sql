-- Shared fixed rewards for an L1 distribution hierarchy. L1 rules define the
-- total budget; L2 rules define the share allocated from that budget.
ALTER TABLE distribution_settings
    ADD COLUMN IF NOT EXISTS registration_reward_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS recharge_reward_enabled BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS distribution_agent_reward_rules (
    agent_id BIGINT PRIMARY KEY REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    registration_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    registration_reward_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (registration_reward_cny >= 0),
    recharge_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    recharge_threshold_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (recharge_threshold_cny >= 0),
    recharge_reward_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (recharge_reward_cny >= 0),
    updated_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE distribution_agent_reward_rules
    ADD COLUMN IF NOT EXISTS registration_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS recharge_enabled BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS distribution_reward_events (
    id BIGSERIAL PRIMARY KEY,
    customer_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    root_agent_id BIGINT NOT NULL REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    direct_agent_id BIGINT NOT NULL REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    reward_type VARCHAR(32) NOT NULL CHECK (reward_type IN ('registration', 'recharge_threshold')),
    trigger_payment_order_id BIGINT REFERENCES payment_orders(id) ON DELETE RESTRICT,
    threshold_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (threshold_cny >= 0),
    total_reward_cny NUMERIC(20,8) NOT NULL CHECK (total_reward_cny > 0),
    direct_share_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (direct_share_cny >= 0),
    config_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (customer_user_id, reward_type),
    CHECK (direct_share_cny <= total_reward_cny),
    CHECK ((reward_type='registration' AND trigger_payment_order_id IS NULL) OR
           (reward_type='recharge_threshold' AND trigger_payment_order_id IS NOT NULL))
);

CREATE INDEX IF NOT EXISTS idx_distribution_reward_events_payment
    ON distribution_reward_events(trigger_payment_order_id)
    WHERE trigger_payment_order_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS distribution_reward_grants (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES distribution_reward_events(id) ON DELETE RESTRICT,
    beneficiary_agent_id BIGINT NOT NULL REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    beneficiary_role VARCHAR(16) NOT NULL CHECK (beneficiary_role IN ('root', 'direct')),
    original_amount_cny NUMERIC(20,8) NOT NULL CHECK (original_amount_cny > 0),
    reversed_amount_cny NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (reversed_amount_cny >= 0),
    status VARCHAR(16) NOT NULL DEFAULT 'frozen' CHECK (status IN ('frozen', 'available', 'reversed')),
    available_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (event_id, beneficiary_agent_id),
    CHECK (reversed_amount_cny <= original_amount_cny)
);

CREATE INDEX IF NOT EXISTS idx_distribution_reward_grants_maturity
    ON distribution_reward_grants(status, available_at, id);

ALTER TABLE distribution_wallet_ledger
    ADD COLUMN IF NOT EXISTS reward_grant_id BIGINT REFERENCES distribution_reward_grants(id) ON DELETE RESTRICT;

ALTER TABLE distribution_wallet_ledger DROP CONSTRAINT IF EXISTS distribution_wallet_ledger_entry_type_check;
ALTER TABLE distribution_wallet_ledger ADD CONSTRAINT distribution_wallet_ledger_entry_type_check CHECK (entry_type IN (
    'commission_frozen', 'commission_released', 'refund_reversal', 'debt_offset',
    'reward_frozen', 'reward_released', 'reward_reversal',
    'withdrawal_reserved', 'withdrawal_paid', 'withdrawal_released',
    'balance_conversion', 'manual_adjustment'
));

CREATE UNIQUE INDEX IF NOT EXISTS idx_distribution_wallet_ledger_reward_grant_action
    ON distribution_wallet_ledger(reward_grant_id, entry_type)
    WHERE reward_grant_id IS NOT NULL;
