import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ProfileTotpCard from '@/components/user/profile/ProfileTotpCard.vue'

const mocks = vi.hoisted(() => ({
  getStatus: vi.fn(),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

vi.mock('@/api', () => ({
  totpAPI: {
    getStatus: mocks.getStatus,
  },
}))

describe('ProfileTotpCard load state', () => {
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    mocks.getStatus.mockReset()
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined)
  })

  afterEach(() => {
    consoleErrorSpy.mockRestore()
  })

  it('hides the disabled state and enable action after a failed status request, then retries', async () => {
    mocks.getStatus
      .mockRejectedValueOnce(new Error('status unavailable'))
      .mockResolvedValueOnce({ feature_enabled: true, enabled: false })

    const wrapper = mount(ProfileTotpCard, {
      global: {
        stubs: {
          Icon: true,
          TotpSetupModal: true,
          TotpDisableDialog: true,
        },
      },
    })

    await flushPromises()

    const errorState = wrapper.get('[data-testid="profile-totp-load-error"]')
    expect(errorState.text()).toContain('profile.totp.loadFailed')
    expect(wrapper.text()).not.toContain('profile.totp.notEnabled')
    expect(wrapper.text()).not.toContain('profile.totp.enable')

    await errorState.get('button').trigger('click')
    await flushPromises()

    expect(mocks.getStatus).toHaveBeenCalledTimes(2)
    expect(wrapper.find('[data-testid="profile-totp-load-error"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('profile.totp.notEnabled')
    expect(wrapper.text()).toContain('profile.totp.enable')
  })
})
