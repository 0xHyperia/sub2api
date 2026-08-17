package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AppAuthIssuer           = "https://usa0.top"
	AppAuthAudience         = "https://usa0.top/api/v1/app"
	AppAuthTokenUse         = "app"
	AuthorizationRequestTTL = 5 * time.Minute
	AuthorizationCodeTTL    = time.Minute
	AppAccessTokenTTL       = 10 * time.Minute
	AppRefreshTokenTTL      = 30 * 24 * time.Hour
	MaxAppAuthStateLength   = 1024
	MaxInstallationIDLength = 128
	AppAuthTokenVersion     = 2
)

var (
	ErrAppAuthorizationNotFound = errors.New("app authorization not found")
	ErrAppAuthorizationRevoked  = errors.New("app authorization revoked")
	ErrAppAuthCacheMiss         = errors.New("app auth cache entry not found")
	ErrAppRefreshTokenReused    = errors.New("app refresh token reused")
)

// OAuthError is safe to serialize as a standard OAuth error response.
type OAuthError struct {
	Code        string
	Description string
}

func (e *OAuthError) Error() string {
	if e.Description == "" {
		return e.Code
	}
	return e.Code + ": " + e.Description
}

func oauthError(code, description string) error {
	return &OAuthError{Code: code, Description: description}
}

type AppOAuthClient struct {
	ID               string
	Platform         string
	Name             string
	Confidential     bool
	AllowedScopes    map[string]struct{}
	ValidateRedirect func(string) bool
}

var appAllowedScopes = map[string]struct{}{
	"profile:read":        {},
	"groups:read":         {},
	"keys:read":           {},
	"keys:write":          {},
	"subscriptions:read":  {},
	"profile:write":       {},
	"usage:read":          {},
	"execution:authorize": {},
	"offline_access":      {},
}

var appWebRedirects = struct {
	sync.RWMutex
	values map[string]struct{}
}{values: map[string]struct{}{}}

var appOAuthClients = map[string]AppOAuthClient{
	"zeroagent-desktop": {
		ID: "zeroagent-desktop", Platform: "desktop", Name: "ZeroAgent Desktop",
		AllowedScopes: appAllowedScopes, ValidateRedirect: validateZeroAgentLoopbackRedirect,
	},
	"zeroagent-web": {
		ID: "zeroagent-web", Platform: "web", Name: "ZeroAgent Web", Confidential: true,
		AllowedScopes: appAllowedScopes, ValidateRedirect: validateZeroAgentWebRedirect,
	},
	"zero-canvas-web": {
		ID: "zero-canvas-web", Platform: "web", Name: "ZeroCanvas Web",
		AllowedScopes: appAllowedScopes, ValidateRedirect: validateZeroCanvasWebRedirect,
	},
	"zeroagent-android": {
		ID: "zeroagent-android", Platform: "android", Name: "ZeroAgent Android",
		AllowedScopes: appAllowedScopes, ValidateRedirect: exactRedirect("top.usa0.zeroagent:/oauth/callback"),
	},
}

func ConfigureAppAuthWebRedirectURIs(values []string) {
	next := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if validConfiguredWebRedirect(value) {
			next[value] = struct{}{}
		}
	}
	appWebRedirects.Lock()
	appWebRedirects.values = next
	appWebRedirects.Unlock()
}

func AppOAuthClientByID(clientID string) (AppOAuthClient, bool) {
	client, ok := appOAuthClients[clientID]
	return client, ok
}

