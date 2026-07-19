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
            <div class="grid items-start gap-5 min-[1600px]:grid-cols-[minmax(0,1fr)_400px]">
              <article class="card min-w-0 overflow-hidden">
                <div
                  class="flex min-h-14 items-center gap-2 border-b border-outline bg-surface-subtle p-2.5 sm:p-3"
                  :class="{ 'min-[1600px]:hidden': tabs.length === 1 }"
                >
                  <div v-if="tabs.length > 1" class="purchase-tabs min-w-0 flex-1 sm:max-w-[360px]" role="tablist" :aria-label="t('payment.title')">
                    <button
                      v-for="(tab, index) in tabs"
                      :id="'purchase-tab-' + tab.key"
                      :key="tab.key"
                      type="button"
                      class="purchase-tab min-w-0 flex-1"
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
                  <div class="ml-auto flex shrink-0 items-center gap-1 min-[1600px]:hidden" :aria-label="t('purchaseWorkspace.toolsTitle')">
                    <button
                      type="button"
                      data-testid="purchase-mobile-tools-action"
                      class="btn btn-ghost btn-sm md:hidden"
                      @click="auxiliaryMode = 'combined'"
                    >
                      <Icon name="gift" size="sm" aria-hidden="true" />
                      <span>{{ t('purchaseWorkspace.toolsTitle') }}</span>
                    </button>
                    <button
                      type="button"
                      data-testid="purchase-redeem-action"
                      class="btn btn-ghost btn-sm hidden md:inline-flex"
                      @click="auxiliaryMode = 'redeem'"
                    >
                      <Icon name="gift" size="sm" aria-hidden="true" />
                      <span>{{ t('redeem.redeemCodeLabel') }}</span>
                    </button>
                    <button
                      v-if="affiliateEnabled"
                      type="button"
                      data-testid="purchase-affiliate-action"
                      class="btn btn-ghost btn-sm hidden md:inline-flex"
                      @click="auxiliaryMode = 'affiliate'"
                    >
                      <Icon name="users" size="sm" aria-hidden="true" />
                      <span>{{ t('purchaseWorkspace.affiliateTitle') }}</span>
                    </button>
                  </div>
                </div>

                <section
                  v-if="activeTab === 'recharge'"
                  id="purchase-panel-recharge"
                  class="min-w-0"
                  role="tabpanel"
                  :aria-labelledby="tabs.length > 1 ? 'purchase-tab-recharge' : undefined"
                  :aria-label="tabs.length === 1 ? t('payment.tabTopUp') : undefined"
                >
                  <div v-if="enabledMethods.length === 0">
                    <EmptyState :title="t('payment.notAvailable')" />
                  </div>

                  <template v-else>
                    <div class="grid min-w-0 xl:grid-cols-[minmax(0,1fr)_minmax(380px,0.8fr)]">
                      <div class="min-w-0 p-4 sm:p-5 lg:p-6">
                      <AmountInput
                        v-model="amount"
                        :amounts="checkout.quick_recharge_amounts"
                        :min="globalMinAmount"
                        :max="globalMaxAmount"
                        :custom-enabled="checkout.custom_recharge_amount_enabled"
                        :currency="BALANCE_CURRENCY"
                      />
                      <p v-if="amountError" class="mt-3 text-sm text-warning-foreground" role="alert">{{ amountError }}</p>

                      <div class="mt-4 flex items-center gap-2 rounded-control bg-info-subtle px-3 py-2.5 text-sm text-info-foreground">
                        <Icon name="arrowsUpDown" size="sm" class="shrink-0" aria-hidden="true" />
                        <span>{{ t('payment.fixedBalanceConversion') }}</span>
                      </div>

                      <RechargeValueEstimator
                        v-if="modelMarketplaceEnabled"
                        class="mt-5"
                        :credited-amount="creditedAmount"
                        :recharge-amount="validAmount"
                        :balance-recharge-multiplier="balanceRechargeMultiplier"
                        :official-usd-to-cny-rate="subscriptionUsdToCnyRate"
                        :quick-recharge-amounts="checkout.quick_recharge_amounts"
                      />

                      </div>

                      <aside class="min-w-0 border-t border-outline bg-surface-subtle p-4 sm:p-5 lg:p-6 xl:border-l xl:border-t-0" :aria-label="t('payment.checkoutSummary')">
                        <PaymentMethodSelector
                          compact
                          :methods="methodOptions"
                          :selected="selectedMethod"
                          @select="selectedMethod = $event"
                        />
                        <p
                          v-if="methodSwitchNotice"
                          class="mt-3 flex items-start gap-2 rounded-control border border-info/20 bg-info-subtle px-3 py-2 text-xs leading-5 text-info-foreground"
                          role="status"
                          aria-live="polite"
                        >
                          <Icon name="infoCircle" size="sm" class="mt-0.5 shrink-0" aria-hidden="true" />
                          <span>{{ methodSwitchNotice }}</span>
                        </p>

                        <div class="mt-5 divide-y divide-outline border-y border-outline text-sm">
                          <dl class="py-3">
                            <div class="mb-3 flex items-baseline justify-between gap-3">
                              <dt class="font-medium text-foreground">{{ t('payment.balanceSummary') }}</dt>
                              <dd class="text-xl font-semibold tabular-nums text-foreground">{{ formatBalanceAmount(creditedAmount) }}</dd>
                            </div>
                            <div class="flex items-center justify-between gap-3 text-foreground-subtle">
                              <dt>{{ t('payment.baseBalance') }}</dt>
                              <dd class="tabular-nums">{{ formatBalanceAmount(validAmount * balanceRechargeMultiplier) }}</dd>
                            </div>
                            <div v-if="selectedRechargeBonus > 0" class="mt-2 flex items-center justify-between gap-3 text-success-foreground">
                              <dt>{{ t('payment.bonusBalance') }}</dt>
                              <dd class="tabular-nums">+{{ formatBalanceAmount(selectedRechargeBonus) }}</dd>
                            </div>
                          </dl>

                          <dl class="py-3">
                            <div class="mb-3 flex items-baseline justify-between gap-3">
                              <dt class="font-medium text-foreground">{{ t('payment.paymentSummary', { method: selectedMethodLabel }) }}</dt>
                              <dd class="text-xl font-semibold tabular-nums text-foreground">{{ formatGatewayAmount(totalAmount) }}</dd>
                            </div>
                            <div class="flex items-center justify-between gap-3 text-foreground-subtle">
                              <dt>{{ t('payment.paymentAmount') }}</dt>
                              <dd class="tabular-nums">{{ formatGatewayAmount(validAmount) }}</dd>
                            </div>
                            <div v-if="feeRate > 0" class="mt-2 flex items-center justify-between gap-3 text-foreground-subtle">
                              <dt>{{ t('payment.channelFee') }} ({{ feeRate }}%)</dt>
                              <dd class="tabular-nums">+{{ formatGatewayAmount(feeAmount) }}</dd>
                            </div>
                          </dl>
                        </div>
                        <button
                          type="button"
                          data-testid="recharge-confirm-action"
                          class="payment-confirm-button btn btn-lg mt-4 hidden w-full lg:inline-flex"
                          :disabled="!canSubmit || submitting"
                          :aria-busy="submitting"
                          @click="handleSubmitRecharge"
                        >
                          <span v-if="submitting" class="flex items-center justify-center gap-2">
                            <span class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"></span>
                            {{ t('common.processing') }}
                          </span>
                          <span v-else>{{ t('payment.payAmount', { amount: formatGatewayAmount(totalAmount) }) }}</span>
                        </button>
                      </aside>
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
                  <Transition name="subscription-stage" mode="out-in">
                    <div v-if="subscriptionStage === 'plans'" key="subscription-plans" data-testid="subscription-plan-catalog">
                      <header class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                        <h2 class="text-base font-semibold text-foreground">{{ t('payment.selectPlan') }}</h2>
                        <RouterLink to="/subscriptions" class="btn btn-secondary btn-sm shrink-0">
                          <span>{{ t('userSubscriptions.title') }}</span>
                          <Icon name="arrowRight" size="sm" aria-hidden="true" />
                        </RouterLink>
                      </header>

                      <div v-if="checkout.plans.length === 0" class="mt-4">
                        <EmptyState :title="t('payment.noPlans')" />
                      </div>

                      <div v-else data-testid="subscription-plan-grid" :class="['plan-grid mt-4', planGridClass]">
                        <SubscriptionPlanCard
                          v-for="plan in checkout.plans"
                          :key="plan.id"
                          :plan="plan"
                          :selected="selectedPlan?.id === plan.id"
                          @select="selectPlan"
                        />
                      </div>
                    </div>

                    <div
                      v-else-if="selectedPlan"
                      key="subscription-checkout"
                      ref="subscriptionCheckoutRef"
                      data-testid="subscription-checkout"
                      class="scroll-mt-24 outline-none"
                      tabindex="-1"
                    >
                      <header class="flex min-w-0 items-center justify-between gap-3 border-b border-outline pb-4">
                        <div class="flex min-w-0 items-center gap-3">
                          <button
                            type="button"
                            data-testid="subscription-back-action"
                            class="btn btn-ghost btn-icon shrink-0"
                            :aria-label="t('payment.backToPlans')"
                            @click="backToPlanSelection"
                          >
                            <Icon name="arrowLeft" size="sm" aria-hidden="true" />
                          </button>
                          <div class="min-w-0">
                            <h2 class="text-base font-semibold text-foreground">{{ t('payment.confirmSubscription') }}</h2>
                            <p class="mt-0.5 truncate text-xs text-foreground-subtle">{{ selectedPlan.name }}</p>
                          </div>
                        </div>
                        <RouterLink to="/subscriptions" class="btn btn-ghost btn-sm hidden shrink-0 sm:inline-flex">
                          <span>{{ t('userSubscriptions.title') }}</span>
                          <Icon name="arrowRight" size="sm" aria-hidden="true" />
                        </RouterLink>
                      </header>

                      <div class="grid min-w-0 gap-5 pt-5 lg:grid-cols-[minmax(0,1fr)_minmax(340px,0.82fr)]">
                        <section class="min-w-0" :aria-label="t('payment.selectedPlanSummary')">
                          <div class="flex min-w-0 items-start justify-between gap-3">
                            <div class="min-w-0">
                              <p class="text-sm font-medium text-foreground-subtle">{{ t('payment.selectedPlanSummary') }}</p>
                              <h3 class="mt-1 break-words text-xl font-semibold text-foreground">{{ selectedPlan.name }}</h3>
                            </div>
                            <span :class="['badge shrink-0', planBadgeClass]">{{ platformLabel(selectedPlan.group_platform || '') }}</span>
                          </div>

                          <div class="mt-4 flex flex-wrap items-baseline gap-x-2 gap-y-1">
                            <span :class="['text-3xl font-semibold tabular-nums', planTextClass]">{{ formatSelectedSubscriptionPaymentAmount(selectedPlan.price) }}</span>
                            <span class="text-sm text-foreground-subtle">/ {{ planValiditySuffix }}</span>
                            <span v-if="selectedPlan.original_price" class="text-sm text-foreground-subtle line-through">{{ formatSelectedSubscriptionPaymentAmount(selectedPlan.original_price) }}</span>
                          </div>
                          <p v-if="selectedPlan.description" class="mt-2 text-sm leading-6 text-foreground-subtle">{{ selectedPlan.description }}</p>

                          <dl class="mt-5 grid gap-x-5 gap-y-3 border-y border-outline py-4 text-sm sm:grid-cols-2">
                            <div class="flex items-center justify-between gap-3">
                              <dt class="text-foreground-subtle">{{ t('payment.planCard.rate') }}</dt>
                              <dd class="font-semibold text-foreground">{{ formatPlanRate(selectedPlan.rate_multiplier) }}</dd>
                            </div>
                            <div v-if="selectedPlan.daily_limit_usd != null" class="flex items-center justify-between gap-3">
                              <dt class="text-foreground-subtle">{{ t('payment.planCard.dailyLimit') }}</dt>
                              <dd class="font-semibold tabular-nums text-foreground">{{ formatPlanQuota(selectedPlan.daily_limit_usd) }}</dd>
                            </div>
                            <div v-if="selectedPlan.weekly_limit_usd != null" class="flex items-center justify-between gap-3">
                              <dt class="text-foreground-subtle">{{ t('payment.planCard.weeklyLimit') }}</dt>
                              <dd class="font-semibold tabular-nums text-foreground">{{ formatPlanQuota(selectedPlan.weekly_limit_usd) }}</dd>
                            </div>
                            <div v-if="selectedPlan.monthly_limit_usd != null" class="flex items-center justify-between gap-3">
                              <dt class="text-foreground-subtle">{{ t('payment.planCard.monthlyLimit') }}</dt>
                              <dd class="font-semibold tabular-nums text-foreground">{{ formatPlanQuota(selectedPlan.monthly_limit_usd) }}</dd>
                            </div>
                          </dl>

                          <ul v-if="selectedPlan.features.length > 0" class="mt-4 space-y-2 text-sm leading-5 text-foreground-muted">
                            <li v-for="feature in selectedPlan.features" :key="feature" class="flex items-start gap-2">
                              <Icon name="check" size="sm" class="mt-0.5 shrink-0 text-success-foreground" aria-hidden="true" />
                              <span>{{ feature }}</span>
                            </li>
                          </ul>
                        </section>

                        <aside class="min-w-0 rounded-panel border border-outline bg-surface-subtle p-4 sm:p-5 lg:sticky lg:top-24 lg:self-start" :aria-label="t('payment.checkoutSummary')">
                          <PaymentMethodSelector
                            compact
                            :methods="subMethodOptions"
                            :selected="selectedMethod"
                            @select="selectedMethod = $event"
                          />
                          <dl class="mt-5 space-y-3 border-y border-outline py-4 text-sm">
                            <div class="flex items-center justify-between gap-4">
                              <dt class="text-foreground-subtle">{{ t('payment.planAmount') }}</dt>
                              <dd class="font-medium tabular-nums text-foreground">{{ formatSelectedPaymentAmount(subPaymentAmount) }}</dd>
                            </div>
                            <div v-if="feeRate > 0 && selectedPlan.price > 0" class="flex items-center justify-between gap-4">
                              <dt class="text-foreground-subtle">{{ t('payment.fee') }} ({{ feeRate }}%)</dt>
                              <dd class="font-medium tabular-nums text-foreground">+{{ formatSelectedPaymentAmount(subFeeAmount) }}</dd>
                            </div>
                          </dl>
                          <dl class="mt-4">
                            <dt class="text-sm font-medium text-foreground-subtle">{{ t('payment.actualPay') }}</dt>
                            <dd class="mt-1 text-3xl font-semibold tabular-nums text-foreground">{{ formatSelectedPaymentAmount(subTotalAmount) }}</dd>
                          </dl>
                          <button
                            type="button"
                            data-testid="subscription-confirm-action"
                            class="payment-confirm-button btn btn-lg mt-4 hidden w-full lg:inline-flex"
                            :disabled="!canSubmitSubscription || submitting"
                            :aria-busy="submitting"
                            @click="confirmSubscribe"
                          >
                            <span v-if="submitting" class="flex items-center justify-center gap-2">
                              <span class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"></span>
                              {{ t('common.processing') }}
                            </span>
                            <span v-else>{{ t('payment.payAmount', { amount: formatSelectedPaymentAmount(subTotalAmount) }) }}</span>
                          </button>
                        </aside>
                      </div>
                    </div>
                  </Transition>
                </section>

              </article>

              <aside data-testid="purchase-side-tools" class="hidden space-y-4 min-[1600px]:sticky min-[1600px]:top-24 min-[1600px]:block">
                <AffiliateRewardPanel
                  v-if="affiliateEnabled"
                  data-testid="affiliate-reward-panel"
                />
                <section data-testid="purchase-redeem-card" class="card min-w-0 p-5" :aria-label="t('redeem.redeemCodeLabel')">
                  <RedeemCodeForm show-description />
                </section>
              </aside>
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

    <div
      v-if="showMobileRechargeBar"
      class="fixed inset-x-0 bottom-0 z-40 border-t border-outline bg-surface-raised/95 px-3 pt-2 shadow-floating backdrop-blur-md lg:hidden"
      style="padding-bottom: max(0.75rem, env(safe-area-inset-bottom))"
    >
      <div class="mx-auto grid max-w-lg grid-cols-[minmax(0,1fr)_minmax(145px,auto)] items-center gap-3">
        <dl class="grid min-w-0 grid-cols-2 gap-3">
          <div class="min-w-0">
            <dt class="text-[10px] leading-4 text-foreground-subtle">{{ t('payment.balanceSummary') }}</dt>
            <dd class="truncate text-sm font-semibold tabular-nums text-foreground">{{ formatBalanceAmount(creditedAmount) }}</dd>
          </div>
          <div class="min-w-0 border-l border-outline pl-3">
            <dt class="text-[10px] leading-4 text-foreground-subtle">{{ t('payment.actualPay') }}</dt>
            <dd class="truncate text-sm font-semibold tabular-nums text-foreground">{{ formatGatewayAmount(totalAmount) }}</dd>
          </div>
        </dl>
        <button
          type="button"
          data-testid="mobile-recharge-confirm-action"
          class="payment-confirm-button btn h-11 w-full px-3 text-sm"
          :disabled="!canSubmit || submitting"
          :aria-busy="submitting"
          @click="handleSubmitRecharge"
        >
          <span v-if="submitting">{{ t('common.processing') }}</span>
          <span v-else>{{ t('payment.mobilePay', { amount: formatGatewayAmount(totalAmount) }) }}</span>
        </button>
      </div>
    </div>

    <div
      v-if="showMobileSubscriptionBar"
      class="fixed inset-x-0 bottom-0 z-40 border-t border-outline bg-surface-raised/95 px-3 pt-2 shadow-floating backdrop-blur-md lg:hidden"
      style="padding-bottom: max(0.75rem, env(safe-area-inset-bottom))"
    >
      <div class="mx-auto grid max-w-lg grid-cols-[minmax(0,1fr)_minmax(155px,auto)] items-center gap-3">
        <dl class="min-w-0">
          <dt class="truncate text-[10px] leading-4 text-foreground-subtle">{{ selectedPlan?.name }}</dt>
          <dd class="truncate text-base font-semibold tabular-nums text-foreground">{{ formatSelectedPaymentAmount(subTotalAmount) }}</dd>
        </dl>
        <button
          type="button"
          data-testid="mobile-subscription-confirm-action"
          class="payment-confirm-button btn h-11 w-full px-3 text-sm"
          :disabled="!canSubmitSubscription || submitting"
          :aria-busy="submitting"
          @click="confirmSubscribe"
        >
          <span v-if="submitting">{{ t('common.processing') }}</span>
          <span v-else>{{ t('payment.payAmount', { amount: formatSelectedPaymentAmount(subTotalAmount) }) }}</span>
        </button>
      </div>
    </div>

    <PurchaseAuxiliaryDrawer
      :mode="auxiliaryMode"
      :affiliate-enabled="affiliateEnabled"
      @close="auxiliaryMode = null"
    />

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
import { ref, computed, nextTick, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink, useRoute, useRouter, type LocationQueryRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePaymentStore } from '@/stores/payment'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
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
import PurchaseAuxiliaryDrawer, { type PurchaseAuxiliaryMode } from '@/components/payment/PurchaseAuxiliaryDrawer.vue'
import RechargeValueEstimator from '@/components/payment/RechargeValueEstimator.vue'
import PaymentMethodSelector from '@/components/payment/PaymentMethodSelector.vue'
import RedeemCodeForm from '@/components/payment/RedeemCodeForm.vue'
import { METHOD_ORDER, getPaymentPopupFeatures } from '@/components/payment/providerConfig'
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
const BALANCE_CURRENCY = 'USD'
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const paymentStore = usePaymentStore()
const subscriptionStore = useSubscriptionStore()
const appStore = useAppStore()

