<template>
  <AppLayout>
    <div :class="pageContainerClass">
      <h1 class="sr-only">{{ t('payment.title') }}</h1>

      <div v-if="loading" class="card flex min-h-48 items-center justify-center" aria-live="polite">
        <LoadingSpinner />
      </div>

      <div v-else-if="checkoutLoadFailed" class="card" role="alert">
        <EmptyState
          :title="t('common.error')"
          :description="t('errors.tryAgain')"
          :action-text="t('common.refresh')"
          @action="reloadCheckout"
        />
      </div>

      <template v-else>
        <header v-if="paymentPhase === 'select'" class="space-y-5">
          <div class="page-header mb-0 flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div class="min-w-0">
              <p class="text-sm font-medium text-foreground-subtle">{{ t('purchaseWorkspace.accountTitle') }}</p>
              <h2 class="page-title mt-1">{{ t('payment.title') }}</h2>
              <p class="page-description">{{ t('purchaseWorkspace.accountDescription') }}</p>
            </div>
            <RouterLink to="/orders" class="btn btn-secondary shrink-0">
              <Icon name="clipboard" size="sm" aria-hidden="true" />
              <span>{{ t('purchaseWorkspace.orders') }}</span>
            </RouterLink>
          </div>

          <dl class="grid overflow-hidden rounded-panel border border-outline bg-surface sm:grid-cols-2 xl:grid-cols-4">
            <div class="min-w-0 border-b border-outline p-4 sm:border-r sm:p-5 xl:border-b-0">
              <dt class="text-sm text-foreground-subtle">{{ t('payment.currentBalance') }}</dt>
              <dd class="mt-2 text-2xl font-semibold tabular-nums text-foreground">{{ '$' }}{{ user?.balance?.toFixed(2) || '0.00' }}</dd>
            </div>
            <div class="min-w-0 border-b border-outline p-4 sm:p-5 xl:border-b-0 xl:border-r">
              <dt class="text-sm text-foreground-subtle">{{ t('purchaseWorkspace.todaySpent') }}</dt>
              <dd class="mt-2 text-2xl font-semibold tabular-nums text-foreground">
                <span v-if="accountStatsLoading">--</span>
                <span v-else>{{ '$' }}{{ (accountStats?.today_actual_cost || 0).toFixed(2) }}</span>
              </dd>
            </div>
            <div class="min-w-0 border-b border-outline p-4 sm:border-b-0 sm:border-r sm:p-5">
              <dt class="text-sm text-foreground-subtle">{{ t('purchaseWorkspace.totalSpent') }}</dt>
              <dd class="mt-2 text-2xl font-semibold tabular-nums text-foreground">
                <span v-if="accountStatsLoading">--</span>
                <span v-else>{{ '$' }}{{ (accountStats?.total_actual_cost || 0).toFixed(2) }}</span>
              </dd>
            </div>
            <div class="min-w-0 p-4 sm:p-5">
              <dt class="text-sm text-foreground-subtle">{{ t('purchaseWorkspace.totalRequests') }}</dt>
              <dd class="mt-2 text-2xl font-semibold tabular-nums text-foreground">
                <span v-if="accountStatsLoading">--</span>
                <span v-else>{{ formatCount(accountStats?.total_requests || 0) }}</span>
              </dd>
              <p v-if="accountStatsLoadFailed" class="mt-1 text-xs text-warning-foreground">
                {{ t('purchaseWorkspace.statsUnavailable') }}
              </p>
            </div>
          </dl>
        </header>

        <div
          v-if="errorMessage"
          class="rounded-panel border border-danger/20 bg-danger-subtle p-4 text-sm text-danger-foreground"
          role="alert"
        >
          <p class="font-medium">{{ errorMessage }}</p>
          <p v-if="errorHintMessage" class="mt-1 leading-6 opacity-90">{{ errorHintMessage }}</p>
        </div>

        <section v-if="paymentPhase === 'paying'" class="py-2">
          <PaymentStatusPanel
            :order-id="paymentState.orderId"
            :qr-code="paymentState.qrCode"
            :expires-at="paymentState.expiresAt"
            :payment-type="paymentState.paymentType"
            :pay-url="paymentState.payUrl"
            :order-type="paymentState.orderType"
            :currency="paymentState.currency || selectedCurrency"
            @done="onPaymentDone"
            @success="onPaymentSuccess"
            @settled="onPaymentSettled"
          />
        </section>

        <template v-else>
          <div v-if="tabs.length === 0" class="card">
            <EmptyState :title="t('payment.notAvailable')" />
          </div>

          <div v-else class="space-y-5">
            <div
              :class="[
                'grid items-start gap-5',
                workspaceGridClass,
              ]"
            >
              <article class="card min-w-0 overflow-hidden">
                <div v-if="tabs.length > 1" class="border-b border-outline bg-surface-subtle p-2.5 sm:p-3">
                  <div class="tabs grid w-full grid-cols-2" role="tablist" :aria-label="t('payment.title')">
                    <button
                      v-for="(tab, index) in tabs"
                      :id="'purchase-tab-' + tab.key"
                      :key="tab.key"
                      type="button"
                      class="tab purchase-tab min-w-0"
                      :class="{ 'purchase-tab-active': activeTab === tab.key }"
                      role="tab"
                      :aria-selected="activeTab === tab.key"
                      :aria-controls="'purchase-panel-' + tab.key"
                      :tabindex="activeTab === tab.key ? 0 : -1"
                      @click="activeTab = tab.key"
                      @keydown="handleTabKeydown($event, index)"
                    >
                      {{ tab.label }}
                    </button>
                  </div>
                </div>

                <section
                  v-if="activeTab === 'recharge'"
                  id="purchase-panel-recharge"
                  class="p-4 sm:p-5"
                  role="tabpanel"
                  :aria-labelledby="tabs.length > 1 ? 'purchase-tab-recharge' : undefined"
                  :aria-label="tabs.length === 1 ? t('payment.tabTopUp') : undefined"
                >
                  <div v-if="enabledMethods.length === 0">
                    <EmptyState :title="t('payment.notAvailable')" />
                  </div>

                  <template v-else>
                    <header>
                      <h2 class="text-base font-semibold text-foreground">{{ t('payment.amountLabel') }}</h2>
                      <p class="mt-1 text-sm text-foreground-subtle">{{ t('payment.customAmount') }}</p>
                    </header>

                    <div class="mt-4 min-w-0">
                      <AmountInput
                        v-model="amount"
                        :amounts="[10, 20, 50, 100, 200, 500, 1000, 2000, 5000]"
                        :min="globalMinAmount"
                        :max="globalMaxAmount"
                      />
                      <p v-if="amountError" class="mt-3 text-sm text-warning-foreground" role="alert">{{ amountError }}</p>

                      <div class="mt-5 border-t border-outline pt-5">
                        <PaymentMethodSelector
                          compact
                          :methods="methodOptions"
                          :selected="selectedMethod"
                          @select="selectedMethod = $event"
                        />
                      </div>

                      <div class="mt-5 grid gap-5 border-t border-outline pt-5 md:grid-cols-[minmax(0,1fr)_minmax(240px,300px)] md:items-end">
                        <div class="min-w-0">
                          <dl class="space-y-2.5 text-sm">
                            <div class="flex items-center justify-between gap-4">
                              <dt class="text-foreground-subtle">{{ t('payment.paymentAmount') }}</dt>
                              <dd class="font-medium tabular-nums text-foreground">{{ formatSelectedPaymentAmount(validAmount) }}</dd>
                            </div>
                            <div v-if="feeRate > 0" class="flex items-center justify-between gap-4">
                              <dt class="text-foreground-subtle">{{ t('payment.fee') }} ({{ feeRate }}%)</dt>
                              <dd class="font-medium tabular-nums text-foreground">{{ formatSelectedPaymentAmount(feeAmount) }}</dd>
                            </div>
                            <div v-if="balanceRechargeMultiplier !== 1" class="flex items-center justify-between gap-4">
                              <dt class="text-foreground-subtle">{{ t('payment.creditedBalance') }}</dt>
                              <dd class="font-medium tabular-nums text-foreground">{{ '$' }}{{ creditedAmount.toFixed(2) }}</dd>
                            </div>
                            <div class="flex items-center justify-between gap-4 border-t border-outline pt-2.5">
                              <dt class="font-medium text-foreground">{{ t('payment.actualPay') }}</dt>
                              <dd class="text-xl font-semibold tabular-nums text-foreground">{{ formatSelectedPaymentAmount(totalAmount) }}</dd>
                            </div>
                          </dl>
                          <p v-if="balanceRechargeMultiplier !== 1" class="mt-3 text-xs leading-5 text-foreground-subtle">
                            {{ t('payment.rechargeRatePreview', { usd: balanceRechargeMultiplier.toFixed(2) }) }}
                          </p>
                        </div>
                        <button
                          type="button"
                          data-testid="recharge-confirm-action"
                          :class="['btn btn-lg w-full md:min-w-[240px] md:justify-self-end', paymentButtonClass]"
                          :disabled="!canSubmit || submitting"
                          :aria-busy="submitting"
                          @click="handleSubmitRecharge"
                        >
                          <span v-if="submitting" class="flex items-center justify-center gap-2">
                            <span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
                            {{ t('common.processing') }}
                          </span>
                          <span v-else>{{ t('payment.createOrder') }} {{ formatSelectedPaymentAmount(totalAmount) }}</span>
                        </button>
                      </div>
                    </div>
                  </template>
                </section>

                <section
                  v-else-if="activeTab === 'subscription'"
                  id="purchase-panel-subscription"
                  class="p-4 sm:p-5"
                  role="tabpanel"
                  :aria-labelledby="tabs.length > 1 ? 'purchase-tab-subscription' : undefined"
                  :aria-label="tabs.length === 1 ? t('payment.tabSubscribe') : undefined"
                >
                  <header class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <div>
                      <h2 class="text-base font-semibold text-foreground">{{ t('payment.selectPlan') }}</h2>
                      <p class="mt-1 text-sm text-foreground-subtle">{{ t('payment.tabSubscribe') }}</p>
                    </div>
                    <RouterLink to="/subscriptions" class="btn btn-secondary btn-sm shrink-0">
                      <span>{{ t('userSubscriptions.title') }}</span>
                      <Icon name="arrowRight" size="sm" aria-hidden="true" />
                    </RouterLink>
                  </header>

                  <div v-if="checkout.plans.length === 0" class="mt-4">
                    <EmptyState :title="t('payment.noPlans')" />
                  </div>

                  <template v-else>
                    <div data-testid="subscription-plan-grid" :class="['plan-grid mt-4', planGridClass]">
                      <SubscriptionPlanCard
                        v-for="plan in checkout.plans"
                        :key="plan.id"
                        :plan="plan"
                        @select="selectPlan"
                      />
                    </div>

                    <div v-if="selectedPlan" class="mt-5 border-t border-outline pt-5">
                      <div class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_minmax(260px,0.72fr)]">
                        <div class="min-w-0">
                          <div class="flex min-w-0 items-start justify-between gap-3">
                            <div class="min-w-0">
                              <p class="text-sm text-foreground-subtle">{{ t('payment.selectPlan') }}</p>
                              <h3 class="mt-1 break-words text-lg font-semibold text-foreground">{{ selectedPlan.name }}</h3>
                            </div>
                            <span :class="['badge shrink-0', planBadgeClass]">{{ platformLabel(selectedPlan.group_platform || '') }}</span>
                          </div>
                          <div class="mt-4 flex flex-wrap items-baseline gap-x-2 gap-y-1">
                            <span :class="['text-2xl font-semibold tabular-nums', planTextClass]">{{ formatSelectedSubscriptionPaymentAmount(selectedPlan.price) }}</span>
                            <span class="text-sm text-foreground-subtle">/ {{ planValiditySuffix }}</span>
                            <span v-if="selectedPlan.original_price" class="text-sm text-foreground-subtle line-through">{{ formatSelectedSubscriptionPaymentAmount(selectedPlan.original_price) }}</span>
                          </div>
                          <p v-if="selectedPlan.description" class="mt-2 text-sm leading-6 text-foreground-subtle">{{ selectedPlan.description }}</p>
                        </div>

                        <div class="min-w-0 space-y-4 border-t border-outline pt-5 lg:border-l lg:border-t-0 lg:pl-5 lg:pt-0">
                          <PaymentMethodSelector
                            compact
                            :methods="subMethodOptions"
                            :selected="selectedMethod"
                            @select="selectedMethod = $event"
                          />
                          <dl class="space-y-2.5 border-t border-outline pt-4 text-sm">
                            <div class="flex items-center justify-between gap-4">
                              <dt class="text-foreground-subtle">{{ t('payment.amountLabel') }}</dt>
                              <dd class="font-medium tabular-nums text-foreground">{{ formatSelectedPaymentAmount(subPaymentAmount) }}</dd>
                            </div>
                            <div v-if="feeRate > 0 && selectedPlan.price > 0" class="flex items-center justify-between gap-4">
                              <dt class="text-foreground-subtle">{{ t('payment.fee') }} ({{ feeRate }}%)</dt>
                              <dd class="font-medium tabular-nums text-foreground">{{ formatSelectedPaymentAmount(subFeeAmount) }}</dd>
                            </div>
                          </dl>
                        </div>
                      </div>

                      <div class="mt-5 flex flex-col gap-4 border-t border-outline pt-5 sm:flex-row sm:items-end sm:justify-between">
                        <dl class="min-w-0">
                          <dt class="text-sm font-medium text-foreground-subtle">{{ t('payment.actualPay') }}</dt>
                          <dd class="mt-1 text-2xl font-semibold tabular-nums text-foreground">{{ formatSelectedPaymentAmount(subTotalAmount) }}</dd>
                        </dl>
                        <button
                          type="button"
                          data-testid="subscription-confirm-action"
                          :class="['btn btn-lg w-full sm:w-auto sm:min-w-[240px]', paymentButtonClass]"
                          :disabled="!canSubmitSubscription || submitting"
                          :aria-busy="submitting"
                          @click="confirmSubscribe"
                        >
                          <span v-if="submitting" class="flex items-center justify-center gap-2">
                            <span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
                            {{ t('common.processing') }}
                          </span>
                          <span v-else>{{ t('payment.createOrder') }} {{ formatSelectedPaymentAmount(subTotalAmount) }}</span>
                        </button>
                      </div>
                    </div>
                  </template>
                </section>

                <div class="border-t border-outline p-4 sm:p-5">
                  <RedeemCodeForm />
                </div>
              </article>

              <AffiliateRewardPanel
                v-if="affiliateEnabled"
                data-testid="affiliate-reward-panel"
                :class="affiliatePanelClass"
              />
            </div>

            <section v-if="showInstantHelp" class="grid gap-4 lg:grid-cols-2">
              <article v-if="showInstantHelp" class="card">
                <div class="flex min-w-0 flex-col gap-4 p-4 sm:flex-row sm:items-center sm:p-5">
                  <button
                    v-if="checkout.help_image_url"
                    type="button"
                    class="shrink-0 rounded-panel border border-outline bg-surface-subtle p-1 focus:outline-none focus-visible:ring-2 focus-visible:ring-focus"
                    :aria-label="t('common.view')"
                    @click="previewImage = checkout.help_image_url"
                  >
                    <img :src="checkout.help_image_url" :alt="checkout.help_text || t('payment.title')" class="h-20 w-28 rounded-control object-contain" />
                  </button>
                  <p v-if="checkout.help_text" class="text-sm leading-6 text-foreground-subtle">{{ checkout.help_text }}</p>
                </div>
              </article>
            </section>

          </div>
        </template>
      </template>
    </div>

    <BaseDialog :show="showRenewalModal" :title="t('payment.selectPlan')" width="wide" @close="closeRenewalModal">
      <div class="grid gap-4 sm:grid-cols-2">
        <SubscriptionPlanCard
          v-for="plan in renewalPlans"
          :key="plan.id"
          :plan="plan"
          @select="selectPlanFromModal"
        />
      </div>
    </BaseDialog>

    <BaseDialog :show="!!previewImage" :title="t('payment.title')" width="wide" :z-index="60" @close="previewImage = ''">
      <img v-if="previewImage" :src="previewImage" :alt="checkout.help_text || t('payment.title')" class="mx-auto max-h-[75dvh] max-w-full rounded-panel object-contain" />
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePaymentStore } from '@/stores/payment'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import usageAPI, { type UserDashboardStats } from '@/api/usage'
import { extractApiErrorMessage, extractI18nErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import type { SubscriptionPlan, CheckoutInfoResponse, CreateOrderResult, OrderType } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import AmountInput from '@/components/payment/AmountInput.vue'
import AffiliateRewardPanel from '@/components/payment/AffiliateRewardPanel.vue'
import PaymentMethodSelector from '@/components/payment/PaymentMethodSelector.vue'
import RedeemCodeForm from '@/components/payment/RedeemCodeForm.vue'
import { METHOD_ORDER, getPaymentPopupFeatures, isBuiltInAlipayMethod, isBuiltInWxpayMethod } from '@/components/payment/providerConfig'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  buildCreateOrderPayload,
  clearPaymentRecoverySnapshot,
  decidePaymentLaunch,
  getVisibleMethods,
  normalizeVisibleMethod,
  readPaymentRecoverySnapshot,
  type PaymentRecoverySnapshot,
  writePaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import { platformBadgeClass, platformTextClass, platformLabel } from '@/utils/platformColors'
import SubscriptionPlanCard from '@/components/payment/SubscriptionPlanCard.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import { DEFAULT_PAYMENT_CURRENCY, formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'
import type { PaymentMethodOption } from '@/components/payment/PaymentMethodSelector.vue'
import { buildPaymentErrorToastMessage, describePaymentScenarioError } from './paymentUx'
import { hasWechatResumeQuery, parseWechatResumeRoute, stripWechatResumeQuery } from './paymentWechatResume'

const i18n = useI18n()
const { t } = i18n
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const paymentStore = usePaymentStore()
const subscriptionStore = useSubscriptionStore()
const appStore = useAppStore()

const user = computed(() => authStore.user)

const loading = ref(true)
const checkoutLoadFailed = ref(false)
const accountStats = ref<UserDashboardStats | null>(null)
const accountStatsLoading = ref(true)
const accountStatsLoadFailed = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const errorHintMessage = ref('')
type PurchaseTab = 'recharge' | 'subscription'

const activeTab = ref<PurchaseTab>('recharge')
const amount = ref<number | null>(null)
const selectedMethod = ref('')
const selectedPlan = ref<SubscriptionPlan | null>(null)
const previewImage = ref('')

const paymentPhase = ref<'select' | 'paying'>('select')

const affiliateEnabled = computed(() => appStore.cachedPublicSettings?.affiliate_enabled === true)
const workspaceGridClass = computed(() => {
  if (!affiliateEnabled.value) return 'grid-cols-1'
  return activeTab.value === 'subscription'
    ? 'grid-cols-1 2xl:grid-cols-[minmax(0,1fr)_minmax(380px,440px)]'
    : 'grid-cols-1 xl:grid-cols-[minmax(0,1fr)_minmax(380px,460px)]'
})
const affiliatePanelClass = computed(() => activeTab.value === 'subscription'
  ? 'w-full max-w-[720px] 2xl:sticky 2xl:top-24 2xl:max-w-none'
  : 'xl:sticky xl:top-24'
)

function formatCount(value: number): string {
  return value.toLocaleString()
}

async function loadAccountStats(): Promise<void> {
  accountStatsLoading.value = true
  accountStatsLoadFailed.value = false
  try {
    accountStats.value = await usageAPI.getDashboardStats()
  } catch (error) {
    console.error('Failed to load purchase account stats:', error)
    accountStatsLoadFailed.value = true
  } finally {
    accountStatsLoading.value = false
  }
}

interface CreateOrderOptions {
  openid?: string
  wechatResumeToken?: string
  paymentType?: string
  isResume?: boolean
  mobileQrFallbackAttempted?: boolean
}

interface WeixinJSBridgeLike {
  invoke(
    action: string,
    payload: Record<string, unknown>,
    callback: (result: Record<string, unknown>) => void,
  ): void
}

function emptyPaymentState(): PaymentRecoverySnapshot {
  return {
    orderId: 0,
    amount: 0,
    qrCode: '',
    expiresAt: '',
    paymentType: '',
    payUrl: '',
    outTradeNo: '',
    clientSecret: '',
    intentId: '',
    currency: '',
    countryCode: '',
    paymentEnv: '',
    payAmount: 0,
    orderType: '',
    paymentMode: '',
    resumeToken: '',
    createdAt: 0,
  }
}

function getWeixinJSBridge(): WeixinJSBridgeLike | undefined {
  return (window as Window & { WeixinJSBridge?: WeixinJSBridgeLike }).WeixinJSBridge
}

function waitForWeixinJSBridge(timeoutMs = 4000): Promise<WeixinJSBridgeLike | null> {
  const existing = getWeixinJSBridge()
  if (existing) return Promise.resolve(existing)

  return new Promise((resolve) => {
    let settled = false
    const finish = (bridge: WeixinJSBridgeLike | null) => {
      if (settled) return
      settled = true
      document.removeEventListener('WeixinJSBridgeReady', handleReady)
      document.removeEventListener('onWeixinJSBridgeReady', handleReady)
      window.clearTimeout(timer)
      resolve(bridge)
    }
    const handleReady = () => finish(getWeixinJSBridge() ?? null)
    const timer = window.setTimeout(() => finish(getWeixinJSBridge() ?? null), timeoutMs)
    document.addEventListener('WeixinJSBridgeReady', handleReady, false)
    document.addEventListener('onWeixinJSBridgeReady', handleReady, false)
  })
}

async function invokeWechatJsapiPayment(payload: Record<string, unknown>): Promise<Record<string, unknown>> {
  const bridge = await waitForWeixinJSBridge()
  if (!bridge) {
    throw new Error('WECHAT_JSAPI_UNAVAILABLE')
  }
  return new Promise((resolve) => {
    bridge.invoke('getBrandWCPayRequest', payload, (result) => resolve(result || {}))
  })
}

const paymentState = ref<PaymentRecoverySnapshot>(emptyPaymentState())

function persistRecoverySnapshot(snapshot: PaymentRecoverySnapshot) {
  if (typeof window === 'undefined' || !snapshot.orderId) return
  writePaymentRecoverySnapshot(window.localStorage, snapshot, PAYMENT_RECOVERY_STORAGE_KEY)
}

function removeRecoverySnapshot() {
  if (typeof window === 'undefined') return
  clearPaymentRecoverySnapshot(window.localStorage, PAYMENT_RECOVERY_STORAGE_KEY)
}

function resetPayment() {
  paymentPhase.value = 'select'
  paymentState.value = emptyPaymentState()
  removeRecoverySnapshot()
}

async function redirectToPaymentResult(state: PaymentRecoverySnapshot): Promise<void> {
  const query: Record<string, string | undefined> = {}
  if (state.orderId > 0) {
    query.order_id = String(state.orderId)
  }
  if (state.outTradeNo) {
    query.out_trade_no = state.outTradeNo
  }
  if (state.resumeToken) {
    query.resume_token = state.resumeToken
  }
  await router.push({
    path: '/payment/result',
    query,
  })
}

function buildWechatOAuthAuthorizeUrl(
  authorizeUrl: string,
  context: { paymentType: string; orderType: OrderType; planId?: number; orderAmount: number },
): string {
  const normalizedUrl = authorizeUrl.trim()
  if (!normalizedUrl || typeof window === 'undefined') {
    return normalizedUrl
  }

  try {
    const targetUrl = new URL(normalizedUrl, window.location.origin)
    const redirectPath = targetUrl.searchParams.get('redirect') || '/purchase'
    const redirectUrl = new URL(redirectPath, window.location.origin)

    const paymentType = normalizeVisibleMethod(context.paymentType) || context.paymentType.trim() || 'wxpay'

    redirectUrl.searchParams.set('payment_type', paymentType)
    redirectUrl.searchParams.set('order_type', context.orderType)

    if (context.planId) {
      redirectUrl.searchParams.set('plan_id', String(context.planId))
    } else {
      redirectUrl.searchParams.delete('plan_id')
    }

    if (context.orderAmount > 0) {
      redirectUrl.searchParams.set('amount', String(context.orderAmount))
    } else {
      redirectUrl.searchParams.delete('amount')
    }

    targetUrl.searchParams.set('redirect', `${redirectUrl.pathname}${redirectUrl.search}`)
    return targetUrl.toString()
  } catch {
    return normalizedUrl
  }
}

function onPaymentDone() {
  const wasSubscription = paymentState.value.orderType === 'subscription'
  resetPayment()
  selectedPlan.value = null
  if (wasSubscription) {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
}

function onPaymentSuccess() {
  removeRecoverySnapshot()
  authStore.refreshUser()
  if (paymentState.value.orderType === 'subscription') {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
}

function onPaymentSettled() {
  removeRecoverySnapshot()
}

// All checkout data from single API call
const checkout = ref<CheckoutInfoResponse>({
  methods: {}, global_min: 0, global_max: 0,
  plans: [], instant_enabled: true, balance_disabled: false, balance_recharge_multiplier: 1, subscription_usd_to_cny_rate: 0, recharge_fee_rate: 0, help_text: '', help_image_url: '', stripe_publishable_key: '',
})

const tabs = computed(() => {
  const result: { key: PurchaseTab; label: string }[] = []
  if (checkout.value.instant_enabled !== false) {
    if (!checkout.value.balance_disabled) result.push({ key: 'recharge', label: t('payment.tabTopUp') })
    result.push({ key: 'subscription', label: t('payment.tabSubscribe') })
  }
  return result
})

function handleTabKeydown(event: KeyboardEvent, index: number): void {
  let nextIndex = index
  if (event.key === 'ArrowRight') nextIndex = (index + 1) % tabs.value.length
  else if (event.key === 'ArrowLeft') nextIndex = (index - 1 + tabs.value.length) % tabs.value.length
  else if (event.key === 'Home') nextIndex = 0
  else if (event.key === 'End') nextIndex = tabs.value.length - 1
  else return

  event.preventDefault()
  const nextTab = tabs.value[nextIndex]
  if (!nextTab) return
  activeTab.value = nextTab.key
  const tabList = (event.currentTarget as HTMLElement).parentElement
  window.requestAnimationFrame(() => {
    tabList?.querySelectorAll<HTMLElement>('[role="tab"]')[nextIndex]?.focus()
  })
}

function reloadCheckout(): void {
  if (typeof window !== 'undefined') window.location.reload()
}

const showInstantHelp = computed(() =>
  (checkout.value.help_text || checkout.value.help_image_url)
    && paymentPhase.value === 'select'
)

const pageContainerClass = computed(() => 'mx-auto w-full max-w-[1440px] space-y-5')

function ensureVisiblePurchaseTab(preferred?: PurchaseTab) {
  const available = tabs.value.map(tab => tab.key)
  if (preferred && available.includes(preferred)) {
    activeTab.value = preferred
    return
  }
  if (!available.includes(activeTab.value)) {
    activeTab.value = available[0] || 'recharge'
  }
}

const visibleMethods = computed(() => getVisibleMethods(checkout.value.methods))
const enabledMethods = computed(() => Object.keys(visibleMethods.value))
const validAmount = computed(() => amount.value ?? 0)
const balanceRechargeMultiplier = computed(() => {
  const multiplier = checkout.value.balance_recharge_multiplier
  return Number.isFinite(multiplier) && multiplier > 0 ? multiplier : 1
})
// 订阅 CNY 换算汇率（1 USD = X CNY）。0 = 未配置，订阅保持 price 直付（与后端 opt-in 条件严格镜像）。
const subscriptionUsdToCnyRate = computed(() => {
  const rate = checkout.value.subscription_usd_to_cny_rate
  return Number.isFinite(rate) && rate > 0 ? rate : 0
})
const creditedAmount = computed(() => Math.round((validAmount.value * balanceRechargeMultiplier.value) * 100) / 100)

const planGridClass = computed(() => {
  if (checkout.value.plans.length === 1) {
    return 'sm:max-w-md'
  }
  return ''
})

// Check if an amount fits a method's [min, max]. 0 = no limit.
function amountFitsMethod(amt: number, methodType: string): boolean {
  if (amt <= 0) return true
  const ml = visibleMethods.value[methodType]
  if (!ml) return false
  if (ml.single_min > 0 && amt < ml.single_min) return false
  if (ml.single_max > 0 && amt > ml.single_max) return false
  return true
}

// Visible methods decide the amount range shown to users.
const globalMinAmount = computed(() => {
  const limits = Object.values(visibleMethods.value)
  if (limits.length === 0) return 0
  if (limits.some(limit => limit.single_min <= 0)) return 0
  return Math.min(...limits.map(limit => limit.single_min))
})
const globalMaxAmount = computed(() => {
  const limits = Object.values(visibleMethods.value)
  if (limits.length === 0) return 0
  if (limits.some(limit => limit.single_max <= 0)) return 0
  return Math.max(...limits.map(limit => limit.single_max))
})

// Selected method's limits (for validation and error messages)
const selectedLimit = computed(() => visibleMethods.value[selectedMethod.value])
const selectedCurrency = computed(() => normalizePaymentCurrency(selectedLimit.value?.currency))
const localeCode = computed(() => {
  const raw = i18n.locale as unknown
  if (typeof raw === 'string') return raw
  if (raw && typeof raw === 'object' && 'value' in raw) {
    return String((raw as { value?: string }).value || '')
  }
  return undefined
})

function currencyFractionDigits(currency: string): number {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
    }).resolvedOptions().maximumFractionDigits ?? 2
  } catch {
    return 2
  }
}

