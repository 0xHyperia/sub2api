<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Loading State -->
      <div
        v-if="loading && !stats"
        class="flex items-center justify-center py-12"
        role="status"
        :aria-label="t('common.loading')"
      >
        <LoadingSpinner />
      </div>

      <template v-else>
        <div
          v-if="dashboardError"
          class="dashboard-load-error"
          role="alert"
          data-testid="dashboard-load-error"
        >
          <Icon name="exclamationTriangle" size="md" aria-hidden="true" />
          <div>
            <p>{{ t('admin.dashboard.failedToLoad') }}</p>
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :disabled="loading || chartsLoading"
              @click="loadDashboardStats"
            >
              <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading || chartsLoading }" />
              {{ t('common.retry') }}
            </button>
          </div>
        </div>

        <template v-if="stats">
        <section class="dashboard-kpi-strip" :aria-label="t('admin.dashboard.title')">
          <DashboardMetric
            :label="t('admin.dashboard.todayRequests')"
            :value="formatNumber(stats.today_requests)"
            :detail="`${t('common.total')}: ${formatNumber(stats.total_requests)}`"
            icon="chart"
            tone="success"
          />
          <DashboardMetric
            :label="t('admin.dashboard.todayTokens')"
            :value="formatTokens(stats.today_tokens)"
            :detail="`${t('admin.dashboard.actual')}: $${formatCost(stats.today_actual_cost)}`"
            icon="cube"
          />
          <DashboardMetric
            :label="t('admin.dashboard.performance')"
            :value="`${formatTokens(stats.rpm)} RPM`"
            :detail="`${formatTokens(stats.tpm)} TPM`"
            icon="bolt"
          />
          <DashboardMetric
            :label="t('admin.dashboard.avgResponse')"
            :value="formatDuration(stats.average_duration_ms)"
            :detail="`${stats.active_users} ${t('admin.dashboard.activeUsers')}`"
            icon="clock"
            :tone="stats.average_duration_ms > 5000 ? 'warning' : 'neutral'"
          />
        </section>

        <section class="dashboard-health-panel">
          <div class="dashboard-section-heading">
            <div>
              <p class="dashboard-eyebrow">{{ t('admin.dashboard.performance') }}</p>
              <h2>{{ t('admin.dashboard.accounts') }}</h2>
            </div>
            <span
              class="dashboard-health-state"
              :data-state="stats.error_accounts > 0 ? 'warning' : 'healthy'"
            >
              <span aria-hidden="true"></span>
              {{ stats.error_accounts > 0 ? t('common.error') : t('common.active') }}
            </span>
          </div>

          <div class="dashboard-health-grid">
            <div>
              <span>{{ t('admin.dashboard.accounts') }}</span>
              <strong>{{ formatNumber(stats.normal_accounts) }} / {{ formatNumber(stats.total_accounts) }}</strong>
            </div>
            <div>
              <span>{{ t('admin.dashboard.apiKeys') }}</span>
              <strong>{{ formatNumber(stats.active_api_keys) }} / {{ formatNumber(stats.total_api_keys) }}</strong>
            </div>
            <div>
              <span>{{ t('admin.dashboard.users') }}</span>
              <strong>{{ formatNumber(stats.active_users) }} / {{ formatNumber(stats.total_users) }}</strong>
            </div>
            <div>
              <span>{{ t('admin.dashboard.totalTokens') }}</span>
              <strong>{{ formatTokens(stats.total_tokens) }}</strong>
            </div>
          </div>

          <div v-if="stats.error_accounts > 0" class="dashboard-alert" role="status">
            <Icon name="exclamationTriangle" size="sm" aria-hidden="true" />
            <span>{{ stats.error_accounts }} {{ t('common.error') }}</span>
            <button type="button" @click="router.push('/admin/accounts')">
              {{ t('admin.dashboard.manageAccounts') }}
            </button>
          </div>
        </section>

        <section class="dashboard-actions-panel" :aria-label="t('admin.dashboard.quickActions')">
          <div class="dashboard-section-heading dashboard-section-heading-compact">
            <div>
              <p class="dashboard-eyebrow">{{ t('admin.dashboard.quickActions') }}</p>
              <h2>{{ t('admin.dashboard.groupPricing') }}</h2>
            </div>
          </div>
          <div class="dashboard-action-list">
            <button type="button" @click="router.push('/admin/groups')">
              <span class="dashboard-action-icon"><Icon name="grid" size="sm" /></span>
              <span>
                <strong>{{ t('admin.dashboard.groupPricing') }}</strong>
                <small>{{ t('admin.dashboard.groupPricingDesc') }}</small>
              </span>
              <Icon name="chevronRight" size="sm" aria-hidden="true" />
            </button>
            <button
              v-if="canUseBatchImage"
              type="button"
              @click="router.push('/batch-image')"
            >
              <span class="dashboard-action-icon"><Icon name="sparkles" size="sm" /></span>
              <span>
                <strong>{{ t('admin.dashboard.batchImage') }}</strong>
                <small>{{ t('admin.dashboard.batchImageDesc') }}</small>
              </span>
              <Icon name="chevronRight" size="sm" aria-hidden="true" />
            </button>
          </div>
        </section>

        <section class="dashboard-analytics">
          <div class="dashboard-chart-toolbar">
            <div class="dashboard-toolbar-copy">
              <p class="dashboard-eyebrow">{{ t('admin.dashboard.recentUsage') }}</p>
              <h2>{{ t('admin.dashboard.tokenUsageTrend') }}</h2>
            </div>
            <div class="dashboard-chart-controls">
              <div class="dashboard-control-group">
                <span class="dashboard-control-label">{{ t('admin.dashboard.timeRange') }}</span>
                <DateRangePicker
                  v-model:start-date="startDate"
                  v-model:end-date="endDate"
                  @change="onDateRangeChange"
                />
              </div>
              <div class="dashboard-control-group dashboard-granularity-control">
                <span class="dashboard-control-label">{{ t('admin.dashboard.granularity') }}</span>
                <div class="w-28">
                  <Select
                    v-model="granularity"
                    :options="granularityOptions"
                    @change="loadChartData"
                  />
                </div>
              </div>
              <button
                type="button"
                @click="loadDashboardStats"
                :disabled="chartsLoading"
                class="btn btn-secondary"
              >
                <Icon name="refresh" size="sm" :class="chartsLoading ? 'animate-spin' : ''" />
                {{ t('common.refresh') }}
              </button>
            </div>
          </div>

          <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <ModelDistributionChart
              :model-stats="modelStats"
              :enable-ranking-view="true"
              :ranking-items="rankingItems"
              :ranking-total-actual-cost="rankingTotalActualCost"
              :ranking-total-requests="rankingTotalRequests"
              :ranking-total-tokens="rankingTotalTokens"
              :loading="chartsLoading"
              :ranking-loading="rankingLoading"
              :ranking-error="rankingError"
              :start-date="startDate"
              :end-date="endDate"
              @ranking-click="goToUserUsage"
            />
            <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
          </div>

          <div class="dashboard-chart-card">
            <h3>
              {{ t('admin.dashboard.recentUsage') }} (Top 12)
            </h3>
            <div class="h-64">
              <div v-if="userTrendLoading" class="flex h-full items-center justify-center">
                <LoadingSpinner size="md" />
              </div>
              <div
                v-else-if="userTrendError"
                class="flex h-full flex-col items-center justify-center gap-3 px-4 text-center text-sm text-danger-foreground"
                role="alert"
              >
                <span>{{ t('admin.dashboard.failedToLoad') }}</span>
                <button type="button" class="btn btn-secondary btn-sm" @click="loadUsersTrend">
                  <Icon name="refresh" size="sm" />
                  {{ t('common.retry') }}
                </button>
              </div>
              <Line v-else-if="userTrendChartData" :data="userTrendChartData" :options="lineOptions" />
              <div
                v-else
                class="flex h-full items-center justify-center text-sm text-foreground-subtle"
              >
                {{ t('admin.dashboard.noDataAvailable') }}
              </div>
            </div>
          </div>
        </section>
        </template>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
