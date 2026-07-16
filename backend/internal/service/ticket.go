package service

import (
	"context"
	"io"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	TicketStatusOpen     = "open"
	TicketStatusAnswered = "answered"
	TicketStatusClosed   = "closed"

	TicketSenderUser   = "user"
	TicketSenderAdmin  = "admin"
	TicketSenderSystem = "system"
)

var (
	ErrTicketNotFound         = infraerrors.NotFound("TICKET_NOT_FOUND", "ticket not found")
	ErrTicketCategoryNotFound = infraerrors.NotFound("TICKET_CATEGORY_NOT_FOUND", "ticket category not found")
	ErrTicketCategoryInactive = infraerrors.BadRequest("TICKET_CATEGORY_INACTIVE", "ticket category is inactive")
	ErrTicketClosed           = infraerrors.Conflict("TICKET_CLOSED", "reopen the ticket before replying")
	ErrTicketOpenLimit        = infraerrors.Conflict("TICKET_OPEN_LIMIT", "too many open tickets")
	ErrTicketContentRequired  = infraerrors.BadRequest("TICKET_CONTENT_REQUIRED", "message content or attachment is required")
	ErrTicketContentTooLong   = infraerrors.BadRequest("TICKET_CONTENT_TOO_LONG", "message content is too long")
	ErrTicketAttachmentOff    = infraerrors.BadRequest("TICKET_ATTACHMENTS_UNAVAILABLE", "ticket attachments are unavailable")
	ErrTicketAttachmentType   = infraerrors.BadRequest("TICKET_ATTACHMENT_TYPE", "attachment type is not allowed")
	ErrTicketAttachmentLimit  = infraerrors.BadRequest("TICKET_ATTACHMENT_LIMIT", "attachment limit exceeded")
	ErrTicketCategoryConflict = infraerrors.Conflict("TICKET_CATEGORY_CONFLICT", "ticket category code already exists")
	ErrTicketCategoryInvalid  = infraerrors.BadRequest("TICKET_CATEGORY_INVALID", "ticket category is invalid")
)

type TicketCategory struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	NameZH    string    `json:"name_zh"`
	NameEN    string    `json:"name_en"`
	Active    bool      `json:"active"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TicketAttachment struct {
	ID           int64     `json:"id"`
	MessageID    int64     `json:"message_id"`
	ObjectKey    string    `json:"-"`
	OriginalName string    `json:"original_name"`
	ContentType  string    `json:"content_type"`
	SizeBytes    int64     `json:"size_bytes"`
	SHA256       string    `json:"sha256"`
	CreatedAt    time.Time `json:"created_at"`
}

type TicketMessage struct {
	ID           int64              `json:"id"`
	TicketID     int64              `json:"ticket_id"`
	SenderUserID *int64             `json:"sender_user_id,omitempty"`
	SenderType   string             `json:"sender_type"`
	EventType    string             `json:"event_type"`
	Content      string             `json:"content"`
	Attachments  []TicketAttachment `json:"attachments"`
	CreatedAt    time.Time          `json:"created_at"`
}

type Ticket struct {
	ID               int64           `json:"id"`
	Number           string          `json:"number"`
	UserID           int64           `json:"user_id"`
	UserEmail        string          `json:"user_email,omitempty"`
	UserName         string          `json:"user_name,omitempty"`
	CategoryID       int64           `json:"category_id"`
	Category         TicketCategory  `json:"category"`
	Subject          string          `json:"subject"`
	Status           string          `json:"status"`
	UserUnreadCount  int             `json:"user_unread_count"`
	AdminUnreadCount int             `json:"admin_unread_count"`
	LastActorType    string          `json:"last_actor_type"`
	LastMessageAt    time.Time       `json:"last_message_at"`
	ClosedAt         *time.Time      `json:"closed_at,omitempty"`
	ClosedByUserID   *int64          `json:"closed_by_user_id,omitempty"`
	ClosedByRole     *string         `json:"closed_by_role,omitempty"`
	Messages         []TicketMessage `json:"messages,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type TicketListFilters struct {
	UserID     *int64
	Status     string
	CategoryID int64
	Search     string
	UnreadOnly bool
}

type TicketUpload struct {
	Name        string
	ContentType string
	Data        []byte
}

type TicketAttachmentCapabilities struct {
	AttachmentsAvailable bool     `json:"attachments_available"`
	MaxFileBytes         int64    `json:"max_file_bytes"`
	MaxFilesPerMessage   int      `json:"max_files_per_message"`
	MaxTotalBytes        int64    `json:"max_total_bytes"`
	AllowedExtensions    []string `json:"allowed_extensions"`
}

type TicketRepository interface {
	ListCategories(ctx context.Context, activeOnly bool) ([]TicketCategory, error)
	GetCategory(ctx context.Context, id int64) (*TicketCategory, error)
	CreateCategory(ctx context.Context, category *TicketCategory) error
	UpdateCategory(ctx context.Context, category *TicketCategory) error
	ReorderCategories(ctx context.Context, ids []int64) error
	CountActiveByUser(ctx context.Context, userID int64) (int, error)
	Create(ctx context.Context, ticket *Ticket, message *TicketMessage, attachments []TicketAttachment) error
	List(ctx context.Context, params pagination.PaginationParams, filters TicketListFilters) ([]Ticket, *pagination.PaginationResult, error)
	Get(ctx context.Context, number string, userID *int64) (*Ticket, error)
	AppendMessage(ctx context.Context, number string, userID int64, isAdmin bool, content string, attachments []TicketAttachment) (*Ticket, error)
	ChangeStatus(ctx context.Context, number string, userID int64, isAdmin bool, reopen bool) (*Ticket, error)
	MarkRead(ctx context.Context, number string, userID int64, isAdmin bool) error
	UnreadCount(ctx context.Context, userID *int64) (int, error)
	GetAttachment(ctx context.Context, id int64, userID int64, isAdmin bool) (*TicketAttachment, error)
}

type TicketStorage interface {
	Active() bool
	Put(ctx context.Context, key, contentType string, body io.Reader, size int64) error
	Delete(ctx context.Context, key string) error
	PresignGet(ctx context.Context, key, filename, contentType string) (string, error)
}
