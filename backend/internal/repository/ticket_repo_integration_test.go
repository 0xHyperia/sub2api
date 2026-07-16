//go:build integration

package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/ent/ticketcategory"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestTicketRepositoryOwnershipAndLifecycle(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	owner, err := integrationEntClient.User.Create().SetEmail(fmt.Sprintf("ticket-owner-%d@example.com", suffix)).SetPasswordHash("hash").Save(ctx)
	require.NoError(t, err)
	other, err := integrationEntClient.User.Create().SetEmail(fmt.Sprintf("ticket-other-%d@example.com", suffix)).SetPasswordHash("hash").Save(ctx)
	require.NoError(t, err)
	category, err := integrationEntClient.TicketCategory.Query().Where(ticketcategory.CodeEQ("other")).Only(ctx)
	require.NoError(t, err)

	repo := NewTicketRepository(integrationEntClient)
	number := fmt.Sprintf("20990101-%06d", suffix%1000000)
	ticket := &service.Ticket{Number: number, UserID: owner.ID, CategoryID: category.ID, Subject: "Integration ticket", Status: service.TicketStatusOpen, AdminUnreadCount: 1, LastActorType: service.TicketSenderUser, LastMessageAt: time.Now()}
	message := &service.TicketMessage{SenderUserID: &owner.ID, SenderType: service.TicketSenderUser, Content: "Initial message"}
	require.NoError(t, repo.Create(ctx, ticket, message, nil))

	owned, err := repo.Get(ctx, ticket.Number, &owner.ID)
	require.NoError(t, err)
	require.Len(t, owned.Messages, 1)
	_, err = repo.Get(ctx, ticket.Number, &other.ID)
	require.ErrorIs(t, err, service.ErrTicketNotFound)
	_, err = repo.AppendMessage(ctx, ticket.Number, other.ID, false, "not allowed", nil)
	require.ErrorIs(t, err, service.ErrTicketNotFound)

	answered, err := repo.AppendMessage(ctx, ticket.Number, other.ID, true, "Support reply", nil)
	require.NoError(t, err)
	require.Equal(t, service.TicketStatusAnswered, answered.Status)
	require.Equal(t, 1, answered.UserUnreadCount)
	require.Zero(t, answered.AdminUnreadCount)
	require.NoError(t, repo.MarkRead(ctx, ticket.Number, owner.ID, false))

	closed, err := repo.ChangeStatus(ctx, ticket.Number, owner.ID, false, false)
	require.NoError(t, err)
	require.Equal(t, service.TicketStatusClosed, closed.Status)
	_, err = repo.AppendMessage(ctx, ticket.Number, other.ID, true, "closed reply", nil)
	require.True(t, errors.Is(err, service.ErrTicketClosed))

	reopened, err := repo.ChangeStatus(ctx, ticket.Number, other.ID, true, true)
	require.NoError(t, err)
	require.Equal(t, service.TicketStatusOpen, reopened.Status)
	require.GreaterOrEqual(t, reopened.UserUnreadCount, 1)
}
