package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type appAuthRepoStub struct {
	grant          *AppGrant
	session        *AppSession
	active         bool
	authorizeCalls int
	activateCalls  int
}

func (r *appAuthRepoStub) AuthorizeGrant(_ context.Context, userID int64, clientID string, scopes []string, proposedGrantID string, now time.Time) (*AppGrant, error) {
	r.authorizeCalls++
	if r.grant == nil || r.grant.Status == "revoked" {
		r.grant = &AppGrant{ID: 1, UserID: userID, ClientID: clientID, GrantID: proposedGrantID, Scopes: append([]string(nil), scopes...), Status: "active", GrantVersion: 1, FirstAuthorizedAt: now, LastAuthorizedAt: now}
	} else {
		r.grant.LastAuthorizedAt = now
	}
	return r.grant, nil
}
func (r *appAuthRepoStub) ActivateSession(_ context.Context, _ string, installationIDHash, deviceName, platform string, scopes []string, proposedSessionID, proposedFamilyID string, now time.Time) (*AppSession, error) {
	r.activateCalls++
	r.session = &AppSession{ID: 1, AppGrantID: 1, SessionID: proposedSessionID, InstallationIDHash: installationIDHash, TokenFamilyID: proposedFamilyID, DeviceName: deviceName, Platform: platform, Scopes: append([]string(nil), scopes...), Status: "active", CreatedAt: now, UpdatedAt: now}
	return r.session, nil
}
func (r *appAuthRepoStub) GetGrantSession(context.Context, string, string) (*AppGrant, *AppSession, error) {
	if r.grant == nil || r.session == nil {
		return nil, nil, ErrAppAuthorizationNotFound
	}
	return r.grant, r.session, nil
}
func (r *appAuthRepoStub) ListGrantsByUserID(context.Context, int64) ([]*AppGrant, error) {
	return nil, nil
}
func (r *appAuthRepoStub) ListSessionsByGrantIDForUser(context.Context, int64, int64, time.Time) ([]*AppSession, error) {
	return nil, nil
}
func (r *appAuthRepoStub) UpdateSessionNameByIDForUser(_ context.Context, _ int64, _ int64, deviceName string) (*AppSession, error) {
	if r.session == nil {
		return nil, ErrAppAuthorizationNotFound
	}
	r.session.DeviceName = deviceName
	return r.session, nil
}
func (r *appAuthRepoStub) RevokeSessionByIDForUser(context.Context, int64, int64) (*AppSession, error) {
	if r.session == nil {
		return nil, ErrAppAuthorizationNotFound
	}
	r.session.Status = "revoked"
	return r.session, nil
}
func (r *appAuthRepoStub) RevokeOtherSessionsByGrantIDForUser(context.Context, int64, int64, int64) (int64, error) {
	return 0, nil
}
func (r *appAuthRepoStub) RevokeGrantByIDForUser(context.Context, int64, int64) (*AppGrant, error) {
	return r.grant, nil
}
func (r *appAuthRepoStub) RevokeByGrantID(context.Context, string) error {
	if r.grant == nil {
		return ErrAppAuthorizationNotFound
	}
	r.grant.Status = "revoked"
	return nil
}
func (r *appAuthRepoStub) RevokeSessionFamily(context.Context, string, string) error {
	if r.session == nil {
		return ErrAppAuthorizationNotFound
	}
	r.session.Status = "revoked"
	return nil
}
func (r *appAuthRepoStub) TouchSession(context.Context, string, string, time.Time) error { return nil }
func (r *appAuthRepoStub) IsUserActive(context.Context, int64) (bool, error)             { return r.active, nil }

type appAuthCacheStub struct {
	request      *AuthorizationRequest
	code         *AuthorizationCode
	refresh      *AppRefreshTokenRecord
	rotateResult RefreshRotationResult
	rotateRecord *AppRefreshTokenRecord
	revoked      bool
}

