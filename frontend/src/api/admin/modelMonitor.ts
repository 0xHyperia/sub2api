import { apiClient } from '../client'

export type ModelMonitorStatus = 'operational' | 'degraded' | 'failed' | 'error'

export interface ModelMonitorTimelinePoint {
  status: ModelMonitorStatus
  latency_ms: number | null
  checked_at: string
}

export interface ModelMonitorSummary {
  status: ModelMonitorStatus | ''
  latency_ms: number | null
  availability_7d: number | null
  last_checked_at: string | null
  timeline: ModelMonitorTimelinePoint[]
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
}

export interface ModelMonitorHistoryItem extends ModelMonitorTimelinePoint {
  id: number
  monitor_id: number
  attempts: number
  message: string
  group_id: number | null
  group_name: string
}

export async function updateGroups(payload: { platform: string; model: string; group_ids: number[] }): Promise<ModelMonitorRow> {
  const { data } = await apiClient.put('/admin/model-monitors/groups', payload)
  return data
}

export async function list(): Promise<{ items: ModelMonitorRow[] }> {
  const { data } = await apiClient.get('/admin/model-monitors')
  return data
}

export async function updateConfig(payload: { platform: string; model: string; enabled: boolean; interval_seconds: number; display_order: number; label: string }): Promise<ModelMonitorRow> {
  const { data } = await apiClient.put('/admin/model-monitors/config', payload)
  return data
}

export async function runNow(platform: string, model: string): Promise<ModelMonitorHistoryItem> {
  const { data } = await apiClient.post('/admin/model-monitors/run', { platform, model })
  return data
}

export async function history(id: number, limit = 100): Promise<{ items: ModelMonitorHistoryItem[] }> {
  const { data } = await apiClient.get(`/admin/model-monitors/${id}/history`, { params: { limit } })
  return data
}

export default { list, updateConfig, updateGroups, runNow, history }
