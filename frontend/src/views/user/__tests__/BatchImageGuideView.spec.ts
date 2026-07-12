import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import type { BatchImageJob } from '@/api/batchImage'
import type { ApiKey } from '@/types'
import BatchImageGuideView from '../BatchImageGuideView.vue'

const {
  cancelBatchImageJob,
  deleteBatchImageJobRecord,
  fetchPublicSettings,
  getBatchImageItemContent,
  getBatchImageJob,
  listApiKeys,
  listBatchImageItems,
  listBatchImageJobs,
  listBatchImageModels,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  cancelBatchImageJob: vi.fn(),
  deleteBatchImageJobRecord: vi.fn(),
  fetchPublicSettings: vi.fn(),
  getBatchImageItemContent: vi.fn(),
  getBatchImageJob: vi.fn(),
  listApiKeys: vi.fn(),
  listBatchImageItems: vi.fn(),
  listBatchImageJobs: vi.fn(),
  listBatchImageModels: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api', () => ({
  keysAPI: {
    list: listApiKeys,
  },
}))

vi.mock('@/api/batchImage', () => ({
  cancelBatchImageJob,
  deleteBatchImageJobRecord,
  downloadBatchImageZip: vi.fn(),
  getBatchImageItemContent,
  getBatchImageJob,
  listBatchImageItems,
  listBatchImageJobs,
  listBatchImageModels,
  saveBlob: vi.fn(),
  submitBatchImageJob: vi.fn(),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    apiBaseUrl: 'https://api.example.test',
    fetchPublicSettings,
    showError,
    showSuccess,
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn(),
  }),
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

const AppLayoutStub = {
  template: '<main><slot /></main>',
}

const TablePageLayoutStub = {
  template: `
    <section>
      <slot name="filters" />
      <slot name="table" />
      <slot name="pagination" />
    </section>
  `,
}

const DataTableStub = {
  props: ['data'],
  template: `
    <div data-test="batch-table">
      <div data-test="select-all"><slot name="header-select" /></div>
      <div v-for="row in data" :key="row.id" data-test="batch-row">
        <slot name="cell-select" :row="row" />
        <slot name="cell-id" :row="row" />
        <slot name="cell-actions" :row="row" />
      </div>
      <slot v-if="data.length === 0" name="empty" />
    </div>
  `,
}

const BaseDialogStub = {
  props: ['show', 'title'],
  emits: ['close'],
  template: `
    <section v-if="show" data-test="base-dialog" :aria-label="title">
      <slot />
      <footer><slot name="footer" /></footer>
    </section>
  `,
}

const ConfirmDialogStub = {
  props: ['show', 'title', 'message', 'confirmText', 'cancelText', 'danger'],
  emits: ['confirm', 'cancel'],
  template: `
    <section v-if="show" data-test="confirm-dialog" :data-danger="danger">
      <h2>{{ title }}</h2>
      <p>{{ message }}</p>
      <button type="button" data-test="cancel-confirmation" @click="$emit('cancel')">{{ cancelText }}</button>
      <button type="button" data-test="confirm-action" @click="$emit('confirm')">{{ confirmText }}</button>
    </section>
  `,
}

const SelectStub = {
  props: ['modelValue', 'options'],
  emits: ['update:modelValue', 'change'],
  template: '<select :value="modelValue" @change="$emit(\'update:modelValue\', $event.target.value)"></select>',
}

const SearchInputStub = {
  props: ['modelValue'],
  emits: ['update:modelValue', 'search'],
  template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
}

const IconStub = {
  props: ['name'],
  template: '<span aria-hidden="true">{{ name }}</span>',
}

const createApiKey = (): ApiKey => ({
  id: 7,
  user_id: 1,
  key: 'sk-batch-test',
  name: 'Gemini production',
  group_id: 3,
  status: 'active',
  ip_whitelist: [],
  ip_blacklist: [],
  last_used_at: null,
  quota: 0,
  quota_used: 0,
  expires_at: null,
  created_at: '2026-07-12T00:00:00Z',
  updated_at: '2026-07-12T00:00:00Z',
  current_concurrency: 0,
  group: {
    platform: 'gemini',
    allow_batch_image_generation: true,
  } as NonNullable<ApiKey['group']>,
  rate_limit_5h: 0,
  rate_limit_1d: 0,
  rate_limit_7d: 0,
  usage_5h: 0,
  usage_1d: 0,
  usage_7d: 0,
  window_5h_start: null,
  window_1d_start: null,
  window_7d_start: null,
  reset_5h_at: null,
  reset_1d_at: null,
  reset_7d_at: null,
})

const createJob = (): BatchImageJob => ({
  id: 'job-1',
  object: 'image.batch',
  task_name: '夏季海报',
  parent_batch_id: null,
  status: 'completed',
  model: 'gemini-2.5-flash-image',
  provider: 'gemini_api',
  item_count: 1,
  success_count: 1,
  fail_count: 0,
  estimated_cost: 0.04,
  hold_amount: 0.04,
  actual_cost: 0.04,
  created_at: 1_783_814_400,
  submitted_at: 1_783_814_401,
  settled_at: 1_783_814_460,
  downloaded_at: null,
})

const mountedWrappers: Array<ReturnType<typeof mount>> = []

async function mountView() {
  const wrapper = mount(BatchImageGuideView, {
    attachTo: document.body,
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        TablePageLayout: TablePageLayoutStub,
        DataTable: DataTableStub,
        BaseDialog: BaseDialogStub,
        ConfirmDialog: ConfirmDialogStub,
        Select: SelectStub,
        SearchInput: SearchInputStub,
        Icon: IconStub,
        Teleport: true,
      },
    },
  })
  mountedWrappers.push(wrapper)
  await flushPromises()
  await nextTick()
  return wrapper
}

