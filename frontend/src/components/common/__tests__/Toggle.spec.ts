import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import Toggle from '../Toggle.vue'

describe('Toggle', () => {
  it('supports an explicit aria-label without changing its model event', async () => {
    const wrapper = mount(Toggle, {
      props: {
        modelValue: false,
        ariaLabel: 'Enable monitoring'
      }
    })
    const toggle = wrapper.get('[role="switch"]')

    expect(toggle.attributes('aria-label')).toBe('Enable monitoring')
    expect(toggle.attributes('aria-checked')).toBe('false')
    await toggle.trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([[true]])
  })

  it('supports aria-labelledby for visible external labels', () => {
    const wrapper = mount(Toggle, {
      props: {
        modelValue: true,
        ariaLabelledby: 'monitoring-label'
      }
    })

    expect(wrapper.get('[role="switch"]').attributes('aria-labelledby')).toBe('monitoring-label')
  })

  it('warns in development when the accessible name is blank', () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => undefined)

    mount(Toggle, {
      props: {
        modelValue: false,
        ariaLabel: ' '
      }
    })

    expect(warn).toHaveBeenCalledWith(
      '[Toggle] Accessible name required: provide a non-empty aria-label or aria-labelledby.'
    )
    warn.mockRestore()
  })
})
