import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import DistributionAnalyticsRange from '../DistributionAnalyticsRange.vue'
import { defaultDistributionAnalyticsRange, rangeForPreset } from '../distributionAnalyticsRange'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key.split('.').at(-1) || key, locale: ref('zh-CN') }) }))

describe('DistributionAnalyticsRange', () => {
  it('uses complete calendar periods for previous week and month presets', () => {
    const lastWeek = rangeForPreset('week', 1)
    const from = new Date(`${lastWeek.date_from}T00:00:00Z`)
    const to = new Date(`${lastWeek.date_to}T00:00:00Z`)
    expect(from.getUTCDay()).toBe(1)
    expect(to.getUTCDay()).toBe(0)
    expect((to.getTime() - from.getTime()) / 86_400_000).toBe(6)

    const lastMonth = rangeForPreset('month', 1)
    expect(lastMonth.date_from.endsWith('-01')).toBe(true)
    const monthEnd = new Date(`${lastMonth.date_to}T00:00:00Z`)
    expect(monthEnd.getUTCDate()).toBe(new Date(Date.UTC(monthEnd.getUTCFullYear(), monthEnd.getUTCMonth() + 1, 0)).getUTCDate())
  })

  it('emits one applied range when a preset is selected', async () => {
    const wrapper = mount(DistributionAnalyticsRange, { props: { modelValue: defaultDistributionAnalyticsRange() } })
    const day = wrapper.findAll('.analytics-segment button').find(button => button.text() === 'day')!
    await day.trigger('click')
    expect(wrapper.emitted('change')).toHaveLength(1)
    expect(wrapper.emitted('change')?.[0]?.[0]).toMatchObject({ granularity: 'day', preset: 'day-0' })
  })

  it('keeps custom edits local until confirmation and reports invalid ranges', async () => {
    const initial = defaultDistributionAnalyticsRange()
    const wrapper = mount(DistributionAnalyticsRange, { props: { modelValue: initial } })
    const custom = wrapper.findAll('.analytics-segment button').find(button => button.text() === 'custom')!
    await custom.trigger('click')
    expect(wrapper.emitted('change')).toBeUndefined()

    const inputs = wrapper.findAll('input[type="date"]')
    await inputs[0].setValue('2026-08-10')
    await inputs[1].setValue('2026-08-01')
    await wrapper.find('.btn-primary').trigger('click')
    expect(wrapper.find('[role="alert"]').exists()).toBe(true)
    expect(wrapper.emitted('change')).toBeUndefined()

    await inputs[1].setValue('2026-08-12')
    await wrapper.find('.btn-primary').trigger('click')
    expect(wrapper.emitted('change')?.[0]?.[0]).toMatchObject({ granularity: 'custom', date_from: '2026-08-10', date_to: '2026-08-12' })
  })
})
