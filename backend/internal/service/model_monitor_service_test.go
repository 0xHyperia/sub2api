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

func TestModelMonitorAcceptsAllConcretePlatforms(t *testing.T) {
	for _, platform := range []string{PlatformAnthropic, PlatformOpenAI, PlatformGemini, PlatformAntigravity, PlatformGrok,
		PlatformKimi, PlatformZhipu, PlatformDeepseek, PlatformMiniMax, PlatformOpenCodeGo} {
		require.True(t, validModelMonitorPlatform(platform), platform)
		group := Group{ID: 7, Platform: platform, Status: StatusActive}
		account := Account{Platform: platform, Credentials: map[string]any{
			"model_mapping": map[string]any{"public-model": "upstream-model"},
		}}
		service := &ModelMonitorService{
			groupRepo:   modelMonitorTestGroupRepo{groups: []Group{group}},
			accountRepo: modelMonitorTestAccountRepo{accounts: map[int64][]Account{group.ID: {account}}},
		}
		catalog, err := service.DiscoverCatalog(context.Background())
		require.NoError(t, err)
		require.Len(t, catalog, 1, platform)
		require.Equal(t, platform, catalog[0].Platform)
		require.Equal(t, "public-model", catalog[0].Model)
	}
	require.False(t, validModelMonitorPlatform(PlatformComposite))
	require.False(t, validModelMonitorPlatform("unknown"))
}

