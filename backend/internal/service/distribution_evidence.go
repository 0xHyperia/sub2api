package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	distributionEvidenceMaxFiles = 5
	distributionEvidenceMaxBytes = int64(10 * 1024 * 1024)
)

type DistributionEvidenceCapabilities struct {
	Available         bool     `json:"available"`
	MaxFileBytes      int64    `json:"max_file_bytes"`
	MaxFilesPerUpload int      `json:"max_files_per_upload"`
	MaxTotalBytes     int64    `json:"max_total_bytes"`
	AllowedExtensions []string `json:"allowed_extensions"`
}

func (s *DistributionService) EvidenceCapabilities() DistributionEvidenceCapabilities {
	maxFile := s.evidenceConfig.MaxFileBytes
	if maxFile <= 0 || maxFile > distributionEvidenceMaxBytes {
		maxFile = distributionEvidenceMaxBytes
	}
	maxTotal := s.evidenceConfig.MaxTotalBytes
	if maxTotal <= 0 || maxTotal > distributionEvidenceMaxFiles*distributionEvidenceMaxBytes {
		maxTotal = distributionEvidenceMaxFiles * distributionEvidenceMaxBytes
	}
	return DistributionEvidenceCapabilities{
		Available:    s.evidenceStorage != nil && s.evidenceStorage.Active(),
		MaxFileBytes: maxFile, MaxFilesPerUpload: distributionEvidenceMaxFiles, MaxTotalBytes: maxTotal,
		AllowedExtensions: []string{".png", ".jpg", ".jpeg", ".webp", ".pdf"},
	}
}

func (s *DistributionService) AdminUploadWithdrawalEvidence(ctx context.Context, withdrawalID, adminID int64, evidenceType, note string, uploads []TicketUpload) ([]DistributionWithdrawalAttachment, error) {
	capabilities := s.EvidenceCapabilities()
	if !capabilities.Available {
		return nil, infraerrors.BadRequest("DISTRIBUTION_EVIDENCE_UNAVAILABLE", "payment evidence storage is unavailable")
	}
	if withdrawalID <= 0 || adminID <= 0 || len(uploads) == 0 || len(uploads) > capabilities.MaxFilesPerUpload {
		return nil, infraerrors.BadRequest("DISTRIBUTION_EVIDENCE_INVALID", "invalid payment evidence upload")
	}
	if evidenceType != "payment_receipt" && evidenceType != "bank_statement" && evidenceType != "other" {
		evidenceType = "payment_receipt"
	}
	note = strings.TrimSpace(note)
	var total int64
	items := make([]DistributionWithdrawalAttachment, 0, len(uploads))
	keys := make([]string, 0, len(uploads))
	cleanup := func() {
		for _, key := range keys {
			_ = s.evidenceStorage.Delete(ctx, key)
		}
	}
	for _, upload := range uploads {
		size := int64(len(upload.Data))
		total += size
		if size <= 0 || size > capabilities.MaxFileBytes || total > capabilities.MaxTotalBytes {
			cleanup()
			return nil, infraerrors.BadRequest("DISTRIBUTION_EVIDENCE_LIMIT", "payment evidence upload limit exceeded")
		}
		contentType, ok := allowedTicketContentType(upload.Name, upload.Data)
		if !ok || (!strings.HasPrefix(contentType, "image/") && contentType != "application/pdf") {
			cleanup()
			return nil, infraerrors.BadRequest("DISTRIBUTION_EVIDENCE_TYPE", "payment evidence must be an image or PDF")
		}
		token, err := randomToken(24)
		if err != nil {
			cleanup()
			return nil, err
		}
		key := "distribution/withdrawals/" + time.Now().Format("2006/01") + "/" + token
		if err = s.evidenceStorage.Put(ctx, key, contentType, bytes.NewReader(upload.Data), size); err != nil {
			cleanup()
			return nil, fmt.Errorf("store withdrawal evidence: %w", err)
		}
		sum := sha256.Sum256(upload.Data)
		items = append(items, DistributionWithdrawalAttachment{WithdrawalID: withdrawalID, ObjectKey: key, OriginalName: sanitizeTicketFilename(upload.Name), ContentType: contentType, SizeBytes: size, SHA256: hex.EncodeToString(sum[:]), EvidenceType: evidenceType, Note: note})
		keys = append(keys, key)
	}
	if err := s.repo.AdminCreateWithdrawalAttachments(ctx, withdrawalID, adminID, items); err != nil {
		cleanup()
		return nil, err
	}
	return items, nil
}

func (s *DistributionService) AdminWithdrawalAttachmentURL(ctx context.Context, withdrawalID, attachmentID int64) (string, error) {
	if s.evidenceStorage == nil || !s.evidenceStorage.Active() {
		return "", infraerrors.BadRequest("DISTRIBUTION_EVIDENCE_UNAVAILABLE", "payment evidence storage is unavailable")
	}
	detail, err := s.repo.AdminGetWithdrawal(ctx, withdrawalID)
	if err != nil {
		return "", err
	}
	for _, attachment := range detail.Attachments {
		if attachment.ID == attachmentID {
			return s.evidenceStorage.PresignGet(ctx, attachment.ObjectKey, attachment.OriginalName, attachment.ContentType)
		}
	}
	return "", infraerrors.NotFound("DISTRIBUTION_EVIDENCE_NOT_FOUND", "payment evidence not found")
}