const loading = ref(true)
const checkoutLoadFailed = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const errorHintMessage = ref('')
type PurchaseTab = 'recharge' | 'subscription'
type SubscriptionStage = 'plans' | 'checkout'

const activeTab = ref<PurchaseTab>('recharge')
const amount = ref<number | null>(null)
const selectedMethod = ref('')
const selectedPlan = ref<SubscriptionPlan | null>(null)
const subscriptionStage = ref<SubscriptionStage>('plans')
const subscriptionCheckoutRef = ref<HTMLElement | null>(null)
const methodSwitchNotice = ref('')
const previewImage = ref('')
const auxiliaryMode = ref<PurchaseAuxiliaryMode | null>(null)

const paymentPhase = ref<'select' | 'paying'>('select')

const affiliateEnabled = computed(() => appStore.cachedPublicSettings?.affiliate_enabled === true)
const modelMarketplaceEnabled = computed(() => appStore.cachedPublicSettings?.model_marketplace_enabled === true)

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
    subscriptionStage.value = 'plans'
    const query: LocationQueryRaw = { ...route.query, tab: 'subscription' }
    delete query.plan
    delete query.group
    void router.replace({ path: route.path, query })
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
const defaultQuickRechargeAmounts = [10, 20, 50, 100, 200, 500, 1000, 2000, 5000]
  .map(amount => ({ amount, bonus: 0 }))

