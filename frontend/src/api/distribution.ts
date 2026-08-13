import type { PaginatedResponse } from "@/types";
import { apiClient } from "./client";

export interface DistributionAgent {
  id: number;
  user_id: number;
  depth: 1 | 2;
  parent_agent_id?: number;
  promotion_code: string;
  rate_override_bps?: number;
  effective_rate_bps: number;
  max_child_rate_bps: number;
  can_recruit_subagents: boolean;
  can_view_promotion_stats: boolean;
  status: "active" | "suspended" | "revoked";
  available_cny: string;
  frozen_cny: string;
  reserved_cny: string;
  debt_cny: string;
  total_earned_cny: string;
  total_withdrawn_cny: string;
  total_converted_cny: string;
  email?: string;
  username?: string;
  user_status?: string;
  customer_count?: number;
  team_count?: number;
  paying_customer_count: number;
  customer_paid_cny: string;
  this_month_commission_cny: string;
  period_customer_count: number;
  period_paying_customers: number;
  period_customer_paid_cny: string;
  period_commission_cny: string;
  team_customer_count: number;
  team_paying_customers: number;
  team_customer_paid_cny: string;
  team_commission_cny: string;
  last_commission_at?: string;
  created_at?: string;
}

export interface DistributionBusinessMetrics {
  new_customers: number;
  paying_customers: number;
  paid_orders: number;
  customer_paid_cny: string;
  commission_cny: string;
  conversion_rate: string;
  average_order_cny: string;
}

export interface DistributionPeriodComparison {
  current: DistributionBusinessMetrics;
  previous: DistributionBusinessMetrics;
  paid_growth_rate?: string;
  commission_growth_rate?: string;
}

export interface DistributionDailyMetric {
  date: string;
  new_customers: number;
  paying_customers: number;
  customer_paid_cny: string;
  commission_cny: string;
}

export interface DistributionBusinessAnalytics {
  days: number;
  date_from: string;
  date_to: string;
  trend_resolution: 'hour' | 'day' | 'week' | 'month';
  direct: DistributionPeriodComparison;
  team: DistributionPeriodComparison;
  total: DistributionPeriodComparison;
  daily_direct: DistributionDailyMetric[];
  daily_team: DistributionDailyMetric[];
  active_agents: number;
}

export interface DistributionAgentRanking {
  agent_id: number;
  email: string;
  username: string;
  depth: 1 | 2;
  customer_paid_cny: string;
  commission_cny: string;
  new_customers: number;
  paying_customers: number;
  conversion_rate: string;
}

export interface DistributionAgentAnalytics {
  agent: DistributionAgent;
  analytics: DistributionBusinessAnalytics;
  ranking: DistributionAgentRanking[];
}
export interface DistributionRewardRule {
  agent_id: number;
  registration_enabled: boolean;
  registration_reward_cny: string;
  recharge_enabled: boolean;
  recharge_threshold_cny: string;
  recharge_reward_cny: string;
}

export interface DistributionOverview {
  agent: DistributionAgent;
  distribution_enabled: boolean;
  promotion_tracking_enabled: boolean;
  customer_count: number;
  team_count: number;
  paying_customer_count: number;
  new_customers_this_month: number;
  customer_paid_cny: string;
  this_month_customer_paid_cny: string;
  this_month_commission_cny: string;
  analytics: DistributionBusinessAnalytics;
  team_ranking: DistributionAgentRanking[];
}

export interface DistributionAccess {
  enabled: boolean;
  is_agent: boolean;
  depth?: 1 | 2;
  status?: DistributionAgent["status"];
  can_recruit_subagents: boolean;
  can_view_promotion_stats: boolean;
}

