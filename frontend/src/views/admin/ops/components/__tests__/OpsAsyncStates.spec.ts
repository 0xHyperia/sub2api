import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import OpsAlertEventsCard from '../OpsAlertEventsCard.vue'
import OpsAlertRulesCard from '../OpsAlertRulesCard.vue'
import OpsConcurrencyCard from '../OpsConcurrencyCard.vue'

const mocks = vi.hoisted(() => ({
  listAlertEvents: vi.fn(),
  getAlertEvent: vi.fn(),
  updateAlertEventStatus: vi.fn(),
  listAlertRules: vi.fn(),
  updateAlertRule: vi.fn(),
  getConcurrencyStats: vi.fn(),
  getAccountAvailabilityStats: vi.fn(),
  getUserConcurrencyStats: vi.fn(),
  getGroups: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    listAlertEvents: mocks.listAlertEvents,
    getAlertEvent: mocks.getAlertEvent,
    updateAlertEventStatus: mocks.updateAlertEventStatus,
    listAlertRules: mocks.listAlertRules,
    updateAlertRule: mocks.updateAlertRule,
    getConcurrencyStats: mocks.getConcurrencyStats,
    getAccountAvailabilityStats: mocks.getAccountAvailabilityStats,
    getUserConcurrencyStats: mocks.getUserConcurrencyStats,
  },
}))

vi.mock('@/api', () => ({
  adminAPI: {
    groups: {
      getAll: mocks.getGroups,
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: mocks.showError,
    showSuccess: mocks.showSuccess,
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => params ? `${key}:${JSON.stringify(params)}` : key,
    }),
  }
})

const SelectStub = defineComponent({
  name: 'SelectControlStub',
  props: {
    modelValue: { type: [String, Number], default: '' },
  },
  emits: ['change', 'update:modelValue'],
  template: '<div class="select-stub" />',
})

const EmptyStateStub = defineComponent({
  name: 'EmptyState',
  inheritAttrs: true,
  props: {
    title: { type: String, default: '' },
    description: { type: String, default: '' },
    actionText: { type: String, default: '' },
    actionIcon: { type: Boolean, default: true },
  },
  emits: ['action'],
  template: `
    <div class="empty-state-stub">
      <span>{{ title }}</span>
      <span>{{ description }}</span>
      <button v-if="actionText" type="button" class="retry-action" @click="$emit('action')">{{ actionText }}</button>
    </div>
  `,
})

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

function alertEvent(id: number, title: string) {
  return {
    id,
    rule_id: 1,
    severity: 'P1',
    status: 'firing',
    title,
    description: `${title} message`,
    dimensions: { platform: 'openai', model: 'gpt-5', user: 'alice' },
    fired_at: `2026-01-01T00:00:${String(id % 60).padStart(2, '0')}Z`,
    email_sent: true,
    created_at: `2026-01-01T00:00:${String(id % 60).padStart(2, '0')}Z`,
  }
}

function concurrencyResponse(platform: string) {
  return {
    enabled: true,
    platform: {
      [platform]: {
        platform,
        current_in_use: 1,
        max_capacity: 4,
        load_percentage: 25,
        waiting_in_queue: 0,
      },
    },
    group: {},
    account: {},
  }
}

function availabilityResponse(platform: string) {
  return {
    enabled: true,
    platform: {
      [platform]: {
        platform,
        total_accounts: 2,
        available_count: 2,
        rate_limit_count: 0,
        error_count: 0,
      },
    },
    group: {},
    account: {},
  }
}

const commonStubs = {
  Select: SelectStub,
  EmptyState: EmptyStateStub,
  BaseDialog: true,
  ConfirmDialog: true,
  Icon: true,
}

