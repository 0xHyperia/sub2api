package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *distributionRepository) PromotionTrackingEnabled(ctx context.Context) (bool, error) {
	var enabled bool
	if err := r.db.QueryRowContext(ctx, `SELECT enabled AND promotion_tracking_enabled FROM distribution_settings WHERE id=1`).Scan(&enabled); err != nil {
		return false, fmt.Errorf("get distribution promotion tracking status: %w", err)
	}
	return enabled, nil
}

func (r *distributionRepository) TrackPromotionVisit(ctx context.Context, input service.DistributionPromotionVisitInput, visitorTokenHash, ipHash, deviceType string, isBot bool) (*service.DistributionPromotionVisitResult, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var tracking, attribution, collectSource, collectDevice bool
	var days int
	var model string
	if err = tx.QueryRowContext(ctx, `SELECT enabled AND promotion_tracking_enabled,promotion_attribution_enabled,
		promotion_collect_source,promotion_collect_device,promotion_attribution_days,promotion_attribution_model
		FROM distribution_settings WHERE id=1`).Scan(&tracking, &attribution, &collectSource, &collectDevice, &days, &model); err != nil {
		return nil, err
	}
	result := &service.DistributionPromotionVisitResult{Tracked: false, AttributionDays: days}
	if !tracking {
		return result, nil
	}

	var agentID int64
	err = tx.QueryRowContext(ctx, `SELECT a.id FROM distribution_agents a
		JOIN distribution_agent_levels l ON l.id=a.level_id AND l.is_active=TRUE
		LEFT JOIN distribution_agents p ON p.id=a.parent_agent_id
		WHERE a.promotion_code=$1 AND a.status='active' AND (l.depth=1 OR p.status='active')`, input.PromotionCode).Scan(&agentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrDistributionCodeInvalid
	}
	if err != nil {
		return nil, err
	}
	if !collectSource {
		input.Referrer, input.UTMSource, input.UTMMedium, input.UTMCampaign = "", "", "", ""
	}
	if !collectDevice {
		deviceType = "unknown"
	}
	var dedupeBucket time.Time
	if err = tx.QueryRowContext(ctx, `SELECT date_trunc('minute',CURRENT_TIMESTAMP)`).Scan(&dedupeBucket); err != nil {
		return nil, err
	}
	var visitID int64
	err = tx.QueryRowContext(ctx, `INSERT INTO distribution_promotion_visits
		(agent_id,promotion_code,visitor_token_hash,ip_hash,landing_path,referrer_host,utm_source,utm_medium,utm_campaign,device_type,is_bot,dedupe_bucket)
		VALUES($1,$2,$3,NULLIF($4,''),$5,NULLIF($6,''),NULLIF($7,''),NULLIF($8,''),NULLIF($9,''),$10,$11,$12)
		ON CONFLICT DO NOTHING
		RETURNING id`, agentID, input.PromotionCode, visitorTokenHash, ipHash, input.LandingPath, input.Referrer, input.UTMSource,
		input.UTMMedium, input.UTMCampaign, deviceType, isBot, dedupeBucket).Scan(&visitID)
	deduplicated := false
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `SELECT id FROM distribution_promotion_visits
			WHERE agent_id=$1 AND visitor_token_hash=$2 AND dedupe_bucket=$3 AND is_bot=$4
			  AND landing_path=$5 AND referrer_host IS NOT DISTINCT FROM NULLIF($6,'')
			  AND utm_source IS NOT DISTINCT FROM NULLIF($7,'') AND utm_medium IS NOT DISTINCT FROM NULLIF($8,'')
			  AND utm_campaign IS NOT DISTINCT FROM NULLIF($9,'') AND device_type=$10
			ORDER BY id DESC LIMIT 1`, agentID, visitorTokenHash, dedupeBucket, isBot, input.LandingPath,
			input.Referrer, input.UTMSource, input.UTMMedium, input.UTMCampaign, deviceType).Scan(&visitID)
		deduplicated = err == nil
	}
	if err != nil {
		return nil, err
	}
	if attribution && !isBot {
		_, err = tx.ExecContext(ctx, `INSERT INTO distribution_promotion_attributions
			(visitor_token_hash,agent_id,promotion_code,first_visit_id,last_visit_id,expires_at)
			VALUES($1,$2,$3,$4,$4,NOW()+($5::text||' days')::interval)
			ON CONFLICT(visitor_token_hash) DO UPDATE SET
				agent_id=CASE WHEN $6='last_touch' OR distribution_promotion_attributions.expires_at<=NOW()
					THEN EXCLUDED.agent_id ELSE distribution_promotion_attributions.agent_id END,
				promotion_code=CASE WHEN $6='last_touch' OR distribution_promotion_attributions.expires_at<=NOW()
					THEN EXCLUDED.promotion_code ELSE distribution_promotion_attributions.promotion_code END,
				first_visit_id=CASE WHEN distribution_promotion_attributions.expires_at<=NOW()
						OR ($6='last_touch' AND distribution_promotion_attributions.agent_id<>EXCLUDED.agent_id)
					THEN EXCLUDED.first_visit_id ELSE distribution_promotion_attributions.first_visit_id END,
				last_visit_id=CASE WHEN $6='first_touch' AND distribution_promotion_attributions.expires_at>NOW()
						AND distribution_promotion_attributions.agent_id<>EXCLUDED.agent_id
					THEN distribution_promotion_attributions.last_visit_id ELSE EXCLUDED.last_visit_id END,
				expires_at=CASE WHEN $6='last_touch' OR distribution_promotion_attributions.expires_at<=NOW()
					THEN EXCLUDED.expires_at ELSE distribution_promotion_attributions.expires_at END,
				updated_at=NOW()
			WHERE distribution_promotion_attributions.converted_user_id IS NULL`,
			visitorTokenHash, agentID, input.PromotionCode, visitID, days, model)
		if err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	result.Tracked = true
	result.Deduplicated = deduplicated
	return result, nil
}

