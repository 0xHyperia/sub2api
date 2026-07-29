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

func (r *appAuthorizationRepository) AuthorizeGrant(ctx context.Context, userID int64, clientID string, scopes []string, proposedGrantID string, now time.Time) (*service.AppGrant, error) {
	encodedScopes, err := json.Marshal(scopes)
	if err != nil {
		return nil, fmt.Errorf("marshal app grant scopes: %w", err)
	}
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO app_oauth_grants
			(user_id, grant_id, client_id, scopes, status, grant_version, first_authorized_at, last_authorized_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4::jsonb, 'active', 1, $5, $5, $5, $5)
		ON CONFLICT (user_id, client_id) DO UPDATE SET
			grant_id = CASE WHEN app_oauth_grants.status = 'revoked' THEN EXCLUDED.grant_id ELSE app_oauth_grants.grant_id END,
			scopes = CASE
				WHEN app_oauth_grants.status = 'revoked' THEN EXCLUDED.scopes
				ELSE (SELECT COALESCE(jsonb_agg(value ORDER BY value), '[]'::jsonb)
				      FROM (SELECT DISTINCT jsonb_array_elements_text(app_oauth_grants.scopes || EXCLUDED.scopes) AS value) merged)
			END,
			status = 'active',
			grant_version = CASE WHEN app_oauth_grants.status = 'revoked' THEN app_oauth_grants.grant_version + 1 ELSE app_oauth_grants.grant_version END,
			last_authorized_at = EXCLUDED.last_authorized_at,
			revoked_at = NULL,
			updated_at = EXCLUDED.updated_at
		RETURNING id, user_id, grant_id, client_id, scopes, status, grant_version,
		          first_authorized_at, last_authorized_at, revoked_at, created_at, updated_at`,
		userID, proposedGrantID, clientID, string(encodedScopes), now)
	return scanAppGrant(row, false)
}

func (r *appAuthorizationRepository) ActivateSession(ctx context.Context, grantID, installationIDHash, deviceName, platform string, scopes []string, proposedSessionID, proposedFamilyID string, now time.Time) (*service.AppSession, error) {
	encodedScopes, err := json.Marshal(scopes)
	if err != nil {
		return nil, fmt.Errorf("marshal app session scopes: %w", err)
	}
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO app_oauth_sessions
			(app_grant_id, session_id, installation_id_hash, token_family_id, device_name, platform, scopes, status, created_at, updated_at)
		SELECT id, $2, $3, $4, $5, $6, $7::jsonb, 'active', $8, $8
		FROM app_oauth_grants
		WHERE grant_id = $1 AND status = 'active'
		ON CONFLICT (app_grant_id, installation_id_hash) DO UPDATE SET
			session_id = EXCLUDED.session_id,
			token_family_id = EXCLUDED.token_family_id,
			device_name = EXCLUDED.device_name,
			platform = EXCLUDED.platform,
			scopes = EXCLUDED.scopes,
			status = 'active',
			last_used_at = NULL,
			revoked_at = NULL,
			created_at = EXCLUDED.created_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id, app_grant_id, session_id, installation_id_hash, token_family_id,
		          device_name, platform, scopes, status, last_used_at, revoked_at, created_at, updated_at`,
		grantID, proposedSessionID, installationIDHash, proposedFamilyID, deviceName, platform, string(encodedScopes), now)
	return scanAppSession(row)
}

