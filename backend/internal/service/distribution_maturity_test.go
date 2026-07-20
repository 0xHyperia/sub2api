package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type distributionMaturityRepoStub struct {
	DistributionRepository
	released int
	err      error
	calls    int
}

func (s *distributionMaturityRepoStub) ReleaseMaturedCommissions(context.Context) (int, error) {
	s.calls++
	return s.released, s.err
}

func TestDistributionMaturityRuntimeRecordsSuccessfulSweep(t *testing.T) {
	repo := &distributionMaturityRepoStub{released: 7}
	runtime := newDistributionMaturityRuntime(repo, nil, nil, 0)
	runtime.runOnce()

	status := runtime.Status()
	require.Equal(t, 1, repo.calls)
	require.Equal(t, "success", status.LastOutcome)
	require.Equal(t, 7, status.LastReleased)
	require.NotNil(t, status.LastAttemptAt)
	require.NotNil(t, status.LastSuccessAt)
	require.Empty(t, status.LastError)
}

func TestDistributionMaturityRuntimeExposesFailure(t *testing.T) {
	repo := &distributionMaturityRepoStub{err: errors.New("database unavailable")}
	runtime := newDistributionMaturityRuntime(repo, nil, nil, 0)
	runtime.runOnce()

	status := runtime.Status()
	require.Equal(t, "error", status.LastOutcome)
	require.Contains(t, status.LastError, "database unavailable")
	require.Nil(t, status.LastSuccessAt)
}
