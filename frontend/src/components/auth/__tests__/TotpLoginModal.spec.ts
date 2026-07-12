import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import TotpLoginModal from '@/components/auth/TotpLoginModal.vue'

const { showErrorMock } = vi.hoisted(() => ({
  showErrorMock: vi.fn(),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError: (...args: any[]) => showErrorMock(...args),
  }),
}))

describe('TotpLoginModal', () => {
  beforeEach(() => {
    showErrorMock.mockReset()
  })

  it('sends verification errors to toast and does not render inline red text', async () => {
    const wrapper = mount(TotpLoginModal, {
      props: {
        tempToken: 'temp-token',
        userEmailMasked: 'u***@example.com',
      },
    })

    ;(wrapper.vm as unknown as { setError: (message: string) => void }).setError('Invalid code')
    await wrapper.vm.$nextTick()

    expect(showErrorMock).toHaveBeenCalledWith('Invalid code')
    expect(wrapper.text()).not.toContain('Invalid code')
    expect(wrapper.find('.bg-red-50').exists()).toBe(false)
  })

  it('exposes dialog semantics, supports Escape, and restores focus on close', async () => {
    const opener = document.createElement('button')
    document.body.appendChild(opener)
    opener.focus()

    const wrapper = mount(TotpLoginModal, {
      attachTo: document.body,
      props: {
        tempToken: 'temp-token',
        userEmailMasked: 'u***@example.com',
      },
    })
    await wrapper.vm.$nextTick()

    const dialog = wrapper.get('[role="dialog"]')
    expect(dialog.attributes('aria-modal')).toBe('true')
    expect(dialog.attributes('aria-labelledby')).toBe('totp-login-title')
    expect(document.activeElement).toBe(wrapper.get('input[aria-label]').element)

    await dialog.trigger('keydown', { key: 'Escape' })
    expect(wrapper.emitted('cancel')).toHaveLength(1)

    wrapper.unmount()
    expect(document.activeElement).toBe(opener)
    opener.remove()
  })

  it('keeps keyboard focus inside the dialog', async () => {
    const wrapper = mount(TotpLoginModal, {
      attachTo: document.body,
      props: {
        tempToken: 'temp-token',
      },
    })
    await wrapper.vm.$nextTick()

    const cancelButton = wrapper.get('button')
    cancelButton.element.focus()
    await wrapper.get('[role="dialog"]').trigger('keydown', { key: 'Tab' })

    expect(document.activeElement).toBe(wrapper.get('input[aria-label]').element)
    wrapper.unmount()
  })
})
