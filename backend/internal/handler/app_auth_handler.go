package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const maxAppAuthJSONBody = 32 << 10

type AppAuthHandler struct {
	service *service.AppAuthService
}

func NewAppAuthHandler(appAuthService *service.AppAuthService) *AppAuthHandler {
	return &AppAuthHandler{service: appAuthService}
}

type createAppAuthorizationRequest struct {
	ResponseType        string `json:"response_type"`
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	Scope               string `json:"scope"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	State               string `json:"state"`
	DeviceName          string `json:"device_name"`
	Platform            string `json:"platform"`
	InstallationID      string `json:"installation_id"`
}

func (h *AppAuthHandler) CreateAuthorizationRequest(c *gin.Context) {
	var request createAppAuthorizationRequest
	if err := decodeStrictJSON(c, &request); err != nil {
		writeOAuthError(c, serviceOAuthError("invalid_request", "invalid authorization request"))
		return
	}
	created, err := h.service.CreateAuthorizationRequest(c.Request.Context(), service.AuthorizationRequestInput{
		ResponseType: request.ResponseType, ClientID: request.ClientID, RedirectURI: request.RedirectURI,
		Scope: request.Scope, CodeChallenge: request.CodeChallenge, CodeChallengeMethod: request.CodeChallengeMethod,
		State: request.State, DeviceName: request.DeviceName, Platform: request.Platform, InstallationID: request.InstallationID,
	})
	if err != nil {
		writeOAuthError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"request_id":        created.RequestID,
		"expires_in":        int64(service.AuthorizationRequestTTL.Seconds()),
		"authorization_uri": "/oauth/authorize?request_id=" + url.QueryEscape(created.RequestID),
	})
}

func (h *AppAuthHandler) AuthorizationContext(c *gin.Context) {
	request, err := h.service.GetAuthorizationContext(c.Request.Context(), strings.TrimSpace(c.Query("request_id")))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	client, _ := service.AppOAuthClientByID(request.ClientID)
	response.Success(c, gin.H{
		"request_id": request.RequestID, "client_id": request.ClientID, "client_name": client.Name,
		"device_name": request.DeviceName, "platform": request.Platform, "scopes": request.Scopes,
	})
}

type appAuthorizationDecisionRequest struct {
	RequestID string `json:"request_id"`
	Decision  string `json:"decision"`
}

func (h *AppAuthHandler) DecideAuthorization(c *gin.Context) {
	subject, ok := servermiddleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var request appAuthorizationDecisionRequest
	if err := decodeStrictJSON(c, &request); err != nil || (request.Decision != "allow" && request.Decision != "deny") {
		response.BadRequest(c, "Invalid authorization decision")
		return
	}
	decision, err := h.service.DecideAuthorization(c.Request.Context(), subject.UserID, request.RequestID, request.Decision == "allow")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	callback, err := url.Parse(decision.Request.RedirectURI)
	if err != nil {
		response.InternalError(c, "Invalid registered redirect URI")
		return
	}
	query := callback.Query()
	query.Set("state", decision.Request.State)
	if decision.AccessDenied {
		query.Set("error", "access_denied")
	} else {
		query.Set("code", decision.Code)
	}
	callback.RawQuery = query.Encode()
	response.Success(c, gin.H{"redirect_uri": callback.String()})
}

func (h *AppAuthHandler) Token(c *gin.Context) {
	if !strings.HasPrefix(strings.ToLower(c.GetHeader("Content-Type")), "application/x-www-form-urlencoded") {
		writeOAuthError(c, serviceOAuthError("invalid_request", "form encoding is required"))
		return
	}
	if err := c.Request.ParseForm(); err != nil {
		writeOAuthError(c, serviceOAuthError("invalid_request", "invalid form body"))
		return
	}
	var token *service.AppTokenResponse
	var err error
	clientID, clientSecret, credentialErr := oauthClientCredentials(c)
	if credentialErr != nil {
		writeOAuthError(c, credentialErr)
		return
	}
	switch c.PostForm("grant_type") {
	case "authorization_code":
		token, err = h.service.ExchangeAuthorizationCode(c.Request.Context(), clientID, clientSecret, c.PostForm("code"), c.PostForm("redirect_uri"), c.PostForm("code_verifier"))
	case "refresh_token":
		token, err = h.service.Refresh(c.Request.Context(), clientID, clientSecret, c.PostForm("refresh_token"))
	default:
		err = serviceOAuthError("unsupported_grant_type", "unsupported grant_type")
	}
	if err != nil {
		writeOAuthError(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.JSON(http.StatusOK, token)
}

func oauthClientCredentials(c *gin.Context) (string, string, error) {
	formClientID := strings.TrimSpace(c.PostForm("client_id"))
	clientID, clientSecret, hasBasic := c.Request.BasicAuth()
	clientID = strings.TrimSpace(clientID)
	if hasBasic {
		if formClientID != "" && formClientID != clientID {
			return "", "", serviceOAuthError("invalid_client", "conflicting client credentials")
		}
		return clientID, clientSecret, nil
	}
	return formClientID, "", nil
}

func (h *AppAuthHandler) Revoke(c *gin.Context) {
	if !strings.HasPrefix(strings.ToLower(c.GetHeader("Content-Type")), "application/x-www-form-urlencoded") {
		writeOAuthError(c, serviceOAuthError("invalid_request", "form encoding is required"))
		return
	}
	if err := c.Request.ParseForm(); err != nil {
		writeOAuthError(c, serviceOAuthError("invalid_request", "invalid form body"))
		return
	}
	if err := h.service.RevokeToken(c.Request.Context(), c.PostForm("token"), c.PostForm("token_type_hint")); err != nil {
		writeOAuthError(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{})
}

func (h *AppAuthHandler) ListGrants(c *gin.Context) {
	subject, ok := servermiddleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	grants, err := h.service.ListAuthorizations(c.Request.Context(), subject.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list authorized apps")
		return
	}
	result := make([]gin.H, 0, len(grants))
	for _, grant := range grants {
		client, _ := service.AppOAuthClientByID(grant.ClientID)
		result = append(result, gin.H{
			"id": grant.ID, "client_id": grant.ClientID, "client_name": client.Name,
			"platform": client.Platform, "scopes": grant.Scopes, "session_count": grant.SessionCount,
			"first_authorized_at": grant.FirstAuthorizedAt, "last_authorized_at": grant.LastAuthorizedAt,
			"last_used_at": grant.LastUsedAt,
		})
	}
	response.Success(c, result)
}

func (h *AppAuthHandler) RevokeGrant(c *gin.Context) {
	subject, ok := servermiddleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid authorization ID")
		return
	}
	if err := h.service.RevokeAuthorization(c.Request.Context(), subject.UserID, id); err != nil {
		response.NotFound(c, "Authorized app not found")
		return
	}
	response.Success(c, gin.H{"revoked": true})
}

func (h *AppAuthHandler) ListGrantSessions(c *gin.Context) {
	subject, id, ok := appAuthManagementSubjectAndID(c, "authorization")
	if !ok {
		return
	}
	sessions, err := h.service.ListAuthorizationSessions(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list application sessions")
		return
	}
	result := make([]gin.H, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, gin.H{
			"id": session.ID, "device_name": session.DeviceName, "platform": session.Platform,
			"created_at": session.CreatedAt, "last_used_at": session.LastUsedAt,
		})
	}
	response.Success(c, result)
}

func (h *AppAuthHandler) RenameSession(c *gin.Context) {
	subject, id, ok := appAuthManagementSubjectAndID(c, "session")
	if !ok {
		return
	}
	var request struct {
		DeviceName string `json:"device_name"`
	}
	if err := decodeStrictJSON(c, &request); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}
	session, err := h.service.RenameAuthorizationSession(c.Request.Context(), subject.UserID, id, request.DeviceName)
	if err != nil {
		if errors.Is(err, service.ErrAppAuthorizationNotFound) {
			response.NotFound(c, "Application session not found")
		} else {
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.Success(c, gin.H{"id": session.ID, "device_name": session.DeviceName})
}

func (h *AppAuthHandler) RevokeSession(c *gin.Context) {
	subject, id, ok := appAuthManagementSubjectAndID(c, "session")
	if !ok {
		return
	}
	if err := h.service.RevokeAuthorizationSession(c.Request.Context(), subject.UserID, id); err != nil {
		response.NotFound(c, "Application session not found")
		return
	}
	response.Success(c, gin.H{"revoked": true})
}

func (h *AppAuthHandler) RevokeOtherSessions(c *gin.Context) {
	subject, grantID, ok := appAuthManagementSubjectAndID(c, "authorization")
	if !ok {
		return
	}
	keepSessionID, err := strconv.ParseInt(c.Param("sessionId"), 10, 64)
	if err != nil || keepSessionID <= 0 {
		response.BadRequest(c, "Invalid session ID")
		return
	}
	revoked, err := h.service.RevokeOtherAuthorizationSessions(c.Request.Context(), subject.UserID, grantID, keepSessionID)
	if err != nil {
		response.NotFound(c, "Application session not found")
		return
	}
	response.Success(c, gin.H{"revoked_count": revoked})
}

func appAuthManagementSubjectAndID(c *gin.Context, resource string) (servermiddleware.AuthSubject, int64, bool) {
	subject, ok := servermiddleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return servermiddleware.AuthSubject{}, 0, false
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid "+resource+" ID")
		return servermiddleware.AuthSubject{}, 0, false
	}
	return subject, id, true
}

func decodeStrictJSON(c *gin.Context, target any) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAppAuthJSONBody)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON value")
	}
	return nil
}

func serviceOAuthError(code, description string) error {
	return &service.OAuthError{Code: code, Description: description}
}

func writeOAuthError(c *gin.Context, err error) {
	var oauthErr *service.OAuthError
	if !errors.As(err, &oauthErr) {
		oauthErr = &service.OAuthError{Code: "server_error", Description: "authorization service unavailable"}
	}
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	status := http.StatusBadRequest
	if oauthErr.Code == "server_error" {
		status = http.StatusInternalServerError
	}
	c.JSON(status, gin.H{"error": oauthErr.Code, "error_description": oauthErr.Description})
}
