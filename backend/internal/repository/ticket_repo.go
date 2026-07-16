package repository

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/predicate"
	"github.com/Wei-Shaw/sub2api/ent/ticket"
	"github.com/Wei-Shaw/sub2api/ent/ticketattachment"
	"github.com/Wei-Shaw/sub2api/ent/ticketcategory"
	"github.com/Wei-Shaw/sub2api/ent/ticketmessage"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type ticketRepository struct{ client *dbent.Client }

func NewTicketRepository(client *dbent.Client) service.TicketRepository {
	return &ticketRepository{client: client}
}

func (r *ticketRepository) ListCategories(ctx context.Context, activeOnly bool) ([]service.TicketCategory, error) {
	q := r.client.TicketCategory.Query()
	if activeOnly {
		q = q.Where(ticketcategory.ActiveEQ(true))
	}
	rows, err := q.Order(dbent.Asc(ticketcategory.FieldSortOrder), dbent.Asc(ticketcategory.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]service.TicketCategory, 0, len(rows))
	for _, row := range rows {
		out = append(out, ticketCategoryToService(row))
	}
	return out, nil
}

func (r *ticketRepository) GetCategory(ctx context.Context, id int64) (*service.TicketCategory, error) {
	row, err := r.client.TicketCategory.Get(ctx, id)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrTicketCategoryNotFound, nil)
	}
	out := ticketCategoryToService(row)
	return &out, nil
}

func (r *ticketRepository) CreateCategory(ctx context.Context, c *service.TicketCategory) error {
	row, err := r.client.TicketCategory.Create().SetCode(c.Code).SetNameZh(c.NameZH).SetNameEn(c.NameEN).SetActive(c.Active).SetSortOrder(c.SortOrder).Save(ctx)
	if err != nil {
		return translatePersistenceError(err, nil, service.ErrTicketCategoryConflict)
	}
	*c = ticketCategoryToService(row)
	return nil
}

func (r *ticketRepository) UpdateCategory(ctx context.Context, c *service.TicketCategory) error {
	row, err := r.client.TicketCategory.UpdateOneID(c.ID).SetNameZh(c.NameZH).SetNameEn(c.NameEN).SetActive(c.Active).SetSortOrder(c.SortOrder).Save(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrTicketCategoryNotFound, nil)
	}
	*c = ticketCategoryToService(row)
	return nil
}

func (r *ticketRepository) ReorderCategories(ctx context.Context, ids []int64) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for i, id := range ids {
		if _, err := tx.TicketCategory.UpdateOneID(id).SetSortOrder((i + 1) * 10).Save(ctx); err != nil {
			return translatePersistenceError(err, service.ErrTicketCategoryNotFound, nil)
		}
	}
	return tx.Commit()
}

func (r *ticketRepository) CountActiveByUser(ctx context.Context, userID int64) (int, error) {
	return r.client.Ticket.Query().Where(ticket.UserIDEQ(userID), ticket.StatusNEQ(service.TicketStatusClosed)).Count(ctx)
}

func (r *ticketRepository) Create(ctx context.Context, t *service.Ticket, m *service.TicketMessage, attachments []service.TicketAttachment) error {
	for attempt := 0; attempt < 4; attempt++ {
		tx, err := r.client.Tx(ctx)
		if err != nil {
			return err
		}
		row, err := tx.Ticket.Create().SetNumber(t.Number).SetUserID(t.UserID).SetCategoryID(t.CategoryID).SetSubject(t.Subject).SetStatus(t.Status).SetUserUnreadCount(t.UserUnreadCount).SetAdminUnreadCount(t.AdminUnreadCount).SetLastActorType(t.LastActorType).SetLastMessageAt(t.LastMessageAt).Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			if isUniqueConstraintViolation(err) {
				t.Number = time.Now().Format("20060102") + "-" + repositoryTicketSuffix()
				continue
			}
			return err
		}
		msg, err := tx.TicketMessage.Create().SetTicketID(row.ID).SetNillableSenderUserID(m.SenderUserID).SetSenderType(m.SenderType).SetEventType(m.EventType).SetContent(m.Content).Save(ctx)
		if err == nil {
			err = createTicketAttachments(ctx, tx, msg.ID, attachments)
		}
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if err = tx.Commit(); err != nil {
			return err
		}
		t.ID = row.ID
		m.ID = msg.ID
		m.TicketID = row.ID
		return nil
	}
	return service.ErrTicketCategoryConflict
}

