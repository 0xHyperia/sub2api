import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const source = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../ModelMarketplaceDetailDrawer.vue'),
  'utf8',
)

describe('ModelMarketplaceDetailDrawer workspace', () => {
  it('uses a responsive right drawer with a three-tab detail workspace', () => {
    expect(source).toContain('class="fixed inset-0 z-50 bg-black/25 backdrop-blur-[1px]"')
    expect(source).toContain('absolute inset-y-0 right-0')
    expect(source).toContain('w-full max-w-5xl')
    expect(source).toContain('@click.self="emit(\'close\')"')
    expect(source).toContain("type DetailTab = 'overview' | 'performance' | 'api'")
    expect(source).toContain("activeTab === 'overview'")
    expect(source).toContain("activeTab === 'performance'")
    expect(source).toContain('ModelMarketplacePerformanceCharts')
    expect(source).toContain('translateX(100%)')
  })

  it('provides pricing, API examples, parameters, and real RPM limits', () => {
    expect(source).toContain('groupPricingRows')
    expect(source).toContain('groupPerformanceRows')
    expect(source).toContain("type CodeLanguage = 'curl' | 'python' | 'typescript' | 'javascript'")
    expect(source).toContain("activeProtocol.value === 'anthropic'")
    expect(source).toContain("activeProtocol.value === 'gemini'")
    expect(source).toContain("`${apiBaseUrl.value}/v1/chat/completions`")
    expect(source).toContain('apiParameters')
    expect(source).toContain('group.rpm_limit')
    expect(source).not.toContain('TPS')
    expect(source).not.toContain('TTFT')
  })
})