func (c *appAuthCacheStub) PutAuthorizationRequest(_ context.Context, request *AuthorizationRequest, _ time.Duration) error {
	c.request = request
	return nil
}
func (c *appAuthCacheStub) GetAuthorizationRequest(context.Context, string) (*AuthorizationRequest, error) {
	return c.request, nil
}
func (c *appAuthCacheStub) ConsumeAuthorizationRequest(context.Context, string) (*AuthorizationRequest, error) {
	request := c.request
	c.request = nil
	if request == nil {
		return nil, ErrAppAuthCacheMiss
	}
	return request, nil
}
func (c *appAuthCacheStub) PutAuthorizationCode(_ context.Context, _ string, code *AuthorizationCode, _ time.Duration) error {
	c.code = code
	return nil
}
func (c *appAuthCacheStub) ConsumeAuthorizationCode(context.Context, string) (*AuthorizationCode, error) {
	code := c.code
	c.code = nil
	if code == nil {
		return nil, ErrAppAuthCacheMiss
	}
	return code, nil
}
func (c *appAuthCacheStub) PutRefreshToken(_ context.Context, _ string, record *AppRefreshTokenRecord, _ time.Duration) error {
	c.refresh = record
	return nil
}
func (*appAuthCacheStub) GetRefreshToken(context.Context, string) (*AppRefreshTokenRecord, error) {
	return nil, ErrAppAuthCacheMiss
}
func (c *appAuthCacheStub) RotateRefreshToken(context.Context, string, string, *AppRefreshTokenRecord, time.Duration, time.Duration) (RefreshRotationResult, *AppRefreshTokenRecord, error) {
	return c.rotateResult, c.rotateRecord, nil
}
func (c *appAuthCacheStub) SetGrantRevoked(context.Context, string, time.Duration) error {
	c.revoked = true
	return nil
}
func (c *appAuthCacheStub) IsGrantRevoked(context.Context, string) (bool, error) {
	return c.revoked, nil
}

func TestCreateAuthorizationRequestValidatesRedirectAndScopes(t *testing.T) {
	svc := NewAppAuthService(&appAuthRepoStub{}, &appAuthCacheStub{}, "website-secret")
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	request, err := svc.CreateAuthorizationRequest(context.Background(), AuthorizationRequestInput{
		ResponseType: "code", ClientID: "zeroagent-desktop",
		RedirectURI: "http://127.0.0.1:43123/oauth/callback",
		Scope:       "profile:read offline_access", State: "state",
		CodeChallenge: challenge, CodeChallengeMethod: "S256",
		InstallationID: "installation-1",
	})
	require.NoError(t, err)
	require.Equal(t, []string{"offline_access", "profile:read"}, request.Scopes)

	_, err = svc.CreateAuthorizationRequest(context.Background(), AuthorizationRequestInput{
		ResponseType: "code", ClientID: "zeroagent-desktop",
		RedirectURI: "http://localhost:43123/oauth/callback", Scope: "profile:read",
		State: "state", CodeChallenge: challenge, CodeChallengeMethod: "S256",
		InstallationID: "installation-1",
	})
	var oauthErr *OAuthError
	require.True(t, errors.As(err, &oauthErr))
	require.Equal(t, "invalid_request", oauthErr.Code)
}

func TestCreateAuthorizationRequestAcceptsExactMobileRedirects(t *testing.T) {
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	for _, clientID := range []string{"zeroagent-android"} {
		t.Run(clientID, func(t *testing.T) {
			svc := NewAppAuthService(&appAuthRepoStub{}, &appAuthCacheStub{}, "website-secret")
			request, err := svc.CreateAuthorizationRequest(context.Background(), AuthorizationRequestInput{
				ResponseType: "code", ClientID: clientID, RedirectURI: "top.usa0.zeroagent:/oauth/callback",
				Scope: "profile:read", State: "state", CodeChallenge: challenge, CodeChallengeMethod: "S256",
				InstallationID: "installation-1",
			})
			require.NoError(t, err)
			require.Equal(t, "top.usa0.zeroagent:/oauth/callback", request.RedirectURI)
		})
	}
}