type AppGrant struct {
	ID                int64      `json:"id"`
	UserID            int64      `json:"user_id"`
	GrantID           string     `json:"grant_id"`
	ClientID          string     `json:"client_id"`
	Scopes            []string   `json:"scopes"`
	Status            string     `json:"status"`
	GrantVersion      int        `json:"grant_version"`
	SessionCount      int        `json:"session_count"`
	LastUsedAt        *time.Time `json:"last_used_at,omitempty"`
	FirstAuthorizedAt time.Time  `json:"first_authorized_at"`
	LastAuthorizedAt  time.Time  `json:"last_authorized_at"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type AppSession struct {
	ID                 int64      `json:"id"`
	AppGrantID         int64      `json:"app_grant_id"`
	SessionID          string     `json:"session_id"`
	InstallationIDHash string     `json:"installation_id_hash"`
	TokenFamilyID      string     `json:"token_family_id"`
	DeviceName         string     `json:"device_name"`
	Platform           string     `json:"platform"`
	Scopes             []string   `json:"scopes"`
	Status             string     `json:"status"`
	LastUsedAt         *time.Time `json:"last_used_at,omitempty"`
	RevokedAt          *time.Time `json:"revoked_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type AppAuthorizationRepository interface {
	AuthorizeGrant(ctx context.Context, userID int64, clientID string, scopes []string, proposedGrantID string, now time.Time) (*AppGrant, error)
	ActivateSession(ctx context.Context, grantID, installationIDHash, deviceName, platform string, scopes []string, proposedSessionID, proposedFamilyID string, now time.Time) (*AppSession, error)
	GetGrantSession(ctx context.Context, grantID, sessionID string) (*AppGrant, *AppSession, error)
	ListGrantsByUserID(ctx context.Context, userID int64) ([]*AppGrant, error)
	ListSessionsByGrantIDForUser(ctx context.Context, grantID, userID int64, activeSince time.Time) ([]*AppSession, error)
	UpdateSessionNameByIDForUser(ctx context.Context, sessionID, userID int64, deviceName string) (*AppSession, error)
	RevokeSessionByIDForUser(ctx context.Context, sessionID, userID int64) (*AppSession, error)
	RevokeOtherSessionsByGrantIDForUser(ctx context.Context, grantID, keepSessionID, userID int64) (int64, error)
	RevokeGrantByIDForUser(ctx context.Context, id, userID int64) (*AppGrant, error)
	RevokeByGrantID(ctx context.Context, grantID string) error
	RevokeSessionFamily(ctx context.Context, sessionID, familyID string) error
	TouchSession(ctx context.Context, sessionID, familyID string, usedAt time.Time) error
	IsUserActive(ctx context.Context, userID int64) (bool, error)
}

type AuthorizationRequest struct {
	RequestID      string   `json:"request_id"`
	ClientID       string   `json:"client_id"`
	RedirectURI    string   `json:"redirect_uri"`
	Scopes         []string `json:"scopes"`
	State          string   `json:"state"`
	CodeChallenge  string   `json:"code_challenge"`
	DeviceName     string   `json:"device_name"`
	Platform       string   `json:"platform"`
	InstallationID string   `json:"installation_id"`
}

type AuthorizationRequestInput struct {
	ResponseType        string
	ClientID            string
	RedirectURI         string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	DeviceName          string
	Platform            string
	InstallationID      string
}

type AuthorizationCode struct {
	UserID             int64    `json:"user_id"`
	ClientID           string   `json:"client_id"`
	RedirectURI        string   `json:"redirect_uri"`
	Scopes             []string `json:"scopes"`
	CodeChallenge      string   `json:"code_challenge"`
	GrantID            string   `json:"grant_id"`
	InstallationIDHash string   `json:"installation_id_hash"`
	DeviceName         string   `json:"device_name"`
	Platform           string   `json:"platform"`
}

type AppRefreshTokenRecord struct {
	UserID    int64    `json:"user_id"`
	ClientID  string   `json:"client_id"`
	Scopes    []string `json:"scopes"`
	GrantID   string   `json:"grant_id"`
	SessionID string   `json:"session_id"`
	FamilyID  string   `json:"family_id"`
}

type RefreshRotationResult int

const (
	RefreshRotationMissing RefreshRotationResult = iota
	RefreshRotationSucceeded
	RefreshRotationReused
)

