import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import HelpTooltip from '@/components/common/HelpTooltip.vue'

function getTooltipElement(): HTMLDivElement {
  const tooltip = document.body.querySelector('[role="tooltip"]')
  if (!(tooltip instanceof HTMLDivElement)) {
    throw new Error('tooltip element not found')
  }
  return tooltip
}

describe('HelpTooltip', () => {
  afterEach(() => {
    vi.restoreAllMocks()
    document.body.innerHTML = ''
  })

  it('keeps the existing hover interaction by default', async () => {
    const wrapper = mount(HelpTooltip, {
      attachTo: document.body,
      props: {
        content: 'hover details',
      },
    })

    const trigger = wrapper.get('.group')
    const tooltip = getTooltipElement()
    const triggerButton = trigger.get('button')

    expect(tooltip.style.display).toBe('none')
    expect(triggerButton.attributes('aria-describedby')).toBe(tooltip.id)
    expect(triggerButton.attributes('aria-expanded')).toBe('false')

    await trigger.trigger('mouseenter')
    await nextTick()
    expect(tooltip.style.display).not.toBe('none')
    expect(triggerButton.attributes('aria-expanded')).toBe('true')

    await trigger.trigger('mouseleave')
    await nextTick()
    expect(tooltip.style.display).toBe('none')

    wrapper.unmount()
  })

  it('opens from keyboard focus, closes with Escape, and clamps to the viewport', async () => {
    vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(320)
    vi.spyOn(window, 'innerHeight', 'get').mockReturnValue(720)
    vi.spyOn(window, 'scrollY', 'get').mockReturnValue(500)

    const wrapper = mount(HelpTooltip, {
      attachTo: document.body,
      props: {
        content: 'keyboard details',
        widthClass: 'w-64',
      },
    })

    const trigger = wrapper.get('.group')
    const triggerButton = trigger.get('button')
    const tooltip = getTooltipElement()
    vi.spyOn(trigger.element, 'getBoundingClientRect').mockReturnValue({
      x: 300,
      y: 200,
      left: 300,
      right: 316,
      top: 200,
      bottom: 216,
      width: 16,
      height: 16,
      toJSON: () => ({}),
    })
    vi.spyOn(tooltip, 'getBoundingClientRect').mockReturnValue({
      x: 0,
      y: 0,
      left: 0,
      right: 256,
      top: 0,
      bottom: 80,
      width: 256,
      height: 80,
      toJSON: () => ({}),
    })

    await triggerButton.trigger('focusin')
    await nextTick()

    expect(tooltip.style.display).not.toBe('none')
    expect(tooltip.style.left).toBe('176px')
    expect(tooltip.style.top).toBe('192px')
    expect(tooltip.classList.contains('max-w-[calc(100vw-2rem)]')).toBe(true)

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await nextTick()
    expect(tooltip.style.display).toBe('none')

    wrapper.unmount()
  })

  it('supports click-to-toggle details and closes on outside click', async () => {
    const wrapper = mount(HelpTooltip, {
      attachTo: document.body,
      props: {
        content: 'click details',
        trigger: 'click',
      },
    })

    const trigger = wrapper.get('.group')
    const tooltip = getTooltipElement()

    expect(tooltip.style.display).toBe('none')

    await trigger.trigger('click')
    await nextTick()
    expect(tooltip.style.display).not.toBe('none')
    expect(tooltip.textContent).toContain('click details')

    const closeButton = tooltip.querySelector('button[aria-label="Close"]')
    if (!(closeButton instanceof HTMLButtonElement)) {
      throw new Error('close button not found')
    }
    closeButton.click()
    await nextTick()
    expect(tooltip.style.display).toBe('none')

    await trigger.trigger('click')
    await nextTick()
    expect(tooltip.style.display).not.toBe('none')

    document.body.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await nextTick()
    expect(tooltip.style.display).toBe('none')

    wrapper.unmount()
  })
})
