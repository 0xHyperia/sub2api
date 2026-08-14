package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type BusinessAnalyticsHandler struct {
	service *service.BusinessAnalyticsService
}

func parseBusinessAnalyticsAgentID(c *gin.Context) (int64, bool) {
	value := strings.TrimSpace(c.Query("agent_id"))
	if value == "" {
		return 0, true
	}
	agentID, err := strconv.ParseInt(value, 10, 64)
	if err != nil || agentID <= 0 {
		response.BadRequest(c, "invalid distribution agent")
		return 0, false
	}
	return agentID, true
}

func NewBusinessAnalyticsHandler(analyticsService *service.BusinessAnalyticsService) *BusinessAnalyticsHandler {
	return &BusinessAnalyticsHandler{service: analyticsService}
}

func (h *BusinessAnalyticsHandler) Get(c *gin.Context) {
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	parseDate := func(key string) (time.Time, error) {
		return time.ParseInLocation("2006-01-02", strings.TrimSpace(c.Query(key)), location)
	}
	from, err := parseDate("date_from")
	if err != nil {
		response.BadRequest(c, "invalid analytics date range")
		return
	}
	to, err := parseDate("date_to")
	if err != nil {
		response.BadRequest(c, "invalid analytics date range")
		return
	}
	agentID, ok := parseBusinessAnalyticsAgentID(c)
	if !ok {
		return
	}
	result, err := h.service.Get(c.Request.Context(), service.BusinessAnalyticsFilter{
		DateFrom: from, DateTo: to, Channel: c.Query("channel"), AgentID: agentID,
		AgentScope: c.Query("agent_scope"), Comparison: true, Section: c.Param("section"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *BusinessAnalyticsHandler) GetBalance(c *gin.Context) {
	window, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("activity_window", "7")))
	if err != nil {
		response.BadRequest(c, "invalid activity window")
		return
	}
	agentID, ok := parseBusinessAnalyticsAgentID(c)
	if !ok {
		return
	}
	result, err := h.service.GetBalance(c.Request.Context(), window, agentID, c.Query("agent_scope"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}
