<template>
  <div class="space-y-4">
    <div v-if="customEnabled">
      <label for="custom-payment-amount" class="mb-2 block text-sm font-medium text-foreground">
        {{ t('payment.customAmount') }}
      </label>
      <div class="relative">
        <span class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-lg font-semibold text-foreground-subtle" aria-hidden="true">
          {{ currencySymbol }}
        </span>
        <input
          id="custom-payment-amount"
          type="text"
          inputmode="decimal"
          :value="customText"
          :placeholder="placeholderText"
          class="input min-h-[52px] w-full pl-10 pr-4 text-lg font-semibold tabular-nums"
          @input="handleInput"
        />
      </div>
    </div>

    <fieldset v-if="filteredAmounts.length" class="space-y-3">
      <legend class="text-sm font-medium text-foreground">
        {{ t('payment.quickAmounts') }}
      </legend>
      <div class="grid grid-cols-3 gap-2 min-[520px]:grid-cols-5">
        <button
          v-for="item in filteredAmounts"
          :key="item.amount"
          type="button"
          :aria-pressed="modelValue === item.amount"
          :data-testid="`quick-amount-${item.amount}`"
          :class="[
            'quick-amount-tile relative flex min-h-[48px] items-center justify-center rounded-control border px-2.5 py-2 text-center transition-colors',
            modelValue === item.amount
              ? 'border-focus bg-info-subtle text-info-foreground shadow-card'
              : quoteFor(item.amount).percent > 0
                ? 'border-danger/40 bg-surface text-foreground hover:border-danger hover:bg-surface-subtle'
                : 'border-outline bg-surface text-foreground hover:border-outline-strong hover:bg-surface-subtle',
          ]"
          @click="selectAmount(item.amount)"
        >
          <!-- 促销价签：仅命中优惠阶梯的金额显示 -->
          <span
            v-if="quoteFor(item.amount).percent > 0"
            class="pointer-events-none absolute -right-1.5 -top-2.5 z-10 rotate-12"
            data-testid="quick-amount-bonus-badge"
          >
            <span class="flex items-center gap-1 whitespace-nowrap rounded-control bg-danger px-1.5 py-0.5 text-[11px] font-extrabold leading-tight text-danger-solid-foreground shadow-card">{{ badgeText(item.amount) }}</span>
          </span>
          <span class="flex min-w-0 flex-col items-center justify-center gap-0.5">
            <span class="text-sm font-semibold tabular-nums">{{ formatAmount(item.amount) }}</span>
            <!-- 配置了优惠阶梯时所有按钮都显示第二行以保持高度一致：赠金显示到账 USD，折扣显示折后实付 -->
            <span
              v-if="showSecondLine"
              :class="[
                'text-[11px] font-normal leading-4',
                quoteFor(item.amount).percent > 0 ? 'text-danger-foreground' : 'text-foreground-subtle',
              ]"
              data-testid="quick-amount-credited"
            >{{ secondLine(item.amount) }}</span>
          </span>
        </button>
      </div>
    </fieldset>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { QuickRechargeAmount, RechargeBonusTier } from '@/types/payment'
import { formatRechargeBonusNumber, quoteRechargeBonus, type RechargeBonusMode } from '@/utils/rechargeBonus'
import { formatPaymentAmount } from './currency'

const props = withDefaults(defineProps<{
  /** 快捷金额；兼容纯数字与带固定赠送金额的对象 */
  amounts?: Array<QuickRechargeAmount | number>
  modelValue: number | null
  min?: number
  max?: number
  customEnabled?: boolean
  /** 充值优惠阶梯（按 min_amount 升序）；为空时不显示价签与第二行 */
  bonusTiers?: RechargeBonusTier[]
  /** 阶梯模式：bonus 赠金 / discount 折扣 */
  bonusMode?: RechargeBonusMode
  /** 充值倍率（1 支付币种 = multiplier USD），用于计算到账金额 */
  multiplier?: number
  /** 支付币种（货币符号、折扣模式第二行实付金额的币种与精度） */
  currency?: string
}>(), {
  amounts: () => [10, 20, 50, 100, 200, 500, 1000, 2000, 5000].map(amount => ({ amount })),
  min: 0,
  max: 0,
  customEnabled: true,
  bonusTiers: () => [],
  bonusMode: 'bonus',
  multiplier: 1,
  currency: 'USD',
})

const emit = defineEmits<{
  'update:modelValue': [value: number | null]
}>()

const { t } = useI18n()

const customText = ref('')

const filteredAmounts = computed<QuickRechargeAmount[]>(() =>
  props.amounts
    .map((item) => (typeof item === 'number' ? { amount: item } : item))
    .filter((item) => (props.min <= 0 || item.amount >= props.min) && (props.max <= 0 || item.amount <= props.max))
)

const currencySymbol = computed(() => {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency: props.currency,
      currencyDisplay: 'narrowSymbol',
    }).formatToParts(0).find(part => part.type === 'currency')?.value || props.currency
  } catch {
    return props.currency
  }
})

function formatAmount(value: number): string {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency: props.currency,
      currencyDisplay: 'narrowSymbol',
      minimumFractionDigits: Number.isInteger(value) ? 0 : 2,
      maximumFractionDigits: 2,
    }).format(value)
  } catch {
    return `${currencySymbol.value}${value}`
  }
}

const showSecondLine = computed(() => props.bonusTiers.length > 0)

function currencyDigits(): number {
  if (!props.currency) return 2
  try {
    return new Intl.NumberFormat(undefined, { style: 'currency', currency: props.currency }).resolvedOptions().maximumFractionDigits ?? 2
  } catch {
    return 2
  }
}

function quoteFor(amt: number) {
  return quoteRechargeBonus(props.bonusTiers, amt, {
    multiplier: props.multiplier,
    mode: props.bonusMode,
    currencyDigits: currencyDigits(),
  })
}

// 价签文案：赠金「+20%」，折扣「20% OFF」
function badgeText(amt: number): string {
  const percent = formatRechargeBonusNumber(quoteFor(amt).percent)
  return props.bonusMode === 'discount' ? `${percent}% OFF` : `+${percent}%`
}

function secondLine(amt: number): string {
  const quote = quoteFor(amt)
  if (props.bonusMode === 'discount') {
    return t('payment.rechargeBonus.payShort', { amount: formatPaymentAmount(quote.payBase, props.currency) })
  }
  return t('payment.rechargeBonus.creditedShort', { amount: '$' + quote.credited.toFixed(2) })
}

const placeholderText = computed(() => {
  if (props.min > 0 && props.max > 0) return `${props.min} - ${props.max}`
  if (props.min > 0) return `>= ${props.min}`
  if (props.max > 0) return `<= ${props.max}`
  return t('payment.enterAmount')
})

const AMOUNT_PATTERN = /^\d*(\.\d{0,2})?$/

function selectAmount(amt: number) {
  customText.value = String(amt)
  emit('update:modelValue', amt)
}

function handleInput(e: Event) {
  const target = e.target as HTMLInputElement
  const val = target.value
  if (!AMOUNT_PATTERN.test(val)) {
    target.value = customText.value
    return
  }
  customText.value = val
  if (val === '') {
    emit('update:modelValue', null)
    return
  }
  const num = parseFloat(val)
  if (!isNaN(num) && num > 0) {
    emit('update:modelValue', num)
  } else {
    emit('update:modelValue', null)
  }
}

watch(() => props.modelValue, (v) => {
  if (v !== null && String(v) !== customText.value) {
    customText.value = String(v)
  }
}, { immediate: true })
</script>
