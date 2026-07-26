//go:build integration

package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestDistributionAndAffiliateOwnershipAreConcurrentExclusive(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	createUser := func(role string) int64 {
		u, err := integrationEntClient.User.Create().
			SetEmail(fmt.Sprintf("ownership-%s-%d@example.com", role, suffix)).
			SetPasswordHash("hash").Save(ctx)
		require.NoError(t, err)
		return u.ID
	}
	adminID := createUser("admin")
	agentUserID := createUser("agent")
	inviterID := createUser("inviter")
	customerID := createUser("customer")

	repo := &distributionRepository{db: integrationDB}
	distributionSvc := service.NewDistributionService(repo)
	affiliateRepo := NewAffiliateRepository(integrationEntClient, integrationDB)
	original, err := distributionSvc.AdminGetSettings(ctx)
	require.NoError(t, err)
	settings := *original
	settings.Enabled = true
	require.NoError(t, distributionSvc.AdminUpdateSettings(ctx, settings, adminID))
	code := fmt.Sprintf("OWN%d", suffix%1_000_000_000)
	agent, err := distributionSvc.AdminGrantAgent(ctx, service.DistributionGrantAgentInput{
		UserID: agentUserID, Depth: 1, PromotionCode: code, GrantedBy: adminID,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_binding_claims WHERE user_id=$1`, customerID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_promotion_conversions WHERE user_id=$1`, customerID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_customer_bindings WHERE user_id=$1`, customerID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM user_affiliates WHERE user_id IN ($1,$2)`, customerID, inviterID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM user_promotion_ownerships WHERE user_id=$1`, customerID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_wallets WHERE agent_id=$1`, agent.ID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_agent_events WHERE agent_id=$1`, agent.ID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_agents WHERE id=$1`, agent.ID)
		_ = distributionSvc.AdminUpdateSettings(cleanupCtx, *original, adminID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `UPDATE distribution_settings SET updated_by=NULL WHERE updated_by=$1`, adminID)
		_, _ = integrationDB.ExecContext(cleanupCtx, `DELETE FROM users WHERE id IN ($1,$2,$3,$4)`, adminID, agentUserID, inviterID, customerID)
	})

	require.NoError(t, repo.QueueDistributionBindingClaim(ctx, customerID, code, "integration"))
	claimCode, pending, err := repo.PendingDistributionBindingClaim(ctx, customerID)
	require.NoError(t, err)
	require.True(t, pending)
	require.Equal(t, code, claimCode)

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	var distributionErr, affiliateErr error
	var affiliateBound bool
	go func() {
		defer wg.Done()
		<-start
		distributionErr = distributionSvc.BindCustomerByCode(ctx, customerID, code)
	}()
	go func() {
		defer wg.Done()
		<-start
		affiliateBound, affiliateErr = affiliateRepo.BindInviter(ctx, customerID, inviterID)
	}()
	close(start)
	wg.Wait()

	var distributionCount, affiliateCount int
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_customer_bindings WHERE user_id=$1`, customerID).Scan(&distributionCount))
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_affiliates WHERE user_id=$1 AND inviter_id IS NOT NULL`, customerID).Scan(&affiliateCount))
	require.Equal(t, 1, distributionCount+affiliateCount)
	if distributionCount == 1 {
		require.NoError(t, distributionErr)
		require.ErrorIs(t, affiliateErr, service.ErrDistributionCodeConflict)
		require.False(t, affiliateBound)
	} else {
		require.NoError(t, affiliateErr)
		require.True(t, affiliateBound)
		require.ErrorIs(t, distributionErr, service.ErrDistributionCodeConflict)
	}
}
