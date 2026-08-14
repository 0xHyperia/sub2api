<template>
  <AppLayout>
    <div class="usage-page">
      <UsageStatsCards :stats="usageStats" :show-account-cost="false" :strike-standard-cost="true" />

      <section class="usage-analytics" aria-labelledby="usage-analytics-title">
        <div class="usage-chart-toolbar">
          <div class="usage-toolbar-copy">
            <p>{{ t('usage.tabs.usage') }}</p>
            <h2 id="usage-analytics-title">{{ t('admin.dashboard.tokenUsageTrend') }}</h2>
          </div>
          <div class="usage-chart-controls">
            <div class="usage-control-group">
              <span class="usage-control-label">{{ t('admin.dashboard.timeRange') }}</span>
              <DateRangePicker
                v-model:start-date="startDate"
                v-model:end-date="endDate"
                @change="onDateRangeChange"
              />
            </div>
            <div class="usage-control-group usage-granularity-control">
              <span class="usage-control-label">{{ t('admin.dashboard.granularity') }}</span>
              <div class="w-28">
                <Select v-model="granularity" :options="granularityOptions" @change="loadChartData" />
              </div>
            </div>
          </div>
        </div>

        <div class="grid min-w-0 grid-cols-1 gap-4 lg:grid-cols-2">
          <ModelDistributionChart
            v-model:metric="modelDistributionMetric"
            :model-stats="requestedModelStats"
            :loading="modelStatsLoading"
            :show-source-toggle="false"
            :show-metric-toggle="true"
            :enable-breakdown="false"
            :show-account-cost="false"
            :start-date="startDate"
            :end-date="endDate"
          />
          <GroupDistributionChart
            v-model:metric="groupDistributionMetric"
            :group-stats="groupStats"
            :loading="chartsLoading"
            :show-metric-toggle="true"
            :enable-breakdown="false"
            :show-account-cost="false"
            :start-date="startDate"
            :end-date="endDate"
          />
        </div>

        <div class="grid min-w-0 grid-cols-1 gap-4 lg:grid-cols-2">
          <EndpointDistributionChart
            v-model:source="endpointDistributionSource"
            v-model:metric="endpointDistributionMetric"
            :endpoint-stats="inboundEndpointStats"
            :upstream-endpoint-stats="upstreamEndpointStats"
            :endpoint-path-stats="endpointPathStats"
            :loading="endpointStatsLoading"
            :show-source-toggle="false"
            :show-metric-toggle="true"
            :enable-breakdown="false"
            :title="t('usage.endpointDistribution')"
            :start-date="startDate"
            :end-date="endDate"
          />
          <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
        </div>
      </section>

      <section class="usage-records" aria-labelledby="usage-records-title">
        <h2 id="usage-records-title" class="sr-only">{{ t('usage.tabs.usage') }}</h2>
        <div class="usage-mobile-records-toolbar sm:hidden">
          <div class="flex min-w-0 items-center justify-between gap-2">
            <div
              class="grid min-w-0 flex-1 rounded-control bg-surface-subtle p-1"
              :class="errorViewEnabled ? 'grid-cols-2' : 'grid-cols-1'"
              role="tablist"
              :aria-label="t('usage.tabs.usage')"
            >
              <button
                id="usage-tab-mobile"
                type="button"
                role="tab"
                class="usage-mobile-tab"
                :class="activeTab === 'usage' ? 'usage-mobile-tab-active' : ''"
                :aria-selected="activeTab === 'usage'"
                aria-controls="usage-panel"
                @click="activateUsageTab('usage')"
              >
                {{ t('usage.tabs.usage') }}
              </button>
              <button
                v-if="errorViewEnabled"
                id="errors-tab-mobile"
                type="button"
                role="tab"
                class="usage-mobile-tab"
                :class="activeTab === 'errors' ? 'usage-mobile-tab-active' : ''"
                :aria-selected="activeTab === 'errors'"
                aria-controls="errors-panel"
                @click="activateUsageTab('errors')"
              >
                {{ t('usage.tabs.errors') }}
              </button>
            </div>
            <button
              type="button"
              class="btn btn-secondary relative h-10 shrink-0 px-3"
              :aria-label="t('common.filter')"
              :aria-expanded="mobileFiltersOpen"
              @click="mobileFiltersOpen = true"
            >
              <Icon name="filter" size="sm" />
              <span>{{ t('common.filter') }}</span>
              <span v-if="activeFilterChips.length" class="flex h-4 min-w-4 items-center justify-center rounded-full bg-brand px-1 text-[9px] font-semibold tabular-nums text-brand-foreground">
                {{ activeFilterChips.length }}
              </span>
            </button>
            <button
              type="button"
              class="btn btn-secondary btn-icon h-10 w-10 shrink-0"
              :disabled="activeTab === 'errors' ? errorLoading : loading"
              :title="t('common.refresh')"
              :aria-label="t('common.refresh')"
              @click="refreshData"
            >
              <Icon name="refresh" size="sm" :class="(activeTab === 'errors' ? errorLoading : loading) ? 'animate-spin' : ''" />
            </button>
          </div>
          <div v-if="activeFilterChips.length" class="-mx-1 mt-2 flex gap-1.5 overflow-x-auto px-1 pb-0.5" :aria-label="t('common.filter')">
            <button
              v-for="chip in activeFilterChips"
              :key="chip.key"
              type="button"
              class="inline-flex h-7 shrink-0 items-center gap-1 rounded-control border border-outline bg-surface-subtle px-2 text-[10px] text-foreground-muted"
              :aria-label="`${t('common.clear')} ${chip.label}`"
              @click="removeActiveFilter(chip.key)"
            >
              <span class="max-w-40 truncate">{{ chip.label }}</span>
              <Icon name="x" size="xs" aria-hidden="true" />
            </button>
          </div>
        </div>

        <div class="usage-filters hidden sm:block">
          <div class="usage-filter-row">
            <div v-if="activeTab === 'errors'" class="usage-filter-fields">
            <div class="w-full sm:w-auto sm:min-w-[220px]">
              <label class="input-label">{{ t('usage.errors.keyName') }}</label>
              <Select v-model="errorFilter.api_key_id" :options="errorKeyOptions" @change="applyErrorFilters" />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[220px]">
              <label class="input-label">{{ t('usage.errors.model') }}</label>
              <Select
                v-model="errorFilter.model"
                :options="errorModelOptions"
                searchable
                creatable
                clearable
                :placeholder="t('usage.errors.modelPlaceholder')"
                @change="applyErrorFilters"
              />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[200px]">
              <label class="input-label">{{ t('usage.errors.category') }}</label>
              <Select v-model="errorFilter.category" :options="errorCategoryOptions" @change="applyErrorFilters" />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[180px]">
              <label class="input-label">{{ t('usage.errors.status') }}</label>
              <Select v-model="errorFilter.status_code" :options="errorStatusOptions" @change="applyErrorFilters" />
            </div>
          </div>
            <div v-else class="usage-filter-fields">
            <div class="w-full sm:w-auto sm:min-w-[220px]">
              <label class="input-label">{{ t('usage.apiKeyFilter') }}</label>
              <Select v-model="filters.api_key_id" :options="apiKeyOptions" @change="applyFilters" />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[220px]">
              <label class="input-label">{{ t('usage.model') }}</label>
              <Select v-model="filters.model" :options="modelOptions" searchable @change="applyFilters" />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[200px]">
              <label class="input-label">{{ t('admin.usage.group') }}</label>
              <Select v-model="filters.group_id" :options="groupOptions" searchable @change="applyFilters" />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[180px]">
              <label class="input-label">{{ t('usage.type') }}</label>
              <Select v-model="filters.request_type" :options="requestTypeOptions" @change="applyFilters" />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[200px]">
              <label class="input-label">{{ t('admin.usage.billingType') }}</label>
              <Select v-model="filters.billing_type" :options="billingTypeOptions" @change="applyFilters" />
            </div>
            <div class="w-full sm:w-auto sm:min-w-[200px]">
              <label class="input-label">{{ t('admin.usage.billingMode') }}</label>
              <Select v-model="filters.billing_mode" :options="billingModeOptions" @change="applyFilters" />
            </div>
          </div>

            <div class="usage-filter-actions">
            <button
              type="button"
              @click="refreshData"
              :disabled="activeTab === 'errors' ? errorLoading : loading"
              class="btn btn-secondary"
              :aria-label="t('common.refresh')"
              :title="t('common.refresh')"
            >
              <Icon name="refresh" size="sm" :class="(activeTab === 'errors' ? errorLoading : loading) ? 'animate-spin' : ''" />
              {{ t('common.refresh') }}
            </button>
            <button type="button" @click="resetFilters" class="btn btn-secondary">
              {{ t('common.reset') }}
            </button>
            <div class="relative" ref="columnDropdownRef">
              <button
                :id="columnMenuTriggerId"
                ref="columnMenuTriggerRef"
                type="button"
                @click="toggleColumnMenu"
                @keydown="handleColumnTriggerKeydown"
                class="btn btn-secondary px-2 md:px-3"
                :title="t('admin.users.columnSettings')"
                :aria-label="t('admin.users.columnSettings')"
                :aria-expanded="showColumnDropdown"
                :aria-controls="showColumnDropdown ? columnMenuId : undefined"
                aria-haspopup="menu"
              >
                <Icon name="grid" size="sm" />
                <span class="hidden md:inline">{{ t('admin.users.columnSettings') }}</span>
              </button>
              <div
                v-if="showColumnDropdown"
                :id="columnMenuId"
                ref="columnMenuRef"
                role="menu"
                :aria-labelledby="columnMenuTriggerId"
                class="absolute right-0 top-full z-50 mt-1 max-h-80 w-52 overflow-y-auto rounded-panel border border-outline bg-surface py-1 shadow-floating"
                @keydown="handleColumnMenuKeydown"
              >
                <button
                  v-for="col in currentToggleableColumns"
                  :key="col.key"
                  type="button"
                  role="menuitemcheckbox"
                  :aria-checked="isCurrentColumnVisible(col.key)"
                  @click="toggleCurrentColumn(col.key)"
                  class="flex min-h-10 w-full items-center justify-between px-3 py-2 text-left text-sm text-foreground-muted hover:bg-surface-subtle hover:text-foreground"
                >
                  <span>{{ col.label }}</span>
                  <Icon v-if="isCurrentColumnVisible(col.key)" name="check" size="sm" class="text-brand" />
                </button>
              </div>
            </div>
            <button v-if="activeTab !== 'errors'" type="button" @click="exportToCSV" :disabled="exporting" class="btn btn-primary">
              <Icon name="download" size="sm" />
              {{ exporting ? t('usage.exporting') : t('usage.exportCsv') }}
            </button>
            </div>
          </div>
        </div>

        <div class="usage-records-header hidden sm:block">
          <div v-if="errorViewEnabled" class="usage-tabs" role="tablist" :aria-label="t('usage.tabs.usage')">
            <button
              id="usage-tab"
              type="button"
              role="tab"
              class="usage-tab"
              :class="{ 'usage-tab-active': activeTab === 'usage' }"
              :aria-selected="activeTab === 'usage'"
              :tabindex="activeTab === 'usage' ? 0 : -1"
              aria-controls="usage-panel"
              @click="activeTab = 'usage'"
              @keydown="handleUsageTabKeydown($event, 'usage')"
            >
              {{ t('usage.tabs.usage') }}
            </button>
            <button
              id="errors-tab"
              type="button"
              role="tab"
              class="usage-tab"
              :class="{ 'usage-tab-active': activeTab === 'errors' }"
              :aria-selected="activeTab === 'errors'"
              :tabindex="activeTab === 'errors' ? 0 : -1"
              aria-controls="errors-panel"
              @click="switchToErrors"
              @keydown="handleUsageTabKeydown($event, 'errors')"
            >
              {{ t('usage.tabs.errors') }}
            </button>
          </div>
          <div v-else class="usage-tabs" aria-hidden="true">
            <span class="usage-tab usage-tab-active">{{ t('usage.tabs.usage') }}</span>
          </div>
        </div>

        <div
          id="usage-panel"
          v-show="activeTab === 'usage'"
          class="usage-table-panel"
          role="tabpanel"
          :aria-labelledby="errorViewEnabled ? 'usage-tab' : undefined"
        >
          <UsageTable
            :data="usageLogs"
            :loading="loading"
            :columns="visibleColumns"
            :server-side-sort="true"
            :show-account-billing="false"
            :show-upstream-endpoint="false"
            user-mobile-card
            default-sort-key="created_at"
            default-sort-order="desc"
            @sort="handleSort"
            @ipGeoBatchFailed="handleIpGeoBatchFailed"
          />

          <Pagination
            v-if="pagination.total > 0"
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="handlePageChange"
            @update:pageSize="handlePageSizeChange"
          />
        </div>

        <div
          v-if="errorViewEnabled"
          id="errors-panel"
          v-show="activeTab === 'errors'"
          class="usage-table-panel"
          role="tabpanel"
          aria-labelledby="errors-tab"
        >
          <UserErrorRequestsTable
            :rows="errorRows"
            :total="errorTotal"
            :loading="errorLoading"
            :page="errorPage"
            :page-size="errorPageSize"
            :visible-column-keys="errVisibleColumnKeys"
            @sort="onErrorSort"
            @update:page="onErrorPage"
            @update:pageSize="onErrorPageSize"
            @ipGeoBatchFailed="handleIpGeoBatchFailed"
          />
        </div>
      </section>
    </div>

    <UsageFilterDrawer
      :open="mobileFiltersOpen"
      :mode="activeTab"
      :usage-filter="mobileUsageFilterValue"
      :error-filter="mobileErrorFilterValue"
      :api-key-options="apiKeyOptions"
      :group-options="groupOptions"
      :model-options="modelOptions"
      :request-type-options="requestTypeOptions"
      :billing-type-options="billingTypeOptions"
      :billing-mode-options="billingModeOptions"
      :error-key-options="errorKeyOptions"
      :error-model-options="errorModelOptions"
      :error-category-options="errorCategoryOptions"
      :error-status-options="errorStatusOptions"
      @close="mobileFiltersOpen = false"
      @apply-usage="applyMobileUsageFilters"
      @apply-errors="applyMobileErrorFilters"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { keysAPI, usageAPI, userGroupsAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import UsageStatsCards from '@/components/admin/usage/UsageStatsCards.vue'
import UsageTable from '@/components/admin/usage/UsageTable.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import GroupDistributionChart from '@/components/charts/GroupDistributionChart.vue'
import EndpointDistributionChart from '@/components/charts/EndpointDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import Icon from '@/components/icons/Icon.vue'
import UserErrorRequestsTable from '@/components/user/UserErrorRequestsTable.vue'
import UsageFilterDrawer from '@/components/user/UsageFilterDrawer.vue'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { useDropdownMenu } from '@/composables/useDropdownMenu'
import { formatReasoningEffort } from '@/utils/format'
import { getBillingModeLabel, getDisplayBillingMode as resolveDisplayBillingMode } from '@/utils/billingMode'
import { resolveUsageRequestType, requestTypeToLegacyStream } from '@/utils/usageRequestType'
import type {
  ApiKey,
  EndpointStat,
  Group,
  GroupStat,
  ModelStat,
  TrendDataPoint,
  UsageLog,
  UsageQueryParams,
  UsageStatsResponse,
  UserErrorRequest,
} from '@/types'
import type { Column } from '@/components/common/types'
import { COMMON_ERROR_STATUS_CODES } from '@/utils/errorBadges'

const { t } = useI18n()
const appStore = useAppStore()

type DistributionMetric = 'tokens' | 'actual_cost'
type EndpointSource = 'inbound' | 'upstream' | 'path'

const usageStats = ref<UsageStatsResponse | null>(null)
const usageLogs = ref<UsageLog[]>([])
const trendData = ref<TrendDataPoint[]>([])
const requestedModelStats = ref<ModelStat[]>([])
const groupStats = ref<GroupStat[]>([])
const inboundEndpointStats = ref<EndpointStat[]>([])
const upstreamEndpointStats = ref<EndpointStat[]>([])
const endpointPathStats = ref<EndpointStat[]>([])

const loading = ref(false)
const chartsLoading = ref(false)
const modelStatsLoading = ref(false)
const endpointStatsLoading = ref(false)
const exporting = ref(false)
const errorRows = ref<UserErrorRequest[]>([])
const errorLoading = ref(false)
const errorPage = ref(1)
const errorPageSize = ref(20)
const errorSortBy = ref('created_at')
const errorSortOrder = ref<'asc' | 'desc'>('desc')
const errorTotal = ref(0)
const errorFilter = ref<{ model: string | null; category: string; api_key_id: number | null; status_code: number | null }>({
  model: '',
  category: '',
  api_key_id: null,
  status_code: null,
})

const errorKeyOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('usage.errors.allKeys') },
  ...apiKeys.value.map((k) => ({ value: k.id, label: k.name })),
])

