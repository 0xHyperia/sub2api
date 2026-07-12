import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import LoginView from '@/views/auth/LoginView.vue'

const { getPublicSettingsMock, loginMock, showErrorMock } = vi.hoisted(() => ({
  getPublicSettingsMock: vi.fn(),
  loginMock: vi.fn(),
  showErrorMock: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    currentRoute: ref({ query: {} }),
    push: vi.fn(),
  }),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    locale: ref('en'),
    t: (key: string) => key,
  }),
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({ login: loginMock, login2FA: vi.fn() }),
  useAppStore: () => ({
    showError: showErrorMock,
    showSuccess: vi.fn(),
    showWarning: vi.fn(),
  }),
}))

vi.mock('@/api/auth', () => ({
  getPublicSettings: (...args: unknown[]) => getPublicSettingsMock(...args),
  isTotp2FARequired: () => false,
  isWeChatWebOAuthEnabled: () => false,
}))

const publicSettings = {
  turnstile_enabled: false,
  turnstile_site_key: '',
  linuxdo_oauth_enabled: false,
  dingtalk_oauth_enabled: false,
  backend_mode_enabled: false,
  oidc_oauth_enabled: false,
  github_oauth_enabled: false,
  google_oauth_enabled: false,
  password_reset_enabled: true,
  login_agreement_enabled: false,
  login_agreement_documents: [],
}

function mountLoginView(attachToDocument = false) {
  return mount(LoginView, {
    ...(attachToDocument ? { attachTo: document.body } : {}),
    global: {
      stubs: {
        GatewayAuthLayout: { template: '<main><slot name="heading" /><slot /></main>' },
        RouterLink: { template: '<a href="#"><slot /></a>' },
        EmailOAuthButtons: true,
        LinuxDoOAuthSection: true,
        DingTalkOAuthSection: true,
        OidcOAuthSection: true,
        WechatOAuthSection: true,
        LoginAgreementPrompt: true,
        TotpLoginModal: true,
        TurnstileWidget: true,
        Icon: true,
      },
    },
  })
}

describe('LoginView accessibility', () => {
  beforeEach(() => {
    getPublicSettingsMock.mockReset().mockResolvedValue(publicSettings)
    loginMock.mockReset()
    showErrorMock.mockReset()
    localStorage.clear()
    sessionStorage.clear()
  })

  it('associates inline validation errors with their fields', async () => {
    const wrapper = mountLoginView(true)
    await flushPromises()
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    const email = wrapper.get('#email')
    const password = wrapper.get('#password')
    expect(email.attributes('aria-invalid')).toBe('true')
    expect(email.attributes('aria-describedby')).toBe('login-email-error')
    expect(wrapper.get('#login-email-error').text()).toBe('auth.emailRequired')
    expect(password.attributes('aria-invalid')).toBe('true')
    expect(password.attributes('aria-describedby')).toBe('login-password-error')
    expect(document.activeElement).toBe(email.element)
    wrapper.unmount()
  })

  it('gives the password visibility control a stateful accessible name', async () => {
    const wrapper = mountLoginView()
    await flushPromises()

    const toggle = wrapper.get('button[aria-label="Show password"]')
    expect(wrapper.get('#password').attributes('type')).toBe('password')
    await toggle.trigger('click')
    expect(wrapper.get('#password').attributes('type')).toBe('text')
    expect(toggle.attributes('aria-label')).toBe('Hide password')
    expect(toggle.attributes('aria-pressed')).toBe('true')
  })

  it('renders authentication failures as an inline alert', async () => {
    loginMock.mockRejectedValue(new Error('Invalid credentials'))
    const wrapper = mountLoginView()
    await flushPromises()
    await wrapper.get('#email').setValue('user@example.com')
    await wrapper.get('#password').setValue('secret123')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.get('#login-form-error').attributes('role')).toBe('alert')
    expect(wrapper.get('#login-form-error').text()).toContain('Invalid credentials')
  })
})
