<template>
  <div class="space-y-4">
    <div class="card p-4">
      <div class="flex flex-wrap items-center gap-3">
        <div class="flex-1 sm:max-w-64">
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('payment.admin.searchOrders')"
            class="input"
            @input="handleSearch"
          />
        </div>
        <Select
          v-model="filters.status"
          :options="statusFilterOptions"
          class="w-36"
          @change="emitFiltersChanged"
        />
        <Select
          v-model="filters.payment_type"
          :options="paymentTypeFilterOptions"
          class="w-40"
          @change="emitFiltersChanged"
        />
        <Select
          v-model="filters.order_type"
          :options="orderTypeFilterOptions"
          class="w-36"
          @change="emitFiltersChanged"
        />
        <div class="flex flex-1 flex-wrap items-center justify-end gap-2">
          <button
            @click="emit('refresh')"
            :disabled="loading"
            class="btn btn-secondary"
            :title="t('common.refresh')"
          >
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </div>
    </div>

    <DataTable :columns="columns" :data="orders" :loading="loading">
      <template #mobile-card="{ row }">
        <article class="space-y-3" :data-test="`admin-order-mobile-card-${row.id}`">
          <header class="flex min-w-0 items-start justify-between gap-3">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-1.5">
                <span class="badge badge-gray">
                  {{ t('payment.admin.' + row.order_type + 'Order', row.order_type) }}
                </span>
                <span :class="['badge', statusBadgeClass(row.status)]">
                  {{ t('payment.status.' + row.status.toLowerCase(), row.status) }}
                </span>
              </div>
              <p class="mt-1.5 font-mono text-sm font-semibold text-foreground">#{{ row.id }}</p>
            </div>
            <div class="flex-none text-right">
              <p class="text-xs text-foreground-subtle">{{ t('payment.orders.userId') }}</p>
              <p class="mt-0.5 font-mono text-sm font-medium text-foreground">#{{ row.user_id }}</p>
            </div>
          </header>

          <dl class="grid grid-cols-2 overflow-hidden rounded-panel border border-outline bg-surface-subtle">
            <div class="min-w-0 px-3 py-2.5">
              <dt class="text-[11px] text-foreground-subtle">{{ t('payment.orders.payAmount') }}</dt>
              <dd class="mt-1 text-base font-semibold tabular-nums text-foreground">
                {{ paymentAmountSymbol(row) }}{{ row.pay_amount.toFixed(2) }}
              </dd>
              <p v-if="shouldShowCreditedAmount(row)" class="mt-0.5 truncate text-[11px] text-foreground-subtle">
                {{ t('payment.orders.creditedAmount') }}: {{ creditedAmountSymbol }}{{ row.amount.toFixed(2) }}
              </p>
            </div>
            <div class="min-w-0 border-l border-outline px-3 py-2.5">
              <dt class="text-[11px] text-foreground-subtle">{{ t('payment.orders.paymentMethod') }}</dt>
              <dd class="mt-1 truncate text-sm font-semibold text-foreground">
                {{ t('payment.methods.' + row.payment_type, row.payment_type) }}
              </dd>
              <p v-if="row.fee_rate > 0" class="mt-0.5 text-[11px] text-foreground-subtle">
                {{ t('payment.orders.fee') }} {{ row.fee_rate }}%
              </p>
            </div>
          </dl>

          <div class="flex items-center justify-between gap-3 text-xs">
            <span class="text-foreground-subtle">{{ t('payment.orders.createdAt') }}</span>
            <time class="text-right text-foreground-muted" :datetime="row.created_at">{{ formatDateTime(row.created_at) }}</time>
          </div>

          <footer class="flex flex-wrap items-center gap-2 border-t border-outline pt-3">
            <button type="button" class="btn btn-secondary min-w-0 flex-1" :data-test="`admin-order-mobile-detail-${row.id}`" @click="emit('detail', row)">
              <Icon name="eye" size="sm" />
              {{ t('common.view') }}
            </button>
            <button v-if="row.status === 'PENDING'" type="button" class="btn btn-secondary text-warning-foreground" :data-test="`admin-order-mobile-cancel-${row.id}`" @click="emit('cancel', row)">
              <Icon name="x" size="sm" />
              {{ t('payment.orders.cancel') }}
            </button>
            <button v-if="row.status === 'FAILED'" type="button" class="btn btn-secondary" :data-test="`admin-order-mobile-retry-${row.id}`" @click="emit('retry', row)">
              <Icon name="refresh" size="sm" />
              {{ t('payment.admin.retry') }}
            </button>
            <button v-if="canRefundRow(row)" type="button" class="btn btn-secondary text-danger-foreground" :data-test="`admin-order-mobile-refund-${row.id}`" @click="emit('refund', row)">
              <Icon name="dollar" size="sm" />
              {{ t('payment.admin.refund') }}
            </button>
          </footer>
        </article>
      </template>
      <template #cell-id="{ value }">
        <span class="font-mono text-sm">#{{ value }}</span>
      </template>

      <template #cell-user_id="{ value }">
        <span class="text-sm text-foreground-subtle">#{{ value }}</span>
      </template>

      <template #cell-pay_amount="{ value, row }">
        <div class="text-sm">
          <span class="font-medium text-foreground">{{ paymentAmountSymbol(row) }}{{ value.toFixed(2) }}</span>
          <span v-if="row.fee_rate > 0" class="ml-1 text-xs text-foreground-subtle" :title="t('payment.orders.fee') + ': ' + row.fee_rate + '%'">
            ({{ row.fee_rate }}%)
          </span>
          <div v-if="shouldShowCreditedAmount(row)" class="text-xs text-foreground-subtle">
            {{ t('payment.orders.creditedAmount') }}: {{ creditedAmountSymbol }}{{ row.amount.toFixed(2) }}
          </div>
        </div>
      </template>

      <template #cell-payment_type="{ value }">
        <span class="text-sm text-foreground-muted">
          {{ t('payment.methods.' + value, value) }}
        </span>
      </template>

      <template #cell-status="{ value }">
        <span :class="['badge', statusBadgeClass(value)]">
          {{ t('payment.status.' + value.toLowerCase(), value) }}
        </span>
      </template>

      <template #cell-order_type="{ value }">
        <span class="text-sm text-foreground-muted">
          {{ t('payment.admin.' + value + 'Order', value) }}
        </span>
      </template>

      <template #cell-created_at="{ value }">
        <span class="text-xs text-foreground-subtle">{{ formatDateTime(value) }}</span>
      </template>

      <template #cell-actions="{ row }">
        <div class="flex items-center gap-2">
          <button
            @click="emit('detail', row)"
            class="flex flex-col items-center gap-0.5 rounded-panel p-1.5 text-foreground-subtle transition-colors hover:bg-surface/20 hover:text-foreground-muted"
          >
            <Icon name="eye" size="sm" />
            <span class="text-xs">{{ t('common.view') }}</span>
          </button>
          <button
            v-if="row.status === 'PENDING'"
            @click="emit('cancel', row)"
            class="flex flex-col items-center gap-0.5 rounded-panel p-1.5 text-foreground-subtle transition-colors hover:bg-warning-subtle/20 hover:text-warning-foreground"
          >
            <Icon name="x" size="sm" />
            <span class="text-xs">{{ t('payment.orders.cancel') }}</span>
          </button>
          <button
            v-if="row.status === 'FAILED'"
            @click="emit('retry', row)"
            class="flex flex-col items-center gap-0.5 rounded-panel p-1.5 text-foreground-subtle transition-colors hover:bg-info-subtle/20 hover:text-info-foreground"
          >
            <Icon name="refresh" size="sm" />
            <span class="text-xs">{{ t('payment.admin.retry') }}</span>
          </button>
          <button
            v-if="canRefundRow(row)"
            @click="emit('refund', row)"
            class="flex flex-col items-center gap-0.5 rounded-panel p-1.5 text-foreground-subtle transition-colors hover:bg-danger-subtle hover:text-danger-foreground"
          >
            <Icon name="dollar" size="sm" />
            <span class="text-xs">{{ t('payment.admin.refund') }}</span>
          </button>
        </div>
      </template>
    </DataTable>

    <Pagination
      v-if="total > 0"
      :page="page"
      :total="total"
      :page-size="pageSize"
      @update:page="emit('update:page', $event)"
      @update:pageSize="emit('update:pageSize', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PaymentOrder } from '@/types/payment'
