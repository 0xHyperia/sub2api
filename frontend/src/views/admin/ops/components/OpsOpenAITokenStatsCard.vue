<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { opsAPI, type OpsOpenAITokenStatsResponse, type OpsOpenAITokenStatsTimeRange } from '@/api/admin/ops'
import { formatNumber } from '@/utils/format'

interface Props {
  platformFilter?: string
  groupIdFilter?: number | null
  refreshToken: number
}

type ViewMode = 'topn' | 'pagination'

const props = withDefaults(defineProps<Props>(), {
  platformFilter: '',
  groupIdFilter: null
})

const { t } = useI18n()

// 与 DataTable 一致：< 768px 切换为卡片视图，避免宽表在移动端被截断。

const loading = ref(false)
const errorMessage = ref('')
const response = ref<OpsOpenAITokenStatsResponse | null>(null)
let requestSequence = 0

const timeRange = ref<OpsOpenAITokenStatsTimeRange>('30d')
const viewMode = ref<ViewMode>('topn')
const topN = ref<number>(20)
const page = ref<number>(1)
const pageSize = ref<number>(20)

const items = computed(() => response.value?.items ?? [])
const total = computed(() => response.value?.total ?? 0)
const totalPages = computed(() => {
  if (viewMode.value !== 'pagination') return 1
  const size = pageSize.value > 0 ? pageSize.value : 20
  return Math.max(1, Math.ceil(total.value / size))
})

const timeRangeOptions = computed(() => [
  { value: '30m', label: t('admin.ops.timeRange.30m') },
  { value: '1h', label: t('admin.ops.timeRange.1h') },
  { value: '1d', label: t('admin.ops.timeRange.1d') },
  { value: '15d', label: t('admin.ops.timeRange.15d') },
  { value: '30d', label: t('admin.ops.timeRange.30d') }
])

const viewModeOptions = computed(() => [
  { value: 'topn', label: t('admin.ops.openaiTokenStats.viewModeTopN') },
  { value: 'pagination', label: t('admin.ops.openaiTokenStats.viewModePagination') }
])

const topNOptions = computed(() => [
  { value: 10, label: 'Top 10' },
  { value: 20, label: 'Top 20' },
  { value: 50, label: 'Top 50' },
  { value: 100, label: 'Top 100' }
])

const pageSizeOptions = computed(() => [
  { value: 10, label: '10' },
  { value: 20, label: '20' },
  { value: 50, label: '50' },
  { value: 100, label: '100' }
])

function formatRate(v?: number | null): string {
  if (typeof v !== 'number' || !Number.isFinite(v)) return '-'
  return v.toFixed(2)
}

function formatInt(v?: number | null): string {
  if (typeof v !== 'number' || !Number.isFinite(v)) return '-'
  return formatNumber(Math.round(v))
}

function buildParams() {
  const params: Record<string, any> = {
    time_range: timeRange.value,
    platform: props.platformFilter || undefined,
    group_id: typeof props.groupIdFilter === 'number' && props.groupIdFilter > 0 ? props.groupIdFilter : undefined
  }

  if (viewMode.value === 'topn') {
    params.top_n = topN.value
  } else {
    params.page = page.value
    params.page_size = pageSize.value
  }
  return params
}

async function loadData() {
  const requestId = ++requestSequence
  const params = buildParams()
  const requestedMode = viewMode.value
  loading.value = true
  errorMessage.value = ''
  try {
    const nextResponse = await opsAPI.getOpenAITokenStats(params)
    if (requestId !== requestSequence) return

    // 防御：若 total 变化导致当前页超出最大页，则回退到末页；page watcher 会发起新请求。
    if (requestedMode === 'pagination') {
      const size = typeof params.page_size === 'number' && params.page_size > 0 ? params.page_size : 20
      const maxPage = Math.max(1, Math.ceil(nextResponse.total / size))
      if (typeof params.page === 'number' && params.page > maxPage) {
        page.value = maxPage
        return
      }
    }
    response.value = nextResponse
  } catch (err: any) {
    if (requestId !== requestSequence) return
    console.error('[OpsOpenAITokenStatsCard] Failed to load data', err)
    errorMessage.value = err?.response?.data?.detail || err?.message || t('admin.ops.openaiTokenStats.failedToLoad')
  } finally {
    if (requestId === requestSequence) loading.value = false
  }
}

watch(
  () => ({
    timeRange: timeRange.value,
    viewMode: viewMode.value,
    topN: topN.value,
    page: page.value,
    pageSize: pageSize.value,
    platform: props.platformFilter,
    groupId: props.groupIdFilter,
    refreshToken: props.refreshToken
  }),
  (next, prev) => {
    // 避免“筛选变化 -> 重置页码 -> 触发两次请求”：
    // 先只重置页码，等待下一次 watch（仅 page 变化）再发起请求。
    const filtersChanged = !prev ||
      next.timeRange !== prev.timeRange ||
      next.viewMode !== prev.viewMode ||
      next.pageSize !== prev.pageSize ||
      next.platform !== prev.platform ||
      next.groupId !== prev.groupId

    if (next.viewMode === 'pagination' && filtersChanged && next.page !== 1) {
      page.value = 1
      return
    }

    void loadData()
  },
  { immediate: true }
)

