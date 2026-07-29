import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AppAuthorizationView from '@/views/AppAuthorizationView.vue'

const mocks = vi.hoisted(() => ({
  query: {} as Record<string, string>,
  replace: vi.fn(),
  authenticated: false,
  createRequest: vi.fn(),
  getContext: vi.fn(),
  submitDecision: vi.fn()
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: mocks.query }),
  useRouter: () => ({ replace: mocks.replace })
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ isAuthenticated: mocks.authenticated })
}))

vi.mock('@/api/appAuth', () => ({
  createAuthorizationRequest: mocks.createRequest,
  getAuthorizationContext: mocks.getContext,
  submitAuthorizationDecision: mocks.submitDecision
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, string>) => {
      if (key === 'appAuthorization.requestDescription') {
        return `${params?.app} 正在请求访问您的 USA0 账户。`
      }
      const scopes: Record<string, string> = {
        'appAuthorization.scopes.execution_authorize': '授权执行受保护操作',
        'appAuthorization.scopes.profile_read': '查看您的个人资料',
        'appAuthorization.scopes.profile_write': '更新您的个人资料',
        'appAuthorization.scopes.usage_read': '查看您的用量信息'
      }
      return scopes[key] ?? key
    }
  })
}))

describe('AppAuthorizationView', () => {
  beforeEach(() => {
    mocks.query = {}
    mocks.replace.mockReset().mockResolvedValue(undefined)
    mocks.createRequest.mockReset()
    mocks.getContext.mockReset()
    mocks.submitDecision.mockReset()
    mocks.authenticated = false
  })

  it('initializes raw OAuth parameters and sends an unauthenticated user to login with only request_id', async () => {
    mocks.query = {
      response_type: 'code',
      client_id: 'zeroagent-desktop',
      redirect_uri: 'http://127.0.0.1:49152/oauth/callback',
      scope: 'profile:read keys:read',
      code_challenge: 'challenge',
      code_challenge_method: 'S256',
      state: 'opaque-state',
      device_name: 'Workstation',
      platform: 'windows'
    }
    mocks.createRequest.mockResolvedValue({ request_id: 'request-123' })

    mount(AppAuthorizationView)
    await flushPromises()

    expect(mocks.createRequest).toHaveBeenCalledWith(expect.objectContaining({
      client_id: 'zeroagent-desktop',
      state: 'opaque-state',
      device_name: 'Workstation'
    }))
    expect(mocks.replace).toHaveBeenNthCalledWith(1, {
      name: 'AppAuthorization',
      query: { request_id: 'request-123' }
    })
    expect(mocks.replace).toHaveBeenNthCalledWith(2, {
      path: '/login',
      query: { redirect: '/oauth/authorize?request_id=request-123' }
    })
    expect(mocks.getContext).not.toHaveBeenCalled()
  })

  it('loads a validated request and submits an explicit decision', async () => {
    mocks.authenticated = true
    mocks.query = { request_id: 'request-456' }
    mocks.getContext.mockResolvedValue({
      request_id: 'request-456',
      client_id: 'zeroagent-desktop',
      client_name: 'ZeroAgent',
      device_name: 'Workstation',
      platform: 'windows',
      scopes: ['execution:authorize', 'profile:read', 'profile:write', 'usage:read']
    })
    mocks.submitDecision.mockReturnValue(new Promise(() => undefined))

    const wrapper = mount(AppAuthorizationView)
    await flushPromises()

    expect(wrapper.text()).toContain('ZeroAgent')
    expect(wrapper.text()).toContain('ZeroAgent 正在请求访问您的 USA0 账户。')
    expect(wrapper.text()).toContain('授权执行受保护操作')
    expect(wrapper.text()).toContain('查看您的个人资料')
    expect(wrapper.text()).toContain('更新您的个人资料')
    expect(wrapper.text()).toContain('查看您的用量信息')
    expect(wrapper.text()).not.toContain('execution:authorize')
    expect(wrapper.text()).not.toContain('profile:write')
    expect(wrapper.text()).not.toContain('usage:read')
    await wrapper.get('button.btn-secondary').trigger('click')

    expect(mocks.submitDecision).toHaveBeenCalledWith('request-456', 'deny')
  })

  it('uses the semantic authorization surface', async () => {
    mocks.authenticated = true
    mocks.query = { request_id: 'request-789' }
    mocks.getContext.mockResolvedValue({
      request_id: 'request-789',
      client_id: 'zeroagent-desktop',
      client_name: 'ZeroAgent',
      device_name: 'Workstation',
      platform: 'windows',
      scopes: ['profile:read']
    })

    const wrapper = mount(AppAuthorizationView)
    await flushPromises()

    expect(wrapper.get('section').classes()).toContain('bg-surface')
    expect(wrapper.html()).not.toMatch(/(?:bg|text|border)-gray-|dark:/)
  })
})
