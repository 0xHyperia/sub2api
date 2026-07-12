import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import UserErrorDetailModal from '../UserErrorDetailModal.vue'

const { getMyErrorDetail } = vi.hoisted(() => ({
  getMyErrorDetail: vi.fn()
}))

vi.mock('@/api/usage', () => ({
  getMyErrorDetail
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

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

function detail(id: number, message = `message-${id}`) {
  return {
    id,
    created_at: '2026-07-12T00:00:00Z',
    model: `model-${id}`,
    inbound_endpoint: '/v1/messages',
    status_code: 500,
    category: 'upstream',
    platform: 'anthropic',
    message,
    key_name: 'test-key',
    key_deleted: false,
    error_body: `body-${id}`,
    upstream_status_code: 502
  }
}

function mountModal(errorId = 1) {
  return mount(UserErrorDetailModal, {
    props: {
      show: true,
      errorId
    },
    global: {
      stubs: {
        BaseDialog: {
          props: ['show'],
          template: '<div v-if="show"><slot /></div>'
        },
        LoadingSpinner: true,
        Icon: true
      }
    }
  })
}

describe('UserErrorDetailModal async state', () => {
  beforeEach(() => {
    getMyErrorDetail.mockReset()
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('loads when mounted open and supports an in-dialog retry', async () => {
    getMyErrorDetail
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce(detail(1))

    const wrapper = mountModal()
    await flushPromises()

    const errorState = wrapper.get('[data-testid="user-error-detail-load-error"]')
    expect(errorState.text()).toContain('usage.errors.detail.loadFailed')

    await errorState.get('button').trigger('click')
    await flushPromises()

    expect(getMyErrorDetail.mock.calls).toEqual([[1], [1]])
    expect(wrapper.text()).toContain('message-1')
    expect(wrapper.find('[data-testid="user-error-detail-load-error"]').exists()).toBe(false)
  })

  it('ignores a response for a previously selected error', async () => {
    const firstRequest = deferred<ReturnType<typeof detail>>()
    const secondRequest = deferred<ReturnType<typeof detail>>()
    getMyErrorDetail
      .mockReturnValueOnce(firstRequest.promise)
      .mockReturnValueOnce(secondRequest.promise)

    const wrapper = mountModal(1)
    await wrapper.setProps({ errorId: 2 })

    firstRequest.resolve(detail(1, 'stale message'))
    await flushPromises()
    expect(wrapper.text()).not.toContain('stale message')

    secondRequest.resolve(detail(2, 'current message'))
    await flushPromises()
    expect(wrapper.text()).toContain('current message')
    expect(wrapper.text()).not.toContain('stale message')
  })
})