type AppAuthCache interface {
	PutAuthorizationRequest(ctx context.Context, request *AuthorizationRequest, ttl time.Duration) error
	GetAuthorizationRequest(ctx context.Context, requestID string) (*AuthorizationRequest, error)
	ConsumeAuthorizationRequest(ctx context.Context, requestID string) (*AuthorizationRequest, error)
	PutAuthorizationCode(ctx context.Context, codeHash string, code *AuthorizationCode, ttl time.Duration) error
	ConsumeAuthorizationCode(ctx context.Context, codeHash string) (*AuthorizationCode, error)
	PutRefreshToken(ctx context.Context, tokenHash string, record *AppRefreshTokenRecord, ttl time.Duration) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*AppRefreshTokenRecord, error)
	RotateRefreshToken(ctx context.Context, oldTokenHash, newTokenHash string, next *AppRefreshTokenRecord, tokenTTL, tombstoneTTL time.Duration) (RefreshRotationResult, *AppRefreshTokenRecord, error)
	SetGrantRevoked(ctx context.Context, grantID string, ttl time.Duration) error
	IsGrantRevoked(ctx context.Context, grantID string) (bool, error)
}

type AppAuthService struct {
	repository      AppAuthorizationRepository
	cache           AppAuthCache
	signingKey      []byte
	webClientSecret string
	now             func() time.Time
}

func NewAppAuthService(repository AppAuthorizationRepository, cache AppAuthCache, signingSecret string, webClientSecret ...string) *AppAuthService {
	secret := ""
	if len(webClientSecret) > 0 {
		secret = strings.TrimSpace(webClientSecret[0])
	}
	return &AppAuthService{repository: repository, cache: cache, signingKey: []byte(signingSecret), webClientSecret: secret, now: time.Now}
}

func (s *AppAuthService) CreateAuthorizationRequest(ctx context.Context, input AuthorizationRequestInput) (*AuthorizationRequest, error) {
	if input.ResponseType != "code" {
		return nil, oauthError("unsupported_response_type", "response_type must be code")
	}
	client, ok := AppOAuthClientByID(input.ClientID)
	if !ok {
		return nil, oauthError("invalid_client", "unknown public client")
	}
	if err := validateRedirectURI(client, input.RedirectURI); err != nil {
		return nil, err
	}
	if input.CodeChallengeMethod != "S256" || !validPKCEChallenge(input.CodeChallenge) {
		return nil, oauthError("invalid_request", "S256 PKCE is required")
	}
	scopes, err := normalizeScopes(client, input.Scope)
	if err != nil {
		return nil, err
	}
	if input.State == "" {
		return nil, oauthError("invalid_request", "state is required")
	}
	if len(input.State) > MaxAppAuthStateLength {
		return nil, oauthError("invalid_request", "state is too long")
	}
	requestID, err := appAuthRandomOpaqueToken(32)
	if err != nil {
		return nil, fmt.Errorf("generate request id: %w", err)
	}
	request := &AuthorizationRequest{
		RequestID: requestID, ClientID: client.ID, RedirectURI: input.RedirectURI,
		Scopes: scopes, State: input.State, CodeChallenge: input.CodeChallenge,
		DeviceName: strings.TrimSpace(input.DeviceName), Platform: strings.TrimSpace(input.Platform),
		InstallationID: strings.TrimSpace(input.InstallationID),
	}
	if len(request.DeviceName) > 200 || len(request.Platform) > 40 {
		return nil, oauthError("invalid_request", "device metadata is too long")
	}
	if request.InstallationID == "" || len(request.InstallationID) > MaxInstallationIDLength {
		return nil, oauthError("invalid_request", "installation_id is required and must not exceed 128 characters")
	}
	if err := s.cache.PutAuthorizationRequest(ctx, request, AuthorizationRequestTTL); err != nil {
		return nil, fmt.Errorf("store authorization request: %w", err)
	}
	return request, nil
}

func (s *AppAuthService) GetAuthorizationContext(ctx context.Context, requestID string) (*AuthorizationRequest, error) {
	request, err := s.cache.GetAuthorizationRequest(ctx, requestID)
	if errors.Is(err, ErrAppAuthCacheMiss) {
		return nil, oauthError("invalid_request", "authorization request expired or invalid")
	}
	return request, err
}

type AuthorizationDecision struct {
	Request      *AuthorizationRequest
	Code         string
	AccessDenied bool
}

