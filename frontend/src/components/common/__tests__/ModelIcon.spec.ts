import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ModelIcon from '../ModelIcon.vue'

describe('ModelIcon', () => {
  it('renders the OpenAI mark in its brand green', () => {
    const wrapper = mount(ModelIcon, {
      props: { model: 'gpt-5.4' }
    })

    expect(wrapper.find('path').attributes('fill')).toBe('#10A37F')
  })

  it('renders the model vendor icon instead of the transport platform', () => {
    const wrapper = mount(ModelIcon, {
      props: { model: 'deepseek-chat' }
    })

    expect(wrapper.find('svg').exists()).toBe(true)
    expect(wrapper.find('path').attributes('fill')).toBe('#4D6BFE')
  })

  it('accepts normalized vendor aliases when no model name is available', () => {
    const wrapper = mount(ModelIcon, {
      props: { vendor: 'anthropic' }
    })

    expect(wrapper.find('svg').exists()).toBe(true)
    expect(wrapper.find('path').attributes('fill')).toBe('#D97706')
  })

  it('falls back to a stable text marker for unknown models', () => {
    const wrapper = mount(ModelIcon, {
      props: { model: 'unknown-model' }
    })

    expect(wrapper.find('.model-icon-fallback').text()).toBe('U')
  })
})
