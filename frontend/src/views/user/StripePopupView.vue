<template>
  <main class="flex min-h-screen items-center justify-center bg-canvas p-4 text-foreground">
    <div
      class="w-full max-w-md space-y-5 rounded-panel border border-outline bg-surface p-5 shadow-modal sm:p-6"
    >
      <!-- Amount + Order ID -->
      <div v-if="amount" class="text-center">
        <p class="text-3xl font-semibold tabular-nums text-foreground">¥{{ amount }}</p>
        <p v-if="orderId" class="mt-1 break-all text-sm text-foreground-muted">
          {{ t('payment.orders.orderId') }}: {{ orderId }}
        </p>
      </div>

      <!-- Error -->
      <div v-if="error" class="space-y-3" role="alert">
        <div
          class="rounded-panel border border-danger/30 bg-danger-subtle p-3 text-sm text-danger-foreground"
        >
          {{ error }}
        </div>
        <button
          type="button"
          class="btn btn-secondary w-full"
          @click="closeWindow"
        >
          {{ t('common.close') }}
        </button>
      </div>

      <!-- Success -->
      <div v-else-if="success" class="space-y-3 py-4 text-center" role="status">
        <span class="mx-auto inline-flex h-14 w-14 items-center justify-center rounded-panel bg-success-subtle text-success-foreground">
          <Icon name="check" size="lg" />
        </span>
        <p class="font-semibold text-foreground">{{ t('payment.result.success') }}</p>
        <button
          type="button"
          class="btn btn-secondary"
          @click="closeWindow"
        >
          {{ t('common.close') }}
        </button>
      </div>

      <!-- Loading / Redirecting -->
      <div v-else class="flex items-center justify-center py-8" role="status">
        <Icon name="refresh" size="lg" class="animate-spin" :style="{ color: methodColor }" />
        <span class="ml-3 text-sm text-foreground-muted">{{ hint }}</span>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import { buildApiUrl } from '@/api/client'
import Icon from '@/components/icons/Icon.vue'

interface StripeWithWechatPay {
  confirmWechatPayPayment(clientSecret: string, options: Record<string, unknown>): Promise<{ error?: { message?: string }; paymentIntent?: { status: string } }>
}

const METHOD_COLORS: Record<string, string> = {
  alipay: '#00AEEF',
  wechat_pay: '#07C160',
}
const DEFAULT_METHOD_COLOR = '#635bff'

const { t } = useI18n()
const route = useRoute()

const orderId = String(route.query.order_id || '')
const method = String(route.query.method || 'alipay')
const amount = String(route.query.amount || '')

const methodColor = computed(() => METHOD_COLORS[method] || DEFAULT_METHOD_COLOR)

const error = ref('')
const success = ref(false)
const hint = ref(t('payment.stripePopup.redirecting'))

let pollTimer: ReturnType<typeof setInterval> | null = null
let initTimeoutTimer: ReturnType<typeof setTimeout> | null = null
let messageHandler: ((event: MessageEvent) => void) | null = null

function closeWindow() { window.close() }

function clearInitTimeout() {
  if (initTimeoutTimer) {
    clearTimeout(initTimeoutTimer)
    initTimeoutTimer = null
  }
}

onMounted(() => {
  messageHandler = (event: MessageEvent) => {
    if (event.origin !== window.location.origin) return
    if (event.data?.type !== 'STRIPE_POPUP_INIT') return
    error.value = ''
    clearInitTimeout()
    if (messageHandler) {
      window.removeEventListener('message', messageHandler)
      messageHandler = null
    }
    initStripe(event.data.clientSecret, event.data.publishableKey)
  }
  window.addEventListener('message', messageHandler)

  if (window.opener) {
    window.opener.postMessage({ type: 'STRIPE_POPUP_READY' }, window.location.origin)
  }

  initTimeoutTimer = setTimeout(() => {
    if (!error.value && !success.value) {
      error.value = t('payment.stripePopup.timeout')
    }
  }, 15000)
})

onUnmounted(() => {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  clearInitTimeout()
  if (messageHandler) {
    window.removeEventListener('message', messageHandler)
    messageHandler = null
  }
})

async function initStripe(clientSecret: string, publishableKey: string) {
  if (!clientSecret || !publishableKey) {
    error.value = t('payment.stripeMissingParams')
    return
  }
  try {
    const { loadStripe } = await import('@stripe/stripe-js')
    const stripe = await loadStripe(publishableKey)
    if (!stripe) { error.value = t('payment.stripeLoadFailed'); return }

    const returnUrl = window.location.origin + '/payment/result?order_id=' + orderId + '&status=success'

    if (method === 'alipay') {
      // Alipay: redirect this popup to Alipay payment page
      const { error: err } = await stripe.confirmAlipayPayment(clientSecret, { return_url: returnUrl })
      if (err) error.value = err.message || t('payment.result.failed')
    } else if (method === 'wechat_pay') {
      // WeChat: Stripe shows its built-in QR dialog, user scans, promise resolves
      hint.value = t('payment.stripePopup.loadingQr')
      const result = await (stripe as unknown as StripeWithWechatPay).confirmWechatPayPayment(clientSecret, {
        payment_method_options: { wechat_pay: { client: isMobileDevice() ? 'mobile_web' : 'web' } },
      })
      if (result.error) {
        error.value = result.error.message || t('payment.result.failed')
      } else if (result.paymentIntent?.status === 'succeeded') {
        success.value = true
        setTimeout(closeWindow, 2000)
      } else {
        // Payment not completed (user closed QR dialog)
        startPolling()
      }
    }
  } catch (err: unknown) {
    error.value = extractI18nErrorMessage(err, t, 'payment.errors', t('payment.stripeLoadFailed'))
  }
}

function startPolling() {
  let inFlight = false
  pollTimer = setInterval(async () => {
    // 防重入：接口响应慢于轮询间隔时避免并发重叠请求。
    if (inFlight) return
    inFlight = true
    try {
      // access token 存储在 localStorage 的 'auth_token' 键下（见 api/client.ts），
      // 之前误读 'token' 导致轮询请求不带认证、永远 401，支付成功无法被检测到。
      const token = localStorage.getItem('auth_token') || ''
      const res = await fetch(buildApiUrl(`/payment/orders/${orderId}`), {
        headers: token ? { Authorization: 'Bearer ' + token } : {},
        credentials: 'include',
      })
      if (!res.ok) return
      const data = await res.json()
      const status = data?.data?.status
      if (status === 'COMPLETED' || status === 'PAID') {
        if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
        success.value = true
        setTimeout(closeWindow, 2000)
      }
    } catch { /* ignore */ } finally {
      inFlight = false
    }
  }, 3000)
}
</script>
