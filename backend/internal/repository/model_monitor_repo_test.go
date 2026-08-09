package repository

import (
	"context"
	"database/sql/driver"
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

func TestModelMonitorDueGroupsUseAlignedSlotsAndFairCompensationPriority(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	now := time.Date(2026, 8, 3, 12, 30, 0, 0, time.UTC)
	createdAt := now.Add(-time.Hour)
	lastTraffic := now.Add(-10 * time.Minute)
	lastProbe := now.Add(-5 * time.Minute)
	columns := []string{
		"id", "platform", "model", "model_enabled", "model_interval", "display_order", "label", "last_checked_at", "created_by", "created_at", "updated_at",
		"group_id", "priority", "group_enabled", "group_interval", "compensation_enabled", "compensation_pending",
		"last_traffic_at", "last_probe_at", "last_scheduled_slot_at", "next_compensation_at", "consecutive_failures", "group_name",
	}
	lastSlot := now.Add(-5 * time.Minute)
	mock.ExpectQuery(`(?s)to_timestamp\(floor\(extract\(epoch FROM \$1::timestamptz\).*ORDER BY ROW_NUMBER\(\) OVER.*PARTITION BY mg\.failure_compensation_pending.*CASE WHEN mg\.failure_compensation_pending`).
		WithArgs(now, 100).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			3, "openai", "gpt-5", false, 300, 0, "", nil, 0, createdAt, createdAt,
			7, 0, true, 300, true, false, lastTraffic, lastProbe, lastSlot, nil, 0, "Standard",
		))

	repo := &modelMonitorRepository{db: db}
	due, err := repo.ListEnabledDueGroups(context.Background(), now, 100)
	require.NoError(t, err)
	require.Len(t, due, 1)
	require.Equal(t, lastTraffic, *due[0].Group.LastTrafficAt)
	require.Equal(t, lastProbe, *due[0].Group.LastProbeAt)
	require.Equal(t, lastSlot, *due[0].Group.LastScheduledSlotAt)
	require.True(t, due[0].Group.FailureCompensationEnabled)

	claimedUntil := now.Add(3 * time.Minute)
	mock.ExpectExec(`(?s)UPDATE model_monitor_groups mg.*failure_compensation_pending=FALSE.*\$5 > COALESCE\(mg\.last_scheduled_slot_at`).
		WithArgs(int64(3), int64(7), now, claimedUntil, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	claimed, err := repo.ClaimDueGroup(context.Background(), 3, 7, now, claimedUntil, now, false)
	require.NoError(t, err)
	require.True(t, claimed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModelMonitorCompensationClaimRequiresPendingRetry(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	now := time.Date(2026, 8, 3, 12, 31, 0, 0, time.UTC)
	claimedUntil := now.Add(3 * time.Minute)
	mock.ExpectExec(`(?s)failure_compensation_enabled=TRUE AND mg\.failure_compensation_pending=TRUE.*next_compensation_at <= \$3`).
		WithArgs(int64(3), int64(7), now, claimedUntil, now).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := &modelMonitorRepository{db: db}
	claimed, err := repo.ClaimDueGroup(context.Background(), 3, 7, now, claimedUntil, now, true)
	require.NoError(t, err)
	require.True(t, claimed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModelMonitorUpsertGroupConfigCastsIntervalForTimestampAlignment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectExec(`(?s)INSERT INTO model_monitor_groups.*\$4::integer.*to_timestamp\(floor\(extract\(epoch FROM NOW\(\)\) / \(\(\$4::integer\)::double precision\)\) \* \(\(\$4::integer\)::double precision\)\)`).
		WithArgs(int64(3), int64(7), true, 300, true).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := &modelMonitorRepository{db: db}
	err = repo.UpsertGroupConfig(context.Background(), service.ModelMonitorGroupConfig{
		MonitorID: 3, GroupID: 7, Enabled: true, IntervalSeconds: 300,
		FailureCompensationEnabled: true,
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModelMonitorPersistGroupProbeResultCommitsAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	checkedAt := time.Date(2026, 8, 9, 12, 31, 7, 0, time.UTC)
	slot := checkedAt.Truncate(time.Minute)
	groupID := int64(7)
	latency, ttft, cost := 1200, 180, 0.012
	monitor := &service.ModelMonitor{ID: 3}
	history := &service.ModelMonitorHistory{
		MonitorID: 3, GroupID: &groupID, GroupName: "Standard", Status: service.MonitorStatusOperational,
		LatencyMs: &latency, FirstTokenMs: &ttft, OutputTokens: 20, GenerationMs: 1000,
		ProbeCost: &cost, ProbeCostKnown: true, CheckedAt: checkedAt,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)INSERT INTO model_monitor_histories.*scheduled_slot_at.*ON CONFLICT.*RETURNING id`).
		WithArgs(anyArgs(14)...).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(99)))
	mock.ExpectExec(`(?s)INSERT INTO model_monitor_metric_buckets.*ON CONFLICT`).
		WithArgs(anyArgs(13)...).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)UPDATE model_monitor_groups.*last_probe_at.*probe_claimed_until=CASE`).
		WithArgs(monitor.ID, groupID, slot, checkedAt, true, false).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE model_monitors SET last_checked_at=GREATEST`).
		WithArgs(monitor.ID, checkedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := &modelMonitorRepository{db: db}
	require.NoError(t, repo.PersistGroupProbeResult(context.Background(), monitor, history, &slot))
	require.EqualValues(t, 99, history.ID)
	require.Equal(t, checkedAt, *monitor.LastCheckedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModelMonitorPersistGroupProbeResultReusesCommittedSlot(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	checkedAt := time.Date(2026, 8, 9, 12, 31, 7, 0, time.UTC)
	existingCheckedAt := checkedAt.Add(-2 * time.Second)
	slot := checkedAt.Truncate(time.Minute)
	groupID := int64(7)
	monitor := &service.ModelMonitor{ID: 3}
	history := &service.ModelMonitorHistory{MonitorID: 3, GroupID: &groupID, Status: service.MonitorStatusOperational, CheckedAt: checkedAt}

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)INSERT INTO model_monitor_histories.*ON CONFLICT`).
		WithArgs(anyArgs(14)...).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT id,status,checked_at FROM model_monitor_histories`).
		WithArgs(monitor.ID, groupID, slot).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "checked_at"}).AddRow(int64(88), service.MonitorStatusFailed, existingCheckedAt))
	mock.ExpectExec(`UPDATE model_monitor_groups SET probe_claimed_until=NULL`).
		WithArgs(monitor.ID, groupID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := &modelMonitorRepository{db: db}
	require.NoError(t, repo.PersistGroupProbeResult(context.Background(), monitor, history, &slot))
	require.EqualValues(t, 88, history.ID)
	require.Equal(t, service.MonitorStatusFailed, history.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModelMonitorTrafficRefreshClaimIsMultiInstanceSafe(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	now := time.Date(2026, 8, 9, 12, 31, 0, 0, time.UTC)
	until := now.Add(2 * time.Minute)
	cursor := now.Add(-10 * time.Minute)

	mock.ExpectQuery(`(?s)UPDATE model_monitor_runtime_state SET traffic_claimed_until=.*RETURNING traffic_cursor_at`).
		WithArgs(now, until).
		WillReturnRows(sqlmock.NewRows([]string{"traffic_cursor_at"}).AddRow(cursor))
	repo := &modelMonitorRepository{db: db}
	got, claimed, err := repo.ClaimTrafficMetricsRefresh(context.Background(), now, until)
	require.NoError(t, err)
	require.True(t, claimed)
	require.Equal(t, cursor, *got)

	mock.ExpectQuery(`(?s)UPDATE model_monitor_runtime_state SET traffic_claimed_until=.*RETURNING traffic_cursor_at`).
		WithArgs(now, until).
		WillReturnRows(sqlmock.NewRows([]string{"traffic_cursor_at"}))
	got, claimed, err = repo.ClaimTrafficMetricsRefresh(context.Background(), now, until)
	require.NoError(t, err)
	require.False(t, claimed)
	require.Nil(t, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModelMonitorPassiveRefreshExcludesClientCancellation(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	from := time.Date(2026, 8, 9, 12, 30, 23, 0, time.UTC)
	to := from.Add(2 * time.Minute)
	normalizedFrom := from.Truncate(time.Minute)
	normalizedTo := to.Truncate(time.Minute).Add(time.Minute)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM model_monitor_metric_buckets`).
		WithArgs(normalizedFrom, normalizedTo).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(`(?s)WITH successful AS.*status_code=499.*error_type.*cancel.*error_owner.*client.*INSERT INTO model_monitor_metric_buckets`).
		WithArgs(normalizedFrom, normalizedTo).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)UPDATE model_monitor_groups mg SET last_traffic_at=.*status_code=499.*error_type.*cancel.*error_owner.*client`).
		WithArgs(normalizedFrom, normalizedTo).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)UPDATE model_monitor_groups mg.*failure_compensation_pending=FALSE`).
		WithArgs(normalizedFrom, normalizedTo).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := &modelMonitorRepository{db: db}
	require.NoError(t, repo.RefreshTrafficMetrics(context.Background(), from, to))
	require.NoError(t, mock.ExpectationsWereMet())
}

func anyArgs(count int) []driver.Value {
	args := make([]driver.Value, count)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	return args
}
