package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnifyRechargeBonusIntoTiersMigration(t *testing.T) {
	content, err := FS.ReadFile("270_unify_recharge_bonus_into_tiers.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "key = 'QUICK_RECHARGE_AMOUNTS'")
	require.Contains(t, sql, "key = 'RECHARGE_BONUS_TIERS'")
	require.Contains(t, sql, "'RECHARGE_BONUS_MODE', 'bonus'")
	require.Contains(t, sql, "e - 'bonus'")
	require.Contains(t, sql, "preset_count <= 20")
}