import { adminAPI } from '@/api/admin'
import type {
  DashboardStats,
  TrendDataPoint,
  ModelStat,
  UserUsageTrendPoint,
  UserSpendingRankingItem
} from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import DashboardMetric from '@/components/admin/DashboardMetric.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import { useBatchImageAccess } from '@/composables/useBatchImageAccess'
import { useChartTheme } from '@/composables/useChartTheme'

import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
)

const appStore = useAppStore()
const router = useRouter()
const { canUseBatchImage, refreshBatchImageAccess } = useBatchImageAccess()
const { chartTheme } = useChartTheme()
const stats = ref<DashboardStats | null>(null)
const loading = ref(false)
const chartsLoading = ref(false)
const dashboardError = ref(false)
const userTrendLoading = ref(false)
const userTrendError = ref(false)
const rankingLoading = ref(false)
const rankingError = ref(false)

// Chart data
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const userTrend = ref<UserUsageTrendPoint[]>([])
const rankingItems = ref<UserSpendingRankingItem[]>([])
const rankingTotalActualCost = ref(0)
const rankingTotalRequests = ref(0)
const rankingTotalTokens = ref(0)
let chartLoadSeq = 0
let usersTrendLoadSeq = 0
let rankingLoadSeq = 0
const rankingLimit = 12

