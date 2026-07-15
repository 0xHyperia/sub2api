<template>
  <div class="space-y-4">
    <div
      v-if="pollUnavailable && !outcome"
      class="rounded-panel border border-warning/30 bg-warning-subtle p-3 text-warning-foreground"
      role="status"
    >
      <p class="text-sm">{{ t('payment.qr.statusUnavailable') }}</p>
      <button
        type="button"
        class="btn btn-secondary btn-sm mt-3"
        :disabled="retryingStatus"
        @click="retryStatus"
      >
        <Icon name="refresh" size="sm" :class="{ 'animate-spin': retryingStatus }" />
        {{ t('common.retry') }}
      </button>
    </div>

    <template v-if="outcome === 'success'">
      <div class="card p-4 sm:p-5" role="status" aria-live="polite">
        <div class="flex flex-col items-center space-y-4 py-3">
          <div class="flex h-14 w-14 items-center justify-center rounded-full bg-success-subtle">
            <Icon name="check" size="lg" class="text-success-foreground" aria-hidden="true" />
          </div>
          <p class="text-lg font-semibold text-foreground">{{ props.orderType === 'subscription' ? t('payment.result.subscriptionSuccess') : t('payment.result.success') }}</p>
          <dl v-if="paidOrder" class="w-full border-y border-outline px-1 text-sm">
            <div class="flex justify-between gap-4 py-2">
              <dt class="text-foreground-subtle">{{ t('payment.orders.orderId') }}</dt>
              <dd class="font-medium text-foreground">#{{ paidOrder.id }}</dd>
            </div>
            <div v-if="paidOrder.out_trade_no" class="flex min-w-0 justify-between gap-4 border-t border-outline py-2">
              <dt class="text-foreground-subtle">{{ t('payment.orders.orderNo') }}</dt>
              <dd class="min-w-0 break-all text-right font-medium text-foreground">{{ paidOrder.out_trade_no }}</dd>
            </div>
            <div class="flex justify-between gap-4 border-t border-outline py-2">
              <dt class="text-foreground-subtle">{{ t('payment.orders.amount') }}</dt>
              <dd class="font-medium tabular-nums text-foreground">{{ creditedAmountSymbol }}{{ paidOrder.amount.toFixed(2) }}</dd>
            </div>
            <div class="flex justify-between gap-4 border-t border-outline py-2">
              <dt class="text-foreground-subtle">{{ t('payment.orders.payAmount') }}</dt>
              <dd class="font-medium tabular-nums text-foreground">{{ formatGatewayAmount(paidOrder.pay_amount, paidOrder.currency) }}</dd>
            </div>
          </dl>
          <div v-if="displayedCardCodes.length" class="w-full rounded-panel border border-outline bg-surface-subtle p-4">
            <div class="mb-2 flex items-center justify-between gap-3">
              <div>
                <span class="text-sm font-medium text-foreground">{{ cardCodesTitle }}</span>
                <p v-if="cardCodesHint" class="mt-0.5 text-xs text-foreground-subtle">{{ cardCodesHint }}</p>
              </div>
              <button class="btn btn-secondary btn-sm" type="button" @click="copyCardCodes">
                {{ t('common.copy') }}
              </button>
            </div>
            <div class="space-y-2">
              <code
                v-for="code in displayedCardCodes"
                :key="code"
                class="block break-all rounded-control border border-outline bg-surface px-3 py-2 text-sm text-foreground"
              >
                {{ code }}
              </code>
            </div>
          </div>
          <button type="button" class="btn btn-primary" @click="handleDone">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>

    <!-- Cancelled -->
    <template v-else-if="outcome === 'cancelled'">
      <div class="card p-4 sm:p-5" role="status" aria-live="polite">
        <div class="flex flex-col items-center space-y-4 py-4">
          <div class="flex h-16 w-16 items-center justify-center rounded-full bg-surface-subtle">
            <Icon name="x" size="lg" class="text-foreground-subtle" aria-hidden="true" />
          </div>
          <p class="text-lg font-semibold text-foreground">{{ t('payment.qr.cancelled') }}</p>
          <p class="text-center text-sm leading-6 text-foreground-subtle">{{ t('payment.qr.cancelledDesc') }}</p>
          <button type="button" class="btn btn-primary" @click="handleDone">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>

    <!-- Expired / Failed -->
    <template v-else-if="outcome === 'expired'">
      <div class="card p-4 sm:p-5" role="status" aria-live="polite">
        <div class="flex flex-col items-center space-y-4 py-4">
          <div class="flex h-16 w-16 items-center justify-center rounded-full bg-warning-subtle">
            <Icon name="clock" size="lg" class="text-warning-foreground" aria-hidden="true" />
          </div>
          <p class="text-lg font-semibold text-foreground">{{ t('payment.qr.expired') }}</p>
          <p class="text-center text-sm leading-6 text-foreground-subtle">{{ t('payment.qr.expiredDesc') }}</p>
          <button type="button" class="btn btn-primary" @click="handleDone">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>

    <template v-else-if="qrUrl">
      <div class="card p-4 sm:p-5">
        <div class="flex flex-col items-center space-y-4">
          <p class="text-lg font-semibold text-foreground">{{ scanTitle }}</p>
          <div :class="['relative rounded-panel border bg-white p-4', qrBorderClass]">
            <canvas ref="qrCanvas" class="mx-auto" role="img" :aria-label="scanTitle"></canvas>
            <div class="pointer-events-none absolute inset-0 flex items-center justify-center">
              <span :class="['rounded-full p-2 shadow ring-2 ring-white', qrLogoBgClass]">
                <img :src="qrLogoIcon" alt="" class="h-5 w-5 brightness-0 invert" />
              </span>
            </div>
          </div>
          <p v-if="scanHint" class="text-center text-sm leading-6 text-foreground-subtle">{{ scanHint }}</p>
          <button v-if="payUrl" type="button" class="btn btn-secondary text-sm" @click="reopenPopup">
            {{ t('payment.qr.openPayWindow') }}
          </button>
        </div>
      </div>
      <div class="card p-4 text-center">
        <p class="text-sm text-foreground-subtle">{{ t('payment.qr.expiresIn') }}</p>
        <p class="mt-1 text-2xl font-semibold tabular-nums text-foreground" role="timer">{{ countdownDisplay }}</p>
        <p class="mt-1 text-xs text-foreground-subtle">{{ t('payment.qr.waitingPayment') }}</p>
      </div>
      <button type="button" class="btn btn-secondary w-full" :disabled="cancelling" @click="handleCancel">
        {{ cancelling ? t('common.processing') : t('payment.qr.cancelOrder') }}
      </button>
    </template>

    <template v-else>
      <div class="card p-4 sm:p-5">
        <div class="flex flex-col items-center space-y-4 py-4">
          <LoadingSpinner />
          <p class="text-center text-sm leading-6 text-foreground-subtle">{{ t('payment.qr.payInNewWindowHint') }}</p>
          <button v-if="payUrl" type="button" class="btn btn-secondary" @click="reopenPopup">
            {{ t('payment.qr.openPayWindow') }}
          </button>
        </div>
      </div>
      <div class="card p-4 text-center">
        <p class="mt-1 text-2xl font-semibold tabular-nums text-foreground" role="timer">{{ countdownDisplay }}</p>
        <p class="mt-1 text-xs text-foreground-subtle">{{ t('payment.qr.waitingPayment') }}</p>
      </div>
      <button type="button" class="btn btn-secondary w-full" :disabled="cancelling" @click="handleCancel">
        {{ cancelling ? t('common.processing') : t('payment.qr.cancelOrder') }}
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePaymentStore } from '@/stores/payment'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { getPaymentPopupFeatures, isBuiltInAlipayMethod, isBuiltInWxpayMethod } from '@/components/payment/providerConfig'
import { currencySymbol, formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'
import type { PaymentOrder } from '@/types/payment'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import QRCode from 'qrcode'
import alipayIcon from '@/assets/icons/alipay.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'
import paymentIcon from '@/assets/icons/payment.svg'

const props = defineProps<{
  orderId: number
  qrCode: string
  expiresAt: string
  paymentType: string
  payUrl?: string
  orderType?: string
  currency?: string
}>()

type PaymentOutcome = 'success' | 'cancelled' | 'expired'

const emit = defineEmits<{ done: []; success: []; settled: [outcome: PaymentOutcome] }>()

const i18n = useI18n()
const { t } = i18n
const paymentStore = usePaymentStore()
const appStore = useAppStore()

const qrCanvas = ref<HTMLCanvasElement | null>(null)
const qrUrl = ref('')
const remainingSeconds = ref(0)
const cancelling = ref(false)
const paidOrder = ref<PaymentOrder | null>(null)
const pollFailureCount = ref(0)
const pollUnavailable = ref(false)
const retryingStatus = ref(false)
const paymentCurrency = computed(() => normalizePaymentCurrency(props.currency))
const creditedAmountSymbol = currencySymbol('USD')
const localeCode = computed(() => {
  const raw = i18n.locale as unknown
  if (typeof raw === 'string') return raw
  if (raw && typeof raw === 'object' && 'value' in raw) {
    return String((raw as { value?: string }).value || '')
  }
  return undefined
})

// Terminal outcome: null = still active, 'success' | 'cancelled' | 'expired'
const outcome = ref<PaymentOutcome | null>(null)

let pollTimer: ReturnType<typeof setInterval> | null = null
let countdownTimer: ReturnType<typeof setInterval> | null = null
let verifyAttempts = 0
let lastVerifyAt = 0

const VERIFY_RETRY_INTERVAL_MS = 15000
const VERIFY_RETRY_MAX_ATTEMPTS = 6

const isAlipay = computed(() => isBuiltInAlipayMethod(props.paymentType))
const isWxpay = computed(() => isBuiltInWxpayMethod(props.paymentType))
const shouldVerifyPendingOrder = computed(() => {
  return isWxpay.value || props.orderType === 'card' || props.paymentType === 'ldxp'
})

const qrBorderClass = computed(() => {
  if (isAlipay.value) return 'border-[#00AEEF] border-[#00AEEF]/70 bg-info-subtle'
  if (isWxpay.value) return 'border-[#2BB741] border-[#2BB741]/70 bg-success-subtle'
  return 'border-outline-strong bg-surface'
})

const qrLogoBgClass = computed(() => {
  if (isAlipay.value) return 'bg-[#00AEEF]'
  if (isWxpay.value) return 'bg-[#2BB741]'
  return 'bg-outline-strong'
})

const qrLogoIcon = computed(() => {
  if (isAlipay.value) return alipayIcon
  if (isWxpay.value) return wxpayIcon
  return paymentIcon
})

const scanTitle = computed(() => {
  if (isAlipay.value) return t('payment.qr.scanAlipay')
  if (isWxpay.value) return t('payment.qr.scanWxpay')
  return t('payment.qr.scanToPay')
})

const scanHint = computed(() => {
  if (isAlipay.value) return t('payment.qr.scanAlipayHint')
  if (isWxpay.value) return t('payment.qr.scanWxpayHint')
  return ''
})

const cardCodesTitle = computed(() => {
  if (props.orderType === 'card' && paidOrder.value?.card_auto_redeem) {
    return t('payment.card.redeemedCodes')
  }
  return t('payment.card.codes')
})

const cardCodesHint = computed(() => {
  if (props.orderType !== 'card') return ''
  if (paidOrder.value?.card_auto_redeem) {
    return t('payment.card.redeemedCodesHint')
  }
  return t('payment.card.codesHint')
})

const displayedCardCodes = computed(() => {
  if (!paidOrder.value) return []
  if (paidOrder.value.card_auto_redeem && paidOrder.value.redeemed_card_codes?.length) {
    return paidOrder.value.redeemed_card_codes
  }
  return paidOrder.value.card_codes || []
})

const countdownDisplay = computed(() => {
  const m = Math.floor(remainingSeconds.value / 60)
  const s = remainingSeconds.value % 60
  return m.toString().padStart(2, '0') + ':' + s.toString().padStart(2, '0')
})

function formatGatewayAmount(value: number, currency?: string | null): string {
  return formatPaymentAmount(value, currency || paymentCurrency.value, localeCode.value)
}

function isSuccessStatus(status: string | null | undefined): boolean {
  return status === 'COMPLETED' || status === 'PAID' || status === 'RECHARGING'
}

function reopenPopup() {
  if (props.payUrl) {
    const win = window.open(props.payUrl, 'paymentPopup', getPaymentPopupFeatures())
    if (!win || win.closed) {
      window.location.href = props.payUrl
    }
  }
}

function setOutcome(next: PaymentOutcome) {
  if (outcome.value === next) return
  outcome.value = next
  emit('settled', next)
}

async function copyCardCodes() {
  const codes = displayedCardCodes.value
  if (!codes.length || typeof navigator === 'undefined' || !navigator.clipboard) return
  await navigator.clipboard.writeText(codes.join('\n'))
  appStore.showSuccess(t('common.copied'))
}

async function renderQR() {
  await nextTick()
  if (!qrCanvas.value || !qrUrl.value) return
  await QRCode.toCanvas(qrCanvas.value, qrUrl.value, {
    width: 220, margin: 2,
    errorCorrectionLevel: 'M',
  })
}

async function tryVerifyPendingOrder(order: PaymentOrder): Promise<PaymentOrder> {
  if (!shouldVerifyPendingOrder.value) return order
  const outTradeNo = String(order.out_trade_no || '').trim()
  if (!outTradeNo) return order
  const normalizedStatus = String(order.status || '').trim().toUpperCase()
  if (normalizedStatus !== 'PENDING') return order
  const now = Date.now()
  if (verifyAttempts >= VERIFY_RETRY_MAX_ATTEMPTS || now - lastVerifyAt < VERIFY_RETRY_INTERVAL_MS) {
    return order
  }

  lastVerifyAt = now
  verifyAttempts += 1
  try {
    const result = await paymentAPI.verifyOrder(outTradeNo)
    return result.data ?? order
  } catch {
    return order
  }
}

let pollInFlight = false
async function pollStatus() {
  if (!props.orderId || outcome.value) return
  if (pollInFlight) return
  pollInFlight = true
  try {
    let order = await paymentStore.pollOrderStatus(props.orderId)
    pollFailureCount.value = 0
    pollUnavailable.value = false
    if (!order) return
    if (outcome.value) return
    order = await tryVerifyPendingOrder(order)
    if (outcome.value) return
    if (isSuccessStatus(order.status)) {
      cleanup()
      paidOrder.value = order
      setOutcome('success')
      emit('success')
    } else if (order.status === 'CANCELLED') {
      cleanup()
      setOutcome('cancelled')
    } else if (order.status === 'EXPIRED' || order.status === 'FAILED') {
      cleanup()
      setOutcome('expired')
    }
  } catch {
    pollFailureCount.value += 1
    if (pollFailureCount.value >= 3) pollUnavailable.value = true
  } finally {
    pollInFlight = false
  }
}

async function retryStatus() {
  if (retryingStatus.value) return
  retryingStatus.value = true
  pollFailureCount.value = 0
  pollUnavailable.value = false
  try {
    await pollStatus()
  } finally {
    retryingStatus.value = false
  }
}

function startCountdown(seconds: number) {
  remainingSeconds.value = Math.max(0, seconds)
  if (remainingSeconds.value <= 0) { setOutcome('expired'); return }
  countdownTimer = setInterval(() => {
    remainingSeconds.value--
    if (remainingSeconds.value <= 0) { setOutcome('expired'); cleanup() }
  }, 1000)
}

async function handleCancel() {
  if (!props.orderId || cancelling.value) return
  cancelling.value = true
  try {
    await paymentAPI.cancelOrder(props.orderId)
    cleanup()
    setOutcome('cancelled')
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    cancelling.value = false
  }
}

function handleDone() { cleanup(); emit('done') }

function cleanup() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null }
}

// Initialize on mount
qrUrl.value = props.qrCode
verifyAttempts = 0
lastVerifyAt = 0
let seconds = 30 * 60
if (props.expiresAt) {
  seconds = Math.floor((new Date(props.expiresAt).getTime() - Date.now()) / 1000)
}
startCountdown(seconds)
pollTimer = setInterval(pollStatus, 3000)
renderQR()

watch(() => qrUrl.value, () => renderQR())
onUnmounted(() => cleanup())
</script>
