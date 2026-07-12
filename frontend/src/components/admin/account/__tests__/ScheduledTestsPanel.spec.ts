import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ScheduledTestsPanel from '../ScheduledTestsPanel.vue'
import type { ScheduledTestPlan, ScheduledTestResult } from '@/types'

const api = vi.hoisted(() => ({
  listByAccount: vi.fn(),
  listResults: vi.fn(),
  create: vi.fn(),
  update: vi.fn(),
  delete: vi.fn(),
}))

const notifications = vi.hoisted(() => ({
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    scheduledTests: api,
  },
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

const plan: ScheduledTestPlan = {
  id: 7,
  account_id: 3,
  model_id: 'claude-sonnet-4-20250514-with-a-long-name',
  cron_expression: '*/30 * * * *',
  enabled: true,
  max_results: 100,
  auto_recover: true,
  last_run_at: '2026-07-12T01:00:00Z',
  next_run_at: '2026-07-12T01:30:00Z',
  created_at: '2026-07-11T00:00:00Z',
  updated_at: '2026-07-11T00:00:00Z',
}

const result: ScheduledTestResult = {
  id: 11,
  plan_id: plan.id,
  status: 'success',
  response_text: 'a'.repeat(500),
  error_message: '',
  latency_ms: 128,
  started_at: '2026-07-12T01:00:00Z',
  finished_at: '2026-07-12T01:00:01Z',
  created_at: '2026-07-12T01:00:01Z',
}

const BaseDialogStub = {
  props: ['show'],
  template: '<div v-if="show"><slot /><slot name="footer" /></div>',
}

const ConfirmDialogStub = {
  props: ['show'],
  emits: ['confirm', 'cancel'],
  template: `
    <div v-if="show" data-test="confirm-dialog">
      <button type="button" data-test="confirm-delete" @click="$emit('confirm')">confirm</button>
      <button type="button" @click="$emit('cancel')">cancel</button>
    </div>
  `,
}

const mountPanel = () => mount(ScheduledTestsPanel, {
  props: {
    show: false,
    accountId: 3,
    modelOptions: [{ value: 'claude-sonnet-4-20250514', label: 'Claude Sonnet' }],
  },
  global: {
    stubs: {
      BaseDialog: BaseDialogStub,
      ConfirmDialog: ConfirmDialogStub,
      HelpTooltip: { template: '<span><slot name="trigger" /><slot /></span>' },
      Select: { template: '<button type="button">select</button>' },
      Input: { template: '<input />' },
      Toggle: { template: '<button type="button" role="switch">toggle</button>' },
      Icon: true,
    },
  },
})

describe('ScheduledTestsPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.listByAccount.mockResolvedValue([plan])
    api.listResults.mockResolvedValue([result])
    api.delete.mockResolvedValue(undefined)
  })

  it('uses keyboard-operable disclosure controls and mobile-safe surfaces', async () => {
    const wrapper = mountPanel()
    await wrapper.setProps({ show: true })
    await flushPromises()

    const addButton = wrapper.get<HTMLButtonElement>('[aria-controls="scheduled-test-add-form"]')
    expect(addButton.attributes('aria-expanded')).toBe('false')
    expect(addButton.classes()).toContain('w-full')

    const expandButton = wrapper.get<HTMLButtonElement>('[aria-controls="scheduled-test-results-7"]')
    expect(expandButton.attributes('aria-expanded')).toBe('false')
    expect(expandButton.element.closest('.rounded-panel')).not.toBeNull()

    await expandButton.trigger('click')
    await flushPromises()

    expect(api.listResults).toHaveBeenCalledWith(plan.id, 20)
    expect(expandButton.attributes('aria-expanded')).toBe('true')
    const resultToggle = wrapper.get<HTMLButtonElement>('[aria-controls="scheduled-result-detail-11"]')
    expect(resultToggle.attributes('aria-expanded')).toBe('false')
    expect(wrapper.get('#scheduled-test-results-7 .badge-success').text()).toContain('admin.scheduledTests.success')

    await resultToggle.trigger('click')
    const detail = wrapper.get('#scheduled-result-detail-11')
    expect(resultToggle.attributes('aria-expanded')).toBe('true')
    expect(detail.classes()).toContain('break-all')
    expect(detail.text()).toBe(result.response_text)

    wrapper.unmount()
  })

  it('keeps plan deletion behind the confirmation dialog', async () => {
    const wrapper = mountPanel()
    await wrapper.setProps({ show: true })
    await flushPromises()

    await wrapper.get<HTMLButtonElement>('[aria-label="admin.scheduledTests.deletePlan"]').trigger('click')
    expect(api.delete).not.toHaveBeenCalled()
    expect(wrapper.find('[data-test="confirm-dialog"]').exists()).toBe(true)

    await wrapper.get('[data-test="confirm-delete"]').trigger('click')
    await flushPromises()

    expect(api.delete).toHaveBeenCalledWith(plan.id)
    expect(wrapper.find('[aria-controls="scheduled-test-results-7"]').exists()).toBe(false)
    wrapper.unmount()
  })
})
