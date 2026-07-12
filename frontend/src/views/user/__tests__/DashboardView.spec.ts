import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import DashboardView from '@/views/user/DashboardView.vue'

const mocks = vi.hoisted(() => ({
  getDashboardStats: vi.fn(),
  getDashboardTrend: vi.fn(),
  getDashboardModels: vi.fn(),
  getByDateRange: vi.fn(),
  getMyPlatformQuotas: vi.fn(),
  refreshUser: vi.fn(),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: { balance: 12 },
    isSimpleMode: false,
    refreshUser: mocks.refreshUser,
  }),
}))

vi.mock('@/api/usage', () => ({
  usageAPI: {
    getDashboardStats: mocks.getDashboardStats,
    getDashboardTrend: mocks.getDashboardTrend,
    getDashboardModels: mocks.getDashboardModels,
    getByDateRange: mocks.getByDateRange,
  },
}))

vi.mock('@/api/user', () => ({
  getMyPlatformQuotas: mocks.getMyPlatformQuotas,
}))

describe('DashboardView load state', () => {
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    mocks.getDashboardStats.mockReset()
    mocks.getDashboardTrend.mockReset().mockResolvedValue({ trend: [] })
    mocks.getDashboardModels.mockReset().mockResolvedValue({ models: [] })
    mocks.getByDateRange.mockReset().mockResolvedValue({ items: [] })
    mocks.getMyPlatformQuotas.mockReset().mockResolvedValue({ platform_quotas: [] })
    mocks.refreshUser.mockReset().mockResolvedValue(undefined)
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined)
  })

  afterEach(() => {
    consoleErrorSpy.mockRestore()
  })

  it('shows an inline retry state and recovers after the dashboard request fails', async () => {
    mocks.getDashboardStats
      .mockRejectedValueOnce(new Error('dashboard unavailable'))
      .mockResolvedValueOnce({ total_api_keys: 1, active_api_keys: 1 })

    const wrapper = mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
          LoadingSpinner: true,
          UserDashboardStats: { template: '<div data-testid="dashboard-stats" />' },
          UserDashboardCharts: true,
          UserDashboardRecentUsage: true,
          UserDashboardQuickActions: true,
        },
      },
    })

    await flushPromises()

    const errorState = wrapper.get('[data-testid="dashboard-load-error"]')
    expect(errorState.text()).toContain('dashboard.loadFailed')
    expect(wrapper.find('[data-testid="dashboard-stats"]').exists()).toBe(false)

    await errorState.get('button').trigger('click')
    await flushPromises()

    expect(mocks.getDashboardStats).toHaveBeenCalledTimes(2)
    expect(wrapper.find('[data-testid="dashboard-load-error"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="dashboard-stats"]').exists()).toBe(true)
  })

  it('keeps loaded stats visible when a later refresh fails', async () => {
    mocks.getDashboardStats
      .mockResolvedValueOnce({ total_api_keys: 1, active_api_keys: 1 })
      .mockRejectedValueOnce(new Error('refresh unavailable'))

    const wrapper = mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
          LoadingSpinner: true,
          UserDashboardStats: { template: '<div data-testid="dashboard-stats" />' },
          UserDashboardCharts: true,
          UserDashboardRecentUsage: true,
          UserDashboardQuickActions: true,
        },
      },
    })

    await flushPromises()
    expect(wrapper.find('[data-testid="dashboard-stats"]').exists()).toBe(true)

    await wrapper.get('header button').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="dashboard-stats"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="dashboard-refresh-error"]').text()).toContain('dashboard.loadFailed')
  })

  it('passes chart and recent-usage failures to their local recovery surfaces', async () => {
    mocks.getDashboardStats.mockResolvedValue({ total_api_keys: 1, active_api_keys: 1 })
    mocks.getDashboardTrend.mockRejectedValueOnce(new Error('charts unavailable'))
    mocks.getByDateRange.mockRejectedValueOnce(new Error('recent unavailable'))

    const wrapper = mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
          LoadingSpinner: true,
          UserDashboardStats: true,
          UserDashboardCharts: {
            props: ['error'],
            template: '<div data-testid="charts-error">{{ error }}</div>',
          },
          UserDashboardRecentUsage: {
            props: ['error'],
            template: '<div data-testid="recent-error">{{ error }}</div>',
          },
          UserDashboardQuickActions: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.get('[data-testid="charts-error"]').text()).toBe('true')
    expect(wrapper.get('[data-testid="recent-error"]').text()).toBe('true')
  })
})
