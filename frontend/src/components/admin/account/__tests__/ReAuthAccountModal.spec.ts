import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import ReAuthAccountModal from '../ReAuthAccountModal.vue'

const mocks = vi.hoisted(() => {
  const state = <T>(value: T) => ({ value })
  const client = () => ({
    authUrl: state(''),
    sessionId: state('session-id'),
    loading: state(false),
    error: state(''),
    resetState: vi.fn(),
    generateAuthUrl: vi.fn()
  })

  return {
    claudeOAuth: {
      ...client(),
      buildExtraInfo: vi.fn()
    },
    openaiOAuth: {
      ...client(),
      oauthState: state('oauth-state'),
      exchangeAuthCode: vi.fn(),
      buildCredentials: vi.fn(),
      buildExtraInfo: vi.fn()
    },
    geminiOAuth: {
      ...client(),
      state: state('oauth-state'),
      exchangeAuthCode: vi.fn(),
      buildCredentials: vi.fn()
    },
    antigravityOAuth: {
      ...client(),
      state: state('oauth-state'),
      exchangeAuthCode: vi.fn(),
      buildCredentials: vi.fn()
    },
    grokOAuth: {
      ...client(),
      state: state('oauth-state'),
      exchangeAuthCode: vi.fn(),
      buildCredentials: vi.fn(),
      buildExtraInfo: vi.fn()
    },
    applyOAuthCredentials: vi.fn(),
    updateAccount: vi.fn(),
    clearError: vi.fn(),
    exchangeCode: vi.fn(),
    showSuccess: vi.fn(),
    showError: vi.fn()
  }
})

vi.mock('@/composables/useAccountOAuth', () => ({
  useAccountOAuth: () => mocks.claudeOAuth
}))

vi.mock('@/composables/useOpenAIOAuth', () => ({
  useOpenAIOAuth: () => mocks.openaiOAuth
}))

vi.mock('@/composables/useGeminiOAuth', () => ({
  useGeminiOAuth: () => mocks.geminiOAuth
}))

vi.mock('@/composables/useAntigravityOAuth', () => ({
  useAntigravityOAuth: () => mocks.antigravityOAuth
}))

vi.mock('@/composables/useGrokOAuth', () => ({
  useGrokOAuth: () => mocks.grokOAuth
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      applyOAuthCredentials: mocks.applyOAuthCredentials,
      update: mocks.updateAccount,
      clearError: mocks.clearError,
      exchangeCode: mocks.exchangeCode
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: mocks.showSuccess,
    showError: mocks.showError
  })
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

const BaseDialogStub = {
  props: ['show', 'title', 'width'],
  emits: ['close'],
  template: '<div v-if="show"><slot /><footer><slot name="footer" /></footer></div>'
}

const OAuthAuthorizationFlowStub = {
  props: ['error', 'loading', 'platform'],
  emits: ['generate-url', 'cookie-auth'],
  setup(_props: unknown, { expose }: { expose: (value: Record<string, unknown>) => void }) {
    expose({
      authCode: 'authorization-code',
      oauthState: 'oauth-state',
      projectId: 'project-id',
      sessionKey: '',
      inputMethod: 'manual',
      reset: vi.fn()
    })
  },
  template: `
    <div data-testid="admin-oauth-flow" :data-platform="platform">
      <p v-if="error" data-testid="admin-oauth-visible-error">{{ error }}</p>
      <button type="button" data-testid="admin-generate-auth-url" @click="$emit('generate-url')">generate</button>
    </div>
  `
}

type TestPlatform = 'openai' | 'gemini' | 'antigravity' | 'grok' | 'anthropic'

function account(id: number, platform: TestPlatform = 'openai') {
  return {
    id,
    name: `Account ${id}`,
    platform,
    type: 'oauth',
    proxy_id: null,
    credentials: {},
    status: 'active'
  } as any
}

function mountModal(selectedAccount = account(1)) {
  return mount(ReAuthAccountModal, {
    props: {
      show: true,
      account: selectedAccount
    },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        OAuthAuthorizationFlow: OAuthAuthorizationFlowStub,
        Icon: true
      }
    }
  })
}

function exposeOAuthFlow(wrapper: ReturnType<typeof mountModal>) {
  const setupState = (wrapper.vm as any).$?.setupState
  setupState.oauthFlowRef = {
    authCode: 'authorization-code',
    oauthState: 'oauth-state',
    projectId: 'project-id',
    sessionKey: '',
    inputMethod: 'manual',
    reset: vi.fn()
  }
}

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

