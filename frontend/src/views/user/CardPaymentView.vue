<template>
  <div :class="embedded ? 'space-y-6' : 'mx-auto max-w-6xl space-y-6 p-4 sm:p-6'">
    <div v-if="!embedded" class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('payment.card.title') }}</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('payment.card.subtitle') }}</p>
      </div>
      <button class="btn btn-secondary" :disabled="loading" @click="loadCheckoutInfo">
        {{ t('common.refresh') }}
      </button>
    </div>

    <PaymentStatusPanel
      v-if="createdOrder"
      :order-id="createdOrder.order_id"
      :qr-code="createdOrder.qr_code || ''"
      :pay-url="createdOrder.pay_url"
      :expires-at="createdOrder.expires_at"
      :payment-type="createdOrder.payment_type || 'ldxp'"
      :order-type="'card'"
      currency="CNY"
      @done="createdOrder = null"
    />

    <div v-else class="grid gap-4 lg:grid-cols-[280px_minmax(0,1fr)]">
      <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
        <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.card.shop') }}</h2>
        <div v-if="loading" class="mt-4 text-sm text-gray-500">{{ t('common.loading') }}</div>
        <div v-else-if="shops.length === 0" class="mt-4 text-sm text-gray-500">
          {{ t('payment.card.empty') }}
        </div>
        <div v-else class="mt-3 space-y-2">
          <button
            v-for="shop in shops"
            :key="shop.provider_instance_id"
            type="button"
            class="w-full rounded-md border px-3 py-2 text-left text-sm transition"
            :class="selectedShop?.provider_instance_id === shop.provider_instance_id
              ? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-900/20 dark:text-primary-300'
              : 'border-gray-200 text-gray-700 hover:border-gray-300 dark:border-dark-600 dark:text-gray-300'"
            @click="selectShop(shop.provider_instance_id)"
          >
            <div class="font-medium">{{ shop.shop?.nickname || shop.name }}</div>
            <div class="mt-0.5 truncate text-xs opacity-75">{{ shop.shop?.token }}</div>
          </button>
        </div>
      </section>

      <section class="space-y-4">
        <div class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex flex-wrap gap-2">
            <button
              type="button"
              class="rounded-md px-3 py-1.5 text-sm transition"
              :class="selectedCategoryId === 0 ? activeChipClass : idleChipClass"
              @click="selectedCategoryId = 0"
            >
              {{ t('common.all') }}
            </button>
            <button
              v-for="category in selectedShop?.categories || []"
              :key="category.id"
              type="button"
              class="rounded-md px-3 py-1.5 text-sm transition"
              :class="selectedCategoryId === category.id ? activeChipClass : idleChipClass"
              @click="selectedCategoryId = category.id"
            >
              {{ category.name }}
            </button>
          </div>
        </div>

        <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          <button
            v-for="goods in filteredGoods"
            :key="goods.goods_key"
            type="button"
            class="flex min-h-[104px] flex-col justify-between rounded-lg border bg-white p-4 text-left transition disabled:cursor-not-allowed disabled:opacity-60 dark:bg-dark-800"
            :class="selectedGoods?.goods_key === goods.goods_key
              ? 'border-primary-500 ring-2 ring-primary-100 dark:border-primary-400 dark:ring-primary-900/40'
              : 'border-gray-200 hover:border-gray-300 dark:border-dark-700'"
            :disabled="goods.stock_count <= 0"
            @click="selectedGoodsKey = goods.goods_key"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="line-clamp-2 min-w-0 text-sm font-semibold leading-5 text-gray-900 dark:text-white">{{ goods.name }}</div>
              <div class="shrink-0 text-base font-semibold text-primary-600">¥{{ goods.price.toFixed(2) }}</div>
            </div>
            <div class="mt-4">
              <span
                class="rounded-full px-2.5 py-1 text-xs font-medium"
                :class="stockStatusClass(goods.stock_count)"
              >
                {{ stockStatusLabel(goods.stock_count) }}
              </span>
            </div>
          </button>
        </div>

        <div class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
          <div v-if="embedded" class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('payment.card.title') }}</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('payment.card.subtitle') }}</p>
            </div>
            <button class="btn btn-secondary" :disabled="loading" @click="loadCheckoutInfo">
              {{ t('common.refresh') }}
            </button>
          </div>
          <div class="grid gap-4 md:grid-cols-4">
            <label class="block">
              <span class="input-label">{{ t('payment.card.quantity') }}</span>
              <input v-model.number="quantity" class="input" type="number" min="1" :max="quantityMax || undefined" />
              <span v-if="quantityError" class="mt-1 block text-xs text-amber-600 dark:text-amber-300">{{ quantityError }}</span>
            </label>
            <label class="block">
              <span class="input-label">{{ t('payment.card.coupon') }}</span>
              <input v-model="couponCode" class="input" type="text" />
            </label>
            <div class="md:col-span-4">
              <span class="input-label">{{ t('payment.card.channel') }}</span>
              <div class="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
                <button
                  v-for="channel in enabledChannels"
                  :key="channel.id"
                  type="button"
                  class="flex min-h-[52px] items-center gap-3 rounded-lg border px-3 py-2 text-left transition"
                  :class="channelId === channel.id
                    ? 'border-primary-500 bg-primary-50 ring-2 ring-primary-100 dark:border-primary-400 dark:bg-primary-900/20 dark:ring-primary-900/40'
                    : 'border-gray-200 hover:border-gray-300 dark:border-dark-700'"
                  @click="channelId = channel.id"
                >
                  <span class="flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-md bg-gray-100 dark:bg-dark-700">
                    <img v-if="channelIconUrl(channel.icon)" :src="channelIconUrl(channel.icon)" alt="" class="h-5 w-5 object-contain" />
                    <span v-else class="text-xs font-semibold text-gray-500">{{ channelInitial(channel) }}</span>
                  </span>
                  <span class="min-w-0">
                    <span class="block truncate text-sm font-medium text-gray-900 dark:text-white">{{ channel.show_name || channel.name }}</span>
                    <span v-if="channel.code" class="block truncate text-xs text-gray-500 dark:text-gray-400">{{ channel.code }}</span>
                  </span>
                </button>
              </div>
            </div>
            <div class="flex items-center justify-between gap-4 rounded-lg border border-gray-100 p-3 dark:border-dark-700 md:col-span-4">
              <span>
                <span class="block text-sm font-medium text-gray-900 dark:text-white">{{ t('payment.card.autoRedeem') }}</span>
                <span class="mt-0.5 block text-xs text-gray-500 dark:text-gray-400">{{ t('payment.card.autoRedeemHint') }}</span>
              </span>
              <button
                type="button"
                role="switch"
                :aria-checked="autoRedeem"
                class="relative h-7 w-12 shrink-0 rounded-full transition"
                :class="autoRedeem ? 'bg-primary-600' : 'bg-gray-300 dark:bg-dark-600'"
                @click="autoRedeem = !autoRedeem"
              >
                <span
                  class="absolute top-1 h-5 w-5 rounded-full bg-white shadow transition"
                  :class="autoRedeem ? 'left-6' : 'left-1'"
                />
              </button>
            </div>
          </div>
          <div class="mt-4 flex flex-col gap-3 border-t border-gray-100 pt-4 dark:border-dark-700 sm:flex-row sm:items-center sm:justify-between">
            <div class="text-sm text-gray-600 dark:text-gray-300">
              {{ t('payment.card.payAmount') }}
              <span class="ml-2 text-xl font-semibold text-gray-900 dark:text-white">¥{{ totalAmount.toFixed(2) }}</span>
            </div>
            <button class="btn btn-primary" :disabled="submitting || !canSubmit" @click="createOrder">
              {{ submitting ? t('common.processing') : t('payment.card.buy') }}
            </button>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { paymentAPI } from '@/api/payment'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import type { CardCheckoutShop, CardGoods, CardPrice, CreateOrderResult } from '@/types/payment'

