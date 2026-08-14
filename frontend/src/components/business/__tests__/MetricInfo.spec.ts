import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import MetricInfo from '../MetricInfo.vue'

vi.mock('@/components/icons/Icon.vue', async () => {
  const { defineComponent, h } = await import('vue')
  return { default: defineComponent({ setup: () => () => h('span', { class: 'icon' }) }) }
})

describe('MetricInfo', () => {
  afterEach(() => {
    document.body.querySelectorAll('.metric-tooltip').forEach(node => node.remove())
  })

  it('renders an inline trigger and teleports the tooltip outside clipped cards', async () => {
    const wrapper = mount(MetricInfo, {
      attachTo: document.body,
      props: { label: '新增用户', description: '周期内完成注册的用户数。' },
    })

    await wrapper.get('button').trigger('click')

    const tooltip = document.body.querySelector<HTMLElement>('.metric-tooltip')
    expect(wrapper.get('.metric-info').classes()).toContain('metric-info')
    expect(wrapper.find('.metric-tooltip').exists()).toBe(false)
    expect(tooltip?.textContent).toContain('周期内完成注册的用户数。')
    expect(wrapper.get('button').attributes('aria-describedby')).toBe(tooltip?.id)

    wrapper.unmount()
  })

  it('closes with Escape and restores focus to the trigger', async () => {
    const wrapper = mount(MetricInfo, {
      attachTo: document.body,
      props: { label: '活跃率', description: '活跃用户占比。' },
    })
    const button = wrapper.get('button')

    await button.trigger('click')
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await wrapper.vm.$nextTick()

    expect(wrapper.get('button').attributes('aria-expanded')).toBe('false')
    expect(document.activeElement).toBe(button.element)
    wrapper.unmount()
  })

  it('only lets the open tooltip claim focus when Escape is pressed', async () => {
    const first = mount(MetricInfo, {
      attachTo: document.body,
      props: { label: '净实付', description: '当前周期净现金流。' },
    })
    const second = mount(MetricInfo, {
      attachTo: document.body,
      props: { label: '退款率', description: '退款金额占比。' },
    })

    await first.get('button').trigger('click')
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await first.vm.$nextTick()

    expect(document.activeElement).toBe(first.get('button').element)
    expect(second.get('button').attributes('aria-expanded')).toBe('false')
    first.unmount()
    second.unmount()
  })
})
