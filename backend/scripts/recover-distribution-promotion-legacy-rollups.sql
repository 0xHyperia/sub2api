-- Run after restoring distribution_promotion_visits and
-- distribution_promotion_conversions from a pre-cleanup backup. The script
-- removes legacy aggregates only when the restored raw rows reproduce every
-- legacy group exactly. It is safe to rerun.
BEGIN;

DO $$
DECLARE
    legacy_count BIGINT;
    mismatch_count BIGINT;
BEGIN
    SELECT COUNT(*) INTO legacy_count
    FROM distribution_promotion_daily_rollups WHERE rollup_kind='legacy';
    IF legacy_count=0 THEN
        RAISE NOTICE 'no legacy promotion rollups require recovery';
        RETURN;
    END IF;

    WITH raw AS (
        SELECT (v.visited_at AT TIME ZONE 'Asia/Hong_Kong')::date stat_date,
               v.agent_id,
               COALESCE(NULLIF(v.utm_source,''),NULLIF(v.referrer_host,''),'直接访问') source,
               v.device_type,v.is_bot,
               COUNT(*) visits,
               COUNT(DISTINCT v.visitor_token_hash) unique_visitors,
               COUNT(DISTINCT c.visitor_token_hash) FILTER(WHERE c.user_id IS NOT NULL) converted_visitors,
               COUNT(DISTINCT c.user_id) tracked_registrations
        FROM distribution_promotion_visits v
        LEFT JOIN distribution_promotion_conversions c
          ON c.attribution_visit_id=v.id AND c.agent_id=v.agent_id
        WHERE (v.visited_at AT TIME ZONE 'Asia/Hong_Kong')::date IN (
            SELECT stat_date FROM distribution_promotion_daily_rollups WHERE rollup_kind='legacy'
        )
        GROUP BY 1,2,3,4,5
    ), legacy AS (
        SELECT stat_date,agent_id,source,device_type,is_bot,visits,unique_visitors,converted_visitors,tracked_registrations
        FROM distribution_promotion_daily_rollups WHERE rollup_kind='legacy'
    )
    SELECT COUNT(*) INTO mismatch_count
    FROM legacy l FULL OUTER JOIN raw r
      USING(stat_date,agent_id,source,device_type,is_bot)
    WHERE l.stat_date IS NULL OR r.stat_date IS NULL
       OR l.visits<>r.visits OR l.unique_visitors<>r.unique_visitors
       OR l.converted_visitors<>r.converted_visitors
       OR l.tracked_registrations<>r.tracked_registrations;

    IF mismatch_count>0 THEN
        RAISE EXCEPTION 'legacy promotion recovery refused: % aggregate groups do not match restored raw data', mismatch_count;
    END IF;

    DELETE FROM distribution_promotion_daily_rollups WHERE rollup_kind='legacy';
    RAISE NOTICE 'legacy promotion rollups recovered from exact raw data';
END;
$$;

COMMIT;
