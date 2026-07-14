<template>
  <AppLayout>
    <div class="payment-dashboard space-y-4">
      <div class="payment-toolbar flex flex-wrap items-center justify-end gap-2">
        <div class="payment-range" role="group" :aria-label="t('admin.dashboard.timeRange')">
          <button
            v-for="d in DAYS_OPTIONS"
            :key="d"
            type="button"
            :aria-pressed="days === d"
            @click="days = d"
          >
            {{ d }}{{ t('payment.admin.daySuffix') }}
          </button>
        </div>
        <button
          type="button"
          @click="loadDashboard"
          :disabled="loading"
          class="btn btn-secondary px-2.5"
          :title="t('common.refresh')"
          :aria-label="t('common.refresh')"
        >
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>

      <div
        v-if="loading && !stats"
        class="payment-panel flex min-h-48 items-center justify-center"
        role="status"
        :aria-label="t('common.loading')"
      >
        <LoadingSpinner />
      </div>
      <div v-else-if="loadError && !stats" class="payment-load-error" role="alert">
        <Icon name="exclamationTriangle" size="md" aria-hidden="true" />
        <p>{{ t('payment.admin.dashboardLoadFailed') }}</p>
        <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="loadDashboard">
          <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
          {{ t('common.retry') }}
        </button>
      </div>
      <template v-else-if="stats">
        <div v-if="loadError" class="payment-load-error" role="alert">
          <Icon name="exclamationTriangle" size="md" aria-hidden="true" />
          <p>{{ t('payment.admin.dashboardLoadFailed') }}</p>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="loadDashboard">
            <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
            {{ t('common.retry') }}
          </button>
        </div>
        <section class="payment-kpis" :aria-label="t('payment.admin.orders')">
          <OrderStatsCards :stats="stats" />
        </section>
        <section class="payment-chart">
          <DailyRevenueChart :data="stats.daily_series || []" :loading="loading" />
        </section>
        <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <section class="payment-panel">
            <h2>{{ t('payment.admin.paymentDistribution') }}</h2>
            <div v-if="!stats.payment_methods?.length" class="payment-empty">{{ t('payment.admin.noData') }}</div>
            <div v-else class="payment-list">
              <div v-for="method in stats.payment_methods" :key="method.type">
                <div class="flex items-center gap-2">
                  <span :class="['payment-method-swatch', methodColor(method.type)]" aria-hidden="true"></span>
                  <span>{{ t('payment.methods.' + method.type, method.type) }}</span>
                </div>
                <div class="text-right tabular-nums">
                  <strong>&yen;{{ method.amount.toFixed(2) }}</strong>
                  <small>{{ method.count }} {{ t('payment.admin.orders') }}</small>
                </div>
              </div>
            </div>
          </section>
          <section class="payment-panel">
            <h2>{{ t('payment.admin.topUsers') }}</h2>
            <div v-if="!stats.top_users?.length" class="payment-empty">{{ t('payment.admin.noData') }}</div>
            <ol v-else class="payment-list">
              <li v-for="(user, idx) in stats.top_users" :key="user.user_id">
                <div class="flex min-w-0 items-center gap-3">
                  <span class="payment-rank">{{ idx + 1 }}</span>
                  <span class="truncate">{{ user.email }}</span>
                </div>
                <strong class="shrink-0 tabular-nums">&yen;{{ user.amount.toFixed(2) }}</strong>
              </li>
            </ol>
          </section>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { DashboardStats } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderStatsCards from '@/components/admin/payment/OrderStatsCards.vue'
import DailyRevenueChart from '@/components/admin/payment/DailyRevenueChart.vue'

const { t } = useI18n()
const appStore = useAppStore()

const DAYS_OPTIONS = [7, 30, 90] as const
const days = ref<number>(30)
const loading = ref(false)
const stats = ref<DashboardStats | null>(null)
const loadError = ref(false)
let loadSequence = 0

function methodColor(type: string): string {
  const c: Record<string, string> = {
    alipay: 'bg-info', wxpay: 'bg-success',
    alipay_direct: 'bg-info', wxpay_direct: 'bg-success',
    stripe: 'bg-brand',
  }
  return c[type] || 'bg-outline-strong'
}

