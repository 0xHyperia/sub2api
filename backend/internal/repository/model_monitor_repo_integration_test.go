//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestModelMonitorRefreshTrafficMetricsExecutesOnMigratedSchema(t *testing.T) {
	to := time.Now().UTC().Truncate(time.Minute)
	from := to.Add(-time.Minute)
	repo := &modelMonitorRepository{db: integrationDB}

	require.NoError(t, repo.RefreshTrafficMetrics(context.Background(), from, to))
}
