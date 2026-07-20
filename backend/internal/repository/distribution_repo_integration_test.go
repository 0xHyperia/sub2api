//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestDistributionDifferentialCommissionAndRefund(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	createUser := func(role string) int64 {
		u, err := integrationEntClient.User.Create().
			SetEmail(fmt.Sprintf("distribution-%s-%d@example.com", role, suffix)).
			SetPasswordHash("hash").Save(ctx)
		require.NoError(t, err)
		return u.ID
	}
	adminID := createUser("admin")
	l1UserID := createUser("l1")
	l2UserID := createUser("l2")
	customerID := createUser("customer")

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, err := integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_wallet_ledger WHERE agent_id IN (SELECT id FROM distribution_agents WHERE user_id IN ($1,$2))`, l1UserID, l2UserID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_commission_entries WHERE source_id IN (SELECT id FROM distribution_commission_sources WHERE customer_user_id=$1)`, customerID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_commission_sources WHERE customer_user_id=$1`, customerID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM payment_orders WHERE user_id=$1`, customerID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_binding_events WHERE customer_user_id=$1`, customerID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_customer_bindings WHERE user_id=$1`, customerID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_wallets WHERE agent_id IN (SELECT id FROM distribution_agents WHERE user_id IN ($1,$2))`, l1UserID, l2UserID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_agent_events WHERE agent_id IN (SELECT id FROM distribution_agents WHERE user_id IN ($1,$2))`, l1UserID, l2UserID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_agents WHERE user_id=$1`, l2UserID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_agents WHERE user_id=$1`, l1UserID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM users WHERE id IN ($1,$2,$3,$4)`, adminID, l1UserID, l2UserID, customerID)
		require.NoError(t, err)
	})

	repo := NewDistributionRepository(integrationDB)
	svc := service.NewDistributionService(repo)
	settings, err := svc.AdminGetSettings(ctx)
	require.NoError(t, err)
	originalSettings := *settings
	t.Cleanup(func() {
		require.NoError(t, svc.AdminUpdateSettings(context.Background(), originalSettings, adminID))
		_, cleanupErr := integrationDB.ExecContext(context.Background(), `DELETE FROM distribution_fx_rates WHERE currency='EUR'`)
		require.NoError(t, cleanupErr)
	})
	settings.Enabled = true
	settings.FreezeHours = 0
	require.NoError(t, svc.AdminUpdateSettings(ctx, *settings, adminID))
	require.NoError(t, svc.AdminSetFXRate(ctx, "USD", decimal.NewFromInt(7), adminID))

	l1, err := svc.AdminGrantAgent(ctx, service.DistributionGrantAgentInput{UserID: l1UserID, Depth: 1, PromotionCode: "L1TESTCODE", GrantedBy: adminID})
	require.NoError(t, err)
	require.NoError(t, svc.AdminUpdateAgentRecruitmentPermission(ctx, l1.ID, adminID, true, "integration recruitment"))
	l2, err := svc.GrantL2Agent(ctx, l1UserID, service.DistributionGrantChildAgentInput{Email: fmt.Sprintf("distribution-l2-%d@example.com", suffix), PromotionCode: "L2TESTCODE"})
	require.NoError(t, err)
	require.NoError(t, svc.BindCustomerByCode(ctx, customerID, l2.PromotionCode))

	now := time.Now().UTC()
	order, err := integrationEntClient.PaymentOrder.Create().
		SetUserID(customerID).SetUserEmail("customer@example.com").SetUserName("customer").
		SetAmount(100).SetPayAmount(100).SetRechargeCode(fmt.Sprintf("DIST-%d", suffix)).
		SetOutTradeNo(fmt.Sprintf("DISTORDER%d", suffix)).SetPaymentType("stripe").SetPaymentTradeNo(fmt.Sprintf("trade-%d", suffix)).
		SetProviderSnapshot(map[string]any{"currency": "USD"}).SetStatus("COMPLETED").
		SetClientIP("127.0.0.1").SetSrcHost("localhost").SetExpiresAt(now.Add(time.Hour)).SetPaidAt(now).Save(ctx)
	require.NoError(t, err)

	result, err := svc.AccruePaidOrder(ctx, service.DistributionCommissionInput{
		PaymentOrderID: order.ID, CustomerUserID: customerID, PaymentType: "stripe", PaymentCurrency: "USD",
		ActualPaid: decimal.NewFromInt(100), PaidAt: now,
	})
	require.NoError(t, err)
	require.True(t, result.Created)
	require.True(t, result.BaseCNY.Equal(decimal.NewFromInt(700)))
	require.True(t, result.Commission.Equal(decimal.NewFromInt(70)))

	l2Overview, err := svc.GetOverview(ctx, l2UserID)
	require.NoError(t, err)
	require.True(t, l2Overview.Agent.AvailableCNY.Equal(decimal.NewFromInt(42)))
	l1Overview, err := svc.GetOverview(ctx, l1UserID)
	require.NoError(t, err)
	require.True(t, l1Overview.Agent.AvailableCNY.Equal(decimal.NewFromInt(28)))

	require.NoError(t, svc.ReverseRefund(ctx, order.ID, decimal.NewFromInt(50)))
	l2Overview, err = svc.GetOverview(ctx, l2UserID)
	require.NoError(t, err)
	require.True(t, l2Overview.Agent.AvailableCNY.Equal(decimal.NewFromInt(21)))
	l1Overview, err = svc.GetOverview(ctx, l1UserID)
	require.NoError(t, err)
	require.True(t, l1Overview.Agent.AvailableCNY.Equal(decimal.NewFromInt(14)))

	secondOrder, err := integrationEntClient.PaymentOrder.Create().SetUserID(customerID).SetUserEmail("customer@example.com").SetUserName("customer").SetAmount(100).SetPayAmount(100).SetRechargeCode(fmt.Sprintf("DIST-EUR-%d", suffix)).SetOutTradeNo(fmt.Sprintf("DISTEUR%d", suffix)).SetPaymentType("stripe").SetPaymentTradeNo(fmt.Sprintf("trade-eur-%d", suffix)).SetProviderSnapshot(map[string]any{"currency": "EUR"}).SetStatus("COMPLETED").SetClientIP("127.0.0.1").SetSrcHost("localhost").SetExpiresAt(now.Add(time.Hour)).SetPaidAt(now).Save(ctx)
	require.NoError(t, err)
	pending, err := svc.AccruePaidOrder(ctx, service.DistributionCommissionInput{PaymentOrderID: secondOrder.ID, CustomerUserID: customerID, PaymentType: "stripe", PaymentCurrency: "EUR", ActualPaid: decimal.NewFromInt(100), PaidAt: now})
	require.NoError(t, err)
	require.Equal(t, "pending_fx", pending.Status)
	require.NoError(t, svc.AdminSetFXRate(ctx, "EUR", decimal.NewFromInt(8), adminID))
	l2Overview, err = svc.GetOverview(ctx, l2UserID)
	require.NoError(t, err)
	require.True(t, l2Overview.Agent.AvailableCNY.Equal(decimal.NewFromInt(69)))
	l1Overview, err = svc.GetOverview(ctx, l1UserID)
	require.NoError(t, err)
	require.True(t, l1Overview.Agent.AvailableCNY.Equal(decimal.NewFromInt(46)))
}

func TestDistributionAdminManagementQueries(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	createUser := func(role string) int64 {
		u, err := integrationEntClient.User.Create().
			SetEmail(fmt.Sprintf("distribution-management-%s-%d@example.com", role, suffix)).
			SetUsername("management-" + role).
			SetPasswordHash("hash").Save(ctx)
		require.NoError(t, err)
		return u.ID
	}

	adminID := createUser("admin")
	agentUserID := createUser("agent")
	candidateUserID := createUser("candidate")
	customerUserID := createUser("customer")

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, err := integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_binding_events WHERE customer_user_id=$1`, customerUserID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_customer_bindings WHERE user_id=$1`, customerUserID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_wallets WHERE agent_id IN (SELECT id FROM distribution_agents WHERE user_id IN ($1,$2))`, agentUserID, candidateUserID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_agent_events WHERE agent_id IN (SELECT id FROM distribution_agents WHERE user_id IN ($1,$2))`, agentUserID, candidateUserID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_agents WHERE user_id=$1`, candidateUserID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_agents WHERE user_id=$1`, agentUserID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM users WHERE id IN ($1,$2,$3,$4)`, adminID, agentUserID, candidateUserID, customerUserID)
		require.NoError(t, err)
	})

	repo := NewDistributionRepository(integrationDB)
	svc := service.NewDistributionService(repo)
	promotionCode := fmt.Sprintf("MGMT%d", suffix%1_000_000_000)
	agent, err := svc.AdminGrantAgent(ctx, service.DistributionGrantAgentInput{
		UserID: agentUserID, Depth: 1, PromotionCode: promotionCode, GrantedBy: adminID,
	})
	require.NoError(t, err)
	require.NoError(t, svc.AdminCorrectCustomerBinding(ctx, customerUserID, agent.ID, adminID, "integration test"))

	candidates, err := svc.AdminLookupAgentCandidates(ctx, "distribution-management")
	require.NoError(t, err)
	require.NotEmpty(t, candidates)
	selectable := false
	alreadyAgent := false
	for _, candidate := range candidates {
		switch candidate.UserID {
		case candidateUserID:
			selectable = candidate.Selectable
		case agentUserID:
			alreadyAgent = !candidate.Selectable && candidate.UnavailableReason == "already_agent"
		}
	}
	require.True(t, selectable)
	require.True(t, alreadyAgent)

	_, err = svc.GrantL2Agent(ctx, agentUserID, service.DistributionGrantChildAgentInput{Email: fmt.Sprintf("distribution-management-candidate-%d@example.com", suffix)})
	require.ErrorIs(t, err, service.ErrSubagentRecruitmentDenied)
	require.NoError(t, svc.AdminUpdateAgentRecruitmentPermission(ctx, agent.ID, adminID, true, "approved recruitment"))
	access, err := svc.GetAccess(ctx, agentUserID)
	require.NoError(t, err)
	require.True(t, access.CanRecruitSubagents)
	_, err = svc.GrantL2Agent(ctx, agentUserID, service.DistributionGrantChildAgentInput{Email: "management-candidate"})
	require.ErrorIs(t, err, service.ErrSubagentCandidateUnavailable)
	_, err = repo.LookupEligibleUserByExactEmail(ctx, "management-candidate")
	require.Error(t, err)
	exactCandidateID, err := repo.LookupEligibleUserByExactEmail(ctx, fmt.Sprintf("distribution-management-candidate-%d@example.com", suffix))
	require.NoError(t, err)
	require.Equal(t, candidateUserID, exactCandidateID)

	agentOptions, err := svc.AdminLookupAgents(ctx, "management-agent")
	require.NoError(t, err)
	require.Len(t, agentOptions, 1)
	require.Equal(t, agent.ID, agentOptions[0].AgentID)

	agents, totalAgents, err := svc.AdminListAgents(ctx, service.DistributionAdminListFilter{
		Page: 1, PageSize: 20, Search: "management-agent", Depth: 1, Status: "active",
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, totalAgents)
	require.Len(t, agents, 1)
	require.Contains(t, agents[0].Email, "distribution-management-agent")
	require.EqualValues(t, 1, agents[0].CustomerCount)
	require.Nil(t, agents[0].RateOverrideBPS)

	overrideRate := 850
	require.NoError(t, svc.AdminUpdateAgentRate(ctx, agent.ID, adminID, &overrideRate, "enterprise agreement"))
	agents, totalAgents, err = svc.AdminListAgents(ctx, service.DistributionAdminListFilter{
		Page: 1, PageSize: 20, Search: "management-agent", SortBy: "effective_rate", SortOrder: "desc",
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, totalAgents)
	require.NotNil(t, agents[0].RateOverrideBPS)
	require.Equal(t, overrideRate, *agents[0].RateOverrideBPS)
	require.Equal(t, overrideRate, agents[0].EffectiveRateBPS)
	var rateEventCount int
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM distribution_agent_events WHERE agent_id=$1 AND event_type='rate_changed' AND reason='enterprise agreement'`, agent.ID).Scan(&rateEventCount))
	require.Equal(t, 1, rateEventCount)

	customers, totalCustomers, err := svc.AdminListCustomers(ctx, service.DistributionAdminListFilter{
		Page: 1, PageSize: 20, Search: "management-customer", AgentID: agent.ID, Depth: 1,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, totalCustomers)
	require.Len(t, customers, 1)
	require.Equal(t, customerUserID, customers[0].UserID)
	require.Equal(t, agentUserID, customers[0].AgentUserID)
	require.Contains(t, customers[0].AgentEmail, "distribution-management-agent")

	userCustomers, totalUserCustomers, err := svc.ListCustomers(ctx, agentUserID, service.DistributionUserListFilter{
		Page: 1, PageSize: 10, Search: "management-customer",
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, totalUserCustomers)
	require.Len(t, userCustomers, 1)
	require.Contains(t, userCustomers[0].Email, "***@example.com")
	require.Zero(t, userCustomers[0].OrderCount)

	userCommissions, totalUserCommissions, err := svc.ListCommissions(ctx, agentUserID, service.DistributionUserListFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Zero(t, totalUserCommissions)
	require.Empty(t, userCommissions)

	team, totalTeam, err := svc.ListTeam(ctx, agentUserID, service.DistributionUserListFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Zero(t, totalTeam)
	require.Empty(t, team)

	withdrawals, totalWithdrawals, err := svc.ListWithdrawals(ctx, agentUserID, service.DistributionUserListFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Zero(t, totalWithdrawals)
	require.Empty(t, withdrawals)

	overview, err := svc.GetOverview(ctx, agentUserID)
	require.NoError(t, err)
	require.EqualValues(t, 1, overview.CustomerCount)
	require.Zero(t, overview.PayingCustomerCount)
	require.True(t, overview.CustomerPaidCNY.IsZero())

	rules, err := svc.GetSettlementRules(ctx)
	require.NoError(t, err)
	require.True(t, rules.CNYPerPlatformUSD.IsPositive())

	child, err := svc.GrantL2Agent(ctx, agentUserID, service.DistributionGrantChildAgentInput{Email: fmt.Sprintf("distribution-management-candidate-%d@example.com", suffix)})
	require.NoError(t, err)
	team, totalTeam, err = svc.ListTeam(ctx, agentUserID, service.DistributionUserListFilter{Page: 1, PageSize: 10, Search: "management-candidate"})
	require.NoError(t, err)
	require.EqualValues(t, 1, totalTeam)
	require.Len(t, team, 1)
	require.Contains(t, team[0].Email, "distribution-management-candidate")
	require.NoError(t, svc.UpdateTeamAgentStatus(ctx, agentUserID, child.ID, "suspended"))
	team, totalTeam, err = svc.ListTeam(ctx, agentUserID, service.DistributionUserListFilter{Page: 1, PageSize: 10, Status: "suspended"})
	require.NoError(t, err)
	require.EqualValues(t, 1, totalTeam)
	require.Equal(t, "suspended", team[0].Status)

	require.NoError(t, svc.AdminUpdateAgentStatus(ctx, agent.ID, adminID, "revoked", "contract ended"))
	require.False(t, svc.IsAgent(ctx, agentUserID))

	require.NoError(t, svc.AdminUpdateAgentStatus(ctx, agent.ID, adminID, "active", "approved reactivation"))
	require.True(t, svc.IsAgent(ctx, agentUserID))
	var status, reason string
	require.NoError(t, integrationDB.QueryRowContext(ctx,
		`SELECT status,status_reason FROM distribution_agents WHERE id=$1`, agent.ID,
	).Scan(&status, &reason))
	require.Equal(t, "active", status)
	require.Equal(t, "approved reactivation", reason)
}

func TestDistributionWithdrawalAuditWorkflow(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	createUser := func(role string) int64 {
		u, err := integrationEntClient.User.Create().
			SetEmail(fmt.Sprintf("distribution-withdrawal-%s-%d@example.com", role, suffix)).
			SetUsername("withdrawal-" + role).
			SetPasswordHash("hash").Save(ctx)
		require.NoError(t, err)
		return u.ID
	}
	adminID := createUser("admin")
	paymentAdminID := createUser("payment-admin")
	agentUserID := createUser("agent")

	repo := NewDistributionRepository(integrationDB)
	svc := service.NewDistributionService(repo)
	agent, err := svc.AdminGrantAgent(ctx, service.DistributionGrantAgentInput{
		UserID: agentUserID, Depth: 1, PromotionCode: fmt.Sprintf("WD%d", suffix%1_000_000_000), GrantedBy: adminID,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, err := integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_withdrawal_attachments WHERE withdrawal_id IN (SELECT id FROM distribution_withdrawals WHERE agent_id=$1)`, agent.ID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_withdrawal_events WHERE withdrawal_id IN (SELECT id FROM distribution_withdrawals WHERE agent_id=$1)`, agent.ID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_wallet_ledger WHERE agent_id=$1`, agent.ID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_withdrawals WHERE agent_id=$1`, agent.ID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_payout_accounts WHERE agent_id=$1`, agent.ID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_wallets WHERE agent_id=$1`, agent.ID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_agent_events WHERE agent_id=$1`, agent.ID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM distribution_agents WHERE id=$1`, agent.ID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(cleanupCtx, `DELETE FROM users WHERE id IN ($1,$2,$3)`, adminID, paymentAdminID, agentUserID)
		require.NoError(t, err)
	})

	settings, err := svc.AdminGetSettings(ctx)
	require.NoError(t, err)
	originalSettings := *settings
	t.Cleanup(func() {
		require.NoError(t, svc.AdminUpdateSettings(context.Background(), originalSettings, adminID))
	})
	settings.Enabled = true
	settings.WithdrawalEnabled = true
	settings.MinimumWithdrawalCNY = decimal.NewFromInt(1)
	settings.MaximumWithdrawalCNY = decimal.NewFromInt(50000)
	require.NoError(t, svc.AdminUpdateSettings(ctx, *settings, adminID))
	require.NoError(t, svc.UpsertPayoutAccount(ctx, agentUserID, service.DistributionPayoutAccount{AlipayName: "Audit Agent", AlipayAccount: "audit@example.com"}))
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_wallets SET available_cny=500,total_earned_cny=500 WHERE agent_id=$1`, agent.ID)
	require.NoError(t, err)

	withdrawal, err := svc.RequestWithdrawal(ctx, agentUserID, decimal.NewFromInt(100))
	require.NoError(t, err)
	require.NotEmpty(t, withdrawal.RequestNo)
	require.Equal(t, "pending", withdrawal.Status)
	require.NoError(t, svc.AdminReviewWithdrawal(ctx, withdrawal.ID, adminID, "approved", "identity and amount checked", ""))
	require.NoError(t, svc.AdminReviewWithdrawal(ctx, withdrawal.ID, adminID, "paying", "bank transfer initiated", ""))

	require.NoError(t, repo.AdminCreateWithdrawalAttachments(ctx, withdrawal.ID, adminID, []service.DistributionWithdrawalAttachment{{
		ObjectKey: fmt.Sprintf("distribution/test/%d", suffix), OriginalName: "receipt.png", ContentType: "image/png",
		SizeBytes: 128, SHA256: fmt.Sprintf("%064d", suffix%1_000_000), EvidenceType: "payment_receipt", Note: "integration evidence",
	}}))
	require.NoError(t, svc.AdminReviewWithdrawal(ctx, withdrawal.ID, adminID, "paid", "payment confirmed", fmt.Sprintf("BANK-%d", suffix)))

	detail, err := svc.AdminGetWithdrawal(ctx, withdrawal.ID)
	require.NoError(t, err)
	require.Equal(t, "paid", detail.Withdrawal.Status)
	require.Len(t, detail.Attachments, 1)
	require.GreaterOrEqual(t, len(detail.Events), 4)
	require.Equal(t, "payment_receipt", detail.Attachments[0].EvidenceType)
	var available, reserved, withdrawn decimal.Decimal
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT available_cny,reserved_cny,total_withdrawn_cny FROM distribution_wallets WHERE agent_id=$1`, agent.ID).Scan(&available, &reserved, &withdrawn))
	require.True(t, available.Equal(decimal.NewFromInt(400)))
	require.True(t, reserved.IsZero())
	require.True(t, withdrawn.Equal(decimal.NewFromInt(100)))

	withoutEvidence, err := svc.RequestWithdrawal(ctx, agentUserID, decimal.NewFromInt(50))
	require.NoError(t, err)
	require.NoError(t, svc.AdminReviewWithdrawal(ctx, withoutEvidence.ID, adminID, "approved", "identity and amount checked", ""))
	require.NoError(t, svc.AdminReviewWithdrawal(ctx, withoutEvidence.ID, adminID, "paying", "bank transfer initiated", ""))
	require.NoError(t, svc.AdminReviewWithdrawal(ctx, withoutEvidence.ID, adminID, "paid", "payment confirmed without optional evidence", fmt.Sprintf("BANK-NO-EVIDENCE-%d", suffix)))
	withoutEvidenceDetail, err := svc.AdminGetWithdrawal(ctx, withoutEvidence.ID)
	require.NoError(t, err)
	require.Equal(t, "paid", withoutEvidenceDetail.Withdrawal.Status)
	require.Empty(t, withoutEvidenceDetail.Attachments)

	batchA, err := svc.RequestWithdrawal(ctx, agentUserID, decimal.NewFromInt(50))
	require.NoError(t, err)
	batchB, err := svc.RequestWithdrawal(ctx, agentUserID, decimal.NewFromInt(60))
	require.NoError(t, err)
	batchResult, err := svc.AdminBatchReviewWithdrawals(ctx, service.DistributionBatchWithdrawalReviewInput{
		WithdrawalIDs: []int64{batchA.ID, batchB.ID}, Status: "approved", Note: "batch identity review",
	}, adminID)
	require.NoError(t, err)
	require.Equal(t, 2, batchResult.Updated)

	// A mixed-state batch must roll back every item instead of partially updating the queue.
	_, err = svc.AdminBatchReviewWithdrawals(ctx, service.DistributionBatchWithdrawalReviewInput{
		WithdrawalIDs: []int64{batchA.ID, withdrawal.ID}, Status: "rejected", Note: "invalid mixed batch",
	}, adminID)
	require.Error(t, err)
	batchADetail, err := svc.AdminGetWithdrawal(ctx, batchA.ID)
	require.NoError(t, err)
	require.Equal(t, "approved", batchADetail.Withdrawal.Status)

	batchResult, err = svc.AdminBatchReviewWithdrawals(ctx, service.DistributionBatchWithdrawalReviewInput{
		WithdrawalIDs: []int64{batchA.ID, batchB.ID}, Status: "rejected", Note: "batch request declined",
	}, adminID)
	require.NoError(t, err)
	require.Equal(t, 2, batchResult.Updated)

	overdue, err := svc.RequestWithdrawal(ctx, agentUserID, decimal.NewFromInt(40))
	require.NoError(t, err)
	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_withdrawals SET created_at=NOW()-INTERVAL '2 days' WHERE id=$1`, overdue.ID)
	require.NoError(t, err)
	anomalies, anomalyTotal, err := svc.AdminListAnomalies(ctx, service.DistributionAdminAnomalyListFilter{
		Page: 1, PageSize: 20, Type: "overdue_withdrawal", Search: overdue.RequestNo,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, anomalyTotal)
	require.Len(t, anomalies, 1)
	require.Equal(t, overdue.ID, anomalies[0].EntityID)
	require.Equal(t, "high", anomalies[0].Severity)

	settings, err = svc.AdminGetSettings(ctx)
	require.NoError(t, err)
	settings.WithdrawalDualApproval = true
	require.NoError(t, svc.AdminUpdateSettings(ctx, *settings, adminID))
	dualControl, err := svc.RequestWithdrawal(ctx, agentUserID, decimal.NewFromInt(30))
	require.NoError(t, err)
	require.NoError(t, svc.AdminReviewWithdrawal(ctx, dualControl.ID, adminID, "approved", "maker review", ""))
	err = svc.AdminReviewWithdrawal(ctx, dualControl.ID, adminID, "paying", "same administrator", "")
	require.Error(t, err)
	require.NoError(t, svc.AdminReviewWithdrawal(ctx, dualControl.ID, paymentAdminID, "paying", "independent payment review", ""))
	dualDetail, err := svc.AdminGetWithdrawal(ctx, dualControl.ID)
	require.NoError(t, err)
	require.NotNil(t, dualDetail.Withdrawal.ApprovedBy)
	require.Equal(t, adminID, *dualDetail.Withdrawal.ApprovedBy)
	require.Contains(t, dualDetail.Withdrawal.ApproverEmail, "distribution-withdrawal-admin")

	_, err = integrationDB.ExecContext(ctx, `UPDATE distribution_wallets SET debt_cny=5 WHERE agent_id=$1`, agent.ID)
	require.NoError(t, err)
	anomalies, anomalyTotal, err = svc.AdminListAnomalies(ctx, service.DistributionAdminAnomalyListFilter{
		Page: 1, PageSize: 20, Type: "agent_debt", Search: "distribution-withdrawal-agent",
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, anomalyTotal)
	require.Len(t, anomalies, 1)
	require.True(t, anomalies[0].AmountCNY.Equal(decimal.NewFromInt(5)))
}
