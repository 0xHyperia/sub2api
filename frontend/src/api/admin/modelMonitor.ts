import { apiClient } from '../client'

export type ModelMonitorStatus = 'operational' | 'degraded' | 'failed' | 'error'

export interface ModelMonitorTimelinePoint {
  status: ModelMonitorStatus
  latency_ms: number | null
  checked_at: string
}

export type ModelMonitorResolution = 'minute' | 'hour'

export interface ModelMonitorMetricBucket {
  started_at: string
  success_rate: number | null
}

export interface ModelMonitorGroupMetrics {
  tps: number | null
  ttft_ms: number | null
  average_latency_ms: number | null
  success_rate: number | null
  probe_cost: number | null
  buckets: ModelMonitorMetricBucket[]
}

export interface ModelMonitorSummary {
  status: ModelMonitorStatus | ''
  latency_ms: number | null
  availability_7d: number | null
  last_checked_at: string | null
  timeline: ModelMonitorTimelinePoint[]
  metrics?: ModelMonitorGroupMetrics | null
}

export interface ModelMonitorRow {
  id: number
  platform: string
  model: string
  enabled: boolean
  interval_seconds: number
  display_order: number
  label: string
  last_checked_at: string | null
  configured: boolean
  catalog_available: boolean
  summary: ModelMonitorSummary | null
  groups: ModelMonitorGroupOption[]
  groups_configured: boolean
}

export interface ModelMonitorGroupOption {
  group_id: number
  name: string
  rate_multiplier: number
  priority: number
  selected: boolean
  enabled: boolean
  interval_seconds: number
  last_traffic_at: string | null
  last_probe_at: string | null
  metrics: ModelMonitorGroupMetrics | null
}

export interface ModelMonitorHistoryItem extends ModelMonitorTimelinePoint {
  id: number
  monitor_id: number
  attempts: number
  message: string
  group_id: number | null
  group_name: string
  first_token_ms: number | null
  input_tokens: number
  output_tokens: number
  generation_ms: number
  probe_cost: number | null
}

export async function list(resolution: ModelMonitorResolution = 'minute'): Promise<{ items: ModelMonitorRow[] }> {
  const { data } = await apiClient.get('/admin/model-monitors', { params: { resolution } })
  return data
}

export async function updateConfig(payload: { platform: string; model: string; enabled: boolean; interval_seconds: number; display_order: number; label: string }): Promise<ModelMonitorRow> {
  const { data } = await apiClient.put('/admin/model-monitors/config', payload)
  return data
}

export async function updateGroupConfig(payload: { platform: string; model: string; group_id: number; enabled: boolean; interval_seconds: number }): Promise<ModelMonitorRow> {
  const { data } = await apiClient.put('/admin/model-monitors/group', payload)
  return data
}

export async function runNow(platform: string, model: string, groupId?: number): Promise<ModelMonitorHistoryItem> {
  const { data } = await apiClient.post('/admin/model-monitors/run', { platform, model, group_id: groupId })
  return data
}

export async function history(id: number, limit = 100): Promise<{ items: ModelMonitorHistoryItem[] }> {
  const { data } = await apiClient.get(`/admin/model-monitors/${id}/history`, { params: { limit } })
  return data
}

export default { list, updateConfig, updateGroupConfig, runNow, history }
