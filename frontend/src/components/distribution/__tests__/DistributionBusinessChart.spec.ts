import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import DistributionBusinessChart from '../DistributionBusinessChart.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key.split('.').at(-1) || key }) }))

const metric = (date: string, value = 10) => ({
  date,
  new_customers: value,
  paying_customers: Math.floor(value / 2),
  customer_paid_cny: String(value * 10),
  commission_cny: String(value),
})

describe('DistributionBusinessChart', () => {
  it('renders readable axes and hourly bucket ranges', async () => {
    const wrapper = mount(DistributionBusinessChart, {
      props: {
        direct: [metric('2026-08-13T09:00:00+08:00'), metric('2026-08-13T10:00:00+08:00', 20)],
        resolution: 'hour',
        title: 'Trend',
      },
    })
    expect(wrapper.findAll('.chart-y-axis span')).toHaveLength(5)
    expect(wrapper.findAll('.chart-x-axis span')).toHaveLength(2)
    expect(wrapper.findAll('.chart-marker')).toHaveLength(2)

    Object.defineProperty(wrapper.find('.chart-plot').element, 'getBoundingClientRect', {
      value: () => ({ left: 0, width: 200, top: 0, bottom: 200, right: 200, height: 200 }),
    })
    await wrapper.find('.chart-plot').trigger('pointermove', { clientX: 200 })
    expect(wrapper.find('.tooltip-title').text()).toContain('10:00')
    expect(wrapper.find('.tooltip-title').text()).toContain('10:59')
  })

  it('limits x labels and removes permanent markers for long series', () => {
    const direct = Array.from({ length: 40 }, (_, index) => metric(`2026-07-${String(index % 28 + 1).padStart(2, '0')}`, index + 1))
    const wrapper = mount(DistributionBusinessChart, { props: { direct, resolution: 'week' } })
    expect(wrapper.findAll('.chart-x-axis span').length).toBeLessThanOrEqual(6)
    expect(wrapper.findAll('.chart-marker')).toHaveLength(0)
  })

  it('switches the displayed metric without changing the time series', async () => {
    const wrapper = mount(DistributionBusinessChart, { props: { direct: [metric('2026-08-12', 4), metric('2026-08-13', 6)] } })
    const newCustomers = wrapper.findAll('.chart-segments button').find(button => button.text() === 'newCustomers')!
    await newCustomers.trigger('click')
    expect(wrapper.find('.chart-current strong').text()).toBe('10')
  })
})
