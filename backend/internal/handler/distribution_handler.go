package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type DistributionHandler struct{ service *service.DistributionService }

const distributionVisitorCookie = "sub2api_distribution_visitor"

// DistributionVisitorContextMiddleware makes promotion attribution available
// to every registration path, including multi-step OAuth callbacks.
func DistributionVisitorContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if token, err := c.Cookie(distributionVisitorCookie); err == nil {
			c.Request = c.Request.WithContext(service.WithDistributionVisitorToken(c.Request.Context(), token))
		}
		c.Next()
	}
}

func NewDistributionHandler(distributionService *service.DistributionService) *DistributionHandler {
	return &DistributionHandler{service: distributionService}
}

func distributionUserID(c *gin.Context) (int64, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return 0, false
	}
	return subject.UserID, true
}

func distributionAgentUserID(c *gin.Context, distributionService *service.DistributionService) (int64, bool) {
	userID, ok := distributionUserID(c)
	if !ok {
		return 0, false
	}
	if err := distributionService.EnsureAgentAccess(c.Request.Context(), userID); err != nil {
		response.ErrorFrom(c, err)
		return 0, false
	}
	return userID, true
}

func distributionAnalyticsFilter(c *gin.Context) (service.DistributionAnalyticsFilter, error) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	filter := service.DistributionAnalyticsFilter{Days: days}
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	for raw, target := range map[string]**time.Time{"date_from": &filter.DateFrom, "date_to": &filter.DateTo} {
		value := strings.TrimSpace(c.Query(raw))
		if value == "" {
			continue
		}
		parsed, err := time.ParseInLocation("2006-01-02", value, location)
		if err != nil {
			return service.DistributionAnalyticsFilter{}, err
		}
		*target = &parsed
	}
	return filter, nil
}

func (h *DistributionHandler) GetAccess(c *gin.Context) {
	uid, ok := distributionUserID(c)
	if !ok {
		return
	}
	access, err := h.service.GetAccess(c.Request.Context(), uid)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, access)
}

func (h *DistributionHandler) TrackPromotionVisit(c *gin.Context) {
	var req service.DistributionPromotionVisitInput
	if err := c.ShouldBindJSON(&req); err != nil {
		if _, ok := extractMaxBytesError(err); ok {
			response.Error(c, http.StatusRequestEntityTooLarge, "request body exceeds 16 KiB limit")
			return
		}
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if token, err := c.Cookie(distributionVisitorCookie); err == nil {
		req.VisitorToken = token
	}
	// Match the public rate limiter's Gin trusted-proxy chain. Promotion
	// fingerprints must never accept the legacy raw forwarded-header override.
	req.ClientIP = ip.GetTrustedClientIP(c)
	req.UserAgent = c.GetHeader("User-Agent")
	result, err := h.service.TrackPromotionVisit(c.Request.Context(), req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if result.Tracked && result.VisitorToken != "" {
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(distributionVisitorCookie, result.VisitorToken, result.AttributionDays*24*60*60, "/", "", isRequestHTTPS(c), true)
	} else if !result.Tracked {
		clearDistributionVisitorCookie(c)
	}
	response.Success(c, result)
}

func (h *DistributionHandler) GetPromotionTrackingStatus(c *gin.Context) {
	enabled, err := h.service.PromotionTrackingEnabled(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if !enabled {
		clearDistributionVisitorCookie(c)
	}
	response.Success(c, gin.H{"enabled": enabled})
}

func clearDistributionVisitorCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(distributionVisitorCookie, "", -1, "/", "", isRequestHTTPS(c), true)
}

func parsePromotionStatsFilter(c *gin.Context) service.DistributionPromotionStatsFilter {
	page, pageSize := response.ParsePagination(c)
	filter := service.DistributionPromotionStatsFilter{Page: page, PageSize: pageSize,
		Source: strings.TrimSpace(c.Query("source")), Device: strings.TrimSpace(c.Query("device")),
		AttributionType: strings.TrimSpace(c.Query("attribution_type"))}
	location := time.FixedZone("Asia/Hong_Kong", 8*60*60)
	if value, err := time.ParseInLocation("2006-01-02", c.Query("date_from"), location); err == nil {
		filter.From = value
	} else if c.Query("date_from") != "" {
		filter.From, filter.To = time.Unix(1, 0), time.Unix(1, 0)
	}
	if value, err := time.ParseInLocation("2006-01-02", c.Query("date_to"), location); err == nil {
		filter.To = value.Add(24 * time.Hour)
	} else if c.Query("date_to") != "" {
		filter.From, filter.To = time.Unix(1, 0), time.Unix(1, 0)
	}
	return filter
}

func (h *DistributionHandler) GetPromotionAnalytics(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	value, err := h.service.GetPromotionAnalytics(c.Request.Context(), uid, parsePromotionStatsFilter(c), false)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, value)
}

func (h *DistributionHandler) ListPromotionVisits(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	filter := parsePromotionStatsFilter(c)
	items, total, err := h.service.ListPromotionVisits(c.Request.Context(), uid, filter, false)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, filter.Page, filter.PageSize)
}
func (h *DistributionHandler) GetPayoutAccount(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	v, err := h.service.GetPayoutAccount(c.Request.Context(), uid)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, v)
}
func (h *DistributionHandler) GetSettlementRules(c *gin.Context) {
	_, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	rules, err := h.service.GetSettlementRules(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, rules)
}
func (h *DistributionHandler) UpdatePayoutAccount(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	var req service.DistributionPayoutAccount
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.service.UpsertPayoutAccount(c.Request.Context(), uid, req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": true})
}
func (h *DistributionHandler) RequestWithdrawal(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	var req struct {
		Amount decimal.Decimal `json:"amount_cny" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	v, err := h.service.RequestWithdrawal(c.Request.Context(), uid, req.Amount)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, v)
}
func (h *DistributionHandler) ListWithdrawals(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	v, total, err := h.service.ListWithdrawals(c.Request.Context(), uid, service.DistributionUserListFilter{
		Page: page, PageSize: pageSize, Status: c.Query("status"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, v, total, page, pageSize)
}
func (h *DistributionHandler) ConvertToBalance(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	var req struct {
		Amount decimal.Decimal `json:"amount_cny" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	credit, err := h.service.ConvertToBalance(c.Request.Context(), uid, req.Amount)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"credited_platform_usd": credit})
}
func (h *DistributionHandler) GrantL2Agent(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	var req service.DistributionGrantChildAgentInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	agent, err := h.service.GrantL2Agent(c.Request.Context(), uid, req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, agent)
}