// 模型候选取自当前已加载错误中出现过的模型；creatable 允许输入任意片段做后端模糊。
const errorModelOptions = computed<SelectOption[]>(() => {
  const seen = new Set<string>()
  const opts: SelectOption[] = []
  for (const r of errorRows.value) {
    if (r.model && !seen.has(r.model)) {
      seen.add(r.model)
      opts.push({ value: r.model, label: r.model })
    }
  }
  return opts
})

const errorCategoryCodes = ['auth', 'rate_limit', 'quota', 'invalid_request', 'service_unavailable', 'upstream', 'internal', 'cyber']

const errorCategoryOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('usage.errors.allCategories') },
  ...errorCategoryCodes.map((c) => ({ value: c, label: t('usage.errors.categories.' + c) })),
])

// 状态码候选用固定常用列表(与管理端 UsageFilters 共用常量),不受当前页数据限制:
// 后端 status_code 过滤对全量生效,若只列当前页出现过的码,用户就选不到仅在后续页的码。
const errorStatusOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('usage.errors.allStatuses') },
  ...COMMON_ERROR_STATUS_CODES.map((c) => ({ value: c, label: String(c) })),
])

const applyErrorFilters = () => {
  errorPage.value = 1
  void loadErrors()
}

let abortController: AbortController | null = null
let chartReqSeq = 0
let statsReqSeq = 0
let modelStatsReqSeq = 0

