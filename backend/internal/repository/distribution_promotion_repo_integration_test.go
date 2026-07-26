//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	dbmigrations "github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestPromotionBaselineMigrationReplaysTransactionally(t *testing.T) {
	ctx := context.Background()
	tx, err := integrationDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	var before int
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_promotion_visits`).Scan(&before))
	content, err := dbmigrations.FS.ReadFile("192_distribution_promotion_tracking.sql")
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, string(content))
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, string(content))
	require.NoError(t, err)
	var after int
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_promotion_visits`).Scan(&after))
	require.Equal(t, before, after)
}

func TestPromotionBaselineSchemaIsFinal(t *testing.T) {
	ctx := context.Background()
	var patchRows int
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations
		WHERE filename ~ '^(19[3-9]|20[0-1])_'
		  AND (filename ILIKE '%distribution%' OR filename ILIKE '%promotion%')`).Scan(&patchRows))
	require.Zero(t, patchRows)

	var dedupeDefinition string
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT pg_get_indexdef('idx_distribution_promotion_visits_dedupe'::regclass)`).Scan(&dedupeDefinition))
	for _, fragment := range []string{"landing_path", "referrer_host", "utm_source", "utm_medium", "utm_campaign", "device_type"} {
		require.Contains(t, dedupeDefinition, fragment)
	}

	var finalColumns, finalTriggers, finalFunctions, finalView int
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.columns
		WHERE (table_name='distribution_promotion_visits' AND column_name IN ('dedupe_bucket','detail_expired_at'))
		   OR (table_name='distribution_promotion_conversions' AND column_name IN ('attribution_visit_id','attribution_device_type','attribution_is_bot'))
		   OR (table_name='distribution_promotion_maintenance' AND column_name IN ('last_privacy_scrub_at','last_archive_skipped_at'))`).Scan(&finalColumns))
	require.Equal(t, 7, finalColumns)
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM pg_trigger
		WHERE NOT tgisinternal AND tgname IN ('distribution_customer_ownership_exclusivity','affiliate_inviter_ownership_exclusivity')`).Scan(&finalTriggers))
	require.Equal(t, 2, finalTriggers)
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM pg_proc
		WHERE proname IN ('scrub_distribution_promotion_details','cleanup_distribution_promotion_details','enforce_user_promotion_ownership_exclusivity')`).Scan(&finalFunctions))
	require.Equal(t, 3, finalFunctions)
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.views
		WHERE table_name='distribution_promotion_legacy_rollup_status'`).Scan(&finalView))
	require.Equal(t, 1, finalView)
}

