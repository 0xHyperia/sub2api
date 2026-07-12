package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type AppResourceHandler struct {
	userService         *service.UserService
	apiKeyService       *service.APIKeyService
	subscriptionService *service.SubscriptionService
}

func NewAppResourceHandler(userService *service.UserService, apiKeyService *service.APIKeyService, subscriptionService *service.SubscriptionService) *AppResourceHandler {
	return &AppResourceHandler{userService: userService, apiKeyService: apiKeyService, subscriptionService: subscriptionService}
}

func (h *AppResourceHandler) Me(c *gin.Context) {
	subject, ok := appSubject(c)
	if !ok {
		return
	}
	user, err := h.userService.GetByID(c.Request.Context(), subject.UserID)
	if err != nil {
		appResourceError(c, http.StatusNotFound, "user_not_found", "User not found")
		return
	}
	appResourceSuccess(c, gin.H{
		"id": user.ID, "email": user.Email, "username": user.Username, "role": user.Role,
		"balance": user.Balance, "frozen_balance": user.FrozenBalance, "concurrency": user.Concurrency,
		"rpm_limit": user.RPMLimit, "status": user.Status, "allowed_groups": user.AllowedGroups,
		"created_at": user.CreatedAt, "last_active_at": user.LastActiveAt,
	})
}

func (h *AppResourceHandler) Groups(c *gin.Context) {
	subject, ok := appSubject(c)
	if !ok {
		return
	}
	groups, err := h.apiKeyService.GetAvailableGroups(c.Request.Context(), subject.UserID)
	if err != nil {
		appResourceError(c, http.StatusInternalServerError, "groups_unavailable", "Failed to list groups")
		return
	}
	rates, err := h.apiKeyService.GetUserGroupRates(c.Request.Context(), subject.UserID)
	if err != nil {
		appResourceError(c, http.StatusInternalServerError, "groups_unavailable", "Failed to resolve group rates")
		return
	}
	items := make([]gin.H, 0, len(groups))
	for _, group := range groups {
		userRate, effective := resolveAppGroupRates(group.RateMultiplier, rates[group.ID], hasMapKey(rates, group.ID))
		items = append(items, gin.H{
			"id": group.ID, "name": group.Name, "description": group.Description, "platform": group.Platform,
			"status": group.Status, "subscription_type": group.SubscriptionType,
			"rate_multiplier": group.RateMultiplier, "user_rate_multiplier": userRate, "effective_rate_multiplier": effective,
		})
	}
	appResourceSuccess(c, items)
}

func resolveAppGroupRates(groupRate, override float64, hasOverride bool) (userRate, effective float64) {
	if hasOverride {
		return override, override
	}
	return 1, groupRate
}

func hasMapKey(values map[int64]float64, key int64) bool {
	_, ok := values[key]
	return ok
}

func (h *AppResourceHandler) Keys(c *gin.Context) {
	subject, ok := appSubject(c)
	if !ok {
		return
	}
	page, pageSize := parseAppPagination(c)
	keys, result, err := h.apiKeyService.List(c.Request.Context(), subject.UserID, pagination.PaginationParams{Page: page, PageSize: pageSize, SortBy: "created_at", SortOrder: "desc"}, service.APIKeyListFilters{})
	if err != nil {
		appResourceError(c, http.StatusInternalServerError, "keys_unavailable", "Failed to list API keys")
		return
	}
	items := make([]dto.APIKey, 0, len(keys))
	for i := range keys {
		items = append(items, *dto.APIKeyFromService(&keys[i]))
	}
	appResourceSuccess(c, gin.H{"items": items, "total": result.Total})
}

type createAppKeyRequest struct {
	Name string `json:"name"`
}

func (h *AppResourceHandler) CreateKey(c *gin.Context) {
	subject, ok := appSubject(c)
	if !ok {
		return
	}
	groupID, err := strconv.ParseInt(c.Param("groupId"), 10, 64)
	if err != nil || groupID <= 0 {
		appResourceError(c, http.StatusBadRequest, "invalid_group", "Invalid group ID")
		return
	}
	var request createAppKeyRequest
	if err := decodeStrictJSON(c, &request); err != nil || strings.TrimSpace(request.Name) == "" || len(request.Name) > 100 {
		appResourceError(c, http.StatusBadRequest, "invalid_request", "Only a non-empty name is accepted")
		return
	}
	key, err := h.apiKeyService.Create(c.Request.Context(), subject.UserID, service.CreateAPIKeyRequest{Name: strings.TrimSpace(request.Name), GroupID: &groupID})
	if err != nil {
		appResourceError(c, http.StatusBadRequest, "key_create_failed", err.Error())
		return
	}
	appResourceSuccess(c, dto.APIKeyFromService(key))
}

func (h *AppResourceHandler) Subscriptions(c *gin.Context) {
	subject, ok := appSubject(c)
	if !ok {
		return
	}
	subscriptions, err := h.subscriptionService.ListUserSubscriptions(c.Request.Context(), subject.UserID)
	if err != nil {
		appResourceError(c, http.StatusInternalServerError, "subscriptions_unavailable", "Failed to list subscriptions")
		return
	}
	items := make([]dto.UserSubscription, 0, len(subscriptions))
	for i := range subscriptions {
		items = append(items, *dto.UserSubscriptionFromService(&subscriptions[i]))
	}
	appResourceSuccess(c, items)
}

func appSubject(c *gin.Context) (servermiddleware.AppAuthSubject, bool) {
	subject, ok := servermiddleware.GetAppAuthSubject(c)
	if !ok {
		appResourceError(c, http.StatusUnauthorized, "invalid_token", "App authorization is required")
	}
	return subject, ok
}

func parseAppPagination(c *gin.Context) (int, int) {
	page, pageSize := 1, 100
	if value, err := strconv.Atoi(c.Query("page")); err == nil && value > 0 {
		page = value
	}
	if value, err := strconv.Atoi(c.Query("page_size")); err == nil && value > 0 && value <= 100 {
		pageSize = value
	}
	return page, pageSize
}

func appResourceSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data, "message": "success"})
}

func appResourceError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"success": false, "message": message, "error": gin.H{"code": code, "message": message}})
}
