package repository

import (
	"os"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"
)

func TestAdminAgentLifetimePaidUsesNetSettledAmount(t *testing.T) {
	source, err := os.ReadFile("distribution_repo.go")
	if err != nil {
		t.Fatal(err)
	}
	query := string(source)
	if !strings.Contains(query, "s.status<>'void' AND COALESCE(s.commission_base_cny,0)>0) AS paying_customer_count") {
		t.Fatal("agent lifetime paying customers must exclude void commission sources")
	}
	if !strings.Contains(query, "SUM(GREATEST(COALESCE(s.commission_base_cny,0)-s.refunded_amount_cny,0))") {
		t.Fatal("agent lifetime customer paid amount must deduct refunds")
	}
}

func TestDistributionAnalyticsMetricHelpers(t *testing.T) {
	t.Parallel()
	direct := service.DistributionBusinessMetrics{
		NewCustomers: 10, PayingCustomers: 6, PaidOrders: 8,
		CustomerPaidCNY: decimal.NewFromInt(800), CommissionCNY: decimal.NewFromInt(80),
	}
	team := service.DistributionBusinessMetrics{
		NewCustomers: 5, PayingCustomers: 2, PaidOrders: 2,
		CustomerPaidCNY: decimal.NewFromInt(200), CommissionCNY: decimal.NewFromInt(30),
	}
	total := addDistributionMetrics(direct, team)
	if !total.CustomerPaidCNY.Equal(decimal.NewFromInt(1000)) || !total.CommissionCNY.Equal(decimal.NewFromInt(110)) {
		t.Fatalf("unexpected totals: %+v", total)
	}
	if !total.ConversionRate.Equal(decimal.NewFromFloat(53.33)) {
		t.Fatalf("conversion rate = %s, want 53.33", total.ConversionRate)
	}
	if !total.AverageOrderCNY.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("average order = %s, want 100", total.AverageOrderCNY)
	}
}

func TestDistributionGrowthWithoutPreviousValueIsNotComparable(t *testing.T) {
	t.Parallel()
	if value := distributionGrowth(decimal.NewFromInt(100), decimal.Zero); value != nil {
		t.Fatalf("expected nil growth for zero baseline, got %s", value.String())
	}
	value := distributionGrowth(decimal.NewFromInt(120), decimal.NewFromInt(100))
	if value == nil || !value.Equal(decimal.NewFromInt(20)) {
		t.Fatalf("expected 20%% growth, got %v", value)
	}
}

func TestDistributionTrendResolution(t *testing.T) {
	t.Parallel()
	tests := []struct {
		days       int
		resolution string
		interval   string
	}{
		{1, "hour", "1 hour"},
		{2, "day", "1 day"},
		{31, "day", "1 day"},
		{32, "week", "1 week"},
		{92, "week", "1 week"},
		{93, "month", "1 month"},
		{366, "month", "1 month"},
	}
	for _, test := range tests {
		resolution, interval := distributionTrendResolution(test.days)
		if resolution != test.resolution || interval != test.interval {
			t.Fatalf("distributionTrendResolution(%d) = (%q, %q), want (%q, %q)", test.days, resolution, interval, test.resolution, test.interval)
		}
	}
}

func TestTeamAgentAnalyticsRequiresDirectParentRelationship(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("distribution_repo.go")
	if err != nil {
		t.Fatal(err)
	}
	query := string(source)
	if !strings.Contains(query, "child.id=$2 AND parent.user_id=$1 AND child.status<>'revoked'") {
		t.Fatal("team agent analytics must be scoped to a direct child of the authenticated parent agent")
	}
}

func TestAdminCustomerListIncludesRegistrationTime(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("distribution_repo.go")
	if err != nil {
		t.Fatal(err)
	}
	query := string(source)
	if !strings.Contains(query, "COALESCE(u.username,''),u.created_at") {
		t.Fatal("admin customer list must select the platform registration time")
	}
}
