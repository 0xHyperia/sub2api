import { describe, expect, it } from 'vitest'

import type { ModelMonitorRow } from '@/api/admin/modelMonitor'
import { compareModelMonitorRows } from '../modelMonitorSorting'

function row(model: string, value: number | null, displayOrder = 0): ModelMonitorRow {
  return {
    id: 1,
    platform: 'openai',
    model,
    enabled: false,
    interval_seconds: 300,
    display_order: displayOrder,
    label: '',
    last_checked_at: null,
    configured: true,
    catalog_available: true,
    groups_configured: true,
    groups: [],
    summary: {
      status: '', latency_ms: null, availability_7d: null, last_checked_at: null, timeline: [],
      metrics: { tps: value, ttft_ms: value, average_latency_ms: value, success_rate: value, probe_cost: null, buckets: [] },
    },
  }
}

describe('compareModelMonitorRows', () => {
  it('sorts metrics in both directions and always keeps missing values last', () => {
    const rows = [row('missing', null), row('slow', 10), row('fast', 50)]
    expect([...rows].sort((a, b) => compareModelMonitorRows(a, b, 'tps', 'desc')).map(item => item.model)).toEqual(['fast', 'slow', 'missing'])
    expect([...rows].sort((a, b) => compareModelMonitorRows(a, b, 'tps', 'asc')).map(item => item.model)).toEqual(['slow', 'fast', 'missing'])
  })

  it('uses display priority and then model name when no metric sort is active', () => {
    const rows = [row('zeta', 10, 2), row('beta', 20, 5), row('alpha', 30, 2)]
    expect([...rows].sort((a, b) => compareModelMonitorRows(a, b, null, 'desc')).map(item => item.model)).toEqual(['beta', 'alpha', 'zeta'])
  })
})
