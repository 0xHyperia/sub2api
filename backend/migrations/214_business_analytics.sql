-- Shared facts and aggregates for platform and distribution business analytics.
-- Historical rows are populated lazily by analytics queries/backfill jobs; these
-- tables provide stable incremental targets without changing payment semantics.
CREATE TABLE IF NOT EXISTS business_user_facts (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    first_activated_at TIMESTAMPTZ,
    first_paid_at TIMESTAMPTZ,
    channel VARCHAR(24) NOT NULL DEFAULT 'unknown'
        CHECK (channel IN ('distribution','affiliate','campaign','organic','unknown')),
    channel_ref_id BIGINT,
    attribution_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_business_user_facts_first_activated
    ON business_user_facts(first_activated_at) WHERE first_activated_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_business_user_facts_first_paid
    ON business_user_facts(first_paid_at) WHERE first_paid_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_business_user_facts_channel
    ON business_user_facts(channel, user_id);

-- Sparse, durable user/hour facts keep business reports independent from the
-- shorter raw usage-log retention window. One row exists only when a user has
-- successful billable inference in that hour.
CREATE TABLE IF NOT EXISTS business_usage_hourly_facts (
    bucket_start TIMESTAMPTZ NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    first_used_at TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ NOT NULL,
    consumed_revenue NUMERIC(20,10) NOT NULL DEFAULT 0,
    supplier_cost NUMERIC(20,10) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (bucket_start, user_id)
);

CREATE INDEX IF NOT EXISTS idx_business_usage_hourly_facts_user_time
    ON business_usage_hourly_facts(user_id, bucket_start);

ALTER TABLE business_usage_hourly_facts
    ADD COLUMN IF NOT EXISTS first_used_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMPTZ;

-- Only relevant while iterating this unpublished migration: an earlier local
-- draft may have created rows before exact event bounds were added.
UPDATE business_usage_hourly_facts
SET first_used_at = COALESCE(first_used_at, bucket_start),
    last_used_at = COALESCE(last_used_at, bucket_start)
WHERE first_used_at IS NULL OR last_used_at IS NULL;

ALTER TABLE business_usage_hourly_facts
    ALTER COLUMN first_used_at SET NOT NULL,
    ALTER COLUMN last_used_at SET NOT NULL;

CREATE TABLE IF NOT EXISTS business_analytics_buckets (
    bucket_start TIMESTAMPTZ NOT NULL,
    resolution VARCHAR(8) NOT NULL CHECK (resolution IN ('hour','day')),
    scope VARCHAR(24) NOT NULL CHECK (scope IN ('platform','channel','agent_hierarchy','agent')),
    scope_id VARCHAR(64) NOT NULL DEFAULT '',
    metrics JSONB NOT NULL DEFAULT '{}'::jsonb,
    estimated BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (resolution, bucket_start, scope, scope_id)
);

CREATE INDEX IF NOT EXISTS idx_business_analytics_buckets_scope_time
    ON business_analytics_buckets(scope, scope_id, resolution, bucket_start DESC);

CREATE TABLE IF NOT EXISTS business_payment_facts (
    payment_order_id BIGINT PRIMARY KEY REFERENCES payment_orders(id) ON DELETE CASCADE,
    currency VARCHAR(3) NOT NULL,
    gross_amount NUMERIC(20,8) NOT NULL DEFAULT 0,
    refunded_amount NUMERIC(20,8) NOT NULL DEFAULT 0,
    fx_rate_to_cny NUMERIC(20,8),
    fx_estimated BOOLEAN NOT NULL DEFAULT FALSE,
    paid_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    channel VARCHAR(24) NOT NULL DEFAULT 'unknown'
        CHECK (channel IN ('distribution','affiliate','campaign','organic','unknown')),
    channel_ref_id BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE business_payment_facts
    ADD COLUMN IF NOT EXISTS channel VARCHAR(24) NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS channel_ref_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_business_payment_facts_paid_at
    ON business_payment_facts(paid_at) WHERE paid_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_business_payment_facts_refunded_at
    ON business_payment_facts(refunded_at) WHERE refunded_at IS NOT NULL;

-- Keep a retry cursor separate from the established dashboard watermark. A
-- failed supplementary aggregation must neither block the primary dashboard
-- nor permanently strand its own missing buckets.
CREATE TABLE IF NOT EXISTS business_analytics_aggregation_state (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    last_aggregated_at TIMESTAMPTZ,
    pending_from TIMESTAMPTZ,
    usage_coverage_from TIMESTAMPTZ,
    invalidation_version BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE business_analytics_aggregation_state
    ADD COLUMN IF NOT EXISTS invalidation_version BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS usage_coverage_from TIMESTAMPTZ;

INSERT INTO business_analytics_aggregation_state (id)
VALUES (1)
ON CONFLICT (id) DO NOTHING;

-- Persist the recoverable boundary once. It must not move forward when the
-- durable fact retention job later removes old rows. On a fresh installation
-- without usage, NOW() is the only boundary that can be stated safely.
UPDATE business_analytics_aggregation_state
SET usage_coverage_from = COALESCE(
    usage_coverage_from,
    (SELECT MIN(created_at) FROM usage_logs WHERE actual_cost > 0),
    NOW()
)
WHERE id = 1;

-- Mark the earliest event time whose hourly and daily buckets must be rebuilt.
-- The monotonically increasing version prevents an aggregation that was already
-- running from clearing a newer invalidation raised during that transaction.
CREATE OR REPLACE FUNCTION business_invalidate_analytics(invalid_from TIMESTAMPTZ)
RETURNS VOID AS $$
BEGIN
    IF invalid_from IS NULL THEN
        RETURN;
    END IF;
    INSERT INTO business_analytics_aggregation_state (
        id, pending_from, invalidation_version, updated_at
    ) VALUES (1, invalid_from, 1, NOW())
    ON CONFLICT (id) DO UPDATE SET
        pending_from = LEAST(
            COALESCE(business_analytics_aggregation_state.pending_from, EXCLUDED.pending_from),
            EXCLUDED.pending_from
        ),
        invalidation_version = business_analytics_aggregation_state.invalidation_version + 1,
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- Backfill immutable lifecycle and attribution facts once. Unknown is deliberate:
-- registrations without reliable evidence must never be reported as organic.
INSERT INTO business_user_facts (
    user_id, first_activated_at, first_paid_at, channel, channel_ref_id,
    attribution_snapshot, created_at, updated_at
)
SELECT
    u.id,
    NULL,
    (SELECT MIN(po.paid_at) FROM payment_orders po
        WHERE po.user_id = u.id
          AND po.status IN (
              'PAID', 'RECHARGING', 'COMPLETED', 'FAILED',
              'REFUND_REQUESTED', 'REFUNDING', 'REFUND_PENDING',
              'REFUND_FAILED', 'PARTIALLY_REFUNDED', 'REFUNDED'
          )
          AND po.paid_at IS NOT NULL
          AND NOT (
              po.status = 'REFUNDED'
              AND COALESCE(po.refund_amount, 0) >= COALESCE(po.amount, 0)
          )),
    CASE
        WHEN dcb.user_id IS NOT NULL THEN 'distribution'
        WHEN ua.inviter_id IS NOT NULL THEN 'affiliate'
        ELSE 'unknown'
    END,
    COALESCE(dcb.agent_id, ua.inviter_id),
    CASE
        WHEN dcb.user_id IS NOT NULL THEN jsonb_build_object('source', 'distribution_binding', 'promotion_code', dcb.promotion_code)
        WHEN ua.inviter_id IS NOT NULL THEN jsonb_build_object('source', 'legacy_affiliate')
        ELSE '{}'::jsonb
    END,
    u.created_at,
    NOW()
FROM users u
LEFT JOIN distribution_customer_bindings dcb ON dcb.user_id = u.id
LEFT JOIN user_affiliates ua ON ua.user_id = u.id
ON CONFLICT (user_id) DO UPDATE SET
    first_activated_at = COALESCE(business_user_facts.first_activated_at, EXCLUDED.first_activated_at),
    first_paid_at = COALESCE(business_user_facts.first_paid_at, EXCLUDED.first_paid_at),
    channel = CASE WHEN business_user_facts.channel='unknown' THEN EXCLUDED.channel ELSE business_user_facts.channel END,
    channel_ref_id = CASE WHEN business_user_facts.channel='unknown' THEN EXCLUDED.channel_ref_id ELSE business_user_facts.channel_ref_id END,
    attribution_snapshot = CASE WHEN business_user_facts.channel='unknown' THEN EXCLUDED.attribution_snapshot ELSE business_user_facts.attribution_snapshot END,
    updated_at = NOW();

CREATE OR REPLACE FUNCTION business_effective_usd_to_cny_rate()
RETURNS NUMERIC AS $$
DECLARE
    auto_enabled BOOLEAN := FALSE;
    manual_value TEXT;
    auto_value TEXT;
    selected_value TEXT;
BEGIN
    SELECT value INTO manual_value FROM settings WHERE key = 'currency_usd_to_cny_manual_rate';
    SELECT value INTO auto_value FROM settings WHERE key = 'currency_usd_to_cny_auto_rate';
    SELECT COALESCE(value, 'false') = 'true' INTO auto_enabled
      FROM settings WHERE key = 'currency_exchange_rate_auto_sync';
    selected_value := CASE WHEN auto_enabled THEN auto_value ELSE manual_value END;
    IF selected_value ~ '^[0-9]+([.][0-9]+)?$' AND selected_value::NUMERIC > 0 THEN
        RETURN selected_value::NUMERIC;
    END IF;
    RETURN 7.20;
END;
$$ LANGUAGE plpgsql STABLE;

CREATE OR REPLACE FUNCTION business_effective_fx_rate_to_cny(payment_currency TEXT)
RETURNS NUMERIC AS $$
DECLARE
    normalized_currency TEXT := UPPER(COALESCE(payment_currency, ''));
    configured_rate NUMERIC;
BEGIN
    IF normalized_currency = 'CNY' THEN
        RETURN 1;
    END IF;
    IF normalized_currency = 'USD' THEN
        RETURN business_effective_usd_to_cny_rate();
    END IF;
    SELECT rate_to_cny INTO configured_rate
      FROM distribution_fx_rates
     WHERE currency = normalized_currency AND rate_to_cny > 0;
    RETURN configured_rate;
END;
$$ LANGUAGE plpgsql STABLE;

CREATE OR REPLACE FUNCTION business_sync_payment_fact()
RETURNS TRIGGER AS $$
DECLARE
    payment_currency TEXT;
    original_amount NUMERIC;
    refund_in_currency NUMERIC;
    snapshot_rate NUMERIC;
    estimated BOOLEAN := FALSE;
    invalid_from TIMESTAMPTZ;
BEGIN
    invalid_from := LEAST(NEW.paid_at, COALESCE(NEW.refund_at, NEW.paid_at));
    IF TG_OP = 'UPDATE' THEN
        invalid_from := LEAST(
            COALESCE(invalid_from, OLD.paid_at, OLD.refund_at),
            COALESCE(OLD.paid_at, OLD.refund_at, invalid_from),
            COALESCE(OLD.refund_at, OLD.paid_at, invalid_from)
        );
    END IF;
    IF invalid_from IS NOT NULL THEN
        PERFORM business_invalidate_analytics(invalid_from);
    END IF;
    -- A paid order remains revenue while a refund is requested, processing,
    -- pending, or failed. Only orders that were never paid lose their fact.
    IF NEW.status NOT IN (
        'PAID', 'RECHARGING', 'COMPLETED', 'FAILED',
        'REFUND_REQUESTED', 'REFUNDING', 'REFUND_PENDING',
        'REFUND_FAILED', 'PARTIALLY_REFUNDED', 'REFUNDED'
    ) OR NEW.paid_at IS NULL THEN
        DELETE FROM business_payment_facts WHERE payment_order_id = NEW.id;
        RETURN NEW;
    END IF;

    payment_currency := UPPER(COALESCE(NULLIF(NEW.provider_snapshot->>'currency', ''), 'CNY'));
    original_amount := COALESCE(NULLIF(NEW.provider_amount, 0), NEW.pay_amount, 0);
    refund_in_currency := CASE
        WHEN NEW.status IN ('PARTIALLY_REFUNDED', 'REFUNDED') AND NEW.amount > 0
        THEN original_amount * COALESCE(NEW.refund_amount, 0) / NEW.amount
        ELSE 0
    END;
    snapshot_rate := business_effective_fx_rate_to_cny(payment_currency);
    estimated := snapshot_rate IS NULL;

    INSERT INTO business_payment_facts (
        payment_order_id, currency, gross_amount, refunded_amount,
        fx_rate_to_cny, fx_estimated, paid_at, refunded_at, channel, channel_ref_id, updated_at
    ) VALUES (
        NEW.id, payment_currency, original_amount, refund_in_currency,
        snapshot_rate, estimated, NEW.paid_at,
        CASE WHEN NEW.status IN ('PARTIALLY_REFUNDED', 'REFUNDED') THEN NEW.refund_at ELSE NULL END,
        COALESCE((SELECT channel FROM business_user_facts WHERE user_id=NEW.user_id),'unknown'),
        (SELECT channel_ref_id FROM business_user_facts WHERE user_id=NEW.user_id), NOW()
    )
    ON CONFLICT (payment_order_id) DO UPDATE SET
        currency = EXCLUDED.currency,
        gross_amount = EXCLUDED.gross_amount,
        refunded_amount = EXCLUDED.refunded_amount,
        -- Preserve the event-time rate when an existing order is later refunded.
        fx_rate_to_cny = COALESCE(business_payment_facts.fx_rate_to_cny, EXCLUDED.fx_rate_to_cny),
		fx_estimated = business_payment_facts.fx_estimated OR EXCLUDED.fx_estimated,
        paid_at = EXCLUDED.paid_at,
        refunded_at = EXCLUDED.refunded_at,
        -- Payment attribution is an event-time snapshot. Refund updates must not
        -- move historical revenue to the customer's current channel.
        channel = business_payment_facts.channel,
        channel_ref_id = business_payment_facts.channel_ref_id,
        updated_at = NOW();

    INSERT INTO business_user_facts (user_id, first_paid_at, created_at, updated_at)
    SELECT NEW.user_id,
        (SELECT MIN(f.paid_at)
         FROM business_payment_facts f
         JOIN payment_orders paid_order ON paid_order.id = f.payment_order_id
         WHERE paid_order.user_id = NEW.user_id
           AND f.paid_at IS NOT NULL
           AND f.gross_amount > f.refunded_amount),
        u.created_at, NOW()
    FROM users u WHERE u.id = NEW.user_id
    ON CONFLICT (user_id) DO UPDATE SET
        first_paid_at = EXCLUDED.first_paid_at,
        updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Install the trigger before the historical scan. Changes committed while the
-- backfill is running are then either visible to the scan or captured by the
-- trigger, closing the deployment race between those two operations.
DROP TRIGGER IF EXISTS business_payment_fact_sync ON payment_orders;
CREATE TRIGGER business_payment_fact_sync
AFTER INSERT OR UPDATE OF status, paid_at, refund_amount, refund_at, provider_amount, pay_amount, provider_snapshot
ON payment_orders FOR EACH ROW EXECUTE FUNCTION business_sync_payment_fact();

INSERT INTO business_payment_facts (
    payment_order_id, currency, gross_amount, refunded_amount,
    fx_rate_to_cny, fx_estimated, paid_at, refunded_at, channel, channel_ref_id, updated_at
)
SELECT
    po.id,
    UPPER(COALESCE(NULLIF(po.provider_snapshot->>'currency', ''), 'CNY')),
    COALESCE(NULLIF(po.provider_amount, 0), po.pay_amount, 0),
    CASE WHEN po.status IN ('PARTIALLY_REFUNDED', 'REFUNDED') AND po.amount > 0
        THEN COALESCE(NULLIF(po.provider_amount, 0), po.pay_amount, 0) * COALESCE(po.refund_amount, 0) / po.amount
        ELSE 0 END,
    business_effective_fx_rate_to_cny(UPPER(COALESCE(NULLIF(po.provider_snapshot->>'currency', ''), 'CNY'))),
    -- Historical USD orders do not have an event-time FX snapshot. They use the
    -- current base rate during backfill and must remain visibly estimated.
    business_effective_fx_rate_to_cny(UPPER(COALESCE(NULLIF(po.provider_snapshot->>'currency', ''), 'CNY'))) IS NULL
        OR UPPER(COALESCE(NULLIF(po.provider_snapshot->>'currency', ''), 'CNY')) <> 'CNY',
    po.paid_at,
    CASE WHEN po.status IN ('PARTIALLY_REFUNDED', 'REFUNDED') THEN po.refund_at ELSE NULL END,
    COALESCE(buf.channel,'unknown'),
    buf.channel_ref_id,
    NOW()
FROM payment_orders po
LEFT JOIN business_user_facts buf ON buf.user_id=po.user_id
WHERE po.status IN (
    'PAID', 'RECHARGING', 'COMPLETED', 'FAILED',
    'REFUND_REQUESTED', 'REFUNDING', 'REFUND_PENDING',
    'REFUND_FAILED', 'PARTIALLY_REFUNDED', 'REFUNDED'
) AND po.paid_at IS NOT NULL
ON CONFLICT (payment_order_id) DO UPDATE SET
    currency = EXCLUDED.currency,
    gross_amount = EXCLUDED.gross_amount,
    refunded_amount = EXCLUDED.refunded_amount,
    fx_rate_to_cny = COALESCE(business_payment_facts.fx_rate_to_cny, EXCLUDED.fx_rate_to_cny),
    fx_estimated = business_payment_facts.fx_estimated OR business_payment_facts.fx_rate_to_cny IS NULL,
    paid_at = EXCLUDED.paid_at,
    refunded_at = EXCLUDED.refunded_at,
    updated_at = NOW();

-- Rebuild first-payment facts from effective payments after the payment fact
-- backfill. A fully refunded order is cash history, but it is not a retained
-- first payment or a repurchase baseline.
UPDATE business_user_facts user_fact
SET first_paid_at = effective.first_paid_at, updated_at = NOW()
FROM (
    SELECT u.id AS user_id, MIN(f.paid_at) FILTER (WHERE f.gross_amount > f.refunded_amount) AS first_paid_at
    FROM users u
    LEFT JOIN payment_orders po ON po.user_id = u.id
    LEFT JOIN business_payment_facts f ON f.payment_order_id = po.id
    GROUP BY u.id
) effective
WHERE user_fact.user_id = effective.user_id
  AND user_fact.first_paid_at IS DISTINCT FROM effective.first_paid_at;

CREATE OR REPLACE FUNCTION business_sync_usage_fact()
RETURNS TRIGGER AS $$
BEGIN
    -- Ordinary realtime usage is already covered by the scheduler lookback.
    -- Only imported or delayed historical usage needs the shared invalidation
    -- cursor; updating it for every request would create a hot database row.
    IF NEW.actual_cost > 0 AND NEW.created_at < NOW() - interval '10 minutes' THEN
        PERFORM business_invalidate_analytics(NEW.created_at);
    END IF;
    IF NEW.actual_cost > 0 THEN
        INSERT INTO business_user_facts (user_id, first_activated_at, created_at, updated_at)
        SELECT NEW.user_id, NEW.created_at, u.created_at, NOW() FROM users u WHERE u.id = NEW.user_id
        ON CONFLICT (user_id) DO UPDATE SET
            first_activated_at = LEAST(COALESCE(business_user_facts.first_activated_at, EXCLUDED.first_activated_at), EXCLUDED.first_activated_at),
            updated_at = NOW()
        WHERE business_user_facts.first_activated_at IS NULL
           OR business_user_facts.first_activated_at > EXCLUDED.first_activated_at;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS business_usage_fact_sync ON usage_logs;
CREATE TRIGGER business_usage_fact_sync
AFTER INSERT ON usage_logs FOR EACH ROW EXECUTE FUNCTION business_sync_usage_fact();

-- Backfill once after installing the trigger. New rows committed during or
-- after this scan are handled by the trigger and the scheduler's lookback.
INSERT INTO business_usage_hourly_facts (
    bucket_start, user_id, first_used_at, last_used_at,
    consumed_revenue, supplier_cost, updated_at
)
SELECT
    date_trunc('hour', timezone('Asia/Shanghai', ul.created_at)) AT TIME ZONE 'Asia/Shanghai',
    ul.user_id, MIN(ul.created_at), MAX(ul.created_at),
    COALESCE(SUM(ul.actual_cost), 0),
    COALESCE(SUM(COALESCE(ul.account_stats_cost, ul.total_cost) * COALESCE(ul.account_rate_multiplier, 1)), 0),
    NOW()
FROM usage_logs ul
WHERE ul.actual_cost > 0
GROUP BY 1, 2
ON CONFLICT (bucket_start, user_id) DO UPDATE SET
    consumed_revenue = EXCLUDED.consumed_revenue,
    supplier_cost = EXCLUDED.supplier_cost,
    first_used_at = EXCLUDED.first_used_at,
    last_used_at = EXCLUDED.last_used_at,
    updated_at = NOW();

INSERT INTO business_user_facts (
    user_id, first_activated_at, created_at, updated_at
)
SELECT fact.user_id, MIN(fact.first_used_at), MIN(u.created_at), NOW()
FROM business_usage_hourly_facts fact
JOIN users u ON u.id = fact.user_id
GROUP BY fact.user_id
ON CONFLICT (user_id) DO UPDATE SET
    first_activated_at = LEAST(
        COALESCE(business_user_facts.first_activated_at, EXCLUDED.first_activated_at),
        EXCLUDED.first_activated_at
    ),
    updated_at = NOW();

CREATE OR REPLACE FUNCTION business_refresh_user_attribution(target_user_id BIGINT)
RETURNS VOID AS $$
DECLARE
    registration_time TIMESTAMPTZ;
BEGIN
    SELECT created_at INTO registration_time FROM users WHERE id = target_user_id;
    INSERT INTO business_user_facts (
        user_id, channel, channel_ref_id, attribution_snapshot, created_at, updated_at
    )
    SELECT
        u.id,
        CASE WHEN dcb.user_id IS NOT NULL THEN 'distribution'
             WHEN ua.inviter_id IS NOT NULL THEN 'affiliate' ELSE 'unknown' END,
        COALESCE(dcb.agent_id, ua.inviter_id),
        CASE WHEN dcb.user_id IS NOT NULL THEN jsonb_build_object('source', 'distribution_binding', 'promotion_code', dcb.promotion_code)
             WHEN ua.inviter_id IS NOT NULL THEN jsonb_build_object('source', 'legacy_affiliate') ELSE '{}'::jsonb END,
        u.created_at,
        NOW()
    FROM users u
    LEFT JOIN distribution_customer_bindings dcb ON dcb.user_id = u.id
    LEFT JOIN user_affiliates ua ON ua.user_id = u.id
    WHERE u.id = target_user_id
    ON CONFLICT (user_id) DO UPDATE SET
        channel = EXCLUDED.channel,
        channel_ref_id = EXCLUDED.channel_ref_id,
        attribution_snapshot = EXCLUDED.attribution_snapshot,
        updated_at = NOW()
    -- Registration attribution is immutable after reliable evidence is stored.
    -- Current agent bindings remain available from distribution tables and must
    -- not rewrite historical registration, usage or revenue reports.
    WHERE business_user_facts.channel = 'unknown' AND EXCLUDED.channel <> 'unknown';
    IF FOUND THEN
        PERFORM business_invalidate_analytics(registration_time);
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION business_attribution_trigger()
RETURNS TRIGGER AS $$
DECLARE
    target_user_id BIGINT;
BEGIN
    target_user_id := CASE WHEN TG_OP = 'DELETE' THEN OLD.user_id ELSE NEW.user_id END;
    PERFORM business_refresh_user_attribution(target_user_id);
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS business_distribution_attribution_sync ON distribution_customer_bindings;
CREATE TRIGGER business_distribution_attribution_sync
AFTER INSERT OR UPDATE OR DELETE ON distribution_customer_bindings
FOR EACH ROW EXECUTE FUNCTION business_attribution_trigger();

DROP TRIGGER IF EXISTS business_affiliate_attribution_sync ON user_affiliates;
CREATE TRIGGER business_affiliate_attribution_sync
AFTER INSERT OR UPDATE OF inviter_id OR DELETE ON user_affiliates
FOR EACH ROW EXECUTE FUNCTION business_attribution_trigger();
