/**
 * User Channels API endpoints (non-admin)
 * 用户侧「可用渠道」聚合查询：渠道 + 用户可访问的分组 + 支持模型（含定价）。
 */

import { apiClient } from './client'
import type { BillingMode } from '@/constants/channel'

export interface UserAvailableGroup {
  id: number
  name: string
  platform: string
  /** 'standard' | 'subscription' — 订阅分组视觉加深，和 API 密钥页保持一致。 */
  subscription_type: string
  /** 分组默认倍率。用户专属倍率（若有）通过 /groups/rates 获取后在前端 join。 */
  rate_multiplier: number
  peak_rate_enabled: boolean
  peak_start: string
  peak_end: string
  peak_rate_multiplier: number
  /** true = 专属分组（小范围授权）；false = 公开分组。 */
  is_exclusive: boolean
  /** 分组级每分钟请求上限；0 表示该层不限制。 */
  rpm_limit?: number
}

export interface UserMarketplaceGroup extends UserAvailableGroup {
  /** 图片生成使用独立倍率时，不再叠加用户专属或普通分组倍率。 */
  image_rate_independent: boolean
  image_rate_multiplier: number
  /** USD / image；null 表示该尺寸使用模型默认价格。 */
  image_price_1k: number | null
  image_price_2k: number | null
  image_price_4k: number | null
  pricing?: UserSupportedModelPricing | null
}

export interface UserPricingInterval {
  min_tokens: number
  max_tokens: number | null
  tier_label?: string
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_write_1h_price?: number | null
  cache_read_price: number | null
  input_multiplier?: number | null
  output_multiplier?: number | null
  cache_write_multiplier?: number | null
  cache_read_multiplier?: number | null
  per_request_price: number | null
}

export interface UserSupportedModelPricing {
  time_pricing?: {
    timezone: string
    weekdays_only?: boolean
    periods: { start_time: string; end_time: string; multiplier: number }[]
  } | null
  billing_mode: BillingMode
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_write_1h_price?: number | null
  cache_read_price: number | null
  image_input_price: number | null
  image_output_price: number | null
  per_request_price: number | null
  intervals: UserPricingInterval[]
}

export interface UserSupportedModel {
  name: string
  platform: string
  pricing: UserSupportedModelPricing | null
}

export interface UserMarketplaceModel extends UserSupportedModel {
  groups: UserMarketplaceGroup[]
  capabilities?: string[]
  monitor_status?: UserModelMonitorSummary | null
}

export interface UserModelMonitorTimelinePoint {
  status: 'operational' | 'degraded' | 'failed' | 'error'
  latency_ms: number | null
  checked_at: string
  group_id?: number | null
  group_name?: string
}

export interface UserModelMonitorSummary {
  status: UserModelMonitorTimelinePoint['status'] | ''
  latency_ms: number | null
  availability_7d: number | null
  last_checked_at: string | null
  timeline: UserModelMonitorTimelinePoint[]
  display_order: number
  label: string
  groups?: UserModelMonitorGroupMetrics[]
  metrics?: UserModelMonitorGroupMetrics['metrics'] | null
  hourly_metrics?: UserModelMonitorGroupMetrics['metrics'] | null
}

export interface UserModelMonitorMetricBucket {
  started_at: string
  success_rate: number | null
  ttft_ms: number | null
}

export interface UserModelMonitorGroupMetrics {
  group_id: number
  name: string
  metrics: {
    tps: number | null
    ttft_ms: number | null
    average_latency_ms: number | null
    success_rate: number | null
    buckets: UserModelMonitorMetricBucket[]
  }
}

export interface UserMarketplacePlatform {
  platform: string
  groups: UserMarketplaceGroup[]
  supported_models: UserMarketplaceModel[]
}

/**
 * 渠道下单个平台的子视图：用户可访问的分组 + 该平台支持的模型。
 * 后端把一个渠道按平台聚合成 sections，前端可以把渠道名作为 row-group
 * 一次渲染，后面按 sections 顺序用 rowspan 铺开。
 */
export interface UserChannelPlatformSection {
  platform: string
  groups: UserAvailableGroup[]
  supported_models: UserSupportedModel[]
}

export interface UserAvailableChannel {
  name: string
  description: string
  platforms: UserChannelPlatformSection[]
}

/** 列出当前用户可见的「可用渠道」（与 /groups/available 保持一致，返回平数组）。 */
export async function getAvailable(options?: { signal?: AbortSignal }): Promise<UserAvailableChannel[]> {
  const { data } = await apiClient.get<UserAvailableChannel[]>('/channels/available', {
    signal: options?.signal
  })
  return data
}

/** List the current user's model marketplace catalog. */
export async function getMarketplace(options?: { signal?: AbortSignal; resolution?: 'minute' | 'hour' }): Promise<UserMarketplacePlatform[]> {
  const { data } = await apiClient.get<UserMarketplacePlatform[]>('/models/marketplace', {
    signal: options?.signal,
    params: { resolution: options?.resolution ?? 'hour' },
  })
  return data
}

export const userChannelsAPI = { getAvailable, getMarketplace }

export default userChannelsAPI
