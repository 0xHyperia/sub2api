package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AppAuthIssuer           = "https://usa0.top"
	AppAuthAudience         = "sub2api-app-api"
	AppAuthTokenUse         = "app"
	AuthorizationRequestTTL = 5 * time.Minute
	AuthorizationCodeTTL    = time.Minute
	AppAccessTokenTTL       = 10 * time.Minute
	AppRefreshTokenTTL      = 30 * 24 * time.Hour
	MaxAppAuthStateLength   = 1024
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

type AppPublicClient struct {
	ID               string
	Platform         string
	Name             string
	AllowedScopes    map[string]struct{}
	ValidateRedirect func(string) bool
}

var appAllowedScopes = map[string]struct{}{
	// Temporary ZeroBox compatibility scope. No id_token or OIDC UserInfo is
	// issued in the first integration phase.
	"openid":             {},
	"profile:read":       {},
	"groups:read":        {},
	"keys:read":          {},
	"keys:write":         {},
	"subscriptions:read": {},
	"offline_access":     {},
}

var appPublicClients = map[string]AppPublicClient{
	"zerobox-desktop": {
		ID: "zerobox-desktop", Platform: "desktop", Name: "ZeroBox",
		AllowedScopes: appAllowedScopes, ValidateRedirect: validateZeroBoxLoopbackRedirect,
	},
	"zerobox-android": {
		ID: "zerobox-android", Platform: "android", Name: "ZeroBox",
		AllowedScopes: appAllowedScopes, ValidateRedirect: exactRedirect("zerobox://oauth/callback"),
	},
	"zerobox-ios": {
		ID: "zerobox-ios", Platform: "ios", Name: "ZeroBox",
		AllowedScopes: appAllowedScopes, ValidateRedirect: exactRedirect("zerobox://oauth/callback"),
	},
}

func AppPublicClientByID(clientID string) (AppPublicClient, bool) {
	client, ok := appPublicClients[clientID]
	return client, ok
}

