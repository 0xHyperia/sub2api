package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type businessAnalyticsRepoStub struct {
	filter            BusinessAnalyticsFilter
	agentDepth        int
	balanceAgentID    int64
	balanceAgentScope string
}

func (s *businessAnalyticsRepoStub) GetBusinessAnalytics(_ context.Context, filter BusinessAnalyticsFilter) (*BusinessAnalyticsSnapshot, error) {
	s.filter = filter
	return &BusinessAnalyticsSnapshot{Granularity: filter.Granularity}, nil
}

func (s *businessAnalyticsRepoStub) GetBusinessBalance(_ context.Context, activityWindowDays int, agentID int64, agentScope string) (*BusinessBalanceSnapshot, error) {
	s.balanceAgentID = agentID
	s.balanceAgentScope = agentScope
	return &BusinessBalanceSnapshot{ActivityWindowDays: activityWindowDays}, nil
}

func (s *businessAnalyticsRepoStub) GetBusinessAgentDepth(_ context.Context, _ int64) (int, error) {
	if s.agentDepth == 0 {
		return 1, nil
	}
	return s.agentDepth, nil
}

func TestBusinessAnalyticsServiceNormalizesRangeAndGranularity(t *testing.T) {
	repo := &businessAnalyticsRepoStub{}
	svc := NewBusinessAnalyticsService(repo)
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	result, err := svc.Get(context.Background(), BusinessAnalyticsFilter{
		DateFrom: time.Date(2026, 8, 1, 0, 0, 0, 0, location),
		DateTo:   time.Date(2026, 8, 13, 0, 0, 0, 0, location),
		Channel:  " DISTRIBUTION ",
	})
	require.NoError(t, err)
	require.Equal(t, "day", result.Granularity)
	require.Equal(t, "distribution", repo.filter.Channel)
	require.Empty(t, repo.filter.Section)
	require.Equal(t, "Asia/Shanghai", repo.filter.Timezone)
	require.Equal(t, time.Date(2026, 8, 13, 0, 0, 0, 0, location), repo.filter.ReportDateTo)
	require.False(t, repo.filter.DateTo.Before(time.Date(2026, 8, 13, 0, 0, 0, 0, location)))
	require.False(t, repo.filter.DateTo.After(time.Date(2026, 8, 14, 0, 0, 0, 0, location)))
}

func TestBusinessAnalyticsServiceValidatesBalanceWindow(t *testing.T) {
	repo := &businessAnalyticsRepoStub{}
	svc := NewBusinessAnalyticsService(repo)
	result, err := svc.GetBalance(context.Background(), 7, 0, "")
	require.NoError(t, err)
	require.Equal(t, 7, result.ActivityWindowDays)
	_, err = svc.GetBalance(context.Background(), 14, 0, "")
	require.Error(t, err)
}

func TestBusinessAnalyticsServiceNormalizesAgentScope(t *testing.T) {
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	filter := BusinessAnalyticsFilter{DateFrom: time.Now().In(location), DateTo: time.Now().In(location), AgentID: 9, AgentScope: "direct", Channel: "distribution"}
	repo := &businessAnalyticsRepoStub{agentDepth: 1}
	_, err := NewBusinessAnalyticsService(repo).Get(context.Background(), filter)
	require.NoError(t, err)
	require.Equal(t, "direct", repo.filter.AgentScope)
	require.Empty(t, repo.filter.Channel)

	repo = &businessAnalyticsRepoStub{agentDepth: 2}
	filter.AgentScope = "team"
	_, err = NewBusinessAnalyticsService(repo).Get(context.Background(), filter)
	require.NoError(t, err)
	require.Equal(t, "direct", repo.filter.AgentScope)
}