// Helper function to format date in local timezone
const formatLocalDate = (date: Date): string => {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const getLast24HoursRangeDates = (): { start: string; end: string } => {
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  return {
    start: formatLocalDate(start),
    end: formatLocalDate(end)
  }
}

// Date range
const granularity = ref<'day' | 'hour'>('hour')
const defaultRange = getLast24HoursRangeDates()
const startDate = ref(defaultRange.start)
const endDate = ref(defaultRange.end)

// Granularity options for Select component
const granularityOptions = computed(() => [
  { value: 'day', label: t('admin.dashboard.day') },
  { value: 'hour', label: t('admin.dashboard.hour') }
])

// Chart colors
const chartColors = computed(() => ({
  text: chartTheme.value.foregroundMuted,
  grid: chartTheme.value.outline,
}))

// Line chart options (for user trend chart)
const lineOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        padding: 15,
        font: {
          size: 11
        }
      }
    },
    tooltip: {
      itemSort: (a: any, b: any) => {
        const aValue = typeof a?.raw === 'number' ? a.raw : Number(a?.parsed?.y ?? 0)
        const bValue = typeof b?.raw === 'number' ? b.raw : Number(b?.parsed?.y ?? 0)
        return bValue - aValue
      },
      callbacks: {
        label: (context: any) => {
          return `${context.dataset.label}: ${formatTokens(context.raw)}`
        }
      }
    }
  },
  scales: {
    x: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        }
      }
    },
    y: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        },
        callback: (value: string | number) => formatTokens(Number(value))
      }
    }
  }
}))

// User trend chart data
const userTrendChartData = computed(() => {
  if (!userTrend.value?.length) return null

  const getDisplayName = (point: UserUsageTrendPoint): string => {
    const username = point.username?.trim()
    if (username) {
      return username
    }

    const email = point.email?.trim()
    if (email) {
      return email
    }

    return t('admin.redeem.userPrefix', { id: point.user_id })
  }

  // Group by user_id to avoid merging different users with the same display name
  const userGroups = new Map<number, { name: string; data: Map<string, number> }>()
  const allDates = new Set<string>()

  userTrend.value.forEach((point) => {
    allDates.add(point.date)
    const key = point.user_id
    if (!userGroups.has(key)) {
      userGroups.set(key, { name: getDisplayName(point), data: new Map() })
    }
    userGroups.get(key)!.data.set(point.date, point.tokens)
  })

  const sortedDates = Array.from(allDates).sort()
  const colors = [
    '#3b82f6',
    '#10b981',
    '#f59e0b',
    '#ef4444',
    '#8b5cf6',
    '#ec4899',
    '#14b8a6',
    '#f97316',
    '#6366f1',
    '#84cc16',
    '#06b6d4',
    '#a855f7'
  ]

  const datasets = Array.from(userGroups.values()).map((group, idx) => ({
    label: group.name,
    data: sortedDates.map((date) => group.data.get(date) || 0),
    borderColor: colors[idx % colors.length],
    backgroundColor: `${colors[idx % colors.length]}20`,
    fill: false,
    tension: 0.3
  }))

  return {
    labels: sortedDates,
    datasets
  }
})