defineProps<{ embedded?: boolean }>()

const { t } = useI18n()

const loading = ref(false)
const submitting = ref(false)
const shops = ref<CardCheckoutShop[]>([])
const selectedProviderId = ref('')
const selectedCategoryId = ref(0)
const selectedGoodsKey = ref('')
const quantity = ref(1)
const channelId = ref(0)
const couponCode = ref('')
const autoRedeem = ref(true)
const price = ref<CardPrice | null>(null)
const createdOrder = ref<CreateOrderResult | null>(null)

const activeChipClass = 'bg-primary-600 text-white'
const idleChipClass = 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-300'

const selectedShop = computed(() => shops.value.find(shop => shop.provider_instance_id === selectedProviderId.value) || null)
const filteredGoods = computed(() => {
  const goods = selectedShop.value?.goods || []
  if (!selectedCategoryId.value) return goods
  return goods.filter(item => item.category_id === selectedCategoryId.value)
})
const selectedGoods = computed<CardGoods | null>(() => selectedShop.value?.goods.find(item => item.goods_key === selectedGoodsKey.value) || null)
const enabledChannels = computed(() => (selectedShop.value?.channels || []).filter(channel => channel.status === 1))
const totalAmount = computed(() => price.value?.total_amount ?? (selectedGoods.value ? selectedGoods.value.price * Math.max(1, quantity.value || 1) : 0))
const quantityMax = computed(() => {
  if (!selectedGoods.value) return 0
  const limits = [selectedGoods.value.stock_count]
  if (selectedGoods.value.limit_count > 0) limits.push(selectedGoods.value.limit_count)
  return Math.max(0, Math.min(...limits))
})
const stockEnough = computed(() => !!selectedGoods.value && selectedGoods.value.stock_count > 0 && quantity.value <= selectedGoods.value.stock_count)
const limitEnough = computed(() => !selectedGoods.value || selectedGoods.value.limit_count <= 0 || quantity.value <= selectedGoods.value.limit_count)
const quantityError = computed(() => {
  if (!selectedGoods.value || quantity.value <= 0) return ''
  if (selectedGoods.value.stock_count <= 0) return t('payment.card.outOfStock')
  if (!stockEnough.value) return t('payment.card.stockNotEnough')
  if (!limitEnough.value) return t('payment.card.limitExceeded')
  return ''
})
const canSubmit = computed(() => !!selectedShop.value && !!selectedGoods.value && quantity.value > 0 && channelId.value > 0 && stockEnough.value && limitEnough.value)

