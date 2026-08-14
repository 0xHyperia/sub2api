<template>
  <BaseDialog
    :show="show"
    :title="t('payment.admin.orderDetail')"
    width="wide"
    @close="emit('close')"
  >
    <div v-if="order" class="min-w-0 space-y-4">
      <dl class="grid grid-cols-1 gap-x-4 gap-y-3 min-[420px]:grid-cols-2">
        <div class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.orders.orderId') }}</dt>
          <dd class="break-all font-mono text-sm font-medium text-foreground">#{{ order.id }}</dd>
        </div>
        <div class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.orders.status') }}</dt>
          <dd class="mt-0.5">
            <span :class="['badge', statusBadgeClass(order.status)]">
              {{ t('payment.status.' + order.status.toLowerCase(), order.status) }}
            </span>
          </dd>
        </div>
        <div class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.orders.paymentPrincipal') }}</dt>
          <dd class="break-all text-sm font-medium text-foreground">{{ paymentAmountSymbol }}{{ paymentPrincipal.toFixed(2) }}</dd>
        </div>
        <div v-if="order.fee_rate > 0" class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.orders.fee') }} ({{ order.fee_rate }}%)</dt>
          <dd class="break-all text-sm font-medium text-foreground">{{ paymentAmountSymbol }}{{ surchargeAmount.toFixed(2) }}</dd>
        </div>
        <div class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.orders.payAmount') }}</dt>
          <dd class="break-all text-sm font-medium text-foreground">{{ paymentAmountSymbol }}{{ order.pay_amount.toFixed(2) }}</dd>
        </div>
        <div v-if="providerAmount !== order.pay_amount" class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.orders.providerAmount') }}</dt>
          <dd class="break-all text-sm font-medium text-foreground">{{ paymentAmountSymbol }}{{ providerAmount.toFixed(2) }}</dd>
        </div>
        <div v-if="entitlementPrincipal !== order.amount" class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.orders.entitlementPrincipal') }}</dt>
          <dd class="break-all text-sm font-medium text-foreground">{{ creditedAmountSymbol }}{{ entitlementPrincipal.toFixed(2) }}</dd>
        </div>
        <div v-if="bonusAmount > 0" class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.orders.bonusAmount') }}</dt>
          <dd class="break-all text-sm font-medium text-success-foreground">+{{ creditedAmountSymbol }}{{ bonusAmount.toFixed(2) }}</dd>
        </div>
        <div class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.orders.creditedAmount') }}</dt>
          <dd class="break-all text-sm font-medium text-foreground">{{ creditedAmountSymbol }}{{ order.amount.toFixed(2) }}</dd>
        </div>
        <div class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.orders.paymentMethod') }}</dt>
          <dd class="break-words text-sm text-foreground">
            {{ t('payment.methods.' + order.payment_type, order.payment_type) }}
          </dd>
        </div>
        <div class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.admin.orderType') }}</dt>
          <dd class="break-words text-sm text-foreground">
            {{ t('payment.admin.' + order.order_type + 'Order', order.order_type) }}
          </dd>
        </div>
        <div class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.orders.userId') }}</dt>
          <dd class="break-all text-sm text-foreground">#{{ order.user_id }}</dd>
        </div>
        <div class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.orders.createdAt') }}</dt>
          <dd class="break-words text-sm text-foreground">{{ formatDateTime(order.created_at) }}</dd>
        </div>
        <div class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.admin.expiresAt') }}</dt>
          <dd class="break-words text-sm text-foreground">{{ formatDateTime(order.expires_at) }}</dd>
        </div>
        <div v-if="order.paid_at" class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.admin.paidAt') }}</dt>
          <dd class="break-words text-sm text-foreground">{{ formatDateTime(order.paid_at) }}</dd>
        </div>
        <div v-if="order.completed_at" class="min-w-0">
          <dt class="text-xs text-foreground-muted">{{ t('payment.admin.completedAt') }}</dt>
          <dd class="break-words text-sm text-foreground">{{ formatDateTime(order.completed_at) }}</dd>
        </div>
      </dl>

      <div
        v-if="order.refund_amount"
        class="rounded-panel border border-danger/30 bg-danger-subtle p-3 text-danger-foreground"
      >
        <h4 class="mb-2 text-sm font-semibold">
          {{ t('payment.admin.refundInfo') }}
        </h4>
        <div class="grid grid-cols-1 gap-2 text-sm min-[420px]:grid-cols-2">
          <div class="min-w-0">
            <span>{{ t('payment.admin.refundAmount') }}:</span>
            <span class="ml-1 break-all font-medium">{{ creditedAmountSymbol }}{{ order.refund_amount.toFixed(2) }}</span>
          </div>
          <div v-if="order.refund_reason" class="min-w-0 min-[420px]:col-span-2">
            <span>{{ t('payment.admin.refundReason') }}:</span>
            <span class="ml-1 break-words">{{ order.refund_reason }}</span>
          </div>
        </div>
      </div>

      <div class="flex flex-col-reverse gap-2 border-t border-outline pt-4 min-[420px]:flex-row min-[420px]:items-center min-[420px]:justify-end">
        <button
          v-if="order.status === 'PENDING'"
          type="button"
          @click="emit('cancel', order)"
          class="btn btn-secondary btn-sm w-full text-warning-foreground min-[420px]:w-auto"
        >
          {{ t('payment.orders.cancel') }}
        </button>
        <button
          v-if="order.status === 'FAILED'"
          type="button"
          @click="emit('retry', order)"
          class="btn btn-sm btn-secondary w-full min-[420px]:w-auto"
        >
          {{ t('payment.admin.retry') }}
        </button>
        <button
          v-if="canRefund(order)"
          type="button"
          @click="emit('refund', order)"
          class="btn btn-danger btn-sm w-full min-[420px]:w-auto"
        >
          {{ t('payment.admin.refund') }}
        </button>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { PaymentOrder } from '@/types/payment'
import { statusBadgeClass, canRefund as canRefundStatus, formatOrderDateTime } from '@/components/payment/orderUtils'
import { currencySymbol } from '@/components/payment/currency'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  order: PaymentOrder | null
}>()

const creditedAmountSymbol = currencySymbol('USD')

const paymentAmountSymbol = computed(() => currencySymbol(props.order?.currency))

const paymentPrincipal = computed(() => {
  if (!props.order) return 0
  if ((props.order.payment_principal_amount || 0) > 0) return props.order.payment_principal_amount || 0
  const feeRate = Number(props.order.fee_rate) || 0
  if (feeRate <= 0) return props.order.pay_amount
  return props.order.pay_amount / (1 + feeRate / 100)
})

const surchargeAmount = computed(() => {
  if (!props.order) return 0
  if ((props.order.payment_principal_amount || 0) > 0) return props.order.surcharge_amount || 0
  const feeRate = Number(props.order.fee_rate) || 0
  if (feeRate <= 0) return 0
  return props.order.pay_amount - paymentPrincipal.value
})

const providerAmount = computed(() => props.order?.provider_amount || props.order?.pay_amount || 0)
const entitlementPrincipal = computed(() => props.order?.entitlement_principal_amount || props.order?.amount || 0)
const bonusAmount = computed(() => Math.max(0, (props.order?.amount || 0) - entitlementPrincipal.value))

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'cancel', order: PaymentOrder): void
  (e: 'retry', order: PaymentOrder): void
  (e: 'refund', order: PaymentOrder): void
}>()

function canRefund(order: PaymentOrder): boolean {
  return canRefundStatus(order.status)
}

function formatDateTime(dateStr: string): string {
  return formatOrderDateTime(dateStr)
}
</script>
