package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"golang.org/x/sync/errgroup"
)

type businessAnalyticsRepository struct{ db *sql.DB }

const businessAnalyticsRetryChunk = 24 * time.Hour
const businessAnalyticsRetention = 735 * 24 * time.Hour

func NewBusinessAnalyticsRepository(db *sql.DB) *businessAnalyticsRepository {
	return &businessAnalyticsRepository{db: db}
}

func businessAgentFilter(alias string, agentParam, scopeParam int) string {
	agent := fmt.Sprintf("$%d", agentParam)
	scope := fmt.Sprintf("$%d", scopeParam)
	return `(` + agent + `=0 OR EXISTS (SELECT 1 FROM distribution_customer_bindings b JOIN distribution_agents a ON a.id=b.agent_id WHERE b.user_id=` + alias + `.id AND (a.id=` + agent + ` OR (` + scope + `='team' AND a.parent_agent_id=` + agent + `))))`
}

func (r *businessAnalyticsRepository) GetBusinessAgentDepth(ctx context.Context, agentID int64) (int, error) {
	var depth int
	err := r.db.QueryRowContext(ctx, `SELECT l.depth FROM distribution_agents a JOIN distribution_agent_levels l ON l.id=a.level_id WHERE a.id=$1 AND a.status<>'revoked'`, agentID).Scan(&depth)
	if err == sql.ErrNoRows {
		return 0, service.ErrBusinessAnalyticsAgentNotFound
	}
	return depth, err
}

// AggregateRange refreshes the sparse hourly and daily business buckets. The
// source window is intentionally re-read on every run so late payment/refund
// events correct the open bucket without double counting.
func (r *businessAnalyticsRepository) AggregateRange(ctx context.Context, start, end time.Time) error {
	if r == nil || r.db == nil || !end.After(start) {
		return nil
	}
	requestedStart := start
	var pendingFrom sql.NullTime
	var invalidationVersion int64
	if err := r.db.QueryRowContext(ctx, `SELECT pending_from,invalidation_version FROM business_analytics_aggregation_state WHERE id=1`).Scan(&pendingFrom, &invalidationVersion); err != nil && err != sql.ErrNoRows {
		return err
	}
	nextPending := any(nil)
	if pendingFrom.Valid && pendingFrom.Time.Before(requestedStart) {
		retryEnd := pendingFrom.Time.Add(businessAnalyticsRetryChunk)
		if retryEnd.After(requestedStart) {
			retryEnd = requestedStart
		}
		if err := r.aggregateBusinessRange(ctx, pendingFrom.Time, retryEnd); err != nil {
			return r.recordBusinessAggregationFailure(err, pendingFrom.Time)
		}
		if retryEnd.Before(requestedStart) {
			nextPending = retryEnd
		}
	} else if pendingFrom.Valid && !pendingFrom.Time.Before(end) {
		// A manual historical backfill may end before a newer invalidation. Keep
		// that future cursor until a scheduled range actually covers it.
		nextPending = pendingFrom.Time
	}

	// Refresh the requested interval even while an older backlog is catching up.
	// This keeps current dashboards live without combining the complete backlog
	// into one large transaction.
	if err := r.aggregateBusinessRange(ctx, requestedStart, end); err != nil {
		return r.recordBusinessAggregationFailure(err, requestedStart)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO business_analytics_aggregation_state(id,last_aggregated_at,pending_from,invalidation_version,updated_at)
		VALUES(1,$1,$2,$3,NOW()) ON CONFLICT(id) DO UPDATE SET
		last_aggregated_at=EXCLUDED.last_aggregated_at,pending_from=EXCLUDED.pending_from,updated_at=NOW()
		WHERE business_analytics_aggregation_state.invalidation_version=EXCLUDED.invalidation_version`, end, nextPending, invalidationVersion)
	return err
}

func (r *businessAnalyticsRepository) recordBusinessAggregationFailure(aggregateErr error, invalidFrom time.Time) error {
	stateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, stateErr := r.db.ExecContext(stateCtx, `SELECT business_invalidate_analytics($1)`, invalidFrom)
	if stateErr != nil {
		return fmt.Errorf("aggregate business analytics: %w (record retry state: %v)", aggregateErr, stateErr)
	}
	return aggregateErr
}

func (r *businessAnalyticsRepository) aggregateBusinessRange(ctx context.Context, start, end time.Time) error {
	loc := time.FixedZone("Asia/Shanghai", 8*60*60)
	start = start.In(loc).Truncate(time.Hour)
	end = end.In(loc)
	if end.Truncate(time.Hour).Equal(end) {
		// Keep the completed boundary out of the source scan; the bucket itself is
		// still refreshed by the preceding interval.
	} else {
		end = end.Truncate(time.Hour).Add(time.Hour)
	}
	if !end.After(start) {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `INSERT INTO business_payment_facts (
		payment_order_id,currency,gross_amount,refunded_amount,fx_rate_to_cny,fx_estimated,
		paid_at,refunded_at,channel,channel_ref_id,updated_at
	)
	SELECT po.id,
		UPPER(COALESCE(NULLIF(po.provider_snapshot->>'currency',''),'CNY')),
		COALESCE(NULLIF(po.provider_amount,0),po.pay_amount,0),
		CASE WHEN po.status IN ('PARTIALLY_REFUNDED','REFUNDED') AND po.amount>0
			THEN COALESCE(NULLIF(po.provider_amount,0),po.pay_amount,0)*COALESCE(po.refund_amount,0)/po.amount ELSE 0 END,
		business_effective_fx_rate_to_cny(UPPER(COALESCE(NULLIF(po.provider_snapshot->>'currency',''),'CNY'))),
			business_effective_fx_rate_to_cny(UPPER(COALESCE(NULLIF(po.provider_snapshot->>'currency',''),'CNY'))) IS NULL
				OR UPPER(COALESCE(NULLIF(po.provider_snapshot->>'currency',''),'CNY')) <> 'CNY',
		po.paid_at,CASE WHEN po.status IN ('PARTIALLY_REFUNDED','REFUNDED') THEN po.refund_at ELSE NULL END,
		COALESCE(buf.channel,'unknown'),buf.channel_ref_id,NOW()
	FROM payment_orders po LEFT JOIN business_user_facts buf ON buf.user_id=po.user_id
	WHERE po.status IN ('PAID','RECHARGING','COMPLETED','FAILED','REFUND_REQUESTED','REFUNDING','REFUND_PENDING','REFUND_FAILED','PARTIALLY_REFUNDED','REFUNDED')
		AND po.paid_at IS NOT NULL
		AND (po.updated_at >= $1 AND po.updated_at < $2 OR po.paid_at >= $1 AND po.paid_at < $2 OR po.refund_at >= $1 AND po.refund_at < $2)
	ON CONFLICT(payment_order_id) DO UPDATE SET
		currency=EXCLUDED.currency,gross_amount=EXCLUDED.gross_amount,refunded_amount=EXCLUDED.refunded_amount,
		fx_rate_to_cny=COALESCE(business_payment_facts.fx_rate_to_cny,EXCLUDED.fx_rate_to_cny),
		fx_estimated=business_payment_facts.fx_estimated OR EXCLUDED.fx_estimated,
		paid_at=EXCLUDED.paid_at,refunded_at=EXCLUDED.refunded_at,updated_at=NOW()`, start, end); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO business_user_facts(user_id,first_paid_at,created_at,updated_at)
		SELECT affected.user_id,MIN(fact.paid_at) FILTER(WHERE fact.gross_amount>fact.refunded_amount),MIN(u.created_at),NOW()
		FROM (SELECT DISTINCT user_id FROM payment_orders WHERE updated_at >= $1 AND updated_at < $2 OR paid_at >= $1 AND paid_at < $2 OR refund_at >= $1 AND refund_at < $2) affected
		JOIN users u ON u.id=affected.user_id
		LEFT JOIN payment_orders po ON po.user_id=affected.user_id
		LEFT JOIN business_payment_facts fact ON fact.payment_order_id=po.id
		GROUP BY affected.user_id
		ON CONFLICT(user_id) DO UPDATE SET first_paid_at=EXCLUDED.first_paid_at,updated_at=NOW()`, start, end); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO business_usage_hourly_facts
		(bucket_start,user_id,first_used_at,last_used_at,consumed_revenue,supplier_cost,updated_at)
		SELECT date_trunc('hour',timezone('Asia/Shanghai',ul.created_at)) AT TIME ZONE 'Asia/Shanghai',
			ul.user_id,MIN(ul.created_at),MAX(ul.created_at),COALESCE(SUM(ul.actual_cost),0),
			COALESCE(SUM(COALESCE(ul.account_stats_cost,ul.total_cost)*COALESCE(ul.account_rate_multiplier,1)),0),NOW()
		FROM usage_logs ul WHERE ul.actual_cost>0 AND ul.created_at >= $1 AND ul.created_at < $2
		GROUP BY 1,2
		ON CONFLICT(bucket_start,user_id) DO UPDATE SET
			first_used_at=EXCLUDED.first_used_at,last_used_at=EXCLUDED.last_used_at,
			consumed_revenue=EXCLUDED.consumed_revenue,supplier_cost=EXCLUDED.supplier_cost,updated_at=NOW()`, start, end); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO business_user_facts(user_id,first_activated_at,created_at,updated_at)
		SELECT fact.user_id,MIN(fact.first_used_at),MIN(u.created_at),NOW()
		FROM business_usage_hourly_facts fact JOIN users u ON u.id=fact.user_id
		WHERE fact.bucket_start >= $1 AND fact.bucket_start < $2 GROUP BY fact.user_id
		ON CONFLICT(user_id) DO UPDATE SET
		first_activated_at=LEAST(COALESCE(business_user_facts.first_activated_at,EXCLUDED.first_activated_at),EXCLUDED.first_activated_at),updated_at=NOW()`, start, end); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
