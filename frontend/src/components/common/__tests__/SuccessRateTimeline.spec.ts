import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import SuccessRateTimeline from '../SuccessRateTimeline.vue'

describe('SuccessRateTimeline', () => {
  it('always renders 30 fixed time buckets and pads missing values as empty', () => {
    const wrapper = mount(SuccessRateTimeline, {
      props: {
        buckets: [{ started_at: '2026-08-03T00:00:00Z', success_rate: 100 }],
        successRate: 100,
        resolution: 'minute',
      },
    })

    const buckets = wrapper.find('[role="img"]').findAll(':scope > span')
    expect(buckets).toHaveLength(30)
    expect(wrapper.text()).toContain('100.0%')
  })

  it.each([
    [99.9, '100%'],
    [99, '88%'],
    [95, '72%'],
    [90, '55%'],
    [70, '48%'],
    [69, '40%'],
  ])('uses the expected height tier for %s%%', (rate, height) => {
    const wrapper = mount(SuccessRateTimeline, {
      props: {
        buckets: [{ started_at: '2026-08-03T00:00:00Z', success_rate: rate }],
        successRate: rate,
      },
    })

    expect(wrapper.find('[role="img"]').findAll(':scope > span').at(29)?.find('button').attributes('style')).toContain(`height: ${height}`)
  })

  it('uses gray for empty buckets and health colors for populated buckets', () => {
    const wrapper = mount(SuccessRateTimeline, {
      props: {
        buckets: [
          { started_at: '2026-08-02T23:00:00Z', success_rate: null },
          { started_at: '2026-08-03T00:00:00Z', success_rate: 50 },
        ],
      },
    })
    const buckets = wrapper.find('[role="img"]').findAll(':scope > span')
    expect(buckets.at(28)?.find('button').classes()).toContain('bg-outline-strong')
    expect(buckets.at(29)?.find('button').classes()).toContain('bg-danger')
  })

  it('renders model availability with red through 30% and warning below 70%', () => {
    const wrapper = mount(SuccessRateTimeline, {
      props: {
        buckets: [
          { started_at: '2026-08-02T20:00:00Z', success_rate: null },
          { started_at: '2026-08-02T21:00:00Z', success_rate: 30 },
          { started_at: '2026-08-02T22:00:00Z', success_rate: 30.1 },
          { started_at: '2026-08-02T23:00:00Z', success_rate: 70 },
          { started_at: '2026-08-03T00:00:00Z', success_rate: 99.9 },
        ],
        variant: 'availability',
        showOverall: false,
      },
    })

    const buckets = wrapper.find('[role="img"]').findAll(':scope > span')
    expect(buckets.at(25)?.find('button').classes()).toContain('bg-outline-strong')
    expect(buckets.at(26)?.find('button').classes()).toContain('bg-danger')
    expect(buckets.at(27)?.find('button').classes()).toContain('bg-warning')
    expect(buckets.at(28)?.find('button').classes()).toContain('bg-success/70')
    expect(buckets.at(29)?.find('button').classes()).toContain('bg-success')
    expect(buckets.at(29)?.find('button').attributes('style')).toContain('height: 100%')
    expect(wrapper.text()).not.toContain('100.0%')
  })

  it('provides a local-time custom hover and keyboard tooltip for every bucket', async () => {
    const wrapper = mount(SuccessRateTimeline, {
      props: {
        buckets: [{ started_at: '2026-08-03T00:00:00Z', success_rate: 57.5 }],
        successRate: 57.5,
      },
    })

    const bucket = wrapper.find('[role="img"]').findAll(':scope > span').at(29)
    const button = bucket?.find('button')
    expect(button?.attributes('aria-label')).toContain('57.50%')
    expect(button?.attributes('aria-label')).not.toContain('2026-08-03T00:00:00.000Z')
    await button?.trigger('focus')
    expect(document.body.textContent).toContain('57.50%')
  })

  it('fills missing time slices in the middle of the returned range', () => {
    const wrapper = mount(SuccessRateTimeline, {
      props: {
        buckets: [
          { started_at: '2026-08-03T00:00:00Z', success_rate: 100 },
          { started_at: '2026-08-03T02:00:00Z', success_rate: 50 },
        ],
        resolution: 'hour',
      },
    })

    const buckets = wrapper.find('[role="img"]').findAll(':scope > span')
    expect(buckets.at(28)?.find('button').classes()).toContain('bg-outline-strong')
    expect(buckets.at(29)?.find('button').classes()).toContain('bg-danger')
  })
})
