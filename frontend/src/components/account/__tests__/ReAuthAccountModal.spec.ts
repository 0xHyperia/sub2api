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

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
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
    <div data-testid="oauth-flow">
      <p v-if="error" data-testid="oauth-visible-error">{{ error }}</p>
      <button type="button" data-testid="generate-auth-url" @click="$emit('generate-url')">generate</button>
    </div>
  `
}

function account(id: number, platform: 'openai' | 'gemini' | 'antigravity' | 'anthropic' = 'openai') {
  return {
    id,
    name: `Account ${id}`,
    platform,
    type: 'oauth',
    proxy_id: null,
    credentials: {}
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

describe('ReAuthAccountModal', () => {
  beforeEach(() => {
    for (const oauth of [
      mocks.claudeOAuth,
      mocks.openaiOAuth,
      mocks.geminiOAuth,
      mocks.antigravityOAuth
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
    mocks.claudeOAuth.buildExtraInfo.mockReset()
    mocks.updateAccount.mockReset().mockResolvedValue(undefined)
    mocks.clearError.mockReset().mockResolvedValue(undefined)
    mocks.exchangeCode.mockReset()
    mocks.showSuccess.mockReset()
    mocks.showError.mockReset()
  })

  it('uses a neutral account summary and mobile-stacked actions', () => {
    const wrapper = mountModal()
    const summary = wrapper.get('[data-testid="reauth-account-summary"]')

    expect(summary.classes()).toContain('rounded-panel')
    expect(summary.classes()).toContain('bg-surface-subtle')
    expect(wrapper.find('.bg-gradient-to-br').exists()).toBe(false)
    expect(wrapper.find('[class*="from-green"]').exists()).toBe(false)
    expect(wrapper.get('#reauth-account-platform').text()).toBe('admin.accounts.openaiAccount')
    expect(wrapper.get('[aria-busy="false"]').exists()).toBe(true)

    const footerLayout = wrapper.get('footer > div')
    expect(footerLayout.classes()).toContain('flex-col-reverse')
    expect(footerLayout.findAll('button').every((button) => button.classes().includes('w-full'))).toBe(true)
  })

  it('groups the Anthropic authorization methods as touch-sized labelled radios', () => {
    const wrapper = mountModal(account(1, 'anthropic'))
    const fieldset = wrapper.get('fieldset')
    const radios = fieldset.findAll('input[name="reauthorization-method"]')

    expect(fieldset.get('legend').text()).toBe('admin.accounts.oauth.authMethod')
    expect(radios).toHaveLength(2)
    expect(radios.every((radio) => radio.element.closest('label')?.classList.contains('min-h-touch'))).toBe(true)
  })

  it('keeps OAuth errors inline and exposes the same generate action for retry', async () => {
    mocks.openaiOAuth.error.value = 'network unavailable'
    const wrapper = mountModal()

    expect(wrapper.get('[data-testid="oauth-visible-error"]').text()).toBe('network unavailable')
    expect(wrapper.get('#reauthorization-error').attributes('role')).toBe('alert')

    await wrapper.get('[data-testid="generate-auth-url"]').trigger('click')
    await flushPromises()

    expect(mocks.openaiOAuth.generateAuthUrl).toHaveBeenCalledWith(null)
  })

  it('updates and clears the selected account after a current OAuth exchange succeeds', async () => {
    const wrapper = mountModal(account(12))
    const setupState = (wrapper.vm as any).$?.setupState
    setupState.oauthFlowRef = {
      authCode: 'authorization-code',
      oauthState: 'oauth-state',
      projectId: 'project-id',
      sessionKey: '',
      inputMethod: 'manual',
      reset: vi.fn()
    }
    await wrapper.vm.$nextTick()

    await wrapper.get('button.btn-primary').trigger('click')
    await flushPromises()

    expect(mocks.openaiOAuth.exchangeAuthCode).toHaveBeenCalledWith(
      'authorization-code',
      'session-id',
      'oauth-state',
      null
    )
    expect(mocks.updateAccount).toHaveBeenCalledWith(12, {
      type: 'oauth',
      credentials: { access_token: 'token' },
      extra: {}
    })
    expect(mocks.clearError).toHaveBeenCalledWith(12)
    expect(mocks.showSuccess).toHaveBeenCalledWith('admin.accounts.reAuthorizedSuccess')
    expect(wrapper.emitted('reauthorized')).toHaveLength(1)
    expect(wrapper.emitted('close')).toHaveLength(1)
  })

  it('ignores an OAuth exchange response after the selected account changes', async () => {
    const pendingExchange = deferred<{ access_token: string }>()
    mocks.openaiOAuth.exchangeAuthCode.mockReturnValueOnce(pendingExchange.promise)
    const wrapper = mountModal(account(1))
    const setupState = (wrapper.vm as any).$?.setupState
    setupState.oauthFlowRef = {
      authCode: 'authorization-code',
      oauthState: 'oauth-state',
      projectId: 'project-id',
      sessionKey: '',
      inputMethod: 'manual',
      reset: vi.fn()
    }
    await wrapper.vm.$nextTick()

    const completeButton = wrapper.get('button.btn-primary')
    expect((completeButton.element as HTMLButtonElement).disabled).toBe(false)
    await completeButton.trigger('click')

    await wrapper.setProps({ account: account(2) })
    pendingExchange.resolve({ access_token: 'stale-token' })
    await flushPromises()

    expect(mocks.updateAccount).not.toHaveBeenCalled()
    expect(mocks.clearError).not.toHaveBeenCalled()
    expect(mocks.showSuccess).not.toHaveBeenCalled()
    expect(wrapper.emitted('reauthorized')).toBeUndefined()
  })
})