func (r *ticketRepository) List(ctx context.Context, params pagination.PaginationParams, filters service.TicketListFilters) ([]service.Ticket, *pagination.PaginationResult, error) {
	preds := ticketPredicates(filters)
	q := r.client.Ticket.Query().Where(preds...)
	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	rows, err := q.WithCategory().WithUser().Order(dbent.Desc(ticket.FieldLastMessageAt), dbent.Desc(ticket.FieldID)).Offset(params.Offset()).Limit(params.Limit()).All(ctx)
	if err != nil {
		return nil, nil, err
	}
	out := make([]service.Ticket, 0, len(rows))
	for _, row := range rows {
		out = append(out, ticketToService(row, false))
	}
	return out, paginationResultFromTotal(int64(total), params), nil
}

func ticketPredicates(filters service.TicketListFilters) []predicate.Ticket {
	preds := make([]predicate.Ticket, 0, 6)
	if filters.UserID != nil {
		preds = append(preds, ticket.UserIDEQ(*filters.UserID))
	}
	if filters.Status != "" {
		preds = append(preds, ticket.StatusEQ(filters.Status))
	}
	if filters.CategoryID > 0 {
		preds = append(preds, ticket.CategoryIDEQ(filters.CategoryID))
	}
	if filters.UnreadOnly {
		if filters.UserID == nil {
			preds = append(preds, ticket.AdminUnreadCountGT(0))
		} else {
			preds = append(preds, ticket.UserUnreadCountGT(0))
		}
	}
	if search := strings.TrimSpace(filters.Search); search != "" {
		preds = append(preds, ticket.Or(ticket.NumberContainsFold(search), ticket.SubjectContainsFold(search), ticket.HasUserWith(user.Or(user.EmailContainsFold(search), user.UsernameContainsFold(search)))))
	}
	return preds
}

func (r *ticketRepository) Get(ctx context.Context, number string, userID *int64) (*service.Ticket, error) {
	preds := []predicate.Ticket{ticket.NumberEQ(number)}
	if userID != nil {
		preds = append(preds, ticket.UserIDEQ(*userID))
	}
	row, err := r.client.Ticket.Query().Where(preds...).WithCategory().WithUser().WithMessages(func(q *dbent.TicketMessageQuery) {
		q.Order(dbent.Asc(ticketmessage.FieldCreatedAt), dbent.Asc(ticketmessage.FieldID)).WithAttachments(func(aq *dbent.TicketAttachmentQuery) { aq.Order(dbent.Asc(ticketattachment.FieldID)) })
	}).Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrTicketNotFound, nil)
	}
	out := ticketToService(row, true)
	return &out, nil
}

