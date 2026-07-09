import { apiClient } from './client'

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

export const homeAPI = {
  getHomeMetrics
}

export default homeAPI
