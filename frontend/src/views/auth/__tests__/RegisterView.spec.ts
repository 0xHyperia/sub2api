import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import RegisterView from '@/views/auth/RegisterView.vue'

const {
  getPublicSettingsMock,
  pushMock,
  replaceMock,
  showErrorMock,
  showSuccessMock,
  registerMock,
  getPromotionTrackingStatusMock,
  trackPromotionVisitMock,
  validatePromoCodeMock,
  validateInvitationCodeMock,
  routeQuery
} = vi.hoisted(() => ({
  getPublicSettingsMock: vi.fn(),
  pushMock: vi.fn(),
  replaceMock: vi.fn(),
  showErrorMock: vi.fn(),
  showSuccessMock: vi.fn(),
  registerMock: vi.fn(),
  getPromotionTrackingStatusMock: vi.fn(),
  trackPromotionVisitMock: vi.fn(),
  validatePromoCodeMock: vi.fn(),
  validateInvitationCodeMock: vi.fn(),
  routeQuery: {} as Record<string, string>
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ path: '/register', fullPath: '/register', query: routeQuery }),
  useRouter: () => ({ push: pushMock, replace: replaceMock })
}))

vi.mock('@/api/distribution', () => ({
  getPromotionTrackingStatus: (...args: unknown[]) => getPromotionTrackingStatusMock(...args),
  trackPromotionVisit: (...args: unknown[]) => trackPromotionVisitMock(...args)
}))

