<template>
  <main class="flex min-h-screen items-center bg-canvas px-4 py-8 text-foreground sm:px-6">
    <section class="mx-auto w-full max-w-lg" aria-live="polite">
      <div v-if="loading" class="flex min-h-64 items-center justify-center" role="status" :aria-label="t('common.loading')">
        <Icon name="refresh" size="lg" class="animate-spin text-foreground-subtle" />
      </div>

      <template v-else>
        <header class="text-center">
          <span
            v-if="isSuccess"
            class="mx-auto inline-flex h-16 w-16 items-center justify-center rounded-panel bg-success-subtle text-success-foreground"
          >
            <Icon name="check" size="xl" />
          </span>
          <span
            v-else-if="isPending"
            class="mx-auto inline-flex h-16 w-16 items-center justify-center rounded-panel bg-warning-subtle text-warning-foreground"
          >
            <Icon name="refresh" size="xl" class="animate-spin" />
          </span>
          <span
            v-else
            class="mx-auto inline-flex h-16 w-16 items-center justify-center rounded-panel bg-danger-subtle text-danger-foreground"
          >
            <Icon name="x" size="xl" />
          </span>
          <h1 class="mt-4 text-2xl font-semibold text-foreground">{{ statusTitle }}</h1>
          <p v-if="isPending" class="mx-auto mt-2 max-w-md text-sm text-foreground-muted">
            {{ t('payment.result.processingHint') }}
          </p>
        </header>

        <div
          v-if="isPending && statusRefreshExhausted"
          class="mt-5 rounded-panel border border-warning/30 bg-warning-subtle p-4 text-warning-foreground"
          role="status"
        >
          <p class="text-sm font-medium">{{ t('payment.result.confirmationDelayed') }}</p>
          <p v-if="lastCheckedLabel" class="mt-1 text-xs opacity-80">
            {{ t('payment.result.lastChecked', { time: lastCheckedLabel }) }}
          </p>
          <button
            type="button"
            class="btn btn-secondary btn-sm mt-3"
            :disabled="retrying"
            @click="retryStatus"
          >
            <Icon name="refresh" size="sm" :class="{ 'animate-spin': retrying }" />
            {{ t('common.retry') }}
          </button>
        </div>

        <dl
          v-if="order"
          class="mt-6 divide-y divide-outline overflow-hidden rounded-panel border border-outline bg-surface px-4 shadow-card"
        >
          <div v-if="hasOrderId(order)" class="flex items-start justify-between gap-4 py-3 text-sm">
            <dt class="text-foreground-muted">{{ t('payment.orders.orderId') }}</dt>
            <dd class="font-medium text-foreground">#{{ order.id }}</dd>
          </div>
          <div v-if="order.out_trade_no" class="flex min-w-0 items-start justify-between gap-4 py-3 text-sm">
            <dt class="flex-shrink-0 text-foreground-muted">{{ t('payment.orders.orderNo') }}</dt>
            <dd class="min-w-0 break-all text-right font-medium text-foreground">{{ order.out_trade_no }}</dd>
          </div>
          <div v-if="hasAmountFields(order)" class="flex items-start justify-between gap-4 py-3 text-sm">
            <dt class="text-foreground-muted">{{ t('payment.orders.baseAmount') }}</dt>
            <dd class="font-medium tabular-nums text-foreground">{{ formatGatewayAmount(baseAmount) }}</dd>
          </div>
          <div v-if="hasAmountFields(order) && order.fee_rate > 0" class="flex items-start justify-between gap-4 py-3 text-sm">
            <dt class="text-foreground-muted">{{ t('payment.orders.fee') }} ({{ order.fee_rate }}%)</dt>
            <dd class="font-medium tabular-nums text-foreground">{{ formatGatewayAmount(feeAmount) }}</dd>
          </div>
          <div v-if="hasAmountFields(order)" class="flex items-start justify-between gap-4 py-3 text-sm">
            <dt class="text-foreground-muted">{{ t('payment.orders.payAmount') }}</dt>
            <dd class="font-semibold tabular-nums text-foreground">{{ formatGatewayAmount(order.pay_amount) }}</dd>
          </div>
          <div v-if="hasAmountFields(order) && order.amount !== order.pay_amount" class="flex items-start justify-between gap-4 py-3 text-sm">
            <dt class="text-foreground-muted">{{ t('payment.orders.creditedAmount') }}</dt>
            <dd class="font-medium tabular-nums text-foreground">{{ order.order_type === 'balance' ? '$' + order.amount.toFixed(2) : formatGatewayAmount(order.amount) }}</dd>
          </div>
          <div v-if="hasPaymentType(order)" class="flex items-start justify-between gap-4 py-3 text-sm">
            <dt class="text-foreground-muted">{{ t('payment.orders.paymentMethod') }}</dt>
            <dd class="text-right font-medium text-foreground">{{ t(paymentMethodI18nKey(order.payment_type), normalizedOrderPaymentType(order.payment_type)) }}</dd>
          </div>
          <div class="flex items-center justify-between gap-4 py-3 text-sm">
            <dt class="text-foreground-muted">{{ t('payment.orders.status') }}</dt>
            <dd><OrderStatusBadge :status="displayOrderStatus(order.status)" /></dd>
          </div>
        </dl>

        <dl
          v-else-if="returnInfo"
          class="mt-6 divide-y divide-outline overflow-hidden rounded-panel border border-outline bg-surface px-4 shadow-card"
        >
          <div v-if="returnInfo.outTradeNo" class="flex min-w-0 items-start justify-between gap-4 py-3 text-sm">
            <dt class="flex-shrink-0 text-foreground-muted">{{ t('payment.orders.orderId') }}</dt>
            <dd class="min-w-0 break-all text-right font-medium text-foreground">{{ returnInfo.outTradeNo }}</dd>
          </div>
          <div v-if="returnInfo.money" class="flex items-start justify-between gap-4 py-3 text-sm">
            <dt class="text-foreground-muted">{{ t('payment.orders.payAmount') }}</dt>
            <dd class="font-medium tabular-nums text-foreground">{{ formatGatewayAmount(Number(returnInfo.money) || 0) }}</dd>
          </div>
          <div v-if="returnInfo.type" class="flex items-start justify-between gap-4 py-3 text-sm">
            <dt class="text-foreground-muted">{{ t('payment.orders.paymentMethod') }}</dt>
            <dd class="text-right font-medium text-foreground">{{ t(paymentMethodI18nKey(returnInfo.type), normalizedOrderPaymentType(returnInfo.type)) }}</dd>
          </div>
        </dl>

        <div class="mt-6 flex flex-col-reverse gap-2 sm:flex-row">
          <button type="button" class="btn btn-secondary flex-1" @click="router.push('/purchase')">
            {{ t('payment.result.backToRecharge') }}
          </button>
          <button type="button" class="btn btn-primary flex-1" @click="router.push('/orders')">
            {{ t('payment.result.viewOrders') }}
          </button>
        </div>
      </template>
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref, computed, onBeforeUnmount, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  clearPaymentRecoverySnapshot,
  readPaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import { usePaymentStore } from '@/stores/payment'
