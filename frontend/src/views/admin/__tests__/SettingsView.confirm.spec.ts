import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'

import SettingsView from '../SettingsView.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

const {
  getSettings,
  getWebSearchEmulationConfig,
  getAdminApiKey,
  regenerateAdminApiKey,
  deleteAdminApiKey,
  resetWebSearchUsage,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  getSettings: vi.fn(),
  getWebSearchEmulationConfig: vi.fn(),
  getAdminApiKey: vi.fn(),
  regenerateAdminApiKey: vi.fn(),
  deleteAdminApiKey: vi.fn(),
  resetWebSearchUsage: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: {} }),
  useRouter: () => ({ replace: vi.fn() }),
}))

vi.mock('@/api', () => ({
  adminAPI: {
    settings: {
      getSettings,
      getWebSearchEmulationConfig,
      getAdminApiKey,
      regenerateAdminApiKey,
      deleteAdminApiKey,
      resetWebSearchUsage,
      getOverloadCooldownSettings: vi.fn().mockResolvedValue({}),
      getRateLimit429CooldownSettings: vi.fn().mockResolvedValue({}),
      getStreamTimeoutSettings: vi.fn().mockResolvedValue({}),
      getRectifierSettings: vi.fn().mockResolvedValue({}),
      getBetaPolicySettings: vi.fn().mockResolvedValue({ rules: [] }),
    },
    groups: { getAll: vi.fn().mockResolvedValue([]) },
    proxies: { list: vi.fn().mockResolvedValue({ items: [] }) },
    payment: { getProviders: vi.fn().mockResolvedValue({ data: [] }) },
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
    showWarning: vi.fn(),
    showInfo: vi.fn(),
    fetchPublicSettings: vi.fn(),
  }),
}))

vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => ({ fetch: vi.fn() }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard: vi.fn() }),
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: (_error: unknown, fallback: string) => fallback,
  extractI18nErrorMessage: (_error: unknown, _t: unknown, _prefix: string, fallback: string) => fallback,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      locale: { value: 'en' },
      t: (key: string) => key,
    }),
  }
})

const webSearchProvider = {
  type: 'brave',
  api_key: '',
  api_key_configured: true,
  quota_limit: 10,
  quota_used: 9,
  subscribed_at: null,
  proxy_id: null,
  expires_at: null,
}

function mountView() {
  return shallowMount(SettingsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        ConfirmDialog: true,
        RouterLink: true,
      },
    },
  })
}

function findButton(wrapper: ReturnType<typeof mountView>, text: string) {
  const button = wrapper.findAll('button').find((item) => item.text().includes(text))
  if (!button) throw new Error(`button not found: ${text}`)
  return button
}

function activeDialog(wrapper: ReturnType<typeof mountView>) {
  const dialog = wrapper.findAllComponents(ConfirmDialog).find((item) => item.props('show'))
  if (!dialog) throw new Error('active confirmation dialog not found')
  return dialog
}

describe('admin SettingsView confirmations', () => {
  beforeEach(() => {
    for (const mock of [
      getSettings,
      getWebSearchEmulationConfig,
      getAdminApiKey,
      regenerateAdminApiKey,
      deleteAdminApiKey,
      resetWebSearchUsage,
      showError,
      showSuccess,
    ]) {
      mock.mockReset()
    }

    getSettings.mockResolvedValue({
      payment_load_balance_strategy: 'round-robin',
      backend_mode_enabled: false,
      login_agreement_documents: [],
      default_subscriptions: [],
      default_platform_quotas: {},
      registration_email_suffix_whitelist: [],
      table_page_size_options: [10, 20, 50],
      wechat_connect_open_enabled: false,
      wechat_connect_mp_enabled: false,
      wechat_connect_mobile_enabled: false,
    })
    getWebSearchEmulationConfig.mockResolvedValue({
      enabled: true,
      providers: [{ ...webSearchProvider }],
    })
    getAdminApiKey.mockResolvedValue({ exists: true, masked_key: 'sk-admin...1234' })
    regenerateAdminApiKey.mockResolvedValue({ key: 'sk-new-admin-key' })
    deleteAdminApiKey.mockResolvedValue(undefined)
    resetWebSearchUsage.mockResolvedValue(undefined)
  })

  it('confirms API key regeneration and deletion with danger tone', async () => {
    const wrapper = mountView()
    await flushPromises()

    await findButton(wrapper, 'admin.settings.adminApiKey.regenerate').trigger('click')
    let dialog = activeDialog(wrapper)
    expect(dialog.props('danger')).toBe(true)
    expect(regenerateAdminApiKey).not.toHaveBeenCalled()

    dialog.vm.$emit('cancel')
    await flushPromises()
    expect(regenerateAdminApiKey).not.toHaveBeenCalled()

    await findButton(wrapper, 'admin.settings.adminApiKey.regenerate').trigger('click')
    dialog = activeDialog(wrapper)
    dialog.vm.$emit('confirm')
    await flushPromises()
    expect(regenerateAdminApiKey).toHaveBeenCalledOnce()

    await findButton(wrapper, 'admin.settings.adminApiKey.delete').trigger('click')
    dialog = activeDialog(wrapper)
    expect(dialog.props('danger')).toBe(true)
    expect(deleteAdminApiKey).not.toHaveBeenCalled()

    dialog.vm.$emit('confirm')
    await flushPromises()
    expect(deleteAdminApiKey).toHaveBeenCalledOnce()
  })

  it('resets provider usage only after destructive confirmation', async () => {
    const wrapper = mountView()
    await flushPromises()

    const view = wrapper.vm as unknown as {
      resetWebSearchUsage: (index: number) => void
    }
    view.resetWebSearchUsage(0)
    await flushPromises()

    let dialog = activeDialog(wrapper)
    expect(dialog.props('danger')).toBe(true)
    expect(dialog.props('message')).toBe(
      'admin.settings.webSearchEmulation.resetUsageConfirm',
    )
    expect(resetWebSearchUsage).not.toHaveBeenCalled()

    dialog.vm.$emit('cancel')
    await flushPromises()
    expect(resetWebSearchUsage).not.toHaveBeenCalled()

    view.resetWebSearchUsage(0)
    await flushPromises()
    dialog = activeDialog(wrapper)
    dialog.vm.$emit('confirm')
    await flushPromises()

    expect(resetWebSearchUsage).toHaveBeenCalledOnce()
    expect(resetWebSearchUsage).toHaveBeenCalledWith({ provider_type: 'brave' })
  })
})
