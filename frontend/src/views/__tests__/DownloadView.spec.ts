import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import DownloadView from '../DownloadView.vue'
import { SOFTWARE_CATALOG, type SoftwareReleaseResult } from '@/api/clientRelease'
import { normalizeSoftwareAssets } from '@/utils/clientRelease'

const { getSoftwareCenterReleasesMock, getSoftwareReleaseMock, checkAuthMock, fetchPublicSettingsMock } = vi.hoisted(() => ({
  getSoftwareCenterReleasesMock: vi.fn(),
  getSoftwareReleaseMock: vi.fn(),
  checkAuthMock: vi.fn(),
  fetchPublicSettingsMock: vi.fn()
}))

vi.mock('@/api/clientRelease', async () => {
  const actual = await vi.importActual<typeof import('@/api/clientRelease')>('@/api/clientRelease')
  return {
    ...actual,
    getSoftwareCenterReleases: getSoftwareCenterReleasesMock,
    getSoftwareRelease: getSoftwareReleaseMock
  }
})

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    cachedPublicSettings: { site_name: 'USA-零' },
    siteName: 'USA-零',
    publicSettingsLoaded: true,
    fetchPublicSettings: fetchPublicSettingsMock
  }),
  useAuthStore: () => ({
    isAuthenticated: false,
    isAdmin: false,
    checkAuth: checkAuthMock
  })
}))

vi.mock('@/composables/useTheme', () => ({
  useTheme: () => ({ resolvedTheme: ref('light'), toggleTheme: vi.fn() })
}))

function releaseResults(): SoftwareReleaseResult[] {
  return SOFTWARE_CATALOG.map((entry) => ({
    entry,
    release: entry.fallbackRelease,
    assets: normalizeSoftwareAssets(entry, entry.fallbackRelease.assets),
    state: 'ready'
  }))
}

function mountView() {
  return mount(DownloadView, {
    global: {
      stubs: {
        RouterLink: {
          props: ['to'],
          template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>'
        }
      }
    }
  })
}

describe('DownloadView software center', () => {
  beforeEach(() => {
    localStorage.clear()
    getSoftwareCenterReleasesMock.mockReset().mockResolvedValue(releaseResults())
    getSoftwareReleaseMock.mockReset().mockImplementation(async (entry) => releaseResults().find((result) => result.entry.id === entry.id))
    checkAuthMock.mockReset()
    fetchPublicSettingsMock.mockReset()
    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue(null)
    Object.defineProperty(window.navigator, 'platform', { configurable: true, value: 'Win32' })
    Object.defineProperty(window.navigator, 'userAgent', { configurable: true, value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)' })
  })

  it('promotes ZeroAgent and renders the three selected applications', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('.featured-product').attributes('data-software-id')).toBe('zeroagent')
    expect(wrapper.get('.featured-product').text()).not.toContain('备用数据')
    expect(wrapper.get('.featured-product').text()).not.toContain('缓存数据')
    expect(wrapper.findAll('.software-card')).toHaveLength(3)
    expect(wrapper.text()).toContain('CC Switch')
    expect(wrapper.text()).toContain('Codex++')
    expect(wrapper.text()).not.toContain('ZeroBox')
    expect(wrapper.get('.hero-actions .primary').attributes('href')).toContain('https://ghfast.top/https://github.com/USA-Zero/ZeroAgent/releases/download/')
    expect(wrapper.get('.webui-button').attributes('href')).toBe('https://agent.usa0.top')
    expect(wrapper.get('.webui-button').attributes('target')).toBe('_blank')
    wrapper.unmount()
  })

  it('filters by name and platform without losing the catalog', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('.search-box input').attributes('type')).toBe('text')
	 expect(wrapper.findAll('.search-clear')).toHaveLength(0)

    await wrapper.get('.search-box input').setValue('Codex++')
    expect(wrapper.findAll('.software-card')).toHaveLength(1)
    expect(wrapper.get('.software-card').attributes('data-software-id')).toBe('codex-plus-plus')

    await wrapper.get('.search-clear').trigger('click')
    expect(wrapper.findAll('.software-card')).toHaveLength(3)
    const androidFilter = wrapper.findAll('.platform-filter button').find((button) => button.text().startsWith('Android'))!
    await androidFilter.trigger('click')
    expect(wrapper.findAll('.software-card')).toHaveLength(1)
    expect(wrapper.get('.software-card').attributes('data-software-id')).toBe('zeroagent')
    wrapper.unmount()
  })

  it('shows accelerated and original links in the package dialog and closes with Escape', async () => {
    const wrapper = mountView()
    await flushPromises()

    const windowsPlatform = wrapper.findAll('.platform-row button').find((button) => button.text().includes('Windows'))!
    await windowsPlatform.trigger('click')
    expect(wrapper.get('[role="dialog"]').text()).toContain('选择 ZeroAgent 安装版本')
    expect(wrapper.findAll('.asset-group')).toHaveLength(1)
    expect(wrapper.get('.asset-group').classes()).toContain('is-current-platform')
    expect(wrapper.get('.asset-row.is-recommended').text()).toContain('推荐')
    await wrapper.findAll('.dialog-platform-tabs button').find((button) => button.text() === '全部')!.trigger('click')
    expect(wrapper.findAll('.asset-group')).toHaveLength(4)
    expect(wrapper.get('.accelerated-link').attributes('href')).toMatch(/^https:\/\/ghfast\.top\/https:\/\/github\.com\//)
    expect(wrapper.get('.original-link').attributes('href')).toMatch(/^https:\/\/github\.com\//)

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await flushPromises()
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('opens sanitized release notes', async () => {
    const results = releaseResults()
    results[0] = {
      ...results[0],
      release: { ...results[0].release, body: '## 安全更新\n\n- 修复登录状态\n- 改进下载体验\n\n<script>alert(1)</script>\n[说明](docs/release.md)\n\n![界面](assets/release.png)' }
    }
    getSoftwareCenterReleasesMock.mockResolvedValue(results)
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('.featured-links button').trigger('click')
    expect(wrapper.get('.release-body').html()).toContain('安全更新')
    expect(wrapper.get('.release-body').html()).not.toContain('<script>')
    expect(wrapper.get('.release-body ul').text()).toContain('改进下载体验')
    expect(wrapper.get('.release-body a').attributes('rel')).toBe('noopener noreferrer')
    expect(wrapper.get('.release-body a').attributes('href')).toBe('https://github.com/USA-Zero/ZeroAgent/blob/v0.3.15/docs/release.md')
    expect(wrapper.get('.release-body img').attributes('src')).toBe('https://raw.githubusercontent.com/USA-Zero/ZeroAgent/v0.3.15/assets/release.png')
    expect(wrapper.get('.dialog-footer a').attributes('href')).toBe('https://github.com/USA-Zero/ZeroAgent/releases/tag/v0.3.15')
    wrapper.unmount()
  })

  it('renders an empty catalog when every application is unpublished', async () => {
    getSoftwareCenterReleasesMock.mockResolvedValue([])
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.get('#software-title').text()).toBe('软件中心')
    expect(wrapper.text()).toContain('暂时没有已发布的软件')
    expect(wrapper.findAll('.software-card')).toHaveLength(0)
    wrapper.unmount()
  })
})
