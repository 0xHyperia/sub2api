import { apiClient } from "../client";
import type {
  DistributionAgent,
  DistributionAgentAnalytics,
  DistributionAgentRanking,
  DistributionBusinessAnalytics,
  DistributionCommission,
  DistributionCustomer,
} from "../distribution";
import type { PaginatedResponse } from "@/types";

export interface DistributionSettings {
  enabled: boolean;
  l1_default_rate_bps: number;
  l1_max_child_rate_bps: number;
  l2_default_rate_bps: number;
  freeze_hours: number;
  withdrawal_enabled: boolean;
  withdrawal_dual_approval_enabled: boolean;
  minimum_withdrawal_cny: string;
  maximum_withdrawal_cny: string;
  withdrawal_fee_rate_bps: number;
  withdrawal_fee_fixed_cny: string;
  daily_withdrawal_limit_cny: string;
  monthly_withdrawal_limit_cny: string;
  cny_per_platform_usd: string;
  usd_to_cny: string;
  promotion_tracking_enabled: boolean;
  promotion_attribution_enabled: boolean;
  promotion_attribution_days: number;
  promotion_attribution_model: 'first_touch' | 'last_touch';
  promotion_collect_source: boolean;
  promotion_collect_device: boolean;
  promotion_bot_filter_enabled: boolean;
  promotion_detail_retention_days: number;
  registration_reward_enabled: boolean;
  recharge_reward_enabled: boolean;
}
export interface DistributionRewardRule {
  agent_id: number;
  registration_enabled: boolean;
  registration_reward_cny: string;
  recharge_enabled: boolean;
  recharge_threshold_cny: string;
  recharge_reward_cny: string;
}

export interface DistributionAdminOverview {
  total_agents: number;
  active_agents: number;
  suspended_agents: number;
  new_agents_this_month: number;
  total_customers: number;
  paying_customers: number;
  new_customers_this_month: number;
  customer_paid_cny: string;
  total_commission_cny: string;
  month_commission_cny: string;
  reversed_commission_cny: string;
  available_commission_cny: string;
  frozen_commission_cny: string;
  reserved_commission_cny: string;
  debt_commission_cny: string;
  pending_withdrawals: number;
  pending_withdrawal_cny: string;
  paying_withdrawals: number;
  paying_withdrawal_cny: string;
  paid_this_month_cny: string;
  overdue_pending_count: number;
  period_registration_reward_cny: string;
  period_recharge_reward_cny: string;
  analytics: DistributionBusinessAnalytics;
  agent_ranking: DistributionAgentRanking[];
}

export interface DistributionAdminAnomaly {
  id: string;
  type:
    | "overdue_withdrawal"
    | "pending_fx"
    | "agent_debt"
    | "inactive_agent_customers"
    | "matured_commission";
  severity: "critical" | "high" | "medium";
  entity_type:
    "withdrawal" | "commission_source" | "commission" | "agent" | "customer";
  entity_id: number;
  reference: string;
  agent_id?: number;
  agent_email?: string;
  subject: string;
  amount_cny: string;
  detected_at: string;
  description: string;
}

export interface DistributionMaturityStatus {
  running: boolean;
  last_outcome: "success" | "standby" | "error" | "disabled" | "";
  last_attempt_at?: string;
  last_success_at?: string;
  last_released: number;
  last_error?: string;
}

export async function getOverview(params: number | DistributionListParams = 30) {
  const query = typeof params === 'number' ? { days: params } : params;
  const { data } = await apiClient.get<DistributionAdminOverview>(
    "/admin/distribution/overview", { params: query },
  );
  return data;
}

export async function getAgentAnalytics(agentId: number, params: number | DistributionListParams = 30) {
  const query = typeof params === 'number' ? { days: params } : params;
  const { data } = await apiClient.get<DistributionAgentAnalytics>(
    `/admin/distribution/agents/${agentId}/analytics`, { params: query },
  );
  return data;
}