export interface DistributionPromotionSummary {
  total_visits: number;
  unique_visitors: number;
  bot_visits: number;
  converted_visitors: number;
  tracked_registrations: number;
  registrations: number;
  direct_registrations: number;
  persisted_registrations: number;
  untracked_direct: number;
  conversion_rate: number;
}
export interface DistributionPromotionAnalytics {
  summary: DistributionPromotionSummary;
  daily: DistributionPromotionDailyStat[];
  sources: DistributionPromotionSourceStat[];
  meta?: DistributionPromotionAnalyticsMeta;
}
export interface DistributionPromotionAnalyticsMeta {
  cohort: 'visit';
  tracking_enabled: boolean;
  attribution_enabled: boolean;
  attribution_days: number;
  attribution_model: 'first_touch' | 'last_touch';
  bot_filter_enabled: boolean;
  detail_retention_days: number;
  raw_retention_days: number;
  timezone: string;
  generated_at?: string;
}
export interface DistributionPromotionDailyStat {
  date: string;
  visits: number;
  visitors: number;
  conversions: number;
  registrations: number;
}
export interface DistributionPromotionSourceStat {
  source: string;
  visits: number;
  conversions: number;
  registrations: number;
}
export interface DistributionPromotionVisit {
  id: number;
  agent_id: number;
  promotion_code: string;
  landing_path: string;
  source: string;
  device_type: string;
  is_bot: boolean;
  visited_at: string;
  registered_at?: string;
  attribution_type?: 'direct' | 'persisted' | 'unregistered';
  agent_email?: string;
}

export interface DistributionPayoutAccount {
  alipay_name: string;
  alipay_account: string;
}

export interface DistributionWithdrawal {
  id: number;
  amount_cny: string;
  fee_cny: string;
  payout_cny: string;
  status: string;
  created_at: string;
}

export interface DistributionSettlementRules {
  withdrawal_enabled: boolean;
  minimum_withdrawal_cny: string;
  maximum_withdrawal_cny: string;
  withdrawal_fee_rate_bps: number;
  withdrawal_fee_fixed_cny: string;
  daily_withdrawal_limit_cny: string;
  monthly_withdrawal_limit_cny: string;
  cny_per_platform_usd: string;
  freeze_hours: number;
}

export interface DistributionCommission {
  id: number;
  customer_user_id: number;
  customer_email: string;
  customer_username: string;
  entry_type: "direct" | "team";
  payment_type: string;
  payment_order_id: number;
  order_no: string;
  beneficiary_agent_id: number;
  agent_email: string;
  agent_username: string;
  payment_currency: string;
  actual_paid_amount: string;
  fx_rate_to_cny: string;
  rate_bps: number;
  commission_base_cny: string;
  original_amount_cny: string;
  reversed_amount_cny: string;
  status: string;
  available_at: string;
  created_at: string;
}

export interface DistributionCustomer {
  user_id: number;
  email: string;
  username: string;
  registered_at?: string;
  bound_at: string;
  order_count: number;
  total_paid_cny: string;
  commission_cny: string;
  refunded_cny: string;
  last_paid_at?: string;
  agent_id?: number;
  agent_user_id?: number;
  agent_promotion_code?: string;
  agent_email?: string;
  agent_username?: string;
  agent_depth?: 1 | 2;
}

export interface DistributionUserOption {
  user_id: number;
  email: string;
  username: string;
  status: string;
  selectable: boolean;
  unavailable_reason?: string;
}

export interface DistributionListParams {
  page?: number;
  page_size?: number;
  search?: string;
  status?: string;
  entry_type?: string;
  days?: number;
  date_from?: string;
  date_to?: string;
}

export interface DistributionAnalyticsParams {
  days?: number;
  date_from?: string;
  date_to?: string;
}

export async function getDistributionAccess() {
  const { data } = await apiClient.get<DistributionAccess>(
    "/distribution/access",
  );
  return data;
}

export async function trackPromotionVisit(payload: {
  promotion_code: string;
  landing_path?: string;
  referrer?: string;
  utm_source?: string;
  utm_medium?: string;
  utm_campaign?: string;
}) {
  const { data } = await apiClient.post<DistributionPromotionTrackResult>(
    '/distribution/track', payload,
  );
  return data;
}