WITH buckets AS (
  SELECT date_trunc('hour', timezone('Asia/Shanghai', gs)) AT TIME ZONE 'Asia/Shanghai' AS bucket_start
  FROM generate_series($1::timestamptz, $2::timestamptz - interval '1 hour', interval '1 hour') gs
), scoped AS (
  SELECT u.id, COALESCE(buf.channel, 'unknown') AS channel
  FROM users u LEFT JOIN business_user_facts buf ON buf.user_id=u.id
), usage_hour AS (
  SELECT fact.bucket_start, su.channel, COUNT(DISTINCT fact.user_id) active_users,
         COALESCE(SUM(fact.consumed_revenue),0) consumed,
         COALESCE(SUM(fact.supplier_cost),0) supplier
  FROM business_usage_hourly_facts fact JOIN scoped su ON su.id=fact.user_id
  WHERE fact.bucket_start >= $1 AND fact.bucket_start < $2
  GROUP BY 1,2
), payment_hour AS (
  SELECT date_trunc('hour', timezone('Asia/Shanghai', bpf.paid_at)) AT TIME ZONE 'Asia/Shanghai' bucket_start,
         bpf.channel, COUNT(DISTINCT po.user_id) FILTER (WHERE bpf.gross_amount>bpf.refunded_amount) paying_users,
         COUNT(DISTINCT po.user_id) FILTER (WHERE buf.first_paid_at >= date_trunc('hour', timezone('Asia/Shanghai', bpf.paid_at)) AT TIME ZONE 'Asia/Shanghai' AND buf.first_paid_at < (date_trunc('hour', timezone('Asia/Shanghai', bpf.paid_at)) + interval '1 hour') AT TIME ZONE 'Asia/Shanghai') first_paid_users,
         COUNT(DISTINCT po.user_id) FILTER (WHERE bpf.gross_amount>bpf.refunded_amount AND buf.first_paid_at < date_trunc('hour', timezone('Asia/Shanghai', bpf.paid_at)) AT TIME ZONE 'Asia/Shanghai') repurchase_users,
		COALESCE(SUM(bpf.gross_amount*bpf.fx_rate_to_cny),0) gross_paid
  FROM business_payment_facts bpf JOIN payment_orders po ON po.id=bpf.payment_order_id
  JOIN scoped su ON su.id=po.user_id
  LEFT JOIN business_user_facts buf ON buf.user_id=po.user_id
  WHERE bpf.paid_at >= $1 AND bpf.paid_at < $2
  GROUP BY 1,2
), refund_hour AS (
  SELECT date_trunc('hour', timezone('Asia/Shanghai', bpf.refunded_at)) AT TIME ZONE 'Asia/Shanghai' bucket_start,
         bpf.channel, COALESCE(SUM(bpf.refunded_amount*bpf.fx_rate_to_cny),0) refunded
  FROM business_payment_facts bpf JOIN payment_orders po ON po.id=bpf.payment_order_id
  JOIN scoped su ON su.id=po.user_id
  WHERE bpf.refunded_at >= $1 AND bpf.refunded_at < $2
  GROUP BY 1,2
), registrations_hour AS (
	  SELECT date_trunc('hour', timezone('Asia/Shanghai', u.created_at)) AT TIME ZONE 'Asia/Shanghai' bucket_start,
	         COALESCE(buf.channel, 'unknown') channel, COUNT(*) new_users
	  FROM users u LEFT JOIN business_user_facts buf ON buf.user_id=u.id
	  WHERE u.created_at >= $1 AND u.created_at < $2 GROUP BY 1,2
), activations_hour AS (
	  SELECT date_trunc('hour', timezone('Asia/Shanghai', buf.first_activated_at)) AT TIME ZONE 'Asia/Shanghai' bucket_start,
	         buf.channel, COUNT(*) activated_users
	  FROM business_user_facts buf
	  WHERE buf.first_activated_at >= $1 AND buf.first_activated_at < $2 GROUP BY 1,2
), rows AS (
			SELECT b.bucket_start, c.channel, COALESCE(reg.new_users,0) new_users, COALESCE(act.activated_users,0) activated_users,
    COALESCE(x.active_users,0) active_users, COALESCE(p.paying_users,0) paying_users, COALESCE(p.first_paid_users,0) first_paid_users,
		COALESCE(p.repurchase_users,0) repurchase_users, COALESCE(p.gross_paid,0) gross_paid, COALESCE(r.refunded,0) refunded, COALESCE(x.consumed,0) consumed, COALESCE(x.supplier,0) supplier
  FROM buckets b CROSS JOIN (SELECT unnest(ARRAY['distribution','affiliate','campaign','organic','unknown']) channel) c
	  LEFT JOIN registrations_hour reg ON reg.bucket_start=b.bucket_start AND reg.channel=c.channel
	  LEFT JOIN activations_hour act ON act.bucket_start=b.bucket_start AND act.channel=c.channel
  LEFT JOIN usage_hour x ON x.bucket_start=b.bucket_start AND x.channel=c.channel
  LEFT JOIN payment_hour p ON p.bucket_start=b.bucket_start AND p.channel=c.channel
  LEFT JOIN refund_hour r ON r.bucket_start=b.bucket_start AND r.channel=c.channel
)
INSERT INTO business_analytics_buckets (bucket_start,resolution,scope,scope_id,metrics,estimated,updated_at)
	SELECT bucket_start,'hour','channel',channel,jsonb_build_object('new_users',new_users,'activated_users',activated_users,'active_users',active_users,'paying_users',paying_users,'first_paid_users',first_paid_users,'repurchase_users',repurchase_users,'gross_paid_cny',gross_paid,'refunded_cny',refunded,'consumed_revenue',consumed,'supplier_cost',supplier),FALSE,NOW() FROM rows
