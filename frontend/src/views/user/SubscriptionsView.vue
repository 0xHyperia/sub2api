<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1280px] space-y-5">
      <header class="page-header mb-0 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div class="min-w-0">
          <h1 class="page-title">{{ t('userSubscriptions.title') }}</h1>
          <p class="page-description">{{ t('userSubscriptions.description') }}</p>
        </div>
        <button type="button" class="btn btn-primary shrink-0" @click="router.push('/purchase?tab=subscription')">
          <Icon name="plus" size="sm" aria-hidden="true" />
          <span>{{ t('payment.tabSubscribe') }}</span>
        </button>
      </header>

      <div v-if="loading" class="card flex min-h-48 items-center justify-center" aria-live="polite">
        <LoadingSpinner />
      </div>

      <div v-else-if="loadFailed" class="card" role="alert">
        <EmptyState
          :title="t('userSubscriptions.failedToLoad')"
          :description="t('errors.tryAgain')"
          :action-text="t('common.refresh')"
          @action="loadSubscriptions"
        />
      </div>

      <div v-else-if="subscriptions.length === 0" class="card">
        <EmptyState
          :title="t('userSubscriptions.noActiveSubscriptions')"
          :description="t('userSubscriptions.noActiveSubscriptionsDesc')"
          :action-text="t('payment.tabSubscribe')"
          action-to="/purchase?tab=subscription"
        />
      </div>

      <section v-else class="grid gap-4 xl:grid-cols-2" :aria-label="t('userSubscriptions.title')">
        <article
          v-for="subscription in subscriptions"
          :key="subscription.id"
          class="card min-w-0 overflow-hidden"
        >
          <header class="flex min-w-0 flex-col gap-4 border-b border-outline p-4 sm:p-5">
            <div class="flex min-w-0 items-start justify-between gap-4">
              <div class="flex min-w-0 items-start gap-3">
                <span :class="['mt-1.5 h-2.5 w-2.5 shrink-0 rounded-full', platformAccentDotClass(subscription.group?.platform || '')]" aria-hidden="true" />
                <div class="min-w-0">
                  <div class="flex min-w-0 flex-wrap items-center gap-2">
                    <h2 class="break-words text-lg font-semibold text-foreground">
                      {{ subscription.group?.name || 'Group #' + subscription.group_id }}
                    </h2>
                    <span :class="['badge', platformBadgeClass(subscription.group?.platform || '')]">
                      {{ platformLabel(subscription.group?.platform || '') }}
                    </span>
                  </div>
                  <p v-if="subscription.group?.description" class="mt-2 break-words text-sm leading-6 text-foreground-subtle">
                    {{ subscription.group.description }}
                  </p>
                </div>
              </div>

              <span
                :class="[
                  'badge shrink-0',
                  subscription.status === 'active'
                    ? 'badge-success'
                    : subscription.status === 'expired'
                      ? 'badge-gray'
                      : 'badge-danger'
                ]"
              >
                {{ t('userSubscriptions.status.' + subscription.status) }}
              </span>
            </div>

            <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <div class="flex flex-wrap gap-x-4 gap-y-1 text-sm text-foreground-subtle">
                <span>{{ t('payment.planCard.rate') }}: x{{ subscription.group?.rate_multiplier ?? 1 }}</span>
                <span v-if="subscriptionHasPeakRate(subscription)" class="text-warning-foreground">
                  {{ t('payment.planCard.peakRate') }}: {{ subscriptionPeakRateLabel(subscription) }}
                </span>
              </div>
              <button
                v-if="subscription.status === 'active'"
                type="button"
                class="btn btn-secondary btn-sm self-start sm:self-auto"
                @click="router.push({ path: '/purchase', query: { tab: 'subscription', group: String(subscription.group_id) } })"
              >
                {{ t('payment.renewNow') }}
              </button>
            </div>
          </header>

          <div class="space-y-6 p-4 sm:p-5">
            <dl class="grid gap-3 sm:grid-cols-2">
              <div class="rounded-panel bg-surface-subtle px-3 py-3">
                <dt class="text-xs text-foreground-subtle">{{ t('userSubscriptions.expires') }}</dt>
                <dd v-if="subscription.expires_at" :class="['mt-1 text-sm font-semibold', getExpirationClass(subscription.expires_at)]">
                  {{ formatExpirationDate(subscription.expires_at) }}
                </dd>
                <dd v-else class="mt-1 text-sm font-semibold text-foreground-muted">
                  {{ t('userSubscriptions.noExpiration') }}
                </dd>
              </div>
              <div class="rounded-panel bg-surface-subtle px-3 py-3">
                <dt class="text-xs text-foreground-subtle">{{ t('userSubscriptions.usage') }}</dt>
                <dd class="mt-1 text-sm font-semibold text-foreground">
                  {{ subscription.group?.daily_limit_usd || subscription.group?.weekly_limit_usd || subscription.group?.monthly_limit_usd ? t('userSubscriptions.usage') : t('userSubscriptions.unlimited') }}
                </dd>
              </div>
            </dl>

            <div
              v-if="
                subscription.group?.daily_limit_usd ||
                subscription.group?.weekly_limit_usd ||
                subscription.group?.monthly_limit_usd
              "
              class="space-y-5 border-t border-outline pt-5"
            >
              <div v-if="subscription.group?.daily_limit_usd" class="space-y-2">
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <span class="text-sm font-medium text-foreground">{{ t('userSubscriptions.daily') }}</span>
                  <span class="text-sm tabular-nums text-foreground-subtle">
                    {{ '$' }}{{ (subscription.daily_usage_usd || 0).toFixed(2) }} / {{ '$' }}{{ subscription.group.daily_limit_usd.toFixed(2) }}
                  </span>
                </div>
                <div
                  class="progress"
                  role="progressbar"
                  :aria-label="t('userSubscriptions.daily')"
                  :aria-valuenow="getProgressValue(subscription.daily_usage_usd, subscription.group.daily_limit_usd)"
                  aria-valuemin="0"
                  aria-valuemax="100"
                >
                  <div
                    class="h-full rounded-full transition-[width]"
                    :class="getProgressBarClass(subscription.daily_usage_usd, subscription.group.daily_limit_usd)"
                    :style="{ width: getProgressWidth(subscription.daily_usage_usd, subscription.group.daily_limit_usd) }"
                  />
                </div>
                <p v-if="subscription.daily_window_start" class="text-xs text-foreground-subtle">
                  {{ formatDailyUsageWindow(subscription) }}
                </p>
              </div>

              <div v-if="subscription.group?.weekly_limit_usd" class="space-y-2">
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <span class="text-sm font-medium text-foreground">{{ t('userSubscriptions.weekly') }}</span>
                  <span class="text-sm tabular-nums text-foreground-subtle">
                    {{ '$' }}{{ (subscription.weekly_usage_usd || 0).toFixed(2) }} / {{ '$' }}{{ subscription.group.weekly_limit_usd.toFixed(2) }}
                  </span>
                </div>
                <div
                  class="progress"
                  role="progressbar"
                  :aria-label="t('userSubscriptions.weekly')"
                  :aria-valuenow="getProgressValue(subscription.weekly_usage_usd, subscription.group.weekly_limit_usd)"
                  aria-valuemin="0"
                  aria-valuemax="100"
                >
                  <div
                    class="h-full rounded-full transition-[width]"
                    :class="getProgressBarClass(subscription.weekly_usage_usd, subscription.group.weekly_limit_usd)"
                    :style="{ width: getProgressWidth(subscription.weekly_usage_usd, subscription.group.weekly_limit_usd) }"
                  />
                </div>
                <p v-if="subscription.weekly_window_start" class="text-xs text-foreground-subtle">
                  {{ t('userSubscriptions.resetIn', { time: formatResetTime(subscription.weekly_window_start, 168) }) }}
                </p>
              </div>

              <div v-if="subscription.group?.monthly_limit_usd" class="space-y-2">
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <span class="text-sm font-medium text-foreground">{{ t('userSubscriptions.monthly') }}</span>
                  <span class="text-sm tabular-nums text-foreground-subtle">
                    {{ '$' }}{{ (subscription.monthly_usage_usd || 0).toFixed(2) }} / {{ '$' }}{{ subscription.group.monthly_limit_usd.toFixed(2) }}
                  </span>
                </div>
                <div
                  class="progress"
                  role="progressbar"
                  :aria-label="t('userSubscriptions.monthly')"
                  :aria-valuenow="getProgressValue(subscription.monthly_usage_usd, subscription.group.monthly_limit_usd)"
                  aria-valuemin="0"
                  aria-valuemax="100"
                >
                  <div
                    class="h-full rounded-full transition-[width]"
                    :class="getProgressBarClass(subscription.monthly_usage_usd, subscription.group.monthly_limit_usd)"
                    :style="{ width: getProgressWidth(subscription.monthly_usage_usd, subscription.group.monthly_limit_usd) }"
                  />
                </div>
                <p v-if="subscription.monthly_window_start" class="text-xs text-foreground-subtle">
                  {{ t('userSubscriptions.resetIn', { time: formatResetTime(subscription.monthly_window_start, 720) }) }}
                </p>
              </div>
            </div>

            <div v-else class="flex items-center gap-2 border-t border-outline pt-5 text-sm text-success-foreground">
              <Icon name="checkCircle" size="sm" aria-hidden="true" />
              <div>
                <p class="font-medium">{{ t('userSubscriptions.unlimited') }}</p>
                <p class="mt-1 text-xs text-foreground-subtle">{{ t('userSubscriptions.unlimitedDesc') }}</p>
              </div>
            </div>
          </div>
        </article>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import subscriptionsAPI from '@/api/subscriptions'
