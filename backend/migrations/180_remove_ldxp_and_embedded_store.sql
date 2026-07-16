-- Permanently remove the retired LDXP card-purchase integration and the
-- legacy embedded-store settings. Fulfilled balances/subscriptions are not
-- reversed; only integration-owned records and configuration are deleted.

DELETE FROM payment_audit_logs
WHERE order_id IN (
    SELECT id::text
    FROM payment_orders
    WHERE order_type = 'card'
       OR payment_type = 'ldxp'
       OR provider_key = 'ldxp'
);

DELETE FROM payment_orders
WHERE order_type = 'card'
   OR payment_type = 'ldxp'
   OR provider_key = 'ldxp';

DELETE FROM payment_provider_instances
WHERE provider_key = 'ldxp';

DELETE FROM settings
WHERE lower(key) IN (
    'payment_card_enabled',
    'purchase_subscription_enabled',
    'purchase_subscription_url'
);

DO $$
DECLARE
    menu_value TEXT;
BEGIN
    SELECT value INTO menu_value
    FROM settings
    WHERE lower(key) = 'custom_menu_items'
    LIMIT 1;

    IF menu_value IS NULL OR btrim(menu_value) = '' THEN
        RETURN;
    END IF;

    BEGIN
        UPDATE settings
        SET value = COALESCE((
            SELECT jsonb_agg(item ORDER BY ordinal)
            FROM jsonb_array_elements(menu_value::jsonb) WITH ORDINALITY AS entries(item, ordinal)
            WHERE item ->> 'id' <> 'migrated_purchase_subscription'
        ), '[]'::jsonb)::text
        WHERE lower(key) = 'custom_menu_items';
    EXCEPTION
        WHEN invalid_text_representation OR data_exception THEN
            RAISE NOTICE 'Skipping invalid custom_menu_items JSON while removing embedded store entry';
    END;
END $$;
