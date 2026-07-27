package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaymentFeeModesMigration(t *testing.T) {
	t.Parallel()

	content, err := FS.ReadFile("201_payment_fee_modes.sql")
	require.NoError(t, err)
	sql := strings.ToLower(string(content))
	require.Contains(t, sql, "provider_amount")
	require.Contains(t, sql, "fee_mode")
	require.Contains(t, sql, "set provider_amount = pay_amount")
	for _, mode := range []string{"platform", "provider", "merchant"} {
		require.Contains(t, sql, "'"+mode+"'")
	}
}