describe('BatchImageGuideView interactions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 1024 })
    Object.defineProperty(window, 'innerHeight', { configurable: true, value: 768 })

    fetchPublicSettings.mockResolvedValue(undefined)
    listApiKeys.mockResolvedValue({ items: [createApiKey()] })
    listBatchImageJobs.mockResolvedValue({ object: 'list', data: [createJob()], has_more: false })
    listBatchImageModels.mockResolvedValue({ object: 'list', data: [] })
    getBatchImageJob.mockResolvedValue(createJob())
    listBatchImageItems.mockResolvedValue({
      object: 'list',
      data: [{
        custom_id: 'poster-1',
        status: 'succeeded',
        prompt_preview: '明亮的夏季海报，海边日落与清晰产品主体',
        mime_type: 'image/png',
        file_extension: 'png',
        image_count: 1,
        error: null,
      }],
      has_more: false,
    })
    cancelBatchImageJob.mockResolvedValue(createJob())
    deleteBatchImageJobRecord.mockResolvedValue(undefined)
    getBatchImageItemContent.mockResolvedValue(new Blob())
  })

  afterEach(() => {
    while (mountedWrappers.length) mountedWrappers.pop()?.unmount()
    document.body.innerHTML = ''
  })

  it('names selection controls and manages the more menu as an accessible menu', async () => {
    const wrapper = await mountView()
    const selectAll = wrapper.get('input[aria-label="选择本页全部任务"]')
    const selectJob = wrapper.get('input[aria-label="选择任务：夏季海报"]')
    const trigger = wrapper.get('button[title="更多操作"]')

    expect(selectAll.attributes('type')).toBe('checkbox')
    expect(selectJob.attributes('type')).toBe('checkbox')
    expect(trigger.attributes('aria-haspopup')).toBe('menu')
    expect(trigger.attributes('aria-expanded')).toBe('false')
    expect(trigger.attributes('aria-controls')).toBe('batch-job-menu-job-1')

    await trigger.trigger('click')
    await flushPromises()

    const menu = wrapper.get('[role="menu"]')
    const firstMenuItem = menu.get('[role="menuitem"]')
    expect(menu.attributes('aria-labelledby')).toBe('batch-job-menu-trigger-job-1')
    expect(trigger.attributes('aria-expanded')).toBe('true')
    expect(document.activeElement).toBe(firstMenuItem.element)

    await menu.trigger('keydown', { key: 'Escape' })
    await flushPromises()

    expect(wrapper.find('[role="menu"]').exists()).toBe(false)
    expect(trigger.attributes('aria-expanded')).toBe('false')
    expect(document.activeElement).toBe(trigger.element)
  })

  it('requires the shared confirmation dialog before deleting a job', async () => {
    const wrapper = await mountView()

    await wrapper.get('button[title="更多操作"]').trigger('click')
    await flushPromises()
    await wrapper.get('[role="menuitem"]').trigger('click')
    await nextTick()

    const dialog = wrapper.get('[data-test="confirm-dialog"]')
    expect(dialog.text()).toContain('删除任务记录')
    expect(dialog.text()).toContain('账务记录仍会保留')
    expect(deleteBatchImageJobRecord).not.toHaveBeenCalled()

    await dialog.get('[data-test="confirm-action"]').trigger('click')
    await flushPromises()

    expect(deleteBatchImageJobRecord).toHaveBeenCalledTimes(1)
    expect(deleteBatchImageJobRecord).toHaveBeenCalledWith('sk-batch-test', 'job-1')
    expect(wrapper.find('[data-test="confirm-dialog"]').exists()).toBe(false)
  })

  it('keeps detail actions usable at 320px and bounds the prompt popover', async () => {
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 320 })
    const wrapper = await mountView()

    await wrapper.get('button[title="查看详情"]').trigger('click')
    await flushPromises()

    const footer = wrapper.get('.batch-detail-actions')
    expect(footer.classes()).toContain('flex-col-reverse')
    for (const button of footer.findAll('button')) {
      expect(button.classes()).toContain('w-full')
      expect(button.classes()).toContain('sm:w-auto')
    }

    const promptTrigger = wrapper.get('.batch-prompt-trigger')
    expect(promptTrigger.classes()).toContain('focus-visible:ring-2')
    await promptTrigger.trigger('click')
    await nextTick()

    const popover = wrapper.get('.batch-prompt-popover')
    const style = (popover.element as HTMLElement).style
    expect(style.width).toBe('304px')
    expect(style.maxWidth).toBe('calc(100vw - 16px)')
  })
})
