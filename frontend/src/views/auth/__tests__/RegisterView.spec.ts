import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import RegisterView from '@/views/auth/RegisterView.vue'

const {
  getPublicSettingsMock,
  pushMock,
  replaceMock,
  showErrorMock,
  showSuccessMock,
  registerMock
} = vi.hoisted(() => ({
  getPublicSettingsMock: vi.fn(),
  pushMock: vi.fn(),
  replaceMock: vi.fn(),
  showErrorMock: vi.fn(),
  showSuccessMock: vi.fn(),
  registerMock: vi.fn()
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ path: '/register', query: {} }),
  useRouter: () => ({ push: pushMock, replace: replaceMock })
}))

vi.mock('vue-i18n', () => ({
  createI18n: () => ({
    global: {
      t: (key: string) => key
    }
  }),
  useI18n: () => ({
    locale: { value: 'zh' },
    t: (key: string) => key
  })
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({ register: registerMock }),
  useAppStore: () => ({
    showError: showErrorMock,
    showSuccess: showSuccessMock,
    showWarning: vi.fn()
  })
}))

vi.mock('@/api/auth', async () => {
  const actual = await vi.importActual<typeof import('@/api/auth')>('@/api/auth')
  return {
    ...actual,
    getPublicSettings: (...args: unknown[]) => getPublicSettingsMock(...args),
    isWeChatWebOAuthEnabled: () => false
  }
})

describe('RegisterView', () => {
  beforeEach(() => {
    getPublicSettingsMock.mockReset()
    pushMock.mockReset()
    replaceMock.mockReset()
    showErrorMock.mockReset()
    showSuccessMock.mockReset()
    registerMock.mockReset()
    localStorage.clear()
  })

  it('does not override the cached backend site name while settings are loading', () => {
    getPublicSettingsMock.mockReturnValue(new Promise(() => {}))

    const wrapper = mount(RegisterView, {
      global: {
        stubs: {
          GatewayAuthLayout: {
            name: 'GatewayAuthLayout',
            props: ['siteName'],
            template: '<main><slot name="heading" /><slot /></main>'
          },
          EmailOAuthButtons: true,
          Icon: true,
          LoginAgreementPrompt: true,
          TurnstileWidget: true,
          transition: false
        }
      }
    })

    expect(wrapper.getComponent({ name: 'GatewayAuthLayout' }).props('siteName')).toBe('')
  })

  it('does not reveal the registration form before settings confirm registration is enabled', async () => {
    let resolveSettings!: (settings: Record<string, unknown>) => void
    getPublicSettingsMock.mockReturnValue(
      new Promise<Record<string, unknown>>((resolve) => {
        resolveSettings = resolve
      })
    )

    const wrapper = mount(RegisterView, {
      global: {
        stubs: {
          GatewayAuthLayout: {
            template: '<main><slot name="heading" /><slot /></main>'
          },
          EmailOAuthButtons: true,
          Icon: true,
          LoginAgreementPrompt: true,
          TurnstileWidget: true,
          transition: false
        }
      }
    })

    expect(wrapper.find('form').exists()).toBe(false)

    resolveSettings({
      registration_enabled: false,
      email_verify_enabled: false,
      promo_code_enabled: false,
      invitation_code_enabled: false,
      turnstile_enabled: false,
      site_name: 'USA-零',
      linuxdo_oauth_enabled: false,
      oidc_oauth_enabled: false,
      github_oauth_enabled: false,
      google_oauth_enabled: false,
      registration_email_suffix_whitelist: []
    })
    await flushPromises()

    expect(wrapper.find('form').exists()).toBe(false)
    expect(wrapper.text()).toContain('auth.registrationDisabled')
  })

  it('shows the registration form after settings confirm registration is enabled', async () => {
    getPublicSettingsMock.mockResolvedValue({
      registration_enabled: true,
      email_verify_enabled: false,
      promo_code_enabled: false,
      invitation_code_enabled: false,
      turnstile_enabled: false,
      site_name: 'USA-零',
      linuxdo_oauth_enabled: false,
      oidc_oauth_enabled: false,
      github_oauth_enabled: false,
      google_oauth_enabled: false,
      registration_email_suffix_whitelist: []
    })

    const wrapper = mount(RegisterView, {
      global: {
        stubs: {
          GatewayAuthLayout: {
            template: '<main><slot name="heading" /><slot /></main>'
          },
          EmailOAuthButtons: true,
          Icon: true,
          LoginAgreementPrompt: true,
          TurnstileWidget: true,
          transition: false
        }
      }
    })
    await flushPromises()

    expect(wrapper.find('form').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('auth.registrationDisabled')
  })

  it('switches to inline email verification without leaving the registration page', async () => {
    getPublicSettingsMock.mockResolvedValue({
      registration_enabled: true,
      email_verify_enabled: true,
      promo_code_enabled: false,
      invitation_code_enabled: false,
      turnstile_enabled: false,
      site_name: 'USA-零',
      linuxdo_oauth_enabled: false,
      oidc_oauth_enabled: false,
      github_oauth_enabled: false,
      google_oauth_enabled: false,
      registration_email_suffix_whitelist: []
    })

    const wrapper = mount(RegisterView, {
      global: {
        stubs: {
          GatewayAuthLayout: {
            template: '<main><slot name="heading" /><slot /></main>'
          },
          EmailVerificationStep: {
            template: '<div data-testid="email-verification-step" />'
          },
          EmailOAuthButtons: true,
          Icon: true,
          LoginAgreementPrompt: true,
          TurnstileWidget: true,
          transition: false
        }
      }
    })
    await flushPromises()

    await wrapper.get('#email').setValue('new@example.com')
    await wrapper.get('#password').setValue('secret-123')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.find('[data-testid="email-verification-step"]').exists()).toBe(true)
    expect(pushMock).not.toHaveBeenCalledWith('/email-verify')
    expect(replaceMock).toHaveBeenCalledWith({
      path: '/register',
      query: { step: 'verify' }
    })
  })
})