func (r *appAuthorizationRepository) GetGrantSession(ctx context.Context, grantID, sessionID string) (*service.AppGrant, *service.AppSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT g.id, g.user_id, g.grant_id, g.client_id, g.scopes, g.status, g.grant_version,
		       g.first_authorized_at, g.last_authorized_at, g.revoked_at, g.created_at, g.updated_at,
		       s.id, s.app_grant_id, s.session_id, s.installation_id_hash, s.token_family_id,
		       s.device_name, s.platform, s.scopes, s.status, s.last_used_at, s.revoked_at, s.created_at, s.updated_at
		FROM app_oauth_grants g
		JOIN app_oauth_sessions s ON s.app_grant_id = g.id
		WHERE g.grant_id = $1 AND s.session_id = $2`, grantID, sessionID)
	return scanAppGrantSession(row)
}

func (r *appAuthorizationRepository) ListGrantsByUserID(ctx context.Context, userID int64) ([]*service.AppGrant, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.user_id, g.grant_id, g.client_id, g.scopes, g.status, g.grant_version,
		       g.first_authorized_at, g.last_authorized_at, g.revoked_at, g.created_at, g.updated_at,
		       COUNT(s.id) FILTER (WHERE s.status = 'active' AND COALESCE(s.last_used_at, s.created_at) >= NOW() - INTERVAL '30 days') AS session_count,
		       MAX(s.last_used_at) FILTER (WHERE s.status = 'active' AND COALESCE(s.last_used_at, s.created_at) >= NOW() - INTERVAL '30 days') AS last_used_at
		FROM app_oauth_grants g
		LEFT JOIN app_oauth_sessions s ON s.app_grant_id = g.id
		WHERE g.user_id = $1 AND g.status = 'active'
		GROUP BY g.id
		ORDER BY g.last_authorized_at DESC, g.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make([]*service.AppGrant, 0)
	for rows.Next() {
		grant, err := scanAppGrant(rows, true)
		if err != nil {
			return nil, err
		}
		result = append(result, grant)
	}
	return result, rows.Err()
}

func (r *appAuthorizationRepository) ListSessionsByGrantIDForUser(ctx context.Context, grantID, userID int64, activeSince time.Time) ([]*service.AppSession, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.id, s.app_grant_id, s.session_id, s.installation_id_hash, s.token_family_id,
		       s.device_name, s.platform, s.scopes, s.status, s.last_used_at, s.revoked_at, s.created_at, s.updated_at
		FROM app_oauth_sessions s
		JOIN app_oauth_grants g ON g.id = s.app_grant_id
		WHERE g.id = $1 AND g.user_id = $2 AND g.status = 'active'
		  AND s.status = 'active' AND COALESCE(s.last_used_at, s.created_at) >= $3
		ORDER BY COALESCE(s.last_used_at, s.created_at) DESC, s.id DESC`, grantID, userID, activeSince)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make([]*service.AppSession, 0)
	for rows.Next() {
		session, err := scanAppSession(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, session)
	}
	return result, rows.Err()
}

func (r *appAuthorizationRepository) UpdateSessionNameByIDForUser(ctx context.Context, sessionID, userID int64, deviceName string) (*service.AppSession, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE app_oauth_sessions s
		SET device_name = $3, updated_at = NOW()
		FROM app_oauth_grants g
		WHERE s.id = $1 AND g.id = s.app_grant_id AND g.user_id = $2
		  AND g.status = 'active' AND s.status = 'active'
		RETURNING s.id, s.app_grant_id, s.session_id, s.installation_id_hash, s.token_family_id,
		          s.device_name, s.platform, s.scopes, s.status, s.last_used_at, s.revoked_at, s.created_at, s.updated_at`,
		sessionID, userID, deviceName)
	return scanAppSession(row)
}

func (r *appAuthorizationRepository) RevokeSessionByIDForUser(ctx context.Context, sessionID, userID int64) (*service.AppSession, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE app_oauth_sessions s
		SET status = 'revoked', revoked_at = COALESCE(s.revoked_at, NOW()), updated_at = NOW()
		FROM app_oauth_grants g
		WHERE s.id = $1 AND g.id = s.app_grant_id AND g.user_id = $2
		  AND g.status = 'active' AND s.status = 'active'
		RETURNING s.id, s.app_grant_id, s.session_id, s.installation_id_hash, s.token_family_id,
		          s.device_name, s.platform, s.scopes, s.status, s.last_used_at, s.revoked_at, s.created_at, s.updated_at`,
		sessionID, userID)
	return scanAppSession(row)
}

func (r *appAuthorizationRepository) RevokeOtherSessionsByGrantIDForUser(ctx context.Context, grantID, keepSessionID, userID int64) (int64, error) {
	var keepExists bool
	if err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM app_oauth_sessions s
			JOIN app_oauth_grants g ON g.id = s.app_grant_id
			WHERE g.id = $1 AND g.user_id = $2 AND g.status = 'active'
			  AND s.id = $3 AND s.status = 'active'
		)`, grantID, userID, keepSessionID).Scan(&keepExists); err != nil {
		return 0, err
	}
	if !keepExists {
		return 0, service.ErrAppAuthorizationNotFound
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE app_oauth_sessions s
		SET status = 'revoked', revoked_at = COALESCE(s.revoked_at, NOW()), updated_at = NOW()
		FROM app_oauth_grants g
		WHERE g.id = $1 AND g.user_id = $2 AND g.id = s.app_grant_id
		  AND g.status = 'active' AND s.status = 'active' AND s.id <> $3`, grantID, userID, keepSessionID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *appAuthorizationRepository) RevokeGrantByIDForUser(ctx context.Context, id, userID int64) (*service.AppGrant, error) {
	return r.revokeGrant(ctx, `id = $1 AND user_id = $2`, id, userID)
}