func (s *AppAuthService) DecideAuthorization(ctx context.Context, userID int64, requestID string, allow bool) (*AuthorizationDecision, error) {
	request, err := s.cache.ConsumeAuthorizationRequest(ctx, requestID)
	if errors.Is(err, ErrAppAuthCacheMiss) {
		return nil, oauthError("invalid_request", "authorization request expired or already used")
	}
	if err != nil {
		return nil, err
	}
	if !allow {
		return &AuthorizationDecision{Request: request, AccessDenied: true}, nil
	}
	proposedGrantID, err := appAuthRandomOpaqueToken(24)
	if err != nil {
		return nil, err
	}
	grant, err := s.repository.AuthorizeGrant(ctx, userID, request.ClientID, request.Scopes, proposedGrantID, s.now())
	if err != nil {
		return nil, fmt.Errorf("authorize app grant: %w", err)
	}
	code, err := appAuthRandomOpaqueToken(32)
	if err != nil {
		return nil, err
	}
	entry := &AuthorizationCode{
		UserID: userID, ClientID: request.ClientID, RedirectURI: request.RedirectURI,
		Scopes: request.Scopes, CodeChallenge: request.CodeChallenge, GrantID: grant.GrantID,
		InstallationIDHash: hashOpaqueToken(request.InstallationID), DeviceName: request.DeviceName, Platform: request.Platform,
	}
	if err := s.cache.PutAuthorizationCode(ctx, hashOpaqueToken(code), entry, AuthorizationCodeTTL); err != nil {
		return nil, fmt.Errorf("store authorization code: %w", err)
	}
	return &AuthorizationDecision{Request: request, Code: code}, nil
}

type AppTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
}

func (s *AppAuthService) ExchangeAuthorizationCode(ctx context.Context, clientID, clientSecret, code, redirectURI, verifier string) (*AppTokenResponse, error) {
	client, err := s.authenticateClient(clientID, clientSecret)
	if err != nil {
		return nil, err
	}
	if err := validateRedirectURI(client, redirectURI); err != nil {
		return nil, err
	}
	entry, err := s.cache.ConsumeAuthorizationCode(ctx, hashOpaqueToken(code))
	if errors.Is(err, ErrAppAuthCacheMiss) {
		return nil, oauthError("invalid_grant", "authorization code expired, invalid, or already used")
	}
	if err != nil {
		return nil, err
	}
	if entry.ClientID != clientID || entry.RedirectURI != redirectURI || !verifyPKCE(verifier, entry.CodeChallenge) {
		return nil, oauthError("invalid_grant", "authorization code binding mismatch")
	}
	sessionID, err := appAuthRandomOpaqueToken(24)
	if err != nil {
		return nil, err
	}
	familyID, err := appAuthRandomOpaqueToken(24)
	if err != nil {
		return nil, err
	}
	session, err := s.repository.ActivateSession(ctx, entry.GrantID, entry.InstallationIDHash, entry.DeviceName, entry.Platform, entry.Scopes, sessionID, familyID, s.now())
	if err != nil {
		return nil, fmt.Errorf("activate app session: %w", err)
	}
	return s.issueTokenPair(ctx, entry.UserID, entry.ClientID, entry.Scopes, entry.GrantID, session.SessionID, session.TokenFamilyID, false, "")
}

func (s *AppAuthService) Refresh(ctx context.Context, clientID, clientSecret, refreshToken string) (*AppTokenResponse, error) {
	if _, err := s.authenticateClient(clientID, clientSecret); err != nil {
		return nil, err
	}
	if refreshToken == "" {
		return nil, oauthError("invalid_grant", "refresh token is required")
	}
	return s.issueTokenPair(ctx, 0, clientID, nil, "", "", "", true, refreshToken)
}

