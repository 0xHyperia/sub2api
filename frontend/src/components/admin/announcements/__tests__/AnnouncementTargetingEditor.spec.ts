import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import AnnouncementTargetingEditor from '../AnnouncementTargetingEditor.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const SelectStub = {
  inheritAttrs: true,
  props: ['modelValue', 'options'],
  emits: ['update:modelValue'],
  template: `
    <button
      type="button"
      role="combobox"
      @click="$emit('update:modelValue', modelValue === 'subscription' ? 'balance' : 'gt')"
    >
      {{ modelValue }}
    </button>
  `
}

const GroupSelectorStub = {
  inheritAttrs: true,
  props: ['modelValue', 'groups'],
  emits: ['update:modelValue'],
  template: `
    <div>
      <button
        type="button"
        data-testid="select-subscription-groups"
        @click="$emit('update:modelValue', [7, 8])"
      >
        groups
      </button>
    </div>
  `
}

function mountEditor(modelValue: Record<string, unknown>) {
  return mount(AnnouncementTargetingEditor, {
    props: {
      modelValue,
      groups: []
    } as any,
    global: {
      stubs: {
        Select: SelectStub,
        GroupSelector: GroupSelectorStub,
        Icon: true
      }
    }
  })
}

describe('AnnouncementTargetingEditor', () => {
  it('exposes the targeting mode as a labelled radio group and defaults to a 320px-safe stack', async () => {
    const wrapper = mountEditor({ any_of: [] })
    const radios = wrapper.findAll('input[name="announcement-targeting-mode"]')

    expect(radios).toHaveLength(2)
    expect(radios.every((radio) => radio.attributes('aria-describedby') === 'announcement-targeting-mode-description')).toBe(true)
    expect(wrapper.get('#announcement-targeting-mode-description').exists()).toBe(true)
    expect(wrapper.find('.rounded-2xl').exists()).toBe(false)
    expect(wrapper.find('.shadow-sm').exists()).toBe(false)

    await wrapper.get('#announcement-targeting-custom').setValue()

    expect(wrapper.emitted('update:modelValue')?.at(-1)?.[0]).toEqual({
      any_of: [
        {
          all_of: [
            {
              type: 'subscription',
              operator: 'in',
              group_ids: []
            }
          ]
        }
      ]
    })
  })

  it('associates dynamic controls and validation errors with stable accessible ids', () => {
    const wrapper = mountEditor({
      any_of: [
        {
          all_of: [
            { type: 'subscription', operator: 'in', group_ids: [] },
            { type: 'balance', operator: 'gte', value: 10 }
          ]
        }
      ]
    })

    const root = wrapper.get('.rounded-panel')
    const validationId = root.attributes('aria-describedby')
    expect(root.attributes('aria-invalid')).toBe('true')
    expect(validationId).toBe('announcement-targeting-validation-summary')
    expect(wrapper.get(`#${validationId}`).attributes('role')).toBe('alert')

    const group = wrapper.get('[role="group"][aria-labelledby="announcement-targeting-group-0-label"]')
    expect(wrapper.get(`#${group.attributes('aria-labelledby')}`).exists()).toBe(true)

    const typeControl = wrapper.get('[role="combobox"][aria-labelledby="announcement-targeting-condition-0-0-type-label"]')
    expect(wrapper.get(`#${typeControl.attributes('aria-labelledby')}`).text()).toContain('conditionType')

    const packageControl = wrapper.get('[role="group"][aria-labelledby="announcement-targeting-condition-0-0-packages-label"]')
    expect(packageControl.attributes('aria-invalid')).toBe('true')
    expect(wrapper.get(`#${packageControl.attributes('aria-describedby')}`).text()).toContain('selectPackages')

    const balanceInput = wrapper.get('#announcement-targeting-condition-0-1-value')
    expect(wrapper.get('label[for="announcement-targeting-condition-0-1-value"]').exists()).toBe(true)
    expect(balanceInput.attributes('type')).toBe('number')

    const deleteButtons = wrapper.findAll('button[aria-label^="common.delete"]')
    expect(deleteButtons).toHaveLength(3)
    expect(
      wrapper.findAll('button').some((button) =>
        button.classes().includes('w-full') && button.classes().includes('sm:w-auto')
      )
    ).toBe(true)
  })

  it('writes package and condition changes directly to the targeted condition', async () => {
    const wrapper = mountEditor({
      any_of: [
        {
          all_of: [{ type: 'subscription', operator: 'in', group_ids: [] }]
        }
      ]
    })

    await wrapper.get('[data-testid="select-subscription-groups"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')?.at(-1)?.[0]).toEqual({
      any_of: [
        {
          all_of: [{ type: 'subscription', operator: 'in', group_ids: [7, 8] }]
        }
      ]
    })

    await wrapper.get('[role="combobox"][aria-labelledby="announcement-targeting-condition-0-0-type-label"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')?.at(-1)?.[0]).toEqual({
      any_of: [
        {
          all_of: [{ type: 'balance', operator: 'gte', value: 0 }]
        }
      ]
    })
  })
})
