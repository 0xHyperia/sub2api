<template>
  <fieldset>
    <legend class="mb-2 text-sm font-medium text-foreground">
      {{ t('payment.paymentMethod') }}
    </legend>
    <div
      class="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3"
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
          'relative flex min-h-[56px] min-w-0 items-center rounded-control border px-3 py-2 text-left transition-colors',
          !method.available
            ? 'cursor-not-allowed border-outline bg-surface-subtle opacity-50'
            : selected === method.type
              ? methodSelectedClass(method.type)
              : 'border-outline-strong bg-surface text-foreground-muted hover:border-focus hover:text-foreground',
        ]"
        @click="method.available && emit('select', method.type)"
        @keydown="handleRadioKeydown($event, index)"
      >
        <span class="flex min-w-0 items-center gap-2.5">
          <img :src="methodIcon(method.type)" alt="" class="h-7 w-7 shrink-0 object-contain" aria-hidden="true" />
          <span class="flex min-w-0 flex-col items-start gap-1 leading-none">
            <span class="break-words text-sm font-semibold leading-5">{{ methodLabel(method) }}</span>
            <span
              v-if="method.fee_rate > 0"
              class="text-xs text-foreground-subtle"
            >
              {{ t('payment.fee') }} {{ method.fee_rate }}%
            </span>
          </span>
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
  if (isBuiltInAlipayMethod(type)) return 'border-[#02A9F1] bg-info-subtle text-foreground'
  if (isBuiltInWxpayMethod(type)) return 'border-[#09BB07] bg-success-subtle text-foreground'
  if (type === 'stripe') return 'border-[#676BE5] bg-surface-subtle text-foreground'
  if (type === 'airwallex') return 'border-[#FF6B3D] bg-warning-subtle text-foreground'
  return 'border-brand bg-info-subtle text-foreground'
}
</script>
