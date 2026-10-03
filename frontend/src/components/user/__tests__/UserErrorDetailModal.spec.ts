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
  reject: (error: Error) => void
}

function deferred<T>(): Deferred<T> {
  let resolve!: (value: T) => void
  let reject!: (error: Error) => void
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return { promise, resolve, reject }
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

const modelDetail = (model: string) => ({ model, status_code: 500, category: 'upstream', created_at: '2026-09-20' })

async function openClosed() {
  const wrapper = mount(UserErrorDetailModal, {
    props: { show: false, errorId: 1 },
    global: { stubs: { BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /></div>' } } },
  })
  await wrapper.setProps({ show: true })
  return wrapper
}

describe('user error detail request ownership', () => {
  it('keeps the newer detail when an earlier request finishes last', async () => {
    const old = deferred<unknown>(); const current = deferred<unknown>()
    getMyErrorDetail.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const wrapper = await openClosed()
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true, errorId: 2 })
    current.resolve(modelDetail('current-model')); await flushPromises()
    old.resolve(modelDetail('old-model')); await flushPromises()
    expect(wrapper.text()).toContain('current-model')
    expect(wrapper.text()).not.toContain('old-model')
  })

  it('keeps loading until the current request settles', async () => {
    const old = deferred<unknown>(); const current = deferred<unknown>()
    getMyErrorDetail.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const wrapper = await openClosed()
    await wrapper.setProps({ errorId: 2 })
    old.resolve(modelDetail('old-model')); await flushPromises()
    expect(wrapper.find('[role="status"]').exists()).toBe(true)
    current.resolve(modelDetail('current-model')); await flushPromises()
    expect(wrapper.find('[role="status"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('current-model')
  })

  it('ignores an obsolete failure, including when reopening the same record', async () => {
    const old = deferred<unknown>()
    getMyErrorDetail.mockReturnValueOnce(old.promise).mockResolvedValueOnce(modelDetail('current-model'))
    const wrapper = await openClosed()
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true }); await flushPromises()
    old.reject(new Error('obsolete')); await flushPromises()
    expect(wrapper.text()).toContain('current-model')
    expect(wrapper.text()).not.toContain('usage.errors.detail.loadFailed')
  })

  it('still reports a failure for the current record', async () => {
    getMyErrorDetail.mockRejectedValueOnce(new Error('current failure'))
    const wrapper = await openClosed(); await flushPromises()
    expect(wrapper.text()).toContain('usage.errors.detail.loadFailed')
    expect(wrapper.find('[role="status"]').exists()).toBe(false)
  })
})
