<template>
  <article
    class="group relative flex min-w-0 flex-col overflow-hidden rounded-panel border border-outline bg-surface shadow-card transition-colors hover:border-outline-strong"
  >
    <div :class="['h-1', accentClass]" aria-hidden="true" />

    <div class="flex min-w-0 flex-1 flex-col p-4">
      <div class="mb-3 flex min-w-0 flex-col gap-3">
        <div class="min-w-0 flex-1">
          <div class="flex min-w-0 flex-wrap items-center gap-2">
            <h3 class="break-words text-base font-semibold text-foreground">{{ plan.name }}</h3>
            <span :class="['badge shrink-0', badgeLightClass]">
              {{ pLabel }}
            </span>
          </div>
          <p v-if="plan.description" class="mt-1 break-words text-xs leading-5 text-foreground-subtle line-clamp-2">
            {{ plan.description }}
          </p>
        </div>
        <div class="shrink-0 text-left">
          <div class="flex items-baseline gap-1">
            <span class="text-xs text-foreground-subtle">$</span>
            <span :class="['text-xl font-semibold tabular-nums', textClass]">{{ plan.price }}</span>
          </div>
          <span class="text-xs text-foreground-subtle">/ {{ validitySuffix }}</span>
          <div v-if="plan.original_price" class="mt-1 flex items-center gap-1.5">
            <span class="text-xs text-foreground-subtle line-through">${{ plan.original_price }}</span>
            <span :class="['badge', discountClass]">{{ discountText }}</span>
          </div>
        </div>
      </div>

      <dl class="mb-3 grid grid-cols-2 gap-x-3 gap-y-2 border-y border-outline py-3 text-xs">
        <div class="flex min-w-0 items-center justify-between gap-2">
          <dt class="text-foreground-subtle">{{ t('payment.planCard.rate') }}</dt>
          <dd class="font-medium text-foreground-muted">{{ rateDisplay }}</dd>
        </div>
        <div v-if="hasPeakRate" class="col-span-2 flex items-center justify-between gap-2">
          <dt class="text-foreground-subtle">{{ t('payment.planCard.peakRate') }}</dt>
          <dd class="text-right font-medium text-warning-foreground">{{ peakRateDisplay }}</dd>
        </div>
        <div v-if="plan.daily_limit_usd != null" class="flex min-w-0 items-center justify-between gap-2">
          <dt class="text-foreground-subtle">{{ t('payment.planCard.dailyLimit') }}</dt>
          <dd class="font-medium tabular-nums text-foreground-muted">${{ plan.daily_limit_usd }}</dd>
        </div>
        <div v-if="plan.weekly_limit_usd != null" class="flex min-w-0 items-center justify-between gap-2">
          <dt class="text-foreground-subtle">{{ t('payment.planCard.weeklyLimit') }}</dt>
          <dd class="font-medium tabular-nums text-foreground-muted">${{ plan.weekly_limit_usd }}</dd>
        </div>
        <div v-if="plan.monthly_limit_usd != null" class="flex min-w-0 items-center justify-between gap-2">
          <dt class="text-foreground-subtle">{{ t('payment.planCard.monthlyLimit') }}</dt>
          <dd class="font-medium tabular-nums text-foreground-muted">${{ plan.monthly_limit_usd }}</dd>
        </div>
        <div v-if="plan.daily_limit_usd == null && plan.weekly_limit_usd == null && plan.monthly_limit_usd == null" class="flex min-w-0 items-center justify-between gap-2">
          <dt class="text-foreground-subtle">{{ t('payment.planCard.quota') }}</dt>
          <dd class="font-medium text-foreground-muted">{{ t('payment.planCard.unlimited') }}</dd>
        </div>
        <div v-if="modelScopeLabels.length > 0" class="col-span-2 flex items-center justify-between">
          <dt class="text-foreground-subtle">{{ t('payment.planCard.models') }}</dt>
          <dd class="flex flex-wrap justify-end gap-1">
            <span v-for="scope in modelScopeLabels" :key="scope"
              class="badge badge-gray">
              {{ scope }}
            </span>
          </dd>
        </div>
      </dl>

      <div v-if="plan.features.length > 0" class="mb-3 space-y-1">
        <div v-for="feature in plan.features" :key="feature" class="flex items-start gap-1.5">
          <svg :class="['mt-0.5 h-3.5 w-3.5 flex-shrink-0', iconClass]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
          </svg>
          <span class="break-words text-xs leading-5 text-foreground-muted">{{ feature }}</span>
        </div>
      </div>

      <div class="flex-1" />

      <button
        type="button"
        class="btn btn-primary w-full"
        @click="emit('select', plan)"
      >
        {{ isRenewal ? t('payment.renewNow') : t('payment.subscribeNow') }}
      </button>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'
import { useAppStore } from '@/stores/app'
import { hasPeakRate as groupHasPeakRate, formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'
import {
  platformAccentBarClass,
  platformBadgeLightClass,
  platformTextClass,
  platformIconClass,
  platformDiscountClass,
  platformLabel,
} from '@/utils/platformColors'

const props = defineProps<{ plan: SubscriptionPlan; activeSubscriptions?: UserSubscription[] }>()
const emit = defineEmits<{ select: [plan: SubscriptionPlan] }>()
const { t } = useI18n()

const platform = computed(() => props.plan.group_platform || '')
const isRenewal = computed(() =>
  props.activeSubscriptions?.some(s => s.group_id === props.plan.group_id && s.status === 'active') ?? false
)

// Derived color classes from central config
const accentClass = computed(() => platformAccentBarClass(platform.value))
const badgeLightClass = computed(() => platformBadgeLightClass(platform.value))
const textClass = computed(() => platformTextClass(platform.value))
const iconClass = computed(() => platformIconClass(platform.value))
const discountClass = computed(() => platformDiscountClass(platform.value))
const pLabel = computed(() => platformLabel(platform.value))

const discountText = computed(() => {
  if (!props.plan.original_price || props.plan.original_price <= 0) return ''
  const pct = Math.round((1 - props.plan.price / props.plan.original_price) * 100)
  return pct > 0 ? `-${pct}%` : ''
})

const rateDisplay = computed(() => {
  const rate = props.plan.rate_multiplier ?? 1
  return `×${Number(rate.toPrecision(10))}`
})

const appStore = useAppStore()

const hasPeakRate = computed(() => groupHasPeakRate(props.plan))

const peakRateDisplay = computed(() => {
  return formatPeakRateWindow(props.plan, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
})

const MODEL_SCOPE_LABELS: Record<string, string> = {
  claude: 'Claude',
  gemini_text: 'Gemini',
  gemini_image: 'Imagen',
}

const modelScopeLabels = computed(() => {
  if (platform.value !== 'antigravity') return []
  const scopes = props.plan.supported_model_scopes
  if (!scopes || scopes.length === 0) return []
  return scopes.map(s => MODEL_SCOPE_LABELS[s] || s)
})

const validitySuffix = computed(() => {
  const u = props.plan.validity_unit || 'day'
  if (u === 'month') return t('payment.perMonth')
  if (u === 'year') return t('payment.perYear')
  return `${props.plan.validity_days}${t('payment.days')}`
})
</script>
