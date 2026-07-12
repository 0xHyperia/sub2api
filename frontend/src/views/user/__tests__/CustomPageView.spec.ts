import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import CustomPageView from '../CustomPageView.vue'

const { appState, authState, routeState, fetchSettingsMock } = vi.hoisted(() => ({
  appState: {
    publicSettingsLoaded: true,
    cachedPublicSettings: null as null | Record<string, unknown>,
  },
  authState: {
    isAdmin: false,
    token: 'test-token',
    user: { id: 7 },
  },
  routeState: { params: { id: 'docs' } },
  fetchSettingsMock: vi.fn(),
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
      locale: { value: 'en' },
      t: (key: string, params?: Record<string, unknown>) => params?.title ? `${key}:${params.title}` : key,
    }),
  }
})

vi.mock('@/stores', () => ({
  useAppStore: () => ({ ...appState, fetchPublicSettings: fetchSettingsMock }),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authState,
}))

vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => ({ customMenuItems: [] }),
}))

function response(ok: boolean, status: number, body = ''): Response {
  return {
    ok,
    status,
    text: vi.fn().mockResolvedValue(body),
  } as unknown as Response
}

function mountView() {
  return mount(CustomPageView, {
    global: {
      stubs: {
        AppLayout: { template: '<main><slot /></main>' },
        Icon: true,
      },
    },
  })
}

describe('CustomPageView', () => {
  const originalInnerWidth = window.innerWidth

  beforeEach(() => {
    routeState.params = { id: 'docs' }
    appState.publicSettingsLoaded = true
    appState.cachedPublicSettings = {
      custom_menu_items: [{
        id: 'docs',
        label: 'Documentation',
        icon_svg: '',
        url: 'md:guide',
        page_slug: 'guide',
        visibility: 'user',
        sort_order: 1,
      }],
    }
    fetchSettingsMock.mockReset()
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 390 })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: originalInnerWidth })
  })

  it('anchors the mobile contents drawer to the content surface', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response(true, 200, '# Overview\n\n```ts\nconst ready = true\n```')))
    const wrapper = mountView()
    await flushPromises()

    const content = wrapper.get('.markdown-page-content')
    expect(content.element.parentElement?.classList.contains('relative')).toBe(true)

    const toggle = wrapper.get('.toc-toggle-btn')
    expect(toggle.attributes('aria-controls')).toBe('custom-page-toc')
    await toggle.trigger('click')
    expect(wrapper.get('#custom-page-toc').isVisible()).toBe(true)
    expect(wrapper.get('.toc-backdrop').isVisible()).toBe(true)

    const copyButton = wrapper.get('.copy-btn')
    expect(copyButton.attributes('type')).toBe('button')
    expect(copyButton.attributes('aria-label')).toBe('customPage.copyCode')
  })

  it('shows a retryable markdown error instead of injecting hard-coded HTML', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(response(false, 503))
      .mockResolvedValueOnce(response(true, 200, '# Restored'))
    vi.stubGlobal('fetch', fetchMock)

    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain('customPage.contentLoadFailedTitle')

    await wrapper.get('[role="alert"] button').trigger('click')
    await flushPromises()
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.get('.markdown-page-content h1').text()).toBe('Restored')
    expect(fetchMock).toHaveBeenCalledTimes(2)
  })

  it('gives embedded custom pages an accessible title', async () => {
    appState.cachedPublicSettings = {
      custom_menu_items: [{
        id: 'docs',
        label: 'Partner portal',
        icon_svg: '',
        url: 'https://example.com/portal',
        visibility: 'user',
        sort_order: 1,
      }],
    }
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('iframe').attributes('title')).toBe('customPage.embedTitle:Partner portal')
    expect(wrapper.get('iframe').attributes('referrerpolicy')).toBe('strict-origin-when-cross-origin')
  })
})