describe('ops async request states', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.getGroups.mockResolvedValue([])
    mocks.getAlertEvent.mockImplementation(async (id: number) => alertEvent(id, `detail-${id}`))
    mocks.updateAlertEventStatus.mockResolvedValue(undefined)
    mocks.updateAlertRule.mockImplementation(async (_id: number, patch: Record<string, unknown>) => patch)
    mocks.getUserConcurrencyStats.mockResolvedValue({ enabled: true, user: {} })
  })

  it('does not append a load-more response from a previous alert filter', async () => {
    const staleLoadMore = deferred<ReturnType<typeof alertEvent>[]>()
    const firstPage = Array.from({ length: 10 }, (_, index) => alertEvent(index + 1, `initial-${index + 1}`))
    mocks.listAlertEvents
      .mockResolvedValueOnce(firstPage)
      .mockImplementationOnce(() => staleLoadMore.promise)
      .mockResolvedValueOnce([alertEvent(101, 'current-filter')])

    const wrapper = mount(OpsAlertEventsCard, { global: { stubs: commonStubs } })
    await flushPromises()

    await wrapper.find('.max-h-\\[600px\\]').trigger('scroll')
    expect(mocks.listAlertEvents).toHaveBeenCalledTimes(2)

    const severitySelect = wrapper.findAllComponents(SelectStub)[1]
    severitySelect.vm.$emit('change', 'P1')
    await nextTick()
    await flushPromises()

    expect(wrapper.text()).toContain('current-filter')
    staleLoadMore.resolve([alertEvent(202, 'stale-load-more')])
    await flushPromises()

    expect(wrapper.text()).not.toContain('stale-load-more')
    expect(wrapper.text()).toContain('current-filter')
    wrapper.unmount()
  })

  it('shows a persistent alert-event error and recovers through retry', async () => {
    mocks.listAlertEvents
      .mockRejectedValueOnce(new Error('alerts unavailable'))
      .mockResolvedValueOnce([alertEvent(1, 'alerts recovered')])

    const wrapper = mount(OpsAlertEventsCard, { global: { stubs: commonStubs } })
    await flushPromises()

    expect(wrapper.find('[data-testid="alert-events-load-error"]').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('admin.ops.alertEvents.empty')
    await wrapper.find('[data-testid="alert-events-load-error"] .retry-action').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('alerts recovered')
    expect(wrapper.find('[data-testid="alert-events-load-error"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('keeps existing alert events visible when a refresh fails', async () => {
    mocks.listAlertEvents
      .mockResolvedValueOnce([alertEvent(1, 'preserved alert')])
      .mockRejectedValueOnce(new Error('refresh unavailable'))

    const wrapper = mount(OpsAlertEventsCard, { global: { stubs: commonStubs } })
    await flushPromises()

    const refreshButton = wrapper.findAll('button').find((button) => button.text().includes('common.refresh'))
    expect(refreshButton).toBeDefined()
    await refreshButton!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('preserved alert')
    expect(wrapper.find('[data-testid="alert-events-refresh-error"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('renders alert events as mobile action cards while preserving the desktop table', async () => {
    const event = alertEvent(12, 'Queue depth exceeded')
    mocks.listAlertEvents.mockResolvedValue([event])

    const wrapper = mount(OpsAlertEventsCard, { global: { stubs: commonStubs } })
    await flushPromises()

    const mobileCard = wrapper.get('[data-test="mobile-alert-event-12"]')
    expect(wrapper.get('[data-test="mobile-alert-events"]').classes()).toContain('md:hidden')
    expect(wrapper.get('[data-test="desktop-alert-events"]').classes()).toEqual(expect.arrayContaining(['hidden', 'md:block']))
    expect(wrapper.get('[data-test="desktop-alert-events"] > div').classes()).toContain('min-w-[900px]')
    expect(mobileCard.text()).toContain('P1')
    expect(mobileCard.text()).toContain('admin.ops.alertEvents.status.firing')
    expect(mobileCard.text()).toContain('Queue depth exceeded')
    expect(mobileCard.text()).toContain('alice')
    expect(mobileCard.text()).toContain('gpt-5')

    await wrapper.get('[data-test="mobile-alert-view-12"]').trigger('click')
    await flushPromises()
    expect(mocks.getAlertEvent).toHaveBeenCalledWith(12)

    await wrapper.get('[data-test="mobile-alert-resolve-12"]').trigger('click')
    await flushPromises()
    expect(mocks.updateAlertEventStatus).toHaveBeenCalledWith(12, 'manual_resolved')
    expect(wrapper.get('[data-test="mobile-alert-event-12"]').text()).toContain('admin.ops.alertEvents.status.manualResolved')
    wrapper.unmount()
  })

  it('keeps alert rules out of the empty state after a failed initial load and retries', async () => {
    mocks.listAlertRules
      .mockRejectedValueOnce(new Error('rules unavailable'))
      .mockResolvedValueOnce([{
        id: 7,
        name: 'Recovered Rule',
        enabled: true,
        metric_type: 'error_rate',
        operator: '>',
        threshold: 1,
        window_minutes: 1,
        sustained_minutes: 2,
        severity: 'P1',
        cooldown_minutes: 10,
        notify_email: true,
      }])

    const wrapper = mount(OpsAlertRulesCard, { global: { stubs: commonStubs } })
    await flushPromises()

    expect(wrapper.find('[data-testid="alert-rules-load-error"]').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('admin.ops.alertRules.empty')
    await wrapper.find('[data-testid="alert-rules-load-error"] .retry-action').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Recovered Rule')
    wrapper.unmount()
  })

  it('keeps existing alert rules visible when a refresh fails', async () => {
    const existingRule = {
      id: 8,
      name: 'Preserved Rule',
      enabled: true,
      metric_type: 'error_rate',
      operator: '>',
      threshold: 1,
      window_minutes: 1,
      sustained_minutes: 2,
      severity: 'P1',
      cooldown_minutes: 10,
      notify_email: true,
    }
    mocks.listAlertRules
      .mockResolvedValueOnce([existingRule])
      .mockRejectedValueOnce(new Error('refresh unavailable'))

    const wrapper = mount(OpsAlertRulesCard, { global: { stubs: commonStubs } })
    await flushPromises()

    const refreshButton = wrapper.findAll('button').find((button) => button.text().includes('common.refresh'))
    expect(refreshButton).toBeDefined()
    await refreshButton!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Preserved Rule')
    expect(wrapper.find('[data-testid="alert-rules-refresh-error"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('renders mobile alert-rule cards with edit, toggle, and delete actions', async () => {
    const rule = {
      id: 18,
      name: 'High error rate',
      description: 'Notify when errors exceed the threshold',
      enabled: true,
      metric_type: 'error_rate',
      operator: '>',
      threshold: 5,
      window_minutes: 5,
      sustained_minutes: 2,
      severity: 'P1',
      cooldown_minutes: 10,
      notify_email: true,
    }
    mocks.listAlertRules.mockResolvedValue([rule])
    const DialogStateStub = defineComponent({
      props: ['show'],
      template: '<div v-if="show" data-test="rule-editor"><slot /><slot name="footer" /></div>',
    })
    const ConfirmStateStub = defineComponent({
      props: ['show'],
      template: '<div v-if="show" data-test="rule-delete-confirm" />',
    })

    const wrapper = mount(OpsAlertRulesCard, {
      global: { stubs: { ...commonStubs, BaseDialog: DialogStateStub, ConfirmDialog: ConfirmStateStub } },
    })
    await flushPromises()

    const mobileCard = wrapper.get('[data-test="mobile-alert-rule-18"]')
    expect(wrapper.get('[data-test="mobile-alert-rules"]').classes()).toContain('md:hidden')
    expect(wrapper.get('[data-test="desktop-alert-rules"]').classes()).toEqual(expect.arrayContaining(['hidden', 'md:block']))
    expect(wrapper.get('[data-test="desktop-alert-rules"] > div').classes()).toContain('min-w-[680px]')
    expect(mobileCard.text()).toContain('High error rate')
    expect(mobileCard.text()).toContain('error_rate')
    expect(mobileCard.text()).toContain('> 5')
    expect(mobileCard.text()).toContain('P1')
    expect(mobileCard.text()).toContain('admin.ops.alertRules.form.notifyEmail')

    await wrapper.get('[data-test="mobile-rule-edit-18"]').trigger('click')
    expect(wrapper.find('[data-test="rule-editor"]').exists()).toBe(true)

    await wrapper.get('[data-test="mobile-rule-toggle-18"]').trigger('click')
    await flushPromises()
    expect(mocks.updateAlertRule).toHaveBeenCalledWith(18, { enabled: false })
    expect(wrapper.get('[data-test="mobile-alert-rule-18"]').text()).toContain('common.disabled')

    await wrapper.get('[data-test="mobile-rule-delete-18"]').trigger('click')
    expect(wrapper.find('[data-test="rule-delete-confirm"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('lets the newest concurrency refresh win', async () => {
    const oldConcurrency = deferred<ReturnType<typeof concurrencyResponse>>()
    const oldAvailability = deferred<ReturnType<typeof availabilityResponse>>()
    mocks.getConcurrencyStats
      .mockImplementationOnce(() => oldConcurrency.promise)
      .mockResolvedValueOnce(concurrencyResponse('new'))
    mocks.getAccountAvailabilityStats
      .mockImplementationOnce(() => oldAvailability.promise)
      .mockResolvedValueOnce(availabilityResponse('new'))

    const wrapper = mount(OpsConcurrencyCard, {
      props: { refreshToken: 0 },
      global: { stubs: { EmptyState: EmptyStateStub } },
    })
    await nextTick()
    await wrapper.setProps({ refreshToken: 1 })
    await flushPromises()

    expect(wrapper.text()).toContain('NEW')
    oldConcurrency.resolve(concurrencyResponse('old'))
    oldAvailability.resolve(availabilityResponse('old'))
    await flushPromises()

    expect(wrapper.text()).toContain('NEW')
    expect(wrapper.text()).not.toContain('OLD')
    wrapper.unmount()
  })

  it('ignores a stale concurrency error after switching view modes', async () => {
    mocks.getConcurrencyStats.mockResolvedValue(concurrencyResponse('platform'))
    mocks.getAccountAvailabilityStats.mockResolvedValue(availabilityResponse('platform'))
    const staleUserRequest = deferred<{ enabled: boolean; user: Record<string, never> }>()
    mocks.getUserConcurrencyStats.mockImplementationOnce(() => staleUserRequest.promise)

    const wrapper = mount(OpsConcurrencyCard, {
      props: { refreshToken: 0 },
      global: { stubs: { EmptyState: EmptyStateStub } },
    })
    await flushPromises()

    const modeButton = wrapper.find('button[title="admin.ops.concurrency.switchToUser"]')
    await modeButton.trigger('click')
    await nextTick()
    await wrapper.find('button[title="admin.ops.concurrency.switchToPlatform"]').trigger('click')
    await flushPromises()

    staleUserRequest.reject(new Error('stale user error'))
    await flushPromises()

    expect(wrapper.text()).toContain('PLATFORM')
    expect(wrapper.text()).not.toContain('admin.ops.concurrency.loadFailed')
    wrapper.unmount()
  })

  it('shows a retryable concurrency error instead of an empty state', async () => {
    mocks.getConcurrencyStats
      .mockRejectedValueOnce(new Error('concurrency unavailable'))
      .mockResolvedValueOnce(concurrencyResponse('recovered'))
    mocks.getAccountAvailabilityStats.mockResolvedValue(availabilityResponse('recovered'))

    const wrapper = mount(OpsConcurrencyCard, {
      props: { refreshToken: 0 },
      global: { stubs: { EmptyState: EmptyStateStub } },
    })
    await flushPromises()

    expect(wrapper.find('[data-testid="concurrency-load-error"]').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('admin.ops.concurrency.empty')
    await wrapper.find('[data-testid="concurrency-load-error"] .retry-action').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('RECOVERED')
    wrapper.unmount()
  })

  it('keeps existing concurrency rows visible when auto refresh fails', async () => {
    mocks.getConcurrencyStats
      .mockResolvedValueOnce(concurrencyResponse('preserved'))
      .mockRejectedValueOnce(new Error('refresh unavailable'))
    mocks.getAccountAvailabilityStats.mockResolvedValue(availabilityResponse('preserved'))

    const wrapper = mount(OpsConcurrencyCard, {
      props: { refreshToken: 0 },
      global: { stubs: { EmptyState: EmptyStateStub } },
    })
    await flushPromises()

    await wrapper.setProps({ refreshToken: 1 })
    await flushPromises()

    expect(wrapper.text()).toContain('PRESERVED')
    expect(wrapper.find('[data-testid="concurrency-refresh-error"]').exists()).toBe(true)
    wrapper.unmount()
  })
})