ON CONFLICT (resolution,bucket_start,scope,scope_id) DO UPDATE SET metrics=EXCLUDED.metrics,updated_at=NOW()`, start, end); err != nil {
		return err
	}
	dailyStart := time.Date(start.In(loc).Year(), start.In(loc).Month(), start.In(loc).Day(), 0, 0, 0, 0, loc)
	lastIncluded := end.Add(-time.Nanosecond).In(loc)
	dailyEnd := time.Date(lastIncluded.Year(), lastIncluded.Month(), lastIncluded.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
	if _, err = tx.ExecContext(ctx, `WITH days AS (
	 SELECT date_trunc('day', timezone('Asia/Shanghai', gs)) AT TIME ZONE 'Asia/Shanghai' bucket_start
	 FROM generate_series(
	   date_trunc('day', timezone('Asia/Shanghai',$1::timestamptz)) AT TIME ZONE 'Asia/Shanghai',
	   date_trunc('day', timezone('Asia/Shanghai',$2::timestamptz-interval '1 second')) AT TIME ZONE 'Asia/Shanghai',
	   interval '1 day') gs
), channels AS (SELECT unnest(ARRAY['distribution','affiliate','campaign','organic','unknown']) channel),
scoped AS (SELECT u.id,COALESCE(buf.channel,'unknown') channel FROM users u LEFT JOIN business_user_facts buf ON buf.user_id=u.id),
registrations AS (
	 SELECT date_trunc('day',timezone('Asia/Shanghai',u.created_at)) AT TIME ZONE 'Asia/Shanghai' bucket_start,
	   COALESCE(buf.channel,'unknown') channel,COUNT(*) new_users
	 FROM users u LEFT JOIN business_user_facts buf ON buf.user_id=u.id
	 WHERE u.created_at >= $1 AND u.created_at < $2 GROUP BY 1,2
), activations AS (
	 SELECT date_trunc('day',timezone('Asia/Shanghai',first_activated_at)) AT TIME ZONE 'Asia/Shanghai' bucket_start,
	   channel,COUNT(*) activated_users
	 FROM business_user_facts WHERE first_activated_at >= $1 AND first_activated_at < $2 GROUP BY 1,2
), usage_daily AS (
	 SELECT date_trunc('day',timezone('Asia/Shanghai',fact.bucket_start)) AT TIME ZONE 'Asia/Shanghai' bucket_start,
	   su.channel,COUNT(DISTINCT fact.user_id) active_users,COALESCE(SUM(fact.consumed_revenue),0) consumed,
	   COALESCE(SUM(fact.supplier_cost),0) supplier
	 FROM business_usage_hourly_facts fact JOIN scoped su ON su.id=fact.user_id
	 WHERE fact.bucket_start >= $1 AND fact.bucket_start < $2 GROUP BY 1,2
), payment_daily AS (
	 SELECT date_trunc('day',timezone('Asia/Shanghai',bpf.paid_at)) AT TIME ZONE 'Asia/Shanghai' bucket_start,
	   bpf.channel,COUNT(DISTINCT po.user_id) FILTER(WHERE bpf.gross_amount>bpf.refunded_amount) paying_users,
	   COUNT(DISTINCT po.user_id) FILTER(WHERE buf.first_paid_at >= date_trunc('day',timezone('Asia/Shanghai',bpf.paid_at)) AT TIME ZONE 'Asia/Shanghai' AND buf.first_paid_at < (date_trunc('day',timezone('Asia/Shanghai',bpf.paid_at))+interval '1 day') AT TIME ZONE 'Asia/Shanghai') first_paid_users,
	   COUNT(DISTINCT po.user_id) FILTER(WHERE bpf.gross_amount>bpf.refunded_amount AND buf.first_paid_at < date_trunc('day',timezone('Asia/Shanghai',bpf.paid_at)) AT TIME ZONE 'Asia/Shanghai') repurchase_users,
	   COALESCE(SUM(bpf.gross_amount*bpf.fx_rate_to_cny),0) gross_paid
	 FROM business_payment_facts bpf JOIN payment_orders po ON po.id=bpf.payment_order_id JOIN scoped su ON su.id=po.user_id
	 LEFT JOIN business_user_facts buf ON buf.user_id=po.user_id
	 WHERE bpf.paid_at >= $1 AND bpf.paid_at < $2 GROUP BY 1,2
), refund_daily AS (
	 SELECT date_trunc('day',timezone('Asia/Shanghai',bpf.refunded_at)) AT TIME ZONE 'Asia/Shanghai' bucket_start,
	   bpf.channel,COALESCE(SUM(bpf.refunded_amount*bpf.fx_rate_to_cny),0) refunded
	 FROM business_payment_facts bpf JOIN payment_orders po ON po.id=bpf.payment_order_id JOIN scoped su ON su.id=po.user_id
	 WHERE bpf.refunded_at >= $1 AND bpf.refunded_at < $2 GROUP BY 1,2
), daily AS (
	 SELECT d.bucket_start,c.channel scope_id,jsonb_build_object(
	   'new_users',COALESCE(reg.new_users,0),'activated_users',COALESCE(act.activated_users,0),
	   'active_users',COALESCE(u.active_users,0),'paying_users',COALESCE(p.paying_users,0),
	   'first_paid_users',COALESCE(p.first_paid_users,0),'repurchase_users',COALESCE(p.repurchase_users,0),
	   'gross_paid_cny',COALESCE(p.gross_paid,0),'refunded_cny',COALESCE(r.refunded,0),
	   'consumed_revenue',COALESCE(u.consumed,0),'supplier_cost',COALESCE(u.supplier,0)) metrics
	 FROM days d CROSS JOIN channels c
	 LEFT JOIN registrations reg ON reg.bucket_start=d.bucket_start AND reg.channel=c.channel
	 LEFT JOIN activations act ON act.bucket_start=d.bucket_start AND act.channel=c.channel
	 LEFT JOIN usage_daily u ON u.bucket_start=d.bucket_start AND u.channel=c.channel
	 LEFT JOIN payment_daily p ON p.bucket_start=d.bucket_start AND p.channel=c.channel
	 LEFT JOIN refund_daily r ON r.bucket_start=d.bucket_start AND r.channel=c.channel
)
INSERT INTO business_analytics_buckets (bucket_start,resolution,scope,scope_id,metrics,estimated,updated_at)
SELECT bucket_start,'day','channel',scope_id,metrics,FALSE,NOW() FROM daily
ON CONFLICT (resolution,bucket_start,scope,scope_id) DO UPDATE SET metrics=EXCLUDED.metrics,updated_at=NOW()`, dailyStart, dailyEnd); err != nil {
		return err
	}
	retentionCutoff := time.Now().UTC().Add(-businessAnalyticsRetention)
	if _, err = tx.ExecContext(ctx, `DELETE FROM business_usage_hourly_facts WHERE bucket_start < $1`, retentionCutoff); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM business_analytics_buckets WHERE bucket_start < $1`, retentionCutoff); err != nil {
		return err
	}
	return tx.Commit()
}

const businessChannelExpr = `COALESCE(buf.channel, 'unknown')`

type businessPeriodValues struct {
	population, newUsers, activated, firstPaid, paying, repurchase, active int64
	cohortActivated, cohortFirstPaid, cohortRepurchased                    int64
	gross, refunded, consumed, supplier, commission, rewards               float64
	estimated                                                              bool
}

func businessChange(current, previous float64) *float64 {
	if previous == 0 {
		return nil
	}
	v := math.Round(((current-previous)/previous)*10000) / 100
	return &v
}

func businessMetric(current, previous float64, currency string, estimated bool) service.BusinessMetric {
	comparisonType := "relative"
	changeRate := businessChange(current, previous)
	if previous < 0 && current >= 0 {
		comparisonType, changeRate = "turned_positive", nil
	} else if previous >= 0 && current < 0 {
		comparisonType, changeRate = "turned_negative", nil
	} else if previous == 0 {
		comparisonType = "unavailable"
	}
	changeValue := current - previous
	return service.BusinessMetric{Value: current, Previous: previous, ChangeRate: changeRate, ChangeValue: &changeValue, ComparisonType: comparisonType, Currency: currency, Estimated: estimated}
}

func businessRateMetric(current, previous float64, estimated bool) service.BusinessMetric {
	change := current - previous
	return service.BusinessMetric{Value: current, Previous: previous, ChangeValue: &change, ComparisonType: "percentage_point", Estimated: estimated}
}

func (r *businessAnalyticsRepository) effectiveUSDToCNY(ctx context.Context) (float64, bool) {
	values := map[string]string{}
	rows, err := r.db.QueryContext(ctx, `SELECT key,value FROM settings WHERE key IN ('currency_usd_to_cny_manual_rate','currency_exchange_rate_auto_sync','currency_usd_to_cny_auto_rate')`)
	if err == nil {
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var key, value string
			if rows.Scan(&key, &value) == nil {
				values[key] = value
			}
		}
	}
	key := "currency_usd_to_cny_manual_rate"
	if values["currency_exchange_rate_auto_sync"] == "true" && strings.TrimSpace(values["currency_usd_to_cny_auto_rate"]) != "" {
		key = "currency_usd_to_cny_auto_rate"
	}
	rate, parseErr := strconv.ParseFloat(strings.TrimSpace(values[key]), 64)
	if parseErr != nil || rate <= 0 {
		return service.DefaultUSDToCNYRate, true
	}
	return rate, false
}

