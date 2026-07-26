package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDistributionVisitorContextMiddlewareCoversOAuthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(DistributionVisitorContextMiddleware())
	router.GET("/auth/oauth/callback", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"token": service.DistributionVisitorTokenFromContext(c.Request.Context())})
	})
	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/callback", nil)
	req.AddCookie(&http.Cookie{Name: distributionVisitorCookie, Value: "visitor-token"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"token":"visitor-token"`)
}
