import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const getDashboard = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    getDashboard,
  },
  default: {
    getDashboard,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

import AdminPaymentDashboardView from '../AdminPaymentDashboardView.vue'

const dashboardStats = {
  today_amount: 0,
  total_amount: 0,
  today_count: 0,
  total_count: 0,
  avg_amount: 0,
  daily_series: [],
  payment_methods: [],
  top_users: [],
}

function mountView() {
  return mount(AdminPaymentDashboardView, {
    global: {
      stubs: {
        AppLayout: { template: '<main><slot /></main>' },
        LoadingSpinner: true,
        Icon: true,
        OrderStatsCards: true,
        DailyRevenueChart: true,
      },
    },
  })
}

describe('AdminPaymentDashboardView', () => {
  beforeEach(() => {
    getDashboard.mockReset()
    showError.mockReset()
  })

  it('shows an inline error and can recover after retry', async () => {
    getDashboard.mockRejectedValueOnce(new Error('unavailable'))
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('payment.admin.dashboardLoadFailed')
    expect(wrapper.text()).toContain('common.retry')

    getDashboard.mockResolvedValueOnce({ data: dashboardStats })
    const retryButton = wrapper.findAll('button').find(button => button.text().includes('common.retry'))
    expect(retryButton).toBeDefined()
    await retryButton!.trigger('click')
    await flushPromises()

    expect(getDashboard).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).not.toContain('payment.admin.dashboardLoadFailed')
    expect(wrapper.findComponent({ name: 'OrderStatsCards' }).exists()).toBe(true)
  })
})
