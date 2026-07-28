package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	ExecutionTargetSwitchPurpose = "execution_target_switch"
	ExecutionStepUpProofTTL      = 60 * time.Second
	executionStepUpMaxAttempts   = 5
)

var (
	ErrExecutionStepUpPurposeInvalid  = infraerrors.BadRequest("STEP_UP_PURPOSE_INVALID", "unsupported step-up purpose")
	ErrExecutionStepUpTargetRequired  = infraerrors.BadRequest("STEP_UP_TARGET_REQUIRED", "target fingerprint is required")
	ErrExecutionStepUpProofInvalid    = infraerrors.Unauthorized("STEP_UP_PROOF_INVALID", "step-up proof is invalid or expired")
	ErrExecutionStepUpTargetMismatch  = infraerrors.Forbidden("STEP_UP_TARGET_MISMATCH", "step-up proof does not match the requested target")
	ErrExecutionStepUpTooManyAttempts = infraerrors.TooManyRequests(
		"STEP_UP_TOO_MANY_ATTEMPTS",
		"too many password verification attempts, please try again later",
	)
	ErrExecutionStepUpPasswordNotSet = infraerrors.Conflict(
		"STEP_UP_PASSWORD_NOT_SET",
		"bind an email address and set a password before switching execution targets",
	)
)

type ExecutionStepUpGrant struct {
	UserID            int64     `json:"user_id"`
	Purpose           string    `json:"purpose"`
	TargetFingerprint string    `json:"target_fingerprint"`
	ExpiresAt         time.Time `json:"expires_at"`
}

type ExecutionStepUpCache interface {
	StoreExecutionProof(ctx context.Context, proofHash string, grant ExecutionStepUpGrant, ttl time.Duration) error
	ConsumeExecutionProof(ctx context.Context, proofHash string) (*ExecutionStepUpGrant, error)
	GetExecutionStepUpAttempts(ctx context.Context, userID int64) (int, error)
	IncrementExecutionStepUpAttempts(ctx context.Context, userID int64) (int, error)
	ClearExecutionStepUpAttempts(ctx context.Context, userID int64) error
}

type ExecutionStepUpIssueResult struct {
	Proof     string
	ExpiresAt time.Time
}

type ExecutionStepUpService struct {
	userRepo UserRepository
	cache    ExecutionStepUpCache
	now      func() time.Time
}

func NewExecutionStepUpService(userRepo UserRepository, cache ExecutionStepUpCache) *ExecutionStepUpService {
	return &ExecutionStepUpService{userRepo: userRepo, cache: cache, now: time.Now}
}

func (s *ExecutionStepUpService) Issue(
	ctx context.Context,
	userID int64,
	password string,
	purpose string,
	targetFingerprint string,
) (*ExecutionStepUpIssueResult, error) {
	purpose, targetFingerprint, err := validateExecutionStepUpBinding(purpose, targetFingerprint)
	if err != nil {
		return nil, err
	}

	attempts, err := s.cache.GetExecutionStepUpAttempts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get execution step-up attempts: %w", err)
	}
	if attempts >= executionStepUpMaxAttempts {
		return nil, ErrExecutionStepUpTooManyAttempts
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user for execution step-up: %w", err)
	}
	if strings.TrimSpace(user.PasswordHash) == "" {
		return nil, ErrExecutionStepUpPasswordNotSet
	}
	if password == "" {
		return nil, ErrPasswordRequired
	}
	if !user.CheckPassword(password) {
		count, incrementErr := s.cache.IncrementExecutionStepUpAttempts(ctx, userID)
		if incrementErr != nil {
			return nil, fmt.Errorf("increment execution step-up attempts: %w", incrementErr)
		}
		slog.Warn("execution_step_up_password_rejected", "user_id", userID, "attempts", count)
		if count >= executionStepUpMaxAttempts {
			return nil, ErrExecutionStepUpTooManyAttempts
		}
		return nil, ErrPasswordIncorrect
	}

	if err := s.cache.ClearExecutionStepUpAttempts(ctx, userID); err != nil {
		return nil, fmt.Errorf("clear execution step-up attempts: %w", err)
	}
	return s.issueProof(ctx, userID, purpose, targetFingerprint)
}

// IssueAuthorized issues a proof for an app token that already carries the
// execution:authorize scope. Password verification remains website-only.
func (s *ExecutionStepUpService) IssueAuthorized(ctx context.Context, userID int64, purpose, targetFingerprint string) (*ExecutionStepUpIssueResult, error) {
	purpose, targetFingerprint, err := validateExecutionStepUpBinding(purpose, targetFingerprint)
	if err != nil {
		return nil, err
	}
	return s.issueProof(ctx, userID, purpose, targetFingerprint)
}

func (s *ExecutionStepUpService) issueProof(ctx context.Context, userID int64, purpose, targetFingerprint string) (*ExecutionStepUpIssueResult, error) {
	proof, err := newExecutionStepUpProof()
	if err != nil {
		return nil, fmt.Errorf("generate execution step-up proof: %w", err)
	}
	expiresAt := s.now().UTC().Add(ExecutionStepUpProofTTL)
	grant := ExecutionStepUpGrant{
		UserID:            userID,
		Purpose:           purpose,
		TargetFingerprint: targetFingerprint,
		ExpiresAt:         expiresAt,
	}
	if err := s.cache.StoreExecutionProof(ctx, hashExecutionStepUpProof(proof), grant, ExecutionStepUpProofTTL); err != nil {
		return nil, fmt.Errorf("store execution step-up proof: %w", err)
	}
	slog.Info("execution_step_up_issued", "user_id", userID, "purpose", purpose, "target_fingerprint", targetFingerprint)
	return &ExecutionStepUpIssueResult{Proof: proof, ExpiresAt: expiresAt}, nil
}

func (s *ExecutionStepUpService) Consume(
	ctx context.Context,
	userID int64,
	proof string,
	purpose string,
	targetFingerprint string,
) (*ExecutionStepUpGrant, error) {
	purpose, targetFingerprint, err := validateExecutionStepUpBinding(purpose, targetFingerprint)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(proof) == "" {
		return nil, ErrExecutionStepUpProofInvalid
	}

	grant, err := s.cache.ConsumeExecutionProof(ctx, hashExecutionStepUpProof(proof))
	if err != nil {
		return nil, fmt.Errorf("consume execution step-up proof: %w", err)
	}
	if grant == nil || grant.UserID != userID || !grant.ExpiresAt.After(s.now().UTC()) {
		return nil, ErrExecutionStepUpProofInvalid
	}
	if grant.Purpose != purpose || grant.TargetFingerprint != targetFingerprint {
		return nil, ErrExecutionStepUpTargetMismatch
	}
	slog.Info("execution_step_up_consumed", "user_id", userID, "purpose", purpose, "target_fingerprint", targetFingerprint)
	return grant, nil
}

func validateExecutionStepUpBinding(purpose, targetFingerprint string) (string, string, error) {
	purpose = strings.TrimSpace(purpose)
	targetFingerprint = strings.TrimSpace(targetFingerprint)
	if purpose != ExecutionTargetSwitchPurpose {
		return "", "", ErrExecutionStepUpPurposeInvalid
	}
	if targetFingerprint == "" || len(targetFingerprint) > 512 {
		return "", "", ErrExecutionStepUpTargetRequired
	}
	return purpose, targetFingerprint, nil
}

func newExecutionStepUpProof() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashExecutionStepUpProof(proof string) string {
	sum := sha256.Sum256([]byte(proof))
	return hex.EncodeToString(sum[:])
}
