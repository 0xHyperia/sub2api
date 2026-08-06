import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const source = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../ModelMonitorView.vue'),
  'utf8',
)

describe('model monitor presentation controls', () => {
  it('supports label and display priority configuration', () => {
    expect(source).toContain('openPresentation(row)')
    expect(source).toContain('presentationLabel')
    expect(source).toContain('presentationOrder')
    expect(source).toContain('display_order:')
  })

  it('silently refreshes visible monitor data every 30 seconds', () => {
    expect(source).toContain('window.setInterval')
    expect(source).toContain('30_000')
    expect(source).toContain("document.visibilityState === 'visible'")
    expect(source).toContain('void load(true)')
    expect(source).toContain('window.clearInterval')
  })

  it('shows expandable group metrics and independent probe controls', () => {
    expect(source).toContain('SuccessRateTimeline')
    expect(source).toContain('group.metrics?.tps')
    expect(source).toContain('group.metrics?.ttft_ms')
    expect(source).toContain('group.metrics?.average_latency_ms')
    expect(source).toContain('row.summary?.metrics?.success_rate')
    expect(source).toContain('row.summary?.metrics?.success_count')
    expect(source).toContain('row.summary?.metrics?.failure_count')
    expect(source).toContain("t('admin.modelMonitor.requestResults')")
    expect(source).toContain('toggleAllModels')
    expect(source).toContain("toggleMetricSort('tps')")
    expect(source).toContain("toggleMetricSort('ttft')")
    expect(source).toContain("toggleMetricSort('latency')")
    expect(source).toContain("toggleMetricSort('successRate')")
    expect(source).toContain('filteredRows.value.map(modelKey)')
    expect(source).not.toContain('expansionInitialized')
    expect(source).toContain('overflow-y-auto overscroll-contain')
    expect(source).toContain('group.metrics?.success_count')
    expect(source).toContain('group.metrics?.failure_count')
    expect(source).toContain('xl:min-w-[1200px]')
    expect(source).toContain('w-full items-center justify-center gap-1 font-mono text-xs tabular-nums xl:mt-0 xl:flex')
    expect(source).toContain('xl:justify-self-end')
    expect(source).toContain('xl:justify-self-center')
    expect(source).toContain('xl:w-[112px]')
    expect(source).toContain('<PlatformIcon')
    expect(source).toContain("t('admin.modelMonitor.displayOrderHint')")
    expect(source).toContain('updateGroupConfig')
    expect(source).toContain('failure_compensation_enabled')
    expect(source).toContain('failure_compensation_pending')
    expect(source).toContain('showUnavailable')
    expect(source).toContain('!showUnavailable.value && !row.catalog_available')
    expect(source).toContain('intervalOptions = [60, 300, 600, 900, 1800, 3600]')
    expect(source).toContain('xl:hidden')
    expect(source).not.toContain('TablePageLayout')
    expect(source).toContain("localStorage.getItem('usa0:model-monitor-resolution') === 'minute' ? 'minute' : 'hour'")
  })
})
