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
			order := &dbent.PaymentOrder{
				OrderType: orderType, Amount: 110, PayAmount: 104, FeeRate: 4,
				PaymentPrincipalAmount: 100, EntitlementPrincipalAmount: 100,
			}
			require.Equal(t, 100.0, commissionableOrderAmount(order))
			require.Equal(t, 100.0, affiliateRebateBaseAmount(order))
		})
	}
}

func TestOrderAmountBasesKeepLegacyBehaviorWithoutBackfill(t *testing.T) {
	t.Parallel()
	legacy := &dbent.PaymentOrder{OrderType: payment.OrderTypeBalance, Amount: 52, PayAmount: 52}
	require.Equal(t, 52.0, commissionableOrderAmount(legacy))
	require.Equal(t, 52.0, affiliateRebateBaseAmount(legacy))
}

func TestOrderAmountBasesSeparatePaymentAndEntitlementValues(t *testing.T) {
	t.Parallel()
	order := &dbent.PaymentOrder{
		OrderType: payment.OrderTypeBalance, Amount: 9, PayAmount: 52,
		PaymentPrincipalAmount: 50, EntitlementPrincipalAmount: 7,
	}
	require.Equal(t, 50.0, commissionableOrderAmount(order))
	require.Equal(t, 7.0, affiliateRebateBaseAmount(order))
}

func TestCommissionableOrderAmountRejectsUnsupportedOrders(t *testing.T) {
	t.Parallel()
	require.Zero(t, commissionableOrderAmount(nil))
	require.Zero(t, commissionableOrderAmount(&dbent.PaymentOrder{OrderType: "other", Amount: 100, PayAmount: 103}))
}
