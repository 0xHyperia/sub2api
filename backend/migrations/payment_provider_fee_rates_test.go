package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaymentProviderFeeRatesMigration(t *testing.T) {
	t.Parallel()

	content, err := FS.ReadFile("202_payment_provider_fee_rates.sql")
	require.NoError(t, err)
	sql := strings.ToLower(string(content))
	require.Contains(t, sql, "add column if not exists fee_rates")
	require.Contains(t, sql, "alipay_recharge_fee_rate")
	require.Contains(t, sql, "wxpay_recharge_fee_rate")
	require.Contains(t, sql, "jsonb_object_agg")
	require.Contains(t, sql, "trim(instance.fee_rates) = ''")
}
