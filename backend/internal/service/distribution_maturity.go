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
	distributionMaturityLeaderLockKey = "distribution:commission:maturity:leader"
	distributionMaturityLeaderLockTTL = 2 * time.Minute
)

type DistributionMaturityStatus struct {
	Running       bool       `json:"running"`
	LastOutcome   string     `json:"last_outcome"`
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`
	LastSuccessAt *time.Time `json:"last_success_at,omitempty"`
	LastReleased  int        `json:"last_released"`
	LastError     string     `json:"last_error,omitempty"`
}

type distributionMaturityRuntime struct {
	repo       DistributionRepository
	lockCache  LeaderLockCache
	db         *sql.DB
	interval   time.Duration
	instanceID string
	stopCh     chan struct{}
	stopOnce   sync.Once
	wg         sync.WaitGroup
	mu         sync.RWMutex
	status     DistributionMaturityStatus
}

func newDistributionMaturityRuntime(repo DistributionRepository, lockCache LeaderLockCache, db *sql.DB, interval time.Duration) *distributionMaturityRuntime {
	return &distributionMaturityRuntime{repo: repo, lockCache: lockCache, db: db, interval: interval, instanceID: uuid.NewString(), stopCh: make(chan struct{})}
}

func (r *distributionMaturityRuntime) Start() {
	if r == nil || r.repo == nil || r.interval <= 0 {
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

func (r *distributionMaturityRuntime) Stop() {
	if r == nil {
		return
	}
	r.stopOnce.Do(func() { close(r.stopCh) })
	r.wg.Wait()
}

func (r *distributionMaturityRuntime) Status() DistributionMaturityStatus {
	if r == nil {
		return DistributionMaturityStatus{LastOutcome: "disabled"}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.status
}

func (r *distributionMaturityRuntime) runOnce() {
	now := time.Now().UTC()
	r.mu.Lock()
	r.status.LastAttemptAt = &now
	r.status.Running = true
	r.status.LastError = ""
	r.mu.Unlock()

	lockCtx, lockCancel := context.WithTimeout(context.Background(), 2*time.Second)
	release, ok := tryAcquireSingletonLeaderLock(lockCtx, r.lockCache, r.db, distributionMaturityLeaderLockKey, r.instanceID, distributionMaturityLeaderLockTTL)
	lockCancel()
	if !ok {
		r.mu.Lock()
		r.status.Running = false
		r.status.LastOutcome = "standby"
		r.mu.Unlock()
		return
	}
	defer release()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	released, err := r.repo.ReleaseMaturedCommissions(ctx)
	cancel()
	completedAt := time.Now().UTC()
	r.mu.Lock()
	r.status.Running = false
	if err != nil {
		r.status.LastOutcome = "error"
		r.status.LastError = err.Error()
	} else {
		r.status.LastOutcome = "success"
		r.status.LastSuccessAt = &completedAt
		r.status.LastReleased = released
	}
	r.mu.Unlock()
	if err != nil {
		slog.Error("distribution commission maturity sweep failed", "error", err)
		return
	}
	if released > 0 {
		slog.Info("distribution commissions matured", "released", released)
	}
}

func (s *DistributionService) Stop() {
	if s != nil && s.maturity != nil {
		s.maturity.Stop()
	}
}

func (s *DistributionService) MaturityStatus() DistributionMaturityStatus {
	if s == nil || s.maturity == nil {
		return DistributionMaturityStatus{LastOutcome: "disabled"}
	}
	return s.maturity.Status()
}
