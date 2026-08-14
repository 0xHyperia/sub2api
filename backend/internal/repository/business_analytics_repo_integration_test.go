//go:build integration

package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func createBusinessPaymentOrder(t *testing.T, user *service.User, status string, paidAt *time.Time, currency string, amount float64) int64 {
	t.Helper()
	suffix := time.Now().UnixNano()
	builder := integrationEntClient.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName("business-order").
		SetAmount(amount).
		SetPayAmount(amount).
		SetProviderAmount(amount).
		SetRechargeCode(fmt.Sprintf("BIZ-%d", suffix)).
		SetOutTradeNo(fmt.Sprintf("BIZORDER%d", suffix)).
		SetPaymentType("alipay").
		SetPaymentTradeNo(fmt.Sprintf("biz-trade-%d", suffix)).
		SetProviderSnapshot(map[string]any{"currency": currency}).
		SetStatus(status).
		SetClientIP("127.0.0.1").
		SetSrcHost("localhost").
		SetExpiresAt(time.Now().UTC().Add(time.Hour))
	if paidAt != nil {
		builder.SetPaidAt(*paidAt)
	}
	order, err := builder.Save(context.Background())
	require.NoError(t, err)
	return order.ID
}

func TestBusinessAnalyticsDailyBucketsDeduplicateUsersAndTrackActivationTime(t *testing.T) {
	ctx := context.Background()
	repo := NewBusinessAnalyticsRepository(integrationDB)
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	day := time.Date(2026, 8, 10, 0, 0, 0, 0, location)
	suffix := time.Now().UnixNano()
	user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-daily-%d@example.com", suffix), CreatedAt: day.Add(-24 * time.Hour)})
	apiKey := mustCreateApiKey(t, integrationEntClient, &service.APIKey{UserID: user.ID, Key: fmt.Sprintf("sk-business-%d", suffix), Name: "business"})
	account := mustCreateAccount(t, integrationEntClient, &service.Account{Name: fmt.Sprintf("business-%d", suffix)})
	usageRepo := newUsageLogRepositoryWithSQL(integrationEntClient, integrationDB)
	for index, at := range []time.Time{day.Add(2 * time.Hour), day.Add(7 * time.Hour)} {
		_, err := usageRepo.Create(ctx, &service.UsageLog{UserID: user.ID, APIKeyID: apiKey.ID, AccountID: account.ID, RequestID: fmt.Sprintf("business-%d-%d", suffix, index), Model: "test-model", TotalCost: 1, ActualCost: 1, CreatedAt: at})
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM business_analytics_buckets WHERE bucket_start >= $1 AND bucket_start < $2`, day, day.Add(24*time.Hour))
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, user.ID)
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM accounts WHERE id=$1`, account.ID)
	})

	require.NoError(t, repo.AggregateRange(ctx, day, day.Add(24*time.Hour)))
	var active, activated int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT (metrics->>'active_users')::bigint,(metrics->>'activated_users')::bigint FROM business_analytics_buckets WHERE resolution='day' AND scope='channel' AND scope_id='unknown' AND bucket_start=$1`, day).Scan(&active, &activated))
	require.EqualValues(t, 1, active, "a user active in multiple hours must count once in the daily bucket")
	require.EqualValues(t, 1, activated, "first activity must be counted on its actual day, not registration day")
	var activationHour int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT (metrics->>'activated_users')::bigint FROM business_analytics_buckets WHERE resolution='hour' AND scope='channel' AND scope_id='unknown' AND bucket_start=$1`, day.Add(2*time.Hour)).Scan(&activationHour))
	require.EqualValues(t, 1, activationHour)

	trend, available, err := repo.bucketTrend(ctx, service.BusinessAnalyticsFilter{DateFrom: day, DateTo: day.Add(24 * time.Hour), Granularity: "day"}, 7.2)
	require.NoError(t, err)
	require.True(t, available, "all five channel rows make the platform cache complete")
	require.Len(t, trend, 1)
	require.EqualValues(t, 1, trend[0].ActiveUsers)

	// A later incremental refresh must recompute the complete affected day rather
	// than replace its daily bucket with only the latest hour.
	require.NoError(t, repo.AggregateRange(ctx, day.Add(7*time.Hour), day.Add(8*time.Hour)))
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT (metrics->>'active_users')::bigint FROM business_analytics_buckets WHERE resolution='day' AND scope='channel' AND scope_id='unknown' AND bucket_start=$1`, day).Scan(&active))
	require.EqualValues(t, 1, active)
}

func TestBusinessAnalyticsBucketTrendRejectsPartialFinalBucket(t *testing.T) {
	repo := NewBusinessAnalyticsRepository(integrationDB)
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	day := time.Date(2026, 8, 10, 0, 0, 0, 0, location)

	trend, available, err := repo.bucketTrend(context.Background(), service.BusinessAnalyticsFilter{
		DateFrom: day, DateTo: day.Add(14*time.Hour + 23*time.Minute), Granularity: "hour",
	}, 7.2)
	require.NoError(t, err)
	require.False(t, available)
	require.Nil(t, trend, "partial hours must use exact event-time queries")
}

func TestBusinessAnalyticsDurableUsageFactsSurviveRawLogRetention(t *testing.T) {
	ctx := context.Background()
	repo := NewBusinessAnalyticsRepository(integrationDB)
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	start := time.Date(2026, 4, 1, 0, 0, 0, 0, location)
	suffix := time.Now().UnixNano()
	user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-retained-%d@example.com", suffix), CreatedAt: start.AddDate(0, 0, -30)})
	apiKey := mustCreateApiKey(t, integrationEntClient, &service.APIKey{UserID: user.ID, Key: fmt.Sprintf("sk-business-retained-%d", suffix), Name: "business-retained"})
	account := mustCreateAccount(t, integrationEntClient, &service.Account{Name: fmt.Sprintf("business-retained-%d", suffix)})
	usageRepo := newUsageLogRepositoryWithSQL(integrationEntClient, integrationDB)
	_, err := usageRepo.Create(ctx, &service.UsageLog{UserID: user.ID, APIKeyID: apiKey.ID, AccountID: account.ID, RequestID: fmt.Sprintf("business-retained-%d", suffix), Model: "test-model", TotalCost: 2, ActualCost: 3, CreatedAt: start.Add(2 * time.Hour)})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, user.ID)
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM accounts WHERE id=$1`, account.ID)
	})

	require.NoError(t, repo.AggregateRange(ctx, start, start.AddDate(0, 0, 1)))
	before, err := repo.periodValues(ctx, start, start.AddDate(0, 0, 1), "", 0, "", 7.2, false)
	require.NoError(t, err)
	require.EqualValues(t, 1, before.active)
	require.InDelta(t, 3, before.consumed, 0.0001)

	_, err = integrationDB.ExecContext(ctx, `DELETE FROM usage_logs WHERE user_id=$1`, user.ID)
	require.NoError(t, err)
	after, err := repo.periodValues(ctx, start, start.AddDate(0, 0, 1), "", 0, "", 7.2, false)
	require.NoError(t, err)
	require.EqualValues(t, before.active, after.active)
	require.InDelta(t, before.consumed, after.consumed, 0.0001)
	require.InDelta(t, before.supplier, after.supplier, 0.0001)
}

