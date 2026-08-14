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

	t.Run("preserves configured order and bonuses", func(t *testing.T) {
		cfg := (&PaymentConfigService{}).parsePaymentConfig(map[string]string{
			SettingQuickRechargeAmounts:  `[{"amount":100,"bonus":12.5},{"amount":20,"bonus":1}]`,
			SettingCustomRechargeEnabled: "false",
		})
		require.False(t, cfg.CustomRechargeEnabled)
		require.Equal(t, []QuickRechargeAmount{{Amount: 100, Bonus: 12.5}, {Amount: 20, Bonus: 1}}, cfg.QuickRechargeAmounts)
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

	duplicate := []QuickRechargeAmount{{Amount: 10}, {Amount: 10, Bonus: 1}}
	err = svc.UpdatePaymentConfig(context.Background(), UpdatePaymentConfigRequest{QuickRechargeAmounts: &duplicate})
	require.Error(t, err)

	valid := []QuickRechargeAmount{{Amount: 10, Bonus: 1.25}, {Amount: 50}}
	err = svc.UpdatePaymentConfig(context.Background(), UpdatePaymentConfigRequest{
		QuickRechargeAmounts:  &valid,
		CustomRechargeEnabled: &customDisabled,
	})
	require.NoError(t, err)
	require.JSONEq(t, `[{"amount":10,"bonus":1.25},{"amount":50,"bonus":0}]`, repo.values[SettingQuickRechargeAmounts])
	require.Equal(t, "false", repo.values[SettingCustomRechargeEnabled])
}

func TestRechargeBonusAndCustomAmountValidation(t *testing.T) {
	options := []QuickRechargeAmount{{Amount: 10, Bonus: 2.5}, {Amount: 50}}
	bonus, matched := quickRechargeBonus(10, options)
	require.True(t, matched)
	require.Equal(t, 2.5, bonus)
	require.Equal(t, 9.5, calculateCreditedBalanceWithBonus(10, 0.7, bonus))

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
	require.Equal(t, 9.0, calculateCreditedBalanceWithBonus(50, 0.14, 2))
	require.Equal(t, 25.0, paymentPrincipalRefundAmount(9, 50, 4.5))
	require.Equal(t, 50.0, paymentPrincipalRefundAmount(9, 50, 9))
}
