package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type modelMonitorTestGroupRepo struct {
	GroupRepository
	groups []Group
}

func (r modelMonitorTestGroupRepo) ListActive(context.Context) ([]Group, error) {
	return r.groups, nil
}

type modelMonitorTestAccountRepo struct {
	AccountRepository
	accounts map[int64][]Account
}

func (r modelMonitorTestAccountRepo) ListSchedulableByGroupIDAndPlatform(_ context.Context, groupID int64, _ string) ([]Account, error) {
	return r.accounts[groupID], nil
}

type modelMonitorTestRepo struct {
	ModelMonitorRepository
	configs    []ModelMonitor
	batchCalls *int
	metrics    map[int64]map[int64]ModelMonitorGroupMetrics
	summaries  map[string]ModelMonitorSummary
}

func (r modelMonitorTestRepo) List(context.Context) ([]ModelMonitor, error) {
	return r.configs, nil
}

func (r modelMonitorTestRepo) ListGroupConfigs(context.Context) (map[int64][]ModelMonitorGroupConfig, error) {
	return map[int64][]ModelMonitorGroupConfig{}, nil
}

func (r modelMonitorTestRepo) GroupMetrics(context.Context, int64, []int64, ModelMonitorResolution, time.Time) (map[int64]ModelMonitorGroupMetrics, error) {
	return map[int64]ModelMonitorGroupMetrics{}, nil
}

func (r modelMonitorTestRepo) GroupMetricsBatch(context.Context, []ModelMonitorMetricScope, ModelMonitorResolution, time.Time) (map[int64]map[int64]ModelMonitorGroupMetrics, error) {
	if r.batchCalls != nil {
		*r.batchCalls++
	}
	if r.metrics != nil {
		return r.metrics, nil
	}
	return map[int64]map[int64]ModelMonitorGroupMetrics{}, nil
}

func (r modelMonitorTestRepo) Summaries(context.Context, []ModelCatalogEntry, int) (map[string]ModelMonitorSummary, error) {
	if r.summaries != nil {
		return r.summaries, nil
	}
	return map[string]ModelMonitorSummary{}, nil
}

func TestModelMonitorCatalogModelsUsesExplicitAccountRestrictions(t *testing.T) {
	group := Group{Platform: PlatformOpenAI}
	accounts := []Account{
		{
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-5":         "gpt-5",
					"gpt-5-codex":   "gpt-5-codex",
					"gpt-internal*": "gpt-internal*",
				},
			},
		},
		{
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-5":   "gpt-5.1",
					"o3-mini": "o3-mini",
				},
			},
		},
	}

	require.Equal(t, []string{"gpt-5", "gpt-5-codex", "o3-mini"}, ModelCatalogModels(group, accounts))
}

func TestModelMonitorCatalogModelsDoesNotInferDefaultsFromUnrestrictedAccount(t *testing.T) {
	group := Group{Platform: PlatformOpenAI}
	accounts := []Account{{
		Platform:    PlatformOpenAI,
		Credentials: map[string]any{"api_key": "test"},
	}}

	require.Empty(t, ModelCatalogModels(group, accounts))
}

func TestModelMonitorCatalogModelsIntersectsGroupListWithAccountRestrictions(t *testing.T) {
	group := Group{
		Platform: PlatformOpenAI,
		ModelsListConfig: GroupModelsListConfig{
			Enabled: true,
			Models:  []string{"gpt-5", "gpt-5-codex", "o3-mini", "unsupported", " gpt-5 "},
		},
	}
	accounts := []Account{
		{
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"model_mapping": map[string]any{"gpt-5": "gpt-5", "gpt-5-*": "gpt-5-*"},
			},
		},
		{
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"model_mapping": map[string]string{"o3-mini": "o3-mini"},
			},
		},
	}

	require.Equal(t, []string{"gpt-5", "gpt-5-codex", "o3-mini"}, ModelCatalogModels(group, accounts))
}

func TestModelMonitorCatalogModelsRequiresAccountRestrictionEvenWithGroupList(t *testing.T) {
	group := Group{
		Platform: PlatformOpenAI,
		ModelsListConfig: GroupModelsListConfig{
			Enabled: true,
			Models:  []string{"gpt-5"},
		},
	}
	accounts := []Account{{Platform: PlatformOpenAI, Credentials: map[string]any{}}}

	require.Empty(t, ModelCatalogModels(group, accounts))
}

