<template>
  <fieldset class="space-y-2.5">
    <legend class="text-sm font-medium text-foreground">
      {{ t('payment.paymentMethod') }}
    </legend>
    <div
      :class="compact ? 'grid grid-cols-1 gap-2 min-[420px]:grid-cols-2' : 'grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3'"
      role="radiogroup"
      :aria-label="t('payment.paymentMethod')"
    >
      <button
        v-for="(method, index) in sortedMethods"
        :key="method.type"
        type="button"
        :disabled="!method.available"
        role="radio"
        :aria-checked="selected === method.type"
        :tabindex="method.available && (selected === method.type || (!selectedMethodAvailable && index === firstAvailableIndex)) ? 0 : -1"
        :class="[
          'relative flex min-w-0 items-center gap-2.5 rounded-control border px-3 py-2 text-left transition-colors',
          compact ? 'min-h-[52px]' : 'min-h-[72px] py-3',
          !method.available
            ? 'cursor-not-allowed border-outline bg-surface-subtle opacity-60'
            : selected === method.type
              ? methodSelectedClass(method.type)
              : 'border-outline bg-surface text-foreground hover:border-outline-strong hover:bg-surface-subtle',
        ]"
        @click="method.available && emit('select', method.type)"
        @keydown="handleRadioKeydown($event, index)"
      >
        <span class="flex h-8 w-8 shrink-0 items-center justify-center">
          <img :src="methodIcon(method.type)" alt="" class="h-7 w-7 object-contain" aria-hidden="true" />
        </span>
        <span class="flex min-w-0 flex-1 flex-col justify-center">
          <span class="break-words text-sm font-semibold leading-5 text-foreground">{{ methodLabel(method) }}</span>
          <span v-if="method.fee_rate > 0" class="mt-0.5 text-xs leading-4 text-foreground-subtle">
            {{ t('payment.fee') }} {{ method.fee_rate }}%
          </span>
        </span>
        <span
          data-testid="method-selection-indicator"
          :class="[
            'flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded-full border',
            selected === method.type
              ? 'border-info bg-info-subtle text-info-foreground'
              : 'border-outline-strong bg-surface text-transparent',
          ]"
          aria-hidden="true"
        >
          <span class="h-2 w-2 rounded-full bg-current"></span>
        </span>
      </button>
    </div>
  </fieldset>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { METHOD_ORDER, isBuiltInAlipayMethod, isBuiltInWxpayMethod } from './providerConfig'
import alipayIcon from '@/assets/icons/alipay.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'
import stripeIcon from '@/assets/icons/stripe.svg'
import airwallexIcon from '@/assets/icons/airwallex.svg'
import paymentIcon from '@/assets/icons/payment.svg'

export interface PaymentMethodOption {
  type: string
  display_name?: string
  fee_rate: number
  available: boolean
}

const props = defineProps<{
  methods: PaymentMethodOption[]
  selected: string
  compact?: boolean
}>()

const emit = defineEmits<{
  select: [type: string]
}>()

const { t } = useI18n()

const METHOD_ICONS: Record<string, string> = {
  alipay: alipayIcon,
  wxpay: wxpayIcon,
  stripe: stripeIcon,
  airwallex: airwallexIcon,
  credit_card: paymentIcon,
}

const sortedMethods = computed(() => {
  const order: readonly string[] = METHOD_ORDER
  return [...props.methods].sort((a, b) => {
    const ai = order.indexOf(a.type)
    const bi = order.indexOf(b.type)
    return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
  })
})

const firstAvailableIndex = computed(() => sortedMethods.value.findIndex(method => method.available))
const selectedMethodAvailable = computed(() =>
  sortedMethods.value.some(method => method.type === props.selected && method.available)
)

function handleRadioKeydown(event: KeyboardEvent, index: number): void {
  if (!['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown', 'Home', 'End'].includes(event.key)) return

  const availableIndexes = sortedMethods.value
    .map((method, methodIndex) => method.available ? methodIndex : -1)
    .filter(methodIndex => methodIndex >= 0)
  if (availableIndexes.length === 0) return

  event.preventDefault()
  const currentPosition = Math.max(0, availableIndexes.indexOf(index))
  let nextPosition = currentPosition
  if (event.key === 'Home') nextPosition = 0
  else if (event.key === 'End') nextPosition = availableIndexes.length - 1
  else if (event.key === 'ArrowLeft' || event.key === 'ArrowUp') {
    nextPosition = (currentPosition - 1 + availableIndexes.length) % availableIndexes.length
  } else {
    nextPosition = (currentPosition + 1) % availableIndexes.length
  }

  const nextIndex = availableIndexes[nextPosition]
  const nextMethod = sortedMethods.value[nextIndex]
  if (!nextMethod) return
  emit('select', nextMethod.type)
  const radioGroup = (event.currentTarget as HTMLElement).parentElement
  window.requestAnimationFrame(() => {
    radioGroup?.querySelectorAll<HTMLElement>('[role="radio"]')[nextIndex]?.focus()
  })
}

function methodIcon(type: string): string {
  if (isBuiltInAlipayMethod(type)) return METHOD_ICONS.alipay
  if (isBuiltInWxpayMethod(type)) return METHOD_ICONS.wxpay
  if (type === 'airwallex') return METHOD_ICONS.airwallex
  return METHOD_ICONS[type] || paymentIcon
}

function methodLabel(method: PaymentMethodOption): string {
  return method.display_name || t(`payment.methods.${method.type}`, method.type)
}

function methodSelectedClass(type: string): string {
  if (isBuiltInAlipayMethod(type)) return 'border-info/30 bg-info-subtle text-foreground shadow-card'
  if (isBuiltInWxpayMethod(type)) return 'border-success/30 bg-success-subtle text-foreground shadow-card'
  if (type === 'stripe') return 'border-brand/30 bg-info-subtle text-foreground shadow-card'
  if (type === 'airwallex') return 'border-warning/30 bg-warning-subtle text-foreground shadow-card'
  return 'border-brand/30 bg-info-subtle text-foreground shadow-card'
}
</script>
