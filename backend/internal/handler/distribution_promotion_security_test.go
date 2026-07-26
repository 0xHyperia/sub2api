package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestTrackPromotionVisitRejectsBodyOver16KiB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := &DistributionHandler{}
	router.POST("/api/v1/distribution/track", servermiddleware.RequestBodyLimit(16*1024), handler.TrackPromotionVisit)

	body := `{"promotion_code":"AGENT001","utm_campaign":"` + strings.Repeat("x", 17*1024) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/distribution/track", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
}

func TestClearDistributionVisitorCookieUsesForwardedHTTPS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/distribution/tracking-status", nil)
	ctx.Request.Header.Set("X-Forwarded-Proto", "https")

	clearDistributionVisitorCookie(ctx)

	cookie := recorder.Header().Get("Set-Cookie")
	require.Contains(t, cookie, distributionVisitorCookie+"=")
	require.Contains(t, cookie, "Max-Age=0")
	require.Contains(t, cookie, "HttpOnly")
	require.Contains(t, cookie, "Secure")
	require.Contains(t, cookie, "SameSite=Lax")
}
