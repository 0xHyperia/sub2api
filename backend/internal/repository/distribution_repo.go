package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type distributionRepository struct{ db *sql.DB }

func NewDistributionRepository(db *sql.DB) service.DistributionRepository {
	return &distributionRepository{db: db}
}

func (r *distributionRepository) ValidatePromotionCode(ctx context.Context, code string) error {
	var valid bool
	err := r.db.QueryRowContext(ctx, `SELECT ds.enabled AND EXISTS (
		SELECT 1 FROM distribution_agents a
		JOIN distribution_agent_levels l ON l.id=a.level_id AND l.is_active=TRUE
		LEFT JOIN distribution_agents p ON p.id=a.parent_agent_id
		WHERE a.promotion_code=$1 AND a.status='active' AND (l.depth=1 OR p.status='active')
	) FROM distribution_settings ds WHERE ds.id=1`, code).Scan(&valid)
	if err != nil {
		return err
	}
	if !valid {
		return service.ErrDistributionCodeInvalid
	}
	return nil
}

func (r *distributionRepository) BindCustomerByCode(ctx context.Context, userID int64, code string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('promotion-ownership:' || $1::text,0))`, userID); err != nil {
		return fmt.Errorf("lock promotion ownership: %w", err)
	}

	var enabled bool
	if err = tx.QueryRowContext(ctx, `SELECT enabled FROM distribution_settings WHERE id = 1`).Scan(&enabled); err != nil {
		return err
	}
	if !enabled {
		return service.ErrDistributionCodeInvalid
	}

	var agentID int64
	err = tx.QueryRowContext(ctx, `
		SELECT a.id FROM distribution_agents a
		JOIN distribution_agent_levels l ON l.id = a.level_id
		LEFT JOIN distribution_agents p ON p.id = a.parent_agent_id
		WHERE a.promotion_code = $1 AND a.status = 'active' AND l.is_active = TRUE
		  AND (l.depth = 1 OR (l.depth = 2 AND p.status = 'active'))`, code).Scan(&agentID)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrDistributionCodeInvalid
	}
	if err != nil {
		return err
	}

	var blocked bool
	if err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM distribution_agents WHERE user_id = $1)
		    OR EXISTS (SELECT 1 FROM user_affiliates WHERE user_id = $1 AND inviter_id IS NOT NULL)`, userID).Scan(&blocked); err != nil {
		return err
	}
	if blocked {
		return service.ErrDistributionCodeConflict
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO distribution_customer_bindings (user_id, agent_id, promotion_code) VALUES ($1,$2,$3)`, userID, agentID, code)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return service.ErrDistributionAlreadyBound
		}
		if errors.As(err, &pqErr) && pqErr.Code == "23514" {
			return service.ErrDistributionCodeConflict
		}
		if errors.As(err, &pqErr) && pqErr.Code == "40001" {
			return service.ErrDistributionCodeConflict
		}
		return fmt.Errorf("bind distribution customer: %w", err)
	}
	attr, ok := service.DistributionRegistrationAttributionFromContext(ctx)
	if !ok || attr.Code != code {
		attr = service.DistributionRegistrationAttribution{Code: code, Type: "direct"}
	}
	if attr.Type != "persisted" {
		attr.Type = "direct"
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_promotion_conversions
		(user_id,agent_id,promotion_code,visitor_token_hash,first_visit_id,last_visit_id,attribution_type,
		 attribution_visit_id,attribution_source,attribution_model,attribution_device_type,attribution_is_bot)
		VALUES($1,$2,$3,NULLIF($4,''),$5,$6,$7,$8,COALESCE(NULLIF($9,''),'直接注册'),COALESCE(NULLIF($10,''),'untracked'),
		 (SELECT device_type FROM distribution_promotion_visits WHERE id=$8 AND agent_id=$2),
		 (SELECT is_bot FROM distribution_promotion_visits WHERE id=$8 AND agent_id=$2))
		ON CONFLICT(user_id) DO NOTHING`, userID, agentID, code, attr.VisitorTokenHash, attr.FirstVisitID, attr.LastVisitID,
		attr.Type, attr.AttributionVisitID, attr.AttributionSource, attr.AttributionModel); err != nil {
		return fmt.Errorf("record distribution conversion: %w", err)
	}
	if attr.VisitorTokenHash != "" {
		if _, err = tx.ExecContext(ctx, `UPDATE distribution_promotion_attributions p SET converted_user_id=$2,updated_at=NOW()
			FROM distribution_agents a WHERE p.agent_id=a.id AND p.visitor_token_hash=$1
			AND p.converted_user_id IS NULL AND p.promotion_code=$3 AND a.id=$4`, attr.VisitorTokenHash, userID, code, agentID); err != nil {
			return fmt.Errorf("complete distribution attribution: %w", err)
		}
	}
	if err = accrueDistributionRewardTx(ctx, tx, userID, "registration", 0, decimal.Zero); err != nil {
		return fmt.Errorf("accrue registration reward: %w", err)
	}
	return tx.Commit()
}

func (r *distributionRepository) AccrueRegistrationReward(ctx context.Context, customerUserID int64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err = accrueDistributionRewardTx(ctx, tx, customerUserID, "registration", 0, decimal.Zero); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *distributionRepository) QueueDistributionBindingClaim(ctx context.Context, userID int64, code, signupSource string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO distribution_binding_claims
		(user_id,promotion_code,signup_source,status,attempts,last_error,updated_at,completed_at)
		VALUES($1,$2,$3,'pending',0,NULL,NOW(),NULL)
		ON CONFLICT(user_id) DO UPDATE SET
			promotion_code=EXCLUDED.promotion_code,
			signup_source=EXCLUDED.signup_source,
			status=CASE WHEN distribution_binding_claims.status='completed' THEN 'completed' ELSE 'pending' END,
			last_error=CASE WHEN distribution_binding_claims.status='completed' THEN distribution_binding_claims.last_error ELSE NULL END,
			updated_at=NOW()
		WHERE distribution_binding_claims.status<>'completed'`, userID, code, signupSource)
	if err != nil {
		return fmt.Errorf("queue distribution binding claim: %w", err)
	}
	return nil
}

func (r *distributionRepository) PendingDistributionBindingClaim(ctx context.Context, userID int64) (string, bool, error) {
	var code string
	err := r.db.QueryRowContext(ctx, `SELECT promotion_code FROM distribution_binding_claims
		WHERE user_id=$1 AND status='pending'`, userID).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("get pending distribution binding claim: %w", err)
	}
	return code, true, nil
}

func (r *distributionRepository) FinishDistributionBindingClaim(ctx context.Context, userID int64, bindErr error) error {
	status, lastError := "completed", ""
	if bindErr != nil {
		status, lastError = "pending", bindErr.Error()
		if errors.Is(bindErr, service.ErrDistributionCodeConflict) || errors.Is(bindErr, service.ErrDistributionAlreadyBound) || errors.Is(bindErr, service.ErrDistributionCodeInvalid) {
			status = "conflict"
		}
	}
	_, err := r.db.ExecContext(ctx, `UPDATE distribution_binding_claims SET status=$2,attempts=attempts+1,
		last_error=NULLIF($3,''),updated_at=NOW(),completed_at=CASE WHEN $2='completed' THEN NOW() ELSE NULL END
		WHERE user_id=$1`, userID, status, lastError)
	if err != nil {
		return fmt.Errorf("finish distribution binding claim: %w", err)
	}
	return nil
}

const distributionAgentSelect = `
	SELECT a.id, a.user_id, a.level_id, l.depth, a.parent_agent_id, a.promotion_code,
	       COALESCE(a.rate_override_bps, l.default_rate_bps), l.max_child_rate_bps, a.can_recruit_subagents, a.can_view_promotion_stats, a.status,
	       COALESCE(w.available_cny, 0), COALESCE(w.frozen_cny, 0), COALESCE(w.reserved_cny, 0),
	       COALESCE(w.debt_cny, 0), COALESCE(w.total_earned_cny, 0), COALESCE(w.total_withdrawn_cny, 0)
	FROM distribution_agents a
	JOIN distribution_agent_levels l ON l.id = a.level_id
	LEFT JOIN distribution_wallets w ON w.agent_id = a.id`

func scanDistributionAgent(row *sql.Row) (*service.DistributionAgent, error) {
	var a service.DistributionAgent
	err := row.Scan(&a.ID, &a.UserID, &a.LevelID, &a.Depth, &a.ParentAgentID, &a.PromotionCode,
		&a.EffectiveRateBPS, &a.MaxChildRateBPS, &a.CanRecruitSubagents, &a.CanViewPromotionStats, &a.Status, &a.AvailableCNY, &a.FrozenCNY,
		&a.ReservedCNY, &a.DebtCNY, &a.TotalEarnedCNY, &a.TotalWithdrawnCNY)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrDistributionNotAgent
	}
	return &a, err
}

func (r *distributionRepository) GetAgentByUserID(ctx context.Context, userID int64) (*service.DistributionAgent, error) {
	if err := r.releaseMatured(ctx, userID); err != nil {
		return nil, err
	}
	return scanDistributionAgent(r.db.QueryRowContext(ctx, distributionAgentSelect+` WHERE a.user_id = $1`, userID))
}

func (r *distributionRepository) IsAgent(ctx context.Context, userID int64) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM distribution_agents WHERE user_id=$1 AND status <> 'revoked')`, userID).Scan(&ok)
	return ok, err
}

func (r *distributionRepository) AdminGetSettings(ctx context.Context) (*service.DistributionSettings, error) {
	var s service.DistributionSettings
	err := r.db.QueryRowContext(ctx, `SELECT s.enabled,
		(SELECT default_rate_bps FROM distribution_agent_levels WHERE depth=1),
		(SELECT max_child_rate_bps FROM distribution_agent_levels WHERE depth=1),
		(SELECT default_rate_bps FROM distribution_agent_levels WHERE depth=2),
		s.freeze_hours,s.withdrawal_enabled,s.withdrawal_dual_approval_enabled,s.minimum_withdrawal_cny,s.maximum_withdrawal_cny,
		s.withdrawal_fee_rate_bps,s.withdrawal_fee_fixed_cny,s.daily_withdrawal_limit_cny,s.monthly_withdrawal_limit_cny,s.cny_per_platform_usd,
		COALESCE((SELECT rate_to_cny FROM distribution_fx_rates WHERE currency='USD'), 7),
		s.promotion_tracking_enabled,s.promotion_attribution_enabled,s.promotion_attribution_days,s.promotion_attribution_model,
		s.promotion_collect_source,s.promotion_collect_device,s.promotion_bot_filter_enabled,s.promotion_detail_retention_days,
		s.registration_reward_enabled,s.recharge_reward_enabled
		FROM distribution_settings s WHERE s.id=1`).Scan(&s.Enabled, &s.L1DefaultRateBPS, &s.L1MaxChildRateBPS, &s.L2DefaultRateBPS,
		&s.FreezeHours, &s.WithdrawalEnabled, &s.WithdrawalDualApproval, &s.MinimumWithdrawalCNY, &s.MaximumWithdrawalCNY, &s.WithdrawalFeeRateBPS,
		&s.WithdrawalFeeFixedCNY, &s.DailyWithdrawalLimitCNY, &s.MonthlyWithdrawalLimitCNY, &s.CNYPerPlatformUSD, &s.USDToCNY,
		&s.PromotionTrackingEnabled, &s.PromotionAttributionEnabled, &s.PromotionAttributionDays, &s.PromotionAttributionModel,
		&s.PromotionCollectSource, &s.PromotionCollectDevice, &s.PromotionBotFilterEnabled, &s.PromotionDetailRetentionDays,
		&s.RegistrationRewardEnabled, &s.RechargeRewardEnabled)
	return &s, err
}