const formatLocalDate = (date: Date): string =>
  `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`

const getLast24HoursRangeDates = () => {
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  return { start: formatLocalDate(start), end: formatLocalDate(end) }
}

const getGranularityForRange = (start: string, end: string): 'day' | 'hour' => {
  const startTime = new Date(`${start}T00:00:00`).getTime()
  const endTime = new Date(`${end}T00:00:00`).getTime()
  return Math.ceil((endTime - startTime) / (1000 * 60 * 60 * 24)) <= 1 ? 'hour' : 'day'
}

const defaultRange = getLast24HoursRangeDates()
const startDate = ref(defaultRange.start)
const endDate = ref(defaultRange.end)
const granularity = ref<'day' | 'hour'>(getGranularityForRange(startDate.value, endDate.value))

const modelDistributionMetric = ref<DistributionMetric>('tokens')
const groupDistributionMetric = ref<DistributionMetric>('tokens')
const endpointDistributionMetric = ref<DistributionMetric>('tokens')
const endpointDistributionSource = ref<EndpointSource>('inbound')
const activeTab = ref<'usage' | 'errors'>('usage')
const mobileFiltersOpen = ref(false)
const errorViewEnabled = computed(() => appStore.cachedPublicSettings?.allow_user_view_error_requests ?? false)