import type { Column } from '@/components/common/types'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { statusBadgeClass, canRefund, formatOrderDateTime } from '@/components/payment/orderUtils'
import { currencySymbol } from '@/components/payment/currency'

const { t } = useI18n()

defineProps<{
  orders: PaymentOrder[]
  loading: boolean
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  (e: 'detail', order: PaymentOrder): void
  (e: 'cancel', order: PaymentOrder): void
  (e: 'retry', order: PaymentOrder): void
  (e: 'refund', order: PaymentOrder): void
  (e: 'refresh'): void
  (e: 'update:page', page: number): void
  (e: 'update:pageSize', size: number): void
  (e: 'filter', filters: { keyword?: string; status?: string; payment_type?: string; order_type?: string }): void
}>()

const searchQuery = ref('')
const filters = reactive({ status: '', payment_type: '', order_type: '' })
const creditedAmountSymbol = currencySymbol('USD')

function paymentAmountSymbol(order: PaymentOrder): string {
  return currencySymbol(order.currency)
}

function shouldShowCreditedAmount(order: PaymentOrder): boolean {
  return order.order_type === 'balance' || order.amount !== order.pay_amount
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null
function handleSearch() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => emitFiltersChanged(), 300)
}

function emitFiltersChanged() {
  emit('filter', {
    keyword: searchQuery.value || undefined,
    status: filters.status || undefined,
    payment_type: filters.payment_type || undefined,
    order_type: filters.order_type || undefined,
  })
}