func (h *DistributionHandler) GetTeamAgentRewardRule(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	agentID, err := strconv.ParseInt(c.Param("agent_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid agent id")
		return
	}
	rule, err := h.service.GetAgentRewardRule(c.Request.Context(), uid, agentID, false)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, rule)
}

func (h *DistributionHandler) UpdateTeamAgentRewardRule(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	agentID, err := strconv.ParseInt(c.Param("agent_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid agent id")
		return
	}
	var req service.DistributionRewardRuleInput
	if err = c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err = h.service.UpdateAgentRewardRule(c.Request.Context(), uid, agentID, false, req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": true})
}
func (h *DistributionHandler) ListCustomers(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	v, total, err := h.service.ListCustomers(c.Request.Context(), uid, service.DistributionUserListFilter{
		Page: page, PageSize: pageSize, Search: c.Query("search"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, v, total, page, pageSize)
}
func (h *DistributionHandler) ListCommissions(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	v, total, err := h.service.ListCommissions(c.Request.Context(), uid, service.DistributionUserListFilter{
		Page: page, PageSize: pageSize, Search: c.Query("search"), Status: c.Query("status"), EntryType: c.Query("entry_type"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, v, total, page, pageSize)
}
func (h *DistributionHandler) ListTeam(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	analyticsFilter, filterErr := distributionAnalyticsFilter(c)
	if filterErr != nil {
		response.BadRequest(c, "invalid analytics date range")
		return
	}
	v, total, err := h.service.ListTeam(c.Request.Context(), uid, service.DistributionUserListFilter{
		Page: page, PageSize: pageSize, Search: c.Query("search"), Status: c.Query("status"), Days: analyticsFilter.Days, DateFrom: analyticsFilter.DateFrom, DateTo: analyticsFilter.DateTo,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, v, total, page, pageSize)
}

func (h *DistributionHandler) GetTeamAgentAnalytics(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	agentID, err := strconv.ParseInt(c.Param("agent_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid agent id")
		return
	}
	filter, err := distributionAnalyticsFilter(c)
	if err != nil {
		response.BadRequest(c, "invalid analytics date range")
		return
	}
	result, err := h.service.GetTeamAgentAnalytics(c.Request.Context(), uid, agentID, filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}
func (h *DistributionHandler) UpdateTeamAgentStatus(c *gin.Context) {
	uid, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	agentID, err := strconv.ParseInt(c.Param("agent_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid agent id")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err = h.service.UpdateTeamAgentStatus(c.Request.Context(), uid, agentID, req.Status); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": true})
}

// GetOverview returns the current agent's identity, wallet and team counts.
func (h *DistributionHandler) GetOverview(c *gin.Context) {
	userID, ok := distributionAgentUserID(c, h.service)
	if !ok {
		return
	}
	filter, filterErr := distributionAnalyticsFilter(c)
	if filterErr != nil {
		response.BadRequest(c, "invalid analytics date range")
		return
	}
	overview, err := h.service.GetOverview(c.Request.Context(), userID, filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, overview)
}
