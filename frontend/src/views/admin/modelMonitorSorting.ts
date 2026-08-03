import type { ModelMonitorRow } from '@/api/admin/modelMonitor'

export type ModelMonitorMetricSortField = 'tps' | 'ttft' | 'latency' | 'successRate'
export type ModelMonitorSortDirection = 'asc' | 'desc'

function metricValue(row: ModelMonitorRow, field: ModelMonitorMetricSortField): number | null | undefined {
  const metrics = row.summary?.metrics
  if (field === 'tps') return metrics?.tps
  if (field === 'ttft') return metrics?.ttft_ms
  if (field === 'latency') return metrics?.average_latency_ms
  return metrics?.success_rate
}

export function compareModelMonitorRows(
  leftRow: ModelMonitorRow,
  rightRow: ModelMonitorRow,
  field: ModelMonitorMetricSortField | null,
  direction: ModelMonitorSortDirection,
): number {
  if (!field) return rightRow.display_order - leftRow.display_order || leftRow.model.localeCompare(rightRow.model)
  const left = metricValue(leftRow, field)
  const right = metricValue(rightRow, field)
  if (left == null && right == null) return leftRow.model.localeCompare(rightRow.model)
  if (left == null) return 1
  if (right == null) return -1
  const result = left - right
  return (direction === 'asc' ? result : -result) || leftRow.model.localeCompare(rightRow.model)
}