func (s *AppAuthService) authenticateClient(clientID, clientSecret string) (AppOAuthClient, error) {
	client, ok := AppOAuthClientByID(strings.TrimSpace(clientID))
	if !ok {
		return AppOAuthClient{}, oauthError("invalid_client", "unknown OAuth client")
	}
	if !client.Confidential {
		if strings.TrimSpace(clientSecret) != "" {
			return AppOAuthClient{}, oauthError("invalid_client", "public clients must not use a client secret")
		}
		return client, nil
	}
	expected := []byte(s.webClientSecret)
	provided := []byte(clientSecret)
	if len(expected) < 32 || len(expected) != len(provided) || subtle.ConstantTimeCompare(expected, provided) != 1 {
		return AppOAuthClient{}, oauthError("invalid_client", "client authentication failed")
	}
	return client, nil
}

func (s *AppAuthService) issueTokenPair(ctx context.Context, userID int64, clientID string, scopes []string, grantID, sessionID, familyID string, rotating bool, oldRefreshToken string) (*AppTokenResponse, error) {
	var nextRefreshToken string
	if rotating || containsScope(scopes, "offline_access") {
		var err error
		nextRefreshToken, err = appAuthRandomOpaqueToken(48)
		if err != nil {
			return nil, err
		}
	}
	if rotating {
		// The atomic cache operation returns the authoritative old record.
		result, oldRecord, err := s.cache.RotateRefreshToken(ctx, hashOpaqueToken(oldRefreshToken), hashOpaqueToken(nextRefreshToken), nil, AppRefreshTokenTTL, AppRefreshTokenTTL)
		if err != nil {
			return nil, err
		}
		if result == RefreshRotationMissing {
			return nil, oauthError("invalid_grant", "refresh token expired or invalid")
		}
		if oldRecord == nil {
			return nil, oauthError("invalid_grant", "refresh token state is invalid")
		}
		if result == RefreshRotationReused {
			_ = s.repository.RevokeSessionFamily(ctx, oldRecord.SessionID, oldRecord.FamilyID)
			return nil, oauthError("invalid_grant", "refresh token reuse detected; app session revoked")
		}
		if oldRecord.ClientID != clientID {
			_ = s.repository.RevokeSessionFamily(ctx, oldRecord.SessionID, oldRecord.FamilyID)
			return nil, oauthError("invalid_grant", "refresh token client mismatch")
		}
		userID, scopes, grantID, sessionID, familyID = oldRecord.UserID, oldRecord.Scopes, oldRecord.GrantID, oldRecord.SessionID, oldRecord.FamilyID
	}
	grant, session, err := s.repository.GetGrantSession(ctx, grantID, sessionID)
	if errors.Is(err, ErrAppAuthorizationNotFound) || (err == nil && (grant.Status != "active" || session.Status != "active" || session.TokenFamilyID != familyID)) {
		return nil, oauthError("invalid_grant", "app authorization or session is no longer active")
	}
	if err != nil {
		return nil, fmt.Errorf("load device authorization: %w", err)
	}
	if appSessionExpired(session, s.now()) {
		_ = s.repository.RevokeSessionFamily(ctx, session.SessionID, session.TokenFamilyID)
		return nil, oauthError("invalid_grant", "app session expired")
	}
	active, err := s.repository.IsUserActive(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("check app token user: %w", err)
	}
	if !active {
		_ = s.repository.RevokeByGrantID(ctx, grantID)
		_ = s.cache.SetGrantRevoked(ctx, grantID, AppRefreshTokenTTL)
		return nil, oauthError("invalid_grant", "user account is no longer active")
	}
	if revoked, cacheErr := s.cache.IsGrantRevoked(ctx, grantID); cacheErr != nil {
		return nil, cacheErr
	} else if revoked {
		return nil, oauthError("invalid_grant", "device authorization is revoked")
	}
	if grant.UserID != userID || grant.ClientID != clientID || !sameScopes(scopes, session.Scopes) {
		return nil, oauthError("invalid_grant", "app session binding mismatch")
	}
	accessToken, err := s.signAccessToken(userID, clientID, grantID, sessionID, familyID, scopes)
	if err != nil {
		return nil, err
	}
	if !rotating && nextRefreshToken != "" {
		nextRecord := &AppRefreshTokenRecord{UserID: userID, ClientID: clientID, Scopes: scopes, GrantID: grantID, SessionID: sessionID, FamilyID: familyID}
		if err := s.cache.PutRefreshToken(ctx, hashOpaqueToken(nextRefreshToken), nextRecord, AppRefreshTokenTTL); err != nil {
			return nil, err
		}
	}
	_ = s.repository.TouchSession(ctx, sessionID, familyID, s.now())
	return &AppTokenResponse{AccessToken: accessToken, RefreshToken: nextRefreshToken, TokenType: "Bearer", ExpiresIn: int64(AppAccessTokenTTL.Seconds()), Scope: strings.Join(scopes, " ")}, nil
}

