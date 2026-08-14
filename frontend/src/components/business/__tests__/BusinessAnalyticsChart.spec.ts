import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import BusinessAnalyticsChart from '../BusinessAnalyticsChart.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key.split('.').at(-1) || key }) }))
vi.mock('vue-chartjs', async () => {
  const { defineComponent, h } = await import('vue')
  return { Line: defineComponent({ props: ['data'], setup: props => () => h('div', { class: 'mock-chart' }, JSON.stringify(props.data)) }) }
})

describe('BusinessAnalyticsChart', () => {
  it('uses the period metric instead of summing non-additive trend values', () => {
    const wrapper = mount(BusinessAnalyticsChart, {
      props: {
        title: 'Trend', subtitle: 'Period', granularity: 'day', defaultMetrics: ['active_users'],
        availableMetrics: ['active_users'],
        points: [{ bucket: '2026-08-01', active_users: 8 }, { bucket: '2026-08-02', active_users: 9 }] as never,
      },
    })
    expect(wrapper.find('.totals strong').text()).toBe('17')
  })

  it('supports keyboard traversal through time points', async () => {
    const wrapper = mount(BusinessAnalyticsChart, {
      props: {
        title: 'Trend', subtitle: 'Period', granularity: 'day', defaultMetrics: ['new_users'], availableMetrics: ['new_users'],
        points: [{ bucket: '2026-08-01', new_users: 2 }, { bucket: '2026-08-02', new_users: 4 }] as never,
      },
    })
    expect(wrapper.find('.chart-wrap').attributes('tabindex')).toBe('0')
  })

  it('renders hour labels for a single-day report', () => {
    const wrapper = mount(BusinessAnalyticsChart, { props: { title: 'Trend', subtitle: 'Period', granularity: 'hour', defaultMetrics: ['new_users'], availableMetrics: ['new_users'], points: [{ bucket: '2026-08-13T08:00:00+08:00', new_users: 2 }] as never } })
    expect(wrapper.html()).toContain('08:00')
  })

  it('formats first-paid users as a user count instead of currency', () => {
    const wrapper = mount(BusinessAnalyticsChart, {
      props: {
        title: 'Trend', subtitle: 'Period', granularity: 'day', defaultMetrics: ['first_paid_users'],
        availableMetrics: ['first_paid_users'], totals: { first_paid_users: { value: 19, previous: 28 } },
        points: [{ bucket: '2026-08-14', first_paid_users: 19 }] as never,
      },
    })
    const total = wrapper.get('.totals strong').text()
    expect(total).toBe('19')
    expect(total).not.toMatch(/[¥￥$]/)
  })

  it('keeps paid amount metrics formatted as CNY currency', () => {
    const wrapper = mount(BusinessAnalyticsChart, {
      props: {
        title: 'Trend', subtitle: 'Period', granularity: 'day', defaultMetrics: ['net_paid_cny'],
        availableMetrics: ['net_paid_cny'], totals: { net_paid_cny: { value: 19, previous: 28 } },
        points: [{ bucket: '2026-08-14', net_paid_cny: 19 }] as never,
      },
    })
    expect(wrapper.get('.totals strong').text()).toBe('￥19.00')
  })

  it('formats model consumption metrics with a plain dollar sign', () => {
    const wrapper = mount(BusinessAnalyticsChart, {
      props: {
        title: 'Trend', subtitle: 'Period', granularity: 'day', defaultMetrics: ['consumed_revenue'],
        availableMetrics: ['consumed_revenue'], totals: { consumed_revenue: { value: 19, previous: 28 } },
        points: [{ bucket: '2026-08-14', consumed_revenue: 19 }] as never,
      },
    })
    expect(wrapper.get('.totals strong').text()).toBe('$19.00')
    expect(wrapper.get('.totals strong').text()).not.toContain('US$')
  })

  it('does not mix CNY and USD metrics on one axis', async () => {
    const wrapper = mount(BusinessAnalyticsChart, {
      props: {
        title: 'Trend', subtitle: 'Period', granularity: 'day', defaultMetrics: ['net_paid_cny'],
        availableMetrics: ['net_paid_cny', 'consumed_revenue'],
        points: [{ bucket: '2026-08-14', net_paid_cny: 19, consumed_revenue: 3 }] as never,
      },
    })
    await wrapper.findAll('.segments button')[1].trigger('click')
    expect(wrapper.findAll('.totals strong').map(node => node.text())).toEqual(['$3.00'])
  })

  it('does not report zero growth when the previous period has no baseline', () => {
    const wrapper = mount(BusinessAnalyticsChart, {
      props: {
        title: 'Trend', subtitle: 'Period', granularity: 'day', defaultMetrics: ['new_users'],
        availableMetrics: ['new_users'], totals: { new_users: { value: 8, previous: 0 } },
        points: [{ bucket: '2026-08-14', new_users: 8 }] as never,
      },
    })
    expect(wrapper.get('.totals small').text()).toBe('noComparison')
    expect(wrapper.get('.totals small').classes()).toContain('neutral')
  })
})