// Format helpers
const formatTokens = (value: number | undefined): string => {
  if (value === undefined || value === null) return '0'
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return value.toLocaleString()
}

const toFiniteNumber = (value: unknown): number => {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : 0
}

const formatNumber = (value: number | null | undefined): string => {
  return toFiniteNumber(value).toLocaleString()
}

const formatCost = (value: number | null | undefined): string => {
  const safeValue = toFiniteNumber(value)
  if (safeValue >= 1000) {
    return (safeValue / 1000).toFixed(2) + 'K'
  } else if (safeValue >= 1) {
    return safeValue.toFixed(2)
  } else if (safeValue >= 0.01) {
    return safeValue.toFixed(3)
  }
  return safeValue.toFixed(4)
}

const formatDuration = (ms: number): string => {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)}s`
  }
  return `${Math.round(ms)}ms`
}

const goToUserUsage = (item: UserSpendingRankingItem) => {
  void router.push({
    path: '/admin/usage',
    query: {
      user_id: String(item.user_id),
      start_date: startDate.value,
      end_date: endDate.value
    }
  })
}

// Date range change handler
const onDateRangeChange = (range: {
  startDate: string
  endDate: string
  preset: string | null
}) => {
  // Auto-select granularity based on date range
  const start = new Date(range.startDate)
  const end = new Date(range.endDate)
  const daysDiff = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24))

  // If range is 1 day, use hourly granularity
  if (daysDiff <= 1) {
    granularity.value = 'hour'
  } else {
    granularity.value = 'day'
  }

  loadChartData()
}

// Load data
const loadDashboardSnapshot = async (includeStats: boolean) => {
  const currentSeq = ++chartLoadSeq
  if (includeStats && !stats.value) {
    loading.value = true
  }
  chartsLoading.value = true
  dashboardError.value = false
  try {
    const response = await adminAPI.dashboard.getSnapshotV2({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      include_stats: includeStats,
      include_trend: true,
      include_model_stats: true,
      include_group_stats: false,
      include_users_trend: false
    })
    if (currentSeq !== chartLoadSeq) return
    if (includeStats && response.stats) {
      stats.value = response.stats
    }
    if (includeStats && !stats.value) {
      throw new Error('Dashboard snapshot did not include statistics')
    }
    trendData.value = response.trend || []
    modelStats.value = response.models || []
  } catch (error) {
    if (currentSeq !== chartLoadSeq) return
    dashboardError.value = true
    appStore.showError(t('admin.dashboard.failedToLoad'))
    console.error('Error loading dashboard snapshot:', error)
  } finally {
    if (currentSeq === chartLoadSeq) {
      loading.value = false
      chartsLoading.value = false
    }
  }
}

const loadUsersTrend = async () => {
  const currentSeq = ++usersTrendLoadSeq
  userTrendLoading.value = true
  userTrendError.value = false
  try {
    const response = await adminAPI.dashboard.getUserUsageTrend({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      limit: 12
    })
    if (currentSeq !== usersTrendLoadSeq) return
    userTrend.value = response.trend || []
    userTrendError.value = false
  } catch (error) {
    if (currentSeq !== usersTrendLoadSeq) return
    console.error('Error loading users trend:', error)
    userTrend.value = []
    userTrendError.value = true
  } finally {
    if (currentSeq === usersTrendLoadSeq) {
      userTrendLoading.value = false
    }
  }
}

const loadUserSpendingRanking = async () => {
  const currentSeq = ++rankingLoadSeq
  rankingLoading.value = true
  rankingError.value = false
  try {
    const response = await adminAPI.dashboard.getUserSpendingRanking({
      start_date: startDate.value,
      end_date: endDate.value,
      limit: rankingLimit
    })
    if (currentSeq !== rankingLoadSeq) return
    rankingItems.value = response.ranking || []
    rankingTotalActualCost.value = response.total_actual_cost || 0
    rankingTotalRequests.value = response.total_requests || 0
    rankingTotalTokens.value = response.total_tokens || 0
  } catch (error) {
    if (currentSeq !== rankingLoadSeq) return
    console.error('Error loading user spending ranking:', error)
    rankingItems.value = []
    rankingTotalActualCost.value = 0
    rankingTotalRequests.value = 0
    rankingTotalTokens.value = 0
    rankingError.value = true
  } finally {
    if (currentSeq === rankingLoadSeq) {
      rankingLoading.value = false
    }
  }
}

const loadDashboardStats = async () => {
  await Promise.all([
    loadDashboardSnapshot(true),
    loadUsersTrend(),
    loadUserSpendingRanking()
  ])
}

const loadChartData = async () => {
  await Promise.all([
    loadDashboardSnapshot(false),
    loadUsersTrend(),
    loadUserSpendingRanking()
  ])
}

onMounted(() => {
  void refreshBatchImageAccess()
  loadDashboardStats()
})
</script>

<style scoped>
.dashboard-kpi-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  overflow: hidden;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 8px;
  background: var(--ui-surface, #fff);
  box-shadow: var(--ui-shadow-xs, 0 1px 2px rgba(15, 23, 42, 0.04));
}

.dashboard-load-error {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  border: 1px solid color-mix(in srgb, var(--ui-danger, #dc2626) 30%, transparent);
  border-radius: 8px;
  color: var(--ui-danger-text, #b42318);
  background: var(--ui-danger-subtle, #fef3f2);
}

.dashboard-load-error > div {
  display: flex;
  min-width: 0;
  flex: 1;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.dashboard-load-error p {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

@media (max-width: 479px) {
  .dashboard-load-error > div {
    align-items: flex-start;
    flex-direction: column;
  }
}

.dashboard-health-panel,
.dashboard-actions-panel,
.dashboard-chart-toolbar,
.dashboard-chart-card {
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 8px;
  background: var(--ui-surface, #fff);
  box-shadow: var(--ui-shadow-xs, 0 1px 2px rgba(15, 23, 42, 0.04));
}

.dashboard-health-panel,
.dashboard-actions-panel {
  overflow: hidden;
}

.dashboard-section-heading {
  display: flex;
  min-height: 62px;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--ui-border, #dbe3ee);
}

.dashboard-section-heading-compact {
  min-height: 56px;
}

.dashboard-section-heading h2,
.dashboard-chart-toolbar h2 {
  margin: 2px 0 0;
  color: var(--ui-text, #0f172a);
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.35;
}

.dashboard-eyebrow {
  margin: 0;
  color: var(--ui-text-subtle, #8793a3);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.08em;
  line-height: 1.2;
  text-transform: uppercase;
}

.dashboard-health-state {
  display: inline-flex;
  min-height: 28px;
  align-items: center;
  gap: 7px;
  padding: 4px 9px;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 999px;
  color: var(--ui-text-muted, #667085);
  font-size: 11px;
  font-weight: 650;
  white-space: nowrap;
}

.dashboard-health-state > span {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--ui-success, #059669);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--ui-success, #059669) 12%, transparent);
}

.dashboard-health-state[data-state='warning'] > span {
  background: var(--ui-warning, #d97706);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--ui-warning, #d97706) 12%, transparent);
}

.dashboard-health-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.dashboard-health-grid > div {
  min-width: 0;
  padding: 15px 16px 16px;
  border-right: 1px solid var(--ui-border, #dbe3ee);
}

.dashboard-health-grid > div:last-child {
  border-right: 0;
}

.dashboard-health-grid span,
.dashboard-health-grid strong {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dashboard-health-grid span {
  color: var(--ui-text-muted, #667085);
  font-size: 11px;
  font-weight: 600;
}

.dashboard-health-grid strong {
  margin-top: 7px;
  color: var(--ui-text, #0f172a);
  font-size: 16px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.dashboard-alert {
  display: flex;
  min-height: 42px;
  align-items: center;
  gap: 9px;
  padding: 8px 16px;
  border-top: 1px solid color-mix(in srgb, var(--ui-warning, #d97706) 24%, var(--ui-border, #dbe3ee));
  background: color-mix(in srgb, var(--ui-warning, #d97706) 6%, var(--ui-surface, #fff));
  color: var(--ui-warning-strong, #9a5805);
  font-size: 12px;
  font-weight: 600;
}

.dashboard-alert button {
  margin-left: auto;
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font: inherit;
  text-decoration: underline;
  text-underline-offset: 3px;
}

.dashboard-action-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.dashboard-action-list > button {
  display: grid;
  min-height: 70px;
  grid-template-columns: 32px minmax(0, 1fr) 18px;
  align-items: center;
  gap: 11px;
  padding: 12px 16px;
  border: 0;
  border-right: 1px solid var(--ui-border, #dbe3ee);
  background: transparent;
  color: var(--ui-text-muted, #667085);
  cursor: pointer;
  text-align: left;
  transition: background-color 150ms ease, color 150ms ease;
}

.dashboard-action-list > button:last-child {
  border-right: 0;
}

.dashboard-action-list > button:hover {
  background: var(--ui-surface-subtle, #f4f7fb);
  color: var(--ui-text, #0f172a);
}

.dashboard-action-list > button:focus-visible,
.dashboard-alert button:focus-visible {
  outline: 2px solid var(--ui-focus, #475569);
  outline-offset: -3px;
}

.dashboard-action-icon {
  display: inline-grid;
  width: 32px;
  height: 32px;
  place-items: center;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 7px;
  background: var(--ui-surface-subtle, #f4f7fb);
}

.dashboard-action-list strong,
.dashboard-action-list small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dashboard-action-list strong {
  color: var(--ui-text, #0f172a);
  font-size: 13px;
  font-weight: 700;
}

.dashboard-action-list small {
  margin-top: 3px;
  color: var(--ui-text-muted, #667085);
  font-size: 11px;
}

.dashboard-analytics {
  display: grid;
  gap: 16px;
}

.dashboard-chart-toolbar {
  display: flex;
  min-height: 64px;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 10px 12px 10px 16px;
}

.dashboard-chart-controls,
.dashboard-control-group {
  display: flex;
  align-items: center;
}

.dashboard-chart-controls {
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 10px;
}

.dashboard-control-group {
  gap: 8px;
}

.dashboard-control-label {
  color: var(--ui-text-muted, #667085);
  font-size: 11px;
  font-weight: 650;
  white-space: nowrap;
}

.dashboard-chart-card {
  padding: 16px;
}

.dashboard-chart-card h3 {
  margin: 0 0 14px;
  color: var(--ui-text, #0f172a);
  font-size: 13px;
  font-weight: 700;
}

@media (max-width: 1023px) {
  .dashboard-kpi-strip,
  .dashboard-health-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .dashboard-kpi-strip :deep(.dashboard-metric:nth-child(2)),
  .dashboard-health-grid > div:nth-child(2) {
    border-right: 0;
  }

  .dashboard-kpi-strip :deep(.dashboard-metric:nth-child(-n + 2)),
  .dashboard-health-grid > div:nth-child(-n + 2) {
    border-bottom: 1px solid var(--ui-border, #dbe3ee);
  }

  .dashboard-chart-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .dashboard-chart-controls {
    width: 100%;
    justify-content: flex-start;
  }
}

@media (max-width: 639px) {
  .dashboard-section-heading {
    align-items: flex-start;
  }

  .dashboard-health-grid > div {
    padding: 13px 14px;
  }

  .dashboard-health-grid > div:nth-child(2n) {
    border-right: 0;
  }

  .dashboard-action-list {
    grid-template-columns: 1fr;
  }

  .dashboard-action-list > button {
    min-height: 64px;
    border-right: 0;
    border-bottom: 1px solid var(--ui-border, #dbe3ee);
  }

  .dashboard-action-list > button:last-child {
    border-bottom: 0;
  }

  .dashboard-chart-toolbar {
    padding: 14px;
  }

  .dashboard-chart-controls,
  .dashboard-control-group {
    width: 100%;
  }

  .dashboard-control-group {
    align-items: stretch;
    flex-direction: column;
  }

  .dashboard-granularity-control > div,
  .dashboard-chart-controls > .btn {
    width: 100%;
  }
}
</style>
