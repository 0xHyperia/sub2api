<template>
  <BaseDialog
    :show="show"
    :title="t('payment.admin.refundOrder')"
    width="normal"
    @close="emit('cancel')"
  >
    <form id="refund-form" @submit.prevent="handleSubmit" class="space-y-4">
      <!-- Refund Request Info -->
      <div
        v-if="order?.refund_requested_at || order?.refund_request_reason"
        class="rounded-panel border border-info/30 bg-info-subtle p-3 text-info-foreground"
      >
        <div class="flex items-center gap-2 text-sm font-medium">
          <Icon name="infoCircle" size="sm" aria-hidden="true" />
          {{ t('payment.admin.refundRequestInfo') }}
        </div>
        <div v-if="order?.refund_requested_at" class="mt-2 flex flex-col gap-1 text-sm min-[420px]:flex-row min-[420px]:justify-between">
          <span>{{ t('payment.admin.refundRequestedAt') }}</span>
          <span class="break-words font-medium">{{ formatDateTime(order.refund_requested_at) }}</span>
        </div>
        <div v-if="order?.refund_request_reason" class="mt-1 text-sm">
          <span>{{ t('payment.admin.refundRequestReason') }}:</span>
          <span class="ml-1 break-words font-medium">{{ order.refund_request_reason }}</span>
        </div>
      </div>

      <!-- Order Info -->
      <div class="rounded-panel border border-outline bg-surface-subtle p-3">
        <div class="flex items-start justify-between gap-3 text-sm">
          <span class="text-foreground-muted">{{ t('payment.orders.orderId') }}</span>
          <span class="break-all text-right font-mono text-foreground">#{{ order?.id }}</span>
        </div>
        <div class="mt-1 flex items-start justify-between gap-3 text-sm">
          <span class="text-foreground-muted">{{ t('payment.orders.creditedAmount') }}</span>
          <span class="break-all text-right font-medium text-foreground">{{ creditedAmountSymbol }}{{ order?.amount?.toFixed(2) }}</span>
        </div>
        <div class="mt-1 flex items-start justify-between gap-3 text-sm">
          <span class="text-foreground-muted">{{ t('payment.orders.payAmount') }}</span>
          <span class="break-all text-right font-medium text-foreground">{{ paymentAmountSymbol }}{{ order?.pay_amount?.toFixed(2) }}</span>
        </div>
        <div v-if="actuallyRefunded > 0" class="mt-1 flex items-start justify-between gap-3 text-sm">
          <span class="text-foreground-muted">{{ t('payment.admin.alreadyRefunded') }}</span>
          <span class="break-all text-right font-medium text-danger-foreground">{{ creditedAmountSymbol }}{{ actuallyRefunded.toFixed(2) }}</span>
        </div>
      </div>

      <!-- Deduct Balance -->
      <div>
        <div class="flex flex-wrap items-start gap-x-2 gap-y-1">
          <input
            id="deduct-balance"
            v-model="form.deduct_balance"
            type="checkbox"
            aria-describedby="deduct-balance-hint"
            class="mt-0.5 h-4 w-4 rounded border-outline-strong text-brand focus:ring-focus"
          />
          <label for="deduct-balance" class="text-sm text-foreground">
            {{ t('payment.admin.deductBalance') }}
          </label>
          <span id="deduct-balance-hint" class="basis-full pl-6 text-xs text-foreground-muted">{{ t('payment.admin.deductBalanceHint') }}</span>
        </div>

        <!-- User Balance Info (when deduct_balance is checked) -->
        <div v-if="form.deduct_balance && userBalance != null" class="mt-3 grid grid-cols-1 gap-3 min-[420px]:grid-cols-2">
          <div class="rounded-panel border border-outline bg-surface-subtle p-3 text-sm">
            <div class="text-foreground-muted">{{ t('payment.admin.userBalance') }}</div>
            <div class="mt-1 break-all font-semibold text-foreground">{{ creditedAmountSymbol }}{{ userBalance.toFixed(2) }}</div>
          </div>
          <div class="rounded-panel border border-outline bg-surface-subtle p-3 text-sm">
            <div class="text-foreground-muted">{{ t('payment.admin.orderAmount') }}</div>
            <div class="mt-1 break-all font-semibold text-foreground">{{ creditedAmountSymbol }}{{ order?.amount?.toFixed(2) }}</div>
          </div>
        </div>

        <!-- Insufficient balance warning -->
        <div
          v-if="form.deduct_balance && balanceInsufficient"
          class="mt-2 rounded-panel border border-warning/30 bg-warning-subtle p-3 text-sm text-warning-foreground"
          role="alert"
        >
          {{ t('payment.admin.insufficientBalance') }}
        </div>

        <!-- No deduction info -->
        <div
          v-if="!form.deduct_balance"
          class="mt-2 rounded-panel border border-info/30 bg-info-subtle p-3 text-sm text-info-foreground"
        >
          {{ t('payment.admin.noDeduction') }}
        </div>
      </div>

      <!-- Refund Amount -->
      <div>
        <label for="refund-amount" class="input-label">{{ t('payment.admin.refundAmount') }}</label>
        <div class="relative">
          <span class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-foreground-muted" aria-hidden="true">{{ creditedAmountSymbol }}</span>
          <input
            id="refund-amount"
            v-model.number="form.amount"
            type="number"
            step="0.01"
            min="0.01"
            :max="maxRefundable"
            aria-describedby="refund-amount-limit"
            class="input pl-7"
            required
          />
        </div>
        <p id="refund-amount-limit" class="mt-1 text-xs text-foreground-muted">
          {{ t('payment.admin.maxRefundable') }}: {{ creditedAmountSymbol }}{{ maxRefundable.toFixed(2) }}
        </p>
      </div>

      <!-- Reason -->
      <div>
        <label for="refund-reason" class="input-label">{{ t('payment.admin.refundReason') }}</label>
        <textarea
          id="refund-reason"
          v-model="form.reason"
          rows="3"
          class="input"
          :placeholder="t('payment.admin.refundReasonPlaceholder')"
          required
        ></textarea>
      </div>

      <!-- Warning -->
      <div
        v-if="warning"
        class="rounded-panel border border-warning/30 bg-warning-subtle p-3 text-sm text-warning-foreground"
        role="alert"
      >
        {{ warning }}
      </div>

      <!-- Force Refund -->
      <div v-if="requireForce" class="flex items-start gap-2">
        <input
          id="force-refund"
          v-model="form.force"
          type="checkbox"
          class="mt-0.5 h-4 w-4 rounded border-outline-strong text-danger focus:ring-focus"
        />
        <label for="force-refund" class="text-sm font-medium text-danger-foreground">
          {{ t('payment.admin.forceRefund') }}
        </label>
      </div>
    </form>

    <template #footer>
      <div class="flex w-full flex-col-reverse gap-2 min-[420px]:flex-row min-[420px]:justify-end">
        <button type="button" @click="emit('cancel')" class="btn btn-secondary w-full min-[420px]:w-auto">
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          form="refund-form"
          :disabled="submitting || form.amount <= 0 || (requireForce && !form.force)"
          class="btn btn-danger w-full min-[420px]:w-auto"
        >
          {{ submitting ? t('common.processing') : t('payment.admin.confirmRefund') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { PaymentOrder } from '@/types/payment'
import { formatOrderDateTime } from '@/components/payment/orderUtils'
import { currencySymbol } from '@/components/payment/currency'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  order: PaymentOrder | null
  submitting?: boolean
  userBalance?: number | null
  requireForce?: boolean
  warning?: string
}>()

const emit = defineEmits<{
  (e: 'confirm', data: { amount: number; reason: string; deduct_balance: boolean; force: boolean }): void
  (e: 'cancel'): void
}>()

const creditedAmountSymbol = currencySymbol('USD')

const paymentAmountSymbol = computed(() => currencySymbol(props.order?.currency))

const form = reactive({
  amount: 0,
  reason: '',
  deduct_balance: true,
  force: false,
})

// In REFUND_REQUESTED / REFUND_PENDING status, refund_amount is requested/pending, not actually refunded.
// Only PARTIALLY_REFUNDED / REFUNDED have real refund amounts.
const actuallyRefunded = computed(() => {
  if (!props.order) return 0
  const s = props.order.status
  if (s === 'PARTIALLY_REFUNDED' || s === 'REFUNDED') return props.order.refund_amount || 0
  return 0
})

const maxRefundable = computed(() => {
  if (!props.order) return 0
  return props.order.amount - actuallyRefunded.value
})

const balanceInsufficient = computed(() => {
  if (props.userBalance == null || !props.order) return false
  return props.userBalance < props.order.amount
})

watch(
  () => [props.show, props.order?.id] as const,
  ([isOpen]) => {
    if (isOpen && props.order) {
    // For REFUND_REQUESTED, pre-fill with the requested amount
      if (props.order.status === 'REFUND_REQUESTED' && props.order.refund_amount) {
        form.amount = props.order.refund_amount
      } else {
        form.amount = maxRefundable.value
      }
      form.reason = props.order.refund_request_reason || ''
      form.deduct_balance = true
      form.force = false
    }
  },
  { immediate: true }
)

function formatDateTime(dateStr: string): string {
  return formatOrderDateTime(dateStr)
}

function handleSubmit() {
  if (form.amount <= 0 || form.amount > maxRefundable.value) return
  if (props.requireForce && !form.force) return
  emit('confirm', { ...form })
}
</script>