func TestPromotionAttributionIsAgentScopedAndDirectSourceIsRetained(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	createUser := func(role string) int64 {
		u, err := integrationEntClient.User.Create().
			SetEmail(fmt.Sprintf("promotion-%s-%d@example.com", role, suffix)).
			SetPasswordHash("hash").Save(ctx)
		require.NoError(t, err)
		return u.ID
	}
	adminID := createUser("admin")
	agentAUserID := createUser("agent-a")
	agentBUserID := createUser("agent-b")
	persistedUserID := createUser("persisted")
	directUserID := createUser("direct")
	untrackedUserID := createUser("untracked-direct")

	repo := NewDistributionRepository(integrationDB)
	svc := service.NewDistributionService(repo, "promotion-integration-secret-0123456789")
	original, err := svc.AdminGetSettings(ctx)
	require.NoError(t, err)
	settings := *original
	settings.Enabled = true
	settings.PromotionTrackingEnabled = true
	settings.PromotionAttributionEnabled = true
	settings.PromotionAttributionModel = "last_touch"
	settings.PromotionAttributionDays = 30
	settings.PromotionDetailRetentionDays = 180
	require.NoError(t, svc.AdminUpdateSettings(ctx, settings, adminID))

	codeA := fmt.Sprintf("PRA%d", suffix%1_000_000_000)
	codeB := fmt.Sprintf("PRB%d", suffix%1_000_000_000)
	agentA, err := svc.AdminGrantAgent(ctx, service.DistributionGrantAgentInput{UserID: agentAUserID, Depth: 1, PromotionCode: codeA, GrantedBy: adminID})
	require.NoError(t, err)
	agentB, err := svc.AdminGrantAgent(ctx, service.DistributionGrantAgentInput{UserID: agentBUserID, Depth: 1, PromotionCode: codeB, GrantedBy: adminID})
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_promotion_conversions WHERE user_id IN ($1,$2,$3)`, persistedUserID, directUserID, untrackedUserID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_customer_bindings WHERE user_id IN ($1,$2,$3)`, persistedUserID, directUserID, untrackedUserID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_promotion_attributions WHERE agent_id IN ($1,$2)`, agentA.ID, agentB.ID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_promotion_visits WHERE agent_id IN ($1,$2)`, agentA.ID, agentB.ID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_wallets WHERE agent_id IN ($1,$2)`, agentA.ID, agentB.ID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_agent_events WHERE agent_id IN ($1,$2)`, agentA.ID, agentB.ID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_agents WHERE id IN ($1,$2)`, agentA.ID, agentB.ID)
		_ = svc.AdminUpdateSettings(cleanupCtx, *original, adminID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `UPDATE distribution_settings SET updated_by=NULL WHERE updated_by=$1`, adminID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM users WHERE id IN ($1,$2,$3,$4,$5,$6)`, adminID, agentAUserID, agentBUserID, persistedUserID, directUserID, untrackedUserID)
	})

	sharedToken := "shared-cross-agent-visitor"
	_, err = svc.TrackPromotionVisit(ctx, service.DistributionPromotionVisitInput{PromotionCode: codeA, VisitorToken: sharedToken, LandingPath: "/register", UTMSource: "agent-a"})
	require.NoError(t, err)
	trackedB, err := svc.TrackPromotionVisit(ctx, service.DistributionPromotionVisitInput{PromotionCode: codeB, VisitorToken: sharedToken, LandingPath: "/register", UTMSource: "agent-b-first"})
	require.NoError(t, err)
	// Move the first touch into an earlier minute so the next request is a real
	// second visit rather than a refresh in the deduplication bucket.
	var firstBVisitID int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT id FROM distribution_promotion_visits
		WHERE agent_id=$1 AND promotion_code=$2 ORDER BY id DESC LIMIT 1`, agentB.ID, codeB).Scan(&firstBVisitID))
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_promotion_visits
		SET visited_at=NOW()-INTERVAL '2 minutes',dedupe_bucket=date_trunc('minute',NOW()-INTERVAL '2 minutes')
		WHERE id=$1`, firstBVisitID)
	require.NoError(t, err)
	trackedB, err = svc.TrackPromotionVisit(ctx, service.DistributionPromotionVisitInput{PromotionCode: codeB, VisitorToken: sharedToken, LandingPath: "/register", UTMSource: "agent-b-last"})
	require.NoError(t, err)
	duplicateB, err := svc.TrackPromotionVisit(ctx, service.DistributionPromotionVisitInput{PromotionCode: codeB, VisitorToken: sharedToken, LandingPath: "/register", UTMSource: "agent-b-last"})
	require.NoError(t, err)
	require.True(t, duplicateB.Deduplicated)
	require.Equal(t, trackedB.VisitorToken, duplicateB.VisitorToken)
	differentTouchB, err := svc.TrackPromotionVisit(ctx, service.DistributionPromotionVisitInput{PromotionCode: codeB, VisitorToken: sharedToken, LandingPath: "/pricing", UTMSource: "agent-b-alt"})
	require.NoError(t, err)
	require.False(t, differentTouchB.Deduplicated, "different source or landing touches in one minute must not collapse as refreshes")

	registrationCtx := service.WithDistributionVisitorToken(ctx, sharedToken)
	registrationCtx, resolvedCode, err := svc.PrepareRegistrationAttribution(registrationCtx, "", "")
	require.NoError(t, err)
	require.Equal(t, codeB, resolvedCode)
	require.NoError(t, svc.BindCustomerByCode(registrationCtx, persistedUserID, resolvedCode))

	from, to := time.Now().Add(-time.Hour), time.Now().Add(time.Hour)
	filter := service.DistributionPromotionStatsFilter{From: from, To: to, Page: 1, PageSize: 20}
	visitsA, _, err := svc.ListPromotionVisits(ctx, agentAUserID, filter, false)
	require.NoError(t, err)
	require.Len(t, visitsA, 1)
	require.Nil(t, visitsA[0].RegisteredAt, "agent A must not see agent B's conversion")
	visitsB, _, err := svc.ListPromotionVisits(ctx, agentBUserID, filter, false)
	require.NoError(t, err)
	require.Len(t, visitsB, 3, "cross-minute and distinct same-minute touches remain while exact refreshes collapse")
	for _, visit := range visitsB {
		require.NotNil(t, visit.RegisteredAt, "every pre-registration visit by the converted visitor must show as registered")
		require.Equal(t, "persisted", visit.AttributionType)
	}
	var firstAgentID, lastAgentID int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT f.agent_id,l.agent_id FROM distribution_promotion_conversions c
		JOIN distribution_promotion_visits f ON f.id=c.first_visit_id JOIN distribution_promotion_visits l ON l.id=c.last_visit_id
		WHERE c.user_id=$1`, persistedUserID).Scan(&firstAgentID, &lastAgentID))
	require.Equal(t, agentB.ID, firstAgentID)
	require.Equal(t, agentB.ID, lastAgentID)

	analyticsA, err := svc.GetPromotionAnalytics(ctx, agentAUserID, filter, false)
	require.NoError(t, err)
	require.Zero(t, analyticsA.Summary.Registrations)
	analyticsB, err := svc.GetPromotionAnalytics(ctx, agentBUserID, filter, false)
	require.NoError(t, err)
	require.EqualValues(t, 1, analyticsB.Summary.Registrations)
	require.Equal(t, float64(100), analyticsB.Summary.ConversionRate)
	for i := 1; i < len(analyticsB.Daily); i++ {
		require.Greater(t, analyticsB.Daily[i-1].Date, analyticsB.Daily[i].Date, "daily analytics must list the newest date first")
	}
	missingAgentFilter := filter
	missingAgentFilter.AgentID = agentB.ID + 9_000_000_000
	_, err = svc.GetPromotionAnalytics(ctx, adminID, missingAgentFilter, true)
	require.Error(t, err)
	wrongSourceFilter := filter
	wrongSourceFilter.Source = "agent-a"
	filteredB, err := svc.GetPromotionAnalytics(ctx, agentBUserID, wrongSourceFilter, false)
	require.NoError(t, err)
	require.Zero(t, filteredB.Summary.TotalVisits)
	require.Zero(t, filteredB.Summary.Registrations)
	persistedFilter := filter
	persistedFilter.AttributionType = "persisted"
	filteredB, err = svc.GetPromotionAnalytics(ctx, agentBUserID, persistedFilter, false)
	require.NoError(t, err)
	require.EqualValues(t, 3, filteredB.Summary.TotalVisits)
	require.EqualValues(t, 1, filteredB.Summary.Registrations)
	firstSourceFilter := persistedFilter
	firstSourceFilter.Source = "agent-b-first"
	firstSourceAnalytics, err := svc.GetPromotionAnalytics(ctx, agentBUserID, firstSourceFilter, false)
	require.NoError(t, err)
	require.EqualValues(t, 1, firstSourceAnalytics.Summary.TotalVisits)
	require.Zero(t, firstSourceAnalytics.Summary.ConvertedVisitors)
	require.Zero(t, firstSourceAnalytics.Summary.Registrations)
	lastSourceFilter := persistedFilter
	lastSourceFilter.Source = "agent-b-last"
	lastSourceAnalytics, err := svc.GetPromotionAnalytics(ctx, agentBUserID, lastSourceFilter, false)
	require.NoError(t, err)
	require.EqualValues(t, 1, lastSourceAnalytics.Summary.TotalVisits)
	require.Zero(t, lastSourceAnalytics.Summary.ConvertedVisitors)
	require.Zero(t, lastSourceAnalytics.Summary.Registrations)
	altSourceFilter := persistedFilter
	altSourceFilter.Source = "agent-b-alt"
	altSourceAnalytics, err := svc.GetPromotionAnalytics(ctx, agentBUserID, altSourceFilter, false)
	require.NoError(t, err)
	require.EqualValues(t, 1, altSourceAnalytics.Summary.TotalVisits)
	require.EqualValues(t, 1, altSourceAnalytics.Summary.ConvertedVisitors)
	require.EqualValues(t, 1, altSourceAnalytics.Summary.Registrations)
	directFilter := filter
	directFilter.AttributionType = "direct"
	filteredB, err = svc.GetPromotionAnalytics(ctx, agentBUserID, directFilter, false)
	require.NoError(t, err)
	require.Zero(t, filteredB.Summary.TotalVisits)
	require.Zero(t, filteredB.Summary.Registrations)
	unregisteredFilter := filter
	unregisteredFilter.AttributionType = "unregistered"
	unregisteredB, err := svc.GetPromotionAnalytics(ctx, agentBUserID, unregisteredFilter, false)
	require.NoError(t, err)
	require.Zero(t, unregisteredB.Summary.TotalVisits, "converted visitors must not leak non-anchor touches into the unregistered cohort")
	unregisteredBVisits, unregisteredBTotal, err := svc.ListPromotionVisits(ctx, agentBUserID, unregisteredFilter, false)
	require.NoError(t, err)
	require.Zero(t, unregisteredBTotal)
	require.Empty(t, unregisteredBVisits)
	filteredA, err := svc.GetPromotionAnalytics(ctx, agentAUserID, unregisteredFilter, false)
	require.NoError(t, err)
	require.EqualValues(t, 1, filteredA.Summary.TotalVisits)
	require.Zero(t, filteredA.Summary.Registrations)
	unregisteredVisits, unregisteredTotal, err := svc.ListPromotionVisits(ctx, agentAUserID, unregisteredFilter, false)
	require.NoError(t, err)
	require.EqualValues(t, 1, unregisteredTotal)
	require.Len(t, unregisteredVisits, 1)
	require.Nil(t, unregisteredVisits[0].RegisteredAt)

	directToken := "direct-source-visitor"
	_, err = svc.TrackPromotionVisit(ctx, service.DistributionPromotionVisitInput{PromotionCode: codeA, VisitorToken: directToken, LandingPath: "/register", UTMSource: "newsletter"})
	require.NoError(t, err)
	directCtx := service.WithDistributionVisitorToken(ctx, directToken)
	directCtx, resolvedCode, err = svc.PrepareRegistrationAttribution(directCtx, codeA, "")
	require.NoError(t, err)
	require.Equal(t, codeA, resolvedCode)
	require.NoError(t, svc.BindCustomerByCode(directCtx, directUserID, resolvedCode))

	var source, attributionType, model string
	var attributedVisitID int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT attribution_source,attribution_type,attribution_model,attribution_visit_id
		FROM distribution_promotion_conversions WHERE user_id=$1`, directUserID).Scan(&source, &attributionType, &model, &attributedVisitID))
	require.Equal(t, "newsletter", source)
	require.Equal(t, "direct", attributionType)
	require.Equal(t, "last_touch", model)
	require.Positive(t, attributedVisitID)
	require.NoError(t, svc.BindCustomerByCode(ctx, untrackedUserID, codeA))
	analyticsA, err = svc.GetPromotionAnalytics(ctx, agentAUserID, filter, false)
	require.NoError(t, err)
	require.Equal(t, "visit", analyticsA.Meta.Cohort)
	require.EqualValues(t, 1, analyticsA.Summary.Registrations)
	require.EqualValues(t, 1, analyticsA.Summary.DirectRegistrations)
	require.Zero(t, analyticsA.Summary.UntrackedDirect)
	require.EqualValues(t, 1, analyticsA.Summary.ConvertedVisitors)
	require.EqualValues(t, 1, analyticsA.Summary.TrackedRegistrations)
	for i := range analyticsA.Sources {
		require.NotEqual(t, "直接注册", analyticsA.Sources[i].Source, "visit cohort excludes registrations without a tracked visit")
	}
	caseInsensitiveFilter := filter
	caseInsensitiveFilter.Source = "NEWSLETTER"
	caseInsensitiveAnalytics, err := svc.GetPromotionAnalytics(ctx, agentAUserID, caseInsensitiveFilter, false)
	require.NoError(t, err)
	require.EqualValues(t, 1, caseInsensitiveAnalytics.Summary.TotalVisits)
	caseInsensitiveVisits, caseInsensitiveTotal, err := svc.ListPromotionVisits(ctx, agentAUserID, caseInsensitiveFilter, false)
	require.NoError(t, err)
	require.EqualValues(t, 1, caseInsensitiveTotal)
	require.Len(t, caseInsensitiveVisits, 1)
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_promotion_conversions SET registered_at=$2 WHERE user_id=$1`, directUserID, to.Add(time.Hour))
	require.NoError(t, err)
	visitCohortAnalytics, err := svc.GetPromotionAnalytics(ctx, agentAUserID, filter, false)
	require.NoError(t, err)
	require.EqualValues(t, 1, visitCohortAnalytics.Summary.Registrations,
		"a registration outside the selected dates remains attributed to its in-range visit cohort")
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_agents SET status='revoked' WHERE id=$1`, agentB.ID)
	require.NoError(t, err)
	revokedAgentFilter := filter
	revokedAgentFilter.AgentID = agentB.ID
	revokedAnalytics, err := svc.GetPromotionAnalytics(ctx, adminID, revokedAgentFilter, true)
	require.NoError(t, err)
	require.EqualValues(t, 3, revokedAnalytics.Summary.TotalVisits, "administrators must retain access to revoked-agent history")
	activeOnlyOptions, err := svc.AdminLookupAgents(ctx, codeB)
	require.NoError(t, err)
	require.Empty(t, activeOnlyOptions)
	historicalOptions, err := svc.AdminLookupAgents(ctx, codeB, true)
	require.NoError(t, err)
	require.Len(t, historicalOptions, 1)
	require.Equal(t, "revoked", historicalOptions[0].Status)
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_agents SET status='active' WHERE id=$1`, agentB.ID)
	require.NoError(t, err)
}

func TestPromotionDetailCleanupPreservesActiveAttributionAndConversionSnapshot(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	createUser := func(role string) int64 {
		u, err := integrationEntClient.User.Create().
			SetEmail(fmt.Sprintf("promotion-cleanup-%s-%d@example.com", role, suffix)).
			SetPasswordHash("hash").Save(ctx)
		require.NoError(t, err)
		return u.ID
	}
	adminID := createUser("admin")
	agentUserID := createUser("agent")
	customerUserID := createUser("customer")
	archivedCustomerUserID := createUser("archived-customer")

	repo := NewDistributionRepository(integrationDB)
	svc := service.NewDistributionService(repo, "promotion-integration-secret-0123456789")
	original, err := svc.AdminGetSettings(ctx)
	require.NoError(t, err)
	settings := *original
	settings.Enabled = true
	settings.PromotionTrackingEnabled = true
	settings.PromotionAttributionEnabled = true
	settings.PromotionAttributionModel = "first_touch"
	settings.PromotionAttributionDays = 30
	settings.PromotionDetailRetentionDays = 30
	require.NoError(t, svc.AdminUpdateSettings(ctx, settings, adminID))

	code := fmt.Sprintf("PRC%d", suffix%1_000_000_000)
	agent, err := svc.AdminGrantAgent(ctx, service.DistributionGrantAgentInput{UserID: agentUserID, Depth: 1, PromotionCode: code, GrantedBy: adminID})
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_promotion_conversions WHERE user_id IN ($1,$2)`, customerUserID, archivedCustomerUserID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_customer_bindings WHERE user_id IN ($1,$2)`, customerUserID, archivedCustomerUserID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_promotion_daily_rollups WHERE agent_id=$1`, agent.ID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_promotion_attributions WHERE agent_id=$1`, agent.ID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_promotion_visits WHERE agent_id=$1`, agent.ID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_wallets WHERE agent_id=$1`, agent.ID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_agent_events WHERE agent_id=$1`, agent.ID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_agents WHERE id=$1`, agent.ID)
		_ = svc.AdminUpdateSettings(cleanupCtx, *original, adminID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `UPDATE distribution_settings SET updated_by=NULL WHERE updated_by=$1`, adminID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM users WHERE id IN ($1,$2,$3,$4)`, adminID, agentUserID, customerUserID, archivedCustomerUserID)
	})

	convertedToken := "cleanup-converted-visitor"
	_, err = svc.TrackPromotionVisit(ctx, service.DistributionPromotionVisitInput{PromotionCode: code, VisitorToken: convertedToken, LandingPath: "/register", UTMSource: "retained-source"})
	require.NoError(t, err)
	registrationCtx := service.WithDistributionVisitorToken(ctx, convertedToken)
	registrationCtx, resolvedCode, err := svc.PrepareRegistrationAttribution(registrationCtx, "", "")
	require.NoError(t, err)
	require.NoError(t, svc.BindCustomerByCode(registrationCtx, customerUserID, resolvedCode))

	activeToken := "cleanup-active-visitor"
	_, err = svc.TrackPromotionVisit(ctx, service.DistributionPromotionVisitInput{PromotionCode: code, VisitorToken: activeToken, LandingPath: "/register", UTMSource: "active-source"})
	require.NoError(t, err)
	archivedToken := "cleanup-archived-visitor"
	_, err = svc.TrackPromotionVisit(ctx, service.DistributionPromotionVisitInput{PromotionCode: code, VisitorToken: archivedToken, LandingPath: "/register", UTMSource: "archived-source"})
	require.NoError(t, err)
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_promotion_visits
		SET visited_at=NOW()-INTERVAL '2 minutes',dedupe_bucket=date_trunc('minute',NOW()-INTERVAL '2 minutes')
		WHERE agent_id=$1 AND utm_source='archived-source'`, agent.ID)
	require.NoError(t, err)
	_, err = svc.TrackPromotionVisit(ctx, service.DistributionPromotionVisitInput{PromotionCode: code, VisitorToken: archivedToken, LandingPath: "/pricing", UTMSource: "archived-source"})
	require.NoError(t, err)
	archivedCtx := service.WithDistributionVisitorToken(ctx, archivedToken)
	archivedCtx, resolvedCode, err = svc.PrepareRegistrationAttribution(archivedCtx, "", "")
	require.NoError(t, err)
	require.NoError(t, svc.BindCustomerByCode(archivedCtx, archivedCustomerUserID, resolvedCode))

	var convertedHash, activeHash, archivedHash string
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT visitor_token_hash FROM distribution_promotion_visits WHERE agent_id=$1 AND utm_source='retained-source'`, agent.ID).Scan(&convertedHash))
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT visitor_token_hash FROM distribution_promotion_visits WHERE agent_id=$1 AND utm_source='active-source'`, agent.ID).Scan(&activeHash))
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT visitor_token_hash FROM distribution_promotion_visits WHERE agent_id=$1 AND utm_source='archived-source'`, agent.ID).Scan(&archivedHash))
	settings.PromotionTrackingEnabled = false
	require.NoError(t, svc.AdminUpdateSettings(ctx, settings, adminID))
	resolved, err := repo.ResolvePromotionAttribution(ctx, activeHash)
	require.NoError(t, err)
	require.NotNil(t, resolved, "stopping new visit collection must not invalidate an existing attribution window")
	settings.PromotionAttributionEnabled = false
	require.NoError(t, svc.AdminUpdateSettings(ctx, settings, adminID))
	resolved, err = repo.ResolvePromotionAttribution(ctx, activeHash)
	require.NoError(t, err)
	require.Nil(t, resolved, "the attribution switch must independently disable historical cookie attribution")
	settings.PromotionTrackingEnabled = true
	settings.PromotionAttributionEnabled = true
	require.NoError(t, svc.AdminUpdateSettings(ctx, settings, adminID))
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_promotion_visits SET visited_at=NOW()-INTERVAL '45 days'
		WHERE agent_id=$1 AND visitor_token_hash=$2`, agent.ID, convertedHash)
	require.NoError(t, err)
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_promotion_visits SET visited_at=NOW()-INTERVAL '401 days'
		WHERE agent_id=$1 AND visitor_token_hash=$2`, agent.ID, activeHash)
	require.NoError(t, err)
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_promotion_visits SET visited_at=NOW()-INTERVAL '400 days'
		WHERE agent_id=$1 AND visitor_token_hash=$2`, agent.ID, archivedHash)
	require.NoError(t, err)
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_promotion_attributions SET expires_at=CASE visitor_token_hash
		WHEN $2 THEN NOW()-INTERVAL '1 day' ELSE NOW()+INTERVAL '1 day' END
		WHERE agent_id=$1 AND visitor_token_hash IN ($2,$3)`, agent.ID, convertedHash, activeHash)
	require.NoError(t, err)
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_promotion_attributions SET expires_at=NOW()-INTERVAL '1 day'
		WHERE agent_id=$1 AND visitor_token_hash=$2`, agent.ID, archivedHash)
	require.NoError(t, err)
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_promotion_visits
		SET ip_hash='private-ip-hash',landing_path='/private/path',utm_medium='private-medium',utm_campaign='private-campaign'
		WHERE agent_id=$1 AND visitor_token_hash=$2`, agent.ID, convertedHash)
	require.NoError(t, err)
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_promotion_maintenance SET last_privacy_scrub_at='-infinity' WHERE id=1`)
	require.NoError(t, err)
	var rawBeforeScrub int
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_promotion_visits WHERE agent_id=$1`, agent.ID).Scan(&rawBeforeScrub))
	scrubResult, err := repo.ScrubPromotionDetails(ctx, 100)
	require.NoError(t, err)
	require.Positive(t, scrubResult.ScrubbedRows)
	require.Positive(t, scrubResult.ExpiredAttributions)
	var scrubbedIP, scrubbedMedium, scrubbedCampaign *string
	var scrubbedLanding string
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT ip_hash,landing_path,utm_medium,utm_campaign
		FROM distribution_promotion_visits WHERE agent_id=$1 AND visitor_token_hash=$2`, agent.ID, convertedHash).
		Scan(&scrubbedIP, &scrubbedLanding, &scrubbedMedium, &scrubbedCampaign))
	require.Nil(t, scrubbedIP)
	require.Equal(t, "/expired", scrubbedLanding)
	require.Nil(t, scrubbedMedium)
	require.Nil(t, scrubbedCampaign)
	var rawAfterScrub, expiredAttributionCount int
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_promotion_visits WHERE agent_id=$1`, agent.ID).Scan(&rawAfterScrub))
	require.Equal(t, rawBeforeScrub, rawAfterScrub, "privacy scrub must not physically delete raw visits")
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_promotion_attributions
		WHERE agent_id=$1 AND expires_at<NOW()`, agent.ID).Scan(&expiredAttributionCount))
	require.Zero(t, expiredAttributionCount)
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_promotion_maintenance SET last_cleanup_at='-infinity' WHERE id=1`)
	require.NoError(t, err)

	hongKong, err := time.LoadLocation("Asia/Hong_Kong")
	require.NoError(t, err)
	archiveNow := time.Now().In(hongKong)
	archiveFromDate := archiveNow.AddDate(0, 0, -402)
	archiveToDate := archiveNow.AddDate(0, 0, -399)
	archiveFilter := service.DistributionPromotionStatsFilter{
		From: time.Date(archiveFromDate.Year(), archiveFromDate.Month(), archiveFromDate.Day(), 0, 0, 0, 0, hongKong),
		To:   time.Date(archiveToDate.Year(), archiveToDate.Month(), archiveToDate.Day(), 0, 0, 0, 0, hongKong),
		Page: 1, PageSize: 20,
	}
	beforeArchive, err := svc.GetPromotionAnalytics(ctx, agentUserID, archiveFilter, false)
	require.NoError(t, err)
	require.EqualValues(t, 3, beforeArchive.Summary.TotalVisits)
	require.EqualValues(t, 2, beforeArchive.Summary.UniqueVisitors)
	require.EqualValues(t, 1, beforeArchive.Summary.ConvertedVisitors)

	cleanupResult, err := repo.CleanupPromotionDetails(ctx, 1)
	require.NoError(t, err)
	require.LessOrEqual(t, cleanupResult.ScrubbedRows, 1, "detail scrubbing must honor batch_size")
	require.LessOrEqual(t, cleanupResult.DeletedRows, 1, "physical archival must honor batch_size")
	require.Equal(t, 1, cleanupResult.DeletedRows)
	require.Equal(t, 1, cleanupResult.SkippedDays)
	var hashAfterFirstBatch *string
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT visitor_token_hash FROM distribution_promotion_conversions WHERE user_id=$1`, archivedCustomerUserID).Scan(&hashAfterFirstBatch))
	require.NotNil(t, hashAfterFirstBatch, "later batches still need the conversion hash to classify the same visitor")
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_promotion_maintenance SET last_cleanup_at='-infinity' WHERE id=1`)
	require.NoError(t, err)
	secondCleanupResult, err := repo.CleanupPromotionDetails(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, 1, secondCleanupResult.DeletedRows)
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_promotion_maintenance SET last_cleanup_at='-infinity' WHERE id=1`)
	require.NoError(t, err)
	thirdCleanupResult, err := repo.CleanupPromotionDetails(ctx, 1)
	require.NoError(t, err)
	require.Zero(t, thirdCleanupResult.ScrubbedRows, "independent privacy scrub already removed all expired detail")

	var convertedVisits, activeVisits, convertedAttributions, activeAttributions int
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_promotion_visits WHERE agent_id=$1 AND visitor_token_hash=$2`, agent.ID, convertedHash).Scan(&convertedVisits))
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_promotion_visits WHERE agent_id=$1 AND visitor_token_hash=$2`, agent.ID, activeHash).Scan(&activeVisits))
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_promotion_attributions WHERE agent_id=$1 AND visitor_token_hash=$2`, agent.ID, convertedHash).Scan(&convertedAttributions))
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_promotion_attributions WHERE agent_id=$1 AND visitor_token_hash=$2`, agent.ID, activeHash).Scan(&activeAttributions))
	require.Equal(t, 1, convertedVisits)
	require.Zero(t, convertedAttributions)
	require.Equal(t, 1, activeVisits)
	require.Equal(t, 1, activeAttributions)
	var skippedReason string
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT s.reason FROM distribution_promotion_archive_skips s
		JOIN distribution_promotion_visits v ON s.stat_date=(v.visited_at AT TIME ZONE 'Asia/Hong_Kong')::date
		WHERE v.agent_id=$1 AND v.visitor_token_hash=$2`, agent.ID, activeHash).Scan(&skippedReason))
	require.Equal(t, "active_attribution", skippedReason)

	var source, landingPath string
	var attributedVisitID *int64
	var detailExpiredAt *time.Time
	var ipHash *string
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT c.attribution_source,c.attribution_visit_id,
		v.detail_expired_at,v.ip_hash,v.landing_path FROM distribution_promotion_conversions c
		JOIN distribution_promotion_visits v ON v.id=c.attribution_visit_id WHERE c.user_id=$1`, customerUserID).
		Scan(&source, &attributedVisitID, &detailExpiredAt, &ipHash, &landingPath))
	require.Equal(t, "retained-source", source)
	require.NotNil(t, attributedVisitID)
	require.NotNil(t, detailExpiredAt)
	require.Nil(t, ipHash)
	require.Equal(t, "/expired", landingPath)
	var archivedVisits, archivedRollupVisits int
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_promotion_visits WHERE agent_id=$1 AND visitor_token_hash=$2`, agent.ID, archivedHash).Scan(&archivedVisits))
	require.Zero(t, archivedVisits)
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COALESCE(SUM(visits),0) FROM distribution_promotion_daily_rollups WHERE agent_id=$1 AND source='archived-source'`, agent.ID).Scan(&archivedRollupVisits))
	require.Equal(t, 2, archivedRollupVisits)
	var archivedConversionHash *string
	var archivedFirst, archivedLast, archivedCredited *int64
	var archivedSource string
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT visitor_token_hash,first_visit_id,last_visit_id,attribution_visit_id,attribution_source
		FROM distribution_promotion_conversions WHERE user_id=$1`, archivedCustomerUserID).
		Scan(&archivedConversionHash, &archivedFirst, &archivedLast, &archivedCredited, &archivedSource))
	require.Nil(t, archivedConversionHash)
	require.Nil(t, archivedFirst)
	require.Nil(t, archivedLast)
	require.Nil(t, archivedCredited)
	require.Equal(t, "archived-source", archivedSource)
	afterArchive, err := svc.GetPromotionAnalytics(ctx, agentUserID, archiveFilter, false)
	require.NoError(t, err)
	require.Equal(t, beforeArchive.Summary, afterArchive.Summary)
	require.Equal(t, beforeArchive.Daily, afterArchive.Daily)
	require.ElementsMatch(t, beforeArchive.Sources, afterArchive.Sources)
	archivedSourceFilter := archiveFilter
	archivedSourceFilter.Source = "ARCHIVED-SOURCE"
	archivedSourceAnalytics, err := svc.GetPromotionAnalytics(ctx, agentUserID, archivedSourceFilter, false)
	require.NoError(t, err)
	require.EqualValues(t, 2, archivedSourceAnalytics.Summary.TotalVisits)
	require.EqualValues(t, 1, archivedSourceAnalytics.Summary.ConvertedVisitors)
	persistedArchiveFilter := archivedSourceFilter
	persistedArchiveFilter.AttributionType = "persisted"
	persistedArchiveAnalytics, err := svc.GetPromotionAnalytics(ctx, agentUserID, persistedArchiveFilter, false)
	require.NoError(t, err)
	require.EqualValues(t, 2, persistedArchiveAnalytics.Summary.TotalVisits)
	require.EqualValues(t, 1, persistedArchiveAnalytics.Summary.ConvertedVisitors)
	archivedRegistrationFilter := archivedSourceFilter
	archivedRegistrationFilter.Device = "desktop"
	archivedRegistrationAnalytics, err := svc.GetPromotionAnalytics(ctx, agentUserID, archivedRegistrationFilter, false)
	require.NoError(t, err)
	require.EqualValues(t, 1, archivedRegistrationAnalytics.Summary.Registrations,
		"registration device/source filters must survive raw visit deletion")
	partialArchiveFilter := archiveFilter
	partialDate := archiveNow.AddDate(0, 0, -400)
	partialArchiveFilter.From = time.Date(partialDate.Year(), partialDate.Month(), partialDate.Day(), 1, 0, 0, 0, hongKong)
	partialArchiveFilter.To = partialArchiveFilter.From.Add(12 * time.Hour)
	_, err = svc.GetPromotionAnalytics(ctx, agentUserID, partialArchiveFilter, false)
	require.ErrorIs(t, err, service.ErrPromotionArchiveUnavailable)
	legacyDate := archiveNow.AddDate(0, 0, -398)
	legacyStatDate := time.Date(legacyDate.Year(), legacyDate.Month(), legacyDate.Day(), 0, 0, 0, 0, hongKong)
	_, err = integrationDB.ExecContext(ctx, `INSERT INTO distribution_promotion_daily_rollups(
		stat_date,agent_id,source,device_type,is_bot,rollup_kind,visits,unique_visitors
	) VALUES($1,$2,'legacy-source','desktop',FALSE,'legacy',4,3)`, legacyStatDate, agent.ID)
	require.NoError(t, err)
	legacyFilter := service.DistributionPromotionStatsFilter{From: legacyStatDate, To: legacyStatDate.AddDate(0, 0, 1)}
	_, err = svc.GetPromotionAnalytics(ctx, agentUserID, legacyFilter, false)
	require.ErrorIs(t, err, service.ErrPromotionArchiveUnavailable, "legacy aggregates must never be presented as exact analytics")

	oldFilter := service.DistributionPromotionStatsFilter{From: time.Now().Add(-46 * 24 * time.Hour), To: time.Now().Add(-44 * 24 * time.Hour), Page: 1, PageSize: 20}
	details, total, err := svc.ListPromotionVisits(ctx, agentUserID, oldFilter, false)
	require.NoError(t, err)
	require.Empty(t, details)
	require.Zero(t, total)
	analytics, err := svc.GetPromotionAnalytics(ctx, agentUserID, oldFilter, false)
	require.NoError(t, err)
	require.EqualValues(t, 1, analytics.Summary.TotalVisits)
	require.EqualValues(t, 1, analytics.Summary.ConvertedVisitors)
	require.EqualValues(t, 1, analytics.Summary.TrackedRegistrations)
	require.EqualValues(t, 1, analytics.Summary.Registrations, "visit cohort keeps registration and conversion on the attributed visit date")

	var retryScrubbed, retryDeleted, retrySkipped int
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT * FROM cleanup_distribution_promotion_details(100)`).
		Scan(&retryScrubbed, &retryDeleted, &retrySkipped))
	require.Zero(t, retryScrubbed+retryDeleted+retrySkipped, "the global maintenance guard must prevent concurrent cleanup storms")
	require.Error(t, integrationDB.QueryRowContext(ctx, `SELECT * FROM cleanup_distribution_promotion_details(0)`).
		Scan(&retryScrubbed, &retryDeleted, &retrySkipped))
	require.Error(t, integrationDB.QueryRowContext(ctx, `SELECT * FROM cleanup_distribution_promotion_details(10001)`).
		Scan(&retryScrubbed, &retryDeleted, &retrySkipped))
}
