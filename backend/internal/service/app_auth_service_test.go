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
	authorization *AppAuthorization
	active        bool
}

func (r *appAuthRepoStub) Create(_ context.Context, authorization *AppAuthorization) error {
	r.authorization = authorization
	return nil
}
func (r *appAuthRepoStub) GetByGrantID(context.Context, string) (*AppAuthorization, error) {
	if r.authorization == nil {
		return nil, ErrAppAuthorizationNotFound
	}
	return r.authorization, nil
}
func (r *appAuthRepoStub) ListByUserID(context.Context, int64) ([]*AppAuthorization, error) {
	return nil, nil
}
func (r *appAuthRepoStub) RevokeByIDForUser(context.Context, int64, int64) (*AppAuthorization, error) {
	return r.authorization, nil
}
func (r *appAuthRepoStub) RevokeByGrantID(context.Context, string) error {
	if r.authorization == nil {
		return ErrAppAuthorizationNotFound
	}
	r.authorization.Status = "revoked"
	return nil
}
func (r *appAuthRepoStub) Touch(context.Context, string, time.Time) error    { return nil }
func (r *appAuthRepoStub) IsUserActive(context.Context, int64) (bool, error) { return r.active, nil }

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
		ResponseType: "code", ClientID: "zerobox-desktop",
		RedirectURI: "http://127.0.0.1:43123/oauth/callback",
		Scope:       "openid profile:read offline_access", State: "state",
		CodeChallenge: challenge, CodeChallengeMethod: "S256",
	})
	require.NoError(t, err)
	require.Equal(t, []string{"offline_access", "openid", "profile:read"}, request.Scopes)

	_, err = svc.CreateAuthorizationRequest(context.Background(), AuthorizationRequestInput{
		ResponseType: "code", ClientID: "zerobox-desktop",
		RedirectURI: "http://localhost:43123/oauth/callback", Scope: "profile:read",
		State: "state", CodeChallenge: challenge, CodeChallengeMethod: "S256",
	})
	var oauthErr *OAuthError
	require.True(t, errors.As(err, &oauthErr))
	require.Equal(t, "invalid_request", oauthErr.Code)
}

func TestCreateAuthorizationRequestAcceptsExactMobileRedirects(t *testing.T) {
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	for _, clientID := range []string{"zerobox-android", "zerobox-ios"} {
		t.Run(clientID, func(t *testing.T) {
			svc := NewAppAuthService(&appAuthRepoStub{}, &appAuthCacheStub{}, "website-secret")
			request, err := svc.CreateAuthorizationRequest(context.Background(), AuthorizationRequestInput{
				ResponseType: "code", ClientID: clientID, RedirectURI: "zerobox://oauth/callback",
				Scope: "profile:read", State: "state", CodeChallenge: challenge, CodeChallengeMethod: "S256",
			})
			require.NoError(t, err)
			require.Equal(t, "zerobox://oauth/callback", request.RedirectURI)
		})
	}
}

func TestCreateAuthorizationRequestRejectsMobileRedirectVariants(t *testing.T) {
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	redirects := []string{
		"https://usa0.top/app/zerobox/callback",
		"zerobox://oauth/callback/",
		"zerobox://oauth/callback?next=allowed",
		"ZeroBox://oauth/callback",
	}

	for _, redirectURI := range redirects {
		t.Run(redirectURI, func(t *testing.T) {
			svc := NewAppAuthService(&appAuthRepoStub{}, &appAuthCacheStub{}, "website-secret")
			_, err := svc.CreateAuthorizationRequest(context.Background(), AuthorizationRequestInput{
				ResponseType: "code", ClientID: "zerobox-android", RedirectURI: redirectURI,
				Scope: "profile:read", State: "state", CodeChallenge: challenge, CodeChallengeMethod: "S256",
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
		ResponseType: "code", ClientID: "zerobox-android", RedirectURI: "zerobox://oauth/callback",
		Scope: "profile:read", State: string(make([]byte, MaxAppAuthStateLength+1)),
		CodeChallenge: base64.RawURLEncoding.EncodeToString(sum[:]), CodeChallengeMethod: "S256",
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
	cache.code = &AuthorizationCode{UserID: 9, ClientID: "zerobox-desktop", RedirectURI: "http://127.0.0.1:43123/oauth/callback", Scopes: []string{"profile:read"}, CodeChallenge: base64.RawURLEncoding.EncodeToString(sum[:]), GrantID: "grant", FamilyID: "family"}
	repo.authorization = &AppAuthorization{UserID: 9, ClientID: "zerobox-desktop", GrantID: "grant", Status: "active", Scopes: []string{"profile:read"}}

	response, err := svc.ExchangeAuthorizationCode(context.Background(), "zerobox-desktop", "code", "http://127.0.0.1:43123/oauth/callback", verifier)
	require.NoError(t, err)
	require.NotEmpty(t, response.AccessToken)
	require.Empty(t, response.RefreshToken)
	claims, err := svc.ValidateAccessToken(context.Background(), response.AccessToken)
	require.NoError(t, err)
	require.Equal(t, []string{"profile:read"}, claims.Scope)
	_, err = svc.ExchangeAuthorizationCode(context.Background(), "zerobox-desktop", "code", "http://127.0.0.1:43123/oauth/callback", verifier)
	require.Error(t, err)
}

func TestAuthorizationCodeExchangeRevalidatesMobileRedirect(t *testing.T) {
	cache := &appAuthCacheStub{code: &AuthorizationCode{
		UserID: 9, ClientID: "zerobox-android", RedirectURI: "zerobox://oauth/callback",
		Scopes: []string{"profile:read"}, GrantID: "grant", FamilyID: "family",
	}}
	svc := NewAppAuthService(&appAuthRepoStub{}, cache, "website-secret")

	_, err := svc.ExchangeAuthorizationCode(context.Background(), "zerobox-android", "code", "https://usa0.top/app/zerobox/callback", "verifier")
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
		UserID: 9, ClientID: "zerobox-desktop", GrantID: "grant", FamilyID: "family",
		Scopes: []string{"offline_access", "profile:read"},
	}
	cache.code = &AuthorizationCode{
		UserID: 9, ClientID: record.ClientID, RedirectURI: "http://127.0.0.1:43123/oauth/callback",
		Scopes: record.Scopes, CodeChallenge: base64.RawURLEncoding.EncodeToString(sum[:]), GrantID: record.GrantID, FamilyID: record.FamilyID,
	}
	repo.authorization = &AppAuthorization{UserID: 9, ClientID: record.ClientID, GrantID: record.GrantID, Status: "active", Scopes: record.Scopes}

	initial, err := svc.ExchangeAuthorizationCode(context.Background(), record.ClientID, "code", cache.code.RedirectURI, verifier)
	require.NoError(t, err)
	require.NotEmpty(t, initial.RefreshToken)
	require.Equal(t, record.GrantID, cache.refresh.GrantID)

	cache.rotateResult, cache.rotateRecord = RefreshRotationSucceeded, record
	rotated, err := svc.Refresh(context.Background(), record.ClientID, initial.RefreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, rotated.AccessToken)
	require.NotEmpty(t, rotated.RefreshToken)

	cache.rotateResult = RefreshRotationReused
	reused, err := svc.Refresh(context.Background(), record.ClientID, initial.RefreshToken)
	require.Nil(t, reused)
	var oauthErr *OAuthError
	require.ErrorAs(t, err, &oauthErr)
	require.Equal(t, "invalid_grant", oauthErr.Code)
	require.Equal(t, "revoked", repo.authorization.Status)
	require.True(t, cache.revoked)
}
