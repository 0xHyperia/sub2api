package service

import (
	"context"
	"time"
)

const (
	ModelMonitorDefaultIntervalSeconds = 300
	ModelMonitorMinIntervalSeconds     = 60
	ModelMonitorMaxIntervalSeconds     = 3600
	ModelMonitorHistoryRetentionDays   = 30
	ModelMonitorMinuteRetentionHours   = 48
	ModelMonitorMetricBucketCount      = 30
	ModelMonitorTimelinePoints         = 60
	modelMonitorDegradedThreshold      = 6 * time.Second
	modelMonitorMaxAttempts            = 3
)

type ModelMonitor struct {
	ID              int64      `json:"id"`
	Platform        string     `json:"platform"`
	Model           string     `json:"model"`
	Enabled         bool       `json:"enabled"`
	IntervalSeconds int        `json:"interval_seconds"`
	DisplayOrder    int        `json:"display_order"`
	Label           string     `json:"label"`
	LastCheckedAt   *time.Time `json:"last_checked_at"`
	CreatedBy       int64      `json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ModelMonitorHistory struct {
	ID             int64     `json:"id"`
	MonitorID      int64     `json:"monitor_id"`
	Status         string    `json:"status"`
	LatencyMs      *int      `json:"latency_ms"`
	Attempts       int       `json:"attempts"`
	Message        string    `json:"message"`
	GroupID        *int64    `json:"group_id"`
	GroupName      string    `json:"group_name"`
	FirstTokenMs   *int      `json:"first_token_ms"`
	InputTokens    int       `json:"input_tokens"`
	OutputTokens   int       `json:"output_tokens"`
	GenerationMs   int64     `json:"generation_ms"`
	ProbeCost      *float64  `json:"probe_cost"`
	ProbeCostKnown bool      `json:"-"`
	CheckedAt      time.Time `json:"checked_at"`
}

type ModelMonitorResolution string

const (
	ModelMonitorResolutionMinute ModelMonitorResolution = "minute"
	ModelMonitorResolutionHour   ModelMonitorResolution = "hour"
)

type ModelMonitorMetricBucket struct {
	StartedAt   time.Time `json:"started_at"`
	SuccessRate *float64  `json:"success_rate"`
	TTFTMs      *float64  `json:"ttft_ms,omitempty"`
}

type ModelMonitorGroupMetrics struct {
	TPS              *float64                   `json:"tps,omitempty"`
	TTFTMs           *float64                   `json:"ttft_ms,omitempty"`
	AverageLatencyMs *float64                   `json:"average_latency_ms,omitempty"`
	SuccessRate      *float64                   `json:"success_rate"`
	RequestCount     *int64                     `json:"request_count,omitempty"`
	SuccessCount     *int64                     `json:"success_count,omitempty"`
	FailureCount     *int64                     `json:"failure_count,omitempty"`
	ProbeCost        *float64                   `json:"probe_cost,omitempty"`
	Buckets          []ModelMonitorMetricBucket `json:"buckets"`
}

type ModelMonitorGroupConfig struct {
	MonitorID       int64      `json:"monitor_id"`
	GroupID         int64      `json:"group_id"`
	Priority        int        `json:"priority"`
	Enabled         bool       `json:"enabled"`
	IntervalSeconds int        `json:"interval_seconds"`
	LastTrafficAt   *time.Time `json:"last_traffic_at"`
	LastProbeAt     *time.Time `json:"last_probe_at"`
}

type ModelMonitorDueGroup struct {
	Monitor ModelMonitor
	Group   ModelMonitorGroupConfig
	Name    string
}

type ModelMonitorMetricDelta struct {
	MonitorID      int64
	GroupID        int64
	Source         string
	BucketStart    time.Time
	RequestCount   int64
	SuccessCount   int64
	LatencySumMs   int64
	LatencyCount   int64
	TTFTSumMs      int64
	TTFTCount      int64
	OutputTokens   int64
	GenerationMs   int64
	ProbeCost      float64
	ProbeCostKnown bool
}

type ModelMonitorMetricScope struct {
	MonitorID int64
	GroupIDs  []int64
}

type ModelMonitorSummary struct {
	Status         string                           `json:"status"`
	LatencyMs      *int                             `json:"latency_ms"`
	Availability7d *float64                         `json:"availability_7d"`
	LastCheckedAt  *time.Time                       `json:"last_checked_at"`
	Timeline       []ModelMonitorTimelinePoint      `json:"timeline"`
	DisplayOrder   int                              `json:"display_order"`
	Label          string                           `json:"label"`
	Groups         []ModelMonitorPublicGroupMetrics `json:"groups,omitempty"`
	Metrics        *ModelMonitorGroupMetrics        `json:"metrics,omitempty"`
	HourlyMetrics  *ModelMonitorGroupMetrics        `json:"hourly_metrics,omitempty"`
}

type ModelMonitorPublicGroupMetrics struct {
	GroupID int64                    `json:"group_id"`
	Name    string                   `json:"name"`
	Metrics ModelMonitorGroupMetrics `json:"metrics"`
}

type ModelMonitorTimelinePoint struct {
	Status    string    `json:"status"`
	LatencyMs *int      `json:"latency_ms"`
	CheckedAt time.Time `json:"checked_at"`
	GroupID   *int64    `json:"group_id,omitempty"`
	GroupName string    `json:"group_name,omitempty"`
}

type ModelCatalogEntry struct {
	Platform string                    `json:"platform"`
	Model    string                    `json:"model"`
	Groups   []ModelMonitorGroupOption `json:"groups,omitempty"`
}

type ModelMonitorGroupOption struct {
	GroupID         int64                     `json:"group_id"`
	Name            string                    `json:"name"`
	RateMultiplier  float64                   `json:"rate_multiplier"`
	Priority        int                       `json:"priority"`
	Selected        bool                      `json:"selected"`
	Enabled         bool                      `json:"enabled"`
	IntervalSeconds int                       `json:"interval_seconds"`
	LastTrafficAt   *time.Time                `json:"last_traffic_at"`
	LastProbeAt     *time.Time                `json:"last_probe_at"`
	Metrics         *ModelMonitorGroupMetrics `json:"metrics"`
}

type ModelMonitorRow struct {
	ModelMonitor
	CatalogAvailable bool                      `json:"catalog_available"`
	Configured       bool                      `json:"configured"`
	Summary          *ModelMonitorSummary      `json:"summary"`
	Groups           []ModelMonitorGroupOption `json:"groups"`
	GroupsConfigured bool                      `json:"groups_configured"`
}

type ModelMonitorRepository interface {
	List(ctx context.Context) ([]ModelMonitor, error)
	ListEnabledDue(ctx context.Context, now time.Time) ([]ModelMonitor, error)
	ListEnabledDueGroups(ctx context.Context, now time.Time, limit int) ([]ModelMonitorDueGroup, error)
	ClaimDueGroup(ctx context.Context, monitorID, groupID int64, now, claimedUntil time.Time) (bool, error)
	GetByID(ctx context.Context, id int64) (*ModelMonitor, error)
	GetByKey(ctx context.Context, platform, model string) (*ModelMonitor, error)
	Upsert(ctx context.Context, monitor *ModelMonitor) error
	ListGroupConfigs(ctx context.Context) (map[int64][]ModelMonitorGroupConfig, error)
	ReplaceGroups(ctx context.Context, monitorID int64, groupIDs []int64) error
	UpsertGroupConfig(ctx context.Context, config ModelMonitorGroupConfig) error
	UpdateGroupProbeAt(ctx context.Context, monitorID, groupID int64, checkedAt time.Time) error
	UpdateLastChecked(ctx context.Context, id int64, checkedAt time.Time) error
	InsertHistory(ctx context.Context, history *ModelMonitorHistory) error
	ListHistory(ctx context.Context, monitorID int64, limit int) ([]ModelMonitorHistory, error)
	Summaries(ctx context.Context, keys []ModelCatalogEntry, timelineLimit int) (map[string]ModelMonitorSummary, error)
	RefreshTrafficMetrics(ctx context.Context, from, to time.Time) error
	UpsertMetricDelta(ctx context.Context, delta ModelMonitorMetricDelta) error
	GroupMetrics(ctx context.Context, monitorID int64, groupIDs []int64, resolution ModelMonitorResolution, now time.Time) (map[int64]ModelMonitorGroupMetrics, error)
	GroupMetricsBatch(ctx context.Context, scopes []ModelMonitorMetricScope, resolution ModelMonitorResolution, now time.Time) (map[int64]map[int64]ModelMonitorGroupMetrics, error)
	RollupHourlyMetrics(ctx context.Context, hour time.Time) error
	DeleteMetricBucketsBefore(ctx context.Context, resolution ModelMonitorResolution, cutoff time.Time) (int64, error)
	DeleteHistoryBefore(ctx context.Context, cutoff time.Time) (int64, error)
}

func ModelMonitorKey(platform, model string) string { return platform + "\x00" + model }

// RedactModelMonitorDetailedPerformance keeps public success-rate information
// while removing throughput and latency details.
func RedactModelMonitorDetailedPerformance(summary *ModelMonitorSummary) {
	if summary == nil {
		return
	}
	summary.Metrics = successOnlyModelMonitorMetrics(summary.Metrics)
	summary.HourlyMetrics = successOnlyModelMonitorMetrics(summary.HourlyMetrics)
	groups := append([]ModelMonitorPublicGroupMetrics(nil), summary.Groups...)
	for i := range groups {
		metrics := groups[i].Metrics
		redacted := successOnlyModelMonitorMetrics(&metrics)
		groups[i].Metrics = *redacted
	}
	summary.Groups = groups
}

// RedactModelMonitorSampleCounts keeps operational sample volume private on
// public model marketplace endpoints while retaining success-rate aggregates.
func RedactModelMonitorSampleCounts(summary *ModelMonitorSummary) {
	if summary == nil {
		return
	}
	clearModelMonitorMetricCounts(summary.Metrics)
	clearModelMonitorMetricCounts(summary.HourlyMetrics)
	for i := range summary.Groups {
		clearModelMonitorMetricCounts(&summary.Groups[i].Metrics)
	}
}

func clearModelMonitorMetricCounts(metrics *ModelMonitorGroupMetrics) {
	if metrics == nil {
		return
	}
	metrics.RequestCount = nil
	metrics.SuccessCount = nil
	metrics.FailureCount = nil
}

func successOnlyModelMonitorMetrics(metrics *ModelMonitorGroupMetrics) *ModelMonitorGroupMetrics {
	if metrics == nil {
		return nil
	}
	copy := *metrics
	copy.TPS = nil
	copy.TTFTMs = nil
	copy.AverageLatencyMs = nil
	copy.ProbeCost = nil
	clearModelMonitorMetricCounts(&copy)
	copy.Buckets = append([]ModelMonitorMetricBucket(nil), metrics.Buckets...)
	for i := range copy.Buckets {
		copy.Buckets[i].TTFTMs = nil
	}
	return &copy
}
