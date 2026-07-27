ALTER TABLE payment_provider_instances
    ADD COLUMN IF NOT EXISTS fee_rates TEXT NOT NULL DEFAULT '';

WITH configured_rates AS (
    SELECT
        COALESCE(
            (SELECT value::numeric FROM settings WHERE key = 'RECHARGE_FEE_RATE' AND value ~ '^[0-9]+([.][0-9]+)?$'),
            0
        )::double precision AS default_rate,
        COALESCE(
            (SELECT value::numeric FROM settings WHERE key = 'ALIPAY_RECHARGE_FEE_RATE' AND value ~ '^[0-9]+([.][0-9]+)?$'),
            3
        )::double precision AS alipay_rate,
        COALESCE(
            (SELECT value::numeric FROM settings WHERE key = 'WXPAY_RECHARGE_FEE_RATE' AND value ~ '^[0-9]+([.][0-9]+)?$'),
            3.8
        )::double precision AS wxpay_rate
)
UPDATE payment_provider_instances AS instance
SET fee_rates = (
    SELECT COALESCE(jsonb_object_agg(method_key, fee_rate), '{}'::jsonb)::text AS fee_rates
    FROM (
        SELECT DISTINCT
            CASE
                WHEN instance.provider_key = 'stripe' THEN 'stripe'
                WHEN trim(method) IN ('alipay', 'alipay_direct') THEN 'alipay'
                WHEN trim(method) IN ('wxpay', 'wxpay_direct') THEN 'wxpay'
                ELSE trim(method)
            END AS method_key,
            CASE
                WHEN trim(method) IN ('alipay', 'alipay_direct') THEN configured_rates.alipay_rate
                WHEN trim(method) IN ('wxpay', 'wxpay_direct') THEN configured_rates.wxpay_rate
                ELSE configured_rates.default_rate
            END AS fee_rate
        FROM regexp_split_to_table(
            CASE WHEN instance.provider_key = 'stripe' THEN 'stripe' ELSE instance.supported_types END,
            ','
        ) AS method
        WHERE trim(method) <> ''
    ) AS method_rates
)
FROM configured_rates
WHERE trim(instance.fee_rates) = '';
