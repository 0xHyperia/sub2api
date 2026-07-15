package service

import (
	"context"
	"time"
)

const (
	ModelMonitorDefaultIntervalSeconds = 300
	ModelMonitorMinIntervalSeconds     = 15
	ModelMonitorMaxIntervalSeconds     = 3600
	ModelMonitorHistoryRetentionDays   = 30
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
	ID        int64     `json:"id"`
	MonitorID int64     `json:"monitor_id"`
	Status    string    `json:"status"`
	LatencyMs *int      `json:"latency_ms"`
	Attempts  int       `json:"attempts"`
	Message   string    `json:"message"`
	GroupID   *int64    `json:"group_id"`
	GroupName string    `json:"group_name"`
	CheckedAt time.Time `json:"checked_at"`
}

type ModelMonitorSummary struct {
	Status         string                      `json:"status"`
	LatencyMs      *int                        `json:"latency_ms"`
	Availability7d *float64                    `json:"availability_7d"`
	LastCheckedAt  *time.Time                  `json:"last_checked_at"`
	Timeline       []ModelMonitorTimelinePoint `json:"timeline"`
	DisplayOrder   int                         `json:"display_order"`
	Label          string                      `json:"label"`
}

type ModelMonitorTimelinePoint struct {
	Status    string    `json:"status"`
	LatencyMs *int      `json:"latency_ms"`
	CheckedAt time.Time `json:"checked_at"`
}

type ModelCatalogEntry struct {
	Platform string                    `json:"platform"`
	Model    string                    `json:"model"`
	Groups   []ModelMonitorGroupOption `json:"groups,omitempty"`
}

type ModelMonitorGroupOption struct {
	GroupID        int64   `json:"group_id"`
	Name           string  `json:"name"`
	RateMultiplier float64 `json:"rate_multiplier"`
	Priority       int     `json:"priority"`
	Selected       bool    `json:"selected"`
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
	GetByID(ctx context.Context, id int64) (*ModelMonitor, error)
	GetByKey(ctx context.Context, platform, model string) (*ModelMonitor, error)
	Upsert(ctx context.Context, monitor *ModelMonitor) error
	ListGroupConfigs(ctx context.Context) (map[int64][]int64, error)
	ReplaceGroups(ctx context.Context, monitorID int64, groupIDs []int64) error
	UpdateLastChecked(ctx context.Context, id int64, checkedAt time.Time) error
	InsertHistory(ctx context.Context, history *ModelMonitorHistory) error
	ListHistory(ctx context.Context, monitorID int64, limit int) ([]ModelMonitorHistory, error)
	Summaries(ctx context.Context, keys []ModelCatalogEntry, timelineLimit int) (map[string]ModelMonitorSummary, error)
	DeleteHistoryBefore(ctx context.Context, cutoff time.Time) (int64, error)
}

func ModelMonitorKey(platform, model string) string { return platform + "\x00" + model }