func TestCreateAuthorizationRequestRejectsMobileRedirectVariants(t *testing.T) {
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	redirects := []string{
		"https://usa0.top/app/zeroagent/callback",
		"top.usa0.zeroagent:/oauth/callback/",
		"top.usa0.zeroagent:/oauth/callback?next=allowed",
		"ZeroAgent://oauth/callback",
	}

	for _, redirectURI := range redirects {
		t.Run(redirectURI, func(t *testing.T) {
			svc := NewAppAuthService(&appAuthRepoStub{}, &appAuthCacheStub{}, "website-secret")
			_, err := svc.CreateAuthorizationRequest(context.Background(), AuthorizationRequestInput{
				ResponseType: "code", ClientID: "zeroagent-android", RedirectURI: redirectURI,
				Scope: "profile:read", State: "state", CodeChallenge: challenge, CodeChallengeMethod: "S256",
				InstallationID: "installation-1",
			})
			var oauthErr *OAuthError
			require.ErrorAs(t, err, &oauthErr)
			require.Equal(t, "invalid_request", oauthErr.Code)
		})
	}
}

func TestCreateAuthorizationRequestRejectsOversizedState(t *testing.T) {
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	sum := sha256.Sum256([]byte(verifier))
	svc := NewAppAuthService(&appAuthRepoStub{}, &appAuthCacheStub{}, "website-secret")
	_, err := svc.CreateAuthorizationRequest(context.Background(), AuthorizationRequestInput{
		ResponseType: "code", ClientID: "zeroagent-android", RedirectURI: "top.usa0.zeroagent:/oauth/callback",
		Scope: "profile:read", State: string(make([]byte, MaxAppAuthStateLength+1)),
		CodeChallenge: base64.RawURLEncoding.EncodeToString(sum[:]), CodeChallengeMethod: "S256",
		InstallationID: "installation-1",
	})
	var oauthErr *OAuthError
	require.ErrorAs(t, err, &oauthErr)
	require.Equal(t, "invalid_request", oauthErr.Code)
}

func TestAuthorizationCodePKCEAndSingleUse(t *testing.T) {
	repo := &appAuthRepoStub{active: true}
	cache := &appAuthCacheStub{}
	svc := NewAppAuthService(repo, cache, "website-secret")
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	sum := sha256.Sum256([]byte(verifier))
	cache.code = &AuthorizationCode{UserID: 9, ClientID: "zeroagent-desktop", RedirectURI: "http://127.0.0.1:43123/oauth/callback", Scopes: []string{"profile:read"}, CodeChallenge: base64.RawURLEncoding.EncodeToString(sum[:]), GrantID: "grant", InstallationIDHash: "installation-hash"}
	repo.grant = &AppGrant{UserID: 9, ClientID: "zeroagent-desktop", GrantID: "grant", Status: "active", Scopes: []string{"profile:read"}}

	response, err := svc.ExchangeAuthorizationCode(context.Background(), "zeroagent-desktop", "", "code", "http://127.0.0.1:43123/oauth/callback", verifier)
	require.NoError(t, err)
	require.NotEmpty(t, response.AccessToken)
	require.Empty(t, response.RefreshToken)
	claims, err := svc.ValidateAccessToken(context.Background(), response.AccessToken)
	require.NoError(t, err)
	require.Equal(t, []string{"profile:read"}, claims.Scope)
	_, err = svc.ExchangeAuthorizationCode(context.Background(), "zeroagent-desktop", "", "code", "http://127.0.0.1:43123/oauth/callback", verifier)
	require.Error(t, err)
}

func TestAuthorizationCodeExchangeRevalidatesMobileRedirect(t *testing.T) {
	cache := &appAuthCacheStub{code: &AuthorizationCode{
		UserID: 9, ClientID: "zeroagent-android", RedirectURI: "top.usa0.zeroagent:/oauth/callback",
		Scopes: []string{"profile:read"}, GrantID: "grant", InstallationIDHash: "installation-hash",
	}}
	svc := NewAppAuthService(&appAuthRepoStub{}, cache, "website-secret")

	_, err := svc.ExchangeAuthorizationCode(context.Background(), "zeroagent-android", "", "code", "https://usa0.top/app/zeroagent/callback", "verifier")
	var oauthErr *OAuthError
	require.ErrorAs(t, err, &oauthErr)
	require.Equal(t, "invalid_request", oauthErr.Code)
	require.NotNil(t, cache.code, "an unregistered redirect must be rejected before the code is consumed")
}

