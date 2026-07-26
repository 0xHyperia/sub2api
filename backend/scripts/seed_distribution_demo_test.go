package scripts

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDistributionDemoSeedIsNonProductionOnly(t *testing.T) {
	raw, err := os.ReadFile("seed-distribution-demo.sql")
	require.NoError(t, err)
	sql := string(raw)

	require.Contains(t, sql, "ALLOW_DISTRIBUTION_DEMO_SEED=I_UNDERSTAND_NON_PRODUCTION_ONLY")
	require.Contains(t, sql, "DISTRIBUTION_DEMO_DATABASE")
	require.Contains(t, sql, "demo_database_allowed")
	require.Contains(t, sql, "'!distribution-demo-disabled!'")
	require.Contains(t, sql, "'disabled'")
	require.NotContains(t, strings.ToLower(sql), "select password_hash from users")
	require.NotContains(t, strings.ToLower(sql), "admin.password_hash")
}
