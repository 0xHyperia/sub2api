import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ContactUsPanel from '../ContactUsPanel.vue'
import type { PublicSettings } from '@/types'

const push = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push })
}))

function settings(overrides: Partial<PublicSettings> = {}): PublicSettings {
  return {
    contact_us_enabled: true,
    contact_qq_enabled: true,
    contact_qq_name: '用户交流群',
    contact_qq_url: 'https://qm.qq.com/example',
    contact_telegram_enabled: true,
    contact_telegram_name: 'Telegram 公告群',
    contact_telegram_url: 'https://t.me/example',
    contact_ticket_enabled: true,
    ...overrides
  } as PublicSettings
}

function mountPanel(overrides: Partial<PublicSettings> = {}) {
  return mount(ContactUsPanel, {
    props: { settings: settings(overrides), collapsed: false, locale: 'zh-CN' },
    global: { stubs: { Teleport: true } }
  })
}

afterEach(() => {
  push.mockReset()
  document.body.innerHTML = ''
})

describe('ContactUsPanel', () => {
  it('stays hidden when the master switch is disabled', () => {
    const wrapper = mountPanel({ contact_us_enabled: false })
    expect(wrapper.find('button.contact-trigger').exists()).toBe(false)
  })

  it('shows only enabled actions with safe external-link attributes', async () => {
    const wrapper = mountPanel({
      contact_qq_url: 'javascript:alert(1)',
      contact_ticket_enabled: false
    })
    await wrapper.get('button.contact-trigger').trigger('click')

    const links = wrapper.findAll('a.contact-action')
    expect(links).toHaveLength(1)
    expect(links[0].attributes('href')).toBe('https://t.me/example')
    expect(links[0].attributes('target')).toBe('_blank')
    expect(links[0].attributes('rel')).toBe('noopener noreferrer')
    expect(wrapper.text()).not.toContain('用户交流群')
  })

  it('routes the ticket action to the existing support page', async () => {
    const wrapper = mountPanel({ contact_qq_enabled: false, contact_telegram_enabled: false })
    await wrapper.get('button.contact-trigger').trigger('click')
    await wrapper.get('button.contact-action').trigger('click')

    expect(push).toHaveBeenCalledWith('/support')
  })

  it('notifies the shell to close its mobile drawer before opening', async () => {
    const wrapper = mountPanel()
    await wrapper.get('button.contact-trigger').trigger('click')
    expect(wrapper.emitted('open')).toHaveLength(1)
  })

  it('uses a named collapsed trigger and responsive action layout', () => {
    const wrapper = mount(ContactUsPanel, {
      props: { settings: settings(), collapsed: true, locale: 'en-US' },
      global: { stubs: { Teleport: true } }
    })
    const trigger = wrapper.get('button.contact-trigger')
    expect(trigger.attributes('aria-label')).toBe('Contact us')
    expect(trigger.classes()).toContain('contact-trigger-collapsed')
  })
})