function roundPaymentAmount(value: number, currency: string): number {
  if (!Number.isFinite(value)) return 0
  const factor = 10 ** currencyFractionDigits(currency)
  return Math.round(value * factor) / factor
}

function ceilPaymentAmount(value: number, currency: string): number {
  if (!Number.isFinite(value)) return 0
  const factor = 10 ** currencyFractionDigits(currency)
  return Math.ceil(value * factor) / factor
}

function subscriptionPaymentAmountForCurrency(value: number, currency: string): number {
  const rate = subscriptionUsdToCnyRate.value
  if (rate <= 0 || currency !== DEFAULT_PAYMENT_CURRENCY) return roundPaymentAmount(value, currency)
  return roundPaymentAmount(value * rate, currency)
}

function formatSelectedPaymentAmount(value: number): string {
  return formatPaymentAmount(value, selectedCurrency.value, localeCode.value)
}

function formatSelectedSubscriptionPaymentAmount(value: number): string {
  return formatSelectedPaymentAmount(subscriptionPaymentAmountForCurrency(value, selectedCurrency.value))
}


const methodOptions = computed<PaymentMethodOption[]>(() =>
  enabledMethods.value.map((type) => {
    const ml = visibleMethods.value[type]
    return {
      type,
      display_name: ml?.display_name,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(validAmount.value, type),
    }
  })
)

