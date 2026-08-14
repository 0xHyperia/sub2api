package service

import (
	"context"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var ErrBusinessAnalyticsAgentNotFound = infraerrors.NotFound("BUSINESS_ANALYTICS_AGENT_NOT_FOUND", "distribution agent not found")

type BusinessAnalyticsFilter struct {
	DateFrom     time.Time
	DateTo       time.Time
	ReportDateTo time.Time
	Timezone     string
	Granularity  string
	Channel      string
	AgentID      int64
	AgentScope   string
	Comparison   bool
	SummaryOnly  bool
	Section      string
}

type BusinessMetric struct {
	Value          float64  `json:"value"`
	Previous       float64  `json:"previous"`
	ChangeRate     *float64 `json:"change_rate,omitempty"`
	ChangeValue    *float64 `json:"change_value,omitempty"`
	ComparisonType string   `json:"comparison_type,omitempty"`
	Currency       string   `json:"currency,omitempty"`
	Estimated      bool     `json:"estimated,omitempty"`
}

type BusinessTrendPoint struct {
	Bucket          string  `json:"bucket"`
	NewUsers        int64   `json:"new_users"`
	ActivatedUsers  int64   `json:"activated_users"`
	FirstPaidUsers  int64   `json:"first_paid_users"`
	PayingUsers     int64   `json:"paying_users"`
	RepurchaseUsers int64   `json:"repurchase_users"`
	ActiveUsers     int64   `json:"active_users"`
	GrossPaidCNY    float64 `json:"gross_paid_cny"`
	RefundedCNY     float64 `json:"refunded_cny"`
	NetPaidCNY      float64 `json:"net_paid_cny"`
	ConsumedRevenue float64 `json:"consumed_revenue"`
	SupplierCost    float64 `json:"supplier_cost"`
}

type BusinessChannelMetric struct {
	Channel         string  `json:"channel"`
	NewUsers        int64   `json:"new_users"`
	ActivatedUsers  int64   `json:"activated_users"`
	FirstPaidUsers  int64   `json:"first_paid_users"`
	PayingUsers     int64   `json:"paying_users"`
	RepurchaseUsers int64   `json:"repurchase_users"`
	ActiveUsers     int64   `json:"active_users"`
	NetPaidCNY      float64 `json:"net_paid_cny"`
	ARPPUCNY        float64 `json:"arppu_cny"`
	D7RetentionRate float64 `json:"d7_retention_rate"`
}

type BusinessCohortRow struct {
	CohortDate string   `json:"cohort_date"`
	Users      int64    `json:"users"`
	D1         *float64 `json:"d1,omitempty"`
	D7         *float64 `json:"d7,omitempty"`
	D30        *float64 `json:"d30,omitempty"`
}

type BusinessLifecycle struct {
	New                int64 `json:"new"`
	Unactivated        int64 `json:"unactivated"`
	NewlyActivated     int64 `json:"newly_activated"`
	ContinuouslyActive int64 `json:"continuously_active"`
	SilentReactivated  int64 `json:"silent_reactivated"`
	ChurnedReactivated int64 `json:"churned_reactivated"`
	Silent             int64 `json:"silent"`
	Churned            int64 `json:"churned"`
}

type BusinessBalanceSegment struct {
	Key        string  `json:"key"`
	Users      int64   `json:"users"`
	BalanceUSD float64 `json:"balance_usd"`
	FrozenUSD  float64 `json:"frozen_usd"`
}

type BusinessBalanceSnapshot struct {
	AsOf                      time.Time                `json:"as_of"`
	ActivityWindowDays        int                      `json:"activity_window_days"`
	BalanceRechargeMultiplier float64                  `json:"balance_recharge_multiplier"`
	AvailableBalanceUSD       float64                  `json:"available_balance_usd"`
	FrozenBalanceUSD          float64                  `json:"frozen_balance_usd"`
	PositiveBalanceUsers      int64                    `json:"positive_balance_users"`
	TotalUsers                int64                    `json:"total_users"`
	AverageBalanceUSD         float64                  `json:"average_balance_usd"`
	LowBalanceEnabled         bool                     `json:"low_balance_enabled"`
	LowBalanceUsers           int64                    `json:"low_balance_users"`
	Segments                  []BusinessBalanceSegment `json:"segments"`
}

func BusinessAnalyticsInclusiveDate(end time.Time) string {
	return end.Add(-time.Nanosecond).Format("2006-01-02")
}

type BusinessFunnelStep struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

type BusinessDurationBucket struct {
	Key        string `json:"key"`
	Activation int64  `json:"activation"`
	FirstPaid  int64  `json:"first_paid"`
}

type BusinessCurrencyBreakdown struct {
	Currency string  `json:"currency"`
	Gross    float64 `json:"gross"`
	Refunded float64 `json:"refunded"`
	Net      float64 `json:"net"`
}

type BusinessAnalyticsSnapshot struct {
	DateFrom             string                      `json:"date_from"`
	DateTo               string                      `json:"date_to"`
	PreviousDateFrom     string                      `json:"previous_date_from"`
	PreviousDateTo       string                      `json:"previous_date_to"`
	Timezone             string                      `json:"timezone"`
	Granularity          string                      `json:"granularity"`
	Currency             string                      `json:"currency"`
	UpdatedAt            time.Time                   `json:"updated_at"`
	UsageDataFrom        string                      `json:"usage_data_from,omitempty"`
	Estimated            bool                        `json:"estimated"`
	Metrics              map[string]BusinessMetric   `json:"metrics"`
	Trend                []BusinessTrendPoint        `json:"trend"`
	PreviousTrend        []BusinessTrendPoint        `json:"previous_trend"`
	Funnel               []BusinessFunnelStep        `json:"funnel"`
	Channels             []BusinessChannelMetric     `json:"channels"`
	RegistrationCohorts  []BusinessCohortRow         `json:"registration_cohorts"`
	ActivationCohorts    []BusinessCohortRow         `json:"activation_cohorts"`
	Lifecycle            BusinessLifecycle           `json:"lifecycle"`
	DurationDistribution []BusinessDurationBucket    `json:"duration_distribution"`
	CurrencyBreakdown    []BusinessCurrencyBreakdown `json:"currency_breakdown"`
	Warnings             []string                    `json:"warnings"`
}

type BusinessAnalyticsRepository interface {
	GetBusinessAnalytics(ctx context.Context, filter BusinessAnalyticsFilter) (*BusinessAnalyticsSnapshot, error)
	GetBusinessBalance(ctx context.Context, activityWindowDays int, agentID int64, agentScope string) (*BusinessBalanceSnapshot, error)
	GetBusinessAgentDepth(ctx context.Context, agentID int64) (int, error)
}

// BusinessAnalyticsAggregator is consumed by the existing dashboard scheduler
// so both analytics families share leader election and batch boundaries.
type BusinessAnalyticsAggregator interface {
	AggregateRange(ctx context.Context, start, end time.Time) error
}

type BusinessAnalyticsService struct{ repo BusinessAnalyticsRepository }

func NewBusinessAnalyticsService(repo BusinessAnalyticsRepository) *BusinessAnalyticsService {
	return &BusinessAnalyticsService{repo: repo}
}

func (s *BusinessAnalyticsService) Get(ctx context.Context, filter BusinessAnalyticsFilter) (*BusinessAnalyticsSnapshot, error) {
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	filter.Timezone = "Asia/Shanghai"
	filter.Channel = strings.ToLower(strings.TrimSpace(filter.Channel))
	filter.Section = strings.ToLower(strings.TrimSpace(filter.Section))
	if err := s.normalizeAgentScope(ctx, &filter); err != nil {
		return nil, err
	}
	validChannels := map[string]bool{"": true, "distribution": true, "affiliate": true, "campaign": true, "organic": true, "unknown": true}
	if !validChannels[filter.Channel] {
		return nil, infraerrors.BadRequest("INVALID_ANALYTICS_CHANNEL", "invalid analytics channel")
	}
	validSections := map[string]bool{"": true, "overview": true, "growth": true, "finance": true, "retention": true, "channels": true}
	if !validSections[filter.Section] {
		return nil, infraerrors.BadRequest("INVALID_ANALYTICS_SECTION", "invalid analytics section")
	}
	from := filter.DateFrom.In(location)
	to := filter.DateTo.In(location)
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, location)
	to = time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, location)
	days := int(to.Sub(from).Hours()/24) + 1
	if days < 1 || days > 366 || to.After(time.Now().In(location)) {
		return nil, infraerrors.BadRequest("INVALID_ANALYTICS_RANGE", "analytics range must contain 1 to 366 days and cannot end in the future")
	}
	filter.DateFrom = from
	filter.ReportDateTo = to
	filter.DateTo = to.AddDate(0, 0, 1)
	if now := time.Now().In(location); now.Before(filter.DateTo) {
		filter.DateTo = now
	}
	switch {
	case days == 1:
		filter.Granularity = "hour"
	case days <= 31:
		filter.Granularity = "day"
	case days <= 92:
		filter.Granularity = "week"
	default:
		filter.Granularity = "month"
	}
	filter.Comparison = true
	return s.repo.GetBusinessAnalytics(ctx, filter)
}

