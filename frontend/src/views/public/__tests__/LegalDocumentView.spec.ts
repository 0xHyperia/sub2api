import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import LegalDocumentView from '../LegalDocumentView.vue'

const { getPublicSettingsMock, routeState } = vi.hoisted(() => ({
  getPublicSettingsMock: vi.fn(),
  routeState: { params: { documentId: 'terms' } },
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return { ...actual, useRoute: () => routeState }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => params?.siteName ? `${key}:${params.siteName}` : key,
    }),
  }
})

vi.mock('@/api/auth', () => ({
  getPublicSettings: (...args: unknown[]) => getPublicSettingsMock(...args),
}))

vi.mock('@/i18n', () => ({ getLocale: () => 'en' }))

function settings() {
  return {
    site_name: 'Example Gateway',
    site_logo: '/brand.png',
    login_agreement_updated_at: '2026-07-12',
    login_agreement_documents: [{ id: 'terms', title: 'Terms of Service', content_md: '# Terms\n\nBe kind.' }],
  }
}

function mountView() {
  return mount(LegalDocumentView, {
    global: {
      stubs: {
        RouterLink: {
          props: ['to'],
          template: '<a :href="to"><slot /></a>',
        },
        Icon: true,
      },
    },
  })
}

describe('LegalDocumentView', () => {
  beforeEach(() => {
    routeState.params = { documentId: 'terms' }
    getPublicSettingsMock.mockReset()
  })

  it('offers an accessible retry after a settings request fails', async () => {
    getPublicSettingsMock
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce(settings())

    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain('legal.loadFailed')

    await wrapper.get('[role="alert"] button').trigger('click')
    await flushPromises()
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.get('article h1').text()).toBe('Terms of Service')
    expect(getPublicSettingsMock).toHaveBeenCalledTimes(2)
  })

  it('uses the configured site name in the logo alternative text', async () => {
    getPublicSettingsMock.mockResolvedValue(settings())
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('header img').attributes('alt')).toBe('legal.siteLogoAlt:Example Gateway')
  })
})
