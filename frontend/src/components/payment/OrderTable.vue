<template>
  <DataTable :columns="columns" :data="orders" :loading="loading">
    <template v-if="userMobileCard" #mobile-card="{ row }">
      <article class="min-w-0">
        <header class="flex min-w-0 items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="text-[10px] font-medium uppercase text-foreground-subtle">{{ orderTypeLabel(row) }}</p>
            <h3 class="mt-0.5 truncate font-mono text-sm font-semibold text-foreground" :title="String(row.out_trade_no)">
              #{{ row.id }} · {{ row.out_trade_no }}
            </h3>
          </div>
          <OrderStatusBadge :status="row.status" class="shrink-0" />
        </header>

        <div class="mt-3 flex items-end justify-between gap-4 rounded-control bg-surface-subtle px-3 py-2.5">
          <div class="min-w-0">
            <p class="text-[10px] text-foreground-subtle">{{ t('payment.orders.payAmount') }}</p>
            <p class="mt-0.5 truncate text-xl font-semibold tabular-nums text-foreground">
              {{ paymentAmountSymbol(row) }}{{ row.pay_amount.toFixed(2) }}
            </p>
          </div>
          <div class="min-w-0 text-right">
            <p class="truncate text-xs font-medium text-foreground-muted" :title="paymentMethodLabel(row.payment_type)">
              {{ paymentMethodLabel(row.payment_type) }}
            </p>
            <time class="mt-1 block text-[10px] text-foreground-subtle">{{ formatDate(row.created_at) }}</time>
          </div>
        </div>

        <dl v-if="isMobileExpanded(row)" class="mt-3 divide-y divide-outline border-y border-outline text-xs">
          <div class="flex min-w-0 items-start justify-between gap-4 py-2">
            <dt class="shrink-0 text-foreground-subtle">{{ t('payment.orders.orderNo') }}</dt>
            <dd class="min-w-0 break-all text-right font-mono text-foreground">{{ row.out_trade_no }}</dd>
          </div>
          <div v-if="shouldShowCreditedAmount(row)" class="flex items-center justify-between gap-4 py-2">
            <dt class="text-foreground-subtle">{{ t('payment.orders.creditedAmount') }}</dt>
            <dd class="font-medium tabular-nums text-foreground">{{ creditedAmountSymbol }}{{ row.amount.toFixed(2) }}</dd>
          </div>
          <div v-if="row.fee_rate > 0" class="flex items-center justify-between gap-4 py-2">
            <dt class="text-foreground-subtle">{{ t('payment.orders.fee') }}</dt>
            <dd class="font-medium tabular-nums text-foreground">{{ row.fee_rate }}%</dd>
          </div>
        </dl>

        <div class="mt-2 flex min-w-0 items-center gap-2">
          <button
            type="button"
            class="btn btn-secondary btn-sm min-w-0 flex-1"
            :aria-expanded="isMobileExpanded(row)"
            @click.stop="toggleMobileDetails(row)"
          >
            <Icon :name="isMobileExpanded(row) ? 'chevronUp' : 'eye'" size="sm" aria-hidden="true" />
            <span>{{ isMobileExpanded(row) ? t('common.collapse') : t('common.view') }}</span>
          </button>
          <div class="order-mobile-actions min-w-0 flex-1" @click.stop>
            <slot name="actions" :row="row" />
          </div>
        </div>
      </article>
    </template>

    <template #cell-id="{ value }">
      <span class="font-mono text-sm text-foreground">#{{ value }}</span>
    </template>
    <template #cell-out_trade_no="{ value }">
      <span class="block max-w-56 truncate text-sm text-foreground" :title="String(value)">{{ value }}</span>
    </template>
    <template v-if="showUser" #cell-user_email="{ value, row }">
      <div class="text-sm">
        <span class="text-foreground">{{ value || row.user_name || '#' + row.user_id }}</span>
        <span v-if="row.user_notes" class="ml-1 text-xs text-foreground-subtle">({{ row.user_notes }})</span>
      </div>
    </template>
    <template #cell-pay_amount="{ value, row }">
      <div class="text-sm">
        <span class="font-medium tabular-nums text-foreground">{{ paymentAmountSymbol(row) }}{{ value.toFixed(2) }}</span>
        <span v-if="row.fee_rate > 0" class="ml-1 text-xs text-foreground-subtle" :title="t('payment.orders.fee') + ': ' + row.fee_rate + '%'">
          ({{ t('payment.orders.fee') }} {{ row.fee_rate }}%)
        </span>
        <div v-if="shouldShowCreditedAmount(row)" class="text-xs text-foreground-subtle">
          {{ t('payment.orders.creditedAmount') }}: {{ creditedAmountSymbol }}{{ row.amount.toFixed(2) }}
        </div>
      </div>
    </template>
    <template #cell-payment_type="{ value }">
      <span class="text-sm text-foreground-muted">{{ t('payment.methods.' + value, value) }}</span>
    </template>
    <template #cell-status="{ value }">
      <OrderStatusBadge :status="value" />
    </template>
    <template #cell-created_at="{ value }">
      <span class="whitespace-nowrap text-xs text-foreground-subtle">{{ formatDate(value) }}</span>
    </template>
    <template #cell-actions="{ row }">
      <slot name="actions" :row="row" />
    </template>
    <template #empty>
      <EmptyState :title="t('payment.orders.empty')" />
    </template>
  </DataTable>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PaymentOrder } from '@/types/payment'
