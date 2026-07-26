package service

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type promotionCleanupRepoStub struct {
	DistributionRepository
	scrubCalls   atomic.Int32
	cleanupCalls atomic.Int32
	batchSize    atomic.Int32
}

func (s *promotionCleanupRepoStub) ScrubPromotionDetails(_ context.Context, batchSize int) (DistributionPromotionPrivacyScrubResult, error) {
	s.batchSize.Store(int32(batchSize))
	s.scrubCalls.Add(1)
	return DistributionPromotionPrivacyScrubResult{ScrubbedRows: 3, ExpiredAttributions: 2}, nil
}

func (s *promotionCleanupRepoStub) CleanupPromotionDetails(_ context.Context, batchSize int) (DistributionPromotionCleanupResult, error) {
	s.batchSize.Store(int32(batchSize))
	s.cleanupCalls.Add(1)
	return DistributionPromotionCleanupResult{DeletedRows: 3}, nil
}

func TestDistributionPromotionPrivacyScrubRunsWhenPhysicalCleanupDisabled(t *testing.T) {
	repo := &promotionCleanupRepoStub{}
	runtime := newDistributionPromotionCleanupRuntime(repo, &fakeLeaderLockCache{}, nil, 10*time.Millisecond, 321, false)
	runtime.Start()
	require.Eventually(t, func() bool { return repo.scrubCalls.Load() >= 2 }, time.Second, 5*time.Millisecond)
	runtime.Stop()
	require.EqualValues(t, 321, repo.batchSize.Load())
	require.Zero(t, repo.cleanupCalls.Load())
}

func TestDistributionPromotionPhysicalCleanupIsExplicitOptIn(t *testing.T) {
	repo := &promotionCleanupRepoStub{}
	runtime := newDistributionPromotionCleanupRuntime(repo, &fakeLeaderLockCache{}, nil, time.Hour, 100, true)
	runtime.runOnce()
	require.EqualValues(t, 1, repo.scrubCalls.Load())
	require.EqualValues(t, 1, repo.cleanupCalls.Load())
}

func TestDistributionPromotionCleanupRuntimeSkipsWhenPeerIsLeader(t *testing.T) {
	repo := &promotionCleanupRepoStub{}
	cache := &fakeLeaderLockCache{}
	ok, err := cache.TryAcquireLeaderLock(context.Background(), distributionPromotionCleanupLeaderLockKey, "peer", time.Minute)
	require.NoError(t, err)
	require.True(t, ok)
	runtime := newDistributionPromotionCleanupRuntime(repo, cache, nil, time.Hour, 100, true)
	runtime.runOnce()
	require.Zero(t, repo.scrubCalls.Load())
	require.Zero(t, repo.cleanupCalls.Load())
}
