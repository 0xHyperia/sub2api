package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBusinessChange(t *testing.T) {
	require.Nil(t, businessChange(10, 0))
	value := businessChange(125, 100)
	require.NotNil(t, value)
	require.Equal(t, 25.0, *value)
}

func TestBusinessBucketBoundaryAligned(t *testing.T) {
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	require.True(t, businessBucketBoundaryAligned(time.Date(2026, 8, 14, 15, 0, 0, 0, location), "hour"))
	require.False(t, businessBucketBoundaryAligned(time.Date(2026, 8, 14, 15, 23, 0, 0, location), "hour"))
	require.True(t, businessBucketBoundaryAligned(time.Date(2026, 8, 15, 0, 0, 0, 0, location), "day"))
	require.False(t, businessBucketBoundaryAligned(time.Date(2026, 8, 14, 15, 0, 0, 0, location), "day"))
	require.False(t, businessBucketBoundaryAligned(time.Now(), "week"))
}

func TestBusinessMetricComparisonSemantics(t *testing.T) {
	metric := businessMetric(125, 100, "CNY", false)
	require.Equal(t, "relative", metric.ComparisonType)
	require.NotNil(t, metric.ChangeRate)
	require.Equal(t, 25.0, *metric.ChangeRate)
	require.NotNil(t, metric.ChangeValue)
	require.Equal(t, 25.0, *metric.ChangeValue)

	turnedPositive := businessMetric(10, -5, "CNY", false)
	require.Equal(t, "turned_positive", turnedPositive.ComparisonType)
	require.Nil(t, turnedPositive.ChangeRate)

	turnedNegative := businessMetric(-2, 5, "CNY", false)
	require.Equal(t, "turned_negative", turnedNegative.ComparisonType)
	require.Nil(t, turnedNegative.ChangeRate)
}

func TestBusinessRateMetricUsesPercentagePointChange(t *testing.T) {
	metric := businessRateMetric(43.4, 30.9, false)
	require.Equal(t, "percentage_point", metric.ComparisonType)
	require.Nil(t, metric.ChangeRate)
	require.NotNil(t, metric.ChangeValue)
	require.InDelta(t, 12.5, *metric.ChangeValue, 0.0001)
}
