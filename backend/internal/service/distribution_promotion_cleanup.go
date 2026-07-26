package service

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	distributionPromotionCleanupLeaderLockKey = "distribution:promotion:cleanup:leader"
	distributionPromotionCleanupLeaderLockTTL = 3 * time.Minute
)

type distributionPromotionCleanupRuntime struct {
	repo                   DistributionRepository
	lockCache              LeaderLockCache
	db                     *sql.DB
	interval               time.Duration
	batchSize              int
	physicalCleanupEnabled bool
	instanceID             string
	stopCh                 chan struct{}
	stopOnce               sync.Once
	wg                     sync.WaitGroup
}

func newDistributionPromotionCleanupRuntime(repo DistributionRepository, lockCache LeaderLockCache, db *sql.DB, interval time.Duration, batchSize int, physicalCleanupEnabled bool) *distributionPromotionCleanupRuntime {
	return &distributionPromotionCleanupRuntime{
		repo: repo, lockCache: lockCache, db: db, interval: interval, batchSize: batchSize,
		physicalCleanupEnabled: physicalCleanupEnabled,
		instanceID:             uuid.NewString(), stopCh: make(chan struct{}),
	}
}

func (r *distributionPromotionCleanupRuntime) Start() {
	if r == nil || r.repo == nil || r.interval <= 0 || r.batchSize <= 0 {
		return
	}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		r.runOnce()
		for {
			select {
			case <-ticker.C:
				r.runOnce()
			case <-r.stopCh:
				return
			}
		}
	}()
}

func (r *distributionPromotionCleanupRuntime) Stop() {
	if r == nil {
		return
	}
	r.stopOnce.Do(func() { close(r.stopCh) })
	r.wg.Wait()
}

func (r *distributionPromotionCleanupRuntime) runOnce() {
	lockCtx, lockCancel := context.WithTimeout(context.Background(), 2*time.Second)
	release, ok := tryAcquireSingletonLeaderLock(lockCtx, r.lockCache, r.db, distributionPromotionCleanupLeaderLockKey, r.instanceID, distributionPromotionCleanupLeaderLockTTL)
	lockCancel()
	if !ok {
		return
	}
	defer release()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	scrubResult, err := r.repo.ScrubPromotionDetails(ctx, r.batchSize)
	if err == nil && r.physicalCleanupEnabled {
		var cleanupResult DistributionPromotionCleanupResult
		cleanupResult, err = r.repo.CleanupPromotionDetails(ctx, r.batchSize)
		if err == nil && (cleanupResult.DeletedRows > 0 || cleanupResult.SkippedDays > 0) {
			slog.Info("distribution promotion physical cleanup completed",
				"deleted_rows", cleanupResult.DeletedRows,
				"skipped_days", cleanupResult.SkippedDays)
		}
	}
	cancel()
	if err != nil {
		slog.Error("distribution promotion privacy maintenance failed", "error", err, "batch_size", r.batchSize,
			"physical_cleanup_enabled", r.physicalCleanupEnabled)
		return
	}
	if scrubResult.ScrubbedRows > 0 || scrubResult.ExpiredAttributions > 0 {
		slog.Info("distribution promotion privacy scrub completed",
			"scrubbed_rows", scrubResult.ScrubbedRows,
			"expired_attributions", scrubResult.ExpiredAttributions)
	}
}
