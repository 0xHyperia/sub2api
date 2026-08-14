package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaymentOrderAmountSnapshotsMigrationDoesNotRewriteHistory(t *testing.T) {
	t.Parallel()
	content, err := FS.ReadFile("215_payment_order_amount_snapshots.sql")
	require.NoError(t, err)
	sql := strings.ToLower(string(content))
	require.Contains(t, sql, "payment_principal_amount")
	require.Contains(t, sql, "entitlement_principal_amount")
	require.Contains(t, sql, "surcharge_amount")
	require.Contains(t, sql, "nullif(new.payment_principal_amount, 0)")
	require.NotContains(t, sql, "update payment_orders")
}
