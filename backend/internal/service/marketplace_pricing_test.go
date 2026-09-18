//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarketplacePricingPreservesTimeScheduleAndUnmultipliedBase(t *testing.T) {
	schedule := &ChannelTimePricing{Timezone: "Asia/Shanghai", WeekdaysOnly: true, Periods: []ChannelTimePricingPeriod{
		{StartTime: "09:00:00", EndTime: "18:00:00", Multiplier: 2},
	}}
	_, resolver := newTokenCostTestEnv(t, PlatformAnthropic, sonnetChannelWithTimePricing(schedule), nil)
	group := enabledGroup(PlatformAnthropic)
	group.RateMultiplier = 0.1
	pricing := resolver.MarketplacePricing(context.Background(), "claude-sonnet-4", PlatformAnthropic, group)
	require.NotNil(t, pricing)
	require.Equal(t, schedule, pricing.TimePricing)
	requirePrice(t, testPtrFloat64(2e-6), pricing.InputPrice, "input")
	group.ID = 200
	other := resolver.MarketplacePricing(context.Background(), "claude-sonnet-4", PlatformAnthropic, group)
	require.NotNil(t, other)
	require.Nil(t, other.TimePricing)
}
