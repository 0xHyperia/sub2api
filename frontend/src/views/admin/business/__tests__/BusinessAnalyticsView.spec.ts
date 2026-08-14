import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const api = vi.hoisted(() => ({
  getBusinessAnalytics: vi.fn(),
  getBusinessBalance: vi.fn(),
  lookupAgents: vi.fn(),
}))
const showError = vi.hoisted(() => vi.fn())
const route = vi.hoisted(() => ({ query: {} as Record<string, string> }))
const routeProxy = vi.hoisted(() => ({ value: null as null | { query: Record<string, string> } }))
const replace = vi.hoisted(() => vi.fn().mockResolvedValue(undefined))

vi.mock('@/api/businessAnalytics', () => ({
  getBusinessAnalytics: api.getBusinessAnalytics,
  getBusinessBalance: api.getBusinessBalance,
}))
vi.mock('@/api/admin/distribution', () => ({ lookupAgents: api.lookupAgents }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError }) }))
vi.mock('vue-router', async () => {
  const { reactive } = await import('vue')
  routeProxy.value = reactive(route)
  return {
    useRoute: () => routeProxy.value,
    useRouter: () => ({ replace, push: vi.fn() }),
  }
})
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key.split('.').at(-1) || key }) }
})

import BusinessAnalyticsView from '../BusinessAnalyticsView.vue'

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => { resolve = res; reject = rej })
  return { promise, resolve, reject }
}

function snapshot(section: string, value = 1) {
  return {
    date_from: '2026-08-14', date_to: '2026-08-14', previous_date_from: '2026-08-13', previous_date_to: '2026-08-13',
    timezone: 'Asia/Shanghai', granularity: 'hour', currency: 'CNY', updated_at: '2026-08-14T08:00:00Z', estimated: false,
    metrics: { new_users: { value, previous: 0 }, net_paid: { value, previous: 0 } },
    trend: [], previous_trend: [], funnel: [], channels: [], registration_cohorts: [], activation_cohorts: [],
    lifecycle: { new: 0, unactivated: 0, newly_activated: 0, continuously_active: 0, silent_reactivated: 0, churned_reactivated: 0, silent: 0, churned: 0 },
    duration_distribution: [], currency_breakdown: [], warnings: [], section,
  }
}

function balance(window: 1 | 7 | 30, amount: number) {
  return {
    as_of: '2026-08-14T08:00:00Z', activity_window_days: window, balance_recharge_multiplier: 1,
    available_balance_usd: amount, frozen_balance_usd: 0, positive_balance_users: 1, total_users: 1,
    average_balance_usd: amount, low_balance_enabled: false, low_balance_users: 0,
    segments: [{ key: 'active', users: 1, balance_usd: amount, frozen_usd: 0 }],
  }
}

function mountView(section: 'overview' | 'finance' = 'overview') {
  return mount(BusinessAnalyticsView, {
    props: { section },
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: true,
        MetricInfo: true,
        BusinessAnalyticsChart: true,
        BusinessChannelTable: true,
        RemoteEntityCombobox: true,
        DistributionAnalyticsRange: true,
        RouterLink: { template: '<a><slot /></a>' },
      },
    },
  })
}