func (r *distributionRepository) AdminGetOverview(ctx context.Context, filter service.DistributionAnalyticsFilter) (*service.DistributionAdminOverview, error) {
	if err := r.releaseAllMatured(ctx); err != nil {
		return nil, err
	}
	var overview service.DistributionAdminOverview
	err := r.db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM distribution_agents),
		(SELECT COUNT(*) FROM distribution_agents WHERE status='active'),
		(SELECT COUNT(*) FROM distribution_agents WHERE status='suspended'),
		(SELECT COUNT(*) FROM distribution_agents WHERE created_at>=DATE_TRUNC('month',NOW())),
		(SELECT COUNT(*) FROM distribution_customer_bindings),
		(SELECT COUNT(DISTINCT customer_user_id) FROM distribution_commission_sources WHERE status<>'void' AND COALESCE(commission_base_cny,0)>0),
		(SELECT COUNT(*) FROM distribution_customer_bindings WHERE bound_at>=DATE_TRUNC('month',NOW())),
		COALESCE((SELECT SUM(GREATEST(commission_base_cny-refunded_amount_cny,0)) FROM distribution_commission_sources WHERE status<>'void'),0),
		COALESCE((SELECT SUM(GREATEST(original_amount_cny-reversed_amount_cny,0)) FROM distribution_commission_entries),0),
		COALESCE((SELECT SUM(GREATEST(original_amount_cny-reversed_amount_cny,0)) FROM distribution_commission_entries WHERE created_at>=DATE_TRUNC('month',NOW())),0),
		COALESCE((SELECT SUM(reversed_amount_cny) FROM distribution_commission_entries),0),
		COALESCE((SELECT SUM(available_cny) FROM distribution_wallets),0),
		COALESCE((SELECT SUM(frozen_cny) FROM distribution_wallets),0),
		COALESCE((SELECT SUM(reserved_cny) FROM distribution_wallets),0),
		COALESCE((SELECT SUM(debt_cny) FROM distribution_wallets),0),
		(SELECT COUNT(*) FROM distribution_withdrawals WHERE status IN ('pending','approved')),
		COALESCE((SELECT SUM(payout_cny) FROM distribution_withdrawals WHERE status IN ('pending','approved')),0),
		(SELECT COUNT(*) FROM distribution_withdrawals WHERE status='paying'),
		COALESCE((SELECT SUM(payout_cny) FROM distribution_withdrawals WHERE status='paying'),0),
		COALESCE((SELECT SUM(payout_cny) FROM distribution_withdrawals WHERE status='paid' AND paid_at>=DATE_TRUNC('month',NOW())),0),
		(SELECT COUNT(*) FROM distribution_withdrawals WHERE status='pending' AND created_at<NOW()-INTERVAL '24 hours')`).
		Scan(&overview.TotalAgents, &overview.ActiveAgents, &overview.SuspendedAgents, &overview.NewAgentsThisMonth,
			&overview.TotalCustomers, &overview.PayingCustomers, &overview.NewCustomersThisMonth, &overview.CustomerPaidCNY,
			&overview.TotalCommissionCNY, &overview.MonthCommissionCNY, &overview.ReversedCommissionCNY,
			&overview.AvailableCommissionCNY, &overview.FrozenCommissionCNY, &overview.ReservedCommissionCNY, &overview.DebtCommissionCNY,
			&overview.PendingWithdrawals, &overview.PendingWithdrawalCNY, &overview.PayingWithdrawals, &overview.PayingWithdrawalCNY,
			&overview.PaidThisMonthCNY, &overview.OverduePendingCount)
	if err != nil {
		return nil, err
	}
	overview.Analytics, err = r.distributionBusinessAnalytics(ctx, 0, filter)
	if err != nil {
		return nil, err
	}
	overview.AgentRanking, err = r.distributionAgentRanking(ctx, 0, filter, 8)
	if err != nil {
		return nil, err
	}
	scope := newDistributionAnalyticsScope(filter)
	err = r.db.QueryRowContext(ctx, `SELECT
		COALESCE(SUM(total_reward_cny) FILTER (WHERE reward_type='registration'),0),
		COALESCE(SUM(total_reward_cny) FILTER (WHERE reward_type='recharge_threshold'),0)
		FROM distribution_reward_events WHERE created_at>=$1 AND created_at<$2`, scope.currentStart, scope.currentEnd).
		Scan(&overview.PeriodRegistrationRewardCNY, &overview.PeriodRechargeRewardCNY)
	if err != nil {
		return nil, err
	}
	return &overview, nil
}

type distributionAnalyticsScope struct {
	currentStart  time.Time
	currentEnd    time.Time
	previousStart time.Time
}

func newDistributionAnalyticsScope(filter service.DistributionAnalyticsFilter) distributionAnalyticsScope {
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	var currentStart, currentEnd time.Time
	if filter.DateFrom != nil && filter.DateTo != nil {
		from, to := filter.DateFrom.In(location), filter.DateTo.In(location)
		currentStart = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, location)
		currentEnd = time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, location).AddDate(0, 0, 1)
	} else {
		now := time.Now().In(location)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
		currentStart = today.AddDate(0, 0, -(filter.Days - 1))
		currentEnd = today.AddDate(0, 0, 1)
	}
	days := int(currentEnd.Sub(currentStart).Hours() / 24)
	return distributionAnalyticsScope{currentStart: currentStart, currentEnd: currentEnd, previousStart: currentStart.AddDate(0, 0, -days)}
}

func distributionTrendResolution(days int) (string, string) {
	switch {
	case days <= 1:
		return "hour", "1 hour"
	case days <= 31:
		return "day", "1 day"
	case days <= 92:
		return "week", "1 week"
	default:
		return "month", "1 month"
	}
}

func distributionMetricRate(numerator decimal.Decimal, denominator int64) decimal.Decimal {
	if denominator <= 0 {
		return decimal.Zero
	}
	return numerator.Div(decimal.NewFromInt(denominator)).Round(2)
}

func distributionGrowth(current, previous decimal.Decimal) *decimal.Decimal {
	if previous.IsZero() {
		return nil
	}
	value := current.Sub(previous).Div(previous).Mul(decimal.NewFromInt(100)).Round(2)
	return &value
}

func finalizeDistributionMetrics(value *service.DistributionBusinessMetrics) {
	value.ConversionRate = distributionMetricRate(decimal.NewFromInt(value.PayingCustomers).Mul(decimal.NewFromInt(100)), value.NewCustomers)
	value.AverageOrderCNY = distributionMetricRate(value.CustomerPaidCNY, value.PaidOrders)
}

func addDistributionMetrics(left, right service.DistributionBusinessMetrics) service.DistributionBusinessMetrics {
	value := service.DistributionBusinessMetrics{
		NewCustomers:    left.NewCustomers + right.NewCustomers,
		PayingCustomers: left.PayingCustomers + right.PayingCustomers,
		PaidOrders:      left.PaidOrders + right.PaidOrders,
		CustomerPaidCNY: left.CustomerPaidCNY.Add(right.CustomerPaidCNY),
		CommissionCNY:   left.CommissionCNY.Add(right.CommissionCNY),
	}
	finalizeDistributionMetrics(&value)
	return value
}

func finalizeDistributionComparison(value *service.DistributionPeriodComparison) {
	finalizeDistributionMetrics(&value.Current)
	finalizeDistributionMetrics(&value.Previous)
	value.PaidGrowthRate = distributionGrowth(value.Current.CustomerPaidCNY, value.Previous.CustomerPaidCNY)
	value.CommissionGrowthRate = distributionGrowth(value.Current.CommissionCNY, value.Previous.CommissionCNY)
}

func (r *distributionRepository) distributionBusinessAnalytics(ctx context.Context, agentID int64, filter service.DistributionAnalyticsFilter) (*service.DistributionBusinessAnalytics, error) {
	scope := newDistributionAnalyticsScope(filter)
	days := int(scope.currentEnd.Sub(scope.currentStart).Hours() / 24)
	trendResolution, trendInterval := distributionTrendResolution(days)
	result := &service.DistributionBusinessAnalytics{Days: days, DateFrom: scope.currentStart.Format("2006-01-02"), DateTo: scope.currentEnd.AddDate(0, 0, -1).Format("2006-01-02"), TrendResolution: trendResolution}

	// Segment 0 is direct business, segment 1 is business produced by an L2 team.
	rows, err := r.db.QueryContext(ctx, `WITH periods AS (
		SELECT 0 AS period,$2::timestamptz AS starts_at,$3::timestamptz AS ends_at
		UNION ALL SELECT 1,$1::timestamptz,$2::timestamptz
	), scoped AS (
		SELECT binding.user_id,binding.bound_at,
			CASE WHEN $4=0 THEN CASE WHEN level.depth=2 THEN 1 ELSE 0 END
			ELSE CASE WHEN agent.parent_agent_id=$4 THEN 1 ELSE 0 END END AS segment
		FROM distribution_customer_bindings binding
		JOIN distribution_agents agent ON agent.id=binding.agent_id
		JOIN distribution_agent_levels level ON level.id=agent.level_id
		WHERE ($4=0 OR agent.id=$4 OR agent.parent_agent_id=$4)
	)
	SELECT period,segment,COUNT(*) FROM periods JOIN scoped ON scoped.bound_at>=starts_at AND scoped.bound_at<ends_at
	GROUP BY period,segment`, scope.previousStart, scope.currentStart, scope.currentEnd, agentID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var period, segment int
		var count int64
		if err = rows.Scan(&period, &segment, &count); err != nil {
			_ = rows.Close()
			return nil, err
		}
		target := distributionMetricsTarget(result, period, segment)
		target.NewCustomers = count
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}

	rows, err = r.db.QueryContext(ctx, `WITH periods AS (
		SELECT 0 AS period,$2::timestamptz AS starts_at,$3::timestamptz AS ends_at
		UNION ALL SELECT 1,$1::timestamptz,$2::timestamptz
	), scoped AS (
		SELECT source.id,source.customer_user_id,source.paid_at,binding.bound_at,GREATEST(COALESCE(source.commission_base_cny,0)-source.refunded_amount_cny,0) AS paid,
			CASE WHEN $4=0 THEN CASE WHEN level.depth=2 THEN 1 ELSE 0 END
			ELSE CASE WHEN direct.parent_agent_id=$4 THEN 1 ELSE 0 END END AS segment
		FROM distribution_commission_sources source
		JOIN distribution_agents direct ON direct.id=source.direct_agent_id
		JOIN distribution_agent_levels level ON level.id=direct.level_id
		JOIN distribution_customer_bindings binding ON binding.user_id=source.customer_user_id
		WHERE source.status<>'void' AND ($4=0 OR direct.id=$4 OR direct.parent_agent_id=$4)
	)
	SELECT period,segment,COUNT(*),COUNT(DISTINCT customer_user_id) FILTER (WHERE bound_at>=starts_at AND bound_at<ends_at),COALESCE(SUM(paid),0)
	FROM periods JOIN scoped ON scoped.paid_at>=starts_at AND scoped.paid_at<ends_at
	GROUP BY period,segment`, scope.previousStart, scope.currentStart, scope.currentEnd, agentID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var period, segment int
		target := service.DistributionBusinessMetrics{}
		if err = rows.Scan(&period, &segment, &target.PaidOrders, &target.PayingCustomers, &target.CustomerPaidCNY); err != nil {
			_ = rows.Close()
			return nil, err
		}
		metrics := distributionMetricsTarget(result, period, segment)
		metrics.PaidOrders, metrics.PayingCustomers, metrics.CustomerPaidCNY = target.PaidOrders, target.PayingCustomers, target.CustomerPaidCNY
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}

	rows, err = r.db.QueryContext(ctx, `WITH periods AS (
		SELECT 0 AS period,$2::timestamptz AS starts_at,$3::timestamptz AS ends_at
		UNION ALL SELECT 1,$1::timestamptz,$2::timestamptz
	)
	SELECT period,CASE WHEN $4=0 THEN CASE WHEN direct.parent_agent_id IS NOT NULL THEN 1 ELSE 0 END
		ELSE CASE WHEN direct.id=$4 THEN 0 ELSE 1 END END AS segment,
		COALESCE(SUM(GREATEST(entry.original_amount_cny-entry.reversed_amount_cny,0)),0)
	FROM periods JOIN distribution_commission_entries entry ON entry.created_at>=starts_at AND entry.created_at<ends_at
	JOIN distribution_commission_sources source ON source.id=entry.source_id
	JOIN distribution_agents direct ON direct.id=source.direct_agent_id
	WHERE ($4=0 OR entry.beneficiary_agent_id=$4)
	GROUP BY period,segment`, scope.previousStart, scope.currentStart, scope.currentEnd, agentID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var period, segment int
		var amount decimal.Decimal
		if err = rows.Scan(&period, &segment, &amount); err != nil {
			_ = rows.Close()
			return nil, err
		}
		distributionMetricsTarget(result, period, segment).CommissionCNY = amount
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}

	trendEnd := scope.currentEnd
	now := time.Now().In(scope.currentEnd.Location())
	if now.Before(trendEnd) {
		trendEnd = now
	}
	trendQueryEnd := trendEnd.Add(-time.Nanosecond)
	rows, err = r.db.QueryContext(ctx, `WITH buckets AS (
		SELECT generate_series(date_trunc($4,timezone('Asia/Shanghai',$1)),date_trunc($4,timezone('Asia/Shanghai',$2)), $5::interval) AS bucket),
	bindings AS (SELECT date_trunc($4,timezone('Asia/Shanghai',binding.bound_at)) AS bucket,CASE WHEN $3=0 THEN CASE WHEN level.depth=2 THEN 1 ELSE 0 END ELSE CASE WHEN agent.parent_agent_id=$3 THEN 1 ELSE 0 END END segment,COUNT(*) count
		FROM distribution_customer_bindings binding JOIN distribution_agents agent ON agent.id=binding.agent_id JOIN distribution_agent_levels level ON level.id=agent.level_id
		WHERE binding.bound_at>=$1 AND binding.bound_at<$2 AND ($3=0 OR agent.id=$3 OR agent.parent_agent_id=$3) GROUP BY 1,2),
	sources AS (SELECT date_trunc($4,timezone('Asia/Shanghai',source.paid_at)) AS bucket,CASE WHEN $3=0 THEN CASE WHEN level.depth=2 THEN 1 ELSE 0 END ELSE CASE WHEN direct.parent_agent_id=$3 THEN 1 ELSE 0 END END segment,
		COUNT(DISTINCT source.customer_user_id) customers,COALESCE(SUM(GREATEST(COALESCE(source.commission_base_cny,0)-source.refunded_amount_cny,0)),0) paid
		FROM distribution_commission_sources source JOIN distribution_agents direct ON direct.id=source.direct_agent_id JOIN distribution_agent_levels level ON level.id=direct.level_id
		WHERE source.status<>'void' AND source.paid_at>=$1 AND source.paid_at<$2 AND ($3=0 OR direct.id=$3 OR direct.parent_agent_id=$3) GROUP BY 1,2),
	commissions AS (SELECT date_trunc($4,timezone('Asia/Shanghai',entry.created_at)) AS bucket,CASE WHEN $3=0 THEN CASE WHEN direct.parent_agent_id IS NOT NULL THEN 1 ELSE 0 END ELSE CASE WHEN direct.id=$3 THEN 0 ELSE 1 END END segment,COALESCE(SUM(GREATEST(entry.original_amount_cny-entry.reversed_amount_cny,0)),0) commission
		FROM distribution_commission_entries entry JOIN distribution_commission_sources source ON source.id=entry.source_id JOIN distribution_agents direct ON direct.id=source.direct_agent_id
		WHERE entry.created_at>=$1 AND entry.created_at<$2 AND ($3=0 OR entry.beneficiary_agent_id=$3) GROUP BY 1,2)
	SELECT buckets.bucket,segment.value,COALESCE(bindings.count,0),COALESCE(sources.customers,0),COALESCE(sources.paid,0),COALESCE(commissions.commission,0)
	FROM buckets CROSS JOIN (VALUES(0),(1)) segment(value)
	LEFT JOIN bindings ON bindings.bucket=buckets.bucket AND bindings.segment=segment.value
	LEFT JOIN sources ON sources.bucket=buckets.bucket AND sources.segment=segment.value
	LEFT JOIN commissions ON commissions.bucket=buckets.bucket AND commissions.segment=segment.value ORDER BY buckets.bucket,segment.value`, scope.currentStart, trendQueryEnd, agentID, trendResolution, trendInterval)
	if err != nil {
		return nil, err
	}
	result.DailyDirect = make([]service.DistributionDailyMetric, 0)
	result.DailyTeam = make([]service.DistributionDailyMetric, 0)
	for rows.Next() {
		var bucket time.Time
		var segment int
		var item service.DistributionDailyMetric
		if err = rows.Scan(&bucket, &segment, &item.NewCustomers, &item.PayingCustomers, &item.CustomerPaidCNY, &item.CommissionCNY); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if trendResolution == "hour" {
			item.Date = bucket.Format("2006-01-02T15:04:05") + "+08:00"
		} else {
			item.Date = bucket.Format("2006-01-02")
		}
		if segment == 0 {
			result.DailyDirect = append(result.DailyDirect, item)
		} else {
			result.DailyTeam = append(result.DailyTeam, item)
		}
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}

	finalizeDistributionComparison(&result.Direct)
	finalizeDistributionComparison(&result.Team)
	result.Total.Current = addDistributionMetrics(result.Direct.Current, result.Team.Current)
	result.Total.Previous = addDistributionMetrics(result.Direct.Previous, result.Team.Previous)
	result.Total.PaidGrowthRate = distributionGrowth(result.Total.Current.CustomerPaidCNY, result.Total.Previous.CustomerPaidCNY)
	result.Total.CommissionGrowthRate = distributionGrowth(result.Total.Current.CommissionCNY, result.Total.Previous.CommissionCNY)
	if agentID == 0 {
		err = r.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT source.direct_agent_id) FROM distribution_commission_sources source WHERE source.status<>'void' AND source.paid_at>=$1 AND source.paid_at<$2`, scope.currentStart, scope.currentEnd).Scan(&result.ActiveAgents)
	} else {
		err = r.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT direct.id) FROM distribution_commission_sources source JOIN distribution_agents direct ON direct.id=source.direct_agent_id WHERE source.status<>'void' AND source.paid_at>=$2 AND source.paid_at<$3 AND direct.parent_agent_id=$1`, agentID, scope.currentStart, scope.currentEnd).Scan(&result.ActiveAgents)
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func distributionMetricsTarget(result *service.DistributionBusinessAnalytics, period, segment int) *service.DistributionBusinessMetrics {
	comparison := &result.Direct
	if segment == 1 {
		comparison = &result.Team
	}
	if period == 1 {
		return &comparison.Previous
	}
	return &comparison.Current
}

func (r *distributionRepository) distributionAgentRanking(ctx context.Context, parentAgentID int64, filter service.DistributionAnalyticsFilter, limit int) ([]service.DistributionAgentRanking, error) {
	scope := newDistributionAnalyticsScope(filter)
	rows, err := r.db.QueryContext(ctx, `SELECT direct.id,COALESCE(agent_user.email,''),COALESCE(agent_user.username,''),level.depth,
		COALESCE((SELECT SUM(GREATEST(COALESCE(source.commission_base_cny,0)-source.refunded_amount_cny,0)) FROM distribution_commission_sources source
			JOIN distribution_agents source_agent ON source_agent.id=source.direct_agent_id
			WHERE (source_agent.id=direct.id OR ($1=0 AND source_agent.parent_agent_id=direct.id)) AND source.status<>'void' AND source.paid_at>=$2 AND source.paid_at<$3),0) paid,
		COALESCE((SELECT SUM(GREATEST(entry.original_amount_cny-entry.reversed_amount_cny,0)) FROM distribution_commission_entries entry
			JOIN distribution_commission_sources entry_source ON entry_source.id=entry.source_id
			JOIN distribution_agents entry_direct ON entry_direct.id=entry_source.direct_agent_id
			WHERE (($1=0 AND (entry_direct.id=direct.id OR entry_direct.parent_agent_id=direct.id))
				OR ($1<>0 AND entry.beneficiary_agent_id=$1 AND entry_direct.id=direct.id))
			AND entry.created_at>=$2 AND entry.created_at<$3),0) commission,
		(SELECT COUNT(*) FROM distribution_customer_bindings binding JOIN distribution_agents binding_agent ON binding_agent.id=binding.agent_id
			WHERE (binding_agent.id=direct.id OR ($1=0 AND binding_agent.parent_agent_id=direct.id)) AND binding.bound_at>=$2 AND binding.bound_at<$3) new_customers,
		(SELECT COUNT(DISTINCT source.customer_user_id) FROM distribution_commission_sources source JOIN distribution_agents source_agent ON source_agent.id=source.direct_agent_id
			JOIN distribution_customer_bindings binding ON binding.user_id=source.customer_user_id
			WHERE (source_agent.id=direct.id OR ($1=0 AND source_agent.parent_agent_id=direct.id)) AND source.status<>'void' AND source.paid_at>=$2 AND source.paid_at<$3 AND binding.bound_at>=$2 AND binding.bound_at<$3) paying_customers
	FROM distribution_agents direct
	JOIN distribution_agent_levels level ON level.id=direct.level_id
	JOIN users agent_user ON agent_user.id=direct.user_id
	WHERE (($1=0 AND direct.parent_agent_id IS NULL) OR direct.parent_agent_id=$1) AND direct.status<>'revoked'
	ORDER BY paid DESC,direct.id LIMIT $4`, parentAgentID, scope.currentStart, scope.currentEnd, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionAgentRanking, 0)
	for rows.Next() {
		var item service.DistributionAgentRanking
		if err = rows.Scan(&item.AgentID, &item.Email, &item.Username, &item.Depth, &item.CustomerPaidCNY, &item.CommissionCNY, &item.NewCustomers, &item.PayingCustomers); err != nil {
			return nil, err
		}
		item.ConversionRate = distributionMetricRate(decimal.NewFromInt(item.PayingCustomers).Mul(decimal.NewFromInt(100)), item.NewCustomers)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *distributionRepository) AdminGetAgentAnalytics(ctx context.Context, agentID int64, filter service.DistributionAnalyticsFilter) (*service.DistributionAgentAnalytics, error) {
	var userID int64
	var email, username string
	err := r.db.QueryRowContext(ctx, `SELECT agent.user_id,COALESCE(users.email,''),COALESCE(users.username,'') FROM distribution_agents agent JOIN users ON users.id=agent.user_id WHERE agent.id=$1`, agentID).Scan(&userID, &email, &username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrDistributionNotAgent
	}
	if err != nil {
		return nil, err
	}
	agent, err := r.GetAgentByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	agent.Email, agent.Username = email, username
	analytics, err := r.distributionBusinessAnalytics(ctx, agentID, filter)
	if err != nil {
		return nil, err
	}
	ranking := make([]service.DistributionAgentRanking, 0)
	if agent.Depth == 1 {
		ranking, err = r.distributionAgentRanking(ctx, agentID, filter, 8)
		if err != nil {
			return nil, err
		}
	}
	return &service.DistributionAgentAnalytics{Agent: agent, Analytics: analytics, Ranking: ranking}, nil
}

func (r *distributionRepository) AdminUpdateSettings(ctx context.Context, s service.DistributionSettings, adminID int64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `UPDATE distribution_settings SET enabled=$1,freeze_hours=$2,withdrawal_enabled=$3,
		minimum_withdrawal_cny=$4,maximum_withdrawal_cny=$5,withdrawal_fee_rate_bps=$6,withdrawal_fee_fixed_cny=$7,
		daily_withdrawal_limit_cny=$8,monthly_withdrawal_limit_cny=$9,cny_per_platform_usd=$10,withdrawal_dual_approval_enabled=$11,
		promotion_tracking_enabled=$12,promotion_attribution_enabled=$13,promotion_attribution_days=$14,promotion_attribution_model=$15,
		promotion_collect_source=$16,promotion_collect_device=$17,promotion_bot_filter_enabled=$18,promotion_detail_retention_days=$19,
		registration_reward_enabled=$20,recharge_reward_enabled=$21,
		updated_by=$22,updated_at=NOW() WHERE id=1`,
		s.Enabled, s.FreezeHours, s.WithdrawalEnabled, s.MinimumWithdrawalCNY, s.MaximumWithdrawalCNY, s.WithdrawalFeeRateBPS, s.WithdrawalFeeFixedCNY,
		s.DailyWithdrawalLimitCNY, s.MonthlyWithdrawalLimitCNY, s.CNYPerPlatformUSD, s.WithdrawalDualApproval,
		s.PromotionTrackingEnabled, s.PromotionAttributionEnabled, s.PromotionAttributionDays, s.PromotionAttributionModel,
		s.PromotionCollectSource, s.PromotionCollectDevice, s.PromotionBotFilterEnabled, s.PromotionDetailRetentionDays,
		s.RegistrationRewardEnabled, s.RechargeRewardEnabled, adminID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE distribution_agent_levels SET default_rate_bps=CASE depth WHEN 1 THEN $1::integer ELSE $3::integer END,max_child_rate_bps=CASE depth WHEN 1 THEN $2::integer ELSE 0 END,updated_at=NOW() WHERE depth IN (1,2)`, s.L1DefaultRateBPS, s.L1MaxChildRateBPS, s.L2DefaultRateBPS); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_fx_rates(currency,rate_to_cny,updated_by) VALUES('USD',$1,$2) ON CONFLICT(currency) DO UPDATE SET rate_to_cny=EXCLUDED.rate_to_cny,updated_by=EXCLUDED.updated_by,updated_at=NOW()`, s.USDToCNY, adminID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *distributionRepository) AdminSetFXRate(ctx context.Context, currency string, rate decimal.Decimal, adminID int64) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO distribution_fx_rates(currency,rate_to_cny,updated_by) VALUES($1,$2,$3)
		ON CONFLICT(currency) DO UPDATE SET rate_to_cny=EXCLUDED.rate_to_cny,updated_by=EXCLUDED.updated_by,updated_at=NOW()`, currency, rate, adminID)
	if err != nil {
		return err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT payment_order_id,customer_user_id,payment_type,payment_currency,actual_paid_amount,provider_snapshot,paid_at FROM distribution_commission_sources WHERE status='pending_fx' AND payment_currency=$1 ORDER BY id`, currency)
	if err != nil {
		return err
	}
	var inputs []service.DistributionCommissionInput
	for rows.Next() {
		var in service.DistributionCommissionInput
		if err = rows.Scan(&in.PaymentOrderID, &in.CustomerUserID, &in.PaymentType, &in.PaymentCurrency, &in.ActualPaid, &in.ProviderSnapshot, &in.PaidAt); err != nil {
			_ = rows.Close()
			return err
		}
		inputs = append(inputs, in)
	}
	if err = rows.Close(); err != nil {
		return err
	}
	for _, in := range inputs {
		if _, err = r.AccruePaidOrder(ctx, in); err != nil {
			return fmt.Errorf("settle pending FX order %d: %w", in.PaymentOrderID, err)
		}
	}
	return nil
}

func (r *distributionRepository) GetAgentRewardRule(ctx context.Context, actorUserID, agentID int64, admin bool) (*service.DistributionRewardRule, error) {
	var allowed bool
	if admin {
		allowed = true
	} else if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM distribution_agents child JOIN distribution_agents parent ON parent.id=child.parent_agent_id WHERE child.id=$2 AND parent.user_id=$1 AND parent.status='active')`, actorUserID, agentID).Scan(&allowed); err != nil {
		return nil, err
	}
	if !allowed {
		return nil, service.ErrDistributionNotAgent
	}
	rule := &service.DistributionRewardRule{AgentID: agentID}
	err := r.db.QueryRowContext(ctx, `SELECT registration_enabled,registration_reward_cny,recharge_enabled,recharge_threshold_cny,recharge_reward_cny FROM distribution_agent_reward_rules WHERE agent_id=$1`, agentID).
		Scan(&rule.RegistrationEnabled, &rule.RegistrationRewardCNY, &rule.RechargeEnabled, &rule.RechargeThresholdCNY, &rule.RechargeRewardCNY)
	if errors.Is(err, sql.ErrNoRows) {
		return rule, nil
	}
	return rule, err
}