func (r *businessAnalyticsRepository) GetBusinessBalance(ctx context.Context, activityWindowDays int, agentID int64, agentScope string) (*service.BusinessBalanceSnapshot, error) {
	now := time.Now().UTC()
	settings := map[string]string{}
	rows, err := r.db.QueryContext(ctx, `SELECT key,value FROM settings WHERE key IN ('balance_low_notify_enabled','balance_low_notify_threshold','BALANCE_RECHARGE_MULTIPLIER')`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var key, value string
		if err = rows.Scan(&key, &value); err != nil {
			_ = rows.Close()
			return nil, err
		}
		settings[key] = value
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	lowEnabled := settings[service.SettingKeyBalanceLowNotifyEnabled] == "true"
	lowThreshold, _ := strconv.ParseFloat(settings[service.SettingKeyBalanceLowNotifyThreshold], 64)
	if lowThreshold <= 0 {
		lowEnabled = false
	}
	rechargeMultiplier, _ := strconv.ParseFloat(settings[service.SettingBalanceRechargeMult], 64)
	if rechargeMultiplier <= 0 {
		rechargeMultiplier = 1
	}

	result := &service.BusinessBalanceSnapshot{AsOf: now, ActivityWindowDays: activityWindowDays, BalanceRechargeMultiplier: rechargeMultiplier, LowBalanceEnabled: lowEnabled}
	query := `WITH activity AS (
		SELECT u.id,u.balance,u.frozen_balance,u.total_recharged,u.balance_notify_enabled,u.balance_notify_threshold,u.balance_notify_threshold_type,
			buf.first_activated_at first_active,last_usage.last_active
		FROM users u LEFT JOIN business_user_facts buf ON buf.user_id=u.id
		LEFT JOIN LATERAL (
			SELECT fact.last_used_at last_active FROM business_usage_hourly_facts fact
			WHERE fact.user_id=u.id ORDER BY fact.last_used_at DESC LIMIT 1
		) last_usage ON TRUE
		WHERE u.deleted_at IS NULL AND ` + businessAgentFilter("u", 5, 6) + `
	), classified AS (
		SELECT *,CASE WHEN first_active IS NULL THEN 'unactivated'
			WHEN last_active >= $2::timestamptz-($1::int*interval '1 day') THEN 'active'
			WHEN last_active >= $2::timestamptz-interval '30 day' THEN 'silent'
			ELSE 'churned' END segment,
			CASE WHEN balance_notify_threshold_type='percentage' AND total_recharged>0
				THEN total_recharged*COALESCE(balance_notify_threshold,$3)/100
				ELSE COALESCE(balance_notify_threshold,$3) END effective_threshold
		FROM activity
	)
	SELECT segment,COUNT(*),COALESCE(SUM(balance),0),COALESCE(SUM(frozen_balance),0),
		COUNT(*) FILTER(WHERE balance>0),
		COUNT(*) FILTER(WHERE $4 AND balance_notify_enabled AND balance<effective_threshold)
	FROM classified GROUP BY segment ORDER BY segment`
	segmentRows, err := r.db.QueryContext(ctx, query, activityWindowDays, now, lowThreshold, lowEnabled, agentID, agentScope)
	if err != nil {
		return nil, err
	}
	defer func() { _ = segmentRows.Close() }()
	segments := map[string]service.BusinessBalanceSegment{}
	for segmentRows.Next() {
		var item service.BusinessBalanceSegment
		var positive, low int64
		if err = segmentRows.Scan(&item.Key, &item.Users, &item.BalanceUSD, &item.FrozenUSD, &positive, &low); err != nil {
			return nil, err
		}
		segments[item.Key] = item
		result.AvailableBalanceUSD += item.BalanceUSD
		result.FrozenBalanceUSD += item.FrozenUSD
		result.TotalUsers += item.Users
		result.PositiveBalanceUsers += positive
		result.LowBalanceUsers += low
	}
	if err = segmentRows.Err(); err != nil {
		return nil, err
	}
	if result.TotalUsers > 0 {
		result.AverageBalanceUSD = result.AvailableBalanceUSD / float64(result.TotalUsers)
	}
	for _, key := range []string{"active", "silent", "churned", "unactivated"} {
		item := segments[key]
		item.Key = key
		result.Segments = append(result.Segments, item)
	}
	return result, nil
}

func (r *businessAnalyticsRepository) periodValues(ctx context.Context, start, end time.Time, channel string, agentID int64, agentScope string, usdToCNY float64, rateEstimated bool) (businessPeriodValues, error) {
	v := businessPeriodValues{estimated: rateEstimated}
	channelFilter := `($3='' OR ` + businessChannelExpr + `=$3)`
	paymentChannelFilter := `($3='' OR bpf.channel=$3)`
	agentFilter := businessAgentFilter("u", 4, 5)
	err := r.db.QueryRowContext(ctx, `WITH people AS (
		SELECT u.id,u.created_at,`+businessChannelExpr+` channel,buf.first_activated_at first_active,buf.first_paid_at first_paid
		FROM users u LEFT JOIN business_user_facts buf ON buf.user_id=u.id WHERE `+channelFilter+` AND `+agentFilter+`
	), paid_period AS (
		SELECT DISTINCT po.user_id FROM business_payment_facts bpf JOIN payment_orders po ON po.id=bpf.payment_order_id JOIN users u ON u.id=po.user_id LEFT JOIN business_user_facts buf ON buf.user_id=u.id
		WHERE bpf.paid_at>=$1 AND bpf.paid_at<$2 AND bpf.gross_amount>bpf.refunded_amount AND `+paymentChannelFilter+` AND `+agentFilter+`
	), active_period AS (
		SELECT DISTINCT usage.user_id FROM business_usage_hourly_facts usage JOIN users u ON u.id=usage.user_id LEFT JOIN business_user_facts buf ON buf.user_id=u.id
		WHERE usage.first_used_at<$2 AND usage.last_used_at>=$1 AND `+channelFilter+` AND `+agentFilter+`
	), acquisition_cohort AS (
		SELECT id,first_active,first_paid FROM people WHERE created_at>=$1 AND created_at<$2
	)
	SELECT COUNT(*) FILTER (WHERE created_at<$2),COUNT(*) FILTER (WHERE created_at>=$1 AND created_at<$2),
		COUNT(*) FILTER (WHERE first_active>=$1 AND first_active<$2),COUNT(*) FILTER (WHERE first_paid>=$1 AND first_paid<$2),
		(SELECT COUNT(*) FROM paid_period),
		(SELECT COUNT(*) FROM paid_period pp WHERE EXISTS (SELECT 1 FROM payment_orders old JOIN business_payment_facts old_fact ON old_fact.payment_order_id=old.id WHERE old.user_id=pp.user_id AND old_fact.paid_at<$1 AND old_fact.gross_amount>old_fact.refunded_amount)),
		(SELECT COUNT(*) FROM active_period),
		(SELECT COUNT(*) FROM acquisition_cohort WHERE first_active<$2),
		(SELECT COUNT(*) FROM acquisition_cohort WHERE first_paid<$2),
		(SELECT COUNT(*) FROM acquisition_cohort ac WHERE (SELECT COUNT(*) FROM payment_orders po JOIN business_payment_facts pf ON pf.payment_order_id=po.id WHERE po.user_id=ac.id AND pf.paid_at<$2 AND pf.gross_amount>pf.refunded_amount)>=2)
		FROM people`, start, end, channel, agentID, agentScope).
		Scan(&v.population, &v.newUsers, &v.activated, &v.firstPaid, &v.paying, &v.repurchase, &v.active, &v.cohortActivated, &v.cohortFirstPaid, &v.cohortRepurchased)
	if err != nil {
		return v, err
	}

	err = r.db.QueryRowContext(ctx, `SELECT
		COALESCE(SUM(bpf.gross_amount*bpf.fx_rate_to_cny) FILTER(WHERE bpf.paid_at>=$1 AND bpf.paid_at<$2 AND bpf.fx_rate_to_cny IS NOT NULL),0),
		COALESCE(SUM(bpf.refunded_amount*bpf.fx_rate_to_cny) FILTER(WHERE bpf.refunded_at>=$1 AND bpf.refunded_at<$2 AND bpf.fx_rate_to_cny IS NOT NULL),0),
		COALESCE(BOOL_OR(bpf.fx_estimated OR bpf.fx_rate_to_cny IS NULL),FALSE)
	FROM business_payment_facts bpf JOIN payment_orders po ON po.id=bpf.payment_order_id JOIN users u ON u.id=po.user_id LEFT JOIN business_user_facts buf ON buf.user_id=u.id
	WHERE ((bpf.paid_at>=$1 AND bpf.paid_at<$2) OR (bpf.refunded_at>=$1 AND bpf.refunded_at<$2)) AND `+paymentChannelFilter+` AND `+agentFilter, start, end, channel, agentID, agentScope).
		Scan(&v.gross, &v.refunded, &v.estimated)
	if err != nil {
		return v, err
	}
	err = r.db.QueryRowContext(ctx, `WITH usage_values AS (
		SELECT usage.user_id,usage.consumed_revenue consumed,usage.supplier_cost supplier
		FROM business_usage_hourly_facts usage
		WHERE usage.bucket_start>=$1 AND usage.bucket_start<date_trunc('hour',timezone('Asia/Shanghai',$2::timestamptz)) AT TIME ZONE 'Asia/Shanghai'
		UNION ALL
		SELECT ul.user_id,COALESCE(SUM(ul.actual_cost),0),COALESCE(SUM(COALESCE(ul.account_stats_cost,ul.total_cost)*COALESCE(ul.account_rate_multiplier,1)),0)
		FROM usage_logs ul
		WHERE ul.actual_cost>0 AND ul.created_at>=GREATEST($1,date_trunc('hour',timezone('Asia/Shanghai',$2::timestamptz)) AT TIME ZONE 'Asia/Shanghai') AND ul.created_at<$2
		GROUP BY ul.user_id
	) SELECT COALESCE(SUM(usage.consumed),0),COALESCE(SUM(usage.supplier),0)
	FROM usage_values usage JOIN users u ON u.id=usage.user_id LEFT JOIN business_user_facts buf ON buf.user_id=u.id WHERE `+channelFilter+` AND `+agentFilter,
		start, end, channel, agentID, agentScope).Scan(&v.consumed, &v.supplier)
	if err != nil {
		return v, err
	}
	// Usage costs are denominated in USD platform credits. Keep them in their
	// native unit; payment FX only applies to cash receipts and refunds.
	if channel == "" || channel == "distribution" {
		err = r.db.QueryRowContext(ctx, `SELECT
			COALESCE((SELECT SUM(GREATEST(entry.original_amount_cny-entry.reversed_amount_cny,0))
				FROM distribution_commission_entries entry
				JOIN distribution_commission_sources source ON source.id=entry.source_id
				JOIN users cu ON cu.id=source.customer_user_id
				LEFT JOIN business_user_facts buf ON buf.user_id=cu.id
				WHERE entry.created_at>=$1 AND entry.created_at<$2 AND ($3='' OR buf.channel=$3)
				AND ($4=0 OR EXISTS (SELECT 1 FROM distribution_customer_bindings cb JOIN distribution_agents ca ON ca.id=cb.agent_id WHERE cb.user_id=source.customer_user_id AND (ca.id=$4 OR ($5='team' AND ca.parent_agent_id=$4))))),0),
			COALESCE((SELECT SUM(GREATEST(reward.original_amount_cny-reward.reversed_amount_cny,0))
				FROM distribution_reward_grants reward
				JOIN distribution_agents reward_agent ON reward_agent.id=reward.beneficiary_agent_id
				WHERE reward.created_at>=$1 AND reward.created_at<$2 AND ($4=0 OR reward.beneficiary_agent_id=$4 OR ($5='team' AND reward_agent.parent_agent_id=$4))),0)`, start, end, channel, agentID, agentScope).Scan(&v.commission, &v.rewards)
	}
	return v, err
}

func businessBucketFormat(granularity string) string {
	if granularity == "hour" {
		return "2006-01-02T15:04:05+08:00"
	}
	return "2006-01-02"
}

func businessInterval(granularity string) string {
	switch granularity {
	case "hour":
		return "1 hour"
	case "week":
		return "1 week"
	case "month":
		return "1 month"
	default:
		return "1 day"
	}
}

func businessBucketBoundaryAligned(value time.Time, granularity string) bool {
	local := value.In(time.FixedZone("Asia/Shanghai", 8*60*60))
	switch granularity {
	case "hour":
		return local.Equal(local.Truncate(time.Hour))
	case "day":
		return local.Hour() == 0 && local.Minute() == 0 && local.Second() == 0 && local.Nanosecond() == 0
	default:
		return false
	}
}

func (r *businessAnalyticsRepository) bucketTrend(ctx context.Context, filter service.BusinessAnalyticsFilter, usdToCNY float64) ([]service.BusinessTrendPoint, bool, error) {
	// Sparse buckets contain complete hours/days. A partially elapsed final
	// bucket must use event-time queries so current and previous periods cover
	// exactly the same duration.
	if filter.AgentID != 0 || !businessBucketBoundaryAligned(filter.DateTo, filter.Granularity) {
		return nil, false, nil
	}
	resolution := filter.Granularity
	start, end := filter.DateFrom, filter.DateTo
	query := `WITH buckets AS (
 SELECT generate_series(date_trunc($1,timezone('Asia/Shanghai',$2::timestamptz)) AT TIME ZONE 'Asia/Shanghai',
   date_trunc($1,timezone('Asia/Shanghai',($3::timestamptz - interval '1 second'))) AT TIME ZONE 'Asia/Shanghai',
   CASE WHEN $1='hour' THEN interval '1 hour' ELSE interval '1 day' END) bucket_start
), agg AS (
 SELECT bucket_start, SUM((metrics->>'new_users')::numeric) new_users,
  SUM((metrics->>'activated_users')::numeric) activated_users,
  SUM((metrics->>'active_users')::numeric) active_users,
  SUM((metrics->>'paying_users')::numeric) paying_users,
  SUM((metrics->>'first_paid_users')::numeric) first_paid_users,
  SUM((metrics->>'repurchase_users')::numeric) repurchase_users,
  SUM((metrics->>'gross_paid_cny')::numeric) gross_paid,
  SUM((metrics->>'refunded_cny')::numeric) refunded,
  SUM((metrics->>'consumed_revenue')::numeric) consumed,
  SUM((metrics->>'supplier_cost')::numeric) supplier
 FROM business_analytics_buckets
 WHERE resolution=$1 AND scope='channel' AND ($4='' OR scope_id=$4)
   AND bucket_start >= $2 AND bucket_start < $3 GROUP BY bucket_start
)
SELECT b.bucket_start, COALESCE(a.new_users,0),COALESCE(a.activated_users,0),COALESCE(a.active_users,0),
 COALESCE(a.paying_users,0),COALESCE(a.first_paid_users,0),COALESCE(a.repurchase_users,0),
 COALESCE(a.gross_paid,0),COALESCE(a.refunded,0),COALESCE(a.consumed,0),COALESCE(a.supplier,0),
	 NOT EXISTS (SELECT 1 FROM buckets bx WHERE
	   (SELECT COUNT(*) FROM business_analytics_buckets x
	    WHERE x.resolution=$1 AND x.scope='channel' AND ($4='' OR x.scope_id=$4) AND x.bucket_start=bx.bucket_start)
	   <> CASE WHEN $4='' THEN 5 ELSE 1 END)
FROM buckets b LEFT JOIN agg a ON a.bucket_start=b.bucket_start ORDER BY b.bucket_start`
	rows, err := r.db.QueryContext(ctx, query, resolution, start, end, filter.Channel)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = rows.Close() }()
	result := make([]service.BusinessTrendPoint, 0)
	var hasRows bool
	for rows.Next() {
		var bucket time.Time
		var newUsers, activated, active, paying, firstPaid, repurchase int64
		var gross, refunded, consumed, supplier float64
		if err := rows.Scan(&bucket, &newUsers, &activated, &active, &paying, &firstPaid, &repurchase, &gross, &refunded, &consumed, &supplier, &hasRows); err != nil {
			return nil, false, err
		}
		result = append(result, service.BusinessTrendPoint{Bucket: bucket.In(filter.DateFrom.Location()).Format(businessBucketFormat(filter.Granularity)), NewUsers: newUsers, ActivatedUsers: activated, ActiveUsers: active, PayingUsers: paying, FirstPaidUsers: firstPaid, RepurchaseUsers: repurchase, GrossPaidCNY: gross, RefundedCNY: refunded, NetPaidCNY: gross - refunded, ConsumedRevenue: consumed, SupplierCost: supplier})
	}
	return result, hasRows, rows.Err()
}

