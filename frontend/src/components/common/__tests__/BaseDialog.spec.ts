import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

import BaseDialog from '../BaseDialog.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))

afterEach(() => {
  document.body.classList.remove('modal-open')
  document.body.innerHTML = ''
})

describe('BaseDialog', () => {
  it('traps focus, closes with Escape, and restores the opener', async () => {
    const opener = document.createElement('button')
    opener.textContent = 'Open dialog'
    document.body.appendChild(opener)
    opener.focus()

    const wrapper = mount(BaseDialog, {
      attachTo: document.body,
      props: {
        show: true,
        title: 'Accessible dialog',
        showCloseButton: false
      },
      slots: {
        default: '<button id="dialog-first">First</button><button id="dialog-last">Last</button>'
      }
    })
    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()

    const first = document.getElementById('dialog-first') as HTMLButtonElement
    const last = document.getElementById('dialog-last') as HTMLButtonElement
    expect(document.body.classList.contains('modal-open')).toBe(true)
    expect(document.activeElement).toBe(first)

    last.focus()
    document.dispatchEvent(new KeyboardEvent('keydown', {
      key: 'Tab',
      bubbles: true,
      cancelable: true
    }))
    expect(document.activeElement).toBe(first)

    first.focus()
    document.dispatchEvent(new KeyboardEvent('keydown', {
      key: 'Tab',
      shiftKey: true,
      bubbles: true,
      cancelable: true
    }))
    expect(document.activeElement).toBe(last)

    document.dispatchEvent(new KeyboardEvent('keydown', {
      key: 'Escape',
      bubbles: true,
      cancelable: true
    }))
    expect(wrapper.emitted('close')).toHaveLength(1)

    await wrapper.setProps({ show: false })
    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()
    expect(document.activeElement).toBe(opener)
    expect(document.body.classList.contains('modal-open')).toBe(false)

    wrapper.unmount()
  })

  it('keeps the body locked and only lets the topmost dialog handle Escape', async () => {
    const first = mount(BaseDialog, {
      attachTo: document.body,
      props: { show: true, title: 'First dialog' }
    })
    const second = mount(BaseDialog, {
      attachTo: document.body,
      props: { show: true, title: 'Second dialog' }
    })
    await first.vm.$nextTick()
    await second.vm.$nextTick()

    document.dispatchEvent(new KeyboardEvent('keydown', {
      key: 'Escape',
      bubbles: true,
      cancelable: true
    }))
    expect(first.emitted('close')).toBeUndefined()
    expect(second.emitted('close')).toHaveLength(1)

    await second.setProps({ show: false })
    await second.vm.$nextTick()
    expect(document.body.classList.contains('modal-open')).toBe(true)

    await first.setProps({ show: false })
    await first.vm.$nextTick()
    expect(document.body.classList.contains('modal-open')).toBe(false)

    second.unmount()
    first.unmount()
  })

  it('resets body scroll position when reopened', async () => {
    const wrapper = mount(BaseDialog, {
      attachTo: document.body,
      props: { show: false, title: 'Details' },
      slots: { default: '<div style="height: 2000px">content</div>' },
      global: { stubs: { Icon: true } }
    })

    await wrapper.setProps({ show: true })
    await nextTick()
    const body = document.body.querySelector<HTMLElement>('.modal-body')
    expect(body).not.toBeNull()
    body!.scrollTop = 480

    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })
    await nextTick()

    expect(document.body.querySelector<HTMLElement>('.modal-body')?.scrollTop).toBe(0)
    wrapper.unmount()
  })
})