func TestBusinessAnalyticsAggregationRepairsMissingUsageFacts(t *testing.T) {
	ctx := context.Background()
	repo := NewBusinessAnalyticsRepository(integrationDB)
	suffix := time.Now().UnixNano()
	at := time.Now().UTC().Add(-2 * time.Minute).Truncate(time.Second)
	user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-usage-repair-%d@example.com", suffix), CreatedAt: at.Add(-time.Hour)})
	apiKey := mustCreateApiKey(t, integrationEntClient, &service.APIKey{UserID: user.ID, Key: fmt.Sprintf("sk-business-usage-repair-%d", suffix), Name: "business-usage-repair"})
	account := mustCreateAccount(t, integrationEntClient, &service.Account{Name: fmt.Sprintf("business-usage-repair-%d", suffix)})
	usageRepo := newUsageLogRepositoryWithSQL(integrationEntClient, integrationDB)
	_, err := usageRepo.Create(ctx, &service.UsageLog{UserID: user.ID, APIKeyID: apiKey.ID, AccountID: account.ID, RequestID: fmt.Sprintf("business-usage-repair-%d", suffix), Model: "test-model", TotalCost: 2, ActualCost: 3, CreatedAt: at})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, user.ID)
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM accounts WHERE id=$1`, account.ID)
	})

	_, err = integrationDB.ExecContext(ctx, `DELETE FROM business_usage_hourly_facts WHERE user_id=$1`, user.ID)
	require.NoError(t, err)
	_, err = integrationDB.ExecContext(ctx, `UPDATE business_user_facts SET first_activated_at=NULL WHERE user_id=$1`, user.ID)
	require.NoError(t, err)
	require.NoError(t, repo.AggregateRange(ctx, at.Add(-time.Hour), at.Add(time.Hour)))

	var first time.Time
	var consumed float64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT first_activated_at FROM business_user_facts WHERE user_id=$1`, user.ID).Scan(&first))
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT consumed_revenue FROM business_usage_hourly_facts WHERE user_id=$1`, user.ID).Scan(&consumed))
	require.WithinDuration(t, at, first, time.Second)
	require.InDelta(t, 3, consumed, 0.0001)
}

func TestBusinessAnalyticsAggregationRepairsMissingPaymentFacts(t *testing.T) {
	ctx := context.Background()
	repo := NewBusinessAnalyticsRepository(integrationDB)
	paidAt := time.Now().UTC().Add(-2 * time.Minute).Truncate(time.Second)
	user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-payment-repair-%d@example.com", time.Now().UnixNano())})
	orderID := createBusinessPaymentOrder(t, user, "COMPLETED", &paidAt, "CNY", 25)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, user.ID)
	})

	_, err := integrationDB.ExecContext(ctx, `DELETE FROM business_payment_facts WHERE payment_order_id=$1`, orderID)
	require.NoError(t, err)
	_, err = integrationDB.ExecContext(ctx, `UPDATE business_user_facts SET first_paid_at=NULL WHERE user_id=$1`, user.ID)
	require.NoError(t, err)
	require.NoError(t, repo.AggregateRange(ctx, paidAt.Add(-time.Hour), paidAt.Add(time.Hour)))

	var first time.Time
	var gross float64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT first_paid_at FROM business_user_facts WHERE user_id=$1`, user.ID).Scan(&first))
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT gross_amount FROM business_payment_facts WHERE payment_order_id=$1`, orderID).Scan(&gross))
	require.WithinDuration(t, paidAt, first, time.Second)
	require.InDelta(t, 25, gross, 0.0001)
}

func TestBusinessAnalyticsAggregationMarksRepairedForeignPaymentFXEstimated(t *testing.T) {
	ctx := context.Background()
	repo := NewBusinessAnalyticsRepository(integrationDB)
	paidAt := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Minute)
	user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-fx-repair-%d@example.com", time.Now().UnixNano())})
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, user.ID)
	})
	orderID := createBusinessPaymentOrder(t, user, "COMPLETED", &paidAt, "USD", 100)
	_, err := integrationDB.ExecContext(ctx, `DELETE FROM business_payment_facts WHERE payment_order_id=$1`, orderID)
	require.NoError(t, err)

	require.NoError(t, repo.AggregateRange(ctx, paidAt.Add(-time.Hour), paidAt.Add(time.Hour)))
	var rate float64
	var estimated bool
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT fx_rate_to_cny,fx_estimated FROM business_payment_facts WHERE payment_order_id=$1`, orderID).Scan(&rate, &estimated))
	require.Positive(t, rate)
	require.True(t, estimated, "a reconstructed foreign-currency order has no trustworthy event-time FX snapshot")
}