describe('BusinessAnalyticsView request state', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    route.query = {}
    api.lookupAgents.mockResolvedValue([])
    api.getBusinessAnalytics.mockResolvedValue(snapshot('overview'))
    api.getBusinessBalance.mockResolvedValue(balance(7, 7))
  })

  it('does not let a late response from the previous section replace the current section', async () => {
    const oldRequest = deferred<ReturnType<typeof snapshot>>()
    api.getBusinessAnalytics
      .mockReturnValueOnce(oldRequest.promise)
      .mockResolvedValueOnce(snapshot('finance', 22))
    const wrapper = mountView('overview')
    await wrapper.setProps({ section: 'finance' })
    await flushPromises()
    oldRequest.resolve(snapshot('overview', 99))
    await flushPromises()

    expect(api.getBusinessAnalytics).toHaveBeenLastCalledWith('finance', expect.any(Object))
    expect(wrapper.text()).toContain('￥22.00')
    expect(wrapper.text()).not.toContain('￥99.00')
  })

  it('accepts the backend null warnings shape without crashing', async () => {
    api.getBusinessAnalytics.mockResolvedValue({ ...snapshot('overview'), warnings: null })
    const wrapper = mountView('overview')
    await flushPromises()

    expect(wrapper.find('.kpis').exists()).toBe(true)
    expect(wrapper.find('.coverage-warning').exists()).toBe(false)
  })

  it('shows recoverable usage coverage warnings', async () => {
    api.getBusinessAnalytics.mockResolvedValue({
      ...snapshot('overview'),
      usage_data_from: '2026-05-01',
      warnings: ['usage_history_incomplete'],
    })
    const wrapper = mountView('overview')
    await flushPromises()

    expect(wrapper.find('.coverage-warning').text()).toContain('usageHistoryIncomplete')
  })

  it('clears a report snapshot as soon as the selected range changes', async () => {
    const wrapper = mountView('overview')
    await flushPromises()
    expect(wrapper.text()).toContain('￥1.00')

    const next = deferred<ReturnType<typeof snapshot>>()
    api.getBusinessAnalytics.mockReturnValueOnce(next.promise)
    const range = wrapper.findComponent({ name: 'DistributionAnalyticsRange' })
    range.vm.$emit('update:modelValue', {
      granularity: 'custom', preset: 'custom', date_from: '2026-08-01', date_to: '2026-08-07',
    })
    range.vm.$emit('change')
    await flushPromises()

    expect(wrapper.text()).not.toContain('￥1.00')
    next.resolve(snapshot('overview', 7))
    await flushPromises()
    expect(wrapper.text()).toContain('￥7.00')
  })

  it('clears stale balance data and exposes an inline retry when a new window fails', async () => {
    route.query = { tab: 'balance' }
    api.getBusinessAnalytics.mockResolvedValue(snapshot('finance'))
    api.getBusinessBalance.mockResolvedValueOnce(balance(7, 7))
    const wrapper = mountView('finance')
    await flushPromises()
    expect(wrapper.text()).toContain('$7.00')

    api.getBusinessBalance.mockRejectedValueOnce(new Error('balance unavailable'))
    const thirty = wrapper.findAll('button').find(button => button.text() === 'balanceWindow30')
    expect(thirty).toBeDefined()
    await thirty!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).not.toContain('$7.00')
    expect(wrapper.find('[role="alert"]').text()).toContain('balance unavailable')
    expect(wrapper.find('[role="alert"]').text()).toContain('retry')
  })

  it('keeps only the newest balance response during rapid window switches', async () => {
    route.query = { tab: 'balance' }
    api.getBusinessAnalytics.mockResolvedValue(snapshot('finance'))
    api.getBusinessBalance.mockResolvedValueOnce(balance(7, 7))
    const wrapper = mountView('finance')
    await flushPromises()

    const first = deferred<ReturnType<typeof balance>>()
    const second = deferred<ReturnType<typeof balance>>()
    api.getBusinessBalance.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise)
    await wrapper.findAll('button').find(button => button.text() === 'balanceWindow1')!.trigger('click')
    await wrapper.findAll('button').find(button => button.text() === 'balanceWindow30')!.trigger('click')
    second.resolve(balance(30, 30))
    await flushPromises()
    first.resolve(balance(1, 1))
    await flushPromises()

    expect(wrapper.text()).toContain('$30.00')
    expect(wrapper.text()).not.toContain('$1.00')
  })

  it('restores date and channel filters when browser history changes the URL', async () => {
    route.query = { date_from: '2026-08-01', date_to: '2026-08-07', channel: 'unknown' }
    const wrapper = mountView('overview')
    await flushPromises()
    api.getBusinessAnalytics.mockClear()

    routeProxy.value!.query = {
      date_from: '2026-07-01',
      date_to: '2026-07-31',
      channel: 'distribution',
    }
    await flushPromises()

    expect(api.getBusinessAnalytics).toHaveBeenCalledWith('overview', expect.objectContaining({
      date_from: '2026-07-01',
      date_to: '2026-07-31',
      channel: 'distribution',
    }))
    expect(wrapper.find('select.channel').element.value).toBe('distribution')
  })
})
