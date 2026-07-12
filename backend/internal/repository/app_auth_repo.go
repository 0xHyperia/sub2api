package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type appAuthorizationRepository struct {
	db *sql.DB
}

func NewAppAuthorizationRepository(db *sql.DB) service.AppAuthorizationRepository {
	return &appAuthorizationRepository{db: db}
}

func (r *appAuthorizationRepository) Create(ctx context.Context, authorization *service.AppAuthorization) error {
	scopes, err := json.Marshal(authorization.Scopes)
	if err != nil {
		return fmt.Errorf("marshal app authorization scopes: %w", err)
	}
	return r.db.QueryRowContext(ctx, `
		INSERT INTO app_authorizations
			(user_id, grant_id, client_id, device_name, platform, scopes, token_family_id, status)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)
		RETURNING id, created_at, updated_at`,
		authorization.UserID, authorization.GrantID, authorization.ClientID,
		authorization.DeviceName, authorization.Platform, string(scopes),
		authorization.TokenFamilyID, authorization.Status,
	).Scan(&authorization.ID, &authorization.CreatedAt, &authorization.UpdatedAt)
}

func (r *appAuthorizationRepository) GetByGrantID(ctx context.Context, grantID string) (*service.AppAuthorization, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, grant_id, client_id, device_name, platform, scopes,
		       token_family_id, status, last_used_at, revoked_at, created_at, updated_at
		FROM app_authorizations WHERE grant_id = $1`, grantID)
	return scanAppAuthorization(row)
}

func (r *appAuthorizationRepository) ListByUserID(ctx context.Context, userID int64) ([]*service.AppAuthorization, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, grant_id, client_id, device_name, platform, scopes,
		       token_family_id, status, last_used_at, revoked_at, created_at, updated_at
		FROM app_authorizations
		WHERE user_id = $1 AND status = 'active'
		ORDER BY created_at DESC, id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*service.AppAuthorization, 0)
	for rows.Next() {
		authorization, err := scanAppAuthorization(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, authorization)
	}
	return result, rows.Err()
}

func (r *appAuthorizationRepository) RevokeByIDForUser(ctx context.Context, id, userID int64) (*service.AppAuthorization, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE app_authorizations
		SET status = 'revoked', revoked_at = COALESCE(revoked_at, NOW()), updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, grant_id, client_id, device_name, platform, scopes,
		          token_family_id, status, last_used_at, revoked_at, created_at, updated_at`, id, userID)
	return scanAppAuthorization(row)
}

func (r *appAuthorizationRepository) RevokeByGrantID(ctx context.Context, grantID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE app_authorizations
		SET status = 'revoked', revoked_at = COALESCE(revoked_at, NOW()), updated_at = NOW()
		WHERE grant_id = $1`, grantID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return service.ErrAppAuthorizationNotFound
	}
	return nil
}

func (r *appAuthorizationRepository) Touch(ctx context.Context, grantID string, usedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE app_authorizations SET last_used_at = $2, updated_at = $2
		WHERE grant_id = $1 AND status = 'active'`, grantID, usedAt)
	return err
}

func (r *appAuthorizationRepository) IsUserActive(ctx context.Context, userID int64) (bool, error) {
	var active bool
	err := r.db.QueryRowContext(ctx, `SELECT status = 'active' FROM users WHERE id = $1`, userID).Scan(&active)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return active, err
}

type appAuthorizationScanner interface {
	Scan(dest ...any) error
}

func scanAppAuthorization(scanner appAuthorizationScanner) (*service.AppAuthorization, error) {
	var authorization service.AppAuthorization
	var scopes []byte
	err := scanner.Scan(
		&authorization.ID, &authorization.UserID, &authorization.GrantID,
		&authorization.ClientID, &authorization.DeviceName, &authorization.Platform,
		&scopes, &authorization.TokenFamilyID, &authorization.Status,
		&authorization.LastUsedAt, &authorization.RevokedAt,
		&authorization.CreatedAt, &authorization.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrAppAuthorizationNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(scopes, &authorization.Scopes); err != nil {
		return nil, fmt.Errorf("decode app authorization scopes: %w", err)
	}
	return &authorization, nil
}
