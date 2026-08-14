-- New orders persist each business amount explicitly. Existing rows retain the
-- zero defaults and continue to use the legacy application fallbacks; this
-- migration intentionally performs no historical backfill or reconciliation.
ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS payment_principal_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS entitlement_principal_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS surcharge_amount DECIMAL(20,2) NOT NULL DEFAULT 0;

-- Keep realtime business facts on the same principal basis as the application
-- aggregator. The zero fallback deliberately retains legacy order semantics.
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
    IF NEW.status NOT IN (
        'PAID', 'RECHARGING', 'COMPLETED', 'FAILED',
        'REFUND_REQUESTED', 'REFUNDING', 'REFUND_PENDING',
        'REFUND_FAILED', 'PARTIALLY_REFUNDED', 'REFUNDED'
    ) OR NEW.paid_at IS NULL THEN
        DELETE FROM business_payment_facts WHERE payment_order_id = NEW.id;
        RETURN NEW;
    END IF;

    payment_currency := UPPER(COALESCE(NULLIF(NEW.provider_snapshot->>'currency', ''), 'CNY'));
    original_amount := COALESCE(NULLIF(NEW.payment_principal_amount, 0), NULLIF(NEW.provider_amount, 0), NEW.pay_amount, 0);
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
        fx_rate_to_cny = COALESCE(business_payment_facts.fx_rate_to_cny, EXCLUDED.fx_rate_to_cny),
        fx_estimated = business_payment_facts.fx_estimated OR EXCLUDED.fx_estimated,
        paid_at = EXCLUDED.paid_at,
        refunded_at = EXCLUDED.refunded_at,
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

DROP TRIGGER IF EXISTS business_payment_fact_sync ON payment_orders;
CREATE TRIGGER business_payment_fact_sync
AFTER INSERT OR UPDATE OF status, paid_at, refund_amount, refund_at,
    payment_principal_amount, provider_amount, pay_amount, provider_snapshot
ON payment_orders FOR EACH ROW EXECUTE FUNCTION business_sync_payment_fact();
