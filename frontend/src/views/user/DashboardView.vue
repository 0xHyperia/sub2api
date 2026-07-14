<template>
  <AppLayout>
    <div class="mx-auto min-w-0 max-w-page space-y-4 pb-8" :aria-busy="isRefreshing">
      <header class="flex flex-col gap-3 border-b border-outline pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div class="min-w-0">
          <h1 class="text-xl font-semibold text-foreground sm:text-2xl">
            {{ t('dashboard.title') }}
          </h1>
          <p class="mt-1 text-sm text-foreground-muted">
            {{ t('dashboard.welcomeMessage') }}
          </p>
        </div>
        <button
          type="button"
          class="btn btn-secondary btn-icon shrink-0"
          :disabled="isRefreshing"
          :title="t('common.refresh')"
          :aria-label="t('common.refresh')"
          @click="refreshAll"
        >
          <Icon name="refresh" size="md" :class="isRefreshing ? 'animate-spin' : ''" />
        </button>
      </header>

      <div
        v-if="statsLoadFailed && stats"
        class="flex flex-col gap-3 rounded-panel border border-danger/30 bg-danger-subtle p-4 text-danger-foreground sm:flex-row sm:items-center sm:justify-between"
        role="alert"
        data-testid="dashboard-refresh-error"
      >
        <span class="text-sm font-medium">{{ t('dashboard.loadFailed') }}</span>
        <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="loadStats">
          <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
          {{ t('common.retry') }}
        </button>
      </div>

      <div v-if="loading && !stats" class="flex min-h-64 items-center justify-center" role="status">
        <LoadingSpinner />
        <span class="sr-only">{{ t('common.loading') }}</span>
      </div>
      <div v-else-if="statsLoadFailed && !stats" data-testid="dashboard-load-error" class="card" role="alert">
        <EmptyState
          :title="t('dashboard.loadFailed')"
          :description="t('errors.tryAgain')"
          :action-text="t('common.refresh')"
          :action-icon="false"
          @action="refreshAll"
        />
      </div>
      <template v-else-if="stats">
        <UserDashboardStats :stats="stats" :balance="user?.balance || 0" :is-simple="authStore.isSimpleMode" :platform-quotas="platformQuotas" />
        <UserDashboardCharts v-model:startDate="startDate" v-model:endDate="endDate" v-model:granularity="granularity" :loading="loadingCharts" :error="chartsLoadFailed" :trend="trendData" :models="modelStats" @dateRangeChange="loadCharts" @granularityChange="loadCharts" @refresh="loadCharts" />
        <div class="grid min-w-0 grid-cols-1 items-start gap-4 lg:grid-cols-3">
          <div class="lg:col-span-2"><UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" :error="recentLoadFailed" @retry="loadRecent" /></div>
          <div class="lg:col-span-1"><UserDashboardQuickActions /></div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'; import EmptyState from '@/components/common/EmptyState.vue'; import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'; import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'; import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import type { UsageLog, TrendDataPoint, ModelStat, PlatformQuotaItem } from '@/types'
import { getMyPlatformQuotas } from '@/api/user'
import { formatDateLocalInput } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore(); const user = computed(() => authStore.user)
const stats = ref<UserStatsType | null>(null); const loading = ref(false); const loadingUsage = ref(false); const loadingCharts = ref(false)
const statsLoadFailed = ref(false)
const chartsLoadFailed = ref(false)
const recentLoadFailed = ref(false)
const trendData = ref<TrendDataPoint[]>([]); const modelStats = ref<ModelStat[]>([]); const recentUsage = ref<UsageLog[]>([])
const platformQuotas = ref<PlatformQuotaItem[] | null>(null)
const isRefreshing = computed(() => loading.value || loadingUsage.value || loadingCharts.value)
let statsLoadSequence = 0
let chartsLoadSequence = 0
let recentLoadSequence = 0

const startDate = ref(formatDateLocalInput(new Date(Date.now() - 6 * 86400000))); const endDate = ref(formatDateLocalInput(new Date())); const granularity = ref('day')

const loadStats = async () => {
  const sequence = ++statsLoadSequence
  loading.value = true
  statsLoadFailed.value = false
  try {
    await authStore.refreshUser()
    const nextStats = await usageAPI.getDashboardStats()
    if (sequence === statsLoadSequence) stats.value = nextStats
  } catch (error) {
    if (sequence === statsLoadSequence) statsLoadFailed.value = true
    console.error('Failed to load dashboard stats:', error)
  } finally {
    if (sequence === statsLoadSequence) loading.value = false
  }
}

const loadCharts = async () => {
  const sequence = ++chartsLoadSequence
  loadingCharts.value = true
  chartsLoadFailed.value = false
  try {
    const [trendResponse, modelResponse] = await Promise.all([
      usageAPI.getDashboardTrend({
        start_date: startDate.value,
        end_date: endDate.value,
        granularity: granularity.value as any,
      }),
      usageAPI.getDashboardModels({ start_date: startDate.value, end_date: endDate.value }),
    ])
    if (sequence !== chartsLoadSequence) return
    trendData.value = trendResponse.trend || []
    modelStats.value = modelResponse.models || []
  } catch (error) {
    if (sequence === chartsLoadSequence) chartsLoadFailed.value = true
    console.error('Failed to load charts:', error)
  } finally {
    if (sequence === chartsLoadSequence) loadingCharts.value = false
  }
}

const loadRecent = async () => {
  const sequence = ++recentLoadSequence
  loadingUsage.value = true
  recentLoadFailed.value = false
  try {
    const response = await usageAPI.getByDateRange(startDate.value, endDate.value)
    if (sequence === recentLoadSequence) recentUsage.value = response.items.slice(0, 5)
  } catch (error) {
    if (sequence === recentLoadSequence) recentLoadFailed.value = true
    console.error('Failed to load recent usage:', error)
  } finally {
    if (sequence === recentLoadSequence) loadingUsage.value = false
  }
}

const loadPlatformQuotas = async () => {
  try {
    const data = await getMyPlatformQuotas()
    platformQuotas.value = data.platform_quotas ?? []
  } catch (error) {
    console.warn('Failed to load platform quotas:', error)
    platformQuotas.value = []
  }
}

const refreshAll = () => {
  void Promise.all([loadStats(), loadCharts(), loadRecent(), loadPlatformQuotas()])
}

onMounted(() => { refreshAll() })
</script>