func (r *ticketRepository) AppendMessage(ctx context.Context, number string, userID int64, isAdmin bool, content string, attachments []service.TicketAttachment) (*service.Ticket, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	preds := []predicate.Ticket{ticket.NumberEQ(number)}
	if !isAdmin {
		preds = append(preds, ticket.UserIDEQ(userID))
	}
	row, err := tx.Ticket.Query().Where(preds...).ForUpdate().Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrTicketNotFound, nil)
	}
	if row.Status == service.TicketStatusClosed {
		return nil, service.ErrTicketClosed
	}
	sender := service.TicketSenderUser
	nextStatus := service.TicketStatusOpen
	userUnread, adminUnread := row.UserUnreadCount, row.AdminUnreadCount
	if isAdmin {
		sender = service.TicketSenderAdmin
		nextStatus = service.TicketStatusAnswered
		userUnread++
		adminUnread = 0
	} else {
		adminUnread++
		userUnread = 0
	}
	now := time.Now()
	msg, err := tx.TicketMessage.Create().SetTicketID(row.ID).SetSenderUserID(userID).SetSenderType(sender).SetContent(content).Save(ctx)
	if err == nil {
		err = createTicketAttachments(ctx, tx, msg.ID, attachments)
	}
	if err != nil {
		return nil, err
	}
	_, err = tx.Ticket.UpdateOneID(row.ID).SetStatus(nextStatus).SetLastActorType(sender).SetLastMessageAt(now).SetUserUnreadCount(userUnread).SetAdminUnreadCount(adminUnread).Save(ctx)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	owner := (*int64)(nil)
	if !isAdmin {
		owner = &userID
	}
	return r.Get(ctx, number, owner)
}

func (r *ticketRepository) ChangeStatus(ctx context.Context, number string, userID int64, isAdmin bool, reopen bool) (*service.Ticket, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	preds := []predicate.Ticket{ticket.NumberEQ(number)}
	if !isAdmin {
		preds = append(preds, ticket.UserIDEQ(userID))
	}
	row, err := tx.Ticket.Query().Where(preds...).ForUpdate().Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrTicketNotFound, nil)
	}
	now := time.Now()
	sender := service.TicketSenderUser
	if isAdmin {
		sender = service.TicketSenderAdmin
	}
	event, nextStatus := "ticket_closed", service.TicketStatusClosed
	if reopen {
		event, nextStatus = "ticket_reopened", service.TicketStatusOpen
	}
	if row.Status == nextStatus {
		_ = tx.Rollback()
		owner := (*int64)(nil)
		if !isAdmin {
			owner = &userID
		}
		return r.Get(ctx, number, owner)
	}
	if reopen && row.Status != service.TicketStatusClosed {
		return nil, service.ErrTicketClosed
	}
	if _, err = tx.TicketMessage.Create().SetTicketID(row.ID).SetSenderUserID(userID).SetSenderType(service.TicketSenderSystem).SetEventType(event).Save(ctx); err != nil {
		return nil, err
	}
	update := tx.Ticket.UpdateOneID(row.ID).SetStatus(nextStatus).SetLastActorType(service.TicketSenderSystem).SetLastMessageAt(now)
	if reopen {
		update.ClearClosedAt().ClearClosedByUserID().ClearClosedByRole()
	} else {
		update.SetClosedAt(now).SetClosedByUserID(userID).SetClosedByRole(sender)
	}
	if isAdmin {
		update.SetUserUnreadCount(row.UserUnreadCount + 1)
		update.SetAdminUnreadCount(0)
	} else {
		update.SetAdminUnreadCount(row.AdminUnreadCount + 1)
		update.SetUserUnreadCount(0)
	}
	if _, err = update.Save(ctx); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	owner := (*int64)(nil)
	if !isAdmin {
		owner = &userID
	}
	return r.Get(ctx, number, owner)
}

func (r *ticketRepository) MarkRead(ctx context.Context, number string, userID int64, isAdmin bool) error {
	preds := []predicate.Ticket{ticket.NumberEQ(number)}
	if !isAdmin {
		preds = append(preds, ticket.UserIDEQ(userID))
	}
	update := r.client.Ticket.Update().Where(preds...)
	if isAdmin {
		update.SetAdminUnreadCount(0)
	} else {
		update.SetUserUnreadCount(0)
	}
	count, err := update.Save(ctx)
	if err != nil {
		return err
	}
	if count == 0 {
		return service.ErrTicketNotFound
	}
	return nil
}