func TestBusinessUsageFactMovesFirstActivationBackForOutOfOrderImport(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-out-of-order-%d@example.com", suffix)})
	apiKey := mustCreateApiKey(t, integrationEntClient, &service.APIKey{UserID: user.ID, Key: fmt.Sprintf("sk-business-out-of-order-%d", suffix), Name: "business-out-of-order"})
	account := mustCreateAccount(t, integrationEntClient, &service.Account{Name: fmt.Sprintf("business-out-of-order-%d", suffix)})
	usageRepo := newUsageLogRepositoryWithSQL(integrationEntClient, integrationDB)
	newer := time.Now().UTC().AddDate(0, 0, -2).Truncate(time.Second)
	older := newer.AddDate(0, 0, -20)
	for index, at := range []time.Time{newer, older} {
		_, err := usageRepo.Create(ctx, &service.UsageLog{UserID: user.ID, APIKeyID: apiKey.ID, AccountID: account.ID, RequestID: fmt.Sprintf("business-out-of-order-%d-%d", suffix, index), Model: "test-model", TotalCost: 1, ActualCost: 1, CreatedAt: at})
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, user.ID)
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM accounts WHERE id=$1`, account.ID)
	})

	var first time.Time
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT first_activated_at FROM business_user_facts WHERE user_id=$1`, user.ID).Scan(&first))
	require.WithinDuration(t, older, first, time.Second)
}

