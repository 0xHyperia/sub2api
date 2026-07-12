import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProfileAuthorizedAppsCard from '@/components/user/profile/ProfileAuthorizedAppsCard.vue'

const mocks = vi.hoisted(() => ({
  list: vi.fn(),
  revoke: vi.fn(),
  showSuccess: vi.fn(),
  showError: vi.fn()
}))

vi.mock('@/api/appAuth', () => ({
  listAuthorizationDevices: mocks.list,
  revokeAuthorizationDevice: mocks.revoke
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showSuccess: mocks.showSuccess, showError: mocks.showError })
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

describe('ProfileAuthorizedAppsCard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.list.mockResolvedValue([{
      id: 'grant-1',
      client_id: 'zerobox-desktop',
      client_name: 'ZeroBox',
      device_name: 'Workstation',
      platform: 'windows',
      scopes: ['profile:read'],
      created_at: '2026-07-01T00:00:00Z',
      last_used_at: null
    }])
    mocks.revoke.mockResolvedValue(undefined)
  })

  it('lists authorized devices and removes one after confirmation', async () => {
    const wrapper = mount(ProfileAuthorizedAppsCard, {
      global: {
        stubs: {
          ConfirmDialog: {
            props: ['show'],
            emits: ['confirm', 'cancel'],
            template: '<button v-if="show" data-testid="confirm-revoke" @click="$emit(\'confirm\')">confirm</button>'
          }
        }
      }
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Workstation')
    await wrapper.get('button[aria-label="profile.authorizedApps.revoke"]').trigger('click')
    await wrapper.get('[data-testid="confirm-revoke"]').trigger('click')
    await flushPromises()

    expect(mocks.revoke).toHaveBeenCalledWith('grant-1')
    expect(wrapper.text()).not.toContain('Workstation')
    expect(mocks.showSuccess).toHaveBeenCalledWith('profile.authorizedApps.revokeSuccess')
  })

  it('shows a persistent load error and retries in place', async () => {
    mocks.list
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce([])

    const wrapper = mount(ProfileAuthorizedAppsCard)
    await flushPromises()

    const errorState = wrapper.get('[data-testid="authorized-apps-load-error"]')
    expect(errorState.attributes('role')).toBe('alert')

    await errorState.get('button').trigger('click')
    await flushPromises()

    expect(mocks.list).toHaveBeenCalledTimes(2)
    expect(wrapper.find('[data-testid="authorized-apps-load-error"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('profile.authorizedApps.empty')
  })
})
