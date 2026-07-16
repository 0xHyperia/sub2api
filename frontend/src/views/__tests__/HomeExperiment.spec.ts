import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import HomeExperiment from '@/views/HomeExperiment.vue'
import { getHomeShowcase } from '@/api/home'

const source = readFileSync(resolve(dirname(fileURLToPath(import.meta.url)), '../HomeExperiment.vue'), 'utf8')

vi.mock('@/api/home', () => ({
  getHomeMetrics: vi.fn().mockResolvedValue({}),
  getHomeShowcase: vi.fn(),
}))

vi.mock('@/components/user/ModelMonitorTimeline.vue', () => ({
  default: {
    props: ['points', 'limit'],
    template: '<div role="list"><span v-for="point in points" :key="point.checked_at" role="listitem"></span></div>',
  },
}))

describe('HomeExperiment', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue(null)
    vi.mocked(getHomeShowcase).mockResolvedValue([
      {
        platform: 'openai',
        model_count: 8,
        models: [
          {
            name: 'gpt-featured',
            platform: 'openai',
            rate_multiplier: 0.8,
            pricing: {
              billing_mode: 'token',
              input_price: 0.0000025,
              output_price: 0.000015,
              cache_write_price: null,
              cache_read_price: null,
              image_output_price: null,
              per_request_price: null,
              intervals: [],
            },
            monitor_status: {
              status: 'operational',
              latency_ms: 120,
              availability_7d: 99.9,
              last_checked_at: '2026-07-15T00:00:00Z',
              timeline: [
                { status: 'operational', latency_ms: 120, checked_at: '2026-07-15T00:00:00Z' },
                { status: 'degraded', latency_ms: 6800, checked_at: '2026-07-14T23:55:00Z' },
              ],
              display_order: 20,
              label: '热门',
            },
          },
          {
            name: 'gpt-mini',
            platform: 'openai',
            rate_multiplier: 1,
            pricing: null,
            monitor_status: null,
          },
        ],
      },
      {
        platform: 'anthropic',
        model_count: 1,
        models: [{
          name: 'claude-featured',
          platform: 'anthropic',
          rate_multiplier: 1,
          pricing: null,
          monitor_status: null,
        }],
      },
    ])
  })

  it('fills the first viewport without leaving a 24px gap below the hero', () => {
    expect(source).toContain('min-height: 100svh')
    expect(source).toContain('min-height: 100dvh')
    expect(source).not.toContain('calc(100svh - 24px)')
    expect(source).not.toContain('calc(100dvh - 24px)')
  })

  it('uses a disclosure navigation without changing the hero document flow', async () => {
    const wrapper = mount(HomeExperiment, {
      props: {
        siteName: 'USA-零',
        siteSubtitle: 'Unified gateway',
        isAuthenticated: false,
        dashboardPath: '/dashboard',
      },
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
        },
      },
    })
    await flushPromises()

    const toggle = wrapper.get('.mobile-nav-toggle')
    expect(toggle.attributes('aria-expanded')).toBe('false')
    await toggle.trigger('click')
    expect(toggle.attributes('aria-expanded')).toBe('true')
    expect(wrapper.get('#home-navigation').classes()).toContain('is-open')

    expect(wrapper.find('.console-shell').element.tagName).toBe('SECTION')
    expect(wrapper.find('[role="img"].console-shell').exists()).toBe(false)
    wrapper.unmount()
  })

  it('supports route tabs and renders the live model showcase', async () => {
    const wrapper = mount(HomeExperiment, {
      props: {
        siteName: 'USA-零',
        siteSubtitle: 'Unified gateway',
        isAuthenticated: false,
        dashboardPath: '/dashboard',
      },
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
          },
        },
      },
    })
    await flushPromises()

    const tabs = wrapper.findAll('.tabs [role="tab"]')
    expect(tabs[0].attributes('aria-selected')).toBe('true')
    await wrapper.get('.tabs').trigger('keydown', { key: 'ArrowRight' })
    expect(tabs[1].attributes('aria-selected')).toBe('true')
    expect(wrapper.get('.route-demo [role="tabpanel"]').attributes('aria-labelledby')).toBe('route-tab-responses')

    expect(wrapper.find('table.pricing-table').exists()).toBe(false)
    expect(wrapper.get('.featured-model-title').text()).toContain('gpt-featured')
    expect(wrapper.get('.featured-prices').text()).toContain('$2')
    expect(wrapper.get('.featured-prices').text()).toContain('/ 1M')
    expect(wrapper.get('.featured-status').text()).toContain('状态正常')
    expect(wrapper.get('.featured-monitor-heading').text()).toContain('最近 2 次')
    expect(wrapper.findAll('.featured-monitor-history [role="listitem"]')).toHaveLength(2)
    expect(wrapper.get('.featured-billing-heading').text()).toContain('已计入公开倍率')
    expect(wrapper.get('.showcase-list-heading').text()).toContain('同厂商模型')
    expect(wrapper.get('.showcase-model-row').text()).toContain('02')
    expect(wrapper.get('.showcase-row-availability').text()).toContain('状态待检测')

    const providerTabs = wrapper.findAll('.showcase-provider-tabs [role="tab"]')
    expect(providerTabs).toHaveLength(2)
    await providerTabs[1].trigger('click')
    expect(wrapper.get('.featured-model-title').text()).toContain('claude-featured')
    expect(wrapper.find('.featured-monitor-history').exists()).toBe(false)
    wrapper.unmount()
  })

  it('hides the showcase cleanly when the public catalog request fails', async () => {
    vi.mocked(getHomeShowcase).mockRejectedValueOnce(new Error('unavailable'))
    const wrapper = mount(HomeExperiment, {
      props: { siteName: 'USA-零', siteSubtitle: 'Unified gateway', isAuthenticated: false, dashboardPath: '/dashboard' },
      global: { stubs: { RouterLink: { props: ['to'], template: '<a><slot /></a>' } } },
    })
    await flushPromises()
    expect(wrapper.find('.model-showcase-inner').exists()).toBe(false)
    expect(wrapper.get('.feature-grid').exists()).toBe(true)
    wrapper.unmount()
  })
})