describe('admin ReAuthAccountModal', () => {
  beforeEach(() => {
    for (const oauth of [
      mocks.claudeOAuth,
      mocks.openaiOAuth,
      mocks.geminiOAuth,
      mocks.antigravityOAuth,
      mocks.grokOAuth
    ]) {
      oauth.authUrl.value = ''
      oauth.sessionId.value = 'session-id'
      oauth.loading.value = false
      oauth.error.value = ''
      oauth.resetState.mockReset()
      oauth.generateAuthUrl.mockReset().mockResolvedValue(undefined)
    }

    mocks.openaiOAuth.oauthState.value = 'oauth-state'
    mocks.openaiOAuth.exchangeAuthCode.mockReset().mockResolvedValue({ access_token: 'token' })
    mocks.openaiOAuth.buildCredentials.mockReset().mockReturnValue({ access_token: 'token' })
    mocks.openaiOAuth.buildExtraInfo.mockReset().mockReturnValue({})
    mocks.geminiOAuth.exchangeAuthCode.mockReset()
    mocks.geminiOAuth.buildCredentials.mockReset()
    mocks.antigravityOAuth.exchangeAuthCode.mockReset()
    mocks.antigravityOAuth.buildCredentials.mockReset()
    mocks.grokOAuth.exchangeAuthCode.mockReset()
    mocks.grokOAuth.buildCredentials.mockReset()
    mocks.grokOAuth.buildExtraInfo.mockReset()
    mocks.claudeOAuth.buildExtraInfo.mockReset()
    mocks.applyOAuthCredentials.mockReset()
    mocks.updateAccount.mockReset()
    mocks.clearError.mockReset()
    mocks.exchangeCode.mockReset()
    mocks.showSuccess.mockReset()
    mocks.showError.mockReset()
  })

  it('uses neutral surfaces, mobile-stacked actions, and preserves Grok platform context', () => {
    const wrapper = mountModal(account(1, 'grok'))
    const summary = wrapper.get('[data-testid="admin-reauth-account-summary"]')

    expect(summary.classes()).toContain('rounded-panel')
    expect(summary.classes()).toContain('bg-surface-subtle')
    expect(wrapper.find('.bg-gradient-to-br').exists()).toBe(false)
    expect(wrapper.find('[class*="from-zinc"]').exists()).toBe(false)
    expect(wrapper.get('#admin-reauth-account-platform').text()).toBe('admin.accounts.grokAccount')
    expect(wrapper.get('[data-testid="admin-oauth-flow"]').attributes('data-platform')).toBe('grok')

    const footerLayout = wrapper.get('footer > div')
    expect(footerLayout.classes()).toContain('flex-col-reverse')
    expect(footerLayout.findAll('button').every((button) => button.classes().includes('w-full'))).toBe(true)
  })

  it('groups Anthropic methods as touch-sized labelled radios', () => {
    const wrapper = mountModal(account(1, 'anthropic'))
    const fieldset = wrapper.get('fieldset')
    const radios = fieldset.findAll('input[name="admin-reauthorization-method"]')

    expect(fieldset.get('legend').text()).toBe('admin.accounts.oauth.authMethod')
    expect(radios).toHaveLength(2)
    expect(radios.every((radio) => radio.element.closest('label')?.classList.contains('min-h-touch'))).toBe(true)
  })

  it('keeps errors inline and leaves the generation action available for retry', async () => {
    mocks.openaiOAuth.error.value = 'network unavailable'
    const wrapper = mountModal()

    expect(wrapper.get('[data-testid="admin-oauth-visible-error"]').text()).toBe('network unavailable')
    expect(wrapper.get('#admin-reauthorization-error').attributes('role')).toBe('alert')

    await wrapper.get('[data-testid="admin-generate-auth-url"]').trigger('click')
    await flushPromises()

    expect(mocks.openaiOAuth.generateAuthUrl).toHaveBeenCalledWith(null)
  })

  it('emits the updated account returned by applyOAuthCredentials', async () => {
    const updatedAccount = account(12)
    updatedAccount.name = 'Updated account'
    mocks.applyOAuthCredentials.mockResolvedValueOnce(updatedAccount)
    const wrapper = mountModal(account(12))
    exposeOAuthFlow(wrapper)
    await wrapper.vm.$nextTick()

    await wrapper.get('button.btn-primary').trigger('click')
    await flushPromises()

    expect(mocks.applyOAuthCredentials).toHaveBeenCalledWith(12, {
      type: 'oauth',
      credentials: { access_token: 'token' },
      extra: {}
    })
    expect(wrapper.emitted('reauthorized')).toEqual([[updatedAccount]])
    expect(wrapper.emitted('close')).toHaveLength(1)
  })

  it('does not apply credentials from a stale exchange after the account changes', async () => {
    const pendingExchange = deferred<{ access_token: string }>()
    mocks.openaiOAuth.exchangeAuthCode.mockReturnValueOnce(pendingExchange.promise)
    const wrapper = mountModal(account(1))
    exposeOAuthFlow(wrapper)
    await wrapper.vm.$nextTick()

    await wrapper.get('button.btn-primary').trigger('click')
    await wrapper.setProps({ account: account(2) })
    pendingExchange.resolve({ access_token: 'stale-token' })
    await flushPromises()

    expect(mocks.applyOAuthCredentials).not.toHaveBeenCalled()
    expect(mocks.updateAccount).not.toHaveBeenCalled()
    expect(mocks.showSuccess).not.toHaveBeenCalled()
    expect(wrapper.emitted('reauthorized')).toBeUndefined()
  })
})