func (r *distributionRepository) UpdateAgentRewardRule(ctx context.Context, actorUserID, agentID int64, admin bool, in service.DistributionRewardRuleInput) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var depth int
	var parentID sql.NullInt64
	if err = tx.QueryRowContext(ctx, `SELECT l.depth,a.parent_agent_id FROM distribution_agents a JOIN distribution_agent_levels l ON l.id=a.level_id WHERE a.id=$1 AND a.status<>'revoked' FOR UPDATE OF a`, agentID).Scan(&depth, &parentID); err != nil {
		return service.ErrDistributionNotAgent
	}
	if admin {
		if depth != 1 {
			return infraBadRequest("REWARD_RULE_L1_REQUIRED", "administrator can configure hierarchy rewards only for L1 agents")
		}
	} else {
		if depth != 2 || !parentID.Valid {
			return service.ErrDistributionNotAgent
		}
		var parentUserID int64
		var rootRegistration, rootRecharge decimal.Decimal
		if err = tx.QueryRowContext(ctx, `SELECT parent.user_id,COALESCE(rule.registration_reward_cny,0),COALESCE(rule.recharge_reward_cny,0)
			FROM distribution_agents parent LEFT JOIN distribution_agent_reward_rules rule ON rule.agent_id=parent.id WHERE parent.id=$1 AND parent.status='active' FOR UPDATE OF parent`, parentID.Int64).
			Scan(&parentUserID, &rootRegistration, &rootRecharge); err != nil || parentUserID != actorUserID {
			return service.ErrDistributionNotAgent
		}
		if in.RegistrationRewardCNY.GreaterThan(rootRegistration) || in.RechargeRewardCNY.GreaterThan(rootRecharge) {
			return infraBadRequest("REWARD_SHARE_EXCEEDS_BUDGET", "L2 reward share exceeds L1 hierarchy budget")
		}
		// The recharge threshold belongs to the hierarchy and is inherited by L2.
		in.RechargeThresholdCNY = decimal.Zero
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO distribution_agent_reward_rules
		(agent_id,registration_enabled,registration_reward_cny,recharge_enabled,recharge_threshold_cny,recharge_reward_cny,updated_by)
		VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(agent_id) DO UPDATE SET registration_enabled=EXCLUDED.registration_enabled,registration_reward_cny=EXCLUDED.registration_reward_cny,
		recharge_enabled=EXCLUDED.recharge_enabled,
		recharge_threshold_cny=EXCLUDED.recharge_threshold_cny,recharge_reward_cny=EXCLUDED.recharge_reward_cny,updated_by=EXCLUDED.updated_by,updated_at=NOW()`,
		agentID, in.RegistrationEnabled, in.RegistrationRewardCNY, in.RechargeEnabled, in.RechargeThresholdCNY, in.RechargeRewardCNY, actorUserID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *distributionRepository) AdminGrantAgent(ctx context.Context, in service.DistributionGrantAgentInput) (*service.DistributionAgent, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var eligible bool
	var oldAgentID sql.NullInt64
	var oldCode sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND deleted_at IS NULL)
		AND NOT EXISTS(SELECT 1 FROM distribution_agents WHERE user_id=$1)
		AND NOT EXISTS(SELECT 1 FROM user_affiliates WHERE user_id=$1 AND inviter_id IS NOT NULL)`, in.UserID).Scan(&eligible)
	if err != nil {
		return nil, err
	}
	if !eligible {
		return nil, infraBadRequest("AGENT_USER_INELIGIBLE", "user is inactive, already an agent, or has an invitation binding")
	}
	err = tx.QueryRowContext(ctx, `SELECT agent_id,promotion_code FROM distribution_customer_bindings WHERE user_id=$1 FOR UPDATE`, in.UserID).Scan(&oldAgentID, &oldCode)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if oldAgentID.Valid && !in.UpgradeCustomer {
		return nil, infraBadRequest("AGENT_USER_INELIGIBLE", "user has a distribution customer binding; upgrade is required")
	}
	if in.UpgradeCustomer && !oldAgentID.Valid {
		return nil, infraBadRequest("CUSTOMER_BINDING_REQUIRED", "upgrade requires an existing distribution customer binding")
	}
	if in.Depth == 2 {
		var parentDepth int
		var parentStatus string
		if err = tx.QueryRowContext(ctx, `SELECT l.depth,a.status FROM distribution_agents a JOIN distribution_agent_levels l ON l.id=a.level_id WHERE a.id=$1 FOR UPDATE`, *in.ParentAgentID).Scan(&parentDepth, &parentStatus); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, infraBadRequest("PARENT_AGENT_NOT_FOUND", "parent agent not found")
			}
			return nil, err
		}
		if parentDepth != 1 || parentStatus != "active" {
			return nil, infraBadRequest("PARENT_AGENT_INVALID", "parent agent must be an active L1 agent")
		}
	}
	var levelID int64
	var defaultRate, maxChild int
	if err = tx.QueryRowContext(ctx, `SELECT id,default_rate_bps,max_child_rate_bps FROM distribution_agent_levels WHERE depth=$1 AND is_active=TRUE`, in.Depth).Scan(&levelID, &defaultRate, &maxChild); err != nil {
		return nil, err
	}
	if in.Depth == 1 {
		in.ParentAgentID = nil
	} else {
		if in.ParentAgentID == nil {
			return nil, infraBadRequest("PARENT_AGENT_REQUIRED", "L2 agent requires an L1 parent")
		}
		var parentRate, parentMax, depth int
		var status string
		err = tx.QueryRowContext(ctx, `SELECT COALESCE(a.rate_override_bps,l.default_rate_bps),l.max_child_rate_bps,l.depth,a.status FROM distribution_agents a JOIN distribution_agent_levels l ON l.id=a.level_id WHERE a.id=$1 FOR UPDATE`, *in.ParentAgentID).Scan(&parentRate, &parentMax, &depth, &status)
		if err != nil || depth != 1 || status != "active" {
			return nil, infraBadRequest("PARENT_AGENT_INVALID", "parent agent is not active L1")
		}
		rate := defaultRate
		if in.RateOverrideBPS != nil {
			rate = *in.RateOverrideBPS
		}
		if rate < 0 || rate > parentRate || rate > parentMax {
			return nil, infraBadRequest("AGENT_RATE_INVALID", "L2 rate exceeds parent limit")
		}
	}
	code := in.PromotionCode
	for attempt := 0; code == "" && attempt < 12; attempt++ {
		candidate, genErr := randomDistributionCode()
		if genErr != nil {
			return nil, genErr
		}
		var exists bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM distribution_agents WHERE promotion_code=$1)`, candidate).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			code = candidate
		}
	}
	if code == "" {
		return nil, fmt.Errorf("generate distribution promotion code")
	}
	var agentID int64
	err = tx.QueryRowContext(ctx, `INSERT INTO distribution_agents(user_id,level_id,parent_agent_id,promotion_code,rate_override_bps,granted_by) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, in.UserID, levelID, in.ParentAgentID, code, in.RateOverrideBPS, in.GrantedBy).Scan(&agentID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, infraBadRequest("AGENT_ALREADY_EXISTS", "agent or promotion code already exists")
		}
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_wallets(agent_id) VALUES($1)`, agentID); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_agent_events
		(agent_id,event_type,new_status,new_rate_override_bps,new_effective_rate_bps,reason,actor_user_id)
		VALUES($1,'created','active',$2,$3,'agent granted',$4)`, agentID, in.RateOverrideBPS, effectiveDistributionRate(in.RateOverrideBPS, defaultRate), in.GrantedBy); err != nil {
		return nil, err
	}
	if oldAgentID.Valid {
		if _, err = tx.ExecContext(ctx, `DELETE FROM distribution_customer_bindings WHERE user_id=$1`, in.UserID); err != nil {
			return nil, err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_binding_events
			(customer_user_id,old_agent_id,new_agent_id,old_promotion_code,new_promotion_code,reason,actor_user_id)
			VALUES($1,$2,$3,$4,$5,$6,$7)`, in.UserID, oldAgentID.Int64, agentID, oldCode.String, code, "customer upgraded to distribution agent", in.GrantedBy); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetAgentByUserID(ctx, in.UserID)
}

func effectiveDistributionRate(override *int, defaultRate int) int {
	if override != nil {
		return *override
	}
	return defaultRate
}

func (r *distributionRepository) AdminUpdateAgentStatus(ctx context.Context, agentID, adminID int64, status, reason string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var current string
	if err = tx.QueryRowContext(ctx, `SELECT status FROM distribution_agents WHERE id=$1 FOR UPDATE`, agentID).Scan(&current); errors.Is(err, sql.ErrNoRows) {
		return service.ErrDistributionNotAgent
	} else if err != nil {
		return err
	}
	if current == status {
		return infraBadRequest("AGENT_STATUS_UNCHANGED", "agent already has requested status")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE distribution_agents SET status=$2,status_reason=$3,updated_at=NOW() WHERE id=$1`, agentID, status, reason); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_agent_events(agent_id,event_type,old_status,new_status,reason,actor_user_id)
		VALUES($1,'status_changed',$2,$3,$4,$5)`, agentID, current, status, reason, adminID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *distributionRepository) AdminUpdateAgentRate(ctx context.Context, agentID, adminID int64, rateOverrideBPS *int, reason string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var oldOverride sql.NullInt64
	var oldEffective, levelDefault, depth, levelMax int
	var parentID sql.NullInt64
	err = tx.QueryRowContext(ctx, `SELECT a.rate_override_bps,COALESCE(a.rate_override_bps,l.default_rate_bps),l.default_rate_bps,l.depth,l.max_child_rate_bps,a.parent_agent_id
		FROM distribution_agents a JOIN distribution_agent_levels l ON l.id=a.level_id WHERE a.id=$1 FOR UPDATE OF a`, agentID).
		Scan(&oldOverride, &oldEffective, &levelDefault, &depth, &levelMax, &parentID)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrDistributionNotAgent
	}
	if err != nil {
		return err
	}
	newEffective := effectiveDistributionRate(rateOverrideBPS, levelDefault)
	if depth == 1 {
		var highestChild int
		if err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(COALESCE(c.rate_override_bps,cl.default_rate_bps)),0)
			FROM distribution_agents c JOIN distribution_agent_levels cl ON cl.id=c.level_id
			WHERE c.parent_agent_id=$1 AND c.status<>'revoked'`, agentID).Scan(&highestChild); err != nil {
			return err
		}
		if newEffective < highestChild {
			return infraBadRequest("AGENT_RATE_CONFLICT", "agent rate is below an existing child rate")
		}
	} else {
		if !parentID.Valid {
			return infraBadRequest("PARENT_AGENT_REQUIRED", "L2 agent requires an L1 parent")
		}
		var parentRate, parentMax int
		var parentStatus string
		if err = tx.QueryRowContext(ctx, `SELECT COALESCE(p.rate_override_bps,pl.default_rate_bps),pl.max_child_rate_bps,p.status
			FROM distribution_agents p JOIN distribution_agent_levels pl ON pl.id=p.level_id WHERE p.id=$1 FOR UPDATE OF p`, parentID.Int64).
			Scan(&parentRate, &parentMax, &parentStatus); err != nil {
			return err
		}
		if parentStatus == "revoked" || newEffective > parentRate || newEffective > parentMax {
			return infraBadRequest("AGENT_RATE_CONFLICT", "child rate exceeds parent policy")
		}
	}
	if (!oldOverride.Valid && rateOverrideBPS == nil) || (oldOverride.Valid && rateOverrideBPS != nil && int(oldOverride.Int64) == *rateOverrideBPS) {
		return infraBadRequest("AGENT_RATE_UNCHANGED", "agent already has requested rate")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE distribution_agents SET rate_override_bps=$2,updated_at=NOW() WHERE id=$1`, agentID, rateOverrideBPS); err != nil {
		return err
	}
	var oldOverrideValue any
	if oldOverride.Valid {
		oldOverrideValue = oldOverride.Int64
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_agent_events
		(agent_id,event_type,old_rate_override_bps,new_rate_override_bps,old_effective_rate_bps,new_effective_rate_bps,reason,actor_user_id)
		VALUES($1,'rate_changed',$2,$3,$4,$5,$6,$7)`, agentID, oldOverrideValue, rateOverrideBPS, oldEffective, newEffective, reason, adminID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *distributionRepository) AdminUpdateAgentRecruitmentPermission(ctx context.Context, agentID, adminID int64, enabled bool, reason string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var current bool
	var depth int
	err = tx.QueryRowContext(ctx, `SELECT a.can_recruit_subagents,l.depth FROM distribution_agents a JOIN distribution_agent_levels l ON l.id=a.level_id WHERE a.id=$1 FOR UPDATE OF a`, agentID).Scan(&current, &depth)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrDistributionNotAgent
	}
	if err != nil {
		return err
	}
	if depth != 1 {
		return infraBadRequest("L1_AGENT_REQUIRED", "only a top-level agent can recruit subagents")
	}
	if current == enabled {
		return infraBadRequest("RECRUITMENT_PERMISSION_UNCHANGED", "agent already has requested recruitment permission")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE distribution_agents SET can_recruit_subagents=$2,updated_at=NOW() WHERE id=$1`, agentID, enabled); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_agent_events(agent_id,event_type,old_status,new_status,reason,actor_user_id) VALUES($1,'permission_changed',$2,$3,$4,$5)`, agentID, fmt.Sprintf("%t", current), fmt.Sprintf("%t", enabled), reason, adminID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *distributionRepository) AdminUpdateAgentPromotionStatsPermission(ctx context.Context, agentID, adminID int64, enabled bool, reason string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var current bool
	err = tx.QueryRowContext(ctx, `SELECT can_view_promotion_stats FROM distribution_agents WHERE id=$1 FOR UPDATE`, agentID).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrDistributionNotAgent
	}
	if err != nil {
		return err
	}
	if current == enabled {
		return infraBadRequest("PROMOTION_STATS_PERMISSION_UNCHANGED", "agent already has requested promotion statistics permission")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE distribution_agents SET can_view_promotion_stats=$2,updated_at=NOW() WHERE id=$1`, agentID, enabled); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_agent_events(agent_id,event_type,old_status,new_status,reason,actor_user_id)
		VALUES($1,'permission_changed',$2,$3,$4,$5)`, agentID, fmt.Sprintf("promotion_stats:%t", current), fmt.Sprintf("promotion_stats:%t", enabled), reason, adminID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *distributionRepository) AdminUpdateLevel(ctx context.Context, depth, defaultRateBPS, maxChildRateBPS int, active bool) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var invalid bool
	if depth == 1 {
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM distribution_agents p JOIN distribution_agent_levels pl ON pl.id=p.level_id AND pl.depth=1 JOIN distribution_agents c ON c.parent_agent_id=p.id JOIN distribution_agent_levels cl ON cl.id=c.level_id WHERE COALESCE(c.rate_override_bps,cl.default_rate_bps)>$1 OR (p.rate_override_bps IS NULL AND COALESCE(c.rate_override_bps,cl.default_rate_bps)>$2))`, maxChildRateBPS, defaultRateBPS).Scan(&invalid)
	} else {
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM distribution_agents c JOIN distribution_agent_levels cl ON cl.id=c.level_id AND cl.depth=2 JOIN distribution_agents p ON p.id=c.parent_agent_id JOIN distribution_agent_levels pl ON pl.id=p.level_id WHERE c.rate_override_bps IS NULL AND ($1>COALESCE(p.rate_override_bps,pl.default_rate_bps) OR $1>pl.max_child_rate_bps))`, defaultRateBPS).Scan(&invalid)
	}
	if err != nil {
		return err
	}
	if invalid {
		return infraBadRequest("AGENT_LEVEL_CONFLICT", "level change conflicts with existing agent rates")
	}
	_, err = tx.ExecContext(ctx, `UPDATE distribution_agent_levels SET default_rate_bps=$2,max_child_rate_bps=$3,is_active=$4,updated_at=NOW() WHERE depth=$1`, depth, defaultRateBPS, maxChildRateBPS, active)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (r *distributionRepository) AdminCorrectCustomerBinding(ctx context.Context, userID, agentID, adminID int64, reason string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var code string
	err = tx.QueryRowContext(ctx, `SELECT promotion_code FROM distribution_agents WHERE id=$1 AND status='active'`, agentID).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrDistributionNotAgent
	}
	if err != nil {
		return err
	}
	var conflict bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_affiliates WHERE user_id=$1 AND inviter_id IS NOT NULL) OR EXISTS(SELECT 1 FROM distribution_agents WHERE user_id=$1)`, userID).Scan(&conflict); err != nil {
		return err
	}
	if conflict {
		return infraBadRequest("PROMOTION_BINDING_CONFLICT", "user cannot be bound to distribution")
	}
	var oldAgentID sql.NullInt64
	var oldCode sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT agent_id,promotion_code FROM distribution_customer_bindings WHERE user_id=$1 FOR UPDATE`, userID).Scan(&oldAgentID, &oldCode)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if oldAgentID.Valid && oldAgentID.Int64 == agentID {
		return infraBadRequest("BINDING_UNCHANGED", "customer already belongs to target agent")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO distribution_customer_bindings(user_id,agent_id,promotion_code,corrected_by,correction_reason) VALUES($1,$2,$3,$4,$5) ON CONFLICT(user_id) DO UPDATE SET agent_id=EXCLUDED.agent_id,promotion_code=EXCLUDED.promotion_code,corrected_by=EXCLUDED.corrected_by,correction_reason=EXCLUDED.correction_reason,updated_at=NOW()`, userID, agentID, code, adminID, reason)
	if err != nil {
		return err
	}
	var oldAgentValue, oldCodeValue any
	if oldAgentID.Valid {
		oldAgentValue = oldAgentID.Int64
	}
	if oldCode.Valid {
		oldCodeValue = oldCode.String
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_binding_events
		(customer_user_id,old_agent_id,new_agent_id,old_promotion_code,new_promotion_code,reason,actor_user_id)
		VALUES($1,$2,$3,$4,$5,$6,$7)`, userID, oldAgentValue, agentID, oldCodeValue, code, reason, adminID); err != nil {
		return err
	}
	return tx.Commit()
}
func (r *distributionRepository) AdminListAgents(ctx context.Context, filter service.DistributionAdminListFilter) ([]service.DistributionAgent, int64, error) {
	if err := r.releaseAllMatured(ctx); err != nil {
		return nil, 0, err
	}
	like := "%" + filter.Search + "%"
	const from = `
		FROM distribution_agents a
		JOIN distribution_agent_levels l ON l.id = a.level_id
		JOIN users u ON u.id = a.user_id
		LEFT JOIN distribution_wallets w ON w.agent_id = a.id
		LEFT JOIN distribution_agents parent ON parent.id = a.parent_agent_id
		LEFT JOIN users parent_user ON parent_user.id = parent.user_id
		WHERE ($1 = '' OR u.email ILIKE $2 OR u.username ILIKE $2 OR a.promotion_code ILIKE $2
			OR parent_user.email ILIKE $2 OR parent_user.username ILIKE $2)
		  AND ($3 = '' OR a.status = $3)
		  AND ($4 = 0 OR l.depth = $4)`
	const analyticsCTE = `WITH customer_scope AS (
		SELECT binding.user_id,binding.bound_at,owner.agent_id
		FROM distribution_customer_bindings binding
		JOIN distribution_agents direct ON direct.id=binding.agent_id
		CROSS JOIN LATERAL (VALUES (direct.id),(direct.parent_agent_id)) owner(agent_id)
		WHERE owner.agent_id IS NOT NULL
	), customer_metrics AS (
		SELECT scope.agent_id,
			COUNT(DISTINCT scope.user_id) FILTER(WHERE scope.bound_at>=$5 AND scope.bound_at<$6 AND fact.first_activated_at>=scope.bound_at AND fact.first_activated_at<$6) period_activated_customers,
			COUNT(DISTINCT scope.user_id) FILTER(WHERE scope.bound_at>=$5 AND scope.bound_at<$6 AND fact.first_paid_at>=scope.bound_at AND fact.first_paid_at<$6) period_cohort_paid_customers,
			COUNT(DISTINCT scope.user_id) total_customers
		FROM customer_scope scope LEFT JOIN business_user_facts fact ON fact.user_id=scope.user_id GROUP BY scope.agent_id
	), payment_metrics AS (
		SELECT scope.agent_id,COUNT(DISTINCT po.user_id) period_all_paying_customers,
			COUNT(DISTINCT po.user_id) FILTER(WHERE EXISTS(SELECT 1 FROM payment_orders old_po JOIN business_payment_facts old_fact ON old_fact.payment_order_id=old_po.id WHERE old_po.user_id=po.user_id AND old_fact.paid_at<$5 AND old_fact.gross_amount>old_fact.refunded_amount)) period_repurchase_customers
		FROM customer_scope scope JOIN payment_orders po ON po.user_id=scope.user_id
		JOIN business_payment_facts fact ON fact.payment_order_id=po.id
		WHERE fact.paid_at>=$5 AND fact.paid_at<$6 AND fact.gross_amount>fact.refunded_amount GROUP BY scope.agent_id
	), activity_metrics AS (
		SELECT scope.agent_id,COUNT(DISTINCT usage.user_id) period_active_customers
		FROM customer_scope scope JOIN business_usage_hourly_facts usage ON usage.user_id=scope.user_id
		WHERE usage.first_used_at<$6 AND usage.last_used_at>=$5 GROUP BY scope.agent_id
	), reward_metrics AS (
		SELECT owner.agent_id,COALESCE(SUM(GREATEST(reward.original_amount_cny-reward.reversed_amount_cny,0)),0) period_reward_cny
		FROM distribution_reward_grants reward JOIN distribution_agents beneficiary ON beneficiary.id=reward.beneficiary_agent_id
		CROSS JOIN LATERAL (VALUES (beneficiary.id),(beneficiary.parent_agent_id)) owner(agent_id)
		WHERE owner.agent_id IS NOT NULL AND reward.created_at>=$5 AND reward.created_at<$6 GROUP BY owner.agent_id
	) `
	const analyticsJoins = `
		LEFT JOIN customer_metrics business_customers ON business_customers.agent_id=a.id
		LEFT JOIN payment_metrics business_payments ON business_payments.agent_id=a.id
		LEFT JOIN activity_metrics business_activity ON business_activity.agent_id=a.id
		LEFT JOIN reward_metrics business_rewards ON business_rewards.agent_id=a.id`

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+from, filter.Search, like, filter.Status, filter.Depth).Scan(&total); err != nil {
		return nil, 0, err
	}

	const selectColumns = `SELECT a.id,a.user_id,a.level_id,l.depth,a.parent_agent_id,a.promotion_code,a.rate_override_bps,
		COALESCE(a.rate_override_bps,l.default_rate_bps),l.max_child_rate_bps,a.can_recruit_subagents,a.can_view_promotion_stats,a.status,
		COALESCE(w.available_cny,0),COALESCE(w.frozen_cny,0),COALESCE(w.reserved_cny,0),
		COALESCE(w.debt_cny,0),COALESCE(w.total_earned_cny,0),COALESCE(w.total_withdrawn_cny,0),
		COALESCE(u.email,''),COALESCE(u.username,''),u.status,
		COALESCE(parent.user_id,0),COALESCE(parent_user.email,''),COALESCE(parent_user.username,''),
		(SELECT COUNT(*) FROM distribution_customer_bindings binding WHERE binding.agent_id=a.id) AS customer_count,
		(SELECT COUNT(*) FROM distribution_agents child WHERE child.parent_agent_id=a.id AND child.status<>'revoked') AS team_count,
		(SELECT COUNT(DISTINCT s.customer_user_id) FROM distribution_commission_sources s
			JOIN distribution_agents direct ON direct.id=s.direct_agent_id
			WHERE (direct.id=a.id OR direct.parent_agent_id=a.id) AND s.status<>'void' AND COALESCE(s.commission_base_cny,0)>0) AS paying_customer_count,
		COALESCE((SELECT SUM(GREATEST(COALESCE(s.commission_base_cny,0)-s.refunded_amount_cny,0)) FROM distribution_commission_sources s
			JOIN distribution_agents direct ON direct.id=s.direct_agent_id
			WHERE (direct.id=a.id OR direct.parent_agent_id=a.id) AND s.status<>'void'),0) AS customer_paid_cny,
		COALESCE((SELECT SUM(GREATEST(e.original_amount_cny-e.reversed_amount_cny,0)) FROM distribution_commission_entries e
			WHERE e.beneficiary_agent_id=a.id AND e.created_at>=DATE_TRUNC('month',NOW())),0) AS this_month_commission_cny,
		(SELECT COUNT(*) FROM distribution_customer_bindings binding WHERE binding.agent_id=a.id AND binding.bound_at>=$5 AND binding.bound_at<$6) AS period_customer_count,
		(SELECT COUNT(DISTINCT source.customer_user_id) FROM distribution_commission_sources source JOIN distribution_customer_bindings binding ON binding.user_id=source.customer_user_id
			WHERE source.direct_agent_id=a.id AND source.status<>'void' AND source.paid_at>=$5 AND source.paid_at<$6 AND binding.bound_at>=$5 AND binding.bound_at<$6) AS period_paying_customers,
		COALESCE((SELECT SUM(GREATEST(COALESCE(source.commission_base_cny,0)-source.refunded_amount_cny,0)) FROM distribution_commission_sources source
			WHERE source.direct_agent_id=a.id AND source.status<>'void' AND source.paid_at>=$5 AND source.paid_at<$6),0) AS period_customer_paid_cny,
		COALESCE((SELECT SUM(GREATEST(entry.original_amount_cny-entry.reversed_amount_cny,0)) FROM distribution_commission_entries entry
			WHERE entry.beneficiary_agent_id=a.id AND entry.entry_type='direct' AND entry.created_at>=$5 AND entry.created_at<$6),0) AS period_commission_cny,
		(SELECT COUNT(*) FROM distribution_customer_bindings binding JOIN distribution_agents child ON child.id=binding.agent_id WHERE child.parent_agent_id=a.id AND binding.bound_at>=$5 AND binding.bound_at<$6) AS team_customer_count,
		(SELECT COUNT(DISTINCT source.customer_user_id) FROM distribution_commission_sources source JOIN distribution_agents child ON child.id=source.direct_agent_id JOIN distribution_customer_bindings binding ON binding.user_id=source.customer_user_id
			WHERE child.parent_agent_id=a.id AND source.status<>'void' AND source.paid_at>=$5 AND source.paid_at<$6 AND binding.bound_at>=$5 AND binding.bound_at<$6) AS team_paying_customers,
		COALESCE((SELECT SUM(GREATEST(COALESCE(source.commission_base_cny,0)-source.refunded_amount_cny,0)) FROM distribution_commission_sources source JOIN distribution_agents child ON child.id=source.direct_agent_id
			WHERE child.parent_agent_id=a.id AND source.status<>'void' AND source.paid_at>=$5 AND source.paid_at<$6),0) AS team_customer_paid_cny,
		COALESCE((SELECT SUM(GREATEST(entry.original_amount_cny-entry.reversed_amount_cny,0)) FROM distribution_commission_entries entry
			WHERE entry.beneficiary_agent_id=a.id AND entry.entry_type='team' AND entry.created_at>=$5 AND entry.created_at<$6),0) AS team_commission_cny,
		COALESCE(business_customers.period_activated_customers,0) AS period_activated_customers,
		COALESCE(business_customers.period_cohort_paid_customers,0) AS period_cohort_paid_customers,
		COALESCE(business_payments.period_all_paying_customers,0) AS period_all_paying_customers,
		COALESCE(business_payments.period_repurchase_customers,0) AS period_repurchase_customers,
		COALESCE(business_activity.period_active_customers,0) AS period_active_customers,
		(SELECT COUNT(*) FROM distribution_customer_bindings binding JOIN distribution_agents direct ON direct.id=binding.agent_id WHERE direct.parent_agent_id=a.id) AS total_team_customers,
		COALESCE(business_rewards.period_reward_cny,0) AS period_reward_cny,
		COALESCE(w.total_converted_cny,0),
		(SELECT MAX(e.created_at) FROM distribution_commission_entries e WHERE e.beneficiary_agent_id=a.id) AS last_commission_at,
		a.created_at`
	offset := (filter.Page - 1) * filter.PageSize
	scope := newDistributionAnalyticsScope(service.DistributionAnalyticsFilter{Days: filter.Days, DateFrom: filter.DateFrom, DateTo: filter.DateTo})
	sortExpr := distributionAgentSortExpression(filter.SortBy)
	// Composite period metrics are SELECT-list aliases. PostgreSQL only permits
	// aliases as a bare ORDER BY item, not inside an expression (for example
	// `period_customer_paid_cny + team_customer_paid_cny`). Wrap the projection
	// so the aliases become real columns before applying the composite sort.
	dataFrom := strings.Replace(from, "\n\t\tWHERE", analyticsJoins+"\n\t\tWHERE", 1)
	query := analyticsCTE + selectColumns + dataFrom + ` ORDER BY ` + sortExpr + ` ` + filter.SortOrder + ` NULLS LAST,a.id DESC LIMIT $7 OFFSET $8`
	if filter.SortBy == "period_customer_paid" || filter.SortBy == "period_commission" || filter.SortBy == "period_customers" || filter.SortBy == "period_paying_customers" {
		query = analyticsCTE + `SELECT * FROM (` + selectColumns + dataFrom + `) AS agent_rows ORDER BY ` + sortExpr + ` ` + filter.SortOrder + ` NULLS LAST,id DESC LIMIT $7 OFFSET $8`
	}
	rows, err := r.db.QueryContext(ctx, query, filter.Search, like, filter.Status, filter.Depth, scope.currentStart, scope.currentEnd, filter.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionAgent, 0)
	for rows.Next() {
		var a service.DistributionAgent
		if err = rows.Scan(&a.ID, &a.UserID, &a.LevelID, &a.Depth, &a.ParentAgentID, &a.PromotionCode, &a.RateOverrideBPS,
			&a.EffectiveRateBPS, &a.MaxChildRateBPS, &a.CanRecruitSubagents, &a.CanViewPromotionStats, &a.Status, &a.AvailableCNY, &a.FrozenCNY,
			&a.ReservedCNY, &a.DebtCNY, &a.TotalEarnedCNY, &a.TotalWithdrawnCNY,
			&a.Email, &a.Username, &a.UserStatus, &a.ParentUserID, &a.ParentEmail, &a.ParentUsername,
			&a.CustomerCount, &a.TeamCount, &a.PayingCustomerCount, &a.CustomerPaidCNY,
			&a.ThisMonthCommissionCNY, &a.PeriodCustomerCount, &a.PeriodPayingCustomers, &a.PeriodCustomerPaidCNY, &a.PeriodCommissionCNY,
			&a.TeamCustomerCount, &a.TeamPayingCustomers, &a.TeamCustomerPaidCNY, &a.TeamCommissionCNY,
			&a.PeriodActivatedCustomers, &a.PeriodCohortPaidCustomers, &a.PeriodAllPayingCustomers, &a.PeriodRepurchaseCustomers, &a.PeriodActiveCustomers, &a.TotalTeamCustomers, &a.PeriodRewardCNY,
			&a.TotalConvertedCNY, &a.LastCommissionAt, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, rows.Err()
}

func distributionAgentSortExpression(sortBy string) string {
	switch sortBy {
	case "email":
		return "u.email"
	case "effective_rate":
		return "COALESCE(a.rate_override_bps,l.default_rate_bps)"
	case "customer_count":
		return "customer_count"
	case "team_count":
		return "team_count"
	case "paying_customer_count":
		return "paying_customer_count"
	case "customer_paid":
		return "customer_paid_cny"
	case "total_earned":
		return "COALESCE(w.total_earned_cny,0)"
	case "available":
		return "COALESCE(w.available_cny,0)"
	case "frozen":
		return "COALESCE(w.frozen_cny,0)"
	case "last_commission_at":
		return "last_commission_at"
	case "status":
		return "a.status"
	case "period_customer_paid":
		return "period_customer_paid_cny + team_customer_paid_cny"
	case "period_commission":
		return "period_commission_cny + team_commission_cny"
	case "period_customers":
		return "period_customer_count + team_customer_count"
	case "period_paying_customers":
		return "period_paying_customers + team_paying_customers"
	default:
		return "a.created_at"
	}
}

func (r *distributionRepository) AdminListCustomers(ctx context.Context, filter service.DistributionAdminListFilter) ([]service.DistributionCustomer, int64, error) {
	like := "%" + filter.Search + "%"
	const from = `
		FROM distribution_customer_bindings b
		JOIN users u ON u.id=b.user_id
		JOIN distribution_agents a ON a.id=b.agent_id
		JOIN distribution_agent_levels l ON l.id=a.level_id
		JOIN users agent_user ON agent_user.id=a.user_id
		WHERE ($1='' OR u.email ILIKE $2 OR u.username ILIKE $2 OR agent_user.email ILIKE $2
			OR agent_user.username ILIKE $2 OR a.promotion_code ILIKE $2)
		  AND ($3=0 OR a.id=$3)
		  AND ($4=0 OR l.depth=$4)`

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+from, filter.Search, like, filter.AgentID, filter.Depth).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (filter.Page - 1) * filter.PageSize
	sortExpr := distributionCustomerSortExpression(filter.SortBy)
	rows, err := r.db.QueryContext(ctx, `SELECT u.id,COALESCE(u.email,''),COALESCE(u.username,''),u.created_at,
		a.id,a.user_id,a.promotion_code,COALESCE(agent_user.email,''),COALESCE(agent_user.username,''),l.depth,b.bound_at,
		(SELECT COUNT(*) FROM distribution_commission_sources s WHERE s.customer_user_id=u.id) AS order_count,
		COALESCE((SELECT SUM(s.commission_base_cny) FROM distribution_commission_sources s WHERE s.customer_user_id=u.id),0) AS total_paid_cny,
		COALESCE((SELECT SUM(GREATEST(e.original_amount_cny-e.reversed_amount_cny,0)) FROM distribution_commission_entries e
			JOIN distribution_commission_sources s ON s.id=e.source_id WHERE s.customer_user_id=u.id),0) AS commission_cny,
		COALESCE((SELECT SUM(s.refunded_amount_cny) FROM distribution_commission_sources s WHERE s.customer_user_id=u.id),0) AS refunded_cny,
		(SELECT MAX(s.paid_at) FROM distribution_commission_sources s WHERE s.customer_user_id=u.id) AS last_paid_at`+from+`
		ORDER BY `+sortExpr+` `+filter.SortOrder+` NULLS LAST,u.id DESC LIMIT $5 OFFSET $6`, filter.Search, like, filter.AgentID, filter.Depth, filter.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionCustomer, 0)
	for rows.Next() {
		var item service.DistributionCustomer
		if err = rows.Scan(&item.UserID, &item.Email, &item.Username, &item.RegisteredAt, &item.AgentID, &item.AgentUserID,
			&item.AgentPromotionCode, &item.AgentEmail, &item.AgentUsername, &item.AgentDepth, &item.BoundAt,
			&item.OrderCount, &item.TotalPaidCNY, &item.CommissionCNY, &item.RefundedCNY, &item.LastPaidAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func distributionCustomerSortExpression(sortBy string) string {
	switch sortBy {
	case "email":
		return "u.email"
	case "agent":
		return "agent_user.email"
	case "order_count":
		return "order_count"
	case "total_paid":
		return "total_paid_cny"
	case "commission":
		return "commission_cny"
	case "refunded":
		return "refunded_cny"
	case "last_paid_at":
		return "last_paid_at"
	case "registered_at":
		return "u.created_at"
	default:
		return "b.bound_at"
	}
}

func (r *distributionRepository) AdminLookupAgentCandidates(ctx context.Context, query string) ([]service.DistributionUserOption, error) {
	like := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, `SELECT u.id,COALESCE(u.email,''),COALESCE(u.username,''),u.status,
		(u.status='active'
			 AND NOT EXISTS(SELECT 1 FROM distribution_agents a WHERE a.user_id=u.id)
			 AND NOT EXISTS(SELECT 1 FROM user_affiliates ua WHERE ua.user_id=u.id AND ua.inviter_id IS NOT NULL)),
		CASE
			WHEN u.status<>'active' THEN 'user_inactive'
			WHEN EXISTS(SELECT 1 FROM distribution_agents a WHERE a.user_id=u.id) THEN 'already_agent'
			WHEN EXISTS(SELECT 1 FROM distribution_customer_bindings b WHERE b.user_id=u.id) THEN 'distribution_customer'
			WHEN EXISTS(SELECT 1 FROM user_affiliates ua WHERE ua.user_id=u.id AND ua.inviter_id IS NOT NULL) THEN 'affiliate_invitee'
			ELSE ''
		END
		FROM users u
		WHERE u.deleted_at IS NULL AND (u.email ILIKE $1 OR u.username ILIKE $1)
		ORDER BY CASE WHEN LOWER(u.email)=LOWER($2) THEN 0 ELSE 1 END,u.email
		LIMIT 20`, like, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionUserOption, 0)
	for rows.Next() {
		var item service.DistributionUserOption
		if err = rows.Scan(&item.UserID, &item.Email, &item.Username, &item.Status, &item.Selectable, &item.UnavailableReason); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *distributionRepository) LookupEligibleUserByExactEmail(ctx context.Context, email string) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(ctx, `SELECT u.id FROM users u
		WHERE u.deleted_at IS NULL AND u.status='active' AND LOWER(u.email)=LOWER($1)
		  AND NOT EXISTS(SELECT 1 FROM distribution_agents a WHERE a.user_id=u.id)
		  AND NOT EXISTS(SELECT 1 FROM distribution_customer_bindings b WHERE b.user_id=u.id)
		  AND NOT EXISTS(SELECT 1 FROM user_affiliates ua WHERE ua.user_id=u.id AND ua.inviter_id IS NOT NULL)`, email).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, service.ErrSubagentCandidateUnavailable
	}
	return userID, err
}

func (r *distributionRepository) AdminLookupAgents(ctx context.Context, query string, includeInactive bool) ([]service.DistributionAgentOption, error) {
	like := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, `SELECT a.id,a.user_id,COALESCE(u.email,''),COALESCE(u.username,''),a.promotion_code,l.depth,a.status
		FROM distribution_agents a
		JOIN users u ON u.id=a.user_id
		JOIN distribution_agent_levels l ON l.id=a.level_id
		WHERE ($3 OR (a.status='active' AND u.deleted_at IS NULL))
		  AND (a.id::text=$2 OR u.email ILIKE $1 OR u.username ILIKE $1 OR a.promotion_code ILIKE $1)
		ORDER BY CASE WHEN a.id::text=$2 THEN 0 WHEN LOWER(u.email)=LOWER($2) THEN 1 ELSE 2 END,u.email
		LIMIT 20`, like, query, includeInactive)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionAgentOption, 0)
	for rows.Next() {
		var item service.DistributionAgentOption
		if err = rows.Scan(&item.AgentID, &item.UserID, &item.Email, &item.Username, &item.PromotionCode, &item.Depth, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *distributionRepository) AdminListAgentEvents(ctx context.Context, agentID int64) ([]service.DistributionAgentEvent, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT event.id,event.agent_id,event.event_type,COALESCE(event.old_status,''),COALESCE(event.new_status,''),
		event.old_rate_override_bps,event.new_rate_override_bps,event.old_effective_rate_bps,event.new_effective_rate_bps,
		event.reason,event.actor_user_id,COALESCE(actor.email,''),event.created_at
		FROM distribution_agent_events event LEFT JOIN users actor ON actor.id=event.actor_user_id
		WHERE event.agent_id=$1 ORDER BY event.created_at DESC,event.id DESC LIMIT 100`, agentID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionAgentEvent, 0)
	for rows.Next() {
		var item service.DistributionAgentEvent
		var oldOverride, newOverride, oldEffective, newEffective sql.NullInt64
		var actor sql.NullInt64
		if err = rows.Scan(&item.ID, &item.AgentID, &item.EventType, &item.OldStatus, &item.NewStatus,
			&oldOverride, &newOverride, &oldEffective, &newEffective, &item.Reason, &actor, &item.ActorEmail, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.OldRateOverrideBPS = distributionNullableInt(oldOverride)
		item.NewRateOverrideBPS = distributionNullableInt(newOverride)
		item.OldEffectiveRateBPS = distributionNullableInt(oldEffective)
		item.NewEffectiveRateBPS = distributionNullableInt(newEffective)
		if actor.Valid {
			value := actor.Int64
			item.ActorUserID = &value
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *distributionRepository) AdminListBindingEvents(ctx context.Context, customerUserID int64) ([]service.DistributionBindingEvent, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT event.id,event.customer_user_id,event.old_agent_id,event.new_agent_id,
		COALESCE(old_user.email,''),COALESCE(new_user.email,''),COALESCE(event.old_promotion_code,''),event.new_promotion_code,
		event.reason,event.actor_user_id,COALESCE(actor.email,''),event.created_at
		FROM distribution_binding_events event
		LEFT JOIN distribution_agents old_agent ON old_agent.id=event.old_agent_id LEFT JOIN users old_user ON old_user.id=old_agent.user_id
		JOIN distribution_agents new_agent ON new_agent.id=event.new_agent_id JOIN users new_user ON new_user.id=new_agent.user_id
		LEFT JOIN users actor ON actor.id=event.actor_user_id
		WHERE event.customer_user_id=$1 ORDER BY event.created_at DESC,event.id DESC LIMIT 100`, customerUserID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionBindingEvent, 0)
	for rows.Next() {
		var item service.DistributionBindingEvent
		var oldAgent, actor sql.NullInt64
		if err = rows.Scan(&item.ID, &item.CustomerUserID, &oldAgent, &item.NewAgentID, &item.OldAgentEmail, &item.NewAgentEmail,
			&item.OldPromotionCode, &item.NewPromotionCode, &item.Reason, &actor, &item.ActorEmail, &item.CreatedAt); err != nil {
			return nil, err
		}
		if oldAgent.Valid {
			value := oldAgent.Int64
			item.OldAgentID = &value
		}
		if actor.Valid {
			value := actor.Int64
			item.ActorUserID = &value
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func distributionNullableInt(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	result := int(value.Int64)
	return &result
}
func (r *distributionRepository) AdminListCommissions(ctx context.Context, filter service.DistributionAdminCommissionListFilter) ([]service.DistributionCommission, int64, error) {
	if err := r.releaseAllMatured(ctx); err != nil {
		return nil, 0, err
	}
	like := "%" + filter.Search + "%"
	const from = ` FROM distribution_commission_entries e
		JOIN distribution_commission_sources s ON s.id=e.source_id
		JOIN payment_orders po ON po.id=s.payment_order_id
		JOIN users customer ON customer.id=s.customer_user_id
		JOIN distribution_agents beneficiary ON beneficiary.id=e.beneficiary_agent_id
		JOIN users agent_user ON agent_user.id=beneficiary.user_id
		WHERE ($1='' OR customer.email ILIKE $2 OR customer.username ILIKE $2 OR agent_user.email ILIKE $2
			OR agent_user.username ILIKE $2 OR po.out_trade_no ILIKE $2 OR beneficiary.promotion_code ILIKE $2)
		  AND ($3='' OR e.status=$3) AND ($4='' OR e.entry_type=$4)
		  AND ($5='' OR s.payment_type=$5) AND ($6=0 OR beneficiary.id=$6)
		  AND ($7::timestamptz IS NULL OR e.created_at >= $7)
		  AND ($8::timestamptz IS NULL OR e.created_at < $8)`
	args := []any{filter.Search, like, filter.Status, filter.EntryType, filter.PaymentType, filter.AgentID, filter.DateFrom, filter.DateTo}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+from, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	sortExpr := distributionCommissionSortExpression(filter.SortBy)
	query := `SELECT e.id,s.customer_user_id,e.entry_type,e.rate_bps,e.original_amount_cny,e.reversed_amount_cny,e.status,
		e.created_at,e.available_at,COALESCE(s.commission_base_cny,0),COALESCE(customer.email,''),COALESCE(customer.username,''),
		s.payment_type,s.payment_order_id,COALESCE(po.out_trade_no,''),beneficiary.id,COALESCE(agent_user.email,''),
		COALESCE(agent_user.username,''),s.payment_currency,s.actual_paid_amount,COALESCE(s.fx_rate_to_cny,0)` + from +
		` ORDER BY ` + sortExpr + ` ` + filter.SortOrder + ` NULLS LAST,e.id DESC LIMIT $9 OFFSET $10`
	rows, err := r.db.QueryContext(ctx, query, append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionCommission, 0)
	for rows.Next() {
		var x service.DistributionCommission
		if err = rows.Scan(&x.ID, &x.CustomerUserID, &x.EntryType, &x.RateBPS, &x.OriginalAmountCNY,
			&x.ReversedAmountCNY, &x.Status, &x.CreatedAt, &x.AvailableAt, &x.CommissionBaseCNY,
			&x.CustomerEmail, &x.CustomerUsername, &x.PaymentType, &x.PaymentOrderID, &x.OrderNo,
			&x.BeneficiaryAgentID, &x.AgentEmail, &x.AgentUsername, &x.PaymentCurrency,
			&x.ActualPaidAmount, &x.FXRateToCNY); err != nil {
			return nil, 0, err
		}
		items = append(items, x)
	}
	return items, total, rows.Err()
}

func distributionCommissionSortExpression(sortBy string) string {
	switch sortBy {
	case "customer":
		return "customer.email"
	case "agent":
		return "agent_user.email"
	case "commission_base":
		return "s.commission_base_cny"
	case "rate":
		return "e.rate_bps"
	case "commission":
		return "(e.original_amount_cny-e.reversed_amount_cny)"
	case "available_at":
		return "e.available_at"
	case "status":
		return "e.status"
	default:
		return "e.created_at"
	}
}

func (r *distributionRepository) AdminReviewWithdrawal(ctx context.Context, id, adminID int64, status, note, reference string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err = adminReviewWithdrawalTx(ctx, tx, id, adminID, status, note, reference); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *distributionRepository) AdminBatchReviewWithdrawals(ctx context.Context, ids []int64, adminID int64, status, note string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, id := range ids {
		if err = adminReviewWithdrawalTx(ctx, tx, id, adminID, status, note, ""); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func adminReviewWithdrawalTx(ctx context.Context, tx *sql.Tx, id, adminID int64, status, note, reference string) error {
	var agentID int64
	var amount decimal.Decimal
	var current string
	var approvedBy sql.NullInt64
	var dualApproval bool
	err := tx.QueryRowContext(ctx, `SELECT withdrawal.agent_id,withdrawal.amount_cny,withdrawal.status,withdrawal.approved_by,settings.withdrawal_dual_approval_enabled
		FROM distribution_withdrawals withdrawal CROSS JOIN distribution_settings settings WHERE withdrawal.id=$1 AND settings.id=1 FOR UPDATE OF withdrawal`, id).
		Scan(&agentID, &amount, &current, &approvedBy, &dualApproval)
	if errors.Is(err, sql.ErrNoRows) {
		return infraerrors.NotFound("WITHDRAWAL_NOT_FOUND", "withdrawal not found")
	}
	if err != nil {
		return err
	}
	allowed := false
	switch status {
	case "approved":
		allowed = current == "pending"
	case "rejected":
		allowed = current == "pending" || current == "approved"
	case "paying":
		allowed = current == "approved"
	case "paid":
		allowed = current == "paying"
	case "failed":
		allowed = current == "paying"
	}
	if !allowed {
		return infraBadRequest("INVALID_WITHDRAWAL_TRANSITION", "invalid withdrawal status transition")
	}
	if status == "paying" && dualApproval && (!approvedBy.Valid || approvedBy.Int64 == adminID) {
		return infraBadRequest("WITHDRAWAL_DUAL_APPROVAL_REQUIRED", "a different administrator must start payment after approval")
	}
	if status == "rejected" || status == "failed" {
		_, err = tx.ExecContext(ctx, `UPDATE distribution_wallets SET reserved_cny=reserved_cny-$2,available_cny=available_cny+$2,updated_at=NOW() WHERE agent_id=$1`, agentID, amount)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO distribution_wallet_ledger(agent_id,withdrawal_id,entry_type,amount_cny,frozen_after_cny,available_after_cny,reserved_after_cny,debt_after_cny,idempotency_key,note) SELECT agent_id,$2,'withdrawal_released',$3,frozen_cny,available_cny,reserved_cny,debt_cny,$4,$5 FROM distribution_wallets WHERE agent_id=$1`, agentID, id, amount, fmt.Sprintf("withdrawal:%d:released", id), note)
		if err != nil {
			return err
		}
	}
	if status == "paid" {
		_, err = tx.ExecContext(ctx, `UPDATE distribution_wallets SET reserved_cny=reserved_cny-$2,total_withdrawn_cny=total_withdrawn_cny+$2,updated_at=NOW() WHERE agent_id=$1`, agentID, amount)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO distribution_wallet_ledger(agent_id,withdrawal_id,entry_type,amount_cny,frozen_after_cny,available_after_cny,reserved_after_cny,debt_after_cny,idempotency_key,note) SELECT agent_id,$2,'withdrawal_paid',$3,frozen_cny,available_cny,reserved_cny,debt_cny,$4,$5 FROM distribution_wallets WHERE agent_id=$1`, agentID, id, amount.Neg(), fmt.Sprintf("withdrawal:%d:paid", id), reference)
		if err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE distribution_withdrawals SET status=$2::varchar,review_note=$3,reviewed_by=$4,reviewed_at=COALESCE(reviewed_at,NOW()),
		approved_by=CASE WHEN $2::varchar='approved' THEN $4 ELSE approved_by END,approved_at=CASE WHEN $2::varchar='approved' THEN NOW() ELSE approved_at END,
		paid_at=CASE WHEN $2::varchar='paid' THEN NOW() ELSE paid_at END,payment_reference=CASE WHEN $2::varchar='paid' THEN $5::varchar ELSE payment_reference END,updated_at=NOW() WHERE id=$1`, id, status, note, adminID, reference)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_withdrawal_events
		(withdrawal_id,from_status,to_status,note,payment_reference,actor_user_id)
		VALUES($1,$2::varchar,$3::varchar,$4,NULLIF($5::text,''),$6)`, id, current, status, note, reference, adminID); err != nil {
		return err
	}
	return nil
}

func (r *distributionRepository) AdminListWithdrawals(ctx context.Context, filter service.DistributionAdminWithdrawalListFilter) ([]service.DistributionWithdrawal, int64, error) {
	like := "%" + filter.Search + "%"
	const from = ` FROM distribution_withdrawals w
		JOIN distribution_agents a ON a.id=w.agent_id
		JOIN users agent_user ON agent_user.id=a.user_id
		LEFT JOIN users reviewer ON reviewer.id=w.reviewed_by
		LEFT JOIN users approver ON approver.id=w.approved_by
		WHERE ($1='' OR w.request_no ILIKE $2 OR agent_user.email ILIKE $2 OR agent_user.username ILIKE $2
			OR w.alipay_account ILIKE $2 OR COALESCE(w.payment_reference,'') ILIKE $2)
		  AND ($3='' OR w.status=$3) AND ($4=0 OR w.agent_id=$4)
		  AND ($5::timestamptz IS NULL OR w.created_at >= $5)
		  AND ($6::timestamptz IS NULL OR w.created_at < $6)`
	args := []any{filter.Search, like, filter.Status, filter.AgentID, filter.DateFrom, filter.DateTo}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+from, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	sortExpr := distributionWithdrawalSortExpression(filter.SortBy)
	query := `SELECT w.id,w.request_no,w.agent_id,a.user_id,COALESCE(agent_user.email,''),COALESCE(agent_user.username,''),
		w.alipay_name,w.alipay_account,w.amount_cny,w.fee_cny,w.payout_cny,w.status,COALESCE(w.review_note,''),
		w.reviewed_by,COALESCE(reviewer.email,''),w.reviewed_at,w.approved_by,COALESCE(approver.email,''),w.approved_at,w.paid_at,COALESCE(w.payment_reference,''),
		(SELECT COUNT(*) FROM distribution_withdrawal_attachments attachment WHERE attachment.withdrawal_id=w.id AND attachment.deleted_at IS NULL),w.created_at` + from +
		` ORDER BY ` + sortExpr + ` ` + filter.SortOrder + ` NULLS LAST,w.id DESC LIMIT $7 OFFSET $8`
	rows, err := r.db.QueryContext(ctx, query, append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionWithdrawal, 0)
	for rows.Next() {
		var w service.DistributionWithdrawal
		var reviewedBy, approvedBy sql.NullInt64
		var reviewedAt, approvedAt, paidAt sql.NullTime
		if err = rows.Scan(&w.ID, &w.RequestNo, &w.AgentID, &w.AgentUserID, &w.AgentEmail, &w.AgentUsername,
			&w.AlipayName, &w.AlipayAccount, &w.AmountCNY, &w.FeeCNY, &w.PayoutCNY, &w.Status,
			&w.ReviewNote, &reviewedBy, &w.ReviewerEmail, &reviewedAt, &approvedBy, &w.ApproverEmail, &approvedAt, &paidAt, &w.PaymentReference,
			&w.AttachmentCount, &w.CreatedAt); err != nil {
			return nil, 0, err
		}
		if reviewedBy.Valid {
			value := reviewedBy.Int64
			w.ReviewedBy = &value
		}
		if reviewedAt.Valid {
			value := reviewedAt.Time
			w.ReviewedAt = &value
		}
		if approvedBy.Valid {
			value := approvedBy.Int64
			w.ApprovedBy = &value
		}
		if approvedAt.Valid {
			value := approvedAt.Time
			w.ApprovedAt = &value
		}
		if paidAt.Valid {
			value := paidAt.Time
			w.PaidAt = &value
		}
		items = append(items, w)
	}
	return items, total, rows.Err()
}

func distributionWithdrawalSortExpression(sortBy string) string {
	switch sortBy {
	case "agent":
		return "agent_user.email"
	case "amount":
		return "w.amount_cny"
	case "payout":
		return "w.payout_cny"
	case "status":
		return "w.status"
	case "reviewed_at":
		return "w.reviewed_at"
	case "paid_at":
		return "w.paid_at"
	default:
		return "w.created_at"
	}
}

func (r *distributionRepository) AdminListAnomalies(ctx context.Context, filter service.DistributionAdminAnomalyListFilter) ([]service.DistributionAdminAnomaly, int64, error) {
	if err := r.releaseAllMatured(ctx); err != nil {
		return nil, 0, err
	}
	const anomaliesCTE = `WITH anomalies AS (
		SELECT 'overdue_withdrawal:'||w.id::text AS id,'overdue_withdrawal'::text AS type,'high'::text AS severity,
			'withdrawal'::text AS entity_type,w.id AS entity_id,w.request_no AS reference,a.id AS agent_id,
			COALESCE(u.email,'') AS agent_email,COALESCE(u.email,w.request_no) AS subject,w.payout_cny AS amount_cny,
			w.created_at + INTERVAL '24 hours' AS detected_at,'待审核超过 24 小时'::text AS description
		FROM distribution_withdrawals w JOIN distribution_agents a ON a.id=w.agent_id JOIN users u ON u.id=a.user_id
		WHERE w.status='pending' AND w.created_at<NOW()-INTERVAL '24 hours'
		UNION ALL
		SELECT 'pending_fx:'||source.id::text,'pending_fx','high','commission_source',source.id,orders.out_trade_no,
			agent.id,COALESCE(agent_user.email,''),COALESCE(customer.email,orders.out_trade_no),0::numeric,source.created_at + INTERVAL '15 minutes','美元订单等待计佣汇率超过 15 分钟'
		FROM distribution_commission_sources source
		JOIN payment_orders orders ON orders.id=source.payment_order_id
		JOIN users customer ON customer.id=source.customer_user_id
		LEFT JOIN distribution_agents agent ON agent.id=source.direct_agent_id
		LEFT JOIN users agent_user ON agent_user.id=agent.user_id
		WHERE source.status='pending_fx' AND source.created_at<NOW()-INTERVAL '15 minutes'
		UNION ALL
		SELECT 'agent_debt:'||agent.id::text,'agent_debt','critical','agent',agent.id,agent.promotion_code,agent.id,
			COALESCE(u.email,''),COALESCE(u.email,agent.promotion_code),wallet.debt_cny,wallet.updated_at,'代理佣金钱包存在负债欠款'
		FROM distribution_wallets wallet JOIN distribution_agents agent ON agent.id=wallet.agent_id JOIN users u ON u.id=agent.user_id
		WHERE wallet.debt_cny>0
		UNION ALL
		SELECT 'inactive_agent_customers:'||binding.user_id::text,'inactive_agent_customers','medium','customer',binding.user_id,
			COALESCE(customer.email,binding.user_id::text),agent.id,COALESCE(agent_user.email,''),COALESCE(customer.email,binding.user_id::text),0::numeric,
			binding.bound_at,'客户归属于已暂停或已撤销代理'
		FROM distribution_customer_bindings binding
		JOIN distribution_agents agent ON agent.id=binding.agent_id JOIN users agent_user ON agent_user.id=agent.user_id
		JOIN users customer ON customer.id=binding.user_id WHERE agent.status IN ('suspended','revoked')
		UNION ALL
		SELECT 'matured_commission:'||entry.id::text,'matured_commission','medium','commission',entry.id,orders.out_trade_no,
			agent.id,COALESCE(agent_user.email,''),COALESCE(customer.email,orders.out_trade_no),
			GREATEST(entry.original_amount_cny-entry.reversed_amount_cny,0),entry.available_at,'佣金已到解冻时间但仍处于冻结状态'
		FROM distribution_commission_entries entry
		JOIN distribution_commission_sources source ON source.id=entry.source_id JOIN payment_orders orders ON orders.id=source.payment_order_id
		JOIN users customer ON customer.id=source.customer_user_id JOIN distribution_agents agent ON agent.id=entry.beneficiary_agent_id
		JOIN users agent_user ON agent_user.id=agent.user_id WHERE entry.status='frozen' AND entry.available_at<=NOW()
	)`
	like := "%" + filter.Search + "%"
	const where = ` WHERE ($1::text='' OR type=$1::text) AND ($2::text='' OR severity=$2::text)
		AND ($3::text='' OR reference ILIKE $4::text OR agent_email ILIKE $4::text OR subject ILIKE $4::text OR description ILIKE $4::text)`
	args := []any{filter.Type, filter.Severity, filter.Search, like}
	var total int64
	if err := r.db.QueryRowContext(ctx, anomaliesCTE+` SELECT COUNT(*) FROM anomalies`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	order := "DESC"
	if filter.SortOrder == "asc" {
		order = "ASC"
	}
	query := anomaliesCTE + ` SELECT id,type,severity,entity_type,entity_id,reference,agent_id,agent_email,subject,amount_cny,detected_at,description
		FROM anomalies` + where + ` ORDER BY CASE severity WHEN 'critical' THEN 3 WHEN 'high' THEN 2 ELSE 1 END DESC,detected_at ` + order + `,id LIMIT $5 OFFSET $6`
	rows, err := r.db.QueryContext(ctx, query, append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionAdminAnomaly, 0)
	for rows.Next() {
		var item service.DistributionAdminAnomaly
		var agentID sql.NullInt64
		if err = rows.Scan(&item.ID, &item.Type, &item.Severity, &item.EntityType, &item.EntityID, &item.Reference, &agentID,
			&item.AgentEmail, &item.Subject, &item.AmountCNY, &item.DetectedAt, &item.Description); err != nil {
			return nil, 0, err
		}
		if agentID.Valid {
			value := agentID.Int64
			item.AgentID = &value
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *distributionRepository) AdminGetWithdrawal(ctx context.Context, id int64) (*service.DistributionWithdrawalDetail, error) {
	filter := service.DistributionAdminWithdrawalListFilter{Page: 1, PageSize: 1, Search: "", SortOrder: "desc"}
	_ = filter
	var w service.DistributionWithdrawal
	var reviewedBy, approvedBy sql.NullInt64
	var reviewedAt, approvedAt, paidAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `SELECT w.id,w.request_no,w.agent_id,a.user_id,COALESCE(agent_user.email,''),COALESCE(agent_user.username,''),
		w.alipay_name,w.alipay_account,w.amount_cny,w.fee_cny,w.payout_cny,w.status,COALESCE(w.review_note,''),
		w.reviewed_by,COALESCE(reviewer.email,''),w.reviewed_at,w.approved_by,COALESCE(approver.email,''),w.approved_at,w.paid_at,COALESCE(w.payment_reference,''),
		(SELECT COUNT(*) FROM distribution_withdrawal_attachments attachment WHERE attachment.withdrawal_id=w.id AND attachment.deleted_at IS NULL),w.created_at
		FROM distribution_withdrawals w JOIN distribution_agents a ON a.id=w.agent_id JOIN users agent_user ON agent_user.id=a.user_id
		LEFT JOIN users reviewer ON reviewer.id=w.reviewed_by LEFT JOIN users approver ON approver.id=w.approved_by WHERE w.id=$1`, id).
		Scan(&w.ID, &w.RequestNo, &w.AgentID, &w.AgentUserID, &w.AgentEmail, &w.AgentUsername, &w.AlipayName,
			&w.AlipayAccount, &w.AmountCNY, &w.FeeCNY, &w.PayoutCNY, &w.Status, &w.ReviewNote,
			&reviewedBy, &w.ReviewerEmail, &reviewedAt, &approvedBy, &w.ApproverEmail, &approvedAt, &paidAt, &w.PaymentReference, &w.AttachmentCount, &w.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("WITHDRAWAL_NOT_FOUND", "withdrawal not found")
	}
	if err != nil {
		return nil, err
	}
	if reviewedBy.Valid {
		value := reviewedBy.Int64
		w.ReviewedBy = &value
	}
	if reviewedAt.Valid {
		value := reviewedAt.Time
		w.ReviewedAt = &value
	}
	if approvedBy.Valid {
		value := approvedBy.Int64
		w.ApprovedBy = &value
	}
	if approvedAt.Valid {
		value := approvedAt.Time
		w.ApprovedAt = &value
	}
	if paidAt.Valid {
		value := paidAt.Time
		w.PaidAt = &value
	}
	attachments, err := r.listDistributionWithdrawalAttachments(ctx, id)
	if err != nil {
		return nil, err
	}
	events, err := r.listDistributionWithdrawalEvents(ctx, id)
	if err != nil {
		return nil, err
	}
	return &service.DistributionWithdrawalDetail{Withdrawal: &w, Attachments: attachments, Events: events}, nil
}

func (r *distributionRepository) listDistributionWithdrawalAttachments(ctx context.Context, id int64) ([]service.DistributionWithdrawalAttachment, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,withdrawal_id,object_key,original_name,content_type,size_bytes,sha256,evidence_type,note,uploaded_by,created_at
		FROM distribution_withdrawal_attachments WHERE withdrawal_id=$1 AND deleted_at IS NULL ORDER BY created_at,id`, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionWithdrawalAttachment, 0)
	for rows.Next() {
		var item service.DistributionWithdrawalAttachment
		var uploader sql.NullInt64
		if err = rows.Scan(&item.ID, &item.WithdrawalID, &item.ObjectKey, &item.OriginalName, &item.ContentType, &item.SizeBytes, &item.SHA256, &item.EvidenceType, &item.Note, &uploader, &item.CreatedAt); err != nil {
			return nil, err
		}
		if uploader.Valid {
			value := uploader.Int64
			item.UploadedBy = &value
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *distributionRepository) listDistributionWithdrawalEvents(ctx context.Context, id int64) ([]service.DistributionWithdrawalEvent, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT event.id,event.withdrawal_id,COALESCE(event.from_status,''),event.to_status,event.note,
		COALESCE(event.payment_reference,''),event.actor_user_id,COALESCE(actor.email,''),event.created_at
		FROM distribution_withdrawal_events event LEFT JOIN users actor ON actor.id=event.actor_user_id
		WHERE event.withdrawal_id=$1 ORDER BY event.created_at,event.id`, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionWithdrawalEvent, 0)
	for rows.Next() {
		var item service.DistributionWithdrawalEvent
		var actor sql.NullInt64
		if err = rows.Scan(&item.ID, &item.WithdrawalID, &item.FromStatus, &item.ToStatus, &item.Note, &item.PaymentReference, &actor, &item.ActorEmail, &item.CreatedAt); err != nil {
			return nil, err
		}
		if actor.Valid {
			value := actor.Int64
			item.ActorUserID = &value
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *distributionRepository) AdminCreateWithdrawalAttachments(ctx context.Context, withdrawalID, adminID int64, attachments []service.DistributionWithdrawalAttachment) error {
	if len(attachments) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM distribution_withdrawals WHERE id=$1)`, withdrawalID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return infraerrors.NotFound("WITHDRAWAL_NOT_FOUND", "withdrawal not found")
	}
	for _, item := range attachments {
		if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_withdrawal_attachments
		(withdrawal_id,object_key,original_name,content_type,size_bytes,sha256,evidence_type,note,uploaded_by)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, withdrawalID, item.ObjectKey, item.OriginalName, item.ContentType, item.SizeBytes, item.SHA256, item.EvidenceType, item.Note, adminID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *distributionRepository) GetPayoutAccount(ctx context.Context, userID int64) (*service.DistributionPayoutAccount, error) {
	var a service.DistributionPayoutAccount
	err := r.db.QueryRowContext(ctx, `SELECT p.alipay_name,p.alipay_account FROM distribution_payout_accounts p JOIN distribution_agents a ON a.id=p.agent_id WHERE a.user_id=$1`, userID).Scan(&a.AlipayName, &a.AlipayAccount)
	if errors.Is(err, sql.ErrNoRows) {
		return &a, nil
	}
	return &a, err
}
func (r *distributionRepository) UpsertPayoutAccount(ctx context.Context, userID int64, a service.DistributionPayoutAccount) error {
	res, err := r.db.ExecContext(ctx, `INSERT INTO distribution_payout_accounts(agent_id,alipay_name,alipay_account) SELECT id,$2,$3 FROM distribution_agents WHERE user_id=$1 AND status='active' ON CONFLICT(agent_id) DO UPDATE SET alipay_name=EXCLUDED.alipay_name,alipay_account=EXCLUDED.alipay_account,updated_at=NOW()`, userID, a.AlipayName, a.AlipayAccount)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return service.ErrDistributionNotAgent
	}
	return nil
}
func (r *distributionRepository) RequestWithdrawal(ctx context.Context, userID int64, amount decimal.Decimal) (*service.DistributionWithdrawal, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var agentID int64
	var available, debt decimal.Decimal
	var enabled bool
	var min, max, fixed, daily, monthly decimal.Decimal
	var feeBPS int
	err = tx.QueryRowContext(ctx, `SELECT a.id,w.available_cny,w.debt_cny,s.withdrawal_enabled,s.minimum_withdrawal_cny,s.maximum_withdrawal_cny,s.withdrawal_fee_fixed_cny,s.daily_withdrawal_limit_cny,s.monthly_withdrawal_limit_cny,s.withdrawal_fee_rate_bps FROM distribution_agents a JOIN distribution_wallets w ON w.agent_id=a.id CROSS JOIN distribution_settings s WHERE a.user_id=$1 AND a.status='active' FOR UPDATE OF w`, userID).Scan(&agentID, &available, &debt, &enabled, &min, &max, &fixed, &daily, &monthly, &feeBPS)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrDistributionNotAgent
	}
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, infraBadRequest("WITHDRAWAL_DISABLED", "withdrawal is disabled")
	}
	if debt.IsPositive() {
		return nil, infraBadRequest("AGENT_DEBT_OUTSTANDING", "outstanding debt must be cleared")
	}
	if amount.LessThan(min) || amount.GreaterThan(max) || amount.GreaterThan(available) {
		return nil, infraBadRequest("WITHDRAWAL_AMOUNT_OUT_OF_RANGE", "withdrawal amount is out of range")
	}
	var usedDay, usedMonth decimal.Decimal
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount_cny) FILTER(WHERE created_at>=date_trunc('day',NOW())),0),COALESCE(SUM(amount_cny) FILTER(WHERE created_at>=date_trunc('month',NOW())),0) FROM distribution_withdrawals WHERE agent_id=$1 AND status NOT IN('rejected','cancelled','failed')`, agentID).Scan(&usedDay, &usedMonth); err != nil {
		return nil, err
	}
	if daily.IsPositive() && usedDay.Add(amount).GreaterThan(daily) {
		return nil, infraBadRequest("DAILY_WITHDRAWAL_LIMIT", "daily withdrawal limit exceeded")
	}
	if monthly.IsPositive() && usedMonth.Add(amount).GreaterThan(monthly) {
		return nil, infraBadRequest("MONTHLY_WITHDRAWAL_LIMIT", "monthly withdrawal limit exceeded")
	}
	var name, account string
	err = tx.QueryRowContext(ctx, `SELECT alipay_name,alipay_account FROM distribution_payout_accounts WHERE agent_id=$1`, agentID).Scan(&name, &account)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraBadRequest("PAYOUT_ACCOUNT_REQUIRED", "Alipay account is required")
	}
	if err != nil {
		return nil, err
	}
	fee := amount.Mul(decimal.NewFromInt(int64(feeBPS))).Div(decimal.NewFromInt(10000)).Add(fixed).Round(8)
	payout := amount.Sub(fee)
	if !payout.IsPositive() {
		return nil, infraBadRequest("WITHDRAWAL_FEE_EXCEEDED", "withdrawal fee exceeds amount")
	}
	var w service.DistributionWithdrawal
	err = tx.QueryRowContext(ctx, `INSERT INTO distribution_withdrawals(agent_id,amount_cny,fee_cny,payout_cny,alipay_name,alipay_account) VALUES($1,$2,$3,$4,$5,$6) RETURNING id,request_no,amount_cny,fee_cny,payout_cny,status,created_at`, agentID, amount, fee, payout, name, account).Scan(&w.ID, &w.RequestNo, &w.AmountCNY, &w.FeeCNY, &w.PayoutCNY, &w.Status, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_withdrawal_events(withdrawal_id,to_status,note,actor_user_id)
		VALUES($1,'pending','withdrawal requested',$2)`, w.ID, userID); err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE distribution_wallets SET available_cny=available_cny-$2,reserved_cny=reserved_cny+$2,updated_at=NOW() WHERE agent_id=$1`, agentID, amount)
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO distribution_wallet_ledger(agent_id,withdrawal_id,entry_type,amount_cny,frozen_after_cny,available_after_cny,reserved_after_cny,debt_after_cny,idempotency_key) SELECT agent_id,$2,'withdrawal_reserved',$3,frozen_cny,available_cny,reserved_cny,debt_cny,$4 FROM distribution_wallets WHERE agent_id=$1`, agentID, w.ID, amount.Neg(), fmt.Sprintf("withdrawal:%d:reserved", w.ID))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &w, nil
}
func (r *distributionRepository) ListWithdrawals(ctx context.Context, userID int64, filter service.DistributionUserListFilter) ([]service.DistributionWithdrawal, int64, error) {
	var total int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_withdrawals w JOIN distribution_agents a ON a.id=w.agent_id WHERE a.user_id=$1 AND ($2='' OR w.status=$2)`, userID, filter.Status).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT w.id,w.request_no,w.agent_id,a.user_id,w.alipay_name,w.alipay_account,w.amount_cny,w.fee_cny,w.payout_cny,w.status,w.created_at FROM distribution_withdrawals w JOIN distribution_agents a ON a.id=w.agent_id WHERE a.user_id=$1 AND ($2='' OR w.status=$2) ORDER BY w.created_at DESC LIMIT $3 OFFSET $4`, userID, filter.Status, filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionWithdrawal, 0)
	for rows.Next() {
		var w service.DistributionWithdrawal
		if err = rows.Scan(&w.ID, &w.RequestNo, &w.AgentID, &w.AgentUserID, &w.AlipayName, &w.AlipayAccount, &w.AmountCNY, &w.FeeCNY, &w.PayoutCNY, &w.Status, &w.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, w)
	}
	return items, total, rows.Err()
}
func (r *distributionRepository) ConvertToBalance(ctx context.Context, userID int64, amount decimal.Decimal) (decimal.Decimal, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return decimal.Zero, err
	}
	defer func() { _ = tx.Rollback() }()
	var agentID int64
	var available, debt decimal.Decimal
	err = tx.QueryRowContext(ctx, `SELECT a.id,w.available_cny,w.debt_cny FROM distribution_agents a JOIN distribution_wallets w ON w.agent_id=a.id WHERE a.user_id=$1 AND a.status='active' FOR UPDATE OF w`, userID).Scan(&agentID, &available, &debt)
	if errors.Is(err, sql.ErrNoRows) {
		return decimal.Zero, service.ErrDistributionNotAgent
	}
	if err != nil {
		return decimal.Zero, err
	}
	if debt.IsPositive() {
		return decimal.Zero, infraBadRequest("AGENT_DEBT_OUTSTANDING", "outstanding debt must be cleared")
	}
	if amount.GreaterThan(available) {
		return decimal.Zero, infraBadRequest("INSUFFICIENT_COMMISSION", "insufficient available commission")
	}
	credit := amount.Round(8)
	_, err = tx.ExecContext(ctx, `UPDATE distribution_wallets SET available_cny=available_cny-$2,total_converted_cny=total_converted_cny+$2,updated_at=NOW() WHERE agent_id=$1`, agentID, amount)
	if err != nil {
		return decimal.Zero, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE users SET balance=balance+$2,updated_at=NOW() WHERE id=$1`, userID, credit)
	if err != nil {
		return decimal.Zero, err
	}
	key := fmt.Sprintf("conversion:%d:%d", agentID, time.Now().UnixNano())
	_, err = tx.ExecContext(ctx, `INSERT INTO distribution_wallet_ledger(agent_id,entry_type,amount_cny,frozen_after_cny,available_after_cny,reserved_after_cny,debt_after_cny,idempotency_key,note) SELECT agent_id,'balance_conversion',$2,frozen_cny,available_cny,reserved_cny,debt_cny,$3,$4 FROM distribution_wallets WHERE agent_id=$1`, agentID, amount.Neg(), key, "credited platform USD "+credit.String())
	if err != nil {
		return decimal.Zero, err
	}
	if err = tx.Commit(); err != nil {
		return decimal.Zero, err
	}
	return credit, nil
}

func randomDistributionCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 12)
	raw := make([]byte, 12)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return string(b), nil
}

