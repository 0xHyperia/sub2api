package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const contextKeyAppAuth = "app_auth_subject"

type AppAuthSubject struct {
	UserID   int64
	ClientID string
	GrantID  string
	Scopes   map[string]struct{}
}

func (s AppAuthSubject) HasScope(scope string) bool {
	_, ok := s.Scopes[scope]
	return ok
}

func GetAppAuthSubject(c *gin.Context) (AppAuthSubject, bool) {
	value, exists := c.Get(contextKeyAppAuth)
	if !exists {
		return AppAuthSubject{}, false
	}
	subject, ok := value.(AppAuthSubject)
	return subject, ok
}

func NewAppJWTAuthMiddleware(appAuthService *service.AppAuthService, userService *service.UserService) AppJWTAuthMiddleware {
	return AppJWTAuthMiddleware(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			abortAppRequest(c, http.StatusUnauthorized, "invalid_token", "A valid app access token is required")
			return
		}

		claims, err := appAuthService.ValidateAccessToken(c.Request.Context(), strings.TrimSpace(parts[1]))
		if err != nil {
			abortAppRequest(c, http.StatusUnauthorized, "invalid_token", "Invalid or expired app access token")
			return
		}
		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil || userID <= 0 {
			abortAppRequest(c, http.StatusUnauthorized, "invalid_token", "Invalid app access token subject")
			return
		}
		user, err := userService.GetByID(c.Request.Context(), userID)
		if err != nil || !user.IsActive() {
			abortAppRequest(c, http.StatusUnauthorized, "invalid_token", "App token user is not active")
			return
		}

		scopes := make(map[string]struct{})
		for _, scope := range claims.Scope {
			scopes[scope] = struct{}{}
		}
		c.Set(contextKeyAppAuth, AppAuthSubject{
			UserID: userID, ClientID: claims.ClientID, GrantID: claims.GrantID, Scopes: scopes,
		})
		c.Next()
	})
}

func RequireAppScope(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		subject, ok := GetAppAuthSubject(c)
		if !ok {
			abortAppRequest(c, http.StatusUnauthorized, "invalid_token", "App authorization is required")
			return
		}
		if !subject.HasScope(scope) {
			abortAppRequest(c, http.StatusForbidden, "insufficient_scope", "Required scope: "+scope)
			return
		}
		c.Next()
	}
}

func abortAppRequest(c *gin.Context, status int, code, message string) {
	c.Header("Cache-Control", "no-store")
	c.JSON(status, gin.H{
		"success": false,
		"message": message,
		"error":   gin.H{"code": code, "message": message},
	})
	c.Abort()
}
