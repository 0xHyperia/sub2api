package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type promotionTrackingStatusRepoStub struct {
	DistributionRepository
	enabled bool
	err     error
}

type promotionHashRotationRepoStub struct {
	DistributionRepository
	matchedHash string
	seen        []string
}

func (s *promotionHashRotationRepoStub) ResolvePromotionAttribution(_ context.Context, hash string, _ ...string) (*DistributionRegistrationAttribution, error) {
	s.seen = append(s.seen, hash)
	if hash == s.matchedHash {
		return &DistributionRegistrationAttribution{Code: "AGENT1"}, nil
	}
	return nil, nil
}

func (s *promotionTrackingStatusRepoStub) PromotionTrackingEnabled(context.Context) (bool, error) {
	return s.enabled, s.err
}

func TestValidTrackingTextUsesDatabaseLimits(t *testing.T) {
	require.True(t, validTrackingText(strings.Repeat("渠", 128), 128))
	require.False(t, validTrackingText(strings.Repeat("渠", 129), 128))
	require.False(t, validTrackingText("source\x00value", 128))
	require.False(t, validTrackingText(string([]byte{0xff}), 128))
}

func TestPromotionTrackingNormalizesPrivacySensitiveURLs(t *testing.T) {
	repo := &promotionVisitCaptureRepoStub{}
	svc := NewDistributionService(repo)
	svc.trackingHashSecrets = []string{"independent-test-secret"}
	_, err := svc.TrackPromotionVisit(context.Background(), DistributionPromotionVisitInput{
		PromotionCode: "AGENT1",
		VisitorToken:  "visitor",
		LandingPath:   "https://platform.example/register?token=secret#step",
		Referrer:      "https://Campaign.Example:8443/path?email=user@example.com",
	})
	require.NoError(t, err)
	require.Equal(t, "/register", repo.input.LandingPath)
	require.Equal(t, "campaign.example", repo.input.Referrer)

	_, err = svc.TrackPromotionVisit(context.Background(), DistributionPromotionVisitInput{
		PromotionCode: "AGENT1", VisitorToken: "visitor", LandingPath: "not-a-path", Referrer: "private-token",
	})
	require.NoError(t, err)
	require.Equal(t, "/register", repo.input.LandingPath)
	require.Empty(t, repo.input.Referrer)
}

type promotionVisitCaptureRepoStub struct {
	DistributionRepository
	input DistributionPromotionVisitInput
}

func (s *promotionVisitCaptureRepoStub) TrackPromotionVisit(_ context.Context, input DistributionPromotionVisitInput, _, _, _ string, _ bool) (*DistributionPromotionVisitResult, error) {
	s.input = input
	return &DistributionPromotionVisitResult{Tracked: true}, nil
}

func TestNormalizePromotionStatsFilterRejectsInvalidRangesAndEnums(t *testing.T) {
	now := time.Now()
	cases := []DistributionPromotionStatsFilter{
		{From: now, To: now},
		{From: now.AddDate(-2, 0, 0), To: now},
		{From: now.Add(-time.Hour), To: now, Device: "phone"},
		{From: now.Add(-time.Hour), To: now, AttributionType: "unknown"},
		{From: now.Add(-time.Hour), To: now, Source: strings.Repeat("s", 256)},
	}
	for _, filter := range cases {
		require.ErrorIs(t, normalizePromotionStatsFilter(&filter, true), ErrPromotionTrackingInput)
	}
}

func TestNormalizePromotionStatsFilterDefaultsAndCapsPagination(t *testing.T) {
	filter := DistributionPromotionStatsFilter{PageSize: 1000}
	require.NoError(t, normalizePromotionStatsFilter(&filter, true))
	require.Equal(t, 1, filter.Page)
	require.Equal(t, 20, filter.PageSize)
	require.True(t, filter.To.After(filter.From))
	require.LessOrEqual(t, filter.To.Sub(filter.From), 31*24*time.Hour)
}

func TestNormalizePromotionStatsFilterRejectsUnboundedPageOffset(t *testing.T) {
	filter := DistributionPromotionStatsFilter{Page: 100_002, PageSize: 100}
	require.ErrorIs(t, normalizePromotionStatsFilter(&filter, true), ErrPromotionTrackingInput)
}

func TestPromotionTrackingEnabledUsesOptionalRepositoryCapability(t *testing.T) {
	svc := NewDistributionService(&promotionTrackingStatusRepoStub{enabled: true})
	svc.trackingHashSecrets = []string{"independent-test-secret"}
	enabled, err := svc.PromotionTrackingEnabled(context.Background())
	require.NoError(t, err)
	require.True(t, enabled)
}

func TestPromotionTrackingFailsClosedWithoutIndependentHashSecret(t *testing.T) {
	svc := NewDistributionService(&promotionTrackingStatusRepoStub{enabled: true})
	enabled, err := svc.PromotionTrackingEnabled(context.Background())
	require.NoError(t, err)
	require.False(t, enabled)

	_, err = svc.TrackPromotionVisit(context.Background(), DistributionPromotionVisitInput{PromotionCode: "AGENT1"})
	require.ErrorContains(t, err, "promotion tracking is not configured")
	require.Empty(t, svc.hashVisitorTokenCandidates("visitor-token"))
}

func TestPromotionHashRotationWritesPrimaryAndReadsPreviousSecrets(t *testing.T) {
	repo := &promotionHashRotationRepoStub{}
	svc := NewDistributionService(repo)
	svc.trackingHashSecrets = []string{"new-secret", "old-secret"}
	token := "visitor-token"
	candidates := svc.hashVisitorTokenCandidates(token)
	require.Len(t, candidates, 2)
	require.Equal(t, candidates[0], svc.hashVisitorToken(token))
	repo.matchedHash = candidates[1]
	ctx := WithDistributionVisitorToken(context.Background(), token)
	code, err := svc.ResolveRegistrationAttribution(ctx, "", "")
	require.NoError(t, err)
	require.Equal(t, "AGENT1", code)
	require.Equal(t, candidates, repo.seen)
}

func TestPromotionAnalyticsJSONContractIncludesCohortsAndMeta(t *testing.T) {
	value := DistributionPromotionAnalytics{
		Summary: DistributionPromotionSummary{ConvertedVisitors: 2, TrackedRegistrations: 3, Registrations: 4, UntrackedDirect: 1},
		Daily:   []DistributionPromotionDailyStat{{Date: "2026-07-22", Conversions: 2, Registrations: 4}},
		Sources: []DistributionPromotionSourceStat{{Source: "直接注册", Conversions: 0, Registrations: 1}},
		Meta:    DistributionPromotionAnalyticsMeta{Cohort: "visit", AttributionDays: 30, AttributionModel: "first_touch", RawRetentionDays: 367},
	}
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	for _, field := range []string{"converted_visitors", "tracked_registrations", "untracked_direct", "conversions", "meta", "cohort", "attribution_days", "attribution_model", "raw_retention_days"} {
		require.Contains(t, string(raw), `"`+field+`"`)
	}
}
