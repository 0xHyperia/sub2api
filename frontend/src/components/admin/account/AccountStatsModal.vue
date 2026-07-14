<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.usageStatistics')"
    width="extra-wide"
    @close="handleClose"
  >
    <div class="space-y-4">
      <!-- Account Info Header -->
      <div
        v-if="account"
        class="flex min-w-0 items-center justify-between gap-3 rounded-panel border border-outline bg-surface-subtle p-3"
      >
        <div class="flex min-w-0 items-center gap-3">
          <div
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-control border border-outline bg-surface text-foreground-muted"
          >
            <Icon name="chartBar" size="md" />
          </div>
          <div class="min-w-0">
            <div class="truncate font-semibold text-foreground" :title="account.name">{{ account.name }}</div>
            <div class="text-xs text-foreground-subtle">
              {{ t('admin.accounts.last30DaysUsage') }}
            </div>
          </div>
        </div>
        <span
          :class="[
            'shrink-0 rounded-control border px-2.5 py-1 text-xs font-semibold',
            account.status === 'active'
              ? 'border-success/30 bg-success-subtle text-success-foreground'
              : 'border-outline bg-surface text-foreground-muted'
          ]"
        >
          {{ account.status }}
        </span>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-12" role="status" :aria-label="t('common.loading')">
        <LoadingSpinner />
      </div>

      <div
        v-else-if="loadError"
        class="flex min-h-48 flex-col items-center justify-center gap-3 rounded-panel border border-danger/30 bg-danger-subtle p-5 text-center text-danger-foreground"
        role="alert"
        data-testid="account-stats-load-error"
      >
        <Icon name="exclamationTriangle" size="lg" aria-hidden="true" />
        <p class="text-sm font-medium">{{ t('admin.accounts.stats.failedToLoad') }}</p>
        <button type="button" class="btn btn-secondary btn-sm" @click="loadStats">
          <Icon name="refresh" size="sm" />
          {{ t('common.retry') }}
        </button>
      </div>

      <template v-else-if="stats">
        <!-- Row 1: Main Stats Cards -->
        <div class="grid grid-cols-1 gap-3 min-[420px]:grid-cols-2 lg:grid-cols-4">
          <!-- 30-Day Total Cost -->
          <div class="rounded-panel border border-outline bg-surface-subtle p-4">
            <div class="mb-2 flex items-center justify-between">
              <span class="text-xs font-medium text-foreground-muted">{{
                t('admin.accounts.stats.totalCost')
              }}</span>
              <div class="flex h-8 w-8 items-center justify-center rounded-control bg-success-subtle text-success-foreground">
                <Icon name="dollar" size="sm" />
              </div>
            </div>
            <p class="text-2xl font-bold text-foreground">
              ${{ formatCost(stats.summary.total_cost) }}
            </p>
            <p class="mt-1 text-xs leading-5 text-foreground-muted">
              {{ t('admin.accounts.stats.accumulatedCost') }}
              <span class="text-foreground-subtle">
                ({{ t('usage.userBilled') }}: ${{ formatCost(stats.summary.total_user_cost) }} ·
                {{ t('admin.accounts.stats.standardCost') }}: ${{
                  formatCost(stats.summary.total_standard_cost)
                }})
              </span>
            </p>
          </div>

          <!-- 30-Day Total Requests -->
          <div class="rounded-panel border border-outline bg-surface-subtle p-4">
            <div class="mb-2 flex items-center justify-between">
              <span class="text-xs font-medium text-foreground-muted">{{
                t('admin.accounts.stats.totalRequests')
              }}</span>
              <div class="flex h-8 w-8 items-center justify-center rounded-control bg-info-subtle text-info-foreground">
                <Icon name="bolt" size="sm" />
              </div>
            </div>
            <p class="text-2xl font-bold text-foreground">
              {{ formatNumber(stats.summary.total_requests) }}
            </p>
            <p class="mt-1 text-xs text-foreground-muted">
              {{ t('admin.accounts.stats.totalCalls') }}
            </p>
          </div>

          <!-- Daily Average Cost -->
          <div class="rounded-panel border border-outline bg-surface-subtle p-4">
            <div class="mb-2 flex items-center justify-between">
              <span class="text-xs font-medium text-foreground-muted">{{
                t('admin.accounts.stats.avgDailyCost')
              }}</span>
              <div class="flex h-8 w-8 items-center justify-center rounded-control bg-warning-subtle text-warning-foreground">
                <Icon name="calculator" size="sm" />
              </div>
            </div>
            <p class="text-2xl font-bold text-foreground">
              ${{ formatCost(stats.summary.avg_daily_cost) }}
            </p>
             <p class="mt-1 text-xs leading-5 text-foreground-muted">
              {{
                t('admin.accounts.stats.basedOnActualDays', {
                  days: stats.summary.actual_days_used
                })
              }}
              <span class="text-foreground-subtle">
                ({{ t('usage.userBilled') }}: ${{ formatCost(stats.summary.avg_daily_user_cost) }})
              </span>
            </p>
          </div>

          <!-- Daily Average Requests -->
          <div class="rounded-panel border border-outline bg-surface-subtle p-4">
            <div class="mb-2 flex items-center justify-between">
              <span class="text-xs font-medium text-foreground-muted">{{
                t('admin.accounts.stats.avgDailyRequests')
              }}</span>
              <div class="flex h-8 w-8 items-center justify-center rounded-control bg-surface text-foreground-muted">
                <Icon name="trendingUp" size="sm" />
              </div>
            </div>
            <p class="text-2xl font-bold text-foreground">
              {{ formatNumber(Math.round(stats.summary.avg_daily_requests)) }}
            </p>
            <p class="mt-1 text-xs text-foreground-muted">
              {{ t('admin.accounts.stats.avgDailyUsage') }}
            </p>
          </div>
        </div>

        <!-- Row 2: Today, Highest Cost, Highest Requests -->
        <div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
          <!-- Today Overview -->
          <div class="rounded-panel border border-outline bg-surface p-4">
            <div class="mb-3 flex items-center gap-2">
              <div class="flex h-8 w-8 items-center justify-center rounded-control bg-surface-subtle text-foreground-muted">
                <Icon name="clock" size="sm" />
              </div>
              <span class="text-sm font-semibold text-foreground">{{
                t('admin.accounts.stats.todayOverview')
              }}</span>
            </div>
            <div class="space-y-2">
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{ t('usage.accountBilled') }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground"
                  >${{ formatCost(stats.summary.today?.cost || 0) }}</span
                >
              </div>
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{ t('usage.userBilled') }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground"
                  >${{ formatCost(stats.summary.today?.user_cost || 0) }}</span
                >
              </div>
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{
                  t('admin.accounts.stats.requests')
                }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground">{{
                  formatNumber(stats.summary.today?.requests || 0)
                }}</span>
              </div>
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{
                  t('admin.accounts.stats.tokens')
                }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground">{{
                  formatTokens(stats.summary.today?.tokens || 0)
                }}</span>
              </div>
            </div>
          </div>

          <!-- Highest Cost Day -->
          <div class="rounded-panel border border-outline bg-surface p-4">
            <div class="mb-3 flex items-center gap-2">
              <div class="flex h-8 w-8 items-center justify-center rounded-control bg-warning-subtle text-warning-foreground">
                <Icon name="fire" size="sm" />
              </div>
              <span class="text-sm font-semibold text-foreground">{{
                t('admin.accounts.stats.highestCostDay')
              }}</span>
            </div>
            <div class="space-y-2">
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{
                  t('admin.accounts.stats.date')
                }}</span>
                <span class="break-words text-right text-sm font-semibold text-foreground">{{
                  stats.summary.highest_cost_day?.label || '-'
                }}</span>
              </div>
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{ t('usage.accountBilled') }}</span>
                <span class="break-all text-right text-sm font-semibold text-warning-foreground"
                  >${{ formatCost(stats.summary.highest_cost_day?.cost || 0) }}</span
                >
              </div>
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{ t('usage.userBilled') }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground"
                  >${{ formatCost(stats.summary.highest_cost_day?.user_cost || 0) }}</span
                >
              </div>
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{
                  t('admin.accounts.stats.requests')
                }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground">{{
                  formatNumber(stats.summary.highest_cost_day?.requests || 0)
                }}</span>
              </div>
            </div>
          </div>

          <!-- Highest Request Day -->
          <div class="rounded-panel border border-outline bg-surface p-4">
            <div class="mb-3 flex items-center gap-2">
              <div class="flex h-8 w-8 items-center justify-center rounded-control bg-info-subtle text-info-foreground">
                <Icon name="trendingUp" size="sm" />
              </div>
              <span class="text-sm font-semibold text-foreground">{{
                t('admin.accounts.stats.highestRequestDay')
              }}</span>
            </div>
            <div class="space-y-2">
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{
                  t('admin.accounts.stats.date')
                }}</span>
                <span class="break-words text-right text-sm font-semibold text-foreground">{{
                  stats.summary.highest_request_day?.label || '-'
                }}</span>
              </div>
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{
                  t('admin.accounts.stats.requests')
                }}</span>
                <span class="break-all text-right text-sm font-semibold text-info-foreground">{{
                  formatNumber(stats.summary.highest_request_day?.requests || 0)
                }}</span>
              </div>
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{ t('usage.accountBilled') }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground"
                  >${{ formatCost(stats.summary.highest_request_day?.cost || 0) }}</span
                >
              </div>
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{ t('usage.userBilled') }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground"
                  >${{ formatCost(stats.summary.highest_request_day?.user_cost || 0) }}</span
                >
              </div>
            </div>
          </div>
        </div>

        <!-- Row 3: Token Stats -->
        <div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
          <!-- Accumulated Tokens -->
          <div class="rounded-panel border border-outline bg-surface p-4">
            <div class="mb-3 flex items-center gap-2">
              <div class="flex h-8 w-8 items-center justify-center rounded-control bg-surface-subtle text-foreground-muted">
                <Icon name="cube" size="sm" />
              </div>
              <span class="text-sm font-semibold text-foreground">{{
                t('admin.accounts.stats.accumulatedTokens')
              }}</span>
            </div>
            <div class="space-y-2">
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{
                  t('admin.accounts.stats.totalTokens')
                }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground">{{
                  formatTokens(stats.summary.total_tokens)
                }}</span>
              </div>
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{
                  t('admin.accounts.stats.dailyAvgTokens')
                }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground">{{
                  formatTokens(Math.round(stats.summary.avg_daily_tokens))
                }}</span>
              </div>
            </div>
          </div>

          <!-- Performance -->
          <div class="rounded-panel border border-outline bg-surface p-4">
            <div class="mb-3 flex items-center gap-2">
              <div class="flex h-8 w-8 items-center justify-center rounded-control bg-surface-subtle text-foreground-muted">
                <Icon name="bolt" size="sm" />
              </div>
              <span class="text-sm font-semibold text-foreground">{{
                t('admin.accounts.stats.performance')
              }}</span>
            </div>
            <div class="space-y-2">
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{
                  t('admin.accounts.stats.avgResponseTime')
                }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground">{{
                  formatDuration(stats.summary.avg_duration_ms)
                }}</span>
              </div>
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{
                  t('admin.accounts.stats.daysActive')
                }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground"
                  >{{ stats.summary.actual_days_used }} / {{ stats.summary.days }}</span
                >
              </div>
            </div>
          </div>

          <!-- Recent Activity -->
          <div class="rounded-panel border border-outline bg-surface p-4">
            <div class="mb-3 flex items-center gap-2">
              <div class="flex h-8 w-8 items-center justify-center rounded-control bg-surface-subtle text-foreground-muted">
                <Icon name="clipboard" size="sm" />
              </div>
              <span class="text-sm font-semibold text-foreground">{{
                t('admin.accounts.stats.recentActivity')
              }}</span>
            </div>
            <div class="space-y-2">
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{
                  t('admin.accounts.stats.todayRequests')
                }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground">{{
                  formatNumber(stats.summary.today?.requests || 0)
                }}</span>
              </div>
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{
                  t('admin.accounts.stats.todayTokens')
                }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground">{{
                  formatTokens(stats.summary.today?.tokens || 0)
                }}</span>
              </div>
              <div class="flex items-start justify-between gap-3">
                <span class="text-xs text-foreground-muted">{{
                  t('admin.accounts.stats.todayCost')
                }}</span>
                <span class="break-all text-right text-sm font-semibold text-foreground"
                  >${{ formatCost(stats.summary.today?.cost || 0) }}</span
                >
              </div>
            </div>
          </div>
        </div>

        <!-- Usage Trend Chart -->
        <div class="rounded-panel border border-outline bg-surface p-4">
          <h3 class="mb-4 text-sm font-semibold text-foreground">
            {{ t('admin.accounts.stats.usageTrend') }}
          </h3>
          <div class="h-64">
            <Line v-if="trendChartData" :data="trendChartData" :options="lineChartOptions" />
            <div
              v-else
              class="flex h-full items-center justify-center text-sm text-foreground-muted"
            >
              {{ t('admin.dashboard.noDataAvailable') }}
            </div>
          </div>
        </div>

        <!-- Model Distribution -->
        <ModelDistributionChart :model-stats="stats.models" :loading="false" />

        <EndpointDistributionChart
          :endpoint-stats="stats.endpoints || []"
          :loading="false"
          :title="t('usage.inboundEndpoint')"
        />

        <EndpointDistributionChart
          :endpoint-stats="stats.upstream_endpoints || []"
          :loading="false"
          :title="t('usage.upstreamEndpoint')"
        />
      </template>

      <!-- No Data State -->
      <div
        v-else-if="!loading"
        class="flex flex-col items-center justify-center py-12 text-foreground-muted"
      >
        <Icon name="chartBar" size="xl" class="mb-4 h-12 w-12" />
        <p class="text-sm">{{ t('admin.accounts.stats.noData') }}</p>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button
          type="button"
          @click="handleClose"
          class="btn btn-secondary"
        >
          {{ t('common.close') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'
import BaseDialog from '@/components/common/BaseDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import EndpointDistributionChart from '@/components/charts/EndpointDistributionChart.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import type { Account, AccountUsageStatsResponse } from '@/types'
import { useChartTheme } from '@/composables/useChartTheme'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const { t } = useI18n()
const { chartTheme } = useChartTheme()

const props = defineProps<{
  show: boolean
  account: Account | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const loading = ref(false)
const stats = ref<AccountUsageStatsResponse | null>(null)
const loadError = ref(false)
let loadSequence = 0

// Chart colors
const chartColors = computed(() => ({
  text: chartTheme.value.foregroundMuted,
  grid: chartTheme.value.outline,
}))

// Line chart data
const trendChartData = computed(() => {
  if (!stats.value?.history?.length) return null

  return {
    labels: stats.value.history.map((h) => h.label),
    datasets: [
      {
        label: t('usage.accountBilled') + ' (USD)',
        data: stats.value.history.map((h) => h.actual_cost),
        borderColor: chartTheme.value.info,
        backgroundColor: chartTheme.value.infoAlpha,
        fill: true,
        tension: 0.3,
        yAxisID: 'y'
      },
      {
        label: t('usage.userBilled') + ' (USD)',
        data: stats.value.history.map((h) => h.user_cost),
        borderColor: chartTheme.value.success,
        backgroundColor: chartTheme.value.successAlpha,
        fill: false,
        tension: 0.3,
        borderDash: [5, 5],
        yAxisID: 'y'
      },
      {
        label: t('admin.accounts.stats.requests'),
        data: stats.value.history.map((h) => h.requests),
        borderColor: chartTheme.value.warning,
        backgroundColor: chartTheme.value.warningAlpha,
        fill: false,
        tension: 0.3,
        yAxisID: 'y1'
      }
    ]
  }
})

// Line chart options with dual Y-axis
const lineChartOptions = computed(() => ({
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
      callbacks: {
        label: (context: any) => {
          const label = context.dataset.label || ''
          const value = context.raw
          if (label.includes('USD')) {
            return `${label}: $${formatCost(value)}`
          }
          return `${label}: ${formatNumber(value)}`
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
        },
        maxRotation: 45,
        minRotation: 0
      }
    },
    y: {
      type: 'linear' as const,
      display: true,
      position: 'left' as const,
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: '#3b82f6',
        font: {
          size: 10
        },
        callback: (value: string | number) => '$' + formatCost(Number(value))
      },
      title: {
        display: true,
        text: t('usage.accountBilled') + ' (USD)',
        color: '#3b82f6',
        font: {
          size: 11
        }
      }
    },
    y1: {
      type: 'linear' as const,
      display: true,
      position: 'right' as const,
      grid: {
        drawOnChartArea: false
      },
      ticks: {
        color: '#f97316',
        font: {
          size: 10
        },
        callback: (value: string | number) => formatNumber(Number(value))
      },
      title: {
        display: true,
        text: t('admin.accounts.stats.requests'),
        color: '#f97316',
        font: {
          size: 11
        }
      }
    }
  }
}))

