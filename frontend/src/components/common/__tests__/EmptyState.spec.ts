import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

import EmptyState from '../EmptyState.vue'

describe('EmptyState', () => {
  it('renders the legacy message prop as the empty-state title', () => {
    const wrapper = mount(EmptyState, {
      props: { message: 'No matching records' },
      global: { stubs: { Icon: true } },
    })

    expect(wrapper.get('h3').text()).toBe('No matching records')
  })

  it('uses a non-submitting button for local actions', async () => {
    const wrapper = mount(EmptyState, {
      props: { actionText: 'Retry', actionIcon: false },
      global: { stubs: { Icon: true } },
    })

    const button = wrapper.get('button')
    expect(button.attributes('type')).toBe('button')
    await button.trigger('click')
    expect(wrapper.emitted('action')).toHaveLength(1)
  })
})