const checkout = ref<CheckoutInfoResponse>({
  methods: {}, global_min: 0, global_max: 0,
  plans: [], instant_enabled: true, balance_disabled: false, balance_recharge_multiplier: 1,
  quick_recharge_amounts: defaultQuickRechargeAmounts,
  custom_recharge_amount_enabled: true,
  subscription_usd_to_cny_rate: 0, recharge_fee_rate: 0, help_text: '', help_image_url: '', stripe_publishable_key: '',
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

const pageContainerClass = computed(() => [
  'mx-auto w-full space-y-5',
  activeTab.value === 'recharge' || (activeTab.value === 'subscription' && subscriptionStage.value === 'checkout')
    ? 'max-w-[1320px] pb-24 lg:pb-0'
    : 'max-w-[1320px]',
])

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
const showMobileRechargeBar = computed(() =>
  paymentPhase.value === 'select'
  && activeTab.value === 'recharge'
  && tabs.value.length > 0
  && enabledMethods.value.length > 0
)
const showMobileSubscriptionBar = computed(() =>
  paymentPhase.value === 'select'
  && activeTab.value === 'subscription'
  && subscriptionStage.value === 'checkout'
  && selectedPlan.value !== null
  && enabledMethods.value.length > 0
)
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
const selectedRechargeBonus = computed(() => {
  const cents = Math.round(validAmount.value * 100)
  return checkout.value.quick_recharge_amounts.find(item => Math.round(item.amount * 100) === cents)?.bonus ?? 0
})
const creditedAmount = computed(() => Math.round((validAmount.value * balanceRechargeMultiplier.value + selectedRechargeBonus.value) * 100) / 100)
const rechargeAmountAllowed = computed(() =>
  checkout.value.custom_recharge_amount_enabled
    || checkout.value.quick_recharge_amounts.some(item => Math.round(item.amount * 100) === Math.round(validAmount.value * 100))
)

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

function formatBalanceAmount(value: number): string {
  return formatPaymentAmount(value, BALANCE_CURRENCY, localeCode.value)
}

function formatGatewayAmount(value: number, currency = selectedCurrency.value): string {
  return formatPaymentAmount(value, currency, localeCode.value)
}

function formatSelectedPaymentAmount(value: number): string {
  return formatGatewayAmount(value)
}

function formatSelectedSubscriptionPaymentAmount(value: number): string {
  return formatSelectedPaymentAmount(subscriptionPaymentAmountForCurrency(value, selectedCurrency.value))
}

function formatPlanRate(value?: number): string {
  const rate = Number.isFinite(value) ? Number(value) : 1
  return `${Number(rate.toPrecision(10))}x`
}

function formatPlanQuota(value: number): string {
  return formatPaymentAmount(value, 'USD', localeCode.value)
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

const selectedMethodLabel = computed(() => {
  const option = methodOptions.value.find(method => method.type === selectedMethod.value)
  if (option?.display_name) return option.display_name
  const key = `payment.methods.${normalizeVisibleMethod(selectedMethod.value) || selectedMethod.value}`
  const translated = t(key)
  return translated === key ? selectedMethod.value : translated
})

const feeRate = computed(() => selectedLimit.value?.fee_rate ?? checkout.value?.recharge_fee_rate ?? 0)
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
  if (!rechargeAmountAllowed.value) return t('payment.rechargeAmountNotAllowed')
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
    && rechargeAmountAllowed.value
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

watch(validAmount, (amt) => {
  methodSwitchNotice.value = ''
  if (amt <= 0 || amountFitsMethod(amt, selectedMethod.value)) return

  const previousMethod = selectedMethod.value
  const available = enabledMethods.value.find((m) => amountFitsMethod(amt, m))
  if (!available) return

  selectedMethod.value = available
  methodSwitchNotice.value = t('payment.methodAutoSwitched', {
    from: paymentMethodDisplayName(previousMethod),
    to: paymentMethodDisplayName(available),
  })
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

function paymentMethodDisplayName(type: string): string {
  return visibleMethods.value[type]?.display_name || t(`payment.methods.${type}`, type)
}

async function enterSubscriptionCheckout(plan: SubscriptionPlan, syncRoute = true) {
  selectedPlan.value = plan
  subscriptionStage.value = 'checkout'
  errorMessage.value = ''

  if (syncRoute) {
    const query: LocationQueryRaw = { ...route.query, tab: 'subscription', plan: String(plan.id) }
    delete query.group
    await router.push({ path: route.path, query })
  }

  await nextTick()
  const target = subscriptionCheckoutRef.value
  if (!target) return
  const reduceMotion = typeof window !== 'undefined'
    && window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
  target.scrollIntoView?.({ behavior: reduceMotion ? 'auto' : 'smooth', block: 'start' })
  target.focus({ preventScroll: true })
}

async function selectPlan(plan: SubscriptionPlan) {
  await enterSubscriptionCheckout(plan)
}

async function backToPlanSelection() {
  subscriptionStage.value = 'plans'
  errorMessage.value = ''
  const query: LocationQueryRaw = { ...route.query, tab: 'subscription' }
  delete query.plan
  delete query.group
  await router.push({ path: route.path, query })
}

async function selectPlanFromModal(plan: SubscriptionPlan) {
  showRenewalModal.value = false
  renewGroupId.value = null
  await enterSubscriptionCheckout(plan)
}

function closeRenewalModal() {
  showRenewalModal.value = false
  renewGroupId.value = null
}

function syncSubscriptionStageFromRoute(): void {
  if (route.query.tab !== 'subscription' || paymentPhase.value !== 'select') return

  const planId = Number(route.query.plan)
  if (Number.isFinite(planId) && planId > 0) {
    const plan = checkout.value.plans.find(item => item.id === planId)
    if (plan) {
      selectedPlan.value = plan
      subscriptionStage.value = 'checkout'
      return
    }
  }

  const groupId = Number(route.query.group)
  if (Number.isFinite(groupId) && groupId > 0) {
    const groupPlans = checkout.value.plans.filter(plan => plan.group_id === groupId)
    if (groupPlans.length === 1) {
      selectedPlan.value = groupPlans[0]
      subscriptionStage.value = 'checkout'
      return
    }
  }

  subscriptionStage.value = 'plans'
}

watch(
  () => [route.query.tab, route.query.plan, route.query.group],
  syncSubscriptionStageFromRoute,
)

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
    if (selectedPlan.value) subscriptionStage.value = 'checkout'
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
  try {
    checkoutLoadFailed.value = false
    const res = await paymentAPI.getCheckoutInfo()
    checkout.value = {
      ...res.data,
      quick_recharge_amounts: Array.isArray(res.data.quick_recharge_amounts)
        ? res.data.quick_recharge_amounts
        : defaultQuickRechargeAmounts,
      custom_recharge_amount_enabled: res.data.custom_recharge_amount_enabled !== false,
    }
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
    // Handle direct checkout navigation: ?tab=subscription&plan=123 or renewal ?group=123.
    if (route.query.tab === 'subscription') {
      ensureVisiblePurchaseTab('subscription')
      const planId = Number(route.query.plan)
      const routePlan = Number.isFinite(planId) && planId > 0
        ? checkout.value.plans.find(plan => plan.id === planId)
        : undefined
      if (routePlan) {
        await enterSubscriptionCheckout(routePlan, false)
      } else if (route.query.group) {
        const groupId = Number(route.query.group)
        const groupPlans = checkout.value.plans.filter(p => p.group_id === groupId)
        if (groupPlans.length === 1) {
          await enterSubscriptionCheckout(groupPlans[0], false)
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
.purchase-tabs {
  display: flex;
  gap: 3px;
  padding: 3px;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 7px;
  background: var(--ui-surface, #fff);
}

.purchase-tab {
  position: relative;
  min-height: 36px;
  padding: 6px 14px;
  border: 0;
  border-radius: 5px;
  background: transparent;
  color: var(--ui-text-muted, #667085);
  font-size: 14px;
  font-weight: 600;
  outline: none;
  transition: color 150ms ease, background-color 150ms ease, box-shadow 150ms ease;
}

.purchase-tab:hover {
  color: var(--ui-text, #0f172a);
  background: var(--ui-surface-subtle, #f4f7fb);
}

.purchase-tab:focus-visible {
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--ui-focus, #2563eb) 16%, transparent);
}

.purchase-tab-active {
  color: var(--ui-focus, #2563eb);
  background: color-mix(in srgb, var(--ui-focus, #2563eb) 9%, var(--ui-surface, #fff));
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--ui-focus, #2563eb) 22%, transparent);
}

.purchase-tab-active::after {
  position: absolute;
  right: 50%;
  bottom: 3px;
  width: 26px;
  height: 2px;
  border-radius: 999px;
  background: var(--ui-focus, #2563eb);
  content: '';
  transform: translateX(50%);
}

.payment-confirm-button {
  color: rgb(var(--color-brand-foreground));
  background: rgb(var(--color-brand));
  border-color: rgb(var(--color-brand));
  box-shadow: var(--shadow-card);
}

.payment-confirm-button:hover:not(:disabled) {
  background: rgb(var(--color-brand-hover));
  border-color: rgb(var(--color-brand-hover));
}

.subscription-stage-enter-active,
.subscription-stage-leave-active {
  transition: opacity 160ms ease, transform 160ms ease;
}

.subscription-stage-enter-from {
  opacity: 0;
  transform: translateX(10px);
}

.subscription-stage-leave-to {
  opacity: 0;
  transform: translateX(-8px);
}

@media (prefers-reduced-motion: reduce) {
  .subscription-stage-enter-active,
  .subscription-stage-leave-active {
    transition: none;
  }
}

.plan-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 250px), 1fr));
  gap: 1rem;
}
</style>
