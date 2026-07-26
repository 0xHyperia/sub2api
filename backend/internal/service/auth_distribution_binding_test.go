package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type durableDistributionBindingRepoStub struct {
	DistributionRepository
	events    []string
	code      string
	pending   bool
	queueErr  error
	bindErr   error
	finishErr error
}

func (s *durableDistributionBindingRepoStub) QueueDistributionBindingClaim(_ context.Context, _ int64, code, _ string) error {
	s.events = append(s.events, "queue")
	if s.queueErr != nil {
		return s.queueErr
	}
	s.code, s.pending = code, true
	return nil
}

func (s *durableDistributionBindingRepoStub) ValidatePromotionCode(context.Context, string) error {
	return nil
}

func (s *durableDistributionBindingRepoStub) PendingDistributionBindingClaim(context.Context, int64) (string, bool, error) {
	s.events = append(s.events, "inspect")
	return s.code, s.pending, nil
}

func (s *durableDistributionBindingRepoStub) FinishDistributionBindingClaim(_ context.Context, _ int64, bindErr error) error {
	s.events = append(s.events, "finish")
	s.finishErr = bindErr
	if bindErr == nil {
		s.pending = false
	}
	return nil
}

func (s *durableDistributionBindingRepoStub) BindCustomerByCode(context.Context, int64, string) error {
	s.events = append(s.events, "bind")
	return s.bindErr
}

func TestDistributionBindingClaimIsDurableBeforeBindAndRetries(t *testing.T) {
	repo := &durableDistributionBindingRepoStub{bindErr: errors.New("database temporarily unavailable")}
	auth := &AuthService{distributionService: NewDistributionService(repo)}

	durable, err := auth.queueAndTryDistributionBinding(context.Background(), 42, "AGENT42", "oauth")
	require.True(t, durable)
	require.EqualError(t, err, "database temporarily unavailable")
	require.Equal(t, []string{"queue", "bind", "finish"}, repo.events)
	require.True(t, repo.pending)
	require.Equal(t, err, repo.finishErr)

	repo.bindErr = nil
	auth.retryPendingDistributionBinding(context.Background(), 42)
	require.Equal(t, []string{"queue", "bind", "finish", "inspect", "bind", "finish"}, repo.events)
	require.False(t, repo.pending)
	require.NoError(t, repo.finishErr)
}

func TestDistributionBindingClaimQueueFailureIsNotDurable(t *testing.T) {
	queueErr := errors.New("claim queue unavailable")
	repo := &durableDistributionBindingRepoStub{queueErr: queueErr}
	auth := &AuthService{distributionService: NewDistributionService(repo)}

	durable, err := auth.queueAndTryDistributionBinding(context.Background(), 42, "AGENT42", "oauth")

	require.False(t, durable)
	require.ErrorIs(t, err, queueErr)
	require.Equal(t, []string{"queue"}, repo.events)
}

func TestBindOAuthPromotionPropagatesOnlyNonDurableFailure(t *testing.T) {
	queueErr := errors.New("claim queue unavailable")
	repo := &durableDistributionBindingRepoStub{queueErr: queueErr}
	auth := &AuthService{distributionService: NewDistributionService(repo)}

	err := auth.bindOAuthPromotion(context.Background(), 42, "", "AGENT42")
	require.ErrorIs(t, err, queueErr)

	repo = &durableDistributionBindingRepoStub{bindErr: errors.New("binding temporarily unavailable")}
	auth.distributionService = NewDistributionService(repo)
	err = auth.bindOAuthPromotion(context.Background(), 42, "", "AGENT42")
	require.NoError(t, err)
	require.Equal(t, []string{"queue", "bind", "finish"}, repo.events)
	require.True(t, repo.pending)
}