func TestAppRefreshRotationAndReuseRevokesGrant(t *testing.T) {
	repo := &appAuthRepoStub{active: true}
	cache := &appAuthCacheStub{}
	svc := NewAppAuthService(repo, cache, "independent-app-signing-secret-32-bytes")
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	sum := sha256.Sum256([]byte(verifier))
	record := &AppRefreshTokenRecord{
		UserID: 9, ClientID: "zeroagent-desktop", GrantID: "grant", SessionID: "session", FamilyID: "family",
		Scopes: []string{"offline_access", "profile:read"},
	}
	cache.code = &AuthorizationCode{
		UserID: 9, ClientID: record.ClientID, RedirectURI: "http://127.0.0.1:43123/oauth/callback",
		Scopes: record.Scopes, CodeChallenge: base64.RawURLEncoding.EncodeToString(sum[:]), GrantID: record.GrantID, InstallationIDHash: "installation-hash",
	}
	repo.grant = &AppGrant{UserID: 9, ClientID: record.ClientID, GrantID: record.GrantID, Status: "active", Scopes: record.Scopes}

	initial, err := svc.ExchangeAuthorizationCode(context.Background(), record.ClientID, "", "code", cache.code.RedirectURI, verifier)
	require.NoError(t, err)
	require.NotEmpty(t, initial.RefreshToken)
	require.Equal(t, record.GrantID, cache.refresh.GrantID)

	record = cache.refresh
	cache.rotateResult, cache.rotateRecord = RefreshRotationSucceeded, record
	rotated, err := svc.Refresh(context.Background(), record.ClientID, "", initial.RefreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, rotated.AccessToken)
	require.NotEmpty(t, rotated.RefreshToken)

	cache.rotateResult = RefreshRotationReused
	reused, err := svc.Refresh(context.Background(), record.ClientID, "", initial.RefreshToken)
	require.Nil(t, reused)
	var oauthErr *OAuthError
	require.ErrorAs(t, err, &oauthErr)
	require.Equal(t, "invalid_grant", oauthErr.Code)
	require.Equal(t, "revoked", repo.session.Status)
	require.False(t, cache.revoked)
}

func TestRepeatedAuthorizationReusesGrantAndInstallationSession(t *testing.T) {
	repo := &appAuthRepoStub{active: true}
	cache := &appAuthCacheStub{}
	svc := NewAppAuthService(repo, cache, "independent-app-signing-secret-32-bytes")
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	sum := sha256.Sum256([]byte(verifier))
	input := AuthorizationRequestInput{
		ResponseType: "code", ClientID: "zeroagent-desktop", RedirectURI: "http://127.0.0.1:43123/oauth/callback",
		Scope: "profile:read offline_access", State: "state", CodeChallenge: base64.RawURLEncoding.EncodeToString(sum[:]),
		CodeChallengeMethod: "S256", InstallationID: "stable-installation",
	}

	var originalGrantID, originalSessionID, originalFamilyID string
	for attempt := 0; attempt < 2; attempt++ {
		request, err := svc.CreateAuthorizationRequest(context.Background(), input)
		require.NoError(t, err)
		decision, err := svc.DecideAuthorization(context.Background(), 9, request.RequestID, true)
		require.NoError(t, err)
		tokens, err := svc.ExchangeAuthorizationCode(context.Background(), input.ClientID, "", decision.Code, input.RedirectURI, verifier)
		require.NoError(t, err)
		require.NotEmpty(t, tokens.RefreshToken)
		if attempt == 0 {
			originalGrantID, originalSessionID, originalFamilyID = repo.grant.GrantID, repo.session.SessionID, repo.session.TokenFamilyID
		}
	}

	require.Equal(t, 2, repo.authorizeCalls)
	require.Equal(t, 2, repo.activateCalls)
	require.Equal(t, originalGrantID, repo.grant.GrantID, "application consent must be reused")
	require.NotEqual(t, originalSessionID, repo.session.SessionID, "a new login replaces the installation session")
	require.NotEqual(t, originalFamilyID, repo.session.TokenFamilyID, "a new login rotates the token family")
}