const filters = ref<UsageQueryParams>({
  start_date: startDate.value,
  end_date: endDate.value,
  request_type: undefined,
  billing_type: null,
  billing_mode: null,
})

const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
})
const sortState = reactive({
  sort_by: 'created_at',
  sort_order: 'desc' as 'asc' | 'desc',
})

const granularityOptions = computed<SelectOption[]>(() => [
  { value: 'day', label: t('admin.dashboard.day') },
  { value: 'hour', label: t('admin.dashboard.hour') },
])
const requestTypeOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allTypes') },
  { value: 'ws_v2', label: t('usage.ws') },
  { value: 'live', label: t('usage.live') },
  { value: 'stream', label: t('usage.stream') },
  { value: 'sync', label: t('usage.sync') },
])
const billingTypeOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allBillingTypes') },
  { value: 0, label: t('admin.usage.billingTypeBalance') },
  { value: 1, label: t('admin.usage.billingTypeSubscription') },
])
const billingModeOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allBillingModes') },
  { value: 'token', label: t('admin.usage.billingModeToken') },
  { value: 'per_request', label: t('admin.usage.billingModePerRequest') },
  { value: 'image', label: t('admin.usage.billingModeImage') },
  { value: 'video', label: t('admin.usage.billingModeVideo') },
])

const apiKeys = ref<ApiKey[]>([])
const groups = ref<Group[]>([])
const modelOptionValues = ref<string[]>([])

const apiKeyOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('usage.allApiKeys') },
  ...apiKeys.value.map((key) => ({ value: key.id, label: key.name })),
])
const groupOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allGroups') },
  ...groups.value.map((group) => ({ value: group.id, label: group.name })),
])
const modelOptions = computed<SelectOption[]>(() => [
  { value: null, label: t('admin.usage.allModels') },
  ...modelOptionValues.value.map((model) => ({ value: model, label: model })),
])

interface ActiveFilterChip {
  key: string
  label: string
}

const findOptionLabel = (options: SelectOption[], value: unknown): string =>
  options.find(option => String(option.value) === String(value))?.label ?? String(value ?? '')

const mobileUsageFilterValue = computed(() => ({
  api_key_id: filters.value.api_key_id ?? null,
  model: filters.value.model ?? null,
  group_id: filters.value.group_id ?? null,
  request_type: filters.value.request_type ?? null,
  billing_type: filters.value.billing_type ?? null,
  billing_mode: filters.value.billing_mode ?? null,
}))

const mobileErrorFilterValue = computed(() => ({
  api_key_id: errorFilter.value.api_key_id ?? null,
  model: errorFilter.value.model ?? null,
  category: errorFilter.value.category || '',
  status_code: errorFilter.value.status_code ?? null,
}))

const activeFilterChips = computed<ActiveFilterChip[]>(() => {
  if (activeTab.value === 'errors') {
    const chips: ActiveFilterChip[] = []
    if (errorFilter.value.api_key_id != null) chips.push({ key: 'error-api-key', label: `${t('usage.errors.keyName')}: ${findOptionLabel(errorKeyOptions.value, errorFilter.value.api_key_id)}` })
    if (errorFilter.value.model) chips.push({ key: 'error-model', label: `${t('usage.errors.model')}: ${errorFilter.value.model}` })
    if (errorFilter.value.category) chips.push({ key: 'error-category', label: `${t('usage.errors.category')}: ${findOptionLabel(errorCategoryOptions.value, errorFilter.value.category)}` })
    if (errorFilter.value.status_code != null) chips.push({ key: 'error-status', label: `${t('usage.errors.status')}: ${errorFilter.value.status_code}` })
    return chips
  }

  const chips: ActiveFilterChip[] = []
  if (filters.value.api_key_id != null) chips.push({ key: 'api-key', label: `${t('usage.apiKeyFilter')}: ${findOptionLabel(apiKeyOptions.value, filters.value.api_key_id)}` })
  if (filters.value.model) chips.push({ key: 'model', label: `${t('usage.model')}: ${filters.value.model}` })
  if (filters.value.group_id != null) chips.push({ key: 'group', label: `${t('admin.usage.group')}: ${findOptionLabel(groupOptions.value, filters.value.group_id)}` })
  if (filters.value.request_type) chips.push({ key: 'request-type', label: `${t('usage.type')}: ${findOptionLabel(requestTypeOptions.value, filters.value.request_type)}` })
  if (filters.value.billing_type != null) chips.push({ key: 'billing-type', label: `${t('admin.usage.billingType')}: ${findOptionLabel(billingTypeOptions.value, filters.value.billing_type)}` })
  if (filters.value.billing_mode) chips.push({ key: 'billing-mode', label: `${t('admin.usage.billingMode')}: ${findOptionLabel(billingModeOptions.value, filters.value.billing_mode)}` })
  return chips
})

