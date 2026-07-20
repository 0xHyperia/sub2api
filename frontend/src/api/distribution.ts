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
  last_commission_at?: string;
  created_at?: string;
}

export interface DistributionOverview {
  agent: DistributionAgent;
  customer_count: number;
  team_count: number;
  paying_customer_count: number;
  new_customers_this_month: number;
  customer_paid_cny: string;
  this_month_customer_paid_cny: string;
  this_month_commission_cny: string;
}

export interface DistributionAccess {
  enabled: boolean;
  is_agent: boolean;
  depth?: 1 | 2;
  status?: DistributionAgent["status"];
  can_recruit_subagents: boolean;
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
}

export async function getDistributionAccess() {
  const { data } = await apiClient.get<DistributionAccess>(
    "/distribution/access",
  );
  return data;
}

export async function getDistributionOverview() {
  const { data } = await apiClient.get<DistributionOverview>(
    "/distribution/overview",
  );
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

export async function updateTeamAgentStatus(
  agentId: number,
  status: "active" | "suspended",
) {
  await apiClient.put(`/distribution/team/${agentId}/status`, { status });
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
