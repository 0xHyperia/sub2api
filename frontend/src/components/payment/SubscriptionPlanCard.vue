<template>
  <article
    class="group flex min-w-0 flex-col rounded-panel border border-outline bg-surface p-4 shadow-card transition-colors hover:border-outline-strong"
  >
    <div class="flex min-w-0 items-start justify-between gap-3">
      <div class="min-w-0">
        <div class="flex min-w-0 flex-wrap items-center gap-2">
          <h3 class="break-words text-base font-semibold text-foreground">{{ plan.name }}</h3>
          <span :class="['badge shrink-0', badgeLightClass]">
            {{ pLabel }}
          </span>
          <span v-if="isRenewal" class="badge badge-success shrink-0">
            {{ t('payment.renewNow') }}
          </span>
        </div>
        <p v-if="plan.description" class="plan-description mt-1.5 break-words text-sm leading-5 text-foreground-subtle">
          {{ plan.description }}
        </p>
      </div>
    </div>

    <div class="mt-3 flex flex-wrap items-end gap-x-2.5 gap-y-1">
      <div class="flex items-baseline gap-1">
        <span class="text-sm text-foreground-subtle">$</span>
        <span :class="['text-2xl font-semibold tabular-nums', textClass]">{{ plan.price }}</span>
        <span v-if="plan.currency" class="text-xs font-medium text-foreground-subtle">{{ plan.currency }}</span>
      </div>
      <span class="text-sm text-foreground-subtle">/ {{ validitySuffix }}</span>
      <span v-if="plan.original_price" class="text-sm text-foreground-subtle line-through">
        \${{ plan.original_price }}<template v-if="plan.currency"> {{ plan.currency }}</template>
      </span>
      <span v-if="discountText" :class="['badge', discountClass]">{{ discountText }}</span>
    </div>

    <dl class="mt-4 grid grid-cols-1 gap-x-4 gap-y-2 border-t border-outline pt-3 text-sm min-[420px]:grid-cols-2">
      <div class="flex items-center justify-between gap-3">
        <dt class="text-xs text-foreground-subtle">{{ t('payment.planCard.rate') }}</dt>
        <dd class="font-semibold text-foreground">{{ rateDisplay }}</dd>
      </div>
      <div v-if="hasPeakRate" class="flex items-center justify-between gap-3">
        <dt class="text-xs text-foreground-subtle">{{ t('payment.planCard.peakRate') }}</dt>
        <dd class="font-semibold text-warning-foreground">{{ peakRateDisplay }}</dd>
      </div>
      <div v-if="plan.daily_limit_usd != null" class="flex items-center justify-between gap-3">
        <dt class="text-xs text-foreground-subtle">{{ t('payment.planCard.dailyLimit') }}</dt>
        <dd class="font-semibold tabular-nums text-foreground">\${{ plan.daily_limit_usd }}</dd>
      </div>
      <div v-if="plan.weekly_limit_usd != null" class="flex items-center justify-between gap-3">
        <dt class="text-xs text-foreground-subtle">{{ t('payment.planCard.weeklyLimit') }}</dt>
        <dd class="font-semibold tabular-nums text-foreground">\${{ plan.weekly_limit_usd }}</dd>
      </div>
      <div v-if="plan.monthly_limit_usd != null" class="flex items-center justify-between gap-3">
        <dt class="text-xs text-foreground-subtle">{{ t('payment.planCard.monthlyLimit') }}</dt>
        <dd class="font-semibold tabular-nums text-foreground">\${{ plan.monthly_limit_usd }}</dd>
      </div>
      <div v-if="plan.daily_limit_usd == null && plan.weekly_limit_usd == null && plan.monthly_limit_usd == null" class="flex items-center justify-between gap-3">
        <dt class="text-xs text-foreground-subtle">{{ t('payment.planCard.quota') }}</dt>
        <dd class="font-semibold text-foreground">{{ t('payment.planCard.unlimited') }}</dd>
      </div>
    </dl>

    <div v-if="modelScopeLabels.length > 0" class="mt-3">
      <p class="text-xs text-foreground-subtle">{{ t('payment.planCard.models') }}</p>
      <div class="mt-2 flex flex-wrap gap-1.5">
        <span v-for="scope in modelScopeLabels" :key="scope" class="badge badge-gray">
          {{ scope }}
        </span>
      </div>
    </div>

    <ul v-if="plan.features.length > 0" class="mt-3 space-y-1.5 text-sm leading-5 text-foreground-muted">
      <li v-for="feature in plan.features" :key="feature" class="flex items-start gap-2">
        <Icon name="check" size="sm" :class="['mt-0.5 shrink-0', iconClass]" aria-hidden="true" />
        <span class="break-words">{{ feature }}</span>
      </li>
    </ul>

    <div class="mt-4 flex-1" />

    <button
      type="button"
      class="btn btn-primary btn-sm w-full"
      @click="emit('select', plan)"
    >
      {{ isRenewal ? t('payment.renewNow') : t('payment.subscribeNow') }}
    </button>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'
import { useAppStore } from '@/stores/app'
import Icon from '@/components/icons/Icon.vue'
import { hasPeakRate as groupHasPeakRate, formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'
import {
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

const badgeLightClass = computed(() => platformBadgeLightClass(platform.value))
const textClass = computed(() => platformTextClass(platform.value))
const iconClass = computed(() => platformIconClass(platform.value))
const discountClass = computed(() => platformDiscountClass(platform.value))
const pLabel = computed(() => platformLabel(platform.value))

const discountText = computed(() => {
  if (!props.plan.original_price || props.plan.original_price <= 0) return ''
  const pct = Math.round((1 - props.plan.price / props.plan.original_price) * 100)
  return pct > 0 ? '-' + pct + '%' : ''
})

const rateDisplay = computed(() => {
  const rate = props.plan.rate_multiplier ?? 1
  return 'x' + Number(rate.toPrecision(10))
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
  return String(props.plan.validity_days) + t('payment.days')
})
</script>

<style scoped>
.plan-description {
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}
</style>
