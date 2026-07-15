import { defineComponent, nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { useOpsChartTheme } from '../useOpsChartTheme'

const ThemeProbe = defineComponent({
  setup() {
    return useOpsChartTheme()
  },
  template: '<span data-testid="mode">{{ isDarkMode ? "dark" : "light" }}</span><span data-testid="surface">{{ chartTheme.surface }}</span>',
})

describe('useOpsChartTheme', () => {
  afterEach(() => {
    document.documentElement.classList.remove('dark')
  })

  it('tracks theme class changes while mounted', async () => {
    const wrapper = mount(ThemeProbe)
    expect(wrapper.get('[data-testid="mode"]').text()).toBe('light')
    const lightSurface = wrapper.get('[data-testid="surface"]').text()

    document.documentElement.classList.add('dark')
    await nextTick()
    await nextTick()

    expect(wrapper.get('[data-testid="mode"]').text()).toBe('dark')
    expect(wrapper.get('[data-testid="surface"]').text()).not.toBe(lightSurface)
    wrapper.unmount()
  })
})
