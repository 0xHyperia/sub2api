import { apiClient } from './client'
import type { UserModelMonitorSummary, UserSupportedModelPricing } from './channels'

export interface HomeMetrics {
  today_tokens: number
  stats_updated_at?: string
  stats_stale?: boolean
  generated_at?: string
}

export async function getHomeMetrics(): Promise<HomeMetrics> {
  const { data } = await apiClient.get<HomeMetrics>('/home/metrics', { timeout: 4500 })
  return data
}

export interface HomeShowcaseModel {
  name: string
  platform: string
  pricing: UserSupportedModelPricing | null
  rate_multiplier: number
  monitor_status: UserModelMonitorSummary | null
}

export interface HomeShowcasePlatform {
  platform: string
  model_count: number
  models: HomeShowcaseModel[]
}

export async function getHomeShowcase(): Promise<HomeShowcasePlatform[]> {
  const { data } = await apiClient.get<HomeShowcasePlatform[]>('/models/showcase', { timeout: 4500 })
  return data
}

export const homeAPI = {
  getHomeMetrics,
  getHomeShowcase
}

export default homeAPI