func TestBusinessAnalyticsServiceDefaultsLevelOneToTeamAndClearsScopeWithoutAgent(t *testing.T) {
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	filter := BusinessAnalyticsFilter{DateFrom: time.Now().In(location), DateTo: time.Now().In(location), AgentID: 9}
	repo := &businessAnalyticsRepoStub{agentDepth: 1}
	_, err := NewBusinessAnalyticsService(repo).Get(context.Background(), filter)
	require.NoError(t, err)
	require.Equal(t, "team", repo.filter.AgentScope)

	filter.AgentID = 0
	filter.AgentScope = "direct"
	_, err = NewBusinessAnalyticsService(repo).Get(context.Background(), filter)
	require.NoError(t, err)
	require.Zero(t, repo.filter.AgentID)
	require.Empty(t, repo.filter.AgentScope)
}

func TestBusinessAnalyticsServiceRejectsInvalidAgentScope(t *testing.T) {
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	_, err := NewBusinessAnalyticsService(&businessAnalyticsRepoStub{agentDepth: 1}).Get(context.Background(), BusinessAnalyticsFilter{
		DateFrom: time.Now().In(location), DateTo: time.Now().In(location), AgentID: 9, AgentScope: "children",
	})
	require.Error(t, err)
}

func TestBusinessAnalyticsServicePassesNormalizedAgentScopeToBalance(t *testing.T) {
	repo := &businessAnalyticsRepoStub{agentDepth: 2}
	_, err := NewBusinessAnalyticsService(repo).GetBalance(context.Background(), 7, 9, "team")
	require.NoError(t, err)
	require.Equal(t, int64(9), repo.balanceAgentID)
	require.Equal(t, "direct", repo.balanceAgentScope)
}

func TestBusinessAnalyticsServiceNormalizesSection(t *testing.T) {
	repo := &businessAnalyticsRepoStub{}
	svc := NewBusinessAnalyticsService(repo)
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	_, err := svc.Get(context.Background(), BusinessAnalyticsFilter{
		DateFrom: time.Date(2026, 8, 1, 0, 0, 0, 0, location),
		DateTo:   time.Date(2026, 8, 13, 0, 0, 0, 0, location),
		Section:  " OVERVIEW ",
	})
	require.NoError(t, err)
	require.Equal(t, "overview", repo.filter.Section)

	_, err = svc.Get(context.Background(), BusinessAnalyticsFilter{
		DateFrom: time.Date(2026, 8, 1, 0, 0, 0, 0, location),
		DateTo:   time.Date(2026, 8, 13, 0, 0, 0, 0, location),
		Section:  "unknown-section",
	})
	require.Error(t, err)
}

func TestBusinessAnalyticsServiceRejectsInvalidChannelAndRange(t *testing.T) {
	svc := NewBusinessAnalyticsService(&businessAnalyticsRepoStub{})
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	_, err := svc.Get(context.Background(), BusinessAnalyticsFilter{DateFrom: time.Now().In(location), DateTo: time.Now().In(location), Channel: "paid-ad"})
	require.Error(t, err)
	_, err = svc.Get(context.Background(), BusinessAnalyticsFilter{DateFrom: time.Now().AddDate(-2, 0, 0), DateTo: time.Now()})
	require.Error(t, err)
}

func TestBusinessAnalyticsServiceRequestsSummaryOnly(t *testing.T) {
	repo := &businessAnalyticsRepoStub{}
	svc := NewBusinessAnalyticsService(repo)
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	_, err := svc.GetSummary(context.Background(), BusinessAnalyticsFilter{
		DateFrom: time.Date(2026, 8, 1, 0, 0, 0, 0, location),
		DateTo:   time.Date(2026, 8, 13, 0, 0, 0, 0, location),
	})
	require.NoError(t, err)
	require.True(t, repo.filter.SummaryOnly)
}

func TestBusinessAnalyticsInclusiveDateUsesLastIncludedDate(t *testing.T) {
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	require.Equal(t, "2026-07-31", BusinessAnalyticsInclusiveDate(time.Date(2026, 8, 1, 0, 0, 0, 0, location)))
	require.Equal(t, "2026-08-12", BusinessAnalyticsInclusiveDate(time.Date(2026, 8, 12, 16, 30, 0, 0, location)))
}
