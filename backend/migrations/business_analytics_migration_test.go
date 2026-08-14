package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBusinessAnalyticsMigrationAvoidsHotTableIndexLocksAndClosesBackfillRaces(t *testing.T) {
	content, err := FS.ReadFile("214_business_analytics.sql")
	require.NoError(t, err)
	sql := strings.Join(strings.Fields(string(content)), " ")

	require.NotContains(t, sql, "CREATE INDEX IF NOT EXISTS idx_usage_logs_",
		"a transactional baseline must not build ordinary indexes on the hot usage table")

	trigger := strings.Index(sql, "CREATE TRIGGER business_payment_fact_sync")
	backfill := strings.LastIndex(sql, "INSERT INTO business_payment_facts")
	require.NotEqual(t, -1, trigger)
	require.NotEqual(t, -1, backfill)
	require.Less(t, trigger, backfill, "payment changes must be captured before historical facts are scanned")

	require.Equal(t, 1, strings.Count(sql, "FROM usage_logs ul WHERE ul.actual_cost > 0 GROUP BY 1, 2"),
		"usage history must be aggregated in one reconciliation scan")
	require.Contains(t, sql, "usage_coverage_from = COALESCE( usage_coverage_from,")
}