func (r *appAuthorizationRepository) RevokeByGrantID(ctx context.Context, grantID string) error {
	_, err := r.revokeGrant(ctx, `grant_id = $1`, grantID)
	return err
}

func (r *appAuthorizationRepository) revokeGrant(ctx context.Context, predicate string, args ...any) (*service.AppGrant, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	query := `UPDATE app_oauth_grants SET status = 'revoked', revoked_at = COALESCE(revoked_at, NOW()), updated_at = NOW() WHERE ` + predicate + ` RETURNING id, user_id, grant_id, client_id, scopes, status, grant_version, first_authorized_at, last_authorized_at, revoked_at, created_at, updated_at`
	grant, err := scanAppGrant(tx.QueryRowContext(ctx, query, args...), false)
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE app_oauth_sessions SET status = 'revoked', revoked_at = COALESCE(revoked_at, NOW()), updated_at = NOW() WHERE app_grant_id = $1 AND status = 'active'`, grant.ID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return grant, nil
}

func (r *appAuthorizationRepository) RevokeSessionFamily(ctx context.Context, sessionID, familyID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE app_oauth_sessions
		SET status = 'revoked', revoked_at = COALESCE(revoked_at, NOW()), updated_at = NOW()
		WHERE session_id = $1 AND token_family_id = $2 AND status = 'active'`, sessionID, familyID)
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

func (r *appAuthorizationRepository) TouchSession(ctx context.Context, sessionID, familyID string, usedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE app_oauth_sessions SET last_used_at = $3, updated_at = $3
		WHERE session_id = $1 AND token_family_id = $2 AND status = 'active'`, sessionID, familyID, usedAt)
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

func scanAppGrant(scanner appAuthorizationScanner, withStats bool) (*service.AppGrant, error) {
	var grant service.AppGrant
	var scopes []byte
	dest := []any{&grant.ID, &grant.UserID, &grant.GrantID, &grant.ClientID, &scopes, &grant.Status, &grant.GrantVersion, &grant.FirstAuthorizedAt, &grant.LastAuthorizedAt, &grant.RevokedAt, &grant.CreatedAt, &grant.UpdatedAt}
	if withStats {
		dest = append(dest, &grant.SessionCount, &grant.LastUsedAt)
	}
	if err := scanner.Scan(dest...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrAppAuthorizationNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal(scopes, &grant.Scopes); err != nil {
		return nil, fmt.Errorf("decode app grant scopes: %w", err)
	}
	return &grant, nil
}

func scanAppSession(scanner appAuthorizationScanner) (*service.AppSession, error) {
	var session service.AppSession
	var scopes []byte
	if err := scanner.Scan(&session.ID, &session.AppGrantID, &session.SessionID, &session.InstallationIDHash, &session.TokenFamilyID, &session.DeviceName, &session.Platform, &scopes, &session.Status, &session.LastUsedAt, &session.RevokedAt, &session.CreatedAt, &session.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrAppAuthorizationNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal(scopes, &session.Scopes); err != nil {
		return nil, fmt.Errorf("decode app session scopes: %w", err)
	}
	return &session, nil
}

func scanAppGrantSession(scanner appAuthorizationScanner) (*service.AppGrant, *service.AppSession, error) {
	var grant service.AppGrant
	var session service.AppSession
	var grantScopes, sessionScopes []byte
	err := scanner.Scan(
		&grant.ID, &grant.UserID, &grant.GrantID, &grant.ClientID, &grantScopes, &grant.Status, &grant.GrantVersion,
		&grant.FirstAuthorizedAt, &grant.LastAuthorizedAt, &grant.RevokedAt, &grant.CreatedAt, &grant.UpdatedAt,
		&session.ID, &session.AppGrantID, &session.SessionID, &session.InstallationIDHash, &session.TokenFamilyID,
		&session.DeviceName, &session.Platform, &sessionScopes, &session.Status, &session.LastUsedAt, &session.RevokedAt, &session.CreatedAt, &session.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, service.ErrAppAuthorizationNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal(grantScopes, &grant.Scopes); err != nil {
		return nil, nil, fmt.Errorf("decode app grant scopes: %w", err)
	}
	if err := json.Unmarshal(sessionScopes, &session.Scopes); err != nil {
		return nil, nil, fmt.Errorf("decode app session scopes: %w", err)
	}
	return &grant, &session, nil
}
