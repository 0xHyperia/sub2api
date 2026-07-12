import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import type { AdminUser } from '@/types'
import UserAllowedGroupsModal from '../UserAllowedGroupsModal.vue'
import UserApiKeysModal from '../UserApiKeysModal.vue'
import UserBalanceModal from '../UserBalanceModal.vue'
import UserBalanceHistoryModal from '../UserBalanceHistoryModal.vue'

const apiMocks = vi.hoisted(() => ({
  getUserApiKeys: vi.fn(),
  getAllGroups: vi.fn(),
  updateApiKeyGroup: vi.fn(),
  getUserBalanceHistory: vi.fn(),
  listGroups: vi.fn(),
  updateUser: vi.fn(),
  updateBalance: vi.fn()
}))

const storeMocks = vi.hoisted(() => ({
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      getUserApiKeys: apiMocks.getUserApiKeys,
      getUserBalanceHistory: apiMocks.getUserBalanceHistory,
      update: apiMocks.updateUser,
      updateBalance: apiMocks.updateBalance
    },
    groups: {
      getAll: apiMocks.getAllGroups,
      list: apiMocks.listGroups
    },
    apiKeys: {
      updateApiKeyGroup: apiMocks.updateApiKeyGroup
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => storeMocks
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
  reject: (reason?: unknown) => void
}

function deferred<T>(): Deferred<T> {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return { promise, resolve, reject }
}

function createUser(id: number, overrides: Partial<AdminUser> = {}): AdminUser {
  return {
    id,
    username: `user-${id}`,
    email: `user-${id}@example.com`,
    role: 'user',
    balance: id,
    concurrency: 1,
    status: 'active',
    allowed_groups: [],
    balance_notify_enabled: false,
    balance_notify_threshold: null,
    balance_notify_extra_emails: [],
    created_at: '2026-07-01T00:00:00Z',
    updated_at: '2026-07-01T00:00:00Z',
    notes: '',
    ...overrides
  }
}

function apiKey(id: number, name: string) {
  return {
    id,
    name,
    key: `sk-${String(id).padEnd(36, 'x')}`,
    status: 'active',
    group_id: null,
    group: null,
    created_at: '2026-07-01T00:00:00Z'
  }
}

function historyItem(id: number, notes: string) {
  return {
    id,
    type: 'admin_balance',
    value: id,
    notes,
    created_at: '2026-07-01T00:00:00Z'
  }
}

function group(id: number, name: string, overrides: Record<string, unknown> = {}) {
  return {
    id,
    name,
    platform: 'openai',
    subscription_type: 'standard',
    status: 'active',
    is_exclusive: true,
    rate_multiplier: 1,
    ...overrides
  }
}

const BaseDialogStub = {
  props: ['show'],
  template: '<div v-if="show"><slot /><slot name="footer" /></div>'
}

const SelectStub = {
  props: ['modelValue'],
  emits: ['update:modelValue', 'change'],
  template: '<div data-testid="select-stub" />'
}

beforeEach(() => {
  Object.values(apiMocks).forEach((mock) => mock.mockReset())
  Object.values(storeMocks).forEach((mock) => mock.mockReset())
  apiMocks.getAllGroups.mockResolvedValue([])
  apiMocks.updateApiKeyGroup.mockResolvedValue({ api_key: apiKey(1, 'updated') })
  apiMocks.getUserBalanceHistory.mockResolvedValue({ items: [], total: 0, total_recharged: 0 })
  apiMocks.listGroups.mockResolvedValue({ items: [], total: 0 })
  apiMocks.updateUser.mockResolvedValue({})
  apiMocks.updateBalance.mockResolvedValue({})
  vi.spyOn(console, 'error').mockImplementation(() => undefined)
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe('UserApiKeysModal async state', () => {
  function mountModal(user: AdminUser) {
    return mount(UserApiKeysModal, {
      props: { show: true, user },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          GroupBadge: true,
          GroupOptionItem: true,
          Teleport: true
        }
      }
    })
  }

  it('clears the previous user and ignores a stale response after rapid switching', async () => {
    const userBRequest = deferred<{ items: ReturnType<typeof apiKey>[] }>()
    const userCRequest = deferred<{ items: ReturnType<typeof apiKey>[] }>()
    apiMocks.getUserApiKeys
      .mockResolvedValueOnce({ items: [apiKey(1, 'key-user-a')] })
      .mockReturnValueOnce(userBRequest.promise)
      .mockReturnValueOnce(userCRequest.promise)

    const wrapper = mountModal(createUser(1))
    await flushPromises()
    expect(wrapper.text()).toContain('key-user-a')

    await wrapper.setProps({ user: createUser(2) })
    expect(wrapper.text()).not.toContain('key-user-a')

    await wrapper.setProps({ user: createUser(3) })
    userBRequest.resolve({ items: [apiKey(2, 'stale-key-user-b')] })
    await flushPromises()
    expect(wrapper.text()).not.toContain('stale-key-user-b')

    userCRequest.resolve({ items: [apiKey(3, 'key-user-c')] })
    await flushPromises()
    expect(wrapper.text()).toContain('key-user-c')
  })

  it('keeps a load error visible and retries with empty data', async () => {
    apiMocks.getUserApiKeys
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce({ items: [apiKey(4, 'retried-key')] })

    const wrapper = mountModal(createUser(4))
    await flushPromises()

    expect(wrapper.text()).toContain('admin.users.failedToLoadApiKeys')
    expect(wrapper.text()).not.toContain('admin.users.noApiKeys')

    await wrapper.get('[data-testid="api-keys-retry"]').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('retried-key')
  })

  it('exposes an ARIA listbox and updates a key group from the keyboard', async () => {
    apiMocks.getUserApiKeys.mockResolvedValueOnce({ items: [apiKey(6, 'keyboard-key')] })
    apiMocks.getAllGroups.mockResolvedValueOnce([group(60, 'Keyboard Group')])
    apiMocks.updateApiKeyGroup.mockResolvedValueOnce({ api_key: apiKey(6, 'keyboard-key') })

    const wrapper = mountModal(createUser(6))
    await flushPromises()

    const trigger = wrapper.get('[data-testid="api-key-group-select"] [role="combobox"]')
    expect(trigger.attributes('aria-haspopup')).toBe('listbox')
    expect(trigger.attributes('aria-label')).toContain('keyboard-key')

    await trigger.trigger('keydown', { key: 'ArrowDown' })
    await flushPromises()
    const listbox = wrapper.get('[role="listbox"]')
    expect(trigger.attributes('aria-controls')).toBe(listbox.attributes('id'))

    await trigger.trigger('keydown', { key: 'ArrowDown' })
    await trigger.trigger('keydown', { key: 'Enter' })
    await flushPromises()
    expect(apiMocks.updateApiKeyGroup).toHaveBeenCalledWith(6, 60)
  })
})

describe('UserBalanceModal async state', () => {
  it('invalidates a pending balance update when the selected user changes', async () => {
    const updateRequest = deferred<Record<string, never>>()
    apiMocks.updateBalance.mockReturnValueOnce(updateRequest.promise)

    const wrapper = mount(UserBalanceModal, {
      props: { show: true, user: createUser(30), operation: 'add' },
      global: { stubs: { BaseDialog: BaseDialogStub } }
    })

    expect(wrapper.get('label[for="balance-amount"]').exists()).toBe(true)
    expect(wrapper.get('label[for="balance-notes"]').exists()).toBe(true)
    await wrapper.get('#balance-amount').setValue('5')
    await wrapper.get('#balance-form').trigger('submit')
    expect(apiMocks.updateBalance).toHaveBeenCalledWith(30, 5, 'add', '')

    await wrapper.setProps({ user: createUser(31) })
    updateRequest.resolve({})
    await flushPromises()

    expect(storeMocks.showSuccess).not.toHaveBeenCalled()
    expect(wrapper.emitted('success')).toBeUndefined()
    expect(wrapper.emitted('close')).toBeUndefined()
  })
})

describe('UserBalanceHistoryModal async state', () => {
  function mountModal(user: AdminUser) {
    return mount(UserBalanceHistoryModal, {
      props: { show: true, user },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          Icon: true
        }
      }
    })
  }

  it('clears totals and history while switching users and ignores stale history', async () => {
    const userBRequest = deferred<{ items: ReturnType<typeof historyItem>[]; total: number; total_recharged: number }>()
    const userCRequest = deferred<{ items: ReturnType<typeof historyItem>[]; total: number; total_recharged: number }>()
    apiMocks.getUserBalanceHistory
      .mockResolvedValueOnce({ items: [historyItem(1, 'history-user-a')], total: 1, total_recharged: 99 })
      .mockReturnValueOnce(userBRequest.promise)
      .mockReturnValueOnce(userCRequest.promise)

    const wrapper = mountModal(createUser(1))
    await flushPromises()
    expect(wrapper.text()).toContain('history-user-a')
    expect(wrapper.text()).toContain('$99.00')

    await wrapper.setProps({ user: createUser(2) })
    expect(wrapper.text()).not.toContain('history-user-a')
    expect(wrapper.text()).not.toContain('$99.00')

    await wrapper.setProps({ user: createUser(3) })
    userBRequest.resolve({ items: [historyItem(2, 'stale-history-user-b')], total: 1, total_recharged: 88 })
    await flushPromises()
    expect(wrapper.text()).not.toContain('stale-history-user-b')

    userCRequest.resolve({ items: [historyItem(3, 'history-user-c')], total: 1, total_recharged: 77 })
    await flushPromises()
    expect(wrapper.text()).toContain('history-user-c')
    expect(wrapper.text()).toContain('$77.00')
  })

  it('shows a persistent error and retries the current history request', async () => {
    apiMocks.getUserBalanceHistory
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce({ items: [historyItem(5, 'retried-history')], total: 1, total_recharged: 5 })

    const wrapper = mountModal(createUser(5))
    await flushPromises()
    expect(wrapper.text()).toContain('admin.users.failedToLoadBalanceHistory')
    expect(wrapper.text()).not.toContain('admin.users.noBalanceHistory')

    await wrapper.get('[data-testid="balance-history-retry"]').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('retried-history')
  })
})