func TestModelMonitorSelectsCompatibleProviderAccounts(t *testing.T) {
	for _, platform := range []string{PlatformKimi, PlatformZhipu, PlatformDeepseek, PlatformMiniMax, PlatformOpenCodeGo} {
		t.Run(platform, func(t *testing.T) {
			account := Account{ID: 1, Platform: platform, Type: AccountTypeAPIKey,
				Status: StatusActive, Schedulable: true,
				Credentials: map[string]any{"model_mapping": map[string]any{"public-model": "upstream-model"}},
			}
			service := &ModelMonitorService{openAIGateway: &OpenAIGatewayService{
				accountRepo: stubOpenAIAccountRepo{accounts: []Account{account}},
			}}
			selected, err := service.selectAccount(context.Background(), platform, "public-model", 7, nil)
			require.NoError(t, err)
			require.Equal(t, account.ID, selected.ID)
			_, err = service.selectAccount(context.Background(), platform, "public-model", 7, map[int64]struct{}{1: {}})
			require.ErrorIs(t, err, ErrNoAvailableAccounts)
		})
	}
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

type modelMonitorTrafficCursorRepo struct {
	ModelMonitorRepository
	cursor      *time.Time
	refreshFrom time.Time
	refreshTo   time.Time
	savedCursor time.Time
	rolledHours []time.Time
}

func (r *modelMonitorTrafficCursorRepo) ClaimTrafficMetricsRefresh(context.Context, time.Time, time.Time) (*time.Time, bool, error) {
	return r.cursor, true, nil
}

func (r *modelMonitorTrafficCursorRepo) RefreshTrafficMetrics(_ context.Context, from, to time.Time) error {
	r.refreshFrom, r.refreshTo = from, to
	return nil
}

func (r *modelMonitorTrafficCursorRepo) FinishTrafficMetricsRefresh(_ context.Context, cursor time.Time) error {
	r.savedCursor = cursor
	return nil
}

func (r *modelMonitorTrafficCursorRepo) ReleaseTrafficMetricsRefresh(context.Context) error {
	return nil
}

func (r *modelMonitorTrafficCursorRepo) RollupHourlyMetrics(_ context.Context, hour time.Time) error {
	r.rolledHours = append(r.rolledHours, hour)
	return nil
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

func TestAlignedModelMonitorSlotUsesNaturalBoundaries(t *testing.T) {
	now := time.Date(2026, 8, 6, 12, 37, 42, 0, time.FixedZone("CST", 8*60*60))
	require.Equal(t, time.Date(2026, 8, 6, 4, 0, 0, 0, time.UTC), alignedModelMonitorSlot(now, 3600))
	require.Equal(t, time.Date(2026, 8, 6, 4, 35, 0, 0, time.UTC), alignedModelMonitorSlot(now, 300))
	require.Equal(t, time.Date(2026, 8, 6, 4, 37, 0, 0, time.UTC), alignedModelMonitorSlot(now, 60))
}

func TestModelMonitorSlotTrafficUsesPreviousNaturalWindow(t *testing.T) {
	slot := time.Date(2026, 8, 6, 13, 0, 0, 0, time.UTC)
	windowStart := slot.Add(-time.Hour)
	group := ModelMonitorGroupConfig{IntervalSeconds: 3600, LastTrafficAt: &windowStart}
	require.True(t, modelMonitorSlotHasTraffic(group, slot))

	oldTraffic := windowStart.Add(-time.Nanosecond)
	group.LastTrafficAt = &oldTraffic
	require.False(t, modelMonitorSlotHasTraffic(group, slot))
}

func TestModelMonitorCompensationUsesNaturalMinuteDueTime(t *testing.T) {
	next := time.Date(2026, 8, 6, 12, 1, 0, 0, time.UTC)
	group := ModelMonitorGroupConfig{
		Enabled: true, FailureCompensationEnabled: true, FailureCompensationPending: true, NextCompensationAt: &next,
	}
	require.False(t, modelMonitorCompensationDue(group, next.Add(-time.Nanosecond)))
	require.True(t, modelMonitorCompensationDue(group, next))
	group.Enabled = false
	require.False(t, modelMonitorCompensationDue(group, next))
}

func TestModelMonitorPassiveRefreshResumesFromPersistentCursorWithOverlap(t *testing.T) {
	now := time.Date(2026, 8, 9, 12, 37, 42, 0, time.UTC)
	cursor := now.Add(-20*time.Minute - 17*time.Second)
	repo := &modelMonitorTrafficCursorRepo{cursor: &cursor}
	runner := NewModelMonitorRunner(&ModelMonitorService{repo: repo}, nil)

	require.NoError(t, runner.refreshTrafficMetrics(context.Background(), now))
	require.Equal(t, cursor.Add(-3*time.Minute).Truncate(time.Minute), repo.refreshFrom)
	require.Equal(t, now.UTC(), repo.refreshTo)
	require.Equal(t, repo.refreshTo, repo.savedCursor)
}

func TestModelMonitorPassiveRefreshColdStartUsesRetentionWindow(t *testing.T) {
	now := time.Date(2026, 8, 9, 12, 37, 42, 0, time.UTC)
	repo := &modelMonitorTrafficCursorRepo{}
	runner := NewModelMonitorRunner(&ModelMonitorService{repo: repo}, nil)

	require.NoError(t, runner.refreshTrafficMetrics(context.Background(), now))
	require.Equal(t, now.Add(-ModelMonitorMinuteRetentionHours*time.Hour).Truncate(time.Minute), repo.refreshFrom)
	require.Equal(t, now.UTC(), repo.savedCursor)
	require.Len(t, repo.rolledHours, ModelMonitorMinuteRetentionHours-1)
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
		ModelAllowlist: GroupModelAllowlist{
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
		ModelAllowlist: GroupModelAllowlist{
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

func TestConfirmedUpstreamModelFailureClassification(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		confirmed bool
	}{
		{name: "upstream 500", message: "upstream request failed with HTTP 500", confirmed: true},
		{name: "upstream 503", message: `provider response: status_code":503`, confirmed: true},
		{name: "upstream 529", message: "upstream status=529", confirmed: true},
		{name: "account test format", message: "API returned 503: service unavailable", confirmed: true},
		{name: "model capacity", message: "selected model is at capacity", confirmed: true},
		{name: "overload code", message: "model_capacity_exhausted", confirmed: true},
		{name: "ordinary 429", message: "HTTP 429 rate limit exceeded", confirmed: false},
		{name: "bad gateway", message: "upstream HTTP 502 bad gateway", confirmed: false},
		{name: "gateway timeout", message: "upstream HTTP 504 gateway timeout", confirmed: false},
		{name: "network timeout", message: "dial tcp: i/o timeout", confirmed: false},
		{name: "misclassified model not found", message: "API returned 500: model not found", confirmed: false},
		{name: "misclassified context error", message: "API returned 503: context length exceeded", confirmed: false},
		{name: "misclassified auth error", message: "API returned 500: invalid API key", confirmed: false},
		{name: "misclassified network error", message: "API returned 503: TLS handshake timeout", confirmed: false},
		{name: "invalid request", message: "invalid request: context length exceeded", confirmed: false},
		{name: "model not found", message: "model not found", confirmed: false},
		{name: "balance", message: "insufficient balance", confirmed: false},
		{name: "authentication", message: "invalid API key", confirmed: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.confirmed, isConfirmedUpstreamModelFailure(tt.message))
		})
	}
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
