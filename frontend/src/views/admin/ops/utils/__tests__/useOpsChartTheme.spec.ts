import { defineComponent, nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { useOpsChartTheme } from '../useOpsChartTheme'

const ThemeProbe = defineComponent({
  setup() {
    return useOpsChartTheme()
  },
  template: '<span>{{ isDarkMode ? "dark" : "light" }}</span>',
})

describe('useOpsChartTheme', () => {
  afterEach(() => {
    document.documentElement.classList.remove('dark')
  })

  it('tracks theme class changes while mounted', async () => {
    const wrapper = mount(ThemeProbe)
    expect(wrapper.text()).toBe('light')

    document.documentElement.classList.add('dark')
    await nextTick()
    await nextTick()

    expect(wrapper.text()).toBe('dark')
    wrapper.unmount()
  })
})