import type { Column } from '@/components/common/types'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import Icon from '@/components/icons/Icon.vue'
import { currencySymbol } from '@/components/payment/currency'

const { t } = useI18n()

const props = defineProps<{
  orders: PaymentOrder[]
  loading: boolean
  showUser?: boolean
  /** 用户订单页手机摘要卡；管理端默认保持通用表格。 */
  userMobileCard?: boolean
}>()

function formatDate(dateStr: string) { return new Date(dateStr).toLocaleString() }

const creditedAmountSymbol = currencySymbol('USD')

function paymentAmountSymbol(order: PaymentOrder): string {
  return currencySymbol(order.currency)
}

function shouldShowCreditedAmount(order: PaymentOrder): boolean {
  return order.order_type === 'balance' || order.amount !== order.pay_amount
}

function paymentMethodLabel(paymentType: string): string {
  return t('payment.methods.' + paymentType, paymentType)
}

function orderTypeLabel(order: PaymentOrder): string {
  return order.order_type === 'subscription' ? t('payment.tabSubscribe') : t('payment.tabTopUp')
}

const mobileExpandedOrders = ref(new Set<number>())
function isMobileExpanded(order: PaymentOrder): boolean {
  return mobileExpandedOrders.value.has(order.id)
}
function toggleMobileDetails(order: PaymentOrder) {
  const next = new Set(mobileExpandedOrders.value)
  if (next.has(order.id)) next.delete(order.id)
  else next.add(order.id)
  mobileExpandedOrders.value = next
}

const columns = computed((): Column[] => {
  const cols: Column[] = [
    { key: 'id', label: t('payment.orders.orderId') },
    { key: 'out_trade_no', label: t('payment.orders.orderNo') },
  ]
  if (props.showUser) {
    cols.push({ key: 'user_email', label: t('payment.admin.colUser') })
  }
  cols.push(
    { key: 'pay_amount', label: t('payment.orders.payAmount') },
    { key: 'payment_type', label: t('payment.orders.paymentMethod') },
    { key: 'status', label: t('payment.orders.status') },
    { key: 'created_at', label: t('payment.orders.createdAt') },
    { key: 'actions', label: t('common.actions') },
  )
  return cols
})
</script>

<style scoped>
.order-mobile-actions:empty {
  display: none;
}
</style>