func TestBusinessPaymentChannelSnapshotDoesNotMoveWithLaterAttribution(t *testing.T) {
	ctx := context.Background()
	repo := NewBusinessAnalyticsRepository(integrationDB)
	now := time.Now().UTC().Truncate(time.Second)
	user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-payment-channel-%d@example.com", time.Now().UnixNano())})
	orderID := createBusinessPaymentOrder(t, user, "COMPLETED", &now, "CNY", 25)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, user.ID)
	})

	var eventChannel string
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT channel FROM business_payment_facts WHERE payment_order_id=$1`, orderID).Scan(&eventChannel))
	require.Equal(t, "unknown", eventChannel)
	_, err := integrationDB.ExecContext(ctx, `UPDATE business_user_facts SET channel='distribution' WHERE user_id=$1`, user.ID)
	require.NoError(t, err)

	values, err := repo.periodValues(ctx, now.Add(-time.Hour), now.Add(time.Hour), "unknown", 0, "", 7.2, false)
	require.NoError(t, err)
	require.InDelta(t, 25, values.gross, 0.0001, "cash revenue must stay in its event-time channel")
}

func TestBusinessPaymentFactPreservesEstimatedFXOnRefund(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	now := time.Now().UTC()
	user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-fx-%d@example.com", suffix)})
	order, err := integrationEntClient.PaymentOrder.Create().SetUserID(user.ID).SetUserEmail(user.Email).SetUserName("business-fx").SetAmount(100).SetPayAmount(100).SetRechargeCode(fmt.Sprintf("BIZFX-%d", suffix)).SetOutTradeNo(fmt.Sprintf("BIZFXORDER%d", suffix)).SetPaymentType("stripe").SetPaymentTradeNo(fmt.Sprintf("biz-fx-%d", suffix)).SetProviderSnapshot(map[string]any{"currency": "USD"}).SetStatus("COMPLETED").SetClientIP("127.0.0.1").SetSrcHost("localhost").SetExpiresAt(now.Add(time.Hour)).SetPaidAt(now).Save(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, user.ID)
	})
	_, err = integrationDB.ExecContext(ctx, `UPDATE business_payment_facts SET fx_estimated=TRUE WHERE payment_order_id=$1`, order.ID)
	require.NoError(t, err)
	_, err = integrationDB.ExecContext(ctx, `UPDATE payment_orders SET status='PARTIALLY_REFUNDED',refund_amount=50,refund_at=$2 WHERE id=$1`, order.ID, now.Add(time.Minute))
	require.NoError(t, err)
	var estimated bool
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT fx_estimated FROM business_payment_facts WHERE payment_order_id=$1`, order.ID).Scan(&estimated))
	require.True(t, estimated)
}

func TestBusinessPaymentFactUsesPrincipalWithoutSurchargeOrBonus(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	now := time.Now().UTC()
	user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-principal-%d@example.com", suffix)})
	order, err := integrationEntClient.PaymentOrder.Create().
		SetUserID(user.ID).SetUserEmail(user.Email).SetUserName("business-principal").
		SetAmount(52).SetPayAmount(52).SetProviderAmount(50).
		SetPaymentPrincipalAmount(50).SetEntitlementPrincipalAmount(50).SetSurchargeAmount(2).
		SetRechargeCode(fmt.Sprintf("BIZPRINCIPAL-%d", suffix)).SetOutTradeNo(fmt.Sprintf("BIZPRINCIPALORDER%d", suffix)).
		SetPaymentType("alipay").SetPaymentTradeNo(fmt.Sprintf("biz-principal-%d", suffix)).
		SetProviderSnapshot(map[string]any{"currency": "CNY"}).SetStatus("COMPLETED").
		SetClientIP("127.0.0.1").SetSrcHost("localhost").SetExpiresAt(now.Add(time.Hour)).SetPaidAt(now).Save(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, user.ID)
	})

	var gross float64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT gross_amount FROM business_payment_facts WHERE payment_order_id=$1`, order.ID).Scan(&gross))
	require.Equal(t, 50.0, gross)
}

func TestBusinessPaymentFactSurvivesRefundLifecycleUntilRefundCompletes(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	now := time.Now().UTC()
	user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-refund-%d@example.com", suffix)})
	order, err := integrationEntClient.PaymentOrder.Create().SetUserID(user.ID).SetUserEmail(user.Email).SetUserName("business-refund").SetAmount(100).SetPayAmount(100).SetProviderAmount(100).SetRechargeCode(fmt.Sprintf("BIZREF-%d", suffix)).SetOutTradeNo(fmt.Sprintf("BIZREFORDER%d", suffix)).SetPaymentType("alipay").SetPaymentTradeNo(fmt.Sprintf("biz-ref-%d", suffix)).SetStatus("COMPLETED").SetClientIP("127.0.0.1").SetSrcHost("localhost").SetExpiresAt(now.Add(time.Hour)).SetPaidAt(now).Save(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, user.ID)
	})

	for _, status := range []string{"REFUND_REQUESTED", "REFUNDING", "REFUND_PENDING", "REFUND_FAILED"} {
		_, err = integrationDB.ExecContext(ctx, `UPDATE payment_orders SET status=$2,refund_amount=100 WHERE id=$1`, order.ID, status)
		require.NoError(t, err)
		var gross, refunded float64
		require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT gross_amount,refunded_amount FROM business_payment_facts WHERE payment_order_id=$1`, order.ID).Scan(&gross, &refunded))
		require.Equal(t, 100.0, gross)
		require.Zero(t, refunded, status+" must not recognize a refund before completion")
	}

	_, err = integrationDB.ExecContext(ctx, `UPDATE payment_orders SET status='REFUNDED',refund_amount=100,refund_at=$2 WHERE id=$1`, order.ID, now.Add(time.Minute))
	require.NoError(t, err)
	var refunded float64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT refunded_amount FROM business_payment_facts WHERE payment_order_id=$1`, order.ID).Scan(&refunded))
	require.Equal(t, 100.0, refunded)
	var firstPaid *time.Time
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT first_paid_at FROM business_user_facts WHERE user_id=$1`, user.ID).Scan(&firstPaid))
	require.Nil(t, firstPaid, "a fully refunded order must not remain a first-payment or repurchase baseline")
}

