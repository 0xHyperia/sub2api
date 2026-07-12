import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UsageCleanupDialog from '../UsageCleanupDialog.vue'
import type { UsageCleanupTask } from '@/api/admin/usage'

const api = vi.hoisted(() => ({
  listCleanupTasks: vi.fn(),
  createCleanupTask: vi.fn(),
  cancelCleanupTask: vi.fn(),
}))

const notifications = vi.hoisted(() => ({
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin/usage', () => ({
  adminUsageAPI: api,
  default: api,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => notifications,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const task: UsageCleanupTask = {
  id: 19,
  status: 'running',
  filters: {
    start_time: '2026-07-01T00:00:00Z',
    end_time: '2026-07-02T00:00:00Z',
  },
  created_by: 1,
  deleted_rows: 120,
  created_at: '2026-07-12T00:00:00Z',
  updated_at: '2026-07-12T00:00:00Z',
}

const BaseDialogStub = {
  props: ['show'],
  emits: ['close'],
  template: `
    <div v-if="show">
      <button type="button" data-test="outer-close" @click="$emit('close')">close</button>
      <slot />
      <slot name="footer" />
    </div>
  `,
}

const ConfirmDialogStub = {
  props: ['show', 'title'],
  emits: ['confirm', 'cancel'],
  template: `
    <div v-if="show" :data-title="title">
      <button type="button" data-test="confirm" @click="$emit('confirm')">confirm</button>
      <button type="button" data-test="cancel" @click="$emit('cancel')">cancel</button>
    </div>
  `,
}

const mountDialog = () => mount(UsageCleanupDialog, {
  props: {
    show: false,
    filters: { user_id: 5, model: 'gpt-5' },
    startDate: '2026-07-01',
    endDate: '2026-07-02',
  },
  global: {
    stubs: {
      BaseDialog: BaseDialogStub,
      ConfirmDialog: ConfirmDialogStub,
      UsageFilters: {
        props: ['modelValue', 'startDate', 'endDate'],
        emits: ['update:modelValue', 'update:startDate', 'update:endDate', 'change'],
        template: '<div data-test="usage-filters" />',
      },
      Pagination: { template: '<div data-test="pagination" />' },
      Icon: true,
    },
  },
})

const findButtonByText = (wrapper: ReturnType<typeof mountDialog>, text: string) => {
  const button = wrapper.findAll('button').find(item => item.text().includes(text))
  if (!button) throw new Error(`Missing button: ${text}`)
  return button
}

describe('UsageCleanupDialog', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.clearAllMocks()
    api.listCleanupTasks.mockResolvedValue({
      items: [task],
      total: 1,
      page: 1,
      page_size: 5,
      pages: 1,
    })
    api.cancelCleanupTask.mockResolvedValue({ id: task.id, status: 'canceled' })
  })

  afterEach(() => {
    vi.clearAllTimers()
    vi.useRealTimers()
  })

  it('keeps cleanup destructive action confirmed and blocks closing while submitting', async () => {
    let resolveCreate!: (value: UsageCleanupTask) => void
    api.createCleanupTask.mockImplementation(() => new Promise(resolve => {
      resolveCreate = resolve
    }))

    const wrapper = mountDialog()
    await wrapper.setProps({ show: true })
    await flushPromises()

    expect(wrapper.get('.badge-primary').text()).toContain('admin.usage.cleanup.status.running')
    const submitButton = findButtonByText(wrapper, 'admin.usage.cleanup.submit')
    await submitButton.trigger('click')

    const confirmDialog = wrapper.get('[data-title="admin.usage.cleanup.confirmTitle"]')
    expect(api.createCleanupTask).not.toHaveBeenCalled()
    await confirmDialog.get('[data-test="confirm"]').trigger('click')
    await flushPromises()

    expect(api.createCleanupTask).toHaveBeenCalledWith(expect.objectContaining({
      start_date: '2026-07-01',
      end_date: '2026-07-02',
      user_id: 5,
      model: 'gpt-5',
      timezone: expect.any(String),
    }))
    expect(submitButton.attributes('disabled')).toBeDefined()

    await wrapper.get('[data-test="outer-close"]').trigger('click')
    expect(wrapper.emitted('close')).toBeUndefined()

    resolveCreate(task)
    await flushPromises()
    expect(notifications.showSuccess).toHaveBeenCalledWith('admin.usage.cleanup.submitSuccess')

    await wrapper.get('[data-test="outer-close"]').trigger('click')
    expect(wrapper.emitted('close')).toHaveLength(1)
    wrapper.unmount()
  })

  it('requires a second danger confirmation before canceling a running task', async () => {
    const wrapper = mountDialog()
    await wrapper.setProps({ show: true })
    await flushPromises()

    await findButtonByText(wrapper, 'admin.usage.cleanup.cancel').trigger('click')
    const confirmDialog = wrapper.get('[data-title="admin.usage.cleanup.cancelConfirmTitle"]')
    expect(api.cancelCleanupTask).not.toHaveBeenCalled()

    await confirmDialog.get('[data-test="confirm"]').trigger('click')
    await flushPromises()

    expect(api.cancelCleanupTask).toHaveBeenCalledWith(task.id)
    expect(notifications.showSuccess).toHaveBeenCalledWith('admin.usage.cleanup.cancelSuccess')
    wrapper.unmount()
  })
})
