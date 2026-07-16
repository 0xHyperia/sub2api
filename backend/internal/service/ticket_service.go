package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const maxTicketContentRunes = 10000

var ticketCategoryCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,63}$`)

type TicketService struct {
	repo    TicketRepository
	storage TicketStorage
	cfg     config.TicketStorageConfig
}

func NewTicketService(repo TicketRepository, storage TicketStorage, cfg *config.Config) *TicketService {
	return &TicketService{repo: repo, storage: storage, cfg: cfg.TicketStorage}
}

func (s *TicketService) Capabilities() TicketAttachmentCapabilities {
	return TicketAttachmentCapabilities{
		AttachmentsAvailable: s.storage != nil && s.storage.Active(),
		MaxFileBytes:         s.cfg.MaxFileBytes, MaxFilesPerMessage: s.cfg.MaxFilesPerMessage, MaxTotalBytes: s.cfg.MaxTotalBytes,
		AllowedExtensions: []string{".png", ".jpg", ".jpeg", ".webp", ".pdf", ".txt", ".log", ".json", ".zip"},
	}
}

func (s *TicketService) ListCategories(ctx context.Context, activeOnly bool) ([]TicketCategory, error) {
	return s.repo.ListCategories(ctx, activeOnly)
}

func (s *TicketService) CreateCategory(ctx context.Context, category *TicketCategory) error {
	if category == nil || !ticketCategoryCodePattern.MatchString(strings.TrimSpace(category.Code)) || strings.TrimSpace(category.NameZH) == "" {
		return ErrTicketCategoryInvalid
	}
	category.Code = strings.TrimSpace(category.Code)
	category.NameZH = strings.TrimSpace(category.NameZH)
	category.NameEN = strings.TrimSpace(category.NameEN)
	return s.repo.CreateCategory(ctx, category)
}

func (s *TicketService) UpdateCategory(ctx context.Context, category *TicketCategory) error {
	if category == nil || category.ID <= 0 || strings.TrimSpace(category.NameZH) == "" {
		return ErrTicketCategoryInvalid
	}
	category.NameZH = strings.TrimSpace(category.NameZH)
	category.NameEN = strings.TrimSpace(category.NameEN)
	return s.repo.UpdateCategory(ctx, category)
}

func (s *TicketService) ReorderCategories(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return s.repo.ReorderCategories(ctx, ids)
}

func (s *TicketService) Create(ctx context.Context, userID, categoryID int64, description string, uploads []TicketUpload) (*Ticket, error) {
	description = strings.TrimSpace(description)
	if description == "" {
		return nil, ErrTicketContentRequired
	}
	if utf8.RuneCountInString(description) > maxTicketContentRunes {
		return nil, ErrTicketContentTooLong
	}
	category, err := s.repo.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if !category.Active {
		return nil, ErrTicketCategoryInactive
	}
	count, err := s.repo.CountActiveByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= 10 {
		return nil, ErrTicketOpenLimit
	}

	number, err := newTicketNumber(time.Now())
	if err != nil {
		return nil, err
	}
	attachments, uploadedKeys, err := s.upload(ctx, uploads)
	if err != nil {
		return nil, err
	}
	cleanup := true
	defer func() {
		if cleanup {
			s.cleanup(ctx, uploadedKeys)
		}
	}()

	ticket := &Ticket{Number: number, UserID: userID, CategoryID: categoryID, Category: *category, Subject: ticketSubject(description), Status: TicketStatusOpen, AdminUnreadCount: 1, LastActorType: TicketSenderUser, LastMessageAt: time.Now()}
	message := &TicketMessage{SenderUserID: &userID, SenderType: TicketSenderUser, Content: description, Attachments: attachments}
	if err := s.repo.Create(ctx, ticket, message, attachments); err != nil {
		return nil, err
	}
	cleanup = false
	return s.repo.Get(ctx, ticket.Number, &userID)
}

func (s *TicketService) List(ctx context.Context, params pagination.PaginationParams, filters TicketListFilters) ([]Ticket, *pagination.PaginationResult, error) {
	return s.repo.List(ctx, params, filters)
}

func (s *TicketService) Get(ctx context.Context, number string, userID *int64) (*Ticket, error) {
	return s.repo.Get(ctx, number, userID)
}

func (s *TicketService) Reply(ctx context.Context, number string, userID int64, isAdmin bool, content string, uploads []TicketUpload) (*Ticket, error) {
	content = strings.TrimSpace(content)
	if content == "" && len(uploads) == 0 {
		return nil, ErrTicketContentRequired
	}
	if utf8.RuneCountInString(content) > maxTicketContentRunes {
		return nil, ErrTicketContentTooLong
	}
	attachments, keys, err := s.upload(ctx, uploads)
	if err != nil {
		return nil, err
	}
	cleanup := true
	defer func() {
		if cleanup {
			s.cleanup(ctx, keys)
		}
	}()
	ticket, err := s.repo.AppendMessage(ctx, number, userID, isAdmin, content, attachments)
	if err != nil {
		return nil, err
	}
	cleanup = false
	return ticket, nil
}

func (s *TicketService) ChangeStatus(ctx context.Context, number string, userID int64, isAdmin, reopen bool) (*Ticket, error) {
	return s.repo.ChangeStatus(ctx, number, userID, isAdmin, reopen)
}

func (s *TicketService) MarkRead(ctx context.Context, number string, userID int64, isAdmin bool) error {
	return s.repo.MarkRead(ctx, number, userID, isAdmin)
}

func (s *TicketService) UnreadCount(ctx context.Context, userID *int64) (int, error) {
	return s.repo.UnreadCount(ctx, userID)
}

func (s *TicketService) AttachmentURL(ctx context.Context, id, userID int64, isAdmin bool) (string, error) {
	attachment, err := s.repo.GetAttachment(ctx, id, userID, isAdmin)
	if err != nil {
		return "", err
	}
	if s.storage == nil || !s.storage.Active() {
		return "", ErrTicketAttachmentOff
	}
	return s.storage.PresignGet(ctx, attachment.ObjectKey, attachment.OriginalName, attachment.ContentType)
}

func (s *TicketService) upload(ctx context.Context, uploads []TicketUpload) ([]TicketAttachment, []string, error) {
	if len(uploads) == 0 {
		return []TicketAttachment{}, nil, nil
	}
	capabilities := s.Capabilities()
	if !capabilities.AttachmentsAvailable {
		return nil, nil, ErrTicketAttachmentOff
	}
	if len(uploads) > capabilities.MaxFilesPerMessage {
		return nil, nil, ErrTicketAttachmentLimit
	}
	var total int64
	attachments := make([]TicketAttachment, 0, len(uploads))
	keys := make([]string, 0, len(uploads))
	for _, upload := range uploads {
		size := int64(len(upload.Data))
		total += size
		if size <= 0 || size > capabilities.MaxFileBytes || total > capabilities.MaxTotalBytes {
			s.cleanup(ctx, keys)
			return nil, nil, ErrTicketAttachmentLimit
		}
		contentType, ok := allowedTicketContentType(upload.Name, upload.Data)
		if !ok {
			s.cleanup(ctx, keys)
			return nil, nil, ErrTicketAttachmentType
		}
		token, err := randomToken(20)
		if err != nil {
			s.cleanup(ctx, keys)
			return nil, nil, err
		}
		key := strings.TrimRight(s.cfg.Prefix, "/") + "/" + time.Now().Format("2006/01") + "/" + token
		if err := s.storage.Put(ctx, key, contentType, bytes.NewReader(upload.Data), size); err != nil {
			s.cleanup(ctx, keys)
			return nil, nil, fmt.Errorf("store ticket attachment: %w", err)
		}
		sum := sha256.Sum256(upload.Data)
		attachments = append(attachments, TicketAttachment{ObjectKey: key, OriginalName: sanitizeTicketFilename(upload.Name), ContentType: contentType, SizeBytes: size, SHA256: hex.EncodeToString(sum[:])})
		keys = append(keys, key)
	}
	return attachments, keys, nil
}

func (s *TicketService) cleanup(ctx context.Context, keys []string) {
	if s.storage == nil {
		return
	}
	for _, key := range keys {
		_ = s.storage.Delete(ctx, key)
	}
}

func allowedTicketContentType(name string, data []byte) (string, bool) {
	ext := strings.ToLower(filepath.Ext(name))
	allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".pdf": true, ".txt": true, ".log": true, ".json": true, ".zip": true}
	if !allowed[ext] {
		return "", false
	}
	detected := http.DetectContentType(data)
	if ext == ".log" || ext == ".txt" || ext == ".json" {
		if strings.HasPrefix(detected, "text/plain") || detected == "application/json" {
			if ext == ".json" {
				return "application/json", true
			}
			return "text/plain; charset=utf-8", true
		}
		return "", false
	}
	expected := mime.TypeByExtension(ext)
	if i := strings.IndexByte(expected, ';'); i >= 0 {
		expected = expected[:i]
	}
	if ext == ".jpg" {
		expected = "image/jpeg"
	}
	return expected, detected == expected || (ext == ".zip" && detected == "application/zip")
}

func ticketSubject(content string) string {
	line := strings.TrimSpace(strings.Split(content, "\n")[0])
	runes := []rune(line)
	if len(runes) > 120 {
		line = string(runes[:120])
	}
	return line
}

func newTicketNumber(now time.Time) (string, error) {
	suffix, err := randomFromAlphabet(6, "23456789ABCDEFGHJKLMNPQRSTUVWXYZ")
	if err != nil {
		return "", fmt.Errorf("generate ticket number: %w", err)
	}
	return now.Format("20060102") + "-" + suffix, nil
}
func randomToken(n int) (string, error) {
	token, err := randomFromAlphabet(n, "abcdefghijklmnopqrstuvwxyz0123456789")
	if err != nil {
		return "", fmt.Errorf("generate ticket attachment key: %w", err)
	}
	return token, nil
}
func randomFromAlphabet(n int, alphabet string) (string, error) {
	b := make([]byte, n)
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return string(b), nil
}
func sanitizeTicketFilename(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, name)
	if name == "" || name == "." {
		return "attachment"
	}
	runes := []rune(name)
	if len(runes) > 200 {
		name = string(runes[:200])
	}
	return name
}