const feeRate = computed(() => checkout.value?.recharge_fee_rate ?? 0)
const feeAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.ceil(((validAmount.value * feeRate.value) / 100) * 100) / 100
    : 0
)
const totalAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.round((validAmount.value + feeAmount.value) * 100) / 100
    : validAmount.value
)

const amountError = computed(() => {
  if (validAmount.value <= 0) return ''
  // No method can handle this amount
  if (!enabledMethods.value.some((m) => amountFitsMethod(validAmount.value, m))) {
    return t('payment.amountNoMethod')
  }
  // Selected method can't handle this amount (but others can)
  const ml = selectedLimit.value
  if (ml) {
    if (ml.single_min > 0 && validAmount.value < ml.single_min) return t('payment.amountTooLow', { min: formatSelectedPaymentAmount(ml.single_min) })
    if (ml.single_max > 0 && validAmount.value > ml.single_max) return t('payment.amountTooHigh', { max: formatSelectedPaymentAmount(ml.single_max) })
  }
  return ''
})

const canSubmit = computed(() =>
  validAmount.value > 0
    && amountFitsMethod(validAmount.value, selectedMethod.value)
    && selectedLimit.value?.available !== false
)

const subPaymentAmount = computed(() => {
  const price = selectedPlan.value?.price ?? 0
  return subscriptionPaymentAmountForCurrency(price, selectedCurrency.value)
})

