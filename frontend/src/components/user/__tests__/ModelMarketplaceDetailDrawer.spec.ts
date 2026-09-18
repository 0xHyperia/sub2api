import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import { describe, expect, it } from 'vitest'

import ModelMarketplaceDetailDrawer from '../ModelMarketplaceDetailDrawer.vue'
import ModelMarketplacePricingTable from '../ModelMarketplacePricingTable.vue'
import type { MarketplaceGroupOption, MarketplaceModelEntry } from '@/views/user/modelMarketplace'
import type { UserSupportedModelPricing } from '@/api/channels'

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
        entry: entry(), groups: [],
        showRechargePrices: false, balanceRechargeMultiplier: 1, officialUsdToCnyRate: 7.2,
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

  it('switches official detail pricing from USD to the currency-converted CNY price', async () => {
    const pricedEntry: MarketplaceModelEntry = {
      ...entry(),
      pricing: {
        billing_mode: 'token',
        input_price: 0.000002,
        output_price: null,
        cache_write_price: null,
        cache_read_price: null,
        image_input_price: null,
        image_output_price: null,
        per_request_price: null,
        intervals: [],
      },
    }
    const wrapper = mount(ModelMarketplaceDetailDrawer, {
      props: {
        entry: pricedEntry, groups: [],
        showRechargePrices: false, balanceRechargeMultiplier: 0.2, officialUsdToCnyRate: 7.2,
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
        stubs: { teleport: true, ModelIcon: true, Icon: true, SuccessRateTimeline: true, ModelMarketplacePerformanceCharts: true },
      },
    })

    expect(wrapper.text()).toContain('$2.00')
    await wrapper.setProps({ showRechargePrices: true })
    expect(wrapper.text()).toContain('¥14.40')

    wrapper.unmount()
  })

  it('uses a responsive right drawer with a three-tab detail workspace', () => {
    expect(source).toContain('class="fixed inset-0 z-50 bg-black/25 backdrop-blur-[1px]"')
    expect(source).toContain('absolute inset-y-0 right-0')
    expect(source).toContain('w-full max-w-[672px]')
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

  it('shows each group schedule, hides unused prices and keeps free primary prices', async () => {
    const pricing: UserSupportedModelPricing = {
      billing_mode: 'token', input_price: 0, output_price: 0.000004,
      cache_read_price: 0.00000002, cache_write_price: 0,
      image_input_price: 0, image_output_price: null, per_request_price: null, intervals: [],
      time_pricing: { timezone: 'UTC', weekdays_only: true, periods: [
        { start_time: '01:00:00', end_time: '04:00:00', multiplier: 2 },
      ] },
    }
    const group: MarketplaceGroupOption = {
      id: 1, name: 'Group A', platform: 'openai', subscription_type: 'standard',
      rate_multiplier: 0.1, effectiveRate: 0.1, peak_rate_enabled: false,
      peak_start: '', peak_end: '', peak_rate_multiplier: 1, is_exclusive: false,
      image_rate_independent: false, image_rate_multiplier: 1,
      image_price_1k: null, image_price_2k: null, image_price_4k: null, pricing,
    }
    const otherGroup = { ...group, id: 2, name: 'Group B', pricing: { ...pricing, time_pricing: null, output_price: 0.000008 } }
    const wrapper = mount(ModelMarketplaceDetailDrawer, {
      props: {
        entry: { ...entry(), pricing, groups: [group, otherGroup] },
        groups: [group, otherGroup],
        showRechargePrices: false, balanceRechargeMultiplier: 1, officialUsdToCnyRate: 7.2,
        monitorResolution: 'hour', showDetailedPerformance: false,
      },
      global: {
        plugins: [createPinia(), createI18n({ legacy: false, locale: 'en', missingWarn: false, fallbackWarn: false, messages: { en: {} } })],
        stubs: { teleport: true, ModelIcon: true, Icon: true, SuccessRateTimeline: true, ModelMarketplacePerformanceCharts: true },
      },
    })
    const schedule = wrapper.get('[data-testid="time-pricing"]')
    expect(schedule.text()).toContain('UTC')
    expect(schedule.text()).toContain('weekdaysOnly')
    expect(schedule.text()).toContain('01:00:00 ~ 04:00:00')
    expect(schedule.text()).toContain('2×')
    expect(wrapper.text()).not.toContain('modelMarketplace.price.imageInput')
    expect(wrapper.text()).not.toContain('modelMarketplace.price.cacheWrite')
    expect(wrapper.text()).toContain('$0.00')
    expect(wrapper.get('[data-testid="group-pricing-1"]').text()).toContain('$0.40')
    expect(wrapper.get('[data-testid="group-pricing-2"]').text()).toContain('$0.80')
    expect(wrapper.get('[data-testid="group-pricing-2"]').text()).not.toContain('01:00:00')
    expect(wrapper.get('[data-testid="group-pricing-1"]').find('button').exists()).toBe(false)
    expect(wrapper.findAll('button').some(button => button.text() === group.name)).toBe(false)
    expect(wrapper.emitted('selectGroup')).toBeUndefined()
    await wrapper.setProps({ groups: [], entry: { ...entry(), pricing: { ...pricing, image_input_price: 0.000001 } } })
    expect(wrapper.text().split('modelMarketplace.price.imageInput')).toHaveLength(2)
    wrapper.unmount()
  })

  it('provides pricing, API examples, parameters, and real RPM limits', () => {
    expect(source).toContain('groupPricingRows')
    expect(source).toContain('ModelMarketplacePricingTable')
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
    expect(source).toContain('class="hidden overflow-x-auto rounded-panel border border-outline sm:block"')
    expect(source).toContain("showDetailedPerformance ? 'min-w-[720px]' : 'min-w-[420px]'")
    expect(source).toContain('<th class="px-4 py-3 text-right font-medium">{{ t(\'modelMarketplace.details.successRate\') }}</th>')
    expect(source).toContain('<div class="flex justify-end"><SuccessRateTimeline')
    expect(source).not.toContain('row.samples')
    expect(source).not.toContain('probe_cost')
  })

  const tokenPricing: UserSupportedModelPricing = {
    billing_mode: 'token', input_price: 0.000005, output_price: 0.000025,
    cache_read_price: 0.0000005, cache_write_price: 0.00000625,
    image_input_price: null, image_output_price: null, per_request_price: null,
    intervals: [
      { min_tokens: 0, max_tokens: 272000, tier_label: '0_272k', input_price: 0.000005, output_price: 0.000025, cache_read_price: 0.0000005, cache_write_price: 0.00000625, per_request_price: null },
      { min_tokens: 272000, max_tokens: null, tier_label: '272k_plus', input_price: 0.00001, output_price: 0.0000375, cache_read_price: 0, cache_write_price: null, per_request_price: null },
    ],
  }
  function pricedGroup(id: number, effectiveRate: number, pricing: UserSupportedModelPricing | null): MarketplaceGroupOption {
    return {
      id, name: `Group ${id}`, platform: 'openai', subscription_type: 'standard',
      rate_multiplier: effectiveRate, effectiveRate, peak_rate_enabled: false,
      peak_start: '', peak_end: '', peak_rate_multiplier: 1, is_exclusive: false,
      image_rate_independent: false, image_rate_multiplier: 1,
      image_price_1k: null, image_price_2k: null, image_price_4k: null, pricing,
    }
  }
  const testPlugins = () => [createPinia(), createI18n({ legacy: false, locale: 'en', missingWarn: false, fallbackWarn: false, messages: { en: {} } })]

  it('shows every group tier using its own prices and rate without filling missing prices from another group', () => {
    const groups = [
      pricedGroup(1, 0.1, tokenPricing),
      pricedGroup(2, 0.2, { ...tokenPricing, intervals: [{ ...tokenPricing.intervals[0]!, tier_label: 'custom', output_price: 0.00005 }] }),
      pricedGroup(3, 1, null),
    ]
    const wrapper = mount(ModelMarketplaceDetailDrawer, {
      props: { entry: { ...entry(), pricing: tokenPricing, groups }, groups, showRechargePrices: false, balanceRechargeMultiplier: 1, officialUsdToCnyRate: 7.2, monitorResolution: 'hour' },
      global: { plugins: testPlugins(), stubs: { teleport: true, ModelIcon: true, Icon: true, SuccessRateTimeline: true, ModelMarketplacePerformanceCharts: true } },
    })
    const cells = (selector: string) => wrapper.findAll(`${selector} tbody tr`).map(row => row.findAll('th, td').map(cell => cell.text()))
    expect(wrapper.get('[data-testid="base-pricing"]').find('table').exists()).toBe(false)
    expect(wrapper.get('[data-testid="base-pricing"]').text()).toContain('$5.00')
    expect(cells('[data-testid="base-tiers"]')).toEqual([
      ['0_272k', '$5.00', '$25.00', '$0.50', '$6.25'],
      ['272k_plus', '$10.00', '$37.50', '$0.00', '-'],
    ])
    expect(cells('[data-testid="group-pricing-1"]')).toEqual([
      ['0_272k', '$0.50', '$2.50', '$0.05', '$0.625'],
      ['272k_plus', '$1.00', '$3.75', '$0.00', '-'],
    ])
    expect(cells('[data-testid="group-pricing-2"]')).toEqual([
      ['custom', '$1.00', '$10.00', '$0.10', '$1.25'],
    ])
    expect(cells('[data-testid="group-pricing-3"]')).toEqual([])
    expect(wrapper.get('[data-testid="group-pricing-3"]').text()).toContain('modelMarketplace.noPricing')
    wrapper.unmount()
  })

  it('uses one comparison table for untiered groups and omits 1h cache prices in every layout', async () => {
    const pricing = { ...tokenPricing, intervals: [], cache_write_1h_price: 0.00002 }
    const groups = [pricedGroup(1, 0.1, pricing), pricedGroup(2, 0.2, { ...pricing, cache_write_price: null })]
    const wrapper = mount(ModelMarketplaceDetailDrawer, {
      props: { entry: { ...entry(), pricing, groups }, groups, showRechargePrices: false, balanceRechargeMultiplier: 1, officialUsdToCnyRate: 7.2, monitorResolution: 'hour' },
      global: { plugins: testPlugins(), stubs: { teleport: true, ModelIcon: true, Icon: true, SuccessRateTimeline: true, ModelMarketplacePerformanceCharts: true } },
    })
    expect(wrapper.find('[data-testid="base-tiers"]').exists()).toBe(false)
    const table = wrapper.get('[data-testid="group-comparison"]')
    expect(table.findAll('table')).toHaveLength(1)
    expect(table.findAll('tbody tr')).toHaveLength(2)
    expect(table.findAll('thead th').map(cell => cell.text())).toEqual([
      'modelMarketplace.details.group', 'modelMarketplace.details.multiplier',
      'modelMarketplace.price.input', 'modelMarketplace.price.output',
      'modelMarketplace.price.cacheRead', 'modelMarketplace.price.cacheWrite',
    ])
    expect(table.get('[data-testid="group-pricing-1"]').findAll('td').map(cell => cell.text())).toEqual(['0.1×', '$0.50', '$2.50', '$0.05', '$0.625'])
    expect(table.get('[data-testid="group-pricing-2"]').findAll('td').map(cell => cell.text())).toEqual(['0.2×', '$1.00', '$5.00', '$0.10', '-'])
    expect(wrapper.text()).not.toContain('1h')
    await wrapper.setProps({ groups: [pricedGroup(1, 0.1, { ...tokenPricing, intervals: tokenPricing.intervals.map(tier => ({ ...tier, cache_write_1h_price: 0.00002 })) })] })
    expect(wrapper.find('[data-testid="group-comparison"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="group-pricing-1"]').findAll('tbody tr')).toHaveLength(2)
    expect(wrapper.text()).not.toContain('1h')
    wrapper.unmount()
  })

  it('preserves image size overrides, free tiers and recharge conversion in the table', async () => {
    const pricing: UserSupportedModelPricing = {
      ...tokenPricing, billing_mode: 'image', per_request_price: 0.1,
      intervals: [{ ...tokenPricing.intervals[0]!, tier_label: '2K', per_request_price: 0.2 }],
    }
    const group = { ...pricedGroup(1, 0.2, pricing), image_price_1k: 0 }
    const wrapper = mount(ModelMarketplacePricingTable, {
      props: { pricing, group, showRechargePrices: false, balanceRechargeMultiplier: 0.5, officialUsdToCnyRate: 7.2 },
      global: { plugins: testPlugins() },
    })
    const cells = () => wrapper.findAll('tbody tr').map(row => row.findAll('th, td').map(cell => cell.text()))
    expect(cells()).toEqual([['1K', '$0.00'], ['2K', '$0.04'], ['4K', '$0.02']])
    await wrapper.setProps({ showRechargePrices: true })
    expect(cells()).toEqual([['1K', '¥0.00'], ['2K', '¥0.08'], ['4K', '¥0.04']])
    wrapper.unmount()
  })
})
