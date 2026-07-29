import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProfileAuthorizedAppsCard from '@/components/user/profile/ProfileAuthorizedAppsCard.vue'

const mocks = vi.hoisted(() => ({
  list: vi.fn(),
  listSessions: vi.fn(),
  renameSession: vi.fn(),
  revoke: vi.fn(),
  revokeSession: vi.fn(),
  revokeOthers: vi.fn(),
  showSuccess: vi.fn(),
  showError: vi.fn()
}))

vi.mock('@/api/appAuth', () => ({
  listAuthorizationGrants: mocks.list,
  listAuthorizationSessions: mocks.listSessions,
  renameAuthorizationSession: mocks.renameSession,
  revokeAuthorizationGrant: mocks.revoke,
  revokeAuthorizationSession: mocks.revokeSession,
  revokeOtherAuthorizationSessions: mocks.revokeOthers
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
      client_id: 'zeroagent-desktop',
      client_name: 'ZeroAgent',
      platform: 'windows',
      scopes: ['profile:read'],
      session_count: 2,
      first_authorized_at: '2026-07-01T00:00:00Z',
      last_authorized_at: '2026-07-02T00:00:00Z',
      last_used_at: null
    }])
    mocks.revoke.mockResolvedValue(undefined)
    mocks.listSessions.mockResolvedValue([{
      id: 'session-1',
      device_name: 'Office PC',
      platform: 'windows',
      created_at: '2026-07-01T00:00:00Z',
      last_used_at: '2026-07-02T00:00:00Z'
    }])
    mocks.renameSession.mockResolvedValue(undefined)
    mocks.revokeSession.mockResolvedValue(undefined)
    mocks.revokeOthers.mockResolvedValue(1)
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

    expect(wrapper.text()).toContain('ZeroAgent')
    await wrapper.get('button[aria-label="profile.authorizedApps.revoke"]').trigger('click')
    await wrapper.get('[data-testid="confirm-revoke"]').trigger('click')
    await flushPromises()

    expect(mocks.revoke).toHaveBeenCalledWith('grant-1')
    expect(wrapper.text()).not.toContain('ZeroAgent')
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

  it('expands and renames an application session', async () => {
    const wrapper = mount(ProfileAuthorizedAppsCard)
    await flushPromises()

    await wrapper.get('button[aria-label="profile.authorizedApps.manageDevices"]').trigger('click')
    await flushPromises()

    expect(mocks.listSessions).toHaveBeenCalledWith('grant-1')
    expect(wrapper.text()).toContain('Office PC')

    await wrapper.get('button[aria-label="profile.authorizedApps.rename"]').trigger('click')
    await wrapper.get('input').setValue('Home PC')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(mocks.renameSession).toHaveBeenCalledWith('session-1', 'Home PC')
    expect(wrapper.text()).toContain('Home PC')
  })
})