func (r *distributionRepository) CleanupPromotionDetails(ctx context.Context, batchSize int) (service.DistributionPromotionCleanupResult, error) {
	var result service.DistributionPromotionCleanupResult
	if err := r.db.QueryRowContext(ctx, `SELECT scrubbed_rows,deleted_rows,skipped_days FROM cleanup_distribution_promotion_details($1)`, batchSize).
		Scan(&result.ScrubbedRows, &result.DeletedRows, &result.SkippedDays); err != nil {
		return service.DistributionPromotionCleanupResult{}, fmt.Errorf("cleanup distribution promotion details: %w", err)
	}
	return result, nil
}

func (r *distributionRepository) ScrubPromotionDetails(ctx context.Context, batchSize int) (service.DistributionPromotionPrivacyScrubResult, error) {
	var result service.DistributionPromotionPrivacyScrubResult
	if err := r.db.QueryRowContext(ctx, `SELECT scrubbed_rows,expired_attributions FROM scrub_distribution_promotion_details($1)`, batchSize).
		Scan(&result.ScrubbedRows, &result.ExpiredAttributions); err != nil {
		return service.DistributionPromotionPrivacyScrubResult{}, fmt.Errorf("scrub distribution promotion details: %w", err)
	}
	return result, nil
}

func (r *distributionRepository) ResolvePromotionAttribution(ctx context.Context, visitorTokenHash string, explicitCode ...string) (*service.DistributionRegistrationAttribution, error) {
	var out service.DistributionRegistrationAttribution
	code := ""
	if len(explicitCode) > 0 {
		code = explicitCode[0]
	}
	err := r.db.QueryRowContext(ctx, `WITH candidates AS (
		SELECT p.promotion_code,p.visitor_token_hash,p.agent_id,p.first_visit_id,p.last_visit_id,
			CASE WHEN s.promotion_attribution_model='last_touch' THEN p.last_visit_id ELSE p.first_visit_id END attribution_visit_id,
			s.promotion_attribution_model attribution_model
		FROM distribution_promotion_attributions p
		JOIN distribution_settings s ON s.id=1 AND s.enabled=TRUE AND s.promotion_attribution_enabled=TRUE
		JOIN distribution_agents a ON a.id=p.agent_id AND a.status='active'
		JOIN distribution_agent_levels l ON l.id=a.level_id AND l.is_active=TRUE
		LEFT JOIN distribution_agents parent ON parent.id=a.parent_agent_id
		WHERE p.visitor_token_hash=$1 AND p.converted_user_id IS NULL AND p.expires_at>NOW()
		  AND ($2='' OR p.promotion_code=$2) AND (l.depth=1 OR parent.status='active')
		UNION ALL
		SELECT v.promotion_code,v.visitor_token_hash,v.agent_id,
			FIRST_VALUE(v.id) OVER (ORDER BY v.visited_at,v.id ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING),
			FIRST_VALUE(v.id) OVER (ORDER BY v.visited_at DESC,v.id DESC ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING),
			CASE WHEN s.promotion_attribution_model='last_touch'
				THEN FIRST_VALUE(v.id) OVER (ORDER BY v.visited_at DESC,v.id DESC ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING)
				ELSE FIRST_VALUE(v.id) OVER (ORDER BY v.visited_at,v.id ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING) END,
			s.promotion_attribution_model
		FROM distribution_promotion_visits v
		JOIN distribution_settings s ON s.id=1 AND s.enabled=TRUE
		JOIN distribution_agents a ON a.id=v.agent_id AND a.status='active'
		JOIN distribution_agent_levels l ON l.id=a.level_id AND l.is_active=TRUE
		LEFT JOIN distribution_agents parent ON parent.id=a.parent_agent_id
		WHERE $2<>'' AND v.visitor_token_hash=$1 AND v.promotion_code=$2 AND NOT v.is_bot
		  AND v.visited_at > NOW() - make_interval(days => s.promotion_attribution_days)
		  AND (l.depth=1 OR parent.status='active')
	)
	SELECT c.promotion_code,c.visitor_token_hash,c.first_visit_id,c.last_visit_id,c.attribution_visit_id,
		COALESCE(NULLIF(v.utm_source,''),NULLIF(v.referrer_host,''),'直接访问'),c.attribution_model
	FROM candidates c JOIN distribution_promotion_visits v ON v.id=c.attribution_visit_id
		AND v.agent_id=c.agent_id AND v.visitor_token_hash=c.visitor_token_hash AND NOT v.is_bot
		ORDER BY c.attribution_visit_id DESC LIMIT 1`, visitorTokenHash, code).
		Scan(&out.Code, &out.VisitorTokenHash, &out.FirstVisitID, &out.LastVisitID, &out.AttributionVisitID, &out.AttributionSource, &out.AttributionModel)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	out.Type = "persisted"
	if code != "" {
		out.Type = "direct"
	}
	return &out, nil
}

