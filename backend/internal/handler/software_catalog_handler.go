package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type SoftwareCatalogHandler struct {
	service  *service.SoftwareCatalogService
	settings *service.SettingService
}

func NewSoftwareCatalogHandler(catalogService *service.SoftwareCatalogService, settings *service.SettingService) *SoftwareCatalogHandler {
	return &SoftwareCatalogHandler{service: catalogService, settings: settings}
}

func (h *SoftwareCatalogHandler) List(c *gin.Context) {
	if !h.settings.IsSoftwareCenterEnabled(c.Request.Context()) {
		response.Success(c, []service.SoftwareCatalogItem{})
		return
	}
	items, err := h.service.ListPublic(c.Request.Context())
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, items)
}