func TestBusinessPaymentFactsUseConsistentPaidStatuses(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	for _, status := range []string{"PAID", "RECHARGING", "COMPLETED", "FAILED"} {
		t.Run(strings.ToLower(status), func(t *testing.T) {
			user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-status-%s-%d@example.com", strings.ToLower(status), time.Now().UnixNano())})
			t.Cleanup(func() {
				_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, user.ID)
			})
			orderID := createBusinessPaymentOrder(t, user, status, &now, "CNY", 10)
			var firstPaid time.Time
			require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT first_paid_at FROM business_user_facts WHERE user_id=$1`, user.ID).Scan(&firstPaid))
			require.WithinDuration(t, now, firstPaid, time.Second)
			var facts int
			require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM business_payment_facts WHERE payment_order_id=$1`, orderID).Scan(&facts))
			require.Equal(t, 1, facts)
		})
	}

	t.Run("unpaid_failed", func(t *testing.T) {
		user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-unpaid-failed-%d@example.com", time.Now().UnixNano())})
		t.Cleanup(func() {
			_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, user.ID)
		})
		orderID := createBusinessPaymentOrder(t, user, "FAILED", nil, "CNY", 10)
		var facts int
		require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM business_payment_facts WHERE payment_order_id=$1`, orderID).Scan(&facts))
		require.Zero(t, facts)
	})
}

func TestBusinessPaymentFXUsesConfiguredRateAndDoesNotInventUnknownCurrencyRate(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	_, err := integrationDB.ExecContext(ctx, `INSERT INTO distribution_fx_rates(currency,rate_to_cny) VALUES('EUR',8.25) ON CONFLICT(currency) DO UPDATE SET rate_to_cny=EXCLUDED.rate_to_cny`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM distribution_fx_rates WHERE currency IN ('EUR','JPY')`)
	})

	for _, tc := range []struct {
		currency  string
		rate      *float64
		estimated bool
	}{
		{currency: "EUR", rate: func() *float64 { v := 8.25; return &v }()},
		{currency: "JPY", rate: nil, estimated: true},
	} {
		t.Run(strings.ToLower(tc.currency), func(t *testing.T) {
			user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-fx-%s-%d@example.com", strings.ToLower(tc.currency), time.Now().UnixNano())})
			t.Cleanup(func() {
				_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, user.ID)
			})
			orderID := createBusinessPaymentOrder(t, user, "COMPLETED", &now, tc.currency, 100)
			var rate *float64
			var estimated bool
			require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT fx_rate_to_cny,fx_estimated FROM business_payment_facts WHERE payment_order_id=$1`, orderID).Scan(&rate, &estimated))
			if tc.rate == nil {
				require.Nil(t, rate)
			} else {
				require.NotNil(t, rate)
				require.InDelta(t, *tc.rate, *rate, 0.0001)
			}
			require.Equal(t, tc.estimated, estimated)
		})
	}
}

func TestBusinessAnalyticsAggregationRetriesFailedRange(t *testing.T) {
	ctx := context.Background()
	repo := NewBusinessAnalyticsRepository(integrationDB)
	now := time.Now().UTC().Truncate(time.Hour)
	start, end := now.Add(-2*time.Hour), now.Add(-time.Hour)
	_, err := integrationDB.ExecContext(ctx, `UPDATE business_analytics_aggregation_state SET pending_from=NULL WHERE id=1;
		CREATE OR REPLACE FUNCTION test_fail_business_bucket() RETURNS trigger AS $$ BEGIN RAISE EXCEPTION 'forced business bucket failure'; END $$ LANGUAGE plpgsql;
		CREATE TRIGGER test_fail_business_bucket BEFORE INSERT OR UPDATE ON business_analytics_buckets FOR EACH STATEMENT EXECUTE FUNCTION test_fail_business_bucket()`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DROP TRIGGER IF EXISTS test_fail_business_bucket ON business_analytics_buckets; DROP FUNCTION IF EXISTS test_fail_business_bucket(); UPDATE business_analytics_aggregation_state SET pending_from=NULL WHERE id=1`)
	})
	require.Error(t, repo.AggregateRange(ctx, start, end))
	var pending time.Time
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT pending_from FROM business_analytics_aggregation_state WHERE id=1`).Scan(&pending))
	require.WithinDuration(t, start, pending, time.Second)
	_, err = integrationDB.ExecContext(ctx, `DROP TRIGGER test_fail_business_bucket ON business_analytics_buckets; DROP FUNCTION test_fail_business_bucket()`)
	require.NoError(t, err)
	require.NoError(t, repo.AggregateRange(ctx, end, now))
	var pendingAfter *time.Time
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT pending_from FROM business_analytics_aggregation_state WHERE id=1`).Scan(&pendingAfter))
	require.Nil(t, pendingAfter)
}

