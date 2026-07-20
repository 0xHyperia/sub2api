package service

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type distributionEvidenceRepoStub struct {
	DistributionRepository
	created []DistributionWithdrawalAttachment
	detail  *DistributionWithdrawalDetail
	err     error
}

func (s *distributionEvidenceRepoStub) AdminCreateWithdrawalAttachments(_ context.Context, _ int64, _ int64, items []DistributionWithdrawalAttachment) error {
	s.created = append([]DistributionWithdrawalAttachment(nil), items...)
	return s.err
}

func (s *distributionEvidenceRepoStub) AdminGetWithdrawal(context.Context, int64) (*DistributionWithdrawalDetail, error) {
	return s.detail, s.err
}

type distributionEvidenceStorageStub struct {
	active      bool
	putKeys     []string
	deletedKeys []string
	url         string
}

func (s *distributionEvidenceStorageStub) Active() bool { return s.active }
func (s *distributionEvidenceStorageStub) Put(_ context.Context, key, _ string, body io.Reader, size int64) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	if int64(len(data)) != size {
		return errors.New("size mismatch")
	}
	s.putKeys = append(s.putKeys, key)
	return nil
}
func (s *distributionEvidenceStorageStub) Delete(_ context.Context, key string) error {
	s.deletedKeys = append(s.deletedKeys, key)
	return nil
}
func (s *distributionEvidenceStorageStub) PresignGet(context.Context, string, string, string) (string, error) {
	return s.url, nil
}

func newDistributionEvidenceService(repo DistributionRepository, storage TicketStorage) *DistributionService {
	return ProvideDistributionService(repo, storage, &config.Config{TicketStorage: config.TicketStorageConfig{
		MaxFileBytes: 1024, MaxTotalBytes: 2048,
	}}, nil, nil)
}

func TestDistributionEvidenceUploadStoresPrivateMetadata(t *testing.T) {
	repo := &distributionEvidenceRepoStub{}
	storage := &distributionEvidenceStorageStub{active: true}
	svc := newDistributionEvidenceService(repo, storage)
	png := []byte{'\x89', 'P', 'N', 'G', '\r', '\n', '\x1a', '\n', 0, 0, 0, 0}

	items, err := svc.AdminUploadWithdrawalEvidence(context.Background(), 42, 7, "payment_receipt", "bank transfer", []TicketUpload{{Name: "receipt.png", Data: png}})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Len(t, storage.putKeys, 1)
	require.Empty(t, storage.deletedKeys)
	require.Len(t, repo.created, 1)
	require.Contains(t, repo.created[0].ObjectKey, "distribution/withdrawals/")
	require.Equal(t, "image/png", repo.created[0].ContentType)
	require.Equal(t, "payment_receipt", repo.created[0].EvidenceType)
	require.Len(t, repo.created[0].SHA256, 64)
}

func TestDistributionEvidenceUploadCleansObjectWhenMetadataFails(t *testing.T) {
	repo := &distributionEvidenceRepoStub{err: errors.New("database unavailable")}
	storage := &distributionEvidenceStorageStub{active: true}
	svc := newDistributionEvidenceService(repo, storage)
	png := []byte{'\x89', 'P', 'N', 'G', '\r', '\n', '\x1a', '\n', 0, 0, 0, 0}

	_, err := svc.AdminUploadWithdrawalEvidence(context.Background(), 42, 7, "payment_receipt", "", []TicketUpload{{Name: "receipt.png", Data: png}})
	require.ErrorContains(t, err, "database unavailable")
	require.Equal(t, storage.putKeys, storage.deletedKeys)
}

func TestDistributionEvidenceURLOnlyAllowsOwnedAttachment(t *testing.T) {
	storage := &distributionEvidenceStorageStub{active: true, url: "https://signed.example/receipt"}
	repo := &distributionEvidenceRepoStub{detail: &DistributionWithdrawalDetail{Attachments: []DistributionWithdrawalAttachment{{ID: 9, WithdrawalID: 42, ObjectKey: "private/key", OriginalName: "receipt.png", ContentType: "image/png"}}}}
	svc := newDistributionEvidenceService(repo, storage)

	url, err := svc.AdminWithdrawalAttachmentURL(context.Background(), 42, 9)
	require.NoError(t, err)
	require.Equal(t, storage.url, url)
	_, err = svc.AdminWithdrawalAttachmentURL(context.Background(), 42, 10)
	require.Error(t, err)
}
