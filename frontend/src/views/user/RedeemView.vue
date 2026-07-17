<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-5xl space-y-5">
      <header class="page-header">
        <h1 class="page-title">{{ t('redeem.title') }}</h1>
        <p class="page-description">{{ t('redeem.description') }}</p>
      </header>

      <dl class="card grid grid-cols-1 divide-y divide-outline overflow-hidden min-[420px]:grid-cols-2 min-[420px]:divide-x min-[420px]:divide-y-0">
        <div class="min-w-0 p-4 sm:p-5">
          <dt class="flex items-center gap-2 text-sm text-foreground-subtle">
            <Icon name="dollar" size="sm" class="text-info-foreground" aria-hidden="true" />
            {{ t('redeem.currentBalance') }}
          </dt>
          <dd class="mt-2 break-words text-2xl font-semibold tabular-nums text-foreground">
            ${{ user?.balance?.toFixed(2) || '0.00' }}
          </dd>
        </div>
        <div class="min-w-0 p-4 sm:p-5">
          <dt class="flex items-center gap-2 text-sm text-foreground-subtle">
            <Icon name="bolt" size="sm" class="text-warning-foreground" aria-hidden="true" />
            {{ t('redeem.concurrency') }}
          </dt>
          <dd class="mt-2 break-words text-2xl font-semibold tabular-nums text-foreground">
            {{ user?.concurrency || 0 }}
            <span class="text-sm font-normal text-foreground-subtle">{{ t('redeem.requests') }}</span>
          </dd>
        </div>
      </dl>

      <div class="grid items-start gap-5 lg:grid-cols-[minmax(0,1fr)_minmax(280px,360px)]">
        <div class="min-w-0 space-y-4">
          <div class="card p-4 sm:p-5">
            <RedeemCodeForm :show-description="true" @redeemed="fetchHistory" />
          </div>
        </div>

        <aside class="card p-4 sm:p-5" aria-labelledby="redeem-about-title">
          <div class="flex items-center gap-2">
            <Icon name="infoCircle" size="md" class="text-info-foreground" aria-hidden="true" />
            <h2 id="redeem-about-title" class="text-base font-semibold text-foreground">
              {{ t('redeem.aboutCodes') }}
            </h2>
          </div>
          <ul class="mt-3 list-disc space-y-2 pl-5 text-sm leading-6 text-foreground-muted">
            <li>{{ t('redeem.codeRule1') }}</li>
            <li>{{ t('redeem.codeRule2') }}</li>
            <li>
              {{ t('redeem.codeRule3') }}
              <span v-if="contactInfo" class="badge badge-gray ml-1 align-middle">{{ contactInfo }}</span>
            </li>
            <li>{{ t('redeem.codeRule4') }}</li>
          </ul>
        </aside>
      </div>

      <section class="card overflow-hidden" aria-labelledby="redeem-history-title">
        <header class="flex items-center justify-between gap-3 border-b border-outline px-4 py-3.5 sm:px-5">
          <h2 id="redeem-history-title" class="text-base font-semibold text-foreground">
            {{ t('redeem.recentActivity') }}
          </h2>
          <button
            type="button"
            class="btn btn-ghost btn-icon"
            :disabled="loadingHistory"
            :title="t('common.refresh')"
            :aria-label="t('common.refresh')"
            @click="fetchHistory"
          >
            <Icon name="refresh" size="sm" :class="{ 'animate-spin': loadingHistory }" aria-hidden="true" />
          </button>
        </header>

        <div v-if="loadingHistory" class="flex min-h-36 items-center justify-center" aria-live="polite">
          <LoadingSpinner />
        </div>

        <div v-else-if="historyError" class="p-4 sm:p-5" role="alert">
          <div class="flex flex-col gap-3 rounded-panel border border-danger/20 bg-danger-subtle p-4 text-sm text-danger-foreground sm:flex-row sm:items-center sm:justify-between">
            <span>{{ t('common.error') }}. {{ t('errors.tryAgain') }}</span>
            <button type="button" class="btn btn-secondary btn-sm" @click="fetchHistory">
              <Icon name="refresh" size="sm" aria-hidden="true" />
              <span>{{ t('common.refresh') }}</span>
            </button>
          </div>
        </div>

        <ul v-else-if="history.length > 0" class="divide-y divide-outline">
          <li
            v-for="item in history"
            :key="item.id"
            class="flex min-w-0 flex-col gap-3 px-4 py-3.5 sm:flex-row sm:items-center sm:justify-between sm:px-5"
          >
            <div class="flex min-w-0 items-start gap-3">
              <span
                :class="[
                  'flex h-9 w-9 shrink-0 items-center justify-center rounded-panel',
                  historyIconSurfaceClass(item),
                ]"
                aria-hidden="true"
              >
                <Icon
                  v-if="isBalanceType(item.type)"
                  name="dollar"
                  size="sm"
                  :class="historyIconClass(item)"
                />
                <Icon
                  v-else-if="isSubscriptionType(item.type)"
                  name="badge"
                  size="sm"
                  class="text-info-foreground"
                />
                <Icon v-else name="bolt" size="sm" :class="historyIconClass(item)" />
              </span>
              <div class="min-w-0">
                <p class="break-words text-sm font-medium text-foreground">{{ getHistoryItemTitle(item) }}</p>
                <p class="mt-0.5 text-xs text-foreground-subtle">{{ formatDateTime(item.used_at) }}</p>
                <p v-if="item.notes" class="mt-1 break-words text-xs italic text-foreground-muted">
                  {{ item.notes }}
                </p>
              </div>
            </div>
            <div class="min-w-0 pl-12 text-left sm:pl-0 sm:text-right">
              <p :class="['break-words text-sm font-semibold tabular-nums', historyValueClass(item)]">
                {{ formatHistoryValue(item) }}
              </p>
              <code v-if="!isAdminAdjustment(item.type)" class="text-xs text-foreground-subtle" :title="item.code">
                {{ item.code.slice(0, 8) }}...
              </code>
              <p v-else class="text-xs text-foreground-subtle">{{ t('redeem.adminAdjustment') }}</p>
            </div>
          </li>
        </ul>

        <EmptyState v-else :title="t('redeem.historyWillAppear')" />
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { redeemAPI, authAPI, type RedeemHistoryItem } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import RedeemCodeForm from '@/components/payment/RedeemCodeForm.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()