import type { UserSubscription } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateOnly } from '@/utils/format'
import { hasPeakRate, formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'
import { platformBadgeClass, platformLabel } from '@/utils/platformColors'
import { getRemainingDurationParts, isOneTimeDailyQuota, type RemainingDurationParts } from '@/utils/subscriptionQuota'

function platformAccentDotClass(p: string): string {
  switch (p) {
    case 'anthropic': return 'bg-warning'
    case 'openai': return 'bg-success'
    case 'antigravity': return 'bg-brand'
    case 'gemini': return 'bg-info'
    default: return 'bg-outline-strong'
  }
}

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const subscriptions = ref<UserSubscription[]>([])
const loading = ref(true)
const loadFailed = ref(false)

function subscriptionHasPeakRate(subscription: UserSubscription): boolean {
  return hasPeakRate(subscription.group)
}

function subscriptionPeakRateLabel(subscription: UserSubscription): string {
  return formatPeakRateWindow(subscription.group, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
}

async function loadSubscriptions() {
  try {
    loading.value = true
    loadFailed.value = false
    subscriptions.value = await subscriptionsAPI.getMySubscriptions()
  } catch (error) {
    console.error('Failed to load subscriptions:', error)
    loadFailed.value = true
    appStore.showError(t('userSubscriptions.failedToLoad'))
  } finally {
    loading.value = false
  }
}

function getProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return '0%'
  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

function getProgressValue(used: number | undefined, limit: number | null | undefined): number {
  if (!limit || limit <= 0) return 0
  return Math.min(Math.round(((used || 0) / limit) * 100), 100)
}

function getProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return 'bg-outline-strong'
  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 90) return 'bg-danger'
  if (percentage >= 70) return 'bg-warning'
  return 'bg-success'
}

