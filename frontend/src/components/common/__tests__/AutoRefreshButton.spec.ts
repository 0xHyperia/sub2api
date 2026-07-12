import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'

import AutoRefreshButton from '../AutoRefreshButton.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${JSON.stringify(params)}` : key
  })
}))

const mountedWrappers: Array<ReturnType<typeof mount>> = []

function mountMenu() {
  const wrapper = mount(AutoRefreshButton, {
    attachTo: document.body,
    props: {
      enabled: true,
      intervalSeconds: 10,
      countdown: 7,
      intervals: [5, 10, 30]
    },
    global: {
      stubs: {
        Icon: true
      }
    }
  })
  mountedWrappers.push(wrapper)
  return wrapper
}

afterEach(() => {
  mountedWrappers.splice(0).forEach((wrapper) => wrapper.unmount())
})

describe('AutoRefreshButton menu keyboard behavior', () => {
  it('focuses the first item on open and supports roving menu navigation', async () => {
    const wrapper = mountMenu()
    const trigger = wrapper.get('button[aria-haspopup="menu"]')

    await trigger.trigger('click')
    await flushPromises()

    const menu = wrapper.get('[role="menu"]')
    const items = menu.findAll<HTMLElement>('[role^="menuitem"]')
    expect(trigger.attributes('aria-expanded')).toBe('true')
    expect(trigger.attributes('aria-controls')).toBe(menu.attributes('id'))
    expect(menu.attributes('aria-labelledby')).toBe(trigger.attributes('id'))
    expect(items).toHaveLength(4)
    expect(items.every((item) => item.attributes('tabindex') === '-1')).toBe(true)
    expect(document.activeElement).toBe(items[0].element)

    await items[0].trigger('keydown', { key: 'ArrowDown' })
    expect(document.activeElement).toBe(items[1].element)

    await items[1].trigger('keydown', { key: 'End' })
    expect(document.activeElement).toBe(items[3].element)

    await items[3].trigger('keydown', { key: 'Home' })
    expect(document.activeElement).toBe(items[0].element)

    await items[0].trigger('keydown', { key: 'ArrowUp' })
    expect(document.activeElement).toBe(items[3].element)

    await items[3].trigger('keydown', { key: 'Escape' })
    await flushPromises()
    expect(wrapper.find('[role="menu"]').exists()).toBe(false)
    expect(document.activeElement).toBe(trigger.element)
  })

  it('opens on ArrowUp at the last item and restores focus after selection', async () => {
    const wrapper = mountMenu()
    const trigger = wrapper.get('button[aria-haspopup="menu"]')

    await trigger.trigger('keydown', { key: 'ArrowUp' })
    await flushPromises()

    const items = wrapper.findAll<HTMLElement>('[role^="menuitem"]')
    expect(document.activeElement).toBe(items[items.length - 1].element)

    await items[1].trigger('click')
    await flushPromises()
    expect(wrapper.emitted('update:interval')).toEqual([[5]])
    expect(wrapper.find('[role="menu"]').exists()).toBe(false)
    expect(document.activeElement).toBe(trigger.element)
  })
})
