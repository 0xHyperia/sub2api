import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const mocks = vi.hoisted(() => ({
  listRequestDetails: vi.fn(),
  copyToClipboard: vi.fn(),
  showError: vi.fn(),
  showWarning: vi.fn(),
}))

vi.mock('@/api/admin/ops', () => ({
  opsAPI: { listRequestDetails: mocks.listRequestDetails },
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard: mocks.copyToClipboard }),
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({ showError: mocks.showError, showWarning: mocks.showWarning }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      locale: { value: 'zh-CN' },
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/components/common/BaseDialog.vue', () => ({
  default: {
    props: ['show', 'title', 'width'],
    template: '<div v-if="show"><slot /></div>',
  },
}))

vi.mock('@/components/common/Pagination.vue', () => ({
  default: {
    props: ['total', 'page', 'pageSize'],
    template: '<div data-pagination />',
  },
}))

import OpsRequestDetailsModal from '../OpsRequestDetailsModal.vue'

const request = {
  kind: 'error',
  created_at: '2026-07-19T01:02:03Z',
  request_id: 'req-mobile-1',
  platform: 'openai',
  model: 'gpt-5',
  duration_ms: 1320,
  first_token_ms: 210,
  status_code: 500,
  error_id: 42,
  input_tokens: 1200,
  output_tokens: 300,
  cache_read_input_tokens: 800,
  cache_creation_input_tokens: 40,
  actual_cost: 0.012345,
  standard_cost: 0.023456,
}

async function mountAndOpen() {
  const wrapper = mount(OpsRequestDetailsModal, {
    props: {
      modelValue: false,
      timeRange: '30m',
      preset: { title: 'Slow requests' },
    },
  })
  await wrapper.setProps({ modelValue: true })
  await flushPromises()
  return wrapper
}

beforeEach(() => {
  vi.clearAllMocks()
  mocks.listRequestDetails.mockResolvedValue({ items: [request], total: 1 })
  mocks.copyToClipboard.mockResolvedValue(true)
})

describe('OpsRequestDetailsModal responsive details', () => {
  it('renders mobile request cards with expandable metric sections and retains the desktop table', async () => {
    const wrapper = await mountAndOpen()
    const mobile = wrapper.get('[data-mobile-layout="request-cards"]')

    expect(mobile.findAll('article')).toHaveLength(1)
    expect(mobile.findAll('details')).toHaveLength(4)
    expect(mobile.text()).toContain('1,500')
    expect(mobile.text()).toContain('$0.012345')
    expect(mobile.text()).toContain('1320 ms')
    expect(wrapper.get('[data-desktop-layout="requests-table"]').classes()).toContain('sm:flex')
    expect(wrapper.get('table').classes()).toContain('min-w-[960px]')
  })

  it('keeps copy and error-detail actions available on mobile', async () => {
    const wrapper = await mountAndOpen()

    await wrapper.get('[data-mobile-action="copy-request-id"]').trigger('click')
    expect(mocks.copyToClipboard).toHaveBeenCalledWith(
      'req-mobile-1',
      'admin.ops.requestDetails.requestIdCopied'
    )

    await wrapper.get('[data-mobile-action="view-error"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')).toContainEqual([false])
    expect(wrapper.emitted('openErrorDetail')).toContainEqual([42])
  })
})
