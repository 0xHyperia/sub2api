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
          :class="[
            'quick-amount-tile flex min-h-[48px] items-center justify-center rounded-control border px-2.5 py-2 text-center transition-colors',
            modelValue === item.amount
              ? 'border-focus bg-info-subtle text-info-foreground shadow-card'
              : 'border-outline bg-surface text-foreground hover:border-outline-strong hover:bg-surface-subtle',
          ]"
          @click="selectAmount(item.amount)"
        >
          <span class="flex min-w-0 flex-col items-center justify-center gap-0.5">
            <span class="text-sm font-semibold tabular-nums">{{ formatAmount(item.amount) }}</span>
            <span v-if="item.bonus > 0" class="text-[11px] font-medium leading-4 text-success-foreground">
              {{ t('payment.quickAmountBonus', { bonus: item.bonus.toFixed(2) }) }}
            </span>
          </span>
        </button>
      </div>
    </fieldset>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { QuickRechargeAmount } from '@/types/payment'

const props = withDefaults(defineProps<{
  amounts?: QuickRechargeAmount[]
  modelValue: number | null
  min?: number
  max?: number
  customEnabled?: boolean
  currency?: string
}>(), {
  amounts: () => [10, 20, 50, 100, 200, 500, 1000, 2000, 5000].map(amount => ({ amount, bonus: 0 })),
  min: 0,
  max: 0,
  customEnabled: true,
  currency: 'USD',
})

const emit = defineEmits<{
  'update:modelValue': [value: number | null]
}>()

const { t } = useI18n()

const customText = ref('')

const filteredAmounts = computed(() =>
  props.amounts.filter((item) => (props.min <= 0 || item.amount >= props.min) && (props.max <= 0 || item.amount <= props.max))
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
