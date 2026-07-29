//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAppAuthorizationRepositoryReusesGrantAndInstallationSession(t *testing.T) {
	ctx := context.Background()
	user, err := integrationEntClient.User.Create().
		SetEmail("app-oauth-repository@example.com").
		SetPasswordHash("unused-test-password-hash").
		Save(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = integrationDB.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID) })
	intruder, err := integrationEntClient.User.Create().
		SetEmail("app-oauth-repository-intruder@example.com").
		SetPasswordHash("unused-test-password-hash").
		Save(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = integrationDB.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, intruder.ID) })

	repo := NewAppAuthorizationRepository(integrationDB)
	now := time.Now().UTC().Truncate(time.Microsecond)
	firstGrant, err := repo.AuthorizeGrant(ctx, user.ID, "zeroagent-desktop", []string{"profile:read"}, "grant-one", now)
	require.NoError(t, err)
	firstSession, err := repo.ActivateSession(ctx, firstGrant.GrantID, "installation-hash", "Workstation", "windows", []string{"profile:read"}, "session-one", "family-one", now)
	require.NoError(t, err)

	secondGrant, err := repo.AuthorizeGrant(ctx, user.ID, "zeroagent-desktop", []string{"offline_access", "profile:read"}, "grant-two-must-not-win", now.Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, firstGrant.ID, secondGrant.ID)
	require.Equal(t, firstGrant.GrantID, secondGrant.GrantID)
	require.ElementsMatch(t, []string{"offline_access", "profile:read"}, secondGrant.Scopes)

	secondSession, err := repo.ActivateSession(ctx, secondGrant.GrantID, "installation-hash", "Workstation", "windows", []string{"offline_access", "profile:read"}, "session-two", "family-two", now.Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, firstSession.ID, secondSession.ID)
	require.NotEqual(t, firstSession.SessionID, secondSession.SessionID)
	require.NotEqual(t, firstSession.TokenFamilyID, secondSession.TokenFamilyID)

	_, err = repo.ActivateSession(ctx, secondGrant.GrantID, "other-installation-hash", "Laptop", "linux", []string{"profile:read"}, "session-three", "family-three", now.Add(2*time.Minute))
	require.NoError(t, err)
	grants, err := repo.ListGrantsByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, grants, 1)
	require.Equal(t, 2, grants[0].SessionCount)

	sessions, err := repo.ListSessionsByGrantIDForUser(ctx, firstGrant.ID, user.ID, now.Add(-time.Hour))
	require.NoError(t, err)
	require.Len(t, sessions, 2)
	intruderSessions, err := repo.ListSessionsByGrantIDForUser(ctx, firstGrant.ID, intruder.ID, now.Add(-time.Hour))
	require.NoError(t, err)
	require.Empty(t, intruderSessions)
	_, err = repo.UpdateSessionNameByIDForUser(ctx, secondSession.ID, intruder.ID, "Not allowed")
	require.ErrorIs(t, err, service.ErrAppAuthorizationNotFound)
	_, err = repo.RevokeSessionByIDForUser(ctx, secondSession.ID, intruder.ID)
	require.ErrorIs(t, err, service.ErrAppAuthorizationNotFound)
	_, err = repo.RevokeOtherSessionsByGrantIDForUser(ctx, firstGrant.ID, secondSession.ID, intruder.ID)
	require.ErrorIs(t, err, service.ErrAppAuthorizationNotFound)
	secondSession, err = repo.UpdateSessionNameByIDForUser(ctx, secondSession.ID, user.ID, "Renamed workstation")
	require.NoError(t, err)
	require.Equal(t, "Renamed workstation", secondSession.DeviceName)

	revokedCount, err := repo.RevokeOtherSessionsByGrantIDForUser(ctx, firstGrant.ID, secondSession.ID, user.ID)
	require.NoError(t, err)
	require.EqualValues(t, 1, revokedCount)
	sessions, err = repo.ListSessionsByGrantIDForUser(ctx, firstGrant.ID, user.ID, now.Add(-time.Hour))
	require.NoError(t, err)
	require.Len(t, sessions, 1)

	_, err = repo.RevokeSessionByIDForUser(ctx, secondSession.ID, user.ID)
	require.NoError(t, err)
	grants, err = repo.ListGrantsByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 0, grants[0].SessionCount)

	revoked, err := repo.RevokeGrantByIDForUser(ctx, firstGrant.ID, user.ID)
	require.NoError(t, err)
	require.Equal(t, "revoked", revoked.Status)
	grants, err = repo.ListGrantsByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Empty(t, grants)
}