export async function getMaturityStatus() {
  const { data } = await apiClient.get<DistributionMaturityStatus>(
    "/admin/distribution/maturity-status",
  );
  return data;
}

export async function getSettings() {
  const { data } = await apiClient.get<DistributionSettings>(
    "/admin/distribution/settings",
  );
  return data;
}
export async function updateSettings(payload: DistributionSettings) {
  await apiClient.put("/admin/distribution/settings", payload);
}
export async function setFXRate(currency: string, rate: string) {
  await apiClient.put("/admin/distribution/fx-rate", {
    currency,
    rate_to_cny: rate,
  });
}
export async function grantAgent(payload: {
  user_id: number;
  depth?: number;
  parent_agent_id?: number;
  rate_override_bps?: number;
  promotion_code?: string;
  upgrade_customer?: boolean;
}) {
  const { data } = await apiClient.post<DistributionAgent>(
    "/admin/distribution/agents",
    payload,
  );
  return data;
}
export async function getAgentRewardRule(agentId: number) {
  const { data } = await apiClient.get<DistributionRewardRule>(`/admin/distribution/agents/${agentId}/reward-rule`);
  return data;
}
export async function updateAgentRewardRule(agentId: number, payload: Omit<DistributionRewardRule, "agent_id">) {
  await apiClient.put(`/admin/distribution/agents/${agentId}/reward-rule`, payload);
}
export interface DistributionWithdrawal {
  id: number;
  request_no: string;
  agent_id: number;
  agent_user_id: number;
  agent_email: string;
  agent_username: string;
  alipay_name: string;
  alipay_account: string;
  amount_cny: string;
  fee_cny: string;
  payout_cny: string;
  status: "pending" | "approved" | "paying" | "paid" | "rejected" | "failed";
  review_note?: string;
  reviewed_by?: number;
  reviewer_email?: string;
  reviewed_at?: string;
  approved_by?: number;
  approver_email?: string;
  approved_at?: string;
  paid_at?: string;
  payment_reference?: string;
  attachment_count: number;
  created_at: string;
}

export interface DistributionWithdrawalAttachment {
  id: number;
  withdrawal_id: number;
  original_name: string;
  content_type: string;
  size_bytes: number;
  sha256: string;
  evidence_type: "payment_receipt" | "bank_statement" | "other";
  note: string;
  uploaded_by?: number;
  created_at: string;
}

export interface DistributionWithdrawalEvent {
  id: number;
  withdrawal_id: number;
  from_status?: string;
  to_status: string;
  note: string;
  payment_reference?: string;
  actor_user_id?: number;
  actor_email?: string;
  created_at: string;
}

export interface DistributionWithdrawalDetail {
  withdrawal: DistributionWithdrawal;
  attachments: DistributionWithdrawalAttachment[];
  events: DistributionWithdrawalEvent[];
}

export interface DistributionEvidenceCapabilities {
  available: boolean;
  max_file_bytes: number;
  max_files_per_upload: number;
  max_total_bytes: number;
  allowed_extensions: string[];
}

export interface DistributionWithdrawalAction {
  status: DistributionWithdrawal["status"];
  note?: string;
  payment_reference?: string;
}
export async function updateLevel(
  depth: number,
  default_rate_bps: number,
  max_child_rate_bps: number,
  active = true,
) {
  await apiClient.put(`/admin/distribution/levels/${depth}`, {
    default_rate_bps,
    max_child_rate_bps,
    active,
  });
}
export async function correctCustomerBinding(
  user_id: number,
  agent_id: number,
  reason: string,
) {
  await apiClient.put("/admin/distribution/customer-binding", {
    user_id,
    agent_id,
    reason,
  });
}

export interface DistributionListParams {
  page?: number;
  page_size?: number;
  search?: string;
  status?: string;
  depth?: number;
  agent_id?: number;
  entry_type?: string;
  payment_type?: string;
  date_from?: string;
  date_to?: string;
  sort_by?: string;
  sort_order?: "asc" | "desc";
  days?: number;
}