func (r *distributionRepository) promotionAnalyticsAgent(ctx context.Context, userID, requestedAgentID int64, admin bool) (int64, error) {
	if admin {
		if requestedAgentID == 0 {
			return 0, nil
		}
		var valid bool
		if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM distribution_agents WHERE id=$1)`, requestedAgentID).Scan(&valid); err != nil {
			return 0, fmt.Errorf("validate promotion analytics agent: %w", err)
		}
		if !valid {
			return 0, infraerrors.BadRequest("INVALID_PROMOTION_AGENT", "promotion analytics agent does not exist")
		}
		return requestedAgentID, nil
	}
	var agentID int64
	var allowed bool
	err := r.db.QueryRowContext(ctx, `SELECT id,can_view_promotion_stats FROM distribution_agents
		WHERE user_id=$1 AND status<>'revoked'`, userID).Scan(&agentID, &allowed)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, service.ErrDistributionNotAgent
	}
	if err != nil {
		return 0, err
	}
	if !allowed {
		return 0, infraerrors.Forbidden("PROMOTION_STATS_DENIED", "promotion statistics permission is required")
	}
	return agentID, nil
}

func (r *distributionRepository) GetPromotionAnalytics(ctx context.Context, userID int64, filter service.DistributionPromotionStatsFilter, admin bool) (*service.DistributionPromotionAnalytics, error) {
	agentID, err := r.promotionAnalyticsAgent(ctx, userID, filter.AgentID, admin)
	if err != nil {
		return nil, err
	}
	var archiveUnavailable bool
	if err = r.db.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM distribution_promotion_daily_rollups r
		WHERE ($1=0 OR r.agent_id=$1)
		  AND (r.stat_date::timestamp AT TIME ZONE 'Asia/Hong_Kong')<$3
		  AND ((r.stat_date+1)::timestamp AT TIME ZONE 'Asia/Hong_Kong')>$2
		  AND (r.rollup_kind='legacy' OR NOT (
			$2<=(r.stat_date::timestamp AT TIME ZONE 'Asia/Hong_Kong')
			AND $3>=((r.stat_date+1)::timestamp AT TIME ZONE 'Asia/Hong_Kong')
		  ))
	)`, agentID, filter.From, filter.To).Scan(&archiveUnavailable); err != nil {
		return nil, err
	}
	if archiveUnavailable {
		return nil, service.ErrPromotionArchiveUnavailable
	}

	const factsCTE = `WITH cfg AS (
		SELECT promotion_bot_filter_enabled filter_bots FROM distribution_settings WHERE id=1
	), raw_facts AS (
		SELECT (v.visited_at AT TIME ZONE 'Asia/Hong_Kong')::date stat_date,v.agent_id,
			COALESCE(NULLIF(v.utm_source,''),NULLIF(v.referrer_host,''),'直接访问') source,
			v.device_type,v.is_bot,v.visitor_token_hash,COUNT(*) visits,
			COUNT(*) FILTER(WHERE EXISTS (SELECT 1 FROM distribution_promotion_conversions c
				WHERE c.agent_id=v.agent_id AND c.visitor_token_hash=v.visitor_token_hash
				AND v.visited_at<=c.registered_at AND c.attribution_type='direct')) direct_visits,
			COUNT(*) FILTER(WHERE EXISTS (SELECT 1 FROM distribution_promotion_conversions c
				WHERE c.agent_id=v.agent_id AND c.visitor_token_hash=v.visitor_token_hash
				AND v.visited_at<=c.registered_at AND c.attribution_type='persisted')) persisted_visits,
			COUNT(*) FILTER(WHERE NOT EXISTS (SELECT 1 FROM distribution_promotion_conversions c
				WHERE c.agent_id=v.agent_id AND c.visitor_token_hash=v.visitor_token_hash
				AND v.visited_at<=c.registered_at)) unregistered_visits,
			BOOL_OR(EXISTS (SELECT 1 FROM distribution_promotion_conversions c WHERE c.agent_id=v.agent_id
				AND c.attribution_visit_id=v.id AND c.attribution_type='direct')) credited_direct,
			BOOL_OR(EXISTS (SELECT 1 FROM distribution_promotion_conversions c WHERE c.agent_id=v.agent_id
				AND c.attribution_visit_id=v.id AND c.attribution_type='persisted')) credited_persisted,
			BOOL_OR(EXISTS (SELECT 1 FROM distribution_promotion_conversions c WHERE c.agent_id=v.agent_id
				AND c.visitor_token_hash=v.visitor_token_hash AND v.visited_at<=c.registered_at
				AND c.attribution_type='direct' AND LOWER(c.attribution_source)=LOWER(COALESCE(NULLIF(v.utm_source,''),NULLIF(v.referrer_host,''),'直接访问')))) source_direct,
			BOOL_OR(EXISTS (SELECT 1 FROM distribution_promotion_conversions c WHERE c.agent_id=v.agent_id
				AND c.visitor_token_hash=v.visitor_token_hash AND v.visited_at<=c.registered_at
				AND c.attribution_type='persisted' AND LOWER(c.attribution_source)=LOWER(COALESCE(NULLIF(v.utm_source,''),NULLIF(v.referrer_host,''),'直接访问')))) source_persisted,
			SUM((SELECT COUNT(*) FROM distribution_promotion_conversions c WHERE c.agent_id=v.agent_id
				AND c.attribution_visit_id=v.id AND c.attribution_type='direct')) credited_direct_registrations,
			SUM((SELECT COUNT(*) FROM distribution_promotion_conversions c WHERE c.agent_id=v.agent_id
				AND c.attribution_visit_id=v.id AND c.attribution_type='persisted')) credited_persisted_registrations
		FROM distribution_promotion_visits v
		WHERE v.visited_at>=$2 AND v.visited_at<$3 AND ($1=0 OR v.agent_id=$1)
		GROUP BY 1,2,3,4,5,6
	), archived_facts AS (
		SELECT r.stat_date,r.agent_id,r.source,r.device_type,r.is_bot,r.visitor_token_hash,r.visits,
			r.direct_visits,r.persisted_visits,r.unregistered_visits,r.credited_direct,r.credited_persisted,
			r.source_direct,r.source_persisted,r.credited_direct_registrations,r.credited_persisted_registrations
		FROM distribution_promotion_daily_rollups r
		WHERE r.rollup_kind='visitor' AND ($1=0 OR r.agent_id=$1)
		  AND (r.stat_date::timestamp AT TIME ZONE 'Asia/Hong_Kong')>=$2
		  AND ((r.stat_date+1)::timestamp AT TIME ZONE 'Asia/Hong_Kong')<=$3
	), facts AS (
		SELECT * FROM raw_facts UNION ALL SELECT * FROM archived_facts
	), selected_facts AS (
		SELECT f.*,
			CASE $6 WHEN 'direct' THEN direct_visits WHEN 'persisted' THEN persisted_visits
				WHEN 'unregistered' THEN unregistered_visits ELSE visits END fact_visits,
			CASE $6 WHEN 'direct' THEN credited_direct WHEN 'persisted' THEN credited_persisted
				WHEN 'unregistered' THEN FALSE ELSE credited_direct OR credited_persisted END fact_credited,
			CASE $6 WHEN 'direct' THEN source_direct WHEN 'persisted' THEN source_persisted
				WHEN 'unregistered' THEN FALSE ELSE source_direct OR source_persisted END source_converted,
			CASE $6 WHEN 'direct' THEN credited_direct_registrations WHEN 'persisted' THEN credited_persisted_registrations
				WHEN 'unregistered' THEN 0 ELSE credited_direct_registrations+credited_persisted_registrations END fact_registrations
		FROM facts f,cfg
		WHERE (NOT cfg.filter_bots OR NOT f.is_bot)
		  AND ($4='' OR LOWER(f.source)=LOWER($4)) AND ($5='' OR f.device_type=$5)
	), filtered_facts AS (
		SELECT * FROM selected_facts WHERE fact_visits>0
	)`

	out := &service.DistributionPromotionAnalytics{Daily: []service.DistributionPromotionDailyStat{}, Sources: []service.DistributionPromotionSourceStat{}}
	out.Meta.Timezone = "Asia/Hong_Kong"
	out.Meta.Cohort = "visit"
	out.Meta.GeneratedAt = time.Now().UTC()
	if err = r.db.QueryRowContext(ctx, `SELECT promotion_bot_filter_enabled,enabled AND promotion_tracking_enabled,
		promotion_attribution_enabled,promotion_attribution_days,promotion_attribution_model,promotion_detail_retention_days,
		GREATEST(promotion_detail_retention_days,367) FROM distribution_settings WHERE id=1`).Scan(
		&out.Meta.BotFilterEnabled, &out.Meta.TrackingEnabled, &out.Meta.AttributionEnabled, &out.Meta.AttributionDays,
		&out.Meta.AttributionModel, &out.Meta.DetailRetentionDays, &out.Meta.RawRetentionDays); err != nil {
		return nil, err
	}
	if err = r.db.QueryRowContext(ctx, factsCTE+`
		SELECT COALESCE((SELECT SUM(fact_visits) FROM filtered_facts),0),
			(SELECT COUNT(DISTINCT (agent_id,visitor_token_hash)) FROM filtered_facts),
			COALESCE((SELECT SUM(visits) FROM facts WHERE is_bot
			  AND ($4='' OR LOWER(source)=LOWER($4)) AND ($5='' OR device_type=$5)),0),
			(SELECT COUNT(DISTINCT (agent_id,visitor_token_hash)) FROM filtered_facts WHERE fact_credited),
			COALESCE((SELECT SUM(fact_registrations) FROM filtered_facts),0),
			COALESCE((SELECT SUM(fact_registrations) FROM filtered_facts),0),
			COALESCE((SELECT SUM(CASE WHEN $6='persisted' OR $6='unregistered' THEN 0 ELSE credited_direct_registrations END) FROM filtered_facts),0),
			COALESCE((SELECT SUM(CASE WHEN $6='direct' OR $6='unregistered' THEN 0 ELSE credited_persisted_registrations END) FROM filtered_facts),0),
			0`, agentID, filter.From, filter.To,
		filter.Source, filter.Device, filter.AttributionType).Scan(
		&out.Summary.TotalVisits, &out.Summary.UniqueVisitors, &out.Summary.BotVisits,
		&out.Summary.ConvertedVisitors, &out.Summary.TrackedRegistrations, &out.Summary.Registrations,
		&out.Summary.DirectRegistrations, &out.Summary.PersistedRegistrations, &out.Summary.UntrackedDirect); err != nil {
		return nil, err
	}
	if out.Summary.UniqueVisitors > 0 {
		out.Summary.ConversionRate = float64(out.Summary.ConvertedVisitors) / float64(out.Summary.UniqueVisitors) * 100
	}
	rows, err := r.db.QueryContext(ctx, factsCTE+`,
		dates AS (SELECT generate_series(($2::timestamptz AT TIME ZONE 'Asia/Hong_Kong')::date,(($3::timestamptz-interval '1 second') AT TIME ZONE 'Asia/Hong_Kong')::date,interval '1 day')::date AS stat_date),
		v AS (SELECT stat_date,SUM(fact_visits) visits,COUNT(DISTINCT (agent_id,visitor_token_hash)) visitors,
			COUNT(DISTINCT (agent_id,visitor_token_hash)) FILTER(WHERE fact_credited) conversions
			FROM filtered_facts GROUP BY stat_date),
		r AS (SELECT stat_date,SUM(fact_registrations) registrations
			FROM filtered_facts GROUP BY stat_date)
		SELECT TO_CHAR(dates.stat_date,'YYYY-MM-DD'),COALESCE(v.visits,0),COALESCE(v.visitors,0),
			COALESCE(v.conversions,0),COALESCE(r.registrations,0)
		FROM dates LEFT JOIN v USING(stat_date) LEFT JOIN r USING(stat_date) ORDER BY dates.stat_date DESC`, agentID, filter.From, filter.To,
		filter.Source, filter.Device, filter.AttributionType)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var item service.DistributionPromotionDailyStat
		if err = rows.Scan(&item.Date, &item.Visits, &item.Visitors, &item.Conversions, &item.Registrations); err != nil {
			_ = rows.Close()
			return nil, err
		}
		out.Daily = append(out.Daily, item)
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	rows, err = r.db.QueryContext(ctx, factsCTE+`,
		v AS (SELECT LOWER(source) source_key,MIN(source) source,SUM(fact_visits) visits,
			COUNT(DISTINCT (agent_id,visitor_token_hash)) FILTER(WHERE source_converted) conversions
			FROM filtered_facts GROUP BY LOWER(source)),
		r AS (SELECT LOWER(source) source_key,MIN(source) source,SUM(fact_registrations) registrations
			FROM filtered_facts GROUP BY LOWER(source))
		SELECT COALESCE(v.source,r.source),COALESCE(v.visits,0),COALESCE(v.conversions,0),COALESCE(r.registrations,0)
		FROM v FULL OUTER JOIN r USING(source_key)
		ORDER BY COALESCE(v.visits,0)+COALESCE(r.registrations,0) DESC LIMIT 8`, agentID, filter.From, filter.To,
		filter.Source, filter.Device, filter.AttributionType)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var item service.DistributionPromotionSourceStat
		if err = rows.Scan(&item.Source, &item.Visits, &item.Conversions, &item.Registrations); err != nil {
			return nil, err
		}
		out.Sources = append(out.Sources, item)
	}
	return out, rows.Err()
}

