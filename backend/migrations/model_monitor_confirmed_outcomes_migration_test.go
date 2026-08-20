package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModelMonitorConfirmedOutcomesMigration(t *testing.T) {
	content, err := FS.ReadFile("245_model_monitor_confirmed_outcomes.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "metric_semantics_version INTEGER NOT NULL DEFAULT 1")
	require.Contains(t, sql, "UPDATE model_monitor_histories SET status = 'error' WHERE status = 'failed'")
	require.Contains(t, sql, "selected model is at capacity")
	require.Contains(t, sql, "api returned 503:")
	require.Contains(t, sql, "model not found")
	require.Contains(t, sql, "deadline exceeded")
	require.Contains(t, sql, "DELETE FROM model_monitor_metric_buckets")
	require.Contains(t, sql, "traffic_cursor_at = NULL")
	require.Contains(t, sql, "metric_semantics_version = 2")
	require.GreaterOrEqual(t, strings.Count(sql, "metric_semantics_version < 2"), 3)
}