const applyMobileUsageFilters = (draft: {
  api_key_id: number | null
  model: string | null
  group_id: number | null
  request_type: string | null
  billing_type: number | null
  billing_mode: string | null
}) => {
  filters.value.api_key_id = draft.api_key_id ?? undefined
  filters.value.model = draft.model || undefined
  filters.value.group_id = draft.group_id ?? undefined
  filters.value.request_type = (draft.request_type || undefined) as UsageQueryParams['request_type']
  filters.value.billing_type = draft.billing_type
  filters.value.billing_mode = draft.billing_mode
  mobileFiltersOpen.value = false
  applyFilters()
}

const applyMobileErrorFilters = (draft: { api_key_id: number | null; model: string | null; category: string; status_code: number | null }) => {
  errorFilter.value = { ...draft }
  mobileFiltersOpen.value = false
  applyErrorFilters()
}

const removeActiveFilter = (key: string) => {
  if (key.startsWith('error-')) {
    if (key === 'error-api-key') errorFilter.value.api_key_id = null
    else if (key === 'error-model') errorFilter.value.model = ''
    else if (key === 'error-category') errorFilter.value.category = ''
    else if (key === 'error-status') errorFilter.value.status_code = null
    applyErrorFilters()
    return
  }
  if (key === 'api-key') filters.value.api_key_id = undefined
  else if (key === 'model') filters.value.model = undefined
  else if (key === 'group') filters.value.group_id = undefined
  else if (key === 'request-type') filters.value.request_type = undefined
  else if (key === 'billing-type') filters.value.billing_type = null
  else if (key === 'billing-mode') filters.value.billing_mode = null
  applyFilters()
}

const normalizedFilters = computed<UsageQueryParams>(() => {
  const requestType = filters.value.request_type
  const legacyStream = requestType ? requestTypeToLegacyStream(requestType) : filters.value.stream
  return {
    ...filters.value,
    start_date: startDate.value,
    end_date: endDate.value,
    stream: legacyStream === null ? undefined : legacyStream,
  }
})

const buildUsageListParams = (page: number, pageSize: number): UsageQueryParams => ({
  page,
  page_size: pageSize,
  ...normalizedFilters.value,
  sort_by: sortState.sort_by,
  sort_order: sortState.sort_order,
})

const loadLogs = async () => {
  abortController?.abort()
  const controller = new AbortController()
  abortController = controller
  loading.value = true
  try {
    const res = await usageAPI.query(buildUsageListParams(pagination.page, pagination.page_size), {
      signal: controller.signal,
    })
    if (!controller.signal.aborted) {
      usageLogs.value = res.items
      pagination.total = res.total
    }
  } catch (error: any) {
    if (error?.name !== 'AbortError' && error?.code !== 'ERR_CANCELED') {
      appStore.showError(t('usage.failedToLoad'))
    }
  } finally {
    if (abortController === controller) loading.value = false
  }
}

const loadStats = async () => {
  const seq = ++statsReqSeq
  endpointStatsLoading.value = true
  try {
    const stats = await usageAPI.getStats(normalizedFilters.value)
    if (seq !== statsReqSeq) return
    usageStats.value = stats
    inboundEndpointStats.value = stats.endpoints || []
    upstreamEndpointStats.value = []
    endpointPathStats.value = []
  } catch (error) {
    if (seq !== statsReqSeq) return
    console.error('Failed to load usage stats:', error)
    inboundEndpointStats.value = []
    upstreamEndpointStats.value = []
    endpointPathStats.value = []
  } finally {
    if (seq === statsReqSeq) endpointStatsLoading.value = false
  }
}

const loadModelStats = async () => {
  const seq = ++modelStatsReqSeq
  modelStatsLoading.value = true
  try {
    const response = await usageAPI.getDashboardModels({
      ...normalizedFilters.value,
      model_source: 'requested',
    })
    if (seq !== modelStatsReqSeq) return
    requestedModelStats.value = response.models || []
    refreshModelOptions(response.models || [])
  } catch (error) {
    if (seq !== modelStatsReqSeq) return
    console.error('Failed to load model stats:', error)
    requestedModelStats.value = []
  } finally {
    if (seq === modelStatsReqSeq) modelStatsLoading.value = false
  }
}

const loadChartData = async () => {
  const seq = ++chartReqSeq
  chartsLoading.value = true
  try {
    const snapshot = await usageAPI.getDashboardSnapshotV2({
      ...normalizedFilters.value,
      granularity: granularity.value,
      include_trend: true,
      include_model_stats: false,
      include_group_stats: true,
    })
    if (seq !== chartReqSeq) return
    trendData.value = snapshot.trend || []
    groupStats.value = snapshot.groups || []
  } catch (error) {
    if (seq !== chartReqSeq) return
    console.error('Failed to load chart data:', error)
    trendData.value = []
    groupStats.value = []
  } finally {
    if (seq === chartReqSeq) chartsLoading.value = false
  }
}

const refreshModelOptions = (models: ModelStat[]) => {
  const current = filters.value.model
  const set = new Set(modelOptionValues.value)
  models.forEach((item) => {
    if (item.model) set.add(item.model)
  })
  if (current) set.add(current)
  modelOptionValues.value = Array.from(set).sort()
}

const applyFilters = () => {
  pagination.page = 1
  void loadLogs()
  void loadStats()
  void loadModelStats()
  void loadChartData()
  resetErrorRows()
}

const refreshData = () => {
  void loadLogs()
  void loadStats()
  void loadModelStats()
  void loadChartData()
  if (activeTab.value === 'errors') void loadErrors()
}

const resetFilters = () => {
  const range = getLast24HoursRangeDates()
  startDate.value = range.start
  endDate.value = range.end
  filters.value = {
    start_date: range.start,
    end_date: range.end,
    request_type: undefined,
    billing_type: null,
    billing_mode: null,
  }
  granularity.value = getGranularityForRange(range.start, range.end)
  applyFilters()
  if (activeTab.value === 'errors') {
    errorFilter.value = { model: '', category: '', api_key_id: null, status_code: null }
    applyErrorFilters()
  }
}

