import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import { describe, expect, it } from 'vitest'

import ModelMarketplaceDetailDrawer from '../ModelMarketplaceDetailDrawer.vue'
import type { MarketplaceModelEntry } from '@/views/user/modelMarketplace'

const source = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../ModelMarketplaceDetailDrawer.vue'),
  'utf8',
)

describe('ModelMarketplaceDetailDrawer workspace', () => {

  function entry(): MarketplaceModelEntry {
    return {
      key: 'openai:gpt-5', name: 'gpt-5', platform: 'openai', groups: [], pricing: null,
      capabilities: [], monitorStatus: null, displayOrder: 0, label: '',
    }
  }

  it('keeps the performance tab active while the parent reloads the same model resolution', async () => {
    const wrapper = mount(ModelMarketplaceDetailDrawer, {
      attachTo: document.body,
      props: {
        entry: entry(), groups: [], activeGroup: null, showEffectivePrices: false,
        monitorResolution: 'hour', showDetailedPerformance: false,
      },
      global: {
        plugins: [createPinia(), createI18n({
          legacy: false,
          locale: 'en',
          missingWarn: false,
          fallbackWarn: false,
          messages: { en: {} },
        })],
        stubs: { teleport: true, PlatformIcon: true, Icon: true, SuccessRateTimeline: true, ModelMarketplacePerformanceCharts: true },
      },
    })

    const performanceTab = wrapper.get('[data-testid="detail-tab-performance"]')
    await performanceTab.trigger('click')
    expect(performanceTab.attributes('aria-current')).toBe('page')

    const minuteButton = wrapper.get('[data-testid="resolution-minute"]')
    await minuteButton.trigger('click')
    expect(wrapper.emitted('update:monitorResolution')?.at(-1)).toEqual(['minute'])
    await wrapper.setProps({ entry: { ...entry(), monitorStatus: { status: '', latency_ms: null, availability_7d: null, last_checked_at: null, timeline: [], display_order: 0, label: '' } }, monitorResolution: 'minute' })
    expect(performanceTab.attributes('aria-current')).toBe('page')

    wrapper.unmount()
  })
  it('uses a responsive right drawer with a three-tab detail workspace', () => {
    expect(source).toContain('class="fixed inset-0 z-50 bg-black/25 backdrop-blur-[1px]"')
    expect(source).toContain('absolute inset-y-0 right-0')
    expect(source).toContain('w-full max-w-5xl')
    expect(source).toContain('@click.self="emit(\'close\')"')
    expect(source).toContain("type DetailTab = 'overview' | 'performance' | 'api'")
    expect(source).toContain("activeTab === 'overview'")
    expect(source).toContain("activeTab === 'performance'")
    expect(source).toContain('SuccessRateTimeline')
    expect(source).toContain('entry.monitorStatus?.metrics?.buckets')
    expect(source).toContain('variant="availability"')
    expect(source).toContain('ModelMarketplacePerformanceCharts')
    expect(source).toContain('space-y-10 transition-opacity')
    expect(source).not.toContain('class="contents transition-opacity')
    expect(source).toContain('table-fixed')
    expect(source).toContain('<colgroup>')
    expect(source).toContain("monitorResolution: 'minute' | 'hour'")
    expect(source).toContain("'update:monitorResolution'")
    expect(source).toContain('entry.key === previousEntry.key')
    expect(source).toContain('translateX(100%)')
  })

  it('provides pricing, API examples, parameters, and real RPM limits', () => {
    expect(source).toContain('groupPricingRows')
    expect(source).toContain("pricing.image_input_price")
    expect(source).toContain("modelMarketplace.price.imageInput")
    expect(source).toContain('groupPerformanceRows')
    expect(source).toContain("type CodeLanguage = 'curl' | 'python' | 'typescript' | 'javascript'")
    expect(source).toContain("activeProtocol.value === 'anthropic'")
    expect(source).toContain("activeProtocol.value === 'gemini'")
    expect(source).toContain("`${apiBaseUrl.value}/v1/chat/completions`")
    expect(source).toContain('apiParameters')
    expect(source).toContain('group.rpm_limit')
    expect(source).toContain('TPS')
    expect(source).toContain('TTFT')
    expect(source).toContain('<th v-if="showDetailedPerformance" class="px-4 py-3 text-right font-medium">TTFT</th>')
    expect(source).toContain("showDetailedPerformance ? '' : 'text-right'")
    expect(source).toContain("showDetailedPerformance ? '' : 'flex justify-end'")
    expect(source).not.toContain('row.samples')
    expect(source).not.toContain('probe_cost')
  })
})