const subFeeAmount = computed(() => {
  if (feeRate.value <= 0 || subPaymentAmount.value <= 0) return 0
  return ceilPaymentAmount((subPaymentAmount.value * feeRate.value) / 100, selectedCurrency.value)
})

const subTotalAmount = computed(() => {
  if (feeRate.value <= 0 || subPaymentAmount.value <= 0) return subPaymentAmount.value
  return roundPaymentAmount(subPaymentAmount.value + subFeeAmount.value, selectedCurrency.value)
})

function subscriptionTotalAmountForCurrency(value: number, currency: string): number {
  const paymentAmount = subscriptionPaymentAmountForCurrency(value, currency)
  if (feeRate.value <= 0 || paymentAmount <= 0) return paymentAmount
  const fee = ceilPaymentAmount((paymentAmount * feeRate.value) / 100, currency)
  return roundPaymentAmount(paymentAmount + fee, currency)
}

// Subscription-specific: method options based on gateway pay amount
const subMethodOptions = computed<PaymentMethodOption[]>(() => {
  const price = selectedPlan.value?.price ?? 0
  return enabledMethods.value.map((type) => {
    const ml = visibleMethods.value[type]
    const currency = normalizePaymentCurrency(ml?.currency)
    return {
      type,
      display_name: ml?.display_name,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(subscriptionTotalAmountForCurrency(price, currency), type),
    }
  })
})