function onPrevPage() {
  if (viewMode.value !== 'pagination') return
  if (page.value > 1) page.value -= 1
}

function onNextPage() {
  if (viewMode.value !== 'pagination') return
  if (page.value < totalPages.value) page.value += 1
}
</script>

<template>
  <section class="card p-4 md:p-5" aria-labelledby="ops-openai-token-title">
    <div class="mb-4 flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
      <h3 id="ops-openai-token-title" class="text-sm font-semibold text-foreground">
        {{ t('admin.ops.openaiTokenStats.title') }}
      </h3>
      <div class="grid grid-cols-2 gap-2 sm:flex sm:flex-wrap sm:items-center">
        <div class="w-full sm:w-36">
          <Select v-model="timeRange" :options="timeRangeOptions" />
        </div>
        <div class="w-full sm:w-36">
          <Select v-model="viewMode" :options="viewModeOptions" />
        </div>
        <div v-if="viewMode === 'topn'" class="col-span-2 w-full sm:col-span-1 sm:w-28">
          <Select v-model="topN" :options="topNOptions" />
        </div>
        <template v-else>
          <div class="w-24">
            <Select v-model="pageSize" :options="pageSizeOptions" />
          </div>
          <button
            class="btn btn-secondary btn-sm"
            :disabled="loading || page <= 1"
            @click="onPrevPage"
          >
            {{ t('admin.ops.openaiTokenStats.prevPage') }}
          </button>
          <button
            class="btn btn-secondary btn-sm"
            :disabled="loading || page >= totalPages"
            @click="onNextPage"
          >
            {{ t('admin.ops.openaiTokenStats.nextPage') }}
          </button>
          <span class="text-xs text-foreground-subtle">
            {{ t('admin.ops.openaiTokenStats.pageInfo', { page, total: totalPages }) }}
          </span>
        </template>
      </div>
    </div>

    <div
      v-if="errorMessage && items.length > 0"
      data-testid="openai-token-refresh-error"
      role="alert"
      class="mb-4 flex flex-wrap items-center justify-between gap-3 rounded-panel border border-danger/20 bg-danger-subtle px-3 py-2 text-xs text-danger-foreground"
    >
      <span>{{ errorMessage }}</span>
      <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="loadData">
        {{ t('common.retry') }}
      </button>
    </div>

    <div v-if="loading && items.length === 0" class="py-8 text-center text-sm text-foreground-subtle">
      {{ t('admin.ops.loadingText') }}
    </div>

    <EmptyState
      v-else-if="errorMessage && items.length === 0"
      data-testid="openai-token-load-error"
      role="alert"
      :title="errorMessage"
      :action-text="t('common.retry')"
      :action-icon="false"
      @action="loadData"
    />

    <EmptyState
      v-else-if="items.length === 0"
      :title="t('common.noData')"
      :description="t('admin.ops.openaiTokenStats.empty')"
    />

    <div v-else class="space-y-3">
      <div class="overflow-hidden rounded-panel border border-outline">
        <div class="max-h-[420px] overflow-auto">
          <table class="min-w-full text-left text-xs md:text-sm">
            <thead class="sticky top-0 z-10 bg-surface">
              <tr class="border-b border-outline text-foreground-subtle">
                <th class="px-2 py-2 font-semibold">{{ t('admin.ops.openaiTokenStats.table.model') }}</th>
                <th class="px-2 py-2 font-semibold">{{ t('admin.ops.openaiTokenStats.table.requestCount') }}</th>
                <th class="px-2 py-2 font-semibold">{{ t('admin.ops.openaiTokenStats.table.avgTokensPerSec') }}</th>
                <th class="px-2 py-2 font-semibold">{{ t('admin.ops.openaiTokenStats.table.avgFirstTokenMs') }}</th>
                <th class="px-2 py-2 font-semibold">{{ t('admin.ops.openaiTokenStats.table.totalOutputTokens') }}</th>
                <th class="px-2 py-2 font-semibold">{{ t('admin.ops.openaiTokenStats.table.avgDurationMs') }}</th>
                <th class="px-2 py-2 font-semibold">{{ t('admin.ops.openaiTokenStats.table.requestsWithFirstToken') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="row in items"
                :key="row.model"
                class="border-b border-outline text-foreground-muted last:border-b-0"
              >
                <td class="px-2 py-2 font-medium">{{ row.model }}</td>
                <td class="px-2 py-2">{{ formatInt(row.request_count) }}</td>
                <td class="px-2 py-2">{{ formatRate(row.avg_tokens_per_sec) }}</td>
                <td class="px-2 py-2">{{ formatRate(row.avg_first_token_ms) }}</td>
                <td class="px-2 py-2">{{ formatInt(row.total_output_tokens) }}</td>
                <td class="px-2 py-2">{{ formatInt(row.avg_duration_ms) }}</td>
                <td class="px-2 py-2">{{ formatInt(row.requests_with_first_token) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div v-if="viewMode === 'topn'" class="mt-3 text-xs text-foreground-subtle">
        {{ t('admin.ops.openaiTokenStats.totalModels', { total }) }}
      </div>
    </div>
  </section>
</template>