vi.mock('vue-i18n', () => ({
  createI18n: () => ({
    global: {
      t: (key: string) => key
    }
  }),
  useI18n: () => ({
    t: (key: string) =>
      key === 'auth.emailDomainRegistrationLimit'
        ? '该邮箱域名无法注册新账户。请使用主流邮箱注册；如需使用企业邮箱，请联系客服添加域名白名单。'
        : key,
    locale: { value: 'zh' }
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
    validatePromoCode: (...args: unknown[]) => validatePromoCodeMock(...args),
    validateInvitationCode: (...args: unknown[]) => validateInvitationCodeMock(...args),
    isWeChatWebOAuthEnabled: () => false
  }
})

const enabledSettings = {
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
}

function mountRegisterView() {
  return mount(RegisterView, {
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
}

describe('RegisterView', () => {
  beforeEach(() => {
    getPublicSettingsMock.mockReset()
    pushMock.mockReset()
    replaceMock.mockReset()
    showErrorMock.mockReset()
    showSuccessMock.mockReset()
    registerMock.mockReset()
    getPromotionTrackingStatusMock.mockReset()
    trackPromotionVisitMock.mockReset()
    validatePromoCodeMock.mockReset()
    validateInvitationCodeMock.mockReset()
    for (const key of Object.keys(routeQuery)) delete routeQuery[key]
    localStorage.clear()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('loads public settings before checking promotion tracking and skips the tracking POST when disabled', async () => {
    routeQuery.agent = 'AGENT001'
    getPublicSettingsMock.mockResolvedValue(enabledSettings)
    getPromotionTrackingStatusMock.mockResolvedValue({ enabled: false })

    mountRegisterView()
    await flushPromises()

    expect(getPublicSettingsMock).toHaveBeenCalledOnce()
    expect(getPromotionTrackingStatusMock).toHaveBeenCalledOnce()
    expect(getPublicSettingsMock.mock.invocationCallOrder[0]).toBeLessThan(
      getPromotionTrackingStatusMock.mock.invocationCallOrder[0],
    )
    expect(trackPromotionVisitMock).not.toHaveBeenCalled()
  })

  it('tracks an enabled promotion only after public settings have loaded', async () => {
    routeQuery.agent = 'AGENT001'
    getPublicSettingsMock.mockResolvedValue(enabledSettings)
    getPromotionTrackingStatusMock.mockResolvedValue({ enabled: true })
    trackPromotionVisitMock.mockResolvedValue({ tracked: true, attribution_days: 30 })

    mountRegisterView()
    await flushPromises()

    expect(trackPromotionVisitMock).toHaveBeenCalledOnce()
    expect(getPublicSettingsMock.mock.invocationCallOrder[0]).toBeLessThan(
      trackPromotionVisitMock.mock.invocationCallOrder[0],
    )
  })

  it('does not delay registration UI while tracking status is pending', async () => {
    routeQuery.agent = 'AGENT001'
    getPublicSettingsMock.mockResolvedValue(enabledSettings)
    getPromotionTrackingStatusMock.mockReturnValue(new Promise(() => {}))

    const wrapper = mountRegisterView()
    await flushPromises()

    expect(wrapper.find('form').exists()).toBe(true)
  })

  it('retries one transient tracking failure and stops after success', async () => {
    vi.useFakeTimers()
    routeQuery.agent = 'AGENT001'
    getPublicSettingsMock.mockResolvedValue(enabledSettings)
    getPromotionTrackingStatusMock.mockResolvedValue({ enabled: true })
    trackPromotionVisitMock
      .mockRejectedValueOnce(new Error('temporary network failure'))
      .mockResolvedValueOnce({ tracked: true, attribution_days: 30 })

    mountRegisterView()
    await flushPromises()
    expect(trackPromotionVisitMock).toHaveBeenCalledOnce()

    await vi.advanceTimersByTimeAsync(1000)
    await flushPromises()
    expect(trackPromotionVisitMock).toHaveBeenCalledTimes(2)

    await vi.advanceTimersByTimeAsync(5000)
    expect(trackPromotionVisitMock).toHaveBeenCalledTimes(2)
  })

  it('cancels the bounded tracking retry when leaving the page', async () => {
    vi.useFakeTimers()
    routeQuery.agent = 'AGENT001'
    getPublicSettingsMock.mockResolvedValue(enabledSettings)
    getPromotionTrackingStatusMock.mockResolvedValue({ enabled: true })
    trackPromotionVisitMock.mockRejectedValue(new Error('temporary network failure'))

    const wrapper = mountRegisterView()
    await flushPromises()
    expect(trackPromotionVisitMock).toHaveBeenCalledOnce()

    wrapper.unmount()
    await vi.advanceTimersByTimeAsync(1000)
    expect(trackPromotionVisitMock).toHaveBeenCalledOnce()
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

  it('shows a retryable settings error instead of reporting registration as disabled', async () => {
    getPublicSettingsMock
      .mockRejectedValueOnce(new Error('network unavailable'))
      .mockResolvedValueOnce(enabledSettings)

    const wrapper = mountRegisterView()
    await flushPromises()

    expect(wrapper.text()).toContain('auth.settingsLoadFailed')
    expect(wrapper.text()).not.toContain('auth.registrationDisabled')
    expect(wrapper.find('form').exists()).toBe(false)

    const retryButton = wrapper.findAll('button').find((button) => button.text().includes('common.retry'))
    expect(retryButton).toBeDefined()
    await retryButton!.trigger('click')
    await flushPromises()

    expect(getPublicSettingsMock).toHaveBeenCalledTimes(2)
    expect(wrapper.find('form').exists()).toBe(true)
  })

  it('keeps only the latest invitation-code validation result', async () => {
    vi.useFakeTimers()
    getPublicSettingsMock.mockResolvedValue({
      ...enabledSettings,
      invitation_code_enabled: true
    })

    let resolveOld!: (value: { valid: boolean; error_code?: string }) => void
    let resolveNew!: (value: { valid: boolean; error_code?: string }) => void
    validateInvitationCodeMock
      .mockReturnValueOnce(new Promise((resolve) => { resolveOld = resolve }))
      .mockReturnValueOnce(new Promise((resolve) => { resolveNew = resolve }))

    const wrapper = mountRegisterView()
    await flushPromises()
    const input = wrapper.get('#invitation_code')

    await input.setValue('OLD-CODE')
    await vi.advanceTimersByTimeAsync(500)
    await input.setValue('NEW-CODE')
    await vi.advanceTimersByTimeAsync(500)

    resolveNew({ valid: true })
    await flushPromises()
    resolveOld({ valid: false, error_code: 'INVITATION_CODE_INVALID' })
    await flushPromises()

    expect(wrapper.text()).toContain('auth.invitationCodeValid')
    expect(input.attributes('aria-invalid')).toBe('false')
  })

  it('keeps only the latest promo-code validation result', async () => {
    vi.useFakeTimers()
    getPublicSettingsMock.mockResolvedValue({
      ...enabledSettings,
      promo_code_enabled: true
    })

    let resolveOld!: (value: { valid: boolean; error_code?: string }) => void
    let resolveNew!: (value: { valid: boolean; bonus_amount?: number }) => void
    validatePromoCodeMock
      .mockReturnValueOnce(new Promise((resolve) => { resolveOld = resolve }))
      .mockReturnValueOnce(new Promise((resolve) => { resolveNew = resolve }))

    const wrapper = mountRegisterView()
    await flushPromises()
    const input = wrapper.get('#promo_code')

    await input.setValue('OLD-PROMO')
    await vi.advanceTimersByTimeAsync(500)
    await input.setValue('NEW-PROMO')
    await vi.advanceTimersByTimeAsync(500)

    resolveNew({ valid: true, bonus_amount: 12 })
    await flushPromises()
    resolveOld({ valid: false, error_code: 'PROMO_CODE_NOT_FOUND' })
    await flushPromises()

    expect(wrapper.text()).toContain('auth.promoCodeValid')
    expect(input.attributes('aria-invalid')).toBe('false')
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

  it('keeps the optional affiliate invitation field before Turnstile', async () => {
    getPublicSettingsMock.mockResolvedValue({
      ...enabledSettings,
      affiliate_enabled: true,
      turnstile_enabled: true,
      turnstile_site_key: 'site-key'
    })

    const wrapper = mountRegisterView()
    await flushPromises()

    const invitationField = wrapper.get('[data-testid="affiliate-invitation-field"]')
    const turnstile = wrapper.get('[data-testid="registration-turnstile"]')
    expect(invitationField.get('input').attributes('id')).toBe('affiliate_code')
    expect(invitationField.text()).toContain('common.optional')
    expect(
      invitationField.element.compareDocumentPosition(turnstile.element) &
        Node.DOCUMENT_POSITION_FOLLOWING
    ).toBeTruthy()
  })

  it('does not duplicate the affiliate field when invitation codes are mandatory', async () => {
    getPublicSettingsMock.mockResolvedValue({
      ...enabledSettings,
      affiliate_enabled: true,
      invitation_code_enabled: true
    })

    const wrapper = mountRegisterView()
    await flushPromises()

    expect(wrapper.find('[data-testid="affiliate-invitation-field"]').exists()).toBe(false)
    expect(wrapper.get('#invitation_code').exists()).toBe(true)
  })

  it('submits a non-whitelist email domain so the backend can enforce its registration quota', async () => {
    getPublicSettingsMock.mockResolvedValueOnce({
      ...enabledSettings,
      turnstile_enabled: false,
      registration_email_suffix_whitelist: ['allowed.com'],
      registration_email_domain_quota_enabled: true
    })

    const wrapper = mountRegisterView()
    await flushPromises()
    await wrapper.get('#email').setValue('first@custom.example')
    await wrapper.get('#password').setValue('secret-123')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(registerMock).toHaveBeenCalledWith(
      expect.objectContaining({ email: 'first@custom.example' })
    )
    expect(showErrorMock).not.toHaveBeenCalled()
  })

  it('shows the localized registration domain quota message returned by the backend', async () => {
    getPublicSettingsMock.mockResolvedValueOnce({
      ...enabledSettings,
      turnstile_enabled: false,
      registration_email_suffix_whitelist: ['allowed.com'],
      registration_email_domain_quota_enabled: true
    })
    registerMock.mockRejectedValueOnce({
      reason: 'EMAIL_DOMAIN_REGISTRATION_LIMIT',
      message: 'raw backend message'
    })

    const wrapper = mountRegisterView()
    await flushPromises()
    await wrapper.get('#email').setValue('second@custom.example')
    await wrapper.get('#password').setValue('secret-123')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(showErrorMock).toHaveBeenCalledWith(
      '该邮箱域名无法注册新账户。请使用主流邮箱注册；如需使用企业邮箱，请联系客服添加域名白名单。'
    )
  })

  // 域名限量注册开关默认关闭：恢复 PR5423 之前的客户端白名单预检，非白名单域名不发起注册请求。
  it('rejects a non-whitelist email domain locally when the domain quota switch is disabled', async () => {
    getPublicSettingsMock.mockResolvedValueOnce({
      ...enabledSettings,
      turnstile_enabled: false,
      registration_email_suffix_whitelist: ['allowed.com']
    })

    const wrapper = mountRegisterView()
    await flushPromises()
    await wrapper.get('#email').setValue('first@custom.example')
    await wrapper.get('#password').setValue('secret-123')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(registerMock).not.toHaveBeenCalled()
    // 校验失败通过 validationToastMessage watcher 弹 toast
    expect(showErrorMock).toHaveBeenCalledWith('auth.emailSuffixNotAllowedWithAllowed')
    expect(wrapper.get('#email').classes()).toContain('input-error')
  })

  it('still submits whitelisted email domains when the domain quota switch is disabled', async () => {
    getPublicSettingsMock.mockResolvedValueOnce({
      ...enabledSettings,
      turnstile_enabled: false,
      registration_email_suffix_whitelist: ['allowed.com']
    })

    const wrapper = mountRegisterView()
    await flushPromises()
    await wrapper.get('#email').setValue('user@allowed.com')
    await wrapper.get('#password').setValue('secret-123')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(registerMock).toHaveBeenCalledWith(
      expect.objectContaining({ email: 'user@allowed.com' })
    )
    expect(showErrorMock).not.toHaveBeenCalled()
  })
})
