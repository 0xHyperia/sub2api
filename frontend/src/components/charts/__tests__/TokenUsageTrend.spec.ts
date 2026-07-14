import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import TokenUsageTrend from '../TokenUsageTrend.vue'

const messages: Record<string, string> = {
  'admin.dashboard.tokenUsageTrend': 'Token Usage Trend',
  'admin.dashboard.noDataAvailable': 'No data available',
}

const themeState = vi.hoisted(() => ({
  isDark: null as unknown as { value: boolean },
  chartTheme: null as unknown as { value: Record<string, string> },
}))

vi.mock('@/composables/useChartTheme', async () => {
  const { computed, ref } = await import('vue')
  themeState.isDark = ref(false)
  themeState.chartTheme = computed(() => ({
    foregroundMuted: themeState.isDark.value ? 'rgb(190, 202, 214)' : 'rgb(71, 85, 105)',
    outline: themeState.isDark.value ? 'rgb(45, 57, 70)' : 'rgb(226, 232, 240)',
    info: 'rgb(37, 99, 235)',
    infoAlpha: 'rgba(37, 99, 235, 0.14)',
    success: 'rgb(5, 150, 105)',
    successAlpha: 'rgba(5, 150, 105, 0.14)',
    warning: 'rgb(217, 119, 6)',
    warningAlpha: 'rgba(217, 119, 6, 0.14)',
    brand: 'rgb(15, 23, 42)',
    brandAlpha: 'rgba(15, 23, 42, 0.14)',
    danger: 'rgb(220, 38, 38)',
  }))
  return {
    useChartTheme: () => ({ chartTheme: themeState.chartTheme }),
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

vi.mock('vue-chartjs', () => ({
  Line: {
    props: ['data', 'options'],
    template: '<div><div class="chart-data">{{ JSON.stringify(data) }}</div><div class="chart-options">{{ JSON.stringify(options) }}</div></div>',
  },
}))

describe('TokenUsageTrend', () => {
  beforeEach(() => {
    themeState.isDark.value = false
  })

  it('calculates cache hit rate against all prompt tokens', () => {
    const wrapper = mount(TokenUsageTrend, {
      props: {
        trendData: [
          {
            date: '2026-05-08',
            requests: 1,
            input_tokens: 500,
            output_tokens: 100,
            cache_creation_tokens: 0,
            cache_read_tokens: 1500,
            cost: 0.01,
            actual_cost: 0.005,
          },
        ],
      },
      global: {
        stubs: {
          LoadingSpinner: true,
        },
      },
    })

    const chartData = JSON.parse(wrapper.find('.chart-data').text())
    const hitRateDataset = chartData.datasets.find(
      (ds: any) => ds.label === 'Cache Hit Rate'
    )
    // Hit rate = 1500 / (500 + 1500 + 0) * 100 = 75%
    expect(hitRateDataset.data[0]).toBe(75)
  })

  it('returns 0 hit rate when all prompt tokens are zero', () => {
    const wrapper = mount(TokenUsageTrend, {
      props: {
        trendData: [
          {
            date: '2026-05-08',
            requests: 0,
            input_tokens: 0,
            output_tokens: 0,
            cache_creation_tokens: 0,
            cache_read_tokens: 0,
            cost: 0,
            actual_cost: 0,
          },
        ],
      },
      global: {
        stubs: {
          LoadingSpinner: true,
        },
      },
    })

    const chartData = JSON.parse(wrapper.find('.chart-data').text())
    const hitRateDataset = chartData.datasets.find(
      (ds: any) => ds.label === 'Cache Hit Rate'
    )
    expect(hitRateDataset.data[0]).toBe(0)
  })

  it('includes cache_creation_tokens in denominator for Anthropic models', () => {
    const wrapper = mount(TokenUsageTrend, {
      props: {
        trendData: [
          {
            date: '2026-05-08',
            requests: 1,
            input_tokens: 200,
            output_tokens: 50,
            cache_creation_tokens: 300,
            cache_read_tokens: 500,
            cost: 0.02,
            actual_cost: 0.01,
          },
        ],
      },
      global: {
        stubs: {
          LoadingSpinner: true,
        },
      },
    })

    const chartData = JSON.parse(wrapper.find('.chart-data').text())
    const hitRateDataset = chartData.datasets.find(
      (ds: any) => ds.label === 'Cache Hit Rate'
    )
    // Hit rate = 500 / (200 + 500 + 300) * 100 = 50%
    expect(hitRateDataset.data[0]).toBe(50)
  })

  it('updates chart colors when the resolved theme changes', async () => {
    const wrapper = mount(TokenUsageTrend, {
      props: {
        trendData: [
          {
            date: '2026-05-08',
            requests: 1,
            input_tokens: 100,
            output_tokens: 50,
            cache_creation_tokens: 0,
            cache_read_tokens: 0,
            cost: 0.01,
            actual_cost: 0.005,
          },
        ],
      },
      global: {
        stubs: {
          LoadingSpinner: true,
        },
      },
    })

    const lightOptions = JSON.parse(wrapper.get('.chart-options').text())
    expect(lightOptions.scales.x.ticks.color).toBe('rgb(71, 85, 105)')

    themeState.isDark.value = true
    await nextTick()

    const darkOptions = JSON.parse(wrapper.get('.chart-options').text())
    expect(darkOptions.scales.x.ticks.color).toBe('rgb(190, 202, 214)')
  })
})
