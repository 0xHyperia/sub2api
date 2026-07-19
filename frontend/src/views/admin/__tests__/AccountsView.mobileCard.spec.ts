import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'

import AccountsView from '../AccountsView.vue'

const {
  listAccounts,
  listWithEtag,
  getBatchTodayStats,
  getUpstreamBillingProbeSettings,
  getAllProxies,
  getAllGroups
} = vi.hoisted(() => ({
  listAccounts: vi.fn(),
  listWithEtag: vi.fn(),
  getBatchTodayStats: vi.fn(),
  getUpstreamBillingProbeSettings: vi.fn(),
  getAllProxies: vi.fn(),
  getAllGroups: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      list: listAccounts,
      listWithEtag,
      getBatchTodayStats,
      getUpstreamBillingProbeSettings,
      delete: vi.fn(),
      batchClearError: vi.fn(),
      batchRefresh: vi.fn(),
      toggleSchedulable: vi.fn()
    },
    proxies: { getAll: getAllProxies },
    groups: { getAll: getAllGroups }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
    showInfo: vi.fn()
  })
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    token: 'test-token',
    isSimpleMode: false
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => params?.name ? `${key}: ${String(params.name)}` : key
    })
  }
})

const account = {
  id: 42,
  name: 'Primary OpenAI',
  notes: 'Business critical account',
  platform: 'openai',
  type: 'apikey',
  credentials: { email: 'owner@example.com' },
  proxy_id: 3,
  proxy: { id: 3, name: 'Hong Kong Proxy' },
  concurrency: 8,
  current_concurrency: 3,
  priority: 50,
  status: 'active',
  error_message: null,
  last_used_at: '2026-07-19T02:00:00Z',
  expires_at: null,
  auto_pause_on_expired: false,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-07-19T02:00:00Z',
  schedulable: true,
  rate_limited_at: null,
  rate_limit_reset_at: null,
  overload_until: null,
  temp_unschedulable_until: null,
  temp_unschedulable_reason: null,
  session_window_start: null,
  session_window_end: null,
  session_window_status: null,
  quota_daily_limit: 25,
  quota_daily_used: 7.5,
  groups: [{ id: 9, name: 'Production' }]
}

const DataTableStub = {
  props: ['data'],
  template: `
    <div data-test="data-table">
      <div v-for="row in data" :key="row.id">
        <div data-test="mobile-slot"><slot name="mobile-card" :row="row" :index="0" :selected="false" :expanded="false" /></div>
        <div data-test="desktop-select-slot"><slot name="cell-select" :row="row" /></div>
      </div>
    </div>
  `
}

const modalStub = (testId: string) => ({
  props: ['show', 'account'],
  template: `<div v-if="show" data-test="${testId}">{{ account?.name }}</div>`
})

const ConfirmDialogStub = {
  props: ['show', 'title', 'message'],
  template: '<div v-if="show" data-test="confirm-dialog">{{ title }} {{ message }}</div>'
}

let wrapper: VueWrapper | null = null

function mountView() {
  wrapper = mount(AccountsView, {
    attachTo: document.body,
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
        DataTable: DataTableStub,
        HelpTooltip: true,
        Pagination: true,
        ConfirmDialog: ConfirmDialogStub,
        AccountTableActions: { template: '<div><slot name="beforeCreate" /><slot name="after" /></div>' },
        AccountTableFilters: true,
        AccountBulkActionsBar: true,
        ImportDataModal: true,
        ReAuthAccountModal: true,
        AccountTestModal: modalStub('test-modal'),
        AccountStatsModal: true,
        ScheduledTestsPanel: true,
        SyncFromCrsModal: true,
        TempUnschedStatusModal: true,
        ErrorPassthroughRulesModal: true,
        TLSFingerprintProfilesModal: true,
        CreateAccountModal: true,
        EditAccountModal: modalStub('edit-modal'),
        BulkEditAccountModal: true,
        PlatformTypeBadge: { props: ['platform', 'type'], template: '<span data-test="platform-type">{{ platform }}/{{ type }}</span>' },
        AccountCapacityCell: true,
        AccountStatusIndicator: { props: ['account'], template: '<span data-test="account-status">{{ account.status }}</span>' },
        AccountTodayStatsCell: true,
        AccountGroupsCell: { props: ['groups'], template: '<span data-test="account-groups">{{ groups.map(group => group.name).join(", ") }}</span>' },
        AccountUsageCell: true,
        UpstreamBillingRateCell: true,
        Icon: true,
        Toggle: true
      }
    }
  })
  return wrapper
}