func (r *businessAnalyticsRepository) trend(ctx context.Context, filter service.BusinessAnalyticsFilter, usdToCNY float64) ([]service.BusinessTrendPoint, error) {
	if cached, available, err := r.bucketTrend(ctx, filter, usdToCNY); err != nil {
		return nil, err
	} else if available {
		return cached, nil
	}
	end := filter.DateTo
	if now := time.Now().In(filter.DateTo.Location()); now.Before(end) {
		end = now
	}
	query := `WITH buckets AS (
		SELECT generate_series(date_trunc($3,timezone('Asia/Shanghai',$1::timestamptz)),date_trunc($3,timezone('Asia/Shanghai',$2::timestamptz-interval '1 second')),$4::interval) bucket
	), scoped_users AS (
		SELECT u.id,u.created_at,buf.first_activated_at,buf.first_paid_at,buf.channel
		FROM users u LEFT JOIN business_user_facts buf ON buf.user_id=u.id
		WHERE ($6=0 OR EXISTS (
			SELECT 1 FROM distribution_customer_bindings b JOIN distribution_agents a ON a.id=b.agent_id
			WHERE b.user_id=u.id AND (a.id=$6 OR ($7='team' AND a.parent_agent_id=$6))))
	), user_events AS (
		SELECT date_trunc($3,timezone('Asia/Shanghai',su.created_at)) bucket,COUNT(*) new_users,0::bigint activated_users
		FROM scoped_users su WHERE su.created_at>=$1 AND su.created_at<$2 AND ($5='' OR su.channel=$5) GROUP BY 1
		UNION ALL
		SELECT date_trunc($3,timezone('Asia/Shanghai',su.first_activated_at)) bucket,0::bigint new_users,COUNT(*) activated_users
		FROM scoped_users su WHERE su.first_activated_at>=$1 AND su.first_activated_at<$2 AND ($5='' OR su.channel=$5) GROUP BY 1
	), users_by_bucket AS (
		SELECT bucket,SUM(new_users) new_users,SUM(activated_users) activated_users FROM user_events GROUP BY bucket
	), payments AS (
		SELECT date_trunc($3,timezone('Asia/Shanghai',bpf.paid_at)) bucket,
			COUNT(DISTINCT po.user_id) FILTER (WHERE bpf.gross_amount>bpf.refunded_amount) paying_users,
			COUNT(DISTINCT po.user_id) FILTER (WHERE date_trunc($3,timezone('Asia/Shanghai',su.first_paid_at))=date_trunc($3,timezone('Asia/Shanghai',bpf.paid_at))) first_paid_users,
			COUNT(DISTINCT po.user_id) FILTER (WHERE bpf.gross_amount>bpf.refunded_amount AND su.first_paid_at < date_trunc($3,timezone('Asia/Shanghai',bpf.paid_at)) AT TIME ZONE 'Asia/Shanghai') repurchase_users,
			COALESCE(SUM(bpf.gross_amount*bpf.fx_rate_to_cny) FILTER (WHERE bpf.fx_rate_to_cny IS NOT NULL),0) gross_paid,
			0::numeric refunded
			FROM business_payment_facts bpf JOIN payment_orders po ON po.id=bpf.payment_order_id JOIN scoped_users su ON su.id=po.user_id
			WHERE bpf.paid_at>=$1 AND bpf.paid_at<$2 AND ($5='' OR bpf.channel=$5) GROUP BY 1
	), refunds AS (
		SELECT date_trunc($3,timezone('Asia/Shanghai',bpf.refunded_at)) bucket,
			COALESCE(SUM(bpf.refunded_amount*bpf.fx_rate_to_cny) FILTER (WHERE bpf.fx_rate_to_cny IS NOT NULL),0) refunded
			FROM business_payment_facts bpf JOIN payment_orders po ON po.id=bpf.payment_order_id JOIN scoped_users su ON su.id=po.user_id
			WHERE bpf.refunded_at>=$1 AND bpf.refunded_at<$2 AND ($5='' OR bpf.channel=$5) GROUP BY 1
		), usage_source AS (
			SELECT fact.user_id,fact.bucket_start event_at,fact.consumed_revenue consumed,fact.supplier_cost supplier
			FROM business_usage_hourly_facts fact
			WHERE fact.bucket_start>=$1 AND fact.bucket_start<date_trunc('hour',timezone('Asia/Shanghai',$2::timestamptz)) AT TIME ZONE 'Asia/Shanghai'
			UNION ALL
			SELECT ul.user_id,date_trunc('hour',timezone('Asia/Shanghai',ul.created_at)) AT TIME ZONE 'Asia/Shanghai',
				COALESCE(SUM(ul.actual_cost),0),COALESCE(SUM(COALESCE(ul.account_stats_cost,ul.total_cost)*COALESCE(ul.account_rate_multiplier,1)),0)
			FROM usage_logs ul
			WHERE ul.actual_cost>0 AND ul.created_at>=GREATEST($1,date_trunc('hour',timezone('Asia/Shanghai',$2::timestamptz)) AT TIME ZONE 'Asia/Shanghai') AND ul.created_at<$2
			GROUP BY ul.user_id,2
		), usage AS (
			SELECT date_trunc($3,timezone('Asia/Shanghai',source.event_at)) bucket,
				COUNT(DISTINCT source.user_id) active_users,COALESCE(SUM(source.consumed),0) consumed,COALESCE(SUM(source.supplier),0) supplier
			FROM usage_source source JOIN scoped_users su ON su.id=source.user_id
			WHERE ($5='' OR su.channel=$5) GROUP BY 1
	)
	SELECT b.bucket,
		COALESCE(u.new_users,0),COALESCE(u.activated_users,0),COALESCE(x.active_users,0),
		COALESCE((SELECT paying_users FROM payments p WHERE p.bucket=b.bucket),0),
		COALESCE((SELECT first_paid_users FROM payments p WHERE p.bucket=b.bucket),0),
		COALESCE((SELECT repurchase_users FROM payments p WHERE p.bucket=b.bucket),0),
		COALESCE((SELECT gross_paid FROM payments p WHERE p.bucket=b.bucket),0),
		COALESCE((SELECT refunded FROM refunds r WHERE r.bucket=b.bucket),0),
		COALESCE(x.consumed,0),COALESCE(x.supplier,0)
	FROM buckets b
	LEFT JOIN users_by_bucket u ON u.bucket=b.bucket
	LEFT JOIN usage x ON x.bucket=b.bucket
	ORDER BY b.bucket`
	rows, err := r.db.QueryContext(ctx, query, filter.DateFrom, end, filter.Granularity, businessInterval(filter.Granularity), filter.Channel, filter.AgentID, filter.AgentScope)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make([]service.BusinessTrendPoint, 0)
	for rows.Next() {
		var local time.Time
		var newUsers, activated, active, paying, firstPaid, repurchase int64
		var gross, refunded, consumed, supplier float64
		if err = rows.Scan(&local, &newUsers, &activated, &active, &paying, &firstPaid, &repurchase, &gross, &refunded, &consumed, &supplier); err != nil {
			return nil, err
		}
		result = append(result, service.BusinessTrendPoint{Bucket: local.Format(businessBucketFormat(filter.Granularity)), NewUsers: newUsers, ActivatedUsers: activated, FirstPaidUsers: firstPaid, PayingUsers: paying, RepurchaseUsers: repurchase, ActiveUsers: active, GrossPaidCNY: gross, RefundedCNY: refunded, NetPaidCNY: gross - refunded, ConsumedRevenue: consumed, SupplierCost: supplier})
	}
	return result, rows.Err()
}