export interface DistributionAnomalyListParams {
  page?: number;
  page_size?: number;
  search?: string;
  type?: string;
  severity?: string;
  sort_order?: "asc" | "desc";
}

export interface DistributionUserOption {
  user_id: number;
  email: string;
  username: string;
  status: string;
  selectable: boolean;
  unavailable_reason?: string;
}

export interface DistributionAgentOption {
  agent_id: number;
  user_id: number;
  email: string;
  username: string;
  promotion_code: string;
  depth: 1 | 2;
  status: string;
}

export interface DistributionAgentEvent {
  id: number;
  agent_id: number;
  event_type:
    "created" | "status_changed" | "rate_changed" | "permission_changed";
  old_status?: string;
  new_status?: string;
  old_rate_override_bps?: number;
  new_rate_override_bps?: number;
  old_effective_rate_bps?: number;
  new_effective_rate_bps?: number;
  reason: string;
  actor_user_id?: number;
  actor_email?: string;
  created_at: string;
}

export interface DistributionBindingEvent {
  id: number;
  customer_user_id: number;
  old_agent_id?: number;
  new_agent_id: number;
  old_agent_email?: string;
  new_agent_email: string;
  old_promotion_code?: string;
  new_promotion_code: string;
  reason: string;
  actor_user_id?: number;
  actor_email?: string;
  created_at: string;
}

export async function listAgents(params: DistributionListParams = {}) {
  const { data } = await apiClient.get<PaginatedResponse<DistributionAgent>>(
    "/admin/distribution/agents",
    { params },
  );
  return data;
}

export async function listCustomers(params: DistributionListParams = {}) {
  const { data } = await apiClient.get<PaginatedResponse<DistributionCustomer>>(
    "/admin/distribution/customers",
    { params },
  );
  return data;
}

export async function lookupAgentCandidates(query: string) {
  const { data } = await apiClient.get<DistributionUserOption[]>(
    "/admin/distribution/agent-candidates",
    { params: { q: query } },
  );
  return data;
}

export async function lookupAgents(query: string, options: { include_inactive?: boolean } = {}) {
  const { data } = await apiClient.get<DistributionAgentOption[]>(
    "/admin/distribution/agents/lookup",
    { params: { q: query, include_inactive: options.include_inactive || undefined } },
  );
  return data;
}

export async function listAgentEvents(agentId: number) {
  const { data } = await apiClient.get<DistributionAgentEvent[]>(
    `/admin/distribution/agents/${agentId}/events`,
  );
  return data;
}

export async function listBindingEvents(customerUserId: number) {
  const { data } = await apiClient.get<DistributionBindingEvent[]>(
    `/admin/distribution/customers/${customerUserId}/events`,
  );
  return data;
}

export async function updateAgentStatus(
  id: number,
  status: string,
  reason = "",
) {
  await apiClient.put(`/admin/distribution/agents/${id}/status`, {
    status,
    reason,
  });
}
export async function updateAgentRate(
  id: number,
  payload: { use_default: boolean; rate_override_bps?: number; reason: string },
) {
  await apiClient.put(`/admin/distribution/agents/${id}/rate`, payload);
}

export async function updateAgentRecruitmentPermission(
  id: number,
  payload: { enabled: boolean; reason: string },
) {
  await apiClient.put(
    `/admin/distribution/agents/${id}/recruitment-permission`,
    payload,
  );
}

export async function updateAgentPromotionStatsPermission(
  id: number,
  payload: { enabled: boolean; reason: string },
) {
  await apiClient.put(`/admin/distribution/agents/${id}/promotion-stats-permission`, payload);
}

export interface DistributionPromotionListParams {
  page?: number;
  page_size?: number;
  agent_id?: number;
  date_from?: string;
  date_to?: string;
  source?: string;
  device?: string;
  attribution_type?: string;
}

