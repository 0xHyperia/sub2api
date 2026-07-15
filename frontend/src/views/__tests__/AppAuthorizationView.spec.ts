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
    t: (key: string, params?: Record<string, string>) =>
      params?.app ? `${key}:${params.app}` : key
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
      client_id: 'zerobox-desktop',
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
      client_id: 'zerobox-desktop',
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
      client_id: 'zerobox-desktop',
      client_name: 'ZeroBox',
      device_name: 'Workstation',
      platform: 'windows',
      scopes: ['profile:read', 'keys:write']
    })
    mocks.submitDecision.mockReturnValue(new Promise(() => undefined))

    const wrapper = mount(AppAuthorizationView)
    await flushPromises()

    expect(wrapper.text()).toContain('ZeroBox')
    expect(wrapper.text()).toContain('profile:read')
    await wrapper.get('button.btn-secondary').trigger('click')

    expect(mocks.submitDecision).toHaveBeenCalledWith('request-456', 'deny')
  })

  it('uses the semantic authorization surface', async () => {
    mocks.authenticated = true
    mocks.query = { request_id: 'request-789' }
    mocks.getContext.mockResolvedValue({
      request_id: 'request-789',
      client_id: 'zerobox-desktop',
      client_name: 'ZeroBox',
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
