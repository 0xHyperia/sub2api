package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type DistributionHandler struct{ service *service.DistributionService }

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
	v, total, err := h.service.ListTeam(c.Request.Context(), uid, service.DistributionUserListFilter{
		Page: page, PageSize: pageSize, Search: c.Query("search"), Status: c.Query("status"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, v, total, page, pageSize)
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
	overview, err := h.service.GetOverview(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, overview)
}
