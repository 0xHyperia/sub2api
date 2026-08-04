import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AccountsView from '../AccountsView.vue'
import AccountBulkActionsBar from '@/components/admin/account/AccountBulkActionsBar.vue'
import AccountActionMenu from '@/components/admin/account/AccountActionMenu.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

const {
  listAccounts,
  listWithEtag,
  getBatchTodayStats,
  getAllProxies,
  getAllGroups,
  deleteAccount,
  batchDelete,
  batchClearError,
  batchRefresh,
  resetAccountQuota,
} = vi.hoisted(() => ({
  listAccounts: vi.fn(),
  listWithEtag: vi.fn(),
  getBatchTodayStats: vi.fn(),
  getAllProxies: vi.fn(),
  getAllGroups: vi.fn(),
  deleteAccount: vi.fn(),
  batchDelete: vi.fn(),
  batchClearError: vi.fn(),
  batchRefresh: vi.fn(),
  resetAccountQuota: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      list: listAccounts,
      listWithEtag,
      getBatchTodayStats,
      delete: deleteAccount,
      batchDelete,
      batchClearError,
      batchRefresh,
      resetAccountQuota,
      toggleSchedulable: vi.fn(),
    },
    proxies: { getAll: getAllProxies },
    groups: { getAll: getAllGroups },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
    showInfo: vi.fn(),
  }),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ token: 'test-token' }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const accounts = [
  { id: 11, name: 'one', platform: 'openai', type: 'oauth' },
  { id: 12, name: 'two', platform: 'openai', type: 'oauth' },
]

function mountView() {
  return mount(AccountsView, {
    attachTo: document.body,
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: {
          template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>',
        },
        DataTable: true,
        Pagination: true,
        ConfirmDialog: true,
        AccountTableActions: { template: '<div><slot name="beforeCreate" /><slot name="after" /></div>' },
        AccountTableFilters: true,
        AccountBulkActionsBar: true,
        AccountActionMenu: true,
        ImportDataModal: true,
        ReAuthAccountModal: true,
        AccountTestModal: true,
        AccountStatsModal: true,
        ScheduledTestsPanel: true,
        SyncFromCrsModal: true,
        TempUnschedStatusModal: true,
        ErrorPassthroughRulesModal: true,
        TLSFingerprintProfilesModal: true,
        CreateAccountModal: true,
        EditAccountModal: true,
        BulkEditAccountModal: true,
        PlatformTypeBadge: true,
        AccountCapacityCell: true,
        AccountStatusIndicator: true,
        AccountTodayStatsCell: true,
        AccountGroupsCell: true,
        AccountUsageCell: true,
        Icon: true,
      },
    },
  })
}

function activeDialog(wrapper: ReturnType<typeof mountView>) {
  const dialog = wrapper.findAllComponents(ConfirmDialog).find((item) => item.props('show'))
  if (!dialog) throw new Error('active confirmation dialog not found')
  return dialog
}

