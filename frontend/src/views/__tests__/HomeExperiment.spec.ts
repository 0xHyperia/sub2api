import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import HomeExperiment from '@/views/HomeExperiment.vue'

vi.mock('@/api/home', () => ({
  getHomeMetrics: vi.fn().mockResolvedValue({}),
}))

describe('HomeExperiment', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue(null)
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

  it('supports arrow-key tabs and renders semantic model pricing tables', async () => {
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

    const tabs = wrapper.findAll('[role="tab"]')
    expect(tabs[0].attributes('aria-selected')).toBe('true')
    await wrapper.get('[role="tablist"]').trigger('keydown', { key: 'ArrowRight' })
    expect(tabs[1].attributes('aria-selected')).toBe('true')
    expect(wrapper.get('[role="tabpanel"]').attributes('aria-labelledby')).toBe('route-tab-responses')

    expect(wrapper.findAll('table.pricing-table')).toHaveLength(3)
    expect(wrapper.findAll('th[scope="col"]').length).toBeGreaterThan(3)
    expect(wrapper.findAll('th[scope="row"]').length).toBeGreaterThan(10)
    wrapper.unmount()
  })
})