func (r *businessAnalyticsRepository) channels(ctx context.Context, filter service.BusinessAnalyticsFilter, usdToCNY float64) ([]service.BusinessChannelMetric, error) {
	_ = usdToCNY // Cash facts already contain their event-time CNY conversion.
	agentFilter := businessAgentFilter("u", 3, 4)
	rows, err := r.db.QueryContext(ctx, `WITH channels(channel,position) AS (VALUES
		('distribution',1),('affiliate',2),('campaign',3),('organic',4),('unknown',5)
	), scoped_users AS (
		SELECT u.id,u.created_at,COALESCE(buf.channel,'unknown') channel,buf.first_activated_at,buf.first_paid_at
		FROM users u LEFT JOIN business_user_facts buf ON buf.user_id=u.id
		WHERE `+agentFilter+`
	), user_metrics AS (
		SELECT channel,
			COUNT(*) FILTER(WHERE created_at>=$1 AND created_at<$2) new_users,
			COUNT(*) FILTER(WHERE first_activated_at>=$1 AND first_activated_at<$2) activated_users,
			COUNT(*) FILTER(WHERE first_paid_at>=$1 AND first_paid_at<$2) first_paid_users
		FROM scoped_users GROUP BY channel
	), paid_users AS (
		SELECT fact.channel,po.user_id,BOOL_OR(EXISTS(
			SELECT 1 FROM payment_orders old_po
			JOIN business_payment_facts old_fact ON old_fact.payment_order_id=old_po.id
				WHERE old_po.user_id=po.user_id AND old_fact.paid_at<$1 AND old_fact.gross_amount>old_fact.refunded_amount
		)) repurchased
		FROM business_payment_facts fact JOIN payment_orders po ON po.id=fact.payment_order_id
		JOIN scoped_users su ON su.id=po.user_id
		WHERE fact.paid_at>=$1 AND fact.paid_at<$2 AND fact.gross_amount>fact.refunded_amount GROUP BY fact.channel,po.user_id
	), payment_metrics AS (
		SELECT channel,COUNT(*) paying_users,COUNT(*) FILTER(WHERE repurchased) repurchase_users
		FROM paid_users GROUP BY channel
		), active_metrics AS (
			SELECT su.channel,COUNT(DISTINCT fact.user_id) active_users
			FROM business_usage_hourly_facts fact JOIN scoped_users su ON su.id=fact.user_id
			WHERE fact.bucket_start>=$1 AND fact.bucket_start<$2 GROUP BY su.channel
	), cash_events AS (
		SELECT fact.channel,
			COALESCE(SUM(fact.gross_amount*fact.fx_rate_to_cny) FILTER(WHERE fact.paid_at>=$1 AND fact.paid_at<$2 AND fact.fx_rate_to_cny IS NOT NULL),0) gross,
			COALESCE(SUM(fact.refunded_amount*fact.fx_rate_to_cny) FILTER(WHERE fact.refunded_at>=$1 AND fact.refunded_at<$2 AND fact.fx_rate_to_cny IS NOT NULL),0) refunded
		FROM business_payment_facts fact JOIN payment_orders po ON po.id=fact.payment_order_id
		JOIN scoped_users su ON su.id=po.user_id
		WHERE (fact.paid_at>=$1 AND fact.paid_at<$2) OR (fact.refunded_at>=$1 AND fact.refunded_at<$2)
		GROUP BY fact.channel
	), mature_cohort AS (
		SELECT id,channel,date_trunc('day',timezone('Asia/Shanghai',created_at)) cohort_day
		FROM scoped_users WHERE created_at>=$1 AND created_at<($2::timestamptz-interval '7 day')
	), retention AS (
		SELECT channel,COUNT(*) cohort_users,COUNT(*) FILTER(WHERE EXISTS(
				SELECT 1 FROM business_usage_hourly_facts fact WHERE fact.user_id=mature_cohort.id
				AND date_trunc('day',timezone('Asia/Shanghai',fact.bucket_start))=mature_cohort.cohort_day+interval '7 day'
		)) retained_users FROM mature_cohort GROUP BY channel
	)
	SELECT c.channel,COALESCE(u.new_users,0),COALESCE(u.activated_users,0),COALESCE(u.first_paid_users,0),
		COALESCE(p.paying_users,0),COALESCE(p.repurchase_users,0),COALESCE(a.active_users,0),
		COALESCE(x.gross,0)-COALESCE(x.refunded,0) net_paid,
		CASE WHEN COALESCE(p.paying_users,0)>0 THEN (COALESCE(x.gross,0)-COALESCE(x.refunded,0))/p.paying_users ELSE 0 END arppu,
		CASE WHEN COALESCE(rt.cohort_users,0)>0 THEN rt.retained_users*100.0/rt.cohort_users ELSE 0 END d7
	FROM channels c LEFT JOIN user_metrics u ON u.channel=c.channel
	LEFT JOIN payment_metrics p ON p.channel=c.channel LEFT JOIN active_metrics a ON a.channel=c.channel
	LEFT JOIN cash_events x ON x.channel=c.channel LEFT JOIN retention rt ON rt.channel=c.channel
	ORDER BY c.position`, filter.DateFrom, filter.DateTo, filter.AgentID, filter.AgentScope)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make([]service.BusinessChannelMetric, 0, 5)
	for rows.Next() {
		var item service.BusinessChannelMetric
		if err = rows.Scan(&item.Channel, &item.NewUsers, &item.ActivatedUsers, &item.FirstPaidUsers,
			&item.PayingUsers, &item.RepurchaseUsers, &item.ActiveUsers, &item.NetPaidCNY,
			&item.ARPPUCNY, &item.D7RetentionRate); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *businessAnalyticsRepository) cohorts(ctx context.Context, filter service.BusinessAnalyticsFilter, kind string) ([]service.BusinessCohortRow, error) {
	origin := "u.created_at"
	if kind == "activation" {
		origin = "buf.first_activated_at"
	}
	query := fmt.Sprintf(`WITH cohort AS (SELECT u.id,date_trunc('day',timezone('Asia/Shanghai',%s)) cohort_day FROM users u LEFT JOIN business_user_facts buf ON buf.user_id=u.id WHERE %s IS NOT NULL
	AND ($3='' OR COALESCE(buf.channel,'unknown')=$3)
	AND ($4=0 OR EXISTS (SELECT 1 FROM distribution_customer_bindings b JOIN distribution_agents da ON da.id=b.agent_id WHERE b.user_id=u.id AND (da.id=$4 OR ($5='team' AND da.parent_agent_id=$4))))),
	active AS (SELECT DISTINCT fact.user_id,date_trunc('day',timezone('Asia/Shanghai',fact.bucket_start)) active_day FROM business_usage_hourly_facts fact
			WHERE fact.bucket_start >= $1 AND fact.bucket_start < LEAST($2::timestamptz+interval '31 day',NOW())),
	rows AS (SELECT cohort_day,COUNT(*) users,COUNT(*) FILTER(WHERE EXISTS(SELECT 1 FROM active a WHERE a.user_id=cohort.id AND a.active_day=cohort_day+interval '1 day')) d1,
	COUNT(*) FILTER(WHERE EXISTS(SELECT 1 FROM active a WHERE a.user_id=cohort.id AND a.active_day=cohort_day+interval '7 day')) d7,
	COUNT(*) FILTER(WHERE EXISTS(SELECT 1 FROM active a WHERE a.user_id=cohort.id AND a.active_day=cohort_day+interval '30 day')) d30 FROM cohort
	WHERE cohort_day >= timezone('Asia/Shanghai',$1) AND cohort_day < timezone('Asia/Shanghai',$2) GROUP BY cohort_day)
	SELECT cohort_day,users,d1,d7,d30 FROM rows ORDER BY cohort_day DESC LIMIT 31`, origin, origin)
	rows, err := r.db.QueryContext(ctx, query, filter.DateFrom, filter.DateTo, filter.Channel, filter.AgentID, filter.AgentScope)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make([]service.BusinessCohortRow, 0)
	now := time.Now().In(filter.DateFrom.Location())
	for rows.Next() {
		var day time.Time
		var users, d1, d7, d30 int64
		if err = rows.Scan(&day, &users, &d1, &d7, &d30); err != nil {
			return nil, err
		}
		item := service.BusinessCohortRow{CohortDate: day.Format("2006-01-02"), Users: users}
		age := int(now.Sub(day).Hours() / 24)
		if age >= 1 {
			x := float64(d1) * 100 / float64(users)
			item.D1 = &x
		}
		if age >= 7 {
			x := float64(d7) * 100 / float64(users)
			item.D7 = &x
		}
		if age >= 30 {
			x := float64(d30) * 100 / float64(users)
			item.D30 = &x
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *businessAnalyticsRepository) lifecycle(ctx context.Context, filter service.BusinessAnalyticsFilter) (service.BusinessLifecycle, error) {
	var result service.BusinessLifecycle
	channelFilter := `($3='' OR ` + businessChannelExpr + `=$3)`
	agentFilter := businessAgentFilter("u", 4, 7)
	err := r.db.QueryRowContext(ctx, `WITH scoped AS (
		SELECT u.id,u.created_at,buf.first_activated_at first_active FROM users u LEFT JOIN business_user_facts buf ON buf.user_id=u.id
		WHERE u.created_at<$2 AND `+channelFilter+` AND `+agentFilter+`
	), usage_summary AS (
		SELECT fact.user_id,
			MIN(fact.first_used_at) FILTER(WHERE fact.first_used_at>=$1 AND fact.first_used_at<$2) first_current,
			BOOL_OR(fact.first_used_at<$6 AND fact.last_used_at>=$5) previous_active,
			MAX(fact.last_used_at) FILTER(WHERE fact.first_used_at<$1) last_before
		FROM business_usage_hourly_facts fact JOIN scoped s ON s.id=fact.user_id
		WHERE fact.first_used_at<$2
		GROUP BY fact.user_id
	), activity AS (
		SELECT s.id,s.created_at,s.first_active,summary.first_current,
			COALESCE(summary.previous_active,FALSE) previous_active,summary.last_before
		FROM scoped s LEFT JOIN usage_summary summary ON summary.user_id=s.id
	), classified AS (
		SELECT CASE
			WHEN created_at>=$1 THEN 'new'
			WHEN first_active IS NULL OR first_active>=$2 THEN 'unactivated'
			WHEN first_active>=$1 THEN 'newly_activated'
			WHEN first_current IS NOT NULL AND previous_active THEN 'continuously_active'
			WHEN first_current IS NOT NULL AND NOT previous_active AND first_current-last_before<=interval '30 day' THEN 'silent_reactivated'
			WHEN first_current IS NOT NULL AND NOT previous_active THEN 'churned_reactivated'
			WHEN first_current IS NULL AND $2-last_before<=interval '30 day' THEN 'silent'
			ELSE 'churned' END segment
		FROM activity
	)
	SELECT COUNT(*) FILTER(WHERE segment='new'),COUNT(*) FILTER(WHERE segment='unactivated'),
		COUNT(*) FILTER(WHERE segment='newly_activated'),COUNT(*) FILTER(WHERE segment='continuously_active'),
		COUNT(*) FILTER(WHERE segment='silent_reactivated'),COUNT(*) FILTER(WHERE segment='churned_reactivated'),
		COUNT(*) FILTER(WHERE segment='silent'),COUNT(*) FILTER(WHERE segment='churned')
	FROM classified`, filter.DateFrom, filter.DateTo, filter.Channel, filter.AgentID,
		filter.DateFrom.AddDate(0, 0, -reportDays(filter)), filter.DateTo.AddDate(0, 0, -reportDays(filter)), filter.AgentScope).
		Scan(&result.New, &result.Unactivated, &result.NewlyActivated, &result.ContinuouslyActive,
			&result.SilentReactivated, &result.ChurnedReactivated, &result.Silent, &result.Churned)
	return result, err
}

func reportDays(filter service.BusinessAnalyticsFilter) int {
	days := int(filter.ReportDateTo.Sub(filter.DateFrom).Hours()/24) + 1
	if days < 1 {
		return 1
	}
	return days
}

func (r *businessAnalyticsRepository) durationDistribution(ctx context.Context, filter service.BusinessAnalyticsFilter) ([]service.BusinessDurationBucket, error) {
	channelFilter := `($3='' OR ` + businessChannelExpr + `=$3)`
	agentFilter := businessAgentFilter("u", 4, 5)
	rows, err := r.db.QueryContext(ctx, `WITH scoped AS (
		SELECT u.created_at,buf.first_activated_at,buf.first_paid_at
		FROM users u LEFT JOIN business_user_facts buf ON buf.user_id=u.id
		WHERE u.created_at>=$1 AND u.created_at<$2 AND `+channelFilter+` AND `+agentFilter+`
	), ranges(key,min_age,max_age,position) AS (VALUES
		('under_1h',interval '0',interval '1 hour',1),
		('1h_24h',interval '1 hour',interval '1 day',2),
		('1d_3d',interval '1 day',interval '3 day',3),
		('3d_7d',interval '3 day',interval '7 day',4),
		('over_7d',interval '7 day',NULL,5)
	) SELECT r.key,
		COUNT(*) FILTER (WHERE s.first_activated_at IS NOT NULL AND s.first_activated_at-s.created_at>=r.min_age AND (r.max_age IS NULL OR s.first_activated_at-s.created_at<r.max_age)),
		COUNT(*) FILTER (WHERE s.first_paid_at IS NOT NULL AND s.first_paid_at-s.created_at>=r.min_age AND (r.max_age IS NULL OR s.first_paid_at-s.created_at<r.max_age))
	FROM ranges r CROSS JOIN scoped s GROUP BY r.key,r.position ORDER BY r.position`, filter.DateFrom, filter.DateTo, filter.Channel, filter.AgentID, filter.AgentScope)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make([]service.BusinessDurationBucket, 0, 5)
	for rows.Next() {
		var item service.BusinessDurationBucket
		if err = rows.Scan(&item.Key, &item.Activation, &item.FirstPaid); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *businessAnalyticsRepository) currencyBreakdown(ctx context.Context, filter service.BusinessAnalyticsFilter) ([]service.BusinessCurrencyBreakdown, error) {
	agentFilter := businessAgentFilter("u", 4, 5)
	rows, err := r.db.QueryContext(ctx, `SELECT bpf.currency,
		COALESCE(SUM(bpf.gross_amount) FILTER (WHERE bpf.paid_at>=$1 AND bpf.paid_at<$2),0),
		COALESCE(SUM(bpf.refunded_amount) FILTER (WHERE bpf.refunded_at>=$1 AND bpf.refunded_at<$2),0)
		FROM business_payment_facts bpf JOIN payment_orders po ON po.id=bpf.payment_order_id
		JOIN users u ON u.id=po.user_id LEFT JOIN business_user_facts buf ON buf.user_id=u.id
		WHERE ((bpf.paid_at>=$1 AND bpf.paid_at<$2) OR (bpf.refunded_at>=$1 AND bpf.refunded_at<$2))
		AND ($3='' OR bpf.channel=$3) AND `+agentFilter+` GROUP BY bpf.currency ORDER BY bpf.currency`, filter.DateFrom, filter.DateTo, filter.Channel, filter.AgentID, filter.AgentScope)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make([]service.BusinessCurrencyBreakdown, 0)
	for rows.Next() {
		var item service.BusinessCurrencyBreakdown
		if err = rows.Scan(&item.Currency, &item.Gross, &item.Refunded); err != nil {
			return nil, err
		}
		item.Net = item.Gross - item.Refunded
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *businessAnalyticsRepository) GetBusinessAnalytics(ctx context.Context, filter service.BusinessAnalyticsFilter) (*service.BusinessAnalyticsSnapshot, error) {
	usdToCNY, rateEstimated := r.effectiveUSDToCNY(ctx)
	days := reportDays(filter)
	previousStart := filter.DateFrom.AddDate(0, 0, -days)
	previousEnd := filter.DateTo.AddDate(0, 0, -days)
	var current, previous businessPeriodValues
	periodGroup, periodCtx := errgroup.WithContext(ctx)
	periodGroup.Go(func() error {
		var err error
		current, err = r.periodValues(periodCtx, filter.DateFrom, filter.DateTo, filter.Channel, filter.AgentID, filter.AgentScope, usdToCNY, rateEstimated)
		return err
	})
	periodGroup.Go(func() error {
		var err error
		previous, err = r.periodValues(periodCtx, previousStart, previousEnd, filter.Channel, filter.AgentID, filter.AgentScope, usdToCNY, rateEstimated)
		return err
	})
	if err := periodGroup.Wait(); err != nil {
		return nil, fmt.Errorf("business analytics period metrics: %w", err)
	}
	net, prevNet := current.gross-current.refunded, previous.gross-previous.refunded
	margin, prevMargin := current.consumed-current.supplier, previous.consumed-previous.supplier
	rate := func(n, d int64) float64 {
		if d <= 0 {
			return 0
		}
		return float64(n) * 100 / float64(d)
	}
	floatRate := func(n, d float64) float64 {
		if d == 0 {
			return 0
		}
		return n * 100 / d
	}
	metrics := map[string]service.BusinessMetric{
		"new_users":            businessMetric(float64(current.newUsers), float64(previous.newUsers), "", false),
		"activated_users":      businessMetric(float64(current.activated), float64(previous.activated), "", false),
		"first_paid_users":     businessMetric(float64(current.firstPaid), float64(previous.firstPaid), "", false),
		"paying_users":         businessMetric(float64(current.paying), float64(previous.paying), "", false),
		"repurchase_users":     businessMetric(float64(current.repurchase), float64(previous.repurchase), "", false),
		"active_users":         businessMetric(float64(current.active), float64(previous.active), "", false),
		"activation_rate":      businessRateMetric(rate(current.cohortActivated, current.newUsers), rate(previous.cohortActivated, previous.newUsers), false),
		"paid_conversion_rate": businessRateMetric(rate(current.cohortFirstPaid, current.newUsers), rate(previous.cohortFirstPaid, previous.newUsers), false),
		"repurchase_rate":      businessRateMetric(rate(current.repurchase, current.paying), rate(previous.repurchase, previous.paying), false),
		"active_rate":          businessRateMetric(rate(current.active, current.population), rate(previous.active, previous.population), false),
		"gross_paid":           businessMetric(current.gross, previous.gross, "CNY", current.estimated),
		"refunded":             businessMetric(current.refunded, previous.refunded, "CNY", current.estimated),
		"net_paid":             businessMetric(net, prevNet, "CNY", current.estimated),
		"refund_rate":          businessRateMetric(floatRate(current.refunded, current.gross), floatRate(previous.refunded, previous.gross), current.estimated),
		"arpu": businessMetric(func() float64 {
			if current.active == 0 {
				return 0
			}
			return net / float64(current.active)
		}(), func() float64 {
			if previous.active == 0 {
				return 0
			}
			return prevNet / float64(previous.active)
		}(), "CNY", current.estimated),
		"arppu": businessMetric(func() float64 {
			if current.paying == 0 {
				return 0
			}
			return net / float64(current.paying)
		}(), func() float64 {
			if previous.paying == 0 {
				return 0
			}
			return prevNet / float64(previous.paying)
		}(), "CNY", current.estimated),
		"consumed_revenue": businessMetric(current.consumed, previous.consumed, "USD", false), "supplier_cost": businessMetric(current.supplier, previous.supplier, "USD", false), "consumption_profit": businessMetric(margin, prevMargin, "USD", false), "consumption_margin_rate": businessRateMetric(floatRate(margin, current.consumed), floatRate(prevMargin, previous.consumed), false),
		"commission_cost_rate": businessRateMetric(floatRate(current.commission, net), floatRate(previous.commission, prevNet), false), "reward_cost_rate": businessRateMetric(floatRate(current.rewards, net), floatRate(previous.rewards, prevNet), false),
	}
	base := &service.BusinessAnalyticsSnapshot{
		DateFrom: filter.DateFrom.Format("2006-01-02"), DateTo: filter.ReportDateTo.Format("2006-01-02"),
		PreviousDateFrom: previousStart.Format("2006-01-02"), PreviousDateTo: service.BusinessAnalyticsInclusiveDate(previousEnd),
		Timezone: filter.Timezone, Granularity: filter.Granularity, Currency: "CNY", UpdatedAt: time.Now().UTC(),
		Estimated: current.estimated, Metrics: metrics, Warnings: make([]string, 0),
	}
	var usageDataFrom sql.NullTime
	if err := r.db.QueryRowContext(ctx, `SELECT usage_coverage_from FROM business_analytics_aggregation_state WHERE id=1`).Scan(&usageDataFrom); err != nil {
		return nil, fmt.Errorf("business analytics usage coverage: %w", err)
	}
	if usageDataFrom.Valid {
		coverageDay := time.Date(usageDataFrom.Time.In(filter.DateFrom.Location()).Year(), usageDataFrom.Time.In(filter.DateFrom.Location()).Month(), usageDataFrom.Time.In(filter.DateFrom.Location()).Day(), 0, 0, 0, 0, filter.DateFrom.Location())
		base.UsageDataFrom = coverageDay.Format("2006-01-02")
		if filter.DateFrom.Before(coverageDay) {
			base.Warnings = append(base.Warnings, "usage_history_incomplete")
		}
	}
	if current.estimated {
		base.Warnings = append(base.Warnings, "estimated_fx")
	}
	if filter.SummaryOnly {
		base.Trend = []service.BusinessTrendPoint{}
		base.PreviousTrend = []service.BusinessTrendPoint{}
		base.Funnel = []service.BusinessFunnelStep{}
		base.Channels = []service.BusinessChannelMetric{}
		base.RegistrationCohorts = []service.BusinessCohortRow{}
		base.ActivationCohorts = []service.BusinessCohortRow{}
		base.DurationDistribution = []service.BusinessDurationBucket{}
		base.CurrencyBreakdown = []service.BusinessCurrencyBreakdown{}
		return base, nil
	}
	trend := []service.BusinessTrendPoint{}
	previousTrend := []service.BusinessTrendPoint{}
	channels := []service.BusinessChannelMetric{}
	registration, activation := []service.BusinessCohortRow{}, []service.BusinessCohortRow{}
	var lifecycle service.BusinessLifecycle
	durations := []service.BusinessDurationBucket{}
	currencies := []service.BusinessCurrencyBreakdown{}
	sectionGroup, sectionCtx := errgroup.WithContext(ctx)
	run := func(enabled bool, label string, fn func(context.Context) error) {
		if enabled {
			sectionGroup.Go(func() error {
				if err := fn(sectionCtx); err != nil {
					return fmt.Errorf("business analytics %s: %w", label, err)
				}
				return nil
			})
		}
	}
	section := filter.Section
	run(section == "" || section == "overview" || section == "growth" || section == "finance" || section == "channels", "trend", func(queryCtx context.Context) (err error) {
		trend, err = r.trend(queryCtx, filter, usdToCNY)
		return err
	})
	run(section == "" || section == "overview" || section == "growth" || section == "finance", "previous trend", func(queryCtx context.Context) (err error) {
		previousFilter := filter
		previousFilter.DateFrom, previousFilter.DateTo = previousStart, previousEnd
		previousFilter.ReportDateTo = previousEnd
		previousTrend, err = r.trend(queryCtx, previousFilter, usdToCNY)
		return err
	})
	run(section == "" || section == "overview" || section == "growth" || section == "channels", "channels", func(queryCtx context.Context) (err error) {
		channels, err = r.channels(queryCtx, filter, usdToCNY)
		return err
	})
	run(section == "" || section == "retention", "registration cohorts", func(queryCtx context.Context) (err error) {
		registration, err = r.cohorts(queryCtx, filter, "registration")
		return err
	})
	run(section == "" || section == "retention", "activation cohorts", func(queryCtx context.Context) (err error) {
		activation, err = r.cohorts(queryCtx, filter, "activation")
		return err
	})
	run(section == "" || section == "retention", "lifecycle", func(queryCtx context.Context) (err error) {
		lifecycle, err = r.lifecycle(queryCtx, filter)
		return err
	})
	run(section == "" || section == "growth", "duration distribution", func(queryCtx context.Context) (err error) {
		durations, err = r.durationDistribution(queryCtx, filter)
		return err
	})
	run(section == "" || section == "finance", "currency breakdown", func(queryCtx context.Context) (err error) {
		currencies, err = r.currencyBreakdown(queryCtx, filter)
		return err
	})
	if err := sectionGroup.Wait(); err != nil {
		return nil, err
	}
	base.Trend = trend
	base.PreviousTrend = previousTrend
	base.Funnel = []service.BusinessFunnelStep{{Key: "registered", Count: current.newUsers}, {Key: "activated", Count: current.cohortActivated}, {Key: "first_paid", Count: current.cohortFirstPaid}, {Key: "repurchased", Count: current.cohortRepurchased}}
	base.Channels = channels
	base.RegistrationCohorts = registration
	base.ActivationCohorts = activation
	base.Lifecycle = lifecycle
	base.DurationDistribution = durations
	base.CurrencyBreakdown = currencies
	return base, nil
}
