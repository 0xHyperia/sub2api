import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

import Select from '../Select.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

const originalInnerWidth = window.innerWidth

afterEach(() => {
  document.body.innerHTML = ''
  Object.defineProperty(window, 'innerWidth', { configurable: true, value: originalInnerWidth })
  vi.restoreAllMocks()
})

describe('Select', () => {
  it('implements select-only combobox keyboard navigation and selection', async () => {
    const wrapper = mount(Select, {
      attachTo: document.body,
      props: {
        modelValue: null,
        searchable: false,
        placeholder: 'Provider',
        options: [
          { value: 'alpha', label: 'Alpha' },
          { value: 'beta', label: 'Beta' },
          { value: 'gamma', label: 'Gamma' }
        ]
      }
    })
    const trigger = wrapper.get('button.select-trigger')

    expect(trigger.attributes('role')).toBe('combobox')
    expect(trigger.attributes('aria-haspopup')).toBe('listbox')
    expect(trigger.attributes('aria-label')).toBe('Provider')

    await trigger.trigger('keydown', { key: 'ArrowDown' })
    await wrapper.vm.$nextTick()

    const listboxId = trigger.attributes('aria-controls')
    const listbox = document.getElementById(listboxId)
    expect(listbox?.getAttribute('role')).toBe('listbox')
    expect(trigger.attributes('aria-activedescendant')).toContain('-option-0')

    await trigger.trigger('keydown', { key: 'ArrowDown' })
    expect(trigger.attributes('aria-activedescendant')).toContain('-option-1')

    await trigger.trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('update:modelValue')).toEqual([['beta']])
    expect(wrapper.emitted('change')?.[0]?.[0]).toBe('beta')
    expect(trigger.attributes('aria-expanded')).toBe('false')

    wrapper.unmount()
  })

  it('supports typeahead while keeping focus on a non-searchable trigger', async () => {
    const wrapper = mount(Select, {
      attachTo: document.body,
      props: {
        modelValue: null,
        searchable: false,
        options: [
          { value: 'alpha', label: 'Alpha' },
          { value: 'beta', label: 'Beta' },
          { value: 'gamma', label: 'Gamma' }
        ]
      }
    })
    const trigger = wrapper.get('button.select-trigger')
    trigger.element.focus()

    await trigger.trigger('keydown', { key: 'g' })
    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()

    expect(document.activeElement).toBe(trigger.element)
    expect(trigger.attributes('aria-activedescendant')).toContain('-option-2')

    await trigger.trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('update:modelValue')).toEqual([['gamma']])

    wrapper.unmount()
  })

  it('connects the searchable combobox to its listbox', async () => {
    const wrapper = mount(Select, {
      attachTo: document.body,
      props: {
        modelValue: null,
        searchable: true,
        searchPlaceholder: 'Search providers',
        options: [
          { value: 'alpha', label: 'Alpha' },
          { value: 'beta', label: 'Beta' }
        ]
      }
    })

    await wrapper.get('button.select-trigger').trigger('click')
    await wrapper.vm.$nextTick()

    const input = document.querySelector<HTMLInputElement>('.select-search-input')
    const listbox = document.querySelector<HTMLElement>('[role="listbox"]')
    expect(input?.getAttribute('role')).toBe('combobox')
    expect(input?.getAttribute('aria-controls')).toBe(listbox?.id)
    expect(input?.getAttribute('aria-activedescendant')).toContain('-option-0')

    wrapper.unmount()
  })

  it('forwards labels and ids to the interactive trigger while preserving root data attributes', () => {
    const wrapper = mount(Select, {
      attrs: {
        id: 'provider-select',
        'aria-labelledby': 'provider-label',
        'aria-describedby': 'provider-hint',
        'data-testid': 'provider-field',
        class: 'w-full'
      },
      props: {
        modelValue: null,
        options: [{ value: 'alpha', label: 'Alpha' }]
      }
    })

    const trigger = wrapper.get('button.select-trigger')
    expect(trigger.attributes('id')).toBe('provider-select')
    expect(trigger.attributes('aria-labelledby')).toBe('provider-label')
    expect(trigger.attributes('aria-describedby')).toBe('provider-hint')
    expect(trigger.attributes('aria-label')).toBeUndefined()
    expect(wrapper.attributes('data-testid')).toBe('provider-field')
    expect(wrapper.classes()).toContain('w-full')
  })

  it('renders clear as a separate named button and clears with pointer or keyboard', async () => {
    const wrapper = mount(Select, {
      props: {
        modelValue: 'alpha',
        clearable: true,
        options: [{ value: 'alpha', label: 'Alpha' }]
      }
    })

    const trigger = wrapper.get('button.select-trigger')
    const clear = wrapper.get('button.select-clear')
    expect(trigger.find('button').exists()).toBe(false)
    expect(clear.attributes('aria-label')).toBe('common.clear')

    await clear.trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([[null]])

    await trigger.trigger('keydown', { key: 'Delete' })
    expect(wrapper.emitted('update:modelValue')).toEqual([[null], [null]])
  })
})

function setViewportWidth(width: number) {
  Object.defineProperty(window, 'innerWidth', { configurable: true, value: width })
}

function mockTriggerRect(left: number, width: number) {
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
    x: left,
    y: 20,
    top: 20,
    right: left + width,
    bottom: 60,
    left,
    width,
    height: 40,
    toJSON: () => ({})
  })
}

async function openSelect() {
  const wrapper = mount(Select, {
    props: {
      modelValue: null,
      options: [{
        value: 'example',
        label: 'very-long-unbroken-option-value-that-must-not-overflow'
      }]
    }
  })

  await wrapper.get('button').trigger('click')
  await nextTick()

  return { wrapper, dropdown: document.body.querySelector<HTMLElement>('.select-dropdown-portal') }
}

describe('Select dropdown viewport constraints', () => {
  it.each([
    { viewport: 1024, left: 20, width: 80, expectedLeft: '20px', minWidth: '200px', maxWidth: '996px' },
    { viewport: 320, left: 220, width: 80, expectedLeft: '220px', minWidth: '92px', maxWidth: '92px' },
    { viewport: 320, left: -20, width: 80, expectedLeft: '8px', minWidth: '200px', maxWidth: '304px' },
    { viewport: 320, left: 400, width: 80, expectedLeft: '312px', minWidth: '0px', maxWidth: '0px' }
  ])('keeps the dropdown inside a $viewport px viewport', async ({ viewport, left, width, expectedLeft, minWidth, maxWidth }) => {
    setViewportWidth(viewport)
    mockTriggerRect(left, width)

    const { wrapper, dropdown } = await openSelect()
    expect(dropdown).not.toBeNull()
    expect(dropdown?.style.left).toBe(expectedLeft)
    expect(dropdown?.style.minWidth).toBe(minWidth)
    expect(dropdown?.style.maxWidth).toBe(maxWidth)
    wrapper.unmount()
  })
})
