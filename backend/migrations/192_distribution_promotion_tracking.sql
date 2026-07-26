-- Final baseline for privacy-aware distribution promotion tracking,
-- registration attribution, exact archives, and exclusive promotion ownership.

ALTER TABLE distribution_settings
    ADD COLUMN IF NOT EXISTS promotion_tracking_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS promotion_attribution_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS promotion_attribution_days INTEGER NOT NULL DEFAULT 30
        CHECK (promotion_attribution_days BETWEEN 1 AND 365),
    ADD COLUMN IF NOT EXISTS promotion_attribution_model VARCHAR(16) NOT NULL DEFAULT 'first_touch'
        CHECK (promotion_attribution_model IN ('first_touch','last_touch')),
    ADD COLUMN IF NOT EXISTS promotion_collect_source BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS promotion_collect_device BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS promotion_bot_filter_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS promotion_detail_retention_days INTEGER NOT NULL DEFAULT 180
        CHECK (promotion_detail_retention_days BETWEEN 30 AND 730);

ALTER TABLE distribution_agents
    ADD COLUMN IF NOT EXISTS can_view_promotion_stats BOOLEAN NOT NULL DEFAULT TRUE;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM distribution_customer_bindings b
        JOIN user_affiliates a ON a.user_id=b.user_id
        WHERE a.inviter_id IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'existing users have both affiliate and distribution ownership; resolve them before migration 192';
    END IF;
END;
$$;

