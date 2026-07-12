import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import AnnouncementsView from '../AnnouncementsView.vue'

const apiMocks = vi.hoisted(() => ({
  list: vi.fn(),
  getAllGroups: vi.fn()
}))

const storeMocks = vi.hoisted(() => ({
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    announcements: {
      list: apiMocks.list,
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn()
    },
    groups: {
      getAll: apiMocks.getAllGroups
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => storeMocks
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

const emptyResponse = {
  items: [],
  total: 0,
  pages: 0,
  page: 1,
  page_size: 20
}

function mountView() {
  return mount(AnnouncementsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: {
          template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
        },
        DataTable: {
          props: ['data', 'loading'],
          template: '<div><slot v-if="!loading && data.length === 0" name="empty" /></div>'
        },
        EmptyState: {
          props: ['title', 'description', 'actionText'],
          emits: ['action'],
          template: '<div data-testid="empty-state">{{ description }}<button type="button" @click="$emit(\'action\')">{{ actionText }}</button></div>'
        },
        BaseDialog: true,
        ConfirmDialog: true,
        Pagination: true,
        Select: true,
        Icon: true,
        AnnouncementTargetingEditor: true,
        AnnouncementReadStatusDialog: true
      }
    }
  })
}

describe('AnnouncementsView list state', () => {
  beforeEach(() => {
    apiMocks.list.mockReset()
    apiMocks.getAllGroups.mockReset()
    storeMocks.showError.mockReset()
    storeMocks.showSuccess.mockReset()
    apiMocks.getAllGroups.mockResolvedValue([])
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('uses a normal empty-list description after a successful request', async () => {
    apiMocks.list.mockResolvedValue(emptyResponse)
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('[data-testid="empty-state"]').text()).toContain(
      'admin.announcements.emptyDescription'
    )
    expect(wrapper.text()).not.toContain('admin.announcements.failedToLoad')
    wrapper.unmount()
  })

  it('shows a persistent load error and retries in place', async () => {
    apiMocks.list
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce(emptyResponse)

    const wrapper = mountView()
    await flushPromises()

    const errorState = wrapper.get('[data-testid="announcements-load-error"]')
    expect(errorState.text()).toContain('admin.announcements.failedToLoad')
    expect(wrapper.find('[data-testid="empty-state"]').exists()).toBe(false)

    await errorState.get('button').trigger('click')
    await flushPromises()

    expect(apiMocks.list).toHaveBeenCalledTimes(2)
    expect(wrapper.find('[data-testid="announcements-load-error"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="empty-state"]').text()).toContain(
      'admin.announcements.emptyDescription'
    )
    wrapper.unmount()
  })
})
