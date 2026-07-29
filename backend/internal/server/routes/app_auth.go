package routes

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	ratelimit "github.com/Wei-Shaw/sub2api/internal/middleware"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RegisterAppAuthRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth servermiddleware.JWTAuthMiddleware,
	appJWTAuth servermiddleware.AppJWTAuthMiddleware,
	redisClient *redis.Client,
) {
	limiter := ratelimit.NewRateLimiter(redisClient)
	appAuth := v1.Group("/app-auth")
	{
		appAuth.POST("/authorize/requests", limiter.LimitWithOptions("app-auth-request", 60, time.Minute, ratelimit.RateLimitOptions{FailureMode: ratelimit.RateLimitFailClose}), h.AppAuth.CreateAuthorizationRequest)
		appAuth.POST("/token", limiter.LimitWithOptions("app-auth-token", 60, time.Minute, ratelimit.RateLimitOptions{FailureMode: ratelimit.RateLimitFailClose}), h.AppAuth.Token)
		appAuth.POST("/revoke", limiter.LimitWithOptions("app-auth-revoke", 60, time.Minute, ratelimit.RateLimitOptions{FailureMode: ratelimit.RateLimitFailClose}), h.AppAuth.Revoke)

		webAuthenticated := appAuth.Group("")
		webAuthenticated.Use(gin.HandlerFunc(jwtAuth))
		{
			webAuthenticated.GET("/authorize/context", h.AppAuth.AuthorizationContext)
			webAuthenticated.POST("/authorize/decision", h.AppAuth.DecideAuthorization)
			webAuthenticated.GET("/grants", h.AppAuth.ListGrants)
			webAuthenticated.DELETE("/grants/:id", h.AppAuth.RevokeGrant)
			webAuthenticated.GET("/grants/:id/sessions", h.AppAuth.ListGrantSessions)
			webAuthenticated.PATCH("/sessions/:id", h.AppAuth.RenameSession)
			webAuthenticated.DELETE("/sessions/:id", h.AppAuth.RevokeSession)
			webAuthenticated.POST("/grants/:id/sessions/:sessionId/revoke-others", h.AppAuth.RevokeOtherSessions)
		}
	}

	app := v1.Group("/app")
	app.Use(gin.HandlerFunc(appJWTAuth))
	{
		app.GET("/me", servermiddleware.RequireAppScope("profile:read"), h.AppResource.Me)
		app.PATCH("/me", servermiddleware.RequireAppScope("profile:write"), h.AppResource.UpdateMe)
		app.GET("/usage", servermiddleware.RequireAppScope("usage:read"), h.AppResource.Usage)
		app.GET("/groups", servermiddleware.RequireAppScope("groups:read"), h.AppResource.Groups)
		app.GET("/keys", servermiddleware.RequireAppScope("keys:read"), h.AppResource.Keys)
		app.POST("/groups/:groupId/keys", servermiddleware.RequireAppScope("keys:write"), h.AppResource.CreateKey)
		app.GET("/subscriptions", servermiddleware.RequireAppScope("subscriptions:read"), h.AppResource.Subscriptions)
		app.POST("/execution/step-up", servermiddleware.RequireAppScope("execution:authorize"), h.AppResource.IssueExecutionStepUp)
		app.POST("/execution/step-up/consume", servermiddleware.RequireAppScope("execution:authorize"), h.AppResource.ConsumeExecutionStepUp)
	}
}
