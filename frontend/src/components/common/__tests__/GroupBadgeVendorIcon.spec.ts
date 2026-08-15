import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import GroupBadge from '../GroupBadge.vue'
import ModelIcon from '../ModelIcon.vue'
import PlatformIcon from '../PlatformIcon.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ cachedPublicSettings: null })
}))

describe('GroupBadge vendor icon', () => {
  it('uses the sole model vendor for a single-vendor group', () => {
    const wrapper = mount(GroupBadge, {
      props: {
        name: 'DeepSeek group',
        platform: 'openai',
        modelNames: ['deepseek-chat', 'deepseek-reasoner']
      }
    })

    expect(wrapper.findComponent(ModelIcon).props('vendor')).toBe('deepseek')
    expect(wrapper.findComponent(PlatformIcon).exists()).toBe(false)
  })

  it('uses the composite icon for a multi-vendor group', () => {
    const wrapper = mount(GroupBadge, {
      props: {
        name: 'Smart group',
        platform: 'openai',
        modelNames: ['gpt-5.4', 'claude-opus-4-6']
      }
    })

    expect(wrapper.findComponent(ModelIcon).exists()).toBe(false)
    expect(wrapper.findComponent(PlatformIcon).props('platform')).toBe('composite')
  })

  it('falls back to the platform when model vendors are unknown', () => {
    const wrapper = mount(GroupBadge, {
      props: {
        name: 'Custom group',
        platform: 'gemini',
        modelNames: ['private-model']
      }
    })

    expect(wrapper.findComponent(PlatformIcon).props('platform')).toBe('gemini')
  })
})