export async function getPromotionTrackingStatus(): Promise<{ enabled: boolean }> {
  const { data } = await apiClient.get<{ enabled: boolean }>(
    '/distribution/tracking-status',
  );
  return data;
}

export interface DistributionPromotionTrackResult {
  tracked: boolean;
  attribution_days: number;
}

export async function getDistributionOverview(params: number | DistributionAnalyticsParams = 30) {
  const query = typeof params === 'number' ? { days: params } : params;
  const { data } = await apiClient.get<DistributionOverview>("/distribution/overview", { params: query });
  return data;
}

export async function getSettlementRules() {
  const { data } = await apiClient.get<DistributionSettlementRules>(
    "/distribution/settlement-rules",
  );
  return data;
}

export async function getPayoutAccount() {
  const { data } = await apiClient.get<DistributionPayoutAccount>(
    "/distribution/payout-account",
  );
  return data;
}

export async function updatePayoutAccount(payload: DistributionPayoutAccount) {
  await apiClient.put("/distribution/payout-account", payload);
}

export async function requestWithdrawal(amount: string) {
  const { data } = await apiClient.post<DistributionWithdrawal>(
    "/distribution/withdrawals",
    { amount_cny: amount },
  );
  return data;
}

export async function listMyWithdrawals(params: DistributionListParams = {}) {
  const { data } = await apiClient.get<
    PaginatedResponse<DistributionWithdrawal>
  >("/distribution/withdrawals", { params });
  return data;
}

export async function convertToBalance(amount: string) {
  const { data } = await apiClient.post<{ credited_platform_usd: string }>(
    "/distribution/convert-to-balance",
    { amount_cny: amount },
  );
  return data;
}

export async function grantL2Agent(payload: {
  email: string;
  rate_override_bps?: number;
  promotion_code?: string;
}) {
  const { data } = await apiClient.post<DistributionAgent>(
    "/distribution/team",
    payload,
  );
  return data;
}

export async function listTeam(params: DistributionListParams = {}) {
  const { data } = await apiClient.get<PaginatedResponse<DistributionAgent>>(
    "/distribution/team",
    { params },
  );
  return data;
}

export async function getTeamAgentAnalytics(agentId: number, params: number | DistributionAnalyticsParams = 30) {
  const query = typeof params === 'number' ? { days: params } : params;
  const { data } = await apiClient.get<DistributionAgentAnalytics>(`/distribution/team/${agentId}/analytics`, { params: query });
  return data;
}

export async function updateTeamAgentStatus(
  agentId: number,
  status: "active" | "suspended",
) {
  await apiClient.put(`/distribution/team/${agentId}/status`, { status });
}
export async function updateTeamAgentRewardRule(agentId: number, payload: Omit<DistributionRewardRule, "agent_id">) {
  await apiClient.put(`/distribution/team/${agentId}/reward-rule`, payload);
}
export async function getTeamAgentRewardRule(agentId: number) {
  const { data } = await apiClient.get<DistributionRewardRule>(`/distribution/team/${agentId}/reward-rule`);
  return data;
}

export async function listCommissions(params: DistributionListParams = {}) {
  const { data } = await apiClient.get<
    PaginatedResponse<DistributionCommission>
  >("/distribution/commissions", { params });
  return data;
}

export async function listCustomers(params: DistributionListParams = {}) {
  const { data } = await apiClient.get<PaginatedResponse<DistributionCustomer>>(
    "/distribution/customers",
    { params },
  );
  return data;
}

export interface DistributionPromotionListParams extends DistributionListParams {
  date_from?: string;
  date_to?: string;
  source?: string;
  device?: string;
  attribution_type?: string;
}

export async function getPromotionAnalytics(params: DistributionPromotionListParams = {}) {
  const { data } = await apiClient.get<DistributionPromotionAnalytics>('/distribution/promotion/analytics', { params });
  return data;
}

export async function listPromotionVisits(params: DistributionPromotionListParams = {}) {
  const { data } = await apiClient.get<PaginatedResponse<DistributionPromotionVisit>>('/distribution/promotion/visits', { params });
  return data;
}