type AppAccessClaims struct {
	TokenUse     string   `json:"token_use"`
	TokenVersion int      `json:"token_version"`
	ClientID     string   `json:"client_id"`
	GrantID      string   `json:"grant_id"`
	SessionID    string   `json:"session_id"`
	FamilyID     string   `json:"family_id"`
	Scope        []string `json:"scope"`
	jwt.RegisteredClaims
}

func (s *AppAuthService) signAccessToken(userID int64, clientID, grantID, sessionID, familyID string, scopes []string) (string, error) {
	now := s.now()
	jti, err := appAuthRandomOpaqueToken(18)
	if err != nil {
		return "", err
	}
	claims := AppAccessClaims{
		TokenUse: AppAuthTokenUse, TokenVersion: AppAuthTokenVersion, ClientID: clientID,
		GrantID: grantID, SessionID: sessionID, FamilyID: familyID, Scope: append([]string(nil), scopes...),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: AppAuthIssuer, Subject: strconv.FormatInt(userID, 10), Audience: jwt.ClaimStrings{AppAuthAudience}, ID: jti,
			IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(AppAccessTokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.signingKey)
}

func (s *AppAuthService) ValidateAccessToken(ctx context.Context, raw string) (*AppAccessClaims, error) {
	claims := &AppAccessClaims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(AppAuthIssuer), jwt.WithAudience(AppAuthAudience), jwt.WithExpirationRequired())
	token, err := parser.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) { return s.signingKey, nil })
	userID, subjectErr := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || subjectErr != nil || userID <= 0 || !token.Valid || claims.TokenUse != AppAuthTokenUse || claims.TokenVersion != AppAuthTokenVersion || claims.ClientID == "" || claims.GrantID == "" || claims.SessionID == "" || claims.FamilyID == "" || claims.ID == "" {
		return nil, oauthError("invalid_token", "invalid app access token")
	}
	client, ok := AppOAuthClientByID(claims.ClientID)
	if !ok {
		return nil, oauthError("invalid_token", "unknown app client")
	}
	if revoked, err := s.cache.IsGrantRevoked(ctx, claims.GrantID); err != nil {
		return nil, err
	} else if revoked {
		return nil, oauthError("invalid_token", "device authorization is revoked")
	}
	grant, session, err := s.repository.GetGrantSession(ctx, claims.GrantID, claims.SessionID)
	if errors.Is(err, ErrAppAuthorizationNotFound) || (err == nil && (grant.Status != "active" || session.Status != "active" || session.TokenFamilyID != claims.FamilyID || grant.UserID != userID || grant.ClientID != claims.ClientID)) {
		return nil, oauthError("invalid_token", "app authorization or session is no longer active")
	}
	if err != nil {
		return nil, fmt.Errorf("load device authorization: %w", err)
	}
	if appSessionExpired(session, s.now()) {
		_ = s.repository.RevokeSessionFamily(ctx, session.SessionID, session.TokenFamilyID)
		return nil, oauthError("invalid_token", "app session expired")
	}
	claimScopes, err := normalizeScopeValues(client, claims.Scope)
	if err != nil || !sameScopes(claimScopes, session.Scopes) {
		return nil, oauthError("invalid_token", "token scopes do not match the app session")
	}
	return claims, nil
}

func (s *AppAuthService) ListAuthorizations(ctx context.Context, userID int64) ([]*AppGrant, error) {
	return s.repository.ListGrantsByUserID(ctx, userID)
}

