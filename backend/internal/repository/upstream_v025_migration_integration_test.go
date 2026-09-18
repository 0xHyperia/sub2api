//go:build integration

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestMigrationsRunner_UpgradeFrom261To264(t *testing.T) {
	ctx := context.Background()
	dbName := fmt.Sprintf("sub2api_upgrade_261_%d", time.Now().UnixNano())
	adminURL, err := url.Parse(integrationDSN)
	require.NoError(t, err)
	adminURL.Path = "/postgres"
	adminDB, err := sql.Open("postgres", adminURL.String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = adminDB.Close() })
	require.NoError(t, execCreateDatabase(ctx, adminDB, dbName))
	adminURL.Path = "/" + dbName
	db, err := sql.Open("postgres", adminURL.String())
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
		_, err := adminDB.ExecContext(ctx, "DROP DATABASE "+pq.QuoteIdentifier(dbName)+" WITH (FORCE)")
		require.NoError(t, err)
	})
	require.NoError(t, applyMigrationsFS(ctx, db, migrationsThrough(t, 261)))
	requireMigrationPrefixCount(t, db, 262, 0)
	var userID int64
	require.NoError(t, db.QueryRowContext(ctx, `INSERT INTO users (email, password_hash)
VALUES ('upgrade-261@example.com', 'test') RETURNING id`).Scan(&userID))
	_, err = db.ExecContext(ctx, `INSERT INTO user_platform_quotas (user_id, platform, daily_limit_usd)
VALUES ($1, 'openai', NULL), ($1, 'minimax', 0), ($1, 'kimi', 10)`, userID)
	require.NoError(t, err)
	require.NoError(t, ApplyMigrations(ctx, db))
	require.NoError(t, ApplyMigrations(ctx, db))
	for version := 262; version <= 264; version++ {
		requireMigrationPrefixCount(t, db, version, 1)
	}
	for _, file := range []string{"262_opencode_go_platform.sql", "263_purge_unlimited_user_platform_quotas.sql", "264_model_monitor_provider_platforms.sql"} {
		body, err := migrations.FS.ReadFile(file)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, string(body))
		require.NoError(t, err, file)
	}
	var count int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM user_platform_quotas WHERE user_id = $1", userID).Scan(&count))
	require.Equal(t, 2, count, "explicit zero and positive quotas must survive cleanup")
	_, err = db.ExecContext(ctx, `INSERT INTO user_platform_quotas (user_id, platform, daily_limit_usd) VALUES ($1, 'opencode_go', 20)`, userID)
	require.NoError(t, err)
	for _, platform := range []string{"openai", "kimi", "zhipu", "deepseek", "minimax", "opencode_go"} {
		_, err = db.ExecContext(ctx, "INSERT INTO model_monitors (platform, model, created_by) VALUES ($1, 'test-model', $2)", platform, userID)
		require.NoError(t, err, platform)
	}
	_, err = db.ExecContext(ctx, "INSERT INTO model_monitors (platform, model, created_by) VALUES ('composite', 'invalid', $1)", userID)
	require.Error(t, err, "composite aliases must not become independent monitor platforms")
}