func TestExpiredInstallationSessionIsRejected(t *testing.T) {
	now := time.Now().UTC()
	repo := &appAuthRepoStub{
		active:  true,
		grant:   &AppGrant{UserID: 9, ClientID: "zeroagent-desktop", GrantID: "grant", Status: "active", Scopes: []string{"profile:read"}},
		session: &AppSession{SessionID: "session", TokenFamilyID: "family", Status: "active", Scopes: []string{"profile:read"}, CreatedAt: now.Add(-AppRefreshTokenTTL - time.Minute)},
	}
	svc := NewAppAuthService(repo, &appAuthCacheStub{}, "independent-app-signing-secret-32-bytes")
	svc.now = func() time.Time { return now }
	token, err := svc.signAccessToken(9, "zeroagent-desktop", "grant", "session", "family", []string{"profile:read"})
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(context.Background(), token)
	require.Nil(t, claims)
	var oauthErr *OAuthError
	require.ErrorAs(t, err, &oauthErr)
	require.Equal(t, "invalid_token", oauthErr.Code)
	require.Equal(t, "revoked", repo.session.Status)
}

func TestZeroAgentWebRedirectUsesExactHTTPSAllowlistAndLocalDevelopment(t *testing.T) {
	ConfigureAppAuthWebRedirectURIs([]string{
		"https://gateway.example.com/api/auth/oauth/callback",
		"https://gateway.example.com/api/auth/oauth/callback?ignored=true",
		"http://gateway.example.com/api/auth/oauth/callback",
	})
	t.Cleanup(func() { ConfigureAppAuthWebRedirectURIs(nil) })
	client, ok := AppOAuthClientByID("zeroagent-web")
	require.True(t, ok)
	tests := []struct {
		uri  string
		want bool
	}{
		{"https://gateway.example.com/api/auth/oauth/callback", true},
		{"https://gateway.example.com/api/auth/oauth/callback/", false},
		{"https://gateway.example.com/api/auth/oauth/callback?next=/", false},
		{"https://other.example.com/api/auth/oauth/callback", false},
		{"http://gateway.example.com/api/auth/oauth/callback", false},
		{"http://127.0.0.1:50052/api/auth/oauth/callback", true},
		{"http://localhost:5173/api/auth/oauth/callback", true},
		{"http://172.16.0.121:5174/api/auth/oauth/callback", true},
		{"http://127.0.0.1/api/auth/oauth/callback", false},
		{"http://127.0.0.1:50052/other", false},
		{"top.usa0.zeroagent:/oauth/callback", false},
	}
	for _, test := range tests {
		require.Equalf(t, test.want, client.ValidateRedirect(test.uri), "redirect %s", test.uri)
	}
}

func TestZeroAgentOAuthClientsHaveNoZeroBoxCompatibility(t *testing.T) {
	for _, clientID := range []string{"zerobox-web", "zerobox-desktop", "zerobox-android", "zerobox-ios"} {
		_, ok := AppOAuthClientByID(clientID)
		require.False(t, ok, clientID)
	}
	_, ok := AppOAuthClientByID("zeroagent-ios")
	require.False(t, ok)
}

func TestZeroAgentWebClientRequiresConfiguredSecret(t *testing.T) {
	const secret = "zeroagent-test-web-client-secret-32-bytes"
	svc := NewAppAuthService(&appAuthRepoStub{}, &appAuthCacheStub{}, "website-secret", secret)

	client, err := svc.authenticateClient("zeroagent-web", secret)
	require.NoError(t, err)
	require.True(t, client.Confidential)

	for _, provided := range []string{"", "wrong-secret", secret + "-suffix"} {
		_, err := svc.authenticateClient("zeroagent-web", provided)
		var oauthErr *OAuthError
		require.ErrorAs(t, err, &oauthErr)
		require.Equal(t, "invalid_client", oauthErr.Code)
	}
}

func TestZeroAgentRejectsOIDCCompatibilityScope(t *testing.T) {
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	sum := sha256.Sum256([]byte(verifier))
	svc := NewAppAuthService(&appAuthRepoStub{}, &appAuthCacheStub{}, "website-secret")
	_, err := svc.CreateAuthorizationRequest(context.Background(), AuthorizationRequestInput{
		ResponseType: "code", ClientID: "zeroagent-desktop",
		RedirectURI: "http://127.0.0.1:43123/oauth/callback", Scope: "openid profile:read",
		State: "state", CodeChallenge: base64.RawURLEncoding.EncodeToString(sum[:]), CodeChallengeMethod: "S256",
		InstallationID: "installation-1",
	})
	var oauthErr *OAuthError
	require.ErrorAs(t, err, &oauthErr)
	require.Equal(t, "invalid_scope", oauthErr.Code)
}