function unwrapAPI<T>(value: T | { data: T }): T {
  return value && typeof value === 'object' && 'data' in value ? (value as { data: T }).data : value as T
}

async function loadCheckoutInfo() {
  loading.value = true
  try {
    const data = unwrapAPI(await paymentAPI.getCardCheckoutInfo())
    shops.value = data.shops || []
    if (!selectedProviderId.value && shops.value.length) {
      selectShop(shops.value[0].provider_instance_id)
    }
  } finally {
    loading.value = false
  }
}

function selectShop(providerId: string) {
  selectedProviderId.value = providerId
  selectedCategoryId.value = 0
  selectedGoodsKey.value = selectedShop.value?.goods.find(goods => goods.stock_count > 0)?.goods_key || selectedShop.value?.goods[0]?.goods_key || ''
  channelId.value = enabledChannels.value[0]?.id || 0
}

function channelIconUrl(icon?: string) {
  const raw = String(icon || '').trim()
  if (!raw) return ''
  if (raw.startsWith('//')) return `https:${raw}`
  if (/^https?:\/\//i.test(raw)) return raw
  return raw
}

function channelInitial(channel: { show_name?: string; name: string }) {
  return (channel.show_name || channel.name || '?').trim().slice(0, 1).toUpperCase()
}

function stockStatusLabel(stock: number) {
  if (stock >= 20) return t('payment.card.stockPlenty')
  if (stock > 5) return t('payment.card.stockNormal')
  if (stock > 0) return t('payment.card.stockLow')
  return t('payment.card.stockEmpty')
}

function stockStatusClass(stock: number) {
  if (stock >= 20) return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300'
  if (stock > 5) return 'bg-sky-50 text-sky-700 dark:bg-sky-900/20 dark:text-sky-300'
  if (stock > 0) return 'bg-amber-50 text-amber-700 dark:bg-amber-900/20 dark:text-amber-300'
  return 'bg-red-50 text-red-700 dark:bg-red-900/20 dark:text-red-300'
}

async function previewPrice() {
  if (!canSubmit.value || !selectedGoods.value) {
    price.value = null
    return
  }
  price.value = unwrapAPI(await paymentAPI.getCardPrice({
    provider_instance_id: selectedProviderId.value,
    goods_key: selectedGoods.value.goods_key,
    quantity: Math.max(1, quantity.value || 1),
    channel_id: channelId.value,
    coupon_code: couponCode.value,
  }))
}

async function createOrder() {
  if (!canSubmit.value || !selectedGoods.value) return
  submitting.value = true
  try {
    const order = unwrapAPI(await paymentAPI.createCardOrder({
      provider_instance_id: selectedProviderId.value,
      goods_key: selectedGoods.value.goods_key,
      quantity: Math.max(1, quantity.value || 1),
      channel_id: channelId.value,
      coupon_code: couponCode.value,
      auto_redeem: autoRedeem.value,
      return_url: window.location.href,
    }))
    createdOrder.value = order
    if (order.pay_url) {
      window.open(order.pay_url, '_blank', 'noopener,noreferrer')
    }
  } finally {
    submitting.value = false
  }
}

watch([selectedProviderId, selectedGoodsKey, quantity, channelId, couponCode], () => {
  void previewPrice()
})

watch(selectedGoods, () => {
  if (quantityMax.value > 0 && quantity.value > quantityMax.value) {
    quantity.value = quantityMax.value
  }
  if (quantity.value <= 0) {
    quantity.value = 1
  }
})

onMounted(loadCheckoutInfo)
</script>
