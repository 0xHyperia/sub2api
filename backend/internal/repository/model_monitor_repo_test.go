package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestModelMonitorGroupMetricsUsesWindowTotalsAndFixedBuckets(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	now := time.Date(2026, 8, 3, 12, 30, 10, 0, time.UTC)
	current := now.Truncate(time.Minute)
	rows := sqlmock.NewRows([]string{
		"monitor_id", "group_id", "bucket_start", "request_count", "success_count",
		"latency_sum_ms", "latency_count", "ttft_sum_ms", "ttft_count",
		"output_tokens", "generation_ms", "probe_cost", "probe_cost_known",
	}).
		AddRow(3, 0, current.Add(-time.Minute), 1, 1, 100, 1, 20, 1, 100, 1000, 0.01, true).
		AddRow(3, 0, current, 9, 0, 0, 0, 0, 0, 0, 0, 0.02, true).
		AddRow(3, 7, current.Add(-time.Minute), 1, 1, 100, 1, 20, 1, 100, 1000, 0.01, true).
		AddRow(3, 7, current, 9, 0, 0, 0, 0, 0, 0, 0, 0.02, true)
	mock.ExpectQuery("WITH requested AS").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), string(service.ModelMonitorResolutionMinute), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	repo := &modelMonitorRepository{db: db}
	metrics, err := repo.GroupMetrics(context.Background(), 3, []int64{7}, service.ModelMonitorResolutionMinute, now)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())

	group := metrics[7]
	require.Len(t, group.Buckets, 30)
	require.Nil(t, group.Buckets[27].SuccessRate)
	require.Equal(t, 100.0, *group.Buckets[28].SuccessRate)
	require.InDelta(t, 20.0, *group.Buckets[28].TTFTMs, 0.0001)
	require.Equal(t, 0.0, *group.Buckets[29].SuccessRate)
	require.Nil(t, group.Buckets[29].TTFTMs)
	require.InDelta(t, 10.0, *group.SuccessRate, 0.0001)
	require.EqualValues(t, 10, *group.RequestCount)
	require.EqualValues(t, 1, *group.SuccessCount)
	require.EqualValues(t, 9, *group.FailureCount)
	require.InDelta(t, 100.0, *group.AverageLatencyMs, 0.0001)
	require.InDelta(t, 20.0, *group.TTFTMs, 0.0001)
	require.InDelta(t, 100.0, *group.TPS, 0.0001)
	require.InDelta(t, 0.03, *group.ProbeCost, 0.0001)

	model := metrics[0]
	require.Len(t, model.Buckets, 30)
	require.Equal(t, 100.0, *model.Buckets[28].SuccessRate)
	require.InDelta(t, 20.0, *model.Buckets[28].TTFTMs, 0.0001)
	require.Equal(t, 0.0, *model.Buckets[29].SuccessRate)
	require.InDelta(t, 10.0, *model.SuccessRate, 0.0001)
	require.EqualValues(t, 10, *model.RequestCount)
	require.EqualValues(t, 1, *model.SuccessCount)
	require.EqualValues(t, 9, *model.FailureCount)
	require.InDelta(t, 100.0, *model.AverageLatencyMs, 0.0001)
	require.InDelta(t, 20.0, *model.TTFTMs, 0.0001)
	require.InDelta(t, 100.0, *model.TPS, 0.0001)
	require.InDelta(t, 0.03, *model.ProbeCost, 0.0001)
}

func TestModelMonitorGroupMetricsBatchKeepsMonitorsSeparateAndReadsHourlyRollups(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	now := time.Date(2026, 8, 3, 12, 30, 10, 0, time.UTC)
	hour := now.Truncate(time.Hour)
	rows := sqlmock.NewRows([]string{
		"monitor_id", "group_id", "bucket_start", "request_count", "success_count",
		"latency_sum_ms", "latency_count", "ttft_sum_ms", "ttft_count",
		"output_tokens", "generation_ms", "probe_cost", "probe_cost_known",
	}).
		AddRow(3, 0, hour, 5, 5, 500, 5, 100, 5, 250, 5000, 0.02, true).
		AddRow(3, 7, hour, 5, 5, 500, 5, 100, 5, 250, 5000, 0.02, true).
		AddRow(4, 0, hour, 4, 2, 800, 2, 240, 2, 100, 2000, 0.03, true).
		AddRow(4, 8, hour, 4, 2, 800, 2, 240, 2, 100, 2000, 0.03, true)
	mock.ExpectQuery(`(?s)resolution='hour'.*resolution='minute'.*NOT EXISTS`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), string(service.ModelMonitorResolutionHour), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	repo := &modelMonitorRepository{db: db}
	metrics, err := repo.GroupMetricsBatch(context.Background(), []service.ModelMonitorMetricScope{
		{MonitorID: 3, GroupIDs: []int64{7}},
		{MonitorID: 4, GroupIDs: []int64{8}},
	}, service.ModelMonitorResolutionHour, now)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
	require.InDelta(t, 100, *metrics[3][7].SuccessRate, 0.0001)
	require.InDelta(t, 50, *metrics[4][8].SuccessRate, 0.0001)
	require.Len(t, metrics[3][0].Buckets, service.ModelMonitorMetricBucketCount)
	require.Equal(t, hour, metrics[3][0].Buckets[service.ModelMonitorMetricBucketCount-1].StartedAt)
}

func TestModelMonitorDueGroupsUseLatestActivityAndConditionalClaim(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	now := time.Date(2026, 8, 3, 12, 30, 0, 0, time.UTC)
	createdAt := now.Add(-time.Hour)
	lastTraffic := now.Add(-10 * time.Minute)
	lastProbe := now.Add(-5 * time.Minute)
	columns := []string{
		"id", "platform", "model", "model_enabled", "model_interval", "display_order", "label", "last_checked_at", "created_by", "created_at", "updated_at",
		"group_id", "priority", "group_enabled", "group_interval", "last_traffic_at", "last_probe_at", "group_name",
	}
	mock.ExpectQuery(`(?s)GREATEST\(COALESCE\(mg\.last_traffic_at,m\.created_at\),COALESCE\(mg\.last_probe_at,m\.created_at\)\).*ORDER BY GREATEST`).
		WithArgs(now, 100).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			3, "openai", "gpt-5", false, 300, 0, "", nil, 0, createdAt, createdAt,
			7, 0, true, 300, lastTraffic, lastProbe, "Standard",
		))

	repo := &modelMonitorRepository{db: db}
	due, err := repo.ListEnabledDueGroups(context.Background(), now, 100)
	require.NoError(t, err)
	require.Len(t, due, 1)
	require.Equal(t, lastTraffic, *due[0].Group.LastTrafficAt)
	require.Equal(t, lastProbe, *due[0].Group.LastProbeAt)

	claimedUntil := now.Add(3 * time.Minute)
	mock.ExpectExec(`(?s)UPDATE model_monitor_groups mg.*GREATEST\(COALESCE\(mg\.last_traffic_at,m\.created_at\),COALESCE\(mg\.last_probe_at,m\.created_at\)\)`).
		WithArgs(int64(3), int64(7), now, claimedUntil).
		WillReturnResult(sqlmock.NewResult(0, 1))
	claimed, err := repo.ClaimDueGroup(context.Background(), 3, 7, now, claimedUntil)
	require.NoError(t, err)
	require.True(t, claimed)
	require.NoError(t, mock.ExpectationsWereMet())
}
