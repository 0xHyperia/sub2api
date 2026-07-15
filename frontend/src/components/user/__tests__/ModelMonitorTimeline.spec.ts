import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import { describe, expect, it, vi } from 'vitest'

import ModelMonitorTimeline from '../ModelMonitorTimeline.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: ref('zh-CN'),
    }),
  }
})

describe('ModelMonitorTimeline', () => {
  it('provides polished keyboard-accessible timeline tooltips', async () => {
    const wrapper = mount(ModelMonitorTimeline, {
      attachTo: document.body,
      props: {
        limit: 4,
        points: [
          { status: 'operational', latency_ms: 128, checked_at: '2026-07-15T10:00:00Z' },
          { status: 'degraded', latency_ms: 8300, checked_at: '2026-07-15T09:00:00Z' },
        ],
      },
    })

    const points = wrapper.findAll('button')
    const tooltips = wrapper.findAll('[role="tooltip"]')
    expect(points).toHaveLength(2)
    expect(tooltips).toHaveLength(2)
    expect(wrapper.findAll('[aria-hidden="true"]')).toHaveLength(2)
    expect(points[0].attributes('aria-describedby')).toBe(tooltips[0].attributes('id'))
    expect(tooltips[0].classes()).toContain('right-0')
    expect(tooltips[1].classes()).toContain('right-0')
    expect(tooltips[0].text()).toContain('modelMarketplace.monitor.degraded')
    expect(tooltips[0].text()).toContain('8300 ms')

    const trigger = points[0].element as HTMLButtonElement
    trigger.focus()
    expect(document.activeElement).toBe(trigger)
    await points[0].trigger('keydown', { key: 'Escape' })
    expect(document.activeElement).not.toBe(trigger)

    wrapper.unmount()
  })

  it('renders an empty track when the API returns a null timeline', () => {
    const wrapper = mount(ModelMonitorTimeline, {
      props: { points: null },
    })
    expect(wrapper.findAll('[aria-hidden="true"]')).toHaveLength(30)
  })
})