const columns = computed<Column[]>(() => [
  { key: 'id', label: t('payment.orders.orderId') },
  { key: 'user_id', label: t('payment.orders.userId') },
  { key: 'pay_amount', label: t('payment.orders.payAmount') },
  { key: 'payment_type', label: t('payment.orders.paymentMethod') },
  { key: 'status', label: t('payment.orders.status') },
  { key: 'order_type', label: t('payment.orders.orderType') },
  { key: 'created_at', label: t('payment.orders.createdAt') },
  { key: 'actions', label: t('payment.orders.actions') },
])

const statusFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allStatuses') },
  { value: 'PENDING', label: t('payment.status.pending') },
  { value: 'PAID', label: t('payment.status.paid') },
  { value: 'COMPLETED', label: t('payment.status.completed') },
  { value: 'EXPIRED', label: t('payment.status.expired') },
  { value: 'CANCELLED', label: t('payment.status.cancelled') },
  { value: 'FAILED', label: t('payment.status.failed') },
  { value: 'REFUNDED', label: t('payment.status.refunded') },
  { value: 'REFUND_REQUESTED', label: t('payment.status.refund_requested') },
  { value: 'REFUND_PENDING', label: t('payment.status.refund_pending') },
  { value: 'REFUND_FAILED', label: t('payment.status.refund_failed') },
])

const paymentTypeFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allPaymentTypes') },
  { value: 'alipay', label: t('payment.methods.alipay') },
  { value: 'wxpay', label: t('payment.methods.wxpay') },
  { value: 'stripe', label: t('payment.methods.stripe') },
  { value: 'airwallex', label: t('payment.methods.airwallex') },
])

const orderTypeFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allOrderTypes') },
  { value: 'balance', label: t('payment.admin.balanceOrder') },
  { value: 'subscription', label: t('payment.admin.subscriptionOrder') },
])

function canRefundRow(order: PaymentOrder): boolean {
  return canRefund(order.status)
}

function formatDateTime(dateStr: string): string {
  return formatOrderDateTime(dateStr)
}
</script>