func (s *AppAuthService) ListAuthorizationSessions(ctx context.Context, userID, grantID int64) ([]*AppSession, error) {
	return s.repository.ListSessionsByGrantIDForUser(ctx, grantID, userID, s.now().Add(-AppRefreshTokenTTL))
}

func (s *AppAuthService) RenameAuthorizationSession(ctx context.Context, userID, sessionID int64, deviceName string) (*AppSession, error) {
	deviceName = strings.TrimSpace(deviceName)
	if deviceName == "" {
		return nil, errors.New("device name is required")
	}
	if len([]rune(deviceName)) > 200 {
		return nil, errors.New("device name must not exceed 200 characters")
	}
	return s.repository.UpdateSessionNameByIDForUser(ctx, sessionID, userID, deviceName)
}

func (s *AppAuthService) RevokeAuthorizationSession(ctx context.Context, userID, sessionID int64) error {
	_, err := s.repository.RevokeSessionByIDForUser(ctx, sessionID, userID)
	return err
}

func (s *AppAuthService) RevokeOtherAuthorizationSessions(ctx context.Context, userID, grantID, keepSessionID int64) (int64, error) {
	return s.repository.RevokeOtherSessionsByGrantIDForUser(ctx, grantID, keepSessionID, userID)
}

func (s *AppAuthService) RevokeAuthorization(ctx context.Context, userID, authorizationID int64) error {
	grant, err := s.repository.RevokeGrantByIDForUser(ctx, authorizationID, userID)
	if err != nil {
		return err
	}
	return s.cache.SetGrantRevoked(ctx, grant.GrantID, AppRefreshTokenTTL)
}

func (s *AppAuthService) RevokeGrant(ctx context.Context, grantID string) error {
	if err := s.repository.RevokeByGrantID(ctx, grantID); err != nil {
		return err
	}
	return s.cache.SetGrantRevoked(ctx, grantID, AppRefreshTokenTTL)
}

// RevokeToken implements RFC 7009 idempotency: an unknown or malformed token
// is indistinguishable from a successfully revoked token.
func (s *AppAuthService) RevokeToken(ctx context.Context, rawToken, tokenTypeHint string) error {
	if rawToken == "" {
		return nil
	}
	_ = tokenTypeHint // A hint never restricts lookup; RFC 7009 requires fallback.
	record, err := s.cache.GetRefreshToken(ctx, hashOpaqueToken(rawToken))
	if err == nil {
		err = s.repository.RevokeSessionFamily(ctx, record.SessionID, record.FamilyID)
		if errors.Is(err, ErrAppAuthorizationNotFound) {
			return nil
		}
		return err
	}
	if !errors.Is(err, ErrAppAuthCacheMiss) {
		return err
	}
	claims, err := s.parseAccessTokenWithoutGrantLookup(rawToken)
	if err != nil {
		return nil
	}
	err = s.repository.RevokeSessionFamily(ctx, claims.SessionID, claims.FamilyID)
	if errors.Is(err, ErrAppAuthorizationNotFound) {
		return nil
	}
	return err
}

func (s *AppAuthService) parseAccessTokenWithoutGrantLookup(raw string) (*AppAccessClaims, error) {
	claims := &AppAccessClaims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(AppAuthIssuer), jwt.WithAudience(AppAuthAudience), jwt.WithExpirationRequired())
	token, err := parser.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) { return s.signingKey, nil })
	if err != nil || !token.Valid || claims.TokenUse != AppAuthTokenUse || claims.TokenVersion != AppAuthTokenVersion || claims.GrantID == "" || claims.SessionID == "" || claims.FamilyID == "" {
		return nil, ErrAppAuthCacheMiss
	}
	return claims, nil
}

func appSessionExpired(session *AppSession, now time.Time) bool {
	lastActivity := session.CreatedAt
	if session.LastUsedAt != nil && session.LastUsedAt.After(lastActivity) {
		lastActivity = *session.LastUsedAt
	}
	return lastActivity.Add(AppRefreshTokenTTL).Before(now)
}

func normalizeScopes(client AppOAuthClient, raw string) ([]string, error) {
	return normalizeScopeValues(client, strings.Fields(raw))
}