CREATE TABLE IF NOT EXISTS distribution_promotion_visits (
    id BIGSERIAL PRIMARY KEY,
    agent_id BIGINT NOT NULL REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    promotion_code VARCHAR(32) NOT NULL,
    visitor_token_hash VARCHAR(64) NOT NULL,
    ip_hash VARCHAR(64),
    landing_path VARCHAR(512) NOT NULL DEFAULT '/register',
    referrer_host VARCHAR(255),
    utm_source VARCHAR(128),
    utm_medium VARCHAR(128),
    utm_campaign VARCHAR(128),
    device_type VARCHAR(16) NOT NULL DEFAULT 'unknown'
        CHECK (device_type IN ('desktop','mobile','tablet','bot','unknown')),
    is_bot BOOLEAN NOT NULL DEFAULT FALSE,
    dedupe_bucket TIMESTAMPTZ,
    detail_expired_at TIMESTAMPTZ,
    visited_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_distribution_promotion_visits_dedupe
    ON distribution_promotion_visits(
        agent_id,visitor_token_hash,dedupe_bucket,is_bot,landing_path,
        COALESCE(referrer_host,''),COALESCE(utm_source,''),COALESCE(utm_medium,''),
        COALESCE(utm_campaign,''),device_type
    ) WHERE dedupe_bucket IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_visits_agent_time
    ON distribution_promotion_visits(agent_id,visited_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_visits_visitor_time
    ON distribution_promotion_visits(visitor_token_hash,visited_at DESC);
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_visits_time
    ON distribution_promotion_visits(visited_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_visits_agent_visitor_time
    ON distribution_promotion_visits(agent_id,visitor_token_hash,visited_at DESC);
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_visits_source_time
    ON distribution_promotion_visits(agent_id,utm_source,referrer_host,visited_at DESC);
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_visits_retention
    ON distribution_promotion_visits(visited_at,id) WHERE detail_expired_at IS NULL;

CREATE TABLE IF NOT EXISTS distribution_promotion_attributions (
    visitor_token_hash VARCHAR(64) PRIMARY KEY,
    agent_id BIGINT NOT NULL REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    promotion_code VARCHAR(32) NOT NULL,
    first_visit_id BIGINT NOT NULL REFERENCES distribution_promotion_visits(id) ON DELETE RESTRICT,
    last_visit_id BIGINT NOT NULL REFERENCES distribution_promotion_visits(id) ON DELETE RESTRICT,
    expires_at TIMESTAMPTZ NOT NULL,
    converted_user_id BIGINT UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_attributions_agent
    ON distribution_promotion_attributions(agent_id,expires_at DESC);
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_attributions_expires
    ON distribution_promotion_attributions(expires_at,visitor_token_hash);
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_attributions_first_visit
    ON distribution_promotion_attributions(first_visit_id);
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_attributions_last_visit
    ON distribution_promotion_attributions(last_visit_id);

CREATE TABLE IF NOT EXISTS distribution_promotion_conversions (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE RESTRICT,
    agent_id BIGINT NOT NULL REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    promotion_code VARCHAR(32) NOT NULL,
    visitor_token_hash VARCHAR(64),
    first_visit_id BIGINT REFERENCES distribution_promotion_visits(id) ON DELETE SET NULL,
    last_visit_id BIGINT REFERENCES distribution_promotion_visits(id) ON DELETE SET NULL,
    attribution_type VARCHAR(16) NOT NULL
        CHECK (attribution_type IN ('direct','persisted')),
    attribution_visit_id BIGINT REFERENCES distribution_promotion_visits(id) ON DELETE SET NULL,
    attribution_source VARCHAR(255) NOT NULL DEFAULT '直接注册',
    attribution_model VARCHAR(16) NOT NULL DEFAULT 'untracked'
        CHECK (attribution_model IN ('first_touch','last_touch','untracked')),
    attribution_device_type VARCHAR(16),
    attribution_is_bot BOOLEAN,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_conversions_agent_time
    ON distribution_promotion_conversions(agent_id,registered_at DESC,user_id DESC);
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_conversions_attributed_visit
    ON distribution_promotion_conversions(attribution_visit_id)
    WHERE attribution_visit_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_conversions_agent_source
    ON distribution_promotion_conversions(agent_id,attribution_source,registered_at DESC);
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_conversions_agent_visitor_registered
    ON distribution_promotion_conversions(agent_id,visitor_token_hash,registered_at)
    WHERE visitor_token_hash IS NOT NULL;

CREATE TABLE IF NOT EXISTS distribution_promotion_daily_rollups (
    stat_date DATE NOT NULL,
    agent_id BIGINT NOT NULL REFERENCES distribution_agents(id) ON DELETE RESTRICT,
    source VARCHAR(255) NOT NULL,
    device_type VARCHAR(16) NOT NULL,
    is_bot BOOLEAN NOT NULL,
    rollup_kind VARCHAR(16) NOT NULL DEFAULT 'visitor'
        CHECK (rollup_kind IN ('legacy','visitor')),
    visitor_token_hash VARCHAR(64),
    visits BIGINT NOT NULL DEFAULT 0,
    unique_visitors BIGINT NOT NULL DEFAULT 0,
    converted_visitors BIGINT NOT NULL DEFAULT 0,
    tracked_registrations BIGINT NOT NULL DEFAULT 0,
    direct_visits BIGINT NOT NULL DEFAULT 0,
    persisted_visits BIGINT NOT NULL DEFAULT 0,
    unregistered_visits BIGINT NOT NULL DEFAULT 0,
    credited_direct BOOLEAN NOT NULL DEFAULT FALSE,
    credited_persisted BOOLEAN NOT NULL DEFAULT FALSE,
    source_direct BOOLEAN NOT NULL DEFAULT FALSE,
    source_persisted BOOLEAN NOT NULL DEFAULT FALSE,
    credited_direct_registrations BIGINT NOT NULL DEFAULT 0,
    credited_persisted_registrations BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT distribution_promotion_rollups_visitor_hash_check CHECK (
        (rollup_kind='legacy' AND visitor_token_hash IS NULL)
        OR (rollup_kind='visitor' AND visitor_token_hash IS NOT NULL)
    )
);
CREATE INDEX IF NOT EXISTS idx_distribution_promotion_daily_rollups_agent_date
    ON distribution_promotion_daily_rollups(agent_id,stat_date DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_distribution_promotion_rollups_legacy_unique
    ON distribution_promotion_daily_rollups(stat_date,agent_id,source,device_type,is_bot)
    WHERE rollup_kind='legacy';
CREATE UNIQUE INDEX IF NOT EXISTS idx_distribution_promotion_rollups_visitor_unique
    ON distribution_promotion_daily_rollups(stat_date,agent_id,source,device_type,is_bot,visitor_token_hash)
    WHERE rollup_kind='visitor';

CREATE TABLE IF NOT EXISTS distribution_promotion_maintenance (
    id SMALLINT PRIMARY KEY CHECK (id=1),
    last_cleanup_at TIMESTAMPTZ NOT NULL DEFAULT '-infinity',
    last_privacy_scrub_at TIMESTAMPTZ NOT NULL DEFAULT '-infinity',
    last_archive_skipped_at TIMESTAMPTZ,
    last_archive_skipped_date DATE,
    last_archive_skip_reason VARCHAR(64),
    archive_skipped_count BIGINT NOT NULL DEFAULT 0
);
INSERT INTO distribution_promotion_maintenance(id) VALUES(1) ON CONFLICT(id) DO NOTHING;

CREATE TABLE IF NOT EXISTS distribution_promotion_archive_skips (
    stat_date DATE PRIMARY KEY,
    first_detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    detection_count BIGINT NOT NULL DEFAULT 1,
    active_attribution_count BIGINT NOT NULL DEFAULT 0,
    reason VARCHAR(64) NOT NULL
);

CREATE TABLE IF NOT EXISTS distribution_binding_claims (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    promotion_code VARCHAR(32) NOT NULL,
    signup_source VARCHAR(32) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending','completed','conflict')),
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts>=0),
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_distribution_binding_claims_pending
    ON distribution_binding_claims(updated_at,user_id) WHERE status='pending';

CREATE TABLE IF NOT EXISTS user_promotion_ownerships (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    owner_type VARCHAR(16) NOT NULL CHECK (owner_type IN ('affiliate','distribution')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
INSERT INTO user_promotion_ownerships(user_id,owner_type)
SELECT user_id,'distribution' FROM distribution_customer_bindings
ON CONFLICT(user_id) DO NOTHING;
INSERT INTO user_promotion_ownerships(user_id,owner_type)
SELECT user_id,'affiliate' FROM user_affiliates WHERE inviter_id IS NOT NULL
ON CONFLICT(user_id) DO NOTHING;

CREATE OR REPLACE FUNCTION enforce_user_promotion_ownership_exclusivity()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    requested_owner_type VARCHAR(16);
    persisted_owner_type VARCHAR(16);
BEGIN
    requested_owner_type := CASE
        WHEN TG_TABLE_NAME='distribution_customer_bindings' THEN 'distribution'
        ELSE 'affiliate'
    END;
    INSERT INTO user_promotion_ownerships(user_id,owner_type)
    VALUES(NEW.user_id,requested_owner_type)
    ON CONFLICT(user_id) DO UPDATE SET owner_type=user_promotion_ownerships.owner_type
    RETURNING owner_type INTO persisted_owner_type;
    IF persisted_owner_type<>requested_owner_type THEN
        RAISE EXCEPTION 'user % already has % promotion ownership', NEW.user_id, persisted_owner_type
            USING ERRCODE='23514';
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS distribution_customer_ownership_exclusivity ON distribution_customer_bindings;
CREATE TRIGGER distribution_customer_ownership_exclusivity
BEFORE INSERT OR UPDATE OF user_id ON distribution_customer_bindings
FOR EACH ROW EXECUTE FUNCTION enforce_user_promotion_ownership_exclusivity();
DROP TRIGGER IF EXISTS affiliate_inviter_ownership_exclusivity ON user_affiliates;
CREATE TRIGGER affiliate_inviter_ownership_exclusivity
BEFORE INSERT OR UPDATE OF inviter_id,user_id ON user_affiliates
FOR EACH ROW WHEN (NEW.inviter_id IS NOT NULL)
EXECUTE FUNCTION enforce_user_promotion_ownership_exclusivity();

CREATE OR REPLACE VIEW distribution_promotion_legacy_rollup_status AS
SELECT COUNT(*) legacy_rows,
       MIN(stat_date) first_legacy_date,
       MAX(stat_date) last_legacy_date,
       COALESCE(SUM(visits),0) legacy_visits
FROM distribution_promotion_daily_rollups
WHERE rollup_kind='legacy';

CREATE OR REPLACE FUNCTION scrub_distribution_promotion_details(batch_size INTEGER DEFAULT 1000)
RETURNS TABLE(scrubbed_rows INTEGER,expired_attributions INTEGER)
LANGUAGE plpgsql
AS $$
DECLARE
    detail_cutoff TIMESTAMPTZ;
    claimed BOOLEAN;
    scrubbed_count INTEGER := 0;
    attribution_count INTEGER := 0;
BEGIN
    IF batch_size<1 OR batch_size>10000 THEN
        RAISE EXCEPTION 'batch_size must be between 1 and 10000';
    END IF;
    UPDATE distribution_promotion_maintenance SET last_privacy_scrub_at=NOW()
    WHERE id=1 AND last_privacy_scrub_at<NOW()-INTERVAL '10 minutes'
    RETURNING TRUE INTO claimed;
    IF NOT COALESCE(claimed,FALSE) THEN
        RETURN QUERY SELECT 0,0;
        RETURN;
    END IF;
    SELECT NOW()-make_interval(days=>promotion_detail_retention_days)
    INTO detail_cutoff FROM distribution_settings WHERE id=1;
    DELETE FROM distribution_promotion_attributions
    WHERE visitor_token_hash IN (
        SELECT visitor_token_hash FROM distribution_promotion_attributions
        WHERE expires_at<NOW() ORDER BY expires_at,visitor_token_hash LIMIT batch_size
    );
    GET DIAGNOSTICS attribution_count=ROW_COUNT;
    WITH expiring AS (
        SELECT id FROM distribution_promotion_visits
        WHERE visited_at<detail_cutoff AND detail_expired_at IS NULL
        ORDER BY visited_at,id LIMIT batch_size
    )
    UPDATE distribution_promotion_visits v
    SET ip_hash=NULL,landing_path='/expired',utm_medium=NULL,utm_campaign=NULL,detail_expired_at=NOW()
    FROM expiring e WHERE v.id=e.id;
    GET DIAGNOSTICS scrubbed_count=ROW_COUNT;
    RETURN QUERY SELECT scrubbed_count,attribution_count;
END;
$$;

CREATE OR REPLACE FUNCTION cleanup_distribution_promotion_details(batch_size INTEGER DEFAULT 1000)
RETURNS TABLE(scrubbed_rows INTEGER,deleted_rows INTEGER,skipped_days INTEGER)
LANGUAGE plpgsql
AS $$
DECLARE
    detail_cutoff TIMESTAMPTZ;
    physical_cutoff TIMESTAMPTZ;
    claimed BOOLEAN;
    scrubbed_count INTEGER := 0;
    deleted_count INTEGER := 0;
    archive_date DATE;
    archive_start TIMESTAMPTZ;
    archive_end TIMESTAMPTZ;
    scan_from TIMESTAMPTZ := '-infinity';
    active_references BIGINT;
    archive_ids BIGINT[];
    skipped_count INTEGER := 0;
BEGIN
    IF batch_size<1 OR batch_size>10000 THEN
        RAISE EXCEPTION 'batch_size must be between 1 and 10000';
    END IF;
    UPDATE distribution_promotion_maintenance SET last_cleanup_at=NOW()
    WHERE id=1 AND last_cleanup_at<NOW()-INTERVAL '10 minutes'
    RETURNING TRUE INTO claimed;
    IF NOT COALESCE(claimed,FALSE) THEN
        RETURN QUERY SELECT 0,0,0;
        RETURN;
    END IF;
    SELECT NOW()-make_interval(days=>promotion_detail_retention_days),
           NOW()-make_interval(days=>GREATEST(promotion_detail_retention_days,367))
    INTO detail_cutoff,physical_cutoff FROM distribution_settings WHERE id=1;
    DELETE FROM distribution_promotion_attributions
    WHERE visitor_token_hash IN (
        SELECT visitor_token_hash FROM distribution_promotion_attributions
        WHERE expires_at<NOW() ORDER BY expires_at LIMIT batch_size
    );
    WITH expiring AS (
        SELECT id FROM distribution_promotion_visits
        WHERE visited_at<detail_cutoff AND detail_expired_at IS NULL
        ORDER BY visited_at,id LIMIT batch_size
    )
    UPDATE distribution_promotion_visits v
    SET ip_hash=NULL,landing_path='/expired',utm_medium=NULL,utm_campaign=NULL,detail_expired_at=NOW()
    FROM expiring e WHERE v.id=e.id;
    GET DIAGNOSTICS scrubbed_count=ROW_COUNT;

    LOOP
        SELECT (v.visited_at AT TIME ZONE 'Asia/Hong_Kong')::date INTO archive_date
        FROM distribution_promotion_visits v
        WHERE v.visited_at<((physical_cutoff AT TIME ZONE 'Asia/Hong_Kong')::date AT TIME ZONE 'Asia/Hong_Kong')
          AND v.visited_at>=scan_from
        ORDER BY v.visited_at,v.id LIMIT 1;
        EXIT WHEN archive_date IS NULL;
        archive_start := archive_date::timestamp AT TIME ZONE 'Asia/Hong_Kong';
        archive_end := (archive_date+1)::timestamp AT TIME ZONE 'Asia/Hong_Kong';
        SELECT COUNT(*) INTO active_references
        FROM distribution_promotion_attributions a
        LEFT JOIN distribution_promotion_visits first_visit ON first_visit.id=a.first_visit_id
        LEFT JOIN distribution_promotion_visits last_visit ON last_visit.id=a.last_visit_id
        WHERE a.expires_at>NOW()
          AND ((first_visit.visited_at>=archive_start AND first_visit.visited_at<archive_end)
            OR (last_visit.visited_at>=archive_start AND last_visit.visited_at<archive_end));
        IF active_references=0 THEN EXIT; END IF;
        INSERT INTO distribution_promotion_archive_skips(stat_date,active_attribution_count,reason)
        VALUES(archive_date,active_references,'active_attribution')
        ON CONFLICT(stat_date) DO UPDATE SET
            last_detected_at=NOW(),
            detection_count=distribution_promotion_archive_skips.detection_count+1,
            active_attribution_count=EXCLUDED.active_attribution_count,
            reason=EXCLUDED.reason;
        UPDATE distribution_promotion_maintenance SET
            last_archive_skipped_at=NOW(),last_archive_skipped_date=archive_date,
            last_archive_skip_reason='active_attribution',archive_skipped_count=archive_skipped_count+1
        WHERE id=1;
        skipped_count := skipped_count+1;
        scan_from := archive_end;
        archive_date := NULL;
    END LOOP;

    IF archive_date IS NOT NULL THEN
        SELECT ARRAY(
            SELECT id FROM distribution_promotion_visits
            WHERE visited_at>=archive_start AND visited_at<archive_end
            ORDER BY visited_at,id LIMIT batch_size
        ) INTO archive_ids;
        WITH visit_facts AS (
            SELECT v.*,
                COALESCE(NULLIF(v.utm_source,''),NULLIF(v.referrer_host,''),'直接访问') fact_source,
                EXISTS(SELECT 1 FROM distribution_promotion_conversions c
                    WHERE c.agent_id=v.agent_id AND c.visitor_token_hash=v.visitor_token_hash
                      AND v.visited_at<=c.registered_at AND c.attribution_type='direct') has_direct,
                EXISTS(SELECT 1 FROM distribution_promotion_conversions c
                    WHERE c.agent_id=v.agent_id AND c.visitor_token_hash=v.visitor_token_hash
                      AND v.visited_at<=c.registered_at AND c.attribution_type='persisted') has_persisted,
                EXISTS(SELECT 1 FROM distribution_promotion_conversions c
                    WHERE c.agent_id=v.agent_id AND c.attribution_visit_id=v.id AND c.attribution_type='direct') credited_direct_visit,
                EXISTS(SELECT 1 FROM distribution_promotion_conversions c
                    WHERE c.agent_id=v.agent_id AND c.attribution_visit_id=v.id AND c.attribution_type='persisted') credited_persisted_visit,
                EXISTS(SELECT 1 FROM distribution_promotion_conversions c
                    WHERE c.agent_id=v.agent_id AND c.visitor_token_hash=v.visitor_token_hash
                      AND v.visited_at<=c.registered_at AND c.attribution_type='direct'
                      AND LOWER(c.attribution_source)=LOWER(COALESCE(NULLIF(v.utm_source,''),NULLIF(v.referrer_host,''),'直接访问'))) source_direct_visit,
                EXISTS(SELECT 1 FROM distribution_promotion_conversions c
                    WHERE c.agent_id=v.agent_id AND c.visitor_token_hash=v.visitor_token_hash
                      AND v.visited_at<=c.registered_at AND c.attribution_type='persisted'
                      AND LOWER(c.attribution_source)=LOWER(COALESCE(NULLIF(v.utm_source,''),NULLIF(v.referrer_host,''),'直接访问'))) source_persisted_visit,
                (SELECT COUNT(*) FROM distribution_promotion_conversions c
                    WHERE c.agent_id=v.agent_id AND c.attribution_visit_id=v.id AND c.attribution_type='direct') credited_direct_count,
                (SELECT COUNT(*) FROM distribution_promotion_conversions c
                    WHERE c.agent_id=v.agent_id AND c.attribution_visit_id=v.id AND c.attribution_type='persisted') credited_persisted_count
            FROM distribution_promotion_visits v WHERE v.id=ANY(archive_ids)
        )
        INSERT INTO distribution_promotion_daily_rollups(
            stat_date,agent_id,source,device_type,is_bot,rollup_kind,visitor_token_hash,
            visits,unique_visitors,converted_visitors,tracked_registrations,
            direct_visits,persisted_visits,unregistered_visits,
            credited_direct,credited_persisted,source_direct,source_persisted,
            credited_direct_registrations,credited_persisted_registrations
        )
        SELECT archive_date,agent_id,fact_source,device_type,is_bot,'visitor',visitor_token_hash,
            COUNT(*),1,
            CASE WHEN BOOL_OR(credited_direct_visit OR credited_persisted_visit) THEN 1 ELSE 0 END,
            SUM(credited_direct_count+credited_persisted_count),
            COUNT(*) FILTER(WHERE has_direct),COUNT(*) FILTER(WHERE has_persisted),
            COUNT(*) FILTER(WHERE NOT has_direct AND NOT has_persisted),
            BOOL_OR(credited_direct_visit),BOOL_OR(credited_persisted_visit),
            BOOL_OR(source_direct_visit),BOOL_OR(source_persisted_visit),
            SUM(credited_direct_count),SUM(credited_persisted_count)
        FROM visit_facts
        GROUP BY agent_id,fact_source,device_type,is_bot,visitor_token_hash
        ON CONFLICT(stat_date,agent_id,source,device_type,is_bot,visitor_token_hash)
            WHERE rollup_kind='visitor'
        DO UPDATE SET
            visits=distribution_promotion_daily_rollups.visits+EXCLUDED.visits,
            unique_visitors=1,
            converted_visitors=CASE WHEN distribution_promotion_daily_rollups.converted_visitors>0
                OR EXCLUDED.converted_visitors>0 THEN 1 ELSE 0 END,
            tracked_registrations=distribution_promotion_daily_rollups.tracked_registrations+EXCLUDED.tracked_registrations,
            direct_visits=distribution_promotion_daily_rollups.direct_visits+EXCLUDED.direct_visits,
            persisted_visits=distribution_promotion_daily_rollups.persisted_visits+EXCLUDED.persisted_visits,
            unregistered_visits=distribution_promotion_daily_rollups.unregistered_visits+EXCLUDED.unregistered_visits,
            credited_direct=distribution_promotion_daily_rollups.credited_direct OR EXCLUDED.credited_direct,
            credited_persisted=distribution_promotion_daily_rollups.credited_persisted OR EXCLUDED.credited_persisted,
            source_direct=distribution_promotion_daily_rollups.source_direct OR EXCLUDED.source_direct,
            source_persisted=distribution_promotion_daily_rollups.source_persisted OR EXCLUDED.source_persisted,
            credited_direct_registrations=distribution_promotion_daily_rollups.credited_direct_registrations+EXCLUDED.credited_direct_registrations,
            credited_persisted_registrations=distribution_promotion_daily_rollups.credited_persisted_registrations+EXCLUDED.credited_persisted_registrations;

        DELETE FROM distribution_promotion_attributions a
        WHERE a.expires_at<=NOW() AND (a.first_visit_id=ANY(archive_ids) OR a.last_visit_id=ANY(archive_ids));
        UPDATE distribution_promotion_conversions c
        SET first_visit_id=CASE WHEN c.first_visit_id=ANY(archive_ids) THEN NULL ELSE c.first_visit_id END,
            last_visit_id=CASE WHEN c.last_visit_id=ANY(archive_ids) THEN NULL ELSE c.last_visit_id END,
            attribution_visit_id=CASE WHEN c.attribution_visit_id=ANY(archive_ids) THEN NULL ELSE c.attribution_visit_id END
        WHERE c.first_visit_id=ANY(archive_ids) OR c.last_visit_id=ANY(archive_ids) OR c.attribution_visit_id=ANY(archive_ids);
        UPDATE distribution_promotion_conversions c SET visitor_token_hash=NULL
        WHERE c.visitor_token_hash IS NOT NULL
          AND EXISTS(SELECT 1 FROM distribution_promotion_visits archived_visit
              WHERE archived_visit.id=ANY(archive_ids) AND archived_visit.agent_id=c.agent_id
                AND archived_visit.visitor_token_hash=c.visitor_token_hash)
          AND NOT EXISTS(SELECT 1 FROM distribution_promotion_visits remaining_visit
              WHERE remaining_visit.agent_id=c.agent_id AND remaining_visit.visitor_token_hash=c.visitor_token_hash
                AND NOT (remaining_visit.id=ANY(archive_ids)));
        DELETE FROM distribution_promotion_visits WHERE id=ANY(archive_ids);
        GET DIAGNOSTICS deleted_count=ROW_COUNT;
        IF NOT EXISTS(SELECT 1 FROM distribution_promotion_visits
            WHERE visited_at>=archive_start AND visited_at<archive_end) THEN
            DELETE FROM distribution_promotion_archive_skips WHERE stat_date=archive_date;
        END IF;
    END IF;
    RETURN QUERY SELECT scrubbed_count,deleted_count,skipped_count;
END;
$$;

COMMENT ON TABLE distribution_promotion_visits IS
    'Anonymous promotion visits; visitor tokens and client IP addresses are stored only as hashes.';
COMMENT ON COLUMN distribution_promotion_conversions.attribution_visit_id IS
    'The first-touch or last-touch visit credited for this conversion and owned by agent_id.';
COMMENT ON COLUMN distribution_promotion_conversions.attribution_source IS
    'Immutable source snapshot retained after anonymous visit detail expires.';
COMMENT ON TABLE distribution_promotion_daily_rollups IS
    'Exact pseudonymous visitor-day facts retained after raw visit deletion; legacy rows are rejected by exact analytics.';
COMMENT ON TABLE distribution_promotion_archive_skips IS
    'Cleanup exceptions for days retained because active attribution still references raw visits.';
COMMENT ON TABLE distribution_binding_claims IS
    'Durable registration-time distribution ownership claims retried on successful authentication.';
COMMENT ON TABLE user_promotion_ownerships IS
    'Immutable registration-time arbiter preventing concurrent affiliate and distribution ownership.';
COMMENT ON FUNCTION scrub_distribution_promotion_details(INTEGER) IS
    'Always-on bounded privacy scrub that never physically deletes promotion visits.';
COMMENT ON FUNCTION cleanup_distribution_promotion_details(INTEGER) IS
    'Bounded exact archive and physical cleanup for promotion visits beyond the analytics window.';
