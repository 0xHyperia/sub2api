package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

type durableDistributionBindingRepository interface {
	QueueDistributionBindingClaim(ctx context.Context, userID int64, code, signupSource string) error
	PendingDistributionBindingClaim(ctx context.Context, userID int64) (string, bool, error)
	FinishDistributionBindingClaim(ctx context.Context, userID int64, bindErr error) error
}

// queueAndTryDistributionBinding records the registration-time claim before
// applying it. A true durable result means a failed bind is recoverable.
func (s *AuthService) queueAndTryDistributionBinding(ctx context.Context, userID int64, code, signupSource string) (bool, error) {
	code = strings.TrimSpace(code)
	if s == nil || s.distributionService == nil || userID <= 0 || code == "" {
		return true, nil
	}
	repo, durable := s.distributionService.repo.(durableDistributionBindingRepository)
	if durable {
		if err := repo.QueueDistributionBindingClaim(ctx, userID, code, signupSource); err != nil {
			return false, err
		}
	}
	bindErr := s.distributionService.BindCustomerByCode(ctx, userID, code)
	if durable {
		finishErr := repo.FinishDistributionBindingClaim(ctx, userID, bindErr)
		if finishErr != nil {
			return true, finishErr
		}
	}
	return durable || bindErr == nil, bindErr
}

func (s *AuthService) retryPendingDistributionBinding(ctx context.Context, userID int64) {
	if s == nil || s.distributionService == nil || userID <= 0 {
		return
	}
	repo, ok := s.distributionService.repo.(durableDistributionBindingRepository)
	if !ok {
		return
	}
	code, pending, err := repo.PendingDistributionBindingClaim(ctx, userID)
	if err != nil || !pending {
		if err != nil {
			logger.LegacyPrintf("service.auth", "[Auth] Failed to inspect pending distribution binding for user %d: %v", userID, err)
		}
		return
	}
	bindErr := s.distributionService.BindCustomerByCode(ctx, userID, code)
	if errors.Is(bindErr, ErrDistributionAlreadyBound) {
		bindErr = nil
	}
	if err = repo.FinishDistributionBindingClaim(ctx, userID, bindErr); err != nil {
		logger.LegacyPrintf("service.auth", "[Auth] Failed to persist distribution binding retry for user %d: %v", userID, err)
		return
	}
	if bindErr != nil {
		logger.LegacyPrintf("service.auth", "[Auth] Pending distribution binding retry failed for user %d: %v", userID, bindErr)
	}
}