export type DistributionPromotionAnalytics = import('../distribution').DistributionPromotionAnalytics;
export type DistributionPromotionVisit = import('../distribution').DistributionPromotionVisit;

export async function getPromotionAnalytics(params: DistributionPromotionListParams = {}) {
  const { data } = await apiClient.get<DistributionPromotionAnalytics>('/admin/distribution/promotion/analytics', { params });
  return data;
}

export async function listPromotionVisits(params: DistributionPromotionListParams = {}) {
  const { data } = await apiClient.get<PaginatedResponse<DistributionPromotionVisit>>('/admin/distribution/promotion/visits', { params });
  return data;
}

export async function listCommissions(params: DistributionListParams = {}) {
  const { data } = await apiClient.get<
    PaginatedResponse<DistributionCommission>
  >("/admin/distribution/commissions", { params });
  return data;
}

export async function listWithdrawals(params: DistributionListParams = {}) {
  const { data } = await apiClient.get<
    PaginatedResponse<DistributionWithdrawal>
  >("/admin/distribution/withdrawals", { params });
  return data;
}

export async function listAnomalies(
  params: DistributionAnomalyListParams = {},
) {
  const { data } = await apiClient.get<
    PaginatedResponse<DistributionAdminAnomaly>
  >("/admin/distribution/anomalies", { params });
  return data;
}

export async function getWithdrawal(id: number) {
  const { data } = await apiClient.get<DistributionWithdrawalDetail>(
    `/admin/distribution/withdrawals/${id}`,
  );
  return data;
}

export async function reviewWithdrawal(
  id: number,
  payload: DistributionWithdrawalAction,
) {
  await apiClient.put(`/admin/distribution/withdrawals/${id}/status`, payload);
}

export async function batchReviewWithdrawals(payload: {
  withdrawal_ids: number[];
  status: "approved" | "rejected";
  note?: string;
}) {
  const { data } = await apiClient.post<{ updated: number }>(
    "/admin/distribution/withdrawals/batch-review",
    payload,
  );
  return data;
}

async function downloadDistributionExport(
  path: string,
  params: DistributionListParams = {},
) {
  const response = await apiClient.get<Blob>(path, {
    params,
    responseType: "blob",
  });
  return response.data;
}

export const exportAgents = (params: DistributionListParams = {}) =>
  downloadDistributionExport("/admin/distribution/agents/export", params);
export const exportCustomers = (params: DistributionListParams = {}) =>
  downloadDistributionExport("/admin/distribution/customers/export", params);
export const exportCommissions = (params: DistributionListParams = {}) =>
  downloadDistributionExport("/admin/distribution/commissions/export", params);
export const exportWithdrawals = (params: DistributionListParams = {}) =>
  downloadDistributionExport("/admin/distribution/withdrawals/export", params);

export async function getWithdrawalEvidenceCapabilities() {
  const { data } = await apiClient.get<DistributionEvidenceCapabilities>(
    "/admin/distribution/withdrawals/evidence-capabilities",
  );
  return data;
}

export async function uploadWithdrawalEvidence(
  id: number,
  files: File[],
  evidenceType: string,
  note: string,
) {
  const form = new FormData();
  files.forEach((file) => form.append("files", file));
  form.append("evidence_type", evidenceType);
  form.append("note", note);
  const { data } = await apiClient.post<DistributionWithdrawalAttachment[]>(
    `/admin/distribution/withdrawals/${id}/attachments`,
    form,
    {
      headers: { "Content-Type": undefined },
    },
  );
  return data;
}

export async function getWithdrawalEvidenceURL(
  withdrawalId: number,
  attachmentId: number,
) {
  const { data } = await apiClient.get<{ url: string }>(
    `/admin/distribution/withdrawals/${withdrawalId}/attachments/${attachmentId}/url`,
  );
  return data.url;
}