const loadStats = async () => {
  const accountId = props.account?.id
  if (!props.show || accountId == null) return

  const requestSequence = ++loadSequence
  loading.value = true
  loadError.value = false
  try {
    const response = await adminAPI.accounts.getStats(accountId, 30)
    if (
      requestSequence !== loadSequence ||
      !props.show ||
      props.account?.id !== accountId
    ) {
      return
    }
    stats.value = response
  } catch (error) {
    if (requestSequence !== loadSequence) return
    console.error('Failed to load account stats:', error)
    loadError.value = true
  } finally {
    if (requestSequence === loadSequence) {
      loading.value = false
    }
  }
}

// Reload for the selected account and invalidate responses from a previous account.
watch(
  () => [props.show, props.account?.id] as const,
  (current, previous) => {
    const [isOpen, accountId] = current
    if (isOpen && accountId != null) {
      if (!previous || previous[1] !== accountId) {
        stats.value = null
      }
      void loadStats()
      return
    }

    loadSequence += 1
    stats.value = null
    loadError.value = false
    loading.value = false
  },
  { immediate: true }
)

const handleClose = () => {
  emit('close')
}

// Format helpers
const formatCost = (value: number): string => {
  if (value >= 1000) {
    return (value / 1000).toFixed(2) + 'K'
  } else if (value >= 1) {
    return value.toFixed(2)
  } else if (value >= 0.01) {
    return value.toFixed(3)
  }
  return value.toFixed(4)
}

const formatNumber = (value: number): string => {
  if (value >= 1_000_000) {
    return (value / 1_000_000).toFixed(2) + 'M'
  } else if (value >= 1_000) {
    return (value / 1_000).toFixed(2) + 'K'
  }
  return value.toLocaleString()
}

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return value.toLocaleString()
}

const formatDuration = (ms: number): string => {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)}s`
  }
  return `${Math.round(ms)}ms`
}
</script>
