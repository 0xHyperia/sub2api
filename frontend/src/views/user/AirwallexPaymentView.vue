<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-lg space-y-4 py-4 sm:py-8">
      <div v-if="loading" class="flex min-h-64 items-center justify-center" role="status" :aria-label="t('common.loading')">
        <Icon name="refresh" size="lg" class="animate-spin text-foreground-subtle" />
      </div>

      <div
        v-else-if="errorMessage"
        class="rounded-panel border border-danger/30 bg-danger-subtle p-5 text-center text-danger-foreground"
        role="alert"
      >
        <span class="mx-auto inline-flex h-12 w-12 items-center justify-center rounded-panel bg-surface">
          <Icon name="exclamationCircle" size="lg" />
        </span>
        <h2 class="mt-3 text-lg font-semibold">{{ t('payment.airwallexLoadFailed') }}</h2>
        <p class="mt-2 break-words text-sm opacity-80">{{ errorMessage }}</p>
        <button type="button" class="btn btn-primary mt-5 w-full sm:w-auto" @click="router.push('/purchase')">{{ t('payment.result.backToRecharge') }}</button>
      </div>

      <div v-else class="rounded-panel border border-outline bg-surface p-6 shadow-card" role="status">
        <div class="flex flex-col items-center space-y-4 py-4">
          <Icon name="refresh" size="lg" class="animate-spin text-foreground-subtle" />
          <p class="text-center text-sm text-foreground-muted">{{ t('payment.qr.payInNewWindowHint') }}</p>
          <button type="button" class="btn btn-secondary btn-sm" @click="router.push('/purchase')">
            {{ t('payment.result.backToRecharge') }}
          </button>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  readPaymentRecoverySnapshot,
  type PaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()

const loading = ref(true)
const errorMessage = ref('')
let redirectTimeout: ReturnType<typeof setTimeout> | null = null

function queryString(key: string): string {
  const value = route.query[key]
  if (Array.isArray(value)) return value[0] || ''
  return typeof value === 'string' ? value : ''
}

function buildSuccessUrl(snapshot: PaymentRecoverySnapshot): string {
  const url = new URL('/payment/result', window.location.origin)
  const orderId = queryString('order_id')
  const outTradeNo = queryString('out_trade_no')
  const resumeToken = queryString('resume_token')

  if (orderId || snapshot.orderId > 0) url.searchParams.set('order_id', orderId || String(snapshot.orderId))
  if (outTradeNo || snapshot.outTradeNo) url.searchParams.set('out_trade_no', outTradeNo || snapshot.outTradeNo)
  if (resumeToken || snapshot.resumeToken) url.searchParams.set('resume_token', resumeToken || snapshot.resumeToken)
  return url.toString()
}

function restoreAirwallexSnapshot(): PaymentRecoverySnapshot | null {
  if (typeof window === 'undefined') {
    return null
  }

  const orderId = Number(queryString('order_id')) || 0
  const outTradeNo = queryString('out_trade_no')
  const resumeToken = queryString('resume_token')
  const snapshot = readPaymentRecoverySnapshot(
    window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY),
    resumeToken ? { resumeToken } : {},
  )

  if (!snapshot || snapshot.paymentType !== 'airwallex') {
    return null
  }
  if (orderId > 0 && snapshot.orderId !== orderId) {
    return null
  }
  if (outTradeNo && snapshot.outTradeNo !== outTradeNo) {
    return null
  }
  if (!snapshot.intentId || !snapshot.clientSecret) {
    return null
  }
  return snapshot
}

onMounted(async () => {
  const snapshot = restoreAirwallexSnapshot()
  const checkoutLocale = locale.value.toLowerCase().startsWith('zh') ? 'zh' : 'en'

  if (!snapshot) {
    loading.value = false
    errorMessage.value = t('payment.airwallexMissingParams')
    return
  }

  try {
    const airwallex = await import('@airwallex/components-sdk')
    const result = await airwallex.init({
      env: snapshot.paymentEnv === 'prod' ? 'prod' : 'demo',
      enabledElements: ['payments'],
      locale: checkoutLocale,
    })

    loading.value = false
    redirectTimeout = setTimeout(() => {
      redirectTimeout = null
      errorMessage.value = t('payment.airwallexLoadFailed')
    }, 20000)
    const checkoutOptions = {
      intent_id: snapshot.intentId,
      client_secret: snapshot.clientSecret,
      currency: snapshot.currency || 'CNY',
      country_code: snapshot.countryCode || 'CN',
      successUrl: buildSuccessUrl(snapshot),
    }
    if (!result.payments) {
      throw new Error(t('payment.airwallexLoadFailed'))
    }
    const redirectResult = result.payments.redirectToCheckout(checkoutOptions)

    if (typeof redirectResult === 'string' && redirectResult) {
      window.location.assign(redirectResult)
    }
  } catch (err: unknown) {
    if (redirectTimeout) {
      clearTimeout(redirectTimeout)
      redirectTimeout = null
    }
    loading.value = false
    errorMessage.value = err instanceof Error && err.message
      ? err.message
      : t('payment.airwallexLoadFailed')
  }
})

onUnmounted(() => {
  if (redirectTimeout) clearTimeout(redirectTimeout)
})
</script>
