import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import PlatformUsageBreakdown from '../PlatformUsageBreakdown.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

describe('PlatformUsageBreakdown', () => {
  it('exposes a viewport-bound tooltip to keyboard users', async () => {
    const wrapper = mount(PlatformUsageBreakdown, {
      props: {
        today: 3,
        total: 10,
        byPlatform: [
          {
            platform: 'openai',
            today_actual_cost: 2,
            total_actual_cost: 8
          }
        ]
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    const trigger = wrapper.get('button')
    const tooltip = wrapper.get('[role="tooltip"]')
    expect(trigger.attributes('aria-describedby')).toBe(tooltip.attributes('id'))
    expect(trigger.attributes('aria-expanded')).toBe('false')
    expect(tooltip.classes()).toContain('w-[min(18rem,calc(100vw-2rem))]')

    await trigger.trigger('focus')
    expect(trigger.attributes('aria-expanded')).toBe('true')
    expect(tooltip.classes()).toContain('visible')

    await trigger.trigger('keydown', { key: 'Escape' })
    expect(trigger.attributes('aria-expanded')).toBe('false')
    expect(tooltip.classes()).toContain('invisible')
    expect(tooltip.text()).toContain('admin.users.platformOther')
  })
})
