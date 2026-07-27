ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS provider_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS fee_mode VARCHAR(20) NOT NULL DEFAULT 'platform';

UPDATE payment_orders
SET provider_amount = pay_amount
WHERE provider_amount = 0;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'payment_orders_fee_mode_check'
          AND conrelid = 'payment_orders'::regclass
    ) THEN
        ALTER TABLE payment_orders
            ADD CONSTRAINT payment_orders_fee_mode_check
            CHECK (fee_mode IN ('platform', 'provider', 'merchant'));
    END IF;
END $$;