func infraBadRequest(code, message string) error { return infraerrors.BadRequest(code, message) }

func (r *distributionRepository) GetOverview(ctx context.Context, userID int64, filter service.DistributionAnalyticsFilter) (*service.DistributionOverview, error) {
	a, err := r.GetAgentByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var overview service.DistributionOverview
	overview.Agent = a
	if err = r.db.QueryRowContext(ctx, `SELECT enabled,promotion_tracking_enabled FROM distribution_settings WHERE id=1`).Scan(
		&overview.DistributionEnabled, &overview.PromotionTrackingEnabled,
	); err != nil {
		return nil, err
	}
	err = r.db.QueryRowContext(ctx, `
		SELECT (SELECT COUNT(*) FROM distribution_customer_bindings WHERE agent_id = $1),
		       (SELECT COUNT(*) FROM distribution_agents WHERE parent_agent_id = $1 AND status <> 'revoked'),
		       (SELECT COUNT(DISTINCT s.customer_user_id) FROM distribution_commission_sources s JOIN distribution_commission_entries e ON e.source_id=s.id WHERE e.beneficiary_agent_id=$1 AND s.status<>'void' AND COALESCE(s.commission_base_cny,0)>0),
		       (SELECT COUNT(*) FROM distribution_customer_bindings WHERE agent_id=$1 AND bound_at>=date_trunc('month',NOW())),
		       COALESCE((SELECT SUM(DISTINCT_SOURCE.base_cny) FROM (SELECT s.id,GREATEST(COALESCE(s.commission_base_cny,0)-s.refunded_amount_cny,0) base_cny FROM distribution_commission_sources s JOIN distribution_commission_entries e ON e.source_id=s.id WHERE e.beneficiary_agent_id=$1 AND s.status<>'void' GROUP BY s.id) DISTINCT_SOURCE),0),
		       COALESCE((SELECT SUM(DISTINCT_SOURCE.base_cny) FROM (SELECT s.id,GREATEST(COALESCE(s.commission_base_cny,0)-s.refunded_amount_cny,0) base_cny FROM distribution_commission_sources s JOIN distribution_commission_entries e ON e.source_id=s.id WHERE e.beneficiary_agent_id=$1 AND s.status<>'void' AND s.paid_at>=date_trunc('month',NOW()) GROUP BY s.id) DISTINCT_SOURCE),0),
		       COALESCE((SELECT SUM(GREATEST(e.original_amount_cny-e.reversed_amount_cny,0)) FROM distribution_commission_entries e WHERE e.beneficiary_agent_id=$1 AND e.created_at>=date_trunc('month',NOW())),0)`, a.ID).Scan(
		&overview.CustomerCount, &overview.TeamCount, &overview.PayingCustomerCount, &overview.NewCustomersThisMonth,
		&overview.CustomerPaidCNY, &overview.ThisMonthCustomerPaid, &overview.ThisMonthCommissionCNY,
	)
	if err != nil {
		return nil, err
	}
	overview.Analytics, err = r.distributionBusinessAnalytics(ctx, a.ID, filter)
	if err != nil {
		return nil, err
	}
	if a.Depth == 1 {
		overview.TeamRanking, err = r.distributionAgentRanking(ctx, a.ID, filter, 8)
		if err != nil {
			return nil, err
		}
	} else {
		overview.TeamRanking = make([]service.DistributionAgentRanking, 0)
	}
	return &overview, nil
}
func (r *distributionRepository) ListCustomers(ctx context.Context, userID int64, filter service.DistributionUserListFilter) ([]service.DistributionCustomer, int64, error) {
	like := "%" + filter.Search + "%"
	const from = ` FROM distribution_customer_bindings b JOIN distribution_agents a ON a.id=b.agent_id JOIN users u ON u.id=b.user_id WHERE a.user_id=$1 AND ($2='' OR u.email ILIKE $3 OR COALESCE(u.username,'') ILIKE $3)`
	var total int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+from, userID, filter.Search, like).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT u.id,COALESCE(u.email,''),COALESCE(u.username,''),b.bound_at,
		(SELECT COUNT(*) FROM distribution_commission_sources s WHERE s.direct_agent_id=a.id AND s.customer_user_id=u.id AND s.status<>'void'),
		COALESCE((SELECT SUM(GREATEST(COALESCE(s.commission_base_cny,0)-s.refunded_amount_cny,0)) FROM distribution_commission_sources s WHERE s.direct_agent_id=a.id AND s.customer_user_id=u.id AND s.status<>'void'),0),
		COALESCE((SELECT SUM(GREATEST(e.original_amount_cny-e.reversed_amount_cny,0)) FROM distribution_commission_entries e JOIN distribution_commission_sources s ON s.id=e.source_id WHERE e.beneficiary_agent_id=a.id AND s.customer_user_id=u.id),0)`+from+` ORDER BY b.bound_at DESC LIMIT $4 OFFSET $5`, userID, filter.Search, like, filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionCustomer, 0)
	for rows.Next() {
		var x service.DistributionCustomer
		if err = rows.Scan(&x.UserID, &x.Email, &x.Username, &x.BoundAt, &x.OrderCount, &x.TotalPaidCNY, &x.CommissionCNY); err != nil {
			return nil, 0, err
		}
		x.Email = maskDistributionEmail(x.Email)
		items = append(items, x)
	}
	return items, total, rows.Err()
}
func (r *distributionRepository) ListCommissions(ctx context.Context, userID int64, filter service.DistributionUserListFilter) ([]service.DistributionCommission, int64, error) {
	like := "%" + filter.Search + "%"
	const from = ` FROM distribution_commission_entries e JOIN distribution_commission_sources s ON s.id=e.source_id JOIN distribution_agents a ON a.id=e.beneficiary_agent_id JOIN users u ON u.id=s.customer_user_id WHERE a.user_id=$1 AND ($2='' OR e.status=$2) AND ($3='' OR e.entry_type=$3) AND ($4='' OR u.email ILIKE $5 OR COALESCE(u.username,'') ILIKE $5)`
	var total int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+from, userID, filter.Status, filter.EntryType, filter.Search, like).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT e.id,s.customer_user_id,e.entry_type,e.rate_bps,e.original_amount_cny,e.reversed_amount_cny,e.status,e.created_at,e.available_at,COALESCE(s.commission_base_cny,0),COALESCE(u.email,''),COALESCE(u.username,''),s.payment_type`+from+` ORDER BY e.created_at DESC LIMIT $6 OFFSET $7`, userID, filter.Status, filter.EntryType, filter.Search, like, filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionCommission, 0)
	for rows.Next() {
		var x service.DistributionCommission
		if err = rows.Scan(&x.ID, &x.CustomerUserID, &x.EntryType, &x.RateBPS, &x.OriginalAmountCNY, &x.ReversedAmountCNY, &x.Status, &x.CreatedAt, &x.AvailableAt, &x.CommissionBaseCNY, &x.CustomerEmail, &x.CustomerUsername, &x.PaymentType); err != nil {
			return nil, 0, err
		}
		x.CustomerEmail = maskDistributionEmail(x.CustomerEmail)
		items = append(items, x)
	}
	return items, total, rows.Err()
}
func (r *distributionRepository) ListTeam(ctx context.Context, userID int64, filter service.DistributionUserListFilter) ([]service.DistributionAgent, int64, error) {
	like := "%" + filter.Search + "%"
	const from = ` FROM distribution_agents a JOIN distribution_agent_levels l ON l.id=a.level_id LEFT JOIN distribution_wallets w ON w.agent_id=a.id JOIN distribution_agents parent ON parent.id=a.parent_agent_id JOIN users u ON u.id=a.user_id WHERE parent.user_id=$1 AND ($2='' OR a.status=$2) AND ($3='' OR u.email ILIKE $4 OR COALESCE(u.username,'') ILIKE $4 OR a.promotion_code ILIKE $4)`
	var total int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+from, userID, filter.Status, filter.Search, like).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	scope := newDistributionAnalyticsScope(service.DistributionAnalyticsFilter{Days: filter.Days, DateFrom: filter.DateFrom, DateTo: filter.DateTo})
	rows, err := r.db.QueryContext(ctx, `SELECT a.id,a.user_id,a.level_id,l.depth,a.parent_agent_id,a.promotion_code,COALESCE(a.rate_override_bps,l.default_rate_bps),l.max_child_rate_bps,a.can_recruit_subagents,a.can_view_promotion_stats,a.status,COALESCE(w.available_cny,0),COALESCE(w.frozen_cny,0),COALESCE(w.reserved_cny,0),COALESCE(w.debt_cny,0),COALESCE(w.total_earned_cny,0),COALESCE(w.total_withdrawn_cny,0),COALESCE(u.email,''),COALESCE(u.username,''),u.status,
		(SELECT COUNT(*) FROM distribution_customer_bindings b WHERE b.agent_id=a.id),
		(SELECT COUNT(DISTINCT source.customer_user_id) FROM distribution_commission_sources source WHERE source.direct_agent_id=a.id AND source.status<>'void'),
		COALESCE((SELECT SUM(GREATEST(COALESCE(source.commission_base_cny,0)-source.refunded_amount_cny,0)) FROM distribution_commission_sources source WHERE source.direct_agent_id=a.id AND source.status<>'void'),0),
		COALESCE((SELECT SUM(GREATEST(entry.original_amount_cny-entry.reversed_amount_cny,0)) FROM distribution_commission_entries entry WHERE entry.beneficiary_agent_id=a.id AND entry.created_at>=DATE_TRUNC('month',NOW())),0),
		(SELECT COUNT(*) FROM distribution_customer_bindings b WHERE b.agent_id=a.id AND b.bound_at>=$5 AND b.bound_at<$6),
		(SELECT COUNT(DISTINCT source.customer_user_id) FROM distribution_commission_sources source JOIN distribution_customer_bindings b ON b.user_id=source.customer_user_id WHERE source.direct_agent_id=a.id AND source.status<>'void' AND b.bound_at>=$5 AND b.bound_at<$6 AND source.paid_at>=$5 AND source.paid_at<$6),
		COALESCE((SELECT SUM(GREATEST(COALESCE(source.commission_base_cny,0)-source.refunded_amount_cny,0)) FROM distribution_commission_sources source WHERE source.direct_agent_id=a.id AND source.status<>'void' AND source.paid_at>=$5 AND source.paid_at<$6),0),
		COALESCE((SELECT SUM(GREATEST(entry.original_amount_cny-entry.reversed_amount_cny,0)) FROM distribution_commission_entries entry WHERE entry.beneficiary_agent_id=a.id AND entry.entry_type='direct' AND entry.created_at>=$5 AND entry.created_at<$6),0),a.created_at`+from+` ORDER BY a.created_at DESC LIMIT $7 OFFSET $8`, userID, filter.Status, filter.Search, like, scope.currentStart, scope.currentEnd, filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.DistributionAgent, 0)
	for rows.Next() {
		var a service.DistributionAgent
		if err = rows.Scan(&a.ID, &a.UserID, &a.LevelID, &a.Depth, &a.ParentAgentID, &a.PromotionCode, &a.EffectiveRateBPS, &a.MaxChildRateBPS, &a.CanRecruitSubagents, &a.CanViewPromotionStats, &a.Status, &a.AvailableCNY, &a.FrozenCNY, &a.ReservedCNY, &a.DebtCNY, &a.TotalEarnedCNY, &a.TotalWithdrawnCNY, &a.Email, &a.Username, &a.UserStatus, &a.CustomerCount, &a.PayingCustomerCount, &a.CustomerPaidCNY, &a.ThisMonthCommissionCNY, &a.PeriodCustomerCount, &a.PeriodPayingCustomers, &a.PeriodCustomerPaidCNY, &a.PeriodCommissionCNY, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, rows.Err()
}

func (r *distributionRepository) GetTeamAgentAnalytics(ctx context.Context, userID, agentID int64, filter service.DistributionAnalyticsFilter) (*service.DistributionAgentAnalytics, error) {
	var childUserID int64
	err := r.db.QueryRowContext(ctx, `SELECT child.user_id FROM distribution_agents child JOIN distribution_agents parent ON parent.id=child.parent_agent_id WHERE child.id=$2 AND parent.user_id=$1 AND child.status<>'revoked'`, userID, agentID).Scan(&childUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrDistributionNotAgent
	}
	if err != nil {
		return nil, err
	}
	agent, err := r.GetAgentByUserID(ctx, childUserID)
	if err != nil {
		return nil, err
	}
	analytics, err := r.distributionBusinessAnalytics(ctx, agentID, filter)
	if err != nil {
		return nil, err
	}
	return &service.DistributionAgentAnalytics{Agent: agent, Analytics: analytics, Ranking: make([]service.DistributionAgentRanking, 0)}, nil
}
func (r *distributionRepository) UpdateTeamAgentStatus(ctx context.Context, userID, agentID int64, status string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var current string
	err = tx.QueryRowContext(ctx, `SELECT child.status FROM distribution_agents child JOIN distribution_agents parent ON parent.id=child.parent_agent_id WHERE child.id=$2 AND parent.user_id=$1 AND child.status<>'revoked' FOR UPDATE OF child`, userID, agentID).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrDistributionNotAgent
	}
	if err != nil {
		return err
	}
	if current == status {
		return infraBadRequest("AGENT_STATUS_UNCHANGED", "agent already has requested status")
	}
	reason := "updated by parent agent"
	if _, err = tx.ExecContext(ctx, `UPDATE distribution_agents SET status=$2,status_reason=$3,updated_at=NOW() WHERE id=$1`, agentID, status, reason); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_agent_events(agent_id,event_type,old_status,new_status,reason,actor_user_id) VALUES($1,'status_changed',$2,$3,$4,$5)`, agentID, current, status, reason, userID); err != nil {
		return err
	}
	return tx.Commit()
}
func maskDistributionEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || len(parts[0]) == 0 {
		return "***"
	}
	return parts[0][:1] + "***@" + parts[1]
}

type commissionAgentRow struct {
	directID, parentID            int64
	depth, directRate, parentRate int
}

func (r *distributionRepository) AccruePaidOrder(ctx context.Context, in service.DistributionCommissionInput) (*service.DistributionCommissionResult, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var enabled bool
	var freezeHours int
	if err = tx.QueryRowContext(ctx, `SELECT enabled, freeze_hours FROM distribution_settings WHERE id=1`).Scan(&enabled, &freezeHours); err != nil {
		return nil, err
	}
	if !enabled {
		return nil, nil
	}

	var ar commissionAgentRow
	err = tx.QueryRowContext(ctx, `
		SELECT a.id, l.depth, COALESCE(a.rate_override_bps,l.default_rate_bps),
		       COALESCE(p.id,0), COALESCE(p.rate_override_bps,pl.default_rate_bps,0)
		FROM distribution_customer_bindings b
		JOIN distribution_agents a ON a.id=b.agent_id AND a.status='active'
		JOIN distribution_agent_levels l ON l.id=a.level_id AND l.is_active=TRUE
		LEFT JOIN distribution_agents p ON p.id=a.parent_agent_id AND p.status='active'
		LEFT JOIN distribution_agent_levels pl ON pl.id=p.level_id AND pl.is_active=TRUE
		WHERE b.user_id=$1 AND NOT EXISTS (SELECT 1 FROM distribution_agents self WHERE self.user_id=b.user_id)
		  AND (l.depth=1 OR p.id IS NOT NULL)`, in.CustomerUserID).Scan(&ar.directID, &ar.depth, &ar.directRate, &ar.parentID, &ar.parentRate)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	fx := decimal.Zero
	if err = tx.QueryRowContext(ctx, `SELECT rate_to_cny FROM distribution_fx_rates WHERE currency=$1`, in.PaymentCurrency).Scan(&fx); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	status := "settled"
	var base any
	baseCNY := decimal.Zero
	if fx.IsZero() {
		status = "pending_fx"
		base = nil
	} else {
		baseCNY = in.ActualPaid.Mul(fx).Round(8)
		base = baseCNY
	}
	snapshot := in.ProviderSnapshot
	if len(snapshot) == 0 || !json.Valid(snapshot) {
		snapshot = json.RawMessage(`{}`)
	}

	var sourceID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO distribution_commission_sources
		(payment_order_id,customer_user_id,direct_agent_id,payment_type,payment_currency,actual_paid_amount,fx_rate_to_cny,commission_base_cny,status,provider_snapshot,paid_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (payment_order_id) DO NOTHING RETURNING id`, in.PaymentOrderID, in.CustomerUserID, ar.directID,
		in.PaymentType, in.PaymentCurrency, in.ActualPaid, nullableDecimal(fx), base, status, snapshot, in.PaidAt).Scan(&sourceID)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `SELECT id,status,COALESCE(commission_base_cny,0) FROM distribution_commission_sources WHERE payment_order_id=$1 FOR UPDATE`, in.PaymentOrderID).Scan(&sourceID, &status, &baseCNY)
		if err != nil {
			return nil, err
		}
		if status != "pending_fx" || fx.IsZero() {
			_ = tx.Commit()
			return &service.DistributionCommissionResult{SourceID: sourceID, Status: status, BaseCNY: baseCNY, Created: false}, nil
		}
		baseCNY = in.ActualPaid.Mul(fx).Round(8)
		status = "settled"
		if _, err = tx.ExecContext(ctx, `UPDATE distribution_commission_sources SET fx_rate_to_cny=$2,commission_base_cny=$3,status='settled',updated_at=NOW() WHERE id=$1 AND status='pending_fx'`, sourceID, fx, baseCNY); err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	if status == "pending_fx" {
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return &service.DistributionCommissionResult{SourceID: sourceID, Status: status, Created: true}, nil
	}
	if err = accrueDistributionRewardTx(ctx, tx, in.CustomerUserID, "recharge_threshold", in.PaymentOrderID, baseCNY); err != nil {
		return nil, err
	}

	availableAt := in.PaidAt.Add(time.Duration(freezeHours) * time.Hour)
	total := decimal.Zero
	entries := []struct {
		agentID int64
		kind    string
		rate    int
	}{{ar.directID, "direct", ar.directRate}}
	if ar.depth == 2 && ar.parentID > 0 && ar.parentRate > ar.directRate {
		entries = append(entries, struct {
			agentID int64
			kind    string
			rate    int
		}{ar.parentID, "team", ar.parentRate - ar.directRate})
	}
	for _, e := range entries {
		amount := baseCNY.Mul(decimal.NewFromInt(int64(e.rate))).Div(decimal.NewFromInt(10000)).Round(8)
		if !amount.IsPositive() {
			continue
		}
		var entryID int64
		if err = tx.QueryRowContext(ctx, `INSERT INTO distribution_commission_entries
			(source_id,beneficiary_agent_id,entry_type,rate_bps,original_amount_cny,available_at)
			VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, sourceID, e.agentID, e.kind, e.rate, amount, availableAt).Scan(&entryID); err != nil {
			return nil, err
		}
		if err = creditFrozenWallet(ctx, tx, e.agentID, entryID, amount); err != nil {
			return nil, err
		}
		total = total.Add(amount)
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &service.DistributionCommissionResult{SourceID: sourceID, Status: status, BaseCNY: baseCNY, Commission: total, Created: true}, nil
}

func nullableDecimal(v decimal.Decimal) any {
	if v.IsZero() {
		return nil
	}
	return v
}

func accrueDistributionRewardTx(ctx context.Context, tx *sql.Tx, customerUserID int64, rewardType string, paymentOrderID int64, paidCNY decimal.Decimal) error {
	var enabled, rewardEnabled bool
	var freezeHours int
	err := tx.QueryRowContext(ctx, `SELECT enabled,freeze_hours,
		CASE $1 WHEN 'registration' THEN registration_reward_enabled ELSE recharge_reward_enabled END
		FROM distribution_settings WHERE id=1`, rewardType).Scan(&enabled, &freezeHours, &rewardEnabled)
	if err != nil || !enabled || !rewardEnabled {
		return err
	}

	var directID, rootID int64
	var depth int
	err = tx.QueryRowContext(ctx, `SELECT direct.id,l.depth,COALESCE(parent.id,direct.id)
		FROM distribution_customer_bindings binding
		JOIN distribution_agents direct ON direct.id=binding.agent_id AND direct.status='active'
		JOIN distribution_agent_levels l ON l.id=direct.level_id AND l.is_active=TRUE
		LEFT JOIN distribution_agents parent ON parent.id=direct.parent_agent_id AND parent.status='active'
		WHERE binding.user_id=$1 AND NOT EXISTS(SELECT 1 FROM distribution_agents self WHERE self.user_id=binding.user_id)
		AND (l.depth=1 OR parent.id IS NOT NULL)`, customerUserID).Scan(&directID, &depth, &rootID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	var rootRegistration, threshold, rootRecharge decimal.Decimal
	var rootRegistrationEnabled, rootRechargeEnabled bool
	err = tx.QueryRowContext(ctx, `SELECT registration_enabled,registration_reward_cny,recharge_enabled,recharge_threshold_cny,recharge_reward_cny
		FROM distribution_agent_reward_rules WHERE agent_id=$1`, rootID).Scan(&rootRegistrationEnabled, &rootRegistration, &rootRechargeEnabled, &threshold, &rootRecharge)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if rewardType == "registration" && !rootRegistrationEnabled || rewardType == "recharge_threshold" && !rootRechargeEnabled {
		return nil
	}
	total := rootRegistration
	if rewardType == "recharge_threshold" {
		total = rootRecharge
		if !threshold.IsPositive() || paidCNY.LessThan(threshold) {
			return nil
		}
	}
	if !total.IsPositive() {
		return nil
	}

	directShare := decimal.Zero
	if depth == 2 {
		var shareRegistration, shareRecharge decimal.Decimal
		var shareRegistrationEnabled, shareRechargeEnabled bool
		err = tx.QueryRowContext(ctx, `SELECT registration_enabled,registration_reward_cny,recharge_enabled,recharge_reward_cny
			FROM distribution_agent_reward_rules WHERE agent_id=$1`, directID).Scan(&shareRegistrationEnabled, &shareRegistration, &shareRechargeEnabled, &shareRecharge)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if err == nil {
			directShare = shareRegistration
			if !shareRegistrationEnabled {
				directShare = decimal.Zero
			}
			if rewardType == "recharge_threshold" {
				directShare = shareRecharge
				if !shareRechargeEnabled {
					directShare = decimal.Zero
				}
			}
		}
		if directShare.GreaterThan(total) {
			directShare = total
		}
	}

	snapshot, _ := json.Marshal(map[string]any{"root_agent_id": rootID, "direct_agent_id": directID, "threshold_cny": threshold.String(), "total_reward_cny": total.String(), "direct_share_cny": directShare.String()})
	var eventID int64
	var trigger any
	if paymentOrderID > 0 {
		trigger = paymentOrderID
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO distribution_reward_events
		(customer_user_id,root_agent_id,direct_agent_id,reward_type,trigger_payment_order_id,threshold_cny,total_reward_cny,direct_share_cny,config_snapshot)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(customer_user_id,reward_type) DO NOTHING RETURNING id`,
		customerUserID, rootID, directID, rewardType, trigger, threshold, total, directShare, snapshot).Scan(&eventID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	availableAt := time.Now().Add(time.Duration(freezeHours) * time.Hour)
	grants := []struct {
		agentID int64
		role    string
		amount  decimal.Decimal
	}{{rootID, "root", total.Sub(directShare)}}
	if depth == 2 && directShare.IsPositive() {
		grants = append(grants, struct {
			agentID int64
			role    string
			amount  decimal.Decimal
		}{directID, "direct", directShare})
	}
	for _, grant := range grants {
		if !grant.amount.IsPositive() {
			continue
		}
		var grantID int64
		if err = tx.QueryRowContext(ctx, `INSERT INTO distribution_reward_grants
			(event_id,beneficiary_agent_id,beneficiary_role,original_amount_cny,available_at)
			VALUES($1,$2,$3,$4,$5) RETURNING id`, eventID, grant.agentID, grant.role, grant.amount, availableAt).Scan(&grantID); err != nil {
			return err
		}
		if err = creditFrozenRewardWallet(ctx, tx, grant.agentID, grantID, grant.amount); err != nil {
			return err
		}
	}
	return nil
}

func creditFrozenRewardWallet(ctx context.Context, tx *sql.Tx, agentID, grantID int64, amount decimal.Decimal) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO distribution_wallets(agent_id,frozen_cny,total_earned_cny) VALUES($1,$2,$2)
		ON CONFLICT(agent_id) DO UPDATE SET frozen_cny=distribution_wallets.frozen_cny+EXCLUDED.frozen_cny,
		total_earned_cny=distribution_wallets.total_earned_cny+EXCLUDED.total_earned_cny,updated_at=NOW()`, agentID, amount)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO distribution_wallet_ledger
		(agent_id,reward_grant_id,entry_type,amount_cny,frozen_after_cny,available_after_cny,reserved_after_cny,debt_after_cny,idempotency_key)
		SELECT agent_id,$2,'reward_frozen',$3,frozen_cny,available_cny,reserved_cny,debt_cny,$4 FROM distribution_wallets WHERE agent_id=$1`,
		agentID, grantID, amount, fmt.Sprintf("reward:%d:frozen", grantID))
	return err
}