describe('admin AccountsView mobile card', () => {
  beforeEach(() => {
    localStorage.clear()
    localStorage.setItem('account-hidden-columns', JSON.stringify(['today_stats', 'usage']))
    localStorage.setItem('account-hidden-columns-version', 'scheduler-score-hidden-by-default')
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 390 })

    listAccounts.mockReset().mockResolvedValue({ items: [account], total: 1, page: 1, page_size: 20, pages: 1 })
    listWithEtag.mockReset().mockResolvedValue({ notModified: true, etag: null, data: null })
    getBatchTodayStats.mockReset().mockResolvedValue({
      stats: { '42': { requests: 120, tokens: 8000, cost: 1.25, standard_cost: 1.5, user_cost: 2 } }
    })
    getUpstreamBillingProbeSettings.mockReset().mockResolvedValue({ enabled: true, interval_minutes: 30 })
    getAllProxies.mockReset().mockResolvedValue([])
    getAllGroups.mockReset().mockResolvedValue([])
  })

  afterEach(() => {
    wrapper?.unmount()
    wrapper = null
    document.body.innerHTML = ''
  })

  it('shows identity, business metrics and details while reusing account selection', async () => {
    const view = mountView()
    await flushPromises()

    const card = view.get('[data-test="account-mobile-card-42"]')
    expect(card.text()).toContain('openai/apikey')
    expect(card.text()).toContain('Primary OpenAI')
    expect(card.text()).toContain('owner@example.com')
    expect(card.text()).toContain('$1.2500')
    expect(card.text()).toContain('$7.50 / $25.00')
    expect(card.text()).toContain('3 / 8')
    expect(getBatchTodayStats).toHaveBeenCalledWith([42])

    const mobileCheckbox = card.get('input[type="checkbox"]')
    const desktopCheckbox = view.get('[data-test="desktop-select-slot"] input[type="checkbox"]')
    await mobileCheckbox.setValue(true)
    expect((desktopCheckbox.element as HTMLInputElement).checked).toBe(true)

    await card.get('[data-test="account-mobile-details-toggle-42"]').trigger('click')
    const details = card.get('[data-test="account-mobile-details-42"]')
    expect(details.text()).toContain('#42')
    expect(details.text()).toContain('Hong Kong Proxy')
    expect(details.text()).toContain('Production')
    expect(details.text()).toContain('Business critical account')
  })

  it('keeps test and edit visible and exposes delete through mobile More', async () => {
    const view = mountView()
    await flushPromises()

    await view.get('[data-test="account-mobile-test-42"]').trigger('click')
    expect(view.get('[data-test="test-modal"]').text()).toContain('Primary OpenAI')

    await view.get('[data-test="account-mobile-edit-42"]').trigger('click')
    expect(view.get('[data-test="edit-modal"]').text()).toContain('Primary OpenAI')

    await view.get('[data-test="account-mobile-more-42"]').trigger('click')
    await flushPromises()
    const deleteAction = document.body.querySelector<HTMLButtonElement>('[data-test="account-action-delete"]')
    expect(deleteAction).not.toBeNull()
    deleteAction?.click()
    await flushPromises()
    expect(view.get('[data-test="confirm-dialog"]').text()).toContain('Primary OpenAI')
  })
})