func (r *ticketRepository) UnreadCount(ctx context.Context, userID *int64) (int, error) {
	q := r.client.Ticket.Query()
	if userID != nil {
		q = q.Where(ticket.UserIDEQ(*userID))
	}
	rows, err := q.All(ctx)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, row := range rows {
		if userID == nil {
			total += row.AdminUnreadCount
		} else {
			total += row.UserUnreadCount
		}
	}
	return total, nil
}

func (r *ticketRepository) GetAttachment(ctx context.Context, id int64, userID int64, isAdmin bool) (*service.TicketAttachment, error) {
	pred := ticketattachment.IDEQ(id)
	if !isAdmin {
		pred = ticketattachment.And(pred, ticketattachment.HasMessageWith(ticketmessage.HasTicketWith(ticket.UserIDEQ(userID))))
	}
	row, err := r.client.TicketAttachment.Query().Where(pred).Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrTicketNotFound, nil)
	}
	out := ticketAttachmentToService(row)
	return &out, nil
}

func createTicketAttachments(ctx context.Context, tx *dbent.Tx, messageID int64, attachments []service.TicketAttachment) error {
	for i := range attachments {
		a := &attachments[i]
		row, err := tx.TicketAttachment.Create().SetMessageID(messageID).SetObjectKey(a.ObjectKey).SetOriginalName(a.OriginalName).SetContentType(a.ContentType).SetSizeBytes(a.SizeBytes).SetSha256(a.SHA256).Save(ctx)
		if err != nil {
			return err
		}
		a.ID, a.MessageID = row.ID, messageID
	}
	return nil
}

func ticketCategoryToService(row *dbent.TicketCategory) service.TicketCategory {
	return service.TicketCategory{ID: row.ID, Code: row.Code, NameZH: row.NameZh, NameEN: row.NameEn, Active: row.Active, SortOrder: row.SortOrder, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}
func ticketAttachmentToService(row *dbent.TicketAttachment) service.TicketAttachment {
	return service.TicketAttachment{ID: row.ID, MessageID: row.MessageID, ObjectKey: row.ObjectKey, OriginalName: row.OriginalName, ContentType: row.ContentType, SizeBytes: row.SizeBytes, SHA256: row.Sha256, CreatedAt: row.CreatedAt}
}
func ticketToService(row *dbent.Ticket, includeMessages bool) service.Ticket {
	out := service.Ticket{ID: row.ID, Number: row.Number, UserID: row.UserID, CategoryID: row.CategoryID, Subject: row.Subject, Status: row.Status, UserUnreadCount: row.UserUnreadCount, AdminUnreadCount: row.AdminUnreadCount, LastActorType: row.LastActorType, LastMessageAt: row.LastMessageAt, ClosedAt: row.ClosedAt, ClosedByUserID: row.ClosedByUserID, ClosedByRole: row.ClosedByRole, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
	if row.Edges.Category != nil {
		out.Category = ticketCategoryToService(row.Edges.Category)
	}
	if row.Edges.User != nil {
		out.UserEmail, out.UserName = row.Edges.User.Email, row.Edges.User.Username
	}
	if includeMessages {
		out.Messages = make([]service.TicketMessage, 0, len(row.Edges.Messages))
		for _, msg := range row.Edges.Messages {
			item := service.TicketMessage{ID: msg.ID, TicketID: msg.TicketID, SenderUserID: msg.SenderUserID, SenderType: msg.SenderType, EventType: msg.EventType, Content: msg.Content, CreatedAt: msg.CreatedAt, Attachments: []service.TicketAttachment{}}
			for _, a := range msg.Edges.Attachments {
				item.Attachments = append(item.Attachments, ticketAttachmentToService(a))
			}
			out.Messages = append(out.Messages, item)
		}
	}
	return out
}

func repositoryTicketSuffix() string {
	const alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	raw := make([]byte, 6)
	if _, err := rand.Read(raw); err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	for i := range raw {
		raw[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return string(raw)
}