func creditFrozenWallet(ctx context.Context, tx *sql.Tx, agentID, entryID int64, amount decimal.Decimal) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO distribution_wallets(agent_id,frozen_cny,total_earned_cny) VALUES($1,$2,$2)
		ON CONFLICT(agent_id) DO UPDATE SET frozen_cny=distribution_wallets.frozen_cny+EXCLUDED.frozen_cny,
		total_earned_cny=distribution_wallets.total_earned_cny+EXCLUDED.total_earned_cny,updated_at=NOW()`, agentID, amount)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO distribution_wallet_ledger
		(agent_id,commission_entry_id,entry_type,amount_cny,frozen_after_cny,available_after_cny,reserved_after_cny,debt_after_cny,idempotency_key)
		SELECT agent_id,$2,'commission_frozen',$3,frozen_cny,available_cny,reserved_cny,debt_cny,$4 FROM distribution_wallets WHERE agent_id=$1`,
		agentID, entryID, amount, fmt.Sprintf("commission:%d:frozen", entryID))
	return err
}

func (r *distributionRepository) releaseMatured(ctx context.Context, userID int64) error {
	_, err := r.releaseMaturedForUser(ctx, userID)
	return err
}

func (r *distributionRepository) releaseAllMatured(ctx context.Context) error {
	_, err := r.releaseMaturedForUser(ctx, 0)
	return err
}