import { paymentAPI } from '@/api/payment'
import type { PublicOrderVerifyResult } from '@/api/payment'
import type { OrderStatus, PaymentOrder } from '@/types/payment'
import { formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'
import { normalizePaymentMethodForDisplay, paymentMethodI18nKey } from './paymentUx'
import Icon from '@/components/icons/Icon.vue'

const i18n = useI18n()
const { t } = i18n
const route = useRoute()
const router = useRouter()
const paymentStore = usePaymentStore()

type ResolvedOrder = PaymentOrder | PublicOrderVerifyResult

const order = ref<ResolvedOrder | null>(null)
const loading = ref(true)
const currency = ref('CNY')
const retrying = ref(false)
const statusRefreshExhausted = ref(false)
const lastCheckedAt = ref<Date | null>(null)

interface ReturnInfo {
  outTradeNo: string
  money: string
  type: string
  tradeStatus: string
}
const returnInfo = ref<ReturnInfo | null>(null)

const SUCCESS_STATUSES = new Set(['COMPLETED', 'PAID', 'RECHARGING'])
const PENDING_STATUSES = new Set(['PENDING', 'CREATED', 'WAITING', 'PROCESSING'])
const STATUS_REFRESH_INTERVAL_MS = 2000
const STATUS_REFRESH_MAX_ATTEMPTS = 15

let statusRefreshTimer: ReturnType<typeof setTimeout> | null = null
let refreshOrderAction: (() => Promise<ResolvedOrder | null>) | null = null
const refreshAttempts = ref(0)

/** 充值金额 = pay_amount / (1 + fee_rate/100)，fee_rate=0 时等于 pay_amount */
const baseAmount = computed(() => {
  if (!hasAmountFields(order.value)) return 0
  const feeRate = Number(order.value.fee_rate) || 0
  if (feeRate <= 0) return order.value.pay_amount ?? 0
  return Math.round((order.value.pay_amount / (1 + feeRate / 100)) * 100) / 100
})

/** 手续费 = pay_amount - baseAmount */
const feeAmount = computed(() => {
  if (!hasAmountFields(order.value)) return 0
  const feeRate = Number(order.value.fee_rate) || 0
  if (feeRate <= 0) return 0
  return Math.round((order.value.pay_amount - baseAmount.value) * 100) / 100
})

const localeCode = computed(() => {
  const raw = i18n.locale as unknown
  if (typeof raw === 'string') return raw
  if (raw && typeof raw === 'object' && 'value' in raw) {
    return String((raw as { value?: string }).value || '')
  }
  return undefined
})

const isSuccess = computed(() => {
  return isSuccessStatus(order.value?.status)
})

const isPending = computed(() => {
  return isPendingStatus(order.value?.status)
})

const statusTitle = computed(() => {
  if (isSuccess.value) {
    return t('payment.result.success')
  }
  if (isPending.value) {
    return t('payment.result.processing')
  }
  return t('payment.result.failed')
})

const lastCheckedLabel = computed(() => {
  if (!lastCheckedAt.value) return ''
  return new Intl.DateTimeFormat(localeCode.value || undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(lastCheckedAt.value)
})

function normalizedOrderPaymentType(paymentType: string): string {
  return normalizePaymentMethodForDisplay(paymentType || '') || paymentType || ''
}

function formatGatewayAmount(value: number): string {
  return formatPaymentAmount(value, currency.value, localeCode.value)
}

function setResolvedOrder(nextOrder: ResolvedOrder | null): void {
  order.value = nextOrder
  lastCheckedAt.value = new Date()
  if (nextOrder && 'currency' in nextOrder && nextOrder.currency) {
    currency.value = normalizePaymentCurrency(nextOrder.currency)
  }
}

function hasOrderId(nextOrder: ResolvedOrder | null): nextOrder is PaymentOrder {
  return !!nextOrder && 'id' in nextOrder && typeof nextOrder.id === 'number'
}

function hasAmountFields(nextOrder: ResolvedOrder | null): nextOrder is PaymentOrder {
  return !!nextOrder && 'pay_amount' in nextOrder && typeof nextOrder.pay_amount === 'number' && 'amount' in nextOrder && typeof nextOrder.amount === 'number'
}

function hasPaymentType(nextOrder: ResolvedOrder | null): nextOrder is PaymentOrder {
  return !!nextOrder && 'payment_type' in nextOrder && typeof nextOrder.payment_type === 'string' && nextOrder.payment_type.trim() !== ''
}

function normalizeOrderStatus(status: string | null | undefined): string {
  return String(status || '').trim().toUpperCase()
}

function displayOrderStatus(status: string): OrderStatus {
  return normalizeOrderStatus(status) as OrderStatus
}

function isSuccessStatus(status: string | null | undefined): boolean {
  return SUCCESS_STATUSES.has(normalizeOrderStatus(status))
}

function isPendingStatus(status: string | null | undefined): boolean {
  return PENDING_STATUSES.has(normalizeOrderStatus(status))
}

function readRouteQueryString(key: string): string {
  const value = route.query[key]
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

function restoreRecoverySnapshot(context: {
  resumeToken: string
  routeOrderId: number
  routeOutTradeNo: string
}) {
  if (typeof window === 'undefined') {
    return null
  }

  const rawSnapshot = window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)
  if (!rawSnapshot) {
    return null
  }

  if (context.resumeToken) {
    return readPaymentRecoverySnapshot(rawSnapshot, {
      resumeToken: context.resumeToken,
    })
  }

  if (!context.routeOrderId && !context.routeOutTradeNo) {
    return null
  }

  const restored = readPaymentRecoverySnapshot(rawSnapshot)
  if (!restored) {
    return null
  }

  if (context.routeOrderId > 0 && restored.orderId !== context.routeOrderId) {
    return null
  }

  if (context.routeOutTradeNo && restored.outTradeNo !== context.routeOutTradeNo) {
    return null
  }

  return restored
}

async function resolveOrderFromResumeToken(resumeToken: string): Promise<ResolvedOrder | null> {
  try {
    const result = await paymentAPI.resolveOrderPublicByResumeToken(resumeToken)
    return result.data
  } catch (_err: unknown) {
    return null
  }
}

async function resolveOrderFromOutTradeNo(outTradeNo: string): Promise<ResolvedOrder | null> {
  try {
    const result = await paymentAPI.verifyOrder(outTradeNo)
    return result.data
  } catch (_err: unknown) {
    try {
      const result = await paymentAPI.verifyOrderPublic(outTradeNo)
      return result.data
    } catch (_innerErr: unknown) {
      return null
    }
  }
}

function clearStatusRefreshTimer(): void {
  if (statusRefreshTimer !== null) {
    clearTimeout(statusRefreshTimer)
    statusRefreshTimer = null
  }
}

function clearRecoverySnapshot(): void {
  if (typeof window === 'undefined') return
  clearPaymentRecoverySnapshot(window.localStorage, PAYMENT_RECOVERY_STORAGE_KEY)
}

function clearRecoverySnapshotForTerminalStatus(status: string | null | undefined): void {
  if (!status) return
  if (!isPendingStatus(status)) {
    clearRecoverySnapshot()
  }
}

function scheduleStatusRefresh(refreshOrder: (() => Promise<ResolvedOrder | null>) | null): void {
  clearStatusRefreshTimer()
  if (!refreshOrder || !isPending.value) {
    return
  }
  if (refreshAttempts.value >= STATUS_REFRESH_MAX_ATTEMPTS) {
    statusRefreshExhausted.value = true
    return
  }

  statusRefreshTimer = setTimeout(async () => {
    refreshAttempts.value += 1
    const refreshedOrder = await refreshOrder()
    if (refreshedOrder) {
      setResolvedOrder(refreshedOrder)
      clearRecoverySnapshotForTerminalStatus(refreshedOrder.status)
    }

    if (isPendingStatus(order.value?.status)) {
      scheduleStatusRefresh(refreshOrder)
    } else {
      statusRefreshExhausted.value = false
    }
  }, STATUS_REFRESH_INTERVAL_MS)
}

async function retryStatus(): Promise<void> {
  if (!refreshOrderAction || retrying.value) return
  retrying.value = true
  statusRefreshExhausted.value = false
  refreshAttempts.value = 0
  try {
    const refreshedOrder = await refreshOrderAction()
    if (refreshedOrder) {
      setResolvedOrder(refreshedOrder)
      clearRecoverySnapshotForTerminalStatus(refreshedOrder.status)
    } else {
      lastCheckedAt.value = new Date()
    }
    if (isPendingStatus(order.value?.status)) {
      scheduleStatusRefresh(refreshOrderAction)
    }
  } finally {
    retrying.value = false
  }
}

onMounted(async () => {
  const resumeToken = readRouteQueryString('resume_token')
  const routeOrderId = Number(readRouteQueryString('order_id')) || 0
  let outTradeNo = readRouteQueryString('out_trade_no')
  let orderId = 0
  let resumeTokenLookupFailed = false

  const restored = restoreRecoverySnapshot({
    resumeToken,
    routeOrderId,
    routeOutTradeNo: outTradeNo,
  })
  if (restored?.orderId) {
    orderId = restored.orderId
  }
  if (restored?.currency) {
    currency.value = normalizePaymentCurrency(restored.currency)
  }
  if (!outTradeNo && restored?.outTradeNo) {
    outTradeNo = restored.outTradeNo
  }

  if (resumeToken) {
    const resolvedOrder = await resolveOrderFromResumeToken(resumeToken)
    if (resolvedOrder) {
      setResolvedOrder(resolvedOrder)
      if (!orderId) {
        orderId = hasOrderId(resolvedOrder) ? resolvedOrder.id : 0
      }
    } else if (routeOrderId > 0) {
      resumeTokenLookupFailed = true
      orderId = routeOrderId
    } else {
      resumeTokenLookupFailed = true
    }
  } else if (routeOrderId > 0) {
    orderId = routeOrderId
  }

  const hasLegacyFallbackContext = readRouteQueryString('trade_status').trim() !== ''
  const shouldUsePublicOutTradeNo = outTradeNo !== '' && (hasLegacyFallbackContext || routeOrderId > 0 || orderId > 0)

  if (!order.value && orderId && (!resumeToken || routeOrderId > 0)) {
    try {
      setResolvedOrder(await paymentStore.pollOrderStatus(orderId))
    } catch (_err: unknown) {
      // Order lookup failed, will try legacy fallback below when possible.
    }
  }

  if (!order.value && shouldUsePublicOutTradeNo && (!resumeToken || resumeTokenLookupFailed)) {
    const legacyOrder = await resolveOrderFromOutTradeNo(outTradeNo)
    if (legacyOrder) {
      setResolvedOrder(legacyOrder)
      if (!orderId) {
        orderId = hasOrderId(legacyOrder) ? legacyOrder.id : 0
      }
    }
  }

  if (!order.value && !orderId && outTradeNo && hasLegacyFallbackContext) {
    returnInfo.value = {
      outTradeNo,
      money: String(route.query.money || ''),
      type: String(route.query.type || ''),
      tradeStatus: String(route.query.trade_status || ''),
    }
  }

  const refreshOrder = async (): Promise<ResolvedOrder | null> => {
    if (resumeToken) {
      const resolvedOrder = await resolveOrderFromResumeToken(resumeToken)
      if (resolvedOrder) {
        return resolvedOrder
      }
    }

    if (orderId) {
      try {
        return await paymentStore.pollOrderStatus(orderId)
      } catch (_err: unknown) {
        // Fall through to legacy public verification when order polling is unavailable.
      }
    }

    if (shouldUsePublicOutTradeNo) {
      return await resolveOrderFromOutTradeNo(outTradeNo)
    }

    return null
  }
  refreshOrderAction = refreshOrder

  if (isPendingStatus(order.value?.status)) {
    scheduleStatusRefresh(refreshOrder)
  } else if (order.value) {
    clearRecoverySnapshotForTerminalStatus(order.value.status)
  } else if (returnInfo.value) {
    clearRecoverySnapshot()
  }
  loading.value = false
})

onBeforeUnmount(() => {
  clearStatusRefreshTimer()
})
</script>
