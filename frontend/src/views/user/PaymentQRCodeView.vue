<template>
  <AppLayout>
    <section class="mx-auto w-full max-w-lg py-4 sm:py-8" aria-live="polite">
      <div
        v-if="missingPaymentData"
        class="rounded-panel border border-danger/30 bg-danger-subtle p-5 text-center text-danger-foreground"
        role="alert"
      >
        <span class="mx-auto inline-flex h-11 w-11 items-center justify-center rounded-panel bg-surface">
          <Icon name="exclamationCircle" size="lg" />
        </span>
        <h2 class="mt-3 text-base font-semibold">{{ t('payment.qr.missingParams') }}</h2>
        <button type="button" class="btn btn-primary mt-5 w-full sm:w-auto" @click="router.push('/purchase')">
          {{ t('payment.result.backToRecharge') }}
        </button>
      </div>

      <div v-else class="rounded-panel border border-outline bg-surface p-4 shadow-card sm:p-6">
        <header class="text-center">
          <h2 class="text-lg font-semibold text-foreground sm:text-xl">
            {{ qrUrl ? scanTitle : t('payment.qr.payInNewWindow') }}
          </h2>
          <p v-if="qrUrl && !expired && scanHint" class="mt-2 text-sm text-foreground-muted">
            {{ scanHint }}
          </p>
        </header>

        <div v-if="qrUrl" class="mx-auto mt-5 w-full max-w-72 rounded-panel border border-outline bg-white p-3">
          <canvas
            ref="qrCanvas"
            class="mx-auto block h-auto w-full max-w-64"
            role="img"
            :aria-label="scanTitle"
          ></canvas>
        </div>

        <div v-if="expired" class="mt-5 text-center" role="status">
          <p class="font-semibold text-danger-foreground">{{ t('payment.qr.expired') }}</p>
          <p class="mt-1 text-sm text-foreground-muted">{{ t('payment.qr.expiredDesc') }}</p>
          <button type="button" class="btn btn-primary mt-5 w-full sm:w-auto" @click="router.push('/purchase')">
            {{ t('payment.result.backToRecharge') }}
          </button>
        </div>

        <template v-else>
          <div
            v-if="pollUnavailable"
            class="mt-5 rounded-panel border border-warning/30 bg-warning-subtle p-3 text-warning-foreground"
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

          <div class="mt-5 border-y border-outline py-4 text-center">
            <p class="text-sm text-foreground-muted">
              {{ qrUrl ? t('payment.qr.expiresIn') : t('payment.qr.payInNewWindowHint') }}
            </p>
            <p class="mt-1 text-2xl font-semibold tabular-nums text-foreground" role="timer">
              {{ countdownDisplay }}
            </p>
            <p class="mt-1 text-xs text-foreground-subtle">{{ t('payment.qr.waitingPayment') }}</p>
          </div>

          <div class="mt-5 flex flex-col gap-2 sm:flex-row">
            <a
              v-if="payUrl && !qrUrl"
              :href="payUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-primary flex-1"
            >
              {{ t('payment.qr.openPayWindow') }}
            </a>
            <button
              v-if="orderId"
              type="button"
              class="btn btn-secondary flex-1"
              :disabled="cancelling"
              @click="handleCancel"
            >
              {{ cancelling ? t('common.processing') : t('payment.qr.cancelOrder') }}
            </button>
          </div>
        </template>
      </div>
    </section>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import { usePaymentStore } from '@/stores/payment'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { useAppStore } from '@/stores'
import { isBuiltInAlipayMethod, isBuiltInWxpayMethod } from '@/components/payment/providerConfig'
import QRCode from 'qrcode'
import alipayIcon from '@/assets/icons/alipay.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const paymentStore = usePaymentStore()
const appStore = useAppStore()

const qrCanvas = ref<HTMLCanvasElement | null>(null)
const qrUrl = ref('')
const payUrl = ref('')
const orderId = ref(0)
const remainingSeconds = ref(0)
const expired = ref(false)
const cancelling = ref(false)
const paymentType = ref('')
const initialized = ref(false)
const pollFailureCount = ref(0)
const pollUnavailable = ref(false)
const retryingStatus = ref(false)

let pollTimer: ReturnType<typeof setInterval> | null = null
let countdownTimer: ReturnType<typeof setInterval> | null = null
let pollingStatus = false