func TestBusinessAnalyticsInvalidatesHistoricalPaymentBucket(t *testing.T) {
	ctx := context.Background()
	repo := NewBusinessAnalyticsRepository(integrationDB)
	historicalPaidAt := time.Now().UTC().AddDate(0, 0, -10).Truncate(time.Hour)
	user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-invalidation-%d@example.com", time.Now().UnixNano())})
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, user.ID)
		_, _ = integrationDB.ExecContext(context.Background(), `UPDATE business_analytics_aggregation_state SET pending_from=NULL WHERE id=1`)
	})
	_, err := integrationDB.ExecContext(ctx, `UPDATE business_analytics_aggregation_state SET pending_from=NULL WHERE id=1`)
	require.NoError(t, err)
	orderID := createBusinessPaymentOrder(t, user, "COMPLETED", &historicalPaidAt, "CNY", 100)
	_, err = integrationDB.ExecContext(ctx, `UPDATE business_analytics_aggregation_state SET pending_from=NULL WHERE id=1`)
	require.NoError(t, err)

	refundAt := time.Now().UTC()
	_, err = integrationDB.ExecContext(ctx, `UPDATE payment_orders SET status='REFUNDED',refund_amount=100,refund_at=$2 WHERE id=$1`, orderID, refundAt)
	require.NoError(t, err)
	var pending time.Time
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT pending_from FROM business_analytics_aggregation_state WHERE id=1`).Scan(&pending))
	require.WithinDuration(t, historicalPaidAt, pending, time.Second, "a late refund must rebuild the original paid bucket as well as its refund bucket")

	require.NoError(t, repo.AggregateRange(ctx, refundAt.Add(-time.Hour), refundAt.Add(time.Hour)))
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	paidBucket := historicalPaidAt.In(location).Truncate(time.Hour)
	var payingUsers int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT (metrics->>'paying_users')::bigint FROM business_analytics_buckets WHERE resolution='hour' AND scope='channel' AND scope_id='unknown' AND bucket_start=$1`, paidBucket).Scan(&payingUsers))
	require.Zero(t, payingUsers, "the rebuilt historical bucket must agree with realtime effective-payment KPIs")
}

