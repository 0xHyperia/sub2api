package admin

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAdminPromotionStatsFilterDoesNotBroadenInvalidAgentID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("GET", "/?agent_id=not-an-id", nil)

	filter := adminPromotionStatsFilter(ctx)
	require.EqualValues(t, -1, filter.AgentID)
}

func TestAdminPromotionStatsFilterUsesHongKongCalendarDays(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("GET", "/?date_from=2026-07-01&date_to=2026-07-01", nil)

	filter := adminPromotionStatsFilter(ctx)
	require.Equal(t, "2026-07-01T00:00:00+08:00", filter.From.Format(time.RFC3339))
	require.Equal(t, "2026-07-02T00:00:00+08:00", filter.To.Format(time.RFC3339))
}
