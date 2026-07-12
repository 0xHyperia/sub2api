import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import HomeView from '@/views/HomeView.vue'

const { appState, authState, fetchSettingsMock, checkAuthMock } = vi.hoisted(() => ({
  appState: {
    cachedPublicSettings: null as null | Record<string, unknown>,
    siteName: 'Sub2API',
    publicSettingsLoaded: true,
  },
  authState: {
    isAuthenticated: false,
    isAdmin: false,
  },
  fetchSettingsMock: vi.fn(),
  checkAuthMock: vi.fn(),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => params?.siteName ? `${key}:${params.siteName}` : key,
    }),
  }
})

vi.mock('@/stores', () => ({
  useAuthStore: () => ({ ...authState, checkAuth: checkAuthMock }),
  useAppStore: () => ({ ...appState, fetchPublicSettings: fetchSettingsMock }),
}))

function mountView() {
  return mount(HomeView, {
    global: {
      stubs: {
        RouterLink: {
          props: ['to'],
          template: '<a :href="to"><slot /></a>',
        },
        HomeExperiment: { template: '<div data-testid="default-home" />' },
        Icon: true,
      },
    },
  })
}

describe('HomeView custom content', () => {
  beforeEach(() => {
    appState.cachedPublicSettings = null
    appState.siteName = 'Sub2API'
    appState.publicSettingsLoaded = true
    authState.isAuthenticated = false
    authState.isAdmin = false
    fetchSettingsMock.mockReset()
    checkAuthMock.mockReset()
  })

  it('keeps external content in a constrained iframe with recovery actions', async () => {
    appState.cachedPublicSettings = {
      site_name: 'Example Gateway',
      home_content: 'https://example.com/custom-home',
    }
    const wrapper = mountView()
    const frame = wrapper.get('iframe')

    expect(frame.attributes('title')).toBe('home.customPageTitle:Example Gateway')
    expect(frame.attributes('sandbox')).not.toContain('allow-top-navigation')
    expect(frame.attributes('referrerpolicy')).toBe('strict-origin-when-cross-origin')
    expect(wrapper.get('a[target="_blank"]').attributes('href')).toBe('https://example.com/custom-home')

    await frame.trigger('error')
    expect(wrapper.get('[role="alert"]').text()).toContain('home.customPageLoadFailed')

    await wrapper.get('[role="alert"] button').trigger('click')
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.get('[role="status"]').text()).toContain('home.loadingCustomPage')
  })

  it('sanitizes administrator HTML while preserving the custom page body', () => {
    appState.cachedPublicSettings = {
      home_content: '<p data-safe="yes">Visible</p><img src="x" onerror="alert(1)"><iframe src="https://example.com/embed" sandbox="allow-top-navigation"></iframe><script>alert(1)</script>',
    }
    const wrapper = mountView()

    expect(wrapper.get('[data-safe="yes"]').text()).toBe('Visible')
    expect(wrapper.find('script').exists()).toBe(false)
    expect(wrapper.get('img').attributes('onerror')).toBeUndefined()
    expect(wrapper.get('iframe').attributes('sandbox')).not.toContain('allow-top-navigation')
    expect(wrapper.get('iframe').attributes('referrerpolicy')).toBe('strict-origin-when-cross-origin')
  })

  it('keeps the redesigned default home when no override is configured', () => {
    const wrapper = mountView()
    expect(wrapper.get('[data-testid="default-home"]').exists()).toBe(true)
  })
})