func (r *distributionRepository) ListPromotionVisits(ctx context.Context, userID int64, filter service.DistributionPromotionStatsFilter, admin bool) ([]service.DistributionPromotionVisit, int64, error) {
	agentID, err := r.promotionAnalyticsAgent(ctx, userID, filter.AgentID, admin)
	if err != nil {
		return nil, 0, err
	}
	args := []any{agentID, filter.From, filter.To, filter.Source, filter.Device, filter.AttributionType}
	const where = ` FROM distribution_promotion_visits v JOIN distribution_agents a ON a.id=v.agent_id JOIN users u ON u.id=a.user_id
		LEFT JOIN LATERAL (SELECT c.registered_at,c.attribution_type,c.user_id FROM distribution_promotion_conversions c
			WHERE c.agent_id=v.agent_id AND c.visitor_token_hash=v.visitor_token_hash AND v.visited_at<=c.registered_at
			ORDER BY c.registered_at,c.user_id LIMIT 1) c ON TRUE
		WHERE v.detail_expired_at IS NULL AND v.visited_at>=$2 AND v.visited_at<$3 AND ($1=0 OR v.agent_id=$1)
		AND ($4='' OR LOWER(COALESCE(NULLIF(v.utm_source,''),NULLIF(v.referrer_host,''),'直接访问'))=LOWER($4))
		AND ($5='' OR v.device_type=$5)
		AND ($6='' OR ($6='unregistered' AND c.user_id IS NULL) OR ($6<>'unregistered' AND c.attribution_type=$6))`
	var total int64
	if err = r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err := r.db.QueryContext(ctx, `SELECT v.id,v.agent_id,COALESCE(u.email,''),v.promotion_code,v.landing_path,
		COALESCE(NULLIF(v.utm_source,''),NULLIF(v.referrer_host,''),'直接访问'),v.device_type,v.is_bot,v.visited_at,
		c.registered_at,COALESCE(c.attribution_type,'')`+where+` ORDER BY v.visited_at DESC,v.id DESC LIMIT $7 OFFSET $8`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionPromotionVisit, 0)
	for rows.Next() {
		var item service.DistributionPromotionVisit
		if err = rows.Scan(&item.ID, &item.AgentID, &item.AgentEmail, &item.PromotionCode, &item.LandingPath, &item.Source,
			&item.DeviceType, &item.IsBot, &item.VisitedAt, &item.RegisteredAt, &item.AttributionType); err != nil {
			return nil, 0, err
		}
		if !admin {
			item.AgentEmail = ""
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list promotion visits: %w", err)
	}
	return items, total, nil
}