const canSubmitSubscription = computed(() =>
  selectedPlan.value !== null
    && amountFitsMethod(subTotalAmount.value, selectedMethod.value)
    && selectedLimit.value?.available !== false
)

// Auto-switch to first available method when current selection can't handle the amount
watch(() => [validAmount.value, selectedMethod.value] as const, ([amt, method]) => {
  if (amt <= 0 || amountFitsMethod(amt, method)) return
  const available = enabledMethods.value.find((m) => amountFitsMethod(amt, m))
  if (available) selectedMethod.value = available
})

// Payment button class: follows selected payment method color
const paymentButtonClass = computed(() => {
  const m = selectedMethod.value
  if (!m) return 'btn-primary'
  if (isBuiltInAlipayMethod(m)) return 'btn-alipay'
  if (isBuiltInWxpayMethod(m)) return 'btn-wxpay'
  if (m === 'stripe') return 'btn-stripe'
  if (m === 'airwallex') return 'btn-airwallex'
  return 'btn-primary'
})

// Subscription confirm: platform accent colors (clean card, no gradient)
const planBadgeClass = computed(() => platformBadgeClass(selectedPlan.value?.group_platform || ''))
const planTextClass = computed(() => platformTextClass(selectedPlan.value?.group_platform || ''))

// Renewal modal state
const showRenewalModal = ref(false)
const renewGroupId = ref<number | null>(null)
const renewalPlans = computed(() => {
  if (renewGroupId.value == null) return []
  return checkout.value.plans.filter(p => p.group_id === renewGroupId.value)
})