func (s *BusinessAnalyticsService) GetBalance(ctx context.Context, activityWindowDays int, agentID int64, agentScope string) (*BusinessBalanceSnapshot, error) {
	if activityWindowDays != 1 && activityWindowDays != 7 && activityWindowDays != 30 {
		return nil, infraerrors.BadRequest("INVALID_ACTIVITY_WINDOW", "activity window must be 1, 7, or 30 days")
	}
	filter := BusinessAnalyticsFilter{AgentID: agentID, AgentScope: agentScope}
	if err := s.normalizeAgentScope(ctx, &filter); err != nil {
		return nil, err
	}
	return s.repo.GetBusinessBalance(ctx, activityWindowDays, filter.AgentID, filter.AgentScope)
}

func (s *BusinessAnalyticsService) normalizeAgentScope(ctx context.Context, filter *BusinessAnalyticsFilter) error {
	filter.AgentScope = strings.ToLower(strings.TrimSpace(filter.AgentScope))
	if filter.AgentID <= 0 {
		filter.AgentID = 0
		filter.AgentScope = ""
		return nil
	}
	// Agent attribution and acquisition-channel attribution are separate views.
	// An explicit agent scope wins so a stale channel query cannot narrow it.
	filter.Channel = ""
	if filter.AgentScope == "" {
		filter.AgentScope = "team"
	}
	if filter.AgentScope != "team" && filter.AgentScope != "direct" {
		return infraerrors.BadRequest("INVALID_AGENT_SCOPE", "agent scope must be team or direct")
	}
	depth, err := s.repo.GetBusinessAgentDepth(ctx, filter.AgentID)
	if err != nil {
		return err
	}
	if depth == 2 {
		filter.AgentScope = "direct"
	}
	return nil
}

// GetSummary returns the same KPI definitions as the complete report while
// skipping trend, channel, cohort and lifecycle queries. It is intended for
// overview surfaces that only render the metric map.
func (s *BusinessAnalyticsService) GetSummary(ctx context.Context, filter BusinessAnalyticsFilter) (*BusinessAnalyticsSnapshot, error) {
	filter.SummaryOnly = true
	return s.Get(ctx, filter)
}
