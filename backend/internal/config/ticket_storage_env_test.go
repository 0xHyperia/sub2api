//go:build unit

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadTicketStorageCredentialsFromEnv(t *testing.T) {
	resetViperWithJWTSecret(t)
	t.Setenv("TICKET_STORAGE_ENABLED", "true")
	t.Setenv("TICKET_STORAGE_ENDPOINT", "https://acct.r2.cloudflarestorage.com")
	t.Setenv("TICKET_STORAGE_BUCKET", "support-attachments")
	t.Setenv("TICKET_STORAGE_ACCESS_KEY_ID", "ticket-ak")
	t.Setenv("TICKET_STORAGE_SECRET_ACCESS_KEY", "ticket-sk")

	cfg, err := Load()
	require.NoError(t, err)

	require.True(t, cfg.TicketStorage.Enabled)
	require.Equal(t, "https://acct.r2.cloudflarestorage.com", cfg.TicketStorage.Endpoint)
	require.Equal(t, "support-attachments", cfg.TicketStorage.Bucket)
	require.Equal(t, "ticket-ak", cfg.TicketStorage.AccessKeyID)
	require.Equal(t, "ticket-sk", cfg.TicketStorage.SecretAccessKey)
	require.True(t, cfg.TicketStorage.IsConfigured())
	require.True(t, cfg.TicketStorage.Active())
}
