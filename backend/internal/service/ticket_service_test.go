package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type ticketRepoStub struct {
	TicketRepository
	category    *TicketCategory
	activeCount int
	createErr   error
	created     *Ticket
}

func (s *ticketRepoStub) GetCategory(context.Context, int64) (*TicketCategory, error) {
	return s.category, nil
}
func (s *ticketRepoStub) CountActiveByUser(context.Context, int64) (int, error) {
	return s.activeCount, nil
}
func (s *ticketRepoStub) Create(_ context.Context, ticket *Ticket, _ *TicketMessage, _ []TicketAttachment) error {
	s.created = ticket
	return s.createErr
}
func (s *ticketRepoStub) Get(_ context.Context, _ string, _ *int64) (*Ticket, error) {
	return s.created, nil
}

type ticketStorageStub struct {
	active               bool
	putKeys, deletedKeys []string
}

func (s *ticketStorageStub) Active() bool { return s.active }
func (s *ticketStorageStub) Put(_ context.Context, key, _ string, _ io.Reader, _ int64) error {
	s.putKeys = append(s.putKeys, key)
	return nil
}
func (s *ticketStorageStub) Delete(_ context.Context, key string) error {
	s.deletedKeys = append(s.deletedKeys, key)
	return nil
}
func (s *ticketStorageStub) PresignGet(context.Context, string, string, string) (string, error) {
	return "", nil
}

func newTicketServiceForTest(repo TicketRepository, storage TicketStorage) *TicketService {
	cfg := &config.Config{TicketStorage: config.TicketStorageConfig{Enabled: true, Bucket: "tickets", AccessKeyID: "key", SecretAccessKey: "secret", Prefix: "tickets/", MaxFileBytes: 1024, MaxFilesPerMessage: 5, MaxTotalBytes: 2048}}
	return NewTicketService(repo, storage, cfg)
}

func TestTicketSubjectUsesFirstLineAndRuneLimit(t *testing.T) {
	require.Equal(t, "first line", ticketSubject("first line\nsecond line"))
	require.Equal(t, 120, len([]rune(ticketSubject(strings.Repeat("工", 140)))))
}

func TestAllowedTicketContentTypeUsesContentSniffing(t *testing.T) {
	contentType, ok := allowedTicketContentType("client.log", []byte("request failed\n"))
	require.True(t, ok)
	require.Equal(t, "text/plain; charset=utf-8", contentType)
	_, ok = allowedTicketContentType("payload.svg", []byte("<svg></svg>"))
	require.False(t, ok)
	_, ok = allowedTicketContentType("fake.pdf", []byte("plain text"))
	require.False(t, ok)
}

func TestCreateTicketRejectsInactiveCategoryAndOpenLimit(t *testing.T) {
	repo := &ticketRepoStub{category: &TicketCategory{ID: 1, Active: false}}
	svc := newTicketServiceForTest(repo, &ticketStorageStub{active: true})
	_, err := svc.Create(context.Background(), 7, 1, "help", nil)
	require.ErrorIs(t, err, ErrTicketCategoryInactive)

	repo.category.Active = true
	repo.activeCount = 10
	_, err = svc.Create(context.Background(), 7, 1, "help", nil)
	require.ErrorIs(t, err, ErrTicketOpenLimit)
}

func TestCreateTicketCleansUploadedObjectsWhenDatabaseWriteFails(t *testing.T) {
	repo := &ticketRepoStub{category: &TicketCategory{ID: 1, Active: true}, createErr: errors.New("database unavailable")}
	storage := &ticketStorageStub{active: true}
	svc := newTicketServiceForTest(repo, storage)

	_, err := svc.Create(context.Background(), 7, 1, "help", []TicketUpload{{Name: "error.log", Data: []byte("request failed\n")}})
	require.ErrorContains(t, err, "database unavailable")
	require.Len(t, storage.putKeys, 1)
	require.Equal(t, storage.putKeys, storage.deletedKeys)
}

func TestCreateTicketKeepsTextAvailableWithoutStorage(t *testing.T) {
	repo := &ticketRepoStub{category: &TicketCategory{ID: 1, Active: true}}
	svc := newTicketServiceForTest(repo, &ticketStorageStub{active: false})

	created, err := svc.Create(context.Background(), 7, 1, "plain text request", nil)
	require.NoError(t, err)
	require.Equal(t, TicketStatusOpen, created.Status)
	require.Regexp(t, `^\d{8}-[23456789ABCDEFGHJKLMNPQRSTUVWXYZ]{6}$`, created.Number)
	require.False(t, svc.Capabilities().AttachmentsAvailable)
}
