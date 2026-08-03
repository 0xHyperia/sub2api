package admin

import (
	"strconv"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ModelMonitorHandler struct {
	service  *service.ModelMonitorService
	settings *service.SettingService
}

func NewModelMonitorHandler(modelMonitorService *service.ModelMonitorService, settings *service.SettingService) *ModelMonitorHandler {
	return &ModelMonitorHandler{service: modelMonitorService, settings: settings}
}

func (h *ModelMonitorHandler) List(c *gin.Context) {
	resolution := service.ModelMonitorResolution(strings.ToLower(strings.TrimSpace(c.DefaultQuery("resolution", "minute"))))
	rows, err := h.service.ListRows(c.Request.Context(), resolution)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": rows})
}

type modelMonitorGroupConfigRequest struct {
	Platform        string `json:"platform" binding:"required"`
	Model           string `json:"model" binding:"required"`
	GroupID         int64  `json:"group_id" binding:"required"`
	Enabled         bool   `json:"enabled"`
	IntervalSeconds int    `json:"interval_seconds" binding:"required"`
}

func (h *ModelMonitorHandler) ConfigureGroup(c *gin.Context) {
	var req modelMonitorGroupConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	subject, _ := middleware2.GetAuthSubjectFromContext(c)
	m, err := h.service.ConfigureGroup(c.Request.Context(), req.Platform, req.Model, req.GroupID, req.Enabled, req.IntervalSeconds, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	response.Success(c, m)
}

type modelMonitorConfigRequest struct {
	Platform        string `json:"platform" binding:"required"`
	Model           string `json:"model" binding:"required"`
	Enabled         bool   `json:"enabled"`
	IntervalSeconds int    `json:"interval_seconds"`
	DisplayOrder    int    `json:"display_order"`
	Label           string `json:"label"`
}

func (h *ModelMonitorHandler) Upsert(c *gin.Context) {
	var req modelMonitorConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	subject, _ := middleware2.GetAuthSubjectFromContext(c)
	m, err := h.service.Upsert(c.Request.Context(), req.Platform, req.Model, req.Enabled, req.IntervalSeconds, req.DisplayOrder, req.Label, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	response.Success(c, m)
}

type modelMonitorGroupsRequest struct {
	Platform string  `json:"platform" binding:"required"`
	Model    string  `json:"model" binding:"required"`
	GroupIDs []int64 `json:"group_ids" binding:"required,min=1"`
}

func (h *ModelMonitorHandler) ConfigureGroups(c *gin.Context) {
	var req modelMonitorGroupsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	subject, _ := middleware2.GetAuthSubjectFromContext(c)
	m, err := h.service.ConfigureGroups(c.Request.Context(), strings.TrimSpace(req.Platform), strings.TrimSpace(req.Model), req.GroupIDs, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	response.Success(c, m)
}

type modelMonitorRunRequest struct {
	Platform string `json:"platform" binding:"required"`
	Model    string `json:"model" binding:"required"`
	GroupID  int64  `json:"group_id"`
}

func (h *ModelMonitorHandler) Run(c *gin.Context) {
	if h.settings == nil || !h.settings.GetModelMonitorRuntime(c.Request.Context()).Enabled {
		response.ErrorFrom(c, infraerrors.BadRequest("MODEL_MONITOR_DISABLED", "model monitor is disabled"))
		return
	}
	var req modelMonitorRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	subject, _ := middleware2.GetAuthSubjectFromContext(c)
	var result *service.ModelMonitorHistory
	var err error
	if req.GroupID > 0 {
		result, err = h.service.RunGroupByKey(c.Request.Context(), strings.TrimSpace(req.Platform), strings.TrimSpace(req.Model), req.GroupID, subject.UserID)
	} else {
		result, err = h.service.RunByKey(c.Request.Context(), strings.TrimSpace(req.Platform), strings.TrimSpace(req.Model), subject.UserID)
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ModelMonitorHandler) History(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", "invalid monitor id"))
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	items, err := h.service.History(c.Request.Context(), id, limit)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": items})
}