const user = computed(() => authStore.user)
const history = ref<RedeemHistoryItem[]>([])
const loadingHistory = ref(false)
const historyError = ref(false)
const contactInfo = ref('')

const isBalanceType = (type: string) => type === 'balance' || type === 'admin_balance'
const isSubscriptionType = (type: string) => type === 'subscription'
const isAdminAdjustment = (type: string) =>
  type === 'admin_balance' || type === 'admin_concurrency'

function getHistoryItemTitle(item: RedeemHistoryItem): string {
  if (item.type === 'balance') return t('redeem.balanceAddedRedeem')
  if (item.type === 'admin_balance') {
    return item.value >= 0 ? t('redeem.balanceAddedAdmin') : t('redeem.balanceDeductedAdmin')
  }
  if (item.type === 'concurrency') return t('redeem.concurrencyAddedRedeem')
  if (item.type === 'admin_concurrency') {
    return item.value >= 0 ? t('redeem.concurrencyAddedAdmin') : t('redeem.concurrencyReducedAdmin')
  }
  if (item.type === 'subscription') return t('redeem.subscriptionAssigned')
  return t('common.unknown')
}

function formatHistoryValue(item: RedeemHistoryItem): string {
  if (isBalanceType(item.type)) {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}$${item.value.toFixed(2)}`
  }
  if (isSubscriptionType(item.type)) {
    const days = item.validity_days || Math.round(item.value)
    const groupName = item.group?.name || ''
    return groupName
      ? `${days}${t('redeem.days')} - ${groupName}`
      : `${days}${t('redeem.days')}`
  }
  const sign = item.value >= 0 ? '+' : ''
  return `${sign}${item.value} ${t('redeem.requests')}`
}

function historyIconSurfaceClass(item: RedeemHistoryItem): string {
  if (isSubscriptionType(item.type)) return 'bg-info-subtle'
  return item.value >= 0 ? 'bg-success-subtle' : 'bg-danger-subtle'
}

function historyIconClass(item: RedeemHistoryItem): string {
  return item.value >= 0 ? 'text-success-foreground' : 'text-danger-foreground'
}

function historyValueClass(item: RedeemHistoryItem): string {
  if (isSubscriptionType(item.type)) return 'text-info-foreground'
  return item.value >= 0 ? 'text-success-foreground' : 'text-danger-foreground'
}

async function fetchHistory(): Promise<void> {
  loadingHistory.value = true
  historyError.value = false
  try {
    history.value = await redeemAPI.getHistory()
  } catch (error) {
    historyError.value = true
    console.error('Failed to fetch history:', error)
  } finally {
    loadingHistory.value = false
  }
}

onMounted(() => {
  void fetchHistory()
  void authAPI
    .getPublicSettings()
    .then((settings) => {
      contactInfo.value = settings.contact_info || ''
    })
    .catch((error) => {
      console.error('Failed to load contact info:', error)
    })
})
</script>
