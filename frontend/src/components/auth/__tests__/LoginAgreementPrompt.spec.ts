import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import LoginAgreementPrompt from '@/components/auth/LoginAgreementPrompt.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

const documents = [
  { id: 'terms', title: 'Terms of Service', content_md: 'Terms' },
  { id: 'privacy', title: 'Privacy Policy', content_md: 'Privacy' },
]

describe('LoginAgreementPrompt', () => {
  it('uses modal dialog semantics, closes with Escape, and restores focus', async () => {
    const opener = document.createElement('button')
    document.body.appendChild(opener)
    opener.focus()

    const wrapper = mount(LoginAgreementPrompt, {
      attachTo: document.body,
      props: {
        accepted: false,
        documents,
        mode: 'modal',
        visible: true,
      },
      global: {
        stubs: {
          RouterLink: {
            props: ['to'],
            template: '<a href="#"><slot /></a>',
          },
        },
      },
    })
    await wrapper.vm.$nextTick()

    const dialog = document.querySelector<HTMLElement>('[aria-labelledby="login-agreement-title"]')
    expect(dialog?.getAttribute('role')).toBe('dialog')
    expect(dialog?.getAttribute('aria-modal')).toBe('true')

    dialog?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await wrapper.vm.$nextTick()
    expect(wrapper.emitted('reject')).toHaveLength(1)

    await wrapper.setProps({ visible: false })
    await wrapper.vm.$nextTick()
    expect(document.activeElement).toBe(opener)

    wrapper.unmount()
    opener.remove()
  })
})
