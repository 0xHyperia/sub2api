package service

import (
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestCommissionableOrderAmountExcludesChannelFee(t *testing.T) {
	t.Parallel()
	for _, orderType := range []string{payment.OrderTypeBalance, payment.OrderTypeSubscription} {
		t.Run(orderType, func(t *testing.T) {
			t.Parallel()
			order := &dbent.PaymentOrder{OrderType: orderType, Amount: 100, PayAmount: 103, FeeRate: 3}
			require.Equal(t, 100.0, commissionableOrderAmount(order))
			require.Equal(t, 100.0, affiliateRebateBaseAmount(order))
		})
	}
}

func TestCommissionableOrderAmountRejectsUnsupportedOrders(t *testing.T) {
	t.Parallel()
	require.Zero(t, commissionableOrderAmount(nil))
	require.Zero(t, commissionableOrderAmount(&dbent.PaymentOrder{OrderType: "other", Amount: 100, PayAmount: 103}))
}
