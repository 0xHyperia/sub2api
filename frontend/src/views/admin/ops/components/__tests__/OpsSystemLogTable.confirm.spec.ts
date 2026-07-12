import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import OpsSystemLogTable from '../OpsSystemLogTable.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

const {
  listSystemLogs,
  getSystemLogSinkHealth,
  getRuntimeLogConfig,
  resetRuntimeLogConfig,
  cleanupSystemLogs,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  listSystemLogs: vi.fn(),
  getSystemLogSinkHealth: vi.fn(),
  getRuntimeLogConfig: vi.fn(),
  resetRuntimeLogConfig: vi.fn(),
  cleanupSystemLogs: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    listSystemLogs,
    getSystemLogSinkHealth,
    getRuntimeLogConfig,
    resetRuntimeLogConfig,
    cleanupSystemLogs,
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({ showError, showSuccess }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const runtimeConfig = {
  level: 'info',
  enable_sampling: false,
  sampling_initial: 100,
  sampling_thereafter: 100,
  caller: true,
  stacktrace_level: 'error',
  retention_days: 30,
}

const EmptyStateStub = defineComponent({
  name: 'EmptyState',
  props: {
    title: { type: String, default: '' },
    description: { type: String, default: '' },
    actionText: { type: String, default: '' },
    actionIcon: { type: Boolean, default: true },
  },
  emits: ['action'],
  template: '<div class="empty-state-stub">{{ title }}|{{ description }}<button v-if="actionText" class="retry-action" @click="$emit(\'action\')">{{ actionText }}</button></div>',
})

function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })
  return { promise, resolve }
}

function mountTable() {
  return mount(OpsSystemLogTable, {
    global: {
      stubs: {
        Select: true,
        Pagination: true,
        ConfirmDialog: true,
        EmptyState: EmptyStateStub,
      },
    },
  })
}

function findButton(wrapper: ReturnType<typeof mountTable>, text: string) {
  const button = wrapper.findAll('button').find((item) => item.text().includes(text))
  if (!button) throw new Error(`button not found: ${text}`)
  return button
}

function activeDialog(wrapper: ReturnType<typeof mountTable>) {
  const dialog = wrapper.findAllComponents(ConfirmDialog).find((item) => item.props('show'))
  if (!dialog) throw new Error('active confirmation dialog not found')
  return dialog
}

describe('OpsSystemLogTable confirmations', () => {
  beforeEach(() => {
    for (const mock of [
      listSystemLogs,
      getSystemLogSinkHealth,
      getRuntimeLogConfig,
      resetRuntimeLogConfig,
      cleanupSystemLogs,
      showError,
      showSuccess,
    ]) {
      mock.mockReset()
    }

    listSystemLogs.mockResolvedValue({ items: [], total: 0 })
    getSystemLogSinkHealth.mockResolvedValue({
      queue_depth: 0,
      queue_capacity: 100,
      dropped_count: 0,
      write_failed_count: 0,
      written_count: 0,
      avg_write_delay_ms: 0,
    })
    getRuntimeLogConfig.mockResolvedValue(runtimeConfig)
    resetRuntimeLogConfig.mockResolvedValue(runtimeConfig)
    cleanupSystemLogs.mockResolvedValue({ deleted: 4 })
  })

  it('resets runtime logging only after a destructive confirmation', async () => {
    const wrapper = mountTable()
    await flushPromises()

    await findButton(wrapper, 'admin.ops.systemLogs.resetDefaults').trigger('click')
    let dialog = activeDialog(wrapper)
    expect(dialog.props('danger')).toBe(true)
    expect(dialog.props('message')).toBe('admin.ops.systemLogs.resetRuntimeConfigConfirm')
    expect(resetRuntimeLogConfig).not.toHaveBeenCalled()

    dialog.vm.$emit('cancel')
    await flushPromises()
    expect(resetRuntimeLogConfig).not.toHaveBeenCalled()

    await findButton(wrapper, 'admin.ops.systemLogs.resetDefaults').trigger('click')
    dialog = activeDialog(wrapper)
    dialog.vm.$emit('confirm')
    await flushPromises()

    expect(resetRuntimeLogConfig).toHaveBeenCalledOnce()
    wrapper.unmount()
  })

  it('cleans matching logs only after a destructive confirmation', async () => {
    const wrapper = mountTable()
    await flushPromises()

    await findButton(wrapper, 'admin.ops.systemLogs.cleanCurrentFilters').trigger('click')
    let dialog = activeDialog(wrapper)
    expect(dialog.props('danger')).toBe(true)
    expect(dialog.props('message')).toBe('admin.ops.systemLogs.cleanupConfirm')
    expect(cleanupSystemLogs).not.toHaveBeenCalled()

    dialog.vm.$emit('cancel')
    await flushPromises()
    expect(cleanupSystemLogs).not.toHaveBeenCalled()

    await findButton(wrapper, 'admin.ops.systemLogs.cleanCurrentFilters').trigger('click')
    dialog = activeDialog(wrapper)
    dialog.vm.$emit('confirm')
    await flushPromises()

    expect(cleanupSystemLogs).toHaveBeenCalledOnce()
    expect(showSuccess).toHaveBeenCalledWith(
      'admin.ops.systemLogs.cleanupSuccess',
    )
    wrapper.unmount()
  })

  it('keeps the newest log refresh when an older request finishes later', async () => {
    const staleRequest = deferred<{ items: Array<Record<string, unknown>>; total: number }>()
    listSystemLogs
      .mockReset()
      .mockImplementationOnce(() => staleRequest.promise)
      .mockResolvedValueOnce({
        items: [{ id: 2, created_at: '2026-01-02T00:00:00Z', level: 'info', message: 'new log' }],
        total: 1,
      })

    const wrapper = mountTable()
    await wrapper.setProps({ refreshToken: 1 })
    await flushPromises()

    expect(wrapper.text()).toContain('new log')
    staleRequest.resolve({
      items: [{ id: 1, created_at: '2026-01-01T00:00:00Z', level: 'warn', message: 'stale log' }],
      total: 1,
    })
    await flushPromises()

    expect(wrapper.text()).toContain('new log')
    expect(wrapper.text()).not.toContain('stale log')
    wrapper.unmount()
  })

  it('shows a retryable initial log error instead of the empty state', async () => {
    listSystemLogs
      .mockReset()
      .mockRejectedValueOnce(new Error('logs unavailable'))
      .mockResolvedValueOnce({
        items: [{ id: 3, created_at: '2026-01-03T00:00:00Z', level: 'error', message: 'recovered log' }],
        total: 1,
      })

    const wrapper = mountTable()
    await flushPromises()

    expect(wrapper.find('[data-testid="system-logs-load-error"]').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('admin.ops.systemLogs.empty')
    await wrapper.find('[data-testid="system-logs-load-error"] .retry-action').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('recovered log')
    expect(wrapper.find('[data-testid="system-logs-load-error"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('keeps existing logs visible when auto refresh fails', async () => {
    listSystemLogs
      .mockReset()
      .mockResolvedValueOnce({
        items: [{ id: 4, created_at: '2026-01-04T00:00:00Z', level: 'info', message: 'preserved log' }],
        total: 1,
      })
      .mockRejectedValueOnce(new Error('refresh unavailable'))

    const wrapper = mountTable()
    await flushPromises()
    await wrapper.setProps({ refreshToken: 1 })
    await flushPromises()

    expect(wrapper.text()).toContain('preserved log')
    expect(wrapper.find('[data-testid="system-logs-refresh-error"]').exists()).toBe(true)
    wrapper.unmount()
  })
})