const onDateRangeChange = (range: { startDate: string; endDate: string; preset: string | null }) => {
  startDate.value = range.startDate
  endDate.value = range.endDate
  filters.value.start_date = range.startDate
  filters.value.end_date = range.endDate
  granularity.value = getGranularityForRange(range.startDate, range.endDate)
  applyFilters()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  void loadLogs()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadLogs()
}

const handleSort = (key: string, order: 'asc' | 'desc') => {
  sortState.sort_by = key
  sortState.sort_order = order
  pagination.page = 1
  void loadLogs()
}

const handleIpGeoBatchFailed = () => {
  appStore.showError(t('usage.ipGeo.batchFailed'))
}

const getRequestTypeExportText = (log: UsageLog): string => {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'cyber') return 'Cyber'
  if (requestType === 'live') return 'Live'
  if (requestType === 'ws_v2') return 'WS'
  if (requestType === 'stream') return 'Stream'
  if (requestType === 'sync') return 'Sync'
  return 'Unknown'
}

const getDisplayBillingMode = (
  row: Pick<UsageLog, 'billing_mode' | 'image_count'> | null | undefined
): string | null | undefined => resolveDisplayBillingMode(row)

const escapeCSVValue = (value: unknown): string => {
  if (value == null) return ''
  const str = String(value)
  const escaped = str.replace(/"/g, '""')
  if (/^[=+\-@\t\r]/.test(str)) return `"\'${escaped}"`
  if (/[,"\n\r]/.test(str)) return `"${escaped}"`
  return str
}

const exportToCSV = async () => {
  if (pagination.total === 0) {
    appStore.showWarning(t('usage.noDataToExport'))
    return
  }
  exporting.value = true
  appStore.showInfo(t('usage.preparingExport'))
  try {
    const allLogs: UsageLog[] = []
    const pageSize = 100
    const totalPages = Math.ceil(pagination.total / pageSize)
    for (let page = 1; page <= totalPages; page++) {
      const response = await usageAPI.query(buildUsageListParams(page, pageSize))
      allLogs.push(...response.items)
    }
    if (allLogs.length === 0) {
      appStore.showWarning(t('usage.noDataToExport'))
      return
    }
    const headers = [
      'Time',
      'API Key Name',
      'Model',
      'Reasoning Effort',
      'Inbound Endpoint',
      'IP Address',
      'Type',
      'Billing Mode',
      'Input Tokens',
      'Output Tokens',
      'Cache Read Tokens',
      'Cache Creation Tokens',
      'Rate Multiplier',
      'Billed Cost',
      'Original Cost',
      'First Token (ms)',
      'Duration (ms)',
    ]
    const rows = allLogs.map((log) => [
      log.created_at,
      log.api_key?.name || '',
      log.model,
      formatReasoningEffort(log.reasoning_effort),
      log.inbound_endpoint || '',
      log.ip_address || '',
      getRequestTypeExportText(log),
      getBillingModeLabel(getDisplayBillingMode(log), t),
      log.input_tokens,
      log.output_tokens,
      log.cache_read_tokens,
      log.cache_creation_tokens,
      log.rate_multiplier,
      log.actual_cost.toFixed(8),
      log.total_cost.toFixed(8),
      log.first_token_ms ?? '',
      log.duration_ms ?? '',
    ].map(escapeCSVValue))
    const csvContent = [
      headers.map(escapeCSVValue).join(','),
      ...rows.map((row) => row.join(',')),
    ].join('\n')
    const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `usage_${startDate.value}_to_${endDate.value}.csv`
    link.click()
    window.URL.revokeObjectURL(url)
    appStore.showSuccess(t('usage.exportSuccess'))
  } catch (error) {
    console.error('CSV Export failed:', error)
    appStore.showError(t('usage.exportFailed'))
  } finally {
    exporting.value = false
  }
}

const ALWAYS_VISIBLE = ['created_at']
const DEFAULT_HIDDEN_COLUMNS = ['user_agent']
const HIDDEN_COLUMNS_KEY = 'user-usage-hidden-columns'

const allColumns = computed<Column[]>(() => [
  { key: 'api_key', label: t('usage.apiKeyFilter'), sortable: false },
  { key: 'model', label: t('usage.model'), sortable: true },
  { key: 'reasoning_effort', label: t('usage.reasoningEffort'), sortable: false },
  { key: 'endpoint', label: t('usage.endpoint'), sortable: false },
  { key: 'ip_address', label: 'IP', sortable: false },
  { key: 'group', label: t('admin.usage.group'), sortable: false },
  { key: 'stream', label: t('usage.type'), sortable: false },
  { key: 'billing_mode', label: t('admin.usage.billingMode'), sortable: false },
  { key: 'tokens', label: t('usage.tokens'), sortable: false },
  { key: 'cost', label: t('usage.cost'), sortable: false },
  { key: 'latency', label: t('usage.latency'), sortable: false },
  { key: 'created_at', label: t('usage.time'), sortable: true },
  { key: 'user_agent', label: t('usage.userAgent'), sortable: false },
])

const hiddenColumns = reactive<Set<string>>(new Set())
const toggleableColumns = computed(() => allColumns.value.filter((col) => !ALWAYS_VISIBLE.includes(col.key)))
const visibleColumns = computed(() =>
  allColumns.value.filter((col) => ALWAYS_VISIBLE.includes(col.key) || !hiddenColumns.has(col.key))
)
const isColumnVisible = (key: string) => !hiddenColumns.has(key)
const toggleColumn = (key: string) => {
  if (hiddenColumns.has(key)) hiddenColumns.delete(key)
  else hiddenColumns.add(key)
  localStorage.setItem(HIDDEN_COLUMNS_KEY, JSON.stringify([...hiddenColumns]))
}
const loadSavedColumns = () => {
  try {
    const saved = localStorage.getItem(HIDDEN_COLUMNS_KEY)
    const values = saved ? JSON.parse(saved) as string[] : DEFAULT_HIDDEN_COLUMNS
    values.forEach((key) => hiddenColumns.add(key))
  } catch {
    DEFAULT_HIDDEN_COLUMNS.forEach((key) => hiddenColumns.add(key))
  }
}

// 错误请求 tab 独立列设置(机制同用量列设置,存储互不影响)
const ERR_ALWAYS_VISIBLE = ['status', 'created_at']
const ERR_DEFAULT_HIDDEN_COLUMNS = ['user_agent']
const ERR_HIDDEN_COLUMNS_KEY = 'user-usage-error-hidden-columns'

// key 须与 UserErrorRequestsTable 的 allColumns 一致
const errAllColumns = computed<Column[]>(() => [
  { key: 'key_name', label: t('usage.errors.keyName') },
  { key: 'model', label: t('usage.errors.model') },
  { key: 'endpoint', label: t('usage.errors.endpoint') },
  { key: 'client_ip', label: 'IP' },
  { key: 'group', label: t('admin.usage.group') },
  { key: 'type', label: t('usage.type') },
  { key: 'platform', label: t('usage.errors.platform') },
  { key: 'category', label: t('usage.errors.category') },
  { key: 'status', label: t('usage.errors.status') },
  { key: 'message', label: t('usage.errors.message') },
  { key: 'created_at', label: t('usage.errors.time') },
  { key: 'user_agent', label: t('usage.userAgent') },
])

const errHiddenColumns = reactive<Set<string>>(new Set())
const errToggleableColumns = computed(() =>
  errAllColumns.value.filter((col) => !ERR_ALWAYS_VISIBLE.includes(col.key))
)
const errVisibleColumnKeys = computed(() =>
  errAllColumns.value
    .filter((col) => ERR_ALWAYS_VISIBLE.includes(col.key) || !errHiddenColumns.has(col.key))
    .map((col) => col.key)
)
const isErrColumnVisible = (key: string) => !errHiddenColumns.has(key)
const toggleErrColumn = (key: string) => {
  if (errHiddenColumns.has(key)) errHiddenColumns.delete(key)
  else errHiddenColumns.add(key)
  localStorage.setItem(ERR_HIDDEN_COLUMNS_KEY, JSON.stringify([...errHiddenColumns]))
}
const loadSavedErrColumns = () => {
  try {
    const saved = localStorage.getItem(ERR_HIDDEN_COLUMNS_KEY)
    const values = saved ? (JSON.parse(saved) as string[]) : ERR_DEFAULT_HIDDEN_COLUMNS
    values.forEach((key) => errHiddenColumns.add(key))
  } catch {
    ERR_DEFAULT_HIDDEN_COLUMNS.forEach((key) => errHiddenColumns.add(key))
  }
}

// 列设置下拉按当前 tab 分发
const currentToggleableColumns = computed(() =>
  activeTab.value === 'errors' ? errToggleableColumns.value : toggleableColumns.value
)
const isCurrentColumnVisible = (key: string) =>
  activeTab.value === 'errors' ? isErrColumnVisible(key) : isColumnVisible(key)
const toggleCurrentColumn = (key: string) => {
  if (activeTab.value === 'errors') toggleErrColumn(key)
  else toggleColumn(key)
}

const columnDropdownRef = ref<HTMLElement | null>(null)
const {
  open: showColumnDropdown,
  triggerRef: columnMenuTriggerRef,
  menuRef: columnMenuRef,
  triggerId: columnMenuTriggerId,
  menuId: columnMenuId,
  closeMenu: closeColumnMenu,
  toggleMenu: toggleColumnMenu,
  handleTriggerKeydown: handleColumnTriggerKeydown,
  handleMenuKeydown: handleColumnMenuKeydown,
} = useDropdownMenu('user-usage-columns')
const handleColumnClickOutside = (event: MouseEvent) => {
  if (columnDropdownRef.value && !columnDropdownRef.value.contains(event.target as HTMLElement)) {
    void closeColumnMenu()
  }
}

const loadFilterOptions = async () => {
  try {
    const [keys, availableGroups] = await Promise.all([
      keysAPI.list(1, 100),
      userGroupsAPI.getAvailable(),
    ])
    apiKeys.value = keys.items
    groups.value = availableGroups
  } catch (error) {
    console.error('Failed to load usage filter options:', error)
  }
}

const resetErrorRows = () => {
  errorPage.value = 1
  if (activeTab.value === 'errors') {
    void loadErrors()
  } else {
    errorRows.value = []
    errorTotal.value = 0
  }
}

const loadErrors = async () => {
  errorLoading.value = true
  try {
    const resp = await usageAPI.listMyErrorRequests({
      page: errorPage.value,
      page_size: errorPageSize.value,
      start_date: startDate.value,
      end_date: endDate.value,
      model: (errorFilter.value.model ?? '').trim() || undefined,
      category: errorFilter.value.category || undefined,
      api_key_id: errorFilter.value.api_key_id ?? undefined,
      status_code: errorFilter.value.status_code ?? undefined,
      sort_by: errorSortBy.value,
      sort_order: errorSortOrder.value,
    })
    errorRows.value = resp.items
    errorTotal.value = resp.total
  } catch (error) {
    console.error('[UsageView] loadErrors failed:', error)
    appStore.showError(t('usage.errors.failedToLoad'))
  } finally {
    errorLoading.value = false
  }
}

const onErrorSort = (sortBy: string, sortOrder: 'asc' | 'desc') => {
  errorSortBy.value = sortBy
  errorSortOrder.value = sortOrder
  errorPage.value = 1
  void loadErrors()
}

const onErrorPage = (page: number) => {
  errorPage.value = page
  void loadErrors()
}

const onErrorPageSize = (pageSize: number) => {
  errorPageSize.value = pageSize
  errorPage.value = 1
  void loadErrors()
}

const switchToErrors = () => {
  activeTab.value = 'errors'
  if (errorRows.value.length === 0) void loadErrors()
}

const activateUsageTab = (tab: 'usage' | 'errors') => {
  if (tab === 'errors') switchToErrors()
  else activeTab.value = 'usage'
}

const handleUsageTabKeydown = (event: KeyboardEvent, current: 'usage' | 'errors') => {
  let target: 'usage' | 'errors' | null = null
  if (event.key === 'ArrowLeft' || event.key === 'ArrowRight') {
    target = current === 'usage' ? 'errors' : 'usage'
  } else if (event.key === 'Home') {
    target = 'usage'
  } else if (event.key === 'End') {
    target = 'errors'
  }
  if (!target) return

  event.preventDefault()
  activateUsageTab(target)
  const mobileTab = window.matchMedia?.('(max-width: 639px)').matches
  window.requestAnimationFrame(() => document.getElementById(`${target}-tab${mobileTab ? '-mobile' : ''}`)?.focus())
}

onMounted(() => {
  loadSavedColumns()
  loadSavedErrColumns()
  document.addEventListener('click', handleColumnClickOutside)
  void loadFilterOptions()
  refreshData()
})

onUnmounted(() => {
  abortController?.abort()
  document.removeEventListener('click', handleColumnClickOutside)
})

watch(endpointDistributionSource, () => {
  // Endpoint source switching is handled by the chart component using already loaded stats.
})
</script>

<style scoped>
.usage-page,
.usage-analytics,
.usage-records {
  display: grid;
  min-width: 0;
  gap: 16px;
}

.usage-page {
  gap: 24px;
}

.usage-mobile-records-toolbar {
  min-width: 0;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 8px;
  padding: 0.625rem;
  background: var(--ui-surface, #fff);
  box-shadow: var(--ui-shadow-xs, 0 1px 2px rgba(15, 23, 42, 0.04));
}

.usage-mobile-tab {
  display: inline-flex;
  min-width: 0;
  min-height: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  padding: 0.375rem 0.625rem;
  color: var(--ui-text-muted);
  font-size: 0.75rem;
  font-weight: 600;
}

.usage-mobile-tab-active {
  background: var(--ui-surface-raised);
  color: var(--ui-text);
  box-shadow: var(--ui-shadow-xs);
}

.usage-mobile-tab:focus-visible {
  outline: 2px solid var(--ui-focus);
  outline-offset: 1px;
}

.usage-chart-toolbar,
.usage-filters,
.usage-records-header,
.usage-table-panel {
  min-width: 0;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 8px;
  background: var(--ui-surface, #fff);
  box-shadow: var(--ui-shadow-xs, 0 1px 2px rgba(15, 23, 42, 0.04));
}

.usage-chart-toolbar {
  display: flex;
  min-height: 64px;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 10px 12px 10px 16px;
}

.usage-toolbar-copy p,
.usage-toolbar-copy h2 {
  margin: 0;
}

.usage-toolbar-copy p {
  color: var(--ui-text-subtle, #8793a3);
  font-size: 10px;
  font-weight: 700;
  line-height: 1.2;
  text-transform: uppercase;
}

.usage-toolbar-copy h2 {
  margin-top: 3px;
  color: var(--ui-text, #0f172a);
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0;
}

.usage-chart-controls,
.usage-control-group,
.usage-filter-actions {
  display: flex;
  align-items: center;
}

.usage-chart-controls {
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 10px;
}

.usage-control-group {
  gap: 8px;
}

.usage-control-label {
  color: var(--ui-text-muted, #667085);
  font-size: 11px;
  font-weight: 650;
  white-space: nowrap;
}

.usage-filters {
  padding: 14px;
}

.usage-filter-row {
  display: flex;
  min-width: 0;
  align-items: flex-end;
  gap: 16px;
}

.usage-filter-fields {
  display: grid;
  min-width: 0;
  flex: 1;
  grid-template-columns: repeat(auto-fit, minmax(168px, 1fr));
  gap: 12px;
}

.usage-filter-fields > * {
  width: 100% !important;
  min-width: 0 !important;
}

.usage-filter-actions {
  flex: 0 0 auto;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.usage-records-header {
  overflow: hidden;
}

.usage-tabs {
  display: flex;
  min-height: 44px;
  align-items: stretch;
  padding: 0 8px;
}

.usage-tab {
  position: relative;
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  padding: 0 12px;
  border: 0;
  background: transparent;
  color: var(--ui-text-muted, #667085);
  cursor: pointer;
  font-size: 13px;
  font-weight: 650;
}

.usage-tab::after {
  position: absolute;
  right: 12px;
  bottom: 0;
  left: 12px;
  height: 2px;
  border-radius: 2px 2px 0 0;
  background: transparent;
  content: '';
}

.usage-tab:hover,
.usage-tab-active {
  color: var(--ui-text, #0f172a);
}

.usage-tab-active::after {
  background: var(--ui-text, #0f172a);
}

.usage-table-panel {
  overflow: hidden;
}

.usage-table-panel :deep(.card) {
  border: 0;
  border-radius: 0;
  box-shadow: none;
}

.usage-table-panel :deep(.pagination-container) {
  border-top: 1px solid var(--ui-border, #dbe3ee);
}

@media (max-width: 1279px) {
  .usage-filter-row {
    align-items: stretch;
    flex-direction: column;
  }

  .usage-filter-actions {
    width: 100%;
  }
}

@media (max-width: 1023px) {
  .usage-chart-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .usage-chart-controls {
    width: 100%;
    justify-content: flex-start;
  }
}

@media (max-width: 639px) {
  .usage-page {
    gap: 20px;
  }

  .usage-chart-toolbar,
  .usage-filters {
    padding: 14px;
  }

  .usage-chart-controls {
    display: grid;
    width: 100%;
    grid-template-columns: minmax(0, 1fr) 112px;
    align-items: end;
    gap: 8px;
  }

  .usage-control-group {
    display: block;
    min-width: 0;
  }

  .usage-granularity-control > div {
    width: 100%;
  }

  .usage-filter-fields {
    grid-template-columns: minmax(0, 1fr);
  }

  .usage-filter-actions {
    justify-content: flex-start;
  }

  .usage-filter-actions > .btn {
    flex: 1 1 auto;
  }

  .usage-table-panel {
    overflow: visible;
    border: 0;
    background: transparent;
    box-shadow: none;
  }
}
</style>
