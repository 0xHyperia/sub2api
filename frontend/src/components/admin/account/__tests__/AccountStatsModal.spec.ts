import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import AccountStatsModal from '../AccountStatsModal.vue'

const { getStats } = vi.hoisted(() => ({
  getStats: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      getStats
    }
  }
}))

vi.mock('@/composables/useTheme', () => ({
  useTheme: () => ({ isDark: { value: false } })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

interface Deferred<T> {
  promise: Promise<T>
  resolve: (value: T) => void
}

function deferred<T>(): Deferred<T> {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise
  })
  return { promise, resolve }
}

function account(id: number, name = `Account ${id}`) {
  return {
    id,
    name,
    platform: 'openai',
    type: 'oauth',
    status: 'active'
  } as any
}

function stats(totalCost: number) {
  return {
    history: [],
    models: [],
    endpoints: [],
    upstream_endpoints: [],
    summary: {
      days: 30,
      actual_days_used: 1,
      total_cost: totalCost,
      total_user_cost: totalCost,
      total_standard_cost: totalCost,
      total_requests: 10,
      total_tokens: 100,
      avg_daily_cost: totalCost,
      avg_daily_user_cost: totalCost,
      avg_daily_requests: 10,
      avg_daily_tokens: 100,
      avg_duration_ms: 120,
      today: null,
      highest_cost_day: null,
      highest_request_day: null
    }
  }
}

function mountModal(selectedAccount = account(1)) {
  return mount(AccountStatsModal, {
    props: {
      show: true,
      account: selectedAccount
    },
    global: {
      stubs: {
        BaseDialog: {
          props: ['show'],
          template: '<div v-if="show"><slot /><slot name="footer" /></div>'
        },
        LoadingSpinner: true,
        ModelDistributionChart: true,
        EndpointDistributionChart: true,
        Line: true,
        Icon: true
      }
    }
  })
}

describe('AccountStatsModal async state', () => {
  beforeEach(() => {
    getStats.mockReset()
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('shows a persistent load error and retries the current account', async () => {
    getStats
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce(stats(42))

    const wrapper = mountModal()
    await flushPromises()

    const errorState = wrapper.get('[data-testid="account-stats-load-error"]')
    expect(errorState.text()).toContain('admin.accounts.stats.failedToLoad')
    expect(wrapper.text()).not.toContain('admin.accounts.stats.noData')

    await errorState.get('button').trigger('click')
    await flushPromises()

    expect(getStats).toHaveBeenCalledTimes(2)
    expect(getStats).toHaveBeenLastCalledWith(1, 30)
    expect(wrapper.find('[data-testid="account-stats-load-error"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('$42.00')
  })

  it('ignores a stale response after the selected account changes', async () => {
    const firstRequest = deferred<ReturnType<typeof stats>>()
    const secondRequest = deferred<ReturnType<typeof stats>>()
    getStats
      .mockReturnValueOnce(firstRequest.promise)
      .mockReturnValueOnce(secondRequest.promise)

    const wrapper = mountModal(account(1, 'First account'))
    await wrapper.setProps({ account: account(2, 'Second account') })

    firstRequest.resolve(stats(11))
    await flushPromises()
    expect(wrapper.text()).not.toContain('$11.00')

    secondRequest.resolve(stats(22))
    await flushPromises()

    expect(getStats.mock.calls).toEqual([
      [1, 30],
      [2, 30]
    ])
    expect(wrapper.text()).toContain('Second account')
    expect(wrapper.text()).toContain('$22.00')
    expect(wrapper.text()).not.toContain('$11.00')
  })
})
