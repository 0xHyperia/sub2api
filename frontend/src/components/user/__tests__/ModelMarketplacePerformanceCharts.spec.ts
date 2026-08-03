import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const source = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../ModelMarketplacePerformanceCharts.vue'),
  'utf8',
)

describe('ModelMarketplacePerformanceCharts interactions', () => {
  it('divides the plot into centered hover bands based on real data points', () => {
    expect(source).toContain('const bandWidth = (chartArea.right - chartArea.left) / pointCount')
    expect(source).toContain('const left = chartArea.left + (activeIndex * bandWidth)')
    expect(source).toContain('ctx.fillRect(left, chartArea.top, bandWidth')
    expect(source).toContain("axis: 'x' as const")
    expect(source).toContain('offset: true')
    expect(source).not.toContain('ctx.fillRect(x - 16')
  })

  it('builds both trends only from buckets containing their metric', () => {
    expect(source).toContain('filter(bucket => bucket.ttft_ms != null)')
    expect(source).toContain('filter(bucket => bucket.success_rate != null)')
  })
})
