import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import Toast from '../Toast.vue'

const store = vi.hoisted(() => ({
  hideToast: vi.fn(),
  toasts: [
    {
      id: 'toast-1',
      type: 'success',
      message: 'Saved',
      duration: 3000
    }
  ]
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => store
}))

describe('Toast mobile layout', () => {
  it('uses viewport gutters instead of a fixed 320px minimum on narrow screens', () => {
    const wrapper = mount(Toast, {
      global: {
        stubs: {
          Icon: true,
          Teleport: true
        }
      }
    })

    const viewport = wrapper.get('[aria-live="polite"]')
    expect(viewport.classes()).toEqual(expect.arrayContaining(['inset-x-4', 'sm:left-auto']))

    const toast = wrapper.get('.pointer-events-auto')
    expect(toast.classes()).toEqual(
      expect.arrayContaining(['w-full', 'min-w-0', 'sm:min-w-[320px]'])
    )
    expect(toast.classes()).not.toContain('min-w-[320px]')
    wrapper.unmount()
  })
})
