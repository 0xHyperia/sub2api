package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestParseQuickRechargeAmounts(t *testing.T) {
	t.Run("uses backward compatible defaults", func(t *testing.T) {
		cfg := (&PaymentConfigService{}).parsePaymentConfig(map[string]string{})
		require.True(t, cfg.CustomRechargeEnabled)
		require.Len(t, cfg.QuickRechargeAmounts, 9)
		require.Equal(t, float64(10), cfg.QuickRechargeAmounts[0].Amount)
	})

	t.Run("preserves configured order and ignores legacy bonus field", func(t *testing.T) {
		cfg := (&PaymentConfigService{}).parsePaymentConfig(map[string]string{
			SettingQuickRechargeAmounts:  `[{"amount":100,"bonus":12.5},{"amount":20,"bonus":1}]`,
			SettingCustomRechargeEnabled: "false",
		})
		require.False(t, cfg.CustomRechargeEnabled)
		require.Equal(t, []QuickRechargeAmount{{Amount: 100}, {Amount: 20}}, cfg.QuickRechargeAmounts)
	})
}

func TestUpdatePaymentConfigValidatesQuickRechargeAmounts(t *testing.T) {
	repo := &paymentConfigSettingRepoStub{values: map[string]string{}}
	svc := &PaymentConfigService{settingRepo: repo}
	customDisabled := false

	empty := []QuickRechargeAmount{}
	err := svc.UpdatePaymentConfig(context.Background(), UpdatePaymentConfigRequest{
		QuickRechargeAmounts:  &empty,
		CustomRechargeEnabled: &customDisabled,
	})
	require.Error(t, err)

	repo.values[SettingQuickRechargeAmounts] = "[]"
	err = svc.UpdatePaymentConfig(context.Background(), UpdatePaymentConfigRequest{
		CustomRechargeEnabled: &customDisabled,
	})
	require.Error(t, err)
	delete(repo.values, SettingQuickRechargeAmounts)

	duplicate := []QuickRechargeAmount{{Amount: 10}, {Amount: 10}}
	err = svc.UpdatePaymentConfig(context.Background(), UpdatePaymentConfigRequest{QuickRechargeAmounts: &duplicate})
	require.Error(t, err)

	valid := []QuickRechargeAmount{{Amount: 10}, {Amount: 50}}
	err = svc.UpdatePaymentConfig(context.Background(), UpdatePaymentConfigRequest{
		QuickRechargeAmounts:  &valid,
		CustomRechargeEnabled: &customDisabled,
	})
	require.NoError(t, err)
	require.JSONEq(t, `[{"amount":10},{"amount":50}]`, repo.values[SettingQuickRechargeAmounts])
	require.Equal(t, "false", repo.values[SettingCustomRechargeEnabled])
}

func TestRechargeBonusAndCustomAmountValidation(t *testing.T) {
	options := []QuickRechargeAmount{{Amount: 10}, {Amount: 50}}
	require.True(t, quickRechargeAmountMatched(10, options))
	require.False(t, quickRechargeAmountMatched(11, options))

	// 赠送统一由阶梯表达：倍率 0.7、10 元命中 25% 赠金 → 到账基数 7，赠送 1.75
	quote := quoteRechargeBonus(&PaymentConfig{
		BalanceRechargeMultiplier: 0.7,
		RechargeBonusMode:         RechargeBonusModeBonus,
		RechargeBonusTiers:        []RechargeBonusTier{{MinAmount: 10, BonusPercent: 25}},
	}, 10, "CNY")
	require.Equal(t, 1.75, quote.Bonus)
	require.Equal(t, 8.75, quote.Credited)
	require.Equal(t, 7.0, rechargeQuotePaidCredit(quote))

	svc := &PaymentService{}
	cfg := &PaymentConfig{
		MinAmount:             1,
		QuickRechargeAmounts:  options,
		CustomRechargeEnabled: false,
	}
	_, err := svc.validateOrderInput(context.Background(), CreateOrderRequest{
		Amount:    11,
		OrderType: payment.OrderTypeBalance,
	}, cfg)
	require.Error(t, err)

	_, err = svc.validateOrderInput(context.Background(), CreateOrderRequest{
		Amount:    10,
		OrderType: payment.OrderTypeBalance,
	}, cfg)
	require.NoError(t, err)
}

func TestPaymentAmountSnapshotsSeparateFeeBonusAndMultiplier(t *testing.T) {
	t.Parallel()
	require.Equal(t, 2.0, calculatePaymentSurcharge(50, 52))
	require.Equal(t, 7.0, calculateCreditedBalance(50, 0.14))
	require.Equal(t, 25.0, paymentPrincipalRefundAmount(9, 50, 4.5))
	require.Equal(t, 50.0, paymentPrincipalRefundAmount(9, 50, 9))
}