func TestBusinessAnalyticsRetryBacklogAdvancesByBoundedChunk(t *testing.T) {
	ctx := context.Background()
	repo := NewBusinessAnalyticsRepository(integrationDB)
	end := time.Now().UTC().Truncate(time.Hour)
	pending := end.Add(-72 * time.Hour)
	_, err := integrationDB.ExecContext(ctx, `UPDATE business_analytics_aggregation_state SET pending_from=$1,invalidation_version=invalidation_version+1 WHERE id=1`, pending)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `UPDATE business_analytics_aggregation_state SET pending_from=NULL WHERE id=1`)
	})

	require.NoError(t, repo.AggregateRange(ctx, end.Add(-time.Hour), end))
	var next time.Time
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT pending_from FROM business_analytics_aggregation_state WHERE id=1`).Scan(&next))
	require.WithinDuration(t, pending.Add(businessAnalyticsRetryChunk), next, time.Second)
}

func TestBusinessAnalyticsDoesNotClearConcurrentInvalidation(t *testing.T) {
	ctx := context.Background()
	repo := NewBusinessAnalyticsRepository(integrationDB)
	now := time.Now().UTC().Truncate(time.Hour)
	concurrentInvalidation := now.Add(-30 * 24 * time.Hour)
	_, err := integrationDB.ExecContext(ctx, `UPDATE business_analytics_aggregation_state SET pending_from=NULL WHERE id=1;
		CREATE OR REPLACE FUNCTION test_invalidate_while_aggregating() RETURNS trigger AS $$
		BEGIN PERFORM business_invalidate_analytics('`+concurrentInvalidation.Format(time.RFC3339)+`'::timestamptz); RETURN NULL; END $$ LANGUAGE plpgsql;
		CREATE TRIGGER test_invalidate_while_aggregating AFTER INSERT OR UPDATE ON business_analytics_buckets FOR EACH STATEMENT EXECUTE FUNCTION test_invalidate_while_aggregating()`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DROP TRIGGER IF EXISTS test_invalidate_while_aggregating ON business_analytics_buckets; DROP FUNCTION IF EXISTS test_invalidate_while_aggregating(); UPDATE business_analytics_aggregation_state SET pending_from=NULL WHERE id=1`)
	})

	require.NoError(t, repo.AggregateRange(ctx, now.Add(-time.Hour), now))
	var pending time.Time
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT pending_from FROM business_analytics_aggregation_state WHERE id=1`).Scan(&pending))
	require.WithinDuration(t, concurrentInvalidation, pending, time.Second)
}

func TestBusinessAnalyticsChannelsAlwaysReturnAllCategories(t *testing.T) {
	repo := NewBusinessAnalyticsRepository(integrationDB)
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	end := time.Now().In(location)
	start := end.Add(-24 * time.Hour)
	rows, err := repo.channels(context.Background(), service.BusinessAnalyticsFilter{
		DateFrom: start, DateTo: end, Timezone: "Asia/Shanghai",
	}, 7.2)
	require.NoError(t, err)
	require.Equal(t, []string{"distribution", "affiliate", "campaign", "organic", "unknown"}, []string{
		rows[0].Channel, rows[1].Channel, rows[2].Channel, rows[3].Channel, rows[4].Channel,
	})
}

func TestBusinessAnalyticsLifecycleIsExclusiveAndHonorsThirtyDayBoundary(t *testing.T) {
	ctx := context.Background()
	repo := NewBusinessAnalyticsRepository(integrationDB)
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	periodStart := time.Date(2026, 8, 1, 0, 0, 0, 0, location)
	periodEnd := periodStart.AddDate(0, 0, 7)
	previousStart := periodStart.AddDate(0, 0, -7)
	filter := service.BusinessAnalyticsFilter{
		DateFrom: periodStart, DateTo: periodEnd, ReportDateTo: periodEnd.Add(-24 * time.Hour), Timezone: "Asia/Shanghai",
	}
	before, err := repo.lifecycle(ctx, filter)
	require.NoError(t, err)
	suffix := time.Now().UnixNano()
	account := mustCreateAccount(t, integrationEntClient, &service.Account{Name: fmt.Sprintf("business-lifecycle-%d", suffix)})
	createdUsers := make([]int64, 0, 8)
	createUserWithUsage := func(name string, createdAt time.Time, usageTimes ...time.Time) {
		user := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("business-lifecycle-%s-%d@example.com", name, suffix), CreatedAt: createdAt})
		createdUsers = append(createdUsers, user.ID)
		key := mustCreateApiKey(t, integrationEntClient, &service.APIKey{UserID: user.ID, Key: fmt.Sprintf("sk-lifecycle-%s-%d", name, suffix), Name: name})
		usageRepo := newUsageLogRepositoryWithSQL(integrationEntClient, integrationDB)
		for index, at := range usageTimes {
			_, err := usageRepo.Create(ctx, &service.UsageLog{UserID: user.ID, APIKeyID: key.ID, AccountID: account.ID, RequestID: fmt.Sprintf("lifecycle-%s-%d-%d", name, suffix, index), Model: "test-model", TotalCost: 1, ActualCost: 1, CreatedAt: at})
			require.NoError(t, err)
		}
	}
	createUserWithUsage("new", periodStart.Add(time.Hour), periodStart.Add(2*time.Hour))
	createUserWithUsage("unactivated", periodStart.AddDate(0, 0, -60))
	createUserWithUsage("newly-active", periodStart.AddDate(0, 0, -60), periodStart.Add(time.Hour))
	createUserWithUsage("continuous", periodStart.AddDate(0, 0, -60), previousStart.Add(time.Hour), periodStart.Add(time.Hour))
	firstRecall := periodStart.Add(2 * time.Hour)
	createUserWithUsage("silent-recalled", periodStart.AddDate(0, 0, -90), firstRecall.AddDate(0, 0, -30), firstRecall)
	createUserWithUsage("churned-recalled", periodStart.AddDate(0, 0, -120), firstRecall.AddDate(0, 0, -30).Add(-time.Second), firstRecall)
	createUserWithUsage("silent", periodStart.AddDate(0, 0, -90), periodEnd.AddDate(0, 0, -30))
	createUserWithUsage("churned", periodStart.AddDate(0, 0, -120), periodEnd.AddDate(0, 0, -30).Add(-time.Second))
	require.NoError(t, repo.AggregateRange(ctx, periodStart.AddDate(0, 0, -40), periodEnd))
	t.Cleanup(func() {
		for _, userID := range createdUsers {
			_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id=$1`, userID)
		}
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM accounts WHERE id=$1`, account.ID)
	})

	result, err := repo.lifecycle(ctx, filter)
	require.NoError(t, err)
	require.Equal(t, service.BusinessLifecycle{
		New: 1, Unactivated: 1, NewlyActivated: 1, ContinuouslyActive: 1,
		SilentReactivated: 1, ChurnedReactivated: 1, Silent: 1, Churned: 1,
	}, service.BusinessLifecycle{
		New:                result.New - before.New,
		Unactivated:        result.Unactivated - before.Unactivated,
		NewlyActivated:     result.NewlyActivated - before.NewlyActivated,
		ContinuouslyActive: result.ContinuouslyActive - before.ContinuouslyActive,
		SilentReactivated:  result.SilentReactivated - before.SilentReactivated,
		ChurnedReactivated: result.ChurnedReactivated - before.ChurnedReactivated,
		Silent:             result.Silent - before.Silent,
		Churned:            result.Churned - before.Churned,
	})
}

func TestBusinessAnalyticsRepositoryExecutesEverySection(t *testing.T) {
	repo := NewBusinessAnalyticsRepository(integrationDB)
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	now := time.Now().In(location)
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	to := now
	if !to.After(from) {
		to = from.Add(time.Minute)
	}

	require.NoError(t, repo.AggregateRange(context.Background(), from.Add(-time.Hour), to))
	for _, section := range []string{"overview", "growth", "finance", "retention", "channels"} {
		t.Run(section, func(t *testing.T) {
			result, err := repo.GetBusinessAnalytics(context.Background(), service.BusinessAnalyticsFilter{
				DateFrom: from, DateTo: to, ReportDateTo: from, Timezone: "Asia/Shanghai",
				Granularity: "hour", Comparison: true, Section: section,
			})
			require.NoError(t, err)
			require.NotNil(t, result)
		})
	}
	for _, scope := range []string{"direct", "team"} {
		t.Run("agent_"+scope, func(t *testing.T) {
			result, err := repo.GetBusinessAnalytics(context.Background(), service.BusinessAnalyticsFilter{
				DateFrom: from, DateTo: to, ReportDateTo: from, Timezone: "Asia/Shanghai",
				Granularity: "hour", Comparison: true, Section: "overview", AgentID: 999999, AgentScope: scope,
			})
			require.NoError(t, err)
			require.NotNil(t, result)

			balance, err := repo.GetBusinessBalance(context.Background(), 7, 999999, scope)
			require.NoError(t, err)
			require.NotNil(t, balance)
		})
	}

	balance, err := repo.GetBusinessBalance(context.Background(), 7, 0, "")
	require.NoError(t, err)
	require.NotNil(t, balance)
	require.Len(t, balance.Segments, 4)
}