func (r *distributionRepository) ReleaseMaturedCommissions(ctx context.Context) (int, error) {
	return r.releaseMaturedForUser(ctx, 0)
}

func (r *distributionRepository) releaseMaturedForUser(ctx context.Context, userID int64) (int, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(ctx, `SELECT e.id,e.beneficiary_agent_id,(e.original_amount_cny-e.reversed_amount_cny)
		FROM distribution_commission_entries e JOIN distribution_agents a ON a.id=e.beneficiary_agent_id
		WHERE ($1=0 OR a.user_id=$1) AND e.status='frozen' AND e.available_at<=NOW() ORDER BY e.id FOR UPDATE OF e`, userID)
	if err != nil {
		return 0, err
	}
	type matured struct {
		id, agent int64
		amount    decimal.Decimal
	}
	var items []matured
	for rows.Next() {
		var m matured
		if err = rows.Scan(&m.id, &m.agent, &m.amount); err != nil {
			_ = rows.Close()
			return 0, err
		}
		items = append(items, m)
	}
	if err = rows.Close(); err != nil {
		return 0, err
	}
	for _, m := range items {
		var frozen, available, reserved, debt decimal.Decimal
		if err = tx.QueryRowContext(ctx, `SELECT frozen_cny,available_cny,reserved_cny,debt_cny FROM distribution_wallets WHERE agent_id=$1 FOR UPDATE`, m.agent).Scan(&frozen, &available, &reserved, &debt); err != nil {
			return 0, err
		}
		offset := decimal.Min(m.amount, debt)
		net := m.amount.Sub(offset)
		_, err = tx.ExecContext(ctx, `UPDATE distribution_wallets SET frozen_cny=frozen_cny-$2,available_cny=available_cny+$3,debt_cny=debt_cny-$4,updated_at=NOW() WHERE agent_id=$1`, m.agent, m.amount, net, offset)
		if err != nil {
			return 0, err
		}
		_, err = tx.ExecContext(ctx, `UPDATE distribution_commission_entries SET status='available',updated_at=NOW() WHERE id=$1`, m.id)
		if err != nil {
			return 0, err
		}
		kind := "commission_released"
		if offset.Equal(m.amount) {
			kind = "debt_offset"
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO distribution_wallet_ledger(agent_id,commission_entry_id,entry_type,amount_cny,frozen_after_cny,available_after_cny,reserved_after_cny,debt_after_cny,idempotency_key)
			SELECT agent_id,$2,$3,$4,frozen_cny,available_cny,reserved_cny,debt_cny,$5 FROM distribution_wallets WHERE agent_id=$1`, m.agent, m.id, kind, net, fmt.Sprintf("commission:%d:release", m.id))
		if err != nil {
			return 0, err
		}
	}
	rewardCount, err := releaseMaturedRewardsTx(ctx, tx, userID)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return len(items) + rewardCount, nil
}

func releaseMaturedRewardsTx(ctx context.Context, tx *sql.Tx, userID int64) (int, error) {
	rows, err := tx.QueryContext(ctx, `SELECT g.id,g.beneficiary_agent_id,(g.original_amount_cny-g.reversed_amount_cny)
		FROM distribution_reward_grants g JOIN distribution_agents a ON a.id=g.beneficiary_agent_id
		WHERE ($1=0 OR a.user_id=$1) AND g.status='frozen' AND g.available_at<=NOW() ORDER BY g.id FOR UPDATE OF g`, userID)
	if err != nil {
		return 0, err
	}
	type maturedReward struct {
		id, agent int64
		amount    decimal.Decimal
	}
	var items []maturedReward
	for rows.Next() {
		var m maturedReward
		if err = rows.Scan(&m.id, &m.agent, &m.amount); err != nil {
			_ = rows.Close()
			return 0, err
		}
		items = append(items, m)
	}
	if err = rows.Close(); err != nil {
		return 0, err
	}
	for _, m := range items {
		var frozen, available, reserved, debt decimal.Decimal
		if err = tx.QueryRowContext(ctx, `SELECT frozen_cny,available_cny,reserved_cny,debt_cny FROM distribution_wallets WHERE agent_id=$1 FOR UPDATE`, m.agent).Scan(&frozen, &available, &reserved, &debt); err != nil {
			return 0, err
		}
		offset := decimal.Min(m.amount, debt)
		net := m.amount.Sub(offset)
		if _, err = tx.ExecContext(ctx, `UPDATE distribution_wallets SET frozen_cny=frozen_cny-$2,available_cny=available_cny+$3,debt_cny=debt_cny-$4,updated_at=NOW() WHERE agent_id=$1`, m.agent, m.amount, net, offset); err != nil {
			return 0, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE distribution_reward_grants SET status='available',updated_at=NOW() WHERE id=$1`, m.id); err != nil {
			return 0, err
		}
		kind := "reward_released"
		if offset.Equal(m.amount) {
			kind = "debt_offset"
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_wallet_ledger(agent_id,reward_grant_id,entry_type,amount_cny,frozen_after_cny,available_after_cny,reserved_after_cny,debt_after_cny,idempotency_key)
			SELECT agent_id,$2,$3,$4,frozen_cny,available_cny,reserved_cny,debt_cny,$5 FROM distribution_wallets WHERE agent_id=$1`, m.agent, m.id, kind, net, fmt.Sprintf("reward:%d:release", m.id)); err != nil {
			return 0, err
		}
	}
	return len(items), nil
}

func (r *distributionRepository) ReverseRefund(ctx context.Context, orderID int64, refundedActual decimal.Decimal) error {
	if contextTx := dbent.TxFromContext(ctx); contextTx != nil {
		return reverseDistributionRefundTx(ctx, contextTx.Client(), orderID, refundedActual)
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err = reverseDistributionRefundTx(ctx, tx, orderID, refundedActual); err != nil {
		return err
	}
	return tx.Commit()
}

type distributionRefundExecutor interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func reverseDistributionRefundTx(ctx context.Context, tx distributionRefundExecutor, orderID int64, refundedActual decimal.Decimal) error {
	var sourceID int64
	var actual, fx, base, already decimal.Decimal
	var status string
	err := scanSingleRow(ctx, tx, `SELECT id,actual_paid_amount,COALESCE(fx_rate_to_cny,0),COALESCE(commission_base_cny,0),refunded_amount_cny,status
		FROM distribution_commission_sources WHERE payment_order_id=$1 FOR UPDATE`, []any{orderID}, &sourceID, &actual, &fx, &base, &already, &status)
	if errors.Is(err, sql.ErrNoRows) || status == "pending_fx" || status == "void" {
		return nil
	}
	if err != nil {
		return err
	}
	refundCNY := refundedActual.Mul(fx).Round(8)
	remainingBase := base.Sub(already)
	if refundCNY.GreaterThan(remainingBase) {
		refundCNY = remainingBase
	}
	if !refundCNY.IsPositive() {
		return nil
	}
	ratio := refundCNY.Div(base)
	rows, err := tx.QueryContext(ctx, `SELECT id,beneficiary_agent_id,original_amount_cny,reversed_amount_cny,status FROM distribution_commission_entries WHERE source_id=$1 FOR UPDATE`, sourceID)
	if err != nil {
		return err
	}
	type entry struct {
		id, agent          int64
		original, reversed decimal.Decimal
		status             string
	}
	var entries []entry
	for rows.Next() {
		var e entry
		if err = rows.Scan(&e.id, &e.agent, &e.original, &e.reversed, &e.status); err != nil {
			_ = rows.Close()
			return err
		}
		entries = append(entries, e)
	}
	_ = rows.Close()
	for _, e := range entries {
		reversal := e.original.Mul(ratio).Round(8)
		cap := e.original.Sub(e.reversed)
		if reversal.GreaterThan(cap) {
			reversal = cap
		}
		if !reversal.IsPositive() {
			continue
		}
		var frozen, available, reserved, debt decimal.Decimal
		if err = scanSingleRow(ctx, tx, `SELECT frozen_cny,available_cny,reserved_cny,debt_cny FROM distribution_wallets WHERE agent_id=$1 FOR UPDATE`, []any{e.agent}, &frozen, &available, &reserved, &debt); err != nil {
			return err
		}
		fromFrozen := decimal.Min(frozen, reversal)
		left := reversal.Sub(fromFrozen)
		fromAvailable := decimal.Min(available, left)
		debtAdd := left.Sub(fromAvailable)
		_, err = tx.ExecContext(ctx, `UPDATE distribution_wallets SET frozen_cny=frozen_cny-$2,available_cny=available_cny-$3,debt_cny=debt_cny+$4,updated_at=NOW() WHERE agent_id=$1`, e.agent, fromFrozen, fromAvailable, debtAdd)
		if err != nil {
			return err
		}
		newReversed := e.reversed.Add(reversal)
		newStatus := e.status
		if newReversed.Equal(e.original) {
			newStatus = "reversed"
		}
		_, err = tx.ExecContext(ctx, `UPDATE distribution_commission_entries SET reversed_amount_cny=$2,status=$3,updated_at=NOW() WHERE id=$1`, e.id, newReversed, newStatus)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("commission:%d:refund:%s", e.id, strings.ReplaceAll(newReversed.String(), ".", "_"))
		_, err = tx.ExecContext(ctx, `INSERT INTO distribution_wallet_ledger(agent_id,commission_entry_id,entry_type,amount_cny,frozen_after_cny,available_after_cny,reserved_after_cny,debt_after_cny,idempotency_key)
			SELECT agent_id,$2,'refund_reversal',$3,frozen_cny,available_cny,reserved_cny,debt_cny,$4 FROM distribution_wallets WHERE agent_id=$1`, e.agent, e.id, reversal.Neg(), key)
		if err != nil {
			return err
		}
	}
	if err = reverseDistributionRewardEventsTx(ctx, tx, orderID); err != nil {
		return err
	}
	newRefunded := already.Add(refundCNY)
	newStatus := "partially_refunded"
	if newRefunded.GreaterThanOrEqual(base) {
		newStatus = "fully_refunded"
	}
	_, err = tx.ExecContext(ctx, `UPDATE distribution_commission_sources SET refunded_amount_cny=$2,status=$3,updated_at=NOW() WHERE id=$1`, sourceID, newRefunded, newStatus)
	if err != nil {
		return err
	}
	return nil
}

func reverseDistributionRewardEventsTx(ctx context.Context, tx distributionRefundExecutor, orderID int64) error {
	rows, err := tx.QueryContext(ctx, `SELECT g.id,g.beneficiary_agent_id,g.original_amount_cny,g.reversed_amount_cny,g.status
		FROM distribution_reward_events event JOIN distribution_reward_grants g ON g.event_id=event.id
		WHERE event.trigger_payment_order_id=$1 FOR UPDATE OF g`, orderID)
	if err != nil {
		return err
	}
	type rewardGrant struct {
		id, agent          int64
		original, reversed decimal.Decimal
		status             string
	}
	var grants []rewardGrant
	for rows.Next() {
		var g rewardGrant
		if err = rows.Scan(&g.id, &g.agent, &g.original, &g.reversed, &g.status); err != nil {
			_ = rows.Close()
			return err
		}
		grants = append(grants, g)
	}
	if err = rows.Close(); err != nil {
		return err
	}
	for _, g := range grants {
		reversal := g.original.Sub(g.reversed)
		if !reversal.IsPositive() || g.status == "reversed" {
			continue
		}
		var frozen, available, reserved, debt decimal.Decimal
		if err = scanSingleRow(ctx, tx, `SELECT frozen_cny,available_cny,reserved_cny,debt_cny FROM distribution_wallets WHERE agent_id=$1 FOR UPDATE`, []any{g.agent}, &frozen, &available, &reserved, &debt); err != nil {
			return err
		}
		fromFrozen := decimal.Min(frozen, reversal)
		left := reversal.Sub(fromFrozen)
		fromAvailable := decimal.Min(available, left)
		debtAdd := left.Sub(fromAvailable)
		if _, err = tx.ExecContext(ctx, `UPDATE distribution_wallets SET frozen_cny=frozen_cny-$2,available_cny=available_cny-$3,debt_cny=debt_cny+$4,updated_at=NOW() WHERE agent_id=$1`, g.agent, fromFrozen, fromAvailable, debtAdd); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE distribution_reward_grants SET reversed_amount_cny=original_amount_cny,status='reversed',updated_at=NOW() WHERE id=$1`, g.id); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO distribution_wallet_ledger(agent_id,reward_grant_id,entry_type,amount_cny,frozen_after_cny,available_after_cny,reserved_after_cny,debt_after_cny,idempotency_key)
			SELECT agent_id,$2,'reward_reversal',$3,frozen_cny,available_cny,reserved_cny,debt_cny,$4 FROM distribution_wallets WHERE agent_id=$1`, g.agent, g.id, reversal.Neg(), fmt.Sprintf("reward:%d:reversal", g.id)); err != nil {
			return err
		}
	}
	return nil
}
