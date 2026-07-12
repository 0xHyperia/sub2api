import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import { useDropdownMenu } from '@/composables/useDropdownMenu'

const DropdownHarness = defineComponent({
  setup() {
    return useDropdownMenu('test-menu')
  },
  template: `
    <div>
      <button
        :id="triggerId"
        ref="triggerRef"
        type="button"
        aria-haspopup="menu"
        :aria-expanded="open"
        :aria-controls="open ? menuId : undefined"
        @click="toggleMenu"
        @keydown="handleTriggerKeydown"
      >
        Open
      </button>
      <div
        v-if="open"
        :id="menuId"
        ref="menuRef"
        role="menu"
        :aria-labelledby="triggerId"
        @keydown="handleMenuKeydown"
      >
        <button type="button" role="menuitem">First</button>
        <button type="button" role="menuitem" disabled>Disabled</button>
        <button type="button" role="menuitemcheckbox" aria-checked="false">Second</button>
        <button type="button" role="menuitemradio" aria-checked="false">Last</button>
      </div>
    </div>
  `,
})

const menuItems = (wrapper: ReturnType<typeof mount>) =>
  wrapper.findAll<HTMLElement>('[role^="menuitem"]:not([disabled])')

describe('useDropdownMenu', () => {
  it('opens from ArrowDown and supports wrapped arrow, Home, and End navigation', async () => {
    const wrapper = mount(DropdownHarness, { attachTo: document.body })
    const trigger = wrapper.get('button[aria-haspopup="menu"]')

    await trigger.trigger('keydown', { key: 'ArrowDown' })
    const items = menuItems(wrapper)
    expect(document.activeElement).toBe(items[0].element)

    await items[0].trigger('keydown', { key: 'ArrowUp' })
    expect(document.activeElement).toBe(items[2].element)

    await items[2].trigger('keydown', { key: 'Home' })
    expect(document.activeElement).toBe(items[0].element)

    await items[0].trigger('keydown', { key: 'End' })
    expect(document.activeElement).toBe(items[2].element)

    await items[2].trigger('keydown', { key: 'ArrowDown' })
    expect(document.activeElement).toBe(items[0].element)
    wrapper.unmount()
  })

  it('opens on the last item from ArrowUp and restores trigger focus on Escape', async () => {
    const wrapper = mount(DropdownHarness, { attachTo: document.body })
    const trigger = wrapper.get('button[aria-haspopup="menu"]')

    await trigger.trigger('keydown', { key: 'ArrowUp' })
    const items = menuItems(wrapper)
    expect(document.activeElement).toBe(items[2].element)

    await items[2].trigger('keydown', { key: 'Escape' })
    expect(wrapper.find('[role="menu"]').exists()).toBe(false)
    expect(document.activeElement).toBe(trigger.element)
    wrapper.unmount()
  })

  it('closes on Tab without moving focus back to the trigger', async () => {
    const wrapper = mount(DropdownHarness, { attachTo: document.body })
    const trigger = wrapper.get('button[aria-haspopup="menu"]')

    await trigger.trigger('click')
    const firstItem = menuItems(wrapper)[0]
    expect(document.activeElement).toBe(firstItem.element)

    await firstItem.trigger('keydown', { key: 'Tab' })
    expect(wrapper.find('[role="menu"]').exists()).toBe(false)
    expect(document.activeElement).not.toBe(trigger.element)
    wrapper.unmount()
  })
})