function formatExpirationDate(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (days < 0) {
    return t('userSubscriptions.status.expired')
  }

  const dateStr = formatDateOnly(expires)

  if (days === 0) {
    return `${dateStr} (${t('common.today')})`
  }
  if (days === 1) {
    return `${dateStr} (${t('common.tomorrow')})`
  }

  return t('userSubscriptions.daysRemaining', { days }) + ` (${dateStr})`
}

function getExpirationClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (days <= 0) return 'font-medium text-danger-foreground'
  if (days <= 3) return 'text-danger-foreground'
  if (days <= 7) return 'text-warning-foreground'
  return 'text-foreground-muted'
}

function formatDurationParts(parts: RemainingDurationParts): string {
  if (parts.days > 0) {
    return `${parts.days}d ${parts.hours}h`
  }

  if (parts.hours > 0) {
    return `${parts.hours}h ${parts.minutes}m`
  }

  return `${parts.minutes}m`
}

function formatDailyUsageWindow(subscription: UserSubscription): string {
  if (isOneTimeDailyQuota(subscription) && subscription.expires_at) {
    const parts = getRemainingDurationParts(subscription.expires_at)
    if (!parts) return t('userSubscriptions.windowNotActive')
    return t('userSubscriptions.quotaEndsIn', { time: formatDurationParts(parts) })
  }

  return t('userSubscriptions.resetIn', {
    time: formatResetTime(subscription.daily_window_start, 24)
  })
}

function formatResetTime(windowStart: string | null, windowHours: number): string {
  if (!windowStart) return t('userSubscriptions.windowNotActive')

  const start = new Date(windowStart)
  const end = new Date(start.getTime() + windowHours * 60 * 60 * 1000)
  const parts = getRemainingDurationParts(end)

  return parts ? formatDurationParts(parts) : t('userSubscriptions.windowNotActive')
}

onMounted(() => {
  loadSubscriptions()
})
</script>
