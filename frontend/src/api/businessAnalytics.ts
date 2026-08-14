import { apiClient } from './client'

export type BusinessAnalyticsSection = 'overview' | 'growth' | 'finance' | 'retention' | 'channels'
export type BusinessChannel = 'distribution' | 'affiliate' | 'campaign' | 'organic' | 'unknown'

export interface BusinessMetric { value: number; previous: number; change_rate?: number; change_value?: number; comparison_type?: 'relative' | 'percentage_point' | 'turned_positive' | 'turned_negative' | 'unavailable'; currency?: string; estimated?: boolean }
export interface BusinessTrendPoint { bucket: string; new_users: number; activated_users: number; first_paid_users: number; paying_users: number; repurchase_users: number; active_users: number; gross_paid_cny: number; refunded_cny: number; net_paid_cny: number; consumed_revenue: number; supplier_cost: number }
export interface BusinessChannelMetric { channel: BusinessChannel; new_users: number; activated_users: number; first_paid_users: number; paying_users: number; repurchase_users: number; active_users: number; net_paid_cny: number; arppu_cny: number; d7_retention_rate: number }
export interface BusinessCohortRow { cohort_date: string; users: number; d1?: number; d7?: number; d30?: number }
export interface BusinessLifecycle { new: number; unactivated: number; newly_activated: number; continuously_active: number; silent_reactivated: number; churned_reactivated: number; silent: number; churned: number }
export interface BusinessBalanceSegment { key: 'active' | 'silent' | 'churned' | 'unactivated'; users: number; balance_usd: number; frozen_usd: number }
export interface BusinessBalanceSnapshot {
  as_of: string; activity_window_days: 1 | 7 | 30; balance_recharge_multiplier: number
  available_balance_usd: number; frozen_balance_usd: number; positive_balance_users: number; total_users: number; average_balance_usd: number
  low_balance_enabled: boolean; low_balance_users: number; segments: BusinessBalanceSegment[]
}
export interface BusinessDurationBucket { key: 'under_1h' | '1h_24h' | '1d_3d' | '3d_7d' | 'over_7d'; activation: number; first_paid: number }
export interface BusinessCurrencyBreakdown { currency: string; gross: number; refunded: number; net: number }
export interface BusinessFunnelStep { key: string; count: number }
export interface BusinessAnalyticsSnapshot {
  date_from: string; date_to: string; previous_date_from: string; previous_date_to: string; timezone: string; granularity: 'hour' | 'day' | 'week' | 'month'; currency: 'CNY'; updated_at: string; usage_data_from?: string; estimated: boolean
  metrics: Record<string, BusinessMetric>; trend: BusinessTrendPoint[]; previous_trend: BusinessTrendPoint[]; funnel: BusinessFunnelStep[]; channels: BusinessChannelMetric[]
  registration_cohorts: BusinessCohortRow[]; activation_cohorts: BusinessCohortRow[]; lifecycle: BusinessLifecycle
  duration_distribution: BusinessDurationBucket[]; currency_breakdown: BusinessCurrencyBreakdown[]; warnings?: string[] | null
}

export async function getBusinessAnalytics(section: BusinessAnalyticsSection, params: Record<string, string | number | undefined>) {
  const { data } = await apiClient.get<BusinessAnalyticsSnapshot>(`/admin/business-analytics/${section}`, { params })
  return data
}

export async function getBusinessBalance(activityWindow: 1 | 7 | 30, agentId?: number, agentScope?: 'team' | 'direct') {
  const { data } = await apiClient.get<BusinessBalanceSnapshot>('/admin/business-analytics/balance', { params: { activity_window: activityWindow, agent_id: agentId, agent_scope: agentScope } })
  return data
}
