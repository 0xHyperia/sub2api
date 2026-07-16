package admin

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type TicketHandler struct{ service *service.TicketService }

func NewTicketHandler(service *service.TicketService) *TicketHandler {
	return &TicketHandler{service: service}
}

func (h *TicketHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	categoryID, _ := strconv.ParseInt(c.Query("category_id"), 10, 64)
	filters := service.TicketListFilters{Status: strings.TrimSpace(c.Query("status")), CategoryID: categoryID, Search: strings.TrimSpace(c.Query("search")), UnreadOnly: c.Query("unread_only") == "1" || c.Query("unread_only") == "true"}
	items, result, err := h.service.List(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: pageSize}, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}
func (h *TicketHandler) Get(c *gin.Context) {
	item, err := h.service.Get(c.Request.Context(), c.Param("number"), nil)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
func (h *TicketHandler) Reply(c *gin.Context) {
	subject, ok := adminTicketSubject(c)
	if !ok {
		return
	}
	uploads, err := parseAdminTicketMultipart(c, h.service.Capabilities())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	item, err := h.service.Reply(c.Request.Context(), c.Param("number"), subject.UserID, true, c.PostForm("content"), uploads)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
func (h *TicketHandler) Close(c *gin.Context)  { h.changeStatus(c, false) }
func (h *TicketHandler) Reopen(c *gin.Context) { h.changeStatus(c, true) }
func (h *TicketHandler) changeStatus(c *gin.Context, reopen bool) {
	subject, ok := adminTicketSubject(c)
	if !ok {
		return
	}
	item, err := h.service.ChangeStatus(c.Request.Context(), c.Param("number"), subject.UserID, true, reopen)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
func (h *TicketHandler) MarkRead(c *gin.Context) {
	subject, ok := adminTicketSubject(c)
	if !ok {
		return
	}
	if err := h.service.MarkRead(c.Request.Context(), c.Param("number"), subject.UserID, true); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}
func (h *TicketHandler) Unread(c *gin.Context) {
	count, err := h.service.UnreadCount(c.Request.Context(), nil)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"count": count})
}

type ticketCategoryRequest struct {
	Code      string `json:"code"`
	NameZH    string `json:"name_zh" binding:"required"`
	NameEN    string `json:"name_en"`
	Active    *bool  `json:"active"`
	SortOrder int    `json:"sort_order"`
}

func (h *TicketHandler) ListCategories(c *gin.Context) {
	items, err := h.service.ListCategories(c.Request.Context(), false)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}
func (h *TicketHandler) CreateCategory(c *gin.Context) {
	var req ticketCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	item := &service.TicketCategory{Code: req.Code, NameZH: req.NameZH, NameEN: req.NameEN, Active: active, SortOrder: req.SortOrder}
	if err := h.service.CreateCategory(c.Request.Context(), item); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, item)
}
func (h *TicketHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid category id")
		return
	}
	current, err := h.service.ListCategories(c.Request.Context(), false)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var item *service.TicketCategory
	for i := range current {
		if current[i].ID == id {
			copy := current[i]
			item = &copy
			break
		}
	}
	if item == nil {
		response.ErrorFrom(c, service.ErrTicketCategoryNotFound)
		return
	}
	var req ticketCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	item.NameZH, item.NameEN, item.SortOrder = req.NameZH, req.NameEN, req.SortOrder
	if req.Active != nil {
		item.Active = *req.Active
	}
	if err := h.service.UpdateCategory(c.Request.Context(), item); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
func (h *TicketHandler) ReorderCategories(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}
	if err := h.service.ReorderCategories(c.Request.Context(), req.IDs); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}
func adminTicketSubject(c *gin.Context) (middleware2.AuthSubject, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
	}
	return subject, ok
}

func parseAdminTicketMultipart(c *gin.Context, capabilities service.TicketAttachmentCapabilities) ([]service.TicketUpload, error) {
	limit := capabilities.MaxTotalBytes + 2*1024*1024
	if limit < 2*1024*1024 {
		limit = 2 * 1024 * 1024
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
	if err := c.Request.ParseMultipartForm(limit); err != nil {
		return nil, service.ErrTicketAttachmentLimit
	}
	files := c.Request.MultipartForm.File["files"]
	if len(files) == 0 {
		return []service.TicketUpload{}, nil
	}
	if !capabilities.AttachmentsAvailable {
		return nil, service.ErrTicketAttachmentOff
	}
	if len(files) > capabilities.MaxFilesPerMessage {
		return nil, service.ErrTicketAttachmentLimit
	}
	uploads := make([]service.TicketUpload, 0, len(files))
	var total int64
	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			return nil, err
		}
		data, readErr := io.ReadAll(io.LimitReader(file, capabilities.MaxFileBytes+1))
		_ = file.Close()
		if readErr != nil {
			return nil, readErr
		}
		total += int64(len(data))
		if int64(len(data)) > capabilities.MaxFileBytes || total > capabilities.MaxTotalBytes {
			return nil, service.ErrTicketAttachmentLimit
		}
		uploads = append(uploads, service.TicketUpload{Name: header.Filename, ContentType: header.Header.Get("Content-Type"), Data: data})
	}
	return uploads, nil
}