type AppAuthorization struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	GrantID       string     `json:"grant_id"`
	ClientID      string     `json:"client_id"`
	DeviceName    string     `json:"device_name"`
	Platform      string     `json:"platform"`
	Scopes        []string   `json:"scopes"`
	TokenFamilyID string     `json:"token_family_id"`
	Status        string     `json:"status"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type AppAuthorizationRepository interface {
	Create(ctx context.Context, authorization *AppAuthorization) error
	GetByGrantID(ctx context.Context, grantID string) (*AppAuthorization, error)
	ListByUserID(ctx context.Context, userID int64) ([]*AppAuthorization, error)
	RevokeByIDForUser(ctx context.Context, id, userID int64) (*AppAuthorization, error)
	RevokeByGrantID(ctx context.Context, grantID string) error
	Touch(ctx context.Context, grantID string, usedAt time.Time) error
	IsUserActive(ctx context.Context, userID int64) (bool, error)
}

type AuthorizationRequest struct {
	RequestID     string   `json:"request_id"`
	ClientID      string   `json:"client_id"`
	RedirectURI   string   `json:"redirect_uri"`
	Scopes        []string `json:"scopes"`
	State         string   `json:"state"`
	CodeChallenge string   `json:"code_challenge"`
	DeviceName    string   `json:"device_name"`
	Platform      string   `json:"platform"`
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
}

type AuthorizationCode struct {
	UserID        int64    `json:"user_id"`
	ClientID      string   `json:"client_id"`
	RedirectURI   string   `json:"redirect_uri"`
	Scopes        []string `json:"scopes"`
	CodeChallenge string   `json:"code_challenge"`
	GrantID       string   `json:"grant_id"`
	FamilyID      string   `json:"family_id"`
}

type AppRefreshTokenRecord struct {
	UserID   int64    `json:"user_id"`
	ClientID string   `json:"client_id"`
	Scopes   []string `json:"scopes"`
	GrantID  string   `json:"grant_id"`
	FamilyID string   `json:"family_id"`
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
	repository AppAuthorizationRepository
	cache      AppAuthCache
	signingKey []byte
	now        func() time.Time
}

func NewAppAuthService(repository AppAuthorizationRepository, cache AppAuthCache, signingSecret string) *AppAuthService {
	return &AppAuthService{repository: repository, cache: cache, signingKey: []byte(signingSecret), now: time.Now}
}

func (s *AppAuthService) CreateAuthorizationRequest(ctx context.Context, input AuthorizationRequestInput) (*AuthorizationRequest, error) {
	if input.ResponseType != "code" {
		return nil, oauthError("unsupported_response_type", "response_type must be code")
	}
	client, ok := AppPublicClientByID(input.ClientID)
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
	}
	if len(request.DeviceName) > 200 || len(request.Platform) > 40 {
		return nil, oauthError("invalid_request", "device metadata is too long")
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
	grantID, err := appAuthRandomOpaqueToken(24)
	if err != nil {
		return nil, err
	}
	familyID, err := appAuthRandomOpaqueToken(24)
	if err != nil {
		return nil, err
	}
	authorization := &AppAuthorization{
		UserID: userID, GrantID: grantID, ClientID: request.ClientID,
		DeviceName: request.DeviceName, Platform: request.Platform,
		Scopes: append([]string(nil), request.Scopes...), TokenFamilyID: familyID, Status: "active",
	}
	if err := s.repository.Create(ctx, authorization); err != nil {
		return nil, fmt.Errorf("create app authorization: %w", err)
	}
	code, err := appAuthRandomOpaqueToken(32)
	if err != nil {
		_ = s.repository.RevokeByGrantID(ctx, grantID)
		return nil, err
	}
	entry := &AuthorizationCode{
		UserID: userID, ClientID: request.ClientID, RedirectURI: request.RedirectURI,
		Scopes: request.Scopes, CodeChallenge: request.CodeChallenge, GrantID: grantID, FamilyID: familyID,
	}
	if err := s.cache.PutAuthorizationCode(ctx, hashOpaqueToken(code), entry, AuthorizationCodeTTL); err != nil {
		_ = s.repository.RevokeByGrantID(ctx, grantID)
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

func (s *AppAuthService) ExchangeAuthorizationCode(ctx context.Context, clientID, code, redirectURI, verifier string) (*AppTokenResponse, error) {
	client, ok := AppPublicClientByID(clientID)
	if !ok {
		return nil, oauthError("invalid_client", "unknown public client")
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
	return s.issueTokenPair(ctx, entry.UserID, entry.ClientID, entry.Scopes, entry.GrantID, entry.FamilyID, false, "")
}

func (s *AppAuthService) Refresh(ctx context.Context, clientID, refreshToken string) (*AppTokenResponse, error) {
	if _, ok := AppPublicClientByID(clientID); !ok {
		return nil, oauthError("invalid_client", "unknown public client")
	}
	if refreshToken == "" {
		return nil, oauthError("invalid_grant", "refresh token is required")
	}
	return s.issueTokenPair(ctx, 0, clientID, nil, "", "", true, refreshToken)
}

func (s *AppAuthService) issueTokenPair(ctx context.Context, userID int64, clientID string, scopes []string, grantID, familyID string, rotating bool, oldRefreshToken string) (*AppTokenResponse, error) {
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
			_ = s.repository.RevokeByGrantID(ctx, oldRecord.GrantID)
			_ = s.cache.SetGrantRevoked(ctx, oldRecord.GrantID, AppRefreshTokenTTL)
			return nil, oauthError("invalid_grant", "refresh token reuse detected; device authorization revoked")
		}
		if oldRecord.ClientID != clientID {
			_ = s.repository.RevokeByGrantID(ctx, oldRecord.GrantID)
			_ = s.cache.SetGrantRevoked(ctx, oldRecord.GrantID, AppRefreshTokenTTL)
			return nil, oauthError("invalid_grant", "refresh token client mismatch")
		}
		userID, scopes, grantID, familyID = oldRecord.UserID, oldRecord.Scopes, oldRecord.GrantID, oldRecord.FamilyID
	}
	authorization, err := s.repository.GetByGrantID(ctx, grantID)
	if errors.Is(err, ErrAppAuthorizationNotFound) || (err == nil && authorization.Status != "active") {
		return nil, oauthError("invalid_grant", "device authorization is no longer active")
	}
	if err != nil {
		return nil, fmt.Errorf("load device authorization: %w", err)
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
	accessToken, err := s.signAccessToken(userID, clientID, grantID, scopes)
	if err != nil {
		return nil, err
	}
	if !rotating && nextRefreshToken != "" {
		nextRecord := &AppRefreshTokenRecord{UserID: userID, ClientID: clientID, Scopes: scopes, GrantID: grantID, FamilyID: familyID}
		if err := s.cache.PutRefreshToken(ctx, hashOpaqueToken(nextRefreshToken), nextRecord, AppRefreshTokenTTL); err != nil {
			return nil, err
		}
	}
	_ = s.repository.Touch(ctx, grantID, s.now())
	return &AppTokenResponse{AccessToken: accessToken, RefreshToken: nextRefreshToken, TokenType: "Bearer", ExpiresIn: int64(AppAccessTokenTTL.Seconds()), Scope: strings.Join(scopes, " ")}, nil
}

type AppAccessClaims struct {
	TokenUse string   `json:"token_use"`
	ClientID string   `json:"client_id"`
	GrantID  string   `json:"grant_id"`
	Scope    []string `json:"scope"`
	jwt.RegisteredClaims
}

func (s *AppAuthService) signAccessToken(userID int64, clientID, grantID string, scopes []string) (string, error) {
	now := s.now()
	jti, err := appAuthRandomOpaqueToken(18)
	if err != nil {
		return "", err
	}
	claims := AppAccessClaims{
		TokenUse: AppAuthTokenUse, ClientID: clientID, GrantID: grantID, Scope: append([]string(nil), scopes...),
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
	if err != nil || subjectErr != nil || userID <= 0 || !token.Valid || claims.TokenUse != AppAuthTokenUse || claims.ClientID == "" || claims.GrantID == "" || claims.ID == "" {
		return nil, oauthError("invalid_token", "invalid app access token")
	}
	client, ok := AppPublicClientByID(claims.ClientID)
	if !ok {
		return nil, oauthError("invalid_token", "unknown app client")
	}
	if revoked, err := s.cache.IsGrantRevoked(ctx, claims.GrantID); err != nil {
		return nil, err
	} else if revoked {
		return nil, oauthError("invalid_token", "device authorization is revoked")
	}
	authorization, err := s.repository.GetByGrantID(ctx, claims.GrantID)
	if errors.Is(err, ErrAppAuthorizationNotFound) || (err == nil && (authorization.Status != "active" || authorization.UserID != userID || authorization.ClientID != claims.ClientID)) {
		return nil, oauthError("invalid_token", "device authorization is no longer active")
	}
	if err != nil {
		return nil, fmt.Errorf("load device authorization: %w", err)
	}
	claimScopes, err := normalizeScopeValues(client, claims.Scope)
	if err != nil || !sameScopes(claimScopes, authorization.Scopes) {
		return nil, oauthError("invalid_token", "token scopes do not match the device authorization")
	}
	return claims, nil
}

func (s *AppAuthService) ListAuthorizations(ctx context.Context, userID int64) ([]*AppAuthorization, error) {
	return s.repository.ListByUserID(ctx, userID)
}

func (s *AppAuthService) RevokeAuthorization(ctx context.Context, userID, authorizationID int64) error {
	authorization, err := s.repository.RevokeByIDForUser(ctx, authorizationID, userID)
	if err != nil {
		return err
	}
	return s.cache.SetGrantRevoked(ctx, authorization.GrantID, AppRefreshTokenTTL)
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
		err = s.RevokeGrant(ctx, record.GrantID)
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
	err = s.RevokeGrant(ctx, claims.GrantID)
	if errors.Is(err, ErrAppAuthorizationNotFound) {
		return nil
	}
	return err
}

func (s *AppAuthService) parseAccessTokenWithoutGrantLookup(raw string) (*AppAccessClaims, error) {
	claims := &AppAccessClaims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(AppAuthIssuer), jwt.WithAudience(AppAuthAudience), jwt.WithExpirationRequired())
	token, err := parser.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) { return s.signingKey, nil })
	if err != nil || !token.Valid || claims.TokenUse != AppAuthTokenUse || claims.GrantID == "" {
		return nil, ErrAppAuthCacheMiss
	}
	return claims, nil
}

func normalizeScopes(client AppPublicClient, raw string) ([]string, error) {
	return normalizeScopeValues(client, strings.Fields(raw))
}

func normalizeScopeValues(client AppPublicClient, values []string) ([]string, error) {
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

func validateRedirectURI(client AppPublicClient, raw string) error {
	if client.ValidateRedirect == nil || !client.ValidateRedirect(raw) {
		return oauthError("invalid_request", "redirect_uri is not registered for this client")
	}
	return nil
}

func exactRedirect(registered string) func(string) bool {
	return func(raw string) bool { return raw == registered }
}

func validateZeroBoxLoopbackRedirect(raw string) bool {
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