const countdownDisplay = computed(() => {
  const m = Math.floor(remainingSeconds.value / 60)
  const s = remainingSeconds.value % 60
  return m.toString().padStart(2, '0') + ':' + s.toString().padStart(2, '0')
})

const isAlipay = computed(() => isBuiltInAlipayMethod(paymentType.value))
const isWxpay = computed(() => isBuiltInWxpayMethod(paymentType.value))
const missingPaymentData = computed(
  () => initialized.value && !orderId.value && !qrUrl.value && !payUrl.value
)

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

function getLogoForType(): string | null {
  if (isAlipay.value) return alipayIcon
  if (isWxpay.value) return wxpayIcon
  return null
}

async function renderQR() {
  await nextTick()
  if (!qrCanvas.value || !qrUrl.value) return

  // Use medium error correction to support logo overlay while keeping QR code scannable
  const logoSrc = getLogoForType()
  await QRCode.toCanvas(qrCanvas.value, qrUrl.value, {
    width: 256,
    margin: 2,
    errorCorrectionLevel: logoSrc ? 'M' : 'L',
  })

  if (!logoSrc) return

  // Draw logo in center of QR code
  const canvas = qrCanvas.value
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const img = new Image()
  img.src = logoSrc
  img.onload = () => {
    const logoSize = 48
    const x = (canvas.width - logoSize) / 2
    const y = (canvas.height - logoSize) / 2
    // White background with rounded corners
    const pad = 5
    ctx.fillStyle = '#FFFFFF'
    ctx.beginPath()
    const r = 6
    ctx.moveTo(x - pad + r, y - pad)
    ctx.arcTo(x + logoSize + pad, y - pad, x + logoSize + pad, y + logoSize + pad, r)
    ctx.arcTo(x + logoSize + pad, y + logoSize + pad, x - pad, y + logoSize + pad, r)
    ctx.arcTo(x - pad, y + logoSize + pad, x - pad, y - pad, r)
    ctx.arcTo(x - pad, y - pad, x + logoSize + pad, y - pad, r)
    ctx.fill()
    // Draw logo
    ctx.drawImage(img, x, y, logoSize, logoSize)
  }
}

async function pollStatus() {
  if (!orderId.value || pollingStatus) return
  pollingStatus = true
  try {
    const order = await paymentStore.pollOrderStatus(orderId.value)
    pollFailureCount.value = 0
    pollUnavailable.value = false
    if (!order) return
    if (order.status === 'COMPLETED' || order.status === 'PAID') {
      cleanup()
      router.push({ path: '/payment/result', query: { order_id: String(orderId.value), status: 'success' } })
    } else if (order.status === 'EXPIRED' || order.status === 'CANCELLED' || order.status === 'FAILED') {
      cleanup()
      expired.value = true
    }
  } catch {
    pollFailureCount.value += 1
    if (pollFailureCount.value >= 3) pollUnavailable.value = true
  } finally {
    pollingStatus = false
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
  if (remainingSeconds.value <= 0) {
    expired.value = true
    return
  }
  countdownTimer = setInterval(() => {
    remainingSeconds.value--
    if (remainingSeconds.value <= 0) {
      expired.value = true
      cleanup()
    }
  }, 1000)
}

async function handleCancel() {
  if (!orderId.value || cancelling.value) return
  cancelling.value = true
  try {
    await paymentAPI.cancelOrder(orderId.value)
    cleanup()
    router.push('/purchase')
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    cancelling.value = false
  }
}

function cleanup() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null }
}

watch(qrUrl, () => renderQR())

onMounted(() => {
  orderId.value = Number(route.query.order_id) || 0
  qrUrl.value = String(route.query.qr || '')
  payUrl.value = String(route.query.pay_url || '')
  paymentType.value = String(route.query.payment_type || '')
  initialized.value = true

  if (missingPaymentData.value) return

  // Calculate countdown from expiresAt
  const expiresAtStr = String(route.query.expires_at || '')
  let seconds = 30 * 60 // fallback: 30 minutes
  if (expiresAtStr) {
    const expiresAt = new Date(expiresAtStr).getTime()
    if (Number.isFinite(expiresAt)) {
      seconds = Math.floor((expiresAt - Date.now()) / 1000)
    }
  }
  startCountdown(seconds)
  if (orderId.value) pollTimer = setInterval(pollStatus, 3000)
  renderQR()
})

onUnmounted(() => cleanup())
</script>