func TestModelMonitorDiscoverCatalogSkipsCompositeAliases(t *testing.T) {
	group := Group{ID: 7, Name: "Composite", Platform: PlatformComposite, Status: StatusActive}
	service := NewModelMonitorService(
		modelMonitorTestRepo{},
		modelMonitorTestGroupRepo{groups: []Group{group}},
		modelMonitorTestAccountRepo{accounts: map[int64][]Account{
			group.ID: {{Platform: PlatformOpenAI, Credentials: map[string]any{"model_mapping": map[string]any{"public-alias": "gpt-upstream"}}}},
		}},
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	catalog, err := service.DiscoverCatalog(context.Background())
	require.NoError(t, err)
	require.Empty(t, catalog)
}

func TestModelMonitorListRowsKeepsUnavailableConfigsForHistory(t *testing.T) {
	group := Group{ID: 1, Name: "OpenAI", Platform: PlatformOpenAI, Status: StatusActive}
	account := Account{
		Platform: PlatformOpenAI,
		Credentials: map[string]any{
			"model_mapping": map[string]any{"gpt-5": "gpt-5"},
		},
	}
	service := NewModelMonitorService(
		modelMonitorTestRepo{configs: []ModelMonitor{
			{ID: 10, Platform: PlatformOpenAI, Model: "gpt-5", Enabled: true},
			{ID: 11, Platform: PlatformOpenAI, Model: "gpt-4-legacy", Enabled: true},
		}},
		modelMonitorTestGroupRepo{groups: []Group{group}},
		modelMonitorTestAccountRepo{accounts: map[int64][]Account{group.ID: {account}}},
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	rows, err := service.ListRows(context.Background(), ModelMonitorResolutionMinute)

	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Equal(t, "gpt-4-legacy", rows[0].Model)
	require.True(t, rows[0].Configured)
	require.False(t, rows[0].CatalogAvailable)
	require.Equal(t, "gpt-5", rows[1].Model)
	require.True(t, rows[1].Configured)
	require.True(t, rows[1].CatalogAvailable)
}

func TestModelMonitorListRowsLoadsAllMetricsInOneBatch(t *testing.T) {
	group := Group{ID: 1, Name: "OpenAI", Platform: PlatformOpenAI, Status: StatusActive}
	account := Account{Platform: PlatformOpenAI, Credentials: map[string]any{
		"model_mapping": map[string]any{"gpt-5": "gpt-5", "gpt-4.1": "gpt-4.1"},
	}}
	batchCalls := 0
	rate := 99.0
	service := NewModelMonitorService(
		modelMonitorTestRepo{
			configs:    []ModelMonitor{{ID: 10, Platform: PlatformOpenAI, Model: "gpt-5"}, {ID: 11, Platform: PlatformOpenAI, Model: "gpt-4.1"}},
			batchCalls: &batchCalls,
			metrics: map[int64]map[int64]ModelMonitorGroupMetrics{
				10: {0: {SuccessRate: &rate}},
				11: {0: {SuccessRate: &rate}},
			},
			summaries: map[string]ModelMonitorSummary{
				ModelMonitorKey(PlatformOpenAI, "gpt-5"):   {},
				ModelMonitorKey(PlatformOpenAI, "gpt-4.1"): {},
			},
		},
		modelMonitorTestGroupRepo{groups: []Group{group}},
		modelMonitorTestAccountRepo{accounts: map[int64][]Account{group.ID: {account}}},
		nil, nil, nil, nil, nil,
	)

	rows, err := service.ListRows(context.Background(), ModelMonitorResolutionHour)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Equal(t, 1, batchCalls)
	require.InDelta(t, rate, *rows[0].Summary.Metrics.SuccessRate, 0.0001)
	require.InDelta(t, rate, *rows[1].Summary.Metrics.SuccessRate, 0.0001)
}

func TestAverageModelMonitorMetricsUsesEqualGroupWeights(t *testing.T) {
	weightedRate := 99.0
	groupOneRate, groupTwoRate := 100.0, 50.0
	groupOneLatency, groupTwoLatency := 100.0, 300.0
	groupOneTTFT, groupTwoTTFT := 40.0, 80.0
	groupOneTPS, groupTwoTPS := 20.0, 60.0
	groupOneBucketRate, groupTwoBucketRate := 100.0, 50.0
	groupOneBucketTTFT, groupTwoBucketTTFT := 30.0, 90.0
	requests, successes, failures := int64(101), int64(100), int64(1)
	probeCost := 0.25
	startedAt := time.Date(2026, time.August, 6, 0, 0, 0, 0, time.UTC)
	metrics, ok := averageModelMonitorMetrics(map[int64]ModelMonitorGroupMetrics{
		0: {SuccessRate: &weightedRate, RequestCount: &requests, SuccessCount: &successes, FailureCount: &failures, ProbeCost: &probeCost, Buckets: []ModelMonitorMetricBucket{{StartedAt: startedAt, SuccessRate: &weightedRate}}},
		1: {SuccessRate: &groupOneRate, AverageLatencyMs: &groupOneLatency, TTFTMs: &groupOneTTFT, TPS: &groupOneTPS, Buckets: []ModelMonitorMetricBucket{{StartedAt: startedAt, SuccessRate: &groupOneBucketRate, TTFTMs: &groupOneBucketTTFT}}},
		2: {SuccessRate: &groupTwoRate, AverageLatencyMs: &groupTwoLatency, TTFTMs: &groupTwoTTFT, TPS: &groupTwoTPS, Buckets: []ModelMonitorMetricBucket{{StartedAt: startedAt, SuccessRate: &groupTwoBucketRate, TTFTMs: &groupTwoBucketTTFT}}},
	}, []int64{1, 2})

	require.True(t, ok)
	require.InDelta(t, 75.0, *metrics.SuccessRate, 0.0001)
	require.InDelta(t, 200.0, *metrics.AverageLatencyMs, 0.0001)
	require.InDelta(t, 60.0, *metrics.TTFTMs, 0.0001)
	require.InDelta(t, 40.0, *metrics.TPS, 0.0001)
	require.Equal(t, requests, *metrics.RequestCount)
	require.Equal(t, successes, *metrics.SuccessCount)
	require.Equal(t, failures, *metrics.FailureCount)
	require.InDelta(t, probeCost, *metrics.ProbeCost, 0.0001)
	require.Len(t, metrics.Buckets, 1)
	require.InDelta(t, 75.0, *metrics.Buckets[0].SuccessRate, 0.0001)
	require.InDelta(t, 60.0, *metrics.Buckets[0].TTFTMs, 0.0001)
}

func TestApplyProbeCostMultiplierUsesAccountRate(t *testing.T) {
	cost := 0.25
	multiplier := 1.6
	result := applyProbeCostMultiplier(&cost, &Account{RateMultiplier: &multiplier})
	require.NotNil(t, result)
	require.InDelta(t, 0.4, *result, 1e-12)
}

func TestProbeCostAccumulatorSumsBillableRetries(t *testing.T) {
	firstCost := 0.12
	secondCost := 0.34
	accumulator := probeCostAccumulator{known: true}
	accumulator.add(&ScheduledTestResult{InputTokens: 10}, &firstCost)
	accumulator.add(&ScheduledTestResult{OutputTokens: 20}, &secondCost)

	require.NotNil(t, accumulator.value())
	require.InDelta(t, 0.46, *accumulator.value(), 1e-12)
}

func TestProbeCostAccumulatorRejectsPartialKnownCost(t *testing.T) {
	knownCost := 0.12
	accumulator := probeCostAccumulator{known: true}
	accumulator.add(&ScheduledTestResult{InputTokens: 10}, &knownCost)
	accumulator.add(&ScheduledTestResult{OutputTokens: 20}, nil)
	accumulator.add(&ScheduledTestResult{}, nil)

	require.Nil(t, accumulator.value())
}

func TestRedactModelMonitorDetailedPerformanceKeepsSuccessRate(t *testing.T) {
	ttft := 120.0
	tps := 42.0
	latency := 900.0
	rate := 99.5
	requests, successes, failures := int64(100), int64(99), int64(1)
	summary := ModelMonitorSummary{
		Metrics: &ModelMonitorGroupMetrics{
			TPS: &tps, TTFTMs: &ttft, AverageLatencyMs: &latency, SuccessRate: &rate,
			RequestCount: &requests, SuccessCount: &successes, FailureCount: &failures,
			Buckets: []ModelMonitorMetricBucket{{StartedAt: time.Now(), SuccessRate: &rate, TTFTMs: &ttft}},
		},
	}

	RedactModelMonitorDetailedPerformance(&summary)

	require.Nil(t, summary.Metrics.TPS)
	require.Nil(t, summary.Metrics.TTFTMs)
	require.Nil(t, summary.Metrics.AverageLatencyMs)
	require.Nil(t, summary.Metrics.RequestCount)
	require.Nil(t, summary.Metrics.SuccessCount)
	require.Nil(t, summary.Metrics.FailureCount)
	require.Equal(t, rate, *summary.Metrics.SuccessRate)
	require.Nil(t, summary.Metrics.Buckets[0].TTFTMs)
	require.Equal(t, rate, *summary.Metrics.Buckets[0].SuccessRate)

	payload, err := json.Marshal(summary)
	require.NoError(t, err)
	jsonText := string(payload)
	require.NotContains(t, jsonText, `"tps"`)
	require.NotContains(t, jsonText, `"ttft_ms"`)
	require.NotContains(t, jsonText, `"average_latency_ms"`)
	require.NotContains(t, jsonText, `"probe_cost"`)
	require.NotContains(t, jsonText, `"request_count"`)
	require.NotContains(t, jsonText, `"success_count"`)
	require.NotContains(t, jsonText, `"failure_count"`)
	require.True(t, strings.Contains(jsonText, `"success_rate":99.5`))
}

func TestRedactModelMonitorSampleCountsKeepsDetailedPerformance(t *testing.T) {
	tps := 42.0
	requests, successes, failures := int64(100), int64(99), int64(1)
	metrics := ModelMonitorGroupMetrics{
		TPS: &tps, RequestCount: &requests, SuccessCount: &successes, FailureCount: &failures,
	}
	summary := ModelMonitorSummary{
		Metrics:       &metrics,
		HourlyMetrics: &metrics,
		Groups:        []ModelMonitorPublicGroupMetrics{{GroupID: 1, Name: "Default", Metrics: metrics}},
	}

	RedactModelMonitorSampleCounts(&summary)

	require.Equal(t, tps, *summary.Metrics.TPS)
	require.Nil(t, summary.Metrics.RequestCount)
	require.Nil(t, summary.HourlyMetrics.SuccessCount)
	require.Nil(t, summary.Groups[0].Metrics.FailureCount)
}
