package migrations

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDistributionPromotionTrackingBaselineContainsFinalSchema(t *testing.T) {
	content, err := FS.ReadFile("192_distribution_promotion_tracking.sql")
	require.NoError(t, err)
	sql := strings.Join(strings.Fields(string(content)), " ")

	for _, fragment := range []string{
		"promotion_attribution_model IN ('first_touch','last_touch')",
		"detail_expired_at TIMESTAMPTZ",
		"COALESCE(utm_campaign,'')",
		"ON DELETE SET NULL",
		"attribution_device_type VARCHAR(16)",
		"rollup_kind IN ('legacy','visitor')",
		"distribution_promotion_rollups_visitor_hash_check",
		"distribution_promotion_archive_skips",
		"last_privacy_scrub_at",
		"scrub_distribution_promotion_details",
		"cleanup_distribution_promotion_details",
		"ORDER BY visited_at,id LIMIT batch_size",
		"WHERE v.id=ANY(archive_ids)",
		"WHERE id=ANY(archive_ids)",
		"distribution_binding_claims",
		"user_promotion_ownerships",
		"ON CONFLICT(user_id) DO UPDATE SET owner_type=user_promotion_ownerships.owner_type",
		"distribution_customer_ownership_exclusivity",
		"affiliate_inviter_ownership_exclusivity",
		"distribution_promotion_legacy_rollup_status",
	} {
		require.Contains(t, sql, fragment)
	}
	require.NotContains(t, sql, "CREATE INDEX CONCURRENTLY")
	require.NotContains(t, sql, "UPDATE distribution_promotion_conversions c SET attribution_device_type")
	require.NotContains(t, sql, "DROP FUNCTION IF EXISTS cleanup_distribution_promotion_details")
}

func TestDistributionPromotionTrackingUsesSingleUnpublishedBaseline(t *testing.T) {
	for version := 193; version <= 201; version++ {
		matches, err := fs.Glob(FS, fmt.Sprintf("%d_*", version))
		require.NoError(t, err)
		require.Emptyf(t, matches, "development patch migration %d must stay squashed into 192", version)
	}
}
