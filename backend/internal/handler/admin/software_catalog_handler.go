package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type SoftwareCatalogHandler struct {
	service *service.SoftwareCatalogService
}

func NewSoftwareCatalogHandler(catalogService *service.SoftwareCatalogService) *SoftwareCatalogHandler {
	return &SoftwareCatalogHandler{service: catalogService}
}

type softwarePreviewRequest struct {
	RepositoryURL string `json:"repository_url" binding:"required,max=500"`
}
type createSoftwareRequest struct {
	SourceType     string                          `json:"source_type" binding:"omitempty,oneof=github manual"`
	RepositoryURL  string                          `json:"repository_url" binding:"omitempty,max=500"`
	SourceURL      string                          `json:"source_url" binding:"omitempty,max=1000"`
	Name           string                          `json:"name" binding:"max=120"`
	Description    string                          `json:"description" binding:"max=2000"`
	LogoURL        string                          `json:"logo_url" binding:"max=1000"`
	Featured       bool                            `json:"featured"`
	Enabled        bool                            `json:"enabled"`
	SortOrder      int                             `json:"sort_order" binding:"min=-10000,max=10000"`
	Version        string                          `json:"version" binding:"max=120"`
	ReleaseName    string                          `json:"release_name" binding:"max=255"`
	ReleaseNotes   string                          `json:"release_notes"`
	PublishedAt    *time.Time                      `json:"published_at"`
	DownloadAssets []service.SoftwareDownloadAsset `json:"asset_variants"`
}
type updateSoftwareRequest struct {
	Name           *string                          `json:"name" binding:"omitempty,max=120"`
	Description    *string                          `json:"description" binding:"omitempty,max=2000"`
	LogoURL        *string                          `json:"logo_url" binding:"omitempty,max=1000"`
	Featured       *bool                            `json:"featured"`
	Enabled        *bool                            `json:"enabled"`
	SortOrder      *int                             `json:"sort_order" binding:"omitempty,min=-10000,max=10000"`
	SourceURL      *string                          `json:"source_url" binding:"omitempty,max=1000"`
	Version        *string                          `json:"version" binding:"omitempty,max=120"`
	ReleaseName    *string                          `json:"release_name" binding:"omitempty,max=255"`
	ReleaseNotes   *string                          `json:"release_notes"`
	PublishedAt    **time.Time                      `json:"published_at"`
	DownloadAssets *[]service.SoftwareDownloadAsset `json:"asset_variants"`
}

func (h *SoftwareCatalogHandler) List(c *gin.Context) {
	items, err := h.service.ListAdmin(c.Request.Context())
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, items)
}

func (h *SoftwareCatalogHandler) Preview(c *gin.Context) {
	var req softwarePreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入有效的 GitHub 项目地址")
		return
	}
	preview, err := h.service.Preview(c.Request.Context(), req.RepositoryURL)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, preview)
}

func (h *SoftwareCatalogHandler) Create(c *gin.Context) {
	var req createSoftwareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "上架信息格式不正确: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Name) == "" && req.Name != "" {
		response.BadRequest(c, "软件名称不能为空")
		return
	}
	item, err := h.service.Create(c.Request.Context(), service.CreateSoftwareCatalogInput{
		SourceType: req.SourceType, RepositoryURL: req.RepositoryURL, SourceURL: req.SourceURL,
		Name: req.Name, Description: req.Description, LogoURL: req.LogoURL,
		Featured: req.Featured, Enabled: req.Enabled, SortOrder: req.SortOrder,
		Version: req.Version, ReleaseName: req.ReleaseName, ReleaseNotes: req.ReleaseNotes,
		PublishedAt: req.PublishedAt, DownloadAssets: req.DownloadAssets,
	})
	if response.ErrorFrom(c, err) {
		return
	}
	response.Created(c, item)
}

func (h *SoftwareCatalogHandler) Update(c *gin.Context) {
	id, ok := softwareID(c)
	if !ok {
		return
	}
	var req updateSoftwareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "软件信息格式不正确: "+err.Error())
		return
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		response.BadRequest(c, "软件名称不能为空")
		return
	}
	item, err := h.service.Update(c.Request.Context(), id, service.UpdateSoftwareCatalogInput{
		Name: req.Name, Description: req.Description, LogoURL: req.LogoURL, Featured: req.Featured,
		Enabled: req.Enabled, SortOrder: req.SortOrder,
		SourceURL: req.SourceURL, Version: req.Version, ReleaseName: req.ReleaseName,
		ReleaseNotes: req.ReleaseNotes, PublishedAt: req.PublishedAt, DownloadAssets: req.DownloadAssets,
	})
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, item)
}

func (h *SoftwareCatalogHandler) Delete(c *gin.Context) {
	id, ok := softwareID(c)
	if !ok {
		return
	}
	if response.ErrorFrom(c, h.service.Delete(c.Request.Context(), id)) {
		return
	}
	response.Success(c, gin.H{"message": "软件已移除"})
}

func (h *SoftwareCatalogHandler) Refresh(c *gin.Context) {
	id, ok := softwareID(c)
	if !ok {
		return
	}
	item, err := h.service.Refresh(c.Request.Context(), id)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, item)
}

func softwareID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "无效的软件 ID")
		return 0, false
	}
	return id, true
}
