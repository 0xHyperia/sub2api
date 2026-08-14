package service

import (
	"math"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/shopspring/decimal"
)

const defaultBalanceRechargeMultiplier = 1.0

func normalizeBalanceRechargeMultiplier(multiplier float64) float64 {
	if math.IsNaN(multiplier) || math.IsInf(multiplier, 0) || multiplier <= 0 {
		return defaultBalanceRechargeMultiplier
	}
	return multiplier
}

// normalizeSubscriptionUSDToCNYRate 将非法值归一为 0（换算关闭）。
// 与余额倍率不同，0 是合法状态：表示订阅保持 price 直付的存量行为。
func normalizeSubscriptionUSDToCNYRate(rate float64) float64 {
	if math.IsNaN(rate) || math.IsInf(rate, 0) || rate < 0 {
		return 0
	}
	return rate
}

func calculateCreditedBalance(paymentAmount, multiplier float64) float64 {
	return decimal.NewFromFloat(paymentAmount).
		Mul(decimal.NewFromFloat(normalizeBalanceRechargeMultiplier(multiplier))).
		Round(2).
		InexactFloat64()
}

func calculateCreditedBalanceWithBonus(paymentAmount, multiplier, bonus float64) float64 {
	return decimal.NewFromFloat(calculateCreditedBalance(paymentAmount, multiplier)).
		Add(decimal.NewFromFloat(math.Max(0, bonus))).
		Round(2).
		InexactFloat64()
}

func calculatePaymentSurcharge(principal, payAmount float64) float64 {
	if principal <= 0 || payAmount <= principal {
		return 0
	}
	return decimal.NewFromFloat(payAmount).
		Sub(decimal.NewFromFloat(principal)).
		Round(2).
		InexactFloat64()
}

// PaymentOrderPrincipalAmount returns the payment-currency principal used by
// new-order cash reporting. The PayAmount fallback preserves the existing
// payment dashboard behavior for legacy rows without rewriting data.
func PaymentOrderPrincipalAmount(order *dbent.PaymentOrder) float64 {
	if order == nil {
		return 0
	}
	if order.PaymentPrincipalAmount > 0 {
		return order.PaymentPrincipalAmount
	}
	if order.PayAmount > 0 {
		return order.PayAmount
	}
	return order.Amount
}

// PaymentOrderEntitlementPrincipalAmount excludes quick-recharge bonuses. The
// legacy fallback preserves the historical affiliate rebate behavior.
func PaymentOrderEntitlementPrincipalAmount(order *dbent.PaymentOrder) float64 {
	if order == nil {
		return 0
	}
	if order.EntitlementPrincipalAmount > 0 {
		return order.EntitlementPrincipalAmount
	}
	return order.Amount
}

func paymentPrincipalRefundAmount(orderAmount, principalAmount, refundAmount float64) float64 {
	if orderAmount <= 0 || principalAmount <= 0 || refundAmount <= 0 {
		return 0
	}
	if refundAmount >= orderAmount {
		return decimal.NewFromFloat(principalAmount).Round(2).InexactFloat64()
	}
	return decimal.NewFromFloat(principalAmount).
		Mul(decimal.NewFromFloat(refundAmount)).
		Div(decimal.NewFromFloat(orderAmount)).
		Round(2).
		InexactFloat64()
}

func quickRechargeBonus(amount float64, options []QuickRechargeAmount) (float64, bool) {
	if !validMoneyAmount(amount, false) {
		return 0, false
	}
	cents := int64(math.Round(amount * 100))
	for _, option := range options {
		if int64(math.Round(option.Amount*100)) == cents {
			return option.Bonus, true
		}
	}
	return 0, false
}

func calculateGatewayRefundAmount(orderAmount, payAmount, refundAmount float64, currency string) float64 {
	if orderAmount <= 0 || payAmount <= 0 || refundAmount <= 0 {
		return 0
	}
	fractionDigits := int32(payment.CurrencyMaxFractionDigits(currency))
	if math.Abs(refundAmount-orderAmount) <= paymentAmountToleranceForCurrency(currency) {
		return decimal.NewFromFloat(payAmount).Round(fractionDigits).InexactFloat64()
	}
	return decimal.NewFromFloat(payAmount).
		Mul(decimal.NewFromFloat(refundAmount)).
		Div(decimal.NewFromFloat(orderAmount)).
		Round(fractionDigits).
		InexactFloat64()
}
