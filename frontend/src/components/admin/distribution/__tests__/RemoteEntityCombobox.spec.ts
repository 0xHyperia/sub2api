import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'

import RemoteEntityCombobox from '../RemoteEntityCombobox.vue'

describe('RemoteEntityCombobox', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  it('selects an eligible user and ignores disabled candidates', async () => {
    vi.useFakeTimers()
    const search = vi.fn().mockResolvedValue([
      {
        id: 1,
        email: 'existing@example.com',
        username: 'existing',
        selectable: false,
        reason: '已经是代理',
      },
      {
        id: 2,
        email: 'eligible@example.com',
        username: 'eligible',
        selectable: true,
        meta: '用户正常',
      },
    ])
    const wrapper = mount(RemoteEntityCombobox, {
      props: {
        modelValue: null,
        label: '选择用户',
        placeholder: '输入用户邮箱或用户名',
        inputId: 'agent-user',
        search,
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    await wrapper.get('input[role="combobox"]').setValue('eligible')
    await vi.advanceTimersByTimeAsync(250)
    await flushPromises()

    expect(search).toHaveBeenCalledWith('eligible')
    const options = wrapper.findAll('[role="option"]')
    expect(options).toHaveLength(2)
    expect(options[0].attributes('aria-disabled')).toBe('true')

    await options[0].trigger('click')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([null])

    await options[1].trigger('click')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([
      expect.objectContaining({ id: 2, email: 'eligible@example.com' }),
    ])
  })
})