const planValiditySuffix = computed(() => {
  if (!selectedPlan.value) return ''
  const u = selectedPlan.value.validity_unit || 'day'
  if (u === 'month') return t('payment.perMonth')
  if (u === 'year') return t('payment.perYear')
  return `${selectedPlan.value.validity_days}${t('payment.days')}`
})

function selectPlan(plan: SubscriptionPlan) {
  selectedPlan.value = plan
  errorMessage.value = ''
}

function selectPlanFromModal(plan: SubscriptionPlan) {
  showRenewalModal.value = false
  renewGroupId.value = null
  selectedPlan.value = plan
  errorMessage.value = ''
}

function closeRenewalModal() {
  showRenewalModal.value = false
  renewGroupId.value = null
}

async function handleSubmitRecharge() {
  if (!canSubmit.value || submitting.value) return
  await createOrder(validAmount.value, 'balance')
}

async function confirmSubscribe() {
  if (!selectedPlan.value || submitting.value) return
  await createOrder(selectedPlan.value.price, 'subscription', selectedPlan.value.id)
}

async function createOrder(orderAmount: number, orderType: OrderType, planId?: number, options: CreateOrderOptions = {}) {
  submitting.value = true
  errorMessage.value = ''
  errorHintMessage.value = ''
  const requestType = normalizeVisibleMethod(options.paymentType || selectedMethod.value) || options.paymentType || selectedMethod.value
  try {
    const payload = buildCreateOrderPayload({
      amount: orderAmount,
      paymentType: requestType,
      orderType,
      planId,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: isMobileDevice(),
      isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
      forceQRCode: !!(checkout.value.alipay_force_qrcode && normalizeVisibleMethod(requestType) === 'alipay'),
    })
    if (options.openid) {
      payload.openid = options.openid
    }
    if (options.wechatResumeToken) {
      payload.wechat_resume_token = options.wechatResumeToken
    }

    const result = await paymentStore.createOrder(payload) as CreateOrderResult & { resume_token?: string }
    const openWindow = (url: string) => {
      const win = window.open(url, 'paymentPopup', getPaymentPopupFeatures())
      if (!win || win.closed) {
        window.location.href = url
      }
    }
    const visibleMethod = normalizeVisibleMethod(requestType) || requestType
    // When user clicks the dedicated Stripe button, leave method blank so the
    // landing page renders Stripe's full Payment Element (card/link/alipay/wxpay).
    const stripeMethod = visibleMethod === 'stripe'
      ? ''
      : visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret && visibleMethod !== 'airwallex'
      ? router.resolve({
        path: '/payment/stripe',

        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const airwallexRouteUrl = result.client_secret && result.intent_id
      ? router.resolve({
        path: '/payment/airwallex',
        query: {
          order_id: String(result.order_id),
          out_trade_no: result.out_trade_no || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType,
      isMobile: isMobileDevice(),
      isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
      forceQRCode: !!(checkout.value.alipay_force_qrcode && visibleMethod === 'alipay'),
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
      airwallexRouteUrl,
    })

    if (decision.kind === 'wechat_oauth' && decision.oauth?.authorize_url) {
      window.location.href = buildWechatOAuthAuthorizeUrl(decision.oauth.authorize_url, {
        paymentType: visibleMethod,
        orderType,
        planId,
        orderAmount,
      })
      return
    }

    if (decision.kind === 'unhandled') {
      applyScenarioError({ reason: 'UNHANDLED_PAYMENT_SCENARIO' }, visibleMethod)
      return
    }

    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    persistRecoverySnapshot(decision.recovery)

    if (decision.kind === 'stripe_popup') {
      openWindow(decision.paymentState.payUrl)
      return
    }
    if (decision.kind === 'stripe_route') {
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'airwallex_route') {
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'wechat_jsapi' && decision.jsapi) {
      try {
        const jsapiResult = await invokeWechatJsapiPayment(decision.jsapi as Record<string, unknown>)
        const errMsg = String(jsapiResult.err_msg || '').toLowerCase()
        if (errMsg.includes('cancel')) {
          appStore.showInfo(t('payment.qr.cancelled'))
          resetPayment()
        } else if (errMsg && !errMsg.includes('ok')) {
          resetPayment()
          const fallbackApplied = await attemptMobileQrFallback(
            { reason: 'WECHAT_JSAPI_FAILED', message: errMsg },
            {
              orderAmount,
              orderType,
              planId,
              paymentType: visibleMethod,
              attempted: options.mobileQrFallbackAttempted === true,
            },
          )
          if (!fallbackApplied) {
            applyScenarioError({ reason: 'WECHAT_JSAPI_FAILED', message: errMsg }, visibleMethod)
          }
        } else {
          const resultState = { ...decision.paymentState }
          resetPayment()
          await redirectToPaymentResult(resultState)
        }
      } catch (err: unknown) {
        resetPayment()
        const fallbackApplied = await attemptMobileQrFallback(err, {
          orderAmount,
          orderType,
          planId,
          paymentType: visibleMethod,
          attempted: options.mobileQrFallbackAttempted === true,
        })
        if (!fallbackApplied) {
          throw err
        }
      }
      return
    }
    if (decision.kind === 'redirect_waiting' && decision.paymentState.payUrl) {
      if (isMobileDevice()) {
        window.location.href = decision.paymentState.payUrl
        return
      }
      openWindow(decision.paymentState.payUrl)
    }
  } catch (err: unknown) {
    const apiErr = err as Record<string, unknown>
    if (apiErr.reason === 'TOO_MANY_PENDING') {
      const metadata = apiErr.metadata as Record<string, unknown> | undefined
      errorMessage.value = t('payment.errors.tooManyPending', { max: metadata?.max || '' })
      errorHintMessage.value = ''
    } else if (apiErr.reason === 'CANCEL_RATE_LIMITED') {
      errorMessage.value = t('payment.errors.cancelRateLimited')
      errorHintMessage.value = ''
    } else if (await attemptMobileQrFallback(err, {
      orderAmount,
      orderType,
      planId,
      paymentType: requestType,
      attempted: options.mobileQrFallbackAttempted === true,
    })) {
      return
    } else {
      const handled = applyScenarioError(
        err,
        normalizeVisibleMethod(options.paymentType || selectedMethod.value) || selectedMethod.value,
      )
      if (!handled) {
        errorMessage.value = extractI18nErrorMessage(err, t, 'payment.errors', extractApiErrorMessage(err, t('payment.result.failed')))
        errorHintMessage.value = ''
      }
      if (handled) {
        return
      }
    }
    appStore.showError(buildPaymentErrorToastMessage(errorMessage.value, errorHintMessage.value))
  } finally {
    submitting.value = false
  }
}

interface MobileQrFallbackContext {
  orderAmount: number
  orderType: OrderType
  planId?: number
  paymentType: string
  attempted: boolean
}

function shouldFallbackToDesktopQr(err: unknown, paymentMethod: string, attempted: boolean): boolean {
  if (attempted || !isMobileDevice()) {
    return false
  }

  const normalizedMethod = normalizeVisibleMethod(paymentMethod) || paymentMethod
  const reason = typeof err === 'object' && err && 'reason' in err && typeof err.reason === 'string'
    ? err.reason
    : ''
  const message = err instanceof Error
    ? err.message
    : (typeof err === 'object' && err && 'message' in err && typeof err.message === 'string'
      ? err.message
      : '')
  const normalizedMessage = message.toLowerCase()

  if (normalizedMethod === 'wxpay') {
    return reason === 'WECHAT_H5_NOT_AUTHORIZED'
      || reason === 'WECHAT_PAYMENT_MP_NOT_CONFIGURED'
      || reason === 'WECHAT_JSAPI_FAILED'
      || reason === 'PAYMENT_GATEWAY_ERROR'
      || reason === 'UNHANDLED_PAYMENT_SCENARIO'
      || normalizedMessage.includes('weixinjsbridge is unavailable')
      || normalizedMessage.includes('wechat_jsapi_unavailable')
  }

  if (normalizedMethod === 'alipay') {
    return reason === 'PAYMENT_GATEWAY_ERROR' || reason === 'UNHANDLED_PAYMENT_SCENARIO'
  }

  return false
}

async function attemptMobileQrFallback(err: unknown, context: MobileQrFallbackContext): Promise<boolean> {
  if (!shouldFallbackToDesktopQr(err, context.paymentType, context.attempted)) {
    return false
  }

  try {
    const visibleMethod = normalizeVisibleMethod(context.paymentType) || context.paymentType
    const payload = buildCreateOrderPayload({
      amount: context.orderAmount,
      paymentType: visibleMethod,
      orderType: context.orderType,
      planId: context.planId,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: false,

      isWechatBrowser: false,
    })
    const result = await paymentStore.createOrder(payload) as CreateOrderResult & { resume_token?: string }
    const stripeMethod = visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType: context.orderType,
      isMobile: false,
      isWechatBrowser: false,
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
    })

    if (decision.kind !== 'qr_waiting' || !decision.paymentState.qrCode) {
      return false
    }

    errorMessage.value = ''
    errorHintMessage.value = ''
    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    persistRecoverySnapshot(decision.recovery)
    appStore.showWarning(t('payment.errors.mobilePaymentFallbackToQr'))
    return true
  } catch {
    return false
  }
}

function applyScenarioError(err: unknown, paymentMethod: string): boolean {
  const descriptor = describePaymentScenarioError(err, {
    paymentMethod,
    isMobile: isMobileDevice(),
    isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
  })
  if (!descriptor) {
    errorMessage.value = ''
    errorHintMessage.value = ''
    return false
  }
  errorMessage.value = t(descriptor.messageKey)
  errorHintMessage.value = descriptor.hintKey ? t(descriptor.hintKey) : ''
  appStore.showError(buildPaymentErrorToastMessage(errorMessage.value, errorHintMessage.value))
  return true
}

async function resumeWechatPaymentFromQuery() {
  const resume = parseWechatResumeRoute(route.query, checkout.value.plans, validAmount.value)
  if (!resume) {
    return
  }

  selectedMethod.value = resume.paymentType
  if (resume.orderType === 'balance' && resume.orderAmount > 0) {
    amount.value = resume.orderAmount
  }
  if (resume.orderType === 'subscription' && resume.planId) {
    selectedPlan.value = checkout.value.plans.find(plan => plan.id === resume.planId) ?? null
  }

  await router.replace({ path: route.path, query: stripWechatResumeQuery(route.query) })

  if (resume.wechatResumeToken) {
    await createOrder(0, resume.orderType, resume.planId, {
      wechatResumeToken: resume.wechatResumeToken,
      paymentType: resume.paymentType,
      isResume: true,
    })
    return
  }

  if (resume.orderAmount > 0 && resume.openid) {
    await createOrder(resume.orderAmount, resume.orderType, resume.planId, {
      openid: resume.openid,
      paymentType: resume.paymentType,
      isResume: true,
    })
  }
}

onMounted(async () => {
  void loadAccountStats()
  try {
    checkoutLoadFailed.value = false
    const res = await paymentAPI.getCheckoutInfo()
    checkout.value = res.data
    ensureVisiblePurchaseTab()
    if (enabledMethods.value.length) {
      const order: readonly string[] = METHOD_ORDER
      const sorted = [...enabledMethods.value].sort((a, b) => {
        const ai = order.indexOf(a)
        const bi = order.indexOf(b)
        return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
      })
      selectedMethod.value = sorted[0]
    }
    if (typeof window !== 'undefined') {
      if (hasWechatResumeQuery(route.query)) {
        removeRecoverySnapshot()
      }
      const routeResumeToken = typeof route.query.resume_token === 'string'
        ? route.query.resume_token
        : typeof route.query.wechat_resume_token === 'string'
          ? route.query.wechat_resume_token
          : undefined
      const restored = readPaymentRecoverySnapshot(
        window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY),
        { resumeToken: routeResumeToken },
      )
      if (restored) {
        paymentState.value = restored
        paymentPhase.value = 'paying'
        const restoredMethod = normalizeVisibleMethod(restored.paymentType)
          || (visibleMethods.value[restored.paymentType] ? restored.paymentType : '')
        if (restoredMethod) {
          selectedMethod.value = restoredMethod
        }
      } else {
        removeRecoverySnapshot()
      }
    }
    await resumeWechatPaymentFromQuery()
    // Handle renewal navigation: ?tab=subscription&group=123
    if (route.query.tab === 'subscription') {
      ensureVisiblePurchaseTab('subscription')
      if (route.query.group) {
        const groupId = Number(route.query.group)
        const groupPlans = checkout.value.plans.filter(p => p.group_id === groupId)
        if (groupPlans.length === 1) {
          selectedPlan.value = groupPlans[0]
        } else if (groupPlans.length > 1) {
          renewGroupId.value = groupId
          showRenewalModal.value = true
        }
      }
    }
  } catch (err: unknown) {
    checkoutLoadFailed.value = true
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
  finally { loading.value = false }
  // Fetch active subscriptions (uses cache, non-blocking)
  subscriptionStore.fetchActiveSubscriptions().catch(() => {})
})
</script>

<style scoped>
.purchase-tab-active {
  color: #fff;
  background: #111;
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.2);
}

.purchase-tab-active:hover {
  color: #fff;
}

.plan-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 250px), 1fr));
  gap: 1rem;
}
</style>
