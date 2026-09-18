<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores'
import { opsAPI, type OpsRequestDetailsParams, type OpsRequestDetail } from '@/api/admin/ops'
import { parseTimeRangeMinutes, formatDateTime } from '../utils/opsFormatters'

export interface OpsRequestDetailsPreset {
  title: string
  kind?: OpsRequestDetailsParams['kind']
  sort?: OpsRequestDetailsParams['sort']
  min_duration_ms?: number
  max_duration_ms?: number
}

interface Props {
  modelValue: boolean
  timeRange: string
  preset: OpsRequestDetailsPreset
  platform?: string
  groupId?: number | null
  resumeState?: boolean
}

interface OpsRequestDetailMetrics extends OpsRequestDetail {
  input_tokens?: number | null
  output_tokens?: number | null
  cache_read_input_tokens?: number | null
  cache_creation_input_tokens?: number | null
  actual_cost?: number | null
  standard_cost?: number | null
  first_token_ms?: number | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'openErrorDetail', errorId: number): void
}>()

const { t, locale } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

// 与 DataTable 一致：< 768px 切换为卡片视图，避免宽表在移动端被截断。

const loading = ref(false)
const items = ref<OpsRequestDetailMetrics[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const close = () => emit('update:modelValue', false)

const showTTFT = computed(() => props.preset.sort === 'ttft_desc')
const latencyLabel = computed(() => t(showTTFT.value ? 'admin.ops.ttftLabel' : 'admin.ops.requestDetails.table.duration'))

function formatLatency(row: OpsRequestDetail): string {
  const value = showTTFT.value ? row.first_token_ms : row.duration_ms
  return typeof value === 'number' ? `${value} ms` : '-'
}

const rangeLabel = computed(() => {
  const minutes = parseTimeRangeMinutes(props.timeRange)
  if (minutes >= 60) return t('admin.ops.requestDetails.rangeHours', { n: Math.round(minutes / 60) })
  return t('admin.ops.requestDetails.rangeMinutes', { n: minutes })
})

function buildTimeParams(): Pick<OpsRequestDetailsParams, 'start_time' | 'end_time'> {
  const minutes = parseTimeRangeMinutes(props.timeRange)
  const endTime = new Date()
  const startTime = new Date(endTime.getTime() - minutes * 60 * 1000)
  return {
    start_time: startTime.toISOString(),
    end_time: endTime.toISOString()
  }
}

const fetchData = async () => {
  if (!props.modelValue) return
  loading.value = true
  try {
    const params: OpsRequestDetailsParams = {
      ...buildTimeParams(),
      page: page.value,
      page_size: pageSize.value,
      kind: props.preset.kind ?? 'all',
      sort: props.preset.sort ?? 'created_at_desc'
    }

    const platform = (props.platform || '').trim()
    if (platform) params.platform = platform
    if (typeof props.groupId === 'number' && props.groupId > 0) params.group_id = props.groupId

    if (typeof props.preset.min_duration_ms === 'number') params.min_duration_ms = props.preset.min_duration_ms
    if (typeof props.preset.max_duration_ms === 'number') params.max_duration_ms = props.preset.max_duration_ms

    const res = await opsAPI.listRequestDetails(params)
    items.value = res.items || []
    total.value = res.total || 0
  } catch (e: any) {
    console.error('[OpsRequestDetailsModal] Failed to fetch request details', e)
    appStore.showError(e?.message || t('admin.ops.requestDetails.failedToLoad'))
    items.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      if (props.resumeState) return
      page.value = 1
      pageSize.value = 10
      fetchData()
    }
  }
)

watch(
  () => [
    props.timeRange,
    props.platform,
    props.groupId,
    props.preset.kind,
    props.preset.sort,
    props.preset.min_duration_ms,
    props.preset.max_duration_ms
  ],
  () => {
    if (!props.modelValue) return
    page.value = 1
    fetchData()
  }
)

function handlePageChange(next: number) {
  page.value = next
  fetchData()
}

function handlePageSizeChange(next: number) {
  pageSize.value = next
  page.value = 1
  fetchData()
}

async function handleCopyRequestId(requestId: string) {
  const ok = await copyToClipboard(requestId, t('admin.ops.requestDetails.requestIdCopied'))
  if (ok) return
  // `useClipboard` already shows toast on failure; this keeps UX consistent with older ops modal.
  appStore.showWarning(t('admin.ops.requestDetails.copyFailed'))
}