describe('admin AccountsView bulk confirmations', () => {
  beforeEach(() => {
    for (const mock of [
      listAccounts,
      listWithEtag,
      getBatchTodayStats,
      getAllProxies,
      getAllGroups,
      deleteAccount,
      batchDelete,
      batchClearError,
      batchRefresh,
      resetAccountQuota,
    ]) {
      mock.mockReset()
    }

    localStorage.clear()

    listAccounts.mockResolvedValue({ items: accounts, total: 2, page: 1, page_size: 20, pages: 1 })
    listWithEtag.mockResolvedValue({ notModified: true, etag: null, data: null })
    getBatchTodayStats.mockResolvedValue({ stats: {} })
    getAllProxies.mockResolvedValue([])
    getAllGroups.mockResolvedValue([])
    deleteAccount.mockResolvedValue(undefined)
    batchDelete.mockResolvedValue({ success: 2, failed: 0, failed_ids: [] })
    batchClearError.mockResolvedValue({ success: 2, failed: 0 })
    batchRefresh.mockResolvedValue({ success: 2, failed: 0 })
    resetAccountQuota.mockResolvedValue(accounts[0])
  })

  it.each([
    { event: 'delete', api: batchDelete, danger: true },
    { event: 'reset-status', api: batchClearError, danger: false },
    { event: 'refresh-token', api: batchRefresh, danger: false },
  ])('runs $event only after confirmation', async ({ event, api, danger }) => {
    const wrapper = mountView()
    await flushPromises()

    const actions = wrapper.findComponent(AccountBulkActionsBar)
    actions.vm.$emit('select-page')
    await flushPromises()
    actions.vm.$emit(event)
    await flushPromises()

    let dialog = activeDialog(wrapper)
    expect(dialog.props('danger')).toBe(danger)
    expect(api).not.toHaveBeenCalled()

    dialog.vm.$emit('cancel')
    await flushPromises()
    expect(api).not.toHaveBeenCalled()

    actions.vm.$emit(event)
    await flushPromises()
    dialog = activeDialog(wrapper)
    dialog.vm.$emit('confirm')
    await flushPromises()

    if (event === 'delete') {
      expect(batchDelete).toHaveBeenCalledOnce()
      expect(batchDelete).toHaveBeenCalledWith([11, 12])
    } else {
      expect(api).toHaveBeenCalledOnce()
      expect(api).toHaveBeenCalledWith([11, 12])
    }
    wrapper.unmount()
  })

  it('runs a single-account quota reset only after a danger confirmation', async () => {
    const wrapper = mountView()
    await flushPromises()

    const actionMenu = wrapper.findComponent(AccountActionMenu)
    actionMenu.vm.$emit('reset-quota', accounts[0])
    await flushPromises()

    let dialog = activeDialog(wrapper)
    expect(dialog.props('title')).toBe('admin.accounts.resetQuota')
    expect(dialog.props('danger')).toBe(true)
    expect(resetAccountQuota).not.toHaveBeenCalled()

    dialog.vm.$emit('cancel')
    await flushPromises()
    expect(resetAccountQuota).not.toHaveBeenCalled()

    actionMenu.vm.$emit('reset-quota', accounts[0])
    await flushPromises()
    dialog = activeDialog(wrapper)
    dialog.vm.$emit('confirm')
    await flushPromises()

    expect(resetAccountQuota).toHaveBeenCalledOnce()
    expect(resetAccountQuota).toHaveBeenCalledWith(11)
    wrapper.unmount()
  })

  it('wires both toolbar menus to the shared keyboard and ARIA model', async () => {
    const wrapper = mountView()
    await flushPromises()

    const autoRefreshTrigger = wrapper.get<HTMLButtonElement>(
      'button[aria-label="admin.accounts.autoRefresh"]'
    )
    await autoRefreshTrigger.trigger('keydown', { key: 'ArrowDown' })
    await flushPromises()

    const autoRefreshMenuId = autoRefreshTrigger.attributes('aria-controls')
    const autoRefreshMenu = wrapper.get(`#${autoRefreshMenuId}`)
    const autoRefreshItems = autoRefreshMenu.findAll<HTMLElement>(
      '[role="menuitemcheckbox"], [role="menuitemradio"]'
    )
    expect(autoRefreshMenu.attributes('aria-labelledby')).toBe(
      autoRefreshTrigger.attributes('id')
    )
    expect(autoRefreshItems.every((item) => item.attributes('tabindex') === '-1')).toBe(true)
    expect(document.activeElement).toBe(autoRefreshItems[0].element)

    await autoRefreshItems[0].trigger('keydown', { key: 'ArrowDown' })
    expect(document.activeElement).toBe(autoRefreshItems[1].element)
    await autoRefreshItems[1].trigger('keydown', { key: 'Escape' })
    await flushPromises()
    expect(wrapper.find(`#${autoRefreshMenuId}`).exists()).toBe(false)
    expect(document.activeElement).toBe(autoRefreshTrigger.element)

    const toolsTrigger = wrapper.get<HTMLButtonElement>(
      'button[aria-label="admin.accounts.moreActions"]'
    )
    await toolsTrigger.trigger('keydown', { key: 'ArrowUp' })
    await flushPromises()

    const toolsMenuId = toolsTrigger.attributes('aria-controls')
    const toolsMenu = wrapper.get(`#${toolsMenuId}`)
    const toolsItems = toolsMenu.findAll<HTMLElement>(
      '[role="menuitem"], [role="menuitemcheckbox"]'
    )
    expect(toolsMenu.attributes('aria-labelledby')).toBe(toolsTrigger.attributes('id'))
    expect(toolsItems.every((item) => item.attributes('tabindex') === '-1')).toBe(true)
    expect(document.activeElement).toBe(toolsItems[toolsItems.length - 1].element)
    wrapper.unmount()
  })
})