async function loadDashboard() {
  const currentSequence = ++loadSequence
  loading.value = true
  loadError.value = false
  try {
    const res = await adminPaymentAPI.getDashboard(days.value)
    if (currentSequence !== loadSequence) return
    stats.value = res.data
  } catch (err: unknown) {
    if (currentSequence !== loadSequence) return
    loadError.value = true
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    if (currentSequence === loadSequence) loading.value = false
  }
}

watch(days, () => loadDashboard())
onMounted(() => loadDashboard())
</script>

<style scoped>
.payment-toolbar,
.payment-panel {
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 8px;
  background: var(--ui-surface, #fff);
  box-shadow: var(--ui-shadow-xs, 0 1px 2px rgba(15, 23, 42, 0.04));
}

.payment-load-error {
  display: flex;
  min-height: 72px;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border: 1px solid color-mix(in srgb, var(--ui-danger, #dc2626) 30%, transparent);
  border-radius: 8px;
  color: var(--ui-danger-text, #b42318);
  background: var(--ui-danger-subtle, #fef3f2);
}

.payment-load-error p {
  min-width: 0;
  flex: 1;
  font-size: 13px;
  font-weight: 600;
}

.payment-toolbar {
  padding: 10px 12px;
}

.payment-range {
  display: inline-grid;
  grid-auto-flow: column;
  overflow: hidden;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 6px;
}

.payment-range button {
  min-width: 52px;
  min-height: 36px;
  padding: 0 12px;
  border-right: 1px solid var(--ui-border, #dbe3ee);
  color: var(--ui-text-muted, #667085);
  background: var(--ui-surface, #fff);
  font-size: 12px;
  font-weight: 600;
}

.payment-range button:last-child {
  border-right: 0;
}

.payment-range button:hover,
.payment-range button[aria-pressed='true'] {
  color: var(--ui-text, #0f172a);
  background: var(--ui-surface-subtle, #f4f7fb);
}

.payment-range button[aria-pressed='true'] {
  box-shadow: inset 0 -2px 0 var(--ui-text, #0f172a);
}

.payment-range button:focus-visible {
  position: relative;
  z-index: 1;
  outline: 2px solid var(--ui-focus, #475569);
  outline-offset: -2px;
}

.payment-panel {
  min-width: 0;
  padding: 16px;
}

.payment-panel h2 {
  margin-bottom: 12px;
  color: var(--ui-text, #0f172a);
  font-size: 13px;
  font-weight: 600;
}

.payment-empty {
  display: flex;
  min-height: 128px;
  align-items: center;
  justify-content: center;
  color: var(--ui-text-muted, #667085);
  font-size: 14px;
}

.payment-list > div,
.payment-list > li {
  display: flex;
  min-height: 48px;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  border-top: 1px solid var(--ui-border, #dbe3ee);
  color: var(--ui-text-muted, #667085);
  font-size: 13px;
}

.payment-list strong {
  color: var(--ui-text, #0f172a);
  font-size: 13px;
  font-weight: 600;
}

.payment-list small {
  display: block;
  margin-top: 2px;
  color: var(--ui-text-subtle, #98a2b3);
  font-size: 11px;
}

.payment-method-swatch {
  width: 10px;
  height: 10px;
  flex: 0 0 auto;
  border-radius: 3px;
}

.payment-rank {
  display: inline-flex;
  width: 24px;
  height: 24px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 6px;
  color: var(--ui-text-muted, #667085);
  background: var(--ui-surface-subtle, #f4f7fb);
  font-size: 11px;
  font-weight: 600;
}

.payment-kpis :deep(.card),
.payment-chart :deep(.card) {
  border-color: var(--ui-border, #dbe3ee);
  border-radius: 8px;
  background: var(--ui-surface, #fff);
  box-shadow: var(--ui-shadow-xs, 0 1px 2px rgba(15, 23, 42, 0.04));
}

.payment-kpis :deep(.card > div > div:first-child) {
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 6px;
  color: var(--ui-text-muted, #667085) !important;
  background: var(--ui-surface-subtle, #f4f7fb) !important;
}

.payment-kpis :deep(.card > div > div:first-child svg) {
  color: currentColor !important;
}

.payment-kpis :deep(.text-xl) {
  font-size: 20px;
  line-height: 28px;
}

@media (max-width: 639px) {
  .payment-toolbar {
    padding: 10px;
  }

  .payment-range {
    flex: 1 1 auto;
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 479px) {
  .payment-load-error {
    align-items: flex-start;
    flex-direction: column;
  }

  .payment-kpis :deep(.grid) {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