function openErrorDetail(errorId: number | null | undefined) {
  if (!errorId) return
  emit('update:modelValue', false)
  emit('openErrorDetail', errorId)
}

const kindBadgeClass = (kind: string) => {
  if (kind === 'error') return 'bg-danger-subtle text-danger-foreground'
  return 'bg-success-subtle text-success-foreground'
}

function localText(zh: string, en: string): string {
  return locale.value.startsWith('zh') ? zh : en
}

function formatMetric(value: number | null | undefined): string {
  return typeof value === 'number' && Number.isFinite(value)
    ? new Intl.NumberFormat(locale.value).format(value)
    : '—'
}

function formatCost(value: number | null | undefined): string {
  return typeof value === 'number' && Number.isFinite(value) ? `$${value.toFixed(6)}` : '—'
}

function totalTokens(row: OpsRequestDetailMetrics): number | null {
  if (typeof row.input_tokens !== 'number' && typeof row.output_tokens !== 'number') return null
  return (row.input_tokens || 0) + (row.output_tokens || 0)
}
</script>

<template>
  <BaseDialog :show="modelValue" :title="props.preset.title || t('admin.ops.requestDetails.title')" width="full" @close="close">
    <template #default>
      <div class="flex h-full min-h-0 flex-col">
        <div class="mb-4 flex flex-shrink-0 items-center justify-between">
          <div class="text-xs text-foreground-subtle">
            {{ t('admin.ops.requestDetails.rangeLabel', { range: rangeLabel }) }}
          </div>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            @click="fetchData"
          >
            {{ t('common.refresh') }}
          </button>
        </div>

        <!-- Loading -->
        <div v-if="loading" class="flex flex-1 items-center justify-center py-16">
          <div class="flex flex-col items-center gap-3">
            <svg class="h-8 w-8 animate-spin text-info-foreground" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            <span class="text-sm font-medium text-foreground-subtle">{{ t('common.loading') }}</span>
          </div>
        </div>

        <!-- Table -->
        <div v-else class="flex min-h-0 flex-1 flex-col">
          <div v-if="items.length === 0" class="rounded-panel border border-dashed border-outline p-8 text-center">
            <div class="text-sm font-medium text-foreground-muted">{{ t('admin.ops.requestDetails.empty') }}</div>
            <div class="mt-1 text-xs text-foreground-subtle">{{ t('admin.ops.requestDetails.emptyHint') }}</div>
          </div>

          <template v-else>
            <div data-mobile-layout="request-cards" class="min-h-0 flex-1 space-y-3 overflow-y-auto sm:hidden">
              <article
                v-for="(row, idx) in items"
                :key="`mobile-${row.request_id || idx}`"
                class="rounded-panel border border-outline bg-surface p-3 shadow-card"
              >
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0">
                    <div class="flex flex-wrap items-center gap-2">
                      <span class="rounded-full px-2 py-1 text-[10px] font-bold" :class="kindBadgeClass(row.kind)">
                        {{ row.kind === 'error' ? t('admin.ops.requestDetails.kind.error') : t('admin.ops.requestDetails.kind.success') }}
                      </span>
                      <span class="text-xs tabular-nums text-foreground-subtle">{{ formatDateTime(row.created_at) }}</span>
                    </div>
                    <h4 class="mt-2 break-words text-sm font-semibold text-foreground">{{ row.model || '-' }}</h4>
                    <p class="mt-0.5 text-xs uppercase text-foreground-muted">{{ row.platform || 'unknown' }}</p>
                  </div>
                  <div class="shrink-0 text-right">
                    <div class="text-sm font-semibold tabular-nums text-foreground">
                      {{ latencyLabel }}: {{ formatLatency(row) }}
                    </div>
                    <div class="mt-1 text-xs tabular-nums text-foreground-subtle">
                      {{ t('admin.ops.requestDetails.table.status') }}: {{ row.status_code ?? '—' }}
                    </div>
                  </div>
                </div>

                <div v-if="row.request_id" class="mt-3 flex min-w-0 items-center gap-2 rounded-control bg-surface-subtle p-2.5">
                  <code class="min-w-0 flex-1 truncate font-mono text-[11px] text-foreground-muted" :title="row.request_id">
                    {{ row.request_id }}
                  </code>
                  <button
                    type="button"
                    data-mobile-action="copy-request-id"
                    class="btn btn-secondary btn-sm shrink-0"
                    @click="handleCopyRequestId(row.request_id)"
                  >
                    {{ t('admin.ops.requestDetails.copy') }}
                  </button>
                </div>

                <div class="mt-3 space-y-2 border-t border-outline pt-3">
                  <details class="group rounded-control border border-outline bg-surface-subtle/40">
                    <summary class="flex min-h-10 cursor-pointer list-none items-center justify-between gap-3 px-3 py-2 text-xs font-medium text-foreground marker:hidden">
                      {{ localText('Token 明细', 'Token details') }}
                      <span class="tabular-nums text-foreground-subtle">{{ formatMetric(totalTokens(row)) }}</span>
                    </summary>
                    <dl class="grid gap-2 border-t border-outline px-3 py-2 text-xs">
                      <div class="flex justify-between gap-3"><dt class="text-foreground-subtle">{{ localText('输入 Token', 'Input tokens') }}</dt><dd class="tabular-nums text-foreground">{{ formatMetric(row.input_tokens) }}</dd></div>
                      <div class="flex justify-between gap-3"><dt class="text-foreground-subtle">{{ localText('输出 Token', 'Output tokens') }}</dt><dd class="tabular-nums text-foreground">{{ formatMetric(row.output_tokens) }}</dd></div>
                    </dl>
                  </details>

                  <details class="group rounded-control border border-outline bg-surface-subtle/40">
                    <summary class="flex min-h-10 cursor-pointer list-none items-center justify-between gap-3 px-3 py-2 text-xs font-medium text-foreground marker:hidden">
                      {{ localText('费用明细', 'Cost details') }}
                      <span class="tabular-nums text-foreground-subtle">{{ formatCost(row.actual_cost) }}</span>
                    </summary>
                    <dl class="grid gap-2 border-t border-outline px-3 py-2 text-xs">
                      <div class="flex justify-between gap-3"><dt class="text-foreground-subtle">{{ localText('实际费用', 'Actual cost') }}</dt><dd class="tabular-nums text-foreground">{{ formatCost(row.actual_cost) }}</dd></div>
                      <div class="flex justify-between gap-3"><dt class="text-foreground-subtle">{{ localText('标准费用', 'Standard cost') }}</dt><dd class="tabular-nums text-foreground">{{ formatCost(row.standard_cost) }}</dd></div>
                    </dl>
                  </details>

                  <details class="group rounded-control border border-outline bg-surface-subtle/40">
                    <summary class="flex min-h-10 cursor-pointer list-none items-center justify-between gap-3 px-3 py-2 text-xs font-medium text-foreground marker:hidden">
                      {{ localText('缓存明细', 'Cache details') }}
                      <span class="text-foreground-subtle">{{ formatMetric(row.cache_read_input_tokens) }}</span>
                    </summary>
                    <dl class="grid gap-2 border-t border-outline px-3 py-2 text-xs">
                      <div class="flex justify-between gap-3"><dt class="text-foreground-subtle">{{ localText('缓存读取 Token', 'Cache read tokens') }}</dt><dd class="tabular-nums text-foreground">{{ formatMetric(row.cache_read_input_tokens) }}</dd></div>
                      <div class="flex justify-between gap-3"><dt class="text-foreground-subtle">{{ localText('缓存写入 Token', 'Cache write tokens') }}</dt><dd class="tabular-nums text-foreground">{{ formatMetric(row.cache_creation_input_tokens) }}</dd></div>
                    </dl>
                  </details>

                  <details class="group rounded-control border border-outline bg-surface-subtle/40">
                    <summary class="flex min-h-10 cursor-pointer list-none items-center justify-between gap-3 px-3 py-2 text-xs font-medium text-foreground marker:hidden">
                      {{ localText('耗时明细', 'Timing details') }}
                      <span class="tabular-nums text-foreground-subtle">{{ typeof row.duration_ms === 'number' ? `${row.duration_ms} ms` : '—' }}</span>
                    </summary>
                    <dl class="grid gap-2 border-t border-outline px-3 py-2 text-xs">
                      <div class="flex justify-between gap-3"><dt class="text-foreground-subtle">{{ localText('总耗时', 'Total duration') }}</dt><dd class="tabular-nums text-foreground">{{ typeof row.duration_ms === 'number' ? `${row.duration_ms} ms` : '—' }}</dd></div>
                      <div class="flex justify-between gap-3"><dt class="text-foreground-subtle">{{ localText('首 Token', 'First token') }}</dt><dd class="tabular-nums text-foreground">{{ typeof row.first_token_ms === 'number' ? `${row.first_token_ms} ms` : '—' }}</dd></div>
                    </dl>
                  </details>
                </div>

                <button
                  v-if="row.kind === 'error' && row.error_id"
                  type="button"
                  data-mobile-action="view-error"
                  class="btn btn-danger btn-sm mt-3 w-full"
                  @click="openErrorDetail(row.error_id)"
                >
                  {{ t('admin.ops.requestDetails.viewError') }}
                </button>
              </article>
            </div>

            <div data-desktop-layout="requests-table" class="hidden min-h-0 flex-1 flex-col overflow-hidden rounded-panel border border-outline sm:flex">
            <div class="min-h-0 flex-1 overflow-auto">
              <table class="min-w-[960px] divide-y divide-outline">
                <thead class="sticky top-0 z-10 bg-canvas">
                <tr>
                  <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                    {{ t('admin.ops.requestDetails.table.time') }}
                  </th>
                  <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                    {{ t('admin.ops.requestDetails.table.kind') }}
                  </th>
                  <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                    {{ t('admin.ops.requestDetails.table.platform') }}
                  </th>
                  <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                    {{ t('admin.ops.requestDetails.table.model') }}
                  </th>
                  <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                    {{ latencyLabel }}
                  </th>
                  <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                    {{ t('admin.ops.requestDetails.table.status') }}
                  </th>
                  <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                    {{ t('admin.ops.requestDetails.table.requestId') }}
                  </th>
                  <th class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                    {{ t('admin.ops.requestDetails.table.actions') }}
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-outline bg-surface">
                <tr v-for="(row, idx) in items" :key="idx" class="hover:bg-surface-subtle/50">
                  <td class="whitespace-nowrap px-4 py-3 text-xs text-foreground-muted">
                    {{ formatDateTime(row.created_at) }}
                  </td>
                  <td class="whitespace-nowrap px-4 py-3">
                    <span class="rounded-full px-2 py-1 text-[10px] font-bold" :class="kindBadgeClass(row.kind)">
                      {{ row.kind === 'error' ? t('admin.ops.requestDetails.kind.error') : t('admin.ops.requestDetails.kind.success') }}
                    </span>
                  </td>
                  <td class="whitespace-nowrap px-4 py-3 text-xs font-medium text-foreground-muted">
                    {{ (row.platform || 'unknown').toUpperCase() }}
                  </td>
                  <td class="max-w-[240px] truncate px-4 py-3 text-xs text-foreground-muted" :title="row.model || ''">
                    {{ row.model || '-' }}
                  </td>
                  <td class="whitespace-nowrap px-4 py-3 text-xs text-foreground-muted">
                    {{ formatLatency(row) }}
                  </td>
                  <td class="whitespace-nowrap px-4 py-3 text-xs text-foreground-muted">
                    {{ row.status_code ?? '-' }}
                  </td>
                  <td class="px-4 py-3">
                    <div v-if="row.request_id" class="flex items-center gap-2">
                      <span class="max-w-[220px] truncate font-mono text-[11px] text-foreground-muted" :title="row.request_id">
                        {{ row.request_id }}
                      </span>
                      <button
                        type="button"
                        class="btn btn-secondary btn-sm px-2"
                        @click="handleCopyRequestId(row.request_id)"
                      >
                        {{ t('admin.ops.requestDetails.copy') }}
                      </button>
                    </div>
                    <span v-else class="text-xs text-foreground-subtle">-</span>
                  </td>
                  <td class="whitespace-nowrap px-4 py-3 text-right">
                    <button
                      v-if="row.kind === 'error' && row.error_id"
                      type="button"
                      class="btn btn-danger btn-sm"
                      @click="openErrorDetail(row.error_id)"
                    >
                      {{ t('admin.ops.requestDetails.viewError') }}
                    </button>
                    <span v-else class="text-xs text-foreground-subtle">-</span>
                  </td>
                </tr>
              </tbody>
            </table>
            </div>

            <Pagination
              :total="total"
              :page="page"
              :page-size="pageSize"
              @update:page="handlePageChange"
              @update:pageSize="handlePageSizeChange"
            />
            </div>

            <div class="sm:hidden">
              <Pagination
                :total="total"
                :page="page"
                :page-size="pageSize"
                @update:page="handlePageChange"
                @update:pageSize="handlePageSizeChange"
              />
            </div>
          </template>
        </div>
      </div>
    </template>
  </BaseDialog>
</template>