func normalizeScopeValues(client AppOAuthClient, values []string) ([]string, error) {
	seen := make(map[string]struct{})
	for _, scope := range values {
		if _, ok := client.AllowedScopes[scope]; !ok {
			return nil, oauthError("invalid_scope", "unsupported scope: "+scope)
		}
		seen[scope] = struct{}{}
	}
	if len(seen) == 0 {
		return nil, oauthError("invalid_scope", "at least one scope is required")
	}
	scopes := make([]string, 0, len(seen))
	for scope := range seen {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)
	return scopes, nil
}

func validateRedirectURI(client AppOAuthClient, raw string) error {
	if client.ValidateRedirect == nil || !client.ValidateRedirect(raw) {
		return oauthError("invalid_request", "redirect_uri is not registered for this client")
	}
	return nil
}

func exactRedirect(registered string) func(string) bool {
	return func(raw string) bool { return raw == registered }
}

func validConfiguredWebRedirect(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Scheme == "https" && parsed.Host != "" && parsed.User == nil &&
		(parsed.Path == "/api/auth/oauth/callback" || parsed.Path == "/oauth/callback") && parsed.RawQuery == "" && parsed.Fragment == ""
}

func validateZeroCanvasWebRedirect(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Path != "/oauth/callback" {
		return false
	}
	if parsed.Scheme == "http" && (isLoopbackHost(parsed.Hostname()) || isPrivateIPHost(parsed.Hostname())) && parsed.Port() != "" {
		return true
	}
	if parsed.Scheme != "https" || parsed.Host == "" {
		return false
	}
	appWebRedirects.RLock()
	_, ok := appWebRedirects.values[raw]
	appWebRedirects.RUnlock()
	return ok
}

func validateZeroAgentWebRedirect(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" ||
		parsed.Path != "/api/auth/oauth/callback" {
		return false
	}
	if parsed.Scheme == "http" && isLoopbackHost(parsed.Hostname()) && parsed.Port() != "" {
		return true
	}
	if parsed.Scheme == "http" && isPrivateIPHost(parsed.Hostname()) && parsed.Port() != "" {
		return true
	}
	if parsed.Scheme != "https" || parsed.Host == "" {
		return false
	}
	appWebRedirects.RLock()
	_, ok := appWebRedirects.values[raw]
	appWebRedirects.RUnlock()
	return ok
}

func isPrivateIPHost(host string) bool {
	ip := net.ParseIP(host)
	return ip != nil && ip.IsPrivate()
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func validateZeroAgentLoopbackRedirect(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Fragment != "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Opaque != "" {
		return false
	}
	if u.Scheme != "http" || u.Hostname() != "127.0.0.1" || u.Path != "/oauth/callback" || u.RawPath != "" {
		return false
	}
	port, err := strconv.Atoi(u.Port())
	return err == nil && port >= 1 && port <= 65535 && net.ParseIP(u.Hostname()) != nil
}

func validPKCEValue(value string) bool {
	if len(value) < 43 || len(value) > 128 {
		return false
	}
	for _, r := range value {
		allowed := (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			strings.ContainsRune("-._~", r)
		if !allowed {
			return false
		}
	}
	return true
}

func validPKCEChallenge(value string) bool {
	if len(value) != 43 || !validPKCEValue(value) {
		return false
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size
}

func verifyPKCE(verifier, challenge string) bool {
	if !validPKCEValue(verifier) {
		return false
	}
	sum := sha256.Sum256([]byte(verifier))
	actual := base64.RawURLEncoding.EncodeToString(sum[:])
	return hmac.Equal([]byte(actual), []byte(challenge))
}

func appAuthRandomOpaqueToken(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashOpaqueToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func containsScope(scopes []string, target string) bool {
	for _, scope := range scopes {
		if scope == target {
			return true
		}
	}
	return false
}

func sameScopes(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	rightCopy := append([]string(nil), right...)
	sort.Strings(rightCopy)
	for i := range left {
		if left[i] != rightCopy[i] {
			return false
		}
	}
	return true
}
