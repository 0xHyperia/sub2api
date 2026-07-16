package handler

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

func (h *TicketHandler) ListCategories(c *gin.Context) {
	items, err := h.service.ListCategories(c.Request.Context(), true)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}
func (h *TicketHandler) Capabilities(c *gin.Context) { response.Success(c, h.service.Capabilities()) }

func (h *TicketHandler) List(c *gin.Context) {
	subject, ok := ticketSubjectFromContext(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	categoryID, _ := strconv.ParseInt(c.Query("category_id"), 10, 64)
	items, result, err := h.service.List(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: pageSize}, service.TicketListFilters{UserID: &subject.UserID, Status: strings.TrimSpace(c.Query("status")), CategoryID: categoryID, Search: strings.TrimSpace(c.Query("search")), UnreadOnly: parseBoolQuery(c.Query("unread_only"))})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}

func (h *TicketHandler) Create(c *gin.Context) {
	subject, ok := ticketSubjectFromContext(c)
	if !ok {
		return
	}
	uploads, err := ParseTicketMultipart(c, h.service.Capabilities())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	categoryID, err := strconv.ParseInt(c.PostForm("category_id"), 10, 64)
	if err != nil || categoryID <= 0 {
		response.BadRequest(c, "invalid category_id")
		return
	}
	item, err := h.service.Create(c.Request.Context(), subject.UserID, categoryID, c.PostForm("description"), uploads)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, item)
}

func (h *TicketHandler) Get(c *gin.Context) {
	subject, ok := ticketSubjectFromContext(c)
	if !ok {
		return
	}
	item, err := h.service.Get(c.Request.Context(), c.Param("number"), &subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
func (h *TicketHandler) Reply(c *gin.Context) {
	subject, ok := ticketSubjectFromContext(c)
	if !ok {
		return
	}
	uploads, err := ParseTicketMultipart(c, h.service.Capabilities())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	item, err := h.service.Reply(c.Request.Context(), c.Param("number"), subject.UserID, false, c.PostForm("content"), uploads)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
func (h *TicketHandler) Close(c *gin.Context)  { h.changeStatus(c, false) }
func (h *TicketHandler) Reopen(c *gin.Context) { h.changeStatus(c, true) }
func (h *TicketHandler) changeStatus(c *gin.Context, reopen bool) {
	subject, ok := ticketSubjectFromContext(c)
	if !ok {
		return
	}
	item, err := h.service.ChangeStatus(c.Request.Context(), c.Param("number"), subject.UserID, false, reopen)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
func (h *TicketHandler) MarkRead(c *gin.Context) {
	subject, ok := ticketSubjectFromContext(c)
	if !ok {
		return
	}
	if err := h.service.MarkRead(c.Request.Context(), c.Param("number"), subject.UserID, false); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}
func (h *TicketHandler) Unread(c *gin.Context) {
	subject, ok := ticketSubjectFromContext(c)
	if !ok {
		return
	}
	count, err := h.service.UnreadCount(c.Request.Context(), &subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"count": count})
}

func (h *TicketHandler) DownloadAttachment(c *gin.Context) {
	subject, ok := ticketSubjectFromContext(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid attachment id")
		return
	}
	role, _ := middleware2.GetUserRoleFromContext(c)
	url, err := h.service.AttachmentURL(c.Request.Context(), id, subject.UserID, role == service.RoleAdmin)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"url": url})
}

func ParseTicketMultipart(c *gin.Context, capabilities service.TicketAttachmentCapabilities) ([]service.TicketUpload, error) {
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

func ticketSubjectFromContext(c *gin.Context) (middleware2.AuthSubject, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
	}
	return subject, ok
}