describe('UserAllowedGroupsModal async state', () => {
  function mountModal(user: AdminUser) {
    return mount(UserAllowedGroupsModal, {
      props: { show: true, user },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          PlatformIcon: true
        }
      }
    })
  }

  it('disables and hard-blocks saving after load failure, then enables retry', async () => {
    apiMocks.listGroups
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce({ items: [group(10, 'Exclusive Ten')], total: 1 })

    const wrapper = mountModal(createUser(10, { allowed_groups: [10] }))
    await flushPromises()

    const saveButton = wrapper.get('[data-testid="allowed-groups-save"]')
    expect(wrapper.text()).toContain('admin.users.failedToLoadGroups')
    expect(saveButton.attributes('disabled')).toBeDefined()
    await saveButton.trigger('click')
    expect(apiMocks.updateUser).not.toHaveBeenCalled()

    await wrapper.get('[data-testid="allowed-groups-retry"]').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Exclusive Ten')
    expect(saveButton.attributes('disabled')).toBeUndefined()

    await saveButton.trigger('click')
    await flushPromises()
    expect(apiMocks.updateUser).toHaveBeenCalledWith(10, {
      allowed_groups: [10],
      group_rates: undefined
    })
  })

  it('ignores group data that resolves for a previously selected user', async () => {
    const userARequest = deferred<{ items: ReturnType<typeof group>[]; total: number }>()
    const userBRequest = deferred<{ items: ReturnType<typeof group>[]; total: number }>()
    apiMocks.listGroups
      .mockReturnValueOnce(userARequest.promise)
      .mockReturnValueOnce(userBRequest.promise)

    const wrapper = mountModal(createUser(20, { allowed_groups: [20] }))
    await wrapper.setProps({ user: createUser(21, { allowed_groups: [21] }) })

    userARequest.resolve({ items: [group(20, 'Stale Group Twenty')], total: 1 })
    await flushPromises()
    expect(wrapper.text()).not.toContain('Stale Group Twenty')
    expect(wrapper.get('[data-testid="allowed-groups-save"]').attributes('disabled')).toBeDefined()

    userBRequest.resolve({ items: [group(21, 'Current Group Twenty One')], total: 1 })
    await flushPromises()
    expect(wrapper.text()).toContain('Current Group Twenty One')
    expect(wrapper.get('[data-testid="allowed-groups-save"]').attributes('disabled')).toBeUndefined()
  })

  it('labels group checkboxes and custom-rate controls without relying on visual text', async () => {
    apiMocks.listGroups.mockResolvedValueOnce({
      items: [
        group(40, 'Exclusive Accessible'),
        group(41, 'Public Accessible', { is_exclusive: false })
      ],
      total: 2
    })

    const wrapper = mountModal(createUser(40))
    await flushPromises()

    const exclusiveCheckbox = wrapper.get('#exclusive-group-40')
    expect(exclusiveCheckbox.attributes('aria-label')).toContain('Exclusive Accessible')
    expect(wrapper.get('label[for="exclusive-group-40"]').exists()).toBe(true)
    expect(wrapper.get('label[for="exclusive-rate-40"]').exists()).toBe(true)
    expect(wrapper.get('#exclusive-rate-40').attributes('aria-describedby')).toBe('group-rate-meta-40')

    const publicCheckbox = wrapper.get('input[aria-label*="Public Accessible"]')
    expect(publicCheckbox.attributes('checked')).toBeDefined()
    expect(publicCheckbox.attributes('disabled')).toBeDefined()
    expect(wrapper.get('label[for="public-rate-41"]').exists()).toBe(true)
  })
})
