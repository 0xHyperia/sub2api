package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestLoadDistributionTrackingHashSecretsFromEnv(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("JWT_SECRET", strings.Repeat("j", 32))
	primary := strings.Repeat("p", 32)
	previous := strings.Repeat("o", 32)
	t.Setenv("DISTRIBUTION_TRACKING_HASH_SECRETS", " "+primary+" , "+previous+","+primary+" ")
	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, []string{primary, previous}, cfg.DistributionTracking.HashSecrets)
}

func TestLoadRejectsWeakDistributionTrackingHashSecret(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("JWT_SECRET", strings.Repeat("j", 32))
	t.Setenv("DISTRIBUTION_TRACKING_HASH_SECRETS", "too-short")

	_, err := Load()
	require.EqualError(t, err, "validate config error: distribution_tracking.hash_secrets entries must be at least 32 bytes")
}

func TestLoadDistributionTrackingCleanupFromEnv(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("JWT_SECRET", strings.Repeat("j", 32))
	t.Setenv("DISTRIBUTION_TRACKING_CLEANUP_ENABLED", "true")
	t.Setenv("DISTRIBUTION_TRACKING_CLEANUP_INTERVAL_MINUTES", "17")
	t.Setenv("DISTRIBUTION_TRACKING_CLEANUP_BATCH_SIZE", "321")

	cfg, err := Load()
	require.NoError(t, err)
	require.True(t, cfg.DistributionTracking.CleanupEnabled)
	require.Equal(t, 17, cfg.DistributionTracking.CleanupIntervalMinutes)
	require.Equal(t, 321, cfg.DistributionTracking.CleanupBatchSize)
}

func TestLoadDistributionTrackingCleanupDefaultsDisabled(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("JWT_SECRET", strings.Repeat("j", 32))

	cfg, err := Load()
	require.NoError(t, err)
	require.False(t, cfg.DistributionTracking.CleanupEnabled)
	require.Equal(t, 10, cfg.DistributionTracking.CleanupIntervalMinutes)
	require.Equal(t, 5000, cfg.DistributionTracking.CleanupBatchSize)
}
