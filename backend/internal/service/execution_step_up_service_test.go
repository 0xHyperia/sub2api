//go:build unit

package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type executionStepUpMemoryCache struct {
	mu       sync.Mutex
	grants   map[string]ExecutionStepUpGrant
	attempts map[int64]int
}

func newExecutionStepUpMemoryCache() *executionStepUpMemoryCache {
	return &executionStepUpMemoryCache{
		grants:   make(map[string]ExecutionStepUpGrant),
		attempts: make(map[int64]int),
	}
}

func (c *executionStepUpMemoryCache) StoreExecutionProof(_ context.Context, hash string, grant ExecutionStepUpGrant, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.grants[hash] = grant
	return nil
}

func (c *executionStepUpMemoryCache) ConsumeExecutionProof(_ context.Context, hash string) (*ExecutionStepUpGrant, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	grant, ok := c.grants[hash]
	if !ok {
		return nil, nil
	}
	delete(c.grants, hash)
	return &grant, nil
}

func (c *executionStepUpMemoryCache) GetExecutionStepUpAttempts(_ context.Context, userID int64) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.attempts[userID], nil
}

func (c *executionStepUpMemoryCache) IncrementExecutionStepUpAttempts(_ context.Context, userID int64) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.attempts[userID]++
	return c.attempts[userID], nil
}

func (c *executionStepUpMemoryCache) ClearExecutionStepUpAttempts(_ context.Context, userID int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.attempts, userID)
	return nil
}

func newExecutionStepUpTestService(t *testing.T) (*ExecutionStepUpService, *executionStepUpMemoryCache) {
	t.Helper()
	user := &User{ID: 42, Email: "user@example.com"}
	require.NoError(t, user.SetPassword("correct-password"))
	cache := newExecutionStepUpMemoryCache()
	return NewExecutionStepUpService(&mockUserRepo{getByIDUser: user}, cache), cache
}

func TestExecutionStepUpIssueAndConsumeOnce(t *testing.T) {
	service, _ := newExecutionStepUpTestService(t)
	ctx := context.Background()

	issued, err := service.Issue(ctx, 42, "correct-password", ExecutionTargetSwitchPurpose, "device-a:workspace-1")
	require.NoError(t, err)
	require.NotEmpty(t, issued.Proof)
	require.WithinDuration(t, time.Now().Add(ExecutionStepUpProofTTL), issued.ExpiresAt, 2*time.Second)

	grant, err := service.Consume(ctx, 42, issued.Proof, ExecutionTargetSwitchPurpose, "device-a:workspace-1")
	require.NoError(t, err)
	require.Equal(t, int64(42), grant.UserID)

	_, err = service.Consume(ctx, 42, issued.Proof, ExecutionTargetSwitchPurpose, "device-a:workspace-1")
	require.ErrorIs(t, err, ErrExecutionStepUpProofInvalid)
}

func TestExecutionStepUpRejectsTargetMismatchAndConsumesProof(t *testing.T) {
	service, _ := newExecutionStepUpTestService(t)
	ctx := context.Background()
	issued, err := service.Issue(ctx, 42, "correct-password", ExecutionTargetSwitchPurpose, "device-a:workspace-1")
	require.NoError(t, err)

	_, err = service.Consume(ctx, 42, issued.Proof, ExecutionTargetSwitchPurpose, "device-b:workspace-1")
	require.ErrorIs(t, err, ErrExecutionStepUpTargetMismatch)
	_, err = service.Consume(ctx, 42, issued.Proof, ExecutionTargetSwitchPurpose, "device-a:workspace-1")
	require.ErrorIs(t, err, ErrExecutionStepUpProofInvalid)
}

func TestExecutionStepUpRejectsExpiredProof(t *testing.T) {
	service, _ := newExecutionStepUpTestService(t)
	now := time.Date(2026, 7, 20, 8, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	issued, err := service.Issue(context.Background(), 42, "correct-password", ExecutionTargetSwitchPurpose, "web:cloud")
	require.NoError(t, err)

	service.now = func() time.Time { return now.Add(ExecutionStepUpProofTTL + time.Second) }
	_, err = service.Consume(context.Background(), 42, issued.Proof, ExecutionTargetSwitchPurpose, "web:cloud")
	require.ErrorIs(t, err, ErrExecutionStepUpProofInvalid)
}

func TestExecutionStepUpRequiresPasswordForOAuthOnlyAccount(t *testing.T) {
	service := NewExecutionStepUpService(
		&mockUserRepo{getByIDUser: &User{ID: 42, Email: "oauth@example.com"}},
		newExecutionStepUpMemoryCache(),
	)
	_, err := service.Issue(context.Background(), 42, "anything", ExecutionTargetSwitchPurpose, "web:cloud")
	require.ErrorIs(t, err, ErrExecutionStepUpPasswordNotSet)
}

func TestExecutionStepUpRateLimitsWrongPasswords(t *testing.T) {
	service, _ := newExecutionStepUpTestService(t)
	ctx := context.Background()
	for i := 0; i < executionStepUpMaxAttempts-1; i++ {
		_, err := service.Issue(ctx, 42, "wrong", ExecutionTargetSwitchPurpose, "web:cloud")
		require.ErrorIs(t, err, ErrPasswordIncorrect)
	}
	_, err := service.Issue(ctx, 42, "wrong", ExecutionTargetSwitchPurpose, "web:cloud")
	require.ErrorIs(t, err, ErrExecutionStepUpTooManyAttempts)
	_, err = service.Issue(ctx, 42, "correct-password", ExecutionTargetSwitchPurpose, "web:cloud")
	require.ErrorIs(t, err, ErrExecutionStepUpTooManyAttempts)
}
