-- 统一充值优惠：快捷充值金额（QUICK_RECHARGE_AMOUNTS）不再携带固定赠额，
-- 所有赠送/折扣只由充值优惠阶梯（RECHARGE_BONUS_TIERS）表达。
--
-- 1. 仅当阶梯为空且存在带赠额的快捷金额时，把旧赠额折算成赠金模式阶梯：
--    bonus_percent = bonus / (amount * 充值倍率) * 100（保留两位小数）。
--    阶梯按「不低于该金额」命中，因此在存在赠额档位时，其余预设金额以 0% 档位截断继承；
--    预设超过 20 个（阶梯上限）时只折算带赠额的档位。
-- 2. 无论是否折算，都从快捷金额 JSON 中移除 bonus 字段。
-- 已配置阶梯的站点保持现有阶梯不变，旧赠额直接丢弃。可重入：移除 bonus 后不再触发折算。
DO $$
DECLARE
    quick_raw text;
    quick jsonb;
    tiers_raw text;
    mult_raw text;
    mult numeric := 1;
    tiers_empty boolean := true;
    has_bonus boolean;
    preset_count integer;
    converted jsonb;
BEGIN
    SELECT value INTO quick_raw FROM settings WHERE key = 'QUICK_RECHARGE_AMOUNTS';
    IF quick_raw IS NULL OR btrim(quick_raw) = '' THEN
        RETURN;
    END IF;
    BEGIN
        quick := quick_raw::jsonb;
    EXCEPTION WHEN others THEN
        RETURN;
    END;
    IF jsonb_typeof(quick) <> 'array' THEN
        RETURN;
    END IF;

    SELECT value INTO tiers_raw FROM settings WHERE key = 'RECHARGE_BONUS_TIERS';
    IF tiers_raw IS NOT NULL AND btrim(tiers_raw) <> '' THEN
        BEGIN
            tiers_empty := jsonb_typeof(tiers_raw::jsonb) <> 'array' OR jsonb_array_length(tiers_raw::jsonb) = 0;
        EXCEPTION WHEN others THEN
            tiers_empty := true;
        END;
    END IF;

    SELECT value INTO mult_raw FROM settings WHERE key = 'BALANCE_RECHARGE_MULTIPLIER';
    IF mult_raw IS NOT NULL AND btrim(mult_raw) ~ '^[0-9]+(\.[0-9]+)?$' THEN
        mult := btrim(mult_raw)::numeric;
        IF mult <= 0 THEN
            mult := 1;
        END IF;
    END IF;

    SELECT EXISTS (
        SELECT 1 FROM jsonb_array_elements(quick) AS e
        WHERE jsonb_typeof(e) = 'object'
          AND COALESCE((e ->> 'bonus')::numeric, 0) > 0
          AND COALESCE((e ->> 'amount')::numeric, 0) > 0
    ) INTO has_bonus;
    SELECT count(*) INTO preset_count FROM jsonb_array_elements(quick) AS e WHERE jsonb_typeof(e) = 'object';

    IF has_bonus AND tiers_empty THEN
        SELECT jsonb_agg(
                   jsonb_build_object(
                       'min_amount', (e ->> 'amount')::numeric,
                       'bonus_percent', LEAST(1000, round(COALESCE((e ->> 'bonus')::numeric, 0)
                           / ((e ->> 'amount')::numeric * mult) * 100, 2))
                   )
                   ORDER BY (e ->> 'amount')::numeric
               )
        INTO converted
        FROM jsonb_array_elements(quick) AS e
        WHERE jsonb_typeof(e) = 'object'
          AND COALESCE((e ->> 'amount')::numeric, 0) > 0
          AND (preset_count <= 20 OR COALESCE((e ->> 'bonus')::numeric, 0) > 0);

        IF converted IS NOT NULL AND jsonb_array_length(converted) > 0 THEN
            INSERT INTO settings (key, value, updated_at)
            VALUES ('RECHARGE_BONUS_TIERS', converted::text, NOW())
            ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();
            INSERT INTO settings (key, value, updated_at)
            VALUES ('RECHARGE_BONUS_MODE', 'bonus', NOW())
            ON CONFLICT (key) DO UPDATE SET value = 'bonus', updated_at = NOW();
        END IF;
    END IF;

    UPDATE settings
    SET value = COALESCE((
            SELECT jsonb_agg(
                       CASE WHEN jsonb_typeof(e) = 'object' THEN e - 'bonus' ELSE e END
                       ORDER BY ord
                   )
            FROM jsonb_array_elements(quick) WITH ORDINALITY AS t(e, ord)
        ), '[]'::jsonb)::text,
        updated_at = NOW()
    WHERE key = 'QUICK_RECHARGE_AMOUNTS';
END $$;
